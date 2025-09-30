import React from 'react';
import { useTranslation } from 'react-i18next';
import { motion } from 'framer-motion';
import { Sparkles, Gem, Heart, ArrowRight, CheckCircle } from 'lucide-react';
import { useNavigate } from 'react-router-dom';

const WhatIsThisPage = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();

  const containerVariants = {
    hidden: { opacity: 0 },
    visible: {
      opacity: 1,
      transition: {
        staggerChildren: 0.2,
        duration: 0.6,
      },
    },
  };

  const itemVariants = {
    hidden: { opacity: 0, y: 20 },
    visible: { opacity: 1, y: 0 },
  };

  const goToHome = () => {
    navigate('/');
  };

  const goToDiamondArt = () => {
    navigate('/diamond-art');
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 via-purple-50 to-pink-50">
      {}
      <section className="relative overflow-hidden py-16 sm:py-20 lg:py-24">
        <div className="absolute inset-0 opacity-40">
          <div
            className="w-full h-full"
            style={{
              backgroundImage: `url("data:image/svg+xml,%3Csvg width='60' height='60' viewBox='0 0 60 60' xmlns='http://www.w3.org/2000/svg'%3E%3Cg fill='none' fill-rule='evenodd'%3E%3Cg fill='%239C92AC' fill-opacity='0.1'%3E%3Ccircle cx='30' cy='30' r='4'/%3E%3C/g%3E%3C/g%3E%3C/svg%3E")`,
            }}
          />
        </div>

        <div className="relative max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <motion.div
            variants={containerVariants}
            initial="hidden"
            animate="visible"
            className="text-center"
          >
            <motion.div
              variants={itemVariants}
              className="flex items-center justify-center w-20 h-20 sm:w-24 sm:h-24 bg-brand-primary/10 rounded-full mx-auto mb-6 sm:mb-8"
            >
              <Sparkles className="w-10 h-10 sm:w-12 sm:h-12 text-brand-primary" />
            </motion.div>

            <motion.h1
              variants={itemVariants}
              className="text-4xl sm:text-5xl md:text-6xl lg:text-7xl font-bold text-gray-900 mb-6 sm:mb-8 leading-tight px-2"
            >
              {t('what_is_this.title')}
            </motion.h1>

            <motion.p
              variants={itemVariants}
              className="text-xl sm:text-2xl md:text-3xl text-gray-600 mb-8 sm:mb-12 max-w-4xl mx-auto px-4"
            >
              {t('what_is_this.subtitle')}
            </motion.p>

            <motion.div
              variants={itemVariants}
              className="flex flex-col sm:flex-row gap-4 justify-center items-center"
            >
              <button
                onClick={goToDiamondArt}
                className="inline-flex items-center space-x-2 bg-brand-primary text-white py-4 px-8 rounded-xl hover:bg-brand-primary/90 font-semibold text-lg sm:text-xl transition-all duration-200 focus:ring-2 focus:ring-brand-primary focus:ring-offset-2"
              >
                <Gem className="w-5 h-5 sm:w-6 sm:h-6" />
                <span>{t('what_is_this.try_diamond_art')}</span>
              </button>

              <button
                onClick={goToHome}
                className="inline-flex items-center space-x-2 bg-white text-brand-primary py-4 px-8 rounded-xl hover:bg-gray-100 font-semibold text-lg sm:text-xl transition-all duration-200 focus:ring-2 focus:ring-white focus:ring-offset-2 border border-brand-primary"
              >
                <ArrowRight className="w-5 h-5 sm:w-6 sm:h-6" />
                <span>{t('what_is_this.back_to_home')}</span>
              </button>
            </motion.div>
          </motion.div>
        </div>
      </section>

      {}
      <section className="py-16 sm:py-20 lg:py-24 bg-white/50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <motion.div
            variants={containerVariants}
            initial="hidden"
            animate="visible"
            className="grid grid-cols-1 lg:grid-cols-2 gap-12 lg:gap-16 items-center"
          >
            <motion.div variants={itemVariants}>
              <h2 className="text-3xl sm:text-4xl md:text-5xl font-bold text-gray-900 mb-6 sm:mb-8">
                {t('what_is_this.diamond_art.title')}
              </h2>
              <p className="text-lg sm:text-xl text-gray-600 mb-8">
                {t('what_is_this.diamond_art.description')}
              </p>
              <ul className="space-y-4">
                {['step1', 'step2', 'step3'].map((step, index) => (
                  <li key={step} className="flex items-start space-x-3">
                    <CheckCircle className="w-6 h-6 text-brand-primary mt-1 flex-shrink-0" />
                    <span className="text-gray-700 text-lg">
                      {t(`what_is_this.diamond_art.${step}`)}
                    </span>
                  </li>
                ))}
              </ul>
            </motion.div>

            <motion.div variants={itemVariants} className="text-center">
              <div className="bg-gradient-to-br from-brand-primary/20 to-brand-secondary/20 rounded-3xl p-12 sm:p-16">
                <Gem className="w-24 h-24 sm:w-32 sm:h-32 text-brand-primary mx-auto mb-6" />
                <h3 className="text-2xl sm:text-3xl font-bold text-gray-900 mb-4">
                  {t('what_is_this.diamond_art.highlight_title')}
                </h3>
                <p className="text-gray-600 text-lg">
                  {t('what_is_this.diamond_art.highlight_description')}
                </p>
              </div>
            </motion.div>
          </motion.div>
        </div>
      </section>

      <section className="py-16 sm:py-20 lg:py-24">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <motion.div
            variants={containerVariants}
            initial="hidden"
            animate="visible"
            className="text-center mb-16"
          >
            <motion.h2
              variants={itemVariants}
              className="text-3xl sm:text-4xl md:text-5xl font-bold text-gray-900 mb-6 sm:mb-8"
            >
              {t('what_is_this.benefits.title')}
            </motion.h2>
            <motion.p
              variants={itemVariants}
              className="text-xl sm:text-2xl text-gray-600 max-w-3xl mx-auto"
            >
              {t('what_is_this.benefits.subtitle')}
            </motion.p>
          </motion.div>

          <motion.div
            variants={containerVariants}
            initial="hidden"
            animate="visible"
            className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8 sm:gap-10 lg:gap-12"
          >
            {['relaxation', 'creativity', 'result'].map((benefit, index) => (
              <motion.div
                key={benefit}
                variants={itemVariants}
                className="bg-white/80 backdrop-blur-sm rounded-2xl shadow-xl p-8 border border-white/20 text-center"
              >
                <div
                  className={`flex items-center justify-center w-16 h-16 ${
                    index === 0
                      ? 'bg-brand-primary/10'
                      : index === 1
                        ? 'bg-brand-secondary/10'
                        : 'bg-brand-accent/10'
                  } rounded-full mx-auto mb-6`}
                >
                  {index === 0 ? (
                    <Heart className="w-8 h-8 text-brand-primary" />
                  ) : index === 1 ? (
                    <Sparkles className="w-8 h-8 text-brand-secondary" />
                  ) : (
                    <Gem className="w-8 h-8 text-brand-accent" />
                  )}
                </div>
                <h3 className="text-2xl font-bold text-gray-900 mb-4">
                  {t(`what_is_this.benefits.${benefit}.title`)}
                </h3>
                <p className="text-gray-600 text-lg">
                  {t(`what_is_this.benefits.${benefit}.description`)}
                </p>
              </motion.div>
            ))}
          </motion.div>
        </div>
      </section>

      {}
      <section className="py-16 sm:py-20 lg:py-24">
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 text-center">
          <motion.div
            variants={containerVariants}
            initial="hidden"
            animate="visible"
          >
            <motion.div
              variants={itemVariants}
              className="bg-gradient-to-r from-brand-primary to-brand-secondary text-white rounded-3xl p-12 sm:p-16 shadow-2xl"
            >
              <h2 className="text-3xl sm:text-4xl md:text-5xl font-bold mb-6 sm:mb-8">
                {t('what_is_this.cta.title')}
              </h2>
              <p className="text-xl sm:text-2xl text-white/90 mb-8 sm:mb-10 max-w-2xl mx-auto">
                {t('what_is_this.cta.description')}
              </p>
              <button
                onClick={goToDiamondArt}
                className="bg-white text-brand-primary py-4 px-8 rounded-xl hover:bg-gray-100 font-semibold text-lg sm:text-xl transition-all duration-200 focus:ring-2 focus:ring-white focus:ring-offset-2 focus:ring-offset-brand-primary"
              >
                {t('what_is_this.cta.button')}
              </button>
            </motion.div>
          </motion.div>
        </div>
      </section>
    </div>
  );
};

export default WhatIsThisPage;
