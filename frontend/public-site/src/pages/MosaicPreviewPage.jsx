import React, { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { motion } from 'framer-motion'
import { Palette, Ruler, ShoppingCart, ArrowRight, Download, Eye, Loader2 } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import { useMutation } from '@tanstack/react-query'
import { MosaicAPI } from '../api/client'
import { useUIStore, usePartnerStore } from '../store/partnerStore'

const MosaicPreviewPage = () => {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { addNotification } = useUIStore()
  const { partner } = usePartnerStore()
  
  const [selectedSize, setSelectedSize] = useState('30x40')
  const [selectedStyle, setSelectedStyle] = useState('max_colors')
  const [previewImage, setPreviewImage] = useState(null)
  const [couponCode, setCouponCode] = useState('')
  const [showActivation, setShowActivation] = useState(false)

  const sizes = [
    { key: '20x20', title: '20×20 см', description: 'Компактный размер для начинающих', stones: '~400 камней' },
    { key: '30x40', title: '30×40 см', description: 'Популярный размер', stones: '~1200 камней' },
    { key: '40x40', title: '40×40 см', description: 'Квадратный формат', stones: '~1600 камней' },
    { key: '40x50', title: '40×50 см', description: 'Прямоугольный формат', stones: '~2000 камней' },
    { key: '40x60', title: '40×60 см', description: 'Широкий формат', stones: '~2400 камней' },
    { key: '50x70', title: '50×70 см', description: 'Максимальный размер', stones: '~3500 камней' }
  ]

  const styles = [
    { key: 'grayscale', title: 'Черно-белый', description: 'Классический стиль', colors: '~20 оттенков' },
    { key: 'skin_tones', title: 'Телесные тона', description: 'Реалистичные оттенки', colors: '~30 оттенков' },
    { key: 'pop_art', title: 'Поп-арт', description: 'Яркие цвета', colors: '~50 оттенков' },
    { key: 'max_colors', title: 'Максимум цветов', description: 'Полная палитра', colors: '~100 оттенков' }
  ]

  // Генерация превью
  const generatePreviewMutation = useMutation({
    mutationFn: async ({ size, style }) => {
      const response = await MosaicAPI.generateMosaicPreview({
        size,
        style,
        sample_image: 'default' // Используем дефолтное изображение для превью
      })
      return response
    },
    onSuccess: (data) => {
      if (data.preview_url) {
        setPreviewImage(data.preview_url)
      }
    },
    onError: (error) => {
      addNotification({
        type: 'error',
        message: 'Не удалось создать превью'
      })
    }
  })

  // Генерируем превью при изменении размера или стиля
  useEffect(() => {
    generatePreviewMutation.mutate({ size: selectedSize, style: selectedStyle })
  }, [selectedSize, selectedStyle])

  // Активация купона
  const activateCouponMutation = useMutation({
    mutationFn: async (code) => {
      const info = await MosaicAPI.validateCoupon(code)
      
      // Проверка домена партнёра
      if (info?.partner_domain && window.location.hostname !== info.partner_domain) {
        const protocol = window.location.protocol
        const redirectUrl = `${protocol}//${info.partner_domain}/mosaic-preview?coupon=${code}&size=${selectedSize}&style=${selectedStyle}`
        window.location.href = redirectUrl
        return info
      }
      
      await MosaicAPI.activateCoupon(code)
      return info || {}
    },
    onSuccess: (couponData) => {
      if (couponData?.partner_domain && window.location.hostname !== couponData.partner_domain) {
        return
      }
      
      addNotification({
        type: 'success',
        message: 'Купон активирован! Переходим к созданию мозаики...'
      })
      
      setTimeout(() => {
        navigate(`/editor?coupon=${couponCode.replace(/-/g, '')}&size=${selectedSize}&style=${selectedStyle}`)
      }, 1500)
    },
    onError: (error) => {
      addNotification({
        type: 'error',
        message: 'Неверный номер купона'
      })
    }
  })

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

  const handleActivateCoupon = () => {
    if (couponCode.replace(/-/g, '').length === 12) {
      activateCouponMutation.mutate(couponCode.replace(/-/g, ''))
    }
  }

  // Получение ссылок на маркетплейсы с артикулами партнёра
  const getMarketplaceLink = (marketplace) => {
    if (!partner) return null
    
    // TODO: Здесь будет логика получения артикулов из базы данных
    // Пока используем прямые ссылки из партнёра
    if (marketplace === 'ozon' && partner.ozonLink) {
      return partner.ozonLink
    }
    if (marketplace === 'wildberries' && partner.wildberriesLink) {
      return partner.wildberriesLink
    }
    
    return null
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 via-purple-50 to-pink-50">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5 }}
          className="text-center mb-8"
        >
          <h1 className="text-4xl font-bold text-gray-900 mb-4">
            Превью алмазной мозаики
          </h1>
          <p className="text-xl text-gray-600">
            Выберите размер и стиль, чтобы увидеть, как будет выглядеть ваша мозаика
          </p>
        </motion.div>

        <div className="grid lg:grid-cols-3 gap-8">
          {/* Левая колонка - выбор параметров */}
          <div className="lg:col-span-1 space-y-6">
            {/* Выбор размера */}
            <motion.div
              initial={{ opacity: 0, x: -20 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ duration: 0.5, delay: 0.1 }}
              className="bg-white rounded-2xl shadow-lg p-6"
            >
              <h3 className="text-lg font-bold text-gray-900 mb-4 flex items-center">
                <Ruler className="w-5 h-5 mr-2 text-brand-primary" />
                Выберите размер
              </h3>
              <div className="space-y-2">
                {sizes.map((size) => (
                  <button
                    key={size.key}
                    onClick={() => setSelectedSize(size.key)}
                    className={`w-full text-left p-3 rounded-lg transition-all ${
                      selectedSize === size.key
                        ? 'bg-brand-primary text-white'
                        : 'bg-gray-50 hover:bg-gray-100 text-gray-900'
                    }`}
                  >
                    <div className="font-semibold">{size.title}</div>
                    <div className={`text-sm ${selectedSize === size.key ? 'text-white/80' : 'text-gray-600'}`}>
                      {size.description}
                    </div>
                    <div className={`text-xs mt-1 ${selectedSize === size.key ? 'text-white/70' : 'text-brand-primary'}`}>
                      {size.stones}
                    </div>
                  </button>
                ))}
              </div>
            </motion.div>

            {/* Выбор стиля */}
            <motion.div
              initial={{ opacity: 0, x: -20 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ duration: 0.5, delay: 0.2 }}
              className="bg-white rounded-2xl shadow-lg p-6"
            >
              <h3 className="text-lg font-bold text-gray-900 mb-4 flex items-center">
                <Palette className="w-5 h-5 mr-2 text-brand-secondary" />
                Выберите стиль
              </h3>
              <div className="space-y-2">
                {styles.map((style) => (
                  <button
                    key={style.key}
                    onClick={() => setSelectedStyle(style.key)}
                    className={`w-full text-left p-3 rounded-lg transition-all ${
                      selectedStyle === style.key
                        ? 'bg-brand-secondary text-white'
                        : 'bg-gray-50 hover:bg-gray-100 text-gray-900'
                    }`}
                  >
                    <div className="font-semibold">{style.title}</div>
                    <div className={`text-sm ${selectedStyle === style.key ? 'text-white/80' : 'text-gray-600'}`}>
                      {style.description}
                    </div>
                    <div className={`text-xs mt-1 ${selectedStyle === style.key ? 'text-white/70' : 'text-brand-secondary'}`}>
                      {style.colors}
                    </div>
                  </button>
                ))}
              </div>
            </motion.div>
          </div>

          {/* Центральная колонка - превью */}
          <div className="lg:col-span-2">
            <motion.div
              initial={{ opacity: 0, scale: 0.95 }}
              animate={{ opacity: 1, scale: 1 }}
              transition={{ duration: 0.5, delay: 0.3 }}
              className="bg-white rounded-2xl shadow-lg p-6"
            >
              <h3 className="text-lg font-bold text-gray-900 mb-4 flex items-center">
                <Eye className="w-5 h-5 mr-2 text-brand-primary" />
                Превью мозаики
              </h3>
              
              <div className="relative aspect-square bg-gray-100 rounded-xl overflow-hidden mb-6">
                {generatePreviewMutation.isPending ? (
                  <div className="absolute inset-0 flex items-center justify-center">
                    <Loader2 className="w-12 h-12 text-brand-primary animate-spin" />
                    <span className="ml-3 text-gray-600">Создаём превью...</span>
                  </div>
                ) : previewImage ? (
                  <img
                    src={previewImage}
                    alt="Превью мозаики"
                    className="w-full h-full object-cover"
                  />
                ) : (
                  <div className="absolute inset-0 flex items-center justify-center">
                    <div className="text-center">
                      <Eye className="w-16 h-16 text-gray-400 mx-auto mb-4" />
                      <p className="text-gray-500">Превью будет здесь</p>
                    </div>
                  </div>
                )}
              </div>

              {/* Кнопка активации купона */}
              {!showActivation ? (
                <button
                  onClick={() => setShowActivation(true)}
                  className="w-full bg-gradient-to-r from-brand-primary to-brand-secondary text-white py-4 rounded-xl hover:from-brand-primary/90 hover:to-brand-secondary/90 font-semibold text-lg transition-all duration-200 flex items-center justify-center"
                >
                  <span>Мне нравится! Хочу создать такую мозаику</span>
                  <ArrowRight className="w-5 h-5 ml-2" />
                </button>
              ) : (
                <div className="space-y-4">
                  <p className="text-gray-600 text-center">
                    Введите код купона для создания мозаики с выбранными параметрами
                  </p>
                  <input
                    type="text"
                    value={couponCode}
                    onChange={handleCouponInput}
                    placeholder="XXXX-XXXX-XXXX"
                    className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-brand-primary focus:border-transparent text-center text-lg tracking-wider"
                    maxLength={14}
                  />
                  <button
                    onClick={handleActivateCoupon}
                    disabled={couponCode.replace(/-/g, '').length !== 12 || activateCouponMutation.isPending}
                    className="w-full bg-brand-primary text-white py-3 rounded-lg hover:bg-brand-primary/90 disabled:opacity-50 disabled:cursor-not-allowed font-semibold transition-all duration-200 flex items-center justify-center"
                  >
                    {activateCouponMutation.isPending ? (
                      <Loader2 className="w-5 h-5 animate-spin" />
                    ) : (
                      <>
                        <span>Активировать купон и создать мозаику</span>
                        <ArrowRight className="w-5 h-5 ml-2" />
                      </>
                    )}
                  </button>
                </div>
              )}
            </motion.div>

            {/* Ссылки на маркетплейсы */}
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.5, delay: 0.4 }}
              className="bg-white rounded-2xl shadow-lg p-6 mt-6"
            >
              <h3 className="text-lg font-bold text-gray-900 mb-4 flex items-center">
                <ShoppingCart className="w-5 h-5 mr-2 text-brand-secondary" />
                Где купить набор
              </h3>
              
              <div className="grid sm:grid-cols-2 gap-4">
                {/* OZON */}
                <a
                  href={getMarketplaceLink('ozon') || 'https://www.ozon.ru/search/?text=алмазная+мозаика'}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="flex items-center justify-center px-4 py-3 bg-gradient-to-r from-orange-500 to-red-500 text-white rounded-lg hover:from-orange-600 hover:to-red-600 font-semibold transition-all duration-200"
                >
                  <span>Купить на OZON</span>
                  <ArrowRight className="w-4 h-4 ml-2" />
                </a>
                
                {/* Wildberries */}
                <a
                  href={getMarketplaceLink('wildberries') || 'https://www.wildberries.ru/catalog/0/search.aspx?search=алмазная+мозаика'}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="flex items-center justify-center px-4 py-3 bg-gradient-to-r from-purple-500 to-pink-500 text-white rounded-lg hover:from-purple-600 hover:to-pink-600 font-semibold transition-all duration-200"
                >
                  <span>Купить на Wildberries</span>
                  <ArrowRight className="w-4 h-4 ml-2" />
                </a>
              </div>
              
              <p className="text-gray-500 text-sm mt-4 text-center">
                После покупки набора вы получите код для создания уникальной схемы
              </p>
            </motion.div>
          </div>
        </div>
      </div>
    </div>
  )
}

export default MosaicPreviewPage