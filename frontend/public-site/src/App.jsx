import React, { Suspense } from 'react';
import { Routes, Route } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import Layout from './components/Layout/Layout';
import { usePartnerStore } from './store/partnerStore';
import { useUIStore } from './store/partnerStore';
import { useBrandingQuery } from './hooks/useBrandingQuery';
import { useDocumentTitle } from './hooks/useDocumentTitle';
import LoadingScreen from './components/ui/LoadingScreen';
import NotificationSystem from './components/ui/NotificationSystem';
import SupportChatWidget from './components/SupportChatWidget';
import BrandingProvider from './components/ui/BrandingProvider';
import './assets/branding.css';

const HomePage = React.lazy(() => import('./pages/HomePage'));
const EditorPage = React.lazy(() => import('./pages/EditorPage'));
const DiamondArtPage = React.lazy(() => import('./pages/DiamondArtPage'));
const ShopPage = React.lazy(() => import('./pages/ShopPage'));
const NotFoundPage = React.lazy(() => import('./pages/NotFoundPage'));
const WhatIsThisPage = React.lazy(() => import('./pages/WhatIsThisPage'));
const MosaicPreviewPage = React.lazy(() => import('./pages/MosaicPreviewPage'));
const PreviewPage = React.lazy(() => import('./pages/PreviewPage'));
const PreviewStylesPage = React.lazy(() => import('./pages/PreviewStylesPage'));
const PreviewAlbumPage = React.lazy(() => import('./pages/PreviewAlbumPage'));

const PreviewPurchasePage = React.lazy(
  () => import('./pages/PreviewPurchasePage')
);
const PreviewSuccessPage = React.lazy(
  () => import('./pages/PreviewSuccessPage')
);
const CouponActivationPage = React.lazy(
  () => import('./pages/CouponActivationPage')
);
const ImageEditorPage = React.lazy(() => import('./pages/ImageEditorPage'));

function App() {
  const { i18n } = useTranslation();
  const { partner, setPartner } = usePartnerStore();
  const { notifications, removeNotification } = useUIStore();

  const { data: brandingData, isLoading } = useBrandingQuery();

  useDocumentTitle();

  React.useEffect(() => {
    if (!brandingData) return;

    console.log('App - Raw branding data:', brandingData);
    console.log('App - Brand name from API:', brandingData.brand_name);
    console.log('App - Is default branding:', brandingData.is_default);
    console.log('App - Partner code:', brandingData.partner_code);

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
      is_default: brandingData.is_default,
    };

    console.log('App - Mapped partner data:', mappedPartner);
    console.log('App - Marketplace links check:', {
      ozonLink: mappedPartner.ozonLink,
      wildberriesLink: mappedPartner.wildberriesLink,
      marketplace_links: mappedPartner.marketplace_links,
      has_marketplace_data: !!(
        mappedPartner.ozonLink || mappedPartner.wildberriesLink
      ),
    });
    setPartner(mappedPartner);
  }, [brandingData, setPartner]);

  React.useEffect(() => {
    const savedLanguage = localStorage.getItem('language') || 'ru';
    const allowed = ['ru', 'en'];
    const lang = allowed.includes(savedLanguage) ? savedLanguage : 'ru';
    if (savedLanguage !== lang) {
      localStorage.setItem('language', lang);
    }
    i18n.changeLanguage(lang);
  }, [i18n]);

  if (isLoading) {
    return <LoadingScreen />;
  }

  return (
    <BrandingProvider>
      <Layout>
        <Suspense fallback={<LoadingScreen />}>
          <Routes>
            <Route path="/" element={<HomePage />} />
            <Route path="/diamond-art" element={<DiamondArtPage />} />
            <Route path="/what-is-this" element={<WhatIsThisPage />} />
            <Route path="/mosaic-preview" element={<MosaicPreviewPage />} />
            <Route path="/editor" element={<EditorPage />} />
            <Route path="/shop" element={<ShopPage />} />

            {}
            <Route path="/preview" element={<PreviewPage />} />
            <Route path="/preview/styles" element={<PreviewStylesPage />} />
            <Route path="/preview/album" element={<PreviewAlbumPage />} />

            <Route path="/preview/purchase" element={<PreviewPurchasePage />} />
            <Route path="/preview/success" element={<PreviewSuccessPage />} />

            {}
            <Route path="/image-editor" element={<ImageEditorPage />} />

            {}
            <Route path="/coupon" element={<CouponActivationPage />} />

            <Route path="*" element={<NotFoundPage />} />
          </Routes>
        </Suspense>
      </Layout>

      <NotificationSystem
        notifications={notifications}
        onRemove={removeNotification}
      />

      <SupportChatWidget />
    </BrandingProvider>
  );
}

export default App;
