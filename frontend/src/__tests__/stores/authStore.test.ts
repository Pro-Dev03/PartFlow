import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { forceLogoutToLogin, markCloudVerificationPending, retrySubscriptionVerification, useAuthStore, validateSubscriptionWithCloud } from '../../stores/authStore';
import { authApi } from '../../services/api/endpoints';
import { CONNECTION_MODE_KEY } from '../../lib/config/app';
import { TokenManager } from '../../lib/token-manager';
import { readAutoLogoutReason } from '../../features/auth/sessionReason';

describe('cloud subscription validation', () => {
  beforeEach(() => {
    vi.useFakeTimers();
    localStorage.clear();
    sessionStorage.clear();
    TokenManager.clearToken();
    TokenManager.clearRefreshToken();
    TokenManager.clearCloudToken();
    useAuthStore.setState({
      isAuthenticated: false,
      sessionVerified: false,
      sessionRestorePending: false,
      cloudVerificationPending: false,
      user: null,
      token: null,
      refreshTokenValue: null,
      cloudToken: null,
      loginError: null,
      isLoading: false,
      isPostLoginVerifying: false,
    });
    localStorage.setItem(CONNECTION_MODE_KEY, 'cloud');
    vi.restoreAllMocks();
    Object.defineProperty(window.navigator, 'onLine', {
      configurable: true,
      value: true,
    });
    window.history.replaceState({}, '', '/');
    window.location.hash = '#/app/dashboard';
    useAuthStore.setState({
      isAuthenticated: true,
      sessionVerified: true,
      sessionRestorePending: false,
      cloudVerificationPending: false,
      user: null,
      token: 'local-token',
      refreshTokenValue: null,
      cloudToken: 'cloud-token',
      loginError: null,
      isLoading: false,
      isPostLoginVerifying: false,
    });
    TokenManager.setCloudToken('cloud-token');
  });

  afterEach(() => {
    vi.clearAllTimers();
    vi.useRealTimers();
  });

  it('never persists access or refresh credentials in browser storage', () => {
    useAuthStore.setState({
      isAuthenticated: true,
      token: 'access-token-secret',
      cloudToken: 'cloud-token-secret',
      refreshTokenValue: 'refresh-token-secret',
    });

    const persisted = JSON.parse(localStorage.getItem('auth-storage') || '{}').state;
    expect(persisted).not.toHaveProperty('token');
    expect(persisted).not.toHaveProperty('cloudToken');
    expect(persisted).not.toHaveProperty('refreshTokenValue');
    expect(JSON.stringify(persisted)).not.toContain('-token-secret');
  });

  it('uses the local backend for local-mode login', async () => {
    localStorage.setItem(CONNECTION_MODE_KEY, 'local');
    const cloudLoginSpy = vi.spyOn(authApi, 'loginWithCloud');
	cloudLoginSpy.mockResolvedValue({ token: 'cloud-access-token' } as any);
  vi.spyOn(authApi, 'createLocalSession').mockResolvedValue({
    user: { id: 'u-1', email: 'owner@partflow.com', first_name: 'Admin', last_name: 'Owner', phone: '+970599000000', is_active: true, role: 'owner' },
    token: 'local-access-token',
  } as any);
	vi.spyOn(globalThis, 'fetch').mockResolvedValue(
		new Response(JSON.stringify({ success: true, data: { valid: true, user: { id: 'u-1' } } }), { status: 200 })
	);

    await useAuthStore.getState().login('owner@partflow.com', 'TestOwnerPassword123!');

    expect(cloudLoginSpy).toHaveBeenCalledWith('owner@partflow.com', 'TestOwnerPassword123!');
    expect(authApi.createLocalSession).toHaveBeenCalledWith('cloud-access-token');
    expect(useAuthStore.getState().isAuthenticated).toBe(true);
    expect(useAuthStore.getState().token).toBe('local-access-token');
    expect(useAuthStore.getState().cloudToken).toBe('cloud-access-token');
  });

  it('restores the HttpOnly-cookie session before showing login on startup', async () => {
    const user = { id: 'u-restore', email: 'owner@example.test', first_name: 'Test', last_name: 'Owner' } as any;
    window.location.hash = '#/login';
    useAuthStore.setState({
      isAuthenticated: false,
      sessionVerified: false,
      sessionRestorePending: false,
      user: null,
      token: null,
      cloudToken: null,
      cloudVerificationPending: false,
      isLoading: false,
    });
    localStorage.setItem('auth-storage', JSON.stringify({ state: { isAuthenticated: true, user }, version: 0 }));
    vi.spyOn(authApi, 'refreshToken').mockResolvedValue({ token: 'restored-access-token', user } as any);
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response(JSON.stringify({ success: true, data: { valid: true, user } }), { status: 200 })
    );

    await useAuthStore.getState().checkAuth();

    expect(authApi.refreshToken).toHaveBeenCalledTimes(1);
    expect(useAuthStore.getState().isAuthenticated).toBe(true);
    expect(useAuthStore.getState().sessionVerified).toBe(true);
    expect(useAuthStore.getState().token).toBe('restored-access-token');
  });

  it('does not restore a cookie session after explicit logout, even after reconnecting', async () => {
    localStorage.setItem(CONNECTION_MODE_KEY, 'local');
    Object.defineProperty(window.navigator, 'onLine', { configurable: true, value: false });
    useAuthStore.setState({
      isAuthenticated: true,
      sessionVerified: true,
      token: 'old-local-token',
      cloudToken: 'old-cloud-token',
      user: { id: 'old-user', email: 'old@example.test' } as any,
    });
    TokenManager.setToken('old-local-token');
    TokenManager.setCloudToken('old-cloud-token');
    vi.spyOn(authApi, 'logout').mockResolvedValue(undefined as any);
    const refreshSpy = vi.spyOn(authApi, 'refreshToken').mockResolvedValue({
      token: 'cookie-restored-token',
    } as any);
    const fetchSpy = vi.spyOn(globalThis, 'fetch');

    useAuthStore.getState().logout();
    Object.defineProperty(window.navigator, 'onLine', { configurable: true, value: true });
    await useAuthStore.getState().checkAuth();

    expect(refreshSpy).not.toHaveBeenCalled();
    expect(fetchSpy).not.toHaveBeenCalled();
    expect(useAuthStore.getState().isAuthenticated).toBe(false);
    expect(useAuthStore.getState().token).toBeNull();
    expect(localStorage.getItem('partflow-manual-logout')).toBe('true');
  });

  it('shares one startup refresh when React invokes auth restoration more than once', async () => {
    const user = { id: 'u-restore', email: 'owner@example.test' } as any;
    TokenManager.clearToken();
    TokenManager.clearCloudToken();
    useAuthStore.setState({
      isAuthenticated: false,
      sessionVerified: false,
      sessionRestorePending: false,
      user: null,
      token: null,
      cloudToken: null,
    });
    localStorage.setItem('auth-storage', JSON.stringify({ state: { isAuthenticated: true, user }, version: 0 }));
    const refreshSpy = vi.spyOn(authApi, 'refreshToken').mockResolvedValue({ token: 'restored-access-token', user } as any);
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response(JSON.stringify({ success: true, data: { valid: true, user } }), { status: 200 })
    );

    await Promise.all([
      useAuthStore.getState().checkAuth(),
      useAuthStore.getState().checkAuth(),
    ]);

    expect(refreshSpy).toHaveBeenCalledTimes(1);
    expect(useAuthStore.getState().isAuthenticated).toBe(true);
  });

  it('validates the subscription in local mode', async () => {
    localStorage.setItem(CONNECTION_MODE_KEY, 'local');
    TokenManager.setCloudToken('cloud-token');
    useAuthStore.setState({ isAuthenticated: true, sessionVerified: true, token: 'local-token', cloudVerificationPending: false });
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response(JSON.stringify({ success: true, data: { valid: true } }), { status: 200 })
    );

    const valid = await validateSubscriptionWithCloud();

    expect(valid).toBe(true);
    expect(useAuthStore.getState().cloudVerificationPending).toBe(false);
    expect(useAuthStore.getState().isAuthenticated).toBe(true);
  });

  it('keeps the session when the cloud service is unavailable', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response(JSON.stringify({ error: 'temporary failure' }), { status: 503 })
    );

    const valid = await validateSubscriptionWithCloud();

    expect(valid).toBe(false);
    expect(useAuthStore.getState().isAuthenticated).toBe(true);
    expect(useAuthStore.getState().cloudVerificationPending).toBe(true);
    expect(TokenManager.getCloudToken()).toBe('cloud-token');
    expect(window.location.hash).toBe('#/app/dashboard');
  });

  it('keeps the session when the network request fails', async () => {
    localStorage.setItem('partflow-cloud-last-validated-at', String(Date.now()));
    vi.spyOn(globalThis, 'fetch').mockRejectedValue(new Error('cloud unavailable'));

    const valid = await validateSubscriptionWithCloud();

    expect(valid).toBe(false);
    expect(useAuthStore.getState().isAuthenticated).toBe(true);
    expect(useAuthStore.getState().cloudVerificationPending).toBe(true);
    expect(TokenManager.getCloudToken()).toBe('cloud-token');
    expect(localStorage.getItem('partflow-cloud-last-validated-at')).toBe(String(Date.now()));
  });

  it('does not turn an unclassified force-logout request into a logout', () => {
    const sessionCleared = vi.fn();
    window.addEventListener('partflow:session-cleared', sessionCleared);
    try {
      forceLogoutToLogin('Internet connection lost');

      expect(sessionCleared).not.toHaveBeenCalled();
      expect(useAuthStore.getState().isAuthenticated).toBe(true);
      expect(useAuthStore.getState().cloudVerificationPending).toBe(true);
      expect(window.location.hash).toBe('#/app/dashboard');
      expect(localStorage.getItem('partflow-reauth-required')).toBeNull();
    } finally {
      window.removeEventListener('partflow:session-cleared', sessionCleared);
    }
  });

  it('retains the active session when connectivity is lost', async () => {
    Object.defineProperty(window.navigator, 'onLine', { configurable: true, value: false });
    markCloudVerificationPending();
    expect(useAuthStore.getState().isAuthenticated).toBe(true);
    expect(useAuthStore.getState().cloudVerificationPending).toBe(true);
    expect(window.location.hash).toBe('#/app/dashboard');
  });

  it('keeps a local-mode session while cloud authorization is pending', () => {
    localStorage.setItem(CONNECTION_MODE_KEY, 'local');

    markCloudVerificationPending();

    expect(useAuthStore.getState().isAuthenticated).toBe(true);
    expect(useAuthStore.getState().cloudVerificationPending).toBe(true);
    expect(TokenManager.getCloudToken()).toBe('cloud-token');
    expect(window.location.hash).toBe('#/app/dashboard');
  });

  it('keeps a cloud-mode session while a live cloud decision is unavailable', () => {
    localStorage.setItem(CONNECTION_MODE_KEY, 'cloud');
    const sessionCleared = vi.fn();
    window.addEventListener('partflow:session-cleared', sessionCleared);
    try {
      markCloudVerificationPending();

      expect(useAuthStore.getState().isAuthenticated).toBe(true);
      expect(useAuthStore.getState().token).toBe('local-token');
      expect(TokenManager.getCloudToken()).toBe('cloud-token');
      expect(sessionCleared).not.toHaveBeenCalled();
      expect(window.location.hash).toBe('#/app/dashboard');
    } finally {
      window.removeEventListener('partflow:session-cleared', sessionCleared);
    }
  });

  it('does not restore access after an explicit subscription expiry response', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response(JSON.stringify({
        code: 'SUBSCRIPTION_EXPIRED',
        message: 'subscription expired',
      }), { status: 403 })
    );

    const valid = await validateSubscriptionWithCloud();

    expect(valid).toBe(false);
    expect(useAuthStore.getState().isAuthenticated).toBe(false);
    expect(window.location.hash).toBe('#/subscription-expired');
  });

  it('keeps the session after a generic cloud 403 rejection', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(new Response('{}', { status: 403 }));

    const valid = await validateSubscriptionWithCloud();

    expect(valid).toBe(false);
    expect(useAuthStore.getState().isAuthenticated).toBe(true);
    expect(useAuthStore.getState().cloudVerificationPending).toBe(true);
    expect(TokenManager.getCloudToken()).toBe('cloud-token');
  });

  it('keeps the session when refresh returns an unclassified 401', async () => {
    localStorage.setItem('auth_token', 'expired-access-token');
    localStorage.setItem('refresh_token', 'expired-refresh-token');
    TokenManager.setCloudToken('cloud-token');
    useAuthStore.setState({
      isAuthenticated: true,
      sessionVerified: true,
      token: 'expired-access-token',
      refreshTokenValue: 'expired-refresh-token',
      cloudToken: 'cloud-token',
    });
    vi.spyOn(authApi, 'refreshToken').mockRejectedValue(Object.assign(new Error('invalid token'), { status: 401 }));

    await useAuthStore.getState().refreshToken();

    expect(useAuthStore.getState().isAuthenticated).toBe(true);
    expect(useAuthStore.getState().sessionVerified).toBe(true);
    expect(useAuthStore.getState().cloudVerificationPending).toBe(true);
    expect(TokenManager.getToken()).toBe('expired-access-token');
  });

  it('routes confirmed account deletion to login with a clear reason', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(new Response(JSON.stringify({
      success: false,
      data: { error: { code: 'ACCOUNT_DELETED', message: 'User was deleted' } },
    }), { status: 401 }));

    await validateSubscriptionWithCloud();

    expect(useAuthStore.getState().isAuthenticated).toBe(false);
    expect(window.location.hash).toBe('#/login');
    expect(readAutoLogoutReason()).toBe('account-deleted');
  });

  it('routes confirmed account suspension to login with a clear reason', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(new Response(JSON.stringify({
      error: { data: { code: 'ACCOUNT_SUSPENDED', message: 'User disabled' } },
    }), { status: 403 }));

    await validateSubscriptionWithCloud();

    expect(useAuthStore.getState().isAuthenticated).toBe(false);
    expect(window.location.hash).toBe('#/login');
    expect(readAutoLogoutReason()).toBe('account-suspended');
  });

  it('keeps the session pending for an unapproved subscription suspension code', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(new Response(JSON.stringify({ code: 'SUBSCRIPTION_SUSPENDED' }), { status: 403 }));

    await validateSubscriptionWithCloud();

    expect(useAuthStore.getState().isAuthenticated).toBe(true);
    expect(useAuthStore.getState().cloudVerificationPending).toBe(true);
    expect(window.location.hash).toBe('#/app/dashboard');
    expect(readAutoLogoutReason()).toBeNull();
  });

  it('does not process a refresh rejection that arrives after manual logout', async () => {
    localStorage.setItem(CONNECTION_MODE_KEY, 'local');
    TokenManager.setToken('access-token');
    TokenManager.setCloudToken('cloud-token');
    useAuthStore.setState({
      isAuthenticated: true,
      sessionVerified: true,
      token: 'access-token',
      cloudToken: 'cloud-token',
    });
    let rejectRefresh!: (reason: unknown) => void;
    const refreshSpy = vi.spyOn(authApi, 'refreshToken').mockReturnValue(new Promise((_, reject) => {
      rejectRefresh = reject;
    }) as any);
    const localLogout = vi.spyOn(authApi, 'logout').mockResolvedValue(undefined as any);
    vi.spyOn(authApi, 'logoutWithCloud').mockResolvedValue(undefined as any);
    const fetchSpy = vi.spyOn(globalThis, 'fetch');
    const warning = vi.spyOn(console, 'warn');

    const refreshPromise = useAuthStore.getState().refreshToken();
    useAuthStore.getState().logout();
    rejectRefresh(Object.assign(new Error('refresh token rejected'), { status: 401 }));
    await refreshPromise;

    expect(refreshSpy.mock.calls[0]?.[0]?.aborted).toBe(true);
    expect(localLogout).toHaveBeenCalledTimes(1);
    expect(useAuthStore.getState().isAuthenticated).toBe(false);
    expect(TokenManager.getToken()).toBeNull();
    expect(warning).not.toHaveBeenCalled();
    expect(fetchSpy).not.toHaveBeenCalled();
  });

  it('retries cloud token refresh when the user presses retry', async () => {
    TokenManager.setCloudToken('cloud-token');
    localStorage.setItem('cloud_refresh_token', 'valid-refresh-token');
    localStorage.setItem('partflow-cloud-refresh-failed', 'true');
    const fetchSpy = vi.spyOn(globalThis, 'fetch')
      .mockResolvedValueOnce(new Response(JSON.stringify({ error: 'expired token' }), { status: 401 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ data: { access_token: 'fresh-cloud-token', refresh_token: 'next-refresh-token' } }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ success: true, data: { token: 'fresh-cloud-token' } }), { status: 200 }));

    const valid = await retrySubscriptionVerification();

    expect(valid).toBe(true);
    expect(fetchSpy).toHaveBeenCalledTimes(3);
    expect(TokenManager.getCloudToken()).toBe('fresh-cloud-token');
    expect(localStorage.getItem('partflow-cloud-refresh-failed')).toBeNull();
    expect(useAuthStore.getState().cloudVerificationPending).toBe(false);
  });
});
