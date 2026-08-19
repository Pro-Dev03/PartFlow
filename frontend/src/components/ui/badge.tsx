import type { HTMLAttributes } from 'react';
import { forwardRef } from 'react';
import { cn } from '../../lib/utils';

export interface BadgeProps extends HTMLAttributes<HTMLDivElement> {
  variant?: 'default' | 'success' | 'warning' | 'danger' | 'info' | 'destructive' | 'secondary' | 'outline';
  size?: 'sm' | 'md' | 'lg';
  dot?: boolean;
}

const Badge = forwardRef<HTMLDivElement, BadgeProps>(
  ({ className, variant = 'default', size = 'md', dot = false, children, ...props }, ref) => {
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
        aria-label={typeof children === 'string' ? children : undefined}
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
      </div>
    );
  }
);

Badge.displayName = 'Badge';

export { Badge };