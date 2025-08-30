import React, { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { motion } from 'framer-motion'
import { Play, Ticket, Ruler, ShoppingCart, ArrowRight, ExternalLink } from 'lucide-react'
import { useNavigate } from 'react-router-dom'

const DiamondArtPage = () => {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const [couponCode, setCouponCode] = useState('')

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
  }

  const goToEditor = () => {
    const clean = couponCode.replace(/-/g, '')
    if (clean.length === 12) {
      navigate(`/editor?coupon=${clean}`)
    }
  }

  const sizeCards = [
    { key: '21x30', title: '21×30', w: 'w-16', h: 'h-24' },
    { key: '30x40', title: '30×40', w: 'w-20', h: 'h-28' },
    { key: '40x40', title: '40×40', w: 'w-24', h: 'h-24' },
    { key: '40x50', title: '40×50', w: 'w-24', h: 'h-32' },
    { key: '40x60', title: '40×60', w: 'w-24', h: 'h-36' },
    { key: '50x70', title: '50×70', w: 'w-28', h: 'h-40' }
  ]

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

      {/* Блок "Я уже купил набор" — переход к вводу кода купона */}
      <section className="py-16">
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
          <motion.div 
            initial={{ opacity: 0, y: 20 }} 
            animate={{ opacity: 1, y: 0 }} 
            transition={{ duration: 0.5, delay: 0.1 }}
            className="bg-gradient-to-r from-brand-primary to-brand-secondary rounded-3xl shadow-2xl p-10 text-white"
          >
            <div className="flex items-center space-x-4 mb-8">
              <div className="w-16 h-16 bg-white/20 backdrop-blur-sm rounded-2xl flex items-center justify-center">
                <Ticket className="w-8 h-8 text-white" />
              </div>
              <div>
                <h2 className="text-3xl font-bold">{t('sections.diamond_art.coupon_section.title')}</h2>
                <p className="text-brand-primary/80 text-lg mt-2">{t('sections.diamond_art.coupon_section.description')}</p>
              </div>
            </div>
            <div className="grid sm:grid-cols-[1fr_auto] gap-6">
              <input
                type="text"
                value={couponCode}
                onChange={handleCouponInput}
                placeholder={t('hero.coupon_banner.placeholder')}
                maxLength={14}
                className="px-6 py-4 bg-white/10 backdrop-blur-sm border border-white/20 rounded-xl focus:ring-2 focus:ring-white/50 focus:border-transparent text-center text-xl tracking-wider text-white placeholder-brand-primary/60"
              />
              <button
                onClick={goToEditor}
                disabled={couponCode.replace(/-/g, '').length !== 12}
                className="inline-flex items-center justify-center px-8 py-4 bg-white text-brand-primary rounded-xl hover:bg-brand-primary/10 disabled:opacity-50 disabled:cursor-not-allowed font-semibold text-lg transition-all duration-200"
              >
                <span>{t('hero.coupon_banner.activate')}</span>
                <ArrowRight className="w-6 h-6 ml-2" />
              </button>
            </div>
            <div className="mt-6 text-center">
              <p className="text-brand-primary/70 text-sm">{t('sections.diamond_art.coupon_section.code_hint')}</p>
            </div>
          </motion.div>
        </div>
      </section>

      {/* Блок "Как будет выглядеть мозаика" — демонстрация размеров */}
      <section className="py-16 bg-gray-50">
        <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8">
          <motion.div 
            initial={{ opacity: 0, y: 20 }} 
            animate={{ opacity: 1, y: 0 }} 
            transition={{ duration: 0.5, delay: 0.2 }}
          >
            <div className="text-center mb-12">
              <div className="flex items-center justify-center space-x-4 mb-6">
                <div className="w-16 h-16 bg-brand-secondary/10 rounded-2xl flex items-center justify-center">
                  <Ruler className="w-8 h-8 text-brand-secondary" />
                </div>
                              <h2 className="text-3xl font-bold text-gray-900">{t('sections.diamond_art.sizes_section.title')}</h2>
            </div>
            <p className="text-xl text-gray-600 max-w-3xl mx-auto">
              {t('sections.diamond_art.sizes_section.description')}
            </p>
            </div>
            <div className="grid grid-cols-2 md:grid-cols-3 gap-8">
              {sizeCards.map((s, index) => (
                <motion.div 
                  key={s.key} 
                  initial={{ opacity: 0, y: 20 }} 
                  animate={{ opacity: 1, y: 0 }} 
                  transition={{ duration: 0.5, delay: 0.1 + index * 0.1 }}
                  className="bg-white rounded-2xl shadow-lg border border-gray-100 p-8 text-center hover:shadow-xl transition-all duration-300 hover:-translate-y-1"
                >
                  <div className={`mx-auto mb-6 rounded-lg bg-gradient-to-br from-brand-primary to-brand-secondary ${s.w} ${s.h} shadow-lg`} />
                  <div className="text-2xl font-bold text-gray-900 mb-2">{s.title} {t('common.cm', { defaultValue: 'см' })}</div>
                  <div className="text-gray-600">
                    {t(`sections.diamond_art.sizes_section.sizes.${s.key}`)}
                  </div>
                </motion.div>
              ))}
            </div>
          </motion.div>
        </div>
      </section>

      {/* Блок "Хочу купить" — переход к выбору площадок для покупки */}
      <section className="py-16">
        <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8">
          <motion.div 
            initial={{ opacity: 0, y: 20 }} 
            animate={{ opacity: 1, y: 0 }} 
            transition={{ duration: 0.5, delay: 0.3 }}
          >
            <div className="text-center mb-12">
              <div className="flex items-center justify-center space-x-4 mb-6">
                <div className="w-16 h-16 bg-brand-secondary/10 rounded-2xl flex items-center justify-center">
                  <ShoppingCart className="w-8 h-8 text-brand-secondary" />
                </div>
                              <h2 className="text-3xl font-bold text-gray-900">{t('sections.diamond_art.purchase_section.title')}</h2>
            </div>
            <p className="text-xl text-gray-600 max-w-3xl mx-auto">
              {t('sections.diamond_art.purchase_section.description')}
            </p>
            </div>
            
            <div className="grid md:grid-cols-2 gap-8">
              {marketplaceLinks.map((marketplace, index) => (
                <motion.div 
                  key={marketplace.name}
                  initial={{ opacity: 0, y: 20 }} 
                  animate={{ opacity: 1, y: 0 }} 
                  transition={{ duration: 0.5, delay: 0.4 + index * 0.1 }}
                  className="bg-white rounded-2xl shadow-lg border border-gray-100 p-8 hover:shadow-xl transition-all duration-300 hover:-translate-y-1"
                >
                  <div className={`w-16 h-16 bg-gradient-to-r ${marketplace.color} rounded-2xl flex items-center justify-center mx-auto mb-6`}>
                    <ShoppingCart className="w-8 h-8 text-white" />
                  </div>
                  <h3 className="text-2xl font-bold text-gray-900 text-center mb-4">{marketplace.name}</h3>
                  <p className="text-gray-600 text-center mb-6">{marketplace.description}</p>
                  <a
                    href={marketplace.url}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="inline-flex items-center justify-center w-full px-6 py-3 bg-gradient-to-r from-brand-primary to-brand-secondary text-white rounded-xl hover:from-brand-primary/90 hover:to-brand-secondary/90 font-semibold transition-all duration-200"
                  >
                    <span>{marketplace.buttonText}</span>
                    <ExternalLink className="w-5 h-5 ml-2" />
                  </a>
                </motion.div>
              ))}
            </div>

            <div className="mt-12 text-center">
              <div className="bg-brand-primary/5 rounded-2xl p-8 border border-brand-primary/20 max-w-4xl mx-auto">
                <h3 className="text-xl font-semibold text-brand-primary mb-4">{t('sections.diamond_art.purchase_section.advice.title')}</h3>
                <p className="text-brand-primary/80">
                  {t('sections.diamond_art.purchase_section.advice.content')}
                </p>
              </div>
            </div>
          </motion.div>
        </div>
      </section>
    </div>
  )
}

export default DiamondArtPage


