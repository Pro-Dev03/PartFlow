import { create } from 'zustand';
import { createJSONStorage, persist } from 'zustand/middleware';
import { authApi } from '../services/api/endpoints';
import { apiClient } from '../services/api/client';
import { TokenManager } from '../lib/token-manager';
import { User } from '../types/models';
import { getCloudApiUrl } from '../lib/config/app';

interface AuthState {
  isAuthenticated: boolean;
  sessionVerified: boolean;
  user: User | null;
  token: string | null;
  refreshTokenValue: string | null;
  cloudToken: string | null;
  isLoading: boolean;
  login: (email: string, password: string) => Promise<void>;
  logout: () => void;
  checkAuth: () => Promise<void>;
  refreshToken: () => Promise<void>;
}

let refreshInterval: ReturnType<typeof setTimeout> | null = null;
let cloudValidationInFlight: Promise<boolean> | null = null;

const authStorage = {
  getItem: (name: string) => localStorage.getItem(name),
  setItem: (name: string, value: string) => {
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
  localStorage.removeItem('cloud_token');
  localStorage.removeItem('cloud_refresh_token');

  try {
    useAuthStore.persist?.clearStorage();
  } catch {
    // Safe no-op if persistence is not initialized yet.
  }
}

export function shouldRedirectToSubscriptionExpired(error: unknown, pathname = window.location.pathname): boolean {
  if (!error || typeof error !== 'object') return false;
  if (pathname.includes('/subscription-expired')) return false;

  const anyError = error as {
    status?: number;
    code?: string;
    message?: string;
    response?: { status?: number; error?: { code?: string; message?: string } };
  };

  const code = String(anyError.code ?? anyError.response?.error?.code ?? '');
  const message = (anyError.message ?? anyError.response?.error?.message ?? '').toLowerCase();
  const combined = `${code} ${message}`.toLowerCase();

  if (code === 'SUBSCRIPTION_EXPIRED') return true;
  if (code === 'CLOUD_AUTH_REQUIRED') {
    return anyError.status === 403 || anyError.response?.status === 403;
  }

  return /subscription.*expired|expired.*subscription|اشتراك.*منتهي|اشتراك.*منتهية/.test(combined);
}

/**
 * Validate the local session against the cloud authority when connectivity is
 * restored. This intentionally bypasses the active local API URL: Desktop can
 * continue using SQLite for business operations while subscription authority
 * remains on Render.
 */
async function refreshCloudAccessToken(): Promise<string | null> {
  if (typeof window === 'undefined') return null;
  const refreshToken = localStorage.getItem('cloud_refresh_token');
  if (!refreshToken) return null;

  const refreshResponse = await fetch(`${getCloudApiUrl()}/auth/refresh`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ refresh_token: refreshToken }),
  });
  const refreshData = await refreshResponse.json().catch(() => ({}));
  if (!refreshResponse.ok) return null;

  const payload = refreshData?.data && typeof refreshData.data === 'object'
    ? refreshData.data
    : refreshData;
  const nextToken = payload?.access_token || payload?.token;
  if (!nextToken) return null;

  localStorage.setItem('cloud_token', nextToken);
  const nextRefresh = payload?.refresh_token || payload?.refreshToken;
  if (nextRefresh) {
    localStorage.setItem('cloud_refresh_token', nextRefresh);
  }
  useAuthStore.setState({ cloudToken: nextToken });
  return nextToken as string;
}

export async function validateSubscriptionWithCloud(): Promise<boolean> {
  if (typeof window === 'undefined') return false;

  const sessionToken = localStorage.getItem('cloud_token');
  if (!sessionToken) return false;
  let cloudToken = sessionToken;

  if (!navigator.onLine) {
    return false;
  }

  if (cloudValidationInFlight) {
    return cloudValidationInFlight;
  }

  const validation = (async () => {
    try {
      const callValidate = (token: string) => fetch(`${getCloudApiUrl()}/auth/validate`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({}),
      });

      let response = await callValidate(cloudToken);
      if (response.status === 401) {
        const refreshed = await refreshCloudAccessToken();
        if (refreshed) {
          cloudToken = refreshed;
          response = await callValidate(cloudToken);
        }
      }

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
      const nextToken = data?.token || data?.access_token || cloudToken;
      const nextRefreshToken = data?.refresh_token || localStorage.getItem('cloud_refresh_token');
      if (nextToken) {
        localStorage.setItem('cloud_token', nextToken);
        useAuthStore.setState({ cloudToken: nextToken });
      }
      if (nextRefreshToken) {
        localStorage.setItem('cloud_refresh_token', nextRefreshToken);
      }

      const currentUser = useAuthStore.getState().user;
      useAuthStore.setState({
        user: data?.user || currentUser,
        isAuthenticated: true,
        sessionVerified: true,
      });
      return true;
    } catch (error) {
      // Preserve session on transient network/connectivity errors instead of
      // forcing logout. Only explicit cloud rejections should clear auth.
      if (error instanceof TypeError) {
        return true;
      }
      return false;
    }
  })();

  cloudValidationInFlight = validation;
  try {
    return await validation;
  } finally {
    if (cloudValidationInFlight === validation) {
      cloudValidationInFlight = null;
    }
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
    /no refresh token available|invalid refresh token|refresh token expired|invalid token|signature is invalid|token expired|jwt|unauthorized/.test(combined)
  );
}

