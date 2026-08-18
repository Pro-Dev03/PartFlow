import { create } from 'zustand';

export interface CartItem {
  id: string;
  productId: string;
  productName: string;
  quantity: number;
  unitPrice: number;
  total: number;
  barcode?: string;
}

interface CartState {
  items: CartItem[];
  subtotal: number;
  discount: number;
  tax: number;
  total: number;
  addItem: (item: Omit<CartItem, 'total'>) => void;
  removeItem: (id: string) => void;
  updateQuantity: (id: string, quantity: number) => void;
  clearCart: () => void;
  setDiscount: (amount: number) => void;
  setTax: (amount: number) => void;
}

export const useCartStore = create<CartState>((set, get) => ({
  items: [],
  subtotal: 0,
  discount: 0,
  tax: 0,
  total: 0,

  addItem: (item) => {
    const items = get().items;
    const existingItem = items.find((i) => i.productId === item.productId);

    let newItems: CartItem[];
    if (existingItem) {
      newItems = items.map((i) =>
        i.productId === item.productId
          ? { ...i, quantity: i.quantity + item.quantity, total: (i.quantity + item.quantity) * i.unitPrice }
          : i
      );
    } else {
      newItems = [...items, { ...item, total: item.quantity * item.unitPrice }];
    }

    const subtotal = newItems.reduce((sum, item) => sum + item.total, 0);
    const discount = get().discount;
    const tax = get().tax;
    const total = subtotal - discount + tax;

    set({ items: newItems, subtotal, total });
  },

  removeItem: (id) => {
    const newItems = get().items.filter((item) => item.id !== id);
    const subtotal = newItems.reduce((sum, item) => sum + item.total, 0);
    const discount = get().discount;
    const tax = get().tax;
    const total = subtotal - discount + tax;

    set({ items: newItems, subtotal, total });
  },

  updateQuantity: (id, quantity) => {
    if (quantity <= 0) {
      get().removeItem(id);
      return;
    }

    const newItems = get().items.map((item) =>
      item.id === id ? { ...item, quantity, total: quantity * item.unitPrice } : item
    );
    const subtotal = newItems.reduce((sum, item) => sum + item.total, 0);
    const discount = get().discount;
    const tax = get().tax;
    const total = subtotal - discount + tax;

    set({ items: newItems, subtotal, total });
  },

  clearCart: () => {
    set({ items: [], subtotal: 0, discount: 0, tax: 0, total: 0 });
  },

  setDiscount: (amount) => {
    const subtotal = get().subtotal;
    const tax = get().tax;
    const total = subtotal - amount + tax;
    set({ discount: amount, total });
  },

  setTax: (amount) => {
    const subtotal = get().subtotal;
    const discount = get().discount;
    const total = subtotal - discount + amount;
    set({ tax: amount, total });
  },
}));