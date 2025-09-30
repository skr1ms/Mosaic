package validatedata

import (
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

func ValidateWhatsappLink(fl validator.FieldLevel) bool {
	link := fl.Field().String()

	if link == "" {
		return true
	}

	patterns := []string{
		`^https://wa\.me/\d{10,15}$`,
		`^https://api\.whatsapp\.com/send\?phone=\d{10,15}$`,
		`^https://chat\.whatsapp\.com/[a-zA-Z0-9]{20,30}$`,
		`^\+\d{10,15}$`,
	}

	for _, pattern := range patterns {
		matched, _ := regexp.MatchString(pattern, link)
		if matched {
			return true
		}
	}

	return false
}

func ProcessWhatsappLink(whatsapp string) string {
	w := strings.TrimSpace(whatsapp)
	whatsappLink := ""
	if w != "" {
		if strings.HasPrefix(w, "http://") || strings.HasPrefix(w, "https://") {
			whatsappLink = w
		} else {
			digitsOnly := strings.Map(func(r rune) rune {
				if r >= '0' && r <= '9' {
					return r
				}
				return -1
			}, w)
			whatsappLink = "https://wa.me/" + digitsOnly
		}
	}

	return whatsappLink
}
