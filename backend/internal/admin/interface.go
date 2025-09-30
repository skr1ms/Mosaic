package admin

import (
	"context"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/skr1ms/mosaic/config"
	"github.com/skr1ms/mosaic/internal/coupon"
	"github.com/skr1ms/mosaic/internal/image"
	"github.com/skr1ms/mosaic/internal/partner"
	"github.com/skr1ms/mosaic/pkg/jwt"
)

type ConfigInterface interface {
	GetServerConfig() config.ServerConfig
	GetRecaptchaConfig() config.RecaptchaConfig
	GetAlfaBankConfig() config.AlphaBankConfig
	GetS3MinioConfig() config.S3MinioConfig
	GetStableDiffusionConfig() config.StableDiffusionConfig
	GetMosaicGeneratorConfig() config.MosaicGeneratorConfig
	GetGitLabConfig() config.GitLabConfig
}

type AdminRepositoryInterface interface {
	Create(admin *Admin) error
	GetByLogin(login string) (*Admin, error)
	GetByID(id uuid.UUID) (*Admin, error)
	GetAll() ([]*Admin, error)
	UpdateLastLogin(id uuid.UUID) error
	UpdatePassword(id uuid.UUID, hashedPassword string) error
	UpdateEmail(id uuid.UUID, newEmail string) error
	GetByEmail(email string) (*Admin, error)
	Delete(id uuid.UUID) error

	CreateProfileChangeLog(log *ProfileChangeLog) error
	GetProfileChangesByPartnerID(partnerID uuid.UUID) ([]*ProfileChangeLog, error)
}

type PartnerRepositoryInterface interface {
	GetAll(ctx context.Context, sortBy string, order string) ([]*partner.Partner, error)
	GetByID(ctx context.Context, id uuid.UUID) (*partner.Partner, error)
	GetByLogin(ctx context.Context, login string) (*partner.Partner, error)
	GetByDomain(ctx context.Context, domain string) (*partner.Partner, error)
	GetByPartnerCode(ctx context.Context, code string) (*partner.Partner, error)
	GetNextPartnerCode(ctx context.Context) (string, error)
	Create(ctx context.Context, p *partner.Partner) error
	Update(ctx context.Context, p *partner.Partner) error
	DeleteWithCoupons(ctx context.Context, id uuid.UUID) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	Search(ctx context.Context, search, status, sortBy, order string) ([]*partner.Partner, error)
	GetActivePartners(ctx context.Context) ([]*partner.Partner, error)
	InitializeArticleGrid(ctx context.Context, partnerID uuid.UUID) error
	GetArticleGrid(ctx context.Context, partnerID uuid.UUID) (map[string]map[string]map[string]string, error)
	UpdateArticleSKU(ctx context.Context, partnerID uuid.UUID, size, style, marketplace, sku string) error
	GetArticleBySizeStyle(ctx context.Context, partnerID uuid.UUID, size, style, marketplace string) (*partner.PartnerArticle, error)
	GetAllArticlesByPartner(ctx context.Context, partnerID uuid.UUID) ([]*partner.PartnerArticle, error)
	DeleteArticleGrid(ctx context.Context, partnerID uuid.UUID) error
}

type CouponRepositoryInterface interface {
	Create(ctx context.Context, coupon *coupon.Coupon) error
	CreateBatch(ctx context.Context, coupons []*coupon.Coupon) error
	GetByID(ctx context.Context, id uuid.UUID) (*coupon.Coupon, error)
	GetByCode(ctx context.Context, code string) (*coupon.Coupon, error)
	GetByPartnerID(ctx context.Context, partnerID uuid.UUID) ([]*coupon.Coupon, error)
	GetAll(ctx context.Context) ([]*coupon.Coupon, error)
	GetRecentActivated(ctx context.Context, limit int) ([]*coupon.Coupon, error)
	Update(ctx context.Context, coupon *coupon.Coupon) error
	Delete(ctx context.Context, id uuid.UUID) error
	BatchDelete(ctx context.Context, ids []uuid.UUID) (int64, error)
	Search(ctx context.Context, code, status, size, style string, partnerID *uuid.UUID) ([]*coupon.Coupon, error)
	SearchWithPagination(
		ctx context.Context,
		code, status, size, style string,
		partnerID *uuid.UUID,
		page, limit int,
		createdFrom, createdTo, usedFrom, usedTo *time.Time,
		sortBy, sortDir string,
	) ([]*coupon.Coupon, int, error)
	GetFiltered(ctx context.Context, filters map[string]any) ([]*coupon.Coupon, error)
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
	GetTopActivatedByPartner(ctx context.Context, limit int) ([]coupon.PartnerCount, error)
	GetTopPurchasedByPartner(ctx context.Context, limit int) ([]coupon.PartnerCount, error)
	GetLastActivityByPartner(ctx context.Context, partnerID uuid.UUID) (*time.Time, error)
	GetTimeSeriesData(ctx context.Context, from, to time.Time, period string, partnerID *uuid.UUID) ([]map[string]any, error)
	CodeExists(ctx context.Context, code string) (bool, error)
	HealthCheck(ctx context.Context) error
	GetCouponsWithAdvancedFilter(ctx context.Context, filter coupon.CouponFilterRequest) ([]*coupon.CouponInfo, int, error)
	GetCouponsForDeletion(ctx context.Context, ids []uuid.UUID) ([]*coupon.CouponDeletePreview, error)
	GetCouponsForExport(ctx context.Context, options coupon.ExportOptionsRequest) (any, error)
	CountCreatedByPartnerInRange(ctx context.Context, partnerID uuid.UUID, from, to *time.Time) (int, error)
	CountActivatedByPartnerInRange(ctx context.Context, partnerID uuid.UUID, from, to *time.Time) (int, error)
	CountPurchasedByPartnerInRange(ctx context.Context, partnerID uuid.UUID, from, to *time.Time) (int, error)
	ActivateCoupon(ctx context.Context, id uuid.UUID, req coupon.ActivateCouponRequest) error
	SendSchema(ctx context.Context, id uuid.UUID, email string) error
	MarkAsPurchased(ctx context.Context, id uuid.UUID, purchaseEmail string) error
	ResetCoupon(ctx context.Context, id uuid.UUID) error
	Reset(ctx context.Context, id uuid.UUID) error
	BatchReset(ctx context.Context, ids []uuid.UUID) ([]uuid.UUID, []uuid.UUID, error)
	GetPartnerCouponsWithFilter(ctx context.Context, partnerID uuid.UUID, filters map[string]any, page, limit int, sortBy, order string) ([]*coupon.Coupon, int, error)
	GetPartnerCouponByCode(ctx context.Context, partnerID uuid.UUID, code string) (*coupon.Coupon, error)
	GetPartnerCouponDetail(ctx context.Context, partnerID uuid.UUID, couponID uuid.UUID) (*coupon.Coupon, error)
	GetPartnerRecentActivity(ctx context.Context, partnerID uuid.UUID, limit int) ([]*coupon.Coupon, error)
	UpdateStatusByPartnerID(ctx context.Context, partnerID uuid.UUID, status bool) error
}

