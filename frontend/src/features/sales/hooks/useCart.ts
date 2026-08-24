import { useState, useCallback } from 'react';
import { CartItem } from '../types/pos.types';
import { playScanSound } from '../../../hooks/useBarcodeContext';

export function useCart(soundEnabled: boolean = true) {
  const [cart, setCart] = useState<CartItem[]>([]);

  const addToCart = useCallback((item: any) => {
    const existingItem = cart.find((c) => c.barcode === item.barcode);
    if (existingItem) {
      setCart(cart.map((c) => 
        c.barcode === item.barcode 
          ? { ...c, quantity: c.quantity + 1, total: (c.quantity + 1) * c.price }
          : c
      ));
      // Play sound for quantity increase
      if (soundEnabled) {
        playScanSound(true);
      }
    } else {
      setCart([...cart, {
        id: item.id,
        name: item.name,
        barcode: item.barcode,
        price: item.price,
        quantity: 1,
        total: item.price,
        stock: item.stock,
        isTradeIn: item.isTradeIn || item.condition === 'USED' || false,
        purchaseCost: item.purchaseCost || 0,
        partType: item.partType,
        partTypeColor: item.partTypeColor,
        condition: item.condition,
        grade: item.grade,
      }]);
      // Play sound for new item
      if (soundEnabled) {
        playScanSound(true);
      }
    }
  }, [cart, soundEnabled]);

  const removeFromCart = useCallback((barcode: string) => {
    setCart(cart.filter((item) => item.barcode !== barcode));
  }, [cart]);

  const updateQuantity = useCallback((barcode: string, quantity: number) => {
    if (quantity <= 0) {
      removeFromCart(barcode);
      return;
    }
    setCart(cart.map((item) =>
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