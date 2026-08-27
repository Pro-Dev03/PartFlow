import type { ButtonHTMLAttributes } from 'react';
import { forwardRef } from 'react';
import { cn } from '../../utils';

export interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'ghost' | 'danger' | 'success' | 'outline' | 'default' | 'warning' | 'info';
  size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl' | 'icon';
  isLoading?: boolean;
  loading?: boolean;
  fullWidth?: boolean;
}

const DEFAULT_SIZE: 'md' = 'md';

const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  ({
    className,
    variant = 'primary',
    size = DEFAULT_SIZE,
    isLoading = false,
    loading = false,
    disabled = false,
    fullWidth = false,
    children,
    'aria-label': ariaLabel,
    'aria-describedby': ariaDescribedby,
    ...props
  }, ref) => {
    const isDisabled = disabled || isLoading || loading;
    const isActuallyLoading = isLoading || loading;
    
    const baseClasses = [
      'inline-flex',
      'items-center',
      'justify-center',
      'gap-2',
      'font-semibold',
      'rounded-[10px]',
      'transition-all',
      'duration-150',
      'ease-out',
      'focus:outline-none',
      'focus:ring-2',
      'focus:ring-offset-2',
      'focus-visible:ring-primary',
      'focus-visible:ring-offset-2',
      'disabled:opacity-50',
      'disabled:cursor-not-allowed',
    ];

    const variantClasses: Record<string, string> = {
      primary: [
        'bg-primary',
        'text-white',
        'border',
        'border-primary/20',
        'hover:bg-primary/90',
        'focus:ring-primary',
      ].join(' '),
      secondary: [
        'bg-surface-elevated',
        'text-text-primary',
        'border',
        'border-border',
        'hover:bg-surface',
        'hover:text-text-primary',
        'focus:ring-primary',
      ].join(' '),
      ghost: [
        'bg-transparent',
        'text-text-primary',
        'hover:bg-surface',
        'hover:text-text-primary',
        'border-transparent',
      ].join(' '),
      danger: [
        'bg-danger/10',
        'text-danger',
        'border',
        'border-danger/20',
        'hover:bg-danger/20',
        'focus:ring-danger',
      ].join(' '),
      success: [
        'bg-success/10',
        'text-success',
        'border',
        'border-success/20',
        'hover:bg-success/20',
        'focus:ring-success',
      ].join(' '),
      outline: [
        'bg-transparent',
        'text-text-primary',
        'border',
        'border-primary/20',
        'hover:bg-primary/5',
        'focus:ring-primary',
      ].join(' '),
      default: [
        'bg-surface',
        'text-text-primary',
        'border',
        'border-border',
        'hover:bg-surface/80',
      ].join(' '),
      warning: [
        'bg-warning/10',
        'text-warning',
        'border',
        'border-warning/20',
        'hover:bg-warning/20',
        'focus:ring-warning',
      ].join(' '),
      info: [
        'bg-info/10',
        'text-info',
        'border',
        'border-info/20',
        'hover:bg-info/20',
        'focus:ring-info',
      ].join(' '),
    };

    const sizeClasses: Record<string, string> = {
      xs: 'h-[var(--button-height-xs)] px-[var(--button-padding-xs)] text-[var(--button-font-size-xs)]',
      sm: 'h-[var(--button-height-sm)] px-[var(--button-padding-sm)] text-[var(--button-font-size-sm)]',
      md: 'h-[var(--button-height-md)] px-[var(--button-padding-md)] text-[var(--button-font-size-md)]',
      lg: 'h-[var(--button-height-lg)] px-[var(--button-padding-lg)] text-[var(--button-font-size-lg)]',
      xl: 'h-[var(--button-height-xl)] px-[var(--button-padding-xl)] text-[var(--button-font-size-xl)]',
      icon: 'h-10 w-10 p-0',
    };

    return (
      <button
        ref={ref}
        className={cn(
          ...baseClasses,
          variantClasses[variant] || variantClasses.default,
          sizeClasses[size] || sizeClasses.sm,
          'select-none',
          'pf-button',
          `pf-button-${variant}`,
          fullWidth && 'w-full',
          className
        )}
        disabled={isDisabled}
        aria-disabled={isDisabled}
        aria-busy={isActuallyLoading}
        aria-label={ariaLabel || (typeof children === 'string' ? children : undefined)}
        aria-describedby={ariaDescribedby}
        {...props}
      >
        {isActuallyLoading && (
          <svg
            className="animate-spin h-4 w-4"
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
        {isActuallyLoading ? (
          <span className="transition-opacity opacity-50">{children}</span>
        ) : (
          children
        )}
      </button>
    );
  }
);

Button.displayName = 'Button';

export { Button };
export default Button;
