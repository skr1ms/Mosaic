import React, { useState } from 'react';
import { Link, useLocation } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { Menu, X, ChevronDown } from 'lucide-react';
import { usePartnerStore } from '../../store/partnerStore';
import LanguageSelector from '../ui/LanguageSelector';

const Header = () => {
  const { t } = useTranslation();
  const location = useLocation();
  const { partner } = usePartnerStore();
  const [isMenuOpen, setIsMenuOpen] = useState(false);

  const scrollToSection = sectionId => {
    if (location.pathname !== '/') {
      window.location.href = `/#${sectionId}`;
      return;
    }

    const element = document.getElementById(sectionId);
    if (element) {
      element.scrollIntoView({
        behavior: 'smooth',
        block: 'start',
        inline: 'nearest',
      });
    }
  };

  const navigation = [
    {
      name: t('navigation.home'),
      href: '/',
      current: location.pathname === '/',
      onClick: null,
    },
    {
      name: t('navigation.diamond_art'),
      href: '/diamond-art',
      current: location.pathname === '/diamond-art',
      onClick: null,
    },
    {
      name: t('navigation.paint_by_numbers'),
      href: '#paint-by-numbers',
      current: false,
      onClick: e => {
        e.preventDefault();
        scrollToSection('paint-by-numbers');
      },
    },
    {
      name: t('navigation.what_is_this'),
      href: '#faq',
      current: false,
      onClick: e => {
        e.preventDefault();
        scrollToSection('faq');
      },
    },
  ];

  return (
    <header className="bg-white shadow-sm">
      <div className="max-w-7xl mx-auto px-2 sm:px-4 lg:px-8">
        <div className="flex justify-between items-center h-20 lg:h-24">
          <div className="flex items-center min-w-0 flex-shrink-0">
            <Link
              to="/"
              className="flex items-center space-x-2 lg:space-x-3 min-w-0"
            >
              <img
                src={partner?.logoUrl || '/logo.svg'}
                alt={partner?.name || t('navigation.company_name')}
                className="h-8 sm:h-10 lg:h-12 w-auto flex-shrink-0"
                onError={e => {
                  console.warn('Logo failed to load:', e.target.src);
                  e.target.src = '/logo.svg';
                }}
              />
              {partner?.name && (
                <span className="text-sm sm:text-lg lg:text-xl font-semibold text-gray-900 hidden sm:block truncate min-w-0">
                  {partner.name}
                </span>
              )}
            </Link>
          </div>

          <nav className="hidden lg:flex space-x-4 xl:space-x-8 flex-1 justify-center px-4">
            {navigation.map(item =>
              item.onClick ? (
                <button
                  key={item.name}
                  onClick={item.onClick}
                  className={`${
                    item.current
                      ? 'border-brand-primary text-brand-primary'
                      : item.highlight
                        ? 'bg-purple-600 text-white px-4 py-2 rounded-lg hover:bg-purple-700'
                        : 'border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700'
                  } ${item.highlight ? '' : 'inline-flex items-center px-1 xl:px-2 pt-1 border-b-2'} text-xs sm:text-sm lg:text-base font-medium transition-colors whitespace-nowrap`}
                >
                  {item.name}
                </button>
              ) : (
                <Link
                  key={item.name}
                  to={item.href}
                  className={`${
                    item.current
                      ? 'border-brand-primary text-brand-primary'
                      : item.highlight
                        ? 'bg-purple-600 text-white px-4 py-2 rounded-lg hover:bg-purple-700'
                        : 'border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700'
                  } ${item.highlight ? '' : 'inline-flex items-center px-1 xl:px-2 pt-1 border-b-2'} text-xs sm:text-sm lg:text-base font-medium transition-colors whitespace-nowrap`}
                >
                  {item.name}
                </Link>
              )
            )}
          </nav>

          <div className="hidden lg:flex items-center flex-shrink-0">
            <LanguageSelector />
          </div>

          <div className="lg:hidden">
            <button
              onClick={() => setIsMenuOpen(!isMenuOpen)}
              className="inline-flex items-center justify-center p-2 rounded-md text-gray-400 hover:text-gray-500 hover:bg-gray-100"
            >
              {isMenuOpen ? (
                <X className="block h-6 w-6" />
              ) : (
                <Menu className="block h-6 w-6" />
              )}
            </button>
          </div>
        </div>
      </div>

      {isMenuOpen && (
        <div className="lg:hidden bg-white border-t border-gray-100">
          <div className="pt-2 pb-3 space-y-1 px-2">
            {navigation.map(item =>
              item.onClick ? (
                <button
                  key={item.name}
                  onClick={e => {
                    item.onClick(e);
                    setIsMenuOpen(false);
                  }}
                  className={`${
                    item.current
                      ? 'bg-brand-primary/10 border-brand-primary text-brand-primary'
                      : 'border-transparent text-gray-600 hover:bg-gray-50 hover:border-gray-300 hover:text-gray-800'
                  } block px-3 py-3 border-l-4 text-base font-medium transition-colors w-full text-left rounded-r-md`}
                >
                  {item.name}
                </button>
              ) : (
                <Link
                  key={item.name}
                  to={item.href}
                  className={`${
                    item.current
                      ? 'bg-brand-primary/10 border-brand-primary text-brand-primary'
                      : 'border-transparent text-gray-600 hover:bg-gray-50 hover:border-gray-300 hover:text-gray-800'
                  } block px-3 py-3 border-l-4 text-base font-medium transition-colors rounded-r-md`}
                  onClick={() => setIsMenuOpen(false)}
                >
                  {item.name}
                </Link>
              )
            )}
          </div>

          <div className="py-3 border-t border-gray-200">
            <div className="flex items-center justify-center px-4">
              <LanguageSelector />
            </div>
          </div>
        </div>
      )}
    </header>
  );
};

export default Header;
