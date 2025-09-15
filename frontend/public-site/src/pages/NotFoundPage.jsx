import React from 'react';
import { Link } from 'react-router-dom';
import { motion } from 'framer-motion';
import { Home, ArrowLeft } from 'lucide-react';
import { useTranslation } from 'react-i18next';

const NotFoundPage = () => {
  const { t } = useTranslation();

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-brand-primary/5 to-brand-secondary/5 px-4">
      <div className="text-center max-w-lg mx-auto">
        <motion.div
          initial={{ opacity: 0, scale: 0.8 }}
          animate={{ opacity: 1, scale: 1 }}
          transition={{ duration: 0.6 }}
          className="mb-6 sm:mb-8"
        >
          <div className="text-6xl sm:text-8xl md:text-9xl font-bold text-transparent bg-gradient-to-r from-brand-primary to-brand-secondary bg-clip-text mb-3 sm:mb-4 leading-none">
            404
          </div>
          <h1 className="text-2xl sm:text-3xl md:text-4xl font-bold text-gray-900 mb-3 sm:mb-4 leading-tight">
            {t('not_found.title')}
          </h1>
          <p className="text-base sm:text-lg md:text-xl text-gray-600 mb-6 sm:mb-8 leading-relaxed">
            {t('not_found.description')}
          </p>
        </motion.div>

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6, delay: 0.3 }}
          className="flex flex-col sm:flex-row gap-3 sm:gap-4 sm:justify-center"
        >
          <Link
            to="/"
            className="inline-flex items-center justify-center px-4 sm:px-6 py-3 bg-brand-primary text-white rounded-lg hover:bg-brand-primary/90 active:bg-brand-primary/80 transition-colors text-sm sm:text-base font-medium touch-target"
          >
            <Home className="w-4 h-4 sm:w-5 sm:h-5 mr-2 flex-shrink-0" />
            {t('not_found.home_button')}
          </Link>

          <button
            onClick={() => window.history.back()}
            className="inline-flex items-center justify-center px-4 sm:px-6 py-3 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 active:bg-gray-100 transition-colors text-sm sm:text-base font-medium touch-target"
          >
            <ArrowLeft className="w-4 h-4 sm:w-5 sm:h-5 mr-2 flex-shrink-0" />
            {t('not_found.back_button')}
          </button>
        </motion.div>

        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ duration: 1, delay: 0.6 }}
          className="mt-8 sm:mt-12"
        >
          <div className="w-48 h-48 sm:w-64 sm:h-64 mx-auto opacity-20">
            <svg viewBox="0 0 100 100" className="w-full h-full">
              <defs>
                <pattern
                  id="diamond"
                  x="0"
                  y="0"
                  width="10"
                  height="10"
                  patternUnits="userSpaceOnUse"
                >
                  <circle
                    cx="5"
                    cy="5"
                    r="2"
                    fill="currentColor"
                    opacity="0.3"
                  />
                </pattern>
              </defs>
              <rect width="100" height="100" fill="url(#diamond)" />
            </svg>
          </div>
        </motion.div>
      </div>
    </div>
  );
};

export default NotFoundPage;
