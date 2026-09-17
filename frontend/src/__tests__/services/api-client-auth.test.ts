import { beforeEach, describe, expect, it, vi } from 'vitest';
import { apiClient } from '../../services/api/client';

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
        headers: expect.objectContaining({
          Authorization: 'Bearer local-token',
          'X-PartFlow-Cloud-Token': 'cloud-token',
        }),
      })
    );
  });
});
