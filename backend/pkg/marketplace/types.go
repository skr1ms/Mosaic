package marketplace

import "github.com/google/uuid"

type Marketplace string

const (
	MarketplaceOzon        Marketplace = "ozon"
	MarketplaceWildberries Marketplace = "wildberries"
)

var SupportedMarketplaces = []Marketplace{
	MarketplaceOzon,
	MarketplaceWildberries,
}

var ValidSizes = []string{
	"21x30", "30x40", "40x40", "40x50", "40x60", "50x70",
}

var ValidStyles = []string{
	"grayscale", "skin_tones", "pop_art", "max_colors",
	"venus", "sun", "moon", "mars",
}

type ProductURLRequest struct {
	PartnerID   uuid.UUID   `json:"partner_id" validate:"required,uuid"`
	Marketplace Marketplace `json:"marketplace" validate:"required"`
	Size        string      `json:"size" validate:"required"`
	Style       string      `json:"style" validate:"required"`
}

type ProductURLResponse struct {
	URL         string      `json:"url"`
	SKU         string      `json:"sku"`
	HasArticle  bool        `json:"has_article"`
	PartnerName string      `json:"partner_name"`
	Marketplace Marketplace `json:"marketplace"`
	Size        string      `json:"size"`
	Style       string      `json:"style"`
}

type ProductAvailabilityRequest struct {
	PartnerID   *uuid.UUID  `json:"partner_id,omitempty"`
	Marketplace Marketplace `json:"marketplace" validate:"required"`
	Size        string      `json:"size,omitempty"`
	Style       string      `json:"style,omitempty"`
	SKU         string      `json:"sku,omitempty"`
}

type ProductAvailabilityResponse struct {
	Marketplace Marketplace `json:"marketplace"`
	SKU         string      `json:"sku"`
	Available   bool        `json:"available"`
	ProductURL  string      `json:"product_url"`
}

type Partner interface {
	GetID() uuid.UUID
	GetBrandName() string
	GetOzonLink() string
	GetOzonLinkTemplate() string
	GetWildberriesLink() string
	GetWildberriesLinkTemplate() string
}

type Article interface {
	GetSKU() string
	GetSize() string
	GetStyle() string
	GetMarketplace() string
}

type PartnerRepository interface {
	GetByID(partnerID uuid.UUID) (Partner, error)
	GetArticleBySizeStyle(partnerID uuid.UUID, size, style, marketplace string) (Article, error)
}
