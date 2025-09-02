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
  }, [previewUrl, rotation, scale, position, selectedSize])

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
      
      // Рисуем светлый фон
      ctx.fillStyle = '#f8fafc'
      ctx.fillRect(0, 0, canvasWidth, canvasHeight)
      
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
      
      // Определяем базовые размеры изображения с учетом поворота
      const maxImageSize = Math.min(canvasWidth, canvasHeight) * 0.8
      let imgAspectRatio = img.width / img.height
      
      // При поворотах на 90° и 270° меняем соотношение сторон
      if (rotation === 90 || rotation === 270) {
        imgAspectRatio = img.height / img.width
      }
      
      let baseImgWidth, baseImgHeight
      
      if (imgAspectRatio > 1) {
        // Горизонтальное изображение (с учетом поворота)
        baseImgWidth = maxImageSize
        baseImgHeight = maxImageSize / imgAspectRatio
      } else {
        // Вертикальное или квадратное изображение (с учетом поворота)
        baseImgHeight = maxImageSize
        baseImgWidth = maxImageSize * imgAspectRatio
      }
      
      // Область кадрирования всегда точно равна размерам изображения (с учетом поворота)
      let cropWidth = baseImgWidth
      let cropHeight = baseImgHeight
      
      // Применяем пользовательский масштаб к базовым размерам
      const imgWidth = baseImgWidth * scale
      const imgHeight = baseImgHeight * scale
      
      // Рисуем изображение
      ctx.drawImage(
        img,
        -imgWidth / 2,
        -imgHeight / 2,
        imgWidth,
        imgHeight
      )
      
      // Восстанавливаем контекст
      ctx.restore()
      
      const cropX = (canvasWidth - cropWidth) / 2
      const cropY = (canvasHeight - cropHeight) / 2
      
      // Затемняем области вне кадрирования
      ctx.fillStyle = 'rgba(0, 0, 0, 0.5)'
      
      // Верх
      ctx.fillRect(0, 0, canvasWidth, cropY)
      // Низ
      ctx.fillRect(0, cropY + cropHeight, canvasWidth, canvasHeight - cropY - cropHeight)
      // Лево
      ctx.fillRect(0, cropY, cropX, cropHeight)
      // Право
      ctx.fillRect(cropX + cropWidth, cropY, canvasWidth - cropX - cropWidth, cropHeight)
      
      // Рисуем рамку области кадрирования (более яркую и толстую)
      ctx.strokeStyle = '#8b5cf6'
      ctx.lineWidth = 3
      ctx.setLineDash([8, 4])
      ctx.strokeRect(cropX, cropY, cropWidth, cropHeight)
      
      // Рисуем сетку внутри области кадрирования (правило третей)
      ctx.strokeStyle = 'rgba(139, 92, 246, 0.4)'
      ctx.lineWidth = 1
      ctx.setLineDash([2, 2])
      
      // Вертикальные линии сетки
      const gridVertical1 = cropX + cropWidth / 3
      const gridVertical2 = cropX + (cropWidth * 2) / 3
      ctx.beginPath()
      ctx.moveTo(gridVertical1, cropY)
      ctx.lineTo(gridVertical1, cropY + cropHeight)
      ctx.stroke()
      
      ctx.beginPath()
      ctx.moveTo(gridVertical2, cropY)
      ctx.lineTo(gridVertical2, cropY + cropHeight)
      ctx.stroke()
      
      // Горизонтальные линии сетки
      const gridHorizontal1 = cropY + cropHeight / 3
      const gridHorizontal2 = cropY + (cropHeight * 2) / 3
      ctx.beginPath()
      ctx.moveTo(cropX, gridHorizontal1)
      ctx.lineTo(cropX + cropWidth, gridHorizontal1)
      ctx.stroke()
      
      ctx.beginPath()
      ctx.moveTo(cropX, gridHorizontal2)
      ctx.lineTo(cropX + cropWidth, gridHorizontal2)
      ctx.stroke()
      
      // Рисуем углы для лучшей видимости (больше и ярче)
      ctx.setLineDash([])
      ctx.lineWidth = 4
      ctx.strokeStyle = '#7c3aed'
      const cornerSize = 30
      
      // Верхний левый угол
      ctx.beginPath()
      ctx.moveTo(cropX, cropY + cornerSize)
      ctx.lineTo(cropX, cropY)
      ctx.lineTo(cropX + cornerSize, cropY)
      ctx.stroke()
      
      // Верхний правый угол
      ctx.beginPath()
      ctx.moveTo(cropX + cropWidth - cornerSize, cropY)
      ctx.lineTo(cropX + cropWidth, cropY)
      ctx.lineTo(cropX + cropWidth, cropY + cornerSize)
      ctx.stroke()
      
      // Нижний левый угол
      ctx.beginPath()
      ctx.moveTo(cropX, cropY + cropHeight - cornerSize)
      ctx.lineTo(cropX, cropY + cropHeight)
      ctx.lineTo(cropX + cornerSize, cropY + cropHeight)
      ctx.stroke()
      
      // Нижний правый угол
      ctx.beginPath()
      ctx.moveTo(cropX + cropWidth - cornerSize, cropY + cropHeight)
      ctx.lineTo(cropX + cropWidth, cropY + cropHeight)
      ctx.lineTo(cropX + cropWidth, cropY + cropHeight - cornerSize)
      ctx.stroke()
      
      // Добавляем иконку перетаскивания в углу области кадрирования
      ctx.fillStyle = 'rgba(139, 92, 246, 0.6)'
      ctx.font = 'bold 20px Arial'
      ctx.fillText('✋', cropX + cropWidth - 25, cropY + 25)
      
      // Генерируем URL отредактированного изображения (только область кадрирования)
      const tempCanvas = document.createElement('canvas')
      tempCanvas.width = cropWidth
      tempCanvas.height = cropHeight
      const tempCtx = tempCanvas.getContext('2d')
      
      tempCtx.drawImage(canvas, cropX, cropY, cropWidth, cropHeight, 0, 0, cropWidth, cropHeight)
      
      const editedDataUrl = tempCanvas.toDataURL('image/jpeg', 0.9)
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
      setScale(1.0) // Базовый масштаб - изображение точно соответствует области кадрирования
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
    setScale(1.0) // Сбрасываем к базовому масштабу
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
    setScale(1.0) // Базовый масштаб
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
              <div className="max-w-4xl mx-auto">
                <div className="bg-white rounded-2xl p-8 border border-gray-200 shadow-lg">
                  <div className="flex items-center justify-between mb-8">
                    <h4 className="text-xl font-semibold text-gray-800 flex items-center">
                      <Crop className="w-6 h-6 mr-3 text-purple-600" />
                      {t('diamond_mosaic_page.image_editor.setup_title')}
                    </h4>
                    <button
                      onClick={handleRemoveImage}
                      className="text-red-500 hover:text-red-600 flex items-center text-sm font-medium"
                    >
                      ✕ {t('diamond_mosaic_page.image_editor.delete_image')}
                    </button>
                  </div>
                  
                  {/* Превью с областью кадрирования */}
                  <div className="relative mb-8">
                    <div className="bg-gray-50 rounded-2xl p-6 flex justify-center">
                      <div className="relative inline-block">
                        {/* Canvas для редактирования */}
                        <canvas
                          ref={canvasRef}
                          width={800}
                          height={800}
                          className="rounded-xl border-2 border-dashed border-purple-300 cursor-move shadow-lg bg-white"
                          style={{ maxWidth: '100%', height: 'auto' }}
                          onMouseDown={handleMouseDown}
                          onMouseMove={handleMouseMove}
                          onMouseUp={handleMouseUp}
                          onMouseLeave={handleMouseUp}
                        />
                      </div>
                    </div>
                  </div>

                  {/* Простые инструменты */}
                  <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
                    
                    {/* Поворот */}
                    <div className="text-center">
                      <label className="block text-sm font-medium text-gray-700 mb-3">
                        🔄 {t('diamond_mosaic_page.image_editor.rotation_section')}
                      </label>
                      <button
                        onClick={handleRotate}
                        className="w-full bg-purple-100 text-purple-700 px-6 py-4 rounded-xl hover:bg-purple-200 transition-colors font-medium flex items-center justify-center text-lg"
                      >
                        <RotateCw className="w-5 h-5 mr-2" />
                        {t('diamond_mosaic_page.image_editor.tools.rotate_90')}
                      </button>
                    </div>

                    {/* Масштаб */}
                    <div className="text-center">
                      <label className="block text-sm font-medium text-gray-700 mb-3">
                        🔍 {t('diamond_mosaic_page.image_editor.scale_section')}: {Math.round(scale * 100)}%
                      </label>
                      <div className="flex space-x-3">
                        <button
                          onClick={() => handleScaleChange(scale - 0.1)}
                          className="flex-1 bg-gray-100 text-gray-700 px-4 py-4 rounded-xl hover:bg-gray-200 transition-colors flex items-center justify-center text-lg"
                        >
                          <ZoomOut className="w-5 h-5" />
                        </button>
                        <button
                          onClick={() => handleScaleChange(scale + 0.1)}
                          className="flex-1 bg-gray-100 text-gray-700 px-4 py-4 rounded-xl hover:bg-gray-200 transition-colors flex items-center justify-center text-lg"
                        >
                          <ZoomIn className="w-5 h-5" />
                        </button>
                      </div>
                    </div>

                    {/* Сброс */}
                    <div className="text-center">
                      <label className="block text-sm font-medium text-gray-700 mb-3">
                        ⚙️ {t('diamond_mosaic_page.image_editor.settings_section')}
                      </label>
                      <button
                        onClick={handleReset}
                        className="w-full bg-orange-100 text-orange-700 px-6 py-4 rounded-xl hover:bg-orange-200 transition-colors font-medium text-lg"
                      >
                        ↺ {t('diamond_mosaic_page.image_editor.tools.reset')}
                      </button>
                    </div>
                  </div>

                  {/* Информация о файле */}
                  <div className="text-center bg-blue-50 rounded-xl p-6 border border-blue-200">
                    <div className="flex items-center justify-center mb-2">
                      <div className="w-2 h-2 bg-green-500 rounded-full mr-2"></div>
                      <p className="font-medium text-gray-800">{selectedFile?.name}</p>
                    </div>
                    <p className="text-gray-600">{t('diamond_mosaic_page.image_editor.file_info')}</p>
                    <p className="text-sm text-blue-600 mt-2">💡 {t('diamond_mosaic_page.image_editor.crop_hint')}</p>
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
