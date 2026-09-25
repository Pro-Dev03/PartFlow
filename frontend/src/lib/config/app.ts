const getEnvValue = (keys: string[], fallback = '') => {
  for (const key of keys) {
    const value = import.meta.env[key];
    if (typeof value === 'string' && value.trim() !== '') {
      return value.trim();
    }
  }
  return fallback;
};

const localApiUrl = getEnvValue(
  ['VITE_API_BASE_URL_LOCAL', 'VITE_API_URL_LOCAL', 'VITE_API_BASE_URL', 'VITE_API_URL'],
  'http://localhost:8080/api/v1',
);

// Business authorization and writes are pinned to the production Render API.
// Database credentials remain in Render's server-side environment.
const cloudApiUrl = 'https://partflow-api.onrender.com/api/v1';
export const CLOUD_API_URL_OVERRIDE_KEY = 'partflow-cloud-api-url';

export type ConnectionMode = 'local' | 'cloud';
export const CONNECTION_MODE_KEY = 'partflow-connection-mode';

export function getConnectionMode(): ConnectionMode {
  if (typeof window === 'undefined') return 'cloud';
  const savedMode = localStorage.getItem(CONNECTION_MODE_KEY);
  if (savedMode === 'cloud' || savedMode === 'local') return savedMode;

  const isElectron = window.location.protocol === 'file:' || /Electron/i.test(window.navigator.userAgent);
  const isLocalHost = /localhost|127\.0\.0\.1|0\.0\.0\.0/.test(window.location.hostname);
  return isElectron || isLocalHost ? 'local' : 'cloud';
}

export function setConnectionMode(mode: ConnectionMode): void {
  if (typeof window === 'undefined') return;
  localStorage.setItem(CONNECTION_MODE_KEY, mode);
  window.dispatchEvent(new CustomEvent('partflow:connection-mode-changed', { detail: { mode } }));
}

/** Local APIs are used only for device-owned SQLite synchronization/setup. */
export function shouldUseLocalApi(hostname = typeof window !== 'undefined' ? window.location.hostname : ''): boolean {
  if (getConnectionMode() === 'cloud') return false;

  const isElectron = typeof window !== 'undefined' && (
    window.location.protocol === 'file:' ||
    /Electron/i.test(window.navigator.userAgent)
  );

  return isElectron || /localhost|127\.0\.0\.1|0\.0\.0\.0/.test(hostname);
}

export function getActiveApiUrl(): string {
  return getConnectionMode() === 'cloud' ? getCloudApiUrl() : localApiUrl;
}

/** Business records and authorization always belong to the cloud API. */
export function getBusinessApiUrl(): string {
  return getCloudApiUrl();
}

/**
 * Get the cloud API URL for connecting to Render
 */
export function getCloudApiUrl(): string {
  return cloudApiUrl;
}

export function setCloudApiUrl(url: string): void {
  const normalized = url.trim().replace(/\/+$/, '');
  if (normalized !== cloudApiUrl) {
    throw new Error('The cloud API is fixed to the configured Render service');
  }
  if (typeof window !== 'undefined') {
    localStorage.removeItem(CLOUD_API_URL_OVERRIDE_KEY);
    window.dispatchEvent(new CustomEvent('partflow:cloud-api-url-changed', { detail: { url: cloudApiUrl } }));
  }
}

/**
 * Get the local API URL for business operations
 */
export function getLocalApiUrl(): string {
  return localApiUrl;
}

export const appConfig = {
  name: 'PartFlow',
  version: '1.0.0',
  apiUrl: getBusinessApiUrl(),
  defaultLanguage: 'ar',
  supportedLanguages: ['ar', 'en'],
  currency: 'ILS',
  timezone: 'Asia/Jerusalem',
  developmentMode: getEnvValue(['VITE_DEVELOPMENT_MODE'], 'false') === 'true',
} as const;

export type AppConfig = typeof appConfig;
