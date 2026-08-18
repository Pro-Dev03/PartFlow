import { useTranslation } from 'react-i18next'
import { useLocation, Link } from 'react-router-dom'
import { clsx } from 'clsx'

export function Sidebar() {
  const { t, i18n } = useTranslation()
  const location = useLocation()
  const isRTL = i18n.dir() === 'rtl'

  const menuItems = [
    { key: 'dashboard', icon: '📊', path: '/' },
    { key: 'sales', icon: '💰', path: '/sales' },
    { key: 'inventory', icon: '📦', path: '/inventory' },
    { key: 'customers', icon: '👥', path: '/customers' },
    { key: 'debts', icon: '💳', path: '/debts' },
    { key: 'purchases', icon: '🚚', path: '/purchases' },
    { key: 'suppliers', icon: '🏭', path: '/suppliers' },
    { key: 'expenses', icon: '💸', path: '/expenses' },
    { key: 'reports', icon: '📊', path: '/reports' },
    { key: 'settings', icon: '⚙️', path: '/settings' },
  ]

  const isActive = (path: string) => {
    if (path === '/') {
      return location.pathname === '/'
    }
    return location.pathname.startsWith(path)
  }

  return (
    <aside className={clsx(
      "w-64 bg-surface border border-border flex flex-col",
      isRTL ? "border-l" : "border-r"
    )}>
      {/* Logo Section */}
      <div className="p-4 border-b border-border">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-lg bg-primary flex items-center justify-center">
            <span className="text-white text-xl">🏪</span>
          </div>
          <div>
            <h1 className="text-lg font-bold text-primary">{t('app.name')}</h1>
            <p className="text-xs text-muted">{t('app.description')}</p>
          </div>
        </div>
      </div>

      {/* Navigation */}
      <nav className="flex-1 p-4 overflow-y-auto scrollbar-thin">
        <ul className="space-y-1">
          {menuItems.map((item) => (
            <li key={item.key}>
              <Link
                to={item.path}
                className={clsx(
                  "flex items-center gap-3 px-4 py-2.5 rounded-lg transition-all duration-150",
                  isActive(item.path)
                    ? "bg-primary text-white shadow-sm"
                    : "text-text hover:bg-background hover:text-primary"
                )}
              >
                <span className="text-lg">{item.icon}</span>
                <span className="font-medium">{t(`nav.${item.key}`)}</span>
              </Link>
            </li>
          ))}
        </ul>
      </nav>

      {/* User Section */}
      <div className="p-4 border-t border-border">
        <div className="flex items-center gap-3">
          <div className="w-8 h-8 rounded-full bg-primary-100 flex items-center justify-center">
            <span className="text-primary text-sm">👤</span>
          </div>
          <div className="flex-1 min-w-0">
            <p className="text-sm font-medium text-text truncate">المستخدم</p>
            <p className="text-xs text-muted truncate">user@example.com</p>
          </div>
        </div>
      </div>
    </aside>
  )
}
