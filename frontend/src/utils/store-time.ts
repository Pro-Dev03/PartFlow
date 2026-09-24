import { DEFAULT_REGIONAL_PROFILE, RegionalProfile } from '../types/regional';

const REGIONAL_PROFILE_KEY = 'partflow-regional-profile';
export const STORE_TIMEZONE = 'Asia/Jerusalem';

export function getRegionalProfile(): RegionalProfile {
  if (typeof window === 'undefined') return DEFAULT_REGIONAL_PROFILE;
  try {
    const parsed = JSON.parse(localStorage.getItem(REGIONAL_PROFILE_KEY) || 'null') as Partial<RegionalProfile> | null;
    if (parsed?.timezone && parsed?.country_code) {
      return { ...DEFAULT_REGIONAL_PROFILE, ...parsed, timezone: STORE_TIMEZONE };
    }
  } catch {
    // Fall back to the safe store default when legacy local storage is invalid.
  }
  return DEFAULT_REGIONAL_PROFILE;
}

export function setRegionalProfile(profile: RegionalProfile): void {
  if (typeof window !== 'undefined') {
    const canonicalProfile = { ...profile, timezone: STORE_TIMEZONE };
    localStorage.setItem(REGIONAL_PROFILE_KEY, JSON.stringify(canonicalProfile));
    window.dispatchEvent(new CustomEvent('partflow-regional-profile-changed', { detail: canonicalProfile }));
  }
}

export function getDeviceTimezone(): string {
  return STORE_TIMEZONE;
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
    .replace(' ', 'T')
    .replace(/([+-]\d{2})(\d{2})$/, '$1:$2');
}

export function parseBackendTimestamp(value: string | Date | null | undefined): Date | null {
  if (value instanceof Date) return Number.isNaN(value.getTime()) ? null : value;
  const raw = String(value || '').trim();
  if (!raw) return null;
  let normalized = normalizeBackendTimestamp(raw);
  if (/^\d{4}-\d{2}-\d{2}$/.test(normalized)) {
    normalized = `${normalized}T00:00:00Z`;
  } else if (/^\d{4}-\d{2}-\d{2}T/.test(normalized) && !/(?:Z|[+-]\d{2}:\d{2})$/i.test(normalized)) {
    // Legacy SQLite timestamps without an offset are UTC by storage contract.
    normalized += 'Z';
  }
  const parsed = new Date(normalized);
  return Number.isNaN(parsed.getTime()) ? null : parsed;
}

export function storeDateToUTCISOString(dateKey: string): string | null {
  const match = /^(\d{4})-(\d{2})-(\d{2})$/.exec(dateKey);
  if (!match) return null;
  const [, yearText, monthText, dayText] = match;
  const year = Number(yearText);
  const month = Number(monthText);
  const day = Number(dayText);
  const targetWallTime = Date.UTC(year, month - 1, day);
  const validationDate = new Date(targetWallTime);
  if (validationDate.getUTCFullYear() !== year || validationDate.getUTCMonth() + 1 !== month || validationDate.getUTCDate() !== day) {
    return null;
  }

  let utcTime = targetWallTime;
  const formatter = new Intl.DateTimeFormat('en-CA', {
    timeZone: STORE_TIMEZONE,
    year: 'numeric', month: '2-digit', day: '2-digit',
    hour: '2-digit', minute: '2-digit', second: '2-digit', hourCycle: 'h23',
  });
  for (let attempt = 0; attempt < 3; attempt += 1) {
    const parts = Object.fromEntries(formatter.formatToParts(new Date(utcTime)).map((part) => [part.type, part.value]));
    const renderedWallTime = Date.UTC(
      Number(parts.year), Number(parts.month) - 1, Number(parts.day),
      Number(parts.hour), Number(parts.minute), Number(parts.second),
    );
    utcTime += targetWallTime - renderedWallTime;
  }
  return new Date(utcTime).toISOString();
}

export function formatStoreDate(value: string | Date | null | undefined, locale = getRegionalProfile().locale): string {
  const parsed = parseBackendTimestamp(value);
  if (!parsed) return 'غير محدد';
  return new Intl.DateTimeFormat(locale, {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    timeZone: STORE_TIMEZONE,
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
    timeZone: STORE_TIMEZONE,
  }).format(parsed);
}

export function formatStoreTime(value: string | Date | null | undefined, locale = getRegionalProfile().locale): string {
  const parsed = parseBackendTimestamp(value);
  if (!parsed) return 'غير محدد';
  const profile = getRegionalProfile();
  return new Intl.DateTimeFormat(locale, {
    hour: '2-digit',
    minute: '2-digit',
    hour12: profile.time_format === '12h',
    timeZone: STORE_TIMEZONE,
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
    year: 'numeric', month: '2-digit', day: '2-digit', timeZone: STORE_TIMEZONE,
  }).formatToParts(parsed);
  const values = Object.fromEntries(parts.map((part) => [part.type, part.value]));
  return `${values.year}-${values.month}-${values.day}`;
}
