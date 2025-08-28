package public

import (
	"github.com/google/uuid"
)

type SendEmailRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type PurchaseCouponRequest struct {
	Size         string  `json:"size" validate:"required,oneof=21x30 30x40 40x40 40x50 40x60 50x70"`
	Style        string  `json:"style" validate:"required,oneof=grayscale skin_tones pop_art max_colors"`
	Email        string  `json:"email" validate:"required,email"`
	PaymentToken string  `json:"payment_token" validate:"required"`
	Amount       float64 `json:"amount" validate:"required"`
}

type CouponInfoResponse struct {
	ID     uuid.UUID `json:"id"`
	Code   string    `json:"code"`
	Size   string    `json:"size"`
	Style  string    `json:"style"`
	Status string    `json:"status"`
	Valid  bool      `json:"valid"`
}

type ImageUploadResponse struct {
	Message  string    `json:"message"`
	ImageID  uuid.UUID `json:"image_id"`
	NextStep string    `json:"next_step"`
}

type ProcessingStatusResponse struct {
	ID          uuid.UUID `json:"id"`
	Status      string    `json:"status"`
	Progress    int       `json:"progress"`
	Error       *string   `json:"error,omitempty"`
	StartedAt   *string   `json:"started_at,omitempty"`
	CompletedAt *string   `json:"completed_at,omitempty"`
}

type PartnerInfoResponse struct {
	BrandName       string `json:"brand_name"`
	Domain          string `json:"domain"`
	LogoURL         string `json:"logo_url"`
	OzonLink        string `json:"ozon_link"`
	WildberriesLink string `json:"wildberries_link"`
	Email           string `json:"email"`
	Phone           string `json:"phone"`
	Address         string `json:"address"`
	TelegramLink    string `json:"telegram_link"`
	WhatsappLink    string `json:"whatsapp_link"`
	AllowPurchases  bool   `json:"allow_purchases"`
}

type SizeInfo struct {
	Size  string `json:"size"`
	Title string `json:"title"`
	Price int    `json:"price"`
}

type StyleInfo struct {
	Style       string `json:"style"`
	Title       string `json:"title"`
	Description string `json:"description"`
}
