import { HTMLAttributes } from 'react'
import { clsx } from 'clsx'

interface EmptyStateProps extends HTMLAttributes<HTMLDivElement> {
  icon?: string
  title: string
  description?: string
  actionLabel?: string
  onAction?: () => void
  variant?: 'default' | 'success' | 'warning' | 'info'
}

export function EmptyState({
  icon = '📭',
  title,
  description,
  actionLabel,
  onAction,
  variant = 'default',
  className,
  ...props
}: EmptyStateProps) {
  const variantStyles = {
    default: 'text-muted',
    success: 'text-success',
    warning: 'text-warning',
    info: 'text-primary',
  }

  return (
    <div className={clsx('flex flex-col items-center justify-center py-12 px-4 text-center', className)} {...props}>
      <div className={`text-6xl mb-4 ${variantStyles[variant]}`}>{icon}</div>
      <h3 className="text-lg font-semibold text-text mb-2">{title}</h3>
      {description && (
        <p className="text-muted max-w-md mb-6">{description}</p>
      )}
      {actionLabel && onAction && (
        <button
          onClick={onAction}
          className="inline-flex items-center gap-2 px-4 py-2 bg-primary text-white rounded-lg hover:bg-primary-600 transition-colors"
        >
          {actionLabel}
        </button>
      )}
    </div>
  )
}