package randomCouponCode

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strconv"

	"github.com/skr1ms/mosaic/pkg/middleware"
)

var logger = middleware.NewLogger()

type CouponRepository interface {
	CodeExists(ctx context.Context, code string) (bool, error)
}

// GenerateCouponCode generates a unique coupon code in the XXXX-XXXX-XXXX format.
// partnerCode - 4-digit partner code (0000 for own coupons, 0001+ for partners)
func GenerateUniqueCouponCode(partnerCode string, repo CouponRepository) (string, error) {
	var prefix string

	if len(partnerCode) == 4 {
		prefix = partnerCode
	} else {
		code := 0
		if parsedCode, err := strconv.Atoi(partnerCode); err == nil {
			code = parsedCode
		}
		prefix = fmt.Sprintf("%04d", code)
	}

	suffix, err := generateUniqueSuffix(prefix, repo)
	if err != nil {
		logger.GetZerologLogger().Error().
			Err(err).
			Str("partner_code", partnerCode).
			Msg("Failed to generate unique coupon code")
		return "", fmt.Errorf("error generating unique coupon code: %w", err)
	}

	fullCode := prefix + suffix
	return fullCode, nil
}

// GenerateCouponCode backward compatibility for own coupons (partner code 0000)
func GenerateCouponCode() string {
	prefix := "0000"
	suffix := generateRandomDigits(8)
	fullCode := prefix + suffix
	return fullCode
}

// GenerateCouponCodeWithPartner backward compatibility, generates code without checking uniqueness
func GenerateCouponCodeWithPartner(partnerCode string) string {
	var prefix string

	if len(partnerCode) == 4 {
		prefix = partnerCode
	} else {
		code := 0
		if parsedCode, err := strconv.Atoi(partnerCode); err == nil {
			code = parsedCode
		}
		prefix = fmt.Sprintf("%04d", code)
	}

	suffix := generateRandomDigits(8)
	fullCode := prefix + suffix
	return fullCode
}

// generateUniqueSuffix generates a unique 8-digit suffix for a given prefix.
func generateUniqueSuffix(prefix string, repo CouponRepository) (string, error) {
	maxAttempts := 1000

	for attempt := 0; attempt < maxAttempts; attempt++ {
		suffix := generateRandomDigits(8)
		fullCodePlain := prefix + suffix
		exists, err := repo.CodeExists(context.Background(), fullCodePlain)
		if err != nil {
			logger.GetZerologLogger().Error().
				Err(err).
				Str("prefix", prefix).
				Str("full_code", fullCodePlain).
				Int("attempt", attempt).
				Msg("Failed to check code existence")
			return "", fmt.Errorf("error checking code existence: %w", err)
		}

		if !exists {
			return suffix, nil
		}
	}

	logger.GetZerologLogger().Error().
		Str("prefix", prefix).
		Int("max_attempts", maxAttempts).
		Msg("Failed to generate unique code after maximum attempts")
	return "", fmt.Errorf("failed to generate unique code after %d attempts", maxAttempts)
}

// generateRandomDigits generates a string of random digits of a specified length.
func generateRandomDigits(length int) string {
	result := ""
	for i := 0; i < length; i++ {
		n, _ := rand.Int(rand.Reader, big.NewInt(10))
		result += fmt.Sprintf("%d", n.Int64())
	}
	return result
}

// formatCouponCode formats a 12-digit code in XXXX-XXXX-XXXX format
func FormatCouponCode(code string) string {
	if len(code) != 12 {
		return code
	}
	return code[0:4] + "-" + code[4:8] + "-" + code[8:12]
}
