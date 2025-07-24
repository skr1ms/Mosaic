package validatepartnerdata

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

// ValidateInternationalPhone проверяет международный номер телефона без ограничений по стране
func ValidateInternationalPhone(fl validator.FieldLevel) bool {
	phone := fl.Field().String()

	cleanPhone := regexp.MustCompile(`[^\d+]`).ReplaceAllString(phone, "")

	var isValidFormat bool

	if cleanPhone[0] == '+' {
		if len(cleanPhone) >= 8 && len(cleanPhone) <= 16 {
			// Проверяем, что после + идут только цифры
			matched, _ := regexp.MatchString(`^\+[1-9]\d{6,14}$`, cleanPhone)
			isValidFormat = matched
		}
	} else {
		if len(cleanPhone) >= 7 && len(cleanPhone) <= 15 {
			matched, _ := regexp.MatchString(`^[1-9]\d{6,14}$`, cleanPhone)
			isValidFormat = matched
		}
	}

	// Дополнительная проверка: оригинальный номер должен содержать только допустимые символы
	validChars, _ := regexp.MatchString(`^[\+\d\s\-\(\)\.]+$`, phone)

	return isValidFormat && validChars
}
