package validatedata

import (
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

// ValidateTelegramLink validates Telegram link
func ValidateTelegramLink(fl validator.FieldLevel) bool {
	link := fl.Field().String()

	if link == "" {
		return true
	}

	patterns := []string{
		`^@[a-zA-Z0-9_]{5,32}$`,
		`^https://t\.me/[a-zA-Z0-9_]{5,32}$`,
		`^t\.me/[a-zA-Z0-9_]{5,32}$`,
		`^https://telegram\.me/[a-zA-Z0-9_]{5,32}$`,
		`^telegram\.me/[a-zA-Z0-9_]{5,32}$`,
	}

	for _, pattern := range patterns {
		matched, _ := regexp.MatchString(pattern, link)
		if matched {
			return true
		}
	}

	return false
}

// ProcessTeleramLink processes social media links
func ProcessTeleramLink(telegram string) string {
	t := strings.TrimSpace(telegram)
	telegramLink := ""
	if t != "" {
		if strings.HasPrefix(t, "http://") || strings.HasPrefix(t, "https://") {
			telegramLink = t
		} else {
			if strings.HasPrefix(t, "@") && len(t) > 1 {
				t = t[1:]
			}
			telegramLink = "https://t.me/" + t
		}
	}

	return telegramLink
}
