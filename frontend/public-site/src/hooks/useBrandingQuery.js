import { useQuery } from '@tanstack/react-query';
import { MosaicAPI } from '../api/client';
import { usePartnerStore } from '../store/partnerStore';

export const useBrandingQuery = () => {
  return useQuery({
    queryKey: ['branding', window.location.hostname],
    queryFn: () => MosaicAPI.getBrandingInfo(),
    staleTime: 10 * 60 * 1000,
    retry: 1,
    refetchOnWindowFocus: false,
  });
};

export const useCouponQuery = code => {
  return useQuery({
    queryKey: ['coupon', code],
    queryFn: () => MosaicAPI.validateCoupon(code),
    enabled: !!code && code.length === 12,
    retry: false,
    refetchOnWindowFocus: false,
  });
};

export const useImageStatusQuery = (imageId, enabled = false) => {
  return useQuery({
    queryKey: ['imageStatus', imageId],
    queryFn: () => MosaicAPI.getProcessingStatus(imageId),
    enabled: !!imageId && enabled,
    refetchInterval: 2000,
    retry: 3,
  });
};

export const useBrandColors = () => {
  const { getBrandColors, getPrimaryColor, getSecondaryColor, getAccentColor } =
    usePartnerStore();

  return {
    brandColors: getBrandColors(),
    primaryColor: getPrimaryColor(),
    secondaryColor: getSecondaryColor(),
    accentColor: getAccentColor(),
    cssVariables: {
      '--brand-primary': getPrimaryColor(),
      '--brand-secondary': getSecondaryColor(),
      '--brand-accent': getAccentColor(),
    },
  };
};
