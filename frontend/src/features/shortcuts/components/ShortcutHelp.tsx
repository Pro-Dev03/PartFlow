import { useState } from 'react'
import { clsx } from 'clsx'
import type { ShortcutCategory } from '../types/shortcuts'

interface ShortcutHelpProps {
  categories: ShortcutCategory[]
  onClose: () => void
  className?: string
}

export function ShortcutHelp({ categories, onClose, className }: ShortcutHelpProps) {
  const [activeCategory, setActiveCategory] = useState<string | null>(null)

  const formatKeys = (keys: string[]) => {
    return keys.map(key => {
      if (key === 'Ctrl') return 'Ctrl'
      if (key === 'Cmd') return '⌘'
      if (key === 'Shift') return 'Shift'
      if (key === 'Alt') return 'Alt'
      if (key.length === 1) return key.toUpperCase()
      return key
    }).join(' + ')
  }

  return (
    <div className={clsx('bg-surface rounded-lg shadow-xl max-w-2xl w-full mx-4', className)}>
      {/* Header */}
      <div className="flex items-center justify-between p-4 border-b border-border">
        <h2 className="text-lg font-semibold text-text">اختصارات لوحة المفاتيح</h2>
        <button
          onClick={onClose}
          className="text-muted hover:text-text p-2 rounded-lg hover:bg-muted-10"
        >
          ✕
        </button>
      </div>

      {/* Content */}
      <div className="p-4 max-h-96 overflow-y-auto">
        {/* Category Tabs */}
        <div className="flex gap-2 mb-4 flex-wrap">
          <button
            onClick={() => setActiveCategory(null)}
            className={clsx(
              'px-3 py-1 rounded-lg text-sm font-medium transition-colors',
              activeCategory === null
                ? 'bg-primary text-white'
                : 'bg-muted text-muted hover:bg-muted-80'
            )}
          >
            الكل
          </button>
          {categories.map((category) => (
            <button
              key={category.id}
              onClick={() => setActiveCategory(category.id)}
              className={clsx(
                'px-3 py-1 rounded-lg text-sm font-medium transition-colors',
                activeCategory === category.id
                  ? 'bg-primary text-white'
                  : 'bg-muted text-muted hover:bg-muted-80'
              )}
            >
              {category.label}
            </button>
          ))}
        </div>

        {/* Shortcuts List */}
        <div className="space-y-4">
          {(activeCategory
            ? categories.find(c => c.id === activeCategory)?.shortcuts || []
            : categories.flatMap(c => c.shortcuts)
          ).map((shortcut) => (
            <div
              key={shortcut.id}
              className="flex items-center justify-between p-3 bg-muted-10 rounded-lg"
            >
              <div className="flex-1">
                <p className="font-medium text-text">{shortcut.description}</p>
              </div>
              <div className="flex gap-1">
                {shortcut.keys.map((key, index) => (
                  <span
                    key={index}
                    className="px-2 py-1 bg-surface border border-border rounded text-sm font-mono text-text"
                  >
                    {key}
                  </span>
                ))}
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Footer */}
      <div className="p-4 border-t border-border">
        <p className="text-sm text-muted text-center">
          اضغط <kbd className="px-2 py-1 bg-muted rounded text-xs">?</kbd> في أي وقت لعرض هذه القائمة
        </p>
      </div>
    </div>
  )
}
