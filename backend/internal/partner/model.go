package partner

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Partner struct {
	bun.BaseModel `bun:"table:partners,alias:p"`

	ID              uuid.UUID  `bun:"id,pk,type:uuid,default:gen_random_uuid()" json:"id"`
	PartnerCode     string     `bun:"partner_code,unique,notnull" json:"partner_code"` // 4-значный код партнера (0000-9999). 0000 зарезервирован для собственных купонов, 0001+ для партнеров
	Login           string     `bun:"login,unique,notnull" json:"login"`
	Password        string     `bun:"password,notnull" json:"password"` // Хешированный пароль
	LastLogin       *time.Time `bun:"last_login" json:"last_login"`
	Domain          string     `bun:"domain,unique,notnull" json:"domain"`
	BrandName       string     `bun:"brand_name,notnull" json:"brand_name"`
	LogoURL         string     `bun:"logo_url" json:"logo_url"`
	OzonLink        string     `bun:"ozon_link" json:"ozon_link"`
	WildberriesLink string     `bun:"wildberries_link" json:"wildberries_link"`
	Email           string     `bun:"email,notnull" json:"email"`
	Address         string     `bun:"address" json:"address"`
	Phone           string     `bun:"phone" json:"phone"`
	Telegram        string     `bun:"telegram" json:"telegram"`
	Whatsapp        string     `bun:"whatsapp" json:"whatsapp"`
	TelegramLink    string     `bun:"telegram_link" json:"telegram_link"` // Полная ссылка на Telegram
	WhatsappLink    string     `bun:"whatsapp_link" json:"whatsapp_link"` // Полная ссылка на WhatsApp
	AllowSales      bool       `bun:"allow_sales,default:true" json:"allow_sales"`
	AllowPurchases  bool       `bun:"allow_purchases,default:true" json:"allow_purchases"` // Разрешить покупки через брендированную версию
	Status          string     `bun:"status,type:partner_status,default:'active'" json:"status"`
	CreatedAt       time.Time  `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt       time.Time  `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`
}

// - idx_partners_partner_code: быстрый поиск по коду партнера
// - idx_partners_login: быстрый поиск по логину партнера
// - idx_partners_domain: быстрый поиск по домену партнера
// - idx_partners_status: фильтрация по статусу (active/blocked)
// - idx_partners_created_at: сортировка по дате создания
// - idx_partners_updated_at: сортировка по дате обновления

func (p *Partner) CreateIndex() string {
	return `
	CREATE INDEX IF NOT EXISTS idx_partners_partner_code ON partners(partner_code);
	CREATE INDEX IF NOT EXISTS idx_partners_login ON partners(login);
	CREATE INDEX IF NOT EXISTS idx_partners_domain ON partners(domain);
	CREATE INDEX IF NOT EXISTS idx_partners_status ON partners(status);
	CREATE INDEX IF NOT EXISTS idx_partners_created_at ON partners(created_at);
	CREATE INDEX IF NOT EXISTS idx_partners_updated_at ON partners(updated_at);
	`
}


