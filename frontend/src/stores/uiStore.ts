import { create } from 'zustand'
import { persist } from 'zustand/middleware'

// ==================== TYPES ====================

export type Theme = 'light' | 'dark' | 'system'
export type Language = 'ar' | 'he' | 'en'
export type NotificationType = 'success' | 'error' | 'warning' | 'info'

export interface Notification {
  id: string
  type: NotificationType
  title: string
  message: string
  duration?: number
}

// ==================== UI STATE ====================

interface UIState {
  // Theme & Language
  theme: Theme
  language: Language
  setTheme: (theme: Theme) => void
  setLanguage: (language: Language) => void

  // Sidebar
  sidebarCollapsed: boolean
  toggleSidebar: () => void
  setSidebarCollapsed: (collapsed: boolean) => void

  // Modals
  activeModal: string | null
  modalData: any
  openModal: (modalId: string, data?: any) => void
  closeModal: () => void

  // Notifications
  notifications: Notification[]
  addNotification: (notification: Omit<Notification, 'id'>) => void
  removeNotification: (id: string) => void
  clearNotifications: () => void

  // Loading
  isLoading: boolean
  loadingMessage?: string
  setLoading: (loading: boolean, message?: string) => void

  // Filters
  activeFilters: Record<string, any>
  setFilter: (key: string, value: any) => void
  setFilters: (filters: Record<string, any>) => void
  clearFilters: () => void

  // Selection
  selectedItems: string[]
  setSelectedItems: (items: string[]) => void
  toggleSelectedItem: (item: string) => void
  clearSelection: () => void
}

// ==================== STORE ====================

export const useUIStore = create<UIState>()(
  persist(
    (set, get) => ({
      // Theme & Language
      theme: 'system',
      language: 'ar',
      setTheme: (theme) => set({ theme }),
      setLanguage: (language) => set({ language }),

      // Sidebar
      sidebarCollapsed: false,
      toggleSidebar: () => set((state) => ({ sidebarCollapsed: !state.sidebarCollapsed })),
      setSidebarCollapsed: (collapsed) => set({ sidebarCollapsed: collapsed }),

      // Modals
      activeModal: null,
      modalData: null,
      openModal: (modalId, data) => set({ activeModal: modalId, modalData: data }),
      closeModal: () => set({ activeModal: null, modalData: null }),

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
      loadingMessage: undefined,
      setLoading: (loading, message) => set({ isLoading: loading, loadingMessage: message }),

      // Filters
      activeFilters: {},
      setFilter: (key, value) =>
        set((state) => ({
          activeFilters: { ...state.activeFilters, [key]: value },
        })),
      setFilters: (filters) => set({ activeFilters: filters }),
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