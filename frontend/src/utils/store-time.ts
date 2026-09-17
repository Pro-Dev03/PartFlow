import { DEFAULT_REGIONAL_PROFILE, RegionalProfile } from '../types/regional';

const REGIONAL_PROFILE_KEY = 'partflow-regional-profile';

export function getRegionalProfile(): RegionalProfile {
  if (typeof window === 'undefined') return DEFAULT_REGIONAL_PROFILE;
  try {
    const parsed = JSON.parse(localStorage.getItem(REGIONAL_PROFILE_KEY) || 'null') as Partial<RegionalProfile> | null;
    if (parsed?.timezone && parsed?.country_code) {
      return { ...DEFAULT_REGIONAL_PROFILE, ...parsed };
    }
  } catch {
    // Fall back to the safe store default when legacy local storage is invalid.
  }
  return DEFAULT_REGIONAL_PROFILE;
}

export function setRegionalProfile(profile: RegionalProfile): void {
  if (typeof window !== 'undefined') {
    localStorage.setItem(REGIONAL_PROFILE_KEY, JSON.stringify(profile));
    window.dispatchEvent(new CustomEvent('partflow-regional-profile-changed', { detail: profile }));
  }
}

export function getDeviceTimezone(): string {
  try {
    return Intl.DateTimeFormat().resolvedOptions().timeZone || DEFAULT_REGIONAL_PROFILE.timezone;
  } catch {
    return DEFAULT_REGIONAL_PROFILE.timezone;
  }
}

export function getStoreToday(value: Date = new Date()): string {
  return getStoreDateKey(value) || '';
}

export function addStoreDays(dateKey: string, days: number): string {
  const date = new Date(`${dateKey}T12:00:00Z`);
  if (Number.isNaN(date.getTime())) return '';
  date.setUTCDate(date.getUTCDate() + days);
  return date.toISOString().slice(0, 10);
}

export function getStoreMonthBounds(value: Date = new Date()): { start: string; end: string } {
  const today = getStoreToday(value);
  if (!today) return { start: '', end: '' };
  const [year, month] = today.split('-').map(Number);
  const nextMonth = new Date(Date.UTC(year, month, 1));
  return {
    start: `${year}-${String(month).padStart(2, '0')}-01`,
    end: nextMonth.toISOString().slice(0, 10),
  };
}

function normalizeBackendTimestamp(value: string): string {
  return value
    .trim()
    .replace(/\s+m=[+-]?\d+(?:\.\d+)?$/, '')
    .replace(/\s+[A-Z]{2,5}$/, '')
    .replace(/\s([+-]\d{2}:?\d{2})$/, '$1')
    .replace(' ', 'T');
}

export function parseBackendTimestamp(value: string | Date | null | undefined): Date | null {
  if (value instanceof Date) return Number.isNaN(value.getTime()) ? null : value;
  const raw = String(value || '').trim();
  if (!raw) return null;
  const parsed = new Date(normalizeBackendTimestamp(raw));
  return Number.isNaN(parsed.getTime()) ? null : parsed;
}

export function formatStoreDate(value: string | Date | null | undefined, locale = getRegionalProfile().locale): string {
  const parsed = parseBackendTimestamp(value);
  if (!parsed) return 'غير محدد';
  const profile = getRegionalProfile();
  return new Intl.DateTimeFormat(locale, {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    timeZone: profile.timezone,
  }).format(parsed);
}

export function formatStoreDateTime(value: string | Date | null | undefined, locale = getRegionalProfile().locale): string {
  const parsed = parseBackendTimestamp(value);
  if (!parsed) return 'غير محدد';
  const profile = getRegionalProfile();
  return new Intl.DateTimeFormat(locale, {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    hour12: profile.time_format === '12h',
    timeZone: profile.timezone,
  }).format(parsed);
}

export function formatStoreActivityDateTime(
  eventTime: string | Date | null | undefined,
  businessDate?: string | null,
  locale = getRegionalProfile().locale,
): string {
  const formattedTime = formatStoreDateTime(eventTime, locale);
  if (!businessDate) return formattedTime;

  const formattedDate = formatStoreDate(businessDate, locale);
  const separatorIndex = Math.max(formattedTime.lastIndexOf('،'), formattedTime.lastIndexOf(','));
  if (formattedDate === 'غير محدد' || separatorIndex < 0) return formattedTime;

  return `${formattedDate}،${formattedTime.slice(separatorIndex + 1)}`;
}

export function getStoreDateKey(value: string | Date | null | undefined): string | null {
  const parsed = parseBackendTimestamp(value);
  if (!parsed) return null;
  const profile = getRegionalProfile();
  const parts = new Intl.DateTimeFormat('en-CA', {
    year: 'numeric', month: '2-digit', day: '2-digit', timeZone: profile.timezone,
  }).formatToParts(parsed);
  const values = Object.fromEntries(parts.map((part) => [part.type, part.value]));
  return `${values.year}-${values.month}-${values.day}`;
}
