import React, { useState, useRef, useEffect } from 'react'
import { motion } from 'framer-motion'
import { 
  Crop, 
  RotateCw, 
  ZoomIn, 
  ZoomOut, 
  Undo, 
  Redo,
  Save,
  X,
  Move,
  Maximize2,
  FlipHorizontal,
  FlipVertical,
  Sun,
  Contrast,
  Palette
} from 'lucide-react'

const ImageEditorEnhanced = ({ imageUrl, onSave, onCancel, initialSettings = {} }) => {
  const canvasRef = useRef(null)
  const imageRef = useRef(null)
  
  const [zoom, setZoom] = useState(1)
  const [rotation, setRotation] = useState(0)
  const [position, setPosition] = useState({ x: 0, y: 0 })
  const [cropArea, setCropArea] = useState(null)
  const [flipH, setFlipH] = useState(false)
  const [flipV, setFlipV] = useState(false)
  const [brightness, setBrightness] = useState(100)
  const [contrast, setContrast] = useState(100)
  const [saturation, setSaturation] = useState(100)
  const [history, setHistory] = useState([])
  const [historyIndex, setHistoryIndex] = useState(-1)
  const [isDragging, setIsDragging] = useState(false)
  const [dragStart, setDragStart] = useState({ x: 0, y: 0 })
  const [isCropping, setIsCropping] = useState(false)
  const [cropStart, setCropStart] = useState(null)

  useEffect(() => {
    loadImage()
  }, [imageUrl])

  const loadImage = () => {
    const img = new Image()
    img.onload = () => {
      imageRef.current = img
      renderCanvas()
    }
    img.src = imageUrl
  }

  const renderCanvas = () => {
    if (!canvasRef.current || !imageRef.current) return
    
    const canvas = canvasRef.current
    const ctx = canvas.getContext('2d')
    const img = imageRef.current
    
    // Clear canvas
    ctx.clearRect(0, 0, canvas.width, canvas.height)
    
    // Save context state
    ctx.save()
    
    // Apply transformations
    ctx.translate(canvas.width / 2, canvas.height / 2)
    ctx.rotate((rotation * Math.PI) / 180)
    ctx.scale(
      zoom * (flipH ? -1 : 1), 
      zoom * (flipV ? -1 : 1)
    )
    ctx.translate(-canvas.width / 2, -canvas.height / 2)
    
    // Apply filters
    ctx.filter = `brightness(${brightness}%) contrast(${contrast}%) saturate(${saturation}%)`
    
    // Draw image
    const imgWidth = img.width * zoom
    const imgHeight = img.height * zoom
    const x = (canvas.width - imgWidth) / 2 + position.x
    const y = (canvas.height - imgHeight) / 2 + position.y
    
    ctx.drawImage(img, x, y, imgWidth, imgHeight)
    
    // Draw crop area if cropping
    if (isCropping && cropArea) {
      ctx.strokeStyle = 'rgba(255, 255, 255, 0.8)'
      ctx.lineWidth = 2
      ctx.setLineDash([5, 5])
      ctx.strokeRect(
        cropArea.x,
        cropArea.y,
        cropArea.width,
        cropArea.height
      )
    }
    
    // Restore context state
    ctx.restore()
  }

  useEffect(() => {
    renderCanvas()
  }, [zoom, rotation, position, flipH, flipV, brightness, contrast, saturation, cropArea])

  const handleZoomIn = () => {
    setZoom(prev => Math.min(prev + 0.1, 3))
    saveToHistory()
  }

  const handleZoomOut = () => {
    setZoom(prev => Math.max(prev - 0.1, 0.5))
    saveToHistory()
  }

  const handleRotate = () => {
    setRotation(prev => (prev + 90) % 360)
    saveToHistory()
  }

  const handleFlipHorizontal = () => {
    setFlipH(prev => !prev)
    saveToHistory()
  }

  const handleFlipVertical = () => {
    setFlipV(prev => !prev)
    saveToHistory()
  }

  const handleReset = () => {
    setZoom(1)
    setRotation(0)
    setPosition({ x: 0, y: 0 })
    setFlipH(false)
    setFlipV(false)
    setBrightness(100)
    setContrast(100)
    setSaturation(100)
    setCropArea(null)
    setIsCropping(false)
    saveToHistory()
  }

  const saveToHistory = () => {
    const state = {
      zoom,
      rotation,
      position,
      flipH,
      flipV,
      brightness,
      contrast,
      saturation
    }
    
    const newHistory = history.slice(0, historyIndex + 1)
    newHistory.push(state)
    setHistory(newHistory)
    setHistoryIndex(newHistory.length - 1)
  }

  const handleUndo = () => {
    if (historyIndex > 0) {
      const prevState = history[historyIndex - 1]
      applyState(prevState)
      setHistoryIndex(historyIndex - 1)
    }
  }

  const handleRedo = () => {
    if (historyIndex < history.length - 1) {
      const nextState = history[historyIndex + 1]
      applyState(nextState)
      setHistoryIndex(historyIndex + 1)
    }
  }

  const applyState = (state) => {
    setZoom(state.zoom)
    setRotation(state.rotation)
    setPosition(state.position)
    setFlipH(state.flipH)
    setFlipV(state.flipV)
    setBrightness(state.brightness)
    setContrast(state.contrast)
    setSaturation(state.saturation)
  }

  const handleMouseDown = (e) => {
    if (isCropping) {
      const rect = canvasRef.current.getBoundingClientRect()
      const x = e.clientX - rect.left
      const y = e.clientY - rect.top
      setCropStart({ x, y })
      setCropArea({ x, y, width: 0, height: 0 })
    } else {
      setIsDragging(true)
      setDragStart({
        x: e.clientX - position.x,
        y: e.clientY - position.y
      })
    }
  }

  const handleMouseMove = (e) => {
    if (isCropping && cropStart) {
      const rect = canvasRef.current.getBoundingClientRect()
      const x = e.clientX - rect.left
      const y = e.clientY - rect.top
      setCropArea({
        x: Math.min(cropStart.x, x),
        y: Math.min(cropStart.y, y),
        width: Math.abs(x - cropStart.x),
        height: Math.abs(y - cropStart.y)
      })
    } else if (isDragging) {
      setPosition({
        x: e.clientX - dragStart.x,
        y: e.clientY - dragStart.y
      })
    }
  }

  const handleMouseUp = () => {
    if (isCropping) {
      setCropStart(null)
    } else {
      setIsDragging(false)
    }
  }

  const handleSave = () => {
    const canvas = canvasRef.current
    canvas.toBlob((blob) => {
      const url = URL.createObjectURL(blob)
      onSave(url, {
        zoom,
        rotation,
        position,
        flipH,
        flipV,
        brightness,
        contrast,
        saturation,
        cropArea
      })
    }, 'image/jpeg', 0.95)
  }

  return (
    <div className="fixed inset-0 z-50 bg-black bg-opacity-90 flex flex-col">
      {/* Header */}
      <div className="bg-gray-900 text-white p-4 flex items-center justify-between">
        <h2 className="text-xl font-semibold">Редактор изображения</h2>
        <div className="flex items-center gap-4">
          <button
            onClick={handleUndo}
            disabled={historyIndex <= 0}
            className="p-2 hover:bg-gray-700 rounded disabled:opacity-50"
            title="Отменить"
          >
            <Undo className="w-5 h-5" />
          </button>
          <button
            onClick={handleRedo}
            disabled={historyIndex >= history.length - 1}
            className="p-2 hover:bg-gray-700 rounded disabled:opacity-50"
            title="Повторить"
          >
            <Redo className="w-5 h-5" />
          </button>
          <button
            onClick={handleReset}
            className="p-2 hover:bg-gray-700 rounded"
            title="Сбросить"
          >
            <Maximize2 className="w-5 h-5" />
          </button>
          <div className="h-6 w-px bg-gray-600" />
          <button
            onClick={handleSave}
            className="px-4 py-2 bg-purple-600 hover:bg-purple-700 rounded-lg flex items-center gap-2"
          >
            <Save className="w-5 h-5" />
            Сохранить
          </button>
          <button
            onClick={onCancel}
            className="p-2 hover:bg-gray-700 rounded"
            title="Закрыть"
          >
            <X className="w-5 h-5" />
          </button>
        </div>
      </div>

      {/* Main editing area */}
      <div className="flex-1 flex">
        {/* Sidebar with tools */}
        <div className="w-64 bg-gray-800 text-white p-4 overflow-y-auto">
          <div className="space-y-6">
            {/* Transform tools */}
            <div>
              <h3 className="text-sm font-semibold mb-3 text-gray-400">Трансформация</h3>
              <div className="grid grid-cols-2 gap-2">
                <button
                  onClick={handleZoomIn}
                  className="p-3 bg-gray-700 hover:bg-gray-600 rounded flex items-center justify-center"
                  title="Увеличить"
                >
                  <ZoomIn className="w-5 h-5" />
                </button>
                <button
                  onClick={handleZoomOut}
                  className="p-3 bg-gray-700 hover:bg-gray-600 rounded flex items-center justify-center"
                  title="Уменьшить"
                >
                  <ZoomOut className="w-5 h-5" />
                </button>
                <button
                  onClick={handleRotate}
                  className="p-3 bg-gray-700 hover:bg-gray-600 rounded flex items-center justify-center"
                  title="Повернуть"
                >
                  <RotateCw className="w-5 h-5" />
                </button>
                <button
                  onClick={() => setIsCropping(!isCropping)}
                  className={`p-3 ${isCropping ? 'bg-purple-600' : 'bg-gray-700'} hover:bg-gray-600 rounded flex items-center justify-center`}
                  title="Кадрировать"
                >
                  <Crop className="w-5 h-5" />
                </button>
                <button
                  onClick={handleFlipHorizontal}
                  className="p-3 bg-gray-700 hover:bg-gray-600 rounded flex items-center justify-center"
                  title="Отразить по горизонтали"
                >
                  <FlipHorizontal className="w-5 h-5" />
                </button>
                <button
                  onClick={handleFlipVertical}
                  className="p-3 bg-gray-700 hover:bg-gray-600 rounded flex items-center justify-center"
                  title="Отразить по вертикали"
                >
                  <FlipVertical className="w-5 h-5" />
                </button>
              </div>
            </div>

            {/* Adjustment sliders */}
            <div>
              <h3 className="text-sm font-semibold mb-3 text-gray-400">Коррекция</h3>
              <div className="space-y-4">
                <div>
                  <label className="flex items-center justify-between mb-1">
                    <span className="text-sm flex items-center gap-2">
                      <Sun className="w-4 h-4" />
                      Яркость
                    </span>
                    <span className="text-xs">{brightness}%</span>
                  </label>
                  <input
                    type="range"
                    min="50"
                    max="150"
                    value={brightness}
                    onChange={(e) => setBrightness(e.target.value)}
                    className="w-full"
                  />
                </div>
                <div>
                  <label className="flex items-center justify-between mb-1">
                    <span className="text-sm flex items-center gap-2">
                      <Contrast className="w-4 h-4" />
                      Контраст
                    </span>
                    <span className="text-xs">{contrast}%</span>
                  </label>
                  <input
                    type="range"
                    min="50"
                    max="150"
                    value={contrast}
                    onChange={(e) => setContrast(e.target.value)}
                    className="w-full"
                  />
                </div>
                <div>
                  <label className="flex items-center justify-between mb-1">
                    <span className="text-sm flex items-center gap-2">
                      <Palette className="w-4 h-4" />
                      Насыщенность
                    </span>
                    <span className="text-xs">{saturation}%</span>
                  </label>
                  <input
                    type="range"
                    min="0"
                    max="200"
                    value={saturation}
                    onChange={(e) => setSaturation(e.target.value)}
                    className="w-full"
                  />
                </div>
              </div>
            </div>

            {/* Info */}
            <div className="text-xs text-gray-400">
              <p>Масштаб: {(zoom * 100).toFixed(0)}%</p>
              <p>Поворот: {rotation}°</p>
              {cropArea && (
                <p>Область кадрирования: {Math.round(cropArea.width)}x{Math.round(cropArea.height)}</p>
              )}
            </div>
          </div>
        </div>

        {/* Canvas area */}
        <div className="flex-1 flex items-center justify-center bg-gray-900 relative overflow-hidden">
          <canvas
            ref={canvasRef}
            width={800}
            height={600}
            className="bg-white cursor-move"
            onMouseDown={handleMouseDown}
            onMouseMove={handleMouseMove}
            onMouseUp={handleMouseUp}
            onMouseLeave={handleMouseUp}
            style={{ cursor: isCropping ? 'crosshair' : 'move' }}
          />
        </div>
      </div>
    </div>
  )
}

export default ImageEditorEnhanced