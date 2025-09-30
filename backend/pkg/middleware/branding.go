package middleware

import (
	"context"
	"net/url"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// Partner represents partner for branding
type Partner struct {
	bun.BaseModel `bun:"table:partners,alias:p"`

	ID              uuid.UUID `bun:"id,pk,type:uuid" json:"id"`
	PartnerCode     string    `bun:"partner_code" json:"partner_code"`
	Domain          string    `bun:"domain" json:"domain"`
	BrandName       string    `bun:"brand_name" json:"brand_name"`
	LogoURL         string    `bun:"logo_url" json:"logo_url"`
	OzonLink        string    `bun:"ozon_link" json:"ozon_link"`
	WildberriesLink string    `bun:"wildberries_link" json:"wildberries_link"`
	Email           string    `bun:"email" json:"email"`
	Address         string    `bun:"address" json:"address"`
	Phone           string    `bun:"phone" json:"phone"`
	Telegram        string    `bun:"telegram" json:"telegram"`
	Whatsapp        string    `bun:"whatsapp" json:"whatsapp"`
	TelegramLink    string    `bun:"telegram_link" json:"telegram_link"`
	WhatsappLink    string    `bun:"whatsapp_link" json:"whatsapp_link"`
	BrandColors     []string  `bun:"brand_colors,array" json:"brand_colors"`
	AllowSales      bool      `bun:"allow_sales" json:"allow_sales"`
	AllowPurchases  bool      `bun:"allow_purchases" json:"allow_purchases"`
	Status          string    `bun:"status" json:"status"`
}

// BrandingData contains all branding information
type BrandingData struct {
	Partner          *Partner         `json:"partner,omitempty"`
	IsDefault        bool             `json:"is_default"`
	BrandName        string           `json:"brand_name"`
	LogoURL          string           `json:"logo_url"`
	ContactEmail     string           `json:"contact_email"`
	ContactAddress   string           `json:"contact_address"`
	ContactPhone     string           `json:"contact_phone"`
	ContactTelegram  string           `json:"contact_telegram"`
	ContactWhatsapp  string           `json:"contact_whatsapp"`
	TelegramLink     string           `json:"telegram_link"`
	WhatsappLink     string           `json:"whatsapp_link"`
	MarketplaceLinks MarketplaceLinks `json:"marketplace_links"`
	AllowSales       bool             `json:"allow_sales"`
	AllowPurchases   bool             `json:"allow_purchases"`
	BrandColors      []string         `json:"brand_colors"`
}

// MarketplaceLinks contains marketplace links
type MarketplaceLinks struct {
	Ozon        string `json:"ozon"`
	Wildberries string `json:"wildberries"`
}

// DefaultBranding contains default branding values
type DefaultBranding struct {
	BrandName       string
	LogoURL         string
	ContactEmail    string
	ContactAddress  string
	ContactPhone    string
	ContactTelegram string
	ContactWhatsapp string
	TelegramLink    string
	WhatsappLink    string
	OzonLink        string
	WildberriesLink string
}

// BrandingMiddleware determines partner by domain
type BrandingMiddleware struct {
	db               *bun.DB
	defaultBranding  DefaultBranding
	logger           *Logger
}

func NewBrandingMiddleware(db *bun.DB, defaultBranding DefaultBranding, logger *Logger) *BrandingMiddleware {
	middleware := &BrandingMiddleware{
		db:              db,
		defaultBranding: defaultBranding,
		logger:          logger,
	}

	return middleware
}

// logPartnerDetectionAsync logs partner detection asynchronously
func (b *BrandingMiddleware) logPartnerDetectionAsync(domain, partnerCode string, found bool) {
	go func() {
		b.logger.GetZerologLogger().Info().
			Str("domain", domain).
			Str("partner_code", partnerCode).
			Bool("found", found).
			Msg("Partner detection")
	}()
}

// logDefaultBrandingAsync logs default branding usage asynchronously
func (b *BrandingMiddleware) logDefaultBrandingAsync(host string) {
	go func() {
		b.logger.GetZerologLogger().Debug().
			Str("domain", host).
			Str("event_type", "default_branding").
			Msg("Using default branding")
	}()
}

// logBrandingDataAsync logs branding data asynchronously
func (b *BrandingMiddleware) logBrandingDataAsync(domain string, brandingData *BrandingData) {
	go func() {
		b.logger.GetZerologLogger().Debug().
			Str("domain", domain).
			Str("brand_name", brandingData.BrandName).
			Str("logo_url", brandingData.LogoURL).
			Msg("Branding data")
	}()
}

// BrandingMiddlewareHandler creates middleware for automatic branding substitution
func (b *BrandingMiddleware) BrandingMiddlewareHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		host := c.Get("Host")
		if host == "" {
			host = c.Hostname()
		}

		if colonIndex := strings.Index(host, ":"); colonIndex != -1 {
			host = host[:colonIndex]
		}

		brandingData := getBrandingData(c.Context(), b.db, host, b.defaultBranding)

		c.Locals("branding", brandingData)

		if brandingData.Partner != nil {
			b.logPartnerDetectionAsync(host, brandingData.Partner.PartnerCode, true)
		} else {
			b.logPartnerDetectionAsync(host, "", false)
			b.logDefaultBrandingAsync(host)
		}

		b.logBrandingDataAsync(host, brandingData)

		return c.Next()
	}
}

