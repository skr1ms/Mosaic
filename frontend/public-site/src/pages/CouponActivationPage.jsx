import React, { useEffect, useRef } from 'react';
import { useTranslation } from 'react-i18next';
import { motion } from 'framer-motion';
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
  Unlock,
} from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { useUIStore } from '../store/partnerStore';
import useCouponStore from '../store/couponStore';
import { MosaicAPI } from '../api/client';
import ImageEditor from '../components/ImageEditor';

const CouponActivationPage = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { addNotification } = useUIStore();
  const fileInputRef = useRef(null);

  const {
    couponCode,
    setCouponCode,
    couponData,
    setCouponData,
    activationStep,
    setActivationStep,
    selectedFile,
    setSelectedFile,
    previewUrl,
    setPreviewUrl,
    editedImageUrl,
    setEditedImageUrl,
    saveImageData,
    clearSession
  } = useCouponStore();

  const [isValidating, setIsValidating] = React.useState(false);
  const [showEditor, setShowEditor] = React.useState(false);
  const [isActivating, setIsActivating] = React.useState(false);
  const [searchPageNumber, setSearchPageNumber] = React.useState('');
  const [resendEmail, setResendEmail] = React.useState('');
  const [isSendingEmail, setIsSendingEmail] = React.useState(false);

  useEffect(() => {
    const urlParams = new URLSearchParams(window.location.search);
    const urlCoupon = urlParams.get('coupon');
    if (urlCoupon && !couponCode) {
      setCouponCode(urlCoupon);
    }
  }, []);

  const validateCoupon = async () => {
    if (!couponCode.trim()) {
      addNotification({
        type: 'error',
        message: t('coupon_activation.please_enter_code'),
      });
      return;
    }

    setIsValidating(true);
    try {
      const response = await MosaicAPI.validateCoupon(couponCode.trim());

      if (response.valid === false) {
        addNotification({
          type: 'error',
          message: t('coupon_activation.coupon_not_found'),
        });
        return;
      }

      setCouponData(response.coupon || response);

      if (
        response.coupon?.status === 'activated' ||
        response.status === 'activated'
      ) {
        try {
          const reactivationData = await MosaicAPI.reactivateCoupon(
            couponCode.trim()
          );
          setCouponData({
            ...(response.coupon || response),
            ...reactivationData,
            activated_at: reactivationData.activated_at,
            preview_url: reactivationData.preview_url,
            stones_count: reactivationData.stones_count,
            archive_url: reactivationData.archive_url,
            page_count: reactivationData.page_count,
            can_download: reactivationData.can_download,
            can_send_email: reactivationData.can_send_email,
          });
        } catch (reactivationError) {
          console.error('Error fetching reactivation data:', reactivationError);
        }
        setActivationStep('activated');
      } else {
        setActivationStep('upload');
      }
    } catch (error) {
      addNotification({
        type: 'error',
        message: error.message || t('coupon_activation.validation_error'),
      });
    } finally {
      setIsValidating(false);
    }
  };

  const handleFileSelect = event => {
    const file = event.target.files[0];
    if (!file) return;

    if (!file.type.startsWith('image/')) {
      addNotification({
        type: 'error',
        message: t('coupon_activation.please_select_image'),
      });
      return;
    }

    if (file.size > 10 * 1024 * 1024) {
      addNotification({
        type: 'error',
        message: t('coupon_activation.file_size_error'),
      });
      return;
    }

    setSelectedFile(file);

    const reader = new FileReader();
    reader.onload = e => {
      setPreviewUrl(e.target.result);
      setShowEditor(true);
      setActivationStep('edit');
    };
    reader.readAsDataURL(file);
  };

  const handleEditorSave = (editedUrl, editorParams) => {
    setEditedImageUrl(editedUrl);
    setShowEditor(false);
    
    saveImageData({
      file: selectedFile,
      previewUrl: previewUrl,
      editedImageUrl: editedUrl,
      editorParams: editorParams
    });

    navigateToPreviewGeneration();
  };

  const handleEditorCancel = () => {
    setShowEditor(false);
    setSelectedFile(null);
    setPreviewUrl(null);
    setEditedImageUrl(null);
    setActivationStep('upload');
  };

  const navigateToPreviewGeneration = () => {
    const finalImageUrl = editedImageUrl || previewUrl;
    
    sessionStorage.setItem('diamondMosaic_fileUrl', finalImageUrl);
    
    navigate('/preview/album', {
      state: {
        couponId: couponData.id,
        couponCode: couponCode,
        size: couponData.size,
        style: couponData.style,
        imageUrl: finalImageUrl
      }
    });
  };

  const handleSearchPage = async () => {
    if (!searchPageNumber) {
      addNotification({
        type: 'error',
        message: t('coupon_activation.please_enter_page_number'),
      });
      return;
    }

    const pageNum = parseInt(searchPageNumber);
    if (isNaN(pageNum) || pageNum < 1) {
      addNotification({
        type: 'error',
        message: t('coupon_activation.invalid_page_number'),
      });
      return;
    }

    try {
      if (couponData?.id) {
        const result = await MosaicAPI.searchSchemaPage(couponData.id, pageNum);
        if (result.found && result.page_url) {
          window.open(result.page_url, '_blank');
          addNotification({
            type: 'success',
            message: t('coupon_activation.page_opened', { page: pageNum, total: result.total_pages }),
          });
        } else {
          addNotification({
            type: 'warning',
            message: t('coupon_activation.page_not_found', { page: pageNum }),
          });
        }
      } else if (couponData?.zip_url) {
        const pageUrl = `${couponData.zip_url}#page=${searchPageNumber}`;
        window.open(pageUrl, '_blank');
      } else {
        addNotification({
          type: 'error',
          message: t('coupon_activation.archive_not_available'),
        });
      }
    } catch (error) {

      if (couponData?.zip_url) {
        const pageUrl = `${couponData.zip_url}#page=${searchPageNumber}`;
        window.open(pageUrl, '_blank');
      } else {
        addNotification({
          type: 'error',
          message: t('coupon_activation.page_search_error'),
        });
      }
    }
  };

  const handleResendEmail = async () => {
    if (!resendEmail || !couponData?.id) {
      addNotification({
        type: 'error',
        message: t('coupon_activation.please_enter_email'),
      });
      return;
    }

    setIsSendingEmail(true);
    try {
      await MosaicAPI.sendSchemaToEmail(couponData.id, { email: resendEmail });
      addNotification({
        type: 'success',
        message: t('coupon_activation.archive_sent', { email: resendEmail }),
      });
      setResendEmail('');
    } catch (error) {
      addNotification({
        type: 'error',
        message: t('coupon_activation.email_send_error'),
      });
    } finally {
      setIsSendingEmail(false);
    }
  };

  const handleDownloadArchive = () => {
    if (couponData?.zip_url) {
      window.open(couponData.zip_url, '_blank');
      addNotification({
        type: 'success',
        message: t('coupon_activation.download_started'),
      });
    } else {
      addNotification({
        type: 'error',
        message: t('coupon_activation.archive_not_available'),
      });
    }
  };

  const renderCouponInput = () => (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className="max-w-md mx-auto"
    >
      <div className="bg-white rounded-xl sm:rounded-2xl shadow-lg p-6 sm:p-8">
        <h2 className="text-xl sm:text-2xl font-bold text-gray-900 mb-4 sm:mb-6 text-center leading-tight">
          {t('coupon_activation.enter_coupon_code')}
        </h2>

        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2 leading-tight">
              {t('coupon_activation.coupon_code_label')}
            </label>
            <input
              type="text"
              value={couponCode}
              onChange={e => setCouponCode(e.target.value.toUpperCase())}
              placeholder={t('coupon_activation.coupon_placeholder')}
              className="w-full px-3 sm:px-4 py-3 border border-gray-300 rounded-lg text-base focus:ring-2 focus:ring-purple-500 focus:border-transparent touch-target"
              onKeyPress={e => e.key === 'Enter' && validateCoupon()}
              style={{ fontSize: '16px' }}
            />
          </div>

          <button
            onClick={validateCoupon}
            disabled={isValidating || !couponCode.trim()}
            className="w-full bg-gradient-to-r from-purple-600 to-pink-600 text-white py-3 rounded-lg text-sm sm:text-base font-semibold hover:from-purple-700 hover:to-pink-700 active:from-purple-800 active:to-pink-800 transition-all disabled:opacity-50 disabled:cursor-not-allowed touch-target"
          >
            {isValidating ? t('coupon_activation.validating') : t('coupon_activation.validate_coupon')}
          </button>
        </div>

        <div className="mt-4 sm:mt-6 text-center text-xs sm:text-sm text-gray-600 leading-relaxed">
          <p>{t('coupon_activation.coupon_hint')}</p>
        </div>
      </div>
    </motion.div>
  );

  const renderImageUpload = () => (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className="max-w-2xl mx-auto"
    >
      <div className="bg-white rounded-xl sm:rounded-2xl shadow-lg p-4 sm:p-6 lg:p-8">
        <div className="mb-4 sm:mb-6">
          <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between mb-3 sm:mb-4 gap-2 sm:gap-4">
            <h2 className="text-lg sm:text-xl lg:text-2xl font-bold text-gray-900 leading-tight">
              {t('coupon_activation.upload_photo')}
            </h2>
            <div className="flex items-center text-green-600">
              <Check className="w-4 h-4 sm:w-5 sm:h-5 mr-2 flex-shrink-0" />
              <span className="font-medium text-sm sm:text-base">
                {t('coupon_activation.coupon')}: {couponCode}
              </span>
            </div>
          </div>

          <div className="bg-blue-50 border border-blue-200 rounded-lg p-3 sm:p-4">
            <div className="flex items-start">
              <Package className="w-4 h-4 sm:w-5 sm:h-5 text-blue-600 mt-1 mr-2 sm:mr-3 flex-shrink-0" />
              <div className="flex-1 min-w-0">
                <p className="text-xs sm:text-sm font-medium text-blue-900 leading-tight">
                  {t('coupon_activation.mosaic_size')}: {couponData?.size || t('coupon_activation.standard')}
                </p>
                <p className="text-xs sm:text-sm text-blue-700 mt-1 leading-tight">
                  {t('coupon_activation.style')}: {couponData?.style || t('coupon_activation.classic')}
                </p>
              </div>
            </div>
          </div>
        </div>

        <div className="border-2 border-dashed border-gray-300 rounded-xl p-6 sm:p-8 lg:p-12 text-center hover:border-purple-400 active:border-purple-500 transition-colors touch-manipulation">
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
            className="cursor-pointer block touch-target"
          >
            <Upload className="w-12 h-12 sm:w-16 sm:h-16 text-gray-400 mx-auto mb-3 sm:mb-4" />
            <p className="text-lg sm:text-xl font-medium text-gray-700 mb-2 leading-tight">
              {t('coupon_activation.select_image')}
            </p>
            <p className="text-sm sm:text-base text-gray-500 mb-3 sm:mb-4 leading-relaxed">
              {t('coupon_activation.file_formats')}
            </p>
            <div className="inline-block bg-purple-600 text-white px-4 sm:px-6 py-3 rounded-lg hover:bg-purple-700 active:bg-purple-800 transition-colors font-semibold text-sm sm:text-base touch-target">
              {t('coupon_activation.choose_file')}
            </div>
          </label>
        </div>
      </div>
    </motion.div>
  );

  const renderImageEditor = () => (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className="max-w-6xl mx-auto"
    >
      <div className="mb-3 sm:mb-4">
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between bg-white rounded-lg px-3 sm:px-4 py-2 sm:py-3 shadow gap-2 sm:gap-4">
          <h2 className="text-lg sm:text-xl font-semibold text-gray-900 leading-tight">
            {t('coupon_activation.step_2_edit_image')}
          </h2>
          <div className="flex items-center text-green-600">
            <Check className="w-4 h-4 sm:w-5 sm:h-5 mr-2 flex-shrink-0" />
            <span className="font-medium text-sm sm:text-base">
              {t('coupon_activation.coupon')}: {couponCode}
            </span>
          </div>
        </div>
      </div>

      <ImageEditor
        imageUrl={previewUrl}
        onSave={handleEditorSave}
        onCancel={handleEditorCancel}
        title={t('diamond_mosaic_page.image_editor.setup_title')}
        showCropHint={true}
        aspectRatio={1}
        fileName={selectedFile?.name}
      />
    </motion.div>
  );

  const renderActivatedCoupon = () => (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className="max-w-4xl mx-auto"
    >
      <div className="bg-white rounded-xl sm:rounded-2xl shadow-lg overflow-hidden">
        {}
        <div className="bg-gradient-to-r from-green-500 to-emerald-600 p-4 sm:p-6 text-white">
          <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 sm:gap-4">
            <div className="flex-1 min-w-0">
              <h2 className="text-xl sm:text-2xl font-bold mb-2 leading-tight">
                {t('coupon_activation.coupon_activated')}
              </h2>
              <p className="text-green-100 text-sm sm:text-base">
                {t('coupon_activation.code')}: {couponCode}
              </p>
            </div>
            <div className="text-left sm:text-right flex-shrink-0">
              <Lock className="w-8 h-8 sm:w-12 sm:h-12 text-white/20 mb-1 sm:mb-2" />
              <p className="text-xs sm:text-sm text-green-100 leading-tight">
                {couponData?.activated_at
                  ? new Date(couponData.activated_at).toLocaleDateString(
                      t('common.locale')
                    )
                  : t('coupon_activation.activation_date_unknown')}
              </p>
            </div>
          </div>
        </div>

        {}
        <div className="p-4 sm:p-6">
          {}
          {couponData?.preview_image_url && (
            <div className="mb-4 sm:mb-6">
              <h3 className="text-base sm:text-lg font-semibold text-gray-900 mb-2 sm:mb-3 leading-tight">
                {t('coupon_activation.preview_of_mosaic')}
              </h3>
              <div className="bg-gray-50 rounded-lg p-3 sm:p-4">
                <img
                  src={couponData.preview_image_url}
                  alt={t('coupon_activation.mosaic_preview')}
                  className="w-full max-w-sm sm:max-w-md mx-auto rounded-lg shadow"
                />
              </div>
            </div>
          )}

          {}
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3 sm:gap-4 mb-4 sm:mb-6">
            <div className="bg-purple-50 rounded-lg p-3 sm:p-4">
              <div className="flex items-center">
                <Package className="w-6 h-6 sm:w-8 sm:h-8 text-purple-600 mr-2 sm:mr-3 flex-shrink-0" />
                <div className="flex-1 min-w-0">
                  <p className="text-xs sm:text-sm text-gray-600 leading-tight">
                    {t('coupon_activation.size')}
                  </p>
                  <p className="text-sm sm:text-lg font-semibold text-gray-900 leading-tight">
                    {couponData?.size || t('coupon_activation.not_specified')}
                  </p>
                </div>
              </div>
            </div>

            <div className="bg-blue-50 rounded-lg p-3 sm:p-4">
              <div className="flex items-center">
                <FileText className="w-6 h-6 sm:w-8 sm:h-8 text-blue-600 mr-2 sm:mr-3 flex-shrink-0" />
                <div className="flex-1 min-w-0">
                  <p className="text-xs sm:text-sm text-gray-600 leading-tight">
                    {t('coupon_activation.pages_in_schema')}
                  </p>
                  <p className="text-sm sm:text-lg font-semibold text-gray-900 leading-tight">
                    {couponData?.page_count || t('coupon_activation.not_specified')}
                  </p>
                </div>
              </div>
            </div>

            <div className="bg-green-50 rounded-lg p-3 sm:p-4">
              <div className="flex items-center">
                <Calendar className="w-6 h-6 sm:w-8 sm:h-8 text-green-600 mr-2 sm:mr-3 flex-shrink-0" />
                <div className="flex-1 min-w-0">
                  <p className="text-xs sm:text-sm text-gray-600 leading-tight">
                    {t('coupon_activation.stones')}
                  </p>
                  <p className="text-sm sm:text-lg font-semibold text-gray-900 leading-tight">
                    {couponData?.stones_count || t('coupon_activation.not_specified')}
                  </p>
                </div>
              </div>
            </div>
          </div>

          {}
          <div className="space-y-3 sm:space-y-4">
            {}
            {couponData?.zip_url && (
              <div className="border border-gray-200 rounded-lg p-3 sm:p-4">
                <h4 className="font-semibold text-gray-900 mb-2 sm:mb-3 text-sm sm:text-base leading-tight">
                  {t('coupon_activation.download_archive')}
                </h4>
                <button
                  onClick={handleDownloadArchive}
                  className="flex items-center justify-center w-full sm:w-auto bg-purple-600 text-white px-4 sm:px-6 py-3 rounded-lg hover:bg-purple-700 active:bg-purple-800 transition-colors text-sm sm:text-base font-medium touch-target"
                >
                  <Download className="w-4 h-4 sm:w-5 sm:h-5 mr-2 flex-shrink-0" />
                  {t('coupon_activation.download_archive_button')}
                </button>
              </div>
            )}

            {}
            <div className="border border-gray-200 rounded-lg p-3 sm:p-4">
              <h4 className="font-semibold text-gray-900 mb-2 sm:mb-3 text-sm sm:text-base leading-tight">
                {t('coupon_activation.send_to_email')}
              </h4>
              <div className="flex flex-col sm:flex-row gap-2 sm:gap-3">
                <input
                  type="email"
                  value={resendEmail}
                  onChange={e => setResendEmail(e.target.value)}
                  placeholder="example@email.com"
                  className="flex-1 px-3 sm:px-4 py-2 sm:py-3 border border-gray-300 rounded-lg text-base focus:ring-2 focus:ring-purple-500 focus:border-transparent touch-target"
                  style={{ fontSize: '16px' }}
                />
                <button
                  onClick={handleResendEmail}
                  disabled={isSendingEmail || !resendEmail}
                  className="flex items-center justify-center w-full sm:w-auto bg-blue-600 text-white px-4 sm:px-6 py-2 sm:py-3 rounded-lg hover:bg-blue-700 active:bg-blue-800 transition-colors disabled:opacity-50 disabled:cursor-not-allowed text-sm sm:text-base font-medium touch-target"
                >
                  <Mail className="w-4 h-4 sm:w-5 sm:h-5 mr-2 flex-shrink-0" />
                  {isSendingEmail ? t('coupon_activation.sending') : t('coupon_activation.send')}
                </button>
              </div>
            </div>

            {}
            {couponData?.page_count > 0 && (
              <div className="border border-gray-200 rounded-lg p-3 sm:p-4">
                <h4 className="font-semibold text-gray-900 mb-2 sm:mb-3 text-sm sm:text-base leading-tight">
                  {t('coupon_activation.search_page_in_archive')}
                </h4>
                <div className="flex flex-col sm:flex-row gap-2 sm:gap-3">
                  <input
                    type="number"
                    value={searchPageNumber}
                    onChange={e => setSearchPageNumber(e.target.value)}
                    placeholder={t('coupon_activation.page_number')}
                    min="1"
                    max={couponData.page_count}
                    className="flex-1 px-3 sm:px-4 py-2 sm:py-3 border border-gray-300 rounded-lg text-base focus:ring-2 focus:ring-purple-500 focus:border-transparent touch-target"
                    style={{ fontSize: '16px' }}
                  />
                  <button
                    onClick={handleSearchPage}
                    disabled={!searchPageNumber}
                    className="flex items-center justify-center w-full sm:w-auto bg-green-600 text-white px-4 sm:px-6 py-2 sm:py-3 rounded-lg hover:bg-green-700 active:bg-green-800 transition-colors disabled:opacity-50 disabled:cursor-not-allowed text-sm sm:text-base font-medium touch-target"
                  >
                    <Search className="w-4 h-4 sm:w-5 sm:h-5 mr-2 flex-shrink-0" />
                    {t('coupon_activation.find_page')}
                  </button>
                </div>
                <p className="text-xs sm:text-sm text-gray-600 mt-2 leading-relaxed">
                  {t('coupon_activation.total_pages')}: {couponData.page_count}
                </p>
              </div>
            )}
          </div>
        </div>
      </div>
    </motion.div>
  );

  return (
    <div className="min-h-screen bg-gradient-to-br from-purple-50 via-pink-50 to-blue-50">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 sm:py-12">
        {}
        <motion.div
          initial={{ opacity: 0, y: -20 }}
          animate={{ opacity: 1, y: 0 }}
          className="text-center mb-8 sm:mb-12"
        >
          <h1 className="text-2xl sm:text-3xl md:text-4xl lg:text-5xl font-bold text-gray-900 mb-3 sm:mb-4 leading-tight">
            {t('coupon_activation.title')}
          </h1>
          <p className="text-base sm:text-lg md:text-xl text-gray-600 max-w-2xl mx-auto leading-relaxed">
            {activationStep === 'activated'
              ? t('coupon_activation.coupon_already_activated')
              : t('coupon_activation.enter_code_and_upload')}
          </p>
        </motion.div>

        {}
        {activationStep === 'input' && renderCouponInput()}
        {activationStep === 'upload' && renderImageUpload()}
        {activationStep === 'edit' && showEditor && renderImageEditor()}
        {activationStep === 'activated' && renderActivatedCoupon()}
      </div>
    </div>
  );
};

export default CouponActivationPage;
