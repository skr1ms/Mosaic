package partner

import (
	"time"

	"github.com/google/uuid"
)

type LoginRequest struct {
	Login    string `json:"login" validate:"required,secure_login"`
	Password string `json:"password" validate:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type CreatePartnerRequest struct {
	PartnerCode     string `json:"partner_code" validate:"required,len=4,numeric"`
	Login           string `json:"login" validate:"required,secure_login"`
	Password        string `json:"password" validate:"required,secure_password"`
	Domain          string `json:"domain" validate:"required,url"`
	BrandName       string `json:"brand_name" validate:"required"`
	LogoURL         string `json:"logo_url" validate:"required,url"`
	OzonLink        string `json:"ozon_link" validate:"required,url"`
	WildberriesLink string `json:"wildberries_link" validate:"required,url"`
	Email           string `json:"email" validate:"required,email"`
	Address         string `json:"address" validate:"required"`
	Phone           string `json:"phone" validate:"required,international_phone"`
	Telegram        string `json:"telegram" validate:"required,telegram_handle"`
	Whatsapp        string `json:"whatsapp" validate:"required,whatsapp_number"`
	TelegramLink    string `json:"telegram_link" validate:"required,telegram_link"`
	WhatsappLink    string `json:"whatsapp_link" validate:"required,whatsapp_link"`
	AllowSales      bool   `json:"allow_sales"`
	AllowPurchases  bool   `json:"allow_purchases"`
	Status          string `json:"status" validate:"omitempty,oneof=active blocked"`
}

type UpdatePartnerRequest struct {
	Login           *string `json:"login" validate:"omitempty,secure_login"`
	Password        *string `json:"password" validate:"omitempty,secure_password"`
	Domain          *string `json:"domain" validate:"omitempty,url"`
	BrandName       *string `json:"brand_name" validate:"omitempty"`
	LogoURL         *string `json:"logo_url" validate:"omitempty,url"`
	OzonLink        *string `json:"ozon_link" validate:"omitempty,url"`
	WildberriesLink *string `json:"wildberries_link" validate:"omitempty,url"`
	Email           *string `json:"email" validate:"omitempty,email"`
	Address         *string `json:"address" validate:"omitempty"`
	Phone           *string `json:"phone" validate:"omitempty,international_phone"`
	Telegram        *string `json:"telegram" validate:"omitempty,telegram_link"`
	Whatsapp        *string `json:"whatsapp" validate:"omitempty,whatsapp_link"`
	AllowSales      *bool   `json:"allow_sales"`
	Status          *string `json:"status" validate:"omitempty,oneof=active inactive pending"`
}

type UpdatePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,secure_password"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
	//Captcha string `json:"captcha" validate:"required"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,secure_password"`
}

type ExportCouponRequest struct {
	CouponCode    string     `json:"coupon_code"`
	PartnerID     uuid.UUID  `json:"partner_id"`
	PartnerStatus string     `json:"partner_status"`
	CouponStatus  string     `json:"coupon_status"`
	Size          string     `json:"size"`
	Style         string     `json:"style"`
	BrandName     string     `json:"brand_name"`
	Email         string     `json:"email"`
	CreatedAt     time.Time  `json:"created_at"`
	UsedAt        *time.Time `json:"used_at,omitempty"`
}
