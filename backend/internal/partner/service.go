package partner

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/skr1ms/mosaic/config"
	"github.com/skr1ms/mosaic/internal/coupon"
	"github.com/skr1ms/mosaic/pkg/bcrypt"
)

type PartnerServiceDeps struct {
	PartnerRepository PartnerRepositoryInterface
	CouponService     CouponServiceInterface
	Recaptcha         RecaptchaInterface
	JwtService        JWTInterface
	MailSender        MailerInterface
	Config            *config.Config
}

type PartnerService struct {
	deps *PartnerServiceDeps
}

func NewPartnerService(deps *PartnerServiceDeps) *PartnerService {
	return &PartnerService{
		deps: deps,
	}
}

func (s *PartnerService) ExportCoupons(partnerID uuid.UUID, status, format string) ([]byte, string, string, error) {
	// Создаем запрос для экспорта
	partnerIDStr := partnerID.String()
	options := coupon.ExportOptionsRequest{
		Format:        coupon.ExportFormatCodes, // Только коды купонов
		PartnerID:     &partnerIDStr,
		Status:        status,
		FileFormat:    format,
		IncludeHeader: false,
	}

	// Используем новый продвинутый метод экспорта
	content, filename, contentType, err := s.deps.CouponService.ExportCouponsAdvanced(options)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to export coupons: %w", err)
	}

	return content, filename, contentType, nil
}

func (s *PartnerService) UpdatePassword(partnerID uuid.UUID, currentPassword, newPassword string) error {
	// Получаем текущего партнера
	partner, err := s.deps.PartnerRepository.GetByID(context.Background(), partnerID)
	if err != nil {
		return fmt.Errorf("failed to find partner by id: %w", err)
	}

	// Проверяем текущий пароль
	if !bcrypt.CheckPassword(currentPassword, partner.Password) {
		return fmt.Errorf("current password is incorrect: %w", err)
	}

	// Хешируем новый пароль
	hashedPassword, err := bcrypt.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Обновляем пароль
	if err := s.deps.PartnerRepository.UpdatePassword(context.Background(), partnerID, hashedPassword); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

func (s *PartnerService) ForgotPassword(ctx context.Context, email string /*captcha string*/) error {
	// Находим партнера по email
	partner, err := s.deps.PartnerRepository.GetByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("failed to find partner by email: %w", err)
	}

	// Проверяем статус партнера
	if partner.Status != "active" {
		return fmt.Errorf("partner status is not active: %w", err)
	}

	// Проверяем капчу
	if s.deps.Config.RecaptchaConfig.Environment == "development" {
		fmt.Println("reCAPTCHA verification is disabled in development mode")
	} /*else {
		valid, err := s.deps.Recaptcha.Verify(captcha, "forgot_password")
		if err != nil || !valid {
			return fmt.Errorf("invalid captcha: %w", err)
		}
	}*/

	// Проверяем статус партнера
	resetToken, err := s.deps.JwtService.CreatePasswordResetToken(partner.ID, partner.Email)
	if err != nil {
		return fmt.Errorf("failed to create token: %w", err)
	}

	// Формируем ссылку для сброса пароля
	resetLink := fmt.Sprintf("%s/reset?token=%s",
		s.deps.Config.ServerConfig.FrontendURL,
		resetToken,
	)

	// Отправляем письмо с ссылкой для сброса пароля
	if err := s.deps.MailSender.SendResetPasswordEmail(partner.Email, resetLink); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func (s *PartnerService) ResetPassword(ctx context.Context, token, newPassword string) error {
	// Валидируем токен сброса пароля
	claims, err := s.deps.JwtService.ValidatePasswordResetToken(token)
	if err != nil {
		return fmt.Errorf("invalid token: %w", err)
	}

	// Находим партнера
	partner, err := s.deps.PartnerRepository.GetByEmail(ctx, claims.Login) // login == email
	if err != nil {
		return fmt.Errorf("failed to find partner by email: %w", err)
	}

	// Проверяем статус партнера
	if partner.Status != "active" {
		return fmt.Errorf("partner status is not active: %w", err)
	}

	// Хешируем новый пароль
	hashedPassword, err := bcrypt.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Обновляем пароль
	if err := s.deps.PartnerRepository.UpdatePassword(ctx, claims.UserID, hashedPassword); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// DeletePartnerWithCoupons удаляет партнера и все его купоны
func (s *PartnerService) DeletePartnerWithCoupons(ctx context.Context, partnerID uuid.UUID) error {
	// Начинаем транзакцию
	err := s.deps.PartnerRepository.DeleteWithCoupons(ctx, partnerID)
	if err != nil {
		return fmt.Errorf("failed to delete partner: %w", err)
	}

	return nil
}
