import React, { useState, useEffect } from 'react'
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
  
  const [email, setEmail] = useState('')
  const [paymentStatus, setPaymentStatus] = useState(null)
  const [paymentUrl, setPaymentUrl] = useState(null)
  const [orderNumber, setOrderNumber] = useState(null)
  
  const { addNotification } = useUIStore()
  

  
  // Обрабатываем возврат с оплаты и проверяем статус заказа
  useEffect(() => {
    // Если есть номер заказа, проверяем его статус
    if (orderNumber && !paymentStatus) {
      const checkOrderStatus = async () => {
        try {
          const statusData = await MosaicAPI.getOrderStatus(orderNumber)
          if (statusData.status === 'paid') {
            setPaymentStatus('success')
            addNotification({ 
              type: 'success', 
              title: 'Оплата прошла успешно!', 
              message: 'Ваш купон будет отправлен на email в течение 5 минут.' 
            })
          } else if (statusData.status === 'failed' || statusData.status === 'cancelled') {
            setPaymentStatus('fail')
            addNotification({ 
              type: 'error', 
              title: 'Ошибка оплаты', 
              message: 'Попробуйте еще раз или обратитесь в поддержку.' 
            })
          }
        } catch (error) {
          console.error('Failed to check order status:', error)
        }
      }
      
      // Проверяем статус через 2 секунды после возврата
      const timer = setTimeout(checkOrderStatus, 2000)
      return () => clearTimeout(timer)
    }
  }, [orderNumber, paymentStatus, addNotification])

  const createOrderMutation = useMutation({
    mutationFn: (payload) => MosaicAPI.initiateCouponOrder(payload),
    onSuccess: (data) => {
      console.log('Payment response:', data)
      if (data?.payment_url) {
        // Сохраняем номер заказа для проверки статуса
        if (data.order_number) {
          setOrderNumber(data.order_number)
        }
        // Открываем страницу оплаты в новой вкладке
        window.open(data.payment_url, '_blank')
        setPaymentUrl(data.payment_url)
        addNotification({ 
          type: 'success', 
          title: 'Заказ создан!', 
          message: 'Откройте страницу оплаты в новой вкладке.' 
        })
      } else {
        addNotification({ type: 'error', title: t('payment.error_title'), message: t('payment.purchase_failed') })
      }
    },
    onError: (err) => {
      console.error('Payment error:', err)
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
    if (!email) {
      alert('Пожалуйста, введите email')
      return
    }
    const payload = {
      size: selectedSize,
      style: selectedStyle,
      email: email,
      return_url: window.location.origin + '/shop',
      fail_url: window.location.origin + '/shop',
      language: 'ru'
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

        {/* Email Input */}
        <motion.div
          {...fadeInUp}
          transition={{ delay: 0.5 }}
          className="max-w-md mx-auto mb-8"
        >
          <label htmlFor="email" className="block text-sm font-medium text-gray-700 mb-2">
            Email для получения купона
          </label>
          <input
            type="email"
            id="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder="your@email.com"
            className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-brand-primary focus:border-transparent"
            required
          />
        </motion.div>

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
            disabled={!selectedSize || !email}
            className={`px-12 py-4 text-lg font-semibold rounded-xl transition-all duration-300 shadow-lg transform ${
              selectedSize && email
                ? 'bg-brand-primary text-white hover:bg-brand-primary/90 hover:shadow-xl hover:scale-105' 
                : 'bg-gray-300 text-gray-500 cursor-not-allowed'
            }`}
          >
            <div className="flex items-center justify-center space-x-2">
              <CreditCard className="w-5 h-5" />
              <span>{t('shop.proceed_to_payment')}</span>
            </div>
          </button>
          
          {/* Payment URL Button - показывается после создания заказа */}
          {paymentUrl && (
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              className="mt-4"
            >
              <button
                onClick={() => window.open(paymentUrl, '_blank')}
                className="px-8 py-3 bg-green-600 text-white font-semibold rounded-lg hover:bg-green-700 transition-all duration-300 shadow-lg"
              >
                <div className="flex items-center justify-center space-x-2">
                  <ExternalLink className="w-4 h-4" />
                  <span>Перейти к оплате</span>
                </div>
              </button>
              <p className="text-sm text-gray-600 mt-2">
                Если страница оплаты не открылась автоматически, нажмите эту кнопку
              </p>
            </motion.div>
          )}
          
          {/* Check Order Status Button - показывается если есть номер заказа */}
          {orderNumber && !paymentStatus && (
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              className="mt-4"
            >
              <button
                onClick={async () => {
                  try {
                    const statusData = await MosaicAPI.getOrderStatus(orderNumber)
                    if (statusData.status === 'paid') {
                      setPaymentStatus('success')
                      addNotification({ 
                        type: 'success', 
                        title: 'Оплата прошла успешно!', 
                        message: 'Ваш купон будет отправлен на email в течение 5 минут.' 
                      })
                    } else if (statusData.status === 'failed' || statusData.status === 'cancelled') {
                      setPaymentStatus('fail')
                      addNotification({ 
                        type: 'error', 
                        title: 'Ошибка оплаты', 
                        message: 'Попробуйте еще раз или обратитесь в поддержку.' 
                      })
                    } else {
                      addNotification({ 
                        type: 'info', 
                        title: 'Статус заказа', 
                        message: `Статус: ${statusData.status}` 
                      })
                    }
                  } catch (error) {
                    console.error('Failed to check order status:', error)
                    addNotification({ 
                      type: 'error', 
                      title: 'Ошибка', 
                      message: 'Не удалось проверить статус заказа' 
                    })
                  }
                }}
                className="px-8 py-3 bg-blue-600 text-white font-semibold rounded-lg hover:bg-blue-700 transition-all duration-300 shadow-lg"
              >
                <div className="flex items-center justify-center space-x-2">
                  <Check className="w-4 h-4" />
                  <span>Проверить статус оплаты</span>
                </div>
              </button>
              <p className="text-sm text-gray-600 mt-2">
                Нажмите эту кнопку после завершения оплаты для проверки статуса
              </p>
            </motion.div>
          )}
          
          {/* Payment Status Display */}
          {paymentStatus && (
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              className={`mt-4 p-4 rounded-lg ${
                paymentStatus === 'success' 
                  ? 'bg-green-100 border border-green-300 text-green-800' 
                  : 'bg-red-100 border border-red-300 text-red-800'
              }`}
            >
              <div className="flex items-center justify-center space-x-2">
                {paymentStatus === 'success' ? (
                  <Check className="w-5 h-5 text-green-600" />
                ) : (
                  <div className="w-5 h-5 bg-red-600 rounded-full" />
                )}
                <span className="font-semibold">
                  {paymentStatus === 'success' 
                    ? 'Оплата прошла успешно!' 
                    : 'Ошибка оплаты'
                  }
                </span>
              </div>
              <p className="text-center mt-2">
                {paymentStatus === 'success' 
                  ? 'Ваш купон будет отправлен на email в течение 5 минут.' 
                  : 'Попробуйте еще раз или обратитесь в поддержку.'
                }
              </p>
            </motion.div>
          )}
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
      </div>
    </div>
  )
}

export default ShopPage