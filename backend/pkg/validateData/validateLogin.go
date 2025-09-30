package validatedata

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

// ValidateLogin checks login compliance with requirements:
// - at least 5 characters
// - no special characters
func ValidateLogin(fl validator.FieldLevel) bool {
	login := fl.Field().String()

	if len(login) < 5 {
		return false
	}

	matched, _ := regexp.MatchString("^[a-zA-Z0-9_]+$", login)
	return matched
}
