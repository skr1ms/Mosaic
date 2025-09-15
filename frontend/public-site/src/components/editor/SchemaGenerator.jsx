import React, { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { motion, AnimatePresence } from 'framer-motion';
import {
  CheckCircle,
  AlertTriangle,
  Download,
  Mail,
  Loader,
} from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { useMutation } from '@tanstack/react-query';
import { MosaicAPI } from '../../api/client';
import { useUIStore } from '../../store/partnerStore';

const SchemaGenerator = ({
  imageId,
  selectedOptions,
  onBack,
  onComplete,
  onLockNavigation,
}) => {
  const { t } = useTranslation();
  const { addNotification } = useUIStore();
  const navigate = useNavigate();
  const [showWarning, setShowWarning] = useState(true);
  const [isGenerating, setIsGenerating] = useState(false);
  const [generationProgress, setGenerationProgress] = useState(0);
  const [schemaData, setSchemaData] = useState(null);
  const [email, setEmail] = useState('');
  const [showEmailForm, setShowEmailForm] = useState(false);

  const generateSchemaMutation = useMutation({
    mutationFn: async params => {
      return await MosaicAPI.generateSchema(imageId, params);
    },
    onSuccess: data => {},
    onError: error => {
      setIsGenerating(false);
      addNotification({
        type: 'error',
        title: t('notifications.schema_error'),
        message: error.message || t('notifications.schema_error_desc'),
      });
    },
  });

  const sendEmailMutation = useMutation({
    mutationFn: async emailData => {
      return await MosaicAPI.sendSchemaToEmail(imageId, emailData);
    },
    onSuccess: () => {
      addNotification({
        type: 'success',
        title: t('notifications.email_sent'),
        message: t('notifications.email_sent_desc'),
      });
      setShowEmailForm(false);
    },
    onError: error => {
      addNotification({
        type: 'error',
        title: t('notifications.email_error'),
        message: error.message || t('notifications.email_error_desc'),
      });
    },
  });

  const handleGenerateSchema = async () => {
    try {
      setIsGenerating(true);
      setGenerationProgress(0);
      try {
        sessionStorage.setItem(`editor:confirmed:${imageId}`, '1');
      } catch {}
      if (onLockNavigation) onLockNavigation(true);

      let currentStatus = null;
      try {
        currentStatus = await MosaicAPI.getProcessingStatus(imageId);
      } catch (_) {}

      const mapParams = opts => {
        const styleMap = {
          original: { style: 'grayscale', use_ai: false },
          enhanced: { style: 'max_colors', use_ai: true },
        };
        const lightingMap = {
          natural: 'sun',
          moonlight: 'moon',
          venus: 'venus',
        };
        const contrastMap = { soft: 'low', strong: 'high' };
        const mapped = styleMap[opts.style] || styleMap.original;
        return {
          style: mapped.style,
          use_ai: mapped.use_ai,
          lighting: lightingMap[opts.lighting] || 'sun',
          contrast: contrastMap[opts.contrast] || undefined,
        };
      };

      try {
        const editParams = {
          crop_x: 0,
          crop_y: 0,
          crop_width: 0,
          crop_height: 0,
          rotation: 0,
          scale: 1,
        };

        try {
          const savedEdits = sessionStorage.getItem(`editor:edits:${imageId}`);
          if (savedEdits) {
            const edits = JSON.parse(savedEdits);
            Object.assign(editParams, edits);
          }
        } catch {}

        await MosaicAPI.editImage(imageId, editParams);
      } catch (error) {
        console.warn('Failed to apply edits:', error);
      }

      if (!currentStatus || currentStatus.status !== 'processed') {
        const processParams = mapParams(selectedOptions);
        await MosaicAPI.processImage(imageId, processParams);
      }

      const waitForProcessed = async () => {
        const start = Date.now();
        const TIMEOUT_MS = 600000;
        while (Date.now() - start < TIMEOUT_MS) {
          const status = await MosaicAPI.getProcessingStatus(imageId);
          if (status?.status === 'processed') return true;
          if (status?.status === 'failed')
            throw new Error(status?.error_message || 'Processing failed');
          await new Promise(r => setTimeout(r, 2000));
        }
        throw new Error('Processing timeout');
      };

      await waitForProcessed();

      await generateSchemaMutation.mutateAsync({ confirmed: true });

      const waitForCompleted = async () => {
        const start = Date.now();
        const TIMEOUT_MS = 600000;
        while (Date.now() - start < TIMEOUT_MS) {
          const status = await MosaicAPI.getProcessingStatus(imageId);
          if (status?.status === 'completed') return status;
          if (status?.status === 'failed')
            throw new Error(status?.error_message || 'Generation failed');
          await new Promise(r => setTimeout(r, 2000));
        }
        throw new Error('Generation timeout');
      };

      const finalStatus = await waitForCompleted();
      const newSchemaData = {
        schema_uuid: imageId,
        preview_url: finalStatus?.preview_url,
        schema_url: finalStatus?.schema_url,
      };
      console.log('Setting schemaData after completion:', newSchemaData);
      setSchemaData(newSchemaData);
      setIsGenerating(false);
      setGenerationProgress(100);

      addNotification({
        type: 'success',
        title: t('notifications.schema_generated'),
        message: t('notifications.schema_generated_desc'),
      });
    } catch (error) {
      setIsGenerating(false);
      if (onLockNavigation) onLockNavigation(false);
      addNotification({
        type: 'error',
        title: t('notifications.schema_error'),
        message: error.message || t('notifications.schema_error_desc'),
      });
    }
  };

  useEffect(() => {
    let intervalId;
    let isMounted = true;

    const pollStatus = async () => {
      try {
        const status = await MosaicAPI.getProcessingStatus(imageId);
        if (!isMounted) return;

        if (status?.status === 'completed') {
          setIsGenerating(false);
          setGenerationProgress(100);
          if (onLockNavigation) onLockNavigation(false);
          const newSchemaData = {
            schema_uuid: imageId,
            preview_url: status?.preview_url,
            schema_url: status?.schema_url,
          };
          console.log('Setting schemaData from polling:', newSchemaData);
          setSchemaData(newSchemaData);
        }
        if (status?.status === 'failed') {
          setIsGenerating(false);
          if (onLockNavigation) onLockNavigation(false);
        }
      } catch (e) {}
    };

    if (isGenerating && !schemaData) {
      pollStatus();
      intervalId = setInterval(pollStatus, 2000);
    }

    return () => {
      isMounted = false;
      if (intervalId) clearInterval(intervalId);
    };
  }, [isGenerating, schemaData, imageId, onLockNavigation]);

  useEffect(() => {
    try {
      const confirmed =
        sessionStorage.getItem(`editor:confirmed:${imageId}`) === '1';
      const savedSchemaData = sessionStorage.getItem(
        `editor:schemaData:${imageId}`
      );
      if (confirmed) {
        setShowWarning(false);
        if (savedSchemaData) {
          const parsed = JSON.parse(savedSchemaData);
          setSchemaData(parsed);
          setIsGenerating(false);
          setGenerationProgress(100);
        } else {
          setIsGenerating(true);
        }
      }
    } catch {}
  }, [imageId]);

  useEffect(() => {
    setShowWarning(true);
    setIsGenerating(false);
    setGenerationProgress(0);
    setSchemaData(null);
    setEmail('');
    setShowEmailForm(false);
  }, [imageId]);

  useEffect(() => {
    if (schemaData) {
      try {
        sessionStorage.setItem(
          `editor:schemaData:${imageId}`,
          JSON.stringify(schemaData)
        );
      } catch {}
    }
  }, [schemaData, imageId]);

  const handleDownloadSchema = () => {
    const uuid = schemaData?.schema_uuid || imageId;
    if (uuid) {
      MosaicAPI.downloadSchemaArchive(uuid);
    }
  };

  const handleSendEmail = e => {
    e.preventDefault();
    if (email) {
      sendEmailMutation.mutate({ email });
    }
  };

  useEffect(() => {
    if (onComplete && schemaData) {
      onComplete(schemaData);
    }
  }, [schemaData, onComplete]);

  return (
    <div className="space-y-6 sm:space-y-8">
      <AnimatePresence mode="wait">
        {showWarning && !isGenerating && !schemaData && (
          <motion.div
            key="warning"
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -20 }}
            transition={{ duration: 0.5 }}
            className="bg-brand-accent/5 border border-brand-accent/20 rounded-xl p-4 sm:p-6"
          >
            <div className="flex items-start space-x-2 sm:space-x-3">
              <div className="flex-shrink-0">
                <div className="w-6 h-6 sm:w-8 sm:h-8 bg-brand-accent/10 rounded-full flex items-center justify-center">
                  <AlertTriangle className="w-4 h-4 sm:w-5 sm:h-5 text-brand-accent" />
                </div>
              </div>
              <div className="flex-1 min-w-0">
                <h4 className="text-base sm:text-lg font-semibold text-brand-accent mb-2 leading-tight">
                  {t('editor.confirmation.warning_title')}
                </h4>
                <p className="text-sm sm:text-base text-brand-accent/80 mb-2 leading-relaxed">
                  {t('editor.confirmation.warning_text')}
                </p>
                <p className="text-sm sm:text-base text-brand-accent font-medium leading-relaxed">
                  {t('editor.confirmation.warning_note')}
                </p>
              </div>
            </div>
          </motion.div>
        )}

        {!isGenerating && !schemaData && (
          <motion.div
            key="confirmation"
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -20 }}
            transition={{ duration: 0.5 }}
            className="text-center px-4"
          >
            <h2 className="text-2xl sm:text-3xl font-bold text-gray-900 mb-4 sm:mb-6 leading-tight">
              {t('editor.confirmation.title')}
            </h2>

            <div className="max-w-2xl mx-auto bg-gray-50 rounded-xl p-4 sm:p-6 mb-6 sm:mb-8">
              <h3 className="text-base sm:text-lg font-semibold text-gray-900 mb-3 sm:mb-4 leading-tight">
                {t('editor.confirmation.selected_options')}
              </h3>

              <div className="grid grid-cols-1 sm:grid-cols-3 gap-3 sm:gap-4 text-xs sm:text-sm">
                <div className="text-center sm:text-left">
                  <span className="font-medium text-gray-700">
                    {t('editor.confirmation.style')}:
                  </span>
                  <p className="text-gray-900 mt-1">
                    {t(
                      `diamond_mosaic_styles.styles.${selectedOptions.style}.title`
                    )}
                  </p>
                </div>
                <div className="text-center sm:text-left">
                  <span className="font-medium text-gray-700">
                    {t('editor.confirmation.lighting')}:
                  </span>
                  <p className="text-gray-900 mt-1">
                    {t(`editor.lighting.${selectedOptions.lighting}.title`)}
                  </p>
                </div>
                <div className="text-center sm:text-left">
                  <span className="font-medium text-gray-700">
                    {t('editor.confirmation.contrast')}:
                  </span>
                  <p className="text-gray-900 mt-1">
                    {t(`editor.contrast.${selectedOptions.contrast}.title`)}
                  </p>
                </div>
              </div>
            </div>

            <div className="flex flex-col sm:flex-row justify-center gap-3 sm:gap-4">
              <button
                onClick={onBack}
                className="w-full sm:w-auto px-6 sm:px-8 py-3 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 active:bg-gray-100 transition-colors text-sm sm:text-base font-medium touch-target"
              >
                {t('editor.confirmation.no_button')}
              </button>

              <button
                onClick={handleGenerateSchema}
                className="w-full sm:w-auto px-6 sm:px-8 py-3 bg-brand-primary text-white rounded-lg hover:bg-brand-primary/90 active:bg-brand-primary/80 transition-colors text-sm sm:text-base font-medium touch-target"
              >
                {t('editor.confirmation.yes_button')}
              </button>
            </div>
          </motion.div>
        )}

        {isGenerating && (
          <motion.div
            key="generating"
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5 }}
            className="text-center px-4"
          >
            <h2 className="text-2xl sm:text-3xl font-bold text-gray-900 mb-6 sm:mb-8 leading-tight">
              {t('editor.generation.title')}
            </h2>

            <div className="max-w-md mx-auto">
              <div className="w-20 h-20 sm:w-24 sm:h-24 border-4 border-brand-primary/20 border-t-brand-primary rounded-full animate-spin mx-auto mb-4 sm:mb-6" />

              <h3 className="text-lg sm:text-xl font-semibold text-gray-900 mb-3 sm:mb-4 leading-tight">
                {t('editor.generation.processing')}
              </h3>

              <p className="text-sm sm:text-base text-gray-600 mb-4 sm:mb-6 leading-relaxed">
                {t('editor.generation.description')}
              </p>
            </div>
          </motion.div>
        )}

        {schemaData && (
          <motion.div
            key="complete"
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5 }}
            className="text-center px-4"
          >
            <div className="w-16 h-16 sm:w-20 sm:h-20 bg-brand-secondary/10 rounded-full flex items-center justify-center mx-auto mb-4 sm:mb-6">
              <CheckCircle className="w-8 h-8 sm:w-10 sm:h-10 text-brand-secondary" />
            </div>

            <h2 className="text-2xl sm:text-3xl font-bold text-gray-900 mb-3 sm:mb-4 leading-tight">
              {t('editor.completion.title')}
            </h2>

            <p className="text-base sm:text-lg text-gray-600 mb-6 sm:mb-8 max-w-2xl mx-auto leading-relaxed">
              {t('editor.completion.description')}
            </p>

            {}
            <div className="flex justify-center mb-6 sm:mb-8">
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 sm:gap-6 max-w-2xl w-full">
                <motion.button
                  whileHover={{ scale: 1.02 }}
                  whileTap={{ scale: 0.98 }}
                  onClick={handleDownloadSchema}
                  className="flex flex-col items-center p-4 sm:p-6 bg-brand-primary/5 rounded-xl hover:bg-brand-primary/10 active:bg-brand-primary/15 transition-colors touch-target"
                >
                  <Download className="w-10 h-10 sm:w-12 sm:h-12 text-brand-primary mb-3 sm:mb-4" />
                  <h3 className="text-base sm:text-lg font-semibold text-gray-900 mb-2 leading-tight">
                    {t('editor.completion.download_schema')}
                  </h3>
                  <p className="text-xs sm:text-sm text-gray-600 text-center leading-relaxed">
                    {t('editor.completion.download_schema_desc')}
                  </p>
                </motion.button>

                <motion.button
                  whileHover={{ scale: 1.02 }}
                  whileTap={{ scale: 0.98 }}
                  onClick={() => setShowEmailForm(true)}
                  className="flex flex-col items-center p-4 sm:p-6 bg-brand-secondary/5 rounded-xl hover:bg-brand-secondary/10 active:bg-brand-secondary/15 transition-colors touch-target"
                >
                  <Mail className="w-10 h-10 sm:w-12 sm:h-12 text-brand-secondary mb-3 sm:mb-4" />
                  <h3 className="text-base sm:text-lg font-semibold text-gray-900 mb-2 leading-tight">
                    {t('editor.completion.send_email')}
                  </h3>
                  <p className="text-xs sm:text-sm text-gray-600 text-center leading-relaxed">
                    {t('editor.completion.send_email_desc')}
                  </p>
                </motion.button>
              </div>
            </div>

            {}
            <div className="mt-6 sm:mt-8 flex justify-center">
              <button
                onClick={() => {
                  try {
                    Object.keys(sessionStorage).forEach(k => {
                      if (k.startsWith('editor:')) {
                        sessionStorage.removeItem(k);
                      }
                    });
                  } catch {}
                  navigate('/');
                }}
                className="w-full sm:w-auto px-6 sm:px-8 py-3 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 active:bg-gray-100 transition-colors text-sm sm:text-base font-medium touch-target"
              >
                {t('editor.completion.back_to_home')}
              </button>
            </div>

            {}
            <AnimatePresence>
              {showEmailForm && (
                <motion.div
                  initial={{ opacity: 0, height: 0 }}
                  animate={{ opacity: 1, height: 'auto' }}
                  exit={{ opacity: 0, height: 0 }}
                  transition={{ duration: 0.3 }}
                  className="max-w-md mx-auto mt-6"
                >
                  <form
                    onSubmit={handleSendEmail}
                    className="space-y-3 sm:space-y-4"
                  >
                    <div>
                      <label
                        htmlFor="email"
                        className="block text-xs sm:text-sm font-medium text-gray-700 mb-2 leading-tight"
                      >
                        {t('editor.completion.email_label')}
                      </label>
                      <input
                        type="email"
                        id="email"
                        value={email}
                        onChange={e => setEmail(e.target.value)}
                        placeholder={t('editor.completion.email_placeholder')}
                        className="w-full px-3 sm:px-4 py-2 sm:py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-brand-primary focus:border-transparent text-base touch-target"
                        style={{ fontSize: '16px' }}
                        required
                      />
                    </div>

                    <div className="flex flex-col sm:flex-row gap-2 sm:gap-3">
                      <button
                        type="button"
                        onClick={() => setShowEmailForm(false)}
                        className="flex-1 px-3 sm:px-4 py-2 sm:py-3 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 active:bg-gray-100 transition-colors text-sm sm:text-base font-medium touch-target"
                      >
                        {t('common.cancel')}
                      </button>

                      <button
                        type="submit"
                        disabled={sendEmailMutation.isPending}
                        className="flex-1 px-3 sm:px-4 py-2 sm:py-3 bg-brand-secondary text-white rounded-lg hover:bg-brand-secondary/90 active:bg-brand-secondary/80 disabled:opacity-50 transition-colors text-sm sm:text-base font-medium touch-target"
                      >
                        {sendEmailMutation.isPending ? (
                          <Loader className="w-4 h-4 animate-spin mx-auto" />
                        ) : (
                          t('editor.completion.send_button')
                        )}
                      </button>
                    </div>
                  </form>
                </motion.div>
              )}
            </AnimatePresence>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
};

export default SchemaGenerator;
