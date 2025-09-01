import React, { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { motion } from 'framer-motion'
import { ExternalLink, AlertCircle, Loader2, ShoppingCart } from 'lucide-react'
import { MosaicAPI } from '../api/client'
import { usePartnerStore, useUIStore } from '../store/partnerStore'

const MarketplaceCards = ({ selectedSize, selectedStyle }) => {
  const { t, i18n } = useTranslation()
  const { partner } = usePartnerStore()
  const { addNotification } = useUIStore()
  
  console.log('i18n debug:', {
    language: i18n.language,
    isInitialized: i18n.isInitialized,
    hasResources: !!i18n.store?.data?.[i18n.language]?.translation,
    testTranslation: t('mosaic_preview.marketplace.ready_to_buy')
  })
  
  const [marketplaceData, setMarketplaceData] = useState({})
  const [loading, setLoading] = useState(false)

  console.log('MarketplaceCards render:', { selectedSize, selectedStyle, partner: !!partner })

  const marketplaces = [
    {
      key: 'ozon',
      name: 'OZON',
      color: '#005BFF',
      backgroundColor: '#E3F2FD',
      icon: <ShoppingCart className="w-5 h-5" />
    },
    {
      key: 'wildberries',
      name: 'Wildberries',
      color: '#CB11AB',
      backgroundColor: '#FCE4EC',
      icon: <ShoppingCart className="w-5 h-5" />
    }
  ]

  useEffect(() => {
    if (!selectedSize || !selectedStyle || !partner) {
      console.log('MarketplaceCards: Missing required data, clearing marketplace data')
      setMarketplaceData({})
      return
    }

    const loadMarketplaceData = async () => {
      setLoading(true)
      const newData = {}

      try {
        // Load data for each marketplace
        for (const marketplace of marketplaces) {
          try {
            // If we have partner_id, try to get specific product URLs
            if (partner.partner_id) {
              const response = await MosaicAPI.generateProductURL(partner.partner_id, {
                marketplace: marketplace.key,
                style: selectedStyle,
                size: selectedSize
              })
              
              newData[marketplace.key] = {
                ...response,
                available: response.has_article, // Available if we have specific article
                has_general_link: !!response.url && !response.has_article, // General link available if no specific article but URL exists
                specific_product: true
              }
            } else {
              // For default partner or when no partner_id, use general marketplace links
              const generalLink = partner?.[`${marketplace.key}Link`] || partner?.marketplace_links?.[marketplace.key]
              
              newData[marketplace.key] = {
                url: generalLink || '',
                sku: '',
                has_article: false,
                available: !!generalLink,
                partner_name: partner.name || '',
                marketplace: marketplace.key,
                size: selectedSize,
                style: selectedStyle
              }
            }
          } catch (error) {
            console.error(`Error loading ${marketplace.key} data:`, error)
            // Fallback to general link even on API error
            const generalLink = partner?.[`${marketplace.key}Link`] || partner?.marketplace_links?.[marketplace.key]
            
            newData[marketplace.key] = {
              url: generalLink || '',
              sku: '',
              has_article: false,
              available: false, // Never show as "available" when API fails - specific product not found
              has_general_link: !!generalLink,
              error: !generalLink,
              partner_name: partner.name || '',
              marketplace: marketplace.key,
              size: selectedSize,
              style: selectedStyle
            }
          }
        }
      } catch (error) {
        console.error('Error loading marketplace data:', error)
      }

      setMarketplaceData(newData)
      setLoading(false)
    }

    loadMarketplaceData()
  }, [selectedSize, selectedStyle, partner])

  const handleProductClick = async (marketplaceKey, data) => {
    if (!data.url) {
      addNotification({
        type: 'error',
        message: t('sections.diamond_art.marketplace.no_link_available')
      })
      return
    }

    try {
      // Copy URL to clipboard
      await navigator.clipboard.writeText(data.url)
      
      // Show appropriate notification
      if (data.has_article && data.sku) {
        addNotification({
          type: 'success',
          message: t('sections.diamond_art.marketplace.url_copied', { sku: data.sku })
        })
      } else {
        // For general marketplace links when specific product not available
        if (data.has_general_link || data.url) {
          addNotification({
            type: 'info',
            message: t('sections.diamond_art.marketplace.redirected_to_general')
          })
        } else {
          addNotification({
            type: 'error',
            message: t('sections.diamond_art.marketplace.no_link_available')
          })
          return
        }
      }

      // Open in new tab
      window.open(data.url, '_blank')
    } catch (error) {
      console.error('Error handling product click:', error)
      // Still try to open the URL even if clipboard fails
      window.open(data.url, '_blank')
    }
  }

  const hasPartner = !!partner
  const marketplaceDataKeys = Object.keys(marketplaceData)

  console.log('MarketplaceCards: rendering with data:', {
    selectedSize,
    selectedStyle,
    hasPartner,
    marketplaceDataKeys,
    loading
  })

  if (!hasPartner) {
    return null
  }

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.5 }}
      className="mt-8 p-6 bg-white rounded-lg shadow-lg"
    >
      <div className="text-center mb-6">
        <h3 className="text-xl font-semibold text-gray-900 mb-2">
{t('sections.diamond_art.marketplace.ready_to_buy')}
        </h3>
        {(() => {
          // Используем размеры без перевода, как на странице активации купонов
          const sizeText = selectedSize // Просто 21x30, 30x40 и т.д.
          // Приводим к единому формату ключей в shop.styles
          const styleKey = selectedStyle === 'skin_tones' ? 'skin_tone' : selectedStyle
          const styleText = t(`shop.styles.${styleKey}.title`) || selectedStyle
          
          console.log('Translation debug:', {
            selectedSize,
            selectedStyle,
            sizeText,
            styleText
          })
          
          return t('sections.diamond_art.marketplace.selected_params', { size: sizeText, style: styleText })
        })()}
      </div>

      {loading ? (
        <div className="flex justify-center items-center py-8">
          <Loader2 className="w-6 h-6 animate-spin text-blue-500 mr-2" />
          <span className="text-gray-600">{t('sections.diamond_art.marketplace.checking_availability')}</span>
        </div>
      ) : (
        <div className="grid md:grid-cols-2 gap-4">
          {marketplaces.map((marketplace) => {
            const data = marketplaceData[marketplace.key] || {}
            const isAvailable = data.available
            const hasError = data.error
            const hasAlternatives = data.has_general_link && !data.available

            console.log(`Rendering ${marketplace.key} card:`, { data, isAvailable, hasError })

            return (
              <motion.div
                key={marketplace.key}
                whileHover={{ scale: 1.02 }}
                whileTap={{ scale: 0.98 }}
                className={`p-4 rounded-lg border-2 transition-all cursor-pointer ${
                  isAvailable
                    ? 'border-green-200 bg-green-50 hover:border-green-300'
                    : hasAlternatives
                    ? 'border-yellow-200 bg-yellow-50 hover:border-yellow-300'
                    : 'border-gray-200 bg-gray-50 hover:border-gray-300'
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
                  <div>
                    <h4 className="font-semibold text-gray-900">{marketplace.name}</h4>
                    {data.sku && (
                      <p className="text-xs text-gray-500">Артикул: {data.sku}</p>
                    )}
                  </div>
                </div>

                <div className="flex items-center justify-between">
                  <div className="flex items-center">
                    {hasError ? (
                      <>
                        <AlertCircle className="w-4 h-4 text-red-500 mr-2" />
                        <span className="text-sm">{t('sections.diamond_art.marketplace.error_loading')}</span>
                      </>
                    ) : isAvailable ? (
                      <>
                        <div className="w-3 h-3 bg-green-500 rounded-full mr-2"></div>
                        <span className="text-sm font-medium">{t('sections.diamond_art.marketplace.in_stock')}</span>
                      </>
                    ) : hasAlternatives ? (
                      <>
                        <div className="w-3 h-3 bg-yellow-500 rounded-full mr-2"></div>
                        <span className="text-sm">{t('sections.diamond_art.marketplace.not_available')}</span>
                      </>
                    ) : (
                      <>
                        <div className="w-3 h-3 bg-gray-400 rounded-full mr-2"></div>
                        <span className="text-sm">{t('sections.diamond_art.marketplace.error_loading')}</span>
                      </>
                    )}
                  </div>
                  
                  <button className={`px-3 py-1 rounded-full text-sm font-medium flex items-center ${
                    isAvailable
                      ? 'bg-green-600 text-white hover:bg-green-700'
                      : hasAlternatives
                      ? 'bg-yellow-600 text-white hover:bg-yellow-700'
                      : 'bg-gray-600 text-white hover:bg-gray-700'
                  }`}>
                    <ExternalLink className="w-3 h-3 mr-1" />
                    {isAvailable ? t('sections.diamond_art.marketplace.buy_now') : t('sections.diamond_art.marketplace.view_alternatives')}
                  </button>
                </div>
              </motion.div>
            )
          })}
        </div>
      )}

      <div className="mt-4 text-center">
        <p className="text-xs text-gray-500">
{t('sections.diamond_art.marketplace.info_text')}
        </p>
      </div>
    </motion.div>
  )
}

export default MarketplaceCards