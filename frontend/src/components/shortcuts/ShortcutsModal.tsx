import { useShortcutStore } from './KeyboardShortcuts'
import { clsx } from 'clsx'
import { X, Keyboard } from 'lucide-react'

export function ShortcutsModal() {
  const { shortcuts, isShortcutsModalOpen, closeShortcutsModal } = useShortcutStore()

  if (!isShortcutsModalOpen) return null

  const groupedShortcuts = shortcuts.reduce((acc, shortcut) => {
    const category = shortcut.category || 'عام'
    if (!acc[category]) {
      acc[category] = []
    }
    acc[category].push(shortcut)
    return acc
  }, {} as Record<string, typeof shortcuts>)

  const formatKey = (key: string) => {
    const keyMap: Record<string, string> = {
      'Control': 'Ctrl',
      'Meta': 'Cmd',
      'ArrowUp': '↑',
      'ArrowDown': '↓',
      'ArrowLeft': '←',
      'ArrowRight': '→',
      ' ': 'Space',
    }
    return keyMap[key] || key
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      {/* Backdrop */}
      <div
        className="absolute inset-0 bg-black/50 backdrop-blur-sm"
        onClick={closeShortcutsModal}
      />

      {/* Modal */}
      <div className="relative w-full max-w-2xl bg-surface rounded-xl shadow-2xl border border-border overflow-hidden">
        {/* Header */}
        <div className="flex items-center justify-between p-4 border-b border-border">
          <div className="flex items-center gap-2">
            <Keyboard className="w-5 h-5 text-primary" />
            <h2 className="text-lg font-semibold">اختصارات لوحة المفاتيح</h2>
          </div>
          <button
            onClick={closeShortcutsModal}
            className="p-2 hover:bg-muted-10 rounded-lg transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Content */}
        <div className="max-h-[60vh] overflow-y-auto p-4">
          {Object.entries(groupedShortcuts).map(([category, categoryShortcuts]) => (
            <div key={category} className="mb-6">
              <h3 className="text-sm font-semibold text-muted uppercase tracking-wider mb-3">
                {category}
              </h3>
              <div className="space-y-2">
                {categoryShortcuts.map((shortcut) => (
                  <div
                    key={shortcut.id}
                    className="flex items-center justify-between p-3 bg-muted-10 rounded-lg"
                  >
                    <span className="text-sm">{shortcut.description}</span>
                    <div className="flex items-center gap-1">
                      {shortcut.keys.map((key, index) => (
                        <span
                          key={index}
                          className="px-2 py-1 bg-surface border border-border rounded text-xs font-mono"
                        >
                          {formatKey(key)}
                        </span>
                      ))}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          ))}
        </div>

        {/* Footer */}
        <div className="p-4 border-t border-border bg-muted-10">
          <button
            onClick={closeShortcutsModal}
            className="w-full py-2 bg-primary text-white rounded-lg hover:bg-primary-90 transition-colors"
          >
            إغلاق
          </button>
        </div>
      </div>
    </div>
  )
}