import React, { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { motion } from 'framer-motion';
import { useNavigate } from 'react-router-dom';
import { useUIStore } from '../store/partnerStore';
import ImageEditor from '../components/ImageEditor';

const ImageEditorPage = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { addNotification } = useUIStore();

  const [imageData, setImageData] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    try {
      const savedImageData = localStorage.getItem(
        'diamondMosaic_selectedImage'
      );
      const editorSettings = localStorage.getItem(
        'diamondMosaic_editorSettings'
      );

      if (!savedImageData) {
        addNotification({
          type: 'error',
          message: t('image_editor.image_not_found'),
        });
        navigate('/preview');
        return;
      }

      const parsedData = JSON.parse(savedImageData);
      const parsedSettings = editorSettings ? JSON.parse(editorSettings) : {};

      const finalImageData = {
        ...parsedData,
        settings: parsedSettings,
      };

      setImageData(finalImageData);
    } catch (error) {
      console.error('Error loading image data:', error);
      addNotification({
        type: 'error',
        message: t('image_editor.error_loading_data'),
      });
      navigate('/preview');
    } finally {
      setLoading(false);
    }
  }, [navigate, addNotification, t]);

  const handleSave = (editedImageUrl, editorParams) => {
    try {
      if (editorParams?.newImage) {
        const newImageData = {
          url: editedImageUrl,
          previewUrl: editedImageUrl,
          fileName: editorParams.fileName || imageData.fileName,
          timestamp: Date.now(),
          settings: imageData.settings,
        };

        localStorage.setItem(
          'diamondMosaic_selectedImage',
          JSON.stringify(newImageData)
        );

        setImageData(newImageData);

        addNotification({
          type: 'success',
          message: t('notifications.image_uploaded'),
        });

        return;
      }

      const updatedImageData = {
        ...imageData,
        previewUrl: editedImageUrl,
        hasEdits: true,
        editorParams,
        timestamp: Date.now(),
      };

      localStorage.setItem(
        'diamondMosaic_selectedImage',
        JSON.stringify(updatedImageData)
      );

      addNotification({
        type: 'success',
        message: t('image_editor.save_success'),
      });

      const returnTo = imageData?.settings?.returnTo || '/preview/styles';

      try {
        sessionStorage.setItem('diamondMosaic_returnedFromEditor', '1');
      } catch {}
      navigate(returnTo);
    } catch (error) {
      console.error('Error saving edited image:', error);
      addNotification({
        type: 'error',
        message: t('image_editor.save_error'),
      });
    }
  };

  const handleCancel = () => {
    const returnTo = imageData?.settings?.returnTo || '/preview';
    navigate(returnTo);
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-purple-50 via-pink-50 to-blue-50 flex items-center justify-center px-4">
        <div className="text-center">
          <div className="w-12 h-12 sm:w-16 sm:h-16 border-4 border-purple-600 border-t-transparent rounded-full animate-spin mx-auto mb-3 sm:mb-4"></div>
          <p className="text-sm sm:text-base text-gray-600">
            {t('image_editor.loading')}
          </p>
        </div>
      </div>
    );
  }

  if (!imageData) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-purple-50 via-pink-50 to-blue-50 flex items-center justify-center px-4">
        <div className="text-center">
          <p className="text-sm sm:text-base text-gray-600">
            {t('image_editor.image_not_found')}
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-purple-50 via-pink-50 to-blue-50">
      <div className="max-w-5xl mx-auto px-4 sm:px-6 lg:px-8 py-4 sm:py-6 lg:py-8">
        <motion.div
          initial={{ opacity: 0, y: -20 }}
          animate={{ opacity: 1, y: 0 }}
          className="text-center mb-4 sm:mb-6"
        >
          <h1 className="text-xl sm:text-2xl lg:text-3xl font-bold text-gray-800 mb-2 sm:mb-3 leading-tight">
            {t('image_editor.page_title')}
          </h1>
          <p className="text-sm sm:text-base lg:text-lg text-gray-600 max-w-xl mx-auto leading-relaxed px-4 sm:px-0">
            {t('image_editor.page_description')}
          </p>
        </motion.div>

        {}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.2 }}
          className="max-w-4xl mx-auto"
        >
          <ImageEditor
            key={imageData.timestamp || imageData.uploadedAt || Date.now()}
            imageUrl={
              imageData.editedUrl || imageData.previewUrl || imageData.url
            }
            onSave={handleSave}
            onCancel={handleCancel}
            title={t('diamond_mosaic_page.image_editor.setup_title')}
            showCropHint={true}
            aspectRatio={1}
            fileName={imageData.fileName}
          />
        </motion.div>

        {}
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ delay: 0.4 }}
          className="mt-4 sm:mt-6 max-w-4xl mx-auto"
        >
          <div className="bg-white/70 backdrop-blur-sm rounded-xl p-4 sm:p-6 shadow-lg">
            <h3 className="text-base sm:text-lg font-semibold text-gray-900 mb-3 sm:mb-4 text-center sm:text-left">
              {t('image_editor.recommendations.title')}
            </h3>
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3 sm:gap-4 text-xs sm:text-sm text-gray-600">
              <div className="flex items-start space-x-2 sm:space-x-3">
                <div className="w-2 h-2 bg-purple-500 rounded-full mt-1.5 sm:mt-2 flex-shrink-0"></div>
                <div className="leading-relaxed">
                  <strong>
                    {t('image_editor.recommendations.crop.title')}
                  </strong>{' '}
                  {t('image_editor.recommendations.crop.description')}
                </div>
              </div>
              <div className="flex items-start space-x-2 sm:space-x-3">
                <div className="w-2 h-2 bg-blue-500 rounded-full mt-1.5 sm:mt-2 flex-shrink-0"></div>
                <div className="leading-relaxed">
                  <strong>
                    {t('image_editor.recommendations.rotate.title')}
                  </strong>{' '}
                  {t('image_editor.recommendations.rotate.description')}
                </div>
              </div>
              <div className="flex items-start space-x-2 sm:space-x-3">
                <div className="w-2 h-2 bg-green-500 rounded-full mt-1.5 sm:mt-2 flex-shrink-0"></div>
                <div className="leading-relaxed">
                  <strong>
                    {t('image_editor.recommendations.scale.title')}
                  </strong>{' '}
                  {t('image_editor.recommendations.scale.description')}
                </div>
              </div>
            </div>
          </div>
        </motion.div>

        {}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.5 }}
          className="mt-6 sm:mt-8 max-w-4xl mx-auto flex justify-center"
        >
          <button
            onClick={() => {
              try {
                const savedImageData = localStorage.getItem(
                  'diamondMosaic_selectedImage'
                );
                if (savedImageData) {
                  const imageData = JSON.parse(savedImageData);

                  try {
                    const keys = Object.keys(localStorage);
                    keys.forEach(key => {
                      if (
                        (key.startsWith('preview_') ||
                          key.startsWith('style_') ||
                          key.startsWith('temp_')) &&
                        key !== 'diamondMosaic_selectedImage'
                      ) {
                        localStorage.removeItem(key);
                      }
                    });
                  } catch (cleanupError) {
                    console.warn(
                      'Error cleaning up localStorage:',
                      cleanupError
                    );
                  }

                  const updatedImageData = {
                    ...imageData,
                    timestamp: Date.now(),
                    hasEdits: true,
                  };

                  localStorage.setItem(
                    'diamondMosaic_selectedImage',
                    JSON.stringify(updatedImageData)
                  );

                  const projectSettings = localStorage.getItem(
                    'diamondMosaic_projectSettings'
                  );
                  let hasProjectSettings = false;

                  if (projectSettings) {
                    try {
                      const settings = JSON.parse(projectSettings);
                      if (settings.selectedStyle && settings.size) {
                        const updatedImageDataWithSettings = {
                          ...updatedImageData,
                          selectedStyle: settings.selectedStyle,
                          size: settings.size,
                        };
                        localStorage.setItem(
                          'diamondMosaic_selectedImage',
                          JSON.stringify(updatedImageDataWithSettings)
                        );
                        hasProjectSettings = true;
                      }
                    } catch (error) {
                      console.error('Error parsing project settings:', error);
                    }
                  }

                  if (
                    hasProjectSettings ||
                    imageData.selectedStyle ||
                    imageData.style
                  ) {
                    navigate('/preview/album');
                  } else {
                    navigate('/preview/styles');
                  }
                } else {
                  navigate('/preview/styles');
                }
              } catch (error) {
                console.error('Error reading saved data:', error);
                navigate('/preview/styles');
              }
            }}
            className="inline-flex items-center gap-2 sm:gap-3 px-6 sm:px-8 py-3 sm:py-4 bg-purple-600 hover:bg-purple-700 active:bg-purple-800 text-white font-semibold rounded-xl transition-all duration-200 shadow-lg hover:shadow-xl text-sm sm:text-base"
          >
            <span>{t('common.continue')}</span>
            <svg
              className="w-4 h-4 sm:w-5 sm:h-5"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M13 7l5 5m0 0l-5 5m5-5H6"
              />
            </svg>
          </button>
        </motion.div>
      </div>
    </div>
  );
};

export default ImageEditorPage;
