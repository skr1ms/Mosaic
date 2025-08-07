package auth

import (
	"context"
	"fmt"

	"github.com/skr1ms/mosaic/internal/admin"
	"github.com/skr1ms/mosaic/internal/partner"
	"github.com/skr1ms/mosaic/pkg/bcrypt"
	"github.com/skr1ms/mosaic/pkg/jwt"
)

type AuthServiceDeps struct {
	PartnerRepository PartnerRepositoryInterface
	AdminRepository   AdminRepositoryInterface
	JwtService        JwtServiceInterface
}

type AuthService struct {
	deps *AuthServiceDeps
}

func NewAuthService(deps *AuthServiceDeps) *AuthService {
	return &AuthService{
		deps: deps,
	}
}

// AdminLogin обрабатывает авторизацию администратора и генерирует JWT токены
func (s *AuthService) AdminLogin(login, password string) (*admin.Admin, *jwt.TokenPair, error) {
	// Находим администратора по логину
	admin, err := s.deps.AdminRepository.GetByLogin(login)
	if err != nil && admin == nil {
		return nil, nil, fmt.Errorf("admin not found: %w", err)
	}

	// Проверяем пароль
	if !bcrypt.CheckPassword(password, admin.Password) {
		return nil, nil, fmt.Errorf("invalid credentials")
	}

	// Генерируем токены
	tokenPair, err := s.deps.JwtService.CreateTokenPair(admin.ID, admin.Login, "admin")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create token pair: %w", err)
	}

	// Обновляем время последнего входа
	if err := s.deps.AdminRepository.UpdateLastLogin(admin.ID); err != nil {
		return nil, nil, fmt.Errorf("failed to update last login: %w", err)
	}

	return admin, tokenPair, nil
}

// PartnerLogin обрабатывает авторизацию партнера и генерирует JWT токены
func (s *AuthService) PartnerLogin(login, password string) (*partner.Partner, *jwt.TokenPair, error) {
	// Находим партнера по логину
	partner, err := s.deps.PartnerRepository.GetByLogin(context.Background(), login)
	if err != nil && partner == nil {
		return nil, nil, fmt.Errorf("partner not found: %w", err)
	}

	// Проверяем пароль
	if !bcrypt.CheckPassword(password, partner.Password) {
		return nil, nil, fmt.Errorf("invalid credentials")
	}

	// Проверяем статус партнера
	if partner.Status == "blocked" {
		return nil, nil, fmt.Errorf("partner blocked")
	}

	// Генерируем токены
	tokenPair, err := s.deps.JwtService.CreateTokenPair(partner.ID, partner.Login, "partner")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create token pair: %w", err)
	}

	// Обновляем время последнего входа
	if err := s.deps.PartnerRepository.UpdateLastLogin(context.Background(), partner.ID); err != nil {
		return nil, nil, fmt.Errorf("failed to update last login: %w", err)
	}

	return partner, tokenPair, nil
}

// RefreshAdminTokens обновляет токены администратора используя refresh токен
func (s *AuthService) RefreshAdminTokens(refreshToken string) (*jwt.TokenPair, error) {
	// Валидируем refresh токен
	claims, err := s.deps.JwtService.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Проверяем роль токена
	if claims.Role != "admin" {
		return nil, fmt.Errorf("invalid token role")
	}

	// Обновляем токены
	tokenPair, err := s.deps.JwtService.RefreshTokens(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh tokens: %w", err)
	}

	return tokenPair, nil
}

// RefreshPartnerTokens обновляет токены партнера используя refresh токен
func (s *AuthService) RefreshPartnerTokens(refreshToken string) (*jwt.TokenPair, error) {
	// Валидируем refresh токен
	claims, err := s.deps.JwtService.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Проверяем роль токена
	if claims.Role != "partner" {
		return nil, fmt.Errorf("invalid token role")
	}

	// Обновляем токены
	tokenPair, err := s.deps.JwtService.RefreshTokens(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh tokens: %w", err)
	}

	return tokenPair, nil
}
