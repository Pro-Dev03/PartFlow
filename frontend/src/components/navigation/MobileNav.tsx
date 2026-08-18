import { useTranslation } from 'react-i18next'
import { useLocation, useNavigate } from 'react-router-dom'
import { Home, ShoppingCart, Camera, Package, MoreHorizontal } from 'lucide-react'

interface NavItem {
  key: string
  icon: React.ReactNode
  path: string
}

export function MobileNav() {
  const { t } = useTranslation()
  const location = useLocation()
  const navigate = useNavigate()

  const menuItems: NavItem[] = [
    { key: 'dashboard', icon: <Home className="w-6 h-6" />, path: '/mobile' },
    { key: 'sales', icon: <ShoppingCart className="w-6 h-6" />, path: '/mobile/sales' },
    { key: 'scan', icon: <Camera className="w-6 h-6" />, path: '/mobile/scan' },
    { key: 'inventory', icon: <Package className="w-6 h-6" />, path: '/mobile/inventory' },
    { key: 'more', icon: <MoreHorizontal className="w-6 h-6" />, path: '/mobile/more' },
  ]

  const isActive = (path: string) => {
    if (path === '/mobile') {
      return location.pathname === '/mobile' || location.pathname === '/mobile/dashboard'
    }
    return location.pathname.startsWith(path)
  }

  const handleNavClick = (path: string) => {
    navigate(path)
  }

  return (
    <nav className="h-16 bg-surface border-t border-border flex items-center justify-around md:hidden">
      {menuItems.map((item) => (
        <button
          key={item.key}
          onClick={() => handleNavClick(item.path)}
          className={`flex flex-col items-center gap-1 p-2 flex-1 transition-colors ${
            isActive(item.path) ? 'text-accent' : 'text-text-faint'
          }`}
        >
          {item.icon}
          <span className="text-xs">{t(`nav.${item.key}`)}</span>
        </button>
      ))}
    </nav>
  )
}
