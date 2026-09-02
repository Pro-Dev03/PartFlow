import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { apiClient } from '../../services/api/client';
import { authApi } from '../../services/api/endpoints';

describe('apiClient auth refresh', () => {
  beforeEach(() => {
    localStorage.clear();
    apiClient.setToken('old-access-token');
    localStorage.setItem('refresh_token', 'refresh-token-123');
    vi.stubGlobal('fetch', vi.fn());
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    localStorage.clear();
  });

  it('retries the original request after a successful refresh when backend returns top-level token fields', async () => {
    const fetchMock = vi.mocked(globalThis.fetch);

    fetchMock
      .mockResolvedValueOnce({
        ok: false,
        status: 401,
        text: async () => JSON.stringify({ error: { message: 'unauthorized' } }),
      } as Response)
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ({
          access_token: 'new-access-token',
          refresh_token: 'new-refresh-token',
          expires_in: 900,
          user: { id: 'u-1', email: 'owner@partflow.com' },
        }),
      } as Response)
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        text: async () => JSON.stringify({ success: true, data: { id: 'ok' } }),
      } as Response);

    const result = await apiClient.get('/customers');

    expect(result.success).toBe(true);
    expect(localStorage.getItem('auth_token')).toBe('new-access-token');
    expect(localStorage.getItem('refresh_token')).toBe('new-refresh-token');
    expect(fetchMock).toHaveBeenCalledTimes(3);
  });

  it('routes admin checks to the cloud auth service instead of the local SQLite API', async () => {
    const fetchMock = vi.mocked(globalThis.fetch);
    fetchMock.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: async () => ({ success: true, data: { admin: true } }),
    } as Response);

    localStorage.setItem('cloud_token', 'cloud-access-token');

    await expect(authApi.checkAdminAccess()).resolves.toMatchObject({ admin: true });

    expect(fetchMock).toHaveBeenCalledWith(
      expect.stringContaining('partflow-api.onrender.com/api/v1/auth/admin-check'),
      expect.objectContaining({
        method: 'GET',
        headers: expect.objectContaining({
          Authorization: 'Bearer cloud-access-token',
          'X-PartFlow-Cloud-Token': 'cloud-access-token',
        }),
      }),
    );
  });

  it('keeps ordinary permission errors separate from subscription expiry', async () => {
    window.location.hash = '#/app/settings';
    const fetchMock = vi.mocked(globalThis.fetch);
    fetchMock.mockResolvedValueOnce({
      ok: false,
      status: 403,
      text: async () => JSON.stringify({
        error: { code: 'ADMIN_REQUIRED', message: 'administrator privileges required' },
      }),
    } as Response);

    await expect(apiClient.get('/auth/admin-check', undefined, false)).rejects.toMatchObject({
      status: 403,
      code: 'ADMIN_REQUIRED',
    });
    expect(window.location.hash).toBe('#/app/settings');
  });

  it('validates mutations with the cloud token instead of the local JWT', async () => {
    const fetchMock = vi.mocked(globalThis.fetch);
    localStorage.setItem('cloud_token', 'cloud-access-token');
    apiClient.setToken('local-jwt-token');

    fetchMock
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ({ success: true, data: { valid: true } }),
      } as Response)
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        text: async () => JSON.stringify({ success: true, data: { id: 'sale-1' } }),
      } as Response);

    await expect(apiClient.post('/sales', { total: 10 })).resolves.toMatchObject({
      success: true,
    });

    expect(fetchMock.mock.calls[0][0]).toEqual(
      expect.stringContaining('partflow-api.onrender.com/api/v1/auth/validate'),
    );
    expect(fetchMock.mock.calls[0][1]).toEqual(expect.objectContaining({
      headers: expect.objectContaining({ Authorization: 'Bearer cloud-access-token' }),
    }));
    expect(fetchMock.mock.calls[1][1]).toEqual(expect.objectContaining({
      headers: expect.objectContaining({
        Authorization: 'Bearer local-jwt-token',
        'X-PartFlow-Cloud-Token': 'cloud-access-token',
      }),
    }));
  });

  it('does not validate mutations with the local JWT when cloud_token is missing', async () => {
    const fetchMock = vi.mocked(globalThis.fetch);
    apiClient.setToken('local-jwt-token');

    await expect(apiClient.post('/sales', { total: 10 })).rejects.toThrow('يلزم تسجيل الدخول');
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it('stops retrying token refresh after the first failed refresh attempt', async () => {
    const fetchMock = vi.mocked(globalThis.fetch);

    fetchMock
      .mockResolvedValueOnce({
        ok: false,
        status: 401,
        text: async () => JSON.stringify({ error: { message: 'unauthorized' } }),
      } as Response)
      .mockResolvedValueOnce({
        ok: false,
        status: 401,
        text: async () => JSON.stringify({ error: { message: 'refresh failed' } }),
      } as Response)
      .mockResolvedValueOnce({
        ok: false,
        status: 401,
        text: async () => JSON.stringify({ error: { message: 'unauthorized' } }),
      } as Response);

    await expect(apiClient.get('/customers-refresh-storm')).rejects.toMatchObject({
      status: 401,
      code: 'AUTH_REFRESH_FAILED',
    });

    await expect(apiClient.get('/customers-refresh-storm')).rejects.toMatchObject({
      status: 401,
      code: 'AUTH_REFRESH_FAILED',
    });

    expect(fetchMock).toHaveBeenCalledTimes(3);
    const urls = fetchMock.mock.calls.map(([url]) => String(url));
    expect(urls.filter((url) => url.includes('/auth/refresh'))).toHaveLength(1);
  });
});
