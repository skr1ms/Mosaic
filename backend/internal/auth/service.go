package auth

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/skr1ms/mosaic/internal/admin"
	"github.com/skr1ms/mosaic/internal/partner"
	"github.com/skr1ms/mosaic/pkg/bcrypt"
	"github.com/skr1ms/mosaic/pkg/jwt"
)

type AuthServiceDeps struct {
	PartnerRepository PartnerRepositoryInterface
	AdminRepository   AdminRepositoryInterface
	JwtService        JWTServiceInterface
	Recaptcha         RecaptchaInterface
	MailSender        MailSenderInterface
	Config            ConfigInterface
}

type AuthService struct {
	deps *AuthServiceDeps
}

func NewAuthService(deps *AuthServiceDeps) *AuthService {
	return &AuthService{
		deps: deps,
	}
}

// AdminLogin authenticates admin and generates JWT tokens
func (s *AuthService) AdminLogin(login, password string) (*admin.Admin, *jwt.TokenPair, error) {
	admin, err := s.deps.AdminRepository.GetByLogin(login)
	if err != nil && admin == nil {
		return nil, nil, fmt.Errorf("admin not found: %w", err)
	}

	if !bcrypt.CheckPassword(password, admin.Password) {
		return nil, nil, fmt.Errorf("invalid credentials")
	}

	tokenPair, err := s.deps.JwtService.CreateTokenPair(admin.ID, admin.Login, admin.Role)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create token pair: %w", err)
	}

	if err := s.deps.AdminRepository.UpdateLastLogin(admin.ID); err != nil {
		return nil, nil, fmt.Errorf("failed to update last login: %w", err)
	}

	return admin, tokenPair, nil
}

// PartnerLogin authenticates partner and generates JWT tokens
func (s *AuthService) PartnerLogin(login, password string) (*partner.Partner, *jwt.TokenPair, error) {
	partner, err := s.deps.PartnerRepository.GetByLogin(context.Background(), login)
	if err != nil && partner == nil {
		return nil, nil, fmt.Errorf("partner not found: %w", err)
	}

	if !bcrypt.CheckPassword(password, partner.Password) {
		return nil, nil, fmt.Errorf("invalid credentials")
	}

	if partner.Status == "blocked" {
		return nil, nil, fmt.Errorf("partner blocked")
	}

	tokenPair, err := s.deps.JwtService.CreateTokenPair(partner.ID, partner.Login, "partner")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create token pair: %w", err)
	}

	if err := s.deps.PartnerRepository.UpdateLastLogin(context.Background(), partner.ID); err != nil {
		return nil, nil, fmt.Errorf("failed to update last login: %w", err)
	}

	return partner, tokenPair, nil
}

// RefreshTokens refreshes JWT tokens using refresh token
func (s *AuthService) RefreshTokens(refreshToken string) (*jwt.TokenPair, error) {
	claims, err := s.deps.JwtService.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	if claims.Role != "partner" && claims.Role != "admin" && claims.Role != "main_admin" {
		return nil, fmt.Errorf("invalid token role")
	}

	tokenPair, err := s.deps.JwtService.RefreshTokens(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh tokens: %w", err)
	}

	return tokenPair, nil
}

// ChangeAdminEmail changes admin email after password verification
func (s *AuthService) ChangeAdminEmail(adminID uuid.UUID, currentPassword, newEmail string) error {
	admin, err := s.deps.AdminRepository.GetByID(adminID)
	if err != nil || admin == nil {
		return fmt.Errorf("admin not found")
	}

	if !bcrypt.CheckPassword(currentPassword, admin.Password) {
		return fmt.Errorf("invalid password")
	}

	if existing, err := s.deps.AdminRepository.GetByEmail(newEmail); err == nil && existing != nil && existing.ID != adminID {
		return fmt.Errorf("email already in use")
	}

	if err := s.deps.AdminRepository.UpdateEmail(adminID, newEmail); err != nil {
		return fmt.Errorf("failed to update email: %w", err)
	}

	return nil
}

