import React, { useState, useRef } from 'react'
import { useTranslation } from 'react-i18next'
import { motion } from 'framer-motion'
import { Upload, Palette, CheckCircle, X } from 'lucide-react'
import { useMutation } from '@tanstack/react-query'
import ImageEditor from '../ImageEditor'
import StyleSelector from './StyleSelector'
import SchemaGenerator from './SchemaGenerator'
import { MosaicAPI } from '../../api/client'
import { useUIStore } from '../../store/partnerStore'

const EditorSteps = ({ couponCode, couponSize, initialImageId = null, initialStep = 1 }) => {
  const { t } = useTranslation()
  const { addNotification } = useUIStore()
  const [currentStep, setCurrentStep] = useState(Math.min(Math.max(initialStep, 1), 3))
  const [imageData, setImageData] = useState(initialImageId ? { id: initialImageId, image_id: initialImageId } : null)
  const [selectedOptions, setSelectedOptions] = useState(null)
  const [isStepLocked, setIsStepLocked] = useState(false)
  
    const [selectedFile, setSelectedFile] = useState(null)
  const [previewUrl, setPreviewUrl] = useState(null)
  const [showEditor, setShowEditor] = useState(false)
  const [dragActive, setDragActive] = useState(false)
  const fileInputRef = useRef(null)
  
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
      handleImageUploaded(data)
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
        : t('notifications.upload_failed')

      addNotification({
        type: 'error',
        title: t('notifications.upload_error'),
        message: msg
      })
    }
  })

    React.useEffect(() => {
    const imageId = initialImageId || imageData?.image_id || imageData?.id
    if (!imageId) return
    try {
      const stored = sessionStorage.getItem(`editor:selectedOptions:${imageId}`)
      if (stored) {
        setSelectedOptions(JSON.parse(stored))
      }
      const storedStep = sessionStorage.getItem(`editor:step:${imageId}`)
      if (storedStep) {
        const stepNum = parseInt(storedStep, 10)
        if (!Number.isNaN(stepNum)) {
          setCurrentStep(Math.min(Math.max(stepNum, 1), 3))
        }
      }
    } catch (e) {
          }
  }, [initialImageId])

    React.useEffect(() => {
    if (couponCode) {
            setImageData(null)
      setSelectedOptions(null)
      setCurrentStep(1)
      setIsStepLocked(false)
      setSelectedFile(null)
      setPreviewUrl(null)
      setShowEditor(false)
    }
  }, [couponCode])

    const handleFileSelect = (file) => {
    if (!file) return
    
        if (!file.type.startsWith('image/')) {
      addNotification({
        type: 'error',
        title: t('notifications.upload_error'),
        message: t('notifications.invalid_file_type')
      })
      return
    }

    setSelectedFile(file)
    const url = URL.createObjectURL(file)
    setPreviewUrl(url)
    setShowEditor(true)
  }

  const handleDrop = (e) => {
    e.preventDefault()
    setDragActive(false)
    
    const files = Array.from(e.dataTransfer.files)
    if (files.length > 0) {
      handleFileSelect(files[0])
    }
  }

  const handleDrag = (e) => {
    e.preventDefault()
    setDragActive(e.type === 'dragenter' || e.type === 'dragover')
  }

  const handleEditorSave = async (editedImageBlob, fileName, cropData, position, rotate) => {
    if (!selectedFile || !editedImageBlob) return

    const formData = new FormData()
    formData.append('image', editedImageBlob, fileName)
    formData.append('coupon_code', couponCode)
    formData.append('coupon_size', couponSize)
    
        if (cropData) {
      formData.append('crop_x', Math.round(cropData.x))
      formData.append('crop_y', Math.round(cropData.y))
      formData.append('crop_width', Math.round(cropData.width))
      formData.append('crop_height', Math.round(cropData.height))
    }
    if (position) {
      formData.append('position_x', position.x)
      formData.append('position_y', position.y)
    }
    if (rotate) {
      formData.append('rotation', rotate)
    }

    uploadImageMutation.mutate(formData)
  }

  const handleRemoveImage = () => {
    setSelectedFile(null)
    setPreviewUrl(null)
    setShowEditor(false)
    setImageData(null)
    if (previewUrl) {
      URL.revokeObjectURL(previewUrl)
    }
  }

  const goToStep = (stepNumber) => {
    const bounded = Math.min(Math.max(stepNumber, 1), steps.length || 3)
    setCurrentStep(bounded)
    const params = new URLSearchParams(window.location.search)
    params.set('step', String(bounded))
    window.history.replaceState({}, '', `${window.location.pathname}?${params.toString()}`)
    try {
      const imageId = imageData?.image_id || imageData?.id || initialImageId
      if (imageId) {
        sessionStorage.setItem(`editor:step:${imageId}`, String(bounded))
      }
    } catch {}
  }

  const handleImageUploaded = (data) => {
    setImageData(data)
        const params = new URLSearchParams(window.location.search)
    if (data?.image_id || data?.id) params.set('image', data.image_id || data.id)
    window.history.replaceState({}, '', `${window.location.pathname}?${params.toString()}`)
    goToStep(2)
  }

  const handleStyleSelected = (options) => {
    setSelectedOptions(options)
    try {
      const imageId = imageData?.image_id || imageData?.id
      if (imageId) {
        sessionStorage.setItem(`editor:selectedOptions:${imageId}`, JSON.stringify(options))
      }
    } catch {}
    goToStep(3)
  }

  const handleSchemaComplete = (schemaData) => {
        console.log('Schema completed:', schemaData)
  }

  const nextStep = () => {
    if (currentStep < steps.length) {
      goToStep(currentStep + 1)
    }
  }

  const prevStep = () => {
    if (currentStep > 1) {
      goToStep(currentStep - 1)
    }
  }

  const canGoNext = () => {
    switch (currentStep) {
      case 1:
        return !!imageData
      case 2:
        return !!selectedOptions
      case 3:
        return false
      default:
        return false
    }
  }

  const canGoPrev = () => {
    return currentStep > 1 && !isStepLocked
  }

  const steps = [
    {
      id: 1,
      name: t('editor.steps.upload'),
      icon: Upload,
      component: (
        !showEditor ? (
                    <div
            className={`border-2 border-dashed rounded-lg p-8 text-center transition-colors ${
              dragActive ? 'border-brand-primary bg-brand-primary/5' : 'border-gray-300 hover:border-gray-400'
            }`}
            onDrop={handleDrop}
            onDragOver={handleDrag}
            onDragEnter={handleDrag}
            onDragLeave={handleDrag}
          >
            <input
              ref={fileInputRef}
              type="file"
              className="hidden"
              accept="image/*"
              onChange={(e) => handleFileSelect(e.target.files?.[0])}
            />
            
            <Upload className="w-12 h-12 text-gray-400 mx-auto mb-4" />
            <h3 className="text-lg font-semibold text-gray-900 mb-2">
              {t('image_uploader.upload_image')}
            </h3>
            <p className="text-gray-600 mb-4">
              {t('image_uploader.drag_drop')}
            </p>
            
            <button
              type="button"
              onClick={() => fileInputRef.current?.click()}
              className="bg-brand-primary text-white px-6 py-2 rounded-lg hover:bg-brand-primary/90 transition-colors"
              disabled={uploadImageMutation.isPending}
            >
              {uploadImageMutation.isPending ? t('common.loading') : t('image_uploader.choose_file')}
            </button>
            
            <p className="text-sm text-gray-500 mt-4">
              {t('image_uploader.supported_formats')}
            </p>
          </div>
        ) : (
          
          <div className="space-y-4">
            <div className="flex justify-between items-center">
              <h3 className="text-lg font-semibold text-gray-900">
                {t('image_editor.edit_image')}
              </h3>
              <button
                onClick={handleRemoveImage}
                className="text-red-600 hover:text-red-800 flex items-center gap-2"
              >
                <X className="w-4 h-4" />
                {t('common.remove')}
              </button>
            </div>
            
            <ImageEditor
              imageUrl={previewUrl}
              fileName={selectedFile?.name}
              onSave={handleEditorSave}
            />
          </div>
        )
      )
    },
    {
      id: 2,
      name: t('editor.steps.styles'),
      icon: Palette,
      component: imageData ? (
        <StyleSelector
          imageId={imageData.image_id || imageData.id}
          initialOptions={selectedOptions || null}
          onStyleSelected={handleStyleSelected}
          onBack={() => goToStep(1)}
        />
      ) : (
        <div className="text-center py-12">
          <Palette className="w-16 h-16 text-gray-400 mx-auto mb-6" />
          <h3 className="text-2xl font-semibold text-gray-900 mb-4">
            {t('editor_steps.style_selection')}
          </h3>
          <p className="text-gray-600">
            {t('editor_steps.style_selection_desc')}
          </p>
        </div>
      )
    },
    {
      id: 3,
      name: t('editor.steps.confirm'),
      icon: CheckCircle,
      component: selectedOptions ? (
        <SchemaGenerator
          imageId={imageData?.image_id || imageData?.id}
          selectedOptions={selectedOptions}
          onBack={() => goToStep(2)}
          onComplete={handleSchemaComplete}
          onLockNavigation={setIsStepLocked}
        />
      ) : (
        <div className="text-center py-12">
          <CheckCircle className="w-16 h-16 text-gray-400 mx-auto mb-6" />
          <h3 className="text-2xl font-semibold text-gray-900 mb-4">
            {t('editor_steps.confirmation')}
          </h3>
          <p className="text-gray-600">
            {t('editor_steps.confirmation_desc')}
          </p>
        </div>
      )
    }
  ]

  return (
    <div className="space-y-8">
      
      <div className="flex items-center justify-between">
        {steps.map((step, index) => (
          <React.Fragment key={step.id}>
            <div className="flex items-center">
              <div
                className={`flex items-center justify-center w-12 h-12 rounded-full border-2 ${
                  currentStep >= step.id
                    ? 'bg-brand-primary border-brand-primary text-white'
                    : 'bg-white border-gray-300 text-gray-500'
                } transition-all duration-300`}
              >
                {currentStep > step.id ? (
                  <CheckCircle className="w-6 h-6" />
                ) : (
                  <step.icon className="w-6 h-6" />
                )}
              </div>
              <div className="ml-4 hidden sm:block">
                <p className={`text-sm font-medium ${
                  currentStep >= step.id ? 'text-brand-primary' : 'text-gray-500'
                }`}>
                  {t('editor_steps.step')} {step.id}
                </p>
                <p className={`text-sm ${
                  currentStep >= step.id ? 'text-gray-900' : 'text-gray-500'
                }`}>
                  {step.name}
                </p>
              </div>
            </div>
            
            {index < steps.length - 1 && (
              <div className={`flex-1 h-0.5 mx-4 ${
                currentStep > step.id ? 'bg-brand-primary' : 'bg-gray-300'
              } transition-all duration-300`} />
            )}
          </React.Fragment>
        ))}
      </div>

      
      <motion.div
        key={currentStep}
        initial={{ opacity: 0, x: 20 }}
        animate={{ opacity: 1, x: 0 }}
        exit={{ opacity: 0, x: -20 }}
        transition={{ duration: 0.3 }}
        className="min-h-96"
      >
        {steps[currentStep - 1].component}
      </motion.div>

      {}
    </div>
  )
}

export default EditorSteps
