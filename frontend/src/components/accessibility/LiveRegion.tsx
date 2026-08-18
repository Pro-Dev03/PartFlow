import { clsx } from 'clsx'

type Politeness = 'polite' | 'assertive' | 'off'

interface LiveRegionProps {
  children: React.ReactNode
  politeness?: Politeness
  className?: string
}

/**
 * LiveRegion - Announces content changes to screen readers
 * Use for dynamic content updates, error messages, loading states, etc.
 */
export function LiveRegion({ children, politeness = 'polite', className }: LiveRegionProps) {
  return (
    <div
      role="status"
      aria-live={politeness}
      aria-atomic="true"
      className={clsx('sr-only', className)}
    >
      {children}
    </div>
  )
}
