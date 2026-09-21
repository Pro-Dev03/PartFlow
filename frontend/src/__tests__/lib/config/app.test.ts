import { describe, expect, it, beforeEach, vi } from 'vitest';
import { getActiveApiUrl, getCloudApiUrl, getConnectionMode, setCloudApiUrl, setConnectionMode, shouldUseLocalApi } from '../../../lib/config/app';

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

  it('routes all business operations to the cloud when cloud mode is selected', () => {
    setConnectionMode('cloud');

    expect(getConnectionMode()).toBe('cloud');
    expect(getActiveApiUrl()).toBe('https://partflow-api.onrender.com/api/v1');
    expect(shouldUseLocalApi('localhost')).toBe(false);
  });

  it('uses a saved cloud API URL override for auth and sync routes', () => {
    setCloudApiUrl('https://cloud.example.com/api/v1/');
    setConnectionMode('cloud');

    expect(getCloudApiUrl()).toBe('https://cloud.example.com/api/v1');
    expect(getActiveApiUrl()).toBe('https://cloud.example.com/api/v1');
    expect(getConnectionMode()).toBe('cloud');
    expect(shouldUseLocalApi('localhost')).toBe(false);
  });

  it('never triggers automatic initial sync because cloud sync is manual', async () => {
    const { isInitialSyncNeeded } = await import('../../../hooks/useInitialDataSync');

    expect(isInitialSyncNeeded()).toBe(false);

    localStorage.setItem('partflow-auto-sync-enabled', 'true');
    expect(isInitialSyncNeeded()).toBe(false);
  });
});
