package middleware

import (
	"github.com/gofiber/fiber/v2"
)

// BrandingHelper contains helper methods for working with branding in handlers
type BrandingHelper struct{}

func NewBrandingHelper() *BrandingHelper {
	return &BrandingHelper{}
}

// GetBranding returns branding data from context
func (bh *BrandingHelper) GetBranding(c *fiber.Ctx) *BrandingData {
	return GetBrandingFromContext(c)
}

// GetPartner returns partner from branding context
func (bh *BrandingHelper) GetPartner(c *fiber.Ctx) *Partner {
	return GetPartnerFromContext(c)
}

// IsDefault checks if default branding is used
func (bh *BrandingHelper) IsDefault(c *fiber.Ctx) bool {
	return IsDefaultBranding(c)
}

// GetBrandName returns brand name for current context
func (bh *BrandingHelper) GetBrandName(c *fiber.Ctx) string {
	branding := GetBrandingFromContext(c)
	if branding == nil {
		return ""
	}
	return branding.BrandName
}

// GetLogoURL returns logo URL for current context
func (bh *BrandingHelper) GetLogoURL(c *fiber.Ctx) string {
	branding := GetBrandingFromContext(c)
	if branding == nil {
		return ""
	}
	return branding.LogoURL
}

// GetContactInfo returns contact information for current context
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

// GetMarketplaceLinks returns marketplace links for current context
func (bh *BrandingHelper) GetMarketplaceLinks(c *fiber.Ctx) MarketplaceLinks {
	branding := GetBrandingFromContext(c)
	if branding == nil {
		return MarketplaceLinks{}
	}
	return branding.MarketplaceLinks
}

// PersonalizeOzonLink returns personalized Ozon link
func (bh *BrandingHelper) PersonalizeOzonLink(c *fiber.Ctx, defaultLink string) string {
	return PersonalizeMarketplaceLink(c, "ozon", defaultLink)
}

// PersonalizeWildberriesLink returns personalized Wildberries link
func (bh *BrandingHelper) PersonalizeWildberriesLink(c *fiber.Ctx, defaultLink string) string {
	return PersonalizeMarketplaceLink(c, "wildberries", defaultLink)
}

// GetContactLinks returns properly formatted contact links
func (bh *BrandingHelper) GetContactLinks(c *fiber.Ctx) map[string]string {
	branding := GetBrandingFromContext(c)
	return BuildContactLinks(branding)
}

// CanSell checks if selling is allowed in current context
func (bh *BrandingHelper) CanSell(c *fiber.Ctx) bool {
	branding := GetBrandingFromContext(c)
	if branding == nil {
		return true
	}
	return branding.AllowSales
}

// CanPurchase checks if purchasing is allowed in current context
func (bh *BrandingHelper) CanPurchase(c *fiber.Ctx) bool {
	branding := GetBrandingFromContext(c)
	if branding == nil {
		return true // By default allow
	}
	return branding.AllowPurchases
}

// CreateBrandingResponse creates structured response with branding data
func (bh *BrandingHelper) CreateBrandingResponse(c *fiber.Ctx) map[string]any {
	return BrandingResponse(c)
}

// AddBrandingToResponse adds branding data to existing response
func (bh *BrandingHelper) AddBrandingToResponse(c *fiber.Ctx, response map[string]any) map[string]any {
	if response == nil {
		response = make(map[string]any)
	}

	response["branding"] = bh.CreateBrandingResponse(c)
	return response
}

// ValidatePartnerAccess checks if current partner has access to resource
func (bh *BrandingHelper) ValidatePartnerAccess(c *fiber.Ctx, requiredPartnerID string) bool {
	partner := bh.GetPartner(c)
	if partner == nil {
		return false
	}

	return partner.ID.String() == requiredPartnerID || partner.PartnerCode == requiredPartnerID
}

// GetPartnerCode returns partner code from context
func (bh *BrandingHelper) GetPartnerCode(c *fiber.Ctx) string {
	partner := bh.GetPartner(c)
	if partner == nil {
		return ""
	}
	return partner.PartnerCode
}

// GetPartnerID returns partner ID from context
func (bh *BrandingHelper) GetPartnerID(c *fiber.Ctx) string {
	partner := bh.GetPartner(c)
	if partner == nil {
		return ""
	}
	return partner.ID.String()
}
