import React, { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { motion } from 'framer-motion';
import { useNavigate } from 'react-router-dom';
import {
  ArrowLeft,
  ArrowRight,
  Palette,
  Sparkles,
  Sun,
  Moon,
  Loader2,
  Check,
  Circle,
} from 'lucide-react';
import { useUIStore } from '../store/partnerStore';
import { cleanupImageStorage, emergencyCleanup } from '../utils/storageCleanup';
import { MosaicAPI } from '../api/client';

const PreviewStylesPage = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { addNotification } = useUIStore();

  const [imageData, setImageData] = useState(null);
  const [selectedStyle, setSelectedStyle] = useState(null);
  const [isGeneratingPreviews, setIsGeneratingPreviews] = useState(false);
  const [stylePreviews, setStylePreviews] = useState({});

  const getSizeTitle = (sizeKey) => {
    const sizeTranslations = {
      '21x30': t('diamond_mosaic_page.size_selection.sizes.21x30'),
      '30x40': t('diamond_mosaic_page.size_selection.sizes.30x40'),
      '40x40': t('diamond_mosaic_page.size_selection.sizes.40x40'),
      '40x50': t('diamond_mosaic_page.size_selection.sizes.40x50'),
      '40x60': t('diamond_mosaic_page.size_selection.sizes.40x60'),
      '50x70': t('diamond_mosaic_page.size_selection.sizes.50x70'),
    };
    return sizeTranslations[sizeKey] || sizeKey;
  };

  const styles = [
    {
      key: 'grayscale',
      title: t('diamond_mosaic_styles.styles.grayscale.title'),
      description: t('diamond_mosaic_styles.styles.grayscale.description'),
      icon: <Circle className="w-6 h-6" />,
      color: 'from-gray-800 to-gray-400',
    },
    {
      key: 'skin_tones',
      title: t('diamond_mosaic_styles.styles.skin_tones.title'),
      description: t('diamond_mosaic_styles.styles.skin_tones.description'),
      icon: <Sun className="w-6 h-6" />,
      color: 'from-orange-400 to-amber-500',
    },
    {
      key: 'pop_art',
      title: t('diamond_mosaic_styles.styles.pop_art.title'),
      description: t('diamond_mosaic_styles.styles.pop_art.description'),
      icon: <Sparkles className="w-6 h-6" />,
      color: 'from-pink-400 to-purple-500',
    },
    {
      key: 'max_colors',
      title: t('diamond_mosaic_styles.styles.max_colors.title'),
      description: t('diamond_mosaic_styles.styles.max_colors.description'),
      icon: <Palette className="w-6 h-6" />,
      color: 'from-red-400 to-yellow-500',
    },
  ];

  useEffect(() => {
    try {
      const savedImageData = localStorage.getItem(
        'diamondMosaic_selectedImage'
      );
      if (!savedImageData) {
        navigate('/preview');
        return;
      }

      const parsedData = JSON.parse(savedImageData);
      setImageData(parsedData);

      generateStylePreviews(parsedData);
    } catch (error) {
      console.error('Error loading image data:', error);
      navigate('/preview');
    }
  }, [navigate]);

  const generateStylePreviews = async data => {
    setIsGeneratingPreviews(true);

    try {
      console.log('Starting real style preview generation for 4 styles...');
      
      let blob;
      
      const fileUrl = sessionStorage.getItem('diamondMosaic_fileUrl');
      if (fileUrl) {
        try {
          const response = await fetch(fileUrl);
          if (response.ok) {
            blob = await response.blob();
            console.log('Successfully got blob from sessionStorage URL');
          }
        } catch (error) {
          console.warn('Failed to fetch from sessionStorage URL:', error);
        }
      }
      
      if (!blob && data.previewUrl) {
        try {
          const response = await fetch(data.previewUrl);
          if (response.ok) {
            blob = await response.blob();
            console.log('Successfully got blob from previewUrl');
          }
        } catch (error) {
          console.warn('Failed to fetch from previewUrl:', error);
        }
      }
      
      if (!blob && data.editedUrl) {
        try {
          const response = await fetch(data.editedUrl);
          if (response.ok) {
            blob = await response.blob();
            console.log('Successfully got blob from editedUrl');
          }
        } catch (error) {
          console.warn('Failed to fetch from editedUrl:', error);
        }
      }
      
      if (!blob) {
        throw new Error('Could not obtain image blob for processing');
      }
      
      const formData = new FormData();
      formData.append('image', blob, 'image.jpg');
      formData.append('size', data.size || '30x40');
      
      console.log('Sending request to generate 4 style variants...');
      
      const response = await MosaicAPI.generateStyleVariants(formData);
      
      if (response && response.previews && response.previews.length > 0) {
        const previews = {};
        
        response.previews.forEach(preview => {
          previews[preview.style] = preview.preview_url;
        });
        
        setStylePreviews(previews);
        
        console.log(`Successfully generated ${response.previews.length} style previews`);
        
        addNotification({
          type: 'success',
          message: t('diamond_mosaic_styles.previews_generated_success')
        });
      } else {
        throw new Error('Invalid response from server: no previews generated');
      }
      
    } catch (error) {
      console.error('Error generating style previews:', error);
      
      const previews = {};
      const fallbackImage = data.editedUrl || data.previewUrl;
      
      for (const style of styles) {
        previews[style.key] = fallbackImage;
      }
      
      setStylePreviews(previews);
      
      addNotification({
        type: 'error',
        message: t('diamond_mosaic_styles.error_generating_previews'),
      });
      
      console.warn('Using fallback: showing original image for all styles');
    } finally {
      setIsGeneratingPreviews(false);
    }
  };

  const handleStyleSelect = async style => {
    setSelectedStyle(style.key);

    try {
      const savedImageData = localStorage.getItem(
        'diamondMosaic_selectedImage'
      );
      if (savedImageData) {
        const parsedData = JSON.parse(savedImageData);
        parsedData.selectedStyle = style.key;
        parsedData.styleTitle = style.title;
        
        const stylePreview = stylePreviews[style.key] || parsedData.previewUrl;
        if (stylePreview && stylePreview !== parsedData.previewUrl) {
          parsedData.stylePreview = stylePreview;
        }
        
        try {
          localStorage.setItem(
            'diamondMosaic_selectedImage',
            JSON.stringify(parsedData)
          );
        } catch (storageError) {
          if (storageError.name === 'QuotaExceededError') {
            console.warn('localStorage quota exceeded, performing cleanup');
            emergencyCleanup();
            
            try {
              localStorage.setItem(
                'diamondMosaic_selectedImage',
                JSON.stringify(parsedData)
              );
            } catch (retryError) {
              console.error('Failed to save even after cleanup:', retryError);
              throw new Error('Storage cleanup failed');
            }
          } else {
            throw storageError;
          }
        }
      }

      const projectSettings = {
        selectedStyle: style.key,
        styleTitle: style.title,
        size: imageData?.size,
        timestamp: Date.now()
      };
      
      try {
        localStorage.setItem(
          'diamondMosaic_projectSettings',
          JSON.stringify(projectSettings)
        );
      } catch (storageError) {
        if (storageError.name === 'QuotaExceededError') {
          emergencyCleanup();
          localStorage.setItem(
            'diamondMosaic_projectSettings',
            JSON.stringify(projectSettings)
          );
        } else {
          throw storageError;
        }
      }

    } catch (error) {
      console.error('Error saving style selection:', error);
      
      const isQuotaError = error.message.includes('quota') || 
                          error.name === 'QuotaExceededError' ||
                          error.message.includes('Storage cleanup failed');
      
      addNotification({
        type: 'error',
        message: isQuotaError 
          ? t('diamond_mosaic_styles.storage_full_error')
          : t('diamond_mosaic_styles.selection_error'),
      });
    }
    
    navigate('/preview/album');
  };

  const handleContinue = () => {
    if (!selectedStyle) {
      addNotification({
        type: 'error',
        message: t('diamond_mosaic_styles.select_style_error'),
      });
      return;
    }

    try {
      const savedImageData = localStorage.getItem(
        'diamondMosaic_selectedImage'
      );
      if (savedImageData) {
        const parsedData = JSON.parse(savedImageData);
        parsedData.selectedStyle = selectedStyle;
        
        try {
          localStorage.setItem(
            'diamondMosaic_selectedImage',
            JSON.stringify(parsedData)
          );
        } catch (storageError) {
          if (storageError.name === 'QuotaExceededError') {
            emergencyCleanup();
            localStorage.setItem(
              'diamondMosaic_selectedImage',
              JSON.stringify(parsedData)
            );
          } else {
            throw storageError;
          }
        }
      }

      const projectSettings = {
        size: imageData.size,
        selectedStyle: selectedStyle,
        timestamp: Date.now(),
      };
      
      try {
        localStorage.setItem(
          'diamondMosaic_projectSettings',
          JSON.stringify(projectSettings)
        );
      } catch (storageError) {
        if (storageError.name === 'QuotaExceededError') {
          emergencyCleanup();
          localStorage.setItem(
            'diamondMosaic_projectSettings',
            JSON.stringify(projectSettings)
          );
        } else {
          throw storageError;
        }
      }

    } catch (error) {
      console.error('Error saving style selection:', error);
      
      const isQuotaError = error.message.includes('quota') || error.name === 'QuotaExceededError';
      addNotification({
        type: 'error',
        message: isQuotaError 
          ? t('diamond_mosaic_styles.storage_full_error')
          : t('diamond_mosaic_styles.selection_error'),
      });
    }
    
    navigate('/preview/album');
  };

  const handleBack = () => {
    navigate('/preview');
  };

  if (!imageData) {
    return (
      <div className="min-h-screen bg-white flex items-center justify-center px-4">
        <div className="text-center max-w-md">
          <div className="w-12 h-12 border-4 border-purple-600 border-t-transparent rounded-full animate-spin mx-auto mb-4"></div>
          <div className="text-gray-600 mb-4">
            {t('diamond_mosaic_styles.loading')}
          </div>
          <button
            onClick={() => navigate('/preview')}
            className="text-purple-600 hover:text-purple-700 underline text-sm"
          >
            ← {t('diamond_mosaic_styles.back_to_editor')}
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-white">
      <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-4 sm:py-6 lg:py-8">
        {}
        <div className="mb-6 sm:mb-8">
          <button
            onClick={handleBack}
            className="flex items-center text-purple-600 hover:text-purple-700 mb-4 transition-colors p-2 -m-2 touch-target"
          >
            <ArrowLeft className="w-4 h-4 sm:w-5 sm:h-5 mr-2" />
            <span className="text-sm sm:text-base">
              {t('diamond_mosaic_styles.back_to_editor')}
            </span>
          </button>

          <div className="text-center">
            <h1 className="text-2xl sm:text-3xl lg:text-4xl font-bold text-gray-900 mb-2 sm:mb-3 leading-tight">
              {t('diamond_mosaic_styles.page_title')}
            </h1>
            <p className="text-sm sm:text-base text-gray-600 leading-relaxed">
              {t('diamond_mosaic_styles.size_label', { size: getSizeTitle(imageData?.size) || 'Unknown' })}
            </p>
          </div>
        </div>

        <div className="mb-6 sm:mb-8 flex justify-center">
          {imageData && (imageData.editedUrl || imageData.previewUrl) ? (
            <img
              src={imageData.editedUrl || imageData.previewUrl}
              alt="Your uploaded image"
              className="w-32 h-32 sm:w-40 sm:h-40 lg:w-48 lg:h-48 object-cover rounded-lg border-2 border-gray-200 shadow-md"
              onError={e => {
                console.error('❌ Image failed to load:', e.target.src);
                e.target.style.display = 'none';
              }}
            />
          ) : (
            <div className="w-32 h-32 sm:w-40 sm:h-40 lg:w-48 lg:h-48 bg-gray-200 rounded-lg border-2 border-gray-300 flex items-center justify-center">
              <span className="text-gray-500 text-sm">No image available</span>
            </div>
          )}
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 sm:gap-6 mb-6 sm:mb-8">
          {styles.map(style => (
          <motion.div
            key={style.key}
            onClick={() => handleStyleSelect(style)}
            whileHover={{ scale: 1.02 }}
            whileTap={{ scale: 0.98 }}
            className={`
              relative overflow-hidden rounded-xl border-2 cursor-pointer transition-all hover:shadow-xl touch-target
              ${
                selectedStyle === style.key
                  ? 'border-purple-500 ring-2 sm:ring-4 ring-purple-200 shadow-lg'
                  : 'border-gray-200 hover:border-purple-300 active:bg-gray-50'
              }
            `}
          >
              {}
              <div className="relative h-32 sm:h-40 lg:h-48 bg-gradient-to-br ${style.color} overflow-hidden">
                {stylePreviews[style.key] ? (
                  <img
                    src={stylePreviews[style.key]}
                    alt={style.title}
                    className="w-full h-full object-cover"
                  />
                ) : (
                  <div className="w-full h-full flex items-center justify-center">
                    {isGeneratingPreviews ? (
                      <Loader2 className="w-8 h-8 text-white animate-spin" />
                    ) : (
                      <div
                        className={`w-full h-full bg-gradient-to-br ${style.color}`}
                      />
                    )}
                  </div>
                )}

                {}
                {selectedStyle === style.key && (
                  <div className="absolute top-2 right-2 w-6 h-6 sm:w-8 sm:h-8 bg-white rounded-full flex items-center justify-center shadow-lg">
                    <Check className="w-4 h-4 sm:w-5 sm:h-5 text-purple-600" />
                  </div>
                )}
              </div>

              {}
              <div className="p-3 sm:p-4 bg-white">
                <div className="flex items-center justify-center mb-2 sm:mb-3">
                  <div
                    className={`text-transparent bg-clip-text bg-gradient-to-r ${style.color}`}
                  >
                    {style.icon}
                  </div>
                </div>
                <h3 className="text-base sm:text-lg font-semibold text-gray-900 mb-1 text-center leading-tight">
                  {style.title}
                </h3>
                <p className="text-xs sm:text-sm text-gray-600 text-center leading-relaxed">
                  {style.description}
                </p>
              </div>
            </motion.div>
          ))}
        </div>

        {isGeneratingPreviews && (
          <motion.div
            className="text-center mb-8"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
          >
            <div className="inline-flex items-center px-6 py-3 bg-purple-100 rounded-full text-purple-700">
              <Loader2 className="w-5 h-5 mr-3 animate-spin" />
              {t('diamond_mosaic_styles.generating_previews')}
            </div>
          </motion.div>
        )}

        {}
        <div className="flex flex-col sm:flex-row gap-3 sm:gap-4 max-w-md mx-auto px-4 sm:px-0">
          <button
            onClick={handleBack}
            className="flex-1 py-3 sm:py-4 px-4 sm:px-6 bg-white border-2 border-gray-300 text-gray-700 rounded-xl font-medium hover:bg-gray-50 active:bg-gray-100 transition-colors text-sm sm:text-base touch-target"
          >
            {t('diamond_mosaic_styles.back')}
          </button>

          <button
            onClick={handleContinue}
            disabled={!selectedStyle}
            className="flex-1 py-3 sm:py-4 px-4 sm:px-6 bg-gradient-to-r from-purple-600 to-pink-600 text-white rounded-xl font-medium hover:from-purple-700 hover:to-pink-700 active:from-purple-800 active:to-pink-800 transition-all disabled:opacity-50 disabled:cursor-not-allowed text-sm sm:text-base touch-target"
          >
            {t('diamond_mosaic_styles.create_preview')}
          </button>
        </div>
      </div>
    </div>
  );
};

export default PreviewStylesPage;
