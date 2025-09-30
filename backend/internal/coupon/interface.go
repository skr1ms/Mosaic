package coupon

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type PartnerRepositoryInterface interface {
	GetByID(ctx context.Context, id uuid.UUID) (*Partner, error)
	GetByPartnerCode(ctx context.Context, code string) (*Partner, error)
}

type Partner struct {
	ID          uuid.UUID `json:"id"`
	PartnerCode string    `json:"partner_code"`
	Domain      string    `json:"domain"`
	BrandName   string    `json:"brand_name"`
}

type RedisClientInterface interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
}

type CouponRepositoryInterface interface {
	Create(ctx context.Context, coupon *Coupon) error
	CreateBatch(ctx context.Context, coupons []*Coupon) error
	GetByID(ctx context.Context, id uuid.UUID) (*Coupon, error)
	GetByCode(ctx context.Context, code string) (*Coupon, error)
	GetByPartnerID(ctx context.Context, partnerID uuid.UUID) ([]*Coupon, error)
	GetAll(ctx context.Context) ([]*Coupon, error)
	Update(ctx context.Context, coupon *Coupon) error
	Delete(ctx context.Context, id uuid.UUID) error
	BatchDelete(ctx context.Context, ids []uuid.UUID) (int64, error)

	Search(ctx context.Context, code, status, size, style string, partnerID *uuid.UUID) ([]*Coupon, error)
	SearchWithPagination(
		ctx context.Context,
		code, status, size, style string,
		partnerID *uuid.UUID,
		page, limit int,
		createdFrom, createdTo, usedFrom, usedTo *time.Time,
		sortBy, sortDir string,
	) ([]*Coupon, int, error)
	GetFiltered(ctx context.Context, filters map[string]any) ([]*Coupon, error)

	GetPartnerCouponsWithFilter(ctx context.Context, partnerID uuid.UUID, filters map[string]any, page, limit int, sortBy, order string) ([]*Coupon, int, error)
	GetPartnerCouponByCode(ctx context.Context, partnerID uuid.UUID, code string) (*Coupon, error)
	GetPartnerCouponDetail(ctx context.Context, partnerID uuid.UUID, couponID uuid.UUID) (*Coupon, error)
	GetPartnerRecentActivity(ctx context.Context, partnerID uuid.UUID, limit int) ([]*Coupon, error)

	UpdateStatusByPartnerID(ctx context.Context, partnerID uuid.UUID, status bool) error
	ActivateCoupon(ctx context.Context, id uuid.UUID, req ActivateCouponRequest) error
	SendSchema(ctx context.Context, id uuid.UUID, email string) error
	MarkAsPurchased(ctx context.Context, id uuid.UUID, purchaseEmail string) error
	ResetCoupon(ctx context.Context, id uuid.UUID) error
	Reset(ctx context.Context, id uuid.UUID) error
	BatchReset(ctx context.Context, ids []uuid.UUID) ([]uuid.UUID, []uuid.UUID, error)

	CountByPartnerID(ctx context.Context, partnerID uuid.UUID) (int, error)
	CountActivatedByPartnerID(ctx context.Context, partnerID uuid.UUID) (int, error)
	CountPurchasedByPartnerID(ctx context.Context, partnerID uuid.UUID) (int, error)
	CountTotal(ctx context.Context) (int64, error)
	CountByStatus(ctx context.Context, status string) (int64, error)
	CountByPartner(ctx context.Context, partnerID uuid.UUID) (int64, error)
	CountByPartnerAndStatus(ctx context.Context, partnerID uuid.UUID, status string) (int64, error)
	CountActivated(ctx context.Context) (int64, error)
	CountActivatedByPartner(ctx context.Context, partnerID uuid.UUID) (int64, error)
	CountBrandedPurchasesByPartner(ctx context.Context, partnerID uuid.UUID) (int64, error)
	CountActivatedInTimeRange(ctx context.Context, from, to time.Time) (int64, error)

	GetStatistics(ctx context.Context, partnerID *uuid.UUID) (map[string]int64, error)
	GetPartnerStatistics(ctx context.Context, partnerID uuid.UUID) (map[string]any, error)
	GetPartnerSalesStatistics(ctx context.Context, partnerID uuid.UUID) (map[string]any, error)
	GetPartnerUsageStatistics(ctx context.Context, partnerID uuid.UUID) (map[string]any, error)
	GetStatusCounts(ctx context.Context, partnerID *uuid.UUID) (map[string]int64, error)
	GetExtendedStatusCounts(ctx context.Context, partnerID *uuid.UUID) (map[string]int64, error)
	GetSizeCounts(ctx context.Context, partnerID *uuid.UUID) (map[string]int64, error)
	GetStyleCounts(ctx context.Context, partnerID *uuid.UUID) (map[string]int64, error)

	GetTopActivatedByPartner(ctx context.Context, limit int) ([]PartnerCount, error)
	GetTopPurchasedByPartner(ctx context.Context, limit int) ([]PartnerCount, error)

	GetLastActivityByPartner(ctx context.Context, partnerID uuid.UUID) (*time.Time, error)
	GetTimeSeriesData(ctx context.Context, from, to time.Time, period string, partnerID *uuid.UUID) ([]map[string]any, error)
	GetRecentActivated(ctx context.Context, limit int) ([]*Coupon, error)
	FindAvailableCoupon(ctx context.Context, size, style string, partnerID *uuid.UUID) (*Coupon, error)

	CodeExists(ctx context.Context, code string) (bool, error)
	HealthCheck(ctx context.Context) error

	GetCouponsWithAdvancedFilter(ctx context.Context, filter CouponFilterRequest) ([]*CouponInfo, int, error)
	GetCouponsForDeletion(ctx context.Context, ids []uuid.UUID) ([]*CouponDeletePreview, error)
	GetCouponsForExport(ctx context.Context, options ExportOptionsRequest) (any, error)
}

type S3Interface interface {
	DeleteFile(ctx context.Context, objectKey string) error
}

type CouponServiceInterface interface {
	GetCouponByID(id uuid.UUID) (*Coupon, error)
	GetCouponByCode(code string) (*Coupon, error)
	ActivateCoupon(id uuid.UUID, req ActivateCouponRequest) error
	ResetCoupon(id uuid.UUID) error
	SendSchema(id uuid.UUID, email string) error
	MarkAsPurchased(id uuid.UUID, purchaseEmail string) error

	SearchCoupons(code, status, size, style string, partnerID *uuid.UUID) ([]*Coupon, error)
	SearchCouponsWithPagination(code, status, size, style string, partnerID *uuid.UUID, page, limit int) ([]*Coupon, int64, error)
	SearchCouponsByPartner(partnerID uuid.UUID, status, size, style string) ([]*Coupon, error)

	GetStatistics(partnerID *uuid.UUID) (map[string]int64, error)

	ValidateCoupon(code string) (*CouponValidationResponse, error)

	ExportCoupons(partnerID *uuid.UUID, status, format string) (string, error)
	ExportCouponsAdvanced(options ExportOptionsRequest) ([]byte, string, string, error)

	DownloadMaterials(id uuid.UUID) ([]byte, string, error)

	BatchResetCoupons(couponIDs []string) (*BatchResetResponse, error)
	PreviewBatchDelete(couponIDs []string) (*BatchDeletePreviewResponse, error)
	ExecuteBatchDelete(request BatchDeleteConfirmRequest) (*BatchDeleteResponse, error)
}
