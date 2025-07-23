package coupon

import (
	"time"

	"github.com/google/uuid"
)

type Coupon struct {
	ID               uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Code             string     `gorm:"unique;not null;size:12" json:"code"` // 12-значный код купона в формате XXXX-XXXX-XXXX
	PartnerID        uuid.UUID  `gorm:"type:uuid;not null" json:"partner_id"`
	Size             string     `gorm:"type:enum('21x30','30x40','40x40','40x50','40x60','50x70');not null" json:"size"`
	Style            string     `gorm:"type:enum('grayscale','skin_tones','pop_art','max_colors');not null" json:"style"`
	Status           string     `gorm:"type:enum('new','used');default:'new'" json:"status"`
	IsPurchased      bool       `gorm:"default:false" json:"is_purchased"`
	PurchaseEmail    *string    `gorm:"size:255" json:"purchase_email"`
	PurchasedAt      *time.Time `json:"purchased_at"`
	UsedAt           *time.Time `json:"used_at"`
	OriginalImageURL *string    `gorm:"size:255" json:"original_image_url"`
	PreviewURL       *string    `gorm:"size:255" json:"preview_url"`
	SchemaURL        *string    `gorm:"size:255" json:"schema_url"`
	SchemaSentEmail  *string    `gorm:"size:255" json:"schema_sent_email"`
	SchemaSentAt     *time.Time `json:"schema_sent_at"`
	CreatedAt        time.Time  `json:"created_at"`
}
