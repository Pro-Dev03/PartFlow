import { clsx } from 'clsx'

interface SkipLinkProps {
  targetId: string
  children?: React.ReactNode
  className?: string
}

/**
 * SkipLink - Allows keyboard users to skip navigation and jump to main content
 * Place this at the top of your app, it becomes visible on focus
 */
export function SkipLink({ targetId, children = 'انتقل للمحتوى الرئيسي', className }: SkipLinkProps) {
  return (
    <a
      href={`#${targetId}`}
      className={clsx(
        'sr-only focus:not-sr-only focus:absolute focus:top-4 focus:left-4 focus:z-50',
        'focus:px-4 focus:py-2 focus:bg-primary focus:text-white focus:rounded-lg',
        'focus:shadow-lg focus:outline-none',
        className
      )}
    >
      {children}
    </a>
  )
}
