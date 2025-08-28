import { useEffect } from 'react'
import { useTranslation } from 'react-i18next'

export const useDocumentTitle = () => {
  const { t, i18n } = useTranslation()

  useEffect(() => {
    // Update title
    document.title = t('navigation.page_title')
    
    // Update meta description
    const metaDescription = document.querySelector('meta[name="description"]')
    if (metaDescription) {
      metaDescription.setAttribute('content', t('navigation.page_description'))
    }
    
    // Update HTML lang attribute
    document.documentElement.lang = i18n.language
  }, [t, i18n.language])
}
