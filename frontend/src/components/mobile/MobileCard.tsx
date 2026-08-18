import { clsx } from 'clsx'

interface MobileCardProps {
  title: string
  subtitle?: string
  icon?: string
  children: React.ReactNode
  onClick?: () => void
  className?: string
}

export function MobileCard({ title, subtitle, icon, children, onClick, className }: MobileCardProps) {
  return (
    <div
      onClick={onClick}
      className={clsx(
        'bg-surface rounded-lg border border-border p-4 active:bg-muted-10 transition-colors',
        onClick && 'cursor-pointer',
        className
      )}
    >
      {/* Header */}
      <div className="flex items-start gap-3 mb-3">
        {icon && (
          <div className="text-2xl flex-shrink-0">{icon}</div>
        )}
        <div className="flex-1 min-w-0">
          <h3 className="font-semibold text-text truncate">{title}</h3>
          {subtitle && (
            <p className="text-sm text-muted truncate">{subtitle}</p>
          )}
        </div>
      </div>

      {/* Content */}
      <div className="space-y-2">
        {children}
      </div>
    </div>
  )
}
