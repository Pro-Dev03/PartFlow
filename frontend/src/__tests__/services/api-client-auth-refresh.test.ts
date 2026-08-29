import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { apiClient } from '../../services/api/client';

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
});
