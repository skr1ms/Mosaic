package validatepartnerdata

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

// ValidateLogin проверяет логин на соответствие требованиям:
// - не менее 5 символов
// - не должно быть специальных символов
func ValidateLogin(fl validator.FieldLevel) bool {
	login := fl.Field().String()

	if len(login) < 5 {
		return false
	}

	// Проверяем, что в логине только буквы, цифры и подчеркивания
	matched, _ := regexp.MatchString("^[a-zA-Z0-9_]+$", login)
	return matched
}
