package validatepartnerdata

import (
	"unicode"

	"github.com/go-playground/validator/v10"
)

// ValidatePassword проверяет пароль на соответствие требованиям:
// - минимум 8 символов
// - наличие заглавной буквы
// - наличие специального символа
// - наличие цифры
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
