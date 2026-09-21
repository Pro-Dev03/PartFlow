import { beforeEach, describe, expect, it, vi } from 'vitest';
import { apiClient } from '../../services/api/client';
import { authApi } from '../../services/api/endpoints';

describe('apiClient auth propagation', () => {
  beforeEach(() => {
    localStorage.clear();
    vi.restoreAllMocks();
  });

  it('adds both local Authorization and cloud token headers to protected requests', async () => {
    localStorage.setItem('auth_token', 'local-token');
    localStorage.setItem('cloud_token', 'cloud-token');

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
          Authorization: 'Bearer local-token',
          'X-PartFlow-Cloud-Token': 'cloud-token',
        }),
      })
    );
  });

  it('keeps the session credentials when a 401 cannot be refreshed', async () => {
    localStorage.setItem('auth_token', 'local-token');
    localStorage.setItem('cloud_token', 'cloud-token');
    apiClient.setToken('local-token');
    apiClient.setCloudToken('cloud-token');

    const invalidated = vi.fn();
    window.addEventListener('partflow:auth-invalidated', invalidated);
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
      code: 'AUTH_REFRESH_FAILED',
    });

    expect(invalidated).toHaveBeenCalledTimes(1);
    expect(localStorage.getItem('auth_token')).toBe('local-token');
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
          Authorization: 'Bearer local-token',
          'X-PartFlow-Cloud-Token': 'cloud-token',
        }),
      })
    );
  });
});
