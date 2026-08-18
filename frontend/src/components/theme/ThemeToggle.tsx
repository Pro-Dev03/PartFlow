import { useTheme } from './ThemeProvider'
import { clsx } from 'clsx'

export function ThemeToggle() {
  const { theme, setTheme, actualTheme } = useTheme()

  const themes: { value: 'light' | 'dark' | 'system'; label: string; icon: string }[] = [
    { value: 'light', label: 'فاتح', icon: '☀️' },
    { value: 'dark', label: 'داكن', icon: '🌙' },
    { value: 'system', label: 'تلقائي', icon: '💻' },
  ]

  return (
    <div className="flex items-center gap-2">
      {themes.map((t) => (
        <button
          key={t.value}
          onClick={() => setTheme(t.value)}
          className={clsx(
            'w-10 h-10 rounded-lg flex items-center justify-center text-xl transition-all',
            'hover:bg-muted-10',
            theme === t.value
              ? 'bg-primary-10 text-primary ring-2 ring-primary-30'
              : 'text-muted'
          )}
          title={t.label}
        >
          {t.icon}
        </button>
      ))}
    </div>
  )
}
