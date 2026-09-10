import { describe, expect, it } from 'vitest';
import { buildManualProductPayload } from './manualProductPayload';

describe('buildManualProductPayload', () => {
  it('includes required backend fields for a manual product', () => {
    const payload = buildManualProductPayload({
      name: 'E2E Product 20260909',
      price: '120',
      quantity: '2',
      barcode: 'E2E-BAR-200',
    });

    expect(payload).toMatchObject({
      name: 'E2E Product 20260909',
      selling_price: 120,
      cost_price: 84,
      stock: 2,
      condition: 'new',
    });
    expect(payload.barcode).toBe('E2E-BAR-200');
    expect(payload.sku).toMatch(/^SKU-/);
  });

  it('creates a fallback SKU when no barcode is provided', () => {
    const payload = buildManualProductPayload({
      name: 'Fallback SKU Product',
      price: '75',
      quantity: '1',
      barcode: '',
    });

    expect(payload.sku).toMatch(/^SKU-/);
    expect(payload.name).toBe('Fallback SKU Product');
    expect(payload.selling_price).toBe(75);
  });
});
