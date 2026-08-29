import { create } from 'zustand';
import { createJSONStorage, persist } from 'zustand/middleware';
import { authApi } from '../services/api/endpoints';
import { apiClient } from '../services/api/client';
import { TokenManager } from '../lib/token-manager';
import { clearSubscriptionGuard, isOfflineSubscriptionBlocked, syncSubscriptionGuard } from '../lib/subscription-guard';
import { User } from '../types/models';

interface AuthState {
  isAuthenticated: boolean;
  user: User | null;
  token: string | null;
  refreshTokenValue: string | null;
  isLoading: boolean;
  login: (email: string, password: string) => Promise<void>;
  logout: () => void;
  checkAuth: () => void;
  refreshToken: () => Promise<void>;
}

let refreshInterval: ReturnType<typeof setTimeout> | null = null;

const authStorage = {
  getItem: (name: string) => localStorage.getItem(name),
  setItem: (name: string, value: string) => {
    if (name === 'auth-storage') {
      try {
        const parsed = JSON.parse(value) as { state?: { isAuthenticated?: boolean; token?: string | null } };
        if (!parsed?.state?.isAuthenticated || !parsed.state.token) {
          localStorage.removeItem(name);
          return;
        }
      } catch {
        localStorage.removeItem(name);
        return;
      }
    }

    localStorage.setItem(name, value);
  },
  removeItem: (name: string) => localStorage.removeItem(name),
};

function clearPersistedAuthStorage() {
  if (typeof window === 'undefined') return;

  localStorage.removeItem('auth-storage');
  localStorage.removeItem('auth_token');
  localStorage.removeItem('token');
  localStorage.removeItem('refresh_token');

  try {
    useAuthStore.persist?.clearStorage();
  } catch {
    // Safe no-op if persistence is not initialized yet.
  }
}

export function shouldRedirectToSubscriptionExpired(error: unknown, pathname = window.location.pathname): boolean {
  if (!error || typeof error !== 'object') return false;

  const anyError = error as {
    status?: number;
    code?: string;
    message?: string;
    response?: { status?: number; error?: { code?: string; message?: string } };
  };

  const status = anyError.status ?? anyError.response?.status;
  const code = anyError.code ?? anyError.response?.error?.code ?? '';
  const message = (anyError.message ?? anyError.response?.error?.message ?? '').toLowerCase();
  const combined = `${status ?? ''} ${code} ${message}`.toLowerCase();

  const isSubscriptionExpired =
    status === 403 ||
    code === 'SUBSCRIPTION_EXPIRED' ||
    /انتهت.*اشتراك|subscription.*expired|expired.*subscription|subscription.*ended|اشتراك.*منتهي|اشتراك.*منتهية/.test(combined);

  return isSubscriptionExpired && !pathname.includes('/subscription-expired');
}

