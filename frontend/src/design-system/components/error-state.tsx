import type { HTMLAttributes, ReactNode } from 'react';
import { forwardRef } from 'react';
import { cn } from '../../utils';

export interface ErrorStateProps extends HTMLAttributes<HTMLDivElement> {
  icon?: ReactNode;
  title: string;
  description?: string;
  action?: ReactNode;
  error?: string;
}

const ErrorState = forwardRef<HTMLDivElement, ErrorStateProps>(
  ({ className, icon, title, description, action, error, children, ...props }, ref) => {
    return (
      <div
        ref={ref}
        className={cn(
          'pf-error-state flex flex-col items-center justify-center py-12 px-4 text-center',
          className
        )}
        role="alert"
        aria-live="assertive"
        {...props}
      >
        {icon && (
          <div className="w-16 h-16 rounded-full flex items-center justify-center mb-4" style={{ background: 'var(--color-danger-10)', color: 'var(--danger)' }} aria-hidden="true">
            {icon}
          </div>
        )}
        <h3 className="text-h3 font-semibold text-text mb-2">
          {title}
        </h3>
        {description && (
          <p className="text-small text-text-muted max-w-sm mb-4">
            {description}
          </p>
        )}
        {error && (
          <p className="text-tiny max-w-sm mb-6 font-mono" style={{ color: 'var(--danger)' }}>
            {error}
          </p>
        )}
        {action && (
          <div className="mt-4">
            {action}
          </div>
        )}
        {children && (
          <div className="mt-6">
            {children}
          </div>
        )}
      </div>
    );
  }
);

ErrorState.displayName = 'ErrorState';

export { ErrorState };
