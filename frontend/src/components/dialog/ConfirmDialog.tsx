import { useEffect } from 'react'
import { clsx } from 'clsx'

interface ConfirmDialogProps {
  isOpen: boolean
  onClose: () => void
  onConfirm: () => void
  title: string
  message: string
  confirmLabel?: string
  cancelLabel?: string
  variant?: 'danger' | 'warning' | 'info'
  type?: 'delete' | 'action' | 'financial'
}

export function ConfirmDialog({
  isOpen,
  onClose,
  onConfirm,
  title,
  message,
  confirmLabel = 'تأكيد',
  cancelLabel = 'إلغاء',
  variant = 'danger',
  type = 'action',
}: ConfirmDialogProps) {
  useEffect(() => {
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && isOpen) {
        onClose()
      }
    }

    window.addEventListener('keydown', handleEscape)
    return () => window.removeEventListener('keydown', handleEscape)
  }, [isOpen, onClose])

  if (!isOpen) return null

  const variantStyles = {
    danger: {
      confirmBg: 'bg-danger',
      confirmHover: 'hover:bg-danger-600',
      icon: '⚠️',
    },
    warning: {
      confirmBg: 'bg-warning',
      confirmHover: 'hover:bg-warning-600',
      icon: '⚠️',
    },
    info: {
      confirmBg: 'bg-primary',
      confirmHover: 'hover:bg-primary-600',
      icon: 'ℹ️',
    },
  }

  const styles = variantStyles[variant]

  const handleConfirm = () => {
    onConfirm()
    onClose()
  }

  return (
    <>
      {/* Backdrop */}
      <div
        className="fixed inset-0 bg-black/50 backdrop-blur-sm z-40 transition-opacity"
        onClick={onClose}
      />

      {/* Dialog */}
      <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div className="bg-surface rounded-xl shadow-2xl max-w-md w-full border border-border">
          {/* Header */}
          <div className="p-6 border-b border-border">
            <div className="flex items-center gap-3">
              <span className="text-2xl">{styles.icon}</span>
              <h3 className="text-lg font-semibold text-text">{title}</h3>
            </div>
          </div>

          {/* Content */}
          <div className="p-6">
            <p className="text-muted">{message}</p>

            {/* Type-specific warnings */}
            {type === 'delete' && (
              <div className="mt-4 p-3 bg-danger-10 border border-danger-30 rounded-lg">
                <p className="text-sm text-danger">
                  ⚠️ هذه العملية لا يمكن التراجع عنها
                </p>
              </div>
            )}

            {type === 'financial' && (
              <div className="mt-4 p-3 bg-warning-10 border border-warning-30 rounded-lg">
                <p className="text-sm text-warning">
                  💰 تأكد من صحة المبلغ قبل التأكيد
                </p>
              </div>
            )}
          </div>

          {/* Actions */}
          <div className="p-6 border-t border-border flex gap-3 justify-end">
            <button
              onClick={onClose}
              className="px-4 py-2 border border-border rounded-lg hover:bg-surface-80 transition-colors"
            >
              {cancelLabel}
            </button>
            <button
              onClick={handleConfirm}
              className={clsx(
                'px-4 py-2 text-white rounded-lg transition-colors',
                styles.confirmBg,
                styles.confirmHover
              )}
            >
              {confirmLabel}
            </button>
          </div>
        </div>
      </div>
    </>
  )
}
