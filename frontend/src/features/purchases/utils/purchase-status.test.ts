import { describe, expect, it } from 'vitest';
import {
  isReceivedPurchaseStatus,
  matchesPurchaseViewFilter,
  normalizePurchaseStatus,
} from './purchase-status';

describe('purchase status utilities', () => {
  it('normalizes legacy and partial statuses to the backend format', () => {
    expect(normalizePurchaseStatus('completed')).toBe('received');
    expect(normalizePurchaseStatus('COMPLETED')).toBe('received');
    expect(normalizePurchaseStatus('partially_received')).toBe('partially_received');
    expect(normalizePurchaseStatus('PARTIALLY_RECEIVED')).toBe('partially_received');
    expect(normalizePurchaseStatus('canceled')).toBe('cancelled');
  });

  it('treats partial and completed receipts as received', () => {
    expect(isReceivedPurchaseStatus('received')).toBe(true);
    expect(isReceivedPurchaseStatus('completed')).toBe(true);
    expect(isReceivedPurchaseStatus('partially_received')).toBe(true);
    expect(isReceivedPurchaseStatus('pending')).toBe(false);
  });

  it('matches the received view filter for real purchase states', () => {
    expect(matchesPurchaseViewFilter('received', 'received')).toBe(true);
    expect(matchesPurchaseViewFilter('completed', 'received')).toBe(true);
    expect(matchesPurchaseViewFilter('partially_received', 'received')).toBe(true);
    expect(matchesPurchaseViewFilter('partially received', 'received')).toBe(true);
    expect(matchesPurchaseViewFilter('PARTIALLY RECEIVED', 'received')).toBe(true);
    expect(matchesPurchaseViewFilter('pending', 'received')).toBe(false);
  });
});
