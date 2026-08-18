import { useTranslation } from 'react-i18next'

export function MobileNav() {
  const { t } = useTranslation()

  const menuItems = [
    { key: 'dashboard', icon: '🏠' },
    { key: 'sales', icon: '💰' },
    { key: 'scan', icon: '📷' },
    { key: 'inventory', icon: '📦' },
    { key: 'more', icon: '⋯' },
  ]

  return (
    <nav className="h-16 bg-surface border-t border-border flex items-center justify-around">
      {menuItems.map((item) => (
        <a
          key={item.key}
          href={`/${item.key}`}
          className="flex flex-col items-center gap-1 p-2"
        >
          <span className="text-xl">{item.icon}</span>
          <span className="text-xs">{t(`nav.${item.key}`)}</span>
        </a>
      ))}
    </nav>
  )
}
