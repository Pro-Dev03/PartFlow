import { useTheme } from './ThemeProvider'
import { clsx } from 'clsx'
import { Sun, Moon, Monitor } from 'lucide-react'

export function ThemeToggle() {
  const { theme, setTheme } = useTheme()

  const themes: { value: 'light' | 'dark' | 'system'; label: string; icon: React.ReactNode }[] = [
    { value: 'light', label: 'فاتح', icon: <Sun className="w-5 h-5" /> },
    { value: 'dark', label: 'داكن', icon: <Moon className="w-5 h-5" /> },
    { value: 'system', label: 'تلقائي', icon: <Monitor className="w-5 h-5" /> },
  ]

  return (
    <div className="flex items-center gap-2">
      {themes.map((t) => (
        <button
          key={t.value}
          onClick={() => setTheme(t.value)}
          className={clsx(
            'w-10 h-10 rounded-lg flex items-center justify-center transition-all',
            'hover:bg-surface-elevated dark:hover:bg-white/10',
            theme === t.value
              ? 'bg-seal-soft text-seal dark:bg-seal/20 dark:text-seal-dark ring-2 ring-seal dark:ring-seal-dark'
              : 'text-text-secondary dark:text-text-dim'
          )}
          title={t.label}
        >
          {t.icon}
        </button>
      ))}
    </div>
  )
}
