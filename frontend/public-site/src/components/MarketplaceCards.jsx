import React, { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { motion } from 'framer-motion';
import { ExternalLink, AlertCircle, Loader2, ShoppingCart } from 'lucide-react';
import { MosaicAPI } from '../api/client';
import { usePartnerStore, useUIStore } from '../store/partnerStore';

const MarketplaceCards = ({ selectedSize, selectedStyle }) => {
  const { t, i18n } = useTranslation();
  const { partner } = usePartnerStore();
  const { addNotification } = useUIStore();

  console.log('i18n debug:', {
    language: i18n.language,
    isInitialized: i18n.isInitialized,
    hasResources: !!i18n.store?.data?.[i18n.language]?.translation,
    testTranslation: t('mosaic_preview.marketplace.ready_to_buy'),
  });

  const [marketplaceData, setMarketplaceData] = useState({});
  const [loading, setLoading] = useState(false);

  console.log('MarketplaceCards render:', {
    selectedSize,
    selectedStyle,
    partner: !!partner,
  });

  const marketplaces = [
    {
      key: 'ozon',
      name: 'OZON',
      color: '#005BFF',
      backgroundColor: '#E3F2FD',
      icon: <ShoppingCart className="w-5 h-5" />,
    },
    {
      key: 'wildberries',
      name: 'Wildberries',
      color: '#CB11AB',
      backgroundColor: '#FCE4EC',
      icon: <ShoppingCart className="w-5 h-5" />,
    },
  ];

  useEffect(() => {
    if (!selectedSize || !selectedStyle || !partner) {
      console.log(
        'MarketplaceCards: Missing required data, clearing marketplace data'
      );
      setMarketplaceData({});
      return;
    }

    const loadMarketplaceData = async () => {
      setLoading(true);
      const newData = {};

      try {
        for (const marketplace of marketplaces) {
          try {
            if (partner.partner_id) {
              const response = await MosaicAPI.generateProductURL(
                partner.partner_id,
                {
                  marketplace: marketplace.key,
                  style: selectedStyle,
                  size: selectedSize,
                }
              );

              let actualAvailability = false;
              try {
                const statusResponse = await MosaicAPI.checkMarketplaceStatus({
                  marketplace: marketplace.key,
                  partnerId: partner.partner_id,
                  size: selectedSize,
                  style: selectedStyle,
                  sku: response.sku
                });
                actualAvailability = statusResponse.available;
              } catch (statusError) {
                console.warn(`Failed to check availability for ${marketplace.key}:`, statusError);
                actualAvailability = response.has_article;
              }

              newData[marketplace.key] = {
                ...response,
                available: actualAvailability,
                has_general_link: !!response.url && !actualAvailability,
                specific_product: true,
              };
            } else {
              const generalLink =
                partner?.[`${marketplace.key}Link`] ||
                partner?.marketplace_links?.[marketplace.key];

              newData[marketplace.key] = {
                url: generalLink || '',
                sku: '',
                has_article: false,
                available: false,
                has_general_link: !!generalLink,
                partner_name: partner.name || '',
                marketplace: marketplace.key,
                size: selectedSize,
                style: selectedStyle,
              };
            }
          } catch (error) {
            console.error(`Error loading ${marketplace.key} data:`, error);
            const generalLink =
              partner?.[`${marketplace.key}Link`] ||
              partner?.marketplace_links?.[marketplace.key];

            newData[marketplace.key] = {
              url: generalLink || '',
              sku: '',
              has_article: false,
              available: false,
              has_general_link: !!generalLink,
              error: !generalLink,
              partner_name: partner.name || '',
              marketplace: marketplace.key,
              size: selectedSize,
              style: selectedStyle,
            };
          }
        }
      } catch (error) {
        console.error('Error loading marketplace data:', error);
      }

      setMarketplaceData(newData);
      setLoading(false);
    };

    loadMarketplaceData();
  }, [selectedSize, selectedStyle, partner]);

  const handleProductClick = async (marketplaceKey, data) => {
    if (!data.url) {
      addNotification({
        type: 'error',
        message: t('sections.diamond_art.marketplace.no_link_available'),
      });
      return;
    }

    try {
      await navigator.clipboard.writeText(data.url);

      if (data.has_article && data.sku) {
        addNotification({
          type: 'success',
          message: t('sections.diamond_art.marketplace.url_copied', {
            sku: data.sku,
          }),
        });
      } else {
        if (data.has_general_link || data.url) {
          addNotification({
            type: 'info',
            message: t(
              'sections.diamond_art.marketplace.redirected_to_general'
            ),
          });
        } else {
          addNotification({
            type: 'error',
            message: t('sections.diamond_art.marketplace.no_link_available'),
          });
          return;
        }
      }

      window.open(data.url, '_blank');
    } catch (error) {
      console.error('Error handling product click:', error);
      window.open(data.url, '_blank');
    }
  };

  const hasPartner = !!partner;
  const marketplaceDataKeys = Object.keys(marketplaceData);

  console.log('MarketplaceCards: rendering with data:', {
    selectedSize,
    selectedStyle,
    hasPartner,
    marketplaceDataKeys,
    loading,
  });

  if (!hasPartner) {
    return null;
  }

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.5 }}
      className="mt-6 sm:mt-8 p-4 sm:p-6 bg-white rounded-lg sm:rounded-xl shadow-lg"
    >
      <div className="text-center mb-4 sm:mb-6">
        <h3 className="text-lg sm:text-xl font-semibold text-gray-900 mb-2 leading-tight">
          {t('sections.diamond_art.marketplace.ready_to_buy')}
        </h3>
        {`Выбранные параметры: ${selectedSize ? `${selectedSize} ${t('common.cm')}` : 'не выбран'}, ${selectedStyle ? t(`diamond_mosaic_styles.styles.${selectedStyle}.title`) || selectedStyle : 'не выбран'}`}
      </div>

      {loading ? (
        <div className="flex justify-center items-center py-8">
          <Loader2 className="w-6 h-6 animate-spin text-blue-500 mr-2" />
          <span className="text-gray-600">
            {t('sections.diamond_art.marketplace.checking_availability')}
          </span>
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 sm:gap-4">
          {marketplaces.map(marketplace => {
            const data = marketplaceData[marketplace.key] || {};
            const isAvailable = data.available;
            const hasError = data.error;
            const hasAlternatives = data.has_general_link && !data.available;

            console.log(`Rendering ${marketplace.key} card:`, {
              data,
              isAvailable,
              hasError,
            });

            return (
              <motion.div
                key={marketplace.key}
                whileHover={{ scale: 1.01 }}
                whileTap={{ scale: 0.99 }}
                className={`p-3 sm:p-4 rounded-lg sm:rounded-xl border-2 transition-all cursor-pointer touch-target ${
                  isAvailable
                    ? 'border-green-200 bg-green-50 hover:border-green-300 active:bg-green-100'
                    : hasAlternatives
                      ? 'border-yellow-200 bg-yellow-50 hover:border-yellow-300 active:bg-yellow-100'
                      : 'border-gray-200 bg-gray-50 hover:border-gray-300 active:bg-gray-100'
                }`}
                onClick={() => handleProductClick(marketplace.key, data)}
              >
                <div className="flex items-center mb-3">
                  <div
                    className="w-10 h-10 rounded-full flex items-center justify-center mr-3"
                    style={{ backgroundColor: marketplace.color }}
                  >
                    {marketplace.icon}
                  </div>
                  <div className="min-w-0 flex-1">
                    <h4 className="text-sm sm:text-base font-semibold text-gray-900 leading-tight">
                      {marketplace.name}
                    </h4>
                    {data.sku && (
                      <p className="text-xs text-gray-500 leading-relaxed break-all">
                        {t('sections.diamond_art.marketplace.sku', {
                          sku: data.sku,
                        })}
                      </p>
                    )}
                  </div>
                </div>

                <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-2 sm:gap-0">
                  <div className="flex items-center">
                    {hasError ? (
                      <>
                        <AlertCircle className="w-3 h-3 sm:w-4 sm:h-4 text-red-500 mr-2 flex-shrink-0" />
                        <span className="text-xs sm:text-sm leading-relaxed">
                          {t('sections.diamond_art.marketplace.error_loading')}
                        </span>
                      </>
                    ) : isAvailable ? (
                      <>
                        <div className="w-2 h-2 sm:w-3 sm:h-3 bg-green-500 rounded-full mr-2 flex-shrink-0"></div>
                        <span className="text-xs sm:text-sm font-medium leading-relaxed">
                          {t('sections.diamond_art.marketplace.in_stock')}
                        </span>
                      </>
                    ) : hasAlternatives ? (
                      <>
                        <div className="w-2 h-2 sm:w-3 sm:h-3 bg-yellow-500 rounded-full mr-2 flex-shrink-0"></div>
                        <span className="text-xs sm:text-sm leading-relaxed">
                          {t('sections.diamond_art.marketplace.not_available')}
                        </span>
                      </>
                    ) : (
                      <>
                        <div className="w-2 h-2 sm:w-3 sm:h-3 bg-gray-400 rounded-full mr-2 flex-shrink-0"></div>
                        <span className="text-xs sm:text-sm leading-relaxed">
                          {t('sections.diamond_art.marketplace.error_loading')}
                        </span>
                      </>
                    )}
                  </div>

                  <button
                    className={`px-3 py-2 sm:py-1 rounded-full text-xs sm:text-sm font-medium flex items-center touch-target w-full sm:w-auto justify-center ${
                      isAvailable
                        ? 'bg-green-600 text-white hover:bg-green-700 active:bg-green-800'
                        : hasAlternatives
                          ? 'bg-yellow-600 text-white hover:bg-yellow-700 active:bg-yellow-800'
                          : 'bg-gray-600 text-white hover:bg-gray-700 active:bg-gray-800'
                    }`}
                  >
                    <ExternalLink className="w-3 h-3 mr-1 flex-shrink-0" />
                    <span className="leading-tight">
                      {isAvailable
                        ? t('sections.diamond_art.marketplace.buy_now')
                        : t(
                            'sections.diamond_art.marketplace.view_alternatives'
                          )}
                    </span>
                  </button>
                </div>
              </motion.div>
            );
          })}
        </div>
      )}

      <div className="mt-3 sm:mt-4 text-center">
        <p className="text-xs text-gray-500 leading-relaxed px-2 sm:px-0">
          {t('sections.diamond_art.marketplace.info_text')}
        </p>
      </div>
    </motion.div>
  );
};

export default MarketplaceCards;
