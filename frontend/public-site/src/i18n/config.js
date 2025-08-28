import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'
import LanguageDetector from 'i18next-browser-languagedetector'

// Динамический импорт локалей для совместимости с Vite
const resources = {
  ru: {
    translation: {}
  },
  en: {
    translation: {}
  }
}

// Функция для загрузки локалей
const loadLocales = async () => {
  try {
    const [ru, en] = await Promise.all([
      import('./locales/ru.json'),
      import('./locales/en.json')
    ])
    
    resources.ru.translation = ru.default || ru
    resources.en.translation = en.default || en
    
    // Переинициализируем i18n с загруженными ресурсами
    i18n.changeLanguage(i18n.language)
  } catch (error) {
    console.error('Failed to load locales:', error)
  }
}

i18n
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    resources,
    fallbackLng: 'ru',
    // В development используем переменную из Docker Compose, в production - false
    debug: import.meta.env.VITE_DEBUG === 'true',
    
    interpolation: {
      escapeValue: false
    },
    
    detection: {
      order: ['localStorage', 'navigator', 'htmlTag'],
      caches: ['localStorage']
    }
  })

// Загружаем локали после инициализации
loadLocales()

export default i18n
