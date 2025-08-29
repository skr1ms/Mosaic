import React from 'react'
import { useTranslation } from 'react-i18next'
import { motion } from 'framer-motion'
import { Ticket, Package, Palette, Code } from 'lucide-react'

const CouponInfo = ({ coupon }) => {
  const { t } = useTranslation()

  if (!coupon) {
    return null
  }

  // В development используем переменную из Docker Compose, в production - false
  const isDevMode = process.env.NODE_ENV === 'development'

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.6 }}
      className="mb-8"
    >
      <div className="bg-gradient-to-r from-brand-primary/5 to-brand-secondary/5 border border-brand-primary/20 rounded-2xl p-6">
        <div className="flex items-center space-x-3 mb-6">
          <div className="w-12 h-12 bg-brand-primary/10 rounded-xl flex items-center justify-center">
            <Ticket className="w-6 h-6 text-brand-primary" />
          </div>
          <div>
            <h2 className="text-xl font-bold text-gray-900">
              {t('editor.coupon_info')}
            </h2>

          </div>
        </div>

        <div className="grid md:grid-cols-3 gap-4">
          <div className="bg-white/60 rounded-lg p-4 border border-white/40">
            <div className="flex items-center space-x-2 mb-2">
              <Ticket className="w-5 h-5 text-gray-500" />
              <span className="text-sm font-medium text-gray-600">{t('coupon_info.coupon_number')}</span>
            </div>
            <p className="text-lg font-bold text-gray-900 font-mono tracking-wider">
              {coupon.code}
            </p>
          </div>

          <div className="bg-white/60 rounded-lg p-4 border border-white/40">
            <div className="flex items-center space-x-2 mb-2">
              <Package className="w-5 h-5 text-gray-500" />
              <span className="text-sm font-medium text-gray-600">{t('coupon_info.size')}</span>
            </div>
            <p className="text-lg font-bold text-gray-900">
              {coupon.size || t('coupon_info.not_specified')}
            </p>
          </div>

          <div className="bg-white/60 rounded-lg p-4 border border-white/40">
            <div className="flex items-center space-x-2 mb-2">
              <Palette className="w-5 h-5 text-gray-500" />
              <span className="text-sm font-medium text-gray-600">{t('coupon_info.style')}</span>
            </div>
            <p className="text-lg font-bold text-gray-900">
              {coupon.style || t('coupon_info.standard')}
            </p>
          </div>
        </div>

        {coupon.activated_at && (
          <div className="mt-4 p-3 bg-brand-secondary/10 border border-brand-secondary/20 rounded-lg">
            <p className="text-sm text-brand-secondary">
              ✅ {t('coupon_info.activated')}: {new Date(coupon.activated_at).toLocaleString()}
            </p>
          </div>
        )}
      </div>
    </motion.div>
  )
}

export default CouponInfo
