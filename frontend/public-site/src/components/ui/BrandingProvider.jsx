import React from 'react'
import { useBrandColors } from '../../hooks/useBrandingQuery'

const BrandingProvider = ({ children }) => {
  const { cssVariables } = useBrandColors()

  React.useEffect(() => {
    // Применяем CSS переменные к document.documentElement
    Object.entries(cssVariables).forEach(([property, value]) => {
      document.documentElement.style.setProperty(property, value)
    })

    // Очистка при размонтировании
    return () => {
      Object.keys(cssVariables).forEach(property => {
        document.documentElement.style.removeProperty(property)
      })
    }
  }, [cssVariables])

  return <>{children}</>
}

export default BrandingProvider
