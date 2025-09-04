import React, { useState, useRef, useCallback, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import ReactCrop from 'react-image-crop'
import { ZoomIn, ZoomOut, RotateCw } from 'lucide-react'
import 'react-image-crop/dist/ReactCrop.css'

const ImageEditor = ({ 
  imageUrl, 
  onSave, 
  onCancel, 
  title = "Настройте изображение",
  showCropHint = true,
  aspectRatio,
  fileName: propFileName
}) => {
  const { t } = useTranslation()
  const imgRef = useRef(null)
  const containerRef = useRef(null)
  const [crop, setCrop] = useState(null) 
  const [completedCrop, setCompletedCrop] = useState(null)
  const [scale, setScale] = useState(1)
  const [rotate, setRotate] = useState(0)
  const [fileName, setFileName] = useState('')
  const [position, setPosition] = useState({ x: 0, y: 0 })
  const [isDragging, setIsDragging] = useState(false)
  const [lastMousePos, setLastMousePos] = useState({ x: 0, y: 0 })

  
  useEffect(() => {
    
    if (propFileName) {
      setFileName(propFileName)
      return
    }
    
    if (imageUrl) {
      
      if (imageUrl.startsWith('data:')) {
        
        try {
          const savedImageData = localStorage.getItem('diamondMosaic_selectedImage')
          if (savedImageData) {
            const parsedData = JSON.parse(savedImageData)
            if (parsedData.fileName) {
              setFileName(parsedData.fileName)
              return
            }
          }
        } catch (error) {
          console.error('Error getting filename from localStorage:', error)
        }
        setFileName('uploaded_image.jpg')
      } else {
        const parts = imageUrl.split('/')
        const name = parts[parts.length - 1]
        setFileName(name)
      }
    }
  }, [imageUrl, propFileName])

  
  function onImageLoad(e) {
    
    
    setCrop(null)
    setCompletedCrop(null)
  }

  
  const generateCroppedImage = useCallback(() => {
    if (!completedCrop || !imgRef.current) {
      return null
    }

    const image = imgRef.current
    const canvas = document.createElement('canvas')
    const ctx = canvas.getContext('2d')

    const scaleX = image.naturalWidth / image.width
    const scaleY = image.naturalHeight / image.height

    
    canvas.width = completedCrop.width * scaleX
    canvas.height = completedCrop.height * scaleY

    
    ctx.save()
    ctx.translate(canvas.width / 2 + position.x, canvas.height / 2 + position.y)
    ctx.rotate((rotate * Math.PI) / 180)
    ctx.scale(scale, scale)
    ctx.translate(-canvas.width / 2, -canvas.height / 2)

    
    ctx.drawImage(
      image,
      completedCrop.x * scaleX,
      completedCrop.y * scaleY,
      completedCrop.width * scaleX,
      completedCrop.height * scaleY,
      0,
      0,
      canvas.width,
      canvas.height
    )

    ctx.restore()

    
    return canvas.toDataURL('image/jpeg')
  }, [completedCrop, rotate, scale, position])

  
  const handleRotate = () => {
    setRotate((prevRotate) => (prevRotate + 90) % 360)
  }

  const handleZoomIn = () => {
    setScale(Math.min(scale + 0.1, 3))
  }

  const handleZoomOut = () => {
    setScale(Math.max(scale - 0.1, 0.1))
  }

  const handleReset = () => {
    setRotate(0)
    setScale(1)
    setPosition({ x: 0, y: 0 })
    setCrop(null) 
  }

  
  const handleMouseDown = (e) => {
    if (e.button === 2) { 
      e.preventDefault()
      setIsDragging(true)
      setLastMousePos({ x: e.clientX, y: e.clientY })
    }
  }

  const handleMouseMove = (e) => {
    if (isDragging) {
      e.preventDefault()
      const deltaX = e.clientX - lastMousePos.x
      const deltaY = e.clientY - lastMousePos.y
      
      setPosition(prev => ({
        x: prev.x + deltaX,
        y: prev.y + deltaY
      }))
      
      setLastMousePos({ x: e.clientX, y: e.clientY })
    }
  }

  const handleMouseUp = (e) => {
    setIsDragging(false)
  }



  
  const handleContextMenu = (e) => {
    e.preventDefault()
  }

  
  useEffect(() => {
    const handleGlobalMouseMove = (e) => {
      if (isDragging) {
        handleMouseMove(e)
      }
    }

    const handleGlobalMouseUp = (e) => {
      if (isDragging) {
        handleMouseUp(e)
      }
    }

    if (isDragging) {
      document.addEventListener('mousemove', handleGlobalMouseMove)
      document.addEventListener('mouseup', handleGlobalMouseUp)
      document.addEventListener('contextmenu', handleContextMenu)
      document.body.style.cursor = 'grabbing'
      document.body.style.overflow = 'hidden' 
      
      return () => {
        document.removeEventListener('mousemove', handleGlobalMouseMove)
        document.removeEventListener('mouseup', handleGlobalMouseUp)
        document.removeEventListener('contextmenu', handleContextMenu)
        document.body.style.cursor = 'default'
        document.body.style.overflow = 'auto' 
      }
    }
  }, [isDragging, lastMousePos])

  const handleSave = () => {
    const croppedImageUrl = generateCroppedImage()
    if (croppedImageUrl) {
      onSave(croppedImageUrl, { crop, rotate, scale, position })
    }
  }

  return (
    <div className="bg-white rounded-2xl shadow-lg">
      
      <div className="flex items-center justify-between p-6 border-b border-gray-200">
        <h2 className="text-xl font-semibold text-gray-900">{title}</h2>
        {onCancel && (
          <button
            onClick={onCancel}
            className="text-red-500 hover:text-red-600 text-sm font-medium"
          >
            ✕ {t('diamond_mosaic_page.image_editor.delete_image')}
          </button>
        )}
      </div>

      
      <div className="p-6">
        <div 
          ref={containerRef}
          className="bg-gray-600 rounded-lg overflow-hidden mb-6 relative w-full flex items-center justify-center"
          style={{
            minHeight: '400px',
            maxHeight: '70vh'
          }}
        >
          <div 
            className="relative"
            style={{
              width: '100%',
              height: '100%',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center'
            }}
            onMouseDown={handleMouseDown}
            onMouseUp={handleMouseUp}
            onContextMenu={handleContextMenu}
          >
            <ReactCrop
              crop={crop}
              onChange={(c) => !isDragging && setCrop(c)}
              onComplete={(c) => !isDragging && setCompletedCrop(c)}
              aspect={aspectRatio}
              disabled={isDragging}
              style={{
                maxWidth: '100%',
                maxHeight: '100%',
                display: 'block',
                pointerEvents: isDragging ? 'none' : 'auto'
              }}
            >
                <img
                  ref={imgRef}
                  src={imageUrl}
                  alt="Edit"
                  style={{
                    transform: `translate(${position.x}px, ${position.y}px) rotate(${rotate}deg) scale(${scale})`,
                    maxWidth: rotate === 90 || rotate === 270 ? '70vh' : '100%',
                    maxHeight: rotate === 90 || rotate === 270 ? '100%' : '70vh',
                    objectFit: 'contain',
                    transformOrigin: 'center center',
                    display: 'block',
                    cursor: isDragging ? 'grabbing' : 'grab',
                    userSelect: 'none'
                  }}
                  onLoad={onImageLoad}
                  draggable={false}
                              />
            </ReactCrop>
          </div>
        </div>

        
        <div className="flex items-end mb-6">
          
          <div className="flex-1 flex justify-start">
            <button
              onClick={handleRotate}
              className="flex items-center px-6 py-3 bg-purple-100 hover:bg-purple-200 text-purple-700 rounded-lg transition-colors font-medium"
              title="Повернуть на 90°"
            >
              <RotateCw className="w-5 h-5 mr-2" />
              Повернуть на 90°
            </button>
          </div>

          
          <div className="flex-1 flex flex-col items-center gap-2">
            <span className="text-sm text-gray-600 font-medium">
              Масштаб: {Math.round(scale * 100)}%
            </span>
            <div className="flex items-center gap-2">
              <button
                onClick={handleZoomOut}
                className="w-12 h-12 bg-gray-100 hover:bg-gray-200 text-gray-700 rounded-lg flex items-center justify-center transition-colors"
                title="Уменьшить"
              >
                <ZoomOut className="w-5 h-5" />
              </button>
              <button
                onClick={handleZoomIn}
                className="w-12 h-12 bg-gray-100 hover:bg-gray-200 text-gray-700 rounded-lg flex items-center justify-center transition-colors"
                title="Увеличить"
              >
                <ZoomIn className="w-5 h-5" />
              </button>
            </div>
          </div>

          
          <div className="flex-1 flex justify-end">
            <button
              onClick={handleReset}
              className="flex items-center px-6 py-3 bg-orange-100 hover:bg-orange-200 text-orange-700 rounded-lg transition-colors font-medium"
            >
              Сбросить
            </button>
          </div>
        </div>

        
        <div className="text-center bg-white border border-gray-200 rounded-lg p-6 mb-6">
          {fileName && (
            <div className="flex items-center justify-center mb-3">
              <div className="w-2 h-2 bg-green-500 rounded-full mr-2"></div>
              <p className="text-base font-medium text-gray-800">{fileName}</p>
            </div>
          )}
          
          <p className="text-base text-gray-700">
            Вы можете выбрать область изображения что будет использована для создания превью!
          </p>
        </div>
      </div>
    </div>
  )
}

export default ImageEditor