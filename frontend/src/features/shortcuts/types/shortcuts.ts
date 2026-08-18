export interface KeyboardShortcut {
  id: string
  keys: string[]
  description: string
  category: 'navigation' | 'actions' | 'search' | 'forms' | 'general'
  action: () => void
  enabled?: boolean
}

export interface ShortcutCategory {
  id: string
  label: string
  shortcuts: KeyboardShortcut[]
}
