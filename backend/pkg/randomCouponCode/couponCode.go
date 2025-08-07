package randomCouponCode

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strconv"
)

type CouponRepository interface {
	CodeExists(ctx context.Context, code string) (bool, error)
}

// GenerateCouponCode генерирует уникальный код купона в формате XXXX-XXXX-XXXX
// partnerCode - 4-значный код партнера (0000 для собственных купонов, 0001+ для партнеров)
func GenerateUniqueCouponCode(partnerCode string, repo CouponRepository) (string, error) {
	// Форматируем код партнера как 4-значную строку
	var prefix string

	if len(partnerCode) == 4 {
		// Если уже строка из 4 цифр, используем как есть
		prefix = partnerCode
	} else {
		// Если число, форматируем как 4-значную строку
		code := 0
		if parsedCode, err := strconv.Atoi(partnerCode); err == nil {
			code = parsedCode
		}
		prefix = fmt.Sprintf("%04d", code)
	}

	suffix, err := generateUniqueSuffix(prefix, repo)
	if err != nil {
		return "", fmt.Errorf("error generating unique coupon code: %w", err)
	}

	// Форматируем как XXXX-XXXX-XXXX
	fullCode := prefix + suffix
	return formatCouponCode(fullCode), nil
}

// GenerateCouponCode - обратная совместимость для собственных купонов (партнер код 0000)
func GenerateCouponCode() string {
	prefix := "0000"
	suffix := generateRandomDigits(8)
	fullCode := prefix + suffix
	return formatCouponCode(fullCode)
}

// GenerateCouponCodeWithPartner - обратная совместимость, генерирует код без проверки уникальности
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
	return formatCouponCode(fullCode)
}

// generateUniqueSuffix генерирует уникальный 8-значный суффикс для данного префикса
func generateUniqueSuffix(prefix string, repo CouponRepository) (string, error) {
	maxAttempts := 1000 // Максимальное количество попыток

	for attempt := 0; attempt < maxAttempts; attempt++ {
		suffix := generateRandomDigits(8)
		fullCodePlain := prefix + suffix
		fullCodeFormatted := formatCouponCode(fullCodePlain)

		// Проверяем, существует ли такой код (проверяем отформатированную версию)
		exists, err := repo.CodeExists(context.Background(), fullCodeFormatted)
		if err != nil {
			return "", fmt.Errorf("error checking code existence: %w", err)
		}

		if !exists {
			return suffix, nil
		}
	}

	return "", fmt.Errorf("failed to generate unique code after %d attempts", maxAttempts)
}

// generateRandomDigits генерирует строку из случайных цифр заданной длины
func generateRandomDigits(length int) string {
	result := ""
	for i := 0; i < length; i++ {
		n, _ := rand.Int(rand.Reader, big.NewInt(10))
		result += fmt.Sprintf("%d", n.Int64())
	}
	return result
}

// formatCouponCode форматирует 12-значный код в формат XXXX-XXXX-XXXX
func formatCouponCode(code string) string {
	if len(code) != 12 {
		return code // Возвращаем как есть, если длина неправильная
	}
	return code[0:4] + "-" + code[4:8] + "-" + code[8:12]
}
