package payment

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// Статусы заказов
const (
	OrderStatusCreated   = "created"
	OrderStatusPending   = "pending"
	OrderStatusPaid      = "paid"
	OrderStatusFailed    = "failed"
	OrderStatusCancelled = "cancelled"
)

// Фиксированная цена для всех размеров
const (
	FixedPriceRub = 100.0 // 100 рублей для всех размеров
)

// Размеры купонов согласно ТЗ
const (
	Size21x30 = "21x30"
	Size30x40 = "30x40"
	Size40x40 = "40x40"
	Size40x50 = "40x50"
	Size40x60 = "40x60"
	Size50x70 = "50x70"
)

// Стили обработки согласно ТЗ
const (
	StyleGrayscale = "grayscale"  // оттенки серого
	StyleSkinTone = "skin_tones" // оттенки телесного
	StylePopArt    = "pop_art"    // поп-арт
	StyleMaxColors = "max_colors" // максимум цветов
)

// Модель заказа для покупки купонов онлайн
type Order struct {
	bun.BaseModel `bun:"table:orders"`

	ID              uuid.UUID  `bun:"id,pk,type:uuid,default:gen_random_uuid()" json:"id"`
	OrderNumber     string     `bun:"order_number,notnull,unique" json:"order_number"`
	AlfaBankOrderID *string    `bun:"alfabank_order_id" json:"alfabank_order_id"`
	PartnerID       *uuid.UUID `bun:"partner_id,type:uuid" json:"partner_id"` // Партнер через чей сайт была покупка

	// Параметры купона
	Size  string `bun:"size,notnull" json:"size"`   // Размер мозаики
	Style string `bun:"style,notnull" json:"style"` // Стиль обработки

	// Данные покупателя
	UserEmail string `bun:"user_email,notnull" json:"user_email"`

	// Финансовые данные
	Amount   int64  `bun:"amount,notnull" json:"amount"` // в копейках
	Currency string `bun:"currency,notnull,default:'RUB'" json:"currency"`

	// Статус и URLs
	Status      string  `bun:"status,notnull,default:'created'" json:"status"`
	PaymentURL  *string `bun:"payment_url" json:"payment_url"`
	ReturnURL   string  `bun:"return_url,notnull" json:"return_url"`
	FailURL     *string `bun:"fail_url" json:"fail_url"`
	Description string  `bun:"description" json:"description"`

	// Созданный купон после успешной оплаты
	CouponID *uuid.UUID `bun:"coupon_id,type:uuid" json:"coupon_id"`

	CreatedAt time.Time  `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt time.Time  `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`
	PaidAt    *time.Time `bun:"paid_at" json:"paid_at"`
}

// - idx_orders_order_number: быстрый поиск по номеру заказа
// - idx_orders_alfabank_order_id: быстрый поиск по номеру заказа в Альфа-Банке
// - idx_orders_partner_id: быстрый поиск по ID партнера
// - idx_orders_status: быстрый поиск по статусу заказа
// - idx_orders_user_email: быстрый поиск по email покупателя
// - idx_orders_created_at: сортировка по дате создания

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
