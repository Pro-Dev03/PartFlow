import { describe, expect, it } from 'vitest';
import { combinePurchasesAndSupplierReports } from '../../../features/reports/hooks/useReports';

describe('combinePurchasesAndSupplierReports', () => {
  it('keeps period purchases separate from current supplier balances', () => {
    const periodSupplier = { supplier_name: 'Period supplier', total_cost: 200, item_count: 4 };
    const currentSupplier = { supplier_name: 'Current supplier', total_purchases: 900, total_paid: 700, outstanding: 200 };

    const report = combinePurchasesAndSupplierReports(
      { data: { total_purchases: 2, total_cost: 200, by_supplier: [periodSupplier] } },
      { data: { total_purchases: 900, total_paid: 700, total_outstanding: 200, by_supplier: [currentSupplier] } },
    );

    expect(report.total_purchases).toBe(2);
    expect(report.total_cost).toBe(200);
    expect(report.by_supplier).toEqual([periodSupplier]);
    expect(report.suppliers_report.total_purchases).toBe(900);
    expect(report.suppliers_report.by_supplier).toEqual([currentSupplier]);
  });
});
