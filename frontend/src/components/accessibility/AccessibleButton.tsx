import { ButtonHTMLAttributes, forwardRef } from 'react'
import { clsx } from 'clsx'
import { useFocusVisible } from './FocusVisible'
import { VisuallyHidden } from './VisuallyHidden'

interface AccessibleButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  icon?: string
  iconOnly?: boolean
  loading?: boolean
  ariaLabel?: string
}

/**
 * AccessibleButton - Button with proper accessibility features
 * - Always has an accessible name (either visible text or aria-label)
 * - Shows focus ring only for keyboard navigation
 * - Handles loading states
 * - Properly disabled
 */
export const AccessibleButton = forwardRef<HTMLButtonElement, AccessibleButtonProps>(
  ({ icon, iconOnly = false, loading = false, ariaLabel, children, disabled, className, ...props }, ref) => {
    const isFocusVisible = useFocusVisible()

    return (
      <button
        ref={ref}
        disabled={disabled || loading}
        aria-label={iconOnly ? ariaLabel || (typeof children === 'string' ? children : undefined) : undefined}
        aria-busy={loading}
        className={clsx(
          'inline-flex items-center justify-center gap-2 px-4 py-2 rounded-lg',
          'font-medium transition-colors',
          'focus:outline-none',
          isFocusVisible && 'focus:ring-2 focus:ring-primary focus:ring-offset-2',
          'disabled:opacity-50 disabled:cursor-not-allowed',
          loading && 'opacity-75 cursor-wait',
          className
        )}
        {...props}
      >
        {loading && (
          <span className="animate-spin" aria-hidden="true">
            ⏳
          </span>
        )}
        {icon && <span aria-hidden="true">{icon}</span>}
        {children}
        {iconOnly && !ariaLabel && typeof children === 'string' && (
          <VisuallyHidden>{children}</VisuallyHidden>
        )}
      </button>
    )
  }
)

AccessibleButton.displayName = 'AccessibleButton'
