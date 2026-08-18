import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import LanguageDetector from 'i18next-browser-languagedetector';

import arTranslations from './locales/ar.json';
import enTranslations from './locales/en.json';

export const languages = [
  { code: 'ar', name: 'Arabic', nativeName: 'العربية', direction: 'rtl' as const, flag: '🇸🇦' },
  { code: 'en', name: 'English', nativeName: 'English', direction: 'ltr' as const, flag: '🇬🇧' },
] as const;

export type Language = typeof languages[number]['code'];

const resources = {
  ar: { translation: arTranslations },
  en: { translation: enTranslations },
};

i18n
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    resources,
    fallbackLng: 'ar',
    lng: 'ar',
    
    interpolation: {
      escapeValue: false,
    },
    
    detection: {
      order: ['localStorage', 'navigator'],
      caches: ['localStorage'],
    },
  }).then(() => {
    // Set initial direction based on language
    const currentLang = i18n.language as Language;
    const langConfig = languages.find(l => l.code === currentLang) || languages[0];
    document.documentElement.dir = langConfig.direction;
    document.documentElement.lang = currentLang;
  });

export default i18n;