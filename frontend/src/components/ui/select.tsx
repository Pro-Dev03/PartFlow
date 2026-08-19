import type { SelectHTMLAttributes } from 'react';
import { forwardRef } from 'react';
import { cn } from '../../lib/utils';

export interface SelectProps extends Omit<SelectHTMLAttributes<HTMLSelectElement>, 'size'> {
  label?: string;
  error?: string;
  helperText?: string;
  fullWidth?: boolean;
  size?: 'sm' | 'md' | 'lg';
  options?: { value: string; label: string }[];
}

const Select = forwardRef<HTMLSelectElement, SelectProps>(
  ({ 
    className, 
    label, 
    error, 
    helperText, 
    id,
    fullWidth = false,
    size = 'md',
    disabled = false,
    options,
    children,
    ...props 
  }, ref) => {
    const selectId = id || `select-${Math.random().toString(36).substr(2, 9)}`;
    
    const hasError = !!error;
    
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
            htmlFor={selectId}
            className="block text-small font-medium text-text mb-1"
          >
            {label}
          </label>
        )}
        <select
          ref={ref}
          id={selectId}
          className={cn(
            'flex w-full rounded-sm border border-border bg-surface',
            'focus-visible:outline-none focus-visible:ring-2 focus-visible:border-transparent',
            'transition-all duration-normal',
            'disabled:cursor-not-allowed disabled:opacity-50',
            sizes[size],
            hasError && 'border-red focus-visible:ring-red shadow-glow-soft',
            !hasError && 'focus-visible:ring-cyan focus-visible:border-cyan/50',
            className
          )}
          disabled={disabled}
          aria-invalid={hasError}
          aria-describedby={hasError ? `${selectId}-error` : helperText ? `${selectId}-helper` : undefined}
          {...props}
        >
          {options ? (
            options.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))
          ) : (
            children
          )}
        </select>
        {error && (
          <p id={`${selectId}-error`} className="mt-1 text-small text-red" role="alert">
            {error}
          </p>
        )}
        {helperText && !error && (
          <p id={`${selectId}-helper`} className="mt-1 text-small text-text-muted">
            {helperText}
          </p>
        )}
      </div>
    );
  }
);

Select.displayName = 'Select';

export { Select };