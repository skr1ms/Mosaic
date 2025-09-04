import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import navigationService from '../services/navigationService'

export function useNavigation() {
  const navigate = useNavigate()
  const [state, setState] = useState(navigationService.getState())

  useEffect(() => {
    // Subscribe to state changes
    const unsubscribe = navigationService.subscribe((newState) => {
      setState(newState)
    })

    return unsubscribe
  }, [])

  // Navigation helpers
  const navigateToStep = (step) => {
    const stepRoutes = {
      'size-selection': '/preview',
      'style-selection': '/preview/styles',
      'preview-album': '/preview/album',
      'purchase': '/preview/purchase',
      'success': '/preview/success'
    }

    if (navigationService.goToStep(step)) {
      navigate(stepRoutes[step] || '/preview')
    }
  }

  const goBack = () => {
    if (navigationService.goBack()) {
      const stepRoutes = {
        'size-selection': '/preview',
        'style-selection': '/preview/styles',
        'preview-album': '/preview/album',
        'purchase': '/preview/purchase'
      }
      navigate(stepRoutes[navigationService.getState().currentStep] || '/preview')
    }
  }

  const goToNextStep = () => {
    const currentStep = state.currentStep
    const nextSteps = {
      'size-selection': 'style-selection',
      'style-selection': 'preview-album',
      'preview-album': 'purchase',
      'purchase': 'success'
    }

    const nextStep = nextSteps[currentStep]
    if (nextStep) {
      navigateToStep(nextStep)
    }
  }

  // Data setters
  const selectSize = (size) => {
    navigationService.selectSize(size)
  }

  const uploadImage = (file, previewUrl) => {
    navigationService.uploadImage(file, previewUrl)
  }

  const updateEditedImage = (editedUrl, editorParams) => {
    navigationService.updateEditedImage(editedUrl, editorParams)
  }

  const selectStyle = (style) => {
    navigationService.selectStyle(style)
  }

  const selectPreview = (previewIndex) => {
    navigationService.selectPreview(previewIndex)
  }

  const setPreviews = (previews) => {
    navigationService.setPreviews(previews)
  }

  const addAIPreviews = (aiPreviews) => {
    navigationService.addAIPreviews(aiPreviews)
  }

  const toggleAI = (enabled) => {
    navigationService.toggleAI(enabled)
  }

  const setPurchaseData = (data) => {
    navigationService.setPurchaseData(data)
  }

  const reset = () => {
    navigationService.reset()
    navigate('/preview')
  }

  // Validation
  const canProceed = () => {
    switch (state.currentStep) {
      case 'size-selection':
        return navigationService.canProceedToStyles()
      case 'style-selection':
        return navigationService.canProceedToAlbum()
      case 'preview-album':
        return navigationService.canProceedToPurchase()
      default:
        return false
    }
  }

  return {
    // State
    state,
    currentStep: state.currentStep,
    selectedSize: state.selectedSize,
    selectedFile: state.selectedFile,
    selectedStyle: state.selectedStyle,
    selectedPreview: state.selectedPreview,
    imageData: state.imageData,
    editorSettings: state.editorSettings,
    previews: state.previews,
    aiPreviews: state.aiPreviews,
    useAI: state.useAI,
    purchaseData: state.purchaseData,
    
    // Navigation
    navigateToStep,
    goBack,
    goToNextStep,
    
    // Actions
    selectSize,
    uploadImage,
    updateEditedImage,
    selectStyle,
    selectPreview,
    setPreviews,
    addAIPreviews,
    toggleAI,
    setPurchaseData,
    reset,
    
    // Validation
    canProceed
  }
}