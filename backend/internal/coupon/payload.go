package coupon

import (
	"time"

	"github.com/google/uuid"
)

type CouponSize string
type CouponStyle string
type CouponStatus string

const (
	Size21x30 CouponSize = "21x30"
	Size30x40 CouponSize = "30x40"
	Size40x40 CouponSize = "40x40"
	Size40x50 CouponSize = "40x50"
	Size40x60 CouponSize = "40x60"
	Size50x70 CouponSize = "50x70"

	StyleGrayscale CouponStyle = "grayscale"
	StyleSkinTones CouponStyle = "skin_tones"
	StylePopArt    CouponStyle = "pop_art"
	StyleMaxColors CouponStyle = "max_colors"

	StatusNew  CouponStatus = "new"
	StatusUsed CouponStatus = "used"
)

type CreateCouponRequest struct {
	Count     int         `json:"count" validate:"required,min=1,max=1000"`
	PartnerID uuid.UUID   `json:"partner_id" validate:"required"`
	Size      CouponSize  `json:"size" validate:"required,oneof=21x30 30x40 40x40 40x50 40x60 50x70"`
	Style     CouponStyle `json:"style" validate:"required,oneof=grayscale skin_tones pop_art max_colors"`
}

type UpdateCouponRequest struct {
	Status           *CouponStatus `json:"status,omitempty" validate:"omitempty,oneof=new used"`
	IsPurchased      *bool         `json:"is_purchased,omitempty"`
	PurchaseEmail    *string       `json:"purchase_email,omitempty" validate:"omitempty,email"`
	PurchasedAt      *time.Time    `json:"purchased_at,omitempty"`
	UsedAt           *time.Time    `json:"used_at,omitempty"`
	OriginalImageURL *string       `json:"original_image_url,omitempty" validate:"omitempty,url"`
	PreviewURL       *string       `json:"preview_url,omitempty" validate:"omitempty,url"`
	SchemaURL        *string       `json:"schema_url,omitempty" validate:"omitempty,url"`
	SchemaSentEmail  *string       `json:"schema_sent_email,omitempty" validate:"omitempty,email"`
	SchemaSentAt     *time.Time    `json:"schema_sent_at,omitempty"`
}

type CouponResponse struct {
	ID               uuid.UUID    `json:"id"`
	Code             string       `json:"code"`
	PartnerID        uuid.UUID    `json:"partner_id"`
	Size             CouponSize   `json:"size"`
	Style            CouponStyle  `json:"style"`
	Status           CouponStatus `json:"status"`
	IsPurchased      bool         `json:"is_purchased"`
	PurchaseEmail    *string      `json:"purchase_email,omitempty"`
	PurchasedAt      *time.Time   `json:"purchased_at,omitempty"`
	UsedAt           *time.Time   `json:"used_at,omitempty"`
	OriginalImageURL *string      `json:"original_image_url,omitempty"`
	PreviewURL       *string      `json:"preview_url,omitempty"`
	SchemaURL        *string      `json:"schema_url,omitempty"`
	SchemaSentEmail  *string      `json:"schema_sent_email,omitempty"`
	SchemaSentAt     *time.Time   `json:"schema_sent_at,omitempty"`
	CreatedAt        time.Time    `json:"created_at"`
}

type BatchDeleteRequest struct {
	CouponIDs []string `json:"coupon_ids" validate:"required,min=1"`
}

type CouponValidationResponse struct {
	Valid   bool       `json:"valid"`
	Message string     `json:"message"`
	Size    *string    `json:"size,omitempty"`
	Style   *string    `json:"style,omitempty"`
	UsedAt  *time.Time `json:"used_at,omitempty"`
}

type CouponStatistics struct {
	Total     int64 `json:"total"`
	New       int64 `json:"new"`
	Used      int64 `json:"used"`
	Purchased int64 `json:"purchased"`
}

type PaginationInfo struct {
	CurrentPage int   `json:"current_page"`
	PerPage     int   `json:"per_page"`
	Total       int64 `json:"total"`
	TotalPages  int64 `json:"total_pages"`
	HasNext     bool  `json:"has_next"`
	HasPrevious bool  `json:"has_previous"`
}

type PaginatedCouponsResponse struct {
	Coupons    []*Coupon      `json:"coupons"`
	Pagination PaginationInfo `json:"pagination"`
}

type ActivateCouponRequest struct {
	OriginalImageURL *string `json:"original_image_url,omitempty" validate:"omitempty,url"`
	PreviewURL       *string `json:"preview_url,omitempty" validate:"omitempty,url"`
	SchemaURL        *string `json:"schema_url,omitempty" validate:"omitempty,url"`
}

type SendSchemaRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type MarkAsPurchasedRequest struct {
	PurchaseEmail string `json:"purchase_email" validate:"required,email"`
}
