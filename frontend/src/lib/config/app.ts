export const appConfig = {
  name: 'PartFlow',
  version: '1.0.0',
  apiUrl: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1',
  defaultLanguage: 'ar',
  supportedLanguages: ['ar', 'en'],
  currency: 'ILS',
  timezone: 'Asia/Jerusalem',
} as const;

export type AppConfig = typeof appConfig;