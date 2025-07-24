package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// generateCouponCode генерирует код купона в формате XXXX-XXXX-XXXX
func GenerateCouponCode(partnerCode int16) string {
	prefix := fmt.Sprintf("%04d", partnerCode)
	suffix := GenerateRandomDigits(8)
	fullCode := prefix + suffix

	return fullCode
}

// generateRandomDigits генерирует строку из случайных цифр заданной длины
func GenerateRandomDigits(length int) string {
	result := ""
	for i := 0; i < length; i++ {
		n, _ := rand.Int(rand.Reader, big.NewInt(10))
		result += fmt.Sprintf("%d", n.Int64())
	}
	return result
}
