package partner

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Partner struct {
	bun.BaseModel `bun:"table:partners,alias:p"`

	ID          uuid.UUID  `bun:"id,pk,type:uuid,default:gen_random_uuid()" json:"id"`
	PartnerCode string     `bun:"partner_code,unique,notnull" json:"partner_code"`
	Login       string     `bun:"login,unique,notnull" json:"login"`
	Password    string     `bun:"password,notnull" json:"password"`
	LastLogin   *time.Time `bun:"last_login" json:"last_login"`
	Domain      string     `bun:"domain,unique,notnull" json:"domain"`
	BrandName   string     `bun:"brand_name,notnull" json:"brand_name"`
	LogoURL     string     `bun:"logo_url" json:"logo_url"`

	OzonLink        string `bun:"ozon_link" json:"ozon_link"`
	WildberriesLink string `bun:"wildberries_link" json:"wildberries_link"`

	OzonLinkTemplate        string `bun:"ozon_link_template" json:"ozon_link_template"`
	WildberriesLinkTemplate string `bun:"wildberries_link_template" json:"wildberries_link_template"`

	Email           string    `bun:"email,notnull" json:"email"`
	Address         string    `bun:"address" json:"address"`
	Phone           string    `bun:"phone" json:"phone"`
	Telegram        string    `bun:"telegram" json:"telegram"`
	Whatsapp        string    `bun:"whatsapp" json:"whatsapp"`
	TelegramLink    string    `bun:"telegram_link" json:"telegram_link"`
	WhatsappLink    string    `bun:"whatsapp_link" json:"whatsapp_link"`
	BrandColors     []string  `bun:"brand_colors,array" json:"brand_colors"`
	AllowSales      bool      `bun:"allow_sales,default:true" json:"allow_sales"`
	AllowPurchases  bool      `bun:"allow_purchases,default:true" json:"allow_purchases"`
	Status          string    `bun:"status,type:partner_status,default:'active'" json:"status"`
	IsBlockedInChat bool      `bun:"is_blocked_in_chat,default:false" json:"is_blocked_in_chat"`
	CreatedAt       time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt       time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`

	Articles []*PartnerArticle `bun:"rel:has-many,join:id=partner_id" json:"articles"`
}

type PartnerArticle struct {
	bun.BaseModel `bun:"table:partner_articles,alias:pa"`

	ID        uuid.UUID `bun:"id,pk,type:uuid,default:gen_random_uuid()" json:"id"`
	PartnerID uuid.UUID `bun:"partner_id,type:uuid,notnull" json:"partner_id"`
	Partner   *Partner  `bun:"rel:belongs-to,join:partner_id=id" json:"-"`

	Size        string `bun:"size,notnull" json:"size"`
	Style       string `bun:"style,notnull" json:"style"`
	Marketplace string `bun:"marketplace,notnull" json:"marketplace"`

	SKU string `bun:"sku" json:"sku"`

	IsActive  bool      `bun:"is_active,default:true" json:"is_active"`
	CreatedAt time.Time `bun:"created_at,default:current_timestamp" json:"created_at"`
	UpdatedAt time.Time `bun:"updated_at,default:current_timestamp" json:"updated_at"`
}

const (
	Size21x30 = "21x30"
	Size30x40 = "30x40"
	Size40x40 = "40x40"
	Size40x50 = "40x50"
	Size40x60 = "40x60"
	Size50x70 = "50x70"

	StyleGrayscale = "grayscale"
	StyleSkinTones = "skin_tones"
	StylePopArt    = "pop_art"
	StyleMaxColors = "max_colors"

	MarketplaceOzon        = "ozon"
	MarketplaceWildberries = "wildberries"
)

var (
	AvailableSizes  = []string{Size21x30, Size30x40, Size40x40, Size40x50, Size40x60, Size50x70}
	AvailableStyles = []string{StyleGrayscale, StyleSkinTones, StylePopArt, StyleMaxColors}
	Marketplaces    = []string{MarketplaceOzon, MarketplaceWildberries}
)

func (p *Partner) CreateIndex() string {
	return `
	CREATE INDEX IF NOT EXISTS idx_partners_partner_code ON partners(partner_code);
	CREATE INDEX IF NOT EXISTS idx_partners_login ON partners(login);
	CREATE INDEX IF NOT EXISTS idx_partners_domain ON partners(domain);
	CREATE INDEX IF NOT EXISTS idx_partners_status ON partners(status);
    CREATE INDEX IF NOT EXISTS idx_partners_is_blocked_in_chat ON partners(is_blocked_in_chat);
	CREATE INDEX IF NOT EXISTS idx_partners_created_at ON partners(created_at);
	CREATE INDEX IF NOT EXISTS idx_partners_updated_at ON partners(updated_at);
	`
}

func CreatePartnerArticlesIndex() string {
	return `
	CREATE INDEX IF NOT EXISTS idx_partner_articles_partner_id ON partner_articles(partner_id);
	CREATE INDEX IF NOT EXISTS idx_partner_articles_size_style ON partner_articles(size, style);
	CREATE INDEX IF NOT EXISTS idx_partner_articles_marketplace ON partner_articles(marketplace);
	CREATE INDEX IF NOT EXISTS idx_partner_articles_sku ON partner_articles(sku);
	CREATE UNIQUE INDEX IF NOT EXISTS idx_partner_articles_unique ON partner_articles(partner_id, size, style, marketplace);
	`
}
