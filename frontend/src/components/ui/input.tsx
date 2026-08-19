import type { InputHTMLAttributes } from 'react';
import { forwardRef } from 'react';
import { cn } from '../../lib/utils';

export interface InputProps extends Omit<InputHTMLAttributes<HTMLInputElement>, 'size'> {
  label?: string;
  error?: string;
  success?: string;
  helperText?: string;
  leftIcon?: React.ReactNode;
  rightIcon?: React.ReactNode;
  fullWidth?: boolean;
  size?: 'sm' | 'md' | 'lg';
}

const Input = forwardRef<HTMLInputElement, InputProps>(
  ({ 
    className, 
    type = 'text', 
    label, 
    error, 
    success,
    helperText, 
    leftIcon, 
    rightIcon, 
    id,
    fullWidth = false,
    size = 'md',
    disabled = false,
    readOnly = false,
    ...props 
  }, ref) => {
    const inputId = id || `input-${Math.random().toString(36).substr(2, 9)}`;
    
    const hasError = !!error;
    const hasSuccess = !!success && !hasError;
    
    const sizes = {
      sm: 'h-8 px-3 text-small',
      md: 'h-10 px-3 py-2 text-body',
      lg: 'h-12 px-4 py-3 text-small',
    };
    
    const containerClass = fullWidth ? 'w-full' : '';
    
    return (
      <div className={cn('w-full', containerClass)}>
        {label && (
          <label
            htmlFor={inputId}
            className="block text-small font-medium text-text mb-1"
          >
            {label}
          </label>
        )}
        <div className="relative">
          {leftIcon && (
            <div className="absolute inset-y-0 start-0 flex items-center ps-3 pointer-events-none">
              <span className={cn('transition-colors duration-normal', hasError ? 'text-red' : 'text-text-muted')} aria-hidden="true">{leftIcon}</span>
            </div>
          )}
          <input
            ref={ref}
            type={type}
            id={inputId}
            className={cn(
              'flex w-full rounded-sm border border-border bg-surface',
              'placeholder:text-text-muted',
              'focus-visible:outline-none focus-visible:ring-2 focus-visible:border-transparent',
              'transition-all duration-normal',
              'disabled:cursor-not-allowed disabled:opacity-50',
              'read-only:cursor-default read-only:bg-surface-2',
              sizes[size],
              leftIcon && 'ps-10',
              rightIcon && 'pe-10',
              hasError && 'border-red focus-visible:ring-red shadow-glow-soft',
              hasSuccess && 'border-green focus-visible:ring-green',
              !hasError && !hasSuccess && 'focus-visible:ring-cyan focus-visible:border-cyan/50',
              className
            )}
            disabled={disabled}
            readOnly={readOnly}
            aria-invalid={hasError}
            aria-describedby={hasError ? `${inputId}-error` : helperText ? `${inputId}-helper` : undefined}
            {...props}
          />
          {rightIcon && (
            <div className="absolute inset-y-0 end-0 flex items-center pe-3 pointer-events-none">
              <span className="text-text-muted" aria-hidden="true">{rightIcon}</span>
            </div>
          )}
        </div>
        {error && (
          <p id={`${inputId}-error`} className="mt-1 text-small text-red" role="alert">
            {error}
          </p>
        )}
        {success && !error && (
          <p id={`${inputId}-success`} className="mt-1 text-small text-green">
            {success}
          </p>
        )}
        {helperText && !error && !success && (
          <p id={`${inputId}-helper`} className="mt-1 text-small text-text-muted">
            {helperText}
          </p>
        )}
      </div>
    );
  }
);

Input.displayName = 'Input';

export { Input };