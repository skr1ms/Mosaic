package auth

import (
	"context"

	"github.com/google/uuid"
	"github.com/skr1ms/mosaic/internal/admin"
	"github.com/skr1ms/mosaic/internal/partner"
	"github.com/skr1ms/mosaic/pkg/jwt"
)

type AuthServiceInterface interface {
	AdminLogin(login, password string) (*admin.Admin, *jwt.TokenPair, error)
	PartnerLogin(login, password string) (*partner.Partner, *jwt.TokenPair, error)
	RefreshAdminTokens(refreshToken string) (*jwt.TokenPair, error)
	RefreshPartnerTokens(refreshToken string) (*jwt.TokenPair, error)
}

type AdminRepositoryInterface interface {
	GetByLogin(login string) (*admin.Admin, error)
	UpdateLastLogin(id uuid.UUID) error
}

type PartnerRepositoryInterface interface {
	GetByLogin(ctx context.Context, login string) (*partner.Partner, error)
	UpdateLastLogin(ctx context.Context, id uuid.UUID) error
}

type JwtServiceInterface interface {
	CreateTokenPair(userID uuid.UUID, login, role string) (*jwt.TokenPair, error)
	ValidateRefreshToken(refreshToken string) (*jwt.Claims, error)
	RefreshTokens(refreshToken string) (*jwt.TokenPair, error)
}
