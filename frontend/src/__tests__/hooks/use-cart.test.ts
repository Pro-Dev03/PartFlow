import { act, renderHook } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { useCart, normalizePosPrice } from '../../features/sales/hooks/useCart';

vi.mock('../../hooks/useBarcodeContext', () => ({
  playScanSound: vi.fn(),
}));

const product = {
  id: 'gpu-1',
  name: 'RTX 4060',
  barcode: '123',
  selling_price: 1450,
  cost_price: 1000,
};

describe('useCart', () => {
  it('normalizes the first finite price value', () => {
    expect(normalizePosPrice(undefined, '', '1450')).toBe(1450);
    expect(normalizePosPrice('invalid', null, undefined)).toBe(0);
  });

  it('groups duplicate products and preserves totals', () => {
    const { result } = renderHook(() => useCart(false));

    act(() => {
      result.current.addToCart(product);
      result.current.addToCart(product, 2);
    });

    expect(result.current.cart).toHaveLength(1);
    expect(result.current.cart[0].quantity).toBe(3);
    expect(result.current.cart[0].total).toBe(4350);
    expect(result.current.total).toBe(4350);
  });

  it('treats product prices as tax-inclusive', () => {
    const { result } = renderHook(() => useCart(false, 10));

    act(() => result.current.addToCart(product));

    expect(result.current.subtotal).toBe(1450);
    expect(result.current.total).toBe(1450);
  });

  it('updates quantity and removes an item when quantity reaches zero', () => {
    const { result } = renderHook(() => useCart(false));

    act(() => result.current.addToCart(product, 2));
    act(() => result.current.updateQuantity('123', 4));
    expect(result.current.cart[0].total).toBe(5800);

    act(() => result.current.updateQuantity('123', 0));
    expect(result.current.cart).toEqual([]);
  });
});
