package partner

import (
	"time"

	"github.com/google/uuid"
)

type Partner struct {
	ID              uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	PartnerCode     int16      `gorm:"unique;not null;index" json:"partner_code"` // 4-значный код партнера (0000-9999)
	Login           string     `gorm:"unique;not null;size:255;index" json:"login"`
	Password        string     `gorm:"not null;size:255" json:"password"` // Хешированный пароль
	LastLogin       *time.Time `gorm:"index:idx_partners_last_login" json:"last_login"`
	Domain          string     `gorm:"unique;not null;size:255;index" json:"domain"`
	BrandName       string     `gorm:"not null;size:255;index:idx_partners_search,priority:1" json:"brand_name"`
	LogoURL         string     `gorm:"size:255" json:"logo_url"`
	OzonLink        string     `gorm:"size:255" json:"ozon_link"`
	WildberriesLink string     `gorm:"size:255" json:"wildberries_link"`
	Email           string     `gorm:"not null;size:255;index:idx_partners_search,priority:2" json:"email"`
	Address         string     `gorm:"type:text" json:"address"`
	Phone           string     `gorm:"size:50" json:"phone"`
	Telegram        string     `gorm:"size:255" json:"telegram"`
	Whatsapp        string     `gorm:"size:255" json:"whatsapp"`
	AllowSales      bool       `gorm:"default:true;index:idx_partners_allow_sales" json:"allow_sales"`
	Status          string     `gorm:"type:partner_status;default:'active';index:idx_partners_status" json:"status"`
	CreatedAt       time.Time  `gorm:"index:idx_partners_created_at" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"index:idx_partners_updated_at" json:"updated_at"`
}

// - idx_partners_status: фильтрация по статусу (active/blocked)
// - idx_partners_search: составной индекс для поиска по имени бренда и email
// - idx_partners_allow_sales: фильтрация партнеров которым разрешены продажи
// - idx_partners_last_login: сортировка по последнему входу
// - idx_partners_created_at: сортировка по дате создания
// - idx_partners_updated_at: сортировка по дате обновления
