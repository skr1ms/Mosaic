import React, { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { motion } from 'framer-motion'
import { useNavigate } from 'react-router-dom'
import { ArrowLeft, ArrowRight, Palette, Sparkles, Sun, Moon, Loader2 } from 'lucide-react'
import { useUIStore } from '../store/partnerStore'
import { MosaicAPI } from '../api/client'

const PreviewStylesPage = () => {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { addNotification } = useUIStore()
  
  const [imageData, setImageData] = useState(null)
  const [selectedStyle, setSelectedStyle] = useState(null)
  const [isGeneratingPreviews, setIsGeneratingPreviews] = useState(false)
  const [stylePreviews, setStylePreviews] = useState({})

  const styles = [
    {
      key: 'natural',
      title: t('diamond_mosaic_styles.styles.natural.title'),
      description: t('diamond_mosaic_styles.styles.natural.description'),
      icon: <Palette className="w-6 h-6" />,
      color: 'from-green-400 to-blue-500'
    },
    {
      key: 'enhanced',
      title: t('diamond_mosaic_styles.styles.enhanced.title'),
      description: t('diamond_mosaic_styles.styles.enhanced.description'),
      icon: <Sparkles className="w-6 h-6" />,
      color: 'from-pink-400 to-purple-500'
    },
    {
      key: 'vintage',
      title: t('diamond_mosaic_styles.styles.vintage.title'),
      description: t('diamond_mosaic_styles.styles.vintage.description'),
      icon: <Sun className="w-6 h-6" />,
      color: 'from-yellow-400 to-orange-500'
    },
    {
      key: 'monochrome',
      title: t('diamond_mosaic_styles.styles.monochrome.title'),
      description: t('diamond_mosaic_styles.styles.monochrome.description'),
      icon: <Moon className="w-6 h-6" />,
      color: 'from-gray-400 to-gray-600'
    }
  ]

  useEffect(() => {
        try {
      const savedImageData = localStorage.getItem('diamondMosaic_selectedImage')
      if (!savedImageData) {
        navigate('/preview')
        return
      }
      
      const parsedData = JSON.parse(savedImageData)
      setImageData(parsedData)
      
            generateStylePreviews(parsedData)
      
    } catch (error) {
      console.error('Error loading image data:', error)
      navigate('/preview')
    }
  }, [navigate])

  const generateStylePreviews = async (data) => {
    setIsGeneratingPreviews(true)
    
    try {
      const fileUrl = sessionStorage.getItem('diamondMosaic_fileUrl')
      if (!fileUrl) {
                console.log('No file URL found, skipping preview generation')
        setIsGeneratingPreviews(false)
        return
      }
      
            let blob
      try {
        const response = await fetch(fileUrl)
        blob = await response.blob()
      } catch (error) {
        console.error('Error fetching file from URL:', error)
        setIsGeneratingPreviews(false)
        return
      }
      
      const previews = {}
      
            for (const style of styles) {
        try {
                              previews[style.key] = data.previewUrl
          
                    setStylePreviews({ ...previews })
          
                    await new Promise(resolve => setTimeout(resolve, 500))
          
        } catch (error) {
          console.error(`Error generating preview for style ${style.key}:`, error)
          previews[style.key] = null
        }
      }
      
    } catch (error) {
      console.error('Error generating style previews:', error)
      addNotification({
        type: 'error',
        message: t('diamond_mosaic_styles.error_generating_previews')
      })
    } finally {
      setIsGeneratingPreviews(false)
    }
  }

  const handleStyleSelect = (styleKey) => {
    setSelectedStyle(styleKey)
  }

  const handleContinue = () => {
    if (!selectedStyle) {
      addNotification({
        type: 'error',
        message: t('diamond_mosaic_styles.select_style_error')
      })
      return
    }

    try {
            const updatedImageData = {
        ...imageData,
        selectedStyle: selectedStyle,
        stylePreview: stylePreviews[selectedStyle]
      }
      
      localStorage.setItem('diamondMosaic_selectedImage', JSON.stringify(updatedImageData))
      
            navigate('/preview/album')
      
    } catch (error) {
      console.error('Error saving style selection:', error)
      addNotification({
        type: 'error',
        message: t('diamond_mosaic_styles.error_saving_style')
      })
    }
  }

  const handleBack = () => {
    navigate('/preview')
  }

  if (!imageData) {
    return (
      <div className="min-h-screen bg-white flex items-center justify-center">
        <div className="text-gray-600">{t('diamond_mosaic_styles.loading')}</div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-white">
      <div className="max-w-6xl mx-auto px-6 py-8">
        
        {}
        <div className="mb-8">
                     <button
             onClick={handleBack}
             className="flex items-center text-purple-600 hover:text-purple-700 mb-4 transition-colors"
           >
             <ArrowLeft className="w-5 h-5 mr-2" />
             {t('diamond_mosaic_styles.back_to_editor')}
           </button>
           
           <h1 className="text-3xl font-bold text-gray-900 text-center mb-2">
             {t('diamond_mosaic_styles.page_title')}
           </h1>
           <p className="text-gray-600 text-center">
             {t('diamond_mosaic_styles.size_label', { size: imageData.size })}
           </p>
        </div>

        
        <div className="mb-8 text-center">
          <img 
            src={imageData.previewUrl} 
            alt="Uploaded image"
            className="w-48 h-48 object-cover rounded-lg mx-auto border-2 border-gray-200"
          />
        </div>

        
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

        
        {isGeneratingPreviews && (
          <motion.div 
            className="text-center mb-8"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
          >
                         <div className="inline-flex items-center px-6 py-3 bg-purple-100 rounded-full text-purple-700">
               <Loader2 className="w-5 h-5 mr-3 animate-spin" />
               {t('diamond_mosaic_styles.generating_previews')}
             </div>
          </motion.div>
        )}

        {}
        <div className="flex gap-4 max-w-md mx-auto">
                     <button
             onClick={handleBack}
             className="flex-1 py-4 px-6 bg-white border-2 border-gray-300 text-gray-700 rounded-xl font-medium hover:bg-gray-50 transition-colors"
           >
             {t('diamond_mosaic_styles.back')}
           </button>
           
           <button
             onClick={handleContinue}
             disabled={!selectedStyle}
             className="flex-1 py-4 px-6 bg-gradient-to-r from-purple-600 to-pink-600 text-white rounded-xl font-medium hover:from-purple-700 hover:to-pink-700 transition-all disabled:opacity-50 disabled:cursor-not-allowed"
           >
             {t('diamond_mosaic_styles.create_preview')}
           </button>
        </div>

      </div>
    </div>
  )
}

export default PreviewStylesPage
