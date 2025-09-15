import React, { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { motion } from 'framer-motion';
import {
  ArrowLeft,
  ArrowRight,
  Edit,
  ShoppingCart,
  Loader2,
  Sparkles,
  Eye,
  Check,
} from 'lucide-react';
import { useNavigate, useLocation } from 'react-router-dom';
import { useUIStore } from '../store/partnerStore';
import useCouponStore from '../store/couponStore';
import MosaicAPIClient, { MosaicAPI } from '../api/client';
import MarketplaceCards from '../components/MarketplaceCards';
import SwipeableAlbum from '../components/SwipeableAlbum';

const PreviewAlbumPage = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const location = useLocation();
  const { addNotification } = useUIStore();

  const [imageData, setImageData] = useState(null);
  const [selectedPreview, setSelectedPreview] = useState(0);
  const [useAI, setUseAI] = useState(false);
  const [previews, setPreviews] = useState([]);
  const [isGeneratingAI, setIsGeneratingAI] = useState(false);
  const [isGeneratingVariants, setIsGeneratingVariants] = useState(false);
  const [hasGeneratedVariants, setHasGeneratedVariants] = useState(false);
  const [generatingProgress, setGeneratingProgress] = useState(0);

  const styleVariants = [
    { style: 'venus', contrast: 'soft', label: t('preview_album.styles.venus_soft') },
    { style: 'venus', contrast: 'strong', label: t('preview_album.styles.venus_strong') },
    { style: 'sun', contrast: 'soft', label: t('preview_album.styles.sun_soft') },
    { style: 'sun', contrast: 'strong', label: t('preview_album.styles.sun_strong') },
    { style: 'moon', contrast: 'soft', label: t('preview_album.styles.moon_soft') },
    { style: 'moon', contrast: 'strong', label: t('preview_album.styles.moon_strong') },
    { style: 'mars', contrast: 'soft', label: t('preview_album.styles.mars_soft') },
    { style: 'mars', contrast: 'strong', label: t('preview_album.styles.mars_strong') },
  ];

  const couponStore = useCouponStore();
  const [isActivating, setIsActivating] = useState(false);

  useEffect(() => {
    if (hasGeneratedVariants || isGeneratingVariants) {
      return;
    }

    try {
      const locationState = location.state;
      
      if (locationState?.couponId && locationState?.couponCode) {
        const data = {
          size: locationState.size || couponStore.selectedSize || '30x40',
          imageUrl: locationState.imageUrl,
          couponId: locationState.couponId,
          couponCode: locationState.couponCode,
          fileName: 'image.jpg'
        };
        setImageData(data);
        generateContrastVariants(data);
        return;
      }

      if (couponStore.couponData && couponStore.editedImageUrl && !imageData) {
        const data = {
          size: couponStore.couponData.size || '30x40',
          imageUrl: couponStore.editedImageUrl || couponStore.previewUrl,
          couponId: couponStore.couponData.id,
          couponCode: couponStore.couponCode,
          fileName: 'image.jpg'
        };
        setImageData(data);
        generateContrastVariants(data);
        return;
      }

      let savedImageData = localStorage.getItem('diamondMosaic_selectedImage');
      let parsedData = null;

      if (savedImageData) {
        parsedData = JSON.parse(savedImageData);
      }

      if (!parsedData || !parsedData.selectedStyle) {
        const projectSettings = localStorage.getItem(
          'diamondMosaic_projectSettings'
        );
        if (projectSettings) {
          const settings = JSON.parse(projectSettings);
          if (settings.selectedStyle && settings.size) {
            parsedData = {
              size: settings.size,
              selectedStyle: settings.selectedStyle,
              timestamp: Date.now(),
              fileName: 'image.jpg',
            };

            localStorage.setItem(
              'diamondMosaic_selectedImage',
              JSON.stringify(parsedData)
            );
          }
        }
      }

      if (!parsedData) {
        if (couponStore.couponCode) {
          navigate('/coupon-activation');
        } else {
          navigate('/preview');
        }
        return;
      }

      if (!parsedData.selectedStyle) {
        navigate('/preview/styles');
        return;
      }

      if (!imageData) {
        setImageData(parsedData);
        generateContrastVariants(parsedData);
      }
    } catch (error) {
      navigate('/preview');
    }
  }, [navigate, location.pathname, hasGeneratedVariants, isGeneratingVariants, imageData]);

  useEffect(() => {
    return () => {
      setHasGeneratedVariants(false);
    };
  }, [location.pathname]);

  const generateContrastVariants = async data => {
    if (isGeneratingVariants || hasGeneratedVariants) {
      console.log('Skipping generation - already in progress or completed');
      return;
    }

    setIsGeneratingVariants(true);
    setGeneratingProgress(0);

    try {
      let imageId = data.imageId;

      if (!imageId) {
        const fileUrl = data.imageUrl || sessionStorage.getItem('diamondMosaic_fileUrl');
        if (!fileUrl) {
          throw new Error('No file URL found');
        }

        const response = await fetch(fileUrl);
        const blob = await response.blob();

        const formData = new FormData();
        formData.append('image', blob, 'image.jpg');
        formData.append('size', data.size);

        const uploadResult = await MosaicAPI.uploadImage(formData);
        imageId = uploadResult.image_id;

        if (imageId) {
          const updatedData = { ...data, imageId };
          setImageData(updatedData);
          setTimeout(() => {
            couponStore.setImageId(imageId);
          }, 0);
        }
      }

      const progressInterval = setInterval(() => {
        setGeneratingProgress(prev => {
          if (prev >= 90) return prev;
          return prev + Math.random() * 10;
        });
      }, 200);

      let imageFile = null;
      try {
        const fileUrl = data.imageUrl || sessionStorage.getItem('diamondMosaic_fileUrl');
        if (fileUrl) {
          const response = await fetch(fileUrl);
          if (response.ok) {
            imageFile = await response.blob();
          }
        }
      } catch (error) {
        console.warn('Could not get image file for fallback:', error);
      }

      const result = await MosaicAPI.generateAllPreviews(
        imageId,
        data.size,
        false,
        imageFile 
      );

      clearInterval(progressInterval);
      setGeneratingProgress(100);

      const generatedPreviews = result.previews.map((preview, index) => {
        const variant = styleVariants.find(
          v => v.style === preview.style && v.contrast === preview.contrast
        );
        return {
          id: index,
          url: preview.url,
          title: variant ? variant.label : preview.label,
          style: preview.style,
          contrast: preview.contrast,
          type: preview.is_ai ? 'ai' : 'style',
          isAI: preview.is_ai,
        };
      });

      setPreviews(generatedPreviews);
      setTimeout(() => {
        couponStore.setPreviews(generatedPreviews);
      }, 0);
      
      addNotification({
        type: 'success',
        message: t('notifications.previews_generated_success')
      });
    } catch (error) {

      try {
        const fileUrl = sessionStorage.getItem('diamondMosaic_fileUrl');
        if (!fileUrl) {
          throw new Error('No file URL found');
        }

        const response = await fetch(fileUrl);
        const blob = await response.blob();

        const generatedPreviews = [];

        for (let i = 0; i < styleVariants.length; i++) {
          const variant = styleVariants[i];

          try {
            const formData = new FormData();
            formData.append('image', blob, 'image.jpg');
            formData.append('size', data.size);
            formData.append('style', variant.style);
            formData.append('contrast_level', variant.contrast);
            formData.append('use_ai', 'false');

            const result = await MosaicAPI.generatePreviewVariant(formData);

            generatedPreviews.push({
              id: i,
              url: result.preview_url,
              title: variant.label,
              style: variant.style,
              contrast: variant.contrast,
              type: 'style',
              variant: variant,
            });

            setPreviews([...generatedPreviews]);
          } catch (error) {
            console.error(`Error generating contrast variant ${i}:`, error);
          }
        }

        setPreviews(generatedPreviews);
        setTimeout(() => {
          couponStore.setPreviews(generatedPreviews);
        }, 0);
        
        if (generatedPreviews.length > 0) {
          addNotification({
            type: 'success',
            message: t('notifications.previews_generated_success')
          });
        }
      } catch (fallbackError) {
        addNotification({
          type: 'error',
          message: t('diamond_mosaic_preview_album.contrast_generation_error'),
        });
      }
    } finally {
      setIsGeneratingVariants(false);
      setHasGeneratedVariants(true);
      setGeneratingProgress(0); 
    }
  };

  const generateAIPreviews = async () => {
    if (!imageData) return;

    setIsGeneratingAI(true);

    try {
      let imageId = imageData?.imageId;

      if (!imageId) {
        const fileUrl = sessionStorage.getItem('diamondMosaic_fileUrl');
        if (!fileUrl) {
          throw new Error('No file URL found for AI generation');
        }

        const response = await fetch(fileUrl);
        const blob = await response.blob();

        const formData = new FormData();
        formData.append('image', blob, 'image.jpg');
        formData.append('size', imageData.size);

        const uploadResult = await MosaicAPI.uploadImage(formData);
        imageId = uploadResult.image_id;

        if (imageId) {
          const updatedData = { ...imageData, imageId };
          localStorage.setItem(
            'diamondMosaic_selectedImage',
            JSON.stringify(updatedData)
          );
          setImageData(updatedData);
        }
      }

      let imageFile = null;
      try {
        const fileUrl = sessionStorage.getItem('diamondMosaic_fileUrl');
        if (fileUrl) {
          const response = await fetch(fileUrl);
          if (response.ok) {
            imageFile = await response.blob();
          }
        }
      } catch (error) {
        console.warn('Could not get image file for AI fallback:', error);
      }

      const result = await MosaicAPI.generateAllPreviews(
        imageId,
        imageData.size,
        true,
        imageFile 
      );

      const aiPreviews = result.previews
        .filter(p => p.is_ai)
        .map((preview, index) => ({
          id: previews.length + index,
          url: preview.url,
          title:
            preview.label ||
            `${t('diamond_mosaic_preview_album.ai_processing')} ${index + 1}`,
          type: 'ai',
          isAI: true,
        }));

      if (aiPreviews.length > 0) {
        setPreviews(prev => [...aiPreviews, ...prev]);
        setSelectedPreview(0);
        
        addNotification({
          type: 'success',
          message: t('notifications.ai_previews_generated')
        });
      }
    } catch (error) {
      console.error('Error generating AI previews:', error);

      try {
        const fileUrl = sessionStorage.getItem('diamondMosaic_fileUrl');
        const response = await fetch(fileUrl);
        const blob = await response.blob();

        const aiPreviews = [];

        for (let i = 0; i < 2; i++) {
          try {
            const formData = new FormData();
            formData.append('image', blob, 'image.jpg');
            formData.append('size', imageData.size);
            formData.append('style', imageData.selectedStyle);
            formData.append('use_ai', 'true');
            formData.append('ai_variant', i.toString());

            const result = await MosaicAPI.generatePreview(formData);

            aiPreviews.push({
              id: previews.length + i,
              url: result.preview_url,
              title: `${t('diamond_mosaic_preview_album.ai_processing')} ${i + 1}`,
              type: 'ai',
            });
          } catch (error) {
            console.error(`Error generating AI preview ${i}:`, error);
          }
        }

        if (aiPreviews.length > 0) {
          setPreviews(prev => [...aiPreviews, ...prev]);
          setSelectedPreview(0);
          
          addNotification({
            type: 'success',
            message: t('notifications.ai_previews_generated')
          });
        }
      } catch (fallbackError) {
        console.error('Fallback AI generation also failed:', fallbackError);
        addNotification({
          type: 'error',
          message: t(
            'diamond_mosaic_preview_album.ai_preview_generation_error'
          ),
        });
      }
    } finally {
      setIsGeneratingAI(false);
    }
  };

  const handleAIToggle = enabled => {
    setUseAI(enabled);

    if (enabled && !previews.some(p => p.type === 'ai')) {
      generateAIPreviews();
    }
  };

  const handlePreviewSelect = index => {
    setSelectedPreview(index);
  };

  const handleEditImage = () => {
    try {
      const editorData = {
        size: imageData.size,
        style: imageData.selectedStyle,
        returnTo: '/preview/album',
      };
      localStorage.setItem(
        'diamondMosaic_editorSettings',
        JSON.stringify(editorData)
      );

      navigate('/image-editor');
    } catch (error) {
      console.error('Error preparing editor data:', error);
      addNotification({
        type: 'error',
        message: 'Ошибка при подготовке редактора',
      });
    }
  };

  const handlePurchase = async () => {
    if (!imageData || !previews[selectedPreview]) {
      addNotification({
        type: 'error',
        message: t('diamond_mosaic_preview_album.select_preview_for_purchase'),
      });
      return;
    }

    const couponData = couponStore.couponData || 
      (imageData.couponId ? { id: imageData.couponId, code: imageData.couponCode } : null);

    if (couponData && couponData.id) {
      setIsActivating(true);
      try {
        const selectedPreviewData = previews[selectedPreview];

        const activationData = {
          preview_image_url: imageData.imageUrl || imageData.previewUrl,
          selected_preview_id: `${selectedPreviewData.style}_${selectedPreviewData.contrast || 'default'}`,
          final_schema_url: selectedPreviewData.url,
          page_count: 100,
          user_email: null,
        };

        await MosaicAPI.activateCoupon(couponData.id, activationData);

        addNotification({
          type: 'success',
          message: t('notifications.coupon_activated'),
        });

        couponStore.updateCouponData({
          ...activationData,
          activated_at: new Date().toISOString(),
          status: 'activated'
        });
        couponStore.setActivationStep('activated');

        navigate('/coupon-activation');
      } catch (error) {
        addNotification({
          type: 'error',
          message: t('notifications.activation_error'),
        });
      } finally {
        setIsActivating(false);
      }
    } else {
      try {
        const purchaseData = {
          size: imageData.size,
          style: imageData.selectedStyle,
          selectedPreview: previews[selectedPreview],
          originalImage: imageData.previewUrl,
        };
        localStorage.setItem(
          'diamondMosaic_purchaseData',
          JSON.stringify(purchaseData)
        );

        navigate('/preview/purchase');
      } catch (error) {
        console.error('Error preparing purchase data:', error);
      }
    }
  };

  const handleBack = () => {
    navigate('/preview/styles');
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

  if (!imageData) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-purple-50 via-pink-50 to-blue-50 flex items-center justify-center">
        <Loader2 className="w-8 h-8 animate-spin text-purple-600" />
      </div>
    );
  }

  const currentPreview = previews[selectedPreview];

  // Beautiful loading overlay for preview generation
  if (isGeneratingVariants) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-purple-50 via-pink-50 to-blue-50 flex items-center justify-center px-4">
        <div className="bg-white/80 backdrop-blur-sm rounded-3xl shadow-2xl p-8 sm:p-12 max-w-md w-full text-center border border-white/20">
          <motion.div
            initial={{ scale: 0.8, opacity: 0 }}
            animate={{ scale: 1, opacity: 1 }}
            transition={{ duration: 0.5 }}
          >
            {/* Animated mosaic icon */}
            <motion.div
              animate={{ rotate: 360 }}
              transition={{ duration: 2, repeat: Infinity, ease: "linear" }}
              className="w-16 h-16 mx-auto mb-6 bg-gradient-to-br from-purple-500 to-pink-500 rounded-xl flex items-center justify-center"
            >
              <Sparkles className="w-8 h-8 text-white" />
            </motion.div>

            {/* Title */}
            <h2 className="text-xl sm:text-2xl font-bold text-gray-800 mb-2">
              {t('diamond_mosaic_preview_album.generating_previews')}
            </h2>
            
            {/* Subtitle */}
            <p className="text-sm sm:text-base text-gray-600 mb-6">
              Создаем 8 уникальных превью с различными стилями освещения
            </p>

            {/* Progress bar */}
            <div className="w-full bg-gray-200 rounded-full h-2 mb-4 overflow-hidden">
              <motion.div
                className="h-2 bg-gradient-to-r from-purple-500 to-pink-500 rounded-full"
                initial={{ width: 0 }}
                animate={{ width: `${generatingProgress}%` }}
                transition={{ duration: 0.3, ease: "easeOut" }}
              />
            </div>

            {/* Progress percentage */}
            <p className="text-sm text-gray-500">
              {Math.round(generatingProgress)}% завершено
            </p>

            {/* Animated dots */}
            <div className="flex justify-center space-x-1 mt-4">
              {[0, 1, 2].map(i => (
                <motion.div
                  key={i}
                  className="w-2 h-2 bg-purple-400 rounded-full"
                  animate={{ scale: [1, 1.5, 1] }}
                  transition={{
                    duration: 1,
                    repeat: Infinity,
                    delay: i * 0.2,
                  }}
                />
              ))}
            </div>
          </motion.div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-purple-50 via-pink-50 to-blue-50 py-4 sm:py-6 lg:py-8 px-4 sm:px-6 lg:px-8">
      <div className="container mx-auto max-w-7xl">
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
              {t('diamond_mosaic_preview_album.back_to_style_selection')}
            </span>
          </button>

          <div className="text-center">
            <h1 className="text-2xl sm:text-3xl md:text-4xl lg:text-5xl font-bold bg-gradient-to-r from-purple-600 to-pink-600 bg-clip-text text-transparent mb-3 sm:mb-4 leading-tight">
              {t('diamond_mosaic_preview_album.title')}
            </h1>
            <p className="text-sm sm:text-base lg:text-lg text-gray-600 leading-relaxed">
              {t('diamond_mosaic_preview_album.size_label')}{' '}
              <span className="font-semibold">
                {imageData.size} {t('common.cm')}
              </span>{' '}
              •{t('diamond_mosaic_preview_album.style_label')}{' '}
              <span className="font-semibold">
                {getStyleTitle(imageData.selectedStyle)}
              </span>
            </p>
          </div>
        </motion.div>

        <div className="flex flex-col items-center max-w-6xl mx-auto">
          {}
          {}
          <motion.div
            initial={{ opacity: 0, y: -20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.1 }}
            className="mb-8 w-full max-w-md"
          >
            <div
              className={`p-4 bg-white rounded-xl border-2 transition-all duration-300 ${
                useAI ? 'border-purple-400 bg-purple-50' : 'border-gray-200'
              }`}
            >
              <label className="flex items-center justify-between cursor-pointer group">
                <div className="flex items-center">
                  <Sparkles
                    className={`w-5 h-5 mr-3 transition-colors ${
                      useAI
                        ? 'text-purple-600'
                        : 'text-gray-400 group-hover:text-purple-500'
                    }`}
                  />
                  <div>
                    <span className="font-medium text-gray-900">
                      {t('diamond_mosaic_preview_album.ai_processing')}
                    </span>
                    <p className="text-sm text-gray-600 mt-1">
                      {t('diamond_mosaic_preview_album.ai_description')}
                    </p>
                  </div>
                </div>
                <div className="relative ml-4">
                  <input
                    type="checkbox"
                    checked={useAI}
                    onChange={e => handleAIToggle(e.target.checked)}
                    className="sr-only"
                  />
                  <div
                    className={`w-14 h-8 rounded-full transition-colors ${
                      useAI ? 'bg-purple-600' : 'bg-gray-300'
                    }`}
                  >
                    <div
                      className={`absolute top-1 left-1 bg-white w-6 h-6 rounded-full transition-transform shadow-sm ${
                        useAI ? 'translate-x-6' : 'translate-x-0'
                      }`}
                    />
                  </div>
                </div>
              </label>

              {}
              {isGeneratingAI && (
                <motion.div
                  initial={{ opacity: 0, height: 0 }}
                  animate={{ opacity: 1, height: 'auto' }}
                  exit={{ opacity: 0, height: 0 }}
                  className="mt-4 pt-4 border-t border-purple-200"
                >
                  <div className="flex items-center justify-center text-purple-600">
                    <Loader2 className="w-5 h-5 animate-spin mr-2" />
                    <span className="text-sm font-medium">
                      {t('diamond_mosaic_preview_album.generating_ai')}
                    </span>
                  </div>
                  <div className="mt-2 bg-purple-100 rounded-full h-2 overflow-hidden">
                    <motion.div
                      className="bg-purple-600 h-full"
                      initial={{ width: '0%' }}
                      animate={{ width: '100%' }}
                      transition={{ duration: 30 }}
                    />
                  </div>
                </motion.div>
              )}

              {}
              {useAI &&
                !isGeneratingAI &&
                previews.some(p => p.type === 'ai') && (
                  <motion.div
                    initial={{ opacity: 0 }}
                    animate={{ opacity: 1 }}
                    className="mt-3 flex items-center justify-center text-green-600"
                  >
                    <Check className="w-4 h-4 mr-2" />
                    <span className="text-sm">
                      {t('diamond_mosaic_preview_album.ai_ready')}
                    </span>
                  </motion.div>
                )}
            </div>
          </motion.div>

          {}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            className="w-full max-w-2xl mb-8"
          >
            <div className="bg-white rounded-xl p-4 sm:p-6 mb-4 sm:mb-6 shadow-lg">
              <div className="aspect-square bg-gray-100 rounded-lg overflow-hidden mb-3 sm:mb-4">
                {currentPreview?.url ? (
                  <img
                    src={currentPreview.url}
                    alt={currentPreview.title}
                    className="w-full h-full object-cover"
                  />
                ) : (
                  <div className="w-full h-full flex items-center justify-center">
                    <Eye className="w-8 h-8 sm:w-12 sm:h-12 text-gray-400" />
                  </div>
                )}
              </div>

              {currentPreview && (
                <div className="text-center">
                  <h3 className="text-lg sm:text-xl font-semibold text-gray-800 mb-1 sm:mb-2 leading-tight">
                    {currentPreview.title}
                  </h3>
                  <p className="text-sm sm:text-base text-gray-600 capitalize">
                    {currentPreview.type}
                  </p>
                </div>
              )}
            </div>
          </motion.div>

          {}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.3 }}
            className="flex justify-center w-full max-w-md mx-auto mb-6 sm:mb-8"
          >
            <button
              onClick={handleEditImage}
              className="w-full bg-white text-purple-600 border-2 border-purple-600 px-4 py-3 sm:px-6 sm:py-3 rounded-xl font-semibold hover:bg-purple-50 active:bg-purple-100 transition-all duration-300 flex items-center justify-center text-sm sm:text-base touch-target"
            >
              <Edit className="w-4 h-4 sm:w-5 sm:h-5 mr-2 flex-shrink-0" />
              <span>{t('diamond_mosaic_preview_album.edit_image')}</span>
            </button>
          </motion.div>

          {}
          <SwipeableAlbum
            previews={previews}
            selectedPreview={selectedPreview}
            onPreviewSelect={handlePreviewSelect}
            isGeneratingVariants={isGeneratingVariants}
          />

          {}
          {currentPreview?.url && (
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.5 }}
              className="w-full max-w-md mx-auto mb-6 sm:mb-8"
            >
              <div className="text-center">
                <p className="text-gray-700 mb-3 sm:mb-4 text-sm sm:text-base lg:text-lg leading-relaxed px-2">
                  {t('diamond_mosaic_preview_album.liked_preview_text')}
                </p>
                <button
                  onClick={handlePurchase}
                  disabled={isActivating}
                  className="w-full bg-gradient-to-r from-purple-600 to-pink-600 text-white px-4 py-3 sm:px-6 sm:py-4 rounded-xl font-semibold text-sm sm:text-base lg:text-lg hover:from-purple-700 hover:to-pink-700 active:from-purple-800 active:to-pink-800 transition-all duration-300 shadow-lg hover:shadow-xl flex items-center justify-center disabled:opacity-50 disabled:cursor-not-allowed touch-target"
                >
                  {isActivating ? (
                    <>
                      <Loader2 className="w-4 h-4 sm:w-5 sm:h-5 mr-2 animate-spin flex-shrink-0" />
                      <span>Активация купона...</span>
                    </>
                  ) : (couponStore.couponData || imageData?.couponId) ? (
                    <>
                      <Check className="w-4 h-4 sm:w-5 sm:h-5 mr-2 flex-shrink-0" />
                      <span>{t('preview_album.create_schema')}</span>
                    </>
                  ) : (
                    <>
                      <ShoppingCart className="w-4 h-4 sm:w-5 sm:h-5 mr-2 flex-shrink-0" />
                      <span className="text-center leading-tight">
                        {t('diamond_mosaic_preview_album.buy_coupon_and_generate')}
                      </span>
                    </>
                  )}
                </button>
              </div>
            </motion.div>
          )}

          {}
          {currentPreview?.url && (
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.3 }}
              className="w-full max-w-4xl"
            >
              <MarketplaceCards
                selectedSize={imageData.size}
                selectedStyle={imageData.selectedStyle}
              />
            </motion.div>
          )}
        </div>
      </div>
    </div>
  );
};

export default PreviewAlbumPage;
