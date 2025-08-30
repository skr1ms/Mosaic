import React, { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { motion } from 'framer-motion'
import { ArrowLeft, Ruler, Palette, ShoppingCart, ExternalLink, Download, Eye } from 'lucide-react'
import { usePartnerStore } from '../store/partnerStore'
import MosaicAPI from '../api/client'

const MosaicPreviewPage = () => {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const { partner } = usePartnerStore()
  
  const [selectedSize, setSelectedSize] = useState('')
  const [selectedStyle, setSelectedStyle] = useState('')
  const [previewImage, setPreviewImage] = useState(null)
  const [isGenerating, setIsGenerating] = useState(false)
  const [marketplaceLinks, setMarketplaceLinks] = useState([])

  // Доступные размеры и стили
  const sizes = [
    { key: '20x20', title: '20×20', description: 'Небольшой размер, идеально для начинающих' },
    { key: '30x40', title: '30×40', description: 'Средний размер, популярный выбор' },
    { key: '40x40', title: '40×40', description: 'Квадратный формат, сбалансированный' },
    { key: '40x50', title: '40×50', description: 'Прямоугольный, для пейзажей' },
    { key: '40x60', title: '40×60', description: 'Широкий формат, для панорам' },
    { key: '50x70', title: '50×70', description: 'Большой размер, для опытных мастеров' }
  ]

  const styles = [
    { key: 'grayscale', title: 'Черно-белый', description: 'Классический стиль, элегантный' },
    { key: 'skin_tones', title: 'Телесные тона', description: 'Реалистичные оттенки кожи' },
    { key: 'pop_art', title: 'Поп-арт', description: 'Яркие, контрастные цвета' },
    { key: 'max_colors', title: 'Максимум цветов', description: 'Богатая цветовая палитра' }
  ]

  // Получаем размер и стиль из URL параметров
  useEffect(() => {
    const size = searchParams.get('size')
    const style = searchParams.get('style')
    if (size && style) {
      setSelectedSize(size)
      setSelectedStyle(style)
      generatePreview(size, style)
    }
  }, [searchParams])

  // Генерируем превью мозаики
  const generatePreview = async (size, style) => {
    if (!size || !style) return

    setIsGenerating(true)
    try {
      // Вызываем API для генерации превью
      const response = await fetch('/api/preview/generate', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          size: size,
          style: style,
          partner_id: partner?.id || 'default',
          user_email: 'preview@example.com' // В реальности можно брать из формы
        })
      })

      if (!response.ok) {
        throw new Error('Failed to generate preview')
      }

      const data = await response.json()
      setPreviewImage(data.preview_url)
      
      // Генерируем ссылки на маркетплейсы
      generateMarketplaceLinks(size, style)
    } catch (error) {
      console.error('Failed to generate preview:', error)
      // В случае ошибки показываем заглушку
      setPreviewImage('/api/preview-placeholder')
    } finally {
      setIsGenerating(false)
    }
  }

  // Генерируем ссылки на маркетплейсы на основе размера и стиля
  const generateMarketplaceLinks = async (size, style) => {
    if (!partner) return

    try {
      // Получаем артикулы партнера
      const response = await MosaicAPI.getPartnerArticleGrid(partner.id)
      const articleGrid = response.data || {}
      
      const links = []
      
      // OZON
      if (partner.ozonLink) {
        const ozonSKU = articleGrid.ozon?.[style]?.[size]
        const ozonUrl = ozonSKU 
          ? `${partner.ozonLink}?sku=${ozonSKU}&size=${size}&style=${style}`
          : partner.ozonLink
        
        links.push({
          name: 'OZON',
          url: ozonUrl,
          description: `Купить набор ${size} в стиле ${style}`,
          color: 'from-orange-500 to-red-500',
          icon: '🟠',
          hasSpecificSKU: !!ozonSKU
        })
      }
      
      // Wildberries
      if (partner.wildberriesLink) {
        const wbSKU = articleGrid.wildberries?.[style]?.[size]
        const wbUrl = wbSKU 
          ? `${partner.wildberriesLink}?sku=${wbSKU}&size=${size}&style=${style}`
          : partner.wildberriesLink
        
        links.push({
          name: 'Wildberries',
          url: wbUrl,
          description: `Купить набор ${size} в стиле ${style}`,
          color: 'from-purple-500 to-pink-500',
          icon: '🟣',
          hasSpecificSKU: !!wbSKU
        })
      }

      setMarketplaceLinks(links)
    } catch (error) {
      console.error('Failed to get partner articles:', error)
      // Fallback к базовым ссылкам
      const links = []
      
      if (partner.ozonLink) {
        links.push({
          name: 'OZON',
          url: partner.ozonLink,
          description: `Купить набор ${size} в стиле ${style}`,
          color: 'from-orange-500 to-red-500',
          icon: '🟠',
          hasSpecificSKU: false
        })
      }
      
      if (partner.wildberriesLink) {
        links.push({
          name: 'Wildberries',
          url: partner.wildberriesLink,
          description: `Купить набор ${size} в стиле ${style}`,
          color: 'from-purple-500 to-pink-500',
          icon: '🟣',
          hasSpecificSKU: false
        })
      }
      
      setMarketplaceLinks(links)
    }
  }

  // Обработчик выбора размера и стиля
  const handleSizeStyleSelect = (size, style) => {
    setSelectedSize(size)
    setSelectedStyle(style)
    generatePreview(size, style)
  }

  // Переход к редактору с выбранными параметрами
  const goToEditor = () => {
    if (selectedSize && selectedStyle) {
      navigate(`/editor?size=${selectedSize}&style=${selectedStyle}`)
    }
  }

  // Проверяем доступность артикулов для конкретного размера и стиля
  const checkArticleAvailability = (size, style) => {
    const articleGrid = partner?.articleGrid || {}
    return {
      ozon: articleGrid.ozon?.[style]?.[size] !== undefined,
      wildberries: articleGrid.wildberries?.[style]?.[size] !== undefined
    }
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 via-purple-50 to-pink-50">
      {/* Header */}
      <div className="bg-white shadow-sm border-b">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <div className="flex items-center justify-between">
            <button
              onClick={() => navigate('/diamond-art')}
              className="inline-flex items-center space-x-2 text-gray-600 hover:text-gray-900 transition-colors"
            >
              <ArrowLeft className="w-5 h-5" />
              <span>Назад к алмазной мозаике</span>
            </button>
            <h1 className="text-2xl font-bold text-gray-900">Превью мозаики</h1>
          </div>
        </div>
      </div>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="grid lg:grid-cols-2 gap-8">
          
          {/* Левая колонка - Выбор размера и стиля */}
          <div className="space-y-8">
            
            {/* Выбор размера */}
            <motion.div 
              initial={{ opacity: 0, x: -20 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ duration: 0.5 }}
              className="bg-white rounded-2xl shadow-lg p-6"
            >
              <div className="flex items-center space-x-3 mb-6">
                <div className="w-10 h-10 bg-brand-primary/10 rounded-xl flex items-center justify-center">
                  <Ruler className="w-5 h-5 text-brand-primary" />
                </div>
                <h2 className="text-xl font-bold text-gray-900">Выберите размер</h2>
              </div>
              
              <div className="grid grid-cols-2 gap-4">
                {sizes.map((size) => (
                  <button
                    key={size.key}
                    onClick={() => handleSizeStyleSelect(size.key, selectedStyle)}
                    className={`p-4 rounded-xl border-2 transition-all duration-200 text-left relative ${
                      selectedSize === size.key
                        ? 'border-brand-primary bg-brand-primary/5'
                        : 'border-gray-200 hover:border-gray-300 hover:bg-gray-50'
                    }`}
                  >
                    <div className="font-semibold text-gray-900 mb-1">{size.title}</div>
                    <div className="text-sm text-gray-600">{size.description}</div>
                    
                    {/* Индикатор доступности артикулов */}
                    {selectedStyle && (
                      <div className="mt-2 flex items-center space-x-2">
                        {checkArticleAvailability(size.key, selectedStyle).ozon && (
                          <span className="inline-flex items-center px-2 py-1 bg-orange-100 text-orange-800 text-xs rounded-full">
                            🟠 OZON
                          </span>
                        )}
                        {checkArticleAvailability(size.key, selectedStyle).wildberries && (
                          <span className="inline-flex items-center px-2 py-1 bg-purple-100 text-purple-800 text-xs rounded-full">
                            🟣 WB
                          </span>
                        )}
                        {!checkArticleAvailability(size.key, selectedStyle).ozon && !checkArticleAvailability(size.key, selectedStyle).wildberries && (
                          <span className="inline-flex items-center px-2 py-1 bg-gray-100 text-gray-600 text-xs rounded-full">
                            ⚠️ Нет артикулов
                          </span>
                        )}
                      </div>
                    )}
                  </button>
                ))}
              </div>
            </motion.div>

            {/* Выбор стиля */}
            <motion.div 
              initial={{ opacity: 0, x: -20 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ duration: 0.5, delay: 0.1 }}
              className="bg-white rounded-2xl shadow-lg p-6"
            >
              <div className="flex items-center space-x-3 mb-6">
                <div className="w-10 h-10 bg-brand-secondary/10 rounded-xl flex items-center justify-center">
                  <Palette className="w-5 h-5 text-brand-secondary" />
                </div>
                <h2 className="text-xl font-bold text-gray-900">Выберите стиль</h2>
              </div>
              
              <div className="space-y-3">
                {styles.map((style) => (
                  <button
                    key={style.key}
                    onClick={() => handleSizeStyleSelect(selectedSize, style.key)}
                    className={`w-full p-4 rounded-xl border-2 transition-all duration-200 text-left ${
                      selectedStyle === style.key
                        ? 'border-brand-secondary bg-brand-secondary/5'
                        : 'border-gray-200 hover:border-gray-300 hover:bg-gray-50'
                    }`}
                  >
                    <div className="font-semibold text-gray-900 mb-1">{style.title}</div>
                    <div className="text-sm text-gray-600">{style.description}</div>
                  </button>
                ))}
              </div>
            </motion.div>

            {/* Кнопка перехода к редактору */}
            {selectedSize && selectedStyle && (
              <motion.div 
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.5, delay: 0.2 }}
                className="bg-gradient-to-r from-brand-primary to-brand-secondary rounded-2xl p-6 text-white"
              >
                <h3 className="text-lg font-bold mb-3">Готово к созданию!</h3>
                <p className="text-brand-primary/80 mb-4">
                  Размер: <strong>{sizes.find(s => s.key === selectedSize)?.title}</strong><br />
                  Стиль: <strong>{styles.find(s => s.key === selectedStyle)?.title}</strong>
                </p>
                <button
                  onClick={goToEditor}
                  className="w-full inline-flex items-center justify-center px-6 py-3 bg-white text-brand-primary rounded-xl hover:bg-brand-primary/10 font-semibold transition-all duration-200"
                >
                  <Eye className="w-5 h-5 mr-2" />
                  Создать мозаику
                </button>
              </motion.div>
            )}
          </div>

          {/* Правая колонка - Превью и покупка */}
          <div className="space-y-6">
            
            {/* Превью мозаики */}
            <motion.div 
              initial={{ opacity: 0, x: 20 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ duration: 0.5, delay: 0.2 }}
              className="bg-white rounded-2xl shadow-lg p-6"
            >
              <h2 className="text-xl font-bold text-gray-900 mb-4">Превью мозаики</h2>
              
              {isGenerating ? (
                <div className="aspect-square bg-gray-100 rounded-xl flex items-center justify-center">
                  <div className="text-center">
                    <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-brand-primary mx-auto mb-4"></div>
                    <p className="text-gray-600">Генерируем превью...</p>
                  </div>
                </div>
              ) : previewImage ? (
                <div className="aspect-square bg-gray-100 rounded-xl overflow-hidden">
                  <img 
                    src={previewImage} 
                    alt="Превью мозаики" 
                    className="w-full h-full object-cover"
                  />
                </div>
              ) : (
                <div className="aspect-square bg-gray-100 rounded-xl flex items-center justify-center">
                  <div className="text-center text-gray-500">
                    <Eye className="w-16 h-16 mx-auto mb-4 opacity-50" />
                    <p>Выберите размер и стиль для просмотра превью</p>
                  </div>
                </div>
              )}
            </motion.div>

            {/* Ссылки на покупку */}
            {marketplaceLinks.length > 0 && (
              <motion.div 
                initial={{ opacity: 0, x: 20 }}
                animate={{ opacity: 1, x: 0 }}
                transition={{ duration: 0.5, delay: 0.3 }}
                className="bg-white rounded-2xl shadow-lg p-6"
              >
                <h2 className="text-xl font-bold text-gray-900 mb-4">Купить набор</h2>
                
                <div className="space-y-4">
                  {marketplaceLinks.map((marketplace, index) => (
                    <motion.div 
                      key={marketplace.name}
                      initial={{ opacity: 0, y: 20 }}
                      animate={{ opacity: 1, y: 0 }}
                      transition={{ duration: 0.5, delay: 0.4 + index * 0.1 }}
                      className="bg-gradient-to-r from-gray-50 to-gray-100 rounded-xl p-4 border border-gray-200"
                    >
                      <div className="flex items-center justify-between">
                        <div className="flex items-center space-x-3">
                          <span className="text-2xl">{marketplace.icon}</span>
                          <div>
                            <h3 className="font-semibold text-gray-900">{marketplace.name}</h3>
                            <p className="text-sm text-gray-600">{marketplace.description}</p>
                            {marketplace.hasSpecificSKU && (
                              <span className="inline-block mt-1 px-2 py-1 bg-green-100 text-green-800 text-xs rounded-full">
                                ✓ Специальный артикул
                              </span>
                            )}
                          </div>
                        </div>
                        <a
                          href={marketplace.url}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="inline-flex items-center px-4 py-2 bg-gradient-to-r from-brand-primary to-brand-secondary text-white rounded-lg hover:from-brand-primary/90 hover:to-brand-secondary/90 font-medium transition-all duration-200"
                        >
                          <ShoppingCart className="w-4 h-4 mr-2" />
                          Купить
                          <ExternalLink className="w-4 h-4 ml-2" />
                        </a>
                      </div>
                    </motion.div>
                  ))}
                </div>
              </motion.div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

export default MosaicPreviewPage
