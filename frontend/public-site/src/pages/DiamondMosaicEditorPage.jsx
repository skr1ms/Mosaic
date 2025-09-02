import React, { useState, useEffect, useRef } from 'react'
import { useTranslation } from 'react-i18next'
import { motion } from 'framer-motion'
import { ArrowLeft, ArrowRight, RotateCw, Move, ZoomIn, ZoomOut, Crop, Upload } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import { useUIStore } from '../store/partnerStore'

const DiamondMosaicEditorPage = () => {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { addNotification } = useUIStore()
  
  const canvasRef = useRef(null)
  const fileInputRef = useRef(null)
  
  const [editorSettings, setEditorSettings] = useState(null)
  const [currentImage, setCurrentImage] = useState(null)
  const [editedImageUrl, setEditedImageUrl] = useState(null)
  
  // Editing parameters
  const [rotation, setRotation] = useState(0)
  const [scale, setScale] = useState(1)
  const [position, setPosition] = useState({ x: 0, y: 0 })
  const [crop, setCrop] = useState({ x: 0, y: 0, width: 100, height: 100 })
  const [isDragging, setIsDragging] = useState(false)
  const [lastMousePos, setLastMousePos] = useState({ x: 0, y: 0 })

  useEffect(() => {
    // Load editor settings
    try {
      const savedSettings = localStorage.getItem('diamondMosaic_editorSettings')
      const savedImageData = localStorage.getItem('diamondMosaic_selectedImage')
      
      if (!savedSettings || !savedImageData) {
        navigate('/diamond-mosaic')
        return
      }
      
      const parsedSettings = JSON.parse(savedSettings)
      const parsedImageData = JSON.parse(savedImageData)
      
      setEditorSettings(parsedSettings)
      setCurrentImage(parsedImageData.previewUrl)
      
    } catch (error) {
      console.error('Error loading editor settings:', error)
      navigate('/diamond-mosaic')
    }
  }, [navigate])

  useEffect(() => {
    if (currentImage && canvasRef.current) {
      drawImage()
    }
  }, [currentImage, rotation, scale, position, crop])

  const drawImage = () => {
    const canvas = canvasRef.current
    if (!canvas || !currentImage) return

    const ctx = canvas.getContext('2d')
    const img = new Image()
    
    img.onload = () => {
      const canvasWidth = canvas.width
      const canvasHeight = canvas.height
      
      // Clear canvas
      ctx.clearRect(0, 0, canvasWidth, canvasHeight)
      
      // Save context
      ctx.save()
      
      // Center transformations
      ctx.translate(canvasWidth / 2, canvasHeight / 2)
      
      // Применяем поворот
      ctx.rotate((rotation * Math.PI) / 180)
      
      // Применяем масштаб
      ctx.scale(scale, scale)
      
      // Применяем позицию
      ctx.translate(position.x, position.y)
      
      // Рисуем изображение
      const imgWidth = img.width
      const imgHeight = img.height
      
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
    
    img.src = currentImage
  }

  const handleNewImageUpload = (event) => {
    const file = event.target.files[0]
    if (!file) return

    if (!file.type.startsWith('image/')) {
      addNotification({
        type: 'error',
        message: t('diamond_mosaic_editor.image_required')
      })
      return
    }

    const reader = new FileReader()
    reader.onload = (e) => {
      setCurrentImage(e.target.result)
      // Сброс параметров редактирования
      setRotation(0)
      setScale(1)
      setPosition({ x: 0, y: 0 })
      setCrop({ x: 0, y: 0, width: 100, height: 100 })
    }
    reader.readAsDataURL(file)
  }

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
    setCrop({ x: 0, y: 0, width: 100, height: 100 })
  }

  const handleContinue = () => {
    if (!editedImageUrl || !editorSettings) return

    try {
      // Конвертируем data URL в файл
      fetch(editedImageUrl)
        .then(res => res.blob())
        .then(blob => {
          const file = new File([blob], 'edited-image.jpg', { type: 'image/jpeg' })
          const fileUrl = URL.createObjectURL(file)
          
          // Обновляем данные изображения
          const updatedImageData = {
            size: editorSettings.size,
            selectedStyle: editorSettings.style,
            fileName: 'edited-image.jpg',
            previewUrl: editedImageUrl,
            timestamp: Date.now()
          }
          
          localStorage.setItem('diamondMosaic_selectedImage', JSON.stringify(updatedImageData))
          sessionStorage.setItem('diamondMosaic_fileUrl', fileUrl)
          
          // Очищаем настройки редактора
          localStorage.removeItem('diamondMosaic_editorSettings')
          
          // Возвращаемся к альбому превью или переходим к следующему шагу
          if (editorSettings.returnTo) {
            navigate(editorSettings.returnTo)
          } else {
            navigate('/diamond-mosaic/preview-album')
          }
        })
    } catch (error) {
      console.error('Error processing edited image:', error)
      addNotification({
        type: 'error',
        message: t('diamond_mosaic_editor.processing_error')
      })
    }
  }

  const handleBack = () => {
    if (editorSettings?.returnTo) {
      navigate(editorSettings.returnTo)
    } else {
      navigate('/diamond-mosaic/preview-album')
    }
  }

  if (!editorSettings) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-purple-50 via-pink-50 to-blue-50 flex items-center justify-center">
      <div className="text-purple-600">{t('diamond_mosaic_editor.loading')}</div>
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
            {t('diamond_mosaic_editor.navigation.back')}
          </button>
          
          <div className="text-center">
            <h1 className="text-4xl md:text-5xl font-bold bg-gradient-to-r from-purple-600 to-pink-600 bg-clip-text text-transparent mb-4">
              {t('diamond_mosaic_editor.title')}
            </h1>
            <p className="text-lg text-gray-600">
              {t('diamond_mosaic_editor.size_label')} <span className="font-semibold">{editorSettings.size} {t('common.cm')}</span> • 
              {t('diamond_mosaic_editor.style_label')} <span className="font-semibold">{editorSettings.style}</span>
            </p>
          </div>
        </motion.div>

        <div className="grid lg:grid-cols-3 gap-8">
          
          {/* Панель инструментов */}
          <motion.div
            initial={{ opacity: 0, x: -20 }}
            animate={{ opacity: 1, x: 0 }}
            className="lg:col-span-1 space-y-6"
          >
            
            {/* Загрузка нового изображения */}
            <div className="bg-white rounded-xl p-6 shadow-lg">
              <h3 className="text-lg font-semibold text-gray-800 mb-4 flex items-center">
                <Upload className="w-5 h-5 mr-2 text-purple-600" />
                {t('diamond_mosaic_editor.sections.new_image')}
              </h3>
              <input
                ref={fileInputRef}
                type="file"
                accept="image/*"
                onChange={handleNewImageUpload}
                className="hidden"
              />
              <button
                onClick={() => fileInputRef.current?.click()}
                className="w-full bg-purple-100 text-purple-700 px-4 py-2 rounded-lg hover:bg-purple-200 transition-colors"
              >
                {t('diamond_mosaic_editor.sections.select_other_image')}
              </button>
            </div>
            
            {/* Поворот */}
            <div className="bg-white rounded-xl p-6 shadow-lg">
              <h3 className="text-lg font-semibold text-gray-800 mb-4 flex items-center">
                <RotateCw className="w-5 h-5 mr-2 text-purple-600" />
                {t('diamond_mosaic_editor.tools.rotate')}
              </h3>
              <div className="space-y-3">
                <button
                  onClick={handleRotate}
                  className="w-full bg-purple-600 text-white px-4 py-2 rounded-lg hover:bg-purple-700 transition-colors"
                >
                  {t('diamond_mosaic_editor.tools.rotate')} 90°
                </button>
                <div>
                  <label className="block text-sm text-gray-600 mb-2">
                    {t('diamond_mosaic_editor.controls.rotation_control', { value: rotation })}
                  </label>
                  <input
                    type="range"
                    min="0"
                    max="360"
                    value={rotation}
                    onChange={(e) => setRotation(Number(e.target.value))}
                    className="w-full"
                  />
                </div>
              </div>
            </div>
            
            {/* Масштаб */}
            <div className="bg-white rounded-xl p-6 shadow-lg">
              <h3 className="text-lg font-semibold text-gray-800 mb-4 flex items-center">
                <ZoomIn className="w-5 h-5 mr-2 text-purple-600" />
                {t('diamond_mosaic_editor.tools.scale')}
              </h3>
              <div className="space-y-3">
                <div className="flex space-x-2">
                  <button
                    onClick={() => handleScaleChange(scale - 0.1)}
                    className="flex-1 bg-gray-100 text-gray-700 px-3 py-2 rounded-lg hover:bg-gray-200 transition-colors"
                  >
                    <ZoomOut className="w-4 h-4 mx-auto" />
                  </button>
                  <button
                    onClick={() => handleScaleChange(scale + 0.1)}
                    className="flex-1 bg-gray-100 text-gray-700 px-3 py-2 rounded-lg hover:bg-gray-200 transition-colors"
                  >
                    <ZoomIn className="w-4 h-4 mx-auto" />
                  </button>
                </div>
                <div>
                  <label className="block text-sm text-gray-600 mb-2">
                    Масштаб: {Math.round(scale * 100)}%
                  </label>
                  <input
                    type="range"
                    min="0.1"
                    max="3"
                    step="0.1"
                    value={scale}
                    onChange={(e) => setScale(Number(e.target.value))}
                    className="w-full"
                  />
                </div>
              </div>
            </div>
            
            {/* Позиция */}
            <div className="bg-white rounded-xl p-6 shadow-lg">
              <h3 className="text-lg font-semibold text-gray-800 mb-4 flex items-center">
                <Move className="w-5 h-5 mr-2 text-purple-600" />
                {t('diamond_mosaic_editor.tools.position')}
              </h3>
              <div className="space-y-3">
                <div>
                  <label className="block text-sm text-gray-600 mb-2">
                    {t('diamond_mosaic_editor.controls.position_x', { value: Math.round(position.x) })}
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
                  <label className="block text-sm text-gray-600 mb-2">
                    {t('diamond_mosaic_editor.controls.position_y', { value: Math.round(position.y) })}
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
            </div>
            
            {/* Сброс */}
            <div className="bg-white rounded-xl p-6 shadow-lg">
              <button
                onClick={handleReset}
                className="w-full bg-gray-600 text-white px-4 py-2 rounded-lg hover:bg-gray-700 transition-colors"
              >
                {t('diamond_mosaic_editor.sections.reset_changes')}
              </button>
            </div>
          </motion.div>

          {/* Канвас редактора */}
          <motion.div
            initial={{ opacity: 0, x: 20 }}
            animate={{ opacity: 1, x: 0 }}
            className="lg:col-span-2"
          >
            <div className="bg-white rounded-xl p-6 shadow-lg">
              <h3 className="text-lg font-semibold text-gray-800 mb-4">{t('diamond_mosaic_editor.sections.preview')}</h3>
              
              <div className="border-2 border-gray-200 rounded-lg overflow-hidden bg-gray-50">
                <canvas
                  ref={canvasRef}
                  width={600}
                  height={600}
                  className="w-full h-auto cursor-move"
                  onMouseDown={handleMouseDown}
                  onMouseMove={handleMouseMove}
                  onMouseUp={handleMouseUp}
                  onMouseLeave={handleMouseUp}
                />
              </div>
              
              <div className="mt-6 text-center">
                <p className="text-sm text-gray-600 mb-4">
                  {t('diamond_mosaic_editor.sections.drag_instruction')}
                </p>
                
                <button
                  onClick={handleContinue}
                  disabled={!editedImageUrl}
                  className="bg-gradient-to-r from-purple-600 to-pink-600 text-white px-8 py-4 rounded-xl font-semibold text-lg hover:from-purple-700 hover:to-pink-700 transition-all duration-300 shadow-lg hover:shadow-xl disabled:opacity-50 disabled:cursor-not-allowed flex items-center mx-auto"
                >
                  {t('diamond_mosaic_editor.continue')}
                  <ArrowRight className="w-5 h-5 ml-2" />
                </button>
              </div>
            </div>
          </motion.div>
        </div>
      </div>
    </div>
  )
}

export default DiamondMosaicEditorPage
