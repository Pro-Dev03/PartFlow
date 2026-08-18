import { clsx } from 'clsx'

interface FormSelectProps {
  label?: string
  name: string
  value: string
  onChange: (value: string) => void
  onBlur?: () => void
  error?: string
  touched?: boolean
  options: { value: string; label: string }[]
  placeholder?: string
  required?: boolean
  disabled?: boolean
  className?: string
}

export function FormSelect({
  label,
  name,
  value,
  onChange,
  onBlur,
  error,
  touched,
  options,
  placeholder,
  required = false,
  disabled = false,
  className,
}: FormSelectProps) {
  const hasError = touched && error

  return (
    <div className={clsx('space-y-1', className)}>
      {label && (
        <label htmlFor={name} className="block text-sm font-medium text-text">
          {label}
          {required && <span className="text-danger ml-1">*</span>}
        </label>
      )}
      <select
        id={name}
        name={name}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        onBlur={onBlur}
        disabled={disabled}
        required={required}
        className={clsx(
          'w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent transition-colors',
          hasError ? 'border-danger focus:border-danger focus:ring-danger' : 'border-border',
          disabled && 'bg-muted-20 cursor-not-allowed opacity-50'
        )}
      >
        {placeholder && (
          <option value="">{placeholder}</option>
        )}
        {options.map((option) => (
          <option key={option.value} value={option.value}>
            {option.label}
          </option>
        ))}
      </select>
      {hasError && (
        <p className="text-sm text-danger">{error}</p>
      )}
    </div>
  )
}
