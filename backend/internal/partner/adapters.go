package partner

import (
	"github.com/google/uuid"
	"github.com/skr1ms/mosaic/internal/coupon"
	"github.com/skr1ms/mosaic/pkg/email"
	"github.com/skr1ms/mosaic/pkg/jwt"
	"github.com/skr1ms/mosaic/pkg/recaptcha"
)

// Адаптеры для приведения конкретных типов к интерфейсам

type PartnerRepositoryAdapter struct {
	*PartnerRepository
}

func (p *PartnerRepositoryAdapter) Adapt() PartnerRepositoryInterface {
	return p.PartnerRepository
}

type CouponServiceAdapter struct {
	*coupon.CouponService
}

func (c *CouponServiceAdapter) ExportCouponsAdvanced(options coupon.ExportOptionsRequest) ([]byte, string, string, error) {
	return c.CouponService.ExportCouponsAdvanced(options)
}

type JWTAdapter struct {
	*jwt.JWT
}

func (j *JWTAdapter) CreatePasswordResetToken(userID uuid.UUID, email string) (string, error) {
	return j.JWT.CreatePasswordResetToken(userID, email)
}

func (j *JWTAdapter) ValidatePasswordResetToken(token string) (*TokenClaims, error) {
	claims, err := j.JWT.ValidatePasswordResetToken(token)
	if err != nil {
		return nil, err
	}
	return &TokenClaims{
		UserID: claims.UserID,
		Login:  claims.Login,
	}, nil
}

type RecaptchaAdapter struct {
	*recaptcha.Verifier
}

func (r *RecaptchaAdapter) Verify(token, action string) (bool, error) {
	return r.Verifier.Verify(token, action)
}

type MailerAdapter struct {
	*email.Mailer
}

func (m *MailerAdapter) SendResetPasswordEmail(email, resetLink string) error {
	return m.Mailer.SendResetPasswordEmail(email, resetLink)
}
