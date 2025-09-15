import React from 'react';
import { useTranslation } from 'react-i18next';
import { useSearchParams } from 'react-router-dom';
import { motion } from 'framer-motion';
import EditorSteps from '../components/editor/EditorSteps';
import CouponInfo from '../components/editor/CouponInfo';
import { useCouponStore } from '../store/partnerStore';
import { MosaicAPI } from '../api/client';
import { useUIStore } from '../store/partnerStore';

const EditorPage = () => {
  const { t } = useTranslation();
  const [searchParams, setSearchParams] = useSearchParams();
  const { coupon, setCoupon } = useCouponStore();
  const { addNotification } = useUIStore();
  const lastCheckedCodeRef = React.useRef(null);

  React.useEffect(() => {
    try {
      localStorage.removeItem('pendingOrder');

      const keys = Object.keys(localStorage);
      keys.forEach(key => {
        if (
          key.startsWith('preview_') ||
          key.startsWith('temp_') ||
          key.startsWith('shop_')
        ) {
          localStorage.removeItem(key);
        }
      });

      console.log('Cleared stale localStorage data in editor');
    } catch (error) {
      console.error('Error clearing localStorage in editor:', error);
    }

    const couponCode = searchParams.get('coupon');
    const size = searchParams.get('size');
    const style = searchParams.get('style');

    const image = searchParams.get('image');
    const step = parseInt(searchParams.get('step') || '1', 10);

    const cleanedUrlCode = (couponCode || '').replace(/\D/g, '');
    const cleanedStoreCode = (coupon?.code || '').replace(/\D/g, '');

    if (cleanedUrlCode && cleanedUrlCode !== cleanedStoreCode) {
      setCoupon({
        code: couponCode,
        size: size || 'unknown',
        style: style || 'unknown',
      });
      try {
        sessionStorage.setItem('editor:coupon', couponCode);
      } catch {}

      const params = new URLSearchParams(searchParams);
      params.delete('image');
      params.delete('step');
      setSearchParams(params);
      try {
        sessionStorage.removeItem('editor:lastImageId');
        sessionStorage.removeItem('editor:lastPreview');
        Object.keys(sessionStorage).forEach(k => {
          if (
            k.startsWith('editor:confirmed:') ||
            k.startsWith('editor:selectedOptions:') ||
            k.startsWith('editor:step:') ||
            k.startsWith('editor:lastPreview:') ||
            k.startsWith('editor:schemaData:') ||
            k.startsWith('editor:edits:')
          ) {
            sessionStorage.removeItem(k);
          }
        });
      } catch {}
    }

    if (coupon?.code) {
      try {
        sessionStorage.setItem('editor:coupon', coupon.code);
      } catch {}
    }
  }, [searchParams, coupon, setCoupon, setSearchParams]);

  React.useEffect(() => {
    const validateCoupon = async () => {
      if (!coupon?.code) return;
      const imageId = searchParams.get('image');
      if (imageId) return;
      if (lastCheckedCodeRef.current === coupon.code) return;

      const urlSize = searchParams.get('size');
      const urlStyle = searchParams.get('style');
      if (urlSize && urlStyle) {
        lastCheckedCodeRef.current = coupon.code;
        return;
      }

      lastCheckedCodeRef.current = coupon.code;

      try {
        const cleanCode = (coupon.code || '').replace(/\D/g, '');
        if (cleanCode.length !== 12) return;

        const info = await MosaicAPI.validateCoupon(cleanCode);

        if (!info.valid) {
          addNotification({
            type: 'error',
            title: t('notifications.activation_error'),
            message: t('notifications.invalid_coupon'),
          });
          return;
        } else if (info.status === 'used') {
          setCoupon({
            code: coupon.code,
            size: info.size || coupon.size || 'unknown',
            style: info.style || coupon.style || 'unknown',
          });
          return;
        }

        try {
          const activationResult = await MosaicAPI.activateCoupon(cleanCode);

          setCoupon({
            code: coupon.code,
            size:
              activationResult.size || info.size || coupon.size || 'unknown',
            style:
              activationResult.style || info.style || coupon.style || 'unknown',
          });

          if (
            activationResult.message ===
            t('notifications.coupon_already_activated')
          ) {
          } else {
          }
        } catch (activationError) {
          console.error('EditorPage: Activation failed:', activationError);

          setCoupon({
            code: coupon.code,
            size: info.size || coupon.size || 'unknown',
            style: info.style || coupon.style || 'unknown',
          });

          addNotification({
            type: 'error',
            title: t('notifications.activation_error'),
            message:
              activationError.message || t('notifications.activation_error'),
          });
        }
      } catch (e) {
        console.error('EditorPage: Validation error:', e);
        const status = e?.status;
        if (status === 404) {
          addNotification({
            type: 'error',
            title: t('notifications.activation_error'),
            message: t('notifications.invalid_coupon'),
          });
        }
      }
    };
    validateCoupon();
  }, [coupon?.code, setCoupon, addNotification, t, searchParams]);

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6 sm:py-8">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6 }}
        >
          <h1 className="text-2xl sm:text-3xl font-bold text-gray-900 mb-6 sm:mb-8 leading-tight">
            {t('editor.title')}
          </h1>

          {coupon && <CouponInfo coupon={coupon} />}

          <EditorSteps
            couponCode={coupon?.code}
            couponSize={coupon?.size}
            initialImageId={searchParams.get('image') || null}
            initialStep={parseInt(searchParams.get('step') || '1', 10)}
          />
        </motion.div>
      </div>
    </div>
  );
};

export default EditorPage;
