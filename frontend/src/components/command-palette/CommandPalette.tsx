import { useState, useEffect, useCallback } from 'react'
import { clsx } from 'clsx'

interface Command {
  id: string
  label: string
  icon: string
  shortcut?: string
  category: string
  action: () => void
}

interface CommandPaletteProps {
  commands: Command[]
  onClose: () => void
  className?: string
}

export function CommandPalette({ commands, onClose, className }: CommandPaletteProps) {
  const [isOpen, setIsOpen] = useState(true)
  const [query, setQuery] = useState('')
  const [selectedIndex, setSelectedIndex] = useState(0)

  const filteredCommands = commands.filter(command =>
    command.label.toLowerCase().includes(query.toLowerCase()) ||
    command.category.toLowerCase().includes(query.toLowerCase())
  )

  const groupedCommands = filteredCommands.reduce((acc, command) => {
    if (!acc[command.category]) {
      acc[command.category] = []
    }
    acc[command.category].push(command)
    return acc
  }, {} as Record<string, Command[]>)

  const handleKeyDown = useCallback((e: KeyboardEvent) => {
    if (!isOpen) return

    switch (e.key) {
      case 'ArrowDown':
        e.preventDefault()
        setSelectedIndex(prev => 
          prev < filteredCommands.length - 1 ? prev + 1 : prev
        )
        break
      case 'ArrowUp':
        e.preventDefault()
        setSelectedIndex(prev => prev > 0 ? prev - 1 : 0)
        break
      case 'Enter':
        e.preventDefault()
        if (filteredCommands[selectedIndex]) {
          filteredCommands[selectedIndex].action()
          handleClose()
        }
        break
      case 'Escape':
        e.preventDefault()
        handleClose()
        break
    }
  }, [isOpen, filteredCommands, selectedIndex])

  const handleClose = () => {
    setIsOpen(false)
    setQuery('')
    setSelectedIndex(0)
    onClose()
  }

  useEffect(() => {
    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [handleKeyDown])

  if (!isOpen) return null

  return (
    <div className={clsx('fixed inset-0 z-50 flex items-start justify-center pt-24 px-4', className)}>
      {/* Backdrop */}
      <div
        className="absolute inset-0 bg-black/50 backdrop-blur-sm"
        onClick={handleClose}
      />

      {/* Modal */}
      <div className="relative w-full max-w-2xl bg-surface rounded-xl shadow-2xl border border-border overflow-hidden">
        {/* Search Input */}
        <div className="p-4 border-b border-border">
          <div className="relative">
            <span className="absolute right-3 top-1/2 -translate-y-1/2 text-muted">🔍</span>
            <input
              type="text"
              value={query}
              onChange={(e) => {
                setQuery(e.target.value)
                setSelectedIndex(0)
              }}
              placeholder="ابحث عن أمر أو انتقل..."
              className="w-full pr-10 pl-4 py-3 bg-muted-10 border border-border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent text-lg"
              autoFocus
            />
          </div>
        </div>

        {/* Commands List */}
        <div className="max-h-96 overflow-y-auto p-2">
          {filteredCommands.length === 0 ? (
            <div className="text-center py-8 text-muted">
              لا توجد نتائج
            </div>
          ) : (
            Object.entries(groupedCommands).map(([category, commands]) => (
              <div key={category} className="mb-4">
                <div className="px-3 py-2 text-xs font-medium text-muted uppercase tracking-wider">
                  {category}
                </div>
                {commands.map((command, index) => {
                  const globalIndex = filteredCommands.indexOf(command)
                  return (
                    <button
                      key={command.id}
                      onClick={() => {
                        command.action()
                        handleClose()
                      }}
                      className={clsx(
                        'w-full px-3 py-2 rounded-lg flex items-center gap-3 text-right transition-colors',
                        globalIndex === selectedIndex
                          ? 'bg-primary-10 text-primary'
                          : 'hover:bg-muted-10 text-text'
                      )}
                    >
                      <span className="text-xl">{command.icon}</span>
                      <span className="flex-1 font-medium">{command.label}</span>
                      {command.shortcut && (
                        <span className="text-xs text-muted bg-muted-10 px-2 py-1 rounded">
                          {command.shortcut}
                        </span>
                      )}
                    </button>
                  )
                })}
              </div>
            ))
          )}
        </div>

        {/* Footer */}
        <div className="p-3 border-t border-border bg-muted-10 flex items-center justify-between text-xs text-muted">
          <div className="flex items-center gap-4">
            <span className="flex items-center gap-1">
              <kbd className="px-1.5 py-0.5 bg-surface border border-border rounded">↑↓</kbd>
              للتنقل
            </span>
            <span className="flex items-center gap-1">
              <kbd className="px-1.5 py-0.5 bg-surface border border-border rounded">Enter</kbd>
              للاختيار
            </span>
            <span className="flex items-center gap-1">
              <kbd className="px-1.5 py-0.5 bg-surface border border-border rounded">Esc</kbd>
              للإغلاق
            </span>
          </div>
        </div>
      </div>
    </div>
  )
}
