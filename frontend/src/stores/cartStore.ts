import { create } from 'zustand'
import { persist } from 'zustand/middleware'

interface CartItem {
  id: string
  productId: string
  name: string
  price: number
  quantity: number
  barcode?: string
  serialNumber?: string
  condition?: 'new' | 'used'
  discount?: number
}

interface CartState {
  items: CartItem[]
  customerId?: string
  customerName?: string
  notes?: string
  
  // Actions
  addItem: (item: Omit<CartItem, 'id'>) => void
  updateItem: (id: string, updates: Partial<CartItem>) => void
  removeItem: (id: string) => void
  clearCart: () => void
  
  // Customer
  setCustomer: (customerId: string, customerName: string) => void
  clearCustomer: () => void
  
  // Notes
  setNotes: (notes: string) => void
  
  // Computed
  getSubtotal: () => number
  getTotalDiscount: () => number
  getTotal: () => number
  getItemCount: () => number
}

export const useCartStore = create<CartState>()(
  persist(
    (set, get) => ({
      items: [],
      customerId: undefined,
      customerName: undefined,
      notes: undefined,

      addItem: (item) => {
        const id = Date.now().toString()
        set((state) => ({
          items: [...state.items, { ...item, id }],
        }))
      },

      updateItem: (id, updates) => {
        set((state) => ({
          items: state.items.map((item) =>
            item.id === id ? { ...item, ...updates } : item
          ),
        }))
      },

      removeItem: (id) => {
        set((state) => ({
          items: state.items.filter((item) => item.id !== id),
        }))
      },

      clearCart: () => {
        set({
          items: [],
          customerId: undefined,
          customerName: undefined,
          notes: undefined,
        })
      },

      setCustomer: (customerId, customerName) => {
        set({ customerId, customerName })
      },

      clearCustomer: () => {
        set({ customerId: undefined, customerName: undefined })
      },

      setNotes: (notes) => {
        set({ notes })
      },

      getSubtotal: () => {
        return get().items.reduce(
          (total, item) => total + item.price * item.quantity,
          0
        )
      },

      getTotalDiscount: () => {
        return get().items.reduce(
          (total, item) => total + (item.discount || 0) * item.quantity,
          0
        )
      },

      getTotal: () => {
        return get().getSubtotal() - get().getTotalDiscount()
      },

      getItemCount: () => {
        return get().items.reduce((count, item) => count + item.quantity, 0)
      },
    }),
    {
      name: 'partflow-cart-storage',
    }
  )
)