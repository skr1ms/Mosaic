package partner

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/skr1ms/mosaic/config"
	"github.com/skr1ms/mosaic/pkg/bcrypt"
	"github.com/skr1ms/mosaic/pkg/email"
	"github.com/skr1ms/mosaic/pkg/jwt"
	"github.com/skr1ms/mosaic/pkg/recaptcha"
	"gorm.io/gorm"
)

type PartnerService struct {
	PartnerRepository *PartnerRepository
	Recaptcha         *recaptcha.Verifier
	JwtService        *jwt.JWT
	MailSender        *email.Mailer
	Config            *config.Config
	Logger            *zerolog.Logger
}

func NewPartnerService(userRepository *PartnerRepository, recaptcha *recaptcha.Verifier, jwt *jwt.JWT, mailSender *email.Mailer, config *config.Config, logger *zerolog.Logger) *PartnerService {
	return &PartnerService{
		PartnerRepository: userRepository,
		Recaptcha:         recaptcha,
		JwtService:        jwt,
		MailSender:        mailSender,
		Config:            config,
		Logger:            logger,
	}
}

func (s *PartnerService) ExportCoupons(partnerID uuid.UUID, status, format string) (string, string, error) {
	// Получаем купоны партнера со статусом "new"
	coupons, err := s.PartnerRepository.GetPartnerCouponsForExport(partnerID, "new")
	if err != nil {
		return "", "", ErrFailedToFetchCoupons
	}

	// Если нет купонов, возвращаем ошибку
	if len(coupons) == 0 {
		return "", "", ErrNoCouponsFound
	}

	// Генерируем содержимое файла
	var content strings.Builder
	filename := fmt.Sprintf("partner_coupons_%s.%s", time.Now().Format("20060102_150405"), format)

	// Если CSV формат, то генерируем CSV
	if format == "csv" {
		content.WriteString("Coupon Code,Partner Status,Coupon Status,Size,Style,Created At\n")
		for _, coupon := range coupons {
			content.WriteString(fmt.Sprintf("%s,%s,%s,%s,%s,%s\n",
				coupon.CouponCode,
				coupon.PartnerStatus,
				coupon.CouponStatus,
				coupon.Size,
				coupon.Style,
				coupon.CreatedAt.Format("2006-01-02 15:04:05"),
			))
		}
	} else {
		// Если TXT формат, то генерируем TXT
		content.WriteString("Partner Coupons Export\n")
		content.WriteString("======================\n\n")
		for _, coupon := range coupons {
			content.WriteString(fmt.Sprintf("Code: %s\n", coupon.CouponCode))
			content.WriteString(fmt.Sprintf("Partner Status: %s\n", coupon.PartnerStatus))
			content.WriteString(fmt.Sprintf("Coupon Status: %s\n", coupon.CouponStatus))
			content.WriteString(fmt.Sprintf("Size: %s\n", coupon.Size))
			content.WriteString(fmt.Sprintf("Style: %s\n", coupon.Style))
			content.WriteString(fmt.Sprintf("Created: %s\n", coupon.CreatedAt.Format("2006-01-02 15:04:05")))
			content.WriteString("---\n")
		}
	}

	return content.String(), filename, nil
}

func (s *PartnerService) UpdatePassword(partnerID uuid.UUID, currentPassword, newPassword string) error {
	// Получаем текущего партнера
	partner, err := s.PartnerRepository.GetByID(partnerID)
	if err != nil {
		return ErrFailedToFindPartnerByID
	}

	// Проверяем текущий пароль
	if !bcrypt.CheckPassword(currentPassword, partner.Password) {
		return ErrCurrentPasswordIsIncorrect
	}

	// Хешируем новый пароль
	hashedPassword, err := bcrypt.HashPassword(newPassword)
	if err != nil {
		return ErrFailedToHashPassword
	}

	// Обновляем пароль
	if err := s.PartnerRepository.UpdatePassword(partnerID, hashedPassword); err != nil {
		return ErrFailedToUpdatePassword
	}

	return nil
}

func (s *PartnerService) ForgotPassword(email string /*captcha string*/) error {
	// Находим партнера по email
	partner, err := s.PartnerRepository.GetByEmail(email)
	if err != nil {
		return ErrPartnerNotFound
	}

	// Проверяем статус партнера
	if partner.Status != "active" {
		return ErrPartnerStatusNotActive
	}

	// Проверяем капчу
	if s.Config.RecaptchaConfig.Environment == "development" {
		s.Logger.Warn().Msg("reCAPTCHA verification is disabled in development mode")
	} else {
		//valid, err := s.Recaptcha.Verify(captcha, "forgot_password")
		//if err != nil || !valid {
		//	return ErrInvalidCaptcha
		//}
	}

	// Проверяем статус партнера
	resetToken, err := s.JwtService.CreatePasswordResetToken(partner.ID, partner.Email)
	if err != nil {
		return ErrFailedToCreateToken
	}

	// Формируем ссылку для сброса пароля
	resetLink := fmt.Sprintf("%s/reset?token=%s",
		s.Config.ServerConfig.FrontendURL,
		resetToken,
	)

	// Отправляем письмо с ссылкой для сброса пароля
	if err := s.MailSender.SendResetPasswordEmail(partner.Email, resetLink); err != nil {
		return ErrFailedToSendEmail
	}

	return nil
}

func (s *PartnerService) ResetPassword(token, newPassword string) error {
	// Валидируем токен сброса пароля
	claims, err := s.JwtService.ValidatePasswordResetToken(token)
	if err != nil {
		return ErrInvalidToken
	}

	// Находим партнера
	partner, err := s.PartnerRepository.GetByEmail(claims.Login) // login == email
	if err != nil {
		return ErrFailedToFindPartnerByEmail
	}

	// Проверяем статус партнера
	if partner.Status != "active" {
		return ErrPartnerStatusNotActive
	}

	// Хешируем новый пароль
	hashedPassword, err := bcrypt.HashPassword(newPassword)
	if err != nil {
		return ErrFailedToHashPassword
	}

	// Обновляем пароль
	if err := s.PartnerRepository.UpdatePassword(claims.UserID, hashedPassword); err != nil {
		return ErrFailedToUpdatePassword
	}

	return nil
}

// DeletePartnerWithCoupons удаляет партнера и все его купоны
func (s *PartnerService) DeletePartnerWithCoupons(partnerID uuid.UUID) error {
	// Начинаем транзакцию
	return s.PartnerRepository.db.Transaction(func(tx *gorm.DB) error {
		// Удаляем все купоны партнера
		if err := tx.Exec("DELETE FROM coupons WHERE partner_id = ?", partnerID).Error; err != nil {
			return err
		}

		// Удаляем самого партнера
		if err := tx.Delete(&Partner{}, partnerID).Error; err != nil {
			return err
		}

		return nil
	})
}
