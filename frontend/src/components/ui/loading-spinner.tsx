import type { HTMLAttributes } from 'react';
import { forwardRef } from 'react';
import { cn } from '../../utils';

export interface LoadingSpinnerProps extends HTMLAttributes<HTMLDivElement> {
  size?: 'sm' | 'md' | 'lg';
  color?: 'cyan' | 'white' | 'text';
  'aria-label'?: string;
  'aria-live'?: 'polite' | 'assertive' | 'off';
}

const LoadingSpinner = forwardRef<HTMLDivElement, LoadingSpinnerProps>(
  ({ className, size = 'md', color = 'cyan', 'aria-label': ariaLabel, 'aria-live': ariaLive = 'polite', ...props }, ref) => {
  const sizes = {
    sm: 'w-4 h-4 border-2',
    md: 'w-5 h-5 border-2',
    lg: 'w-12 h-12 border-3',
  };
  
  const colors = {
    cyan: 'border-cyan border-t-transparent',
    white: 'border-white border-t-transparent',
    text: 'border-text border-t-transparent',
  };
  
  return (
    <div
      ref={ref}
      className={cn(
        'animate-spin rounded-full',
        sizes[size],
        colors[color],
        className
      )}
      role="status"
      aria-label={ariaLabel || 'جاري التحميل'}
      aria-live={ariaLive}
      aria-busy="true"
      {...props}
    />
  );
});

LoadingSpinner.displayName = 'LoadingSpinner';

export { LoadingSpinner };