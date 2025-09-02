import React, { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { motion } from 'framer-motion'
import { Upload, Ruler, ArrowRight, Info, Image } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import { useUIStore } from '../store/partnerStore'

const DiamondMosaicPage = () => {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { addNotification } = useUIStore()
  
  const [selectedSize, setSelectedSize] = useState('')
  const [selectedFile, setSelectedFile] = useState(null)
  const [previewUrl, setPreviewUrl] = useState(null)
  const [isUploading, setIsUploading] = useState(false)

  // Очистка localStorage при загрузке страницы
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

  const sizes = [
    { 
      key: '21x30', 
      title: '21×30 см', 
      desc: 'Начальный уровень детализации',
      price: 'от 990 ₽',
      details: '8,000 страз'
    },
    { 
      key: '30x40', 
      title: '30×40 см', 
      desc: 'Хорошая детализация',
      price: 'от 1,490 ₽',
      details: '16,000 страз'
    },
    { 
      key: '40x40', 
      title: '40×40 см', 
      desc: 'Высокая детализация',
      price: 'от 1,990 ₽',
      details: '21,000 страз'
    },
    { 
      key: '40x50', 
      title: '40×50 см', 
      desc: 'Очень высокая детализация',
      price: 'от 2,490 ₽',
      details: '26,000 страз'
    },
    { 
      key: '40x60', 
      title: '40×60 см', 
      desc: 'Максимальная детализация',
      price: 'от 2,990 ₽',
      details: '32,000 страз'
    },
    { 
      key: '50x70', 
      title: '50×70 см', 
      desc: 'Профессиональный уровень',
      price: 'от 3,990 ₽',
      details: '46,000 страз'
    }
  ]

  const handleFileSelect = (event) => {
    const file = event.target.files[0]
    if (!file) return

    // Проверка типа файла
    if (!file.type.startsWith('image/')) {
      addNotification({
        type: 'error',
        message: 'Пожалуйста, выберите изображение'
      })
      return
    }

    // Проверка размера файла (максимум 10MB)
    if (file.size > 10 * 1024 * 1024) {
      addNotification({
        type: 'error',
        message: 'Размер файла не должен превышать 10MB'
      })
      return
    }

    setSelectedFile(file)
    
    // Создаем превью
    const reader = new FileReader()
    reader.onload = (e) => {
      setPreviewUrl(e.target.result)
    }
    reader.readAsDataURL(file)
  }

  const handleSizeSelect = (sizeKey) => {
    setSelectedSize(sizeKey)
  }

  const handleContinue = () => {
    if (!selectedFile || !selectedSize) {
      addNotification({
        type: 'error',
        message: 'Пожалуйста, выберите размер и загрузите изображение'
      })
      return
    }

    // Сохраняем данные для следующего шага
    const imageData = {
      file: selectedFile,
      previewUrl: previewUrl,
      size: selectedSize,
      timestamp: Date.now()
    }

    try {
      localStorage.setItem('diamondMosaic_selectedImage', JSON.stringify({
        size: selectedSize,
        fileName: selectedFile.name,
        previewUrl: previewUrl,
        timestamp: Date.now()
      }))
      
      // Создаем временный URL для файла
      const fileUrl = URL.createObjectURL(selectedFile)
      sessionStorage.setItem('diamondMosaic_fileUrl', fileUrl)
      
    } catch (error) {
      console.error('Error saving image data:', error)
    }

    // Переходим к выбору стиля
    navigate('/diamond-mosaic/styles')
  }

  const handleRemoveImage = () => {
    setSelectedFile(null)
    setPreviewUrl(null)
    
    // Очищаем input
    const fileInput = document.getElementById('image-upload')
    if (fileInput) {
      fileInput.value = ''
    }
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-purple-50 via-pink-50 to-blue-50 py-8 px-4">
      <div className="container mx-auto max-w-6xl">
        
        {/* Заголовок */}
        <motion.div
          initial={{ opacity: 0, y: -20 }}
          animate={{ opacity: 1, y: 0 }}
          className="text-center mb-12"
        >
          <h1 className="text-4xl md:text-5xl font-bold bg-gradient-to-r from-purple-600 to-pink-600 bg-clip-text text-transparent mb-4">
            Алмазная мозаика
          </h1>
          <p className="text-lg md:text-xl text-gray-600 max-w-2xl mx-auto">
            Чем больше размер, тем детальнее получится картинка. 
            Выберите размер и загрузите изображение для создания превью.
          </p>
        </motion.div>

        <div className="grid lg:grid-cols-2 gap-12 items-start">
          
          {/* Выбор размера */}
          <motion.div
            initial={{ opacity: 0, x: -20 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ delay: 0.1 }}
          >
            <h2 className="text-2xl font-bold text-gray-800 mb-6 flex items-center">
              <Ruler className="w-6 h-6 mr-2 text-purple-600" />
              Выберите размер
            </h2>
            
            <div className="grid gap-4">
              {sizes.map((size, index) => (
                <motion.div
                  key={size.key}
                  initial={{ opacity: 0, y: 20 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ delay: 0.1 + index * 0.05 }}
                  className={`p-4 rounded-xl border-2 cursor-pointer transition-all duration-300 ${
                    selectedSize === size.key
                      ? 'border-purple-500 bg-purple-50 shadow-lg'
                      : 'border-gray-200 bg-white hover:border-purple-300 hover:shadow-md'
                  }`}
                  onClick={() => handleSizeSelect(size.key)}
                >
                  <div className="flex justify-between items-center">
                    <div>
                      <h3 className="font-semibold text-gray-800">{size.title}</h3>
                      <p className="text-sm text-gray-600">{size.desc}</p>
                      <p className="text-xs text-gray-500 mt-1">{size.details}</p>
                    </div>
                    <div className="text-right">
                      <p className="font-bold text-purple-600">{size.price}</p>
                    </div>
                  </div>
                </motion.div>
              ))}
            </div>
          </motion.div>

          {/* Загрузка изображения */}
          <motion.div
            initial={{ opacity: 0, x: 20 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ delay: 0.2 }}
          >
            <h2 className="text-2xl font-bold text-gray-800 mb-6 flex items-center">
              <Image className="w-6 h-6 mr-2 text-purple-600" />
              Загрузите изображение
            </h2>

            {!previewUrl ? (
              <div className="border-2 border-dashed border-gray-300 rounded-xl p-8 text-center bg-white hover:border-purple-400 transition-colors">
                <input
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
                  <Upload className="w-12 h-12 text-gray-400 mx-auto mb-4" />
                  <p className="text-lg font-medium text-gray-700 mb-2">
                    Выберите изображение
                  </p>
                  <p className="text-sm text-gray-500 mb-4">
                    PNG, JPG до 10MB
                  </p>
                  <div className="inline-block bg-purple-600 text-white px-6 py-2 rounded-lg hover:bg-purple-700 transition-colors">
                    Выбрать файл
                  </div>
                </label>
              </div>
            ) : (
              <div className="bg-white rounded-xl p-4 border border-gray-200">
                <div className="relative">
                  <img
                    src={previewUrl}
                    alt="Preview"
                    className="w-full h-64 object-cover rounded-lg"
                  />
                  <button
                    onClick={handleRemoveImage}
                    className="absolute top-2 right-2 bg-red-500 text-white w-8 h-8 rounded-full flex items-center justify-center hover:bg-red-600 transition-colors"
                  >
                    ×
                  </button>
                </div>
                <p className="text-sm text-gray-600 mt-2">
                  {selectedFile?.name}
                </p>
              </div>
            )}

            {/* Информация */}
            <div className="mt-6 p-4 bg-blue-50 rounded-lg border border-blue-200">
              <div className="flex items-start">
                <Info className="w-5 h-5 text-blue-600 mt-0.5 mr-2 flex-shrink-0" />
                <div className="text-sm text-blue-800">
                  <p className="font-medium mb-1">Рекомендации:</p>
                  <ul className="list-disc list-inside space-y-1 text-blue-700">
                    <li>Используйте изображения с хорошим контрастом</li>
                    <li>Избегайте слишком мелких деталей</li>
                    <li>Лучше всего подходят портреты и пейзажи</li>
                  </ul>
                </div>
              </div>
            </div>
          </motion.div>
        </div>

        {/* Кнопка продолжения */}
        {selectedSize && previewUrl && (
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            className="text-center mt-12"
          >
            <button
              onClick={handleContinue}
              disabled={isUploading}
              className="bg-gradient-to-r from-purple-600 to-pink-600 text-white px-8 py-4 rounded-xl font-semibold text-lg hover:from-purple-700 hover:to-pink-700 transition-all duration-300 shadow-lg hover:shadow-xl disabled:opacity-50 disabled:cursor-not-allowed flex items-center mx-auto"
            >
              Продолжить
              <ArrowRight className="w-5 h-5 ml-2" />
            </button>
          </motion.div>
        )}
      </div>
    </div>
  )
}

export default DiamondMosaicPage
