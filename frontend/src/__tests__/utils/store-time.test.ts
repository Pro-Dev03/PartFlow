import { describe, expect, it, beforeEach } from 'vitest';
import {
  formatStoreDate,
  formatStoreDateTime,
  getRegionalProfile,
  getStoreDateKey,
  getStoreMonthBounds,
  parseBackendTimestamp,
  setRegionalProfile,
  storeDateToUTCISOString,
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

  it('interprets naive stored timestamps as UTC regardless of the browser timezone', () => {
    expect(parseBackendTimestamp('2026-09-16 23:59:00')?.toISOString()).toBe('2026-09-16T23:59:00.000Z');
    expect(formatStoreDateTime('2026-09-16 21:00:00')).toBe(formatStoreDateTime('2026-09-16T21:00:00Z'));
  });

  it('uses the persisted store timezone and applies its winter and summer offsets', () => {
    setRegionalProfile({
      country_code: 'IL',
      country_name: 'Israel',
      timezone: 'America/New_York',
      date_format: 'DD/MM/YYYY',
      time_format: '24h',
      currency: 'ILS',
      locale: 'en-GB',
    });

    expect(getRegionalProfile().timezone).toBe('America/New_York');
    expect(getStoreDateKey('2026-01-16T04:59:00Z')).toBe('2026-01-15');
    expect(getStoreDateKey('2026-01-16T05:00:00Z')).toBe('2026-01-16');
    expect(getStoreDateKey('2026-06-16T03:59:00Z')).toBe('2026-06-15');
    expect(getStoreDateKey('2026-06-16T04:00:00Z')).toBe('2026-06-16');
    expect(storeDateToUTCISOString('2026-01-15')).toBe('2026-01-15T05:00:00.000Z');
    expect(storeDateToUTCISOString('2026-06-15')).toBe('2026-06-15T04:00:00.000Z');
    expect(storeDateToUTCISOString('2026-02-30')).toBeNull();
  });

  it('changes business date when the configured timezone changes', () => {
    const profile = getRegionalProfile();
    const instant = '2026-09-16T23:30:00Z';
    expect(getStoreDateKey(instant)).toBe('2026-09-17');

    setRegionalProfile({ ...profile, timezone: 'America/New_York' });
    expect(getStoreDateKey(instant)).toBe('2026-09-16');
    expect(formatStoreDate('2026-09-16', 'en-US')).toBe('09/16/2026');

    setRegionalProfile({ ...profile, timezone: 'Asia/Tokyo' });
    expect(getStoreDateKey(instant)).toBe('2026-09-17');
  });

  it('finds the first real instant of a store day when local midnight is skipped', () => {
    const profile = getRegionalProfile();
    setRegionalProfile({ ...profile, timezone: 'America/Sao_Paulo' });

    expect(storeDateToUTCISOString('2018-11-04')).toBe('2018-11-04T03:00:00.000Z');

    setRegionalProfile({ ...profile, timezone: 'Pacific/Apia' });
    expect(storeDateToUTCISOString('2011-12-30')).toBeNull();
  });

  it('calculates the store month independently from the browser timezone', () => {
    expect(getStoreMonthBounds(new Date('2026-09-16T21:00:00Z'))).toEqual({
      start: '2026-09-01',
      end: '2026-10-01',
    });
  });
});
