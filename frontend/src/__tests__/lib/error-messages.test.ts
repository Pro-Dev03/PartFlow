import { describe, expect, it } from 'vitest';
import { isNetworkError } from '../../lib/error-messages';

describe('isNetworkError', () => {
  it.each([502, 503, 504])('treats HTTP %i as a service connection failure', (status) => {
    expect(isNetworkError({ status })).toBe(true);
  });

  it('recognizes fetch failures from Node and Chromium', () => {
    expect(isNetworkError(new TypeError('fetch failed'))).toBe(true);
    expect(isNetworkError(new TypeError('Failed to fetch'))).toBe(true);
  });

  it('recognizes DNS and connection errors from the underlying cause', () => {
    expect(isNetworkError({ cause: { code: 'ENOTFOUND' } })).toBe(true);
    expect(isNetworkError({ cause: { code: 'ECONNREFUSED' } })).toBe(true);
  });

  it('does not treat authentication failures as network errors', () => {
    expect(isNetworkError({ status: 401, message: 'invalid credentials' })).toBe(false);
  });
});
