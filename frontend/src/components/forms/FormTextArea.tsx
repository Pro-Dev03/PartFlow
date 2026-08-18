import { clsx } from 'clsx'

interface FormTextAreaProps {
  label?: string
  name: string
  value: string
  onChange: (value: string) => void
  onBlur?: () => void
  error?: string
  touched?: boolean
  placeholder?: string
  required?: boolean
  disabled?: boolean
  rows?: number
  maxLength?: number
  className?: string
}

export function FormTextArea({
  label,
  name,
  value,
  onChange,
  onBlur,
  error,
  touched,
  placeholder,
  required = false,
  disabled = false,
  rows = 4,
  maxLength,
  className,
}: FormTextAreaProps) {
  const hasError = touched && error
  const currentLength = value.length

  return (
    <div className={clsx('space-y-1', className)}>
      {label && (
        <label htmlFor={name} className="block text-sm font-medium text-text">
          {label}
          {required && <span className="text-danger ml-1">*</span>}
        </label>
      )}
      <textarea
        id={name}
        name={name}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        onBlur={onBlur}
        placeholder={placeholder}
        disabled={disabled}
        required={required}
        rows={rows}
        maxLength={maxLength}
        className={clsx(
          'w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent transition-colors resize-none',
          hasError ? 'border-danger focus:border-danger focus:ring-danger' : 'border-border',
          disabled && 'bg-muted-20 cursor-not-allowed opacity-50'
        )}
      />
      {maxLength && (
        <div className="flex justify-between">
          {hasError && (
            <p className="text-sm text-danger">{error}</p>
          )}
          <p className={clsx('text-xs text-muted', hasError && 'ml-auto')}>
            {currentLength}/{maxLength}
          </p>
        </div>
      )}
      {hasError && !maxLength && (
        <p className="text-sm text-danger">{error}</p>
      )}
    </div>
  )
}
