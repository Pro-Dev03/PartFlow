import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'

import en from './locales/en.json'
import ar from './locales/ar.json'
import he from './locales/he.json'

i18n
  .use(initReactI18next)
  .init({
    resources: {
      en: { translation: en },
      ar: { translation: ar },
      he: { translation: he },
    },
    lng: 'ar',
    fallbackLng: 'en',
    interpolation: {
      escapeValue: false,
    },
    // RTL/LTR configuration
    react: {
      useSuspense: false,
    },
  })

// Set document direction based on language
const setDocumentDirection = (lng: string) => {
  const isRTL = lng === 'ar' || lng === 'he'
  document.documentElement.dir = isRTL ? 'rtl' : 'ltr'
  document.documentElement.lang = lng
}

// Set initial direction
setDocumentDirection(i18n.language)

// Update direction on language change
i18n.on('languageChanged', (lng) => {
  setDocumentDirection(lng)
})

export default i18n
