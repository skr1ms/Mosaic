import React from 'react';
import { useBrandColors } from '../../hooks/useBrandingQuery';

const BrandingProvider = ({ children }) => {
  const { cssVariables } = useBrandColors();

  React.useEffect(() => {
    Object.entries(cssVariables).forEach(([property, value]) => {
      document.documentElement.style.setProperty(property, value);
    });

    return () => {
      Object.keys(cssVariables).forEach(property => {
        document.documentElement.style.removeProperty(property);
      });
    };
  }, [cssVariables]);

  return <>{children}</>;
};

export default BrandingProvider;
