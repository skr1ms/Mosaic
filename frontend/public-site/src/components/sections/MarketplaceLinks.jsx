import React from 'react'
import { useTranslation } from 'react-i18next'
import { motion } from 'framer-motion'
import { ShoppingBag, ExternalLink } from 'lucide-react'
import { usePartnerStore } from '../../store/partnerStore'

const MarketplaceLinks = () => {
  const { t } = useTranslation()
  const { partner } = usePartnerStore()

    if (!partner?.ozonLink && !partner?.wildberriesLink) {
    return null
  }

  return (
    <section className="py-16 bg-gray-50">
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6 }}
          viewport={{ once: true }}
          className="text-center mb-12"
        >
          <div className="flex items-center justify-center w-16 h-16 bg-brand-secondary/10 rounded-full mx-auto mb-6">
            <ShoppingBag className="w-8 h-8 text-brand-secondary" />
          </div>
          <h2 className="text-3xl md:text-4xl font-bold text-gray-900 mb-4">
            {t('marketplace.title')}
          </h2>
          <p className="text-xl text-gray-600">
            {t('marketplace_links.description')}
          </p>
        </motion.div>

        <div className="grid md:grid-cols-2 gap-6 max-w-2xl mx-auto">
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
              className="group bg-white rounded-2xl shadow-lg border border-gray-100 p-8 hover:shadow-xl transition-all duration-300"
            >
              <div className="text-center">
                <div className="w-16 h-16 bg-brand-primary/10 rounded-full flex items-center justify-center mx-auto mb-4">
                  <span className="text-3xl">🛒</span>
                </div>
                <h3 className="text-2xl font-bold text-gray-900 mb-2 group-hover:text-brand-primary transition-colors">
                  OZON
                </h3>
                <p className="text-gray-600 mb-4">
                  {t('marketplace_links.ozon_description')}
                </p>
                <div className="flex items-center justify-center space-x-2 text-brand-primary font-medium">
                  <span>{t('marketplace_links.go_to_shop')}</span>
                  <ExternalLink className="w-4 h-4" />
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
              className="group bg-white rounded-2xl shadow-lg border border-gray-100 p-8 hover:shadow-xl transition-all duration-300"
            >
              <div className="text-center">
                <div className="w-16 h-16 bg-brand-accent/10 rounded-full flex items-center justify-center mx-auto mb-4">
                  <span className="text-3xl">🛍️</span>
                </div>
                <h3 className="text-2xl font-bold text-gray-900 mb-2 group-hover:text-brand-accent transition-colors">
                  Wildberries
                </h3>
                <p className="text-gray-600 mb-4">
                  {t('marketplace_links.wb_description')}
                </p>
                <div className="flex items-center justify-center space-x-2 text-brand-accent font-medium">
                  <span>{t('marketplace_links.go_to_shop')}</span>
                  <ExternalLink className="w-4 h-4" />
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
          className="mt-12 text-center"
        >
          <div className="bg-brand-primary/5 rounded-xl p-6 max-w-2xl mx-auto">
            <h3 className="text-lg font-semibold text-gray-900 mb-2">
              💡 {t('marketplace_links.tip_title')}
            </h3>
            <p className="text-gray-600">
              {t('marketplace_links.tip_text')}
            </p>
          </div>
        </motion.div>
      </div>
    </section>
  )
}

export default MarketplaceLinks
