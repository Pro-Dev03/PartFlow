export interface RegionalProfile {
  country_code: string;
  country_name: string;
  timezone: string;
  date_format: string;
  time_format: '12h' | '24h';
  currency: string;
  locale: string;
}

export const DEFAULT_REGIONAL_PROFILE: RegionalProfile = {
  country_code: 'IL',
  country_name: 'إسرائيل',
  timezone: 'Asia/Jerusalem',
  date_format: 'DD/MM/YYYY',
  time_format: '12h',
  currency: 'ILS',
  locale: 'ar',
};
