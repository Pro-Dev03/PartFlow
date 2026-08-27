import type { HTMLAttributes } from 'react';
import { forwardRef } from 'react';
import { cn } from '../../utils';

export interface BadgeProps extends HTMLAttributes<HTMLDivElement> {
  variant?: 'default' | 'success' | 'warning' | 'danger' | 'info' | 'destructive' | 'secondary' | 'outline';
  size?: 'xs' | 'sm' | 'md' | 'lg';
  dot?: boolean;
  'aria-label'?: string;
}

const Badge = forwardRef<HTMLDivElement, BadgeProps>(
  ({ className, variant = 'default', size = 'md', dot = false, children, 'aria-label': ariaLabel, ...props }, ref) => {
    const variants = {
      default: 'bg-surface border border-border text-text',
      success: 'bg-success/8 text-success border border-success/20',
      warning: 'bg-warning/8 text-warning border border-warning/20',
      danger: 'bg-danger/8 text-danger border border-danger/20',
      info: 'bg-info/8 text-info border border-info/20',
      destructive: 'bg-danger text-white border border-danger',
      secondary: 'bg-surface-2 text-text border border-border',
      outline: 'border border-border text-text bg-transparent',
    };

    const sizeClasses = {
      xs: 'px-[var(--badge-padding-xs)] text-[var(--badge-font-size-xs)]',
      sm: 'px-[var(--badge-padding-sm)] text-[var(--badge-font-size-sm)]',
      md: 'px-[var(--badge-padding-md)] text-[var(--badge-font-size-md)]',
      lg: 'px-[var(--badge-padding-lg)] text-[var(--badge-font-size-lg)]',
    };

    const getStatusText = () => {
      switch (variant) {
        case 'success': return 'نشط';
        case 'warning': return 'تحذير';
        case 'danger': return 'خطر';
        case 'info': return 'معلومات';
        case 'default': return 'افتراضي';
        default: return 'حالة';
      }
    };

    return (
      <div
        ref={ref}
        className={cn(
          'inline-flex items-center gap-1.5 font-medium tracking-wide transition-colors duration-normal',
          'rounded-[var(--badge-border-radius)]',
          variants[variant],
          sizeClasses[size] || sizeClasses.md,
          className
        )}
        role="status"
        aria-label={ariaLabel || (typeof children === 'string' ? children : undefined)}
        {...props}
      >
        {dot && (
          <span
            className={cn(
              'w-1.5 h-1.5 rounded-full',
              variant === 'success' && 'bg-success',
              variant === 'warning' && 'bg-warning',
              variant === 'danger' && 'bg-danger',
              variant === 'info' && 'bg-info',
              variant === 'default' && 'bg-text-muted'
            )}
            aria-hidden="true"
          />
        )}
        {children}
        {!ariaLabel && typeof children !== 'string' && (
          <span className="sr-only" aria-hidden="true">
            {getStatusText()}
          </span>
        )}
      </div>
    );
  }
);

Badge.displayName = 'Badge';

export { Badge };
