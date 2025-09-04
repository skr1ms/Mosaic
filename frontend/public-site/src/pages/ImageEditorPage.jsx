import React, { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { motion } from 'framer-motion'
import { useNavigate } from 'react-router-dom'
import { useUIStore } from '../store/partnerStore'
import ImageEditor from '../components/ImageEditor'

const ImageEditorPage = () => {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { addNotification } = useUIStore()
  
  const [imageData, setImageData] = useState(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
        try {
      const savedImageData = localStorage.getItem('diamondMosaic_selectedImage')
      const editorSettings = localStorage.getItem('diamondMosaic_editorSettings')
      
      if (!savedImageData) {
        addNotification({
          type: 'error',
          message: t('image_editor.image_not_found')
        })
        navigate('/preview')
        return
      }
      
      const parsedData = JSON.parse(savedImageData)
      const parsedSettings = editorSettings ? JSON.parse(editorSettings) : {}
      
      setImageData({
        ...parsedData,
        settings: parsedSettings
      })
      
    } catch (error) {
      console.error('Error loading image data:', error)
      addNotification({
        type: 'error',
        message: t('image_editor.error_loading_data')
      })
      navigate('/preview')
    } finally {
      setLoading(false)
    }
  }, [navigate, addNotification, t])

  const handleSave = (editedImageUrl, editorParams) => {
    try {
            const updatedImageData = {
        ...imageData,
        previewUrl: editedImageUrl,
        hasEdits: true,
        editorParams,
        timestamp: Date.now()
      }
      
      localStorage.setItem('diamondMosaic_selectedImage', JSON.stringify(updatedImageData))
      
      addNotification({
        type: 'success',
        message: t('image_editor.save_success')
      })
      
            const returnTo = imageData?.settings?.returnTo || '/preview/styles'
      navigate(returnTo)
      
    } catch (error) {
      console.error('Error saving edited image:', error)
      addNotification({
        type: 'error',
        message: t('image_editor.save_error')
      })
    }
  }

  const handleCancel = () => {
        try {
      localStorage.removeItem('diamondMosaic_selectedImage')
      localStorage.removeItem('diamondMosaic_editorSettings')  
      sessionStorage.removeItem('diamondMosaic_fileUrl')
    } catch (error) {
      console.error('Error clearing image data:', error)
    }
    
        navigate('/preview')
  }

  if (loading) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-purple-50 via-pink-50 to-blue-50 flex items-center justify-center">
        <div className="text-center">
          <div className="w-16 h-16 border-4 border-purple-600 border-t-transparent rounded-full animate-spin mx-auto mb-4"></div>
          <p className="text-gray-600">{t('image_editor.loading')}</p>
        </div>
      </div>
    )
  }

  if (!imageData) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-purple-50 via-pink-50 to-blue-50 flex items-center justify-center">
        <div className="text-center">
          <p className="text-gray-600">{t('image_editor.image_not_found')}</p>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-purple-50 via-pink-50 to-blue-50">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        
        
        <motion.div
          initial={{ opacity: 0, y: -20 }}
          animate={{ opacity: 1, y: 0 }}
          className="text-center mb-8"
        >
          <h1 className="text-4xl font-bold text-gray-800 mb-4">
            {t('image_editor.page_title')}
          </h1>
          <p className="text-xl text-gray-600 max-w-2xl mx-auto">
            {t('image_editor.page_description')}
          </p>
          

        </motion.div>

        {}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.2 }}
          className="max-w-6xl mx-auto"
        >
          <ImageEditor
            imageUrl={imageData.previewUrl}
            onSave={handleSave}
            onCancel={handleCancel}
            title="Настройте изображение"
            showCropHint={true}
            aspectRatio={1}             fileName={imageData.fileName}
          />
        </motion.div>

        {}
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ delay: 0.4 }}
          className="mt-8 max-w-4xl mx-auto"
        >
          <div className="bg-white/70 backdrop-blur-sm rounded-xl p-6 shadow-lg">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">
              Рекомендации по редактированию:
            </h3>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4 text-sm text-gray-600">
              <div className="flex items-start space-x-3">
                <div className="w-2 h-2 bg-purple-500 rounded-full mt-2 flex-shrink-0"></div>
                <div>
                  <strong>Кадрирование:</strong> Выберите наиболее важную часть изображения для мозаики
                </div>
              </div>
              <div className="flex items-start space-x-3">
                <div className="w-2 h-2 bg-blue-500 rounded-full mt-2 flex-shrink-0"></div>
                <div>
                  <strong>Поворот:</strong> Используйте поворот для правильной ориентации изображения
                </div>
              </div>
              <div className="flex items-start space-x-3">
                <div className="w-2 h-2 bg-green-500 rounded-full mt-2 flex-shrink-0"></div>
                <div>
                  <strong>Масштаб:</strong> Увеличьте важные детали для лучшего результата
                </div>
              </div>
            </div>
          </div>
        </motion.div>
      </div>
    </div>
  )
}

export default ImageEditorPage
