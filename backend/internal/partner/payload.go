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
	Login           string   `json:"login" validate:"required,secure_login"`
	Password        string   `json:"password" validate:"required,secure_password"`
	Domain          string   `json:"domain" validate:"required,url"`
	BrandName       string   `json:"brand_name" validate:"required"`
	LogoURL         string   `json:"logo_url" validate:"required,url"`
	OzonLink        string   `json:"ozon_link" validate:"required,url"`
	WildberriesLink string   `json:"wildberries_link" validate:"required,url"`
	Email           string   `json:"email" validate:"required,email"`
	Address         string   `json:"address" validate:"required"`
	Phone           string   `json:"phone" validate:"required,international_phone"`
	Telegram        string   `json:"telegram" validate:"required,telegram_link"`
	Whatsapp        string   `json:"whatsapp" validate:"required,whatsapp_link"`
	TelegramLink    string   `json:"telegram_link" validate:"required,telegram_link"`
	WhatsappLink    string   `json:"whatsapp_link" validate:"required,whatsapp_link"`
	BrandColors     []string `json:"brand_colors" validate:"max=3,dive,hex_color"`
	AllowSales      bool     `json:"allow_sales"`
	AllowPurchases  bool     `json:"allow_purchases"`
	Status          string   `json:"status" validate:"omitempty,oneof=active blocked"`
}

type UpdatePartnerRequest struct {
	Login           *string   `json:"login" validate:"omitempty,secure_login"`
	Password        *string   `json:"password" validate:"omitempty,secure_password"`
	Domain          *string   `json:"domain" validate:"omitempty,url"`
	BrandName       *string   `json:"brand_name" validate:"omitempty"`
	LogoURL         *string   `json:"logo_url" validate:"omitempty,url"`
	OzonLink        *string   `json:"ozon_link" validate:"omitempty,url"`
	WildberriesLink *string   `json:"wildberries_link" validate:"omitempty,url"`
	Email           *string   `json:"email" validate:"omitempty,email"`
	Address         *string   `json:"address" validate:"omitempty"`
	Phone           *string   `json:"phone" validate:"omitempty,international_phone"`
	Telegram        *string   `json:"telegram" validate:"omitempty,telegram_link"`
	Whatsapp        *string   `json:"whatsapp" validate:"omitempty,whatsapp_link"`
	BrandColors     *[]string `json:"brand_colors" validate:"omitempty,max=3,dive,hex_color"`
	AllowSales      *bool     `json:"allow_sales"`
	Status          *string   `json:"status" validate:"omitempty,oneof=active inactive pending"`
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
	ActivatedAt   *time.Time `json:"activated_at,omitempty"`
	UsedAt        *time.Time `json:"used_at,omitempty"`
}

type PartnerCouponFilterRequest struct {
	Status string `json:"status" query:"status" validate:"omitempty,oneof=new activated used completed"`
	Size   string `json:"size" query:"size" validate:"omitempty,oneof=21x30 30x30 40x40 50x50 60x60"`
	Style  string `json:"style" query:"style" validate:"omitempty,oneof=classic modern vintage"`

	CreatedFrom   *time.Time `json:"created_from" query:"created_from"`
	CreatedTo     *time.Time `json:"created_to" query:"created_to"`
	ActivatedFrom *time.Time `json:"activated_from" query:"activated_from"`
	ActivatedTo   *time.Time `json:"activated_to" query:"activated_to"`

	Search string `json:"search" query:"search"`

	SortBy string `json:"sort_by" query:"sort_by" validate:"omitempty,oneof=created_at activated_at used_at code status"`
	Order  string `json:"order" query:"order" validate:"omitempty,oneof=asc desc"`

	Page     int `json:"page" query:"page" validate:"min=1"`
	PageSize int `json:"page_size" query:"page_size" validate:"min=1,max=100"`
}

