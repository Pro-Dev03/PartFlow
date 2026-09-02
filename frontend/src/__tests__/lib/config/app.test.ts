import { describe, expect, it, beforeEach, vi } from 'vitest';
import { getActiveApiUrl, shouldUseLocalApi } from '../../../lib/config/app';

describe('app config', () => {
  beforeEach(() => {
    localStorage.clear();
    vi.restoreAllMocks();
  });

  it('prefers the local backend when running on localhost', () => {
    expect(shouldUseLocalApi('localhost')).toBe(true);
    expect(shouldUseLocalApi('127.0.0.1')).toBe(true);
    expect(shouldUseLocalApi('partflow-api.onrender.com')).toBe(false);
  });

  it('always uses the local backend for business operations', () => {
    expect(getActiveApiUrl()).toBe('http://localhost:8080/api/v1');
  });

  it('does not trigger automatic initial sync unless the user explicitly enables it', async () => {
    const { isInitialSyncNeeded } = await import('../../../hooks/useInitialDataSync');

    expect(isInitialSyncNeeded()).toBe(false);

    localStorage.setItem('partflow-auto-sync-enabled', 'true');
    expect(isInitialSyncNeeded()).toBe(true);
  });
});
