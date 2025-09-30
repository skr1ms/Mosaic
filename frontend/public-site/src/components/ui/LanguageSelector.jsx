import React from 'react';
import { useTranslation } from 'react-i18next';
import { Globe } from 'lucide-react';

const LanguageSelector = () => {
  const { i18n, t } = useTranslation();

  const languages = [
    { code: 'ru', name: t('languages.russian'), flag: '🇷🇺' },
    { code: 'en', name: t('languages.english'), flag: '🇺🇸' },
  ];

  const currentLanguage =
    languages.find(lang => lang.code === i18n.language) || languages[0];

  const handleLanguageChange = languageCode => {
    i18n.changeLanguage(languageCode);
    try {
      localStorage.setItem('language', languageCode);
    } catch {}
  };

  return (
    <div className="relative group">
      <button className="flex items-center space-x-1 sm:space-x-2 px-2 sm:px-3 py-2 text-gray-700 hover:text-gray-900 active:text-gray-600 transition-colors touch-target">
        <Globe className="w-4 h-4 flex-shrink-0" />
        <span className="text-sm font-medium hidden sm:inline">
          {currentLanguage.flag} {currentLanguage.name}
        </span>
        <span className="text-sm font-medium sm:hidden">
          {currentLanguage.flag}
        </span>
      </button>

      <div className="absolute right-0 mt-2 w-32 sm:w-48 bg-white rounded-md shadow-lg border border-gray-200 opacity-0 invisible group-hover:opacity-100 group-hover:visible transition-all duration-200 z-50 max-w-[calc(100vw-2rem)]">
        {languages.map(language => (
          <button
            key={language.code}
            onClick={() => handleLanguageChange(language.code)}
            className={`w-full text-left px-3 sm:px-4 py-2 text-sm hover:bg-gray-100 active:bg-gray-200 transition-colors touch-target ${
              i18n.language === language.code
                ? 'bg-brand-primary/10 text-brand-primary'
                : 'text-gray-700'
            }`}
          >
            <span className="mr-1 sm:mr-2">{language.flag}</span>
            <span className="hidden sm:inline">{language.name}</span>
            <span className="sm:hidden text-xs">
              {language.code.toUpperCase()}
            </span>
          </button>
        ))}
      </div>
    </div>
  );
};

export default LanguageSelector;
