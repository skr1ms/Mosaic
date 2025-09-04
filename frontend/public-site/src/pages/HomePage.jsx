import React from 'react'
import { useTranslation } from 'react-i18next'
import { motion } from 'framer-motion'
import { Gem, Palette, ShoppingCart, Ticket, ArrowRight, Image, Sparkles, Gift } from 'lucide-react'
import HeroSection from '../components/sections/HeroSection'
import SectionCard from '../components/cards/SectionCard'
import FAQ from '../components/sections/FAQ'
import MarketplaceLinks from '../components/sections/MarketplaceLinks'
import { usePartnerStore } from '../store/partnerStore'

const HomePage = () => {
  const { t } = useTranslation()
  const { partner } = usePartnerStore()

    const isOwnDomain = partner?.partner_code === '0000' || partner?.is_default

  const fadeInUp = {
    initial: { opacity: 0, y: 20 },
    animate: { opacity: 1, y: 0 },
    transition: { duration: 0.6 }
  }

  return (
    <div className="min-h-screen">
      <HeroSection />
      
      {/* Coupon Activation Section */}
      <section className="py-12 bg-gradient-to-r from-purple-600 to-pink-600">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <motion.div {...fadeInUp} className="text-center text-white">
            <Gift className="w-16 h-16 mx-auto mb-4" />
            <h2 className="text-3xl font-bold mb-4">У вас есть купон?</h2>
            <p className="text-xl mb-8 max-w-2xl mx-auto">
              Активируйте купон и создайте уникальную схему мозаики из вашей фотографии
            </p>
            <button
              onClick={() => window.location.href = '/coupon'}
              className="bg-white text-purple-600 px-8 py-4 rounded-xl font-semibold text-lg hover:bg-gray-100 transition-all duration-300 shadow-lg hover:shadow-xl inline-flex items-center"
            >
              <Ticket className="w-6 h-6 mr-3" />
              Активировать купон
              <ArrowRight className="w-5 h-5 ml-3" />
            </button>
          </motion.div>
        </div>
      </section>
      
      {isOwnDomain && (
        <section id="diamond-art" className="py-16 bg-gray-50">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <motion.div 
              {...fadeInUp}
              className="grid md:grid-cols-3 gap-8 max-w-6xl mx-auto items-stretch"
            >
              <SectionCard
                icon={<Gem className="w-8 h-8" />}
                title={t('sections.diamond_art.title')}
                description={t('sections.diamond_art.description')}
                buttonText={t('sections.diamond_art.button_details')}
                buttonIcon={<ArrowRight className="w-5 h-5" />}
                onClick={() => window.location.href = '/diamond-art'}
                className="hover-lift"
                gradient="from-brand-primary to-brand-secondary"
                active
              />
              
              <SectionCard
                icon={<Image className="w-8 h-8" />}
                title={t('diamond_art.preview_section.create_preview')}
                description={t('diamond_art.preview_section.description')}
                buttonText={t('diamond_art.preview_section.create_preview')}
                buttonIcon={<Sparkles className="w-5 h-5" />}
                onClick={() => window.location.href = '/preview'}
                className="hover-lift"
                gradient="from-purple-600 to-pink-600"
                active
              />
              
              <div id="paint-by-numbers">
                <SectionCard
                  icon={<Palette className="w-8 h-8" />}
                  title={t('sections.paint_by_numbers.title')}
                  description={t('sections.paint_by_numbers.description')}
                  comingSoon={t('sections.paint_by_numbers.coming_soon')}
                  gradient="from-gray-400 to-gray-500"
                  disabled
                />
              </div>
            </motion.div>
          </div>
        </section>
      )}

      <div id="faq">
        <FAQ />
      </div>
      <MarketplaceLinks />
    </div>
  )
}

export default HomePage
