import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import LanguageDetector from 'i18next-browser-languagedetector';

import ru from './locales/ru.json';
import en from './locales/en.json';

const resources = {
  ru: {
    translation: ru,
  },
  en: {
    translation: en,
  },
};

console.log('i18n resources loaded:', {
  ru: !!ru,
  en: !!en,
  ruMarketplace: !!ru?.mosaic_preview?.marketplace,
  enMarketplace: !!en?.mosaic_preview?.marketplace,
  ruReadyToBuy: ru?.mosaic_preview?.marketplace?.ready_to_buy,
  enReadyToBuy: en?.mosaic_preview?.marketplace?.ready_to_buy,
});

i18n
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    resources,
    fallbackLng: 'ru',
    debug: true,
    interpolation: {
      escapeValue: false,
    },

    detection: {
      order: ['localStorage', 'navigator', 'htmlTag'],
      caches: ['localStorage'],
    },

    missingKeyHandler: (lng, ns, key, fallbackValue) => {
      console.warn(`Missing translation key: ${key} for language: ${lng}`);
      return fallbackValue || key;
    },
  });

export default i18n;
