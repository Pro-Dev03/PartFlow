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

// Cloud API endpoint - connects to Render's centralized API
// No database credentials here, those stay in Render's environment
const cloudApiUrl = getEnvValue(
  ['VITE_API_BASE_URL_CLOUD', 'VITE_API_URL_CLOUD'],
  'https://partflow-api.onrender.com/api/v1',
);
export const CLOUD_API_URL_OVERRIDE_KEY = 'partflow-cloud-api-url';

export type ConnectionMode = 'local' | 'cloud';
export const CONNECTION_MODE_KEY = 'partflow-connection-mode';

export function getConnectionMode(): ConnectionMode {
  if (typeof window === 'undefined') return 'local';
  return localStorage.getItem(CONNECTION_MODE_KEY) === 'cloud' ? 'cloud' : 'local';
}

export function setConnectionMode(mode: ConnectionMode): void {
  if (typeof window === 'undefined') return;
  localStorage.setItem(CONNECTION_MODE_KEY, mode);
  window.dispatchEvent(new CustomEvent('partflow:connection-mode-changed', { detail: { mode } }));
}

/**
 * Business operations always use the local SQLite API. The cloud URL is used
 * only by authentication/subscription validation helpers.
 */
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

/**
 * Get the cloud API URL for connecting to Render
 */
export function getCloudApiUrl(): string {
  if (typeof window !== 'undefined') {
    const override = localStorage.getItem(CLOUD_API_URL_OVERRIDE_KEY)?.trim();
    if (override) return override.replace(/\/+$/, '');
  }
  return cloudApiUrl;
}

export function setCloudApiUrl(url: string): void {
  if (typeof window === 'undefined') return;
  const normalized = url.trim().replace(/\/+$/, '');
  localStorage.setItem(CLOUD_API_URL_OVERRIDE_KEY, normalized);
  window.dispatchEvent(new CustomEvent('partflow:cloud-api-url-changed', { detail: { url: normalized } }));
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
  apiUrl: getActiveApiUrl(),
  defaultLanguage: 'ar',
  supportedLanguages: ['ar', 'en'],
  currency: 'ILS',
  timezone: 'Asia/Jerusalem',
  developmentMode: getEnvValue(['VITE_DEVELOPMENT_MODE'], 'false') === 'true',
} as const;

export type AppConfig = typeof appConfig;
