import React, { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { motion } from 'framer-motion'
import { Ticket, ShoppingCart, ArrowRight } from 'lucide-react'
import { useMutation } from '@tanstack/react-query'
import { MosaicAPI } from '../../api/client'
import { useUIStore, usePartnerStore } from '../../store/partnerStore'
import { useNavigate, useLocation } from 'react-router-dom'

const HeroSection = () => {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const location = useLocation()
  const { addNotification } = useUIStore()
  const { partner } = usePartnerStore()
  const [couponCode, setCouponCode] = useState('')
  const [email, setEmail] = useState('')
  
  // Check for coupon in URL params (for white-label redirect)
  React.useEffect(() => {
    const urlParams = new URLSearchParams(location.search)
    const couponFromUrl = urlParams.get('coupon')
    if (couponFromUrl) {
      // Format the coupon code
      const cleanCode = couponFromUrl.replace(/-/g, '')
      if (cleanCode.length === 12) {
        const formattedCode = cleanCode.substring(0, 4) + '-' + cleanCode.substring(4, 8) + '-' + cleanCode.substring(8)
        setCouponCode(formattedCode)
        // Auto-activate the coupon
        setTimeout(() => {
          activateCouponMutation.mutate(cleanCode)
        }, 500)
      }
    }
  }, [location.search])

  const activateCouponMutation = useMutation({
    mutationFn: async (code) => {
      // Не бросаем локальные ошибки по флагам valid, опираемся только на HTTP-статусы
      const info = await MosaicAPI.validateCoupon(code)
      
      // Check if coupon belongs to a partner with a different domain
      if (info?.partner_domain && window.location.hostname !== info.partner_domain) {
        // Redirect to partner's white-label site with the coupon
        const protocol = window.location.protocol
        const redirectUrl = `${protocol}//${info.partner_domain}/?coupon=${code}`
        window.location.href = redirectUrl
        return info // Return early to prevent further processing
      }
      
      await MosaicAPI.activateCoupon(code)
      return info || {}
    },
    onSuccess: (couponData) => {
      // Don't process if we're redirecting to partner site
      if (couponData?.partner_domain && window.location.hostname !== couponData.partner_domain) {
        return
      }
      
      // Показываем уведомление об успешной активации
      addNotification({
        type: 'success',
        message: t('notifications.coupon_ready_for_use')
      })
      
      // Даем небольшую задержку, чтобы уведомление успело показаться
      setTimeout(() => {
        const size = couponData?.size || '30x40'
        const style = couponData?.style || 'max_colors'
        navigate(`/editor?coupon=${couponCode.replace(/-/g, '')}&size=${size}&style=${style}`)
      }, 500)
    },
    onError: (error) => {
      const msg = error?.status === 404
        ? t('notifications.invalid_coupon')
        : error?.status === 409
        ? t('notifications.activation_error')
        : t('notifications.activation_error')
      addNotification({ type: 'error', title: t('notifications.activation_error'), message: msg })
    }
  })

  const handleCouponSubmit = (e) => {
    e.preventDefault()
    
    if (!couponCode) {
      addNotification({
        type: 'error',
        message: t('notifications.coupon_required')
      })
      return
    }
    
    if (!/^\d{4}-\d{4}-\d{4}$/.test(couponCode)) {
      addNotification({
        type: 'error',
        message: t('notifications.invalid_coupon_format')
      })
      return
    }
    
    // Проверяем, не выполняется ли уже мутация
    if (activateCouponMutation.isPending) {
      return
    }
    
    // Очищаем код от дефисов для отправки на бэкенд
    const cleanCode = couponCode.replace(/-/g, '')
    
    // Проверяем, что получилось ровно 12 цифр
    if (cleanCode.length !== 12) {
      addNotification({
        type: 'error',
        message: t('notifications.invalid_coupon_format')
      })
      return
    }
    
    activateCouponMutation.mutate(cleanCode)
  }

  const handleCouponInput = (e) => {
    // Убираем все нецифровые символы и ограничиваем до 12 цифр
    const digitsOnly = e.target.value.replace(/[^0-9]/g, '').substring(0, 12)
    
    // Форматируем код с дефисами: XXXX-XXXX-XXXX
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

  const goToShop = () => {
    navigate('/diamond-art')
  }

  const containerVariants = {
    hidden: { opacity: 0 },
    visible: {
      opacity: 1,
      transition: {
        staggerChildren: 0.2,
        duration: 0.6
      }
    }
  }

  const itemVariants = {
    hidden: { opacity: 0, y: 20 },
    visible: { opacity: 1, y: 0 }
  }

  return (
    <section className="relative overflow-hidden bg-gradient-to-br from-blue-50 via-purple-50 to-pink-50 py-12 sm:py-16 lg:py-20">
      <div className="absolute inset-0 opacity-40">
        <div className="w-full h-full" style={{
          backgroundImage: `url("data:image/svg+xml,%3Csvg width='60' height='60' viewBox='0 0 60 60' xmlns='http://www.w3.org/2000/svg'%3E%3Cg fill='none' fill-rule='evenodd'%3E%3Cg fill='%239C92AC' fill-opacity='0.05'%3E%3Ccircle cx='30' cy='30' r='3'/%3E%3C/g%3E%3C/g%3E%3C/svg%3E")`
        }} />
      </div>
      
      <div className="relative max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <motion.div 
          variants={containerVariants}
          initial="hidden"
          animate="visible"
          className="text-center"
        >
          <motion.h1 
            variants={itemVariants}
            className="text-3xl sm:text-4xl md:text-5xl lg:text-6xl font-bold text-gray-900 mb-4 sm:mb-6 leading-tight px-2"
          >
            {t('hero.title')}
          </motion.h1>
          
          <motion.p 
            variants={itemVariants}
            className="text-lg sm:text-xl md:text-2xl text-gray-600 mb-8 sm:mb-10 lg:mb-12 max-w-3xl mx-auto px-4"
          >
            {t('hero.subtitle')}
          </motion.p>

          <motion.div 
            variants={itemVariants}
            className="grid grid-cols-1 lg:grid-cols-2 gap-6 sm:gap-8 max-w-5xl mx-auto"
          >
            {/* Coupon Activation Card */}
            <div className="bg-white/80 backdrop-blur-sm rounded-2xl shadow-xl p-6 sm:p-8 border border-white/20 mx-4 lg:mx-0 flex flex-col h-full min-h-[400px]">
              <div className="flex items-center justify-center w-14 h-14 sm:w-16 sm:h-16 bg-brand-primary/10 rounded-full mx-auto mb-4 sm:mb-6">
                <Ticket className="w-7 h-7 sm:w-8 sm:h-8 text-brand-primary" />
              </div>
              
              <h3 className="text-xl sm:text-2xl font-bold text-gray-900 mb-4 px-2">
                {t('hero.coupon_banner.title')}
              </h3>
              
              <p className="text-gray-600 mb-6 px-2 text-sm sm:text-base flex-grow">
                {t('hero.coupon_banner.description')}
              </p>
              
              <form onSubmit={handleCouponSubmit} className="space-y-4 mt-auto">
                <div>
                  <input
                    type="text"
                    value={couponCode}
                    onChange={handleCouponInput}
                    placeholder={t('hero.coupon_banner.placeholder')}
                    className="w-full px-3 sm:px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-brand-primary focus:border-transparent text-center text-base sm:text-lg tracking-wider"
                    maxLength={14}
                  />
                </div>
                
                <button
                  type="submit"
                  disabled={activateCouponMutation.isPending}
                  className="w-full bg-brand-primary text-white py-3 px-4 sm:px-6 rounded-lg hover:bg-brand-primary/90 disabled:opacity-50 disabled:cursor-not-allowed font-semibold text-base sm:text-lg transition-all duration-200 flex items-center justify-center space-x-2 focus:ring-2 focus:ring-brand-primary focus:ring-offset-2 min-h-[48px]"
                >
                  {activateCouponMutation.isPending ? (
                    <div className="w-5 h-5 border-2 border-white border-t-transparent rounded-full animate-spin" />
                  ) : (
                    <>
                      <span>{t('hero.coupon_banner.activate')}</span>
                      <ArrowRight className="w-4 h-4 sm:w-5 sm:h-5" />
                    </>
                  )}
                </button>
              </form>
            </div>

            {/* Shop Card */}
            <div className="bg-white/80 backdrop-blur-sm rounded-2xl shadow-xl p-6 sm:p-8 border border-white/20 mx-4 lg:mx-0 flex flex-col h-full min-h-[400px]">
              <div className="flex items-center justify-center w-14 h-14 sm:w-16 sm:h-16 bg-brand-secondary/10 rounded-full mx-auto mb-4 sm:mb-6">
                <ShoppingCart className="w-7 h-7 sm:w-8 sm:h-8 text-brand-secondary" />
              </div>
              
              <h3 className="text-xl sm:text-2xl font-bold text-gray-900 mb-4 px-2">
                {t('hero.shop_banner.title')}
              </h3>
              
              <p className="text-gray-600 mb-6 px-2 text-sm sm:text-base flex-grow">
                {t('hero.shop_banner.description')}
              </p>
              
              <button
                onClick={goToShop}
                className="w-full bg-brand-secondary text-white py-3 px-4 sm:px-6 rounded-lg hover:bg-brand-secondary/90 font-semibold text-base sm:text-lg transition-all duration-200 flex items-center justify-center space-x-2 focus:ring-2 focus:ring-brand-secondary focus:ring-offset-2 min-h-[48px] mt-auto"
              >
                <ShoppingCart className="w-4 h-4 sm:w-5 sm:h-5" />
                <span>{t('hero.shop_banner.button')}</span>
              </button>
            </div>
          </motion.div>
        </motion.div>
      </div>
    </section>
  )
}

export default HeroSection