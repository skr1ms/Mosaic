package partner

import (
	"time"

	"github.com/google/uuid"
)

type Partner struct {
	ID              uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	PartnerCode     int16      `gorm:"unique;not null" json:"partner_code"` // 4-значный код партнера (0000-9999)
	Login           string     `gorm:"unique;not null;size:255" json:"login"`
	Password        string     `gorm:"not null;size:255" json:"password"` // Хешированный пароль
	LastLogin       *time.Time `json:"last_login"`
	Domain          string     `gorm:"unique;not null;size:255" json:"domain"`
	BrandName       string     `gorm:"not null;size:255" json:"brand_name"`
	LogoURL         string     `gorm:"size:255" json:"logo_url"`
	OzonLink        string     `gorm:"size:255" json:"ozon_link"`
	WildberriesLink string     `gorm:"size:255" json:"wildberries_link"`
	Email           string     `gorm:"not null;size:255" json:"email"`
	Address         string     `gorm:"type:text" json:"address"`
	Phone           string     `gorm:"size:50" json:"phone"`
	Telegram        string     `gorm:"size:255" json:"telegram"`
	Whatsapp        string     `gorm:"size:255" json:"whatsapp"`
	AllowSales      bool       `gorm:"default:true" json:"allow_sales"`
	Status          string     `gorm:"type:enum('active','blocked');default:'active'" json:"status"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}
