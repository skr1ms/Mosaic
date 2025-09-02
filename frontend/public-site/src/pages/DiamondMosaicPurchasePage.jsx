import React, { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { motion } from 'framer-motion'
import { ArrowLeft, ShoppingCart, CreditCard, Check, Package, Star } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import { useUIStore } from '../store/partnerStore'

const DiamondMosaicPurchasePage = () => {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { addNotification } = useUIStore()
  
  const [purchaseData, setPurchaseData] = useState(null)
  const [selectedPackage, setSelectedPackage] = useState('')
  const [isProcessing, setIsProcessing] = useState(false)

  // Пакеты для покупки
  const packages = [
    {
      id: 'basic',
      name: 'Базовый',
      price: 299,
      originalPrice: 499,
      features: [
        'Схема алмазной мозаики',
        'Инструкция по сборке',
        'Поддержка 24/7'
      ],
      popular: false
    },
    {
      id: 'premium',
      name: 'Премиум',
      price: 499,
      originalPrice: 799,
      features: [
        'Схема алмазной мозаики',
        'Инструкция по сборке',
        'Поддержка 24/7',
        'HD качество схемы',
        'Дополнительные варианты обработки'
      ],
      popular: true
    },
    {
      id: 'professional',
      name: 'Профессиональный',
      price: 799,
      originalPrice: 1299,
      features: [
        'Схема алмазной мозаики',
        'Инструкция по сборке',
        'Поддержка 24/7',
        'HD качество схемы',
        'Дополнительные варианты обработки',
        'Файлы для печати',
        'Персональная консультация'
      ],
      popular: false
    }
  ]

  useEffect(() => {
    // Загружаем данные для покупки
    try {
      const savedPurchaseData = localStorage.getItem('diamondMosaic_purchaseData')
      
      if (!savedPurchaseData) {
        navigate('/diamond-mosaic')
        return
      }
      
      const parsedData = JSON.parse(savedPurchaseData)
      setPurchaseData(parsedData)
      
      // Устанавливаем премиум пакет по умолчанию
      setSelectedPackage('premium')
      
    } catch (error) {
      console.error('Error loading purchase data:', error)
      navigate('/diamond-mosaic')
    }
  }, [navigate])

  const handlePackageSelect = (packageId) => {
    setSelectedPackage(packageId)
  }

  const handlePurchase = async () => {
    if (!selectedPackage || !purchaseData) {
      addNotification({
        type: 'error',
        message: 'Выберите пакет для покупки'
      })
      return
    }

    setIsProcessing(true)

    try {
      // Здесь должна быть интеграция с платежной системой
      // Пока что имитируем процесс покупки
      
      const selectedPkg = packages.find(pkg => pkg.id === selectedPackage)
      
      // Симулируем процесс покупки
      await new Promise(resolve => setTimeout(resolve, 2000))
      
      // Сохраняем данные о покупке
      const orderData = {
        orderId: `DMO_${Date.now()}`,
        package: selectedPkg,
        imageData: purchaseData,
        purchaseDate: new Date().toISOString(),
        status: 'completed'
      }
      
      localStorage.setItem('diamondMosaic_lastOrder', JSON.stringify(orderData))
      
      // Очищаем временные данные
      localStorage.removeItem('diamondMosaic_purchaseData')
      localStorage.removeItem('diamondMosaic_selectedImage')
      sessionStorage.removeItem('diamondMosaic_fileUrl')
      
      addNotification({
        type: 'success',
        title: 'Покупка завершена!',
        message: 'Купон активирован. Переходим к генерации схемы...'
      })
      
      // Переходим к странице генерации схемы или результата
      navigate('/diamond-mosaic/success', { 
        state: { orderData } 
      })
      
    } catch (error) {
      console.error('Error processing purchase:', error)
      addNotification({
        type: 'error',
        title: 'Ошибка покупки',
        message: 'Не удалось завершить покупку. Попробуйте ещё раз.'
      })
    } finally {
      setIsProcessing(false)
    }
  }

  const handleBack = () => {
    navigate('/diamond-mosaic/preview-album')
  }

  const getStyleTitle = (styleKey) => {
    const styleMap = {
      'max_colors': 'Максимум цветов',
      'pop_art': 'Поп-арт',
      'grayscale': 'Чёрно-белый',
      'skin_tones': 'Телесные тона'
    }
    return styleMap[styleKey] || styleKey
  }

  if (!purchaseData) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-purple-50 via-pink-50 to-blue-50 flex items-center justify-center">
        <div className="text-purple-600">Загрузка...</div>
      </div>
    )
  }

  const selectedPkg = packages.find(pkg => pkg.id === selectedPackage)

  return (
    <div className="min-h-screen bg-gradient-to-br from-purple-50 via-pink-50 to-blue-50 py-8 px-4">
      <div className="container mx-auto max-w-6xl">
        
        {/* Заголовок и навигация */}
        <motion.div
          initial={{ opacity: 0, y: -20 }}
          animate={{ opacity: 1, y: 0 }}
          className="mb-8"
        >
          <button
            onClick={handleBack}
            className="flex items-center text-purple-600 hover:text-purple-700 mb-4 transition-colors"
          >
            <ArrowLeft className="w-5 h-5 mr-2" />
            Назад к превью
          </button>
          
          <div className="text-center">
            <h1 className="text-4xl md:text-5xl font-bold bg-gradient-to-r from-purple-600 to-pink-600 bg-clip-text text-transparent mb-4">
              Покупка купона
            </h1>
            <p className="text-lg text-gray-600">
              Генерация схемы алмазной мозаики
            </p>
          </div>
        </motion.div>

        <div className="grid lg:grid-cols-3 gap-8">
          
          {/* Информация о заказе */}
          <motion.div
            initial={{ opacity: 0, x: -20 }}
            animate={{ opacity: 1, x: 0 }}
            className="lg:col-span-1"
          >
            <div className="bg-white rounded-xl p-6 shadow-lg mb-6">
              <h3 className="text-xl font-semibold text-gray-800 mb-4">Ваш заказ</h3>
              
              {/* Превью изображения */}
              <div className="mb-4">
                <div className="aspect-square bg-gray-100 rounded-lg overflow-hidden mb-2">
                  <img
                    src={purchaseData.selectedPreview?.url || purchaseData.originalImage}
                    alt="Preview"
                    className="w-full h-full object-cover"
                  />
                </div>
                <p className="text-sm text-gray-600 text-center">
                  {purchaseData.selectedPreview?.title || 'Выбранное превью'}
                </p>
              </div>
              
              {/* Детали заказа */}
              <div className="space-y-2 text-sm">
                <div className="flex justify-between">
                  <span className="text-gray-600">Размер:</span>
                  <span className="font-medium">{purchaseData.size} см</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-600">Стиль:</span>
                  <span className="font-medium">{getStyleTitle(purchaseData.style)}</span>
                </div>
                {purchaseData.selectedPreview?.type && (
                  <div className="flex justify-between">
                    <span className="text-gray-600">Тип обработки:</span>
                    <span className="font-medium capitalize">{purchaseData.selectedPreview.type}</span>
                  </div>
                )}
              </div>
            </div>

            {/* Что входит в покупку */}
            <div className="bg-white rounded-xl p-6 shadow-lg">
              <h3 className="text-lg font-semibold text-gray-800 mb-4 flex items-center">
                <Package className="w-5 h-5 mr-2 text-purple-600" />
                Что вы получите
              </h3>
              <ul className="space-y-2 text-sm text-gray-600">
                <li className="flex items-center">
                  <Check className="w-4 h-4 text-green-500 mr-2 flex-shrink-0" />
                  Подробная схема алмазной мозаики
                </li>
                <li className="flex items-center">
                  <Check className="w-4 h-4 text-green-500 mr-2 flex-shrink-0" />
                  Список необходимых материалов
                </li>
                <li className="flex items-center">
                  <Check className="w-4 h-4 text-green-500 mr-2 flex-shrink-0" />
                  Пошаговая инструкция
                </li>
                <li className="flex items-center">
                  <Check className="w-4 h-4 text-green-500 mr-2 flex-shrink-0" />
                  Файлы для скачивания
                </li>
              </ul>
            </div>
          </motion.div>

          {/* Выбор пакета */}
          <motion.div
            initial={{ opacity: 0, x: 20 }}
            animate={{ opacity: 1, x: 0 }}
            className="lg:col-span-2"
          >
            <h2 className="text-2xl font-bold text-gray-800 mb-6">Выберите пакет</h2>
            
            <div className="grid md:grid-cols-3 gap-6 mb-8">
              {packages.map((pkg, index) => (
                <motion.div
                  key={pkg.id}
                  initial={{ opacity: 0, y: 20 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ delay: 0.1 + index * 0.1 }}
                  className={`relative bg-white rounded-xl p-6 cursor-pointer transition-all duration-300 ${
                    selectedPackage === pkg.id
                      ? 'ring-4 ring-purple-500 shadow-xl scale-105'
                      : 'shadow-lg hover:shadow-xl hover:scale-102'
                  } ${pkg.popular ? 'border-2 border-purple-500' : 'border border-gray-200'}`}
                  onClick={() => handlePackageSelect(pkg.id)}
                >
                  {pkg.popular && (
                    <div className="absolute -top-3 left-1/2 transform -translate-x-1/2">
                      <div className="bg-purple-500 text-white px-4 py-1 rounded-full text-sm font-medium flex items-center">
                        <Star className="w-3 h-3 mr-1" />
                        Популярный
                      </div>
                    </div>
                  )}
                  
                  <div className="text-center">
                    <h3 className="text-xl font-semibold text-gray-800 mb-2">{pkg.name}</h3>
                    
                    <div className="mb-4">
                      <div className="flex items-center justify-center space-x-2">
                        <span className="text-3xl font-bold text-purple-600">{pkg.price}₽</span>
                        <span className="text-lg text-gray-400 line-through">{pkg.originalPrice}₽</span>
                      </div>
                      <div className="text-sm text-green-600 font-medium">
                        Скидка {Math.round((1 - pkg.price / pkg.originalPrice) * 100)}%
                      </div>
                    </div>
                    
                    <ul className="space-y-2 text-sm text-gray-600 mb-6">
                      {pkg.features.map((feature, idx) => (
                        <li key={idx} className="flex items-center">
                          <Check className="w-4 h-4 text-green-500 mr-2 flex-shrink-0" />
                          {feature}
                        </li>
                      ))}
                    </ul>
                  </div>
                </motion.div>
              ))}
            </div>

            {/* Кнопка покупки */}
            {selectedPkg && (
              <motion.div
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                className="bg-white rounded-xl p-6 shadow-lg"
              >
                <div className="flex items-center justify-between mb-6">
                  <div>
                    <h3 className="text-lg font-semibold text-gray-800">
                      {selectedPkg.name} пакет
                    </h3>
                    <p className="text-gray-600">
                      Итого к оплате: <span className="text-2xl font-bold text-purple-600">{selectedPkg.price}₽</span>
                    </p>
                  </div>
                  <div className="text-right">
                    <div className="text-sm text-gray-500 line-through">
                      {selectedPkg.originalPrice}₽
                    </div>
                    <div className="text-green-600 font-medium">
                      Экономия {selectedPkg.originalPrice - selectedPkg.price}₽
                    </div>
                  </div>
                </div>
                
                <button
                  onClick={handlePurchase}
                  disabled={isProcessing}
                  className="w-full bg-gradient-to-r from-purple-600 to-pink-600 text-white px-6 py-4 rounded-xl font-semibold text-lg hover:from-purple-700 hover:to-pink-700 transition-all duration-300 shadow-lg hover:shadow-xl disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center"
                >
                  {isProcessing ? (
                    <>
                      <div className="w-5 h-5 border-2 border-white border-t-transparent rounded-full animate-spin mr-2"></div>
                      Обработка...
                    </>
                  ) : (
                    <>
                      <CreditCard className="w-5 h-5 mr-2" />
                      Купить и сгенерировать схему
                    </>
                  )}
                </button>
                
                <p className="text-xs text-gray-500 text-center mt-3">
                  Нажимая кнопку, вы соглашаетесь с условиями использования
                </p>
              </motion.div>
            )}
          </motion.div>
        </div>
      </div>
    </div>
  )
}

export default DiamondMosaicPurchasePage