// getBrandingData gets branding data based on domain
func getBrandingData(ctx context.Context, db *bun.DB, domain string, defaultBranding DefaultBranding) *BrandingData {
	partner, err := getPartnerByDomain(ctx, db, domain)
	if err != nil || partner == nil || partner.Status != "active" {
		return &BrandingData{
			Partner:         nil,
			IsDefault:       true,
			BrandName:       defaultBranding.BrandName,
			LogoURL:         defaultBranding.LogoURL,
			ContactEmail:    defaultBranding.ContactEmail,
			ContactAddress:  defaultBranding.ContactAddress,
			ContactPhone:    defaultBranding.ContactPhone,
			ContactTelegram: defaultBranding.ContactTelegram,
			ContactWhatsapp: defaultBranding.ContactWhatsapp,
			TelegramLink:    defaultBranding.TelegramLink,
			WhatsappLink:    defaultBranding.WhatsappLink,
			MarketplaceLinks: MarketplaceLinks{
				Ozon:        defaultBranding.OzonLink,
				Wildberries: defaultBranding.WildberriesLink,
			},
			AllowSales:     true,
			AllowPurchases: true,
			BrandColors:    nil,
		}
	}

	return &BrandingData{
		Partner:         partner,
		IsDefault:       false,
		BrandName:       partner.BrandName,
		LogoURL:         partner.LogoURL,
		ContactEmail:    partner.Email,
		ContactAddress:  partner.Address,
		ContactPhone:    partner.Phone,
		ContactTelegram: partner.Telegram,
		ContactWhatsapp: partner.Whatsapp,
		TelegramLink:    partner.TelegramLink,
		WhatsappLink:    partner.WhatsappLink,
		MarketplaceLinks: MarketplaceLinks{
			Ozon:        partner.OzonLink,
			Wildberries: partner.WildberriesLink,
		},
		AllowSales:     partner.AllowSales,
		AllowPurchases: partner.AllowPurchases,
		BrandColors:    partner.BrandColors,
	}
}

// getPartnerByDomain gets partner by domain from database
func getPartnerByDomain(ctx context.Context, db *bun.DB, domain string) (*Partner, error) {
	queryCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	partner := new(Partner)
	err := db.NewSelect().
		Model(partner).
		Where("domain = ? AND status = ?", domain, "active").
		Scan(queryCtx)

	if err != nil {
		return nil, err
	}

	return partner, nil
}

// GetBrandingFromContext extracts branding data from Fiber context
func GetBrandingFromContext(c *fiber.Ctx) *BrandingData {
	branding := c.Locals("branding")
	if branding == nil {
		return nil
	}

	brandingData, ok := branding.(*BrandingData)
	if !ok {
		return nil
	}

	return brandingData
}

