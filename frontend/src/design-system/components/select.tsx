import type { SelectHTMLAttributes } from 'react';
import { forwardRef, useId, useRef } from 'react';
import { cn } from '../../utils';
import { ChevronDown } from 'lucide-react';

export interface SelectProps extends Omit<SelectHTMLAttributes<HTMLSelectElement>, 'size'> {
  label?: string;
  error?: string;
  helperText?: string;
  fullWidth?: boolean;
  size?: 'sm' | 'md' | 'lg';
  options?: { value: string; label: string }[];
  loading?: boolean;
  emptyMessage?: string;
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
    loading = false,
    emptyMessage = 'لا توجد بيانات',
    children,
    ...props 
  }, ref) => {
    const generatedId = useId().replace(/:/g, '');
    const selectId = id || `select-${generatedId}`;
    const arrowRef = useRef<SVGSVGElement>(null);
    
    const hasError = !!error;
    const isLoading = loading;
    const hasOptions = options && options.length > 0;
    
    const sizes = {
      sm: { className: 'text-sm', height: 'var(--input-height-sm)' },
      md: { className: 'text-sm', height: 'var(--input-height-md)' },
      lg: { className: 'text-base', height: 'var(--input-height-lg)' },
    };
    
    const containerClass = fullWidth ? 'w-full' : '';
    
    return (
      <div className={cn('w-full', containerClass)}>
        {label && (
          <label
            htmlFor={selectId}
            className="mb-1.5 block text-sm font-semibold transition-colors duration-200"
            style={{ color: 'var(--text-primary)', letterSpacing: '0.3px' }}
          >
            {label}
          </label>
        )}
        <div className="select-control-wrapper relative w-full">
          <select
            ref={ref}
            id={selectId}
            className={cn(
              'pf-select-control',
              'select-custom',
              'flex w-full appearance-none rounded-xl border bg-[var(--input-bg)] text-[var(--text-primary)] cursor-pointer',
              'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[var(--color-primary-30)] focus-visible:border-[var(--primary)]',
              'transition-[border-color,box-shadow,background-color] duration-150',
              'disabled:cursor-not-allowed disabled:opacity-60 disabled:bg-[var(--bg-surface-3)]',
              'hover:border-[var(--primary)]',
              sizes[size].className,
              hasError && 'border-[var(--danger)] focus-visible:ring-[var(--color-danger-30)] focus-visible:border-[var(--danger)]',
              !hasError && 'border-[var(--input-border)]',
              'pe-10',
              className
            )}
            disabled={disabled || isLoading}
            aria-invalid={hasError}
            aria-describedby={hasError ? `${selectId}-error` : helperText ? `${selectId}-helper` : undefined}
            {...props}
            style={{ ...props.style, boxSizing: 'border-box', minHeight: sizes[size].height, height: sizes[size].height }}
            onFocus={(e) => {
              if (arrowRef.current) {
                arrowRef.current.style.transform = 'rotate(180deg)';
              }
              if (props.onFocus) props.onFocus(e);
            }}
            onBlur={(e) => {
              if (arrowRef.current) {
                arrowRef.current.style.transform = 'rotate(0deg)';
              }
              if (props.onBlur) props.onBlur(e);
            }}
          >
            {isLoading ? (
              <option value="" disabled>
                جاري التحميل...
              </option>
            ) : options ? (
              hasOptions ? (
                options.map((option) => (
                  <option 
                    key={option.value} 
                    value={option.value}
                    className="text-sm font-medium"
                  >
                    {option.label}
                  </option>
                ))
              ) : (
                <option value="" disabled>
                  {emptyMessage}
                </option>
              )
            ) : (
              children
            )}
          </select>
          {/* Chevron Down Arrow */}
          <div className="select-control-arrow absolute inset-y-0 end-0 flex items-center justify-center pointer-events-none">
             <ChevronDown 
               ref={arrowRef}
               className={cn(
                 'transition-transform duration-150 text-[var(--text-tertiary)]',
                 size === 'sm' ? 'w-4 h-4' : size === 'md' ? 'w-5 h-5' : 'w-6 h-6'
               )}
             />
          </div>
        </div>
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
