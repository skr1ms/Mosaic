package validatepartnerdata

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

// ValidateRussianPhone проверяет российский номер телефона
func ValidateRussianPhone(fl validator.FieldLevel) bool {
	phone := fl.Field().String()

	// Российские номера: +7XXXXXXXXXX, 8XXXXXXXXXX, 7XXXXXXXXXX
	patterns := []string{
		`^\+7\d{10}$`, // +7XXXXXXXXXX
		`^8\d{10}$`,   // 8XXXXXXXXXX
		`^7\d{10}$`,   // 7XXXXXXXXXX
		`^\+7\s?\(\d{3}\)\s?\d{3}-?\d{2}-?\d{2}$`, // +7 (XXX) XXX-XX-XX
		`^8\s?\(\d{3}\)\s?\d{3}-?\d{2}-?\d{2}$`,   // 8 (XXX) XXX-XX-XX
	}

	for _, pattern := range patterns {
		matched, _ := regexp.MatchString(pattern, phone)
		if matched {
			return true
		}
	}

	return false
}
