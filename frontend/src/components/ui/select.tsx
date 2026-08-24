import type { SelectHTMLAttributes } from 'react';
import { forwardRef, useRef, useState } from 'react';
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
    const [isFocused, setIsFocused] = useState(false);
    
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
              'pe-10', // space for arrow
              className
            )}
            style={{
              width: '100%',
              background: 'linear-gradient(135deg, rgba(30, 41, 59, 0.9) 0%, rgba(30, 41, 59, 0.7) 100%)',
              border: isFocused || hasError 
                ? (hasError ? '2px solid #ef4444' : '2px solid #6366f1') 
                : '1px solid #6366f1',
              backdropFilter: 'blur(10px)',
              color: 'var(--text-primary)',
              fontSize: '14px',
              fontWeight: '500',
              letterSpacing: '0.2px',
              boxShadow: isFocused ? '0 0 0 3px rgba(99, 102, 241, 0.3), 0 4px 20px rgba(99, 102, 241, 0.4)' : (hasError ? '0 2px 8px rgba(239, 68, 68, 0.15)' : '0 2px 8px rgba(0, 0, 0, 0.1)'),
              transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)'
            }}
            disabled={disabled || isLoading}
            aria-invalid={hasError}
            aria-describedby={hasError ? `${selectId}-error` : helperText ? `${selectId}-helper` : undefined}
            {...props}
            onFocus={(e) => {
              setIsFocused(true);
              if (arrowRef.current) {
                arrowRef.current.style.transform = 'rotate(180deg)';
              }
              if (props.onFocus) props.onFocus(e);
            }}
            onBlur={(e) => {
              setIsFocused(false);
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
                    style={{
                      background: 'rgba(17, 24, 39, 0.98)',
                      color: 'var(--text-primary)',
                      padding: '10px 14px',
                      fontSize: '14px',
                      fontWeight: '500',
                      letterSpacing: '0.2px'
                    }}
                  >
                    {option.label}
                  </option>
                ))
              ) : (
                <option value="" disabled style={{ 
                  background: 'rgba(17, 24, 39, 0.98)',
                  color: 'var(--text-secondary)',
                  padding: '10px 14px'
                }}>
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
                'transition-all duration-300',
                isFocused ? 'text-cyan-400 scale-110' : 'text-text-muted/50 scale-100',
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