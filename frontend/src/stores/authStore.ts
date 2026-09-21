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
  cloudVerificationPending: boolean;
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
let cloudValidationRetryTimer: ReturnType<typeof setTimeout> | null = null;
let cloudValidationRetryAttempt = 0;
// Keep this key only to remove the marker written by older builds. It is never
// used to authorize a session after a failed cloud validation.
const CLOUD_LAST_VALIDATED_AT_KEY = 'partflow-cloud-last-validated-at';
const CLOUD_REFRESH_FAILED_KEY = 'partflow-cloud-refresh-failed';
const CLOUD_REFRESH_LOCK_KEY = 'partflow-cloud-refresh-lock';
const CLOUD_REFRESH_LOCK_TTL_MS = 15_000;
const CLOUD_VALIDATION_RETRY_DELAYS_MS = [5_000, 15_000, 30_000, 60_000, 300_000];

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

function cancelCloudValidationRetry(): void {
  if (cloudValidationRetryTimer) {
    clearTimeout(cloudValidationRetryTimer);
    cloudValidationRetryTimer = null;
  }
  cloudValidationRetryAttempt = 0;
}

function scheduleCloudValidationRetry(): void {
  if (cloudValidationRetryTimer || typeof window === 'undefined') return;

  const state = useAuthStore.getState();
  if (!state.isAuthenticated || !state.token || !state.cloudToken) return;

  const delay = CLOUD_VALIDATION_RETRY_DELAYS_MS[Math.min(
    cloudValidationRetryAttempt,
    CLOUD_VALIDATION_RETRY_DELAYS_MS.length - 1,
  )];
  cloudValidationRetryAttempt += 1;
  cloudValidationRetryTimer = setTimeout(() => {
    cloudValidationRetryTimer = null;
    void validateSubscriptionWithCloud();
  }, delay);
}

export function markCloudVerificationPending(): void {
  if (typeof window !== 'undefined' && getConnectionMode() === 'local') {
    clearCloudVerificationPending();
    return;
  }

  const state = useAuthStore.getState();
  const hasActiveSession = Boolean(
    state.isAuthenticated
      || state.token
      || state.cloudToken
      || TokenManager.getToken()
      || (typeof window !== 'undefined' && localStorage.getItem('cloud_token')),
  );
  if (!hasActiveSession) return;

  useAuthStore.setState({
    isAuthenticated: true,
    sessionVerified: true,
    cloudVerificationPending: true,
  });
  scheduleCloudValidationRetry();
}

