import React from 'react'
import { useTranslation } from 'react-i18next'
import { motion } from 'framer-motion'
import { Palette, Brush, Star, Clock, ArrowRight } from 'lucide-react'
import { useNavigate } from 'react-router-dom'

const PaintByNumbersPage = () => {
  const { t } = useTranslation()
  const navigate = useNavigate()

  const containerVariants = {
    hidden: { opacity: 0 },
    visible: {
      opacity: 1,
      transition: {
        staggerChildren: 0.2,
        duration: 0.6
      }
    }
  }

  const itemVariants = {
    hidden: { opacity: 0, y: 20 },
    visible: { opacity: 1, y: 0 }
  }

  const goToHome = () => {
    navigate('/')
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 via-purple-50 to-pink-50">
      {/* Hero Section */}
      <section className="relative overflow-hidden py-16 sm:py-20 lg:py-24">
        <div className="absolute inset-0 opacity-40">
          <div className="w-full h-full" style={{
            backgroundImage: `url("data:image/svg+xml,%3Csvg width='60' height='60' viewBox='0 0 60 60' xmlns='http://www.w3.org/2000/svg'%3E%3Cg fill='none' fill-rule='evenodd'%3E%3Cg fill='%239C92AC' fill-opacity='0.05'%3E%3Ccircle cx='30' cy='30' r='3'/%3E%3C/g%3E%3C/g%3E%3C/svg%3E")`
          }} />
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
              <Palette className="w-10 h-10 sm:w-12 sm:h-12 text-brand-primary" />
            </motion.div>
            
            <motion.h1 
              variants={itemVariants}
              className="text-4xl sm:text-5xl md:text-6xl lg:text-7xl font-bold text-gray-900 mb-6 sm:mb-8 leading-tight px-2"
            >
              {t('paint_by_numbers.title')}
            </motion.h1>
            
            <motion.p 
              variants={itemVariants}
              className="text-xl sm:text-2xl md:text-3xl text-gray-600 mb-8 sm:mb-12 max-w-4xl mx-auto px-4"
            >
              {t('paint_by_numbers.subtitle')}
            </motion.p>

            <motion.button
              variants={itemVariants}
              onClick={goToHome}
              className="inline-flex items-center space-x-2 bg-brand-primary text-white py-4 px-8 rounded-xl hover:bg-brand-primary/90 font-semibold text-lg sm:text-xl transition-all duration-200 focus:ring-2 focus:ring-brand-primary focus:ring-offset-2"
            >
              <ArrowRight className="w-5 h-5 sm:w-6 sm:h-6" />
              <span>{t('paint_by_numbers.back_to_home')}</span>
            </motion.button>
          </motion.div>
        </div>
      </section>

      {/* Features Section */}
      <section className="py-16 sm:py-20 lg:py-24 bg-white/50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <motion.div 
            variants={containerVariants}
            initial="hidden"
            animate="visible"
            className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8 sm:gap-10 lg:gap-12"
          >
            <motion.div 
              variants={itemVariants}
              className="bg-white/80 backdrop-blur-sm rounded-2xl shadow-xl p-8 border border-white/20 text-center"
            >
              <div className="flex items-center justify-center w-16 h-16 bg-brand-primary/10 rounded-full mx-auto mb-6">
                <Brush className="w-8 h-8 text-brand-primary" />
              </div>
              <h3 className="text-2xl font-bold text-gray-900 mb-4">
                {t('paint_by_numbers.features.creative.title')}
              </h3>
              <p className="text-gray-600 text-lg">
                {t('paint_by_numbers.features.creative.description')}
              </p>
            </motion.div>

            <motion.div 
              variants={itemVariants}
              className="bg-white/80 backdrop-blur-sm rounded-2xl shadow-xl p-8 border border-white/20 text-center"
            >
              <div className="flex items-center justify-center w-16 h-16 bg-brand-secondary/10 rounded-full mx-auto mb-6">
                <Star className="w-8 h-8 text-brand-secondary" />
              </div>
              <h3 className="text-2xl font-bold text-gray-900 mb-4">
                {t('paint_by_numbers.features.quality.title')}
              </h3>
              <p className="text-gray-600 text-lg">
                {t('paint_by_numbers.features.quality.description')}
              </p>
            </motion.div>

            <motion.div 
              variants={itemVariants}
              className="bg-white/80 backdrop-blur-sm rounded-2xl shadow-xl p-8 border border-white/20 text-center md:col-span-2 lg:col-span-1"
            >
              <div className="flex items-center justify-center w-16 h-16 bg-brand-accent/10 rounded-full mx-auto mb-6">
                <Clock className="w-8 h-8 text-brand-accent" />
              </div>
              <h3 className="text-2xl font-bold text-gray-900 mb-4">
                {t('paint_by_numbers.features.coming_soon.title')}
              </h3>
              <p className="text-gray-600 text-lg">
                {t('paint_by_numbers.features.coming_soon.description')}
              </p>
            </motion.div>
          </motion.div>
        </div>
      </section>

      {/* Coming Soon Section */}
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
                {t('paint_by_numbers.coming_soon.title')}
              </h2>
              <p className="text-xl sm:text-2xl text-white/90 mb-8 sm:mb-10 max-w-2xl mx-auto">
                {t('paint_by_numbers.coming_soon.description')}
              </p>
              <button
                onClick={goToHome}
                className="bg-white text-brand-primary py-4 px-8 rounded-xl hover:bg-gray-100 font-semibold text-lg sm:text-xl transition-all duration-200 focus:ring-2 focus:ring-white focus:ring-offset-2 focus:ring-offset-brand-primary"
              >
                {t('paint_by_numbers.coming_soon.button')}
              </button>
            </motion.div>
          </motion.div>
        </div>
      </section>
    </div>
  )
}

export default PaintByNumbersPage
