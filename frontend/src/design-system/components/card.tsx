import type { HTMLAttributes } from 'react';
import { forwardRef } from 'react';
import { cn } from '../../utils';

export interface CardProps extends HTMLAttributes<HTMLDivElement> {
  noPadding?: boolean;
  hoverable?: boolean;
  variant?: 'default' | 'open' | 'interactive' | 'featured' | 'warning' | 'ai' | 'danger' | 'success' | 'info';
  fullWidth?: boolean;
  'aria-label'?: string;
}

const Card = forwardRef<HTMLDivElement, CardProps>(
  ({
    className,
    noPadding = false,
    hoverable = false,
    variant = 'default',
    fullWidth = false,
    children,
    onClick,
    'aria-label': ariaLabel,
    onKeyDown,
    ...props
  }, ref) => {
    const getVariantStyle = () => {
  const baseStyle: Record<string, string> = {
    border: '1px solid var(--card-border)',
    borderRadius: '16px',
    background: 'var(--card-bg)',
    boxShadow: 'var(--shadow-sm)',
    transition: 'transform 160ms ease, border-color 160ms ease, background-color 160ms ease, box-shadow 160ms ease'
  };

      const variantStyles: Record<string, Record<string, string>> = {
        open: {
          border: 'none',
          borderRadius: '0',
          background: 'transparent',
          boxShadow: 'none',
          transition: 'none',
          padding: '0',
        },
        default: {
          ...baseStyle,
          boxShadow: 'var(--shadow-sm)'
        },
        interactive: {
          ...baseStyle,
          cursor: 'pointer'
        },
        featured: {
          ...baseStyle,
          borderColor: 'var(--color-primary-30)',
          boxShadow: 'var(--shadow-sm)'
        },
        warning: {
          ...baseStyle,
          borderColor: 'var(--color-warning-30)',
          background: 'var(--color-warning-05)',
          boxShadow: 'var(--shadow-card)'
        },
        ai: {
          ...baseStyle,
          borderColor: 'var(--color-primary-20)',
          background: 'var(--card-bg)',
          boxShadow: 'var(--shadow-sm)'
        },
        danger: {
          ...baseStyle,
          borderColor: 'var(--color-danger-30)',
          background: 'var(--color-danger-05)',
          boxShadow: 'var(--shadow-sm)'
        },
        success: {
          ...baseStyle,
          borderColor: 'var(--color-success-30)',
          background: 'var(--color-success-05)',
          boxShadow: 'var(--shadow-sm)'
        },
        info: {
          ...baseStyle,
          borderColor: 'var(--color-info-30)',
          background: 'var(--color-info-05)',
          boxShadow: 'var(--shadow-sm)'
        }
      };

      return variantStyles[variant] || variantStyles.default;
    };

    const responsive = fullWidth ? 'w-full' : '';

    const isInteractive = hoverable || variant === 'interactive';

    const handleKeyDown = (e: React.KeyboardEvent<HTMLDivElement>) => {
      // Call custom onKeyDown if provided
      if (onKeyDown) {
        onKeyDown(e);
      }

      // Default keyboard behavior for interactive cards
      if (isInteractive && (e.key === 'Enter' || e.key === ' ')) {
        e.preventDefault();
        if (onClick) {
          // Create a mouse-like event from the keyboard event
          const syntheticEvent = {
            ...e,
            target: e.currentTarget,
            currentTarget: e.currentTarget,
          } as unknown as React.MouseEvent<HTMLDivElement>;
          onClick(syntheticEvent);
        }
      }
    };

    return (
      <div
        ref={ref}
        style={getVariantStyle()}
        className={cn(
          responsive,
          'pf-card',
          `pf-card-${variant}`,
          noPadding && 'p-0',
          isInteractive && 'cursor-pointer',
          isInteractive && 'pf-card-interactive',
          className
        )}
        onClick={onClick}
        onKeyDown={handleKeyDown}
        tabIndex={isInteractive ? 0 : undefined}
        role={isInteractive ? 'button' : undefined}
        aria-label={isInteractive ? ariaLabel : undefined}
        {...props}
      >
        {children}
      </div>
    );
  }
);

Card.displayName = 'Card';

const CardHeader = forwardRef<HTMLDivElement, HTMLAttributes<HTMLDivElement>>(
  ({ className, ...props }, ref) => (
    <div
      ref={ref}
      className={cn('pf-card-header flex justify-between items-center', className)}
      {...props}
    />
  )
);

CardHeader.displayName = 'CardHeader';

const CardTitle = forwardRef<HTMLParagraphElement, HTMLAttributes<HTMLHeadingElement>>(
  ({ className, ...props }, ref) => (
    <h3
      ref={ref}
      className={cn('font-semibold text-text', className)}
      style={{ fontSize: 'var(--font-size-body)' }}
      {...props}
    />
  )
);

CardTitle.displayName = 'CardTitle';

const CardDescription = forwardRef<HTMLParagraphElement, HTMLAttributes<HTMLParagraphElement>>(
  ({ className, ...props }, ref) => (
    <p
      ref={ref}
      className={cn('text-text-muted', className)}
      style={{ fontSize: 'var(--font-size-caption)' }}
      {...props}
    />
  )
);

CardDescription.displayName = 'CardDescription';

const CardContent = forwardRef<HTMLDivElement, HTMLAttributes<HTMLDivElement> & { noPadding?: boolean }>(
  ({ className, noPadding, ...props }, ref) => (
    <div 
      ref={ref} 
      className={cn('pf-card-content', noPadding && 'p-0', className)}
      {...props} 
    />
  )
);

CardContent.displayName = 'CardContent';

const CardFooter = forwardRef<HTMLDivElement, HTMLAttributes<HTMLDivElement>>(
  ({ className, ...props }, ref) => (
    <div
      ref={ref}
      className={cn('pf-card-footer flex items-center border-t border-border', className)}
      {...props}
    />
  )
);

CardFooter.displayName = 'CardFooter';

export { Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter };
