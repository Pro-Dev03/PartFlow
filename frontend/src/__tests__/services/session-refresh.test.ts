import { afterEach, describe, expect, it, vi } from 'vitest';
import { authApi } from '../../services/api/endpoints';
import { classifyAuthFailure, extractAuthErrorCode, isDefinitiveRefreshRejection, refreshSessionCookie } from '../../services/api/session-refresh';
import { getLocalApiUrl } from '../../lib/config/app';

describe('shared session refresh', () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('shares one rotating-cookie refresh across callers in the same tab', async () => {
    let resolveResponse!: (response: Response) => void;
    const fetchSpy = vi.spyOn(globalThis, 'fetch').mockImplementation(() => new Promise((resolve) => {
      resolveResponse = resolve;
    }));

    const fromStore = refreshSessionCookie(getLocalApiUrl());
    const fromApi = authApi.refreshToken();
    await Promise.resolve();

    expect(fetchSpy).toHaveBeenCalledTimes(1);
    expect(fetchSpy).toHaveBeenCalledWith(
      expect.stringContaining('/auth/refresh'),
      expect.objectContaining({ method: 'POST', credentials: 'include' }),
    );

    resolveResponse(new Response(JSON.stringify({ data: { access_token: 'rotated-access-token' } }), { status: 200 }));
    const [storePayload, apiPayload] = await Promise.all([fromStore, fromApi]);
    expect(storePayload.access_token).toBe('rotated-access-token');
    expect(apiPayload.access_token).toBe('rotated-access-token');
    expect(fetchSpy).toHaveBeenCalledTimes(1);
  });

  it.each([
    [{ response: { data: { error: { code: 'SUBSCRIPTION_EXPIRED' } } } }, 'subscription_expired'],
    [{ error: { data: { code: 'ACCOUNT_DELETED' } } }, 'user_deleted'],
    [{ data: { error: { code: 'ACCOUNT_SUSPENDED' } } }, 'user_disabled'],
    [{ status: 403, code: 'SUBSCRIPTION_SUSPENDED' }, 'unclassified'],
    [{ status: 401, response: { data: { error: { message: 'invalid token' } } } }, 'session_invalid'],
    [{ status: 503, response: { data: { error: 'Render unavailable' } } }, 'temporary'],
  ] as const)('classifies nested API errors %#', (value, expected) => {
    expect(classifyAuthFailure(value)).toBe(expected);
  });

  it('prefers a confirmed block code inside a generic API wrapper', () => {
    expect(extractAuthErrorCode({
      code: 'AUTH_REFRESH_PENDING',
      response: { data: { error: { code: 'ACCOUNT_DELETED' } } },
    })).toBe('ACCOUNT_DELETED');
  });

  it('does not treat an unclassified or temporary 401/403/503 as a forced logout', () => {
    expect(isDefinitiveRefreshRejection({ status: 401, code: 'INVALID_TOKEN' })).toBe(false);
    expect(isDefinitiveRefreshRejection({ status: 403, code: 'PERMISSION_DENIED' })).toBe(false);
    expect(isDefinitiveRefreshRejection({ status: 503, code: 'SUBSCRIPTION_EXPIRED' })).toBe(false);
    expect(isDefinitiveRefreshRejection({ status: 403, code: 'SUBSCRIPTION_EXPIRED' })).toBe(true);
  });
});
