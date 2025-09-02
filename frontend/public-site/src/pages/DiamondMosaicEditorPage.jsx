import React, { useState, useEffect, useRef, useCallback } from 'react'
import { useTranslation } from 'react-i18next'
import { ArrowLeft, Download, Minus, Plus, RotateCw, Undo2 } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import { useUIStore } from '../store/partnerStore'

const DiamondMosaicEditorPage = () => {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { addNotification } = useUIStore()
  
  const canvasRef = useRef(null)
  const imageRef = useRef(null)
  const containerRef = useRef(null)
  
  const [editorSettings, setEditorSettings] = useState(null)
  const [currentImage, setCurrentImage] = useState(null)
  const [editedImageUrl, setEditedImageUrl] = useState(null)
  
  // Image editing state
  const [scale, setScale] = useState(1.5) // Увеличиваем начальный масштаб
  const [rotation, setRotation] = useState(0)
  const [position, setPosition] = useState({ x: 0, y: 0 })
  const [isDragging, setIsDragging] = useState(false)
  const [dragStart, setDragStart] = useState({ x: 0, y: 0 })
  const [positionStart, setPositionStart] = useState({ x: 0, y: 0 })
  const [imageSize, setImageSize] = useState({ width: 0, height: 0 })
  
  // Crop area state - изначально весь экран
  const [cropArea, setCropArea] = useState({
    x: 0,
    y: 0, 
    width: 100,
    height: 100
  })
  
  // Crop resizing state
  const [isResizing, setIsResizing] = useState(false)
  const [resizeHandle, setResizeHandle] = useState(null)
  const [resizeStart, setResizeStart] = useState({ x: 0, y: 0 })
  const [cropStart, setCropStart] = useState(null)
  
  // Fixed crop area - изначально весь контейнер
  
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
      
      // Загружаем изображение
      const img = new Image()
      img.onload = () => {
        setImageSize({ width: img.width, height: img.height })
        setCurrentImage(parsedImageData.previewUrl)
      }
      img.src = parsedImageData.previewUrl
      
    } catch (error) {
      console.error('Error loading editor settings:', error)
      navigate('/diamond-mosaic')
    }
  }, [navigate])

  // Рендер изображения на канвасе
  useEffect(() => {
    if (currentImage && canvasRef.current && containerRef.current) {
      renderImage()
    }
  }, [currentImage, scale, position, rotation])

  const renderImage = () => {
    const canvas = canvasRef.current
    const container = containerRef.current
    if (!canvas || !currentImage || !container) return

    const ctx = canvas.getContext('2d')
    
    // Устанавливаем размер канваса равным размеру контейнера
    canvas.width = container.offsetWidth
    canvas.height = container.offsetHeight
    
    // Очищаем канвас
    ctx.clearRect(0, 0, canvas.width, canvas.height)
    
    const img = new Image()
    img.onload = () => {
      // Сохраняем контекст для трансформаций
      ctx.save()
      
      // Перемещаем точку отсчета в центр канваса
      ctx.translate(canvas.width / 2, canvas.height / 2)
      
      // Применяем поворот
      ctx.rotate((rotation * Math.PI) / 180)
      
      // Применяем масштаб
      ctx.scale(scale, scale)
      
      // Применяем позицию (смещение)
      ctx.translate(position.x, position.y)
      
      // Вычисляем размеры для отображения
      const containerAspect = canvas.width / canvas.height
      const imageAspect = img.width / img.height
      
      let renderWidth, renderHeight
      
      // Масштабируем изображение чтобы полностью заполнить контейнер
      if (imageAspect > containerAspect) {
        renderWidth = canvas.width * 1.0 // Заполняем контейнер полностью
        renderHeight = renderWidth / imageAspect
      } else {
        renderHeight = canvas.height * 1.0 // Заполняем контейнер полностью
        renderWidth = renderHeight * imageAspect
      }
      
      // Рисуем изображение относительно центра
      ctx.drawImage(img, -renderWidth / 2, -renderHeight / 2, renderWidth, renderHeight)
      
      // Восстанавливаем контекст
      ctx.restore()
      
      // Создаем обрезанную версию для области кадрирования (20% отступ = 60% область)
      createCroppedImage(canvas, ctx)
    }
    
    img.src = currentImage
  }

  const createCroppedImage = (canvas, ctx) => {
    // Размеры области кадрирования в пикселях
    const cropX = canvas.width * (cropArea.x / 100)
    const cropY = canvas.height * (cropArea.y / 100)
    const cropWidth = canvas.width * (cropArea.width / 100)
    const cropHeight = canvas.height * (cropArea.height / 100)
    
    // Создаем новый канвас для обрезанного изображения
    const cropCanvas = document.createElement('canvas')
    cropCanvas.width = cropWidth
    cropCanvas.height = cropHeight
    const cropCtx = cropCanvas.getContext('2d')
    
    // Копируем область кадрирования
    const imageData = ctx.getImageData(cropX, cropY, cropWidth, cropHeight)
    cropCtx.putImageData(imageData, 0, 0)
    
    // Создаем URL для обрезанного изображения
    const croppedDataUrl = cropCanvas.toDataURL('image/jpeg', 0.95)
    setEditedImageUrl(croppedDataUrl)
  }

  // Обработчики мыши для перетаскивания
  const handleMouseDown = useCallback((e) => {
    e.preventDefault()
    setIsDragging(true)
    setDragStart({ x: e.clientX, y: e.clientY })
    setPositionStart({ ...position })
  }, [position])

  const handleMouseMove = useCallback((e) => {
    if (!isDragging) return
    
    const deltaX = e.clientX - dragStart.x
    const deltaY = e.clientY - dragStart.y
    
    setPosition({
      x: positionStart.x + deltaX,
      y: positionStart.y + deltaY
    })
  }, [isDragging, dragStart, positionStart])

  const handleMouseUp = useCallback(() => {
    setIsDragging(false)
  }, [])

  // Обработчики для изменения размеров области кадрирования
  const handleCropMouseDown = useCallback((e, handle) => {
    e.preventDefault()
    e.stopPropagation()
    setIsResizing(true)
    setResizeHandle(handle)
    setResizeStart({ x: e.clientX, y: e.clientY })
    setCropStart({ ...cropArea })
  }, [cropArea])

  const handleCropMouseMove = useCallback((e) => {
    if (!isResizing || !resizeHandle || !cropStart) return
    
    const container = containerRef.current
    if (!container) return
    
    const rect = container.getBoundingClientRect()
    const deltaX = ((e.clientX - resizeStart.x) / rect.width) * 100
    const deltaY = ((e.clientY - resizeStart.y) / rect.height) * 100
    
    let newCrop = { ...cropStart }
    
    switch (resizeHandle) {
      case 'nw':
        newCrop.x = Math.max(0, Math.min(cropStart.x + deltaX, cropStart.x + cropStart.width - 5))
        newCrop.y = Math.max(0, Math.min(cropStart.y + deltaY, cropStart.y + cropStart.height - 5))
        newCrop.width = cropStart.width - (newCrop.x - cropStart.x)
        newCrop.height = cropStart.height - (newCrop.y - cropStart.y)
        break
      case 'ne':
        newCrop.y = Math.max(0, Math.min(cropStart.y + deltaY, cropStart.y + cropStart.height - 5))
        newCrop.width = Math.max(5, Math.min(100 - cropStart.x, cropStart.width + deltaX))
        newCrop.height = cropStart.height - (newCrop.y - cropStart.y)
        break
      case 'sw':
        newCrop.x = Math.max(0, Math.min(cropStart.x + deltaX, cropStart.x + cropStart.width - 5))
        newCrop.width = cropStart.width - (newCrop.x - cropStart.x)
        newCrop.height = Math.max(5, Math.min(100 - cropStart.y, cropStart.height + deltaY))
        break
      case 'se':
        newCrop.width = Math.max(5, Math.min(100 - cropStart.x, cropStart.width + deltaX))
        newCrop.height = Math.max(5, Math.min(100 - cropStart.y, cropStart.height + deltaY))
        break
    }
    
    setCropArea(newCrop)
  }, [isResizing, resizeHandle, resizeStart, cropStart])

  const handleCropMouseUp = useCallback(() => {
    setIsResizing(false)
    setResizeHandle(null)
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

  useEffect(() => {
    if (isResizing) {
      document.addEventListener('mousemove', handleCropMouseMove)
      document.addEventListener('mouseup', handleCropMouseUp)
      return () => {
        document.removeEventListener('mousemove', handleCropMouseMove)
        document.removeEventListener('mouseup', handleCropMouseUp)
      }
    }
  }, [isResizing, handleCropMouseMove, handleCropMouseUp])

  // Обработчики управления
  const handleZoomIn = () => setScale(prev => Math.min(prev + 0.2, 3))
  const handleZoomOut = () => setScale(prev => Math.max(prev - 0.2, 0.3))
  const handleRotate = () => setRotation(prev => (prev + 90) % 360)
  
  const handleReset = () => {
    setScale(1.5)
    setPosition({ x: 0, y: 0 })
    setRotation(0)
    setCropArea({ x: 0, y: 0, width: 100, height: 100 })
  }

  const handleCropToCenter = () => {
    setCropArea({
      x: 20,
      y: 20,
      width: 60,
      height: 60
    })
  }

  const handleContinue = () => {
    if (!editedImageUrl || !editorSettings) return

    try {
      fetch(editedImageUrl)
        .then(res => res.blob())
        .then(blob => {
          const file = new File([blob], 'edited-image.jpg', { type: 'image/jpeg' })
          const fileUrl = URL.createObjectURL(file)
          
          const updatedImageData = {
            size: editorSettings.size,
            selectedStyle: editorSettings.style,
            fileName: 'edited-image.jpg',
            previewUrl: editedImageUrl,
            timestamp: Date.now()
          }
          
          localStorage.setItem('diamondMosaic_selectedImage', JSON.stringify(updatedImageData))
          sessionStorage.setItem('diamondMosaic_fileUrl', fileUrl)
          localStorage.removeItem('diamondMosaic_editorSettings')
          
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
      <div className="min-h-screen bg-white flex items-center justify-center">
        <div className="text-gray-600">{t('diamond_mosaic_editor.loading')}</div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-white">
      <div className="max-w-4xl mx-auto px-6 py-8">
        
        {/* Заголовок */}
        <div className="text-center mb-8">
          <h1 className="text-2xl font-bold text-gray-900 mb-2">
            {t('diamond_mosaic_editor.main_title')}
          </h1>
          <h2 className="text-2xl font-bold text-gray-900">
            {t('diamond_mosaic_editor.subtitle')}
          </h2>
        </div>

        {/* Основная область редактирования */}
        <div className="bg-gray-200 rounded-lg p-8 mb-8">
          <div 
            ref={containerRef}
            className="relative bg-gray-400 mx-auto rounded-lg overflow-hidden"
            style={{ 
              width: '100%', 
              maxWidth: '800px',
              height: '600px',
              margin: '0 auto'
            }}
          >
            {/* Канвас с изображением */}
            <canvas
              ref={canvasRef}
              className="absolute inset-0 w-full h-full cursor-move"
              onMouseDown={handleMouseDown}
              style={{ touchAction: 'none' }}
            />
            
            {/* Область кадрирования с сеткой и ручками изменения размера */}
            <div 
              className="absolute border-2 border-white shadow-lg"
              style={{
                left: `${cropArea.x}%`,
                top: `${cropArea.y}%`,
                width: `${cropArea.width}%`,
                height: `${cropArea.height}%`,
                backgroundColor: 'rgba(255,255,255,0.05)'
              }}
            >
              {/* Сетка для помощи в кадрировании */}
              <div className="relative w-full h-full">
                {/* Вертикальные линии */}
                <div className="absolute left-1/3 top-0 w-px h-full bg-white opacity-30"></div>
                <div className="absolute left-2/3 top-0 w-px h-full bg-white opacity-30"></div>
                {/* Горизонтальные линии */}
                <div className="absolute top-1/3 left-0 w-full h-px bg-white opacity-30"></div>
                <div className="absolute top-2/3 left-0 w-full h-px bg-white opacity-30"></div>
                
                {/* Ручки для изменения размера */}
                <div 
                  className="absolute w-3 h-3 bg-white border-2 border-purple-500 rounded-full cursor-nw-resize -top-1.5 -left-1.5"
                  onMouseDown={(e) => handleCropMouseDown(e, 'nw')}
                />
                <div 
                  className="absolute w-3 h-3 bg-white border-2 border-purple-500 rounded-full cursor-ne-resize -top-1.5 -right-1.5"
                  onMouseDown={(e) => handleCropMouseDown(e, 'ne')}
                />
                <div 
                  className="absolute w-3 h-3 bg-white border-2 border-purple-500 rounded-full cursor-sw-resize -bottom-1.5 -left-1.5"
                  onMouseDown={(e) => handleCropMouseDown(e, 'sw')}
                />
                <div 
                  className="absolute w-3 h-3 bg-white border-2 border-purple-500 rounded-full cursor-se-resize -bottom-1.5 -right-1.5"
                  onMouseDown={(e) => handleCropMouseDown(e, 'se')}
                />
              </div>
            </div>
          </div>
        </div>

        {/* Панель управления */}
        <div className="flex justify-center items-center gap-4 mb-8">
          
          {/* Загрузить изображение */}
          <button className="w-12 h-12 bg-purple-200 rounded-xl flex items-center justify-center hover:bg-purple-300 transition-colors">
            <Download className="w-6 h-6 text-purple-700" />
          </button>

          {/* Уменьшить */}
          <button 
            onClick={handleZoomOut}
            className="w-12 h-12 bg-purple-200 rounded-xl flex items-center justify-center hover:bg-purple-300 transition-colors"
          >
            <Minus className="w-6 h-6 text-purple-700" />
          </button>

          {/* Увеличить */}
          <button 
            onClick={handleZoomIn}
            className="w-12 h-12 bg-purple-200 rounded-xl flex items-center justify-center hover:bg-purple-300 transition-colors"
          >
            <Plus className="w-6 h-6 text-purple-700" />
          </button>

          {/* Поворот */}
          <button 
            onClick={handleRotate}
            className="w-12 h-12 bg-purple-200 rounded-xl flex items-center justify-center hover:bg-purple-300 transition-colors"
          >
            <RotateCw className="w-6 h-6 text-purple-700" />
          </button>

          {/* Назад */}
          <button 
            onClick={handleBack}
            className="w-12 h-12 bg-purple-200 rounded-xl flex items-center justify-center hover:bg-purple-300 transition-colors"
          >
            <ArrowLeft className="w-6 h-6 text-purple-700" />
          </button>

          {/* Сброс */}
          <button 
            onClick={handleReset}
            className="w-12 h-12 bg-purple-200 rounded-xl flex items-center justify-center hover:bg-purple-300 transition-colors"
          >
            <Undo2 className="w-6 h-6 text-purple-700" />
          </button>

          {/* Кадрирование по центру */}
          <button 
            onClick={handleCropToCenter}
            className="w-12 h-12 bg-orange-200 rounded-xl flex items-center justify-center hover:bg-orange-300 transition-colors"
            title="Кадрировать по центру"
          >
            <span className="text-orange-700 font-bold text-sm">□</span>
          </button>
        </div>

        {/* Кнопки действий */}
        <div className="flex gap-4 max-w-md mx-auto">
          <button
            onClick={handleBack}
            className="flex-1 py-4 px-6 bg-white border-2 border-purple-300 text-gray-700 rounded-full font-medium hover:bg-gray-50 transition-colors"
          >
            {t('diamond_mosaic_editor.buttons.back')}
          </button>
          
          <button
            onClick={handleContinue}
            disabled={!editedImageUrl}
            className="flex-1 py-4 px-6 bg-gradient-to-r from-purple-400 to-purple-600 text-white rounded-full font-medium hover:from-purple-500 hover:to-purple-700 transition-all disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
          >
            {t('diamond_mosaic_editor.buttons.continue')}
            <span>→</span>
          </button>
        </div>

      </div>
    </div>
  )
}

export default DiamondMosaicEditorPage
