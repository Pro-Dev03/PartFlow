import { ReactNode } from 'react'
import { clsx } from 'clsx'

interface FormFieldProps {
  label?: string
  error?: string
  required?: boolean
  children: ReactNode
  className?: string
}

export function FormField({ label, error, required, children, className }: FormFieldProps) {
  return (
    <div className={clsx('space-y-1', className)}>
      {label && (
        <label className="block text-sm font-medium text-text">
          {label}
          {required && <span className="text-danger ml-1">*</span>}
        </label>
      )}
      {children}
      {error && (
        <p className="text-sm text-danger">{error}</p>
      )}
    </div>
  )
}

interface FormSectionProps {
  title?: string
  description?: string
  children: ReactNode
  className?: string
}

export function FormSection({ title, description, children, className }: FormSectionProps) {
  return (
    <div className={clsx('space-y-4', className)}>
      {title && (
        <div>
          <h3 className="text-lg font-semibold text-text">{title}</h3>
          {description && (
            <p className="text-sm text-muted mt-1">{description}</p>
          )}
        </div>
      )}
      {children}
    </div>
  )
}

interface FormActionsProps {
  children: ReactNode
  className?: string
  align?: 'left' | 'center' | 'right'
}

export function FormActions({ children, className, align = 'right' }: FormActionsProps) {
  const alignClasses = {
    left: 'justify-start',
    center: 'justify-center',
    right: 'justify-end',
  }

  return (
    <div className={clsx('flex gap-3 pt-4 border-t border-border', alignClasses[align], className)}>
      {children}
    </div>
  )
}