import React, { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { motion } from 'framer-motion'
import { ShoppingBag, Check, ExternalLink, CreditCard, Building2 } from 'lucide-react'
import { useMutation } from '@tanstack/react-query'
import { MosaicAPI } from '../api/client'
import { useUIStore } from '../store/partnerStore'

const ShopPage = () => {
  const { t } = useTranslation()
  
  // Восстанавливаем состояние из sessionStorage
  const [selectedSize, setSelectedSize] = useState(() => {
    try {
      return sessionStorage.getItem('shop:selectedSize') || null
    } catch {
      return null
    }
  })
  
  const [selectedStyle, setSelectedStyle] = useState(() => {
    try {
      return sessionStorage.getItem('shop:selectedStyle') || 'grayscale'
    } catch {
      return 'grayscale'
    }
  })
  
  const { addNotification } = useUIStore()

  const createOrderMutation = useMutation({
    mutationFn: (payload) => MosaicAPI.initiateCouponOrder(payload),
    onSuccess: (data) => {
      if (data?.payment_url) {
        window.location.href = data.payment_url
      } else if (data?.redirect_url) {
        window.location.href = data.redirect_url
      } else {
        addNotification({ type: 'error', title: t('payment.error_title'), message: t('payment.purchase_failed') })
      }
    },
    onError: (err) => {
      addNotification({ type: 'error', title: t('payment.error_title'), message: err?.message || t('payment.purchase_failed') })
    }
  })

  const fadeInUp = {
    initial: { opacity: 0, y: 20 },
    animate: { opacity: 1, y: 0 },
    transition: { duration: 0.6 }
  }

  const sizeKeys = ['21x30', '30x40', '40x40', '40x50', '40x60', '50x70']
  // Ключи стилей должны соответствовать бэкенду: grayscale, skin_tone, pop_art, max_colors
  const styleKeys = ['grayscale', 'skin_tone', 'pop_art', 'full_color']

  // Rectangle sizes for visual representation (proportional to actual sizes)
  const rectangleSizes = {
    '21x30': { width: 'w-12', height: 'h-16' },
    '30x40': { width: 'w-14', height: 'h-20' },
    '40x40': { width: 'w-16', height: 'h-16' },
    '40x50': { width: 'w-16', height: 'h-20' },
    '40x60': { width: 'w-16', height: 'h-24' },
    '50x70': { width: 'w-20', height: 'h-28' }
  }

  // Сохраняем состояние в sessionStorage при изменении
  const handleSizeChange = (size) => {
    setSelectedSize(size)
    try {
      sessionStorage.setItem('shop:selectedSize', size)
    } catch {}
  }

  const handleStyleChange = (style) => {
    setSelectedStyle(style)
    try {
      sessionStorage.setItem('shop:selectedStyle', style)
    } catch {}
  }

  const handlePurchase = () => {
    if (!selectedSize) {
      alert(t('shop.alerts.select_size'))
      return
    }
    const payload = {
      size: selectedSize,
      style: selectedStyle,
      // все завершение будет через вебхуки; редирект на маркетплейс банка
      return_url: window.location.origin,
      fail_url: window.location.origin
    }
    createOrderMutation.mutate(payload)
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-16">
        <motion.div {...fadeInUp} className="text-center mb-12">
          <h1 className="text-4xl font-bold text-gray-900 mb-4">
            {t('shop.title')}
          </h1>
          <p className="text-xl text-gray-600 max-w-2xl mx-auto">
            {t('shop.subtitle')}
          </p>
        </motion.div>

        {/* Size Selection */}
        <motion.section 
          {...fadeInUp}
          transition={{ delay: 0.2 }}
          className="mb-16"
        >
          <h2 className="text-2xl font-bold text-gray-900 mb-8 flex items-center justify-center">
            {t('shop.select_size')}
          </h2>
          <div className="grid grid-cols-2 md:grid-cols-3 gap-6">
            {sizeKeys.map((sizeKey, index) => {
              const sizeData = t(`shop.sizes.${sizeKey}`, { returnObjects: true })
              const isSelected = selectedSize === sizeKey
              const rectangleSize = rectangleSizes[sizeKey]

              return (
                <motion.div
                  key={sizeKey}
                  initial={{ opacity: 0, y: 20 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ duration: 0.5, delay: index * 0.1 }}
                  onClick={() => handleSizeChange(sizeKey)}
                  className={`relative bg-white rounded-2xl shadow-lg p-6 cursor-pointer transition-all duration-300 transform hover:scale-105 hover:shadow-xl ${
                    isSelected 
                      ? 'ring-4 ring-brand-primary border-brand-primary' 
                      : 'border border-gray-200 hover:border-gray-300'
                  }`}
                >
                  {isSelected && (
                    <div className="absolute -top-2 -right-2 w-6 h-6 bg-brand-primary rounded-full flex items-center justify-center">
                      <Check className="w-4 h-4 text-white" />
                    </div>
                  )}
                  
                  <div className="flex flex-col items-center text-center">
                    <div className={`mb-4 rounded ${rectangleSize.width} ${rectangleSize.height} ${
                      isSelected ? 'bg-brand-primary' : 'bg-brand-secondary'
                    }`} />
                    
                    <h3 className="text-xl font-semibold text-gray-900 mb-2">
                      {sizeData.title}
                    </h3>
                    
                    <p className="text-gray-600">
                      {sizeData.description}
                    </p>
                  </div>
                </motion.div>
              )
            })}
          </div>
        </motion.section>

        {/* Style Selection */}
        <motion.section 
          {...fadeInUp}
          transition={{ delay: 0.4 }}
          className="mb-16"
        >
          <h2 className="text-2xl font-bold text-gray-900 mb-8 flex items-center justify-center">
            {t('shop.select_style')}
          </h2>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-6">
            {styleKeys.map((styleKey, index) => {
              const styleData = t(`shop.styles.${styleKey}`, { returnObjects: true })
              const isSelected = selectedStyle === styleKey

              // Style preview colors
              const styleColors = {
                'grayscale': 'linear-gradient(135deg, #000000, #ffffff)',
                'skin_tone': 'linear-gradient(135deg, #8B4513, #DEB887)',
                'pop_art': 'linear-gradient(135deg, #FF6B35, #00FF41, #0099FF, #FF0080)',
                'full_color': 'linear-gradient(135deg, #FF0080, #00FF41, #0099FF, #FFD700)'
              }

              return (
                <motion.div
                  key={styleKey}
                  initial={{ opacity: 0, y: 20 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ duration: 0.5, delay: index * 0.1 }}
                  onClick={() => handleStyleChange(styleKey)}
                  className={`relative bg-white rounded-2xl shadow-lg p-6 cursor-pointer transition-all duration-300 transform hover:scale-105 hover:shadow-xl ${
                    isSelected 
                      ? 'ring-4 ring-brand-primary border-brand-primary' 
                      : 'border border-gray-200 hover:border-gray-300'
                  }`}
                >
                  {isSelected && (
                    <div className="absolute -top-2 -right-2 w-6 h-6 bg-brand-primary rounded-full flex items-center justify-center">
                      <Check className="w-4 h-4 text-white" />
                    </div>
                  )}
                  
                  <div className="flex flex-col items-center text-center">
                    <div 
                      className="w-16 h-12 mb-4 rounded"
                      style={{ background: styleColors[styleKey] }}
                    />
                    
                    <h3 className="text-lg font-semibold text-gray-900 mb-2">
                      {styleData.title}
                    </h3>
                    
                    <p className="text-sm text-gray-600">
                      {styleData.description}
                    </p>
                  </div>
                </motion.div>
              )
            })}
          </div>
        </motion.section>

        {/* Fixed Price and Purchase Button */}
        <motion.div
          {...fadeInUp}
          transition={{ delay: 0.6 }}
          className="text-center mb-8"
        >
          <div className="text-4xl font-bold text-brand-primary mb-6">
            {t('shop.final_price')}
          </div>
          <button
            onClick={handlePurchase}
            disabled={!selectedSize}
            className={`px-12 py-4 text-lg font-semibold rounded-xl transition-all duration-300 shadow-lg transform ${
              selectedSize 
                ? 'bg-brand-primary text-white hover:bg-brand-primary/90 hover:shadow-xl hover:scale-105' 
                : 'bg-gray-300 text-gray-500 cursor-not-allowed'
            }`}
          >
            <div className="flex items-center justify-center space-x-2">
              <CreditCard className="w-5 h-5" />
              <span>{t('shop.proceed_to_payment')}</span>
            </div>
          </button>
        </motion.div>

        {/* How it works */}
        <motion.div 
          {...fadeInUp}
          transition={{ delay: 0.8 }}
          className="bg-brand-primary/5 rounded-2xl p-8 text-center mb-8"
        >
          <ShoppingBag className="w-16 h-16 text-brand-primary mx-auto mb-4" />
          <h2 className="text-2xl font-semibold mb-4">{t('shop.how_it_works')}</h2>
          <div className="grid md:grid-cols-3 gap-6 text-left">
            {['step1', 'step2', 'step3'].map((step, index) => {
              const stepData = t(`shop.steps.${step}`, { returnObjects: true })
              return (
                <div key={step}>
                  <div className="w-8 h-8 bg-brand-primary text-white rounded-full flex items-center justify-center mb-3 font-semibold">
                    {index + 1}
                  </div>
                  <h3 className="font-semibold mb-2">{stepData.title}</h3>
                  <p className="text-gray-600">{stepData.description}</p>
                </div>
              )
            })}
          </div>
        </motion.div>

        {/* Company Information */}
        <motion.div 
          {...fadeInUp}
          transition={{ delay: 1.0 }}
          className="bg-gradient-to-r from-blue-50 to-indigo-50 rounded-2xl shadow-lg p-8 border-2 border-blue-200"
        >
          <div className="flex items-center justify-center mb-6">
            <Building2 className="w-8 h-8 text-blue-600 mr-3" />
            <h2 className="text-2xl font-bold text-gray-900">
              {t('shop.company_info.title')}
            </h2>
          </div>
          <div className="grid md:grid-cols-2 gap-8 text-center md:text-left">
            <div className="space-y-4">
              <div className="text-xl font-semibold text-gray-900">
                {t('shop.company_info.name')}
              </div>
              <div className="text-lg text-gray-700 font-medium">
                {t('shop.company_info.inn')}
              </div>
            </div>
            <div className="space-y-4">
              <div className="text-lg text-gray-700">
                {t('shop.company_info.address')}
              </div>
              <div className="text-lg text-gray-700 font-medium">
                {t('shop.company_info.phone')}
              </div>
            </div>
          </div>
        </motion.div>
      </div>
    </div>
  )
}

export default ShopPage