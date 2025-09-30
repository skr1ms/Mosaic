import React, { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { motion } from 'framer-motion';
import {
  CheckCircle,
  Download,
  ArrowRight,
  FileText,
  Image,
  Package,
  Home,
} from 'lucide-react';
import { useNavigate, useLocation } from 'react-router-dom';
import { useUIStore } from '../store/partnerStore';

const PreviewSuccessPage = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const location = useLocation();
  const { addNotification } = useUIStore();

  const [orderData, setOrderData] = useState(null);
  const [couponData, setCouponData] = useState(null);
  const [isGenerating, setIsGenerating] = useState(true);
  const [downloadLinks, setDownloadLinks] = useState([]);

  useEffect(() => {
    const activatedCoupon = localStorage.getItem('activatedCoupon');
    if (activatedCoupon) {
      try {
        const couponInfo = JSON.parse(activatedCoupon);
        setCouponData(couponInfo);

        generateSchemaForCoupon(couponInfo);
        return;
      } catch (error) {
        console.error('Error loading coupon data:', error);
      }
    }

    let orderInfo = null;

    if (location.state?.orderData) {
      orderInfo = location.state.orderData;
    } else {
      try {
        const savedOrder = localStorage.getItem('diamondMosaic_lastOrder');
        if (savedOrder) {
          orderInfo = JSON.parse(savedOrder);
        }
      } catch (error) {
        console.error('Error loading order data:', error);
      }
    }

    if (!orderInfo) {
      navigate('/preview');
      return;
    }

    setOrderData(orderInfo);

    generateSchemaFiles(orderInfo);
  }, [location, navigate]);

  const generateSchemaForCoupon = async coupon => {
    setIsGenerating(true);

    try {
      await new Promise(resolve => setTimeout(resolve, 2000));

      const files = [
        {
          id: 1,
          name: 'Схема мозаики (ZIP архив)',
          filename: `mosaic_schema_${coupon.code}.zip`,
          type: 'zip',
          size: '15 MB',
          description: 'Полный архив со всеми страницами схемы',
          url: coupon.final_schema_url || '#',
          icon: <Package className="w-5 h-5" />,
        },
        {
          id: 2,
          name: 'Превью мозаики',
          filename: `preview_${coupon.code}.jpg`,
          type: 'image',
          size: '1.2 MB',
          description: 'Финальное превью вашей мозаики',
          url: coupon.preview_image_url || '#',
          icon: <Image className="w-5 h-5" />,
        },
        {
          id: 3,
          name: 'Инструкция по сборке',
          filename: `instructions_${coupon.code}.pdf`,
          type: 'pdf',
          size: '800 KB',
          description: 'Подробная инструкция по созданию мозаики',
          url: '#',
          icon: <FileText className="w-5 h-5" />,
        },
      ];

      setDownloadLinks(files);
      setIsGenerating(false);

      localStorage.removeItem('activatedCoupon');
    } catch (error) {
      console.error('Error generating schema for coupon:', error);
      addNotification({
        type: 'error',
        message: 'Ошибка при генерации схемы',
      });
      setIsGenerating(false);
    }
  };

  const generateSchemaFiles = async order => {
    setIsGenerating(true);

    try {
      await new Promise(resolve => setTimeout(resolve, 3000));

      const files = [
        {
          id: 1,
          name: t('diamond_mosaic_success.files.schema.name'),
          filename: `schema_${order.orderId}.pdf`,
          type: 'pdf',
          size: '2.5 MB',
          description: t('diamond_mosaic_success.files.schema.description'),
          url: '#',
          icon: <FileText className="w-5 h-5" />,
        },
        {
          id: 2,
          name: t('diamond_mosaic_success.preview_result'),
          filename: `preview_${order.orderId}.jpg`,
          type: 'image',
          size: '1.2 MB',
          description: t('diamond_mosaic_success.files.schema.description'),
          url: order.imageData?.selectedPreview?.url || '#',
          icon: <Image className="w-5 h-5" />,
        },
        {
          id: 3,
          name: t('diamond_mosaic_success.files.instructions.name'),
          filename: `instructions_${order.orderId}.pdf`,
          type: 'pdf',
          size: '800 KB',
          description: t(
            'diamond_mosaic_success.files.instructions.description'
          ),
          url: '#',
          icon: <FileText className="w-5 h-5" />,
        },
      ];

      if (
        order.package.id === 'premium' ||
        order.package.id === 'professional'
      ) {
        files.push({
          id: 4,
          name: t('diamond_mosaic_success.files.hd_schema.name'),
          filename: `schema_hd_${order.orderId}.pdf`,
          type: 'pdf',
          size: '5.1 MB',
          description: t('diamond_mosaic_success.files.hd_schema.description'),
          url: '#',
          icon: <FileText className="w-5 h-5" />,
        });
      }

      if (order.package.id === 'professional') {
        files.push({
          id: 5,
          name: t('diamond_mosaic_success.files.print_files.name'),
          filename: `print_files_${order.orderId}.zip`,
          type: 'archive',
          size: '12.3 MB',
          description: t(
            'diamond_mosaic_success.files.print_files.description'
          ),
          url: '#',
          icon: <Package className="w-5 h-5" />,
        });
      }

      setDownloadLinks(files);
    } catch (error) {
      console.error('Error generating schema files:', error);
      addNotification({
        type: 'error',
        message: t('diamond_mosaic_success.schema_generation_error'),
      });
    } finally {
      setIsGenerating(false);
    }
  };

  const handleDownload = file => {
    addNotification({
      type: 'success',
      message: `Загрузка начата: ${file.name}`,
    });
  };

  const handleDownloadAll = () => {
    downloadLinks.forEach(file => {
      setTimeout(() => handleDownload(file), Math.random() * 1000);
    });
  };

  const handleNewProject = () => {
    try {
      localStorage.removeItem('diamondMosaic_lastOrder');
      localStorage.removeItem('diamondMosaic_selectedImage');
      localStorage.removeItem('diamondMosaic_purchaseData');
      sessionStorage.removeItem('diamondMosaic_fileUrl');
    } catch (error) {
      console.error('Error clearing data:', error);
    }

    navigate('/preview');
  };

  const handleGoHome = () => {
    navigate('/');
  };

  if (!orderData) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-purple-50 via-pink-50 to-blue-50 flex items-center justify-center">
        <div className="text-purple-600">
          {t('diamond_mosaic_purchase.notifications.loading')}
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-purple-50 via-pink-50 to-blue-50 py-4 sm:py-6 lg:py-8 px-4 sm:px-6 lg:px-8">
      <div className="container mx-auto max-w-4xl">
        <motion.div
          initial={{ opacity: 0, y: -20 }}
          animate={{ opacity: 1, y: 0 }}
          className="text-center mb-6 sm:mb-8"
        >
          <motion.div
            initial={{ scale: 0 }}
            animate={{ scale: 1 }}
            transition={{ delay: 0.2, type: 'spring', stiffness: 200 }}
            className="inline-block"
          >
            <CheckCircle className="w-12 h-12 sm:w-16 sm:h-16 text-green-500 mx-auto mb-3 sm:mb-4" />
          </motion.div>

          <h1 className="text-2xl sm:text-3xl md:text-4xl lg:text-5xl font-bold bg-gradient-to-r from-purple-600 to-pink-600 bg-clip-text text-transparent mb-3 sm:mb-4 leading-tight">
            {couponData
              ? 'Купон успешно активирован!'
              : t('diamond_mosaic_success.title')}
          </h1>
          <p className="text-sm sm:text-base lg:text-lg text-gray-600 leading-relaxed px-4 sm:px-0">
            {couponData
              ? `Код купона: ${couponData.code} • Активирован: ${new Date(couponData.activatedAt).toLocaleDateString('ru-RU')}`
              : `Номер заказа: ${orderData?.orderId} • Пакет: ${orderData?.package?.name}`}
          </p>
        </motion.div>

        {}
        {isGenerating && (
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            className="bg-white rounded-xl p-6 shadow-lg mb-8 text-center"
          >
            <div className="w-8 h-8 border-4 border-purple-600 border-t-transparent rounded-full animate-spin mx-auto mb-4"></div>
            <h3 className="text-lg font-semibold text-gray-800 mb-2">
              {t('diamond_mosaic_success.generating_files')}
            </h3>
            <p className="text-gray-600">
              {t('diamond_mosaic_success.generating_description')}
            </p>
          </motion.div>
        )}

        {}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.3 }}
          className="bg-white rounded-xl p-4 sm:p-6 lg:p-8 shadow-lg mb-6 sm:mb-8"
        >
          <h2 className="text-xl sm:text-2xl font-bold text-gray-800 mb-4 sm:mb-6">
            {t('diamond_mosaic_success.order_details')}
          </h2>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 sm:gap-6">
            <div className="mb-6 md:mb-0">
              <h3 className="text-base sm:text-lg font-semibold text-gray-800 mb-3">
                {t('diamond_mosaic_success.your_mosaic')}
              </h3>
              <div className="aspect-square bg-gray-100 rounded-lg overflow-hidden mb-3 max-w-xs mx-auto md:max-w-none">
                <img
                  src={
                    orderData.imageData?.selectedPreview?.url ||
                    orderData.imageData?.originalImage
                  }
                  alt="Preview"
                  className="w-full h-full object-cover"
                />
              </div>
              <p className="text-xs sm:text-sm text-gray-600 text-center leading-relaxed">
                {orderData.imageData?.selectedPreview?.title ||
                  t('diamond_mosaic_success.selected_preview')}
              </p>
            </div>

            {}
            <div>
              <h3 className="text-base sm:text-lg font-semibold text-gray-800 mb-3">
                {t('diamond_mosaic_success.characteristics')}
              </h3>
              <div className="space-y-2 sm:space-y-3">
                <div className="flex flex-col sm:flex-row sm:justify-between py-2 border-b border-gray-100">
                  <span className="text-xs sm:text-sm text-gray-600 mb-1 sm:mb-0">
                    {t('diamond_mosaic_success.size_label')}
                  </span>
                  <span className="text-sm sm:text-base font-medium">
                    {orderData.imageData?.size} {t('common.cm')}
                  </span>
                </div>
                <div className="flex flex-col sm:flex-row sm:justify-between py-2 border-b border-gray-100">
                  <span className="text-xs sm:text-sm text-gray-600 mb-1 sm:mb-0">
                    {t('diamond_mosaic_success.style_label')}
                  </span>
                  <span className="text-sm sm:text-base font-medium">
                    {orderData.imageData?.style}
                  </span>
                </div>
                <div className="flex flex-col sm:flex-row sm:justify-between py-2 border-b border-gray-100">
                  <span className="text-xs sm:text-sm text-gray-600 mb-1 sm:mb-0">
                    {t('diamond_mosaic_success.package')}
                  </span>
                  <span className="text-sm sm:text-base font-medium">
                    {orderData.package.name}
                  </span>
                </div>
                <div className="flex flex-col sm:flex-row sm:justify-between py-2 border-b border-gray-100">
                  <span className="text-xs sm:text-sm text-gray-600 mb-1 sm:mb-0">
                    {t('diamond_mosaic_success.cost')}
                  </span>
                  <span className="text-base sm:text-lg font-bold text-purple-600">
                    {orderData.package.price}₽
                  </span>
                </div>
                <div className="flex flex-col sm:flex-row sm:justify-between py-2">
                  <span className="text-xs sm:text-sm text-gray-600 mb-1 sm:mb-0">
                    {t('diamond_mosaic_success.purchase_date')}
                  </span>
                  <span className="text-sm sm:text-base font-medium">
                    {new Date(orderData.purchaseDate).toLocaleDateString(
                      'ru-RU'
                    )}
                  </span>
                </div>
              </div>
            </div>
          </div>
        </motion.div>

        {}
        {!isGenerating && downloadLinks.length > 0 && (
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.5 }}
            className="bg-white rounded-xl p-6 shadow-lg mb-8"
          >
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-2xl font-bold text-gray-800">
                {t('diamond_mosaic_success.your_files')}
              </h2>
              <button
                onClick={handleDownloadAll}
                className="bg-purple-600 text-white px-4 py-2 rounded-lg hover:bg-purple-700 transition-colors flex items-center"
              >
                <Download className="w-4 h-4 mr-2" />
                {t('diamond_mosaic_success.download_all')}
              </button>
            </div>

            <div className="space-y-3 sm:space-y-4">
              {downloadLinks.map((file, index) => (
                <motion.div
                  key={file.id}
                  initial={{ opacity: 0, x: -20 }}
                  animate={{ opacity: 1, x: 0 }}
                  transition={{ delay: 0.6 + index * 0.1 }}
                  className="flex flex-col sm:flex-row items-start sm:items-center justify-between p-3 sm:p-4 border border-gray-200 rounded-lg hover:border-purple-300 transition-colors gap-3 sm:gap-0"
                >
                  <div className="flex items-center w-full sm:w-auto">
                    <div className="w-8 h-8 sm:w-10 sm:h-10 bg-purple-100 rounded-lg flex items-center justify-center mr-3 sm:mr-4 flex-shrink-0">
                      {file.icon}
                    </div>
                    <div className="min-w-0 flex-1">
                      <h4 className="text-sm sm:text-base font-semibold text-gray-800 leading-tight">
                        {file.name}
                      </h4>
                      <p className="text-xs sm:text-sm text-gray-600 leading-relaxed">
                        {file.description}
                      </p>
                      <p className="text-xs text-gray-500 break-all">
                        {file.filename} • {file.size}
                      </p>
                    </div>
                  </div>
                  <button
                    onClick={() => handleDownload(file)}
                    className="bg-gray-100 text-gray-700 px-4 py-2 sm:px-4 sm:py-2 rounded-lg hover:bg-gray-200 active:bg-gray-300 transition-colors flex items-center text-sm sm:text-base touch-target w-full sm:w-auto justify-center"
                  >
                    <Download className="w-4 h-4 mr-1 sm:mr-2 flex-shrink-0" />
                    <span>{t('diamond_mosaic_success.download')}</span>
                  </button>
                </motion.div>
              ))}
            </div>
          </motion.div>
        )}

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
            {t('diamond_mosaic_success.create_new')}
            <ArrowRight className="w-5 h-5 ml-2" />
          </button>

          <button
            onClick={handleGoHome}
            className="bg-white text-purple-600 border-2 border-purple-600 px-8 py-4 rounded-xl font-semibold text-lg hover:bg-purple-50 transition-all duration-300 flex items-center justify-center"
          >
            <Home className="w-5 h-5 mr-2" />
            {t('diamond_mosaic_success.go_home')}
          </button>
        </motion.div>

        {}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.9 }}
          className="mt-8 text-center text-gray-600"
        >
          <p className="mb-2">
            {t('diamond_mosaic_success.email_notification')}
          </p>
          <p className="text-sm">
            {t('diamond_mosaic_success.support_contact')}
          </p>
        </motion.div>
      </div>
    </div>
  );
};

export default PreviewSuccessPage;
