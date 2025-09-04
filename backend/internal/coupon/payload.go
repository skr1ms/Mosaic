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

	StatusNew       CouponStatus = "new"
	StatusActivated CouponStatus = "activated"
	StatusUsed      CouponStatus = "used"
	StatusCompleted CouponStatus = "completed"
)

type CreateCouponRequest struct {
	Count     int         `json:"count" validate:"required,min=1,max=1000"`
	PartnerID uuid.UUID   `json:"partner_id" validate:"required"`
	Size      CouponSize  `json:"size" validate:"required,oneof=21x30 30x40 40x40 40x50 40x60 50x70"`
	Style     CouponStyle `json:"style" validate:"required,oneof=grayscale skin_tones pop_art max_colors"`
}

type UpdateCouponRequest struct {
	Status        *CouponStatus `json:"status,omitempty" validate:"omitempty,oneof=new activated used completed"`
	IsPurchased   *bool         `json:"is_purchased,omitempty"`
	PurchaseEmail *string       `json:"purchase_email,omitempty" validate:"omitempty,email"`
	PurchasedAt   *time.Time    `json:"purchased_at,omitempty"`
	UsedAt        *time.Time    `json:"used_at,omitempty"`

	ZipURL          *string    `json:"zip_url,omitempty" validate:"omitempty,url"`
	SchemaSentEmail *string    `json:"schema_sent_email,omitempty" validate:"omitempty,email"`
	SchemaSentAt    *time.Time `json:"schema_sent_at,omitempty"`
}

