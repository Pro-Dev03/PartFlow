import { describe, expect, it } from 'vitest';
import { createProductCsvTemplate, parseProductCsv } from './productCsvImport';

describe('product CSV import', () => {
  it('parses product rows and preserves barcode leading zeros', () => {
    const rows = parseProductCsv('name,barcode,cost_price,selling_price\n"Coffee, Large",08993242596979,2.5,4');
    expect(rows).toEqual([{ name: 'Coffee, Large', barcode: '08993242596979', costPrice: '2.5', sellingPrice: '4', quantity: '0', minStockLevel: '0', supplierId: '' }]);
  });

  it('rejects files without required identity columns', () => {
    expect(() => parseProductCsv('name,cost_price\nCoffee,2')).toThrow('barcode');
  });

  it('provides an import template without operational defaults', () => {
    expect(createProductCsvTemplate()).toContain('name,barcode,cost_price,selling_price');
    expect(createProductCsvTemplate()).toContain('8854419001507');
  });
});