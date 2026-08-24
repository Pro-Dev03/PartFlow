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
      sm: isMobile ? 'h-11 px-4 text-sm' : 'h-10 px-4 text-sm', // worktrack: 40px
      md: isMobile ? 'h-12 px-4 py-3 text-sm' : 'h-12 px-4 py-3 text-sm', // worktrack: 48px
      lg: isMobile ? 'h-14 px-5 py-3.5 text-base' : 'h-14 px-5 py-3.5 text-base', // worktrack: 56px
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
              'flex w-full rounded-xl border',
              'focus-visible:outline-none focus-visible:ring-2 focus-visible:border-transparent',
              'transition-all duration-300',
              'disabled:cursor-not-allowed disabled:opacity-50',
              'read-only:cursor-default',
              'placeholder:text-text-muted/70',
              'hover:border-cyan/30 hover:shadow-sm',
              sizes[size],
              hasError && 'border-red-500 focus-visible:ring-red-500/50',
              hasSuccess && 'border-green-500 focus-visible:ring-green-500/50',
              !hasError && !hasSuccess && 'focus-visible:ring-cyan/30 focus-visible:border-cyan/50 focus-visible:shadow-lg',
              className
            )}
            style={{
              background: 'linear-gradient(135deg, rgba(17, 24, 39, 0.8) 0%, rgba(17, 24, 39, 0.6) 100%)',
              border: hasError ? '1px solid rgba(239, 68, 68, 0.3)' : hasSuccess ? '1px solid rgba(34, 197, 94, 0.3)' : '1px solid rgba(99, 102, 241, 0.2)',
              backdropFilter: 'blur(10px)',
              color: 'var(--text-primary)',
              fontSize: '14px',
              fontWeight: '500',
              letterSpacing: '0.2px',
              boxShadow: hasError ? '0 2px 8px rgba(239, 68, 68, 0.15)' : hasSuccess ? '0 2px 8px rgba(34, 197, 94, 0.15)' : 'none',
              transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)'
            }}
            disabled={disabled}
            readOnly={readOnly}
            aria-invalid={hasError}
            aria-describedby={hasError ? `${inputId}-error` : helperText ? `${inputId}-helper` : undefined}
            onMouseEnter={(e) => {
              if (!disabled && !readOnly) {
                if (!hasError && !hasSuccess) {
                  e.currentTarget.style.borderColor = 'rgba(99, 102, 241, 0.4)';
                  e.currentTarget.style.boxShadow = '0 4px 15px rgba(99, 102, 241, 0.15)';
                }
              }
            }}
            onMouseLeave={(e) => {
              if (!disabled && !readOnly) {
                if (!hasError && !hasSuccess) {
                  e.currentTarget.style.borderColor = 'rgba(99, 102, 241, 0.2)';
                  e.currentTarget.style.boxShadow = 'none';
                }
              }
            }}
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