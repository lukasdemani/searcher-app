import i18n from 'i18next';
import LanguageDetector from 'i18next-browser-languagedetector';
import { initReactI18next } from 'react-i18next';
import deTranslations from './locales/de.json';
import enTranslations from './locales/en.json';

const resources = {
  en: {
    translation: enTranslations,
  },
  de: {
    translation: deTranslations,
  },
};

i18n
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    resources,
    fallbackLng: 'en',
    lng: 'en',
    debug: process.env.NODE_ENV === 'development',
    interpolation: {
      escapeValue: false,
    },

    detection: {
      order: ['localStorage', 'navigator', 'htmlTag'],
      lookupLocalStorage: 'i18nextLng',
      caches: ['localStorage'],
    },

    supportedLngs: ['en', 'de'],
    nonExplicitSupportedLngs: true,
    cleanCode: true,
    
    load: 'languageOnly',
  });

if (i18n.language && !['en', 'de'].includes(i18n.language.split('-')[0])) {
  i18n.changeLanguage('en');
  localStorage.removeItem('i18nextLng');
}

export default i18n;
