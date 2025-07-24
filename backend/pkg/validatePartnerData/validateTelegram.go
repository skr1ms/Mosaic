package validatepartnerdata

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

// ValidateTelegramLink проверяет ссылку на Telegram
func ValidateTelegramLink(fl validator.FieldLevel) bool {
	link := fl.Field().String()

	if link == "" {
		return true // omitempty будет обрабатывать пустые значения
	}

	// Паттерны для Telegram: @username, https://t.me/username, t.me/username
	patterns := []string{
		`^@[a-zA-Z0-9_]{5,32}$`,                     // @username
		`^https://t\.me/[a-zA-Z0-9_]{5,32}$`,        // https://t.me/username
		`^t\.me/[a-zA-Z0-9_]{5,32}$`,                // t.me/username
		`^https://telegram\.me/[a-zA-Z0-9_]{5,32}$`, // https://telegram.me/username
		`^telegram\.me/[a-zA-Z0-9_]{5,32}$`,         // telegram.me/username
	}

	for _, pattern := range patterns {
		matched, _ := regexp.MatchString(pattern, link)
		if matched {
			return true
		}
	}

	return false
}
