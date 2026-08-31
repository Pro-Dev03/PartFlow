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

/**
 * Business operations always use the local SQLite API. The cloud URL is used
 * only by authentication/subscription validation helpers.
 */
export function shouldUseLocalApi(hostname = typeof window !== 'undefined' ? window.location.hostname : ''): boolean {
  return /localhost|127\.0\.0\.1|0\.0\.0\.0/.test(hostname);
}

export function getActiveApiUrl(): string {
  return localApiUrl;
}

/**
 * Get the cloud API URL for connecting to Render
 */
export function getCloudApiUrl(): string {
  return cloudApiUrl;
}

/**
 * Get the local API URL for offline mode
 */
export function getLocalApiUrl(): string {
  return localApiUrl;
}

/**
 * Get current operating mode
 */
export function getOperatingMode(): 'offline' | 'online' {
  if (typeof window === 'undefined') {
    return 'offline';
  }

  const storedMode = localStorage.getItem('partflow-operating-mode');
  if (storedMode === 'offline') {
    return 'offline';
  }

  if (storedMode === 'online') {
    return 'online';
  }

  // Safe default: PartFlow is local-first and offline by default unless the user
  // explicitly chooses the cloud/sync mode.
  return 'offline';
}

export function setOperatingModePreference(mode: 'offline' | 'online'): 'offline' | 'online' {
  const nextMode = mode === 'offline' ? 'offline' : 'online';

  if (typeof window !== 'undefined') {
    localStorage.setItem('partflow-operating-mode', nextMode);
  }

  return nextMode;
}

export const appConfig = {
  name: 'PartFlow',
  version: '1.0.0',
  apiUrl: getActiveApiUrl(),
  defaultLanguage: 'ar',
  supportedLanguages: ['ar', 'en'],
  currency: 'ILS',
  timezone: 'Asia/Jerusalem',
} as const;

export type AppConfig = typeof appConfig;
