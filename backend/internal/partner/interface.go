package partner

import (
	"context"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/skr1ms/mosaic/config"
	"github.com/skr1ms/mosaic/internal/coupon"
	"github.com/skr1ms/mosaic/internal/image"
	"github.com/skr1ms/mosaic/pkg/jwt"
)

type PartnerRepositoryInterface interface {
	Create(ctx context.Context, partner *Partner) error
	GetByID(ctx context.Context, id uuid.UUID) (*Partner, error)
	GetByLogin(ctx context.Context, login string) (*Partner, error)
	GetByPartnerCode(ctx context.Context, code string) (*Partner, error)
	GetByDomain(ctx context.Context, domain string) (*Partner, error)
	GetByEmail(ctx context.Context, email string) (*Partner, error)
	Update(ctx context.Context, partner *Partner) error
	UpdatePassword(ctx context.Context, partnerID uuid.UUID, hashedPassword string) error
	UpdateEmail(ctx context.Context, partnerID uuid.UUID, email string) error
	DeleteWithCoupons(ctx context.Context, id uuid.UUID) error
	GetAll(ctx context.Context, sortBy string, order string) ([]*Partner, error)
	GetActivePartners(ctx context.Context) ([]*Partner, error)
	UpdateLastLogin(ctx context.Context, id uuid.UUID) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	Delete(ctx context.Context, id uuid.UUID) error
	Search(ctx context.Context, queryStr string, status string, sortBy string, order string) ([]*Partner, error)
	CountActive(ctx context.Context) (int64, error)
	CountTotal(ctx context.Context) (int64, error)
	GetTopByActivity(ctx context.Context, limit int) ([]*Partner, error)
	GetNextPartnerCode(ctx context.Context) (string, error)
	GetCouponsStatistics(ctx context.Context, partnerID uuid.UUID) (map[string]int64, error)
	GetPartnerCouponsForExport(ctx context.Context, partnerID uuid.UUID, status string) ([]*ExportCouponRequest, error)
	GetAllCouponsForExport(ctx context.Context) ([]*ExportCouponRequest, error)
	InitializeArticleGrid(ctx context.Context, partnerID uuid.UUID) error
	GetArticleGrid(ctx context.Context, partnerID uuid.UUID) (map[string]map[string]map[string]string, error)
	UpdateArticleSKU(ctx context.Context, partnerID uuid.UUID, size, style, marketplace, sku string) error
	GetArticleBySizeStyle(ctx context.Context, partnerID uuid.UUID, size, style, marketplace string) (*PartnerArticle, error)
	GetAllArticlesByPartner(ctx context.Context, partnerID uuid.UUID) ([]*PartnerArticle, error)
	DeleteArticleGrid(ctx context.Context, partnerID uuid.UUID) error
}

type CouponServiceInterface interface {
	ExportCouponsAdvanced(options coupon.ExportOptionsRequest) ([]byte, string, string, error)
}

type RecaptchaInterface interface {
	Verify(token, action string) (bool, error)
}

type JWTInterface interface {
	CreatePasswordResetToken(userID uuid.UUID, email string) (string, error)
	ValidatePasswordResetToken(token string) (*jwt.Claims, error)
}

type MailerInterface interface {
	SendResetPasswordEmail(email, resetLink string) error
}

type ConfigInterface interface {
	GetServerConfig() config.ServerConfig
	GetRecaptchaConfig() config.RecaptchaConfig
}

type PartnerServiceInterface interface {
	DeletePartnerWithCoupons(ctx context.Context, partnerID uuid.UUID) error
	ExportCoupons(partnerID uuid.UUID, status string, format string) ([]byte, string, string, error)
	GetComparisonStatistics(ctx context.Context, partnerID uuid.UUID) (map[string]any, error)
	DownloadCouponMaterials(id uuid.UUID) ([]byte, string, error)

	GetPartnerRepository() PartnerRepositoryInterface
	GetCouponRepository() CouponRepositoryInterface

	InitializeArticleGrid(partnerID uuid.UUID) error
	GetArticleGrid(partnerID uuid.UUID) (map[string]map[string]map[string]string, error)
	UpdateArticleSKU(partnerID uuid.UUID, size, style, marketplace, sku string) error
	GetArticleBySizeStyle(partnerID uuid.UUID, size, style, marketplace string) (*PartnerArticle, error)
	GenerateProductLink(partnerID uuid.UUID, size, style, marketplace string) string
}

type CouponRepositoryInterface interface {
	GetByPartnerID(ctx context.Context, partnerID uuid.UUID) ([]*coupon.Coupon, error)
	GetByCode(ctx context.Context, code string) (*coupon.Coupon, error)
	GetByID(ctx context.Context, id uuid.UUID) (*coupon.Coupon, error)
	CountByPartnerID(ctx context.Context, partnerID uuid.UUID) (int, error)
	GetStatistics(ctx context.Context, partnerID *uuid.UUID) (map[string]int64, error)
	GetRecentActivatedByPartner(ctx context.Context, partnerID uuid.UUID, limit int) ([]*coupon.Coupon, error)
	SearchPartnerCoupons(ctx context.Context, partnerID uuid.UUID, code, status, size, style string, createdFrom, createdTo, usedFrom, usedTo *time.Time, sortBy, sortOrder string, page, limit int) ([]*coupon.Coupon, int, error)
	GetTopActivatedByPartner(ctx context.Context, limit int) ([]coupon.PartnerCount, error)
	GetTopPurchasedByPartner(ctx context.Context, limit int) ([]coupon.PartnerCount, error)
	CountActivatedByPartner(ctx context.Context, partnerID uuid.UUID) (int64, error)
	CountBrandedPurchasesByPartner(ctx context.Context, partnerID uuid.UUID) (int64, error)
}

type JWTServiceInterface interface {
	CreateTokenPair(userID uuid.UUID, login, role string) (*jwt.TokenPair, error)
	ValidateRefreshToken(refreshToken string) (*jwt.Claims, error)
	RefreshTokens(refreshToken string) (*jwt.TokenPair, error)
	CreatePasswordResetToken(userID uuid.UUID, email string) (string, error)
	ValidatePasswordResetToken(token string) (*jwt.Claims, error)
	ValidateAccessToken(token string) (*jwt.Claims, error)
}

type MailSenderInterface interface {
	SendResetPasswordEmail(to, resetLink string) error
}

type BrandingHelperInterface interface {
	GetPartnerBranding(partnerCode string) (map[string]any, error)
	UpdatePartnerBranding(partnerCode string, branding map[string]any) error
	GetDefaultBranding() map[string]any
}

type ImageRepositoryInterface interface {
	GetByCouponID(ctx context.Context, couponID uuid.UUID) (*image.Image, error)
}

type S3ClientInterface interface {
	DownloadFile(ctx context.Context, key string) (io.ReadCloser, error)
}
