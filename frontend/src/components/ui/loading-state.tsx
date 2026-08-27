import { forwardRef } from 'react';
import { cn } from '../../utils';

export interface LoadingStateProps {
  message?: string;
  size?: 'sm' | 'md' | 'lg';
  className?: string;
}

const LoadingState = forwardRef<HTMLDivElement, LoadingStateProps>(
  ({ 
    message = 'جاري التحميل...', 
    size = 'md',
    className
  }, ref) => {
    const sizeStyles = {
      sm: {
        minHeight: 'var(--empty-state-min-height)',
        spinnerSize: 'var(--loading-spinner-size)',
        fontSize: 'var(--font-size-secondary)'
      },
      md: {
        minHeight: 'var(--empty-state-min-height)',
        spinnerSize: 'calc(var(--loading-spinner-size) * 1.5)',
        fontSize: 'var(--font-size-body)'
      },
      lg: {
        minHeight: 'calc(var(--empty-state-min-height) * 1.5)',
        spinnerSize: 'calc(var(--loading-spinner-size) * 2)',
        fontSize: 'var(--font-size-section-title)'
      }
    };

    const currentSize = sizeStyles[size];

    return (
      <div
        ref={ref}
        className={cn(
          'flex flex-col items-center justify-center text-center',
          className
        )}
        style={{
          minHeight: currentSize.minHeight,
          padding: 'var(--spacing-8)'
        }}
      >
        <div 
          className="animate-spin rounded-full border-2 border-text-muted/20 border-t-primary mb-4"
          style={{
            width: currentSize.spinnerSize,
            height: currentSize.spinnerSize,
            borderWidth: 'var(--loading-spinner-border-width)'
          }}
        />
        <p 
          className="text-text-muted"
          style={{ fontSize: currentSize.fontSize }}
        >
          {message}
        </p>
      </div>
    );
  }
);

LoadingState.displayName = 'LoadingState';

export { LoadingState };