package validatedata

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

// ValidateInternationalPhone validates international phone number without country restrictions
func ValidateInternationalPhone(fl validator.FieldLevel) bool {
	phone := fl.Field().String()

	cleanPhone := regexp.MustCompile(`[^\d+]`).ReplaceAllString(phone, "")

	var isValidFormat bool

	if cleanPhone[0] == '+' {
		if len(cleanPhone) >= 8 && len(cleanPhone) <= 16 {
			matched, _ := regexp.MatchString(`^\+[1-9]\d{6,14}$`, cleanPhone)
			isValidFormat = matched
		}
	} else {
		if len(cleanPhone) >= 7 && len(cleanPhone) <= 15 {
			matched, _ := regexp.MatchString(`^[1-9]\d{6,14}$`, cleanPhone)
			isValidFormat = matched
		}
	}

	validChars, _ := regexp.MatchString(`^[\+\d\s\-\(\)\.]+$`, phone)

	return isValidFormat && validChars
}
