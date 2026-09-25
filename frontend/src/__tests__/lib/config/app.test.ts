import { describe, expect, it, beforeEach, vi } from 'vitest';
import { CLOUD_API_URL_OVERRIDE_KEY, getActiveApiUrl, getBusinessApiUrl, getCloudApiUrl, getConnectionMode, setCloudApiUrl, setConnectionMode, shouldUseLocalApi } from '../../../lib/config/app';

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

  it('keeps device connection mode separate from the authoritative business API', () => {
    expect(getActiveApiUrl()).toBe('http://localhost:8080/api/v1');
    expect(getBusinessApiUrl()).toBe('https://partflow-api.onrender.com/api/v1');
  });

  it('routes all business operations to the cloud when cloud mode is selected', () => {
    setConnectionMode('cloud');

    expect(getConnectionMode()).toBe('cloud');
    expect(getActiveApiUrl()).toBe('https://partflow-api.onrender.com/api/v1');
    expect(shouldUseLocalApi('localhost')).toBe(false);
  });

  it('keeps the cloud API pinned to the production Render service', () => {
    expect(() => setCloudApiUrl('https://cloud.example.com/api/v1/')).toThrow(/fixed to the configured Render service/);
    localStorage.setItem(CLOUD_API_URL_OVERRIDE_KEY, 'https://attacker.example/api/v1');
    setConnectionMode('cloud');

    expect(getCloudApiUrl()).toBe('https://partflow-api.onrender.com/api/v1');
    expect(getActiveApiUrl()).toBe('https://partflow-api.onrender.com/api/v1');
    expect(getBusinessApiUrl()).toBe('https://partflow-api.onrender.com/api/v1');
    expect(getConnectionMode()).toBe('cloud');
    expect(shouldUseLocalApi('localhost')).toBe(false);
  });

  it('uses Render for business requests even when the device mode is local', () => {
    setConnectionMode('local');

    expect(getActiveApiUrl()).toBe('http://localhost:8080/api/v1');
    expect(getBusinessApiUrl()).toBe('https://partflow-api.onrender.com/api/v1');
  });

  it('never triggers automatic initial sync because cloud sync is manual', async () => {
    const { isInitialSyncNeeded } = await import('../../../hooks/useInitialDataSync');

    expect(isInitialSyncNeeded()).toBe(false);

    localStorage.setItem('partflow-auto-sync-enabled', 'true');
    expect(isInitialSyncNeeded()).toBe(false);
  });
});
