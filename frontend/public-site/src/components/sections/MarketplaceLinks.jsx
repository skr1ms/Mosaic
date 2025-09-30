import React from 'react';
import { useTranslation } from 'react-i18next';
import { motion } from 'framer-motion';
import { ShoppingBag, ExternalLink } from 'lucide-react';
import { usePartnerStore } from '../../store/partnerStore';

const MarketplaceLinks = () => {
  const { t } = useTranslation();
  const { partner } = usePartnerStore();

  if (!partner?.ozonLink && !partner?.wildberriesLink) {
    return null;
  }

  return (
    <section className="py-12 sm:py-16 bg-gray-50">
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6 }}
          viewport={{ once: true }}
          className="text-center mb-8 sm:mb-12"
        >
          <div className="flex items-center justify-center w-12 h-12 sm:w-16 sm:h-16 bg-brand-secondary/10 rounded-full mx-auto mb-4 sm:mb-6">
            <ShoppingBag className="w-6 h-6 sm:w-8 sm:h-8 text-brand-secondary" />
          </div>
          <h2 className="text-2xl sm:text-3xl md:text-4xl font-bold text-gray-900 mb-3 sm:mb-4 px-4 leading-tight">
            {t('marketplace.title')}
          </h2>
          <p className="text-base sm:text-lg md:text-xl text-gray-600 px-4 leading-relaxed">
            {t('marketplace_links.description')}
          </p>
        </motion.div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 sm:gap-6 max-w-2xl mx-auto">
          {partner?.ozonLink && (
            <motion.a
              href={partner.ozonLink}
              target="_blank"
              rel="noopener noreferrer"
              initial={{ opacity: 0, x: -20 }}
              whileInView={{ opacity: 1, x: 0 }}
              transition={{ duration: 0.6, delay: 0.1 }}
              viewport={{ once: true }}
              whileHover={{ y: -4, scale: 1.02 }}
              whileTap={{ scale: 0.98 }}
              className="group bg-white rounded-xl sm:rounded-2xl shadow-lg border border-gray-100 p-4 sm:p-6 lg:p-8 hover:shadow-xl transition-all duration-300 touch-target active:bg-gray-50"
            >
              <div className="text-center">
                <div className="w-12 h-12 sm:w-16 sm:h-16 bg-brand-primary/10 rounded-full flex items-center justify-center mx-auto mb-3 sm:mb-4">
                  <span className="text-xl sm:text-2xl lg:text-3xl">🛒</span>
                </div>
                <h3 className="text-lg sm:text-xl lg:text-2xl font-bold text-gray-900 mb-2 group-hover:text-brand-primary transition-colors leading-tight">
                  OZON
                </h3>
                <p className="text-sm sm:text-base text-gray-600 mb-3 sm:mb-4 leading-relaxed">
                  {t('marketplace_links.ozon_description')}
                </p>
                <div className="flex items-center justify-center space-x-2 text-brand-primary font-medium text-sm sm:text-base">
                  <span>{t('marketplace_links.go_to_shop')}</span>
                  <ExternalLink className="w-3 h-3 sm:w-4 sm:h-4 flex-shrink-0" />
                </div>
              </div>
            </motion.a>
          )}

          {partner?.wildberriesLink && (
            <motion.a
              href={partner.wildberriesLink}
              target="_blank"
              rel="noopener noreferrer"
              initial={{ opacity: 0, x: 20 }}
              whileInView={{ opacity: 1, x: 0 }}
              transition={{ duration: 0.6, delay: 0.2 }}
              viewport={{ once: true }}
              whileHover={{ y: -4, scale: 1.02 }}
              whileTap={{ scale: 0.98 }}
              className="group bg-white rounded-xl sm:rounded-2xl shadow-lg border border-gray-100 p-4 sm:p-6 lg:p-8 hover:shadow-xl transition-all duration-300 touch-target active:bg-gray-50"
            >
              <div className="text-center">
                <div className="w-12 h-12 sm:w-16 sm:h-16 bg-brand-accent/10 rounded-full flex items-center justify-center mx-auto mb-3 sm:mb-4">
                  <span className="text-xl sm:text-2xl lg:text-3xl">🛍️</span>
                </div>
                <h3 className="text-lg sm:text-xl lg:text-2xl font-bold text-gray-900 mb-2 group-hover:text-brand-accent transition-colors leading-tight">
                  Wildberries
                </h3>
                <p className="text-sm sm:text-base text-gray-600 mb-3 sm:mb-4 leading-relaxed">
                  {t('marketplace_links.wb_description')}
                </p>
                <div className="flex items-center justify-center space-x-2 text-brand-accent font-medium text-sm sm:text-base">
                  <span>{t('marketplace_links.go_to_shop')}</span>
                  <ExternalLink className="w-3 h-3 sm:w-4 sm:h-4 flex-shrink-0" />
                </div>
              </div>
            </motion.a>
          )}
        </div>

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6, delay: 0.3 }}
          viewport={{ once: true }}
          className="mt-8 sm:mt-12 text-center px-4"
        >
          <div className="bg-brand-primary/5 rounded-xl p-4 sm:p-6 max-w-2xl mx-auto">
            <h3 className="text-base sm:text-lg font-semibold text-gray-900 mb-2 leading-tight">
              💡 {t('marketplace_links.tip_title')}
            </h3>
            <p className="text-sm sm:text-base text-gray-600 leading-relaxed">
              {t('marketplace_links.tip_text')}
            </p>
          </div>
        </motion.div>
      </div>
    </section>
  );
};

export default MarketplaceLinks;
