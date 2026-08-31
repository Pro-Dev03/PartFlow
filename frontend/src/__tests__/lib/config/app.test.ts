import { describe, expect, it, beforeEach, vi } from 'vitest';
import { getActiveApiUrl, shouldUseLocalApi } from '../../../lib/config/app';

describe('app config', () => {
  beforeEach(() => {
    localStorage.clear();
    vi.restoreAllMocks();
  });

  it('prefers the local backend when running on localhost without an explicit operating mode', () => {
    expect(shouldUseLocalApi('localhost')).toBe(true);
    expect(shouldUseLocalApi('127.0.0.1')).toBe(true);
    expect(shouldUseLocalApi('partflow-api.onrender.com')).toBe(false);
  });

  it('uses the local backend whenever the browser is offline, even if the app is not explicitly set to offline', () => {
    Object.defineProperty(window.navigator, 'onLine', {
      configurable: true,
      get: () => false,
    });

    expect(getActiveApiUrl()).toBe('http://localhost:8080/api/v1');
  });

  it('defaults to the local backend when no operating mode has been chosen yet', () => {
    expect(getActiveApiUrl()).toBe('http://localhost:8080/api/v1');
  });

  it('uses the local backend when the operating mode is explicitly offline', () => {
    localStorage.setItem('partflow-operating-mode', 'offline');

    expect(getActiveApiUrl()).toBe('http://localhost:8080/api/v1');
  });

  it('keeps business operations local even when a legacy online preference exists', () => {
    localStorage.setItem('partflow-operating-mode', 'online');

    expect(getActiveApiUrl()).toBe('http://localhost:8080/api/v1');
  });
});
