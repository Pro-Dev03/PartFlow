import { clsx } from 'clsx'

interface VisuallyHiddenProps {
  children: React.ReactNode
  className?: string
}

/**
 * VisuallyHidden - Hides content visually but keeps it accessible for screen readers
 * Use this for screen-reader-only text like labels for icon-only buttons
 */
export function VisuallyHidden({ children, className }: VisuallyHiddenProps) {
  return (
    <span
      className={clsx(
        'sr-only absolute w-px h-px p-0 -m-px overflow-hidden whitespace-nowrap border-0',
        className
      )}
    >
      {children}
    </span>
  )
}
