import { beforeEach, describe, expect, it, vi } from 'vitest';

const { lookupProductMetadata } = vi.hoisted(() => ({ lookupProductMetadata: vi.fn() }));

vi.mock('../../services/api/endpoints', () => ({
  barcodeApi: { lookupProductMetadata },
}));

import { clearBarcodeLookupFields, lookupProductByBarcode, normalizeBarcode, normalizeBarcodeLookupResult } from '../productBarcodeLookup';

describe('normalizeBarcode', () => {
  it('removes spaces and hyphens while preserving leading zeros', () => {
    expect(normalizeBarcode(' 0 360-00291452 ')).toBe('036000291452');
  });

  it('rejects invalid barcode values', () => {
    expect(normalizeBarcode('ABC-123')).toBeNull();
    expect(normalizeBarcode('1234567')).toBeNull();
  });
});

describe('normalizeBarcodeLookupResult', () => {
  it('keeps descriptive metadata and excludes operational prices', () => {
    const result = normalizeBarcodeLookupResult({
      barcode: '123456789012',
      source: 'upcitemdb',
      name: 'Wireless Mouse',
      brand: 'Logitech',
      category: 'Electronics',
      description: 'Ergonomic wireless mouse',
      image: 'https://example.com/mouse.png',
      costPrice: 14.5,
      sellingPrice: 29.99,
    });

    expect(result).toMatchObject({ barcode: '123456789012', name: 'Wireless Mouse', source: 'upcitemdb' });
    expect(result).not.toHaveProperty('costPrice');
    expect(result).not.toHaveProperty('sellingPrice');
  });
});

describe('lookupProductByBarcode', () => {
  beforeEach(() => {
    localStorage.clear();
    lookupProductMetadata.mockReset();
  });

  it('uses the backend result for a real barcode without importing prices', async () => {
    lookupProductMetadata.mockResolvedValue({
      data: {
        barcode: '036000291452',
        source: 'upcitemdb',
        name: 'Barcode Product',
        brand: 'Example Brand',
        category: 'General',
        description: 'Product metadata',
        image: 'https://example.com/product.png',
        costPrice: 10,
        sellingPrice: 20,
      },
    });

    const result = await lookupProductByBarcode('036000291452');

    expect(result?.name).toBe('Barcode Product');
    expect(result?.source).toBe('upcitemdb');
    expect(result).not.toHaveProperty('costPrice');
    expect(result).not.toHaveProperty('sellingPrice');
    expect(lookupProductMetadata).toHaveBeenCalledTimes(1);
  });

  it('uses local cache on the second lookup without another backend request', async () => {
    lookupProductMetadata.mockResolvedValue({
      data: { barcode: '511111111111', source: 'upcitemdb', name: 'Cached Product' },
    });

    const first = await lookupProductByBarcode('511111111111');
    const second = await lookupProductByBarcode('511111111111');

    expect(first?.name).toBe('Cached Product');
    expect(second?.source).toBe('cache');
    expect(lookupProductMetadata).toHaveBeenCalledTimes(1);
  });

  it('does not cache or invent data when the backend reports not found', async () => {
    lookupProductMetadata.mockResolvedValue({ data: undefined });

    const result = await lookupProductByBarcode('9999999999999');

    expect(result).toBeNull();
    expect(localStorage.getItem('partflow-product-barcode-cache')).toBeNull();
  });

  it('does not convert a provider failure into a no-result', async () => {
    lookupProductMetadata.mockRejectedValue({ status: 502, code: 'BARCODE_PROVIDER_ERROR' });

    await expect(lookupProductByBarcode('036000291452')).rejects.toMatchObject({
      status: 502,
      code: 'BARCODE_PROVIDER_ERROR',
    });
  });

  it('clears a successful result before an unknown barcode lookup', async () => {
    lookupProductMetadata
      .mockResolvedValueOnce({ data: { barcode: '036000291452', source: 'upcitemdb', name: 'Known Product' } })
      .mockResolvedValueOnce({ data: undefined });

    const known = await lookupProductByBarcode('036000291452');
    let row = {
      name: known?.name || '',
      barcode: '036000291452',
      description: 'Old description',
      image_url: 'old-image.png',
      costPrice: '10',
    };
    row = clearBarcodeLookupFields({ ...row, barcode: '9999999999999' });
    const unknown = await lookupProductByBarcode('9999999999999');

    expect(known?.name).toBe('Known Product');
    expect(unknown).toBeNull();
    expect(row.name).toBe('');
    expect(row.description).toBe('');
    expect(row.image_url).toBe('');
    expect(row.costPrice).toBe('10');
    const cache = JSON.parse(localStorage.getItem('partflow-product-barcode-cache') || '{}');
    expect(cache['036000291452']?.name).toBe('Known Product');
    expect(cache['9999999999999']).toBeUndefined();
  });
});
