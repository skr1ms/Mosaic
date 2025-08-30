import React, { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { motion } from 'framer-motion'
import { Upload, Eye, Download, ArrowLeft, Loader2, AlertCircle } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import { useMutation } from '@tanstack/react-query'
import { MosaicAPI } from '../api/client'
import { useUIStore } from '../store/partnerStore'

const MosaicPreviewPage = () => {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { addNotification } = useUIStore()
  
  const [selectedFile, setSelectedFile] = useState(null)
  const [previewUrl, setPreviewUrl] = useState(null)
  const [selectedSize, setSelectedSize] = useState('30x40')
  const [selectedStyle, setSelectedStyle] = useState('max_colors')
  const [generatedPreview, setGeneratedPreview] = useState(null)

  const sizes = [
    { key: '21x30', title: '21×30 см', desc: 'Компактный размер для начинающих' },
    { key: '30x40', title: '30×40 см', desc: 'Популярный размер' },
    { key: '40x40', title: '40×40 см', desc: 'Квадратный формат' },
    { key: '40x50', title: '40×50 см', desc: 'Прямоугольный формат' },
    { key: '40x60', title: '40×60 см', desc: 'Широкий формат' },
    { key: '50x70', title: '50×70 см', desc: 'Максимальный размер' }
  ]

  const styles = [
    { key: 'grayscale', title: 'Черно-белый', desc: 'Классический, элегантный' },
    { key: 'skin_tones', title: 'Телесные тона', desc: 'Реалистичные оттенки кожи' },
    { key: 'pop_art', title: 'Поп-арт', desc: 'Яркие, контрастные цвета' },
    { key: 'max_colors', title: 'Максимум цветов', desc: 'Богатая цветовая палитра' }
  ]

  const generatePreviewMutation = useMutation({
    mutationFn: async ({ file, size, style }) => {
      const formData = new FormData()
      formData.append('image', file)
      formData.append('size', size)
      formData.append('style', style)
      
      const response = await MosaicAPI.generatePreview(formData)
      return response
    },
    onSuccess: (data) => {
      setGeneratedPreview(data.preview_url)
      addNotification({
        type: 'success',
        message: 'Превью мозаики успешно сгенерировано!'
      })
    },
    onError: (error) => {
      addNotification({
        type: 'error',
        title: 'Ошибка генерации',
        message: error?.message || 'Не удалось сгенерировать превью'
      })
    }
  })

  const handleFileSelect = (event) => {
    const file = event.target.files[0]
    if (!file) return

    // Проверяем тип файла
    if (!file.type.startsWith('image/')) {
      addNotification({
        type: 'error',
        message: 'Пожалуйста, выберите файл изображения'
      })
      return
    }

    // Проверяем размер файла (максимум 15MB)
    if (file.size > 15 * 1024 * 1024) {
      addNotification({
        type: 'error',
        message: 'Размер файла не должен превышать 15MB'
      })
      return
    }

    setSelectedFile(file)
    
    // Создаем превью для отображения
    const reader = new FileReader()
    reader.onload = (e) => {
      setPreviewUrl(e.target.result)
    }
    reader.readAsDataURL(file)
  }

  const handleGeneratePreview = () => {
    if (!selectedFile) {
      addNotification({
        type: 'error',
        message: 'Пожалуйста, выберите изображение'
      })
      return
    }

    generatePreviewMutation.mutate({
      file: selectedFile,
      size: selectedSize,
      style: selectedStyle
    })
  }

  const goBack = () => {
    navigate('/diamond-art')
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <div className="bg-white shadow-sm">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <div className="flex items-center space-x-4">
            <button
              onClick={goBack}
              className="inline-flex items-center text-gray-600 hover:text-gray-900 transition-colors"
            >
              <ArrowLeft className="w-5 h-5 mr-2" />
              Назад к алмазной мозаике
            </button>
            <div className="h-6 w-px bg-gray-300" />
            <h1 className="text-2xl font-bold text-gray-900">Превью мозаики</h1>
          </div>
        </div>
      </div>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="grid lg:grid-cols-2 gap-8">
          
          {/* Левая колонка - загрузка и настройки */}
          <motion.div
            initial={{ opacity: 0, x: -20 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ duration: 0.5 }}
            className="space-y-6"
          >
            {/* Загрузка изображения */}
            <div className="bg-white rounded-2xl shadow-lg p-6">
              <h2 className="text-xl font-bold text-gray-900 mb-4 flex items-center">
                <Upload className="w-6 h-6 mr-2 text-brand-primary" />
                Загрузите изображение
              </h2>
              
              <div className="border-2 border-dashed border-gray-300 rounded-xl p-8 text-center hover:border-brand-primary transition-colors">
                <input
                  type="file"
                  accept="image/*"
                  onChange={handleFileSelect}
                  className="hidden"
                  id="image-upload"
                />
                <label htmlFor="image-upload" className="cursor-pointer">
                  {previewUrl ? (
                    <div className="space-y-4">
                      <img
                        src={previewUrl}
                        alt="Превью загруженного изображения"
                        className="max-h-48 mx-auto rounded-lg shadow-md"
                      />
                      <p className="text-sm text-gray-600">
                        {selectedFile?.name} ({Math.round(selectedFile?.size / 1024)}KB)
                      </p>
                      <p className="text-xs text-brand-primary">Нажмите, чтобы выбрать другое изображение</p>
                    </div>
                  ) : (
                    <div className="space-y-4">
                      <Upload className="w-12 h-12 text-gray-400 mx-auto" />
                      <div>
                        <p className="text-lg font-medium text-gray-900">Выберите изображение</p>
                        <p className="text-sm text-gray-600 mt-1">
                          PNG, JPG до 15MB
                        </p>
                      </div>
                    </div>
                  )}
                </label>
              </div>
            </div>

            {/* Выбор размера */}
            <div className="bg-white rounded-2xl shadow-lg p-6">
              <h3 className="text-lg font-bold text-gray-900 mb-4">Размер мозаики</h3>
              <div className="grid grid-cols-2 gap-3">
                {sizes.map((size) => (
                  <button
                    key={size.key}
                    onClick={() => setSelectedSize(size.key)}
                    className={`p-3 rounded-lg border-2 transition-all ${
                      selectedSize === size.key
                        ? 'border-brand-primary bg-brand-primary/10 text-brand-primary'
                        : 'border-gray-200 hover:border-gray-300'
                    }`}
                  >
                    <div className="font-semibold">{size.title}</div>
                    <div className="text-sm text-gray-600">{size.desc}</div>
                  </button>
                ))}
              </div>
            </div>

            {/* Выбор стиля */}
            <div className="bg-white rounded-2xl shadow-lg p-6">
              <h3 className="text-lg font-bold text-gray-900 mb-4">Стиль мозаики</h3>
              <div className="space-y-3">
                {styles.map((style) => (
                  <button
                    key={style.key}
                    onClick={() => setSelectedStyle(style.key)}
                    className={`w-full p-3 rounded-lg border-2 text-left transition-all ${
                      selectedStyle === style.key
                        ? 'border-brand-secondary bg-brand-secondary/10 text-brand-secondary'
                        : 'border-gray-200 hover:border-gray-300'
                    }`}
                  >
                    <div className="font-semibold">{style.title}</div>
                    <div className="text-sm text-gray-600">{style.desc}</div>
                  </button>
                ))}
              </div>
            </div>

            {/* Кнопка генерации */}
            <button
              onClick={handleGeneratePreview}
              disabled={!selectedFile || generatePreviewMutation.isPending}
              className="w-full bg-gradient-to-r from-brand-primary to-brand-secondary text-white py-4 px-6 rounded-xl hover:from-brand-primary/90 hover:to-brand-secondary/90 disabled:opacity-50 disabled:cursor-not-allowed font-semibold text-lg transition-all duration-200 flex items-center justify-center space-x-2"
            >
              {generatePreviewMutation.isPending ? (
                <>
                  <Loader2 className="w-5 h-5 animate-spin" />
                  <span>Генерация превью...</span>
                </>
              ) : (
                <>
                  <Eye className="w-5 h-5" />
                  <span>Сгенерировать превью</span>
                </>
              )}
            </button>
          </motion.div>

          {/* Правая колонка - результат */}
          <motion.div
            initial={{ opacity: 0, x: 20 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ duration: 0.5, delay: 0.2 }}
            className="bg-white rounded-2xl shadow-lg p-6"
          >
            <h2 className="text-xl font-bold text-gray-900 mb-4 flex items-center">
              <Eye className="w-6 h-6 mr-2 text-brand-secondary" />
              Превью мозаики
            </h2>
            
            {generatedPreview ? (
              <div className="space-y-4">
                <img
                  src={generatedPreview}
                  alt="Превью мозаики"
                  className="w-full rounded-lg shadow-lg"
                  onError={(e) => {
                    console.error('Failed to load preview image:', generatedPreview)
                    e.target.style.display = 'none'
                  }}
                  onLoad={() => {
                    console.log('Preview image loaded successfully:', generatedPreview)
                  }}
                />

                <div className="flex space-x-3">
                  <button
                    onClick={() => {
                      console.log('Opening preview in new tab:', generatedPreview)
                      window.open(generatedPreview, '_blank', 'noopener,noreferrer')
                    }}
                    className="flex-1 bg-brand-primary text-white py-3 px-4 rounded-lg hover:bg-brand-primary/90 font-semibold transition-colors flex items-center justify-center space-x-2"
                  >
                    <Eye className="w-4 h-4" />
                    <span>Открыть в полном размере</span>
                  </button>
                  <button
                    onClick={() => {
                      // Простой способ скачивания через ссылку
                      const link = document.createElement('a')
                      link.href = generatedPreview
                      link.download = 'mosaic-preview.png'
                      link.target = '_blank'
                      document.body.appendChild(link)
                      link.click()
                      document.body.removeChild(link)
                    }}
                    className="flex-1 bg-brand-secondary text-white py-3 px-4 rounded-lg hover:bg-brand-secondary/90 font-semibold transition-colors flex items-center justify-center space-x-2"
                  >
                    <Download className="w-4 h-4" />
                    <span>Скачать</span>
                  </button>
                </div>
              </div>
            ) : (
              <div className="text-center py-12">
                <div className="w-24 h-24 bg-gray-100 rounded-full flex items-center justify-center mx-auto mb-4">
                  <Eye className="w-12 h-12 text-gray-400" />
                </div>
                <h3 className="text-lg font-medium text-gray-900 mb-2">Превью будет здесь</h3>
                <p className="text-gray-600">
                  Загрузите изображение и нажмите "Сгенерировать превью"
                </p>
              </div>
            )}
          </motion.div>
        </div>
      </div>
    </div>
  )
}

export default MosaicPreviewPage