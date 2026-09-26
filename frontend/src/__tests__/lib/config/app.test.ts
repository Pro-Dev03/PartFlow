import { describe, expect, it, beforeEach, vi } from 'vitest';
import { CLOUD_API_URL_OVERRIDE_KEY, getActiveApiUrl, getBusinessApiUrl, getCloudApiUrl, getConnectionMode, isElectronRuntime, setCloudApiUrl, setConnectionMode, shouldUseLocalApi } from '../../../lib/config/app';

describe('app config', () => {
  beforeEach(() => {
    localStorage.clear();
    delete window.partflowDesktop;
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

  it('forces Electron into local device mode even if cloud mode was saved by an older build', () => {
    window.partflowDesktop = { appVersion: 'test', platform: 'win32' };
    localStorage.setItem('partflow-connection-mode', 'cloud');

    expect(getConnectionMode()).toBe('local');
    setConnectionMode('cloud');
    expect(getConnectionMode()).toBe('local');
    expect(localStorage.getItem('partflow-connection-mode')).toBe('local');
    expect(getBusinessApiUrl()).toBe('https://partflow-api.onrender.com/api/v1');
  });

  it('does not classify a hosted browser as Electron from its user agent', () => {
    Object.defineProperty(window.navigator, 'userAgent', {
      configurable: true,
      value: 'Mozilla/5.0 Electron/42.8.1',
    });
    localStorage.setItem('partflow-connection-mode', 'cloud');

    expect(isElectronRuntime()).toBe(false);
    expect(getConnectionMode()).toBe('cloud');
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

  it('routes development business requests locally while keeping cloud auth on Render', async () => {
    vi.stubEnv('VITE_DEPLOYMENT_MODE', 'development');
    vi.stubEnv('VITE_DEVELOPMENT_API_URL', 'http://127.0.0.1:8080/api/v1');
    vi.resetModules();

    const developmentConfig = await import('../../../lib/config/app');
    expect(developmentConfig.appConfig.developmentMode).toBe(true);
    expect(developmentConfig.getBusinessApiUrl()).toBe('http://127.0.0.1:8080/api/v1');
    expect(developmentConfig.getCloudApiUrl()).toBe('https://partflow-api.onrender.com/api/v1');

    vi.unstubAllEnvs();
    vi.resetModules();
  });

  it('keeps production builds pinned to Render even when an API override is present', async () => {
    vi.stubEnv('VITE_DEPLOYMENT_MODE', 'production');
    vi.stubEnv('VITE_CLOUD_API_URL', 'http://127.0.0.1:8080/api/v1');
    vi.resetModules();

    const productionConfig = await import('../../../lib/config/app');
    expect(productionConfig.appConfig.developmentMode).toBe(false);
    expect(productionConfig.getBusinessApiUrl()).toBe('https://partflow-api.onrender.com/api/v1');

    vi.unstubAllEnvs();
    vi.resetModules();
  });

  it('never triggers automatic initial sync because cloud sync is manual', async () => {
    const { isInitialSyncNeeded } = await import('../../../hooks/useInitialDataSync');

    expect(isInitialSyncNeeded()).toBe(false);

    localStorage.setItem('partflow-auto-sync-enabled', 'true');
    expect(isInitialSyncNeeded()).toBe(false);
  });
});
