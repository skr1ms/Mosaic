import React, { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { motion } from 'framer-motion'
import { ShoppingBag, Check, ExternalLink, CreditCard, Building2 } from 'lucide-react'
import { useMutation } from '@tanstack/react-query'
import { MosaicAPI } from '../api/client'
import { useUIStore } from '../store/partnerStore'


const ShopPage = () => {
  const { t } = useTranslation()
  
  // Восстанавлием состояние из sessionStorage
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
  const [couponCode, setCouponCode] = useState(null)
  const [orderData, setOrderData] = useState(null)
  
  // Проверяем сохраненный заказ при загрузке
  useEffect(() => {
    try {
      const savedOrder = localStorage.getItem('pendingOrder')
      if (savedOrder) {
        const order = JSON.parse(savedOrder)
        // Проверяем что заказ не старше 30 минут
        const thirtyMinutes = 30 * 60 * 1000
        if (Date.now() - order.timestamp < thirtyMinutes) {
          // ДОПОЛНИТЕЛЬНАЯ ПРОВЕРКА: пытаемся получить статус заказа
          // Если заказ не существует на сервере - очищаем localStorage
          MosaicAPI.getOrderStatus(order.orderNumber)
            .then(() => {
              // Заказ существует - восстанавливаем данные
              setOrderNumber(order.orderNumber)
              setPaymentUrl(order.paymentUrl)
              setEmail(order.email)
              setSelectedSize(order.size)
              setSelectedStyle(order.style)
              console.log('Restored pending order:', order.orderNumber)
            })
            .catch(() => {
              // Заказ не существует - очищаем localStorage
              console.log('Saved order not found on server, clearing localStorage')
              localStorage.removeItem('pendingOrder')
              localStorage.removeItem('activeCoupon')
            })
        } else {
          // Удаляем старый заказ
          localStorage.removeItem('pendingOrder')
        }
      }
    } catch (error) {
      console.error('Failed to restore pending order:', error)
      // При ошибке тоже очищаем localStorage
      localStorage.removeItem('pendingOrder')
      localStorage.removeItem('activeCoupon')
    }
  }, [])
  
  const { addNotification } = useUIStore()
  

  
  // АВТОМАТИЧЕСКАЯ ПРОВЕРКА СТАТУСА ЗАКАЗА
  useEffect(() => {
    if (!orderNumber || paymentStatus) return
    
    let intervalId = null
    let attempts = 0
    const maxAttempts = 60 // 60 попыток = 3 минуты
    
    const checkOrderStatus = async () => {
      try {
        attempts++
        console.log(`Checking order status... Attempt ${attempts}/${maxAttempts}`)
        
        const statusData = await MosaicAPI.getOrderStatus(orderNumber)
        
        if (statusData.status === 'paid') {
          setPaymentStatus('success')
          clearInterval(intervalId)
          
          // ЕСЛИ ЕСТЬ КУПОН - ПОКАЗЫВАЕМ ЕГО НА СТРАНИЦЕ!
          if (statusData.coupon_code) {
            setCouponCode(statusData.coupon_code)
            setOrderData(statusData)
            localStorage.setItem('activeCoupon', statusData.coupon_code)
            localStorage.removeItem('pendingOrder') // Очищаем pending order
            
            addNotification({ 
              type: 'success', 
              title: t('shop.alerts.coupon_ready_title'), 
              message: t('shop.alerts.coupon_generating') 
            })
          } else {
                      addNotification({
            type: 'success',
            title: t('shop.alerts.payment_success'),
            message: t('shop.alerts.coupon_generating')
          })
          }
        } else if (statusData.status === 'failed' || statusData.status === 'cancelled') {
          setPaymentStatus('fail')
          clearInterval(intervalId)
                    addNotification({
            type: 'error',
            title: t('shop.alerts.payment_error'),
            message: t('shop.alerts.payment_failed')
          })
        } else if (attempts >= maxAttempts) {
          // Превысили максимум попыток
          clearInterval(intervalId)
                    addNotification({
            type: 'warning',
            title: t('shop.alerts.status_check_complete'),
            message: t('shop.alerts.status_check_message')
          })
        }
      } catch (error) {
        console.error('Failed to check order status:', error)
        if (attempts >= maxAttempts) {
          clearInterval(intervalId)
        }
      }
    }
    
    // Начинаем проверку через 3 секунды, затем каждые 3 секунды
    const initialTimer = setTimeout(() => {
      checkOrderStatus() // Первая проверка
      intervalId = setInterval(checkOrderStatus, 3000) // Последующие проверки каждые 3 сек
    }, 3000)
    
    return () => {
      clearTimeout(initialTimer)
      if (intervalId) clearInterval(intervalId)
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
          
          // СОХРАНЯЕМ ДАННЫЕ ЗАКАЗА В LOCALSTORAGE
          try {
            localStorage.setItem('pendingOrder', JSON.stringify({
              orderNumber: data.order_number,
              orderId: data.order_id,
              size: selectedSize,
              style: selectedStyle,
              email: email,
              amount: data.amount,
              timestamp: Date.now(),
              paymentUrl: data.payment_url
            }))
          } catch (error) {
            console.error('Failed to save order to localStorage:', error)
          }
        }
        // Открываем страницу оплаты в новой вкладке
        window.open(data.payment_url, '_blank')
        setPaymentUrl(data.payment_url)
                addNotification({
          type: 'success',
          title: t('shop.alerts.order_created'),
          message: t('shop.alerts.order_created_message')
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
  const styleKeys = ['grayscale', 'skin_tone', 'pop_art', 'max_colors']

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
      alert(t('shop.alerts.email_required'))
      return
    }
    
    let finalStyle = selectedStyle
    
    const payload = {
      size: selectedSize,
      style: finalStyle,
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
                'max_colors': 'linear-gradient(135deg, #FF0080, #00FF41, #0099FF, #FFD700)'
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
            {t('shop.email_label')}
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
          
          {/* Показываем информацию о том, что проверка автоматическая */}
          {orderNumber && !couponCode && (
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              className="mt-4 p-4 bg-blue-50 border border-blue-200 rounded-lg"
            >
              <div className="text-center">
                <div className="animate-spin w-6 h-6 border-2 border-blue-600 border-t-transparent rounded-full mx-auto mb-2"></div>
                <p className="text-blue-800 font-semibold">{t('shop.payment.checking_status')}</p>
                <p className="text-sm text-blue-600 mt-1">{t('shop.payment.auto_check_info')}</p>
              </div>
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

        {/* ОТОБРАЖЕНИЕ ГОТОВОГО КУПОНА */}
        {couponCode && orderData && (
          <motion.section
            initial={{ opacity: 0, scale: 0.9 }}
            animate={{ opacity: 1, scale: 1 }}
            transition={{ duration: 0.5 }}
            className="mt-12"
          >
            <div className="bg-gradient-to-r from-green-50 to-emerald-50 border-2 border-green-200 rounded-2xl p-8 text-center shadow-lg">
              <div className="text-4xl mb-4">🎉</div>
              <h2 className="text-3xl font-bold text-green-800 mb-4">
                {t('shop.alerts.coupon_ready_title')}
              </h2>
              
              <div className="bg-white border-2 border-dashed border-green-300 rounded-xl p-6 mb-6">
                <div className="text-sm text-gray-600 mb-2">{t('shop.alerts.coupon_code_label')}</div>
                <div className="text-3xl font-mono font-bold text-green-700 tracking-wider mb-4">
                  {couponCode}
                </div>
                <button
                  onClick={() => {
                    navigator.clipboard.writeText(couponCode)
                            addNotification({
          type: 'success',
          message: t('shop.alerts.coupon_copied')
        })
                  }}
                  className="bg-green-600 text-white px-6 py-2 rounded-lg hover:bg-green-700 transition-colors font-semibold"
                >
                  {t('shop.alerts.copy_code_button')}
                </button>
              </div>

              <div className="grid grid-cols-2 gap-4 mb-6 text-left">
                <div className="bg-white rounded-lg p-4">
                  <div className="text-sm text-gray-600">{t('shop.alerts.size_label')}</div>
                  <div className="font-semibold text-gray-800">{orderData.size} {t('common.cm')}</div>
                </div>
                <div className="bg-white rounded-lg p-4">
                  <div className="text-sm text-gray-600">{t('shop.alerts.style_label')}</div>
                  <div className="font-semibold text-gray-800">
                    {t(`shop.styles.${orderData.style}.title`)}
                  </div>
                </div>
              </div>

              <div className="flex flex-col sm:flex-row gap-4 justify-center">
                <button
                  onClick={() => {
                    window.location.href = `/editor?coupon=${couponCode}&size=${orderData.size}&style=${orderData.style}`
                  }}
                  className="bg-brand-primary text-white px-8 py-3 rounded-lg hover:bg-brand-primary/90 transition-colors font-semibold text-lg"
                >
                  {t('shop.alerts.create_schema_now')}
                </button>
                <button
                  onClick={() => {
                    window.location.href = '/'
                  }}
                  className="bg-gray-600 text-white px-8 py-3 rounded-lg hover:bg-gray-700 transition-colors font-semibold"
                >
                  {t('shop.alerts.go_home_button')}
                </button>
              </div>

              <div className="mt-6 text-sm text-gray-600">
                <p>{t('shop.alerts.email_sent_info')} <strong>{email}</strong></p>
                <p>{t('shop.alerts.save_code_info')}</p>
              </div>
            </div>
          </motion.section>
        )}
      </div>
    </div>
  )
}

export default ShopPage