import { describe, expect, it } from 'vitest';
import { normalizeSupplier } from './supplier-normalization';

describe('normalizeSupplier', () => {
  it('maps backend snake_case fields to the camelCase shape used by the supplier cards', () => {
    const normalized = normalizeSupplier({
      id: 'supplier-1',
      name: 'مورد 1',
      code: 'SUP-001',
      phone: '0500000000',
      total_purchases: 1500,
      paid_amount: 600,
      outstanding: 900,
      last_purchase: '2026-09-16T00:00:00Z',
    });

    expect(normalized.totalPurchases).toBe(1500);
    expect(normalized.paidAmount).toBe(600);
    expect(normalized.outstanding).toBe(900);
    expect(normalized.lastPurchase).toBe('2026-09-16T00:00:00Z');
  });

  it('keeps camelCase values when the backend already sends them', () => {
    const normalized = normalizeSupplier({
      id: 'supplier-2',
      name: 'مورد 2',
      code: 'SUP-002',
      totalPurchases: 200,
      paidAmount: 50,
      outstanding: 150,
    });

    expect(normalized.totalPurchases).toBe(200);
    expect(normalized.paidAmount).toBe(50);
    expect(normalized.outstanding).toBe(150);
  });
});
