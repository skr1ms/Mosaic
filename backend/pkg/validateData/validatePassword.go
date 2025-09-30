package validatedata

import (
	"unicode"

	"github.com/go-playground/validator/v10"
)

// ValidatePassword checks password compliance with requirements:
// - minimum 8 characters
// - uppercase letter presence
// - special character presence
// - digit presence
func ValidatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	if len(password) < 8 {
		return false
	}

	var hasUpper, hasSpecial, hasDigit bool

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	return hasUpper && hasSpecial && hasDigit
}
