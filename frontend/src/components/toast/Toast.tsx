import { clsx } from 'clsx'
import { useEffect, useState } from 'react'

export type ToastType = 'success' | 'error' | 'warning' | 'info'

export interface ToastProps {
  id: string
  type: ToastType
  title: string
  message?: string
  duration?: number
  action?: {
    label: string
    onClick: () => void
  }
  onClose: () => void
}

export function Toast({ type, title, message, duration = 5000, action, onClose }: ToastProps) {
  const [isVisible, setIsVisible] = useState(true)

  useEffect(() => {
    const timer = setTimeout(() => {
      setIsVisible(false)
      setTimeout(onClose, 300) // Wait for exit animation
    }, duration)

    return () => clearTimeout(timer)
  }, [duration, onClose])

  const typeConfig = {
    success: {
      icon: '✓',
      bgColor: 'bg-success-10',
      borderColor: 'border-success-30',
      iconColor: 'text-success',
    },
    error: {
      icon: '✕',
      bgColor: 'bg-danger-10',
      borderColor: 'border-danger-30',
      iconColor: 'text-danger',
    },
    warning: {
      icon: '⚠',
      bgColor: 'bg-warning-10',
      borderColor: 'border-warning-30',
      iconColor: 'text-warning',
    },
    info: {
      icon: 'ℹ',
      bgColor: 'bg-info-10',
      borderColor: 'border-info-30',
      iconColor: 'text-info',
    },
  }

  const config = typeConfig[type]

  return (
    <div
      className={clsx(
        'flex items-start gap-3 p-4 rounded-lg border shadow-lg min-w-[300px] max-w-md',
        config.bgColor,
        config.borderColor,
        'transform transition-all duration-300',
        isVisible ? 'translate-x-0 opacity-100' : 'translate-x-full opacity-0'
      )}
    >
      {/* Icon */}
      <div className={clsx('flex-shrink-0 w-6 h-6 rounded-full flex items-center justify-center text-sm font-bold', config.iconColor, 'bg-current bg-opacity-10')}>
        {config.icon}
      </div>

      {/* Content */}
      <div className="flex-1 min-w-0">
        <h4 className="font-semibold text-text text-sm">{title}</h4>
        {message && (
          <p className="text-sm text-muted mt-1 break-words">{message}</p>
        )}
        {action && (
          <button
            onClick={action.onClick}
            className="mt-2 text-sm font-medium text-primary hover:text-primary-80 transition-colors"
          >
            {action.label}
          </button>
        )}
      </div>

      {/* Close Button */}
      <button
        onClick={() => {
          setIsVisible(false)
          setTimeout(onClose, 300)
        }}
        className="flex-shrink-0 text-muted hover:text-text transition-colors p-1"
        aria-label="إغلاق"
      >
        ✕
      </button>
    </div>
  )
}