function clearCloudVerificationPending(): void {
  cancelCloudValidationRetry();
  useAuthStore.setState({ cloudVerificationPending: false });
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
  if (localStorage.getItem(CLOUD_REFRESH_FAILED_KEY) === 'true') return null;

  const lockOwner = `${Date.now()}-${Math.random().toString(36).slice(2)}`;
  const lockIsActive = () => {
    const rawLock = localStorage.getItem(CLOUD_REFRESH_LOCK_KEY);
    if (!rawLock) return false;
    try {
      const lock = JSON.parse(rawLock) as { owner?: string; expiresAt?: number };
      return Boolean(lock.owner && lock.owner !== lockOwner && Number(lock.expiresAt) > Date.now());
    } catch {
      return false;
    }
  };

  const waitForOtherRefresh = async () => {
    const deadline = Date.now() + CLOUD_REFRESH_LOCK_TTL_MS;
    while (lockIsActive() && Date.now() < deadline) {
      await new Promise((resolve) => window.setTimeout(resolve, 250));
    }
  };

  if (lockIsActive()) {
    // Another tab owns the refresh rotation. Its result will be visible in
    // storage; do not submit the same refresh token a second time.
    await waitForOtherRefresh();
    return localStorage.getItem('cloud_token');
  }

  localStorage.setItem(CLOUD_REFRESH_LOCK_KEY, JSON.stringify({
    owner: lockOwner,
    expiresAt: Date.now() + CLOUD_REFRESH_LOCK_TTL_MS,
  }));
  if (lockIsActive()) {
    await waitForOtherRefresh();
    return localStorage.getItem('cloud_token');
  }

  try {
    const refreshResponse = await fetch(`${getCloudApiUrl()}/auth/refresh`, {
      method: 'POST',
      credentials: 'include',
      headers: { 'Content-Type': 'application/json' },
      body: '{}',
    });
    const refreshData = await refreshResponse.json().catch(() => ({}));
    if (!refreshResponse.ok) {
      if (refreshResponse.status === 401) {
        localStorage.setItem(CLOUD_REFRESH_FAILED_KEY, 'true');
        markCloudVerificationPending();
      } else {
        const error = new Error('Cloud token refresh is temporarily unavailable') as Error & { status?: number };
        error.status = refreshResponse.status;
        throw error;
      }
      return null;
    }

    const payload = refreshData?.data && typeof refreshData.data === 'object'
      ? refreshData.data
      : refreshData;
    const nextToken = payload?.access_token || payload?.token;
    if (!nextToken) return null;

    localStorage.setItem('cloud_token', nextToken);
    localStorage.removeItem(CLOUD_REFRESH_FAILED_KEY);
    useAuthStore.setState({ cloudToken: nextToken });
    return nextToken as string;
  } finally {
    try {
      const currentLock = JSON.parse(localStorage.getItem(CLOUD_REFRESH_LOCK_KEY) || '{}') as { owner?: string };
      if (currentLock.owner === lockOwner) {
        localStorage.removeItem(CLOUD_REFRESH_LOCK_KEY);
      }
    } catch {
      localStorage.removeItem(CLOUD_REFRESH_LOCK_KEY);
    }
  }
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
  if (getConnectionMode() === 'local') {
    clearCloudVerificationPending();
    const state = useAuthStore.getState();
    if (state.isAuthenticated || state.token || TokenManager.getToken()) {
      useAuthStore.setState({
        isAuthenticated: true,
        sessionVerified: true,
        cloudVerificationPending: false,
      });
      return true;
    }
    return false;
  }

  const sessionToken = localStorage.getItem('cloud_token');
  if (!sessionToken) {
    const state = useAuthStore.getState();
    if (state.isAuthenticated || state.token || state.cloudToken) {
      markCloudVerificationPending();
    }
    return false;
  }
  let cloudToken = sessionToken;

  if (!navigator.onLine) {
    markCloudVerificationPending();
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
            markCloudVerificationPending();
            return false;
          }
          cancelCloudValidationRetry();
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
            cloudVerificationPending: false,
            isLoading: false,
          });
          clearPersistedAuthStorage();
          goToSubscriptionExpiredPage();
          return false;
        } else if (response.status === 401) {
          markCloudVerificationPending();
          return false;
        }
        // Keep an already-authenticated owner in the app for infrastructure
        // outages. Mutations remain blocked until a later validation succeeds.
        markCloudVerificationPending();
        return false;
      }

      const payload = await response.json();
      const data = payload?.data ?? payload;
      const nextToken = data?.token || data?.access_token || cloudToken;
      if (nextToken) {
        localStorage.setItem('cloud_token', nextToken);
        apiClient.setCloudToken(nextToken);
        useAuthStore.setState({ cloudToken: nextToken });
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
        cloudVerificationPending: false,
      });
      clearCloudVerificationPending();
      localStorage.removeItem(CLOUD_LAST_VALIDATED_AT_KEY);
      goToAppDashboard();
      return true;
    } catch {
      // A network or cloud infrastructure failure is temporary. Keep the
      // authenticated UI available, block mutations, and retry with backoff.
      markCloudVerificationPending();
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

export async function retrySubscriptionVerification(): Promise<boolean> {
  if (typeof window !== 'undefined') {
    localStorage.removeItem(CLOUD_REFRESH_FAILED_KEY);
  }
  if (typeof window !== 'undefined' && !localStorage.getItem('cloud_token')) {
    const refreshedToken = await refreshCloudAccessToken();
    if (refreshedToken) {
      apiClient.setCloudToken(refreshedToken);
    }
  }
  return validateSubscriptionWithCloud();
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
  cancelCloudValidationRetry();
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
    cloudVerificationPending: false,
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
      cloudVerificationPending: false,
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
            localStorage.removeItem(CLOUD_REFRESH_FAILED_KEY);
        apiClient.logout();
        cancelCloudValidationRetry();
        set({ isLoading: true, loginError: null, cloudVerificationPending: false });
        try {
          const connectionMode = getConnectionMode();

          if (connectionMode === 'local') {
            const localData = await authApi.login(email, password) as {
              user: User;
              token?: string;
              access_token?: string;
              refresh_token?: string;
            };
            const localToken = localData.token || localData.access_token;
            if (!localToken) {
              throw new Error('Local login response did not include an access token');
            }

            TokenManager.setToken(localToken);
            apiClient.setToken(localToken);
            apiClient.clearCloudToken();
            localStorage.removeItem('cloud_token');

            set({
              isAuthenticated: true,
              sessionVerified: true,
              user: localData.user || null,
              token: localToken,
              refreshTokenValue: null,
              cloudToken: null,
              cloudVerificationPending: false,
              isLoading: false,
            });

            startTokenRefresh();
            return;
          }

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

          TokenManager.setToken(accessToken);
          localStorage.setItem('cloud_token', accessToken);
          apiClient.setCloudToken(accessToken);
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
            refreshTokenValue: null,
            cloudToken: accessToken,
            cloudVerificationPending: false,
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
        cancelCloudValidationRetry();
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
          cloudVerificationPending: false,
        });
        clearPersistedAuthStorage();
      },

      checkAuth: async () => {
        set({ isLoading: true, sessionVerified: false });

        // Migrate credentials created by older builds. Refresh credentials are
        // now HttpOnly cookies and must never remain readable by JavaScript.
        localStorage.removeItem('refresh_token');
        localStorage.removeItem('cloud_refresh_token');

        const persistedState = getPersistedAuthState();
        const currentState = useAuthStore.getState();
        const token = currentState.token || persistedState?.token || TokenManager.getToken();
        const cloudToken = currentState.cloudToken || persistedState?.cloudToken || (typeof window !== 'undefined' ? localStorage.getItem('cloud_token') : null);

        if (token && !TokenManager.getToken()) {
          TokenManager.setToken(token);
        }
        if (cloudToken && !localStorage.getItem('cloud_token')) {
          localStorage.setItem('cloud_token', cloudToken);
        }
        if (cloudToken) {
          apiClient.setCloudToken(cloudToken);
        }

        if (!token) {
          set({ isAuthenticated: false, sessionVerified: false, cloudVerificationPending: false, user: null, token: null, refreshTokenValue: null, isLoading: false });
          return;
        }

        if (!cloudToken) {
          if (getConnectionMode() === 'local') {
            set({ isAuthenticated: true, sessionVerified: true, cloudVerificationPending: false, isLoading: false });
            return;
          }
          markCloudVerificationPending();
          set({ isAuthenticated: true, sessionVerified: true, isLoading: false });
          return;
        }

        apiClient.setToken(token);

        if (!navigator.onLine) {
          markCloudVerificationPending();
          set({ isAuthenticated: true, sessionVerified: true, isLoading: false });
          return;
        }

        const valid = await validateSubscriptionWithCloud();
        if (!valid) {
          const pending = useAuthStore.getState().cloudVerificationPending;
          if (pending) {
            set({ isAuthenticated: true, sessionVerified: true, isLoading: false });
            return;
          }
          set({ isAuthenticated: false, sessionVerified: false, isLoading: false });
          return;
        }

        set({ isAuthenticated: true, sessionVerified: true, cloudVerificationPending: false, isLoading: false });
        startTokenRefresh();
      },

      refreshToken: async () => {
        try {
          const response = await authApi.refreshToken();
          const payload = (response && typeof response === 'object' && 'data' in response && response.data)
            ? response.data
            : response;
          const data = payload as { user?: User; token?: string; access_token?: string };
          const user = data.user;
          const token = data.token || data.access_token;
          if (!user || !token) {
            throw new Error('Refresh response did not include user and token');
          }
          TokenManager.setToken(token);
          apiClient.setToken(token);
          const cloudValid = await validateSubscriptionWithCloud();
          if (!cloudValid) {
            if (useAuthStore.getState().cloudVerificationPending) {
              set({ isAuthenticated: true, sessionVerified: true });
              return;
            }
            markCloudVerificationPending();
            set({ isAuthenticated: true, sessionVerified: true });
            return;
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
            refreshTokenValue: null,
          });
        } catch (error) {
          console.error('Failed to refresh token:', error);

          const status = Number((error as { status?: number })?.status);
          if (status === 401) {
            // A rejected refresh credential is not subscription expiry and is
            // not enough evidence of a network loss. Keep the session visible,
            // block sensitive work, and retry after the next online check.
            markCloudVerificationPending();
            set({ isAuthenticated: true, sessionVerified: true, isLoading: false });
            return;
          }

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
