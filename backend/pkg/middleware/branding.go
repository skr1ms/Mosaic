package middleware

import (
	"context"
	"net/url"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/uptrace/bun"
)

// Partner представляет партнера для брендинга
type Partner struct {
	ID              uuid.UUID `json:"id"`
	PartnerCode     string    `json:"partner_code"`
	Domain          string    `json:"domain"`
	BrandName       string    `json:"brand_name"`
	LogoURL         string    `json:"logo_url"`
	OzonLink        string    `json:"ozon_link"`
	WildberriesLink string    `json:"wildberries_link"`
	Email           string    `json:"email"`
	Address         string    `json:"address"`
	Phone           string    `json:"phone"`
	Telegram        string    `json:"telegram"`
	Whatsapp        string    `json:"whatsapp"`
	TelegramLink    string    `json:"telegram_link"`
	WhatsappLink    string    `json:"whatsapp_link"`
	AllowSales      bool      `json:"allow_sales"`
	AllowPurchases  bool      `json:"allow_purchases"`
	Status          string    `json:"status"`
}

// BrandingData содержит всю информацию для брендинга
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
}

// MarketplaceLinks содержит ссылки на маркетплейсы
type MarketplaceLinks struct {
	Ozon        string `json:"ozon"`
	Wildberries string `json:"wildberries"`
}

// DefaultBranding содержит значения по умолчанию для брендинга
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

// BrandingMiddleware создает middleware для автоматической подстановки брендинга
func BrandingMiddleware(db *bun.DB, defaultBranding DefaultBranding) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Определяем домен из заголовка Host
		host := c.Get("Host")
		if host == "" {
			host = c.Hostname()
		}

		// Очищаем домен от порта
		if colonIndex := strings.Index(host, ":"); colonIndex != -1 {
			host = host[:colonIndex]
		}

		// Получаем данные для брендинга
		brandingData := getBrandingData(c.Context(), db, host, defaultBranding)

		// Сохраняем данные брендинга в контексте
		c.Locals("branding", brandingData)

		// Логируем определение партнера
		if brandingData.Partner != nil {
			go func() {
				log.Info().
					Str("domain", host).
					Str("partner_code", brandingData.Partner.PartnerCode).
					Str("brand_name", brandingData.Partner.BrandName).
					Str("event_type", "partner_detected").
					Msg("Partner detected by domain")
			}()
		} else {
			go func() {
				log.Debug().
					Str("domain", host).
					Str("event_type", "default_branding").
					Msg("Using default branding")
			}()
		}

		return c.Next()
	}
}

// getBrandingData получает данные для брендинга на основе домена
func getBrandingData(ctx context.Context, db *bun.DB, domain string, defaultBranding DefaultBranding) *BrandingData {
	// Пытаемся найти партнера по домену
	partner, err := getPartnerByDomain(ctx, db, domain)
	if err != nil || partner == nil || partner.Status != "active" {
		// Возвращаем данные по умолчанию
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
		}
	}

	// Возвращаем брендированные данные партнера
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
	}
}

// getPartnerByDomain получает партнера по домену из базы данных
func getPartnerByDomain(ctx context.Context, db *bun.DB, domain string) (*Partner, error) {
	// Создаем контекст с таймаутом для запроса к БД
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

// GetBrandingFromContext извлекает данные брендинга из контекста Fiber
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

// GetPartnerFromContext извлекает партнера из контекста брендинга
func GetPartnerFromContext(c *fiber.Ctx) *Partner {
	branding := GetBrandingFromContext(c)
	if branding == nil {
		return nil
	}

	return branding.Partner
}

// IsDefaultBranding проверяет, используется ли брендинг по умолчанию
func IsDefaultBranding(c *fiber.Ctx) bool {
	branding := GetBrandingFromContext(c)
	if branding == nil {
		return true
	}

	return branding.IsDefault
}

// PersonalizeMarketplaceLink персонализирует ссылку на маркетплейс
// Если у партнера есть своя ссылка, используется она, иначе - общая
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

// BuildContactLinks создает правильно отформатированные ссылки на контакты
func BuildContactLinks(branding *BrandingData) map[string]string {
	links := make(map[string]string)

	if branding == nil {
		return links
	}

	// Telegram ссылка
	if branding.TelegramLink != "" {
		links["telegram"] = branding.TelegramLink
	} else if branding.ContactTelegram != "" {
		// Если есть только username, создаем ссылку
		telegram := strings.TrimPrefix(branding.ContactTelegram, "@")
		links["telegram"] = "https://t.me/" + telegram
	}

	// WhatsApp ссылка
	if branding.WhatsappLink != "" {
		links["whatsapp"] = branding.WhatsappLink
	} else if branding.ContactWhatsapp != "" {
		// Если есть только номер, создаем ссылку
		phone := strings.ReplaceAll(branding.ContactWhatsapp, "+", "")
		phone = strings.ReplaceAll(phone, " ", "")
		phone = strings.ReplaceAll(phone, "-", "")
		phone = strings.ReplaceAll(phone, "(", "")
		phone = strings.ReplaceAll(phone, ")", "")
		links["whatsapp"] = "https://wa.me/" + phone
	}

	// Email ссылка
	if branding.ContactEmail != "" {
		links["email"] = "mailto:" + branding.ContactEmail
	}

	// Телефон ссылка
	if branding.ContactPhone != "" {
		phone := strings.ReplaceAll(branding.ContactPhone, " ", "")
		links["phone"] = "tel:" + phone
	}

	return links
}

// ValidateMarketplaceURL проверяет корректность URL маркетплейса
func ValidateMarketplaceURL(urlStr string, marketplace string) bool {
	if urlStr == "" {
		return true // Пустая ссылка допустима
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

// BrandingResponse создает структуру ответа с данными брендинга для API
func BrandingResponse(c *fiber.Ctx) map[string]interface{} {
	branding := GetBrandingFromContext(c)
	if branding == nil {
		return map[string]interface{}{
			"is_default": true,
		}
	}

	contactLinks := BuildContactLinks(branding)

	response := map[string]interface{}{
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
	}

	if branding.Partner != nil {
		response["partner_code"] = branding.Partner.PartnerCode
		response["partner_id"] = branding.Partner.ID
	}

	return response
}
