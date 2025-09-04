import React, { useState, useEffect, useRef } from 'react'
import { useTranslation } from 'react-i18next'
import { motion } from 'framer-motion'
import { 
  Upload, 
  Check, 
  ArrowRight, 
  Download, 
  Mail,
  Search,
  Calendar,
  Package,
  FileText,
  Lock,
  Unlock
} from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import { useUIStore } from '../store/partnerStore'
import { MosaicAPI } from '../api/client'
import ImageEditor from '../components/ImageEditor'

const CouponActivationPage = () => {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { addNotification } = useUIStore()
  const fileInputRef = useRef(null)
  
  // Coupon state
  const [couponCode, setCouponCode] = useState('')
  const [couponData, setCouponData] = useState(null)
  const [isValidating, setIsValidating] = useState(false)
  
  // Image state
  const [selectedFile, setSelectedFile] = useState(null)
  const [previewUrl, setPreviewUrl] = useState(null)
  const [editedImageUrl, setEditedImageUrl] = useState(null)
  const [showEditor, setShowEditor] = useState(false)
  
  // Activation state
  const [isActivating, setIsActivating] = useState(false)
  const [activationStep, setActivationStep] = useState('input') // input, upload, edit, preview, success
  
  // For activated coupons
  const [searchPageNumber, setSearchPageNumber] = useState('')
  const [resendEmail, setResendEmail] = useState('')
  const [isSendingEmail, setIsSendingEmail] = useState(false)

  useEffect(() => {
    // Clear any previous session data
    try {
      localStorage.removeItem('activeCoupon')
      localStorage.removeItem('couponActivation_image')
      sessionStorage.removeItem('couponActivation_fileUrl')
    } catch (error) {
      console.error('Error clearing session data:', error)
    }
  }, [])

  const validateCoupon = async () => {
    if (!couponCode.trim()) {
      addNotification({
        type: 'error',
        message: 'Пожалуйста, введите код купона'
      })
      return
    }

    setIsValidating(true)
    try {
      const response = await MosaicAPI.validateCoupon(couponCode.trim())
      
      if (response.valid === false) {
        addNotification({
          type: 'error',
          message: 'Купон не найден или недействителен'
        })
        return
      }

      setCouponData(response.coupon || response)
      
      // Store coupon in localStorage for session persistence
      localStorage.setItem('activeCoupon', JSON.stringify({
        code: couponCode,
        data: response.coupon || response,
        timestamp: Date.now()
      }))
      
      // Check coupon status
      if (response.coupon?.status === 'activated' || response.status === 'activated') {
        setActivationStep('activated')
      } else {
        setActivationStep('upload')
      }
      
    } catch (error) {
      console.error('Error validating coupon:', error)
      addNotification({
        type: 'error',
        message: error.message || 'Ошибка при проверке купона'
      })
    } finally {
      setIsValidating(false)
    }
  }

  const handleFileSelect = (event) => {
    const file = event.target.files[0]
    if (!file) return

    if (!file.type.startsWith('image/')) {
      addNotification({
        type: 'error',
        message: 'Пожалуйста, выберите изображение'
      })
      return
    }

    if (file.size > 10 * 1024 * 1024) {
      addNotification({
        type: 'error',
        message: 'Размер файла не должен превышать 10 МБ'
      })
      return
    }

    setSelectedFile(file)
    
    const reader = new FileReader()
    reader.onload = (e) => {
      setPreviewUrl(e.target.result)
      setShowEditor(true)
      setActivationStep('edit')
    }
    reader.readAsDataURL(file)
  }

  const handleEditorSave = (editedUrl, editorParams) => {
    setEditedImageUrl(editedUrl)
    setShowEditor(false)
    
    // Save edited image data
    localStorage.setItem('couponActivation_image', JSON.stringify({
      fileName: selectedFile.name,
      previewUrl: editedUrl,
      editorParams,
      timestamp: Date.now()
    }))
    
    // Proceed to preview generation
    navigateToPreviewGeneration()
  }
  
  const handleEditorCancel = () => {
    setShowEditor(false)
    setSelectedFile(null)
    setPreviewUrl(null)
    setEditedImageUrl(null)
    setActivationStep('upload')
  }

  const navigateToPreviewGeneration = () => {
    // Store all necessary data for preview generation
    const finalImageUrl = editedImageUrl || previewUrl
    
    localStorage.setItem('diamondMosaic_selectedImage', JSON.stringify({
      size: couponData.size,
      fileName: selectedFile.name,
      previewUrl: finalImageUrl,
      timestamp: Date.now(),
      hasEdits: editedImageUrl !== null,
      couponId: couponData.id,
      couponCode: couponCode
    }))
    
    // Navigate to preview album page for style selection
    navigate('/preview/album')
  }

  const handleSearchPage = () => {
    if (!searchPageNumber || !couponData?.zip_url) return
    
    // Open specific page in archive (implementation depends on archive structure)
    const pageUrl = `${couponData.zip_url}#page=${searchPageNumber}`
    window.open(pageUrl, '_blank')
  }

  const handleResendEmail = async () => {
    if (!resendEmail || !couponData?.id) {
      addNotification({
        type: 'error',
        message: 'Пожалуйста, введите email'
      })
      return
    }

    setIsSendingEmail(true)
    try {
      await MosaicAPI.sendSchemaToEmail(couponData.id, { email: resendEmail })
      addNotification({
        type: 'success',
        message: `Архив отправлен на ${resendEmail}`
      })
      setResendEmail('')
    } catch (error) {
      console.error('Error sending email:', error)
      addNotification({
        type: 'error',
        message: 'Ошибка при отправке email'
      })
    } finally {
      setIsSendingEmail(false)
    }
  }

  const handleDownloadArchive = () => {
    if (couponData?.zip_url) {
      window.open(couponData.zip_url, '_blank')
    }
  }

  const renderCouponInput = () => (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className="max-w-md mx-auto"
    >
      <div className="bg-white rounded-2xl shadow-lg p-8">
        <h2 className="text-2xl font-bold text-gray-900 mb-6 text-center">
          Введите код купона
        </h2>
        
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Код купона
            </label>
            <input
              type="text"
              value={couponCode}
              onChange={(e) => setCouponCode(e.target.value.toUpperCase())}
              placeholder="Например: ABC123"
              className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-transparent"
              onKeyPress={(e) => e.key === 'Enter' && validateCoupon()}
            />
          </div>
          
          <button
            onClick={validateCoupon}
            disabled={isValidating || !couponCode.trim()}
            className="w-full bg-gradient-to-r from-purple-600 to-pink-600 text-white py-3 rounded-lg font-semibold hover:from-purple-700 hover:to-pink-700 transition-all disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {isValidating ? 'Проверка...' : 'Проверить купон'}
          </button>
        </div>
        
        <div className="mt-6 text-center text-sm text-gray-600">
          <p>Купон можно найти на упаковке набора</p>
        </div>
      </div>
    </motion.div>
  )

  const renderImageUpload = () => (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className="max-w-2xl mx-auto"
    >
      <div className="bg-white rounded-2xl shadow-lg p-8">
        <div className="mb-6">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-2xl font-bold text-gray-900">
              Загрузите фотографию
            </h2>
            <div className="flex items-center text-green-600">
              <Check className="w-5 h-5 mr-2" />
              <span className="font-medium">Купон: {couponCode}</span>
            </div>
          </div>
          
          <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
            <div className="flex items-start">
              <Package className="w-5 h-5 text-blue-600 mt-1 mr-3 flex-shrink-0" />
              <div>
                <p className="text-sm font-medium text-blue-900">
                  Размер мозаики: {couponData?.size || 'Стандартный'}
                </p>
                <p className="text-sm text-blue-700 mt-1">
                  Стиль: {couponData?.style || 'Классический'}
                </p>
              </div>
            </div>
          </div>
        </div>
        
        <div className="border-2 border-dashed border-gray-300 rounded-xl p-12 text-center hover:border-purple-400 transition-colors">
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
            <Upload className="w-16 h-16 text-gray-400 mx-auto mb-4" />
            <p className="text-xl font-medium text-gray-700 mb-2">
              Выберите изображение для мозаики
            </p>
            <p className="text-gray-500 mb-4">
              JPG, PNG до 10 МБ
            </p>
            <div className="inline-block bg-purple-600 text-white px-6 py-3 rounded-lg hover:bg-purple-700 transition-colors font-semibold">
              Выбрать файл
            </div>
          </label>
        </div>
      </div>
    </motion.div>
  )

  const renderImageEditor = () => (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className="max-w-6xl mx-auto"
    >
      <div className="mb-4">
        <div className="flex items-center justify-between bg-white rounded-lg px-4 py-2 shadow">
          <h2 className="text-xl font-semibold text-gray-900">
            Шаг 2: Редактирование изображения
          </h2>
          <div className="flex items-center text-green-600">
            <Check className="w-5 h-5 mr-2" />
            <span className="font-medium">Купон: {couponCode}</span>
          </div>
        </div>
      </div>
      
      <ImageEditor
        imageUrl={previewUrl}
        onSave={handleEditorSave}
        onCancel={handleEditorCancel}
        title="Настройте изображение для мозаики"
        showCropHint={true}
        aspectRatio={1}
        fileName={selectedFile?.name}
      />
    </motion.div>
  )

  const renderActivatedCoupon = () => (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className="max-w-4xl mx-auto"
    >
      <div className="bg-white rounded-2xl shadow-lg overflow-hidden">
        {/* Header */}
        <div className="bg-gradient-to-r from-green-500 to-emerald-600 p-6 text-white">
          <div className="flex items-center justify-between">
            <div>
              <h2 className="text-2xl font-bold mb-2">
                Купон активирован
              </h2>
              <p className="text-green-100">
                Код: {couponCode}
              </p>
            </div>
            <div className="text-right">
              <Lock className="w-12 h-12 text-white/20 mb-2" />
              <p className="text-sm text-green-100">
                {couponData?.activated_at ? 
                  new Date(couponData.activated_at).toLocaleDateString('ru-RU') : 
                  'Дата активации неизвестна'}
              </p>
            </div>
          </div>
        </div>
        
        {/* Content */}
        <div className="p-6">
          {/* Preview Image */}
          {couponData?.preview_image_url && (
            <div className="mb-6">
              <h3 className="text-lg font-semibold text-gray-900 mb-3">
                Превью вашей мозаики
              </h3>
              <div className="bg-gray-50 rounded-lg p-4">
                <img 
                  src={couponData.preview_image_url} 
                  alt="Превью мозаики"
                  className="w-full max-w-md mx-auto rounded-lg shadow"
                />
              </div>
            </div>
          )}
          
          {/* Statistics */}
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
            <div className="bg-purple-50 rounded-lg p-4">
              <div className="flex items-center">
                <Package className="w-8 h-8 text-purple-600 mr-3" />
                <div>
                  <p className="text-sm text-gray-600">Размер</p>
                  <p className="text-lg font-semibold text-gray-900">
                    {couponData?.size || 'Не указан'}
                  </p>
                </div>
              </div>
            </div>
            
            <div className="bg-blue-50 rounded-lg p-4">
              <div className="flex items-center">
                <FileText className="w-8 h-8 text-blue-600 mr-3" />
                <div>
                  <p className="text-sm text-gray-600">Страниц в схеме</p>
                  <p className="text-lg font-semibold text-gray-900">
                    {couponData?.page_count || 'Не указано'}
                  </p>
                </div>
              </div>
            </div>
            
            <div className="bg-green-50 rounded-lg p-4">
              <div className="flex items-center">
                <Calendar className="w-8 h-8 text-green-600 mr-3" />
                <div>
                  <p className="text-sm text-gray-600">Камней</p>
                  <p className="text-lg font-semibold text-gray-900">
                    {couponData?.stones_count || 'Не указано'}
                  </p>
                </div>
              </div>
            </div>
          </div>
          
          {/* Actions */}
          <div className="space-y-4">
            {/* Download Archive */}
            {couponData?.zip_url && (
              <div className="border border-gray-200 rounded-lg p-4">
                <h4 className="font-semibold text-gray-900 mb-3">
                  Скачать архив со схемой
                </h4>
                <button
                  onClick={handleDownloadArchive}
                  className="flex items-center bg-purple-600 text-white px-6 py-3 rounded-lg hover:bg-purple-700 transition-colors"
                >
                  <Download className="w-5 h-5 mr-2" />
                  Скачать архив
                </button>
              </div>
            )}
            
            {/* Send to Email */}
            <div className="border border-gray-200 rounded-lg p-4">
              <h4 className="font-semibold text-gray-900 mb-3">
                Отправить на email
              </h4>
              <div className="flex gap-2">
                <input
                  type="email"
                  value={resendEmail}
                  onChange={(e) => setResendEmail(e.target.value)}
                  placeholder="example@email.com"
                  className="flex-1 px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-transparent"
                />
                <button
                  onClick={handleResendEmail}
                  disabled={isSendingEmail || !resendEmail}
                  className="flex items-center bg-blue-600 text-white px-6 py-2 rounded-lg hover:bg-blue-700 transition-colors disabled:opacity-50"
                >
                  <Mail className="w-5 h-5 mr-2" />
                  {isSendingEmail ? 'Отправка...' : 'Отправить'}
                </button>
              </div>
            </div>
            
            {/* Page Search */}
            {couponData?.page_count > 0 && (
              <div className="border border-gray-200 rounded-lg p-4">
                <h4 className="font-semibold text-gray-900 mb-3">
                  Поиск страницы в архиве
                </h4>
                <div className="flex gap-2">
                  <input
                    type="number"
                    value={searchPageNumber}
                    onChange={(e) => setSearchPageNumber(e.target.value)}
                    placeholder="Номер страницы"
                    min="1"
                    max={couponData.page_count}
                    className="flex-1 px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-transparent"
                  />
                  <button
                    onClick={handleSearchPage}
                    disabled={!searchPageNumber}
                    className="flex items-center bg-green-600 text-white px-6 py-2 rounded-lg hover:bg-green-700 transition-colors disabled:opacity-50"
                  >
                    <Search className="w-5 h-5 mr-2" />
                    Найти страницу
                  </button>
                </div>
                <p className="text-sm text-gray-600 mt-2">
                  Всего страниц: {couponData.page_count}
                </p>
              </div>
            )}
          </div>
        </div>
      </div>
    </motion.div>
  )

  return (
    <div className="min-h-screen bg-gradient-to-br from-purple-50 via-pink-50 to-blue-50">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
        {/* Header */}
        <motion.div
          initial={{ opacity: 0, y: -20 }}
          animate={{ opacity: 1, y: 0 }}
          className="text-center mb-12"
        >
          <h1 className="text-4xl md:text-5xl font-bold text-gray-900 mb-4">
            Активация купона
          </h1>
          <p className="text-lg md:text-xl text-gray-600 max-w-2xl mx-auto">
            {activationStep === 'activated' 
              ? 'Ваш купон уже активирован. Вы можете скачать материалы.'
              : 'Введите код купона и загрузите фото для создания уникальной мозаики'}
          </p>
        </motion.div>

        {/* Content based on step */}
        {activationStep === 'input' && renderCouponInput()}
        {activationStep === 'upload' && renderImageUpload()}
        {activationStep === 'edit' && showEditor && renderImageEditor()}
        {activationStep === 'activated' && renderActivatedCoupon()}
      </div>
    </div>
  )
}

export default CouponActivationPage