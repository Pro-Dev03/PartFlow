import type { HTMLAttributes } from 'react';
import { forwardRef } from 'react';
import { cn } from '../../utils';

export interface CardProps extends HTMLAttributes<HTMLDivElement> {
  noPadding?: boolean;
  hoverable?: boolean;
  variant?: 'default' | 'interactive' | 'featured' | 'warning' | 'ai' | 'danger' | 'success' | 'info';
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
    borderRadius: 'var(--card-border-radius)',
    background: 'var(--card-bg)',
    boxShadow: 'var(--shadow-card)',
    transition: '200ms ease',
    padding: 'var(--card-padding-md)'
  };

      const variantStyles: Record<string, Record<string, string>> = {
        default: {
          ...baseStyle,
          boxShadow: 'var(--shadow-card)'
        },
        interactive: {
          ...baseStyle,
          cursor: 'pointer',
          boxShadow: 'var(--shadow-card)'
        },
        featured: {
          ...baseStyle,
          borderColor: 'var(--color-primary-30)',
          boxShadow: '0 4px 12px rgba(0, 0, 0, 0.08)'
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
          background: 'linear-gradient(145deg, var(--color-primary-08), var(--card-bg))',
          boxShadow: 'var(--shadow-card)'
        },
        danger: {
          ...baseStyle,
          borderColor: 'var(--color-danger-30)',
          background: 'var(--color-danger-05)',
          boxShadow: 'var(--shadow-card)'
        },
        success: {
          ...baseStyle,
          borderColor: 'var(--color-success-30)',
          background: 'var(--color-success-05)',
          boxShadow: 'var(--shadow-card)'
        },
        info: {
          ...baseStyle,
          borderColor: 'var(--color-info-30)',
          background: 'var(--color-info-05)',
          boxShadow: 'var(--shadow-card)'
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
          isInteractive && 'cursor-pointer',
          className
        )}
        onMouseEnter={(e) => {
          if (isInteractive) {
            e.currentTarget.style.borderColor = 'rgba(148, 163, 184, 0.22)';
            e.currentTarget.style.transform = 'translateY(-2px)'; // subtle lift فقط
          }
        }}
        onMouseLeave={(e) => {
          const currentStyle = getVariantStyle();
          e.currentTarget.style.borderColor = currentStyle.borderColor;
          e.currentTarget.style.transform = 'translateY(0)';
        }}
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
      className={cn('flex justify-between items-center', className)}
      style={{ 
        marginBottom: 'var(--spacing-4)',
        padding: 'var(--card-header-padding)'
      }}
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
      className={cn('', className)} 
      style={{ padding: noPadding ? '0' : 'var(--card-body-padding)' }} 
      {...props} 
    />
  )
);

CardContent.displayName = 'CardContent';

const CardFooter = forwardRef<HTMLDivElement, HTMLAttributes<HTMLDivElement>>(
  ({ className, ...props }, ref) => (
    <div
      ref={ref}
      className={cn('flex items-center border-t border-border', className)}
      style={{ 
        paddingTop: 'var(--spacing-4)',
        marginTop: 'var(--spacing-4)',
        padding: 'var(--card-footer-padding)'
      }}
      {...props}
    />
  )
);

CardFooter.displayName = 'CardFooter';

export { Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter };