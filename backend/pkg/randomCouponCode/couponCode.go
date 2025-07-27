package randomCouponCode

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strconv"
)

// CouponRepository интерфейс для проверки существования купонов
type CouponRepository interface {
	CodeExists(code string) (bool, error)
}

// GenerateCouponCode генерирует уникальный код купона в формате XXXX-XXXX-XXXX
// partnerCode - 4-значный код партнера (0000 для собственных купонов, 0001+ для партнеров)
func GenerateUniqueCouponCode(partnerCode string, repo CouponRepository) (string, error) {
	// Форматируем код партнера как 4-значную строку
	var code int
	if len(partnerCode) == 4 {
		// Если уже строка из 4 цифр, используем как есть
		prefix := partnerCode
		suffix, err := generateUniqueSuffix(prefix, repo)
		if err != nil {
			return "", err
		}
		return prefix + suffix, nil
	}

	// Если число, форматируем как 4-значную строку
	if _, err := strconv.Atoi(partnerCode); err != nil {
		code = 0 // Для собственных купонов
	}
	prefix := fmt.Sprintf("%04d", code)
	suffix, err := generateUniqueSuffix(prefix, repo)
	if err != nil {
		return "", err
	}
	return prefix + suffix, nil
}

// GenerateCouponCode - обратная совместимость для собственных купонов (партнер код 0000)
func GenerateCouponCode() string {
	prefix := "0000"
	suffix := generateRandomDigits(8)
	return prefix + suffix
}

// GenerateCouponCodeWithPartner - обратная совместимость, генерирует код без проверки уникальности
// Deprecated: используйте GenerateUniqueCouponCode с репозиторием
func GenerateCouponCodeWithPartner(partnerCode string) string {
	var code int
	if len(partnerCode) == 4 {
		prefix := partnerCode
		suffix := generateRandomDigits(8)
		return prefix + suffix
	}

	if _, err := strconv.Atoi(partnerCode); err != nil {
		code = 0
	}
	prefix := fmt.Sprintf("%04d", code)
	suffix := generateRandomDigits(8)
	return prefix + suffix
}

// generateUniqueSuffix генерирует уникальный 8-значный суффикс для данного префикса
func generateUniqueSuffix(prefix string, repo CouponRepository) (string, error) {
	maxAttempts := 1000 // Максимальное количество попыток

	for attempt := 0; attempt < maxAttempts; attempt++ {
		suffix := generateRandomDigits(8)
		fullCode := prefix + suffix

		// Проверяем, существует ли такой код
		exists, err := repo.CodeExists(fullCode)
		if err != nil {
			return "", fmt.Errorf("ошибка проверки существования кода: %v", err)
		}

		if !exists {
			return suffix, nil
		}
	}

	return "", fmt.Errorf("не удалось сгенерировать уникальный код после %d попыток", maxAttempts)
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

