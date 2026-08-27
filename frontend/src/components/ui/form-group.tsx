import { forwardRef } from 'react';
import { cn } from '../../utils';

export interface FormGroupProps {
  label?: string;
  error?: string;
  success?: string;
  helperText?: string;
  required?: boolean;
  children: React.ReactNode;
  className?: string;
}

const FormGroup = forwardRef<HTMLDivElement, FormGroupProps>(
  ({ 
    label, 
    error, 
    success,
    helperText, 
    required = false,
    children,
    className
  }, ref) => {
    const groupId = `form-group-${Math.random().toString(36).substr(2, 9)}`;
    
    return (
      <div ref={ref} className={cn('w-full', className)}>
        {label && (
          <label
            htmlFor={groupId}
            className="block mb-2 transition-colors duration-200"
            style={{ 
              color: 'var(--form-label-color)', 
              fontSize: 'var(--form-label-font-size)',
              fontWeight: 'var(--form-label-font-weight)'
            }}
          >
            {label}
            {required && <span style={{ color: 'var(--color-danger)' }}>*</span>}
          </label>
        )}
        {children}
        {error && (
          <p 
            id={`${groupId}-error`} 
            className="mt-1 transition-colors duration-200"
            style={{ 
              color: 'var(--form-error-color)', 
              fontSize: 'var(--form-error-font-size)'
            }}
            role="alert"
          >
            {error}
          </p>
        )}
        {success && !error && (
          <p 
            id={`${groupId}-success`} 
            className="mt-1 transition-colors duration-200"
            style={{ 
              color: 'var(--form-success-color)', 
              fontSize: 'var(--form-error-font-size)'
            }}
          >
            {success}
          </p>
        )}
        {helperText && !error && !success && (
          <p 
            id={`${groupId}-helper`} 
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

FormGroup.displayName = 'FormGroup';

export { FormGroup };