package middleware

import (
	"github.com/gofiber/fiber/v2"
)

// BrandingHelper содержит вспомогательные методы для работы с брендингом в handlers
type BrandingHelper struct{}

func NewBrandingHelper() *BrandingHelper {
	return &BrandingHelper{}
}

// GetBranding возвращает данные брендинга из контекста
func (bh *BrandingHelper) GetBranding(c *fiber.Ctx) *BrandingData {
	return GetBrandingFromContext(c)
}

// GetPartner возвращает партнера из контекста брендинга
func (bh *BrandingHelper) GetPartner(c *fiber.Ctx) *Partner {
	return GetPartnerFromContext(c)
}

// IsDefault проверяет, используется ли брендинг по умолчанию
func (bh *BrandingHelper) IsDefault(c *fiber.Ctx) bool {
	return IsDefaultBranding(c)
}

// GetBrandName возвращает название бренда для текущего контекста
func (bh *BrandingHelper) GetBrandName(c *fiber.Ctx) string {
	branding := GetBrandingFromContext(c)
	if branding == nil {
		return ""
	}
	return branding.BrandName
}

// GetLogoURL возвращает URL логотипа для текущего контекста
func (bh *BrandingHelper) GetLogoURL(c *fiber.Ctx) string {
	branding := GetBrandingFromContext(c)
	if branding == nil {
		return ""
	}
	return branding.LogoURL
}

// GetContactInfo возвращает контактную информацию для текущего контекста
func (bh *BrandingHelper) GetContactInfo(c *fiber.Ctx) map[string]string {
	branding := GetBrandingFromContext(c)
	if branding == nil {
		return map[string]string{}
	}

	return map[string]string{
		"email":    branding.ContactEmail,
		"address":  branding.ContactAddress,
		"phone":    branding.ContactPhone,
		"telegram": branding.ContactTelegram,
		"whatsapp": branding.ContactWhatsapp,
	}
}

// GetMarketplaceLinks возвращает ссылки на маркетплейсы для текущего контекста
func (bh *BrandingHelper) GetMarketplaceLinks(c *fiber.Ctx) MarketplaceLinks {
	branding := GetBrandingFromContext(c)
	if branding == nil {
		return MarketplaceLinks{}
	}
	return branding.MarketplaceLinks
}

// PersonalizeOzonLink возвращает персонализированную ссылку на Ozon
func (bh *BrandingHelper) PersonalizeOzonLink(c *fiber.Ctx, defaultLink string) string {
	return PersonalizeMarketplaceLink(c, "ozon", defaultLink)
}

// PersonalizeWildberriesLink возвращает персонализированную ссылку на Wildberries
func (bh *BrandingHelper) PersonalizeWildberriesLink(c *fiber.Ctx, defaultLink string) string {
	return PersonalizeMarketplaceLink(c, "wildberries", defaultLink)
}

// GetContactLinks возвращает правильно отформатированные ссылки на контакты
func (bh *BrandingHelper) GetContactLinks(c *fiber.Ctx) map[string]string {
	branding := GetBrandingFromContext(c)
	return BuildContactLinks(branding)
}

// CanSell проверяет, разрешена ли продажа в текущем контексте
func (bh *BrandingHelper) CanSell(c *fiber.Ctx) bool {
	branding := GetBrandingFromContext(c)
	if branding == nil {
		return true // По умолчанию разрешаем
	}
	return branding.AllowSales
}

// CanPurchase проверяет, разрешена ли покупка в текущем контексте
func (bh *BrandingHelper) CanPurchase(c *fiber.Ctx) bool {
	branding := GetBrandingFromContext(c)
	if branding == nil {
		return true // По умолчанию разрешаем
	}
	return branding.AllowPurchases
}

// CreateBrandingResponse создает структурированный ответ с данными брендинга
func (bh *BrandingHelper) CreateBrandingResponse(c *fiber.Ctx) map[string]interface{} {
	return BrandingResponse(c)
}

// AddBrandingToResponse добавляет данные брендинга к существующему ответу
func (bh *BrandingHelper) AddBrandingToResponse(c *fiber.Ctx, response map[string]interface{}) map[string]interface{} {
	if response == nil {
		response = make(map[string]interface{})
	}

	response["branding"] = bh.CreateBrandingResponse(c)
	return response
}

// ValidatePartnerAccess проверяет, имеет ли текущий партнер доступ к ресурсу
func (bh *BrandingHelper) ValidatePartnerAccess(c *fiber.Ctx, requiredPartnerID string) bool {
	partner := bh.GetPartner(c)
	if partner == nil {
		return false
	}

	return partner.ID.String() == requiredPartnerID || partner.PartnerCode == requiredPartnerID
}

// GetPartnerCode возвращает код партнера из контекста
func (bh *BrandingHelper) GetPartnerCode(c *fiber.Ctx) string {
	partner := bh.GetPartner(c)
	if partner == nil {
		return ""
	}
	return partner.PartnerCode
}

// GetPartnerID возвращает ID партнера из контекста
func (bh *BrandingHelper) GetPartnerID(c *fiber.Ctx) string {
	partner := bh.GetPartner(c)
	if partner == nil {
		return ""
	}
	return partner.ID.String()
}
