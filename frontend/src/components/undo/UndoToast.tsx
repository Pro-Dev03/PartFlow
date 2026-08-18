import { clsx } from 'clsx'
import type { UndoAction } from './useUndo'

interface UndoToastProps {
  action: UndoAction
  onUndo?: () => void
  onDismiss?: () => void
  className?: string
}

export function UndoToast({ action, onUndo, onDismiss, className }: UndoToastProps) {
  const handleUndo = async () => {
    await action.undo()
    onUndo?.()
  }

  const handleDismiss = () => {
    onDismiss?.()
  }

  return (
    <div className={clsx('flex items-center gap-3 p-3 bg-surface border border-border rounded-lg shadow-lg', className)}>
      <span className="text-sm text-muted">{action.description}</span>
      <button
        onClick={handleUndo}
        className="px-3 py-1 text-sm font-medium text-primary hover:text-primary-80 transition-colors"
      >
        تراجع
      </button>
      <button
        onClick={handleDismiss}
        className="text-muted hover:text-text transition-colors"
        aria-label="إغلاق"
      >
        ✕
      </button>
    </div>
  )
}
