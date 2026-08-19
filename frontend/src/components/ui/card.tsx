import type { HTMLAttributes } from 'react';
import { forwardRef } from 'react';
import { cn } from '../../lib/utils';

export interface CardProps extends HTMLAttributes<HTMLDivElement> {
  noPadding?: boolean;
  hoverable?: boolean;
  variant?: 'default' | 'interactive' | 'featured' | 'warning' | 'ai' | 'danger' | 'success' | 'info';
  fullWidth?: boolean;
}

const Card = forwardRef<HTMLDivElement, CardProps>(
  ({ 
    className, 
    noPadding = false, 
    hoverable = false, 
    variant = 'default', 
    fullWidth = false,
    children, 
    ...props 
  }, ref) => {
    const variants = {
      default: 'border border-border bg-card-gradient shadow-card',
      interactive: 'border border-border bg-card-gradient shadow-card hover:border-border/22 hover:-translate-y-2 cursor-pointer focus-visible:ring-2 focus-visible:ring-cyan focus-visible:ring-offset-2 focus-visible:ring-offset-bg',
      featured: 'border border-cyan/30 bg-card-gradient shadow-glow hover:border-cyan/50 hover:shadow-glow-strong backdrop-blur-xl',
      warning: 'border border-yellow/30 bg-yellow/5 hover:border-yellow/50',
      ai: 'border border-cyan/18 bg-card-ai-gradient hover:border-cyan/30',
      danger: 'border border-red/30 bg-red/5 hover:border-red/50',
      success: 'border border-green/30 bg-green/5 hover:border-green/50',
      info: 'border border-cyan/30 bg-cyan/5 hover:border-cyan/50',
    };
    
    const responsive = fullWidth ? 'w-full' : '';
    
    return (
      <div
        ref={ref}
        className={cn(
          'rounded-md transition-all duration-slow',
          variants[variant],
          !noPadding && 'p-lg',
          responsive,
          hoverable && variant === 'default' && 'hover:border-border/22 hover:-translate-y-2 cursor-pointer',
          className
        )}
        tabIndex={hoverable || variant === 'interactive' ? 0 : undefined}
        role={hoverable || variant === 'interactive' ? 'button' : undefined}
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
      className={cn('flex justify-between items-center mb-5', className)}
      {...props}
    />
  )
);

CardHeader.displayName = 'CardHeader';

const CardTitle = forwardRef<HTMLParagraphElement, HTMLAttributes<HTMLHeadingElement>>(
  ({ className, ...props }, ref) => (
    <h3
      ref={ref}
      className={cn('text-card-title font-semibold leading-none text-text', className)}
      {...props}
    />
  )
);

CardTitle.displayName = 'CardTitle';

const CardDescription = forwardRef<HTMLParagraphElement, HTMLAttributes<HTMLParagraphElement>>(
  ({ className, ...props }, ref) => (
    <p
      ref={ref}
      className={cn('text-small text-text-muted', className)}
      {...props}
    />
  )
);

CardDescription.displayName = 'CardDescription';

const CardContent = forwardRef<HTMLDivElement, HTMLAttributes<HTMLDivElement>>(
  ({ className, ...props }, ref) => (
    <div ref={ref} className={cn('', className)} {...props} />
  )
);

CardContent.displayName = 'CardContent';

const CardFooter = forwardRef<HTMLDivElement, HTMLAttributes<HTMLDivElement>>(
  ({ className, ...props }, ref) => (
    <div
      ref={ref}
      className={cn('flex items-center pt-4 mt-4 border-t border-border', className)}
      {...props}
    />
  )
);

CardFooter.displayName = 'CardFooter';

export { Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter };