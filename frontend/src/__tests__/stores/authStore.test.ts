import { beforeEach, describe, expect, it, vi } from 'vitest';
import { TokenManager } from '../../lib/token-manager';
import { authApi } from '../../services/api/endpoints';
import { shouldRedirectToSubscriptionExpired, useAuthStore } from '../../stores/authStore';

describe('auth store logout behavior', () => {
  beforeEach(() => {
    localStorage.clear();
    useAuthStore.setState({
      isAuthenticated: false,
      sessionVerified: false,
      user: null,
      token: null,
      refreshTokenValue: null,
      isLoading: false,
    });
    vi.restoreAllMocks();
  });

  it('clears persisted auth state when logging out', () => {
    useAuthStore.setState({
      isAuthenticated: true,
      user: { id: '1', name: 'Test User', email: 'test@example.com' } as any,
      token: 'abc.def.ghi',
      refreshTokenValue: 'refresh-xyz',
    });

    useAuthStore.getState().logout();

    expect(useAuthStore.getState().isAuthenticated).toBe(false);
    expect(useAuthStore.getState().user).toBeNull();
    expect(useAuthStore.getState().token).toBeNull();
    expect(localStorage.getItem('auth-storage')).toBeNull();
    expect(localStorage.getItem('auth_token')).toBeNull();
    expect(localStorage.getItem('token')).toBeNull();
    expect(localStorage.getItem('refresh_token')).toBeNull();
  });

  it('best-effort revokes the cloud session before local logout', () => {
    const fetchMock = vi.fn().mockResolvedValue({ ok: true, json: async () => ({}) });
    vi.stubGlobal('fetch', fetchMock);
    TokenManager.setToken('local-jwt-token');
    TokenManager.setRefreshToken('cloud-refresh-token');
    localStorage.setItem('cloud_token', 'cloud-access-token');
    useAuthStore.setState({
      isAuthenticated: true,
      sessionVerified: true,
      token: 'local-jwt-token',
      refreshTokenValue: 'local-refresh-token',
      cloudToken: 'cloud-access-token',
    });

    useAuthStore.getState().logout();

    expect(fetchMock).toHaveBeenCalledWith(
      expect.stringContaining('partflow-api.onrender.com/api/v1/auth/logout'),
      expect.objectContaining({
        method: 'POST',
        headers: expect.objectContaining({ Authorization: 'Bearer cloud-access-token' }),
      }),
    );
    expect(useAuthStore.getState().isAuthenticated).toBe(false);
  });

  it('sends the stored refresh token when refreshing the session', async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      text: async () => JSON.stringify({
        success: true,
        data: {
          token: 'new-access-token',
          refresh_token: 'new-refresh-token',
          user: { id: '1', email: 'test@example.com' },
        },
      }),
    });

    vi.stubGlobal('fetch', fetchMock);
    TokenManager.setToken('old-access-token');
    TokenManager.setRefreshToken('old-refresh-token');

    await authApi.refreshToken();

    expect(fetchMock).toHaveBeenCalledWith(
      expect.stringContaining('localhost:8080/api/v1/auth/refresh'),
      expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({ refresh_token: 'old-refresh-token' }),
      })
    );
  });

  it('updates the store when the cloud rotates the refresh token', async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({
        data: {
          token: 'rotated-access-token',
          refresh_token: 'rotated-refresh-token',
          user: { id: '1', email: 'test@example.com' },
        },
      }),
    });
    vi.stubGlobal('fetch', fetchMock);
    TokenManager.setToken('old-access-token');
    TokenManager.setRefreshToken('old-refresh-token');
    useAuthStore.setState({
      isAuthenticated: true,
      sessionVerified: true,
      token: 'old-access-token',
      refreshTokenValue: 'old-refresh-token',
      user: { id: '1', email: 'test@example.com' } as any,
    });

    await useAuthStore.getState().refreshToken();

    expect(TokenManager.getToken()).toBe('rotated-access-token');
    expect(TokenManager.getRefreshToken()).toBe('rotated-refresh-token');
    expect(useAuthStore.getState().refreshTokenValue).toBe('rotated-refresh-token');
    expect(fetchMock).toHaveBeenCalledWith(
      expect.stringContaining('localhost:8080/api/v1/auth/refresh'),
      expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({ refresh_token: 'old-refresh-token' }),
      }),
    );
  });

  it('logs out when the stored refresh token is rejected with 401', async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: false,
      status: 401,
      json: async () => ({
        error: {
          code: 'INVALID_REFRESH_TOKEN',
          message: 'Refresh token expired',
        },
      }),
    });
    vi.stubGlobal('fetch', fetchMock);
    TokenManager.setToken('expired-access-token');
    TokenManager.setRefreshToken('expired-refresh-token');
    useAuthStore.setState({
      isAuthenticated: true,
      sessionVerified: true,
      token: 'expired-access-token',
      refreshTokenValue: 'expired-refresh-token',
      user: { id: '1', email: 'test@example.com' } as any,
    });

    await useAuthStore.getState().refreshToken();

    expect(useAuthStore.getState().isAuthenticated).toBe(false);
    expect(useAuthStore.getState().token).toBeNull();
    expect(TokenManager.getToken()).toBeNull();
    expect(TokenManager.getRefreshToken()).toBeNull();
    expect(localStorage.getItem('auth-storage')).toBeNull();
  });

  it('redirects only for real subscription expiry and not for generic refresh failures', () => {
    expect(shouldRedirectToSubscriptionExpired(new Error('No refresh token available'), '/app')).toBe(false);
    expect(shouldRedirectToSubscriptionExpired(new Error('Session expired. Please login again.'), '/app')).toBe(false);

    const permissionError: any = new Error('administrator privileges required');
    permissionError.status = 403;
    permissionError.code = 'ADMIN_REQUIRED';
    expect(shouldRedirectToSubscriptionExpired(permissionError, '/app')).toBe(false);

    const expiredError: any = new Error('انتهت مدة اشتراكك');
    expiredError.status = 403;
    expiredError.code = 'SUBSCRIPTION_EXPIRED';
    expect(shouldRedirectToSubscriptionExpired(expiredError, '/app')).toBe(true);

    const cloudForbidden: any = new Error('الحساب غير نشط أو أن الاشتراك منتهٍ');
    cloudForbidden.status = 403;
    cloudForbidden.code = 'CLOUD_AUTH_REQUIRED';
    expect(shouldRedirectToSubscriptionExpired(cloudForbidden, '/app')).toBe(true);

    const cloudUnauthorized: any = new Error('جلسة الدخول غير صالحة');
    cloudUnauthorized.status = 401;
    cloudUnauthorized.code = 'CLOUD_AUTH_REQUIRED';
    expect(shouldRedirectToSubscriptionExpired(cloudUnauthorized, '/app')).toBe(false);
  });

  it('does not persist authenticated tokens or auth flags across reloads', async () => {
    useAuthStore.setState({
      isAuthenticated: true,
      sessionVerified: true,
      user: { id: '1', name: 'Test User', email: 'test@example.com' } as any,
      token: 'abc.def.ghi',
      refreshTokenValue: 'refresh-xyz',
    });

    await Promise.resolve();

    const persisted = JSON.parse(localStorage.getItem('auth-storage') || '{}');
    expect(persisted.state?.token).toBeUndefined();
    expect(persisted.state?.refreshTokenValue).toBeUndefined();
    expect(persisted.state?.isAuthenticated).toBeUndefined();
    expect(persisted.state?.sessionVerified).toBeUndefined();
  });

  it('requires internet verification before restoring a session', async () => {
    window.history.pushState({}, '', '/app');
    vi.stubGlobal('navigator', { ...navigator, onLine: false });
    TokenManager.setToken('abc.def.ghi');

    await useAuthStore.getState().checkAuth();

    expect(useAuthStore.getState().isAuthenticated).toBe(false);
  });
});
