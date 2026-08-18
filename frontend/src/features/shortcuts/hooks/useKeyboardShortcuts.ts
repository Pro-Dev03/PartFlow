import { useEffect, useCallback } from 'react'
import type { KeyboardShortcut } from '../types/shortcuts'

export function useKeyboardShortcuts(shortcuts: KeyboardShortcut[]) {
  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      for (const shortcut of shortcuts) {
        if (!shortcut.enabled) continue

        const isMatch = shortcut.keys.every(key => {
          if (key === 'Ctrl' || key === 'Cmd') {
            return event.ctrlKey || event.metaKey
          }
          if (key === 'Shift') {
            return event.shiftKey
          }
          if (key === 'Alt') {
            return event.altKey
          }
          if (key.length === 1) {
            return event.key.toLowerCase() === key.toLowerCase()
          }
          return false
        })

        if (isMatch) {
          event.preventDefault()
          shortcut.action()
          break
        }
      }
    }

    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [shortcuts])
}

export function useShortcut(keys: string[], action: () => void, enabled = true) {
  useEffect(() => {
    if (!enabled) return

    const handleKeyDown = (event: KeyboardEvent) => {
      const isMatch = keys.every(key => {
        if (key === 'Ctrl' || key === 'Cmd') {
          return event.ctrlKey || event.metaKey
        }
        if (key === 'Shift') {
          return event.shiftKey
        }
        if (key === 'Alt') {
          return event.altKey
        }
        if (key.length === 1) {
          return event.key.toLowerCase() === key.toLowerCase()
        }
        return false
      })

      if (isMatch) {
        event.preventDefault()
        action()
      }
    }

    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [keys, action, enabled])
}
