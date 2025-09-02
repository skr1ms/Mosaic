import React, { useState, useEffect, useRef } from 'react'
import { useTranslation } from 'react-i18next'
import { motion } from 'framer-motion'
import { Upload, Ruler, ArrowRight, Info, Image as ImageIcon, RotateCw, Move, ZoomIn, ZoomOut, Crop, Check } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import { useUIStore } from '../store/partnerStore'

const DiamondMosaicPage = () => {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { addNotification } = useUIStore()
  
  const canvasRef = useRef(null)
  const fileInputRef = useRef(null)
  
  const [selectedSize, setSelectedSize] = useState('')
  const [selectedFile, setSelectedFile] = useState(null)
  const [previewUrl, setPreviewUrl] = useState(null)
  const [isUploading, setIsUploading] = useState(false)
  
  // Параметры редактирования
  const [rotation, setRotation] = useState(0)
  const [scale, setScale] = useState(1)
  const [position, setPosition] = useState({ x: 0, y: 0 })
  const [isDragging, setIsDragging] = useState(false)
  const [lastMousePos, setLastMousePos] = useState({ x: 0, y: 0 })
  const [editedImageUrl, setEditedImageUrl] = useState(null)

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

  // Отрисовка канваса при изменении параметров
  useEffect(() => {
    if (previewUrl && canvasRef.current) {
      drawImageOnCanvas()
    }
  }, [previewUrl, rotation, scale, position])

  const drawImageOnCanvas = () => {
    const canvas = canvasRef.current
    if (!canvas || !previewUrl) return

    const ctx = canvas.getContext('2d')
    const img = new Image()
    
    img.onload = () => {
      const canvasWidth = canvas.width
      const canvasHeight = canvas.height
      
      // Очищаем канвас
      ctx.clearRect(0, 0, canvasWidth, canvasHeight)
      
      // Сохраняем контекст
      ctx.save()
      
      // Центрируем трансформации
      ctx.translate(canvasWidth / 2, canvasHeight / 2)
      
      // Применяем поворот
      ctx.rotate((rotation * Math.PI) / 180)
      
      // Применяем масштаб
      ctx.scale(scale, scale)
      
      // Применяем позицию
      ctx.translate(position.x, position.y)
      
      // Рисуем изображение
      const imgWidth = Math.min(img.width, 400)
      const imgHeight = Math.min(img.height, 400)
      
      ctx.drawImage(
        img,
        -imgWidth / 2,
        -imgHeight / 2,
        imgWidth,
        imgHeight
      )
      
      // Восстанавливаем контекст
      ctx.restore()
      
      // Генерируем URL отредактированного изображения
      const editedDataUrl = canvas.toDataURL('image/jpeg', 0.9)
      setEditedImageUrl(editedDataUrl)
    }
    
    img.src = previewUrl
  }

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

    // Проверка типа файла
    if (!file.type.startsWith('image/')) {
      addNotification({
        type: 'error',
        message: t('diamond_mosaic_page.upload_section.file_type_error')
      })
      return
    }

    // Проверка размера файла (максимум 10MB)
    if (file.size > 10 * 1024 * 1024) {
      addNotification({
        type: 'error',
        message: t('diamond_mosaic_page.upload_section.file_size_error')
      })
      return
    }

    setSelectedFile(file)
    
    // Создаем превью
    const reader = new FileReader()
    reader.onload = (e) => {
      setPreviewUrl(e.target.result)
      // Сбрасываем параметры редактирования при загрузке нового изображения
      setRotation(0)
      setScale(1)
      setPosition({ x: 0, y: 0 })
    }
    reader.readAsDataURL(file)
  }

  // Функции редактирования
  const handleRotate = () => {
    setRotation(prev => (prev + 90) % 360)
  }

  const handleScaleChange = (newScale) => {
    setScale(Math.max(0.1, Math.min(3, newScale)))
  }

  const handleMouseDown = (e) => {
    setIsDragging(true)
    setLastMousePos({ x: e.clientX, y: e.clientY })
  }

  const handleMouseMove = (e) => {
    if (!isDragging) return

    const deltaX = e.clientX - lastMousePos.x
    const deltaY = e.clientY - lastMousePos.y

    setPosition(prev => ({
      x: prev.x + deltaX * 0.5,
      y: prev.y + deltaY * 0.5
    }))

    setLastMousePos({ x: e.clientX, y: e.clientY })
  }

  const handleMouseUp = () => {
    setIsDragging(false)
  }

  const handleReset = () => {
    setRotation(0)
    setScale(1)
    setPosition({ x: 0, y: 0 })
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

    // Сохраняем данные для редактора
    const finalImageUrl = editedImageUrl || previewUrl

    try {
      // Сохраняем данные изображения
      localStorage.setItem('diamondMosaic_selectedImage', JSON.stringify({
        size: selectedSize,
        fileName: selectedFile.name,
        previewUrl: finalImageUrl,
        timestamp: Date.now(),
        hasEdits: editedImageUrl !== null
      }))
      
      // Сохраняем настройки для следующего шага
      localStorage.setItem('diamondMosaic_editorSettings', JSON.stringify({
        size: selectedSize,
        style: null, // Будет выбран позже
        returnTo: '/diamond-mosaic/styles'
      }))
      
      // Создаем временный URL для файла
      if (editedImageUrl) {
        // Конвертируем canvas в blob
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

    // Переходим к выбору стилей
    navigate('/diamond-mosaic/styles')
  }

  const handleRemoveImage = () => {
    setSelectedFile(null)
    setPreviewUrl(null)
    setEditedImageUrl(null)
    
    // Сбрасываем параметры редактирования
    setRotation(0)
    setScale(1)
    setPosition({ x: 0, y: 0 })
    
    // Очищаем input
    if (fileInputRef.current) {
      fileInputRef.current.value = ''
    }
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-purple-50 via-pink-50 to-blue-50">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
        
        {/* Заголовок */}
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

        {/* Выбор размера - сетка 3x2 как в магазине */}
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
              
              // Визуальные размеры прямоугольников для разных размеров
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

        {/* Загрузка и редактирование изображения */}
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
              <div className="grid lg:grid-cols-2 gap-8">
                {/* Левая колонка: превью изображения */}
                <div className="space-y-4">
                  <div className="bg-white rounded-2xl p-6 border border-gray-200 shadow-lg">
                    <h4 className="font-semibold text-gray-800 mb-4 text-center">{t('diamond_mosaic_page.upload_section.image_preview')}</h4>
                    
                    <div className="relative">
                      {/* Оригинальное изображение (скрытое) */}
                      <img
                        src={previewUrl}
                        alt="Original"
                        className="hidden"
                      />
                      
                      {/* Canvas для редактирования */}
                      <div className="canvas-container bg-gray-50 rounded-xl overflow-hidden">
                        <canvas
                          ref={canvasRef}
                          width={400}
                          height={400}
                          className="w-full h-auto cursor-move border border-gray-200"
                          onMouseDown={handleMouseDown}
                          onMouseMove={handleMouseMove}
                          onMouseUp={handleMouseUp}
                          onMouseLeave={handleMouseUp}
                        />
                      </div>
                      
                      <button
                        onClick={handleRemoveImage}
                        className="absolute top-2 right-2 bg-red-500 text-white w-8 h-8 rounded-full flex items-center justify-center hover:bg-red-600 transition-colors"
                      >
                        ×
                      </button>
                    </div>
                    
                    <p className="text-sm text-gray-600 mt-3 text-center">
                      {selectedFile?.name}
                    </p>
                    
                    <p className="text-xs text-gray-500 mt-2 text-center">
                      {t('diamond_mosaic_page.image_editor.drag_instruction')}
                    </p>
                  </div>
                </div>

                {/* Правая колонка: инструменты редактирования */}
                <div className="space-y-4">
                  <div className="bg-white rounded-2xl p-6 border border-gray-200 shadow-lg">
                    <h4 className="font-semibold text-gray-800 mb-6 flex items-center">
                      <Crop className="w-5 h-5 mr-2 text-purple-600" />
                      {t('diamond_mosaic_page.image_editor.title')}
                    </h4>
                    
                    <div className="space-y-6">
                      {/* Поворот */}
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-3">
                          <RotateCw className="w-4 h-4 inline mr-1" />
                          {t('diamond_mosaic_page.image_editor.rotation_control')}: {rotation}°
                        </label>
                        <div className="flex space-x-3">
                          <button
                            onClick={handleRotate}
                            className="bg-purple-100 text-purple-700 px-4 py-2 rounded-lg hover:bg-purple-200 transition-colors font-medium"
                          >
                            {t('diamond_mosaic_page.image_editor.tools.rotate_90')}
                          </button>
                          <input
                            type="range"
                            min="0"
                            max="360"
                            value={rotation}
                            onChange={(e) => setRotation(Number(e.target.value))}
                            className="flex-1"
                          />
                        </div>
                      </div>

                      {/* Масштаб */}
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-3">
                          <ZoomIn className="w-4 h-4 inline mr-1" />
                          {t('diamond_mosaic_page.image_editor.scale_control')}: {Math.round(scale * 100)}%
                        </label>
                        <div className="flex space-x-3 items-center">
                          <button
                            onClick={() => handleScaleChange(scale - 0.1)}
                            className="bg-gray-100 text-gray-700 p-2 rounded-lg hover:bg-gray-200 transition-colors"
                          >
                            <ZoomOut className="w-4 h-4" />
                          </button>
                          <input
                            type="range"
                            min="0.1"
                            max="3"
                            step="0.1"
                            value={scale}
                            onChange={(e) => setScale(Number(e.target.value))}
                            className="flex-1"
                          />
                          <button
                            onClick={() => handleScaleChange(scale + 0.1)}
                            className="bg-gray-100 text-gray-700 p-2 rounded-lg hover:bg-gray-200 transition-colors"
                          >
                            <ZoomIn className="w-4 h-4" />
                          </button>
                        </div>
                      </div>

                      {/* Позиция */}
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-3">
                          <Move className="w-4 h-4 inline mr-1" />
                          {t('diamond_mosaic_page.image_editor.controls.position_x')}: {Math.round(position.x)}
                        </label>
                        <input
                          type="range"
                          min="-200"
                          max="200"
                          value={position.x}
                          onChange={(e) => setPosition(prev => ({ ...prev, x: Number(e.target.value) }))}
                          className="w-full"
                        />
                      </div>

                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-3">
                          {t('diamond_mosaic_page.image_editor.controls.position_y')}: {Math.round(position.y)}
                        </label>
                        <input
                          type="range"
                          min="-200"
                          max="200"
                          value={position.y}
                          onChange={(e) => setPosition(prev => ({ ...prev, y: Number(e.target.value) }))}
                          className="w-full"
                        />
                      </div>
                    </div>

                    {/* Кнопки действий */}
                    <div className="flex space-x-3 mt-6">
                      <button
                        onClick={handleReset}
                        className="flex-1 bg-gray-100 text-gray-700 px-4 py-2 rounded-lg hover:bg-gray-200 transition-colors font-medium"
                      >
                        {t('diamond_mosaic_page.image_editor.tools.reset')}
                      </button>
                      <button
                        onClick={() => fileInputRef.current?.click()}
                        className="flex-1 bg-blue-100 text-blue-700 px-4 py-2 rounded-lg hover:bg-blue-200 transition-colors font-medium"
                      >
                        {t('diamond_mosaic_page.navigation.other_photo')}
                      </button>
                    </div>
                  </div>

                  {/* Информация */}
                  <div className="bg-blue-50 rounded-xl p-4 border border-blue-200">
                    <div className="flex items-start">
                      <Info className="w-5 h-5 text-blue-600 mt-0.5 mr-2 flex-shrink-0" />
                      <div className="text-sm text-blue-800">
                        <p className="font-medium mb-2">{t('diamond_mosaic_page.navigation.recommendations.title')}</p>
                        <ul className="list-disc list-inside space-y-1 text-blue-700">
                          <li>{t('diamond_mosaic_page.navigation.recommendations.good_contrast')}</li>
                          <li>{t('diamond_mosaic_page.navigation.recommendations.avoid_small_details')}</li>
                          <li>{t('diamond_mosaic_page.navigation.recommendations.best_for')}</li>
                        </ul>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            )}
          </div>
        </motion.section>

        {/* Кнопка продолжения */}
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
            <p className="text-gray-600 mt-3">
              {t('diamond_mosaic_page.size_selection.sizes.' + selectedSize)}: <strong>{getSizes().find(s => s.key === selectedSize)?.title}</strong>
            </p>
          </motion.div>
        )}
      </div>
    </div>
  )
}

export default DiamondMosaicPage
