import { useEffect } from 'react';
import { useTranslation } from 'react-i18next';

export const useDocumentTitle = () => {
  const { t, i18n } = useTranslation();

  useEffect(() => {
    document.title = t('navigation.html_title');

    const metaDescription = document.querySelector('meta[name="description"]');
    if (metaDescription) {
      metaDescription.setAttribute('content', t('navigation.html_description'));
    }

    document.documentElement.lang = i18n.language;
  }, [t, i18n.language]);
};
