import React, { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { motion } from 'framer-motion'
import { ArrowLeft, ArrowRight, Edit, ShoppingCart, Loader2, Sparkles, Eye } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import { useUIStore } from '../store/partnerStore'
import MosaicAPIClient, { MosaicAPI } from '../api/client'
import MarketplaceCards from '../components/MarketplaceCards'

const DiamondMosaicPreviewAlbumPage = () => {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { addNotification } = useUIStore()
  
  const [imageData, setImageData] = useState(null)
  const [selectedPreview, setSelectedPreview] = useState(0)
  const [useAI, setUseAI] = useState(false)
  const [previews, setPreviews] = useState([])
  const [isGeneratingAI, setIsGeneratingAI] = useState(false)
  const [isGeneratingVariants, setIsGeneratingVariants] = useState(false)

  // Contrast configuration - 4 контраста × 2 уровня = 8 превью
  const contrastVariants = [
    { name: 'venus', type: 'soft', label: t('diamond_mosaic_preview_album.contrast_variants.soft') },
    { name: 'venus', type: 'strong', label: t('diamond_mosaic_preview_album.contrast_variants.strong') },
    { name: 'sun', type: 'soft', label: t('diamond_mosaic_preview_album.contrast_variants.soft') },
    { name: 'sun', type: 'strong', label: t('diamond_mosaic_preview_album.contrast_variants.strong') },
    { name: 'moon', type: 'soft', label: t('diamond_mosaic_preview_album.contrast_variants.soft') },
    { name: 'moon', type: 'strong', label: t('diamond_mosaic_preview_album.contrast_variants.strong') },
    { name: 'mars', type: 'soft', label: t('diamond_mosaic_preview_album.contrast_variants.soft') },
    { name: 'mars', type: 'strong', label: t('diamond_mosaic_preview_album.contrast_variants.strong') }
  ]

  useEffect(() => {
    // Load image data
    try {
      const savedImageData = localStorage.getItem('diamondMosaic_selectedImage')
      
      if (!savedImageData) {
        navigate('/diamond-mosaic')
        return
      }
      
      const parsedData = JSON.parse(savedImageData)
      if (!parsedData.selectedStyle) {
        navigate('/diamond-mosaic/styles')
        return
      }
      
      setImageData(parsedData)
      
      // Generate contrast variants
      generateContrastVariants(parsedData)
      
    } catch (error) {
      console.error('Error loading image data:', error)
      navigate('/diamond-mosaic')
    }
  }, [navigate])

  const generateContrastVariants = async (data) => {
    setIsGeneratingVariants(true)
    
    try {
      const fileUrl = sessionStorage.getItem('diamondMosaic_fileUrl')
      if (!fileUrl) {
        throw new Error('No file URL found')
      }
      
      // Get file from URL
      const response = await fetch(fileUrl)
      const blob = await response.blob()
      
      const generatedPreviews = []
      
      // Add main preview as first
      generatedPreviews.push({
        id: 0,
        url: data.stylePreview,
        title: getStyleTitle(data.selectedStyle),
        type: 'original',
        isMain: true
      })
      
      // Generate contrast variants
      for (let i = 0; i < contrastVariants.length; i++) {
        const variant = contrastVariants[i]
        
        try {
          const formData = new FormData()
          formData.append('image', blob, 'image.jpg')
          formData.append('size', data.size)
          formData.append('style', data.selectedStyle)
          formData.append('contrast_type', variant.name)
          formData.append('contrast_level', variant.type)
          formData.append('use_ai', 'false')
          
          // Use modified API for variant generation
          const result = await MosaicAPI.generatePreviewVariant ? 
            await MosaicAPI.generatePreviewVariant(formData) :
            await MosaicAPI.generatePreview(formData)
          
          generatedPreviews.push({
            id: i + 1,
            url: result.preview_url,
            title: `${t(`diamond_mosaic_preview_album.contrast_variants.${variant.name}`)} (${variant.label})`,
            type: 'contrast',
            variant: variant
          })
          
          // Update preview as it generates
          setPreviews([...generatedPreviews])
          
        } catch (error) {
          console.error(`Error generating contrast variant ${i}:`, error)
          // Добавляем placeholder при ошибке
          generatedPreviews.push({
            id: i + 1,
            url: null,
            title: `${t(`diamond_mosaic_preview_album.contrast_variants.${variant.name}`)} (${variant.label})`,
            type: 'contrast',
            variant: variant,
            error: true
          })
        }
      }
      
      setPreviews(generatedPreviews)
      
    } catch (error) {
      console.error('Error generating contrast variants:', error)
      addNotification({
        type: 'error',
        message: t('diamond_mosaic_preview_album.contrast_generation_error')
      })
    } finally {
      setIsGeneratingVariants(false)
    }
  }

  const generateAIPreviews = async () => {
    if (!imageData) return
    
    setIsGeneratingAI(true)
    
    try {
      const fileUrl = sessionStorage.getItem('diamondMosaic_fileUrl')
      const response = await fetch(fileUrl)
      const blob = await response.blob()
      
      const aiPreviews = []
      
      // Генерируем 2+ AI варианта
      for (let i = 0; i < 2; i++) {
        try {
          const formData = new FormData()
          formData.append('image', blob, 'image.jpg')
          formData.append('size', imageData.size)
          formData.append('style', imageData.selectedStyle)
          formData.append('use_ai', 'true')
          formData.append('ai_variant', i.toString())
          
          const result = await MosaicAPI.generatePreview(formData)
          
          aiPreviews.push({
            id: previews.length + i,
            url: result.preview_url,
            title: `${t('diamond_mosaic_preview_album.ai_processing')} ${i + 1}`,
            type: 'ai'
          })
          
        } catch (error) {
          console.error(`Error generating AI preview ${i}:`, error)
        }
      }
      
      // Добавляем AI превью к существующим
      setPreviews(prev => [...prev, ...aiPreviews])
      
    } catch (error) {
      console.error('Error generating AI previews:', error)
      addNotification({
        type: 'error',
        message: t('diamond_mosaic_preview_album.ai_preview_generation_error')
      })
    } finally {
      setIsGeneratingAI(false)
    }
  }

  const handleAIToggle = (enabled) => {
    setUseAI(enabled)
    
    if (enabled && !previews.some(p => p.type === 'ai')) {
      generateAIPreviews()
    }
  }

  const handlePreviewSelect = (index) => {
    setSelectedPreview(index)
  }

  const handleEditImage = () => {
    // Сохраняем текущие настройки для редактора
    try {
      const editorData = {
        size: imageData.size,
        style: imageData.selectedStyle,
        returnTo: '/diamond-mosaic/preview-album'
      }
      localStorage.setItem('diamondMosaic_editorSettings', JSON.stringify(editorData))
      
      // Переходим в редактор
      navigate('/diamond-mosaic/editor')
      
    } catch (error) {
      console.error('Error preparing editor data:', error)
    }
  }

  const handlePurchase = () => {
    if (!imageData || !previews[selectedPreview]) {
      addNotification({
        type: 'error',
        message: t('diamond_mosaic_preview_album.select_preview_for_purchase')
      })
      return
    }
    
    // Сохраняем данные для покупки
    try {
      const purchaseData = {
        size: imageData.size,
        style: imageData.selectedStyle,
        selectedPreview: previews[selectedPreview],
        originalImage: imageData.previewUrl
      }
      localStorage.setItem('diamondMosaic_purchaseData', JSON.stringify(purchaseData))
      
      // Переходим на страницу покупки
      navigate('/diamond-mosaic/purchase')
      
    } catch (error) {
      console.error('Error preparing purchase data:', error)
    }
  }

  const handleBack = () => {
    navigate('/diamond-mosaic/styles')
  }

  const getStyleTitle = (styleKey) => {
    const styleMap = {
      'max_colors': t('diamond_mosaic_preview.style_selection.styles.realistic.title'),
      'pop_art': t('diamond_mosaic_preview.style_selection.styles.bright.title'),
      'grayscale': t('diamond_mosaic_preview.style_selection.styles.monochrome.title'),
      'skin_tones': t('diamond_mosaic_preview.style_selection.styles.warm.title')
    }
    return styleMap[styleKey] || styleKey
  }

  if (!imageData) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-purple-50 via-pink-50 to-blue-50 flex items-center justify-center">
        <Loader2 className="w-8 h-8 animate-spin text-purple-600" />
      </div>
    )
  }

  const currentPreview = previews[selectedPreview]

  return (
    <div className="min-h-screen bg-gradient-to-br from-purple-50 via-pink-50 to-blue-50 py-8 px-4">
      <div className="container mx-auto max-w-7xl">
        
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
            {t('diamond_mosaic_preview_album.back_to_style_selection')}
          </button>
          
          <div className="text-center">
            <h1 className="text-4xl md:text-5xl font-bold bg-gradient-to-r from-purple-600 to-pink-600 bg-clip-text text-transparent mb-4">
              {t('diamond_mosaic_preview_album.title')}
            </h1>
            <p className="text-lg text-gray-600">
              {t('diamond_mosaic_preview_album.size_label')} <span className="font-semibold">{imageData.size} {t('common.cm')}</span> • 
              {t('diamond_mosaic_preview_album.style_label')} <span className="font-semibold">{getStyleTitle(imageData.selectedStyle)}</span>
            </p>
          </div>
        </motion.div>

        <div className="grid lg:grid-cols-3 gap-8">
          
          {/* Мини-альбом превью */}
          <motion.div
            initial={{ opacity: 0, x: -20 }}
            animate={{ opacity: 1, x: 0 }}
            className="lg:col-span-1"
          >
            <h2 className="text-2xl font-bold text-gray-800 mb-6">{t('diamond_mosaic_preview_album.subtitle')}</h2>
            
            {/* AI переключатель */}
            <div className="mb-6 p-4 bg-white rounded-xl border border-gray-200">
              <div className="flex items-center justify-between mb-2">
                <label className="flex items-center cursor-pointer">
                  <Sparkles className="w-5 h-5 text-purple-600 mr-2" />
                  <span className="font-medium">{t('diamond_mosaic_preview_album.ai_processing')}</span>
                </label>
                <input
                  type="checkbox"
                  checked={useAI}
                  onChange={(e) => handleAIToggle(e.target.checked)}
                  className="w-5 h-5 text-purple-600 rounded focus:ring-purple-500"
                />
              </div>
              <p className="text-sm text-gray-600">
                {t('diamond_mosaic_preview_album.ai_description')}
              </p>
              {isGeneratingAI && (
                <div className="mt-2 flex items-center text-purple-600">
                  <Loader2 className="w-4 h-4 animate-spin mr-2" />
                  {t('diamond_mosaic_preview_album.generating_ai')}
                </div>
              )}
            </div>
            
            {/* Список превью */}
            <div className="space-y-3 max-h-96 overflow-y-auto">
              {isGeneratingVariants && previews.length === 0 && (
                <div className="text-center py-8">
                  <Loader2 className="w-8 h-8 animate-spin text-purple-600 mx-auto mb-2" />
                  <p className="text-gray-600">{t('diamond_mosaic_preview_album.generating_variants')}</p>
                </div>
              )}
              
              {previews.map((preview, index) => (
                <motion.div
                  key={preview.id}
                  initial={{ opacity: 0, y: 10 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ delay: index * 0.1 }}
                  className={`flex items-center p-3 rounded-lg cursor-pointer transition-all ${
                    selectedPreview === index
                      ? 'bg-purple-100 border-2 border-purple-500'
                      : 'bg-white border border-gray-200 hover:border-purple-300'
                  }`}
                  onClick={() => handlePreviewSelect(index)}
                >
                  <div className="w-16 h-16 rounded-lg overflow-hidden bg-gray-100 mr-3 flex-shrink-0">
                    {preview.url ? (
                      <img
                        src={preview.url}
                        alt={preview.title}
                        className="w-full h-full object-cover"
                      />
                    ) : preview.error ? (
                      <div className="w-full h-full flex items-center justify-center text-red-400">
                        ❌
                      </div>
                    ) : (
                      <div className="w-full h-full flex items-center justify-center">
                        <Loader2 className="w-4 h-4 animate-spin text-purple-600" />
                      </div>
                    )}
                  </div>
                  <div className="flex-1 min-w-0">
                    <p className="font-medium text-gray-800 truncate">{preview.title}</p>
                    <p className="text-sm text-gray-500 capitalize">{preview.type}</p>
                    {preview.isMain && (
                      <span className="inline-block px-2 py-1 text-xs bg-purple-100 text-purple-700 rounded mt-1">
                        {t('diamond_mosaic_preview_album.main_preview')}
                      </span>
                    )}
                  </div>
                </motion.div>
              ))}
            </div>
          </motion.div>

          {/* Превью и действия */}
          <motion.div
            initial={{ opacity: 0, x: 20 }}
            animate={{ opacity: 1, x: 0 }}
            className="lg:col-span-2"
          >
            {/* Большое превью */}
            <div className="bg-white rounded-xl p-6 mb-6 shadow-lg">
              <div className="aspect-square bg-gray-100 rounded-lg overflow-hidden mb-4">
                {currentPreview?.url ? (
                  <img
                    src={currentPreview.url}
                    alt={currentPreview.title}
                    className="w-full h-full object-cover"
                  />
                ) : (
                  <div className="w-full h-full flex items-center justify-center">
                    <Eye className="w-12 h-12 text-gray-400" />
                  </div>
                )}
              </div>
              
              {currentPreview && (
                <div className="text-center">
                  <h3 className="text-xl font-semibold text-gray-800 mb-2">
                    {currentPreview.title}
                  </h3>
                  <p className="text-gray-600 capitalize">{currentPreview.type}</p>
                </div>
              )}
            </div>
            
            {/* Кнопки действий */}
            <div className="space-y-4 mb-8">
              {/* Кнопка редактирования изображения */}
              <button
                onClick={handleEditImage}
                className="w-full bg-white text-purple-600 border-2 border-purple-600 px-6 py-3 rounded-xl font-semibold hover:bg-purple-50 transition-all duration-300 flex items-center justify-center"
              >
                <Edit className="w-5 h-5 mr-2" />
                {t('diamond_mosaic_preview_album.edit_image')}
              </button>
              
              {/* Кнопка покупки купона */}
              {currentPreview?.url && (
                <div className="text-center">
                  <p className="text-gray-700 mb-4 text-lg">
                    {t('diamond_mosaic_preview_album.liked_preview_text')}
                  </p>
                  <button
                    onClick={handlePurchase}
                    className="w-full bg-gradient-to-r from-purple-600 to-pink-600 text-white px-6 py-4 rounded-xl font-semibold text-lg hover:from-purple-700 hover:to-pink-700 transition-all duration-300 shadow-lg hover:shadow-xl flex items-center justify-center"
                  >
                    <ShoppingCart className="w-5 h-5 mr-2" />
                    {t('diamond_mosaic_preview_album.buy_coupon_and_generate')}
                  </button>
                </div>
              )}
            </div>

            {/* Магазины */}
            {currentPreview?.url && (
              <motion.div
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.3 }}
              >
                <h3 className="text-xl font-bold text-gray-800 mb-4">{t('diamond_mosaic_preview_album.ready_sets_in_stores')}</h3>
                <MarketplaceCards 
                  selectedSize={imageData.size} 
                  selectedStyle={imageData.selectedStyle} 
                />
              </motion.div>
            )}
          </motion.div>
        </div>
      </div>
    </div>
  )
}

export default DiamondMosaicPreviewAlbumPage
