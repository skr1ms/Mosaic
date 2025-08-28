import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';

const LanguageSelector = () => {
  const { i18n } = useTranslation();
  const [isOpen, setIsOpen] = useState(false);

  const languages = [
    { code: 'ru', name: 'RU', flag: '🇷🇺' },
    { code: 'en', name: 'EN', flag: '🇺🇸' },
  ];

  const currentLanguage = languages.find(lang => lang.code === i18n.language) || languages[0];

  const handleLanguageChange = (langCode) => {
    i18n.changeLanguage(langCode);
    localStorage.setItem('dashboard_language', langCode);
    setIsOpen(false);
  };

  return (
    <div className="relative">
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="btn btn-outline-secondary"
        style={{ marginRight: '15px' }}
      >
        <span style={{ marginRight: '8px' }}>{currentLanguage.flag}</span>
        <span>{currentLanguage.name}</span>
        <i className={`pe-7s-angle-down ${isOpen ? 'rotate-180' : ''}`} style={{ marginLeft: '8px', transition: 'transform 0.2s' }}></i>
      </button>

      {isOpen && (
        <>
          <div 
            className="fixed inset-0 z-10" 
            onClick={() => setIsOpen(false)}
          />
          <div className="dropdown-menu show" style={{ position: 'absolute', right: 0, top: '100%', marginTop: '5px', minWidth: '120px', zIndex: 1000 }}>
            {languages.map((language) => (
              <button
                key={language.code}
                onClick={() => handleLanguageChange(language.code)}
                className="dropdown-item"
                style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', width: '100%', padding: '8px 16px', border: 'none', background: 'none', textAlign: 'left' }}
              >
                <div style={{ display: 'flex', alignItems: 'center' }}>
                  <span style={{ marginRight: '8px' }}>{language.flag}</span>
                  <span>{language.name}</span>
                </div>
                {language.code === i18n.language && (
                  <i className="pe-7s-check text-primary"></i>
                )}
              </button>
            ))}
          </div>
        </>
      )}
    </div>
  );
};

export default LanguageSelector;
