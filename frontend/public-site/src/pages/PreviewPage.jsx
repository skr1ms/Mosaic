import React, { useState, useEffect, useRef } from 'react';
import { useTranslation } from 'react-i18next';
import { motion } from 'framer-motion';
import {
  Upload,
  Ruler,
  ArrowRight,
  Info,
  Image as ImageIcon,
  Check,
} from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { useUIStore } from '../store/partnerStore';
import ImageEditor from '../components/ImageEditor';

const PreviewPage = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { addNotification } = useUIStore();

  const fileInputRef = useRef(null);

  const [selectedSize, setSelectedSize] = useState('');
  const [selectedFile, setSelectedFile] = useState(null);
  const [previewUrl, setPreviewUrl] = useState(null);
  const [isUploading, setIsUploading] = useState(false);
  const [editedImageUrl, setEditedImageUrl] = useState(null);
  const [showEditor, setShowEditor] = useState(false);

  useEffect(() => {
    const urlParams = new URLSearchParams(window.location.search);
    const isNewProject = urlParams.get('new') === 'true';
    const returnedFromEditor =
      sessionStorage.getItem('diamondMosaic_returnedFromEditor') === '1';

    if (isNewProject && !returnedFromEditor) {
      try {
        localStorage.removeItem('pendingOrder');
        localStorage.removeItem('activeCoupon');

        const keys = Object.keys(localStorage);
        keys.forEach(key => {
          if (
            key.startsWith('preview_') ||
            key.startsWith('temp_') ||
            key.startsWith('shop_')
          ) {
            localStorage.removeItem(key);
          }
        });

        localStorage.removeItem('diamondMosaic_selectedImage');
        localStorage.removeItem('diamondMosaic_editorSettings');
        sessionStorage.removeItem('diamondMosaic_fileUrl');

        console.log('Starting new project - cleared previous data');

        window.history.replaceState({}, document.title, '/preview');
      } catch (error) {
        console.error('Error clearing storage:', error);
      }
    } else if (returnedFromEditor) {
      sessionStorage.removeItem('diamondMosaic_returnedFromEditor');
    }
  }, []);

  useEffect(() => {
    try {
      const savedImageData = localStorage.getItem(
        'diamondMosaic_selectedImage'
      );
      if (savedImageData) {
        const parsedData = JSON.parse(savedImageData);

        if (parsedData.previewUrl) {
          setPreviewUrl(parsedData.previewUrl);

          if (parsedData.fileName) {
            const file = new File([''], parsedData.fileName, {
              type: 'image/*',
            });
            setSelectedFile(file);
          }

          if (parsedData.hasEdits) {
            setEditedImageUrl(parsedData.previewUrl);
          }
        }

        if (parsedData.size) {
          setSelectedSize(parsedData.size);
        }
      }
    } catch (error) {
      console.error('Error restoring state from localStorage:', error);
    }
  }, []);

  const getSizes = () => [
    {
      key: '21x30',
      title: t('diamond_mosaic_page.size_selection.sizes.21x30'),
      desc: t('diamond_mosaic_page.size_selection.details.21x30.desc'),
      detail: t('diamond_mosaic_page.size_selection.details.21x30.detail'),
    },
    {
      key: '30x40',
      title: t('diamond_mosaic_page.size_selection.sizes.30x40'),
      desc: t('diamond_mosaic_page.size_selection.details.30x40.desc'),
      detail: t('diamond_mosaic_page.size_selection.details.30x40.detail'),
    },
    {
      key: '40x40',
      title: t('diamond_mosaic_page.size_selection.sizes.40x40'),
      desc: t('diamond_mosaic_page.size_selection.details.40x40.desc'),
      detail: t('diamond_mosaic_page.size_selection.details.40x40.detail'),
    },
    {
      key: '40x50',
      title: t('diamond_mosaic_page.size_selection.sizes.40x50'),
      desc: t('diamond_mosaic_page.size_selection.details.40x50.desc'),
      detail: t('diamond_mosaic_page.size_selection.details.40x50.detail'),
    },
    {
      key: '40x60',
      title: t('diamond_mosaic_page.size_selection.sizes.40x60'),
      desc: t('diamond_mosaic_page.size_selection.details.40x60.desc'),
      detail: t('diamond_mosaic_page.size_selection.details.40x60.detail'),
    },
    {
      key: '50x70',
      title: t('diamond_mosaic_page.size_selection.sizes.50x70'),
      desc: t('diamond_mosaic_page.size_selection.details.50x70.desc'),
      detail: t('diamond_mosaic_page.size_selection.details.50x70.detail'),
    },
  ];

  const handleFileSelect = event => {
    const file = event.target.files[0];
    if (!file) return;

    if (!file.type.startsWith('image/')) {
      addNotification({
        type: 'error',
        message: t('diamond_mosaic_page.upload_section.file_type_error'),
      });
      return;
    }

    if (file.size > 10 * 1024 * 1024) {
      addNotification({
        type: 'error',
        message: t('diamond_mosaic_page.upload_section.file_size_error'),
      });
      return;
    }

    setSelectedFile(file);

    const reader = new FileReader();
    reader.onload = e => {
      setPreviewUrl(e.target.result);
      setShowEditor(true);
    };
    reader.readAsDataURL(file);
  };

  const handleEditorSave = (editedUrl, editorParams) => {
    if (editorParams?.newImage) {
      setPreviewUrl(editedUrl);

      if (editorParams.fileName) {
        const file = new File([''], editorParams.fileName, { type: 'image/*' });
        setSelectedFile(file);
      }

      setEditedImageUrl(null);

      return;
    }

    setEditedImageUrl(editedUrl);
    setShowEditor(false);
  };

  const handleEditorCancel = () => {
    setShowEditor(false);
  };

  const handleSizeSelect = sizeKey => {
    setSelectedSize(sizeKey);
  };

  const handleContinue = () => {
    if (!selectedFile || !selectedSize) {
      addNotification({
        type: 'error',
        message: t('diamond_mosaic_page.size_and_image_required'),
      });
      return;
    }

    const finalImageUrl = editedImageUrl || previewUrl;

    try {
      localStorage.setItem(
        'diamondMosaic_selectedImage',
        JSON.stringify({
          size: selectedSize,
          fileName: selectedFile.name,
          previewUrl: finalImageUrl,
          timestamp: Date.now(),
          hasEdits: editedImageUrl !== null,
        })
      );

      localStorage.setItem(
        'diamondMosaic_projectSettings',
        JSON.stringify({
          size: selectedSize,
          timestamp: Date.now(),
        })
      );

      localStorage.setItem(
        'diamondMosaic_editorSettings',
        JSON.stringify({
          size: selectedSize,
          style: null,
          returnTo: '/preview/styles',
        })
      );

      if (editedImageUrl) {
        fetch(editedImageUrl)
          .then(res => res.blob())
          .then(blob => {
            const fileUrl = URL.createObjectURL(blob);
            sessionStorage.setItem('diamondMosaic_fileUrl', fileUrl);
          });
      } else {
        const fileUrl = URL.createObjectURL(selectedFile);
        sessionStorage.setItem('diamondMosaic_fileUrl', fileUrl);
      }
    } catch (error) {
      console.error('Error saving image data:', error);
    }

    navigate('/preview/styles');
  };

  const handleRemoveImage = () => {
    setSelectedFile(null);
    setPreviewUrl(null);
    setEditedImageUrl(null);
    setShowEditor(false);

    try {
      localStorage.removeItem('diamondMosaic_selectedImage');
      localStorage.removeItem('diamondMosaic_editorSettings');
      sessionStorage.removeItem('diamondMosaic_fileUrl');
    } catch (error) {
      console.error('Error clearing localStorage:', error);
    }

    if (fileInputRef.current) {
      fileInputRef.current.value = '';
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-purple-50 via-pink-50 to-blue-50">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
        {}
        <motion.div
          initial={{ opacity: 0, y: -20 }}
          animate={{ opacity: 1, y: 0 }}
          className="flex flex-col items-center text-center mb-8 sm:mb-12 px-4"
        >
          <h1 className="text-2xl sm:text-3xl md:text-4xl lg:text-5xl font-bold text-gray-900 mb-3 sm:mb-4 leading-tight">
            {t('diamond_mosaic_page.title')}
          </h1>
          <p className="text-base sm:text-lg md:text-xl text-gray-600 max-w-2xl mx-auto leading-relaxed">
            {t('diamond_mosaic_page.upload_section.subtitle')}
          </p>
        </motion.div>

        <motion.section
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.1 }}
          className="mb-16"
        >
          <div className="force-center-container mb-8">
            <Ruler
              className="w-6 h-6 sm:w-7 sm:h-7 text-purple-600 mb-2 sm:mb-3"
              style={{ margin: '0 auto' }}
            />
            <h2 className="text-xl sm:text-2xl font-bold text-gray-900 mb-4 leading-tight force-center">
              {t('diamond_mosaic_page.size_selection.title')}
            </h2>
          </div>

          <div className="flex justify-center mb-6 sm:mb-8">
            <motion.div
              initial={{ opacity: 0, y: -10 }}
              animate={{ opacity: 1, y: 0 }}
              className="inline-flex items-center gap-2 py-2 px-4 bg-blue-50 border border-blue-200 rounded-lg"
            >
              <Info className="w-4 h-4 text-blue-600 flex-shrink-0" />
              <p className="text-sm text-blue-800 leading-snug whitespace-nowrap">
                {t('diamond_mosaic_page.size_selection.hint')}
              </p>
            </motion.div>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3 sm:gap-4 lg:gap-6 px-4 sm:px-0">
            {getSizes().map((size, index) => {
              const isSelected = selectedSize === size.key;

              const rectangleClasses = {
                '21x30': 'w-8 h-12',
                '30x40': 'w-10 h-14',
                '40x40': 'w-12 h-12',
                '40x50': 'w-12 h-16',
                '40x60': 'w-12 h-18',
                '50x70': 'w-14 h-22',
              };

              return (
                <motion.div
                  key={size.key}
                  initial={{ opacity: 0, y: 20 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ duration: 0.5, delay: index * 0.1 }}
                  onClick={() => handleSizeSelect(size.key)}
                  className={`relative bg-white rounded-xl sm:rounded-2xl shadow-lg p-4 sm:p-6 cursor-pointer transition-all duration-300 transform hover:scale-[1.02] sm:hover:scale-105 hover:shadow-xl ${
                    isSelected
                      ? 'ring-2 sm:ring-4 ring-purple-500 border-purple-500'
                      : 'border border-gray-200 hover:border-gray-300'
                  }`}
                >
                  {isSelected && (
                    <div className="absolute -top-1 sm:-top-2 -right-1 sm:-right-2 w-5 h-5 sm:w-6 sm:h-6 bg-purple-500 rounded-full flex items-center justify-center">
                      <Check className="w-3 h-3 sm:w-4 sm:h-4 text-white" />
                    </div>
                  )}

                  <div className="flex flex-col items-center text-center">
                    <div
                      className={`mb-3 sm:mb-4 rounded ${rectangleClasses[size.key]} ${
                        isSelected ? 'bg-purple-500' : 'bg-purple-300'
                      }`}
                    />

                    <h3 className="text-base sm:text-lg lg:text-xl font-semibold text-gray-900 mb-1 sm:mb-2 leading-tight">
                      {size.title}
                    </h3>

                    <p className="text-xs sm:text-sm text-gray-600 mb-1 sm:mb-2 leading-relaxed">
                      {size.desc}
                    </p>

                    {size.detail && (
                      <p className="text-xs text-purple-600 font-medium leading-relaxed">
                        {size.detail}
                      </p>
                    )}
                  </div>
                </motion.div>
              );
            })}
          </div>
        </motion.section>

        {}
        <motion.section
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.3 }}
          className="mb-12"
        >
          <div className="force-center-container mb-8">
            <ImageIcon
              className="w-6 h-6 sm:w-7 sm:h-7 text-purple-600 mb-2 sm:mb-3"
              style={{ margin: '0 auto' }}
            />
            <h2 className="text-xl sm:text-2xl font-bold text-gray-900 mb-6 sm:mb-8 leading-tight force-center">
              {t('diamond_mosaic_page.upload_section.title')}
            </h2>
          </div>

          <div className="max-w-4xl mx-auto px-4 sm:px-0">
            {!previewUrl ? (
              <div className="border-2 border-dashed border-gray-300 rounded-xl sm:rounded-2xl p-6 sm:p-8 lg:p-12 text-center bg-white hover:border-purple-400 transition-colors">
                <input
                  ref={fileInputRef}
                  type="file"
                  id="image-upload"
                  accept="image/*"
                  onChange={handleFileSelect}
                  className="hidden"
                />
                <label htmlFor="image-upload" className="cursor-pointer block">
                  <Upload className="w-12 h-12 sm:w-16 sm:h-16 text-gray-400 mx-auto mb-4 sm:mb-6" />
                  <p className="text-lg sm:text-xl lg:text-2xl font-medium text-gray-700 mb-2 sm:mb-3 leading-tight">
                    {t('diamond_mosaic_page.upload_section.select_image')}
                  </p>
                  <p className="text-sm sm:text-base text-gray-500 mb-4 sm:mb-6 leading-relaxed">
                    {t('diamond_mosaic_page.upload_section.formats')}
                  </p>
                  <div className="inline-block bg-purple-600 text-white px-6 py-3 sm:px-8 sm:py-3 rounded-xl hover:bg-purple-700 transition-colors font-semibold text-sm sm:text-base">
                    {t('diamond_mosaic_page.upload_section.button')}
                  </div>
                </label>
              </div>
            ) : (
              <div className="max-w-4xl mx-auto">
                {showEditor ? (
                  <ImageEditor
                    key={previewUrl}
                    imageUrl={previewUrl}
                    onSave={handleEditorSave}
                    onCancel={handleEditorCancel}
                    title={t('diamond_mosaic_page.image_editor.setup_title')}
                    showCropHint={true}
                    fileName={selectedFile?.name}
                  />
                ) : (
                  <div className="bg-white rounded-xl sm:rounded-2xl p-4 sm:p-6 lg:p-8 border border-gray-200 shadow-lg">
                    <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between mb-6 sm:mb-8 gap-3 sm:gap-0">
                      <h4 className="text-lg sm:text-xl font-semibold text-gray-800">
                        {t('diamond_mosaic_page.image_editor.setup_title')}
                      </h4>
                      <button
                        onClick={handleRemoveImage}
                        className="text-red-500 hover:text-red-600 flex items-center text-sm font-medium transition-colors px-2 py-1 rounded hover:bg-red-50"
                      >
                        ✕{' '}
                        <span className="ml-1">
                          {t('diamond_mosaic_page.image_editor.delete_image')}
                        </span>
                      </button>
                    </div>

                    <div className="relative mb-6 sm:mb-8">
                      <div className="bg-gray-50 rounded-xl sm:rounded-2xl p-3 sm:p-4 lg:p-6 flex justify-center">
                        <img
                          src={editedImageUrl || previewUrl}
                          alt="Preview"
                          className="max-w-full h-auto rounded-lg shadow-lg"
                          style={{ maxHeight: '300px' }}
                        />
                      </div>
                    </div>

                    {}
                    <div className="flex justify-center mb-6 sm:mb-8">
                      <button
                        onClick={() => {
                          const editorData = {
                            size: selectedSize,
                            returnTo: '/preview',
                          };
                          localStorage.setItem(
                            'diamondMosaic_editorSettings',
                            JSON.stringify(editorData)
                          );

                          navigate('/image-editor');
                        }}
                        className="bg-purple-100 text-purple-700 px-6 py-3 sm:px-8 sm:py-3 rounded-xl hover:bg-purple-200 transition-colors font-medium text-sm sm:text-base w-full sm:w-auto max-w-xs"
                      >
                        {t('diamond_mosaic_page.image_editor.edit_image')}
                      </button>
                    </div>

                    {}
                    <div className="text-center bg-blue-50 rounded-xl p-4 sm:p-6 border border-blue-200">
                      <div className="flex items-center justify-center mb-2 flex-wrap">
                        <div className="w-2 h-2 bg-green-500 rounded-full mr-2 flex-shrink-0"></div>
                        <p className="font-medium text-gray-800 text-sm sm:text-base break-all">
                          {selectedFile?.name}
                        </p>
                      </div>
                      <p className="text-gray-600 text-xs sm:text-sm leading-relaxed">
                        {t('diamond_mosaic_page.image_editor.file_info')}
                      </p>
                    </div>
                  </div>
                )}
              </div>
            )}
          </div>
        </motion.section>

        {selectedSize && previewUrl && (
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            className="flex justify-center px-4"
          >
            <button
              onClick={handleContinue}
              disabled={isUploading}
              className="bg-gradient-to-r from-purple-600 to-pink-600 text-white px-8 py-4 sm:px-12 sm:py-4 rounded-xl font-semibold text-base sm:text-lg hover:from-purple-700 hover:to-pink-700 transition-all duration-300 shadow-lg hover:shadow-xl disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center w-full sm:w-auto max-w-sm"
            >
              <span>{t('diamond_mosaic_page.page_navigation.continue')}</span>
              <ArrowRight className="w-4 h-4 sm:w-5 sm:h-5 ml-2" />
            </button>
          </motion.div>
        )}
      </div>
    </div>
  );
};

export default PreviewPage;
