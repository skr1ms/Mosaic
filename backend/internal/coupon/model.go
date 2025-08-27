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
	Code          string     `bun:"code,unique,notnull" json:"code"`
	Size          string     `bun:"size,type:coupon_size,notnull" json:"size"`
	Style         string     `bun:"style,type:coupon_style,notnull" json:"style"`
	Status        string     `bun:"status,type:coupon_status,default:'new'" json:"status"`
	IsPurchased   bool       `bun:"is_purchased,default:false" json:"is_purchased"`
	PurchaseEmail *string    `bun:"purchase_email" json:"purchase_email"`
	PurchasedAt   *time.Time `bun:"purchased_at" json:"purchased_at"`

	UserEmail   *string    `bun:"user_email" json:"user_email"`
	ActivatedAt *time.Time `bun:"activated_at" json:"activated_at"`
	UsedAt      *time.Time `bun:"used_at" json:"used_at"`
	CompletedAt *time.Time `bun:"completed_at" json:"completed_at"`

	ZipURL          *string    `bun:"zip_url" json:"zip_url"`
	SchemaSentEmail *string    `bun:"schema_sent_email" json:"schema_sent_email"`
	SchemaSentAt    *time.Time `bun:"schema_sent_at" json:"schema_sent_at"`
	CreatedAt       time.Time  `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
}

func (c *Coupon) CreateIndex() string {
	return `
	CREATE INDEX IF NOT EXISTS idx_coupons_partner_id ON coupons(partner_id);
	CREATE INDEX IF NOT EXISTS idx_coupons_partner_status ON coupons(partner_id, status);
	CREATE INDEX IF NOT EXISTS idx_coupons_status ON coupons(status);
	CREATE INDEX IF NOT EXISTS idx_coupons_filters ON coupons(size, style, status);
	CREATE INDEX IF NOT EXISTS idx_coupons_purchased ON coupons(is_purchased);
	CREATE INDEX IF NOT EXISTS idx_coupons_created_at ON coupons(created_at);
	CREATE INDEX IF NOT EXISTS idx_coupons_purchased_at ON coupons(purchased_at);
	CREATE INDEX IF NOT EXISTS idx_coupons_activated_at ON coupons(activated_at);
	CREATE INDEX IF NOT EXISTS idx_coupons_used_at ON coupons(used_at);
	CREATE INDEX IF NOT EXISTS idx_coupons_completed_at ON coupons(completed_at);
	`
}

type CouponFilterRequest struct {
	PartnerID *uuid.UUID `json:"partner_id" query:"partner_id"`
	Status    string     `json:"status" query:"status"`
	Size      string     `json:"size" query:"size"`
	Style     string     `json:"style" query:"style"`

	CreatedFrom   *time.Time `json:"created_from" query:"created_from"`
	CreatedTo     *time.Time `json:"created_to" query:"created_to"`
	ActivatedFrom *time.Time `json:"activated_from" query:"activated_from"`
	ActivatedTo   *time.Time `json:"activated_to" query:"activated_to"`

	Search string `json:"search" query:"search"`

	SortBy string `json:"sort_by" query:"sort_by"`
	Order  string `json:"order" query:"order"`

	Page     int `json:"page" query:"page"`
	PageSize int `json:"page_size" query:"page_size"`
}

type CouponInfo struct {
	ID          uuid.UUID  `json:"id"`
	Code        string     `json:"code"`
	PartnerID   uuid.UUID  `json:"partner_id"`
	PartnerName string     `json:"partner_name"`
	Status      string     `json:"status"`
	Size        string     `json:"size"`
	Style       string     `json:"style"`
	CreatedAt   time.Time  `json:"created_at"`
	ActivatedAt *time.Time `json:"activated_at"`
	UsedAt      *time.Time `json:"used_at"`
}

type PartnerCount struct {
	PartnerID   uuid.UUID `json:"partner_id"`
	PartnerCode string    `json:"partner_code"`
	BrandName   string    `json:"brand_name"`
	Count       int64     `json:"count"`
}
