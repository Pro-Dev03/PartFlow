import { useState, useCallback, useEffect } from 'react'
import { CommandPalette } from './CommandPalette'

interface Command {
  id: string
  label: string
  icon: string
  shortcut?: string
  category: string
  action: () => void
}

export function useCommandPalette() {
  const [isOpen, setIsOpen] = useState(false)
  const [commands, setCommands] = useState<Command[]>([])

  const open = useCallback(() => setIsOpen(true), [])
  const close = useCallback(() => setIsOpen(false), [])

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Ctrl+K or Cmd+K to open
      if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
        e.preventDefault()
        setIsOpen(prev => !prev)
      }
    }

    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [])

  const registerCommands = useCallback((newCommands: Command[]) => {
    setCommands(prev => [...prev, ...newCommands])
  }, [])

  const CommandPaletteComponent = isOpen ? (
    <CommandPalette
      commands={commands}
      onClose={close}
    />
  ) : null

  return {
    isOpen,
    open,
    close,
    registerCommands,
    CommandPaletteComponent,
  }
}
