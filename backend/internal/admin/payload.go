package admin

import (
	"time"

	"github.com/google/uuid"
	"github.com/skr1ms/mosaic/internal/coupon"
)

type CreateAdminRequest struct {
	Login    string `json:"login" validate:"required,secure_login"`
	Password string `json:"password" validate:"required,secure_password"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,secure_password"`
}

type PartnerDetailResponse struct {
	ID              uuid.UUID  `json:"id"`
	Login           string     `json:"login"`
	BrandName       string     `json:"brand_name"`
	Domain          string     `json:"domain"`
	Email           string     `json:"email"`
	Phone           string     `json:"phone"`
	TelegramLink    string     `json:"telegram_link"`
	WhatsAppLink    string     `json:"whatsapp_link"`
	WildberriesLink string     `json:"wildberries_link"`
	OzonLink        string     `json:"ozon_link"`
	Status          string     `json:"status"`
	CreatedAt       time.Time  `json:"created_at"`
	LastLogin       *time.Time `json:"last_login"`

	// Статистика
	TotalCoupons     int        `json:"total_coupons"`
	ActivatedCoupons int        `json:"activated_coupons"`
	UnusedCoupons    int        `json:"unused_coupons"`
	LastActivity     *time.Time `json:"last_activity"`

	// История изменений профиля
	ProfileChanges []ProfileChange `json:"profile_changes"`
}

type ProfileChange struct {
	ID        uuid.UUID `json:"id"`
	PartnerID uuid.UUID `json:"partner_id"`
	Field     string    `json:"field"`
	OldValue  string    `json:"old_value"`
	NewValue  string    `json:"new_value"`
	ChangedBy string    `json:"changed_by"` // admin login
	ChangedAt time.Time `json:"changed_at"`
	Reason    string    `json:"reason"`
}

type CouponFilterRequest struct {
	// Базовые фильтры
	PartnerID *uuid.UUID `json:"partner_id" query:"partner_id"`
	Status    string     `json:"status" query:"status"`
	Size      string     `json:"size" query:"size"`
	Style     string     `json:"style" query:"style"`

	// Фильтры по датам
	CreatedFrom   *time.Time `json:"created_from" query:"created_from"`
	CreatedTo     *time.Time `json:"created_to" query:"created_to"`
	ActivatedFrom *time.Time `json:"activated_from" query:"activated_from"`
	ActivatedTo   *time.Time `json:"activated_to" query:"activated_to"`

	// Поиск
	Search string `json:"search" query:"search"` // поиск по коду купона

	// Сортировка
	SortBy string `json:"sort_by" query:"sort_by"` // created_at, activated_at, code
	Order  string `json:"order" query:"order"`     // asc, desc

	// Пагинация
	Page     int `json:"page" query:"page"`
	PageSize int `json:"page_size" query:"page_size"`
}

// UserFilter содержит сохраненные пользовательские фильтры
type UserFilter struct {
	ID          uuid.UUID `bun:"id,pk,type:uuid,default:gen_random_uuid()" json:"id"`
	AdminID     uuid.UUID `bun:"admin_id,notnull,type:uuid" json:"admin_id"`
	Name        string    `bun:"name,notnull" json:"name"`
	Description string    `bun:"description" json:"description"`
	FilterType  string    `bun:"filter_type,notnull" json:"filter_type"`
	FilterData  string    `bun:"filter_data,notnull" json:"filter_data"`
	IsDefault   bool      `bun:"is_default,default:false" json:"is_default"`
	CreatedAt   time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt   time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`
}

type CouponFilterResponse struct {
	Coupons    []*coupon.CouponInfo `json:"coupons"`
	Total      int                  `json:"total"`
	Page       int                  `json:"page"`
	PageSize   int                  `json:"page_size"`
	TotalPages int                  `json:"total_pages"`
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

// BatchDownloadMaterialsRequest запрос для массового скачивания материалов
type BatchDownloadMaterialsRequest struct {
	CouponIDs []uuid.UUID `json:"coupon_ids" validate:"required,min=1,max=100"`
}
