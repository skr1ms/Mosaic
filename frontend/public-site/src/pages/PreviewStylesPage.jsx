import React, { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { motion } from 'framer-motion'
import { useNavigate } from 'react-router-dom'
import { ArrowLeft, ArrowRight, Palette, Sparkles, Sun, Moon, Loader2, Check } from 'lucide-react'
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
      key: 'max_color',
      title: 'Max Color',
      description: 'Максимальная насыщенность цветов',
      icon: <Palette className="w-6 h-6" />,
      color: 'from-red-400 to-yellow-500'
    },
    {
      key: 'pop_art',
      title: 'Pop Art',
      description: 'Яркий поп-арт стиль',
      icon: <Sparkles className="w-6 h-6" />,
      color: 'from-pink-400 to-purple-500'
    },
    {
      key: 'watercolor',
      title: 'Акварель',
      description: 'Нежные акварельные переходы',
      icon: <Sun className="w-6 h-6" />,
      color: 'from-blue-400 to-cyan-500'
    },
    {
      key: 'oil_painting',
      title: 'Масляная живопись',
      description: 'Классический стиль масляной живописи',
      icon: <Moon className="w-6 h-6" />,
      color: 'from-amber-400 to-brown-600'
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
            <motion.div
              key={style.key}
              onClick={() => handleStyleSelect(style.key)}
              whileHover={{ scale: 1.05 }}
              whileTap={{ scale: 0.95 }}
              className={`
                relative overflow-hidden rounded-xl border-2 cursor-pointer transition-all hover:shadow-xl
                ${selectedStyle === style.key 
                  ? 'border-purple-500 ring-4 ring-purple-200' 
                  : 'border-gray-200 hover:border-purple-300'
                }
              `}
            >
              {/* Preview Image */}
              <div className="relative h-48 bg-gradient-to-br ${style.color} overflow-hidden">
                {stylePreviews[style.key] ? (
                  <img 
                    src={stylePreviews[style.key]} 
                    alt={style.title}
                    className="w-full h-full object-cover"
                  />
                ) : (
                  <div className="w-full h-full flex items-center justify-center">
                    {isGeneratingPreviews ? (
                      <Loader2 className="w-8 h-8 text-white animate-spin" />
                    ) : (
                      <div className={`w-full h-full bg-gradient-to-br ${style.color}`} />
                    )}
                  </div>
                )}
                
                {/* Selected indicator */}
                {selectedStyle === style.key && (
                  <div className="absolute top-2 right-2 w-8 h-8 bg-white rounded-full flex items-center justify-center shadow-lg">
                    <Check className="w-5 h-5 text-purple-600" />
                  </div>
                )}
              </div>
              
              {/* Style Info */}
              <div className="p-4 bg-white">
                <div className="flex items-center justify-center mb-2">
                  <div className={`text-transparent bg-clip-text bg-gradient-to-r ${style.color}`}>
                    {style.icon}
                  </div>
                </div>
                <h3 className="text-lg font-semibold text-gray-900 mb-1 text-center">
                  {style.title}
                </h3>
                <p className="text-sm text-gray-600 text-center">
                  {style.description}
                </p>
              </div>
            </motion.div>
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
