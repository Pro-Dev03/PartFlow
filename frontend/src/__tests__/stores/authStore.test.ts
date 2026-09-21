import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { useAuthStore, validateSubscriptionWithCloud } from '../../stores/authStore';

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
});
