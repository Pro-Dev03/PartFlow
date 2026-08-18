import { useState, useCallback } from 'react'
import { ConfirmDialog } from './ConfirmDialog'

interface ConfirmOptions {
  title: string
  message: string
  confirmLabel?: string
  cancelLabel?: string
  variant?: 'danger' | 'warning' | 'info'
  type?: 'delete' | 'action' | 'financial'
}

interface ConfirmState extends ConfirmOptions {
  resolve: (result: boolean) => void
}

export function useConfirm() {
  const [confirmState, setConfirmState] = useState<ConfirmState | null>(null)

  const confirm = useCallback((options: ConfirmOptions): Promise<boolean> => {
    return new Promise((resolve) => {
      setConfirmState({
        ...options,
        resolve,
      })
    })
  }, [])

  const handleClose = useCallback(() => {
    if (confirmState) {
      confirmState.resolve(false)
      setConfirmState(null)
    }
  }, [confirmState])

  const handleConfirm = useCallback(() => {
    if (confirmState) {
      confirmState.resolve(true)
      setConfirmState(null)
    }
  }, [confirmState])

  const ConfirmDialogComponent = confirmState ? (
    <ConfirmDialog
      isOpen={!!confirmState}
      onClose={handleClose}
      onConfirm={handleConfirm}
      title={confirmState.title}
      message={confirmState.message}
      confirmLabel={confirmState.confirmLabel}
      cancelLabel={confirmState.cancelLabel}
      variant={confirmState.variant}
      type={confirmState.type}
    />
  ) : null

  return {
    confirm,
    ConfirmDialogComponent,
  }
}
