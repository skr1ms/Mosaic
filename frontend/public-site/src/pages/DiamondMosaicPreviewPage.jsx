import React, { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { motion } from 'framer-motion'
import { ArrowLeft, ArrowRight, Palette, Loader2 } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import { useUIStore } from '../store/partnerStore'
import MosaicAPIClient, { MosaicAPI } from '../api/client'

const DiamondMosaicPreviewPage = () => {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { addNotification } = useUIStore()
  
  const [imageData, setImageData] = useState(null)
  const [selectedStyle, setSelectedStyle] = useState('')
  const [stylePreviews, setStylePreviews] = useState({})
  const [isGenerating, setIsGenerating] = useState(false)
  const [fileUrl, setFileUrl] = useState(null)

  const getStyles = () => [
    { 
      key: 'max_colors', 
      title: t('diamond_mosaic_preview.style_selection.styles.realistic.title'), 
      desc: t('diamond_mosaic_preview.style_selection.styles.realistic.desc'),
      preview: null
    },
    { 
      key: 'pop_art', 
      title: t('diamond_mosaic_preview.style_selection.styles.bright.title'), 
      desc: t('diamond_mosaic_preview.style_selection.styles.bright.desc'),
      preview: null
    },
    { 
      key: 'grayscale', 
      title: t('diamond_mosaic_preview.style_selection.styles.monochrome.title'), 
      desc: t('diamond_mosaic_preview.style_selection.styles.monochrome.desc'),
      preview: null
    },
    { 
      key: 'skin_tones', 
      title: t('diamond_mosaic_preview.style_selection.styles.warm.title'), 
      desc: t('diamond_mosaic_preview.style_selection.styles.warm.desc'),
      preview: null
    }
  ]

  useEffect(() => {
    // Загружаем данные изображения из localStorage
    try {
      const savedImageData = localStorage.getItem('diamondMosaic_selectedImage')
      const savedFileUrl = sessionStorage.getItem('diamondMosaic_fileUrl')
      
      if (!savedImageData) {
        // Если нет данных изображения, возвращаемся на предыдущую страницу
        navigate('/diamond-mosaic')
        return
      }
      
      const parsedData = JSON.parse(savedImageData)
      setImageData(parsedData)
      setFileUrl(savedFileUrl)
      
      // Сразу генерируем превью для всех стилей
      generateAllPreviews(savedFileUrl, parsedData.size)
      
    } catch (error) {
      console.error('Error loading image data:', error)
      navigate('/diamond-mosaic')
    }
  }, [navigate])

  const generateAllPreviews = async (imageUrl, size) => {
    if (!imageUrl || !size) return
    
    setIsGenerating(true)
    
    try {
      // Получаем файл из URL
      const response = await fetch(imageUrl)
      if (!response.ok) {
        throw new Error('Failed to fetch image')
      }
      const blob = await response.blob()
      
      // Генерируем превью для каждого стиля последовательно
      for (const style of getStyles()) {
        try {
          // Создаем новый FormData для каждого запроса
          const formData = new FormData()
          formData.append('image', blob, 'image.jpg')
          formData.append('size', size)
          formData.append('style', style.key)
          formData.append('use_ai', 'false')
          
          console.log(`Generating preview for style: ${style.key}`)
          
          // Используем существующий API endpoint
          const result = await MosaicAPI.generatePreview(formData)
          
          if (result && result.preview_url) {
            // Обновляем состояние для каждого сгенерированного превью
            setStylePreviews(prev => ({
              ...prev,
              [style.key]: result.preview_url
            }))
            console.log(`Preview generated for ${style.key}:`, result.preview_url)
          } else {
            console.warn(`No preview URL returned for style ${style.key}`)
          }
          
        } catch (error) {
          console.error(`Error generating preview for style ${style.key}:`, error)
          setStylePreviews(prev => ({
            ...prev,
            [style.key]: null
          }))
        }
        
        // Небольшая задержка между запросами
        await new Promise(resolve => setTimeout(resolve, 500))
      }
      
    } catch (error) {
      console.error('Error generating previews:', error)
      addNotification({
        type: 'error',
        message: t('diamond_mosaic_preview.notifications.generation_error')
      })
    } finally {
      setIsGenerating(false)
    }
  }

  const handleStyleSelect = (styleKey) => {
    setSelectedStyle(styleKey)
  }

  const handleContinue = () => {
    if (!selectedStyle) {
      addNotification({
        type: 'error',
        message: t('diamond_mosaic_preview.notifications.style_selection_required')
      })
      return
    }

    // Сохраняем выбранный стиль
    try {
      const updatedData = {
        ...imageData,
        selectedStyle: selectedStyle,
        stylePreview: stylePreviews[selectedStyle]
      }
      localStorage.setItem('diamondMosaic_selectedImage', JSON.stringify(updatedData))
      
      // Переходим к странице мини-альбома
      navigate('/diamond-mosaic/preview-album')
      
    } catch (error) {
      console.error('Error saving style selection:', error)
      addNotification({
        type: 'error',
        message: t('diamond_mosaic_preview.notifications.save_error')
      })
    }
  }

  const handleBack = () => {
    navigate('/diamond-mosaic')
  }

  if (!imageData) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-purple-50 via-pink-50 to-blue-50 flex items-center justify-center">
        <Loader2 className="w-8 h-8 animate-spin text-purple-600" />
      </div>
    )
  }

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
            {t('diamond_mosaic_preview.navigation.back')}
          </button>
          
          <div className="text-center">
            <h1 className="text-4xl md:text-5xl font-bold bg-gradient-to-r from-purple-600 to-pink-600 bg-clip-text text-transparent mb-4">
              {t('diamond_mosaic_preview.title')}
            </h1>
            <p className="text-lg md:text-xl text-gray-600 max-w-2xl mx-auto">
              {t('diamond_mosaic_preview.preview_section.size', { size: imageData.size + ' ' + t('common.cm') })}
            </p>
          </div>
        </motion.div>

        {/* Превью исходного изображения */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.1 }}
          className="text-center mb-8"
        >
          <h3 className="text-lg font-medium text-gray-700 mb-4">{t('diamond_mosaic_preview.subtitle')}:</h3>
          <div className="inline-block bg-white p-4 rounded-xl shadow-lg">
            <img
              src={imageData.previewUrl}
              alt="Original"
              className="max-w-xs max-h-48 object-contain rounded-lg"
            />
          </div>
        </motion.div>

        {/* Стили обработки */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.2 }}
          className="mb-12"
        >
          <h2 className="text-2xl font-bold text-gray-800 mb-6 text-center flex items-center justify-center">
            <Palette className="w-6 h-6 mr-2 text-purple-600" />
            {t('diamond_mosaic_preview.style_selection.title')}
          </h2>
          
          {isGenerating && (
            <div className="text-center mb-6">
              <div className="inline-flex items-center text-purple-600">
                <Loader2 className="w-5 h-5 animate-spin mr-2" />
                {t('diamond_mosaic_preview.preview_section.generating')}
              </div>
            </div>
          )}
          
          <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-6">
            {getStyles().map((style, index) => (
              <motion.div
                key={style.key}
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.3 + index * 0.1 }}
                className={`bg-white rounded-xl overflow-hidden shadow-lg cursor-pointer transition-all duration-300 ${
                  selectedStyle === style.key
                    ? 'ring-4 ring-purple-500 shadow-xl scale-105'
                    : 'hover:shadow-xl hover:scale-102'
                }`}
                onClick={() => handleStyleSelect(style.key)}
              >
                <div className="aspect-square bg-gray-100 flex items-center justify-center">
                  {stylePreviews[style.key] ? (
                    <img
                      src={stylePreviews[style.key]}
                      alt={style.title}
                      className="w-full h-full object-cover"
                    />
                  ) : (
                    <div className="text-center">
                      {isGenerating ? (
                        <Loader2 className="w-8 h-8 animate-spin text-purple-600 mx-auto mb-2" />
                      ) : (
                        <Palette className="w-8 h-8 text-gray-400 mx-auto mb-2" />
                      )}
                      <p className="text-sm text-gray-500">
                        {isGenerating ? t('diamond_mosaic_preview.preview_section.generating') : t('diamond_mosaic_preview.preview_section.title')}
                      </p>
                    </div>
                  )}
                </div>
                
                <div className="p-4">
                  <h3 className="font-semibold text-gray-800 mb-2">{style.title}</h3>
                  <p className="text-sm text-gray-600">{style.desc}</p>
                </div>
              </motion.div>
            ))}
          </div>
        </motion.div>

        {/* Кнопка продолжения */}
        {selectedStyle && (
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            className="text-center"
          >
            <button
              onClick={handleContinue}
              className="bg-gradient-to-r from-purple-600 to-pink-600 text-white px-8 py-4 rounded-xl font-semibold text-lg hover:from-purple-700 hover:to-pink-700 transition-all duration-300 shadow-lg hover:shadow-xl flex items-center mx-auto"
            >
              {t('diamond_mosaic_preview.navigation.continue')}
              <ArrowRight className="w-5 h-5 ml-2" />
            </button>
          </motion.div>
        )}
      </div>
    </div>
  )
}

export default DiamondMosaicPreviewPage
