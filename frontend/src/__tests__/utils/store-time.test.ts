import { describe, expect, it, beforeEach } from 'vitest';
import {
  formatStoreDate,
  getStoreDateKey,
  getStoreMonthBounds,
  parseBackendTimestamp,
  setRegionalProfile,
} from '../../utils/store-time';

describe('store time utilities', () => {
  beforeEach(() => {
    setRegionalProfile({
      country_code: 'IL',
      country_name: 'Israel',
      timezone: 'Asia/Jerusalem',
      date_format: 'DD/MM/YYYY',
      time_format: '24h',
      currency: 'ILS',
      locale: 'ar',
    });
  });

  it('keeps store-day boundaries at 23:59, 00:00, and 00:01', () => {
    expect(getStoreDateKey('2026-09-16T20:59:00Z')).toBe('2026-09-16');
    expect(getStoreDateKey('2026-09-16T21:00:00Z')).toBe('2026-09-17');
    expect(getStoreDateKey('2026-09-16T21:01:00Z')).toBe('2026-09-17');
  });

  it('parses legacy Go timestamps without Invalid Date', () => {
    expect(parseBackendTimestamp('2026-09-16 23:59:00 +0300 EEST')).not.toBeNull();
    expect(parseBackendTimestamp('2026-09-16 23:59:00 +0300 EEST m=+6169.648151801')).not.toBeNull();
    expect(formatStoreDate('2026-09-16 23:59:00 +0300 EEST', 'en-GB')).toContain('16');
  });

  it('calculates the store month independently from the browser timezone', () => {
    expect(getStoreMonthBounds(new Date('2026-09-16T21:00:00Z'))).toEqual({
      start: '2026-09-01',
      end: '2026-10-01',
    });
  });
});