type CouponResponse struct {
	ID            uuid.UUID    `json:"id"`
	Code          string       `json:"code"`
	PartnerID     uuid.UUID    `json:"partner_id"`
	Size          CouponSize   `json:"size"`
	Style         CouponStyle  `json:"style"`
	Status        CouponStatus `json:"status"`
	IsPurchased   bool         `json:"is_purchased"`
	PurchaseEmail *string      `json:"purchase_email,omitempty"`
	PurchasedAt   *time.Time   `json:"purchased_at,omitempty"`
	UsedAt        *time.Time   `json:"used_at,omitempty"`

	ZipURL          *string    `json:"zip_url,omitempty"`
	SchemaSentEmail *string    `json:"schema_sent_email,omitempty"`
	SchemaSentAt    *time.Time `json:"schema_sent_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
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

	PartnerID        *uuid.UUID `json:"partner_id,omitempty"`
	PartnerCode      *string    `json:"partner_code,omitempty"`
	PartnerDomain    *string    `json:"partner_domain,omitempty"`
	PartnerBrandName *string    `json:"partner_brand_name,omitempty"`
	IsCorrectDomain  bool       `json:"is_correct_domain"`
	CorrectDomain    *string    `json:"correct_domain,omitempty"`
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
	ZipURL            *string `json:"zip_url,omitempty" validate:"omitempty,url"`
	PreviewImageURL   *string `json:"preview_image_url,omitempty" validate:"omitempty,url"`
	SelectedPreviewID *string `json:"selected_preview_id,omitempty"`
	StonesCount       *int    `json:"stones_count,omitempty"`
	FinalSchemaURL    *string `json:"final_schema_url,omitempty" validate:"omitempty,url"`
	PageCount         *int    `json:"page_count,omitempty"`
	UserEmail         *string `json:"user_email,omitempty" validate:"omitempty,email"`
}

type SendSchemaRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type MarkAsPurchasedRequest struct {
	PurchaseEmail string `json:"purchase_email" validate:"required,email"`
}

type BatchResetRequest struct {
	CouponIDs []string `json:"coupon_ids" validate:"required,min=1,max=1000"`
}

type BatchResetResponse struct {
	Success      []string `json:"success"`
	Failed       []string `json:"failed"`
	SuccessCount int      `json:"success_count"`
	FailedCount  int      `json:"failed_count"`
	Errors       []string `json:"errors,omitempty"`
}

type BatchDeleteConfirmRequest struct {
	CouponIDs       []string `json:"coupon_ids" validate:"required,min=1,max=1000"`
	ConfirmationKey string   `json:"confirmation_key" validate:"required"`
	AdminComment    string   `json:"admin_comment,omitempty"`
}

type BatchDeletePreviewResponse struct {
	TotalCount      int                    `json:"total_count"`
	UsedCount       int                    `json:"used_count"`
	NewCount        int                    `json:"new_count"`
	Coupons         []*CouponDeletePreview `json:"coupons"`
	ConfirmationKey string                 `json:"confirmation_key"`
	ExpiresAt       time.Time              `json:"expires_at"`
}

type CouponDeletePreview struct {
	ID          string     `json:"id"`
	Code        string     `json:"code"`
	Status      string     `json:"status"`
	PartnerName string     `json:"partner_name"`
	CreatedAt   time.Time  `json:"created_at"`
	UsedAt      *time.Time `json:"used_at,omitempty"`
}

type BatchDeleteResponse struct {
	DeletedCount int      `json:"deleted_count"`
	FailedCount  int      `json:"failed_count"`
	Deleted      []string `json:"deleted"`
	Failed       []string `json:"failed"`
	Errors       []string `json:"errors,omitempty"`
}

type ExportFormatType string

const (
	ExportFormatCodes    ExportFormatType = "codes"    // only codes
	ExportFormatBasic    ExportFormatType = "basic"    // basic information
	ExportFormatFull     ExportFormatType = "full"     // full information (all coupons)
	ExportFormatAdmin    ExportFormatType = "admin"    // admin format (new coupons)
	ExportFormatPartner  ExportFormatType = "partner"  // with partner information
	ExportFormatActivity ExportFormatType = "activity" // with user activity
)

type ExportOptionsRequest struct {
	Format       ExportFormatType `json:"format" validate:"required,oneof=codes basic full admin partner activity"`
	PartnerID    *string          `json:"partner_id,omitempty"`
	PartnerCodes []string         `json:"partner_codes,omitempty"`
	Status       string           `json:"status,omitempty"`
	Size         string           `json:"size,omitempty"`
	Style        string           `json:"style,omitempty"`

	CreatedFrom   *time.Time `json:"created_from,omitempty"`
	CreatedTo     *time.Time `json:"created_to,omitempty"`
	ActivatedFrom *time.Time `json:"activated_from,omitempty"`
	ActivatedTo   *time.Time `json:"activated_to,omitempty"`

	FileFormat    string `json:"file_format" validate:"oneof=txt csv xlsx"`
	Delimiter     string `json:"delimiter,omitempty"`
	IncludeHeader bool   `json:"include_header"`
}

type BasicExportRow struct {
	Code      string    `json:"code"`
	Status    string    `json:"status"`
	Size      string    `json:"size"`
	Style     string    `json:"style"`
	CreatedAt time.Time `json:"created_at"`
}

type AdminExportRow struct {
	Code          string    `json:"code"`
	PartnerID     string    `json:"partner_id"`
	PartnerStatus string    `json:"partner_status"`
	CouponStatus  string    `json:"coupon_status"`
	Size          string    `json:"size"`
	Style         string    `json:"style"`
	BrandName     string    `json:"brand_name"`
	Email         string    `json:"email"`
	CreatedAt     time.Time `json:"created_at"`
}

type PartnerExportRow struct {
	Code          string    `json:"code"`
	PartnerStatus string    `json:"partner_status"`
	CouponStatus  string    `json:"coupon_status"`
	Size          string    `json:"size"`
	Style         string    `json:"style"`
	CreatedAt     time.Time `json:"created_at"`
}

type FullExportRow struct {
	Code          string     `json:"code"`
	PartnerID     string     `json:"partner_id"`
	PartnerStatus string     `json:"partner_status"`
	CouponStatus  string     `json:"coupon_status"`
	Size          string     `json:"size"`
	Style         string     `json:"style"`
	BrandName     string     `json:"brand_name"`
	Email         string     `json:"email"`
	CreatedAt     time.Time  `json:"created_at"`
	UsedAt        *time.Time `json:"used_at"`
}

type ActivityExportRow struct {
	Code            string     `json:"code"`
	PartnerName     string     `json:"partner_name"`
	Status          string     `json:"status"`
	Size            string     `json:"size"`
	Style           string     `json:"style"`
	CreatedAt       time.Time  `json:"created_at"`
	ActivatedAt     *time.Time `json:"activated_at"`
	UsedAt          *time.Time `json:"used_at"`
	CompletedAt     *time.Time `json:"completed_at"`
	UserEmail       *string    `json:"user_email"`
	PurchaseEmail   *string    `json:"purchase_email"`
	PurchasedAt     *time.Time `json:"purchased_at"`
	SchemaSentEmail *string    `json:"schema_sent_email"`
	SchemaSentAt    *time.Time `json:"schema_sent_at"`
}
