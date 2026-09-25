import { beforeEach, describe, expect, it, vi } from 'vitest';
import { apiClient } from '../../services/api/client';
import { authApi } from '../../services/api/endpoints';
import { TokenManager } from '../../lib/token-manager';

describe('apiClient auth propagation', () => {
  beforeEach(() => {
    localStorage.clear();
    TokenManager.clearToken();
    TokenManager.clearRefreshToken();
    TokenManager.clearCloudToken();
    apiClient.logout();
    vi.restoreAllMocks();
  });

  it('keeps auth tokens in memory instead of browser storage', () => {
    TokenManager.setToken('memory-token');
    TokenManager.setRefreshToken('memory-refresh-token');

    expect(TokenManager.getToken()).toBe('memory-token');
    expect(TokenManager.getRefreshToken()).toBe('memory-refresh-token');
    expect(localStorage.getItem('auth_token')).toBeNull();
    expect(localStorage.getItem('token')).toBeNull();
    expect(localStorage.getItem('refresh_token')).toBeNull();

    TokenManager.clearToken();
    TokenManager.clearRefreshToken();

    expect(TokenManager.getToken()).toBeNull();
    expect(TokenManager.getRefreshToken()).toBeNull();
  });

  it('sends business requests to the cloud using the cloud bearer token', async () => {
    localStorage.setItem('partflow-connection-mode', 'local');
    TokenManager.setToken('local-token');
    TokenManager.setCloudToken('cloud-token');
    apiClient.setToken('local-token');
    apiClient.setCloudToken('cloud-token');

    const fetchSpy = vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response(JSON.stringify({ success: true, data: { ok: true } }), {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      })
    );

    await apiClient.get('/products', undefined, false);

    expect(fetchSpy).toHaveBeenCalledWith(
      expect.stringContaining('/products'),
      expect.objectContaining({
        method: 'GET',
        credentials: 'include',
        headers: expect.objectContaining({
          Authorization: 'Bearer cloud-token',
          'X-PartFlow-Cloud-Token': 'cloud-token',
        }),
      })
    );
    expect(fetchSpy.mock.calls[0][0]).toContain('partflow-api.onrender.com');
  });

  it('keeps device database maintenance on the local API in local mode', async () => {
    localStorage.setItem('partflow-connection-mode', 'local');
    TokenManager.setToken('local-token');
    apiClient.setToken('local-token');
    const fetchSpy = vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response(JSON.stringify({ success: true, data: { deleted: true } }), {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      })
    );

    await apiClient.delete('/settings/database', { target: 'offline' });

    expect(fetchSpy.mock.calls[0][0]).toBe('http://localhost:8080/api/v1/settings/database?target=offline');
  });

  it('never routes local SQLite maintenance to Render in cloud mode', async () => {
    localStorage.setItem('partflow-connection-mode', 'cloud');
    TokenManager.setToken('local-token');
    TokenManager.setCloudToken('cloud-token');
    apiClient.setToken('local-token');
    apiClient.setCloudToken('cloud-token');
    const fetchSpy = vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response(JSON.stringify({ success: true, data: { deleted: true } }), {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      })
    );

    await apiClient.delete('/settings/database', { target: 'offline' });

    expect(fetchSpy.mock.calls[0][0]).toContain('http://localhost:8080/api/v1/settings/database');
    expect(fetchSpy.mock.calls[0][0]).not.toContain('partflow-api.onrender.com');
  });

  it('downloads CSV through the shared authenticated client', async () => {
    TokenManager.setToken('local-token');
    TokenManager.setCloudToken('cloud-token');
    apiClient.setToken('local-token');
    apiClient.setCloudToken('cloud-token');
    const fetchSpy = vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response('id,action\n1,create', { status: 200, headers: { 'Content-Type': 'text/csv' } })
    );

    const exported = await apiClient.getBlob('/audit/export');

    expect(await exported.text()).toBe('id,action\n1,create');
    expect(fetchSpy).toHaveBeenCalledWith(
      expect.stringContaining('/audit/export'),
      expect.objectContaining({
        method: 'GET',
        credentials: 'include',
        headers: expect.objectContaining({
          Authorization: 'Bearer cloud-token',
          'X-PartFlow-Cloud-Token': 'cloud-token',
          accept: 'text/csv',
        }),
      })
    );
  });

  it('does not automatically retry a transaction create after a server failure', async () => {
    vi.spyOn(console, 'error').mockImplementation(() => undefined);
    const fetchSpy = vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response(JSON.stringify({ error: { code: 'INTERNAL_ERROR', message: 'temporary failure' } }), {
        status: 503,
        headers: { 'Content-Type': 'application/json' },
      })
    );

    await expect(apiClient.post('/payments', { amount: 125 })).rejects.toMatchObject({ status: 503 });

    expect(fetchSpy).toHaveBeenCalledTimes(1);
  });

  it('invalidates the local desktop session when the cloud business API cannot be reached', async () => {
    localStorage.setItem('partflow-connection-mode', 'local');
    TokenManager.setToken('local-token');
    TokenManager.setCloudToken('cloud-token');
    apiClient.setToken('local-token');
    apiClient.setCloudToken('cloud-token');
    vi.spyOn(globalThis, 'fetch').mockRejectedValue(new TypeError('Failed to fetch'));

    const pending = vi.fn();
    window.addEventListener('partflow:cloud-verification-pending', pending);
    await expect(apiClient.post('/sales', { items: [] })).rejects.toThrow();
    expect(pending).toHaveBeenCalledTimes(1);
    window.removeEventListener('partflow:cloud-verification-pending', pending);
  });

  it('invalidates the cloud session when the cloud business API cannot be reached', async () => {
    localStorage.setItem('partflow-connection-mode', 'cloud');
    TokenManager.setToken('cloud-token');
    TokenManager.setCloudToken('cloud-token');
    apiClient.setToken('cloud-token');
    apiClient.setCloudToken('cloud-token');
    vi.spyOn(globalThis, 'fetch').mockRejectedValue(new TypeError('Failed to fetch'));

    const pending = vi.fn();
    window.addEventListener('partflow:cloud-verification-pending', pending);
    await expect(apiClient.post('/sales', { items: [] })).rejects.toThrow();
    expect(pending).toHaveBeenCalledTimes(1);
    window.removeEventListener('partflow:cloud-verification-pending', pending);
  });

  it.each(['cloud', 'local'])('keeps the %s session when a business request returns 500', async (mode) => {
    vi.spyOn(console, 'error').mockImplementation(() => undefined);
    vi.spyOn(apiClient as any, 'sleep').mockResolvedValue(undefined);
    localStorage.setItem('partflow-connection-mode', mode);
    TokenManager.setToken('access-token');
    TokenManager.setCloudToken('cloud-token');
    apiClient.setToken('access-token');
    apiClient.setCloudToken('cloud-token');
    vi.spyOn(globalThis, 'fetch').mockImplementation(async () =>
      new Response(JSON.stringify({ success: false, error: { code: 'INTERNAL_ERROR', message: 'temporary failure' } }), {
        status: 500,
        headers: { 'Content-Type': 'application/json' },
      })
    );

    const pending = vi.fn();
    const invalidated = vi.fn();
    window.addEventListener('partflow:cloud-verification-pending', pending);
    window.addEventListener('partflow:auth-invalidated', invalidated);
    await expect(apiClient.get('/sales/shifts/current')).rejects.toMatchObject({ status: 500 });

    expect(pending).not.toHaveBeenCalled();
    expect(invalidated).not.toHaveBeenCalled();
    expect(TokenManager.getToken()).toBe('access-token');
    expect(TokenManager.getCloudToken()).toBe('cloud-token');
    window.removeEventListener('partflow:cloud-verification-pending', pending);
    window.removeEventListener('partflow:auth-invalidated', invalidated);
  });

  it('keeps the session credentials when a 401 cannot be refreshed', async () => {
    TokenManager.setToken('local-token');
    TokenManager.setCloudToken('cloud-token');
    apiClient.setToken('local-token');
    apiClient.setCloudToken('cloud-token');

    const invalidated = vi.fn();
    const pending = vi.fn();
    window.addEventListener('partflow:auth-invalidated', invalidated);
    window.addEventListener('partflow:cloud-verification-pending', pending);
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response(JSON.stringify({
        success: false,
        error: { code: 'INVALID_TOKEN', message: 'invalid token' },
      }), {
        status: 401,
        headers: { 'Content-Type': 'application/json' },
      })
    );

    await expect(apiClient.get('/products', undefined, false)).rejects.toMatchObject({
      status: 401,
      code: 'AUTH_REFRESH_PENDING',
    });

    expect(invalidated).not.toHaveBeenCalled();
    expect(pending).toHaveBeenCalled();
    expect(TokenManager.getToken()).toBe('local-token');
    window.removeEventListener('partflow:auth-invalidated', invalidated);
    window.removeEventListener('partflow:cloud-verification-pending', pending);
  });

  it('retries refresh after a temporary refresh endpoint failure', async () => {
    TokenManager.setToken('expired-local-token');
    TokenManager.setCloudToken('expired-cloud-token');
    apiClient.setToken('expired-local-token');
    apiClient.setCloudToken('expired-cloud-token');
    const unauthorized = () => new Response(JSON.stringify({ code: 'INVALID_TOKEN', error: 'invalid token' }), {
      status: 401,
      headers: { 'Content-Type': 'application/json' },
    });
    const fetchSpy = vi.spyOn(globalThis, 'fetch')
      .mockResolvedValueOnce(unauthorized())
      .mockResolvedValueOnce(new Response(JSON.stringify({ error: 'temporary failure' }), { status: 503 }))
      .mockResolvedValueOnce(unauthorized())
      .mockResolvedValueOnce(new Response(JSON.stringify({ data: { access_token: 'fresh-cloud-token' } }), {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ success: true, data: { ok: true } }), {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      }));

    await expect(apiClient.get('/products', undefined, false)).rejects.toMatchObject({ code: 'AUTH_REFRESH_PENDING' });
    await expect(apiClient.get('/products', undefined, false)).resolves.toMatchObject({ data: { ok: true } });
    expect(fetchSpy).toHaveBeenCalledTimes(5);
    expect(TokenManager.getCloudToken()).toBe('fresh-cloud-token');
  });

  it('uses one refresh for concurrent 401 responses', async () => {
    TokenManager.setToken('expired-access-token');
    TokenManager.setCloudToken('expired-access-token');
    apiClient.setToken('expired-access-token');
    apiClient.setCloudToken('expired-access-token');
    let refreshCount = 0;
    vi.spyOn(globalThis, 'fetch').mockImplementation(async (input, init) => {
      const url = String(input);
      if (url.endsWith('/auth/refresh')) {
        refreshCount += 1;
        await new Promise((resolve) => setTimeout(resolve, 10));
        return new Response(JSON.stringify({ data: { access_token: 'fresh-access-token' } }), { status: 200 });
      }
      if (url.endsWith('/products') || url.endsWith('/customers')) {
        const authorization = init?.headers;
        const bearer = authorization instanceof Headers
          ? authorization.get('Authorization')
          : (authorization as Record<string, string> | undefined)?.Authorization;
        if (bearer !== 'Bearer fresh-access-token') {
          return new Response(JSON.stringify({ code: 'INVALID_TOKEN', error: 'expired' }), { status: 401 });
        }
        return new Response(JSON.stringify({ success: true, data: { ok: true } }), { status: 200 });
      }
      return new Response('{}', { status: 404 });
    });

    const [products, customers] = await Promise.all([
      apiClient.get('/products', undefined, false),
      apiClient.get('/customers', undefined, false),
    ]);

    expect(products.data).toEqual({ ok: true });
    expect(customers.data).toEqual({ ok: true });
    expect(refreshCount).toBe(1);
    expect(TokenManager.getCloudToken()).toBe('fresh-access-token');
  });

  it.each([
    { code: 'SUBSCRIPTION_EXPIRED', status: 403 },
    { code: 'ACCOUNT_DELETED', status: 401 },
    { code: 'ACCOUNT_SUSPENDED', status: 403 },
  ])('forces logout for the classified retry response $code', async ({ code, status }) => {
    TokenManager.setToken('expired-access-token');
    TokenManager.setCloudToken('expired-access-token');
    apiClient.setToken('expired-access-token');
    apiClient.setCloudToken('expired-access-token');
    const invalidated = vi.fn();
    window.addEventListener('partflow:auth-invalidated', invalidated);
    let businessRequest = 0;
    vi.spyOn(globalThis, 'fetch').mockImplementation(async (input) => {
      const url = String(input);
      if (url.endsWith('/auth/refresh')) {
        return new Response(JSON.stringify({ data: { access_token: 'fresh-access-token' } }), { status: 200 });
      }
      businessRequest += 1;
      if (businessRequest === 1) return new Response(JSON.stringify({ code: 'INVALID_TOKEN' }), { status: 401 });
      return new Response(JSON.stringify({ data: { error: { code } } }), { status });
    });

    await expect(apiClient.get('/products', undefined, false)).rejects.toMatchObject({ status });
    expect(invalidated).toHaveBeenCalledWith(expect.objectContaining({
      detail: expect.objectContaining({ code, definitive: true }),
    }));
    window.removeEventListener('partflow:auth-invalidated', invalidated);
  });

  it('does not log an expected unknown-barcode lookup as a server error', async () => {
    const consoleError = vi.spyOn(console, 'error').mockImplementation(() => undefined);
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response(JSON.stringify({ error: 'barcode not found' }), {
        status: 404,
        headers: { 'Content-Type': 'application/json' },
      })
    );

    await expect(apiClient.get('/barcodes/resolve/8854419001507', undefined, false)).rejects.toMatchObject({
      status: 404,
    });

    expect(consoleError).not.toHaveBeenCalled();
  });

  it('routes admin checks through the shared authenticated client', async () => {
    localStorage.setItem('auth_token', 'local-token');
    localStorage.setItem('cloud_token', 'cloud-token');
    apiClient.setToken('local-token');
    apiClient.setCloudToken('cloud-token');

    const fetchSpy = vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response(JSON.stringify({ success: true, data: { is_admin: true } }), {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      })
    );

    const result = await authApi.checkAdminAccess();

    expect(result).toEqual({ is_admin: true });
    expect(fetchSpy).toHaveBeenCalledWith(
      expect.stringContaining('/auth/admin-check'),
      expect.objectContaining({
        headers: expect.objectContaining({
          Authorization: 'Bearer cloud-token',
          'X-PartFlow-Cloud-Token': 'cloud-token',
        }),
      })
    );
  });
});
