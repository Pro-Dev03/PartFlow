const getEnvValue = (...keys: string[]) => {
  for (const key of keys) {
    const value = import.meta.env[key];
    if (typeof value === 'string' && value.trim() !== '') {
      return value.trim();
    }
  }
  return '';
};

const connectionMode = (getEnvValue('VITE_API_CONNECTION_MODE', 'VITE_API_MODE') || 'default').toLowerCase();

const apiUrl = (() => {
  if (connectionMode === 'local') {
    return getEnvValue('VITE_API_BASE_URL_LOCAL', 'VITE_API_URL_LOCAL', 'VITE_API_URL', 'http://localhost:8080/api/v1');
  }

  if (connectionMode === 'cloud') {
    return getEnvValue('VITE_API_BASE_URL_CLOUD', 'VITE_API_URL_CLOUD', 'VITE_API_BASE_URL', 'VITE_API_URL', 'http://localhost:8080/api/v1');
  }

  return getEnvValue('VITE_API_BASE_URL', 'VITE_API_URL', 'http://localhost:8080/api/v1');
})();

export const appConfig = {
  name: 'PartFlow',
  version: '1.0.0',
  apiUrl,
  defaultLanguage: 'ar',
  supportedLanguages: ['ar', 'en'],
  currency: 'ILS',
  timezone: 'Asia/Jerusalem',
} as const;

export type AppConfig = typeof appConfig;