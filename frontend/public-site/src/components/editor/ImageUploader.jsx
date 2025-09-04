import React, { useState, useRef, useCallback } from 'react'
import { useTranslation } from 'react-i18next'
import { motion, AnimatePresence } from 'framer-motion'
import { Upload, X, RotateCw, ZoomIn, ZoomOut, Crop, RefreshCw } from 'lucide-react'
import { useMutation } from '@tanstack/react-query'
import { MosaicAPI } from '../../api/client'
import { useUIStore } from '../../store/partnerStore'

const ImageUploader = ({ onImageUploaded, couponCode, couponSize, initialImageId = null }) => {
  const { t } = useTranslation()
  const { addNotification } = useUIStore()
  const [dragActive, setDragActive] = useState(false)
  const [selectedFile, setSelectedFile] = useState(null)
  const [preview, setPreview] = useState(null)
  const [imageData, setImageData] = useState(initialImageId ? { image_id: initialImageId, id: initialImageId } : null)
  const [editMode, setEditMode] = useState(false)
  const [scale, setScale] = useState(1)
  const [rotation, setRotation] = useState(0)
  const [cropData, setCropData] = useState(null)
  const fileInputRef = useRef(null)
  const canvasRef = useRef(null)

  const uploadImageMutation = useMutation({
    mutationFn: async (formData) => {
      return await MosaicAPI.uploadImage(formData)
    },
    onSuccess: (data) => {
      addNotification({
        type: 'success',
        title: t('notifications.image_uploaded'),
        message: t('notifications.image_uploaded_desc')
      })
      
      setImageData(data)
      try {
        if (data?.image_id || data?.id) {
          sessionStorage.setItem('editor:lastImageId', data.image_id || data.id)
        }
      } catch {}
      onImageUploaded(data)
    },
    onError: (error) => {
      const msgRaw = (error && (error.message || error?.original?.response?.data?.detail)) || ''
      const msg = /not activated/i.test(msgRaw)
        ? t('notifications.coupon_not_activated')
        : /coupon not found/i.test(msgRaw)
        ? t('notifications.invalid_coupon')
        : /invalid image type/i.test(msgRaw)
        ? t('notifications.invalid_file_type')
        : /file too large/i.test(msgRaw)
        ? t('notifications.file_too_large')
        : t('notifications.upload_error_desc')
      addNotification({ type: 'error', title: t('notifications.upload_error'), message: msg })
    }
  })

  const validateFile = (file) => {
    if (!file || !Number.isFinite(file.size)) {
      addNotification({ type: 'error', message: t('notifications.invalid_file_type') })
      return false
    }
    const validTypes = ['image/jpeg', 'image/jpg', 'image/png']
    const maxSize = 15 * 1024 * 1024 
    if (!validTypes.includes(file.type)) {
      addNotification({
        type: 'error',
        message: t('notifications.invalid_file_type')
      })
      return false
    }

    if (file.size > maxSize) {
      addNotification({
        type: 'error',
        message: t('notifications.file_too_large')
      })
      return false
    }

    return true
  }

  const formatBytes = (bytes) => {
    if (!Number.isFinite(bytes)) return ''
    const units = ['B', 'KB', 'MB', 'GB']
    let i = 0
    let val = bytes
    while (val >= 1024 && i < units.length - 1) {
      val /= 1024
      i++
    }
    return `${val.toFixed(2)} ${units[i]}`
  }

  const handleFile = useCallback((file) => {
    if (!validateFile(file)) return

    setSelectedFile(file)
    
    
    const reader = new FileReader()
    reader.onload = (e) => {
      setPreview(e.target.result)
      setEditMode(true)
      setScale(1)
      setRotation(0)
      setCropData(null)
      try { sessionStorage.setItem('editor:lastPreview', e.target.result) } catch {}
    }
    reader.readAsDataURL(file)
  }, [addNotification, t])

  const handleDrag = useCallback((e) => {
    e.preventDefault()
    e.stopPropagation()
    if (e.type === "dragenter" || e.type === "dragover") {
      setDragActive(true)
    } else if (e.type === "dragleave") {
      setDragActive(false)
    }
  }, [])

  const handleDrop = useCallback((e) => {
    e.preventDefault()
    e.stopPropagation()
    setDragActive(false)
    
    if (e.dataTransfer.files && e.dataTransfer.files[0]) {
      handleFile(e.dataTransfer.files[0])
    }
  }, [handleFile])

  const handleFileInput = (e) => {
    if (e.target.files && e.target.files[0]) {
      handleFile(e.target.files[0])
    }
  }

  const handleUpload = async () => {
    
    const effectiveCoupon = couponCode || (() => { try { return sessionStorage.getItem('editor:coupon') || '' } catch { return '' } })()
    if (!selectedFile || !effectiveCoupon) {
      addNotification({ type: 'error', message: t('notifications.coupon_required') })
      return
    }

    console.log('ImageUploader: handleUpload called with couponCode:', couponCode)
    console.log('ImageUploader: effectiveCoupon:', effectiveCoupon)
    console.log('ImageUploader: selectedFile:', selectedFile)

    const formData = new FormData()
    formData.append('image', selectedFile)
    
    
    const clean = effectiveCoupon.replace(/\D/g, '')
    console.log('ImageUploader: Cleaned coupon code:', clean)
    
    if (clean.length !== 12) {
      addNotification({ type: 'error', message: t('notifications.invalid_coupon_format') })
      return
    }
    formData.append('coupon_code', clean)
    console.log('ImageUploader: FormData prepared with coupon_code:', clean)

    uploadImageMutation.mutate(formData)
  }

  const handleRotate = (direction) => {
    setRotation(prev => {
      const newRotation = direction === 'left' ? prev - 90 : prev + 90
      return newRotation % 360
    })
  }

  const handleZoom = (direction) => {
    setScale(prev => {
      const newScale = direction === 'in' ? prev * 1.2 : prev / 1.2
      return Math.max(0.1, Math.min(5, newScale))
    })
  }

  const handleReset = () => {
    setScale(1)
    setRotation(0)
    setCropData(null)
  }

    React.useEffect(() => {
    try {
      const prev = sessionStorage.getItem('editor:lastPreview')
      if (prev && !preview) {
        setPreview(prev)
        setEditMode(true)
      }
      
            if (imageData?.image_id) {
        const savedEdits = sessionStorage.getItem(`editor:edits:${imageData.image_id}`)
        if (savedEdits) {
          const edits = JSON.parse(savedEdits)
          setRotation(edits.rotation || 0)
          setScale(edits.scale || 1)
          if (edits.crop_width > 0 && edits.crop_height > 0) {
            setCropData({
              x: edits.crop_x || 0,
              y: edits.crop_y || 0,
              width: edits.crop_width || 0,
              height: edits.crop_height || 0
            })
          }
        }
      }
    } catch {}
  }, [imageData?.image_id])

  const handleReplace = () => {
    setSelectedFile(null)
    setPreview(null)
    setImageData(null)
    setEditMode(false)
    setScale(1)
    setRotation(0)
    setCropData(null)
    if (fileInputRef.current) {
      fileInputRef.current.value = ''
    }
  }

  const handleCrop = () => {
            const ratioMap = {
      '21x30': 3/4,
      '30x40': 3/4,
      '40x40': 1,
      '40x50': 4/5,
      '40x60': 2/3,
      '50x70': 5/7,
    }
    const ratio = ratioMap[couponSize] || 3/4
        const widthPct = 80
    const heightPct = Math.min(80 / ratio, 90)
    const xPct = (100 - widthPct) / 2
    const yPct = (100 - heightPct) / 2
    setCropData({ x: xPct, y: yPct, width: widthPct, height: heightPct })
  }

  const applyEdits = async () => {
    if (!imageData?.image_id) return
    const params = {
      crop_x: cropData?.x || 0,
      crop_y: cropData?.y || 0,
      crop_width: cropData?.width || 0,
      crop_height: cropData?.height || 0,
      rotation: ((rotation % 360) + 360) % 360,
      scale: scale,
    }
    
    
    try {
      sessionStorage.setItem(`editor:edits:${imageData.image_id}`, JSON.stringify(params))
    } catch {}
    
    try {
      await MosaicAPI.editImage(imageData.image_id, params)
      addNotification({ 
        type: 'success', 
        message: t('notifications.edits_applied', 'Изменения применены') 
      })
    } catch (e) {
      addNotification({ type: 'error', message: t('notifications.style_error_desc') })
    }
  }

  return (
    <div className="space-y-6">
      <AnimatePresence mode="wait">
        {!editMode ? (
          <motion.div
            key="upload"
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -20 }}
            transition={{ duration: 0.3 }}
          >
            <div
              className={`relative border-2 border-dashed rounded-xl p-12 transition-all duration-300 ${
                dragActive 
                  ? 'border-brand-primary bg-brand-primary/5' 
                  : 'border-gray-300 hover:border-gray-400'
              }`}
              onDragEnter={handleDrag}
              onDragLeave={handleDrag}
              onDragOver={handleDrag}
              onDrop={handleDrop}
            >
              <div className="text-center">
                <Upload className={`w-16 h-16 mx-auto mb-6 transition-colors ${
                  dragActive ? 'text-brand-primary' : 'text-gray-400'
                }`} />
                
                <h3 className="text-2xl font-semibold text-gray-900 mb-4">
                  {dragActive ? t('editor.upload.drag_active') : t('editor.upload.title')}
                </h3>
                
                <p className="text-gray-600 mb-4">
                  {t('editor.upload.description')}
                </p>
                
                <p className="text-sm text-gray-500 mb-6">
                  {t('editor.upload.file_types')}
                </p>
                
                <input
                  ref={fileInputRef}
                  type="file"
                  id="image-upload"
                  accept="image/*"
                  onChange={handleFileInput}
                  className="hidden"
                />
                
                <button 
                  onClick={() => fileInputRef.current?.click()}
                  className="inline-flex items-center px-6 py-3 bg-brand-primary text-white rounded-lg hover:bg-brand-primary/90 transition-colors"
                >
                  <Upload className="w-5 h-5 mr-2" />
                  {t('editor.upload.button')}
                </button>
              </div>
            </div>
          </motion.div>
        ) : (
          <motion.div
            key="editor"
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -20 }}
            transition={{ duration: 0.3 }}
            className="space-y-6"
          >
            
            <div className="flex items-center justify-center gap-4 p-4 bg-gray-50 rounded-lg overflow-x-auto">
              <div className="flex items-center gap-4 min-w-max">
                <button
                  onClick={() => handleRotate('left')}
                  className="flex items-center px-3 py-2 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors flex-shrink-0"
                  title={t('editor.tools.rotate_left')}
                >
                  <RotateCw className="w-4 h-4 mr-2" />
                  {t('editor.tools.rotate_left')}
                </button>
                
                <button
                  onClick={() => handleRotate('right')}
                  className="flex items-center px-3 py-2 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors flex-shrink-0"
                  title={t('editor.tools.rotate_right')}
                >
                  <RotateCw className="w-4 h-4 mr-2 rotate-180" />
                  {t('editor.tools.rotate_right')}
                </button>
                
                <button
                  onClick={() => handleZoom('in')}
                  className="flex items-center px-3 py-2 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors flex-shrink-0"
                  title={t('editor.tools.zoom_in')}
                >
                  <ZoomIn className="w-4 h-4 mr-2" />
                  {t('editor.tools.zoom_in')}
                </button>
                
                <button
                  onClick={() => handleZoom('out')}
                  className="flex items-center px-3 py-2 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors flex-shrink-0"
                  title={t('editor.tools.zoom_out')}
                >
                  <ZoomOut className="w-4 h-4 mr-2" />
                  {t('editor.tools.zoom_out')}
                </button>
                
                <button
                  onClick={handleCrop}
                  className="flex items-center px-3 py-2 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors flex-shrink-0"
                  title={t('editor.tools.crop')}
                >
                  <Crop className="w-4 h-4 mr-2" />
                  {t('editor.tools.crop')}
                </button>
                
                <button
                  onClick={handleReset}
                  className="flex items-center px-3 py-2 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors flex-shrink-0"
                  title={t('editor.tools.reset')}
                >
                  <RefreshCw className="w-4 h-4 mr-2" />
                  {t('editor.tools.reset')}
                </button>
                
                <button
                  onClick={handleReplace}
                  className="flex items-center px-3 py-2 bg-red-50 border border-red-300 text-red-700 rounded-lg hover:bg-red-100 transition-colors flex-shrink-0"
                  title={t('editor.tools.replace')}
                >
                  <X className="w-4 h-4 mr-2" />
                  {t('editor.tools.replace')}
                </button>
              </div>
            </div>

            
            <div className="flex justify-center">
              <div className="relative overflow-hidden rounded-lg border border-gray-300">
                <img
                  src={preview}
                  alt="Preview"
                  className="max-w-full max-h-96 object-contain"
                  style={{
                    transform: `scale(${scale}) rotate(${rotation}deg)`,
                    transition: 'transform 0.3s ease'
                  }}
                />
                
                {cropData && (
                  <div
                    className="absolute border-2 border-brand-primary bg-brand-primary bg-opacity-20"
                    style={{
                      left: `${cropData.x}%`,
                      top: `${cropData.y}%`,
                      width: `${cropData.width}%`,
                      height: `${cropData.height}%`
                    }}
                  />
                )}
              </div>
            </div>

            
            <div className="bg-gray-50 rounded-lg p-4">
              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-3">
                  <div className="w-10 h-10 bg-brand-primary/10 rounded-lg flex items-center justify-center">
                    <Upload className="w-5 h-5 text-brand-primary" />
                  </div>
                  <div>
                    <p className="font-medium text-gray-900">{selectedFile?.name}</p>
                    <p className="text-sm text-gray-500">
                      {formatBytes(selectedFile?.size)}
                    </p>
                  </div>
                </div>
                
                <div className="flex gap-3">
                  <button
                    onClick={applyEdits}
                    disabled={!imageData?.image_id}
                    className="px-6 py-2 border border-gray-300 rounded-lg hover:bg-gray-50"
                  >
                    {t('common.save')}
                  </button>
                  <button
                    onClick={handleUpload}
                    disabled={uploadImageMutation.isPending || !couponCode}
                    className="px-6 py-2 bg-brand-primary text-white rounded-lg hover:bg-brand-primary/90 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                  >
                    {uploadImageMutation.isPending ? (
                      <div className="w-5 h-5 border-2 border-white border-t-transparent rounded-full animate-spin" />
                    ) : (
                      t('editor.upload.upload_button')
                    )}
                  </button>
                </div>
              </div>
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  )
}

export default ImageUploader
