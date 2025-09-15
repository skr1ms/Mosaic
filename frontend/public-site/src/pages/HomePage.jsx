import React from 'react';
import { useTranslation } from 'react-i18next';
import { motion } from 'framer-motion';
import {
  Gem,
  Palette,
  ShoppingCart,
  ArrowRight,
  Image,
  Sparkles,
} from 'lucide-react';
import HeroSection from '../components/sections/HeroSection';
import SectionCard from '../components/cards/SectionCard';
import FAQ from '../components/sections/FAQ';
import MarketplaceLinks from '../components/sections/MarketplaceLinks';
import { usePartnerStore } from '../store/partnerStore';

const HomePage = () => {
  const { t } = useTranslation();
  const { partner } = usePartnerStore();

  const isOwnDomain = partner?.partner_code === '0000' || partner?.is_default;

  const fadeInUp = {
    initial: { opacity: 0, y: 20 },
    animate: { opacity: 1, y: 0 },
    transition: { duration: 0.6 },
  };

  return (
    <div className="min-h-screen">
      <HeroSection />

      <section id="diamond-art" className="py-12 sm:py-16 bg-gray-50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <motion.div
            {...fadeInUp}
            className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6 sm:gap-8 max-w-6xl mx-auto items-stretch"
          >
            <SectionCard
              icon={<Gem className="w-8 h-8" />}
              title={t('sections.diamond_art.title')}
              description={t('sections.diamond_art.description')}
              buttonText={t('sections.diamond_art.button_details')}
              buttonIcon={<ArrowRight className="w-5 h-5" />}
              onClick={() => (window.location.href = '/diamond-art')}
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
              onClick={() => (window.location.href = '/preview?new=true')}
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

      <div id="faq">
        <FAQ />
      </div>

      {isOwnDomain && <MarketplaceLinks />}
    </div>
  );
};

export default HomePage;
