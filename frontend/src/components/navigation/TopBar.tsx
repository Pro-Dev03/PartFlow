import { useTranslation } from 'react-i18next'
import { useState } from 'react'
import { Input } from '../ui/Input'

export function TopBar() {
  const { t, i18n } = useTranslation()
  const [searchQuery, setSearchQuery] = useState('')
  const isRTL = i18n.dir() === 'rtl'

  const toggleLanguage = () => {
    const newLang = i18n.language === 'ar' ? 'en' : 'ar'
    i18n.changeLanguage(newLang)
  }

  return (
    <header className="h-16 bg-surface border-b border-border flex items-center justify-between px-6">
      {/* Search Bar */}
      <div className="flex-1 max-w-md">
        <Input
          placeholder={t('common.search')}
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          leftIcon="🔍"
          className="bg-background"
        />
      </div>

      {/* Right Actions */}
      <div className="flex items-center gap-3">
        {/* Language Toggle */}
        <button
          onClick={toggleLanguage}
          className="p-2 hover:bg-background rounded-lg transition-colors flex items-center gap-1"
          title="Change Language"
        >
          <span className="text-sm font-medium">{i18n.language.toUpperCase()}</span>
        </button>

        {/* Notifications */}
        <button className="p-2 hover:bg-background rounded-lg transition-colors relative">
          🔔
          <span className="absolute top-1 right-1 w-2 h-2 bg-danger rounded-full"></span>
        </button>

        {/* User Menu */}
        <button className="p-2 hover:bg-background rounded-lg transition-colors">
          👤
        </button>
      </div>
    </header>
  )
}
