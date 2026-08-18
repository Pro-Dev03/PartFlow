import { clsx } from 'clsx'

interface NavItem {
  id: string
  label: string
  icon: string
  path: string
}

interface MobileBottomNavProps {
  items: NavItem[]
  activeItem: string
  onNavigate: (path: string) => void
  className?: string
}

export function MobileBottomNav({ items, activeItem, onNavigate, className }: MobileBottomNavProps) {
  return (
    <nav className={clsx('fixed bottom-0 left-0 right-0 bg-surface border-t border-border z-40', className)}>
      <div className="flex items-center justify-around py-2">
        {items.map((item) => (
          <button
            key={item.id}
            onClick={() => onNavigate(item.path)}
            className={clsx(
              'flex flex-col items-center gap-1 px-4 py-2 rounded-lg transition-colors min-w-0',
              activeItem === item.id
                ? 'text-primary'
                : 'text-muted hover:text-text'
            )}
          >
            <span className="text-2xl">{item.icon}</span>
            <span className="text-xs font-medium truncate">{item.label}</span>
          </button>
        ))}
      </div>
    </nav>
  )
}
