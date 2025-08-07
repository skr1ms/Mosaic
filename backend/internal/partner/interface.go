package partner

import (
	"context"

	"github.com/google/uuid"
	"github.com/skr1ms/mosaic/internal/coupon"
)

type TokenClaims struct {
	UserID uuid.UUID
	Login  string
}

type PartnerRepositoryInterface interface {
	Create(ctx context.Context, partner *Partner) error
	GetByID(ctx context.Context, id uuid.UUID) (*Partner, error)
	GetByLogin(ctx context.Context, login string) (*Partner, error)
	GetByPartnerCode(ctx context.Context, code string) (*Partner, error)
	GetByDomain(ctx context.Context, domain string) (*Partner, error)
	GetByEmail(ctx context.Context, email string) (*Partner, error)
	Update(ctx context.Context, partner *Partner) error
	UpdatePassword(ctx context.Context, partnerID uuid.UUID, hashedPassword string) error
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
}

type CouponServiceInterface interface {
	ExportCouponsAdvanced(options coupon.ExportOptionsRequest) ([]byte, string, string, error)
}

type RecaptchaInterface interface {
	Verify(token, action string) (bool, error)
}

type JWTInterface interface {
	CreatePasswordResetToken(userID uuid.UUID, email string) (string, error)
	ValidatePasswordResetToken(token string) (*TokenClaims, error)
}

type MailerInterface interface {
	SendResetPasswordEmail(email, resetLink string) error
}
