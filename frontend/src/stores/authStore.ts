import { create } from 'zustand';
import { createJSONStorage, persist } from 'zustand/middleware';
import { authApi } from '../services/api/endpoints';
import { apiClient } from '../services/api/client';
import { TokenManager } from '../lib/token-manager';
import { User } from '../types/models';
import { getCloudApiUrl } from '../lib/config/app';

interface AuthState {
  isAuthenticated: boolean;
  user: User | null;
  token: string | null;
  refreshTokenValue: string | null;
  isLoading: boolean;
  login: (email: string, password: string) => Promise<void>;
  logout: () => void;
  checkAuth: () => Promise<void>;
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

/**
 * Validate the local session against the cloud authority when connectivity is
 * restored. This intentionally bypasses the active local API URL: Desktop can
 * continue using SQLite for business operations while subscription authority
 * remains on Render.
 */
export async function validateSubscriptionWithCloud(): Promise<boolean> {
  if (typeof window === 'undefined' || !navigator.onLine) return false;

  const token = TokenManager.getToken();
  if (!token) return false;

  try {
    const response = await fetch(`${getCloudApiUrl()}/auth/validate`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify({}),
    });

    if (!response.ok) {
      if (response.status === 403) {
        clearPersistedAuthStorage();
        goToSubscriptionExpiredPage();
      } else if (response.status === 401) {
        forceLogoutToLogin('Cloud session is no longer valid');
      }
      return false;
    }

    const payload = await response.json();
    const data = payload?.data ?? payload;
    const nextToken = data?.token || data?.access_token || token;
    const nextRefreshToken = data?.refresh_token || TokenManager.getRefreshToken();
    if (nextToken) {
      TokenManager.setToken(nextToken);
      apiClient.setToken(nextToken);
    }
    if (nextRefreshToken) TokenManager.setRefreshToken(nextRefreshToken);

    const currentUser = useAuthStore.getState().user;
    useAuthStore.setState({
      token: nextToken,
      refreshTokenValue: nextRefreshToken || null,
      user: data?.user || currentUser,
      isAuthenticated: true,
    });
    return true;
  } catch {
    return false;
  }
}

function goToSubscriptionExpiredPage(): void {
  if (typeof window === 'undefined') return;

  if (window.location.hash.includes('/subscription-expired')) {
    return;
  }

  try {
    window.history.pushState({}, '', '/subscription-expired');
    window.location.hash = '#/subscription-expired';
    window.dispatchEvent(new HashChangeEvent('hashchange'));
  } catch {
    const currentUrl = new URL(window.location.href);
    currentUrl.hash = '#/subscription-expired';
    window.location.href = currentUrl.toString();
  }
}

function isInvalidTokenError(error: unknown): boolean {
  if (!error || typeof error !== 'object') return false;

  const anyError = error as {
    message?: string;
    code?: string;
    status?: number;
    response?: { status?: number; error?: { code?: string; message?: string } };
  };

  const combined = [
    anyError.message,
    anyError.code,
    anyError.response?.error?.code,
    anyError.response?.error?.message,
  ].filter(Boolean).join(' ').toLowerCase();

  return (
    anyError.status === 401 ||
    anyError.response?.status === 401 ||
    /invalid token|signature is invalid|token expired|jwt|unauthorized/.test(combined)
  );
}

function forceLogoutToLogin(reason = 'Session expired') {
  if (typeof window === 'undefined') return;

  clearPersistedAuthStorage();
  if (!window.location.hash.includes('/login')) {
    window.location.hash = '#/login';
  }

  console.warn('Clearing stale auth session:', reason);
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
        clearPersistedAuthStorage();
        set({ isLoading: true });
        try {
          const data = await authApi.loginWithCloud(email, password) as {
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
        set({
          isAuthenticated: false,
          user: null,
          token: null,
          refreshTokenValue: null,
        });
      },

      checkAuth: async () => {
        const token = TokenManager.getToken();
        if (!token || !navigator.onLine) {
          forceLogoutToLogin('Cloud verification requires an internet connection');
          set({ isAuthenticated: false, user: null, token: null, refreshTokenValue: null, isLoading: false });
          return;
        }

        apiClient.setToken(token);
        const valid = await validateSubscriptionWithCloud();
        if (!valid) {
          if (!window.location.hash.includes('/subscription-expired')) {
            forceLogoutToLogin('Cloud verification failed');
          }
          set({ isAuthenticated: false, user: null, token: null, refreshTokenValue: null, isLoading: false });
          return;
        }

        startTokenRefresh();
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

          if (isInvalidTokenError(error)) {
            forceLogoutToLogin('Refresh failed due to invalid token');
            set({
              isAuthenticated: false,
              user: null,
              token: null,
              refreshTokenValue: null,
            });
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
