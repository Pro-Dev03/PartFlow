import { HTMLAttributes } from 'react'
import { clsx } from 'clsx'

interface ErrorStateProps extends HTMLAttributes<HTMLDivElement> {
  icon?: string
  title: string
  description?: string
  actionLabel?: string
  onAction?: () => void
  errorDetails?: string
}

export function ErrorState({
  icon = '⚠️',
  title,
  description,
  actionLabel,
  onAction,
  errorDetails,
  className,
  ...props
}: ErrorStateProps) {
  return (
    <div className={clsx('flex flex-col items-center justify-center py-12 px-4 text-center', className)} {...props}>
      <div className="text-6xl mb-4 text-danger">{icon}</div>
      <h3 className="text-lg font-semibold text-text mb-2">{title}</h3>
      {description && (
        <p className="text-muted max-w-md mb-4">{description}</p>
      )}
      {errorDetails && (
        <div className="bg-danger-50 border border-danger-200 rounded-lg p-4 mb-6 max-w-md text-left">
          <code className="text-xs text-danger-700 break-all">{errorDetails}</code>
        </div>
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