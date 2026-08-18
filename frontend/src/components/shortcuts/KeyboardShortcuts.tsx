import { useEffect, useCallback } from 'react'
import { create } from 'zustand'

export interface Shortcut {
  id: string
  keys: string[]
  description: string
  action: () => void
  category?: string
  disabled?: boolean
}

interface ShortcutStore {
  shortcuts: Shortcut[]
  registerShortcut: (shortcut: Shortcut) => void
  unregisterShortcut: (id: string) => void
  isShortcutsModalOpen: boolean
  openShortcutsModal: () => void
  closeShortcutsModal: () => void
}

export const useShortcutStore = create<ShortcutStore>((set) => ({
  shortcuts: [],
  registerShortcut: (shortcut) =>
    set((state) => ({
      shortcuts: [...state.shortcuts.filter((s) => s.id !== shortcut.id), shortcut],
    })),
  unregisterShortcut: (id) =>
    set((state) => ({
      shortcuts: state.shortcuts.filter((s) => s.id !== id),
    })),
  isShortcutsModalOpen: false,
  openShortcutsModal: () => set({ isShortcutsModalOpen: true }),
  closeShortcutsModal: () => set({ isShortcutsModalOpen: false }),
}))

export function useKeyboardShortcuts() {
  const { shortcuts } = useShortcutStore()

  const handleKeyDown = useCallback(
    (e: KeyboardEvent) => {
      const pressedKeys: string[] = []

      if (e.ctrlKey || e.metaKey) pressedKeys.push('Ctrl')
      if (e.altKey) pressedKeys.push('Alt')
      if (e.shiftKey) pressedKeys.push('Shift')
      if (e.key !== 'Control' && e.key !== 'Alt' && e.key !== 'Shift' && e.key !== 'Meta') {
        pressedKeys.push(e.key)
      }

      const keyCombo = pressedKeys.join('+')

      for (const shortcut of shortcuts) {
        if (shortcut.disabled) continue

        const shortcutCombo = shortcut.keys.join('+')
        if (keyCombo.toLowerCase() === shortcutCombo.toLowerCase()) {
          e.preventDefault()
          shortcut.action()
          break
        }
      }
    },
    [shortcuts]
  )

  useEffect(() => {
    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [handleKeyDown])
}

export function useRegisterShortcut(shortcut: Shortcut) {
  const { registerShortcut, unregisterShortcut } = useShortcutStore()

  useEffect(() => {
    registerShortcut(shortcut)
    return () => unregisterShortcut(shortcut.id)
  }, [shortcut.id, registerShortcut, unregisterShortcut])
}

// Default shortcuts for the application
export const defaultShortcuts: Shortcut[] = [
  {
    id: 'search',
    keys: ['Ctrl', 'K'],
    description: 'البحث العام',
    action: () => console.log('Search'),
    category: 'التنقل',
  },
  {
    id: 'new-sale',
    keys: ['F2'],
    description: 'بيع جديد',
    action: () => console.log('New Sale'),
    category: 'المبيعات',
  },
  {
    id: 'scan',
    keys: ['F3'],
    description: 'مسح باركود',
    action: () => console.log('Scan'),
    category: 'المخزون',
  },
  {
    id: 'customer',
    keys: ['F4'],
    description: 'إضافة عميل',
    action: () => console.log('Add Customer'),
    category: 'العملاء',
  },
  {
    id: 'refresh',
    keys: ['F5'],
    description: 'تحديث',
    action: () => window.location.reload(),
    category: 'عام',
  },
  {
    id: 'help',
    keys: ['F1'],
    description: 'المساعدة',
    action: () => console.log('Help'),
    category: 'عام',
  },
  {
    id: 'close',
    keys: ['Escape'],
    description: 'إغلاق',
    action: () => console.log('Close'),
    category: 'عام',
  },
  {
    id: 'save',
    keys: ['Ctrl', 'S'],
    description: 'حفظ',
    action: () => console.log('Save'),
    category: 'عام',
  },
  {
    id: 'dashboard',
    keys: ['Ctrl', 'D'],
    description: 'الرئيسية',
    action: () => console.log('Dashboard'),
    category: 'التنقل',
  },
  {
    id: 'inventory',
    keys: ['Ctrl', 'I'],
    description: 'المخزون',
    action: () => console.log('Inventory'),
    category: 'التنقل',
  },
]