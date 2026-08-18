import { clsx } from 'clsx'

interface FormInputProps {
  label?: string
  name: string
  type?: 'text' | 'email' | 'tel' | 'password' | 'number' | 'date'
  value: string | number
  onChange: (value: string | number) => void
  onBlur?: () => void
  error?: string
  touched?: boolean
  placeholder?: string
  required?: boolean
  disabled?: boolean
  className?: string
}

export function FormInput({
  label,
  name,
  type = 'text',
  value,
  onChange,
  onBlur,
  error,
  touched,
  placeholder,
  required = false,
  disabled = false,
  className,
}: FormInputProps) {
  const hasError = touched && error

  return (
    <div className={clsx('space-y-1', className)}>
      {label && (
        <label htmlFor={name} className="block text-sm font-medium text-text">
          {label}
          {required && <span className="text-danger ml-1">*</span>}
        </label>
      )}
      <input
        id={name}
        name={name}
        type={type}
        value={value}
        onChange={(e) => onChange(type === 'number' ? parseFloat(e.target.value) || 0 : e.target.value)}
        onBlur={onBlur}
        placeholder={placeholder}
        disabled={disabled}
        required={required}
        className={clsx(
          'w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent transition-colors',
          hasError ? 'border-danger focus:border-danger focus:ring-danger' : 'border-border',
          disabled && 'bg-muted-20 cursor-not-allowed opacity-50'
        )}
      />
      {hasError && (
        <p className="text-sm text-danger">{error}</p>
      )}
    </div>
  )
}
