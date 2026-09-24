import type { InputHTMLAttributes } from 'react';
import { forwardRef, useId } from 'react';
import { cn } from '../../utils';

export interface InputProps extends Omit<InputHTMLAttributes<HTMLInputElement>, 'size'> {
  label?: string;
  error?: string;
  success?: string;
  helperText?: string;
  fullWidth?: boolean;
  size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl';
  required?: boolean;
}

const Input = forwardRef<HTMLInputElement, InputProps>(
  ({ 
    className, 
    type = 'text', 
    label, 
    error, 
    success,
    helperText, 
    id,
    fullWidth = false,
    size = 'md',
    disabled = false,
    readOnly = false,
    required = false,
    placeholder,
    ...props 
  }, ref) => {
    const generatedId = useId().replace(/:/g, '');
    const inputId = id || `input-${generatedId}`;
    const errorId = `${inputId}-error`;
    const successId = `${inputId}-success`;
    const helperId = `${inputId}-helper`;
    
    const hasError = !!error;
    const hasSuccess = !!success && !hasError;
    const hasHelper = !!helperText && !hasError && !hasSuccess;
    
    const sizeClasses = {
      xs: 'h-[var(--input-height-xs)] px-[var(--input-padding-xs)] text-[var(--input-font-size-xs)]',
      sm: 'h-[var(--input-height-sm)] px-[var(--input-padding-sm)] text-[var(--input-font-size-sm)]',
      md: 'h-[var(--input-height-md)] px-[var(--input-padding-md)] text-[var(--input-font-size-md)]',
      lg: 'h-[var(--input-height-lg)] px-[var(--input-padding-lg)] text-[var(--input-font-size-lg)]',
      xl: 'h-[var(--input-height-xl)] px-[var(--input-padding-xl)] text-[var(--input-font-size-xl)]',
    };
    const inputHeight = `var(--input-height-${size})`;
    
    const containerClass = fullWidth ? 'w-full' : '';
    
    return (
      <div className={cn('w-full', containerClass)}>
        {label && (
          <label
            htmlFor={inputId}
            className="mb-1.5 block text-sm font-semibold transition-colors duration-200"
            style={{ 
              color: 'var(--text-primary)'
            }}
          >
            {label}
            {required && <span aria-hidden="true" className="ms-1" style={{ color: 'var(--color-danger)' }}>*</span>}
          </label>
        )}
        <div className="relative group">
          <input
            ref={ref}
            type={type}
            id={inputId}
            className={cn(
              'flex box-border w-full min-w-0 rounded-xl border bg-[var(--input-bg)] text-[var(--text-primary)]',
              'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[var(--color-primary-30)] focus-visible:border-[var(--primary)]',
              'transition-[border-color,box-shadow,background-color] duration-150',
              'disabled:cursor-not-allowed disabled:opacity-60 disabled:bg-[var(--bg-surface-3)]',
              'read-only:cursor-default',
              'placeholder:text-[var(--input-placeholder)]',
              sizeClasses[size] || sizeClasses.md,
              hasError && 'border-[var(--danger)] focus-visible:ring-[var(--color-danger-30)] focus-visible:border-[var(--danger)]',
              hasSuccess && 'border-[var(--success)] focus-visible:ring-[var(--color-success-30)] focus-visible:border-[var(--success)]',
              !hasError && !hasSuccess && 'border-[var(--input-border)]',
              className
            )}
            disabled={disabled}
            readOnly={readOnly}
            required={required}
            placeholder={placeholder}
            aria-invalid={hasError}
            aria-describedby={
              hasError ? errorId :
              hasSuccess ? successId :
              hasHelper ? helperId :
              undefined
            }
            aria-required={required}
            {...props}
            style={{ ...props.style, boxSizing: 'border-box', minHeight: inputHeight, height: inputHeight }}
          />
        </div>
        {error && (
          <p 
            id={errorId} 
            className="mt-1 transition-colors duration-200"
            style={{ 
              color: 'var(--form-error-color)', 
              fontSize: 'var(--form-error-font-size)'
            }}
            role="alert"
            aria-live="polite"
          >
            {error}
          </p>
        )}
        {success && !error && (
          <p 
            id={successId} 
            className="mt-1 transition-colors duration-200"
            style={{ 
              color: 'var(--form-success-color)', 
              fontSize: 'var(--form-error-font-size)'
            }}
            role="status"
            aria-live="polite"
          >
            {success}
          </p>
        )}
        {helperText && !error && !success && (
          <p 
            id={helperId} 
            className="mt-1 transition-colors duration-200"
            style={{ 
              color: 'var(--form-hint-color)', 
              fontSize: 'var(--form-hint-font-size)'
            }}
          >
            {helperText}
          </p>
        )}
      </div>
    );
  }
);

Input.displayName = 'Input';

export { Input };
