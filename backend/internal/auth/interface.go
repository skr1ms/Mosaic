package auth

import (
	"context"

	"github.com/google/uuid"
	"github.com/skr1ms/mosaic/config"
	"github.com/skr1ms/mosaic/internal/admin"
	"github.com/skr1ms/mosaic/internal/partner"
	"github.com/skr1ms/mosaic/pkg/jwt"
)

type AuthServiceInterface interface {
	AdminLogin(login, password string) (*admin.Admin, *jwt.TokenPair, error)
	PartnerLogin(login, password string) (*partner.Partner, *jwt.TokenPair, error)
	RefreshTokens(refreshToken string) (*jwt.TokenPair, error)
	ForgotPassword(ctx context.Context, login, email, captcha string) error
	ResetPassword(ctx context.Context, token, newPassword string) error
	ChangePassword(userID uuid.UUID, userRole, currentPassword, newPassword string) error
	ChangeAdminEmail(adminID uuid.UUID, currentPassword, newEmail string) error
	ChangePartnerEmail(partnerID uuid.UUID, currentPassword, newEmail string) error
}

type AdminRepositoryInterface interface {
	GetByLogin(login string) (*admin.Admin, error)
	GetByEmail(email string) (*admin.Admin, error)
	GetByID(id uuid.UUID) (*admin.Admin, error)
	UpdateLastLogin(id uuid.UUID) error
	UpdatePassword(id uuid.UUID, hashedPassword string) error
	UpdateEmail(id uuid.UUID, email string) error
}

type PartnerRepositoryInterface interface {
	GetByLogin(ctx context.Context, login string) (*partner.Partner, error)
	GetByEmail(ctx context.Context, email string) (*partner.Partner, error)
	GetByID(ctx context.Context, id uuid.UUID) (*partner.Partner, error)
	UpdateLastLogin(ctx context.Context, id uuid.UUID) error
	UpdatePassword(ctx context.Context, id uuid.UUID, hashedPassword string) error
	UpdateEmail(ctx context.Context, id uuid.UUID, email string) error
}

type PartnerServiceInterface interface {
	PartnerLogin(login, password string) (*partner.Partner, *jwt.TokenPair, error)
}

type JWTServiceInterface interface {
	CreateTokenPair(userID uuid.UUID, login, role string) (*jwt.TokenPair, error)
	ValidateRefreshToken(refreshToken string) (*jwt.Claims, error)
	RefreshTokens(refreshToken string) (*jwt.TokenPair, error)
	CreatePasswordResetToken(userID uuid.UUID, email string) (string, error)
	ValidatePasswordResetToken(token string) (*jwt.Claims, error)
}

type RecaptchaInterface interface {
	Verify(token, expectedAction string) (bool, error)
}

type MailSenderInterface interface {
	SendResetPasswordEmail(to, resetLink string) error
}

type ConfigInterface interface {
	GetRecaptchaConfig() config.RecaptchaConfig
	GetServerConfig() config.ServerConfig
}
