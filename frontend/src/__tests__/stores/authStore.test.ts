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
    TokenManager.setToken('cloud-access-token');
    TokenManager.setRefreshToken('cloud-refresh-token');
    useAuthStore.setState({
      isAuthenticated: true,
      sessionVerified: true,
      token: 'cloud-access-token',
      refreshTokenValue: 'cloud-refresh-token',
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
      expect.stringContaining('partflow-api.onrender.com/api/v1/auth/refresh'),
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
      expect.stringContaining('partflow-api.onrender.com/api/v1/auth/refresh'),
      expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({ refresh_token: 'old-refresh-token' }),
      }),
    );
  });

  it('redirects only for real subscription expiry and not for generic refresh failures', () => {
    expect(shouldRedirectToSubscriptionExpired(new Error('No refresh token available'), '/app')).toBe(false);
    expect(shouldRedirectToSubscriptionExpired(new Error('Session expired. Please login again.'), '/app')).toBe(false);

    const expiredError: any = new Error('انتهت مدة اشتراكك');
    expiredError.status = 403;
    expiredError.code = 'SUBSCRIPTION_EXPIRED';
    expect(shouldRedirectToSubscriptionExpired(expiredError, '/app')).toBe(true);
  });

  it('requires internet verification before restoring a session', async () => {
    window.history.pushState({}, '', '/app');
    vi.stubGlobal('navigator', { ...navigator, onLine: false });
    TokenManager.setToken('abc.def.ghi');

    await useAuthStore.getState().checkAuth();

    expect(useAuthStore.getState().isAuthenticated).toBe(false);
  });
});
