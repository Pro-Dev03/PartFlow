import { useEffect } from 'react'
import { useTranslation } from 'react-i18next'

interface DirectionProviderProps {
  children: React.ReactNode
}

export function DirectionProvider({ children }: DirectionProviderProps) {
  const { i18n } = useTranslation()

  useEffect(() => {
    // Set document direction based on current language
    const setDirection = (lang: string) => {
      const rtlLanguages = ['ar', 'he', 'fa', 'ur']
      const direction = rtlLanguages.includes(lang) ? 'rtl' : 'ltr'
      
      document.documentElement.dir = direction
      document.documentElement.lang = lang
    }

    // Set initial direction
    setDirection(i18n.language)

    // Listen for language changes
    const handleLanguageChange = (lng: string) => {
      setDirection(lng)
    }

    i18n.on('languageChanged', handleLanguageChange)

    // Cleanup
    return () => {
      i18n.off('languageChanged', handleLanguageChange)
    }
  }, [i18n])

  return <>{children}</>
}