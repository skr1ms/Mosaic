import React from 'react'
import { Routes, Route } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import Layout from './components/Layout/Layout'
import HomePage from './pages/HomePage'
import EditorPage from './pages/EditorPage'
import DiamondArtPage from './pages/DiamondArtPage'
import ShopPage from './pages/ShopPage'
import NotFoundPage from './pages/NotFoundPage'
import PaintByNumbersPage from './pages/PaintByNumbersPage'
import WhatIsThisPage from './pages/WhatIsThisPage'
import MosaicPreviewPage from './pages/MosaicPreviewPage'
import DiamondMosaicPage from './pages/DiamondMosaicPage'
import DiamondMosaicPreviewPage from './pages/DiamondMosaicPreviewPage'
import DiamondMosaicPreviewAlbumPage from './pages/DiamondMosaicPreviewAlbumPage'
import DiamondMosaicEditorPage from './pages/DiamondMosaicEditorPage'
import DiamondMosaicPurchasePage from './pages/DiamondMosaicPurchasePage'
import DiamondMosaicSuccessPage from './pages/DiamondMosaicSuccessPage'
import { usePartnerStore } from './store/partnerStore'
import { useUIStore } from './store/partnerStore'
import { useBrandingQuery } from './hooks/useBrandingQuery'
import { useDocumentTitle } from './hooks/useDocumentTitle'
import LoadingScreen from './components/ui/LoadingScreen'
import NotificationSystem from './components/ui/NotificationSystem'
import SupportChatWidget from './components/SupportChatWidget'
import BrandingProvider from './components/ui/BrandingProvider'
import './assets/branding.css'

function App() {
  const { i18n } = useTranslation()
  const { partner, setPartner } = usePartnerStore()
  const { notifications, removeNotification } = useUIStore()
  
  const { data: brandingData, isLoading } = useBrandingQuery()
  
  // Update document title and meta tags based on current language
  useDocumentTitle()

  React.useEffect(() => {
    if (!brandingData) return

    console.log('App - Raw branding data:', brandingData)
    console.log('App - Brand name from API:', brandingData.brand_name)
    console.log('App - Is default branding:', brandingData.is_default)
    console.log('App - Partner code:', brandingData.partner_code)

    // Маппим данные API в формат, ожидаемый компонентами
    const mappedPartner = {
      name: brandingData.brand_name,
      email: brandingData.contact_email,
      phone: brandingData.contact_phone,
      telegram: brandingData.contact_telegram,
      whatsapp: brandingData.contact_whatsapp,
      ozonLink: brandingData.marketplace_links?.ozon,
      wildberriesLink: brandingData.marketplace_links?.wildberries,
      marketplace_links: brandingData.marketplace_links,
      address: brandingData.contact_address,
      logoUrl: brandingData.logo_url,
      brandColors: brandingData.brand_colors || [],
      partner_code: brandingData.partner_code,
      partner_id: brandingData.partner_id,
      is_default: brandingData.is_default
    }

    console.log('App - Mapped partner data:', mappedPartner)
    console.log('App - Marketplace links check:', {
      ozonLink: mappedPartner.ozonLink,
      wildberriesLink: mappedPartner.wildberriesLink,
      marketplace_links: mappedPartner.marketplace_links,
      has_marketplace_data: !!(mappedPartner.ozonLink || mappedPartner.wildberriesLink)
    })
    setPartner(mappedPartner)
  }, [brandingData, setPartner])

  React.useEffect(() => {
    const savedLanguage = localStorage.getItem('language') || 'ru'
    const allowed = ['ru', 'en']
    const lang = allowed.includes(savedLanguage) ? savedLanguage : 'ru'
    if (savedLanguage !== lang) {
      localStorage.setItem('language', lang)
    }
    i18n.changeLanguage(lang)
  }, [i18n])

  if (isLoading) {
    return <LoadingScreen />
  }

  return (
    <BrandingProvider>
      <Layout>
        <Routes>
          <Route path="/" element={<HomePage />} />
          <Route path="/diamond-art" element={<DiamondArtPage />} />
          <Route path="/paint-by-numbers" element={<PaintByNumbersPage />} />
          <Route path="/what-is-this" element={<WhatIsThisPage />} />
          <Route path="/mosaic-preview" element={<MosaicPreviewPage />} />
          <Route path="/editor" element={<EditorPage />} />
          <Route path="/shop" element={<ShopPage />} />
          
          {/* Новые роуты для алмазной мозаики */}
          <Route path="/diamond-mosaic" element={<DiamondMosaicPage />} />
          <Route path="/diamond-mosaic/styles" element={<DiamondMosaicPreviewPage />} />
          <Route path="/diamond-mosaic/preview-album" element={<DiamondMosaicPreviewAlbumPage />} />
          <Route path="/diamond-mosaic/editor" element={<DiamondMosaicEditorPage />} />
          <Route path="/diamond-mosaic/purchase" element={<DiamondMosaicPurchasePage />} />
          <Route path="/diamond-mosaic/success" element={<DiamondMosaicSuccessPage />} />
          
          <Route path="*" element={<NotFoundPage />} />
        </Routes>
      </Layout>
      
      <NotificationSystem
        notifications={notifications}
        onRemove={removeNotification}
      />

      <SupportChatWidget />
    </BrandingProvider>
  )
}

export default App
