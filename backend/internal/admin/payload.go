package admin

import (
	"time"

	"github.com/google/uuid"
)

type CreateAdminRequest struct {
	Login    string `json:"login" validate:"required,secure_login"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,secure_password"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,secure_password"`
}

type ChangeEmailRequest struct {
	Password string `json:"password" validate:"required"`
	NewEmail string `json:"new_email" validate:"required,email"`
}

type UpdateAdminPasswordRequest struct {
	Password string `json:"password" validate:"required,secure_password"`
}

type UpdateAdminEmailRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type PartnerDetailResponse struct {
	ID               uuid.UUID       `json:"id"`
	Login            string          `json:"login"`
	BrandName        string          `json:"brand_name"`
	Domain           string          `json:"domain"`
	Email            string          `json:"email"`
	Phone            string          `json:"phone"`
	TelegramLink     string          `json:"telegram_link"`
	WhatsappLink     string          `json:"whatsapp_link"`
	WildberriesLink  string          `json:"wildberries_link"`
	OzonLink         string          `json:"ozon_link"`
	BrandColors      []string        `json:"brand_colors"`
	Status           string          `json:"status"`
	CreatedAt        time.Time       `json:"created_at"`
	LastLogin        *time.Time      `json:"last_login"`
	TotalCoupons     int             `json:"total_coupons"`
	ActivatedCoupons int             `json:"activated_coupons"`
	UnusedCoupons    int             `json:"unused_coupons"`
	LastActivity     *time.Time      `json:"last_activity"`
	ProfileChanges   []ProfileChange `json:"profile_changes"`
}

type ProfileChange struct {
	ID        uuid.UUID `json:"id"`
	PartnerID uuid.UUID `json:"partner_id"`
	Field     string    `json:"field"`
	OldValue  string    `json:"old_value"`
	NewValue  string    `json:"new_value"`
	ChangedBy string    `json:"changed_by"`
	ChangedAt time.Time `json:"changed_at"`
	Reason    string    `json:"reason"`
}

type BatchDownloadMaterialsRequest struct {
	CouponIDs []uuid.UUID `json:"coupon_ids" validate:"required,dive,required"`
}
