import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { retrySubscriptionVerification, useAuthStore, validateSubscriptionWithCloud } from '../../stores/authStore';
import { authApi } from '../../services/api/endpoints';

describe('cloud subscription validation', () => {
  beforeEach(() => {
    vi.useFakeTimers();
    localStorage.clear();
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
      cloudVerificationPending: false,
      user: null,
      token: 'local-token',
      refreshTokenValue: null,
      cloudToken: 'cloud-token',
      loginError: null,
      isLoading: false,
      isPostLoginVerifying: false,
    });
    localStorage.setItem('cloud_token', 'cloud-token');
  });

  afterEach(() => {
    vi.clearAllTimers();
    vi.useRealTimers();
  });

  it('keeps the owner in the app during a temporary cloud outage', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response(JSON.stringify({ error: 'temporary failure' }), { status: 503 })
    );

    const valid = await validateSubscriptionWithCloud();

    expect(valid).toBe(false);
    expect(useAuthStore.getState().isAuthenticated).toBe(true);
    expect(useAuthStore.getState().cloudVerificationPending).toBe(true);
    expect(localStorage.getItem('cloud_token')).toBe('cloud-token');
    expect(window.location.hash).toBe('#/app/dashboard');
  });

  it('keeps the owner in the app during a temporary network failure', async () => {
    localStorage.setItem('partflow-cloud-last-validated-at', String(Date.now()));
    vi.spyOn(globalThis, 'fetch').mockRejectedValue(new Error('cloud unavailable'));

    const valid = await validateSubscriptionWithCloud();

    expect(valid).toBe(false);
    expect(useAuthStore.getState().isAuthenticated).toBe(true);
    expect(useAuthStore.getState().cloudVerificationPending).toBe(true);
    expect(localStorage.getItem('partflow-cloud-last-validated-at')).toBe(String(Date.now()));
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

  it('logs out and redirects when the refresh token is rejected', async () => {
    localStorage.setItem('auth_token', 'expired-access-token');
    localStorage.setItem('refresh_token', 'expired-refresh-token');
    localStorage.setItem('cloud_token', 'cloud-token');
    useAuthStore.setState({
      isAuthenticated: true,
      sessionVerified: true,
      token: 'expired-access-token',
      refreshTokenValue: 'expired-refresh-token',
      cloudToken: 'cloud-token',
    });
    vi.spyOn(authApi, 'refreshToken').mockRejectedValue(Object.assign(new Error('invalid token'), { status: 401 }));

    await useAuthStore.getState().refreshToken();

    expect(useAuthStore.getState().isAuthenticated).toBe(false);
    expect(useAuthStore.getState().token).toBeNull();
    expect(localStorage.getItem('auth_token')).toBeNull();
    expect(localStorage.getItem('refresh_token')).toBeNull();
    expect(window.location.hash).toBe('#/login');
  });

  it('retries cloud token refresh when the user presses retry', async () => {
    localStorage.setItem('cloud_refresh_token', 'valid-refresh-token');
    localStorage.setItem('partflow-cloud-refresh-failed', 'true');
    const fetchSpy = vi.spyOn(globalThis, 'fetch')
      .mockResolvedValueOnce(new Response(JSON.stringify({ error: 'expired token' }), { status: 401 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ data: { access_token: 'fresh-cloud-token', refresh_token: 'next-refresh-token' } }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ success: true, data: { token: 'fresh-cloud-token' } }), { status: 200 }));

    const valid = await retrySubscriptionVerification();

    expect(valid).toBe(true);
    expect(fetchSpy).toHaveBeenCalledTimes(3);
    expect(localStorage.getItem('cloud_token')).toBe('fresh-cloud-token');
    expect(localStorage.getItem('partflow-cloud-refresh-failed')).toBeNull();
    expect(useAuthStore.getState().cloudVerificationPending).toBe(false);
  });
});