function forceLogoutToLogin(reason = 'Session expired') {
  if (typeof window === 'undefined') return;

  stopTokenRefresh();
  apiClient.logout();
  TokenManager.clearToken();
  TokenManager.clearRefreshToken();
  localStorage.removeItem('cloud_token');
  localStorage.removeItem('cloud_refresh_token');
  useAuthStore.setState({
    isAuthenticated: false,
    user: null,
    token: null,
    refreshTokenValue: null,
    cloudToken: null,
    sessionVerified: false,
    isLoading: false,
  });
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
      sessionVerified: false,
      user: null,
      token: null,
      refreshTokenValue: null,
      cloudToken: null,
      isLoading: false,

          login: async (email: string, password: string) => {
        clearPersistedAuthStorage();
        apiClient.logout();
        set({ isLoading: true });
        try {
          const data = await authApi.loginWithCloud(email, password) as {
            user: User;
            token?: string;
            access_token?: string;
            refresh_token?: string;
            subscription_status?: string;
            subscription_expires_at?: string | null;
          };
          const { user } = data;
          const accessToken = data.token || data.access_token;
          if (!accessToken) {
            throw new Error('Cloud login response did not include an access token');
          }
          const cloudRefreshToken = data.refresh_token;
          const localSession = await authApi.createLocalSession(accessToken);
          const localData = localSession as {
            user: User;
            token?: string;
            access_token?: string;
            refresh_token?: string;
          };
          const localToken = localData.token || localData.access_token;
          if (!localToken) {
            throw new Error('Local session response did not include an access token');
          }
          const localRefreshToken = localData.refresh_token;

          TokenManager.setToken(localToken);
          if (localRefreshToken) {
            TokenManager.setRefreshToken(localRefreshToken);
          }
          localStorage.setItem('cloud_token', accessToken);
          if (cloudRefreshToken) {
            localStorage.setItem('cloud_refresh_token', cloudRefreshToken);
          }
          apiClient.setToken(localToken);

          set({
            isAuthenticated: true,
            sessionVerified: true,
            user: localData.user || user || null,
            token: localToken,
            refreshTokenValue: localRefreshToken || null,
            cloudToken: accessToken,
            isLoading: false,
          });

          startTokenRefresh();
        } catch (error) {
          set({ isLoading: false });
          throw error;
        }
      },

      logout: () => {
        const state = useAuthStore.getState();
        const cloudToken = state.cloudToken || (typeof window !== 'undefined' ? localStorage.getItem('cloud_token') : null);
        if (cloudToken && (typeof navigator === 'undefined' || navigator.onLine)) {
          void authApi.logoutWithCloud(cloudToken).catch(() => undefined);
        }
        stopTokenRefresh();
        TokenManager.clearToken();
        TokenManager.clearRefreshToken();
        localStorage.removeItem('cloud_token');
        localStorage.removeItem('cloud_refresh_token');
        apiClient.logout();
        set({
          isAuthenticated: false,
          sessionVerified: false,
          user: null,
          token: null,
          refreshTokenValue: null,
          cloudToken: null,
        });
        clearPersistedAuthStorage();
      },

      checkAuth: async () => {
        set({ isLoading: true, sessionVerified: false });
        const token = TokenManager.getToken();
        const cloudToken = typeof window !== 'undefined' ? localStorage.getItem('cloud_token') : null;
        if (!token || !cloudToken) {
          forceLogoutToLogin('No active cloud session');
          set({ isAuthenticated: false, sessionVerified: false, user: null, token: null, refreshTokenValue: null, isLoading: false });
          return;
        }

        if (!navigator.onLine) {
          set({ isAuthenticated: true, sessionVerified: true, isLoading: false });
          return;
        }

        apiClient.setToken(token);
        const valid = await validateSubscriptionWithCloud();
        if (!valid) {
          if (TokenManager.getToken() && !window.location.hash.includes('/subscription-expired')) {
            forceLogoutToLogin('Cloud verification failed');
          }
          set({ isAuthenticated: false, sessionVerified: false, user: null, token: null, refreshTokenValue: null, isLoading: false });
          return;
        }

        set({ isAuthenticated: true, sessionVerified: true, isLoading: false });
        startTokenRefresh();
      },

      refreshToken: async () => {
        try {
          const response = await authApi.refreshToken();
          const payload = (response && typeof response === 'object' && 'data' in response && response.data)
            ? response.data
            : response;
          const data = payload as { user?: User; token?: string; access_token?: string; refresh_token?: string };
          const user = data.user;
          const token = data.token || data.access_token;
          if (!user || !token) {
            throw new Error('Refresh response did not include user and token');
          }
          const refreshToken = data.refresh_token || TokenManager.getRefreshToken();

          TokenManager.setToken(token);
          if (refreshToken) {
            TokenManager.setRefreshToken(refreshToken);
          }
          apiClient.setToken(token);
          set({
            isAuthenticated: true,
            sessionVerified: true,
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
            return;
          }

          // For transient refresh failures, preserve the current session instead
          // of forcing logout, but still reject the operation so an explicit
          // "verify subscription" action can show the real failure. The
          // background timer catches this rejection and keeps the session.
          throw error;
        }
      },
    }),
    {
      name: 'auth-storage',
      storage: createJSONStorage(() => authStorage),
      // The app must re-verify the cloud session on every load. Persisting
      // tokens or auth flags allows stale sessions to be restored and triggers
      // the 401 refresh/login loop after reload.
      partialize: () => ({}),
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
