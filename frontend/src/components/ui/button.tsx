import type { ButtonHTMLAttributes } from 'react';
import { forwardRef } from 'react';
import { cn } from '../../lib/utils';

export interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'ghost' | 'danger' | 'success' | 'outline' | 'default' | 'warning' | 'info';
  size?: 'sm' | 'md' | 'lg' | 'icon';
  isLoading?: boolean;
  fullWidth?: boolean;
}

const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  ({ 
    className, 
    variant = 'primary', 
    size = 'md', 
    isLoading = false, 
    disabled = false,
    fullWidth = false,
    children, 
    ...props 
  }, ref) => {
    const isDisabled = disabled || isLoading;
    
    const baseStyles = 'inline-flex items-center justify-center rounded-sm font-medium transition-all duration-normal focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-cyan focus-visible:ring-offset-2 focus-visible:ring-offset-bg disabled:pointer-events-none disabled:opacity-50 disabled:cursor-not-allowed active:scale-95';
    
    const variants = {
      primary: 'border border-cyan/35 bg-button-primary-gradient text-cyan-100 hover:-translate-y-px hover:border-cyan/30 hover:shadow-glow active:translate-y-0 active:shadow-none',
      secondary: 'border border-border bg-surface/75 text-blue-100 hover:-translate-y-px hover:border-cyan/30 active:translate-y-0',
      ghost: 'text-text hover:bg-surface/50 active:bg-surface/75',
      danger: 'bg-red text-white hover:bg-red/90 active:bg-red/80',
      success: 'bg-green text-white hover:bg-green/90 active:bg-green/80',
      outline: 'border-2 border-border bg-transparent text-text hover:bg-surface/50 active:bg-surface/75',
      default: 'border border-border bg-surface text-text hover:border-cyan/30',
      warning: 'bg-yellow text-white hover:bg-yellow/90 active:bg-yellow/80',
      info: 'bg-cyan text-white hover:bg-cyan/90 active:bg-cyan/80',
    };
    
    const sizes = {
      sm: 'h-8 px-3 text-small',
      md: 'h-10 px-4 text-body',
      lg: 'h-12 px-6 text-small',
      icon: 'h-10 w-10 p-0',
    };
    
    const responsive = fullWidth ? 'w-full' : '';
    
    return (
      <button
        ref={ref}
        className={cn(
          baseStyles, 
          variants[variant], 
          sizes[size], 
          responsive,
          className
        )}
        disabled={isDisabled}
        aria-disabled={isDisabled}
        aria-busy={isLoading}
        {...props}
      >
        {isLoading && (
          <svg
            className="animate-spin -me-2 h-4 w-4"
            xmlns="http://www.w3.org/2000/svg"
            fill="none"
            viewBox="0 0 24 24"
            aria-hidden="true"
          >
            <circle
              className="opacity-25"
              cx="12"
              cy="12"
              r="10"
              stroke="currentColor"
              strokeWidth="4"
            />
            <path
              className="opacity-75"
              fill="currentColor"
              d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
            />
          </svg>
        )}
        <span className={isLoading ? 'opacity-50' : ''}>{children}</span>
      </button>
    );
  }
);

Button.displayName = 'Button';

export { Button };