import { useState, useCallback, useEffect } from 'react';
import { CartItem } from '../types/pos.types';
import { playScanSound } from '../../../hooks/useBarcodeContext';

export interface PosCartProduct {
  id: string | number;
  inventoryItemId?: string;
  serialNumber?: string;
  name: string;
  barcode?: string;
  sku?: string;
  price?: number;
  sellingPrice?: number;
  selling_price?: number;
  costPrice?: number;
  cost_price?: number;
  stock?: number;
  isTradeIn?: boolean;
  condition?: string;
  partType?: string;
  partTypeColor?: string;
  grade?: string;
}

export function normalizePosPrice(...values: unknown[]): number {
  const value = values.find((candidate) => candidate !== undefined && candidate !== null && candidate !== '');
  const numericValue = typeof value === 'number' ? value : Number(value);

  if (!Number.isFinite(numericValue)) {
    return 0;
  }

  return numericValue;
}

function getCartItemKey(item: PosCartProduct): string {
  return String(item.inventoryItemId || item.barcode || item.sku || item.id);
}

const POS_CART_STORAGE_KEY = 'partflow-pos-cart';

function loadPersistedCart(): CartItem[] {
  if (typeof window === 'undefined') return [];
  try {
    const parsed = JSON.parse(window.localStorage.getItem(POS_CART_STORAGE_KEY) || '[]');
    if (!Array.isArray(parsed)) return [];
    return parsed.filter((item): item is CartItem => (
      item && typeof item === 'object' && typeof item.id !== 'undefined' &&
      typeof item.name === 'string' && typeof item.barcode === 'string' &&
      Number.isFinite(Number(item.price)) && Number(item.quantity) > 0
    ));
  } catch {
    return [];
  }
}

export function useCart(soundEnabled: boolean = true, taxRate: number = 0) {
  const [cart, setCart] = useState<CartItem[]>(loadPersistedCart);

  useEffect(() => {
    window.localStorage.setItem(POS_CART_STORAGE_KEY, JSON.stringify(cart));
  }, [cart]);

  const addToCart = useCallback((item: PosCartProduct, requestedQuantity = 1) => {
    const quantityToAdd = Math.max(1, requestedQuantity);
    const itemKey = getCartItemKey(item);
    const price = normalizePosPrice(item.price, item.sellingPrice, item.selling_price);
    const purchaseCost = normalizePosPrice(item.purchaseCost, item.costPrice, item.cost_price);
    setCart((currentCart) => {
      const existingItem = currentCart.find((c) =>
        item.inventoryItemId
          ? c.inventoryItemId === item.inventoryItemId
          : c.barcode === itemKey
      );
      if (existingItem) {
        if (item.inventoryItemId) return currentCart;
        return currentCart.map((c) =>
          c.barcode === itemKey
            ? { ...c, quantity: c.quantity + quantityToAdd, total: (c.quantity + quantityToAdd) * c.price }
            : c
        );
      }
      return [...currentCart, {
        id: item.id,
        inventoryItemId: item.inventoryItemId,
        serialNumber: item.serialNumber,
        name: item.name,
        barcode: itemKey,
        price,
        quantity: quantityToAdd,
        total: quantityToAdd * price,
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
        ? { ...item, quantity: item.inventoryItemId ? 1 : quantity, total: (item.inventoryItemId ? 1 : quantity) * item.price }
        : item
    ));
  }, [removeFromCart]);

  const clearCart = useCallback(() => {
    setCart([]);
  }, []);

  // Product prices are stored before tax. POS applies tax to the customer total.
  const subtotal = cart.reduce((sum, item) => sum + item.total, 0);
  const total = subtotal;

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