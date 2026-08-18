import { create } from 'zustand'
import { persist } from 'zustand/middleware'

// UI State Store
interface UIState {
  // Theme
  theme: 'light' | 'dark' | 'system'
  setTheme: (theme: 'light' | 'dark' | 'system') => void

  // Language
  language: 'ar' | 'he' | 'en'
  setLanguage: (language: 'ar' | 'he' | 'en') => void

  // Sidebar
  sidebarCollapsed: boolean
  toggleSidebar: () => void
  setSidebarCollapsed: (collapsed: boolean) => void

  // Modals
  activeModal: string | null
  openModal: (modalId: string) => void
  closeModal: () => void

  // Notifications
  notifications: Notification[]
  addNotification: (notification: Omit<Notification, 'id'>) => void
  removeNotification: (id: string) => void
  clearNotifications: () => void

  // Loading states
  isLoading: boolean
  setLoading: (loading: boolean) => void

  // Filters
  activeFilters: Record<string, any>
  setFilter: (key: string, value: any) => void
  clearFilters: () => void

  // Selection
  selectedItems: string[]
  setSelectedItems: (items: string[]) => void
  toggleSelectedItem: (item: string) => void
  clearSelection: () => void
}

interface Notification {
  id: string
  type: 'success' | 'error' | 'warning' | 'info'
  title: string
  message: string
  duration?: number
}

export const useUIStore = create<UIState>()(
  persist(
    (set, get) => ({
      // Theme
      theme: 'system',
      setTheme: (theme) => set({ theme }),

      // Language
      language: 'ar',
      setLanguage: (language) => set({ language }),

      // Sidebar
      sidebarCollapsed: false,
      toggleSidebar: () => set((state) => ({ sidebarCollapsed: !state.sidebarCollapsed })),
      setSidebarCollapsed: (collapsed) => set({ sidebarCollapsed: collapsed }),

      // Modals
      activeModal: null,
      openModal: (modalId) => set({ activeModal: modalId }),
      closeModal: () => set({ activeModal: null }),

      // Notifications
      notifications: [],
      addNotification: (notification) => {
        const id = Date.now().toString()
        const newNotification = { ...notification, id }
        set((state) => ({
          notifications: [...state.notifications, newNotification],
        }))

        // Auto-remove notification after duration
        if (notification.duration !== 0) {
          setTimeout(() => {
            get().removeNotification(id)
          }, notification.duration || 5000)
        }
      },
      removeNotification: (id) =>
        set((state) => ({
          notifications: state.notifications.filter((n) => n.id !== id),
        })),
      clearNotifications: () => set({ notifications: [] }),

      // Loading
      isLoading: false,
      setLoading: (loading) => set({ isLoading: loading }),

      // Filters
      activeFilters: {},
      setFilter: (key, value) =>
        set((state) => ({
          activeFilters: { ...state.activeFilters, [key]: value },
        })),
      clearFilters: () => set({ activeFilters: {} }),

      // Selection
      selectedItems: [],
      setSelectedItems: (items) => set({ selectedItems: items }),
      toggleSelectedItem: (item) =>
        set((state) => ({
          selectedItems: state.selectedItems.includes(item)
            ? state.selectedItems.filter((i) => i !== item)
            : [...state.selectedItems, item],
        })),
      clearSelection: () => set({ selectedItems: [] }),
    }),
    {
      name: 'partflow-ui-storage',
      partialize: (state) => ({
        theme: state.theme,
        language: state.language,
        sidebarCollapsed: state.sidebarCollapsed,
      }),
    }
  )
)