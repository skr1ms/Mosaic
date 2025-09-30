import React from 'react';
import { useTranslation } from 'react-i18next';
import { motion } from 'framer-motion';
import { Ticket, Package, Palette, Code } from 'lucide-react';

const CouponInfo = ({ coupon }) => {
  const { t } = useTranslation();

  if (!coupon) {
    return null;
  }

  const isDevMode = process.env.NODE_ENV === 'development';

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.6 }}
      className="mb-6 sm:mb-8"
    >
      <div className="bg-gradient-to-r from-brand-primary/5 to-brand-secondary/5 border border-brand-primary/20 rounded-xl sm:rounded-2xl p-4 sm:p-6">
        <div className="flex items-center space-x-2 sm:space-x-3 mb-4 sm:mb-6">
          <div className="w-10 h-10 sm:w-12 sm:h-12 bg-brand-primary/10 rounded-lg sm:rounded-xl flex items-center justify-center flex-shrink-0">
            <Ticket className="w-5 h-5 sm:w-6 sm:h-6 text-brand-primary" />
          </div>
          <div className="min-w-0 flex-1">
            <h2 className="text-lg sm:text-xl font-bold text-gray-900 leading-tight">
              {t('editor.coupon_info')}
            </h2>
          </div>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3 sm:gap-4">
          <div className="bg-white/60 rounded-lg p-3 sm:p-4 border border-white/40">
            <div className="flex items-center space-x-2 mb-2">
              <Ticket className="w-4 h-4 sm:w-5 sm:h-5 text-gray-500 flex-shrink-0" />
              <span className="text-xs sm:text-sm font-medium text-gray-600 leading-tight">
                {t('coupon_info.coupon_number')}
              </span>
            </div>
            <p className="text-base sm:text-lg font-bold text-gray-900 font-mono tracking-wider leading-tight break-all">
              {coupon.code}
            </p>
          </div>

          <div className="bg-white/60 rounded-lg p-3 sm:p-4 border border-white/40">
            <div className="flex items-center space-x-2 mb-2">
              <Package className="w-4 h-4 sm:w-5 sm:h-5 text-gray-500 flex-shrink-0" />
              <span className="text-xs sm:text-sm font-medium text-gray-600 leading-tight">
                {t('coupon_info.size')}
              </span>
            </div>
            <p className="text-base sm:text-lg font-bold text-gray-900 leading-tight">
              {coupon.size || t('coupon_info.not_specified')}
            </p>
          </div>

          <div className="bg-white/60 rounded-lg p-3 sm:p-4 border border-white/40">
            <div className="flex items-center space-x-2 mb-2">
              <Palette className="w-4 h-4 sm:w-5 sm:h-5 text-gray-500 flex-shrink-0" />
              <span className="text-xs sm:text-sm font-medium text-gray-600 leading-tight">
                {t('coupon_info.style')}
              </span>
            </div>
            <p className="text-base sm:text-lg font-bold text-gray-900 leading-tight">
              {coupon.style || t('coupon_info.standard')}
            </p>
          </div>
        </div>

        {coupon.activated_at && (
          <div className="mt-3 sm:mt-4 p-3 bg-brand-secondary/10 border border-brand-secondary/20 rounded-lg">
            <p className="text-xs sm:text-sm text-brand-secondary leading-relaxed">
              ✅ {t('coupon_info.activated')}:{' '}
              {new Date(coupon.activated_at).toLocaleString()}
            </p>
          </div>
        )}
      </div>
    </motion.div>
  );
};

export default CouponInfo;
