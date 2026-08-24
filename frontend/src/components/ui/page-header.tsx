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
        className={cn('mb-2xl', className)} 
        style={{ marginBottom: '24px' }} // worktrack: 24px
        {...props}
      >
        {breadcrumbs && (
          <nav className="mb-md" aria-label="Breadcrumb">
            {breadcrumbs}
          </nav>
        )}
        
        <div className="flex items-start justify-between gap-md mb-sm">
          <div className="flex-1">
            {eyebrow && (
              <div className="text-eyebrow text-cyan mb-1" aria-hidden="true" style={{ fontSize: '12px' }}>
                {eyebrow}
              </div>
            )}
            <h1 className="text-h1 font-extrabold text-text tracking-tight mb-1" style={{ fontSize: '18px' }}>
              {title}
            </h1>
            {description && (
              <p className="text-small text-text-muted" style={{ fontSize: '13px' }}>
                {description}
              </p>
            )}
          </div>
          
          {actions && (
            <div className="flex items-center gap-sm">
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