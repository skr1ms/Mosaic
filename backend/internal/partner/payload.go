package partner

type LoginRequest struct {
	Login    string `json:"login" validate:"required,secure_login"`
	Password string `json:"password" validate:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type CreatePartnerRequest struct {
	Login           string `json:"login" validate:"required,secure_login"`
	Password        string `json:"password" validate:"required,secure_password"`
	PartnerCode     int16  `json:"partner_code" validate:"required"`
	Domain          string `json:"domain" validate:"required,url"`
	BrandName       string `json:"brand_name" validate:"required"`
	LogoURL         string `json:"logo_url" validate:"omitempty,url"`
	OzonLink        string `json:"ozon_link" validate:"omitempty,url"`
	WildberriesLink string `json:"wildberries_link" validate:"omitempty,url"`
	Email           string `json:"email" validate:"omitempty,email"`
	Address         string `json:"address" validate:"omitempty"`
	Phone           string `json:"phone" validate:"required,russian_phone"`
	Telegram        string `json:"telegram" validate:"omitempty,telegram_link"`
	Whatsapp        string `json:"whatsapp" validate:"omitempty,whatsapp_link"`
	AllowSales      bool   `json:"allow_sales"`
	Status          string `json:"status" validate:"omitempty,oneof=active inactive pending"`
}

type UpdatePartnerRequest struct {
	Login           *string `json:"login" validate:"omitempty,secure_login"`
	Password        *string `json:"password" validate:"omitempty,secure_password"`
	PartnerCode     *int16  `json:"partner_code" validate:"omitempty"`
	Domain          *string `json:"domain" validate:"omitempty,url"`
	BrandName       *string `json:"brand_name" validate:"omitempty"`
	LogoURL         *string `json:"logo_url" validate:"omitempty,url"`
	OzonLink        *string `json:"ozon_link" validate:"omitempty,url"`
	WildberriesLink *string `json:"wildberries_link" validate:"omitempty,url"`
	Email           *string `json:"email" validate:"omitempty,email"`
	Address         *string `json:"address" validate:"omitempty"`
	Phone           *string `json:"phone" validate:"omitempty,russian_phone"`
	Telegram        *string `json:"telegram" validate:"omitempty,telegram_link"`
	Whatsapp        *string `json:"whatsapp" validate:"omitempty,whatsapp_link"`
	AllowSales      *bool   `json:"allow_sales"`
	Status          *string `json:"status" validate:"omitempty,oneof=active inactive pending"`
}
