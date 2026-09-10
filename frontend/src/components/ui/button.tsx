import type { ButtonHTMLAttributes } from 'react';
import { forwardRef } from 'react';
import { cn } from '../../utils';

export interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'ghost' | 'danger' | 'success' | 'outline' | 'default' | 'warning' | 'info';
  size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl' | 'icon';
  isLoading?: boolean;
  loading?: boolean;
  fullWidth?: boolean;
  tableAction?: boolean;
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
    tableAction = false,
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
      'whitespace-nowrap',
      'font-semibold',
      'tracking-[0.01em]',
      'rounded-[8px]',
      'border',
      'border-transparent',
      'transition-all',
      'duration-200',
      'ease-out',
      'focus-visible:outline-none',
      'focus-visible:ring-2',
      'focus-visible:ring-offset-2',
      'focus-visible:ring-primary',
      'focus-visible:ring-offset-[var(--bg-surface)]',
      'disabled:opacity-50',
      'disabled:cursor-not-allowed',
      'disabled:pointer-events-none',
      'shrink-0',
    ];

    const variantClasses: Record<string, string> = {
      primary: [
        'bg-primary',
        'text-white',
        'border-primary/40',
        'shadow-[0_8px_18px_rgba(37,99,235,0.18)]',
        'hover:bg-primary/90',
        'hover:shadow-[0_10px_22px_rgba(37,99,235,0.22)]',
        'active:bg-primary/80',
        'focus-visible:ring-primary',
      ].join(' '),
      secondary: [
        'bg-surface-elevated',
        'text-text-primary',
        'border-border',
        'shadow-[0_1px_2px_rgba(15,23,42,0.08)]',
        'hover:bg-surface',
        'hover:border-primary/30',
        'hover:text-text-primary',
        'focus-visible:ring-primary',
      ].join(' '),
      ghost: [
        'bg-primary/5',
        'text-primary',
        'border-transparent',
        'hover:bg-primary/10',
        'hover:text-primary',
        'focus-visible:ring-primary',
      ].join(' '),
      danger: [
        'bg-danger/12',
        'text-danger',
        'border-danger/30',
        'hover:bg-danger/20',
        'focus-visible:ring-danger',
      ].join(' '),
      success: [
        'bg-success/12',
        'text-success',
        'border-success/30',
        'hover:bg-success/20',
        'focus-visible:ring-success',
      ].join(' '),
      outline: [
        'bg-transparent',
        'text-primary',
        'border-primary/30',
        'hover:bg-primary/5',
        'hover:text-primary',
        'focus-visible:ring-primary',
      ].join(' '),
      default: [
        'bg-surface',
        'text-text-primary',
        'border-border',
        'hover:bg-surface/80',
        'hover:border-primary/20',
      ].join(' '),
      warning: [
        'bg-warning/12',
        'text-warning',
        'border-warning/30',
        'hover:bg-warning/20',
        'focus-visible:ring-warning',
      ].join(' '),
      info: [
        'bg-info/12',
        'text-info',
        'border-info/30',
        'hover:bg-info/20',
        'focus-visible:ring-info',
      ].join(' '),
    };

    const sizeClasses: Record<string, string> = {
      xs: 'h-[var(--button-height-xs)] px-[var(--button-padding-xs)] text-[var(--button-font-size-xs)]',
      sm: 'h-[var(--button-height-sm)] px-[var(--button-padding-sm)] text-[var(--button-font-size-sm)]',
      md: 'h-[var(--button-height-md)] px-[var(--button-padding-md)] text-[var(--button-font-size-md)]',
      lg: 'h-[var(--button-height-lg)] px-[var(--button-padding-lg)] text-[var(--button-font-size-lg)]',
      xl: 'h-[var(--button-height-xl)] px-[var(--button-padding-xl)] text-[var(--button-font-size-xl)]',
      icon: 'h-5 w-5 min-h-5 min-w-5 p-0 gap-0 text-[0.7rem] leading-none',
    };

    const isAutoTableAction = (size === 'icon' || (size === 'sm' && ['ghost', 'outline', 'danger'].includes(variant))) && !tableAction;
    const tableActionClasses = tableAction || isAutoTableAction
      ? 'pf-table-action-btn h-8 min-h-8 w-8 min-w-8 p-0 gap-0 rounded-[9px] shrink-0 leading-none'
      : '';

    return (
      <button
        ref={ref}
        className={cn(
          ...baseClasses,
          variantClasses[variant] || variantClasses.default,
          sizeClasses[size] || sizeClasses.sm,
          tableActionClasses,
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
