import { create } from 'zustand'
import { persist } from 'zustand/middleware'

// UI State
interface UIState {
  sidebarOpen: boolean
  theme: 'light' | 'dark' | 'system'
  language: 'ar' | 'he' | 'en'
  setSidebarOpen: (open: boolean) => void
  setTheme: (theme: 'light' | 'dark' | 'system') => void
  setLanguage: (language: 'ar' | 'he' | 'en') => void
}

export const useUIStore = create<UIState>()(
  persist(
    (set) => ({
      sidebarOpen: true,
      theme: 'system',
      language: 'ar',
      setSidebarOpen: (open) => set({ sidebarOpen: open }),
      setTheme: (theme) => set({ theme }),
      setLanguage: (language) => set({ language }),
    }),
    {
      name: 'ui-storage',
    }
  )
)

// Notification State
interface NotificationState {
  notifications: Array<{
    id: string
    type: 'success' | 'error' | 'warning' | 'info'
    title: string
    message: string
    duration?: number
  }>
  addNotification: (notification: Omit<NotificationState['notifications'][0], 'id'>) => void
  removeNotification: (id: string) => void
  clearNotifications: () => void
}

export const useNotificationStore = create<NotificationState>((set) => ({
  notifications: [],
  addNotification: (notification) => {
    const id = Date.now().toString()
    set((state) => ({
      notifications: [...state.notifications, { ...notification, id }],
    }))
    // Auto-remove after duration (default 5 seconds)
    const duration = notification.duration || 5000
    setTimeout(() => {
      set((state) => ({
        notifications: state.notifications.filter((n) => n.id !== id),
      }))
    }, duration)
  },
  removeNotification: (id) =>
    set((state) => ({
      notifications: state.notifications.filter((n) => n.id !== id),
    })),
  clearNotifications: () => set({ notifications: [] }),
}))

// Modal State
interface ModalState {
  activeModal: string | null
  modalData: any
  openModal: (modalId: string, data?: any) => void
  closeModal: () => void
}

export const useModalStore = create<ModalState>((set) => ({
  activeModal: null,
  modalData: null,
  openModal: (modalId, data) => set({ activeModal: modalId, modalData: data }),
  closeModal: () => set({ activeModal: null, modalData: null }),
}))

// Loading State
interface LoadingState {
  isLoading: boolean
  loadingMessage?: string
  setLoading: (loading: boolean, message?: string) => void
}

export const useLoadingStore = create<LoadingState>((set) => ({
  isLoading: false,
  loadingMessage: undefined,
  setLoading: (loading, message) => set({ isLoading: loading, loadingMessage: message }),
}))

// Filter State (shared across pages)
interface FilterState {
  filters: Record<string, any>
  setFilter: (key: string, value: any) => void
  setFilters: (filters: Record<string, any>) => void
  clearFilters: () => void
}

export const useFilterStore = create<FilterState>((set) => ({
  filters: {},
  setFilter: (key, value) =>
    set((state) => ({
      filters: { ...state.filters, [key]: value },
    })),
  setFilters: (filters) => set({ filters }),
  clearFilters: () => set({ filters: {} }),
}))

// Cart State (for sales)
interface CartItem {
  productId: string
  productName: string
  quantity: number
  price: number
  total: number
}

interface CartState {
  items: CartItem[]
  customerId?: string
  addToCart: (item: CartItem) => void
  removeFromCart: (productId: string) => void
  updateQuantity: (productId: string, quantity: number) => void
  clearCart: () => void
  setCustomerId: (customerId: string) => void
  getTotal: () => number
  getItemCount: () => number
}

export const useCartStore = create<CartState>((set, get) => ({
  items: [],
  customerId: undefined,
  
  addToCart: (item) => {
    set((state) => {
      const existingItem = state.items.find((i) => i.productId === item.productId)
      if (existingItem) {
        return {
          items: state.items.map((i) =>
            i.productId === item.productId
              ? { ...i, quantity: i.quantity + item.quantity, total: (i.quantity + item.quantity) * i.price }
              : i
          ),
        }
      }
      return { items: [...state.items, item] }
    })
  },
  
  removeFromCart: (productId) =>
    set((state) => ({
      items: state.items.filter((i) => i.productId !== productId),
    })),
  
  updateQuantity: (productId, quantity) => {
    if (quantity <= 0) {
      get().removeFromCart(productId)
      return
    }
    set((state) => ({
      items: state.items.map((i) =>
        i.productId === productId
          ? { ...i, quantity, total: quantity * i.price }
          : i
      ),
    }))
  },
  
  clearCart: () => set({ items: [], customerId: undefined }),
  
  setCustomerId: (customerId) => set({ customerId }),
  
  getTotal: () => {
    return get().items.reduce((total, item) => total + item.total, 0)
  },
  
  getItemCount: () => {
    return get().items.reduce((count, item) => count + item.quantity, 0)
  },
}))

// Combined store types for TypeScript
export type AppStore = {
  ui: UIState
  notifications: NotificationState
  modal: ModalState
  loading: LoadingState
  filters: FilterState
  cart: CartState
}
