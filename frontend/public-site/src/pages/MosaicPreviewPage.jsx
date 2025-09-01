import React, { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { motion } from 'framer-motion'
import { Upload, Eye, Download, ArrowLeft, Loader2, AlertCircle } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import { useMutation } from '@tanstack/react-query'
import { MosaicAPI } from '../api/client'
import { useUIStore } from '../store/partnerStore'
import MarketplaceCards from '../components/MarketplaceCards'

const MosaicPreviewPage = () => {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { addNotification } = useUIStore()
  
  const [selectedFile, setSelectedFile] = useState(null)
  const [previewUrl, setPreviewUrl] = useState(null)
  const [selectedSize, setSelectedSize] = useState('30x40')
  const [selectedStyle, setSelectedStyle] = useState('max_colors')
  const [useAI, setUseAI] = useState(false)
  const [generatedPreview, setGeneratedPreview] = useState(null)

  // ОЧИСТКА localStorage ПРИ ЗАГРУЗКЕ СТРАНИЦЫ
  useEffect(() => {
    // Очищаем данные о заказах и купонах при входе на страницу превью
    const clearStaleData = () => {
      try {
        // Очищаем данные о заказах
        localStorage.removeItem('pendingOrder')
        localStorage.removeItem('activeCoupon')
        
        // Очищаем любые временные данные превью если есть
        const keys = Object.keys(localStorage)
        keys.forEach(key => {
          if (key.startsWith('preview_') || key.startsWith('temp_')) {
            localStorage.removeItem(key)
          }
        })
        
        console.log('Cleared stale localStorage data on preview page')
      } catch (error) {
        console.error('Error clearing localStorage:', error)
      }
    }
    
    clearStaleData()
  }, [])

  const sizes = [
    { key: '21x30', title: t('mosaic_preview.sizes.21x30.title'), desc: t('mosaic_preview.sizes.21x30.desc') },
    { key: '30x40', title: t('mosaic_preview.sizes.30x40.title'), desc: t('mosaic_preview.sizes.30x40.desc') },
    { key: '40x40', title: t('mosaic_preview.sizes.40x40.title'), desc: t('mosaic_preview.sizes.40x40.desc') },
    { key: '40x50', title: t('mosaic_preview.sizes.40x50.title'), desc: t('mosaic_preview.sizes.40x50.desc') },
    { key: '40x60', title: t('mosaic_preview.sizes.40x60.title'), desc: t('mosaic_preview.sizes.40x60.desc') },
    { key: '50x70', title: t('mosaic_preview.sizes.50x70.title'), desc: t('mosaic_preview.sizes.50x70.desc') }
  ]

  const styles = [
    { key: 'grayscale', title: t('mosaic_preview.styles.grayscale.title'), desc: t('mosaic_preview.styles.grayscale.desc') },
    { key: 'skin_tones', title: t('mosaic_preview.styles.skin_tones.title'), desc: t('mosaic_preview.styles.skin_tones.desc') },
    { key: 'pop_art', title: t('mosaic_preview.styles.pop_art.title'), desc: t('mosaic_preview.styles.pop_art.desc') },
    { key: 'max_colors', title: t('mosaic_preview.styles.max_colors.title'), desc: t('mosaic_preview.styles.max_colors.desc') }
  ]

  const generatePreviewMutation = useMutation({
    mutationFn: async ({ file, size, style, useAI }) => {
      const formData = new FormData()
      formData.append('image', file)
      formData.append('size', size)
      formData.append('style', style)
      formData.append('use_ai', useAI ? 'true' : 'false')
      
      const response = await MosaicAPI.generatePreview(formData)
      return response
    },
    onSuccess: (data) => {
      setGeneratedPreview(data.preview_url)
      addNotification({
        type: 'success',
        message: t('mosaic_preview.notifications.success')
      })
    },
    onError: (error) => {
      addNotification({
        type: 'error',
        title: t('mosaic_preview.notifications.error'),
        message: error?.message || t('mosaic_preview.notifications.error')
      })
    }
  })

  const handleFileSelect = (event) => {
    const file = event.target.files[0]
    if (!file) return

        // Check file type
    if (!file.type.startsWith('image/')) {
      addNotification({
        type: 'error',
        message: 'Please select an image file'
      })
      return
    }

    // Check file size (max 15MB)
    if (file.size > 15 * 1024 * 1024) {
      addNotification({
        type: 'error',
        message: 'File size should not exceed 15MB'
      })
      return
    }

    setSelectedFile(file)

    // Create preview for display
    const reader = new FileReader()
    reader.onload = (e) => {
      setPreviewUrl(e.target.result)
    }
    reader.readAsDataURL(file)
  }

  const handleGeneratePreview = () => {
    if (!selectedFile) {
      addNotification({
        type: 'error',
        message: t('mosaic_preview.notifications.select_image')
      })
      return
    }

    generatePreviewMutation.mutate({
      file: selectedFile,
      size: selectedSize,
      style: selectedStyle,
      useAI: useAI
    })
  }

  const goBack = () => {
    navigate('/diamond-art')
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <div className="bg-white shadow-sm">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <div className="flex items-center space-x-4">
            <button
              onClick={goBack}
              className="inline-flex items-center text-gray-600 hover:text-gray-900 transition-colors"
            >
              <ArrowLeft className="w-5 h-5 mr-2" />
              {t('mosaic_preview.back')}
            </button>
            <div className="h-6 w-px bg-gray-300" />
            <h1 className="text-2xl font-bold text-gray-900">{t('mosaic_preview.title')}</h1>
          </div>
        </div>
      </div>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="grid lg:grid-cols-2 gap-8">
          
          {/* Left column - upload and settings */}
          <motion.div
            initial={{ opacity: 0, x: -20 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ duration: 0.5 }}
            className="space-y-6"
          >
            {/* Image upload */}
            <div className="bg-white rounded-2xl shadow-lg p-6">
              <h2 className="text-xl font-bold text-gray-900 mb-4 flex items-center">
                <Upload className="w-6 h-6 mr-2 text-brand-primary" />
                {t('mosaic_preview.upload_section.title')}
              </h2>
              
              <div className="border-2 border-dashed border-gray-300 rounded-xl p-8 text-center hover:border-brand-primary transition-colors">
                <input
                  type="file"
                  accept="image/*"
                  onChange={handleFileSelect}
                  className="hidden"
                  id="image-upload"
                />
                <label htmlFor="image-upload" className="cursor-pointer">
                  {previewUrl ? (
                    <div className="space-y-4">
                      <img
                        src={previewUrl}
                        alt="Preview of uploaded image"
                        className="max-h-48 mx-auto rounded-lg shadow-md"
                      />
                      <p className="text-sm text-gray-600">
                        {selectedFile?.name} ({Math.round(selectedFile?.size / 1024)}KB)
                      </p>
                      <p className="text-xs text-brand-primary">Click to select another image</p>
                    </div>
                  ) : (
                    <div className="space-y-4">
                      <Upload className="w-12 h-12 text-gray-400 mx-auto" />
                      <div>
                        <p className="text-lg font-medium text-gray-900">{t('mosaic_preview.upload_section.description')}</p>
                        <p className="text-sm text-gray-600 mt-1">
                          {t('mosaic_preview.upload_section.formats')}
                        </p>
                      </div>
                    </div>
                  )}
                </label>
              </div>
            </div>

            {/* Size selection */}
            <div className="bg-white rounded-2xl shadow-lg p-4 sm:p-6">
              <h3 className="text-lg font-bold text-gray-900 mb-4">{t('mosaic_preview.size_section.title')}</h3>
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
                {sizes.map((size) => (
                  <button
                    key={size.key}
                    onClick={() => setSelectedSize(size.key)}
                    className={`p-3 rounded-lg border-2 transition-all text-left ${
                      selectedSize === size.key
                        ? 'border-brand-primary bg-brand-primary/10 text-brand-primary'
                        : 'border-gray-200 hover:border-gray-300'
                    }`}
                  >
                    <div className="font-semibold text-sm sm:text-base">{size.title}</div>
                    <div className="text-xs sm:text-sm text-gray-600">{size.desc}</div>
                  </button>
                ))}
              </div>
            </div>

            {/* Style selection */}
            <div className="bg-white rounded-2xl shadow-lg p-4 sm:p-6">
              <h3 className="text-lg font-bold text-gray-900 mb-4">{t('mosaic_preview.style_section.title')}</h3>
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                {styles.map((style) => (
                  <button
                    key={style.key}
                    onClick={() => setSelectedStyle(style.key)}
                    className={`p-3 rounded-lg border-2 text-left transition-all ${
                      selectedStyle === style.key
                        ? 'border-brand-secondary bg-brand-secondary/10 text-brand-secondary'
                        : 'border-gray-200 hover:border-gray-300'
                    }`}
                  >
                    <div className="font-semibold text-sm sm:text-base">{style.title}</div>
                    <div className="text-xs sm:text-sm text-gray-600">{style.desc}</div>
                  </button>
                ))}
              </div>
            </div>

            {/* AI Processing Option */}
            <div className="mb-6">
              <div className="flex items-center space-x-3 p-4 border rounded-lg bg-gray-50">
                <input
                  type="checkbox"
                  id="useAI"
                  checked={useAI}
                  onChange={(e) => setUseAI(e.target.checked)}
                  className="w-4 h-4 text-brand-primary bg-gray-100 border-gray-300 rounded focus:ring-brand-primary focus:ring-2"
                />
                <label htmlFor="useAI" className="flex-1">
                  <div className="font-semibold text-gray-900">{t('mosaic_preview.ai_section.title')}</div>
                  <div className="text-sm text-gray-600">
                    {t('mosaic_preview.ai_section.description')}
                  </div>
                </label>
              </div>
            </div>

            {/* Generate button */}
            <button
              onClick={handleGeneratePreview}
              disabled={!selectedFile || generatePreviewMutation.isPending}
              className="w-full bg-gradient-to-r from-brand-primary to-brand-secondary text-white py-4 px-6 rounded-xl hover:from-brand-primary/90 hover:to-brand-secondary/90 disabled:opacity-50 disabled:cursor-not-allowed font-semibold text-lg transition-all duration-200 flex items-center justify-center space-x-2"
            >
              {generatePreviewMutation.isPending ? (
                <>
                  <Loader2 className="w-5 h-5 animate-spin" />
                  <span>{t('mosaic_preview.generating')}</span>
                </>
              ) : (
                <>
                  <Eye className="w-5 h-5" />
                  <span>{t('mosaic_preview.generate_button')}</span>
                </>
              )}
            </button>

            {/* Marketplace Cards */}
            <MarketplaceCards 
              selectedSize={selectedSize} 
              selectedStyle={selectedStyle} 
            />
          </motion.div>

          {/* Right column - result */}
          <motion.div
            initial={{ opacity: 0, x: 20 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ duration: 0.5, delay: 0.2 }}
            className="bg-white rounded-2xl shadow-lg p-6"
          >
            <h2 className="text-xl font-bold text-gray-900 mb-4 flex items-center">
              <Eye className="w-6 h-6 mr-2 text-brand-secondary" />
              {t('mosaic_preview.preview_section.title')}
            </h2>
            
            {generatedPreview ? (
              <div className="space-y-4">
                <img
                  src={generatedPreview}
                  alt="Mosaic preview"
                  className="w-full rounded-lg shadow-lg"
                  onError={(e) => {
                    console.error('Failed to load preview image:', generatedPreview)
                    e.target.style.display = 'none'
                  }}
                  onLoad={() => {
                    console.log('Preview image loaded successfully:', generatedPreview)
                  }}
                />

                <div className="flex space-x-3">
                  <button
                    onClick={() => {
                      console.log('Opening preview in new tab:', generatedPreview)
                      window.open(generatedPreview, '_blank', 'noopener,noreferrer')
                    }}
                    className="flex-1 bg-brand-primary text-white py-3 px-4 rounded-lg hover:bg-brand-primary/90 font-semibold transition-colors flex items-center justify-center space-x-2"
                  >
                    <Eye className="w-4 h-4" />
                    <span>{t('mosaic_preview.view_full')}</span>
                  </button>
                  <button
                    onClick={() => {
                      // Download through special endpoint with proper headers
                      const previewId = generatedPreview.split('/').pop().replace('.png', '')
                      const downloadUrl = `/api/preview/${previewId}/download`
                      window.open(downloadUrl, '_blank')
                    }}
                    className="flex-1 bg-brand-secondary text-white py-3 px-4 rounded-lg hover:bg-brand-secondary/90 font-semibold transition-colors flex items-center justify-center space-x-2"
                  >
                    <Download className="w-4 h-4" />
                    <span>{t('mosaic_preview.download')}</span>
                  </button>
                </div>
              </div>
            ) : (
              <div className="text-center py-12">
                <div className="w-24 h-24 bg-gray-100 rounded-full flex items-center justify-center mx-auto mb-4">
                  <Eye className="w-12 h-12 text-gray-400" />
                </div>
                <h3 className="text-lg font-medium text-gray-900 mb-2">{t('mosaic_preview.preview_section.preview_will_be_here')}</h3>
                <p className="text-gray-600">
                  {t('mosaic_preview.preview_section.description')}
                </p>
              </div>
            )}
          </motion.div>
        </div>
      </div>
    </div>
  )
}

export default MosaicPreviewPage