// ChangePartnerEmail changes partner email after password verification
func (s *AuthService) ChangePartnerEmail(partnerID uuid.UUID, currentPassword, newEmail string) error {
	partner, err := s.deps.PartnerRepository.GetByID(context.Background(), partnerID)
	if err != nil || partner == nil {
		return fmt.Errorf("partner not found")
	}

	if !bcrypt.CheckPassword(currentPassword, partner.Password) {
		return fmt.Errorf("invalid password")
	}

	if existing, err := s.deps.PartnerRepository.GetByEmail(context.Background(), newEmail); err == nil && existing != nil && existing.ID != partnerID {
		return fmt.Errorf("email already in use")
	}

	if err := s.deps.PartnerRepository.UpdateEmail(context.Background(), partnerID, newEmail); err != nil {
		return fmt.Errorf("failed to update email: %w", err)
	}

	return nil
}

// ForgotPassword handles password reset request for partners and admins
func (s *AuthService) ForgotPassword(ctx context.Context, login, email, captcha string) error {
	log := zerolog.Ctx(ctx)

	if s.deps.Config.GetRecaptchaConfig().Environment == "development" {
		if strings.TrimSpace(captcha) == "" {
			return fmt.Errorf("invalid captcha: empty reCAPTCHA token")
		}
		valid, err := s.deps.Recaptcha.Verify(captcha, "forgot_password")
		if err != nil || !valid {
			return fmt.Errorf("invalid captcha: %w", err)
		}
	} else {
		valid, err := s.deps.Recaptcha.Verify(captcha, "forgot_password")
		if err != nil || !valid {
			return fmt.Errorf("invalid captcha: %w", err)
		}
	}

	admin, adminErr := s.deps.AdminRepository.GetByLogin(login)
	if adminErr == nil && admin != nil {
		if admin.Email != email {
			log.Error().Str("login", login).Str("email", email).Msg("Admin email mismatch")
			return fmt.Errorf("email mismatch")
		}

		resetToken, err := s.deps.JwtService.CreatePasswordResetToken(admin.ID, admin.Email)
		if err != nil {
			log.Error().Err(err).Str("admin_id", admin.ID.String()).Msg("Failed to create reset token")
			return fmt.Errorf("failed to create token: %w", err)
		}

		resetLink := fmt.Sprintf("%s/#/reset?token=%s",
			s.deps.Config.GetServerConfig().FrontendURL,
			resetToken,
		)

		if err := s.deps.MailSender.SendResetPasswordEmail(admin.Email, resetLink); err != nil {
			log.Error().Err(err).Str("admin_email", admin.Email).Msg("Failed to send reset email")
			return fmt.Errorf("failed to send email: %w", err)
		}

		return nil
	}

	partner, partnerErr := s.deps.PartnerRepository.GetByLogin(ctx, login)
	if partnerErr != nil || partner == nil {
		log.Error().Str("login", login).Msg("User not found")
		return fmt.Errorf("user not found")
	}

	if partner.Email != email {
		log.Error().Str("login", login).Str("email", email).Msg("Partner email mismatch")
		return fmt.Errorf("email mismatch")
	}

	if partner.Status != "active" {
		log.Error().Str("partner_id", partner.ID.String()).Str("status", partner.Status).Msg("Partner is not active")
		return fmt.Errorf("partner status is not active")
	}

	resetToken, err := s.deps.JwtService.CreatePasswordResetToken(partner.ID, partner.Email)
	if err != nil {
		log.Error().Err(err).Str("partner_id", partner.ID.String()).Msg("Failed to create reset token")
		return fmt.Errorf("failed to create token: %w", err)
	}

	resetLink := fmt.Sprintf("%s/#/reset?token=%s",
		s.deps.Config.GetServerConfig().FrontendURL,
		resetToken,
	)

	if err := s.deps.MailSender.SendResetPasswordEmail(partner.Email, resetLink); err != nil {
		log.Error().Err(err).Str("partner_email", partner.Email).Msg("Failed to send reset email")
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// ResetPassword resets user password using reset token
func (s *AuthService) ResetPassword(ctx context.Context, token, newPassword string) error {
	log := zerolog.Ctx(ctx)

	claims, err := s.deps.JwtService.ValidatePasswordResetToken(token)
	if err != nil {
		log.Error().Err(err).Msg("Invalid reset token")
		return fmt.Errorf("invalid token: %w", err)
	}

	hashedPassword, err := bcrypt.HashPassword(newPassword)
	if err != nil {
		log.Error().Err(err).Msg("Failed to hash password")
		return fmt.Errorf("failed to hash password: %w", err)
	}

	admin, adminErr := s.deps.AdminRepository.GetByID(claims.UserID)
	if adminErr == nil && admin != nil {
		if admin.Email != claims.Login {
			log.Error().Str("admin_id", admin.ID.String()).Str("token_email", claims.Login).Msg("Admin email mismatch")
			return fmt.Errorf("invalid token")
		}

		if err := s.deps.AdminRepository.UpdatePassword(admin.ID, hashedPassword); err != nil {
			log.Error().Err(err).Str("admin_id", admin.ID.String()).Msg("Failed to update admin password")
			return fmt.Errorf("failed to update password: %w", err)
		}

		return nil
	}

	partner, partnerErr := s.deps.PartnerRepository.GetByID(ctx, claims.UserID)
	if partnerErr != nil || partner == nil {
		log.Error().Str("user_id", claims.UserID.String()).Msg("User not found")
		return fmt.Errorf("user not found")
	}

	if partner.Email != claims.Login {
		log.Error().Str("partner_id", partner.ID.String()).Str("token_email", claims.Login).Msg("Partner email mismatch")
		return fmt.Errorf("invalid token")
	}

	if partner.Status != "active" {
		log.Error().Str("partner_id", partner.ID.String()).Str("status", partner.Status).Msg("Partner is not active")
		return fmt.Errorf("partner status is not active")
	}

	if err := s.deps.PartnerRepository.UpdatePassword(ctx, partner.ID, hashedPassword); err != nil {
		log.Error().Err(err).Str("partner_id", partner.ID.String()).Msg("Failed to update partner password")
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// ChangePassword changes user password after verifying current password
func (s *AuthService) ChangePassword(userID uuid.UUID, userRole, currentPassword, newPassword string) error {
	log := zerolog.Ctx(context.Background())

	hashedPassword, err := bcrypt.HashPassword(newPassword)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("Failed to hash password")
		return fmt.Errorf("failed to hash password: %w", err)
	}

	if userRole == "admin" || userRole == "main_admin" {
		admin, err := s.deps.AdminRepository.GetByID(userID)
		if err != nil {
			log.Error().Err(err).Str("admin_id", userID.String()).Msg("Failed to find admin by ID")
			return fmt.Errorf("failed to find admin by ID: %w", err)
		}

		if !bcrypt.CheckPassword(currentPassword, admin.Password) {
			log.Error().Str("admin_id", userID.String()).Msg("Invalid current password")
			return fmt.Errorf("invalid current password")
		}

		if err := s.deps.AdminRepository.UpdatePassword(userID, hashedPassword); err != nil {
			log.Error().Err(err).Str("admin_id", userID.String()).Msg("Failed to change password")
			return fmt.Errorf("failed to change password: %w", err)
		}

		return nil
	}

	if userRole == "partner" {
		partner, err := s.deps.PartnerRepository.GetByID(context.Background(), userID)
		if err != nil {
			log.Error().Err(err).Str("partner_id", userID.String()).Msg("Failed to find partner by ID")
			return fmt.Errorf("failed to find partner by ID: %w", err)
		}

		if !bcrypt.CheckPassword(currentPassword, partner.Password) {
			log.Error().Str("partner_id", userID.String()).Msg("Invalid current password")
			return fmt.Errorf("invalid current password")
		}

		if err := s.deps.PartnerRepository.UpdatePassword(context.Background(), userID, hashedPassword); err != nil {
			log.Error().Err(err).Str("partner_id", userID.String()).Msg("Failed to change password")
			return fmt.Errorf("failed to change password: %w", err)
		}

		return nil
	}

	return fmt.Errorf("invalid user role: %s", userRole)
}
