import { useTranslation as useI18nextTranslation } from 'react-i18next';
import { languages, type Language } from '../i18n/config';

export function useTranslation() {
  const { t, i18n } = useI18nextTranslation();
  
  const currentLanguage = i18n.language as Language;
  const languageConfig = languages.find(l => l.code === currentLanguage) || languages[0];
  
  const changeLanguage = (lang: Language) => {
    const newLanguageConfig = languages.find(l => l.code === lang) || languages[0];
    i18n.changeLanguage(lang);
    document.documentElement.dir = newLanguageConfig.direction;
    document.documentElement.lang = lang;
  };
  
  return {
    t,
    i18n,
    currentLanguage,
    direction: languageConfig.direction,
    languages,
    changeLanguage,
  };
}