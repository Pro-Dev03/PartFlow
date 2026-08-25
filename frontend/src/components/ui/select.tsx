import type { SelectHTMLAttributes } from 'react';
import { forwardRef, useRef } from 'react';
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
    const selectId = id || `select-${Math.random().toString(36).substr(2, 9)}`;
    const arrowRef = useRef<SVGSVGElement>(null);
    
    const hasError = !!error;
    const isLoading = loading;
    const hasOptions = options && options.length > 0;
    
    const sizes = {
      sm: 'h-10 text-sm',
      md: 'h-12 text-sm',
      lg: 'h-14 text-base',
    };
    
    const containerClass = fullWidth ? 'w-full' : '';
    
    return (
      <div className={cn('w-full', containerClass)}>
        {label && (
          <label
            htmlFor={selectId}
            className="block text-xs font-semibold mb-2 transition-colors duration-200"
            style={{ color: 'var(--text-primary)', letterSpacing: '0.3px' }}
          >
            {label}
          </label>
        )}
        <div className="relative w-full">
          <select
            ref={ref}
            id={selectId}
            className={cn(
              'select-custom',
              'flex w-full rounded-xl border appearance-none cursor-pointer',
              'focus-visible:outline-none focus-visible:ring-2 focus-visible:border-transparent',
              'transition-all duration-300',
              'disabled:cursor-not-allowed disabled:opacity-50',
              'placeholder:text-text-muted/40',
              'hover:border-cyan/30 hover:shadow-sm',
              sizes[size],
              hasError && 'border-red-500 focus-visible:ring-red-500/50',
              !hasError && 'focus-visible:ring-cyan/30 focus-visible:border-cyan/50 focus-visible:shadow-lg',
              'pe-10',
              className
            )}
            disabled={disabled || isLoading}
            aria-invalid={hasError}
            aria-describedby={hasError ? `${selectId}-error` : helperText ? `${selectId}-helper` : undefined}
            {...props}
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
          <div className="absolute inset-y-0 end-0 flex items-center pe-4 pointer-events-none">
             <ChevronDown 
               ref={arrowRef}
               className={cn(
                 'transition-all duration-300 text-text-muted/50',
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