import React, { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { motion } from 'framer-motion'
import { Play, Ticket, Ruler, ShoppingCart, ArrowRight, ExternalLink, AlertTriangle, Palette, Eye } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import MosaicAPI from '../api/client'

const DiamondArtPage = () => {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const [couponCode, setCouponCode] = useState('')
  const [domainWarning, setDomainWarning] = useState(null)
  const [isValidating, setIsValidating] = useState(false)

  const handleCouponInput = (e) => {
    const digitsOnly = e.target.value.replace(/[^0-9]/g, '').substring(0, 12)
    let formattedCode = ''
    if (digitsOnly.length > 0) {
      formattedCode = digitsOnly
      if (digitsOnly.length > 4) {
        formattedCode = digitsOnly.substring(0, 4) + '-' + digitsOnly.substring(4)
      }
      if (digitsOnly.length > 8) {
        formattedCode = digitsOnly.substring(0, 4) + '-' + digitsOnly.substring(4, 8) + '-' + digitsOnly.substring(8)
      }
    }
    setCouponCode(formattedCode)
    
    // Очищаем предупреждение при изменении кода
    if (domainWarning) {
      setDomainWarning(null)
    }
  }

  const goToEditor = () => {
    const clean = couponCode.replace(/-/g, '')
    if (clean.length === 12) {
      navigate(`/editor?coupon=${clean}`)
    }
  }

  const validateCouponDomain = async (code) => {
    if (code.replace(/-/g, '').length !== 12) return;
    
    try {
      setIsValidating(true);
      const response = await MosaicAPI.validateCoupon(code.replace(/-/g, ''));
      
      if (response.valid && !response.is_correct_domain) {
        setDomainWarning({
          partnerDomain: response.correct_domain,
          partnerBrandName: response.partner_brand_name,
          message: response.message
        });
      } else {
        setDomainWarning(null);
      }
    } catch (error) {
      console.error('Failed to validate coupon domain:', error);
    } finally {
      setIsValidating(false);
    }
  };

  const goToPartnerSite = () => {
    if (domainWarning?.partnerDomain) {
      window.open(`https://${domainWarning.partnerDomain}`, '_blank');
    }
  };

  const marketplaceLinks = [
    {
      name: 'OZON',
      url: 'https://www.ozon.ru/search/?text=алмазная+мозаика+набор',
      description: t('sections.diamond_art.purchase_section.marketplaces.ozon.description'),
      buttonText: t('sections.diamond_art.purchase_section.marketplaces.ozon.button'),
      color: 'from-orange-500 to-red-500'
    },
    {
      name: 'Wildberries',
      url: 'https://www.wildberries.ru/catalog/0/search.aspx?search=алмазная+мозаика+набор',
      description: t('sections.diamond_art.purchase_section.marketplaces.wildberries.description'),
      buttonText: t('sections.diamond_art.purchase_section.marketplaces.wildberries.button'),
      color: 'from-purple-500 to-pink-500'
    }
  ]

  return (
    <div className="min-h-screen bg-white">
      {/* Видео-инструкция по созданию алмазной мозаики */}
      <section className="bg-gradient-to-br from-blue-50 via-purple-50 to-pink-50 py-16">
        <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8">
          <motion.div 
            initial={{ opacity: 0, y: 20 }} 
            animate={{ opacity: 1, y: 0 }} 
            transition={{ duration: 0.5 }} 
            className="grid lg:grid-cols-2 gap-12 items-center"
          >
            <div className="aspect-video bg-black rounded-2xl overflow-hidden shadow-2xl">
              <iframe
                className="w-full h-full"
                src="https://www.youtube.com/embed/dQw4w9WgXcQ"
                title="Инструкция по созданию алмазной мозаики"
                frameBorder="0"
                allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share"
                allowFullScreen
              />
            </div>
            <div>
              <div className="flex items-center space-x-4 mb-6">
                <div className="w-16 h-16 bg-brand-primary/10 rounded-2xl flex items-center justify-center">
                  <Play className="w-8 h-8 text-brand-primary" />
                </div>
                <h1 className="text-4xl font-bold text-gray-900">{t('sections.diamond_art.title')}</h1>
              </div>
              <p className="text-xl text-gray-600 leading-relaxed mb-6">
                {t('sections.diamond_art.description')}
              </p>
              <div className="bg-white/60 backdrop-blur-sm rounded-xl p-6 border border-white/20">
                <h3 className="text-lg font-semibold text-gray-800 mb-3">{t('sections.diamond_art.video_section.what_you_learn')}</h3>
                <ul className="space-y-2 text-gray-700">
                  {t('sections.diamond_art.video_section.learn_items', { returnObjects: true }).map((item, index) => (
                    <li key={index} className="flex items-center space-x-2">
                      <div className="w-2 h-2 bg-brand-primary rounded-full"></div>
                      <span>{item}</span>
                    </li>
                  ))}
                </ul>
              </div>
            </div>
          </motion.div>
        </div>
      </section>

      {/* Два основных блока на одном уровне */}
      <section className="py-16">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="grid lg:grid-cols-2 gap-8 items-start">
            
            {/* Блок "Активация купона" */}
            <motion.div 
              initial={{ opacity: 0, y: 20 }} 
              animate={{ opacity: 1, y: 0 }} 
              transition={{ duration: 0.5, delay: 0.1 }}
              className="bg-gradient-to-r from-brand-primary to-brand-secondary rounded-3xl shadow-2xl p-8 lg:p-10 text-white"
            >
              <div className="flex items-center space-x-4 mb-6">
                <div className="w-16 h-16 bg-white/20 backdrop-blur-sm rounded-2xl flex items-center justify-center">
                  <Ticket className="w-8 h-8 text-white" />
                </div>
                <div>
                  <h2 className="text-2xl lg:text-3xl font-bold">{t('sections.diamond_art.coupon_section.title')}</h2>
                  <p className="text-brand-primary/80 text-base lg:text-lg mt-2">{t('sections.diamond_art.coupon_section.description')}</p>
                </div>
              </div>
              
              <div className="space-y-4">
                <input
                  type="text"
                  value={couponCode}
                  onChange={handleCouponInput}
                  onBlur={() => validateCouponDomain(couponCode)}
                  placeholder={t('hero.coupon_banner.placeholder')}
                  maxLength={14}
                  className="w-full px-4 lg:px-6 py-3 lg:py-4 bg-white/10 backdrop-blur-sm border border-white/20 rounded-xl focus:ring-2 focus:ring-white/50 focus:border-transparent text-center text-lg lg:text-xl tracking-wider text-white placeholder-brand-primary/60"
                />
                
                <button
                  onClick={goToEditor}
                  disabled={couponCode.replace(/-/g, '').length !== 12}
                  className="w-full inline-flex items-center justify-center px-6 lg:px-8 py-3 lg:py-4 bg-white text-brand-primary rounded-xl hover:bg-brand-primary/10 disabled:opacity-50 disabled:cursor-not-allowed font-semibold text-base lg:text-lg transition-all duration-200"
                >
                  <span>{t('hero.coupon_banner.activate')}</span>
                  <ArrowRight className="w-5 h-5 ml-2" />
                </button>
                
                {/* Предупреждение о неправильном домене */}
                {domainWarning && (
                  <div className="bg-yellow-500/20 border border-yellow-400/30 rounded-xl p-4">
                    <div className="flex items-start space-x-3">
                      <AlertTriangle className="w-5 h-5 text-yellow-400 mt-0.5 flex-shrink-0" />
                      <div className="flex-1">
                        <p className="text-yellow-100 text-sm font-medium mb-2">
                          {domainWarning.message}
                        </p>
                        <p className="text-yellow-200/80 text-xs mb-3">
                          Этот купон предназначен для сайта партнера: <strong>{domainWarning.partnerBrandName}</strong>
                        </p>
                        <button
                          onClick={goToPartnerSite}
                          className="inline-flex items-center px-4 py-2 bg-yellow-500 hover:bg-yellow-600 text-white text-sm font-medium rounded-lg transition-colors duration-200"
                        >
                          Перейти на сайт партнера
                          <ExternalLink className="w-4 h-4 ml-2" />
                        </button>
                      </div>
                    </div>
                  </div>
                )}
                
                <p className="text-brand-primary/70 text-sm text-center">
                  {t('sections.diamond_art.coupon_section.code_hint')}
                </p>
              </div>
            </motion.div>

            {/* Блок "Покупка купона" */}
            <motion.div 
              initial={{ opacity: 0, y: 20 }} 
              animate={{ opacity: 1, y: 0 }} 
              transition={{ duration: 0.5, delay: 0.2 }}
              className="bg-gradient-to-br from-gray-50 to-gray-100 rounded-3xl shadow-2xl p-8 lg:p-10 border border-gray-200"
            >
              <div className="flex items-center space-x-4 mb-6">
                <div className="w-16 h-16 bg-brand-secondary/10 rounded-2xl flex items-center justify-center">
                  <ShoppingCart className="w-8 h-8 text-brand-secondary" />
                </div>
                <div>
                  <h2 className="text-2xl lg:text-3xl font-bold text-gray-900">{t('sections.diamond_art.purchase_section.title')}</h2>
                  <p className="text-gray-600 text-base lg:text-lg mt-2">{t('sections.diamond_art.purchase_section.description')}</p>
                </div>
              </div>
              
              <div className="space-y-4">
                {marketplaceLinks.map((marketplace, index) => (
                  <motion.div 
                    key={marketplace.name}
                    initial={{ opacity: 0, x: 20 }} 
                    animate={{ opacity: 1, x: 0 }} 
                    transition={{ duration: 0.5, delay: 0.3 + index * 0.1 }}
                    className="bg-white rounded-2xl shadow-lg border border-gray-100 p-6 hover:shadow-xl transition-all duration-300 hover:-translate-y-1"
                  >
                    <div className="flex items-center space-x-4">
                      <div className={`w-12 h-12 bg-gradient-to-r ${marketplace.color} rounded-xl flex items-center justify-center flex-shrink-0`}>
                        <ShoppingCart className="w-6 h-6 text-white" />
                      </div>
                      <div className="flex-1 min-w-0">
                        <h3 className="text-lg font-bold text-gray-900 mb-1">{marketplace.name}</h3>
                        <p className="text-gray-600 text-sm">{marketplace.description}</p>
                      </div>
                    </div>
                    <a
                      href={marketplace.url}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="mt-4 inline-flex items-center justify-center w-full px-4 py-3 bg-gradient-to-r from-brand-primary to-brand-secondary text-white rounded-xl hover:from-brand-primary/90 hover:to-brand-secondary/90 font-semibold transition-all duration-200"
                    >
                      <span>{marketplace.buttonText}</span>
                      <ExternalLink className="w-4 h-4 ml-2" />
                    </a>
                  </motion.div>
                ))}
              </div>
            </motion.div>
          </div>
        </div>
      </section>

      {/* Блок "Информация о размерах и стилях" */}
      <section className="py-16 bg-gray-50">
        <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8">
          <motion.div 
            initial={{ opacity: 0, y: 20 }} 
            animate={{ opacity: 1, y: 0 }} 
            transition={{ duration: 0.5, delay: 0.3 }}
            className="text-center mb-12"
          >
            <div className="flex items-center justify-center space-x-4 mb-6">
              <div className="w-16 h-16 bg-brand-primary/10 rounded-2xl flex items-center justify-center">
                <Ruler className="w-8 h-8 text-brand-primary" />
              </div>
              <h2 className="text-3xl font-bold text-gray-900">Размеры и стили мозаики</h2>
            </div>
            <p className="text-xl text-gray-600 max-w-3xl mx-auto">
              Выберите размер и стиль, чтобы увидеть, как будет выглядеть ваша мозаика
            </p>
          </motion.div>

          <div className="grid md:grid-cols-2 gap-8 mb-12">
            {/* Размеры */}
            <motion.div 
              initial={{ opacity: 0, x: -20 }} 
              animate={{ opacity: 1, x: 0 }} 
              transition={{ duration: 0.5, delay: 0.4 }}
              className="bg-white rounded-2xl shadow-lg p-6"
            >
              <h3 className="text-xl font-bold text-gray-900 mb-4 flex items-center">
                <Ruler className="w-5 h-5 mr-2 text-brand-primary" />
                Доступные размеры
              </h3>
              <div className="grid grid-cols-2 gap-3">
                {[
                  { key: '20x20', title: '20×20', desc: 'Для начинающих', stones: '~400 камней' },
                  { key: '30x40', title: '30×40', desc: 'Популярный', stones: '~1200 камней' },
                  { key: '40x40', title: '40×40', desc: 'Сбалансированный', stones: '~1600 камней' },
                  { key: '40x50', title: '40×50', desc: 'Пейзажи', stones: '~2000 камней' },
                  { key: '40x60', title: '40×60', desc: 'Панорамы', stones: '~2400 камней' },
                  { key: '50x70', title: '50×70', desc: 'Для опытных', stones: '~3500 камней' }
                ].map((size) => (
                  <div key={size.key} className="text-center p-3 bg-gray-50 rounded-lg">
                    <div className="font-semibold text-gray-900">{size.title}</div>
                    <div className="text-sm text-gray-600">{size.desc}</div>
                    <div className="text-xs text-brand-primary font-medium mt-1">{size.stones}</div>
                  </div>
                ))}
              </div>
            </motion.div>

            {/* Стили */}
            <motion.div 
              initial={{ opacity: 0, x: 20 }} 
              animate={{ opacity: 1, x: 0 }} 
              transition={{ duration: 0.5, delay: 0.5 }}
              className="bg-white rounded-2xl shadow-lg p-6"
            >
              <h3 className="text-xl font-bold text-gray-900 mb-4 flex items-center">
                <Palette className="w-5 h-5 mr-2 text-brand-secondary" />
                Доступные стили
              </h3>
              <div className="space-y-3">
                {[
                  { key: 'grayscale', title: 'Черно-белый', desc: 'Классический, элегантный', colors: '~20 оттенков' },
                  { key: 'skin_tones', title: 'Телесные тона', desc: 'Реалистичные оттенки кожи', colors: '~30 оттенков' },
                  { key: 'pop_art', title: 'Поп-арт', description: 'Яркие, контрастные цвета', colors: '~50 оттенков' },
                  { key: 'max_colors', title: 'Максимум цветов', description: 'Богатая цветовая палитра', colors: '~100 оттенков' }
                ].map((style) => (
                  <div key={style.key} className="p-3 bg-gray-50 rounded-lg">
                    <div className="font-semibold text-gray-900">{style.title}</div>
                    <div className="text-sm text-gray-600">{style.description || style.desc}</div>
                    <div className="text-xs text-brand-secondary font-medium mt-1">{style.colors}</div>
                  </div>
                ))}
              </div>
            </motion.div>
          </div>

          {/* Кнопка перехода к превью */}
          <motion.div 
            initial={{ opacity: 0, y: 20 }} 
            animate={{ opacity: 1, y: 0 }} 
            transition={{ duration: 0.5, delay: 0.6 }}
            className="text-center"
          >
            <button
              onClick={() => navigate('/mosaic-preview')}
              className="inline-flex items-center px-8 py-4 bg-gradient-to-r from-brand-primary to-brand-secondary text-white rounded-2xl hover:from-brand-primary/90 hover:to-brand-secondary/90 font-semibold text-lg transition-all duration-200 shadow-lg hover:shadow-xl"
            >
              <Eye className="w-6 h-6 mr-3" />
              Посмотреть превью мозаики
              <ArrowRight className="w-6 h-6 ml-3" />
            </button>
            <p className="text-gray-600 mt-4">
              Выберите размер и стиль, чтобы увидеть, как будет выглядеть ваша мозаика
            </p>
          </motion.div>
        </div>
      </section>
    </div>
  )
}

export default DiamondArtPage


