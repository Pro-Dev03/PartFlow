import type { HTMLAttributes } from 'react';
import { forwardRef } from 'react';
import { cn } from '../../utils';

export interface BadgeProps extends HTMLAttributes<HTMLDivElement> {
  variant?: 'default' | 'success' | 'warning' | 'danger' | 'info' | 'destructive' | 'secondary' | 'outline';
  size?: 'sm' | 'md' | 'lg';
  dot?: boolean;
  'aria-label'?: string;
}

const Badge = forwardRef<HTMLDivElement, BadgeProps>(
  ({ className, variant = 'default', size = 'md', dot = false, children, 'aria-label': ariaLabel, ...props }, ref) => {
    const variants = {
      default: 'bg-surface border border-border text-text',
      success: 'bg-green/8 text-green border border-green/20',
      warning: 'bg-yellow/8 text-yellow border border-yellow/20',
      danger: 'bg-red/8 text-red border border-red/20',
      info: 'bg-cyan/8 text-cyan border border-cyan/20',
      destructive: 'bg-red text-white border border-red',
      secondary: 'bg-surface-2 text-text border border-border',
      outline: 'border border-border text-text bg-transparent',
    };

    const sizes = {
      sm: 'px-2 py-0.5 text-tiny',
      md: 'px-2 py-1 text-tiny',
      lg: 'px-3 py-1 text-small',
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
          'inline-flex items-center gap-1.5 rounded-lg font-medium tracking-wide transition-colors duration-normal',
          variants[variant],
          sizes[size],
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
              variant === 'success' && 'bg-green',
              variant === 'warning' && 'bg-yellow',
              variant === 'danger' && 'bg-red',
              variant === 'info' && 'bg-cyan',
              variant === 'default' && 'bg-text-muted'
            )}
            aria-hidden="true"
          />
        )}
        {children}
        {/* Screen reader text for color-only status */}
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