function goToSubscriptionExpiredPage(): void {
  if (typeof window === 'undefined') return;

  if (window.location.pathname.includes('/subscription-expired')) {
    return;
  }

  try {
    window.history.pushState({}, '', '/subscription-expired');
    window.dispatchEvent(new PopStateEvent('popstate'));
  } catch {
    window.location.href = '/subscription-expired';
  }
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      isAuthenticated: false,
      user: null,
      token: null,
      refreshTokenValue: null,
      isLoading: false,

      login: async (email: string, password: string) => {
        set({ isLoading: true });
        try {
          const response = await authApi.login(email, password);
          const data = response.data as {
            user: User;
            token: string;
            refresh_token?: string;
            subscription_status?: string;
            subscription_expires_at?: string | null;
          };
          const { user, token } = data;
          const refreshToken = data.refresh_token || TokenManager.getRefreshToken();

          // Use TokenManager for consistent token storage
          TokenManager.setToken(token);
          if (refreshToken) {
            TokenManager.setRefreshToken(refreshToken);
          }
          apiClient.setToken(token);

          syncSubscriptionGuard(token, {
            userId: user.id,
            email: user.email,
            subscriptionStatus: data.subscription_status || 'active',
            subscriptionExpiresAt: data.subscription_expires_at || null,
          });

          set({
            isAuthenticated: true,
            user,
            token,
            refreshTokenValue: refreshToken,
            isLoading: false
          });

          // Start auto-refresh
          startTokenRefresh();
        } catch (error) {
          set({ isLoading: false });
          throw error;
        }
      },

      logout: () => {
        // Stop auto-refresh
        stopTokenRefresh();
        TokenManager.clearToken();
        TokenManager.clearRefreshToken();
        apiClient.logout();
        clearPersistedAuthStorage();
        clearSubscriptionGuard();
        set({
          isAuthenticated: false,
          user: null,
          token: null,
          refreshTokenValue: null,
        });
      },

      checkAuth: () => {
        const token = TokenManager.getToken();
        if (token) {
          apiClient.setToken(token);

          if (isOfflineSubscriptionBlocked(token)) {
            goToSubscriptionExpiredPage();
            return;
          }

          void authApi.refreshToken()
            .then((response) => {
              const data = response.data as { user?: User; token?: string; subscription_status?: string; subscription_expires_at?: string | null };
              const liveToken = data.token || token;
              const activeUser = data.user || get().user;

              syncSubscriptionGuard(liveToken, {
                userId: activeUser?.id,
                email: activeUser?.email,
                subscriptionStatus: data.subscription_status || activeUser?.subscription_status || 'active',
                subscriptionExpiresAt: data.subscription_expires_at || activeUser?.subscription_expires_at || null,
              });

              const nextRefreshToken = data.refresh_token || TokenManager.getRefreshToken();
              if (nextRefreshToken) {
                TokenManager.setRefreshToken(nextRefreshToken);
              }

              set({
                isAuthenticated: true,
                user: activeUser || null,
                token: liveToken,
                refreshTokenValue: nextRefreshToken || null,
              });

              startTokenRefresh();
            })
            .catch((error) => {
              if (shouldRedirectToSubscriptionExpired(error, window.location.pathname)) {
                goToSubscriptionExpiredPage();
                return;
              }

              // Generic refresh failures (network/server issues) should not destroy a valid session.
              // Keep the existing auth state and let the user continue to use the app.
              if (!window.location.pathname.includes('/login') && !window.location.pathname.includes('/subscription-expired')) {
                return;
              }

              // If the user is already on login screen, we can clear state there.
              clearPersistedAuthStorage();
              clearSubscriptionGuard();
              set({
                isAuthenticated: false,
                user: null,
                token: null,
                refreshTokenValue: null,
              });
            });
          return;
        }

        clearPersistedAuthStorage();
        clearSubscriptionGuard();
        set({
          isAuthenticated: false,
          user: null,
          token: null,
          refreshTokenValue: null,
        });
      },

      refreshToken: async () => {
        try {
          const response = await authApi.refreshToken();
          const data = response.data as { user: User; token: string; refresh_token?: string };
          const { user, token } = data;
          const refreshToken = data.refresh_token || TokenManager.getRefreshToken();

          TokenManager.setToken(token);
          if (refreshToken) {
            TokenManager.setRefreshToken(refreshToken);
          }
          apiClient.setToken(token);
          set({
            isAuthenticated: true,
            user,
            token,
            refreshTokenValue: refreshToken || null,
          });
        } catch (error) {
          console.error('Failed to refresh token:', error);

          if (shouldRedirectToSubscriptionExpired(error, window.location.pathname)) {
            goToSubscriptionExpiredPage();
            return;
          }

          // For transient refresh failures, preserve the current session instead of forcing logout.
          return;
        }
      },
    }),
    {
      name: 'auth-storage',
      storage: createJSONStorage(() => authStorage),
      partialize: (state) => ({
        isAuthenticated: state.isAuthenticated,
        user: state.user,
        token: state.token,
        refreshTokenValue: state.refreshTokenValue,
      }),
    }
  )
);

// Auto-refresh token every 10 minutes
function startTokenRefresh() {
  stopTokenRefresh(); // Clear any existing interval
  
  refreshInterval = setInterval(async () => {
    const { isAuthenticated, token } = useAuthStore.getState();
    if (isAuthenticated && token) {
      try {
        await useAuthStore.getState().refreshToken();
      } catch (error) {
        console.error('Auto-refresh failed:', error);
      }
    }
  }, 10 * 60 * 1000); // 10 minutes
}

function stopTokenRefresh() {
  if (refreshInterval) {
    clearInterval(refreshInterval);
    refreshInterval = null;
  }
}