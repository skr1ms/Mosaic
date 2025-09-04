import React, { useState, useEffect, useRef } from 'react'
import { useTranslation } from 'react-i18next'
import { motion } from 'framer-motion'
import { Upload, Ruler, ArrowRight, Info, Image as ImageIcon, Check } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import { useUIStore } from '../store/partnerStore'
import ImageEditor from '../components/ImageEditor'

const PreviewPage = () => {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { addNotification } = useUIStore()
  
  const fileInputRef = useRef(null)
  
  const [selectedSize, setSelectedSize] = useState('')
  const [selectedFile, setSelectedFile] = useState(null)
  const [previewUrl, setPreviewUrl] = useState(null)
  const [isUploading, setIsUploading] = useState(false)
  const [editedImageUrl, setEditedImageUrl] = useState(null)
  const [showEditor, setShowEditor] = useState(false)

    useEffect(() => {
    try {
      localStorage.removeItem('pendingOrder')
      localStorage.removeItem('activeCoupon')
      
      const keys = Object.keys(localStorage)
      keys.forEach(key => {
        if (key.startsWith('preview_') || key.startsWith('temp_') || key.startsWith('shop_')) {
          localStorage.removeItem(key)
        }
      })
    } catch (error) {
      console.error('Error clearing localStorage:', error)
    }
  }, [])

    useEffect(() => {
    const checkImageData = () => {
      try {
        const savedImageData = localStorage.getItem('diamondMosaic_selectedImage')
        if (!savedImageData && (selectedFile || previewUrl)) {
                    setSelectedFile(null)
          setPreviewUrl(null)
          setEditedImageUrl(null)
          setShowEditor(false)
        }
      } catch (error) {
        console.error('Error checking image data:', error)
      }
    }

        window.addEventListener('focus', checkImageData)
    checkImageData() 
    return () => {
      window.removeEventListener('focus', checkImageData)
    }
  }, [selectedFile, previewUrl])



  const getSizes = () => [
    { 
      key: '21x30', 
      title: t('diamond_mosaic_page.size_selection.sizes.21x30'), 
      desc: t('diamond_mosaic_page.size_selection.details.21x30.desc')
    },
    { 
      key: '30x40', 
      title: t('diamond_mosaic_page.size_selection.sizes.30x40'), 
      desc: t('diamond_mosaic_page.size_selection.details.30x40.desc')
    },
    { 
      key: '40x40', 
      title: t('diamond_mosaic_page.size_selection.sizes.40x40'), 
      desc: t('diamond_mosaic_page.size_selection.details.40x40.desc')
    },
    { 
      key: '40x50', 
      title: t('diamond_mosaic_page.size_selection.sizes.40x50'), 
      desc: t('diamond_mosaic_page.size_selection.details.40x50.desc')
    },
    { 
      key: '40x60', 
      title: t('diamond_mosaic_page.size_selection.sizes.40x60'), 
      desc: t('diamond_mosaic_page.size_selection.details.40x60.desc')
    },
    { 
      key: '50x70', 
      title: t('diamond_mosaic_page.size_selection.sizes.50x70'), 
      desc: t('diamond_mosaic_page.size_selection.details.50x70.desc')
    }
  ]

  const handleFileSelect = (event) => {
    const file = event.target.files[0]
    if (!file) return

        if (!file.type.startsWith('image/')) {
      addNotification({
        type: 'error',
        message: t('diamond_mosaic_page.upload_section.file_type_error')
      })
      return
    }

        if (file.size > 10 * 1024 * 1024) {
      addNotification({
        type: 'error',
        message: t('diamond_mosaic_page.upload_section.file_size_error')
      })
      return
    }

    setSelectedFile(file)
    
        const reader = new FileReader()
    reader.onload = (e) => {
      setPreviewUrl(e.target.result)
      setShowEditor(true)
    }
    reader.readAsDataURL(file)
  }

  const handleEditorSave = (editedUrl, editorParams) => {
    setEditedImageUrl(editedUrl)
    setShowEditor(false)
  }
  
  const handleEditorCancel = () => {
    setShowEditor(false)
  }

  const handleSizeSelect = (sizeKey) => {
    setSelectedSize(sizeKey)
  }

  const handleContinue = () => {
    if (!selectedFile || !selectedSize) {
      addNotification({
        type: 'error',
        message: t('diamond_mosaic_page.size_and_image_required')
      })
      return
    }

        const finalImageUrl = editedImageUrl || previewUrl

    try {
            localStorage.setItem('diamondMosaic_selectedImage', JSON.stringify({
        size: selectedSize,
        fileName: selectedFile.name,
        previewUrl: finalImageUrl,
        timestamp: Date.now(),
        hasEdits: editedImageUrl !== null
      }))
      
            localStorage.setItem('diamondMosaic_editorSettings', JSON.stringify({
        size: selectedSize,
        style: null,         returnTo: '/preview/styles'
      }))
      
            if (editedImageUrl) {
                fetch(editedImageUrl)
          .then(res => res.blob())
          .then(blob => {
            const fileUrl = URL.createObjectURL(blob)
            sessionStorage.setItem('diamondMosaic_fileUrl', fileUrl)
          })
      } else {
        const fileUrl = URL.createObjectURL(selectedFile)
        sessionStorage.setItem('diamondMosaic_fileUrl', fileUrl)
      }
      
    } catch (error) {
      console.error('Error saving image data:', error)
    }

        navigate('/preview/styles')
  }

  const handleRemoveImage = () => {
        setSelectedFile(null)
    setPreviewUrl(null)
    setEditedImageUrl(null)
    setShowEditor(false)
    
        try {
      localStorage.removeItem('diamondMosaic_selectedImage')
      localStorage.removeItem('diamondMosaic_editorSettings')
      sessionStorage.removeItem('diamondMosaic_fileUrl')
    } catch (error) {
      console.error('Error clearing localStorage:', error)
    }
    
        if (fileInputRef.current) {
      fileInputRef.current.value = ''
    }
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-purple-50 via-pink-50 to-blue-50">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
        
        {}
        <motion.div
          initial={{ opacity: 0, y: -20 }}
          animate={{ opacity: 1, y: 0 }}
          className="text-center mb-12"
        >
          <h1 className="text-4xl md:text-5xl font-bold text-gray-900 mb-4">
            {t('diamond_mosaic_page.title')}
          </h1>
          <p className="text-lg md:text-xl text-gray-600 max-w-2xl mx-auto">
            {t('diamond_mosaic_page.upload_section.subtitle')}
          </p>
        </motion.div>

        
        <motion.section
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.1 }}
          className="mb-16"
        >
          <h2 className="text-2xl font-bold text-gray-900 mb-8 text-center flex items-center justify-center">
            <Ruler className="w-6 h-6 mr-2 text-purple-600" />
            {t('diamond_mosaic_page.size_selection.title')}
          </h2>
          
          <div className="grid grid-cols-2 md:grid-cols-3 gap-6">
            {getSizes().map((size, index) => {
              const isSelected = selectedSize === size.key
              
              
              const rectangleClasses = {
                '21x30': 'w-8 h-12',
                '30x40': 'w-10 h-14', 
                '40x40': 'w-12 h-12',
                '40x50': 'w-12 h-16',
                '40x60': 'w-12 h-18',
                '50x70': 'w-14 h-22'
              }

              return (
                <motion.div
                  key={size.key}
                  initial={{ opacity: 0, y: 20 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ duration: 0.5, delay: index * 0.1 }}
                  onClick={() => handleSizeSelect(size.key)}
                  className={`relative bg-white rounded-2xl shadow-lg p-6 cursor-pointer transition-all duration-300 transform hover:scale-105 hover:shadow-xl ${
                    isSelected 
                      ? 'ring-4 ring-purple-500 border-purple-500' 
                      : 'border border-gray-200 hover:border-gray-300'
                  }`}
                >
                  {isSelected && (
                    <div className="absolute -top-2 -right-2 w-6 h-6 bg-purple-500 rounded-full flex items-center justify-center">
                      <Check className="w-4 h-4 text-white" />
                    </div>
                  )}
                  
                  <div className="flex flex-col items-center text-center">
                    <div className={`mb-4 rounded ${rectangleClasses[size.key]} ${
                      isSelected ? 'bg-purple-500' : 'bg-purple-300'
                    }`} />
                    
                    <h3 className="text-xl font-semibold text-gray-900 mb-2">
                      {size.title}
                    </h3>
                    
                    <p className="text-sm text-gray-600 mb-3">
                      {size.desc}
                    </p>
                  </div>
                </motion.div>
              )
            })}
          </div>
        </motion.section>

        {}
        <motion.section
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.3 }}
          className="mb-12"
        >
          <h2 className="text-2xl font-bold text-gray-900 mb-8 text-center flex items-center justify-center">
            <ImageIcon className="w-6 h-6 mr-2 text-purple-600" />
            {t('diamond_mosaic_page.upload_section.title')}
          </h2>

          <div className="max-w-4xl mx-auto">
            {!previewUrl ? (
              <div className="border-2 border-dashed border-gray-300 rounded-2xl p-12 text-center bg-white hover:border-purple-400 transition-colors">
                <input
                  ref={fileInputRef}
                  type="file"
                  id="image-upload"
                  accept="image/*"
                  onChange={handleFileSelect}
                  className="hidden"
                />
                <label
                  htmlFor="image-upload"
                  className="cursor-pointer block"
                >
                  <Upload className="w-16 h-16 text-gray-400 mx-auto mb-6" />
                  <p className="text-2xl font-medium text-gray-700 mb-3">
                    {t('diamond_mosaic_page.upload_section.select_image')}
                  </p>
                  <p className="text-gray-500 mb-6">
                    {t('diamond_mosaic_page.upload_section.formats')}
                  </p>
                  <div className="inline-block bg-purple-600 text-white px-8 py-3 rounded-xl hover:bg-purple-700 transition-colors font-semibold">
                    {t('diamond_mosaic_page.upload_section.button')}
                  </div>
                </label>
              </div>
            ) : (
              <div className="max-w-4xl mx-auto">
                {showEditor ? (
                  <ImageEditor
                    imageUrl={previewUrl}
                    onSave={handleEditorSave}
                    onCancel={handleEditorCancel}
                    title={t('diamond_mosaic_page.image_editor.setup_title')}
                    showCropHint={true}
                  />
                ) : (
                  <div className="bg-white rounded-2xl p-8 border border-gray-200 shadow-lg">
                    <div className="flex items-center justify-between mb-8">
                      <h4 className="text-xl font-semibold text-gray-800 flex items-center">
                        {t('diamond_mosaic_page.image_editor.setup_title')}
                      </h4>
                      <button
                        onClick={handleRemoveImage}
                        className="text-red-500 hover:text-red-600 flex items-center text-sm font-medium"
                      >
                        ✕ {t('diamond_mosaic_page.image_editor.delete_image')}
                      </button>
                    </div>
                    
                    
                    <div className="relative mb-8">
                      <div className="bg-gray-50 rounded-2xl p-6 flex justify-center">
                        <img 
                          src={editedImageUrl || previewUrl} 
                          alt="Preview"
                          className="max-w-full rounded-lg shadow-lg"
                          style={{ maxHeight: '500px' }}
                        />
                      </div>
                    </div>
                    
                    {}
                    <div className="text-center mb-8">
                      <button
                        onClick={() => setShowEditor(true)}
                        className="bg-purple-100 text-purple-700 px-6 py-3 rounded-xl hover:bg-purple-200 transition-colors font-medium"
                      >
                        Изменить изображение
                      </button>
                    </div>

                    {}
                    <div className="text-center bg-blue-50 rounded-xl p-6 border border-blue-200">
                      <div className="flex items-center justify-center mb-2">
                        <div className="w-2 h-2 bg-green-500 rounded-full mr-2"></div>
                        <p className="font-medium text-gray-800">{selectedFile?.name}</p>
                      </div>
                      <p className="text-gray-600">{t('diamond_mosaic_page.image_editor.file_info')}</p>
                    </div>
                  </div>
                )}
              </div>
            )}
          </div>
        </motion.section>

        
        {selectedSize && previewUrl && (
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            className="text-center"
          >
            <button
              onClick={handleContinue}
              disabled={isUploading}
              className="bg-gradient-to-r from-purple-600 to-pink-600 text-white px-12 py-4 rounded-xl font-semibold text-lg hover:from-purple-700 hover:to-pink-700 transition-all duration-300 shadow-lg hover:shadow-xl disabled:opacity-50 disabled:cursor-not-allowed flex items-center mx-auto"
            >
              {t('diamond_mosaic_page.navigation.continue')}
              <ArrowRight className="w-5 h-5 ml-2" />
            </button>
          </motion.div>
        )}
      </div>
    </div>
  )
}

export default PreviewPage
