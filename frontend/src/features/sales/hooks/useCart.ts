import { useState, useCallback } from 'react';
import { CartItem } from '../types/pos.types';
import { playScanSound } from '../../../hooks/useBarcodeContext';

export function normalizePosPrice(...values: unknown[]): number {
  const value = values.find((candidate) => candidate !== undefined && candidate !== null && candidate !== '');
  const numericValue = typeof value === 'number' ? value : Number(value);

  if (!Number.isFinite(numericValue)) {
    return 0;
  }

  return numericValue;
}

export function useCart(soundEnabled: boolean = true) {
  const [cart, setCart] = useState<CartItem[]>([]);

  const addToCart = useCallback((item: any) => {
    const price = normalizePosPrice(item.price, item.sellingPrice, item.selling_price);
    const purchaseCost = normalizePosPrice(item.purchaseCost, item.costPrice, item.cost_price);
    setCart((currentCart) => {
      const existingItem = currentCart.find((c) => c.barcode === item.barcode);
      if (existingItem) {
        return currentCart.map((c) =>
          c.barcode === item.barcode
            ? { ...c, quantity: c.quantity + 1, total: (c.quantity + 1) * c.price }
            : c
        );
      }
      return [...currentCart, {
        id: item.id,
        name: item.name,
        barcode: item.barcode,
        price,
        quantity: 1,
        total: price,
        stock: item.stock,
        isTradeIn: item.isTradeIn || item.condition === 'USED' || false,
        purchaseCost,
        partType: item.partType,
        partTypeColor: item.partTypeColor,
        condition: item.condition,
        grade: item.grade,
      }];
    });
    if (soundEnabled) playScanSound(true);
  }, [soundEnabled]);

  const removeFromCart = useCallback((barcode: string) => {
    setCart((currentCart) => currentCart.filter((item) => item.barcode !== barcode));
  }, []);

  const updateQuantity = useCallback((barcode: string, quantity: number) => {
    if (quantity <= 0) {
      removeFromCart(barcode);
      return;
    }
    setCart((currentCart) => currentCart.map((item) =>
      item.barcode === barcode
        ? { ...item, quantity, total: quantity * item.price }
        : item
    ));
  }, [cart, removeFromCart]);

  const clearCart = useCallback(() => {
    setCart([]);
  }, []);

  const subtotal = cart.reduce((sum, item) => sum + item.total, 0);
  const total = subtotal; // No tax - total equals subtotal

  return {
    cart,
    addToCart,
    removeFromCart,
    updateQuantity,
    clearCart,
    subtotal,
    total,
    cartCount: cart.length,
  };
}