package validatepartnerdata

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

// ValidateWhatsappLink проверяет ссылку на WhatsApp
func ValidateWhatsappLink(fl validator.FieldLevel) bool {
	link := fl.Field().String()

	if link == "" {
		return true // omitempty будет обрабатывать пустые значения
	}

	// Паттерны для WhatsApp: https://wa.me/7XXXXXXXXXX, https://api.whatsapp.com/send?phone=7XXXXXXXXXX
	patterns := []string{
		`^https://wa\.me/\d{10,15}$`,                         // https://wa.me/7XXXXXXXXXX
		`^https://api\.whatsapp\.com/send\?phone=\d{10,15}$`, // https://api.whatsapp.com/send?phone=7XXXXXXXXXX
		`^https://chat\.whatsapp\.com/[a-zA-Z0-9]{20,30}$`,   // https://chat.whatsapp.com/grouplink
		`^\+\d{10,15}$`, // +7XXXXXXXXXX
	}

	for _, pattern := range patterns {
		matched, _ := regexp.MatchString(pattern, link)
		if matched {
			return true
		}
	}

	return false
}
