import React from 'react';
import { useTranslation } from 'react-i18next';
import { Mail, Phone, MapPin } from 'lucide-react';
import { usePartnerStore } from '../../store/partnerStore';

const Footer = () => {
  const { t } = useTranslation();
  const { partner } = usePartnerStore();

  const currentYear = new Date().getFullYear();

  return (
    <footer className="bg-gray-900 text-white">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 sm:py-12">
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6 sm:gap-8">
          {}
          <div className="text-center sm:text-left">
            <div className="flex flex-col items-center sm:items-start mb-4">
              <div className="flex items-center justify-center sm:justify-start mb-2">
                {partner?.logoUrl ? (
                  <img
                    src={partner.logoUrl}
                    alt={partner.name || t('navigation.company_name')}
                    className="h-8 sm:h-10 w-auto max-w-[150px]"
                    onError={e => {
                      e.target.style.display = 'none';
                      const fallback = e.target.nextElementSibling;
                      if (fallback) fallback.style.display = 'block';
                    }}
                    onLoad={e => {
                      const fallback = e.target.nextElementSibling;
                      if (fallback) fallback.style.display = 'none';
                    }}
                    style={{
                      maxHeight: '40px',
                      objectFit: 'contain',
                      display: 'block',
                      backgroundColor: 'transparent',
                    }}
                  />
                ) : null}

                <img
                  src="/logo.svg"
                  alt={t('navigation.company_name')}
                  className="h-8 sm:h-10 w-auto max-w-[150px]"
                  style={{
                    maxHeight: '40px',
                    objectFit: 'contain',
                    display: partner?.logoUrl ? 'none' : 'block',
                    backgroundColor: 'transparent',
                  }}
                />
              </div>
              {partner?.name && (
                <h2 className="text-lg sm:text-xl font-semibold text-white">
                  {partner.name}
                </h2>
              )}
            </div>
            <p className="text-gray-300 mb-4 text-sm sm:text-base leading-relaxed">
              {t('footer.company_description')}
            </p>
          </div>

          <div className="text-center sm:text-left">
            <h3 className="text-lg font-semibold mb-4">
              {t('footer.contacts')}
            </h3>
            <div className="space-y-3">
              {partner?.email && (
                <div className="flex items-center justify-center sm:justify-start space-x-3">
                  <Mail className="w-4 h-4 sm:w-5 sm:h-5 text-gray-400 flex-shrink-0" />
                  <a
                    href={`mailto:${partner.email}`}
                    className="text-gray-300 hover:text-white active:text-gray-100 transition-colors text-sm sm:text-base break-all touch-target"
                  >
                    {partner.email}
                  </a>
                </div>
              )}

              {partner?.phone && (
                <div className="flex items-center justify-center sm:justify-start space-x-3">
                  <Phone className="w-4 h-4 sm:w-5 sm:h-5 text-gray-400 flex-shrink-0" />
                  <a
                    href={`tel:${partner.phone}`}
                    className="text-gray-300 hover:text-white active:text-gray-100 transition-colors text-sm sm:text-base touch-target"
                  >
                    {partner.phone}
                  </a>
                </div>
              )}

              {partner?.address && (
                <div className="flex items-center justify-center sm:justify-start space-x-3">
                  <MapPin className="w-4 h-4 sm:w-5 sm:h-5 text-gray-400 flex-shrink-0" />
                  <span className="text-gray-300 text-sm sm:text-base">
                    {partner.address}
                  </span>
                </div>
              )}
            </div>
          </div>

          <div className="text-center sm:text-left lg:col-span-1 sm:col-span-2 lg:col-span-1">
            <h3 className="text-lg font-semibold mb-4">
              {t('footer.social_networks')}
            </h3>
            <div className="space-y-3">
              {partner?.telegram && (
                <a
                  href={
                    partner.telegram.startsWith('http')
                      ? partner.telegram
                      : `https://t.me/${partner.telegram.replace('@', '')}`
                  }
                  target="_blank"
                  rel="noopener noreferrer"
                  className="flex items-center justify-center sm:justify-start space-x-3 text-gray-300 hover:text-white active:text-gray-100 transition-colors text-sm sm:text-base touch-target"
                >
                  <span className="text-lg">📱</span>
                  <span>Telegram</span>
                </a>
              )}

              {partner?.whatsapp && (
                <a
                  href={
                    partner.whatsapp.startsWith('http')
                      ? partner.whatsapp
                      : `https://wa.me/${partner.whatsapp}`
                  }
                  target="_blank"
                  rel="noopener noreferrer"
                  className="flex items-center justify-center sm:justify-start space-x-3 text-gray-300 hover:text-white active:text-gray-100 transition-colors text-sm sm:text-base touch-target"
                >
                  <span className="text-lg">💬</span>
                  <span>{t('footer.whatsapp')}</span>
                </a>
              )}
            </div>
          </div>
        </div>

        <div className="border-t border-gray-800 mt-6 sm:mt-8 pt-6 sm:pt-8 text-center">
          <p className="text-gray-400 text-sm sm:text-base">
            {t('footer.copyright')}
          </p>
        </div>
      </div>
    </footer>
  );
};

export default Footer;
