package coupon

import (
	"time"

	"github.com/google/uuid"
)

type Coupon struct {
	ID            uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	PartnerID     uuid.UUID  `gorm:"type:uuid;not null" json:"partner_id"`
	IsBlocked     bool       `gorm:"default:false;index:idx_coupons_blocked" json:"is_blocked"`
	Code          string     `gorm:"unique;not null;size:12;index" json:"code"` // 12-значный код купона в формате XXXX-XXXX-XXXX
	Size          string     `gorm:"type:coupon_size;not null;index:idx_coupons_filters,priority:1" json:"size"`
	Style         string     `gorm:"type:coupon_style;not null;index:idx_coupons_filters,priority:2" json:"style"`
	Status        string     `gorm:"type:coupon_status;default:'new';index:idx_coupons_status;index:idx_coupons_partner_status,priority:2;index:idx_coupons_filters,priority:3" json:"status"`
	IsPurchased   bool       `gorm:"default:false;index:idx_coupons_purchased" json:"is_purchased"`
	PurchaseEmail *string    `gorm:"size:255" json:"purchase_email"`
	PurchasedAt   *time.Time `gorm:"index:idx_coupons_purchased_at" json:"purchased_at"`

	// Поля для пользовательского API
	UserEmail   *string    `gorm:"size:255" json:"user_email"`                         // Email пользователя при активации
	ActivatedAt *time.Time `gorm:"index:idx_coupons_activated_at" json:"activated_at"` // Время активации купона
	UsedAt      *time.Time `gorm:"index:idx_coupons_used_at" json:"used_at"`           // Время использования купона
	CompletedAt *time.Time `gorm:"index:idx_coupons_completed_at" json:"completed_at"` // Время завершения обработки

	OriginalImageURL *string    `gorm:"size:255" json:"original_image_url"`
	PreviewURL       *string    `gorm:"size:255" json:"preview_url"`
	SchemaURL        *string    `gorm:"size:255" json:"schema_url"`
	SchemaSentEmail  *string    `gorm:"size:255" json:"schema_sent_email"`
	SchemaSentAt     *time.Time `json:"schema_sent_at"`
	CreatedAt        time.Time  `gorm:"index:idx_coupons_created_at" json:"created_at"`
}

// - idx_coupons_partner_id: быстрый поиск купонов по партнеру
// - idx_coupons_partner_status: составной индекс для фильтрации купонов партнера по статусу
// - idx_coupons_status: фильтрация по статусу (new/used)
// - idx_coupons_filters: составной индекс для комбинированной фильтрации (size, style, status)
// - idx_coupans_purchased: фильтрация купленных купонов
// - idx_coupons_created_at: сортировка по дате создания
// - idx_coupons_purchased_at: аналитика по датам покупки
// - idx_coupons_used_at: аналитика по датам использования
