import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'
import ru from './locales/ru.json'
import en from './locales/en.json'

const resources = {
  ru: { translation: ru },
  en: { translation: en },
}

i18n
  .use(initReactI18next)
  .init({
    resources,
    lng: localStorage.getItem('dashboard_language') || 'ru',
    fallbackLng: 'ru',
    
    interpolation: {
      escapeValue: false,
    },

    react: {
      useSuspense: false,
    },
  })

export default i18n
