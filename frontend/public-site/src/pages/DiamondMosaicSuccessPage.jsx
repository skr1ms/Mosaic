import React, { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { motion } from 'framer-motion'
import { CheckCircle, Download, ArrowRight, FileText, Image, Package, Home } from 'lucide-react'
import { useNavigate, useLocation } from 'react-router-dom'
import { useUIStore } from '../store/partnerStore'

const DiamondMosaicSuccessPage = () => {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const location = useLocation()
  const { addNotification } = useUIStore()
  
  const [orderData, setOrderData] = useState(null)
  const [isGenerating, setIsGenerating] = useState(true)
  const [downloadLinks, setDownloadLinks] = useState([])

  useEffect(() => {
    // Получаем данные заказа
    let orderInfo = null
    
    if (location.state?.orderData) {
      orderInfo = location.state.orderData
    } else {
      try {
        const savedOrder = localStorage.getItem('diamondMosaic_lastOrder')
        if (savedOrder) {
          orderInfo = JSON.parse(savedOrder)
        }
      } catch (error) {
        console.error('Error loading order data:', error)
      }
    }
    
    if (!orderInfo) {
      navigate('/diamond-mosaic')
      return
    }
    
    setOrderData(orderInfo)
    
    // Симулируем генерацию файлов
    generateSchemaFiles(orderInfo)
  }, [location, navigate])

  const generateSchemaFiles = async (order) => {
    setIsGenerating(true)
    
    try {
      // Симулируем процесс генерации файлов
      await new Promise(resolve => setTimeout(resolve, 3000))
      
      // Имитируем созданные файлы
      const files = [
        {
          id: 1,
          name: 'Схема алмазной мозаики',
          filename: `schema_${order.orderId}.pdf`,
          type: 'pdf',
          size: '2.5 MB',
          description: 'Подробная схема с номерами и цветами страз',
          url: '#', // В реальном приложении здесь будет реальная ссылка
          icon: <FileText className="w-5 h-5" />
        },
        {
          id: 2,
          name: 'Превью результата',
          filename: `preview_${order.orderId}.jpg`,
          type: 'image',
          size: '1.2 MB',
          description: 'Изображение готовой мозаики',
          url: order.imageData?.selectedPreview?.url || '#',
          icon: <Image className="w-5 h-5" />
        },
        {
          id: 3,
          name: 'Инструкция по сборке',
          filename: `instructions_${order.orderId}.pdf`,
          type: 'pdf',
          size: '800 KB',
          description: 'Пошаговое руководство по созданию мозаики',
          url: '#',
          icon: <FileText className="w-5 h-5" />
        }
      ]
      
      // Добавляем дополнительные файлы для премиум и профессионального пакетов
      if (order.package.id === 'premium' || order.package.id === 'professional') {
        files.push({
          id: 4,
          name: 'HD схема',
          filename: `schema_hd_${order.orderId}.pdf`,
          type: 'pdf',
          size: '5.1 MB',
          description: 'Схема в высоком разрешении',
          url: '#',
          icon: <FileText className="w-5 h-5" />
        })
      }
      
      if (order.package.id === 'professional') {
        files.push({
          id: 5,
          name: 'Файлы для печати',
          filename: `print_files_${order.orderId}.zip`,
          type: 'archive',
          size: '12.3 MB',
          description: 'Готовые файлы для печати схемы',
          url: '#',
          icon: <Package className="w-5 h-5" />
        })
      }
      
      setDownloadLinks(files)
      
    } catch (error) {
      console.error('Error generating schema files:', error)
      addNotification({
        type: 'error',
        message: 'Ошибка при генерации файлов схемы'
      })
    } finally {
      setIsGenerating(false)
    }
  }

  const handleDownload = (file) => {
    // В реальном приложении здесь будет загрузка файла
    addNotification({
      type: 'success',
      message: `Скачивание ${file.name} начато`
    })
  }

  const handleDownloadAll = () => {
    // Скачиваем все файлы
    downloadLinks.forEach(file => {
      setTimeout(() => handleDownload(file), Math.random() * 1000)
    })
  }

  const handleNewProject = () => {
    // Очищаем все данные и начинаем новый проект
    try {
      localStorage.removeItem('diamondMosaic_lastOrder')
      localStorage.removeItem('diamondMosaic_selectedImage')
      localStorage.removeItem('diamondMosaic_purchaseData')
      sessionStorage.removeItem('diamondMosaic_fileUrl')
    } catch (error) {
      console.error('Error clearing data:', error)
    }
    
    navigate('/diamond-mosaic')
  }

  const handleGoHome = () => {
    navigate('/')
  }

  if (!orderData) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-purple-50 via-pink-50 to-blue-50 flex items-center justify-center">
        <div className="text-purple-600">Загрузка...</div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-purple-50 via-pink-50 to-blue-50 py-8 px-4">
      <div className="container mx-auto max-w-4xl">
        
        {/* Заголовок успеха */}
        <motion.div
          initial={{ opacity: 0, y: -20 }}
          animate={{ opacity: 1, y: 0 }}
          className="text-center mb-8"
        >
          <motion.div
            initial={{ scale: 0 }}
            animate={{ scale: 1 }}
            transition={{ delay: 0.2, type: "spring", stiffness: 200 }}
            className="inline-block"
          >
            <CheckCircle className="w-16 h-16 text-green-500 mx-auto mb-4" />
          </motion.div>
          
          <h1 className="text-4xl md:text-5xl font-bold bg-gradient-to-r from-purple-600 to-pink-600 bg-clip-text text-transparent mb-4">
            Покупка завершена!
          </h1>
          <p className="text-lg text-gray-600">
            Заказ #{orderData.orderId} • {orderData.package.name} пакет
          </p>
        </motion.div>

        {/* Статус генерации */}
        {isGenerating && (
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            className="bg-white rounded-xl p-6 shadow-lg mb-8 text-center"
          >
            <div className="w-8 h-8 border-4 border-purple-600 border-t-transparent rounded-full animate-spin mx-auto mb-4"></div>
            <h3 className="text-lg font-semibold text-gray-800 mb-2">
              Генерируем ваши файлы...
            </h3>
            <p className="text-gray-600">
              Создаём схему алмазной мозаики и сопутствующие материалы
            </p>
          </motion.div>
        )}

        {/* Информация о заказе */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.3 }}
          className="bg-white rounded-xl p-6 shadow-lg mb-8"
        >
          <h2 className="text-2xl font-bold text-gray-800 mb-6">Детали заказа</h2>
          
          <div className="grid md:grid-cols-2 gap-6">
            {/* Превью */}
            <div>
              <h3 className="text-lg font-semibold text-gray-800 mb-3">Ваша мозаика</h3>
              <div className="aspect-square bg-gray-100 rounded-lg overflow-hidden mb-3">
                <img
                  src={orderData.imageData?.selectedPreview?.url || orderData.imageData?.originalImage}
                  alt="Preview"
                  className="w-full h-full object-cover"
                />
              </div>
              <p className="text-sm text-gray-600 text-center">
                {orderData.imageData?.selectedPreview?.title || 'Выбранное превью'}
              </p>
            </div>
            
            {/* Детали */}
            <div>
              <h3 className="text-lg font-semibold text-gray-800 mb-3">Характеристики</h3>
              <div className="space-y-3">
                <div className="flex justify-between py-2 border-b border-gray-100">
                  <span className="text-gray-600">Размер:</span>
                  <span className="font-medium">{orderData.imageData?.size} см</span>
                </div>
                <div className="flex justify-between py-2 border-b border-gray-100">
                  <span className="text-gray-600">Стиль:</span>
                  <span className="font-medium">{orderData.imageData?.style}</span>
                </div>
                <div className="flex justify-between py-2 border-b border-gray-100">
                  <span className="text-gray-600">Пакет:</span>
                  <span className="font-medium">{orderData.package.name}</span>
                </div>
                <div className="flex justify-between py-2 border-b border-gray-100">
                  <span className="text-gray-600">Стоимость:</span>
                  <span className="font-bold text-purple-600">{orderData.package.price}₽</span>
                </div>
                <div className="flex justify-between py-2">
                  <span className="text-gray-600">Дата покупки:</span>
                  <span className="font-medium">
                    {new Date(orderData.purchaseDate).toLocaleDateString('ru-RU')}
                  </span>
                </div>
              </div>
            </div>
          </div>
        </motion.div>

        {/* Файлы для скачивания */}
        {!isGenerating && downloadLinks.length > 0 && (
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.5 }}
            className="bg-white rounded-xl p-6 shadow-lg mb-8"
          >
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-2xl font-bold text-gray-800">Ваши файлы</h2>
              <button
                onClick={handleDownloadAll}
                className="bg-purple-600 text-white px-4 py-2 rounded-lg hover:bg-purple-700 transition-colors flex items-center"
              >
                <Download className="w-4 h-4 mr-2" />
                Скачать всё
              </button>
            </div>
            
            <div className="space-y-4">
              {downloadLinks.map((file, index) => (
                <motion.div
                  key={file.id}
                  initial={{ opacity: 0, x: -20 }}
                  animate={{ opacity: 1, x: 0 }}
                  transition={{ delay: 0.6 + index * 0.1 }}
                  className="flex items-center justify-between p-4 border border-gray-200 rounded-lg hover:border-purple-300 transition-colors"
                >
                  <div className="flex items-center">
                    <div className="w-10 h-10 bg-purple-100 rounded-lg flex items-center justify-center mr-4">
                      {file.icon}
                    </div>
                    <div>
                      <h4 className="font-semibold text-gray-800">{file.name}</h4>
                      <p className="text-sm text-gray-600">{file.description}</p>
                      <p className="text-xs text-gray-500">{file.filename} • {file.size}</p>
                    </div>
                  </div>
                  <button
                    onClick={() => handleDownload(file)}
                    className="bg-gray-100 text-gray-700 px-4 py-2 rounded-lg hover:bg-gray-200 transition-colors flex items-center"
                  >
                    <Download className="w-4 h-4 mr-1" />
                    Скачать
                  </button>
                </motion.div>
              ))}
            </div>
          </motion.div>
        )}

        {/* Действия */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.7 }}
          className="flex flex-col sm:flex-row gap-4 justify-center"
        >
          <button
            onClick={handleNewProject}
            className="bg-gradient-to-r from-purple-600 to-pink-600 text-white px-8 py-4 rounded-xl font-semibold text-lg hover:from-purple-700 hover:to-pink-700 transition-all duration-300 shadow-lg hover:shadow-xl flex items-center justify-center"
          >
            Создать новую мозаику
            <ArrowRight className="w-5 h-5 ml-2" />
          </button>
          
          <button
            onClick={handleGoHome}
            className="bg-white text-purple-600 border-2 border-purple-600 px-8 py-4 rounded-xl font-semibold text-lg hover:bg-purple-50 transition-all duration-300 flex items-center justify-center"
          >
            <Home className="w-5 h-5 mr-2" />
            На главную
          </button>
        </motion.div>

        {/* Дополнительная информация */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.9 }}
          className="mt-8 text-center text-gray-600"
        >
          <p className="mb-2">
            Файлы также были отправлены на ваш email (если указан)
          </p>
          <p className="text-sm">
            Если у вас есть вопросы, обратитесь в службу поддержки
          </p>
        </motion.div>
      </div>
    </div>
  )
}

export default DiamondMosaicSuccessPage
