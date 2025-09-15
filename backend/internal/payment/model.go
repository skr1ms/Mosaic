package payment

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

const (
	OrderStatusCreated   = "created"
	OrderStatusPending   = "pending"
	OrderStatusPaid      = "paid"
	OrderStatusFailed    = "failed"
	OrderStatusCancelled = "cancelled"
)

const (
	FixedPriceRub = 100.0
)

const (
	Size21x30 = "21x30"
	Size30x40 = "30x40"
	Size40x40 = "40x40"
	Size40x50 = "40x50"
	Size40x60 = "40x60"
	Size50x70 = "50x70"
)

const (
	StyleGrayscale = "grayscale"
	StyleSkinTone  = "skin_tones"
	StylePopArt    = "pop_art"
	StyleMaxColors = "max_colors"
)

type Order struct {
	bun.BaseModel `bun:"table:orders"`

	ID              uuid.UUID  `bun:"id,pk,type:uuid,default:gen_random_uuid()" json:"id"`
	OrderNumber     string     `bun:"order_number,notnull,unique" json:"order_number"`
	AlfaBankOrderID *string    `bun:"alfabank_order_id" json:"alfabank_order_id"`
	PartnerID       *uuid.UUID `bun:"partner_id,type:uuid" json:"partner_id"`

	Size  string `bun:"size,notnull" json:"size"`
	Style string `bun:"style,notnull" json:"style"`

	UserEmail string `bun:"user_email,notnull" json:"user_email"`

	Amount   int64  `bun:"amount,notnull" json:"amount"`
	Currency string `bun:"currency,notnull,default:'RUB'" json:"currency"`

	Status      string  `bun:"status,notnull,default:'created'" json:"status"`
	PaymentURL  *string `bun:"payment_url" json:"payment_url"`
	ReturnURL   string  `bun:"return_url,notnull" json:"return_url"`
	FailURL     *string `bun:"fail_url" json:"fail_url"`
	Description string  `bun:"description" json:"description"`

	CouponID *uuid.UUID `bun:"coupon_id,type:uuid" json:"coupon_id"`

	CreatedAt time.Time  `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt time.Time  `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`
	PaidAt    *time.Time `bun:"paid_at" json:"paid_at"`
}

func (o *Order) CreateIndex() string {
	return `
	CREATE INDEX IF NOT EXISTS idx_orders_order_number ON orders(order_number);
	CREATE INDEX IF NOT EXISTS idx_orders_alfabank_order_id ON orders(alfabank_order_id);
	CREATE INDEX IF NOT EXISTS idx_orders_partner_id ON orders(partner_id);
	CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
	CREATE INDEX IF NOT EXISTS idx_orders_user_email ON orders(user_email);
	CREATE INDEX IF NOT EXISTS idx_orders_created_at ON orders(created_at);
	`
}
