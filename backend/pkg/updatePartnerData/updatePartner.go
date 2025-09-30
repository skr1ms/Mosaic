package updatePartnerData

import "github.com/skr1ms/mosaic/internal/partner"

// UpdatePartnerData updates the partner's data
func UpdatePartnerData(partner *partner.Partner, req *partner.UpdatePartnerRequest) {
	if req.Login != nil && *req.Login != "" {
		partner.Login = *req.Login
	}
	if req.Password != nil && *req.Password != "" {
		partner.Password = *req.Password
	}
	if req.Domain != nil && *req.Domain != "" {
		partner.Domain = *req.Domain
	}
	if req.BrandName != nil && *req.BrandName != "" {
		partner.BrandName = *req.BrandName
	}
	if req.LogoURL != nil && *req.LogoURL != "" {
		partner.LogoURL = *req.LogoURL
	}
	if req.OzonLink != nil && *req.OzonLink != "" {
		partner.OzonLink = *req.OzonLink
	}
	if req.WildberriesLink != nil && *req.WildberriesLink != "" {
		partner.WildberriesLink = *req.WildberriesLink
	}
	if req.Email != nil && *req.Email != "" {
		partner.Email = *req.Email
	}
	if req.Address != nil && *req.Address != "" {
		partner.Address = *req.Address
	}
	if req.Phone != nil && *req.Phone != "" {
		partner.Phone = *req.Phone
	}
	if req.Telegram != nil && *req.Telegram != "" {
		partner.Telegram = *req.Telegram
	}
	if req.Whatsapp != nil && *req.Whatsapp != "" {
		partner.Whatsapp = *req.Whatsapp
	}
	if req.AllowSales != nil && !(*req.AllowSales) {
		partner.AllowSales = *req.AllowSales
	}
	if req.Status != nil && *req.Status != "" {
		partner.Status = *req.Status
	}
	if req.BrandColors != nil {
		partner.BrandColors = *req.BrandColors
	}
}
