package auth

import (
	"github.com/rs/zerolog"
	"github.com/skr1ms/mosaic/internal/admin"
	"github.com/skr1ms/mosaic/internal/partner"
	"github.com/skr1ms/mosaic/pkg/bcrypt"
	"github.com/skr1ms/mosaic/pkg/jwt"
)

type AuthService struct {
	partnerRepository *partner.PartnerRepository
	adminRepository   *admin.AdminRepository
	jwtService        *jwt.JWT
	Logger            *zerolog.Logger
}

func NewAuthService(partnerRepo *partner.PartnerRepository, adminRepo *admin.AdminRepository, jwtService *jwt.JWT, logger *zerolog.Logger) *AuthService {
	return &AuthService{
		partnerRepository: partnerRepo,
		adminRepository:   adminRepo,
		jwtService:        jwtService,
		Logger:            logger,
	}
}

// AdminLogin обрабатывает авторизацию администратора и генерирует JWT токены
func (s *AuthService) AdminLogin(login, password string) (*admin.Admin, *jwt.TokenPair, error) {
	// Находим администратора по логину
	admin, err := s.adminRepository.GetByLogin(login)
	if err != nil && admin == nil {
		s.Logger.Error().Err(err).Msg(ErrAdminNotFound.Error())
		return nil, nil, ErrAdminNotFound
	}

	// Проверяем пароль
	if !bcrypt.CheckPassword(password, admin.Password) {
		s.Logger.Error().Msg(ErrInvalidCredentials.Error())
		return nil, nil, ErrInvalidCredentials
	}

	// Генерируем токены
	tokenPair, err := s.jwtService.CreateTokenPair(admin.ID, admin.Login, "admin")
	if err != nil {
		s.Logger.Error().Err(err).Msg(ErrCreateTokenPair.Error())
		return nil, nil, ErrCreateTokenPair
	}

	// Обновляем время последнего входа
	if err := s.adminRepository.UpdateLastLogin(admin.ID); err != nil {
		s.Logger.Error().Err(err).Msg(ErrUpdateLastLogin.Error())
		return nil, nil, ErrUpdateLastLogin
	}

	return admin, tokenPair, nil
}

// PartnerLogin обрабатывает авторизацию партнера и генерирует JWT токены
func (s *AuthService) PartnerLogin(login, password string) (*partner.Partner, *jwt.TokenPair, error) {
	// Находим партнера по логину
	partner, err := s.partnerRepository.GetByLogin(login)
	if err != nil && partner == nil {
		s.Logger.Error().Err(err).Msg(ErrPartnerNotFound.Error())
		return nil, nil, ErrPartnerNotFound
	}

	// Проверяем пароль
	if !bcrypt.CheckPassword(password, partner.Password) {
		s.Logger.Error().Msg(ErrInvalidCredentials.Error())
		return nil, nil, ErrInvalidCredentials
	}

	// Проверяем статус партнера
	if partner.Status == "blocked" {
		s.Logger.Error().Msg(ErrPartnerBlocked.Error())
		return nil, nil, ErrPartnerBlocked
	}

	// Генерируем токены
	tokenPair, err := s.jwtService.CreateTokenPair(partner.ID, partner.Login, "partner")
	if err != nil {
		s.Logger.Error().Err(err).Msg(ErrCreateTokenPair.Error())
		return nil, nil, ErrCreateTokenPair
	}

	// Обновляем время последнего входа
	if err := s.partnerRepository.UpdateLastLogin(partner.ID); err != nil {
		s.Logger.Error().Err(err).Msg(ErrUpdateLastLogin.Error())
		return nil, nil, ErrUpdateLastLogin
	}

	return partner, tokenPair, nil
}

// RefreshAdminTokens обновляет токены администратора используя refresh токен
func (s *AuthService) RefreshAdminTokens(refreshToken string) (*jwt.TokenPair, error) {
	// Валидируем refresh токен
	claims, err := s.jwtService.ValidateRefreshToken(refreshToken)
	if err != nil {
		s.Logger.Error().Err(err).Msg(ErrInvalidRefreshToken.Error())
		return nil, ErrInvalidRefreshToken
	}

	// Проверяем роль токена
	if claims.Role != "admin" {
		s.Logger.Error().Msg(ErrInvalidTokenRole.Error())
		return nil, ErrInvalidTokenRole
	}

	// Обновляем токены
	tokenPair, err := s.jwtService.RefreshTokens(refreshToken)
	if err != nil {
		s.Logger.Error().Err(err).Msg(ErrRefreshTokens.Error())
		return nil, ErrRefreshTokens
	}

	return tokenPair, nil
}

// RefreshPartnerTokens обновляет токены партнера используя refresh токен
func (s *AuthService) RefreshPartnerTokens(refreshToken string) (*jwt.TokenPair, error) {
	// Валидируем refresh токен
	claims, err := s.jwtService.ValidateRefreshToken(refreshToken)
	if err != nil {
		s.Logger.Error().Err(err).Msg(ErrInvalidRefreshToken.Error())
		return nil, ErrInvalidRefreshToken
	}

	// Проверяем роль токена
	if claims.Role != "partner" {
		s.Logger.Error().Msg(ErrInvalidTokenRole.Error())
		return nil, ErrInvalidTokenRole
	}

	// Обновляем токены
	tokenPair, err := s.jwtService.RefreshTokens(refreshToken)
	if err != nil {
		s.Logger.Error().Err(err).Msg(ErrRefreshTokens.Error())
		return nil, ErrRefreshTokens
	}

	return tokenPair, nil
}
