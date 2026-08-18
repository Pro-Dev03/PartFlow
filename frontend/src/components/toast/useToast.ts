import { useState, useCallback } from 'react'
import { ToastProps, ToastType } from './Toast'

interface ToastOptions {
  title: string
  message?: string
  type?: ToastType
  duration?: number
  action?: {
    label: string
    onClick: () => void
  }
}

export function useToast() {
  const [toasts, setToasts] = useState<ToastProps[]>([])

  const addToast = useCallback((options: ToastOptions) => {
    const id = Math.random().toString(36).substr(2, 9)
    const newToast: ToastProps = {
      id,
      type: options.type || 'info',
      title: options.title,
      message: options.message,
      duration: options.duration,
      action: options.action,
      onClose: () => {
        setToasts(prev => prev.filter(t => t.id !== id))
      },
    }

    setToasts(prev => [...prev, newToast])
    return id
  }, [])

  const removeToast = useCallback((id: string) => {
    setToasts(prev => prev.filter(t => t.id !== id))
  }, [])

  const success = useCallback((title: string, message?: string) => {
    return addToast({ title, message, type: 'success' })
  }, [addToast])

  const error = useCallback((title: string, message?: string) => {
    return addToast({ title, message, type: 'error', duration: 7000 })
  }, [addToast])

  const warning = useCallback((title: string, message?: string) => {
    return addToast({ title, message, type: 'warning' })
  }, [addToast])

  const info = useCallback((title: string, message?: string) => {
    return addToast({ title, message, type: 'info' })
  }, [addToast])

  const withAction = useCallback((options: ToastOptions) => {
    return addToast(options)
  }, [addToast])

  return {
    toasts,
    addToast,
    removeToast,
    success,
    error,
    warning,
    info,
    withAction,
  }
}
