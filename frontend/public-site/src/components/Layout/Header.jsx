import React, { useState } from 'react'
import { Link, useLocation } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { Menu, X, ChevronDown, LogIn } from 'lucide-react'
import { usePartnerStore } from '../../store/partnerStore'
import LanguageSelector from '../ui/LanguageSelector'

const Header = () => {
  const { t } = useTranslation()
  const location = useLocation()
  const { partner } = usePartnerStore()
  // Читаем URL из глобального window.__ENV__ если есть (на случай единого .env из корня)
  const runtimeDashboardUrl = (typeof window !== 'undefined' && window.__ENV__ && window.__ENV__.DASHBOARD_URL) || null
  const [isMenuOpen, setIsMenuOpen] = useState(false)

  // Function to scroll to section
  const scrollToSection = (sectionId) => {
    // If we're not on home page, go to home first
    if (location.pathname !== '/') {
      window.location.href = `/#${sectionId}`
      return
    }
    
    const element = document.getElementById(sectionId)
    if (element) {
      element.scrollIntoView({ 
        behavior: 'smooth',
        block: 'start',
        inline: 'nearest'
      })
    }
  }

  const navigation = [
    {
      name: t('navigation.home'),
      href: '/',
      current: location.pathname === '/',
      onClick: null
    },
    {
      name: t('navigation.diamond_art'),
      href: '/diamond-art',
      current: location.pathname === '/diamond-art',
      onClick: null
    },
    {
      name: t('navigation.paint_by_numbers'),
      href: '#paint-by-numbers',
      current: false,
      onClick: (e) => {
        e.preventDefault()
        scrollToSection('paint-by-numbers')
      }
    },
    {
      name: t('navigation.what_is_this'),
      href: '#faq',
      current: false,
      onClick: (e) => {
        e.preventDefault()
        scrollToSection('faq')
      }
    }
  ]

  const goToDashboard = () => {
    // Очищаем токен авторизации, чтобы всегда показывалась страница логина
    localStorage.removeItem('token')
    localStorage.removeItem('refreshToken')
    localStorage.removeItem('refresh_token')
    localStorage.removeItem('user')
    
    // Переходим на страницу логина admin dashboard c параметром force=1
    const adminUrl = runtimeDashboardUrl || import.meta.env.VITE_DASHBOARD_URL || 'https://adm.doyoupaint.com'
    const target = `${adminUrl.replace(/\/$/, '')}/login?force=1&ts=${Date.now()}`
    window.open(target, '_blank', 'noopener,noreferrer')
  }

  return (
    <header className="bg-white shadow-sm">
      <div className="max-w-7xl mx-auto px-2 sm:px-4 lg:px-8">
        <div className="flex justify-between items-center h-16">
          {/* Левая часть: кнопка входа в самом левом углу и логотип */}
          <div className="flex items-center space-x-2 sm:space-x-4">
            {/* Кнопка входа в самом левом углу */}
            <button
              onClick={goToDashboard}
              className="inline-flex items-center px-2 py-2 sm:px-3 md:px-4 border border-transparent text-sm font-medium rounded-lg text-white bg-brand-primary hover:bg-brand-primary/90 transition-colors flex-shrink-0"
              title={t('navigation.login_tooltip')}
            >
              <LogIn className="w-4 h-4" />
              <span className="hidden sm:inline ml-1 md:ml-2">{t('navigation.login')}</span>
            </button>
            
            <Link to="/" className="flex items-center flex-shrink-0">
              <img 
                src={partner?.logoUrl || '/logo.svg'} 
                alt={partner?.name || t('navigation.company_name')} 
                className="h-8 sm:h-10 w-auto max-w-[120px] sm:max-w-none"
                onError={(e) => {
                  console.warn('Logo failed to load:', e.target.src)
                  e.target.src = '/logo.svg'
                }}
              />
            </Link>
          </div>

          {/* Desktop Navigation */}
          <nav className="hidden md:flex space-x-8">
            {navigation.map((item) => (
              item.onClick ? (
                <button
                  key={item.name}
                  onClick={item.onClick}
                  className={`${
                    item.current
                      ? 'border-brand-primary text-brand-primary'
                      : 'border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700'
                  } inline-flex items-center px-1 pt-1 border-b-2 text-sm font-medium transition-colors`}
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
                      : 'border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700'
                  } inline-flex items-center px-1 pt-1 border-b-2 text-sm font-medium transition-colors`}
                >
                  {item.name}
                </Link>
              )
            ))}
          </nav>

          {/* Right side */}
          <div className="hidden md:flex items-center">
            <LanguageSelector />
          </div>

          {/* Mobile menu button */}
          <div className="md:hidden">
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

      {/* Mobile menu */}
      {isMenuOpen && (
        <div className="md:hidden bg-white border-t border-gray-100">
          <div className="pt-2 pb-3 space-y-1 px-2">
            {navigation.map((item) => (
              item.onClick ? (
                <button
                  key={item.name}
                  onClick={(e) => {
                    item.onClick(e)
                    setIsMenuOpen(false)
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
            ))}
          </div>
          
          <div className="py-3 border-t border-gray-200">
            <div className="flex items-center justify-center px-4">
              <LanguageSelector />
            </div>
          </div>
        </div>
      )}
    </header>
  )
}

export default Header
