import React, { useState, useEffect, useRef, useCallback } from 'react'
import { useTranslation } from 'react-i18next'
import { ArrowLeft, Minus, Plus, RotateCcw, Share2, Upload } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import { useUIStore } from '../store/partnerStore'

const DiamondMosaicEditorPage = () => {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { addNotification } = useUIStore()
  
  const canvasRef = useRef(null)
  const fileInputRef = useRef(null)
  const containerRef = useRef(null)
  
  const [editorSettings, setEditorSettings] = useState(null)
  const [currentImage, setCurrentImage] = useState(null)
  const [editedImageUrl, setEditedImageUrl] = useState(null)
  
  // Editing parameters
  const [scale, setScale] = useState(1)
  const [position, setPosition] = useState({ x: 0, y: 0 })
  const [isDragging, setIsDragging] = useState(false)
  const [dragStart, setDragStart] = useState({ x: 0, y: 0 })
  const [positionStart, setPositionStart] = useState({ x: 0, y: 0 })
  
  // Crop area state (fixed center crop)
  const [cropArea] = useState({ x: 25, y: 25, width: 50, height: 50 })

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
      drawImageOnCanvas()
    }
  }, [currentImage, scale, position])

  const drawImageOnCanvas = () => {
    const canvas = canvasRef.current
    const container = containerRef.current
    if (!canvas || !currentImage || !container) return

    const ctx = canvas.getContext('2d')
    const img = new Image()
    
    img.onload = () => {
      const containerRect = container.getBoundingClientRect()
      canvas.width = containerRect.width
      canvas.height = containerRect.height
      
      // Clear canvas
      ctx.clearRect(0, 0, canvas.width, canvas.height)
      
      // Calculate image dimensions
      const aspectRatio = img.width / img.height
      let drawWidth, drawHeight
      
      if (aspectRatio > 1) {
        drawWidth = Math.min(canvas.width * scale, img.width * scale)
        drawHeight = drawWidth / aspectRatio
      } else {
        drawHeight = Math.min(canvas.height * scale, img.height * scale)
        drawWidth = drawHeight * aspectRatio
      }
      
      // Center the image with offset
      const centerX = canvas.width / 2 + position.x
      const centerY = canvas.height / 2 + position.y
      const drawX = centerX - drawWidth / 2
      const drawY = centerY - drawHeight / 2
      
      // Draw image
      ctx.drawImage(img, drawX, drawY, drawWidth, drawHeight)
      
      // Generate edited image URL
      const editedDataUrl = canvas.toDataURL('image/jpeg', 0.95)
      setEditedImageUrl(editedDataUrl)
    }
    
    img.src = currentImage
  }

  // Mouse handlers for dragging
  const handleMouseDown = useCallback((e) => {
    e.preventDefault()
    setIsDragging(true)
    setDragStart({ x: e.clientX, y: e.clientY })
    setPositionStart(position)
  }, [position])

  const handleMouseMove = useCallback((e) => {
    if (isDragging) {
      const deltaX = e.clientX - dragStart.x
      const deltaY = e.clientY - dragStart.y
      
      setPosition({
        x: positionStart.x + deltaX,
        y: positionStart.y + deltaY
      })
    }
  }, [isDragging, dragStart, positionStart])

  const handleMouseUp = useCallback(() => {
    setIsDragging(false)
  }, [])

  useEffect(() => {
    if (isDragging) {
      document.addEventListener('mousemove', handleMouseMove)
      document.addEventListener('mouseup', handleMouseUp)
      return () => {
        document.removeEventListener('mousemove', handleMouseMove)
        document.removeEventListener('mouseup', handleMouseUp)
      }
    }
  }, [isDragging, handleMouseMove, handleMouseUp])

  // Scale handlers
  const handleZoomIn = () => {
    setScale(prev => Math.min(prev + 0.2, 3))
  }

  const handleZoomOut = () => {
    setScale(prev => Math.max(prev - 0.2, 0.2))
  }

  const handleReset = () => {
    setScale(1)
    setPosition({ x: 0, y: 0 })
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
      setScale(1)
      setPosition({ x: 0, y: 0 })
    }
    reader.readAsDataURL(file)
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
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-purple-600">{t('diamond_mosaic_editor.loading')}</div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="max-w-6xl mx-auto p-6">
        
        {/* Заголовок */}
        <div className="mb-6">
          <button
            onClick={handleBack}
            className="flex items-center text-purple-600 hover:text-purple-700 mb-4 transition-colors"
          >
            <ArrowLeft className="w-5 h-5 mr-2" />
            {t('diamond_mosaic_editor.navigation.back')}
          </button>
          
          <h1 className="text-3xl font-bold text-gray-900 text-center">
            {t('diamond_mosaic_editor.title')}
          </h1>
          <p className="text-gray-600 text-center mt-2">
            {t('diamond_mosaic_editor.size_label')} <span className="font-semibold">{editorSettings.size} {t('common.cm')}</span> • 
            {t('diamond_mosaic_editor.style_label')} <span className="font-semibold">{editorSettings.style}</span>
          </p>
        </div>

        {/* Главный редактор */}
        <div className="bg-white rounded-2xl shadow-lg overflow-hidden">
          
          {/* Область изображения */}
          <div 
            ref={containerRef}
            className="relative bg-gray-100 h-96 md:h-[500px] overflow-hidden"
          >
            <canvas
              ref={canvasRef}
              className="w-full h-full cursor-move select-none"
              onMouseDown={handleMouseDown}
              style={{ touchAction: 'none' }}
            />
            
            {/* Сетка кадрирования */}
            <div 
              className="absolute border-2 border-white shadow-lg pointer-events-none"
              style={{
                left: `${cropArea.x}%`,
                top: `${cropArea.y}%`,
                width: `${cropArea.width}%`,
                height: `${cropArea.height}%`,
              }}
            >
              {/* Сетка 3x3 */}
              <svg className="w-full h-full" viewBox="0 0 100 100" preserveAspectRatio="none">
                <line x1="33.33" y1="0" x2="33.33" y2="100" stroke="white" strokeWidth="0.5" opacity="0.7" />
                <line x1="66.66" y1="0" x2="66.66" y2="100" stroke="white" strokeWidth="0.5" opacity="0.7" />
                <line x1="0" y1="33.33" x2="100" y2="33.33" stroke="white" strokeWidth="0.5" opacity="0.7" />
                <line x1="0" y1="66.66" x2="100" y2="66.66" stroke="white" strokeWidth="0.5" opacity="0.7" />
              </svg>
            </div>
          </div>

          {/* Панель управления */}
          <div className="p-6 bg-white border-t border-gray-200">
            <div className="flex flex-wrap items-center justify-center gap-4">
              
              {/* Загрузка нового изображения */}
              <label className="flex items-center justify-center w-12 h-12 bg-purple-100 hover:bg-purple-200 rounded-xl cursor-pointer transition-colors">
                <Upload className="w-5 h-5 text-purple-600" />
                <input
                  ref={fileInputRef}
                  type="file"
                  accept="image/*"
                  onChange={handleNewImageUpload}
                  className="hidden"
                />
              </label>

              {/* Уменьшить масштаб */}
              <button
                onClick={handleZoomOut}
                className="flex items-center justify-center w-12 h-12 bg-gray-100 hover:bg-gray-200 rounded-xl transition-colors"
              >
                <Minus className="w-5 h-5 text-gray-600" />
              </button>

              {/* Увеличить масштаб */}
              <button
                onClick={handleZoomIn}
                className="flex items-center justify-center w-12 h-12 bg-gray-100 hover:bg-gray-200 rounded-xl transition-colors"
              >
                <Plus className="w-5 h-5 text-gray-600" />
              </button>

              {/* Сброс */}
              <button
                onClick={handleReset}
                className="flex items-center justify-center w-12 h-12 bg-gray-100 hover:bg-gray-200 rounded-xl transition-colors"
              >
                <RotateCcw className="w-5 h-5 text-gray-600" />
              </button>

              {/* Поделиться (будущий функционал) */}
              <button className="flex items-center justify-center w-12 h-12 bg-gray-100 hover:bg-gray-200 rounded-xl transition-colors opacity-50 cursor-not-allowed">
                <Share2 className="w-5 h-5 text-gray-600" />
              </button>
            </div>
          </div>

          {/* Кнопки действий */}
          <div className="p-6 bg-gray-50 border-t border-gray-200">
            <div className="flex gap-4 max-w-md mx-auto">
              <button
                onClick={handleBack}
                className="flex-1 px-6 py-3 bg-white border border-gray-300 text-gray-700 rounded-xl font-medium hover:bg-gray-50 transition-colors"
              >
                {t('diamond_mosaic_editor.navigation.back')}
              </button>
              
              <button
                onClick={handleContinue}
                disabled={!editedImageUrl}
                className="flex-1 px-6 py-3 bg-gradient-to-r from-purple-600 to-pink-600 text-white rounded-xl font-medium hover:from-purple-700 hover:to-pink-700 transition-all disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {t('diamond_mosaic_editor.continue')} →
              </button>
            </div>
          </div>

        </div>
      </div>
    </div>
  )
}

export default DiamondMosaicEditorPageNew