type ImageRepositoryInterface interface {
	GetAll(ctx context.Context) ([]*image.Image, error)
	GetByID(ctx context.Context, id uuid.UUID) (*image.Image, error)
	GetByCouponID(ctx context.Context, couponID uuid.UUID) (*image.Image, error)
	GetByStatus(ctx context.Context, status string) ([]*image.Image, error)
	Update(ctx context.Context, img *image.Image) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type S3ClientInterface interface {
	DownloadFile(ctx context.Context, key string) (io.ReadCloser, error)
	UploadLogo(ctx context.Context, reader io.Reader, size int64, contentType string, partnerID string) (string, error)
	GetLogoURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error)
	DeleteFile(ctx context.Context, objectKey string) error
}

type RedisClientInterface interface {
	Ping(ctx context.Context) *redis.StatusCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	LLen(ctx context.Context, key string) *redis.IntCmd
}

type AdminServiceInterface interface {
	CreateAdmin(req CreateAdminRequest) (*Admin, error)
	GetAdmins() ([]*Admin, error)
	DeleteAdmin(id uuid.UUID) error
	UpdateAdminPassword(id uuid.UUID, newPassword string) error
	UpdateAdminEmail(id uuid.UUID, newEmail string) error
	GetDashboardData() (map[string]any, error)
	GetPartners(search, status, sortBy, order string) ([]*partner.Partner, error)
	GetPartnerDetail(partnerID uuid.UUID) (*PartnerDetailResponse, error)
	CreatePartner(req partner.CreatePartnerRequest) (*partner.Partner, error)
	UpdatePartnerWithHistory(partnerID uuid.UUID, req partner.UpdatePartnerRequest, adminLogin, reason string) (*partner.Partner, error)
	ExportCouponsAdvanced(options coupon.ExportOptionsRequest) ([]byte, string, string, error)
	DownloadCouponMaterials(couponID uuid.UUID) ([]byte, string, error)
	GetStatistics() (map[string]any, error)
	GetSystemStatistics() (map[string]any, error)
	GetImageDetails(imageID uuid.UUID) (*image.Image, error)
	DeleteImageTask(imageID uuid.UUID) error
	RetryImageTask(imageID uuid.UUID) error
	BatchResetCoupons(couponIDs []string) (*coupon.BatchResetResponse, error)

	ResetCoupon(id uuid.UUID) error
	DeleteCoupon(id uuid.UUID) error

	GetActivePartnersWithDomains() ([]partner.Partner, error)
	DeletePartner(id uuid.UUID) error

	GetPartnerRepository() PartnerRepositoryInterface
	GetCouponRepository() CouponRepositoryInterface
	GetImageRepository() ImageRepositoryInterface
	GetS3Client() S3ClientInterface
	DeployNginxConfig() error
}

type JWTServiceInterface interface {
	CreateTokenPair(userID uuid.UUID, login, role string) (*jwt.TokenPair, error)
	ValidateRefreshToken(refreshToken string) (*jwt.Claims, error)
	RefreshTokens(refreshToken string) (*jwt.TokenPair, error)
	CreatePasswordResetToken(userID uuid.UUID, email string) (string, error)
	ValidatePasswordResetToken(token string) (*jwt.Claims, error)
}
