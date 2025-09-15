import React, { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { motion, AnimatePresence } from 'framer-motion';
import { Sparkles, Clock, CheckCircle, AlertCircle } from 'lucide-react';

const ProcessingIndicator = ({
  isProcessing,
  useAI,
  currentStep = 0,
  error = null,
}) => {
  const { t } = useTranslation();
  const [step, setStep] = useState(0);

  const steps = [
    {
      key: 'uploading',
      icon: '📤',
      label: t('editor.processing_steps.uploading', 'Uploading image...'),
    },
    {
      key: 'processing',
      icon: '🤖',
      label: t(
        'editor.processing_steps.processing',
        'Processing through AI...'
      ),
    },
    {
      key: 'enhancing',
      icon: '✨',
      label: t('editor.processing_steps.enhancing', 'Enhancing quality...'),
    },
    {
      key: 'finalizing',
      icon: '🎯',
      label: t(
        'editor.processing_steps.finalizing',
        'Finalizing processing...'
      ),
    },
  ];

  useEffect(() => {
    if (isProcessing) {
      const interval = setInterval(() => {
        setStep(prev => {
          if (prev < steps.length - 1) {
            return prev + 1;
          }
          return prev;
        });
      }, 2000);
      return () => clearInterval(interval);
    } else {
      setStep(0);
    }
  }, [isProcessing, steps.length]);

  if (!isProcessing) return null;

  if (error) {
    return (
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        className="text-center py-6 sm:py-8 px-4 sm:px-6 bg-red-50 rounded-xl border border-red-200"
      >
        <div className="w-12 h-12 sm:w-16 sm:h-16 bg-red-100 rounded-full flex items-center justify-center mx-auto mb-3 sm:mb-4">
          <AlertCircle className="w-6 h-6 sm:w-8 sm:h-8 text-red-600" />
        </div>
        <h3 className="text-base sm:text-lg font-semibold text-red-800 mb-2 leading-tight">
          {t('editor.processing_error', 'Processing Error')}
        </h3>
        <p className="text-red-600 text-xs sm:text-sm max-w-md mx-auto leading-relaxed">
          {error}
        </p>
      </motion.div>
    );
  }

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className="text-center py-6 sm:py-8 px-4 sm:px-6 bg-gradient-to-br from-brand-primary/5 to-brand-secondary/5 rounded-xl border border-brand-primary/20"
    >
      <div className="mb-4 sm:mb-6">
        <div className="w-16 h-16 sm:w-20 sm:h-20 bg-gradient-to-br from-brand-primary to-brand-secondary rounded-full flex items-center justify-center mx-auto mb-3 sm:mb-4">
          {useAI ? (
            <Sparkles className="w-8 h-8 sm:w-10 sm:h-10 text-white animate-pulse" />
          ) : (
            <CheckCircle className="w-8 h-8 sm:w-10 sm:h-10 text-white" />
          )}
        </div>

        <h3 className="text-lg sm:text-xl font-semibold text-gray-900 mb-2 leading-tight">
          {useAI
            ? t('editor.ai_processing', 'AI Image Enhancement...')
            : t('editor.processing', 'Processing Image...')}
        </h3>

        {useAI && (
          <p className="text-gray-600 text-xs sm:text-sm leading-relaxed">
            {t(
              'editor.ai_processing_desc',
              'Using Stable Diffusion to improve quality'
            )}
          </p>
        )}
      </div>

      <div className="max-w-md mx-auto mb-4 sm:mb-6">
        <div className="space-y-2 sm:space-y-3">
          {steps.map((stepItem, index) => (
            <motion.div
              key={stepItem.key}
              initial={{ opacity: 0, x: -20 }}
              animate={{
                opacity: index <= step ? 1 : 0.3,
                x: 0,
              }}
              transition={{ delay: index * 0.1 }}
              className={`flex items-center space-x-2 sm:space-x-3 p-2 sm:p-3 rounded-lg transition-all duration-300 ${
                index <= step
                  ? 'bg-white shadow-sm border border-brand-primary/20'
                  : 'bg-gray-50'
              }`}
            >
              <div
                className={`w-6 h-6 sm:w-8 sm:h-8 rounded-full flex items-center justify-center text-sm sm:text-lg flex-shrink-0 ${
                  index < step
                    ? 'bg-brand-secondary text-white'
                    : index === step
                      ? 'bg-brand-primary text-white animate-pulse'
                      : 'bg-gray-300 text-gray-600'
                }`}
              >
                {index < step ? '✓' : stepItem.icon}
              </div>

              <span
                className={`font-medium text-xs sm:text-sm flex-1 min-w-0 ${
                  index <= step ? 'text-gray-900' : 'text-gray-500'
                }`}
              >
                {stepItem.label}
              </span>

              {index === step && (
                <motion.div
                  animate={{ rotate: 360 }}
                  transition={{ duration: 1, repeat: Infinity, ease: 'linear' }}
                  className="ml-auto flex-shrink-0"
                >
                  <div className="w-3 h-3 sm:w-4 sm:h-4 border-2 border-brand-primary border-t-transparent rounded-full" />
                </motion.div>
              )}
            </motion.div>
          ))}
        </div>
      </div>

      <div className="flex items-center justify-center space-x-2 text-xs sm:text-sm text-gray-600">
        <Clock className="w-3 h-3 sm:w-4 sm:h-4 flex-shrink-0" />
        <span className="text-center">
          {useAI
            ? t('editor.processing_time', 'Estimated time: 2-5 minutes')
            : t('editor.processing_time_fast', 'Estimated time: 1-2 minutes')}
        </span>
      </div>

      {useAI && (
        <div className="mt-3 sm:mt-4 p-2 sm:p-3 bg-brand-primary/10 rounded-lg max-w-md mx-auto">
          <p className="text-xs text-brand-primary/80 text-center leading-relaxed">
            {t(
              'editor.processing_note',
              'AI processing may take some time depending on system load'
            )}
          </p>
        </div>
      )}
    </motion.div>
  );
};

export default ProcessingIndicator;
