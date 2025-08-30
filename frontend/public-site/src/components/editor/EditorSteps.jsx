import React, { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { motion } from 'framer-motion'
import { Upload, Palette, CheckCircle } from 'lucide-react'
import ImageUploader from './ImageUploader'
import StyleSelector from './StyleSelector'
import SchemaGenerator from './SchemaGenerator'

const EditorSteps = ({ couponCode, couponSize, initialImageId = null, initialStep = 1 }) => {
  const { t } = useTranslation()
  const [currentStep, setCurrentStep] = useState(Math.min(Math.max(initialStep, 1), 3))
  const [imageData, setImageData] = useState(initialImageId ? { id: initialImageId, image_id: initialImageId } : null)
  const [selectedOptions, setSelectedOptions] = useState(null)
  const [isStepLocked, setIsStepLocked] = useState(false)

  // Восстановление состояний из sessionStorage по imageId
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
      // ignore parse errors
    }
  }, [initialImageId])

  // Очищаем состояние при смене купона
  React.useEffect(() => {
    if (couponCode) {
      // Сбрасываем состояние при новом купоне
      setImageData(null)
      setSelectedOptions(null)
      setCurrentStep(1)
      setIsStepLocked(false)
    }
  }, [couponCode])

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
    // Сохраняем в URL для восстановления после перезагрузки
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
    // Схема создана успешно
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
        <ImageUploader
          onImageUploaded={handleImageUploaded}
          couponCode={couponCode}
          couponSize={couponSize}
          initialImageId={imageData?.image_id || imageData?.id || initialImageId || null}
        />
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
      {/* Progress Steps */}
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

      {/* Step Content */}
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

      {/* Navigation removed as per request */}
    </div>
  )
}

export default EditorSteps
