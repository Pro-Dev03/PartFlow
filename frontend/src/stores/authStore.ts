import { create } from 'zustand';
import { createJSONStorage, persist } from 'zustand/middleware';
import { authApi } from '../services/api/endpoints';
import { apiClient } from '../services/api/client';
import { TokenManager } from '../lib/token-manager';
import { User } from '../types/models';
import { getCloudApiUrl, getConnectionMode } from '../lib/config/app';
import { isNetworkError } from '../lib/error-messages';
import { saveAutoLogoutReason, type AutoLogoutReason } from '../features/auth/sessionReason';

interface AuthState {
  isAuthenticated: boolean;
  sessionVerified: boolean;
  user: User | null;
  token: string | null;
  refreshTokenValue: string | null;
  cloudToken: string | null;
  loginError: string | null;
  isLoading: boolean;
  isPostLoginVerifying: boolean;
  setPostLoginVerifying: (value: boolean) => void;
  login: (email: string, password: string) => Promise<void>;
  logout: () => void;
  checkAuth: () => Promise<void>;
  refreshToken: () => Promise<void>;
}

let refreshInterval: ReturnType<typeof setTimeout> | null = null;
let cloudValidationInFlight: Promise<boolean> | null = null;
const CLOUD_VALIDATION_GRACE_MS = 72 * 60 * 60 * 1000;
const CLOUD_LAST_VALIDATED_AT_KEY = 'partflow-cloud-last-validated-at';

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
  localStorage.removeItem(CLOUD_LAST_VALIDATED_AT_KEY);
  localStorage.removeItem('partflow-user-phone');

  try {
    useAuthStore.persist?.clearStorage();
  } catch {
    // Safe no-op if persistence is not initialized yet.
  }
}

function getPersistedAuthState() {
  if (typeof window === 'undefined') return null;

  try {
    const raw = localStorage.getItem('auth-storage');
    if (!raw) return null;

    const parsed = JSON.parse(raw);
    return parsed?.state ?? null;
  } catch {
    return null;
  }
}

function hasCloudValidationGrace(): boolean {
  if (typeof window === 'undefined') return false;
  const lastValidatedAt = Number(localStorage.getItem(CLOUD_LAST_VALIDATED_AT_KEY));
  return Number.isFinite(lastValidatedAt)
    && lastValidatedAt > 0
    && Date.now() - lastValidatedAt <= CLOUD_VALIDATION_GRACE_MS;
}

