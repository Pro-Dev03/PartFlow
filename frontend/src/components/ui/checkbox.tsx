import type { InputHTMLAttributes } from 'react';
import { forwardRef } from 'react';
import { cn } from '../../utils';

export interface CheckboxProps extends Omit<InputHTMLAttributes<HTMLInputElement>, 'size'> {
  label?: string;
  error?: string;
  helperText?: string;
  fullWidth?: boolean;
  size?: 'sm' | 'md' | 'lg';
}

const Checkbox = forwardRef<HTMLInputElement, CheckboxProps>(
  ({ 
    className, 
    label, 
    error, 
    helperText, 
    id,
    fullWidth = false,
    size = 'md',
    disabled = false,
    ...props 
  }, ref) => {
    const checkboxId = id || `checkbox-${Math.random().toString(36).substr(2, 9)}`;
    
    const hasError = !!error;
    
    const sizes = {
      sm: 'w-4 h-4',
      md: 'w-5 h-5',
      lg: 'w-6 h-6',
    };
    
    const labelSizes = {
      sm: 'text-tiny',
      md: 'text-small',
      lg: 'text-body',
    };
    
    const containerClass = fullWidth ? 'w-full' : '';
    
    return (
      <div className={cn('flex flex-col gap-1', containerClass)}>
        <label className="flex items-center gap-md cursor-pointer group">
          <div className="relative">
            <input
              ref={ref}
              type="checkbox"
              id={checkboxId}
              className={cn(
                'peer sr-only',
                'disabled:cursor-not-allowed disabled:opacity-50',
                className
              )}
              disabled={disabled}
              aria-invalid={hasError}
              aria-describedby={hasError ? `${checkboxId}-error` : helperText ? `${checkboxId}-helper` : undefined}
              {...props}
            />
            <div className={cn(
              'flex items-center justify-center border-2 rounded-md transition-all duration-normal',
              'bg-surface border-border',
              'peer-hover:border-cyan/50',
              'peer-focus-visible:ring-2 peer-focus-visible:ring-cyan peer-focus-visible:ring-offset-2 peer-focus-visible:ring-offset-bg',
              'peer-checked:bg-cyan/20 peer-checked:border-cyan',
              'peer-disabled:opacity-50 peer-disabled:cursor-not-allowed',
              hasError && 'border-red peer-focus-visible:ring-red',
              sizes[size]
            )}>
              <svg 
                className={cn(
                  'w-3/4 h-3/4 text-cyan opacity-0 transition-opacity duration-normal',
                  'peer-checked:opacity-100',
                  hasError && 'text-red'
                )}
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                strokeWidth="3"
              >
                <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
              </svg>
            </div>
          </div>
          {label && (
            <span className={cn(
              'font-medium transition-colors duration-normal',
              'text-text-secondary group-hover:text-text',
              'peer-disabled:text-text-disabled',
              labelSizes[size]
            )}>
              {label}
            </span>
          )}
        </label>
        {error && (
          <p id={`${checkboxId}-error`} className="text-tiny text-red" role="alert">
            {error}
          </p>
        )}
        {helperText && !error && (
          <p id={`${checkboxId}-helper`} className="text-tiny text-text-muted">
            {helperText}
          </p>
        )}
      </div>
    );
  }
);

Checkbox.displayName = 'Checkbox';

export { Checkbox };
