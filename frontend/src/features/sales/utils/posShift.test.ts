import { describe, expect, it } from 'vitest';
import { resolvePosShiftState } from './posShift';

describe('resolvePosShiftState', () => {
  it('defaults to a closed shift when the backend has no active shift', () => {
    expect(resolvePosShiftState(null)).toMatchObject({
      status: 'closed',
      salesTotal: 0,
      saleCount: 0,
    });
  });

  it('keeps a remote closed shift as closed', () => {
    expect(resolvePosShiftState({ status: 'closed', opened_at: '2026-09-20T00:00:00Z', opening_cash: 100 })).toMatchObject({
      status: 'closed',
      openingCash: 100,
    });
  });

  it('maps an open remote shift without changing business values', () => {
    expect(resolvePosShiftState({ status: 'open', opened_at: '2026-09-20T00:00:00Z', opening_cash: 50, sales_total: 120, sale_count: 3 })).toMatchObject({
      status: 'open',
      openingCash: 50,
      salesTotal: 120,
      saleCount: 3,
    });
  });
});
