import { useQuery } from '@tanstack/react-query'
import { MosaicAPI } from '../api/client'
import { usePartnerStore } from '../store/partnerStore'

export const useBrandingQuery = () => {
  return useQuery({
    queryKey: ['branding', window.location.hostname],
    queryFn: () => MosaicAPI.getBrandingInfo(),
    staleTime: 10 * 60 * 1000, // 10 minutes
    retry: 1,
    refetchOnWindowFocus: false,
  })
}

export const useCouponQuery = (code) => {
  return useQuery({
    queryKey: ['coupon', code],
    queryFn: () => MosaicAPI.validateCoupon(code),
    enabled: !!code && code.length === 12,
    retry: false,
    refetchOnWindowFocus: false,
  })
}

export const useImageStatusQuery = (imageId, enabled = false) => {
  return useQuery({
    queryKey: ['imageStatus', imageId],
    queryFn: () => MosaicAPI.getProcessingStatus(imageId),
    enabled: !!imageId && enabled,
    refetchInterval: 2000, // Poll every 2 seconds
    retry: 3,
  })
}

// Хук для использования цветов брендинга
export const useBrandColors = () => {
  const { getBrandColors, getPrimaryColor, getSecondaryColor, getAccentColor } = usePartnerStore()
  
  return {
    brandColors: getBrandColors(),
    primaryColor: getPrimaryColor(),
    secondaryColor: getSecondaryColor(),
    accentColor: getAccentColor(),
    // CSS переменные для использования в стилях
    cssVariables: {
      '--brand-primary': getPrimaryColor(),
      '--brand-secondary': getSecondaryColor(),
      '--brand-accent': getAccentColor(),
    }
  }
}