function rememberCloudValidation(): void {
  if (typeof window !== 'undefined') {
    localStorage.setItem(CLOUD_LAST_VALIDATED_AT_KEY, String(Date.now()));
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

async function classifyLocalLoginFailure(email: string, password: string): Promise<boolean> {
  try {
    await authApi.login(email, password);
    return false;
  } catch (error) {
    return shouldRedirectToSubscriptionExpired(error);
  }
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
          'X-PartFlow-Cloud-Token': token,
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
          const rejectedPayload = await response.clone().json().catch(() => ({}));
          const rejectionCode = rejectedPayload?.code || rejectedPayload?.error?.code || rejectedPayload?.data?.code;
          if (rejectionCode !== 'SUBSCRIPTION_EXPIRED') {
            return hasCloudValidationGrace();
          }
          stopTokenRefresh();
          apiClient.logout();
          TokenManager.clearToken();
          TokenManager.clearRefreshToken();
          useAuthStore.setState({
            isAuthenticated: false,
            sessionVerified: false,
            user: null,
            token: null,
            refreshTokenValue: null,
            cloudToken: null,
            isLoading: false,
          });
          clearPersistedAuthStorage();
          goToSubscriptionExpiredPage();
        } else if (response.status === 401) {
          // A single unauthorized response may be caused by token rotation or
          // a transient cloud session problem. Keep the session until the
          // server explicitly confirms subscription expiry or the connection
          // is lost.
          return hasCloudValidationGrace();
        }
        return hasCloudValidationGrace();
      }

      const payload = await response.json();
      const data = payload?.data ?? payload;
      const nextToken = data?.token || data?.access_token || cloudToken;
      const nextRefreshToken = data?.refresh_token || localStorage.getItem('cloud_refresh_token');
      if (nextToken) {
        localStorage.setItem('cloud_token', nextToken);
        apiClient.setCloudToken(nextToken);
        useAuthStore.setState({ cloudToken: nextToken });
      }
      if (nextRefreshToken) {
        localStorage.setItem('cloud_refresh_token', nextRefreshToken);
      }

      const currentUser = useAuthStore.getState().user;
      const cloudUser = data?.user as Partial<User> | undefined;
      const storedPhone = localStorage.getItem('partflow-user-phone') || undefined;
      const mergedUser = currentUser && cloudUser
        ? { ...currentUser, ...cloudUser, phone: cloudUser.phone ?? currentUser.phone ?? storedPhone }
        : cloudUser
          ? { ...cloudUser, phone: cloudUser.phone ?? storedPhone }
          : currentUser;
      if (mergedUser?.phone) {
        localStorage.setItem('partflow-user-phone', mergedUser.phone);
      }
      useAuthStore.setState({
        user: mergedUser,
        isAuthenticated: true,
        sessionVerified: true,
      });
      rememberCloudValidation();
      goToAppDashboard();
      return true;
    } catch {
      // A cloud outage is tolerated only during the bounded grace period.
      return hasCloudValidationGrace();
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

function goToAppDashboard(): void {
  if (typeof window === 'undefined' || !window.location.hash.includes('/subscription-expired')) {
    return;
  }

  window.location.hash = '#/app/dashboard';
  window.dispatchEvent(new HashChangeEvent('hashchange'));
}

export function forceLogoutToLogin(reason = 'Session expired') {
  if (typeof window === 'undefined') return;

  const logoutReason: AutoLogoutReason = reason.includes('Internet')
    ? 'offline'
    : reason.includes('subscription') || reason.includes('اشتراك')
      ? 'subscription'
      : reason.includes('No active')
        ? 'missing-session'
        : reason.includes('Cloud')
          ? 'cloud-rejected'
          : 'session-expired';
  saveAutoLogoutReason(logoutReason);
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

  const currentUrl = new URL(window.location.href);
  const shouldResetPath = currentUrl.pathname !== '/' || !currentUrl.hash.includes('/login');

  if (shouldResetPath) {
    currentUrl.pathname = '/';
    currentUrl.hash = '#/login';
    window.history.replaceState({}, '', currentUrl.toString());
  } else if (!window.location.hash.includes('/login')) {
    window.location.hash = '#/login';
  }

  if (reason !== 'No active cloud session') {
    console.warn('Clearing stale auth session:', reason);
  }
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
      loginError: null,
      isLoading: false,
        isPostLoginVerifying: false,
        setPostLoginVerifying: (value: boolean) => set({ isPostLoginVerifying: value }),

          login: async (email: string, password: string) => {
        clearPersistedAuthStorage();
        apiClient.logout();
        set({ isLoading: true, loginError: null });
        try {
              const connectionMode = getConnectionMode();
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

          if (connectionMode === 'cloud') {
            TokenManager.setToken(accessToken);
            if (cloudRefreshToken) {
              TokenManager.setRefreshToken(cloudRefreshToken);
            }
            localStorage.setItem('cloud_token', accessToken);
            apiClient.setCloudToken(accessToken);
            if (cloudRefreshToken) {
              localStorage.setItem('cloud_refresh_token', cloudRefreshToken);
            }
            apiClient.setToken(accessToken);

            const valid = await validateSubscriptionWithCloud();
            if (!valid) {
              throw new Error('Cloud subscription verification failed');
            }

            set({
              isAuthenticated: true,
              sessionVerified: true,
              user: user || null,
              token: accessToken,
              refreshTokenValue: cloudRefreshToken || null,
              cloudToken: accessToken,
              isLoading: false,
            });

            startTokenRefresh();
            return;
          }

          // Local mode authenticates with the cloud authority, then creates a
          // separate local SQLite session for business operations.
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
          apiClient.setCloudToken(accessToken);
          if (cloudRefreshToken) {
            localStorage.setItem('cloud_refresh_token', cloudRefreshToken);
          }
          apiClient.setToken(localToken);

          const valid = await validateSubscriptionWithCloud();
          if (!valid) {
            throw new Error('Cloud subscription verification failed');
          }

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
          return;
        } catch (error) {
          if (shouldRedirectToSubscriptionExpired(error)) {
            goToSubscriptionExpiredPage();
          } else if (error && typeof error === 'object' && (error as { status?: number }).status === 401) {
            if (await classifyLocalLoginFailure(email, password)) {
              goToSubscriptionExpiredPage();
            }
          }
          if (!window.location.hash.includes('/subscription-expired')) {
            stopTokenRefresh();
            apiClient.logout();
            TokenManager.clearToken();
            TokenManager.clearRefreshToken();
            localStorage.removeItem('cloud_token');
            localStorage.removeItem('cloud_refresh_token');
          }
          set({ isAuthenticated: false, sessionVerified: false, user: null, token: null, refreshTokenValue: null, isLoading: false });
          set({
            loginError: shouldRedirectToSubscriptionExpired(error)
              ? null
              : isNetworkError(error)
                ? 'connection error'
                : 'invalid credentials',
          });
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
        apiClient.clearCloudToken();
        apiClient.clearCloudToken();
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

        const persistedState = getPersistedAuthState();
        const currentState = useAuthStore.getState();
        const token = currentState.token || persistedState?.token || TokenManager.getToken();
        const refreshToken = currentState.refreshTokenValue || persistedState?.refreshTokenValue || TokenManager.getRefreshToken();
        const cloudToken = currentState.cloudToken || persistedState?.cloudToken || (typeof window !== 'undefined' ? localStorage.getItem('cloud_token') : null);

        if (token && !TokenManager.getToken()) {
          TokenManager.setToken(token);
        }
        if (refreshToken && !TokenManager.getRefreshToken()) {
          TokenManager.setRefreshToken(refreshToken);
        }
        if (cloudToken && !localStorage.getItem('cloud_token')) {
          localStorage.setItem('cloud_token', cloudToken);
        }
        if (cloudToken) {
          apiClient.setCloudToken(cloudToken);
        }

        if (!token) {
          forceLogoutToLogin('No active cloud session');
          set({ isAuthenticated: false, sessionVerified: false, user: null, token: null, refreshTokenValue: null, isLoading: false });
          return;
        }

        apiClient.setToken(token);

        if (!navigator.onLine) {
          forceLogoutToLogin('Internet connection is required');
          set({ isAuthenticated: false, sessionVerified: false, user: null, token: null, refreshTokenValue: null, isLoading: false });
          return;
        }

        const valid = await validateSubscriptionWithCloud();
        if (!valid) {
          // A transient cloud failure must not sign the owner out. Explicit
          // invalid-token and subscription-expired responses handle logout.
          set({ isAuthenticated: true, sessionVerified: true, isLoading: false });
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
          const cloudValid = await validateSubscriptionWithCloud();
          if (!cloudValid) {
            throw new Error('Cloud subscription verification failed');
          }
          const currentUser = useAuthStore.getState().user;
          const storedPhone = localStorage.getItem('partflow-user-phone') || undefined;
          const mergedUser = currentUser
            ? { ...currentUser, ...user, phone: user.phone ?? currentUser.phone ?? storedPhone }
            : { ...user, phone: user.phone ?? storedPhone };
          if (mergedUser.phone) {
            localStorage.setItem('partflow-user-phone', mergedUser.phone);
          }
          set({
            isAuthenticated: true,
            sessionVerified: true,
            user: mergedUser,
            token,
            refreshTokenValue: refreshToken || null,
          });
        } catch (error) {
          console.error('Failed to refresh token:', error);

          if (shouldRedirectToSubscriptionExpired(error, window.location.pathname)) {
            goToSubscriptionExpiredPage();
            return;
          }

          // Invalid or expired credentials can fail a refresh without proving
          // subscription expiry. Preserve the session and let the next online
          // validation or explicit login resolve it.
          throw error;
        }
      },
    }),
    {
      name: 'auth-storage',
      storage: createJSONStorage(() => authStorage),
      // Persist the authenticated session state so a browser reload can restore
      // the user's local JWT and cloud token, then immediately re-verify the
      // cloud subscription authority before allowing protected routes.
      partialize: (state) => ({
        isAuthenticated: state.isAuthenticated,
        sessionVerified: state.sessionVerified,
        user: state.user,
        token: state.token,
        refreshTokenValue: state.refreshTokenValue,
        cloudToken: state.cloudToken,
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