type PartnerCouponInfo struct {
	ID               uuid.UUID  `json:"id"`
	Code             string     `json:"code"`
	Status           string     `json:"status"`
	Size             string     `json:"size"`
	Style            string     `json:"style"`
	CreatedAt        time.Time  `json:"created_at"`
	ActivatedAt      *time.Time `json:"activated_at"`
	UsedAt           *time.Time `json:"used_at"`
	CompletedAt      *time.Time `json:"completed_at"`
	UserEmail        *string    `json:"user_email"`
	HasOriginalImage bool       `json:"has_original_image"`
	HasPreview       bool       `json:"has_preview"`
	HasSchema        bool       `json:"has_schema"`
	IsPurchased      bool       `json:"is_purchased"`
	PurchaseEmail    *string    `json:"purchase_email"`
	PurchasedAt      *time.Time `json:"purchased_at"`
}

type PartnerCouponDetail struct {
	ID            uuid.UUID  `json:"id"`
	Code          string     `json:"code"`
	Status        string     `json:"status"`
	Size          string     `json:"size"`
	Style         string     `json:"style"`
	CreatedAt     time.Time  `json:"created_at"`
	ActivatedAt   *time.Time `json:"activated_at"`
	UsedAt        *time.Time `json:"used_at"`
	CompletedAt   *time.Time `json:"completed_at"`
	UserEmail     *string    `json:"user_email"`
	IsPurchased   bool       `json:"is_purchased"`
	PurchaseEmail *string    `json:"purchase_email"`
	PurchasedAt   *time.Time `json:"purchased_at"`

	ZipURL              *string    `json:"zip_url"`
	SchemaSentEmail     *string    `json:"schema_sent_email"`
	SchemaSentAt        *time.Time `json:"schema_sent_at"`
	CanDownloadMaterial bool       `json:"can_download_material"`
}

type PartnerCouponsResponse struct {
	Coupons []PartnerCouponInfo `json:"coupons"`
	Total   int                 `json:"total"`
	Page    int                 `json:"page"`
	Limit   int                 `json:"limit"`
	Pages   int                 `json:"pages"`
}

type PartnerDashboardResponse struct {
	Statistics     PartnerStatistics   `json:"statistics"`
	RecentActivity []PartnerCouponInfo `json:"recent_activity"`
	StatusCounts   map[string]int64    `json:"status_counts"`
	SizeCounts     map[string]int64    `json:"size_counts"`
	StyleCounts    map[string]int64    `json:"style_counts"`
}

type PartnerStatistics struct {
	TotalCoupons     int64      `json:"total_coupons"`
	ActivatedCoupons int64      `json:"activated_coupons"`
	UsedCoupons      int64      `json:"used_coupons"`
	CompletedCoupons int64      `json:"completed_coupons"`
	PurchasedCoupons int64      `json:"purchased_coupons"`
	LastActivity     *time.Time `json:"last_activity"`
}

type PartnerSalesStatistics struct {
	TotalSales      int64            `json:"total_sales"`
	SalesThisMonth  int64            `json:"sales_this_month"`
	SalesThisWeek   int64            `json:"sales_this_week"`
	SalesTimeSeries []SalesTimePoint `json:"sales_time_series"`
	TopSizes        []SizeStatistic  `json:"top_sizes"`
	TopStyles       []StyleStatistic `json:"top_styles"`
}

type SalesTimePoint struct {
	Date  time.Time `json:"date"`
	Sales int64     `json:"sales"`
}

type SizeStatistic struct {
	Size  string `json:"size"`
	Count int64  `json:"count"`
}

type StyleStatistic struct {
	Style string `json:"style"`
	Count int64  `json:"count"`
}

type PartnerUsageStatistics struct {
	UsageThisMonth   int64            `json:"usage_this_month"`
	UsageThisWeek    int64            `json:"usage_this_week"`
	UsageTimeSeries  []UsageTimePoint `json:"usage_time_series"`
	ConversionRate   float64          `json:"conversion_rate"`
	CompletionRate   float64          `json:"completion_rate"`
	AverageTimeToUse *int64           `json:"average_time_to_use"`
}

type UsageTimePoint struct {
	Date  time.Time `json:"date"`
	Usage int64     `json:"usage"`
}
