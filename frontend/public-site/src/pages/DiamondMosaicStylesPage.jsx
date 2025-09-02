import React, { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { useNavigate } from 'react-router-dom'
import { ArrowLeft, Palette, Sun, Moon, Venus, Sparkles } from 'lucide-react'
import { useUIStore } from '../store/partnerStore'

const DiamondMosaicStylesPage = () => {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { addNotification } = useUIStore()
  
  const [imageData, setImageData] = useState(null)
  const [selectedStyle, setSelectedStyle] = useState(null)
  const [useAI, setUseAI] = useState(false)

  const styles = [
    {
      key: 'natural',
      title: 'Натуральный',
      description: 'Сохраняет оригинальные цвета',
      icon: <Palette className="w-8 h-8" />,
      preview: '/images/style-natural.jpg'
    },
    {
      key: 'enhanced',
      title: 'Улучшенный',
      description: 'Яркие и насыщенные цвета',
      icon: <Sparkles className="w-8 h-8" />,
      preview: '/images/style-enhanced.jpg'
    },
    {
      key: 'vintage',
      title: 'Винтажный',
      description: 'Приглушенные тона',
      icon: <Sun className="w-8 h-8" />,
      preview: '/images/style-vintage.jpg'
    },
    {
      key: 'monochrome',
      title: 'Монохром',
      description: 'Черно-белый стиль',
      icon: <Moon className="w-8 h-8" />,
      preview: '/images/style-monochrome.jpg'
    }
  ]

  useEffect(() => {
    // Загружаем данные изображения
    try {
      const savedImageData = localStorage.getItem('diamondMosaic_selectedImage')
      if (!savedImageData) {
        navigate('/diamond-mosaic')
        return
      }
      
      const parsedData = JSON.parse(savedImageData)
      setImageData(parsedData)
      
    } catch (error) {
      console.error('Error loading image data:', error)
      navigate('/diamond-mosaic')
    }
  }, [navigate])

  const handleStyleSelect = (styleKey) => {
    setSelectedStyle(styleKey)
  }

  const handleContinue = () => {
    if (!selectedStyle) {
      addNotification({
        type: 'error',
        message: 'Пожалуйста, выберите стиль обработки'
      })
      return
    }

    try {
      // Обновляем данные изображения с выбранным стилем
      const updatedImageData = {
        ...imageData,
        selectedStyle: selectedStyle,
        useAI: useAI
      }
      
      localStorage.setItem('diamondMosaic_selectedImage', JSON.stringify(updatedImageData))
      
      // Переходим к альбому превью
      navigate('/diamond-mosaic/preview-album')
      
    } catch (error) {
      console.error('Error saving style selection:', error)
      addNotification({
        type: 'error',
        message: 'Ошибка при сохранении стиля'
      })
    }
  }

  const handleBack = () => {
    navigate('/diamond-mosaic/editor')
  }

  if (!imageData) {
    return (
      <div className="min-h-screen bg-white flex items-center justify-center">
        <div className="text-gray-600">Загрузка...</div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-white">
      <div className="max-w-6xl mx-auto px-6 py-8">
        
        {/* Заголовок */}
        <div className="mb-8">
          <button
            onClick={handleBack}
            className="flex items-center text-purple-600 hover:text-purple-700 mb-4 transition-colors"
          >
            <ArrowLeft className="w-5 h-5 mr-2" />
            Назад к редактору
          </button>
          
          <h1 className="text-3xl font-bold text-gray-900 text-center mb-2">
            Выберите стиль обработки
          </h1>
          <p className="text-gray-600 text-center">
            Размер: {imageData.size} см
          </p>
        </div>

        {/* Превью изображения */}
        <div className="mb-8 text-center">
          <img 
            src={imageData.previewUrl} 
            alt="Uploaded image"
            className="w-48 h-48 object-cover rounded-lg mx-auto border-2 border-gray-200"
          />
        </div>

        {/* Выбор стилей */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
          {styles.map((style) => (
            <div
              key={style.key}
              onClick={() => handleStyleSelect(style.key)}
              className={`
                p-6 rounded-xl border-2 cursor-pointer transition-all hover:shadow-lg
                ${selectedStyle === style.key 
                  ? 'border-purple-500 bg-purple-50' 
                  : 'border-gray-200 hover:border-purple-300'
                }
              `}
            >
              <div className="text-center">
                <div className="flex justify-center mb-4 text-purple-600">
                  {style.icon}
                </div>
                <h3 className="text-lg font-semibold text-gray-900 mb-2">
                  {style.title}
                </h3>
                <p className="text-sm text-gray-600">
                  {style.description}
                </p>
              </div>
            </div>
          ))}
        </div>

        {/* Опция ИИ */}
        <div className="mb-8 text-center">
          <label className="inline-flex items-center">
            <input
              type="checkbox"
              checked={useAI}
              onChange={(e) => setUseAI(e.target.checked)}
              className="form-checkbox h-5 w-5 text-purple-600 rounded focus:ring-purple-500"
            />
            <span className="ml-2 text-gray-700">
              Использовать ИИ для улучшения (дополнительные превью от нейросети)
            </span>
          </label>
        </div>

        {/* Кнопки действий */}
        <div className="flex gap-4 max-w-md mx-auto">
          <button
            onClick={handleBack}
            className="flex-1 py-4 px-6 bg-white border-2 border-gray-300 text-gray-700 rounded-xl font-medium hover:bg-gray-50 transition-colors"
          >
            Назад
          </button>
          
          <button
            onClick={handleContinue}
            disabled={!selectedStyle}
            className="flex-1 py-4 px-6 bg-gradient-to-r from-purple-600 to-pink-600 text-white rounded-xl font-medium hover:from-purple-700 hover:to-pink-700 transition-all disabled:opacity-50 disabled:cursor-not-allowed"
          >
            Создать превью
          </button>
        </div>

      </div>
    </div>
  )
}

export default DiamondMosaicStylesPage
