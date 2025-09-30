import React, { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { motion } from 'framer-motion';
import {
  ArrowLeft,
  ShoppingCart,
  CreditCard,
  Check,
  Package,
  Star,
} from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { useUIStore } from '../store/partnerStore';

const PreviewPurchasePage = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { addNotification } = useUIStore();

  const [purchaseData, setPurchaseData] = useState(null);
  const [selectedPackage, setSelectedPackage] = useState('');
  const [isProcessing, setIsProcessing] = useState(false);

  const getPackages = () => [
    {
      id: 'basic',
      name: t('diamond_mosaic_purchase.packages.basic.name'),
      price: 299,
      originalPrice: 499,
      features: t('diamond_mosaic_purchase.packages.basic.features'),
      popular: false,
    },
    {
      id: 'premium',
      name: t('diamond_mosaic_purchase.packages.premium.name'),
      price: 499,
      originalPrice: 799,
      features: t('diamond_mosaic_purchase.packages.premium.features'),
      popular: true,
    },
    {
      id: 'professional',
      name: t('diamond_mosaic_purchase.packages.professional.name'),
      price: 799,
      originalPrice: 1299,
      features: t('diamond_mosaic_purchase.packages.professional.features'),
      popular: false,
    },
  ];

  useEffect(() => {
    try {
      const savedPurchaseData = localStorage.getItem(
        'diamondMosaic_purchaseData'
      );

      if (!savedPurchaseData) {
        navigate('/preview');
        return;
      }

      const parsedData = JSON.parse(savedPurchaseData);
      setPurchaseData(parsedData);

      setSelectedPackage('premium');
    } catch (error) {
      console.error('Error loading purchase data:', error);
      navigate('/preview');
    }
  }, [navigate]);

  const handlePackageSelect = packageId => {
    setSelectedPackage(packageId);
  };

  const handlePurchase = async () => {
    if (!selectedPackage || !purchaseData) {
      addNotification({
        type: 'error',
        message: t('diamond_mosaic_purchase.select_package'),
      });
      return;
    }

    setIsProcessing(true);

    try {
      const selectedPkg = getPackages().find(pkg => pkg.id === selectedPackage);

      await new Promise(resolve => setTimeout(resolve, 2000));

      const orderData = {
        orderId: `DMO_${Date.now()}`,
        package: selectedPkg,
        imageData: purchaseData,
        purchaseDate: new Date().toISOString(),
        status: 'completed',
      };

      localStorage.setItem(
        'diamondMosaic_lastOrder',
        JSON.stringify(orderData)
      );

      localStorage.removeItem('diamondMosaic_purchaseData');
      localStorage.removeItem('diamondMosaic_selectedImage');
      sessionStorage.removeItem('diamondMosaic_fileUrl');

      addNotification({
        type: 'success',
        title: t(
          'diamond_mosaic_purchase.notifications.purchase_success_title'
        ),
        message: t(
          'diamond_mosaic_purchase.notifications.purchase_success_message'
        ),
      });

      navigate('/preview/success', {
        state: { orderData },
      });
    } catch (error) {
      console.error('Error processing purchase:', error);
      addNotification({
        type: 'error',
        title: t('diamond_mosaic_purchase.purchase_error'),
        message: t(
          'diamond_mosaic_purchase.notifications.purchase_failed_message'
        ),
      });
    } finally {
      setIsProcessing(false);
    }
  };

  const handleBack = () => {
    navigate('/preview/album');
  };

  const getStyleTitle = styleKey => {
    const styleMap = {
      max_colors: t('diamond_mosaic_styles.styles.max_colors.title'),
      pop_art: t('diamond_mosaic_styles.styles.pop_art.title'),
      grayscale: t('diamond_mosaic_styles.styles.grayscale.title'),
      skin_tones: t('diamond_mosaic_styles.styles.skin_tones.title'),
    };
    return styleMap[styleKey] || styleKey;
  };

  if (!purchaseData) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-purple-50 via-pink-50 to-blue-50 flex items-center justify-center">
        <div className="text-purple-600">
          {t('diamond_mosaic_purchase.notifications.loading')}
        </div>
      </div>
    );
  }

  const selectedPkg = getPackages().find(pkg => pkg.id === selectedPackage);

  return (
    <div className="min-h-screen bg-gradient-to-br from-purple-50 via-pink-50 to-blue-50 py-4 sm:py-6 lg:py-8 px-4 sm:px-6 lg:px-8">
      <div className="container mx-auto max-w-6xl">
        {}
        <motion.div
          initial={{ opacity: 0, y: -20 }}
          animate={{ opacity: 1, y: 0 }}
          className="mb-6 sm:mb-8"
        >
          <button
            onClick={handleBack}
            className="flex items-center text-purple-600 hover:text-purple-700 mb-4 transition-colors p-2 -m-2 touch-target"
          >
            <ArrowLeft className="w-4 h-4 sm:w-5 sm:h-5 mr-2" />
            <span className="text-sm sm:text-base">
              {t('diamond_mosaic_purchase.back_to_preview')}
            </span>
          </button>

          <div className="text-center">
            <h1 className="text-2xl sm:text-3xl md:text-4xl lg:text-5xl font-bold bg-gradient-to-r from-purple-600 to-pink-600 bg-clip-text text-transparent mb-3 sm:mb-4 leading-tight">
              {t('diamond_mosaic_purchase.title')}
            </h1>
            <p className="text-sm sm:text-base lg:text-lg text-gray-600 leading-relaxed">
              {t('diamond_mosaic_purchase.subtitle')}
            </p>
          </div>
        </motion.div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 sm:gap-8">
          {}
          <motion.div
            initial={{ opacity: 0, x: -20 }}
            animate={{ opacity: 1, x: 0 }}
            className="lg:col-span-1 order-2 lg:order-1"
          >
            <div className="bg-white rounded-xl p-4 sm:p-6 shadow-lg mb-4 sm:mb-6">
              <h3 className="text-lg sm:text-xl font-semibold text-gray-800 mb-3 sm:mb-4">
                {t('diamond_mosaic_purchase.order_section.title')}
              </h3>

              <div className="mb-4">
                <div className="aspect-square bg-gray-100 rounded-lg overflow-hidden mb-2 max-w-xs mx-auto lg:max-w-none">
                  <img
                    src={
                      purchaseData.selectedPreview?.url ||
                      purchaseData.originalImage
                    }
                    alt="Preview"
                    className="w-full h-full object-cover"
                  />
                </div>
                <p className="text-xs sm:text-sm text-gray-600 text-center leading-relaxed">
                  {purchaseData.selectedPreview?.title ||
                    t('diamond_mosaic_purchase.selected_preview')}
                </p>
              </div>

              <div className="space-y-2 sm:space-y-3 text-xs sm:text-sm">
                <div className="flex flex-col sm:flex-row sm:justify-between">
                  <span className="text-gray-600 mb-1 sm:mb-0">
                    {t('diamond_mosaic_purchase.size_label')}
                  </span>
                  <span className="font-medium">
                    {purchaseData.size} {t('common.cm')}
                  </span>
                </div>
                <div className="flex flex-col sm:flex-row sm:justify-between">
                  <span className="text-gray-600 mb-1 sm:mb-0">
                    {t('diamond_mosaic_purchase.style_label')}
                  </span>
                  <span className="font-medium">
                    {getStyleTitle(purchaseData.style)}
                  </span>
                </div>
                {purchaseData.selectedPreview?.type && (
                  <div className="flex flex-col sm:flex-row sm:justify-between">
                    <span className="text-gray-600 mb-1 sm:mb-0">
                      {t(
                        'diamond_mosaic_purchase.order_section.processing_type'
                      )}
                    </span>
                    <span className="font-medium capitalize">
                      {purchaseData.selectedPreview.type}
                    </span>
                  </div>
                )}
              </div>
            </div>

            {}
            <div className="bg-white rounded-xl p-4 sm:p-6 shadow-lg">
              <h3 className="text-base sm:text-lg font-semibold text-gray-800 mb-3 sm:mb-4 flex items-center">
                <Package className="w-4 h-4 sm:w-5 sm:h-5 mr-2 text-purple-600 flex-shrink-0" />
                <span>{t('diamond_mosaic_purchase.what_you_get.title')}</span>
              </h3>
              <ul className="space-y-2 sm:space-y-3 text-xs sm:text-sm text-gray-600">
                <li className="flex items-start">
                  <Check className="w-4 h-4 text-green-500 mr-2 sm:mr-3 flex-shrink-0 mt-0.5" />
                  <span className="leading-relaxed">
                    {t('diamond_mosaic_purchase.what_you_get.detailed_schema')}
                  </span>
                </li>
                <li className="flex items-start">
                  <Check className="w-4 h-4 text-green-500 mr-2 sm:mr-3 flex-shrink-0 mt-0.5" />
                  <span className="leading-relaxed">
                    {t('diamond_mosaic_purchase.what_you_get.materials_list')}
                  </span>
                </li>
                <li className="flex items-start">
                  <Check className="w-4 h-4 text-green-500 mr-2 sm:mr-3 flex-shrink-0 mt-0.5" />
                  <span className="leading-relaxed">
                    {t(
                      'diamond_mosaic_purchase.what_you_get.step_instructions'
                    )}
                  </span>
                </li>
                <li className="flex items-start">
                  <Check className="w-4 h-4 text-green-500 mr-2 sm:mr-3 flex-shrink-0 mt-0.5" />
                  <span className="leading-relaxed">
                    {t('diamond_mosaic_purchase.what_you_get.download_files')}
                  </span>
                </li>
              </ul>
            </div>
          </motion.div>

          <motion.div
            initial={{ opacity: 0, x: 20 }}
            animate={{ opacity: 1, x: 0 }}
            className="lg:col-span-2 order-1 lg:order-2"
          >
            <h2 className="text-xl sm:text-2xl font-bold text-gray-800 mb-4 sm:mb-6">
              {t('diamond_mosaic_purchase.labels.select_package')}
            </h2>

            <div className="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-3 gap-4 sm:gap-6 mb-6 sm:mb-8">
              {getPackages().map((pkg, index) => (
                <motion.div
                  key={pkg.id}
                  initial={{ opacity: 0, y: 20 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ delay: 0.1 + index * 0.1 }}
                  className={`relative bg-white rounded-xl p-4 sm:p-6 cursor-pointer transition-all duration-300 touch-target ${
                    selectedPackage === pkg.id
                      ? 'ring-2 sm:ring-4 ring-purple-500 shadow-xl transform scale-[1.02] sm:scale-105'
                      : 'shadow-lg hover:shadow-xl hover:scale-[1.01] sm:hover:scale-102 active:scale-[0.99]'
                  } ${pkg.popular ? 'border-2 border-purple-500' : 'border border-gray-200'}`}
                  onClick={() => handlePackageSelect(pkg.id)}
                >
                  {pkg.popular && (
                    <div className="absolute -top-3 left-1/2 transform -translate-x-1/2">
                      <div className="bg-purple-500 text-white px-4 py-1 rounded-full text-sm font-medium flex items-center">
                        <Star className="w-3 h-3 mr-1" />
                        {t('diamond_mosaic_purchase.labels.popular')}
                      </div>
                    </div>
                  )}

                  <div className="text-center">
                    <h3 className="text-xl font-semibold text-gray-800 mb-2">
                      {pkg.name}
                    </h3>

                    <div className="mb-4">
                      <div className="flex items-center justify-center space-x-2">
                        <span className="text-3xl font-bold text-purple-600">
                          {pkg.price}₽
                        </span>
                        <span className="text-lg text-gray-400 line-through">
                          {pkg.originalPrice}₽
                        </span>
                      </div>
                      <div className="text-sm text-green-600 font-medium">
                        {t('diamond_mosaic_purchase.pricing.discount')}{' '}
                        {Math.round((1 - pkg.price / pkg.originalPrice) * 100)}%
                      </div>
                    </div>

                    <ul className="space-y-2 text-sm text-gray-600 mb-6">
                      {pkg.features.map((feature, idx) => (
                        <li key={idx} className="flex items-center">
                          <Check className="w-4 h-4 text-green-500 mr-2 flex-shrink-0" />
                          {feature}
                        </li>
                      ))}
                    </ul>
                  </div>
                </motion.div>
              ))}
            </div>

            {}
            {selectedPkg && (
              <motion.div
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                className="bg-white rounded-xl p-6 shadow-lg"
              >
                <div className="flex items-center justify-between mb-6">
                  <div>
                    <h3 className="text-lg font-semibold text-gray-800">
                      {selectedPkg.name}{' '}
                      {t('diamond_mosaic_purchase.pricing.package')}
                    </h3>
                    <p className="text-gray-600">
                      {t('diamond_mosaic_purchase.pricing.total_to_pay')}{' '}
                      <span className="text-2xl font-bold text-purple-600">
                        {selectedPkg.price}₽
                      </span>
                    </p>
                  </div>
                  <div className="text-right">
                    <div className="text-sm text-gray-500 line-through">
                      {selectedPkg.originalPrice}₽
                    </div>
                    <div className="text-green-600 font-medium">
                      {t('diamond_mosaic_purchase.pricing.savings')}{' '}
                      {selectedPkg.originalPrice - selectedPkg.price}₽
                    </div>
                  </div>
                </div>

                <button
                  onClick={handlePurchase}
                  disabled={isProcessing}
                  className="w-full bg-gradient-to-r from-purple-600 to-pink-600 text-white px-6 py-4 rounded-xl font-semibold text-lg hover:from-purple-700 hover:to-pink-700 transition-all duration-300 shadow-lg hover:shadow-xl disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center"
                >
                  {isProcessing ? (
                    <>
                      <div className="w-5 h-5 border-2 border-white border-t-transparent rounded-full animate-spin mr-2"></div>
                      {t('diamond_mosaic_purchase.buttons.processing')}
                    </>
                  ) : (
                    <>
                      <CreditCard className="w-5 h-5 mr-2" />
                      {t('diamond_mosaic_purchase.buttons.buy_and_generate')}
                    </>
                  )}
                </button>

                <p className="text-xs text-gray-500 text-center mt-3">
                  {t('diamond_mosaic_purchase.terms')}
                </p>
              </motion.div>
            )}
          </motion.div>
        </div>
      </div>
    </div>
  );
};

export default PreviewPurchasePage;