// GetPartnerFromContext extracts partner from branding context
func GetPartnerFromContext(c *fiber.Ctx) *Partner {
	branding := GetBrandingFromContext(c)
	if branding == nil {
		return nil
	}

	return branding.Partner
}

// IsDefaultBranding checks if default branding is used
func IsDefaultBranding(c *fiber.Ctx) bool {
	branding := GetBrandingFromContext(c)
	if branding == nil {
		return true
	}

	return branding.IsDefault
}

// PersonalizeMarketplaceLink personalizes marketplace link
// If partner has own link, it's used, otherwise - general one
func PersonalizeMarketplaceLink(c *fiber.Ctx, marketplace string, defaultLink string) string {
	branding := GetBrandingFromContext(c)
	if branding == nil || branding.IsDefault {
		return defaultLink
	}

	switch strings.ToLower(marketplace) {
	case "ozon":
		if branding.MarketplaceLinks.Ozon != "" {
			return branding.MarketplaceLinks.Ozon
		}
	case "wildberries", "wb":
		if branding.MarketplaceLinks.Wildberries != "" {
			return branding.MarketplaceLinks.Wildberries
		}
	}

	return defaultLink
}

// BuildContactLinks creates properly formatted contact links
func BuildContactLinks(branding *BrandingData) map[string]string {
	links := make(map[string]string)

	if branding == nil {
		return links
	}

	if branding.TelegramLink != "" {
		links["telegram"] = branding.TelegramLink
	} else if branding.ContactTelegram != "" {
		telegram := strings.TrimPrefix(branding.ContactTelegram, "@")
		links["telegram"] = "https://t.me/" + telegram
	}

	if branding.WhatsappLink != "" {
		links["whatsapp"] = branding.WhatsappLink
	} else if branding.ContactWhatsapp != "" {
		phone := strings.ReplaceAll(branding.ContactWhatsapp, "+", "")
		phone = strings.ReplaceAll(phone, " ", "")
		phone = strings.ReplaceAll(phone, "-", "")
		phone = strings.ReplaceAll(phone, "(", "")
		phone = strings.ReplaceAll(phone, ")", "")
		links["whatsapp"] = "https://wa.me/" + phone
	}

	if branding.ContactEmail != "" {
		links["email"] = "mailto:" + branding.ContactEmail
	}

	if branding.ContactPhone != "" {
		phone := strings.ReplaceAll(branding.ContactPhone, " ", "")
		links["phone"] = "tel:" + phone
	}

	return links
}

// ValidateMarketplaceURL validates marketplace URL correctness
func ValidateMarketplaceURL(urlStr string, marketplace string) bool {
	if urlStr == "" {
		return true
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return false
	}

	switch strings.ToLower(marketplace) {
	case "ozon":
		return strings.Contains(strings.ToLower(parsedURL.Host), "ozon")
	case "wildberries", "wb":
		return strings.Contains(strings.ToLower(parsedURL.Host), "wildberries") ||
			strings.Contains(strings.ToLower(parsedURL.Host), "wb.ru")
	}

	return true
}

// BrandingResponse creates response structure with branding data for API
func BrandingResponse(c *fiber.Ctx) map[string]any {
	branding := GetBrandingFromContext(c)
	if branding == nil {
		return map[string]any{
			"is_default": true,
		}
	}

	contactLinks := BuildContactLinks(branding)

	response := map[string]any{
		"is_default":        branding.IsDefault,
		"brand_name":        branding.BrandName,
		"logo_url":          branding.LogoURL,
		"contact_email":     branding.ContactEmail,
		"contact_address":   branding.ContactAddress,
		"contact_phone":     branding.ContactPhone,
		"contact_telegram":  branding.ContactTelegram,
		"contact_whatsapp":  branding.ContactWhatsapp,
		"contact_links":     contactLinks,
		"marketplace_links": branding.MarketplaceLinks,
		"allow_sales":       branding.AllowSales,
		"allow_purchases":   branding.AllowPurchases,
		"brand_colors":      branding.BrandColors,
	}

	if branding.Partner != nil {
		response["partner_code"] = branding.Partner.PartnerCode
		response["partner_id"] = branding.Partner.ID
	}

	return response
}
