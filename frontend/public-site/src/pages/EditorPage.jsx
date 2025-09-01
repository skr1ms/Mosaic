import React from 'react'
import { useTranslation } from 'react-i18next'
import { useSearchParams } from 'react-router-dom'
import { motion } from 'framer-motion'
import EditorSteps from '../components/editor/EditorSteps'
import CouponInfo from '../components/editor/CouponInfo'
import { useCouponStore } from '../store/partnerStore'
import { MosaicAPI } from '../api/client'
import { useUIStore } from '../store/partnerStore'

const EditorPage = () => {
  const { t } = useTranslation()
  const [searchParams, setSearchParams] = useSearchParams()
  const { coupon, setCoupon } = useCouponStore()
  const { addNotification } = useUIStore()
  const lastCheckedCodeRef = React.useRef(null)

  React.useEffect(() => {
    // ОЧИЩАЕМ localStorage ОТ ДАННЫХ ПОКУПОК И ПРЕВЬЮ ПРИ ВХОДЕ В РЕДАКТОР
    try {
      localStorage.removeItem('pendingOrder')
      
      // Очищаем временные данные, но НЕ трогаем activeCoupon (может понадобиться)
      const keys = Object.keys(localStorage)
      keys.forEach(key => {
        if (key.startsWith('preview_') || key.startsWith('temp_') || key.startsWith('shop_')) {
          localStorage.removeItem(key)
        }
      })
      
      console.log('Cleared stale localStorage data in editor')
    } catch (error) {
      console.error('Error clearing localStorage in editor:', error)
    }
    
    const couponCode = searchParams.get('coupon')
    const size = searchParams.get('size')
    const style = searchParams.get('style')

    // Логируем для отладки
    const image = searchParams.get('image')
    const step = parseInt(searchParams.get('step') || '1', 10)

    const cleanedUrlCode = (couponCode || '').replace(/\D/g, '')
    const cleanedStoreCode = (coupon?.code || '').replace(/\D/g, '')

    // Если в URL есть купон и он отличается от текущего в сторе — принудительно переустанавливаем
    if (cleanedUrlCode && cleanedUrlCode !== cleanedStoreCode) {
      setCoupon({
        code: couponCode,
        size: size || 'unknown',
        style: style || 'unknown'
      })
      try { sessionStorage.setItem('editor:coupon', couponCode) } catch {}

      // Новый купон: очищаем старые параметры image/step в URL и ворк-ключи с прошлой сессии
      const params = new URLSearchParams(searchParams)
      params.delete('image')
      params.delete('step')
      setSearchParams(params)
      try {
        // Удаляем универсальные ключи прошлого изображения
        sessionStorage.removeItem('editor:lastImageId')
        sessionStorage.removeItem('editor:lastPreview')
        // Чистим подтверждение предыдущей схемы и выбранные опции, правки и сохраненные данные схемы
        Object.keys(sessionStorage).forEach((k) => {
          if (
            k.startsWith('editor:confirmed:') ||
            k.startsWith('editor:selectedOptions:') ||
            k.startsWith('editor:step:') ||
            k.startsWith('editor:lastPreview:') ||
            k.startsWith('editor:schemaData:') ||
            k.startsWith('editor:edits:')
          ) {
            sessionStorage.removeItem(k)
          }
        })
      } catch {}
    }

    // Если купон уже есть в сторе, дублируем в sessionStorage
    if (coupon?.code) {
      try { sessionStorage.setItem('editor:coupon', coupon.code) } catch {}
    }
  }, [searchParams, coupon, setCoupon, setSearchParams])

  // Автоматически валидируем купон, если пришли напрямую по ссылке
  React.useEffect(() => {
    const validateCoupon = async () => {
      if (!coupon?.code) return
      // Если уже есть image в URL (успешная загрузка), не шумим уведомлениями
      const imageId = searchParams.get('image')
      if (imageId) return
      if (lastCheckedCodeRef.current === coupon.code) return

      // Если в URL есть size и style, значит мы пришли из HeroSection и купон уже активирован
      const urlSize = searchParams.get('size')
      const urlStyle = searchParams.get('style')
      if (urlSize && urlStyle) {
        lastCheckedCodeRef.current = coupon.code
        return
      }

      lastCheckedCodeRef.current = coupon.code

      try {
        // Очищаем все не-цифры, ожидаем ровно 12 цифр
        const cleanCode = (coupon.code || '').replace(/\D/g, '')
        if (cleanCode.length !== 12) return

        const info = await MosaicAPI.validateCoupon(cleanCode)

        // Показываем уведомления только для ошибок
        if (!info.valid) {
          addNotification({
            type: 'error',
            title: t('notifications.activation_error'),
            message: t('notifications.invalid_coupon')
          })
          return
        } else if (info.status === 'used') {
          // Купон уже активирован, просто обновляем данные без уведомлений
          setCoupon({
            code: coupon.code,
            size: info.size || coupon.size || 'unknown',
            style: info.style || coupon.style || 'unknown'
          })
          return
        }

        // Если валидация прошла успешно и купон еще не активирован, активируем купон
        try {
          const activationResult = await MosaicAPI.activateCoupon(cleanCode)
          
          // Обновим известные параметры с данными активации
          setCoupon({
            code: coupon.code,
            size: activationResult.size || info.size || coupon.size || 'unknown',
            style: activationResult.style || info.style || coupon.style || 'unknown'
          })

          // Проверяем, был ли купон уже активирован
          if (activationResult.message === "Купон уже активирован") {
          } else {
          }
        } catch (activationError) {
          console.error('EditorPage: Activation failed:', activationError)
          
          // Если активация не удалась, но валидация прошла, показываем информацию о валидации
          setCoupon({
            code: coupon.code,
            size: info.size || coupon.size || 'unknown',
            style: info.style || coupon.style || 'unknown'
          })

          // Показываем ошибку активации
          addNotification({
            type: 'error',
            title: t('notifications.activation_error'),
            message: activationError.message || t('notifications.activation_error')
          })
        }
      } catch (e) {
        console.error('EditorPage: Validation error:', e)
        // Игнорируем ошибки, если формат не тот или временные сети; не дублируем сообщения при успешной загрузке
        const status = e?.status
        if (status === 404) {
          addNotification({
            type: 'error',
            title: t('notifications.activation_error'),
            message: t('notifications.invalid_coupon')
          })
        }
      }
    }
    validateCoupon()
  }, [coupon?.code, setCoupon, addNotification, t, searchParams])

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6 }}
        >
          <h1 className="text-3xl font-bold text-gray-900 mb-8">
            {t('editor.title')}
          </h1>

          {coupon && <CouponInfo coupon={coupon} />}
          
          <EditorSteps
            couponCode={coupon?.code}
            couponSize={coupon?.size}
            initialImageId={searchParams.get('image') || null}
            initialStep={parseInt(searchParams.get('step') || '1', 10)}
          />
        </motion.div>
      </div>
    </div>
  )
}

export default EditorPage
