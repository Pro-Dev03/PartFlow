import { useState } from 'react'
import { clsx } from 'clsx'

interface ProgressiveDisclosureProps {
  children: React.ReactNode
  defaultOpen?: boolean
  label: string
  optional?: boolean
  className?: string
}

export function ProgressiveDisclosure({
  children,
  defaultOpen = false,
  label,
  optional = false,
  className,
}: ProgressiveDisclosureProps) {
  const [isOpen, setIsOpen] = useState(defaultOpen)

  return (
    <div className={clsx('border border-border rounded-lg overflow-hidden', className)}>
      <button
        type="button"
        onClick={() => setIsOpen(!isOpen)}
        className="w-full px-4 py-3 flex items-center justify-between bg-surface hover:bg-surface-80 transition-colors"
      >
        <span className="font-medium text-text">
          {label}
          {optional && <span className="text-muted text-sm font-normal ml-2">(اختياري)</span>}
        </span>
        <svg
          className={clsx('w-5 h-5 text-muted transition-transform', isOpen && 'rotate-180')}
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
        </svg>
      </button>
      {isOpen && <div className="p-4 border-t border-border">{children}</div>}
    </div>
  )
}
