import React from 'react';
import { motion } from 'framer-motion';
import { useTranslation } from 'react-i18next';

const LoadingScreen = () => {
  const { t } = useTranslation();
  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-brand-primary/5 to-brand-secondary/5 px-4">
      <div className="text-center">
        <motion.div
          animate={{
            rotate: 360,
            scale: [1, 1.2, 1],
          }}
          transition={{
            rotate: { duration: 2, repeat: Infinity, ease: 'linear' },
            scale: { duration: 1.5, repeat: Infinity, ease: 'easeInOut' },
          }}
          className="w-12 h-12 sm:w-16 sm:h-16 mx-auto mb-4 sm:mb-6 bg-gradient-to-r from-brand-primary to-brand-secondary rounded-full flex items-center justify-center"
        >
          <div className="w-6 h-6 sm:w-8 sm:h-8 bg-white rounded-full opacity-80"></div>
        </motion.div>

        <motion.h2
          animate={{ opacity: [0.5, 1, 0.5] }}
          transition={{ duration: 2, repeat: Infinity }}
          className="text-lg sm:text-xl font-semibold text-gray-700 mb-2 leading-tight"
        >
          {t('loading.title')}
        </motion.h2>

        <p className="text-sm sm:text-base text-gray-500 leading-relaxed">
          {t('loading.subtitle')}
        </p>
      </div>
    </div>
  );
};

export default LoadingScreen;
