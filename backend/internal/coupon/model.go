package coupon

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Coupon struct {
	bun.BaseModel `bun:"table:coupons,alias:c"`

	ID            uuid.UUID  `bun:"id,pk,type:uuid,default:gen_random_uuid()" json:"id"`
	PartnerID     uuid.UUID  `bun:"partner_id,type:uuid,notnull" json:"partner_id"`
	IsBlocked     bool       `bun:"is_blocked,default:false" json:"is_blocked"`
	Code          string     `bun:"code,unique,notnull" json:"code"` // 12-значный код купона в формате XXXX-XXXX-XXXX
	Size          string     `bun:"size,type:coupon_size,notnull" json:"size"`
	Style         string     `bun:"style,type:coupon_style,notnull" json:"style"`
	Status        string     `bun:"status,type:coupon_status,default:'new'" json:"status"`
	IsPurchased   bool       `bun:"is_purchased,default:false" json:"is_purchased"`
	PurchaseEmail *string    `bun:"purchase_email" json:"purchase_email"`
	PurchasedAt   *time.Time `bun:"purchased_at" json:"purchased_at"`

	// Поля для пользовательского API
	UserEmail   *string    `bun:"user_email" json:"user_email"`     // Email пользователя при активации
	ActivatedAt *time.Time `bun:"activated_at" json:"activated_at"` // Время активации купона
	UsedAt      *time.Time `bun:"used_at" json:"used_at"`           // Время использования купона
	CompletedAt *time.Time `bun:"completed_at" json:"completed_at"` // Время завершения обработки

	OriginalImageURL *string    `bun:"original_image_url" json:"original_image_url"`
	PreviewURL       *string    `bun:"preview_url" json:"preview_url"`
	SchemaURL        *string    `bun:"schema_url" json:"schema_url"`
	SchemaSentEmail  *string    `bun:"schema_sent_email" json:"schema_sent_email"`
	SchemaSentAt     *time.Time `bun:"schema_sent_at" json:"schema_sent_at"`
	CreatedAt        time.Time  `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
}

// - idx_coupons_partner_id: быстрый поиск купонов по партнеру
// - idx_coupons_partner_status: составной индекс для фильтрации купонов партнера по статусу
// - idx_coupons_status: фильтрация по статусу (new/used)
// - idx_coupons_filters: составной индекс для комбинированной фильтрации (size, style, status)
// - idx_coupans_purchased: фильтрация купленных купонов
// - idx_coupons_created_at: сортировка по дате создания
// - idx_coupons_purchased_at: аналитика по датам покупки
// - idx_coupons_used_at: аналитика по датам использования

func (c *Coupon) CreateIndex() string {
	return `
	CREATE INDEX IF NOT EXISTS idx_coupons_partner_id ON coupons(partner_id);
	CREATE INDEX IF NOT EXISTS idx_coupons_partner_status ON coupons(partner_id, status);
	CREATE INDEX IF NOT EXISTS idx_coupons_status ON coupons(status);
	CREATE INDEX IF NOT EXISTS idx_coupons_filters ON coupons(size, style, status);
	CREATE INDEX IF NOT EXISTS idx_coupons_purchased ON coupons(is_purchased);
	CREATE INDEX IF NOT EXISTS idx_coupons_created_at ON coupons(created_at);
	CREATE INDEX IF NOT EXISTS idx_coupons_purchased_at ON coupons(purchased_at);
	CREATE INDEX IF NOT EXISTS idx_coupons_used_at ON coupons(used_at);
	`
}
