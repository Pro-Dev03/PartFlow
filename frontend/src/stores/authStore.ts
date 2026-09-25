import { create } from 'zustand';
import { createJSONStorage, persist } from 'zustand/middleware';
import { authApi } from '../services/api/endpoints';
import { apiClient } from '../services/api/client';
import { TokenManager } from '../lib/token-manager';
import { User } from '../types/models';
import { getCloudApiUrl, getConnectionMode } from '../lib/config/app';
import { isNetworkError } from '../lib/error-messages';
import { saveAutoLogoutReason } from '../features/auth/sessionReason';
import { extractAuthErrorCode, refreshSessionCookie } from '../services/api/session-refresh';
import { getLogoutCaller, logSessionDiagnostic } from '../services/api/sessionDiagnostics';

interface AuthState {
  isAuthenticated: boolean;
  sessionVerified: boolean;
  sessionRestorePending: boolean;
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

function getAuthErrorCode(error: unknown): string {
  return extractAuthErrorCode(error) ?? '';
}

function isExplicitAccountBlockCode(code: unknown): boolean {
  return ['ACCOUNT_DELETED', 'ACCOUNT_SUSPENDED'].includes(String(code ?? ''));
}

function isTemporaryAuthFailure(error: unknown): boolean {
  const status = Number((error as { status?: number } | null)?.status);
  return isNetworkError(error)
    || status === 0
    || status === 408
    || status === 429
    || status >= 500;
}

let refreshInterval: ReturnType<typeof setTimeout> | null = null;
let checkAuthInFlight: Promise<void> | null = null;
let authSessionGeneration = 0;
const pendingAuthRefreshControllers = new Set<AbortController>();
let cloudValidationInFlight: Promise<boolean> | null = null;
// Keep this key only to remove the marker written by older builds. It is never
// used to authorize a session after a failed cloud validation.
const CLOUD_LAST_VALIDATED_AT_KEY = 'partflow-cloud-last-validated-at';
const CLOUD_REFRESH_FAILED_KEY = 'partflow-cloud-refresh-failed';
const MANUAL_LOGOUT_KEY = 'partflow-manual-logout';
const REAUTH_REQUIRED_KEY = 'partflow-reauth-required';

function createAuthRefreshController(): AbortController {
  const controller = new AbortController();
  pendingAuthRefreshControllers.add(controller);
  return controller;
}

function abortPendingAuthRefreshes(): void {
  pendingAuthRefreshControllers.forEach((controller) => controller.abort());
  pendingAuthRefreshControllers.clear();
}

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
  localStorage.removeItem(CLOUD_REFRESH_FAILED_KEY);
  localStorage.removeItem(CLOUD_LAST_VALIDATED_AT_KEY);
  localStorage.removeItem('partflow-user-phone');
  TokenManager.clearToken();
  TokenManager.clearRefreshToken();
  TokenManager.clearCloudToken();

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

export function markCloudVerificationPending(): void {
  const state = useAuthStore.getState();
  const hasActiveSession = Boolean(
    state.isAuthenticated
      || state.token
      || state.cloudToken
      || TokenManager.getToken()
      || TokenManager.getCloudToken(),
  );
  if (!hasActiveSession) return;

  logSessionDiagnostic('cloud-verification.pending', { source: 'auth-store' });
  useAuthStore.setState({ cloudVerificationPending: true });
}

function clearCloudVerificationPending(): void {
  useAuthStore.setState({ cloudVerificationPending: false });
}

function clearSessionForSubscriptionBlock(): void {
  logSessionDiagnostic('logout.confirmed', { code: 'SUBSCRIPTION_EXPIRED', caller: getLogoutCaller() });
  authSessionGeneration += 1;
  checkAuthInFlight = null;
  abortPendingAuthRefreshes();
  stopTokenRefresh();
  apiClient.logout();
  window.dispatchEvent(new Event('partflow:session-cleared'));
  TokenManager.clearToken();
  TokenManager.clearRefreshToken();
  TokenManager.clearCloudToken();
  useAuthStore.setState({
    isAuthenticated: false,
    sessionVerified: false,
    sessionRestorePending: false,
    user: null,
    token: null,
    refreshTokenValue: null,
    cloudToken: null,
    cloudVerificationPending: false,
    isLoading: false,
  });
  clearPersistedAuthStorage();
  goToSubscriptionExpiredPage();
}

function clearSessionForAccountBlock(code: string): void {
  logSessionDiagnostic('logout.confirmed', { code, caller: getLogoutCaller() });
  authSessionGeneration += 1;
  checkAuthInFlight = null;
  abortPendingAuthRefreshes();
  stopTokenRefresh();
  apiClient.logout();
  window.dispatchEvent(new Event('partflow:session-cleared'));
  TokenManager.clearToken();
  TokenManager.clearRefreshToken();
  TokenManager.clearCloudToken();
  useAuthStore.setState({
    isAuthenticated: false,
    sessionVerified: false,
    sessionRestorePending: false,
    user: null,
    token: null,
    refreshTokenValue: null,
    cloudToken: null,
    cloudVerificationPending: false,
    isLoading: false,
  });
  clearPersistedAuthStorage();
  saveAutoLogoutReason(code === 'ACCOUNT_DELETED' ? 'account-deleted' : 'account-suspended');
  const currentUrl = new URL(window.location.href);
  currentUrl.pathname = '/';
  currentUrl.hash = '#/login';
  window.history.replaceState({}, '', currentUrl.toString());
  window.dispatchEvent(new HashChangeEvent('hashchange'));
}

export function handleConfirmedAuthBlock(code: string): void {
  if (code === 'SUBSCRIPTION_EXPIRED') {
    clearSessionForSubscriptionBlock();
  } else if (isExplicitAccountBlockCode(code)) {
    clearSessionForAccountBlock(code);
  } else {
    logSessionDiagnostic('auth-block.ignored-unclassified', { code });
    markCloudVerificationPending();
  }
}

export function shouldRedirectToSubscriptionExpired(error: unknown, pathname = window.location.pathname): boolean {
  if (!error || typeof error !== 'object') return false;
  if (pathname.includes('/subscription-expired')) return false;

  return getAuthErrorCode(error) === 'SUBSCRIPTION_EXPIRED';
}

/**
 * Validate the cloud session before cloud business traffic is allowed. Device
 * SQLite is used only by the local synchronization/setup endpoints.
 */
async function refreshCloudAccessToken(): Promise<string | null> {
  if (typeof window === 'undefined') return null;
  const refreshGeneration = authSessionGeneration;
  if (localStorage.getItem(CLOUD_REFRESH_FAILED_KEY) === 'true') return null;
  try {
    const payload = await refreshSessionCookie(getCloudApiUrl());
    if (refreshGeneration !== authSessionGeneration) return null;
    const nextToken = payload?.access_token || payload?.token;
    if (!nextToken) {
      const error: any = new Error('Refresh response did not include an access token');
      error.status = 502;
      error.code = 'AUTH_REFRESH_RESPONSE_INVALID';
      throw error;
    }

    TokenManager.setCloudToken(nextToken);
    apiClient.setCloudToken(nextToken);
    localStorage.removeItem(CLOUD_REFRESH_FAILED_KEY);
    useAuthStore.setState({ cloudToken: nextToken });
    return nextToken as string;
  } catch (error) {
    if (refreshGeneration !== authSessionGeneration) return null;
    if (Number((error as { status?: number } | null)?.status) === 401) {
      localStorage.setItem(CLOUD_REFRESH_FAILED_KEY, 'true');
    }
    throw error;
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
  const validationGeneration = authSessionGeneration;
  const sessionToken = TokenManager.getCloudToken();
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
      if (validationGeneration !== authSessionGeneration) return false;
      if (response.status === 401) {
        const rejectedPayload = await response.clone().json().catch(() => ({}));
        const rejectionCode = extractAuthErrorCode(rejectedPayload);
        if (rejectionCode === 'SUBSCRIPTION_EXPIRED') {
          clearSessionForSubscriptionBlock();
          return false;
        }
        if (isExplicitAccountBlockCode(rejectionCode)) {
          clearSessionForAccountBlock(rejectionCode);
          return false;
        }
        const refreshed = await refreshCloudAccessToken();
        if (validationGeneration !== authSessionGeneration) return false;
        if (refreshed) {
          cloudToken = refreshed;
          response = await callValidate(cloudToken);
        }
      }

      if (!response.ok) {
        if (validationGeneration !== authSessionGeneration) return false;
        if (response.status === 403) {
          const rejectedPayload = await response.clone().json().catch(() => ({}));
          const rejectionCode = extractAuthErrorCode(rejectedPayload);
          if (rejectionCode === 'SUBSCRIPTION_EXPIRED') clearSessionForSubscriptionBlock();
          else if (isExplicitAccountBlockCode(rejectionCode)) clearSessionForAccountBlock(rejectionCode);
          else markCloudVerificationPending();
          return false;
        } else if (response.status === 401) {
          const rejectedPayload = await response.clone().json().catch(() => ({}));
          const rejectionCode = extractAuthErrorCode(rejectedPayload);
          if (rejectionCode === 'SUBSCRIPTION_EXPIRED') clearSessionForSubscriptionBlock();
          else if (isExplicitAccountBlockCode(rejectionCode)) clearSessionForAccountBlock(rejectionCode);
          else markCloudVerificationPending();
          return false;
        }
        markCloudVerificationPending();
        return false;
      }

      const payload = await response.json();
      if (validationGeneration !== authSessionGeneration) return false;
      const data = payload?.data ?? payload;
      const nextToken = data?.token || data?.access_token || cloudToken;
      if (nextToken) {
        TokenManager.setCloudToken(nextToken);
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
        sessionRestorePending: false,
        cloudVerificationPending: false,
      });
      clearCloudVerificationPending();
      localStorage.removeItem(CLOUD_LAST_VALIDATED_AT_KEY);
      goToAppDashboard();
      return true;
    } catch (error) {
      const code = getAuthErrorCode(error);
      if (code === 'SUBSCRIPTION_EXPIRED') {
        clearSessionForSubscriptionBlock();
      } else if (isExplicitAccountBlockCode(code)) clearSessionForAccountBlock(code);
      else markCloudVerificationPending();
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
  if (typeof window !== 'undefined' && !TokenManager.getCloudToken()) {
    try {
      const refreshedToken = await refreshCloudAccessToken();
      if (refreshedToken) {
        apiClient.setCloudToken(refreshedToken);
      } else if (localStorage.getItem(CLOUD_REFRESH_FAILED_KEY) === 'true') {
        markCloudVerificationPending();
        return false;
      }
    } catch (error) {
      const code = getAuthErrorCode(error);
      if (code === 'SUBSCRIPTION_EXPIRED') {
        clearSessionForSubscriptionBlock();
        return false;
      }
      if (isExplicitAccountBlockCode(code)) clearSessionForAccountBlock(code);
      else markCloudVerificationPending();
      return false;
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
  if (reason === 'SUBSCRIPTION_EXPIRED') {
    clearSessionForSubscriptionBlock();
    return;
  }
  if (isExplicitAccountBlockCode(reason)) {
    clearSessionForAccountBlock(reason);
    return;
  }
  logSessionDiagnostic('logout.denied-unconfirmed', { reason, caller: getLogoutCaller() });
  markCloudVerificationPending();
}
export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
    isAuthenticated: false,
    sessionVerified: false,
    sessionRestorePending: false,
      cloudVerificationPending: false,
      user: null,
      token: null,
      refreshTokenValue: null,
      cloudToken: null,
      loginError: null,
      isLoading: true,
        isPostLoginVerifying: false,
        setPostLoginVerifying: (value: boolean) => set({ isPostLoginVerifying: value }),

          login: async (email: string, password: string) => {
        authSessionGeneration += 1;
        checkAuthInFlight = null;
        abortPendingAuthRefreshes();
        clearPersistedAuthStorage();
        localStorage.removeItem(CLOUD_REFRESH_FAILED_KEY);
        apiClient.logout();
        window.dispatchEvent(new Event('partflow:session-cleared'));
        set({ isLoading: true, loginError: null, cloudVerificationPending: false });
        try {
          const connectionMode = getConnectionMode();

          if (connectionMode === 'local') {
            let cloudTokenForLocalSession: string | null = null;
            let localData: {
              user: User;
              token?: string;
              access_token?: string;
              refresh_token?: string;
            };

            try {
              const cloudData = await authApi.loginWithCloud(email, password) as {
                token?: string;
                access_token?: string;
              };
              const cloudToken = cloudData.token || cloudData.access_token;
              if (!cloudToken) {
                throw new Error('Cloud login response did not include an access token');
              }
              cloudTokenForLocalSession = cloudToken;
              localData = await authApi.createLocalSession(cloudToken) as typeof localData;
            } catch (cloudError) {
              const status = cloudError && typeof cloudError === 'object'
                ? (cloudError as { status?: number }).status
                : undefined;
              if (status !== 401 && status !== 404) {
                throw cloudError;
              }

              localData = await authApi.login(email, password) as typeof localData;
            }
            const localToken = localData.token || localData.access_token;
            if (!localToken) {
              throw new Error('Local login response did not include an access token');
            }

            TokenManager.setToken(localToken);
            apiClient.setToken(localToken);
            if (!cloudTokenForLocalSession) {
              const cloudData = await authApi.loginWithCloud(email, password) as {
                token?: string;
                access_token?: string;
              };
              cloudTokenForLocalSession = cloudData.token || cloudData.access_token || null;
            }
            if (!cloudTokenForLocalSession) {
              throw new Error('Cloud subscription verification failed');
            }
            apiClient.setCloudToken(cloudTokenForLocalSession);

          set({
              isAuthenticated: true,
              sessionVerified: true,
              sessionRestorePending: false,
              user: localData.user || null,
              token: localToken,
              refreshTokenValue: null,
              cloudToken: cloudTokenForLocalSession,
              cloudVerificationPending: false,
              isLoading: false,
            });

            if (!(await validateSubscriptionWithCloud())) {
              throw new Error('Cloud subscription verification failed');
            }
            localStorage.removeItem(MANUAL_LOGOUT_KEY);
            localStorage.removeItem(REAUTH_REQUIRED_KEY);
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
          TokenManager.setCloudToken(accessToken);
          apiClient.setCloudToken(accessToken);
          apiClient.setToken(accessToken);

          const valid = await validateSubscriptionWithCloud();
          if (!valid) {
            throw new Error('Cloud subscription verification failed');
          }

          localStorage.removeItem(MANUAL_LOGOUT_KEY);
          localStorage.removeItem(REAUTH_REQUIRED_KEY);

          set({
            isAuthenticated: true,
            sessionVerified: true,
            sessionRestorePending: false,
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
            TokenManager.clearCloudToken();
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
        logSessionDiagnostic('logout.manual', { caller: getLogoutCaller(), connectionMode: getConnectionMode() });
        const cloudToken = state.cloudToken || TokenManager.getCloudToken();
        const isLocalMode = getConnectionMode() === 'local';
        // A remote HttpOnly refresh cookie cannot be cleared while offline.
        // Remember the user's explicit logout locally so a later reconnect
        // cannot silently restore that cookie-backed session.
        localStorage.setItem(MANUAL_LOGOUT_KEY, 'true');
        authSessionGeneration += 1;
        checkAuthInFlight = null;
        abortPendingAuthRefreshes();
        void authApi.logout().catch(() => undefined);
        if (isLocalMode && cloudToken && (typeof navigator === 'undefined' || navigator.onLine)) {
          void authApi.logoutWithCloud(cloudToken).catch(() => undefined);
        }
        stopTokenRefresh();
        TokenManager.clearToken();
        TokenManager.clearRefreshToken();
        TokenManager.clearCloudToken();
        localStorage.removeItem('cloud_refresh_token');
        apiClient.logout();
        window.dispatchEvent(new Event('partflow:session-cleared'));
        apiClient.clearCloudToken();
        set({
          isAuthenticated: false,
          sessionVerified: false,
          sessionRestorePending: false,
          user: null,
          token: null,
          refreshTokenValue: null,
          cloudToken: null,
          cloudVerificationPending: false,
        });
        clearPersistedAuthStorage();
      },

      checkAuth: () => {
        if (checkAuthInFlight) return checkAuthInFlight;

        const checkGeneration = authSessionGeneration;
        const checkPromise = (async () => {
        set({ isLoading: true, sessionVerified: false, sessionRestorePending: false });

        // Older builds wrote REAUTH_REQUIRED_KEY after temporary outages.
        // Manual logout is the only persisted local marker that blocks restore.
        localStorage.removeItem(REAUTH_REQUIRED_KEY);
        if (localStorage.getItem(MANUAL_LOGOUT_KEY) === 'true') {
          apiClient.logout();
          window.dispatchEvent(new Event('partflow:session-cleared'));
          clearPersistedAuthStorage();
          set({
            isAuthenticated: false,
            sessionVerified: false,
            sessionRestorePending: false,
            cloudVerificationPending: false,
            user: null,
            token: null,
            refreshTokenValue: null,
            cloudToken: null,
            isLoading: false,
          });
          return;
        }

        // Migrate credentials created by older builds. Refresh credentials are
        // now HttpOnly cookies and must never remain readable by JavaScript.
        localStorage.removeItem('refresh_token');
        localStorage.removeItem('cloud_refresh_token');
        localStorage.removeItem('auth_token');
        localStorage.removeItem('token');
        TokenManager.clearCloudToken();

        const persistedState = getPersistedAuthState();
        const currentState = useAuthStore.getState();

        const token = currentState.token || persistedState?.token || TokenManager.getToken();
        let cloudToken = currentState.cloudToken || persistedState?.cloudToken || null;

        if (token && !TokenManager.getToken()) {
          TokenManager.setToken(token);
        }
        if (cloudToken) {
          apiClient.setCloudToken(cloudToken);
        }

        if (!token) {
          const hadSavedSession = Boolean(persistedState?.isAuthenticated || persistedState?.user);
          let restoreFailure: unknown = null;
          try {
            const controller = createAuthRefreshController();
            let refreshed;
            try {
              refreshed = await authApi.refreshToken(controller.signal);
            } finally {
              pendingAuthRefreshControllers.delete(controller);
            }
            if (checkGeneration !== authSessionGeneration) return;
            const payload = (refreshed && typeof refreshed === 'object' && 'data' in refreshed && refreshed.data)
              ? refreshed.data
              : refreshed;
            const refreshedToken = payload?.token || payload?.access_token;
            if (refreshedToken) {
              TokenManager.setToken(refreshedToken);
              apiClient.setToken(refreshedToken);
              let refreshedCloudToken: string | null = getConnectionMode() === 'local' ? null : refreshedToken;
              if (getConnectionMode() === 'local') {
                try {
                  refreshedCloudToken = await refreshCloudAccessToken();
                  if (checkGeneration !== authSessionGeneration) return;
                } catch (error) {
                  restoreFailure = error;
                  if (!isTemporaryAuthFailure(error)) throw error;
                }
                if (!refreshedCloudToken && localStorage.getItem(CLOUD_REFRESH_FAILED_KEY) === 'true') {
                  markCloudVerificationPending();
                }
              }
              if (refreshedCloudToken) {
                TokenManager.setCloudToken(refreshedCloudToken);
                apiClient.setCloudToken(refreshedCloudToken);
              }
              set({
                isAuthenticated: true,
                sessionVerified: true,
                sessionRestorePending: false,
                token: refreshedToken,
                cloudToken: refreshedCloudToken,
                user: payload?.user ?? null,
                isLoading: false,
              });
              const subscriptionValid = refreshedCloudToken
                ? await validateSubscriptionWithCloud()
                : false;
              if (checkGeneration !== authSessionGeneration) return;
              if (subscriptionValid) {
                startTokenRefresh();
                return;
              }
              if (useAuthStore.getState().cloudVerificationPending || !refreshedCloudToken) {
                markCloudVerificationPending();
                set({
                  isAuthenticated: true,
                  sessionVerified: true,
                  sessionRestorePending: false,
                  cloudVerificationPending: true,
                  isLoading: false,
                });
                startTokenRefresh();
                return;
              }
            }
          } catch (error) {
            if (checkGeneration !== authSessionGeneration) return;
            restoreFailure = error;
            // A local refresh cookie can become invalid after the local API is
            // rebuilt. Recreate the local session from the cloud cookie before
            // treating the browser session as signed out.
            if (getConnectionMode() === 'local' && navigator.onLine) {
              try {
                const refreshedCloudToken = await refreshCloudAccessToken();
                if (checkGeneration !== authSessionGeneration) return;
                if (refreshedCloudToken) {
                  const localSession = await authApi.createLocalSession(refreshedCloudToken) as {
                    token?: string;
                    access_token?: string;
                    user?: User;
                  };
                  if (checkGeneration !== authSessionGeneration) return;
                  const localToken = localSession.token || localSession.access_token;
                  if (localToken) {
                    TokenManager.setToken(localToken);
                    apiClient.setToken(localToken);
                    set({
                      isAuthenticated: true,
                      sessionVerified: true,
                      sessionRestorePending: false,
                      token: localToken,
                      cloudToken: refreshedCloudToken,
                      user: localSession.user ?? null,
                      isLoading: false,
                    });
                    startTokenRefresh();
                    return;
                  }
                }
              } catch (fallbackError) {
                if (checkGeneration !== authSessionGeneration) return;
                restoreFailure = fallbackError;
                // Fall through to the normal unauthenticated state.
              }
            }
          }

          if (checkGeneration !== authSessionGeneration) return;
          const restoreCode = getAuthErrorCode(restoreFailure);
          if (restoreCode === 'SUBSCRIPTION_EXPIRED') {
            clearSessionForSubscriptionBlock();
            return;
          }
          if (isExplicitAccountBlockCode(restoreCode)) {
            clearSessionForAccountBlock(restoreCode);
            return;
          }

          if (hadSavedSession) {
            set({
              isAuthenticated: false,
              sessionVerified: false,
              sessionRestorePending: true,
              cloudVerificationPending: true,
              user: persistedState?.user ?? currentState.user ?? null,
              token: null,
              refreshTokenValue: null,
              cloudToken: null,
              isLoading: false,
            });
            return;
          }

          TokenManager.clearToken();
          TokenManager.clearRefreshToken();
          TokenManager.clearCloudToken();
          apiClient.logout();
          window.dispatchEvent(new Event('partflow:session-cleared'));
          set({ isAuthenticated: false, sessionVerified: false, sessionRestorePending: false, cloudVerificationPending: false, user: null, token: null, refreshTokenValue: null, cloudToken: null, isLoading: false });
          return;
        }

        if (!cloudToken) {
          if (getConnectionMode() === 'local') {
            try {
              cloudToken = await refreshCloudAccessToken();
              if (checkGeneration !== authSessionGeneration) return;
            } catch (error) {
              if (checkGeneration !== authSessionGeneration) return;
              const code = getAuthErrorCode(error);
              if (code === 'SUBSCRIPTION_EXPIRED') {
                clearSessionForSubscriptionBlock();
                return;
              }
              if (isExplicitAccountBlockCode(code)) {
                clearSessionForAccountBlock(code);
                return;
              }
              markCloudVerificationPending();
            }
            if (cloudToken) {
              TokenManager.setCloudToken(cloudToken);
              apiClient.setCloudToken(cloudToken);
              set({ cloudToken });
            } else if (localStorage.getItem(CLOUD_REFRESH_FAILED_KEY) === 'true') {
              markCloudVerificationPending();
            } else {
              markCloudVerificationPending();
            }
          } else {
            cloudToken = token;
            TokenManager.setCloudToken(token);
            apiClient.setCloudToken(token);
          }
        }

        apiClient.setToken(token);

        if (!navigator.onLine) {
          markCloudVerificationPending();
          set({ isAuthenticated: true, sessionVerified: true, cloudVerificationPending: true, isLoading: false });
          startTokenRefresh();
          return;
        }

        const valid = await validateSubscriptionWithCloud();
        if (checkGeneration !== authSessionGeneration) return;
        if (!valid) {
          markCloudVerificationPending();
          set({ isAuthenticated: true, sessionVerified: true, cloudVerificationPending: true, isLoading: false });
          startTokenRefresh();
          return;
        }

        set({ isAuthenticated: true, sessionVerified: true, cloudVerificationPending: false, isLoading: false });
        startTokenRefresh();
        })();
        checkAuthInFlight = checkPromise;
        const clearInFlight = () => {
          if (checkAuthInFlight === checkPromise) checkAuthInFlight = null;
        };
        void checkPromise.then(clearInFlight, clearInFlight);
        return checkPromise;
      },

      refreshToken: async () => {
        const refreshGeneration = authSessionGeneration;
        try {
          const controller = createAuthRefreshController();
          let response;
          try {
            response = await authApi.refreshToken(controller.signal);
          } finally {
            pendingAuthRefreshControllers.delete(controller);
          }
          if (refreshGeneration !== authSessionGeneration) return;
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
          if (refreshGeneration !== authSessionGeneration) return;
          if (!cloudValid) {
            markCloudVerificationPending();
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
          if (refreshGeneration !== authSessionGeneration) return;
          console.error('Failed to refresh token:', error);

          const authCode = getAuthErrorCode(error);
          if (authCode === 'SUBSCRIPTION_EXPIRED') {
            clearSessionForSubscriptionBlock();
            return;
          }
          if (isExplicitAccountBlockCode(authCode)) {
            clearSessionForAccountBlock(authCode);
            return;
          }

          const currentToken = TokenManager.getToken() || useAuthStore.getState().token;
          if (currentToken) {
            TokenManager.setToken(currentToken);
            apiClient.setToken(currentToken);
          }

          const status = Number((error as { status?: number })?.status);
          if (status === 401) {
            if (getConnectionMode() === 'local' && navigator.onLine) {
              try {
                const refreshedCloudToken = await refreshCloudAccessToken();
                if (refreshGeneration !== authSessionGeneration) return;
                if (refreshedCloudToken) {
                  const localSession = await authApi.createLocalSession(refreshedCloudToken) as {
                    token?: string;
                    access_token?: string;
                    user?: User;
                  };
                  if (refreshGeneration !== authSessionGeneration) return;
                  const localToken = localSession.token || localSession.access_token;
                  if (localToken) {
                    TokenManager.setCloudToken(refreshedCloudToken);
                    TokenManager.setToken(localToken);
                    apiClient.setCloudToken(refreshedCloudToken);
                    apiClient.setToken(localToken);
                    set({
                      isAuthenticated: true,
                      sessionVerified: true,
                      sessionRestorePending: false,
                      cloudVerificationPending: false,
                      cloudToken: refreshedCloudToken,
                      token: localToken,
                      user: localSession.user ?? useAuthStore.getState().user,
                      isLoading: false,
                    });
                    const cloudValid = await validateSubscriptionWithCloud();
                    if (refreshGeneration !== authSessionGeneration) return;
                    if (cloudValid) return;
                    return;
                  }
                }
              } catch (refreshError) {
                if (refreshGeneration !== authSessionGeneration) return;
                const refreshCode = getAuthErrorCode(refreshError);
                if (refreshCode === 'SUBSCRIPTION_EXPIRED') clearSessionForSubscriptionBlock();
                else if (isExplicitAccountBlockCode(refreshCode)) clearSessionForAccountBlock(refreshCode);
                else markCloudVerificationPending();
                return;
              }

              markCloudVerificationPending();
              return;
            }

            markCloudVerificationPending();
            return;
          }

          markCloudVerificationPending();
          return;
        }
      },
    }),
    {
      name: 'auth-storage',
      storage: createJSONStorage(() => authStorage),
      // Keep only non-sensitive UI state in persistence. Access tokens and cloud
      // tokens are protected by HttpOnly refresh cookies and must never be kept
      // in browser storage where JavaScript can read them.
      partialize: (state) => ({
        isAuthenticated: state.isAuthenticated,
        sessionVerified: state.sessionVerified,
        user: state.user,
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
