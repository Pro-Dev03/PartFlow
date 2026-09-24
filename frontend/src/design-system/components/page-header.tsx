import type { HTMLAttributes, ReactNode } from 'react';
import { forwardRef } from 'react';
import { cn } from '../../utils';

export interface PageHeaderProps extends HTMLAttributes<HTMLDivElement> {
  title: string;
  description?: string;
  eyebrow?: string;
  actions?: ReactNode;
  breadcrumbs?: ReactNode;
}

const PageHeader = forwardRef<HTMLDivElement, PageHeaderProps>(
  ({ className, title, description, eyebrow, actions, breadcrumbs, children, ...props }, ref) => {
    return (
      <header 
        ref={ref} 
        className={cn('pf-page-header', className)}
        {...props}
      >
        {breadcrumbs && (
          <nav className="mb-md" aria-label="مسار التنقل">
            {breadcrumbs}
          </nav>
        )}
        
        <div className="pf-page-header-main">
          <div className="pf-page-header-copy">
            {eyebrow && (
              <div className="text-eyebrow text-text-muted mb-1">
                {eyebrow}
              </div>
            )}
            <h1 className="text-h1 font-bold text-text tracking-tight mb-1">
              {title}
            </h1>
            {description && (
              <p className="text-small text-text-muted">
                {description}
              </p>
            )}
          </div>
          
          {actions && (
            <div className="pf-page-header-actions">
              {actions}
            </div>
          )}
        </div>
        
        {children && (
          <div className="mt-md">
            {children}
          </div>
        )}
      </header>
    );
  }
);

PageHeader.displayName = 'PageHeader';

export { PageHeader };
