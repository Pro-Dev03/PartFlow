import type { InputHTMLAttributes } from 'react';
import { forwardRef, useState, useEffect } from 'react';
import { cn } from '../../utils';

export interface InputProps extends Omit<InputHTMLAttributes<HTMLInputElement>, 'size'> {
  label?: string;
  error?: string;
  success?: string;
  helperText?: string;
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
    id,
    fullWidth = false,
    size = 'md',
    disabled = false,
    readOnly = false,
    ...props 
  }, ref) => {
    const [isMobile, setIsMobile] = useState(false);
    
    useEffect(() => {
      const checkMobile = () => {
        setIsMobile(window.innerWidth < 768);
      };
      
      checkMobile();
      window.addEventListener('resize', checkMobile);
      return () => window.removeEventListener('resize', checkMobile);
    }, []);
    
    const inputId = id || `input-${Math.random().toString(36).substr(2, 9)}`;
    
    const hasError = !!error;
    const hasSuccess = !!success && !hasError;
    
    const sizes = {
      sm: isMobile ? 'h-11 px-4 text-sm' : 'h-10 px-4 text-sm', // حسب التقرير النهائي: 32px
      md: isMobile ? 'h-12 px-4 py-3 text-sm' : 'h-12 px-4 py-3 text-sm', // حسب التقرير النهائي: 40px
      lg: isMobile ? 'h-14 px-5 py-3.5 text-base' : 'h-14 px-5 py-3.5 text-base', // حسب التقرير النهائي: 48px
    };
    
    const containerClass = fullWidth ? 'w-full' : '';
    
    return (
      <div className={cn('w-full', containerClass)}>
        {label && (
          <label
            htmlFor={inputId}
            className="block text-xs font-semibold mb-2 transition-colors duration-200"
            style={{ color: 'var(--text-primary)', letterSpacing: '0.3px' }}
          >
            {label}
          </label>
        )}
        <div className="relative group">
          <input
            ref={ref}
            type={type}
            id={inputId}
            className={cn(
              'input-base',
              'flex w-full rounded-xl border',
              'focus-visible:outline-none focus-visible:ring-2 focus-visible:border-transparent',
              'transition-all duration-200',
              'disabled:cursor-not-allowed disabled:opacity-50',
              'read-only:cursor-default',
              'placeholder:text-text-muted/70',
              sizes[size],
              hasError && 'input-error',
              hasSuccess && 'input-success',
              !hasError && !hasSuccess && 'input-default',
              className
            )}
            disabled={disabled}
            readOnly={readOnly}
            aria-invalid={hasError}
            aria-describedby={hasError ? `${inputId}-error` : helperText ? `${inputId}-helper` : undefined}
            {...props}
          />
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