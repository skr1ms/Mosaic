package validatepartnerdata

import (
	"net"
	"net/url"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

// ValidateDomain валидирует доменное имя
func ValidateDomain(fl validator.FieldLevel) bool {
	domain := fl.Field().String()
	if domain == "" {
		return true // Позволяем пустое значение, required обрабатывается отдельно
	}

	// Проверяем, что домен не содержит протокол
	if strings.HasPrefix(domain, "http://") || strings.HasPrefix(domain, "https://") {
		return false
	}

	// Убираем порт, если он есть
	host, _, err := net.SplitHostPort(domain)
	if err != nil {
		host = domain // Если порта нет, используем весь домен
	}

	// Проверяем базовый формат домена
	domainRegex := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*$`)
	if !domainRegex.MatchString(host) {
		return false
	}

	// Проверяем, что домен не слишком длинный
	if len(host) > 253 {
		return false
	}

	// Проверяем, что каждая часть домена не слишком длинная
	parts := strings.Split(host, ".")
	for _, part := range parts {
		if len(part) > 63 {
			return false
		}
	}

	// Проверяем, что домен содержит хотя бы одну точку (не является TLD)
	if !strings.Contains(host, ".") {
		return false
	}

	return true
}

// ValidateBusinessEmail валидирует бизнес email (не общие домены)
func ValidateBusinessEmail(fl validator.FieldLevel) bool {
	email := fl.Field().String()
	if email == "" {
		return true // Позволяем пустое значение, required обрабатывается отдельно
	}

	// Базовая проверка email
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return false
	}

	// Общие домены, которые не желательны для бизнеса
	commonDomains := []string{
		"gmail.com", "yahoo.com", "hotmail.com", "outlook.com",
		"mail.ru", "yandex.ru", "rambler.ru", "bk.ru",
		"list.ru", "inbox.ru", "ya.ru",
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}

	domain := strings.ToLower(parts[1])
	for _, commonDomain := range commonDomains {
		if domain == commonDomain {
			return false
		}
	}

	return true
}

// ValidateMarketplaceURL валидирует URL маркетплейса
func ValidateMarketplaceURL(fl validator.FieldLevel) bool {
	urlStr := fl.Field().String()
	if urlStr == "" {
		return true // Позволяем пустое значение
	}

	// Парсим URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	// Проверяем схему
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return false
	}

	// Проверяем, что хост не пустой
	if parsedURL.Host == "" {
		return false
	}

	return true
}

// ValidateOzonURL валидирует ссылку на Ozon
func ValidateOzonURL(fl validator.FieldLevel) bool {
	urlStr := fl.Field().String()
	if urlStr == "" {
		return true // Позволяем пустое значение
	}

	if !ValidateMarketplaceURL(fl) {
		return false
	}

	parsedURL, _ := url.Parse(urlStr)
	host := strings.ToLower(parsedURL.Host)

	return strings.Contains(host, "ozon") || strings.Contains(host, "ozon.ru")
}

// ValidateWildberriesURL валидирует ссылку на Wildberries
func ValidateWildberriesURL(fl validator.FieldLevel) bool {
	urlStr := fl.Field().String()
	if urlStr == "" {
		return true // Позволяем пустое значение
	}

	if !ValidateMarketplaceURL(fl) {
		return false
	}

	parsedURL, _ := url.Parse(urlStr)
	host := strings.ToLower(parsedURL.Host)

	return strings.Contains(host, "wildberries") || strings.Contains(host, "wb.ru")
}

// ValidatePartnerCode валидирует код партнера (4 цифры, 0001-9999)
func ValidatePartnerCode(fl validator.FieldLevel) bool {
	code := fl.Field().String()
	if code == "" {
		return true // Позволяем пустое значение, required обрабатывается отдельно
	}

	// Проверяем формат: 4 цифры
	codeRegex := regexp.MustCompile(`^\d{4}$`)
	if !codeRegex.MatchString(code) {
		return false
	}

	// Проверяем диапазон: 0001-9999 (0000 зарезервирован)
	if code == "0000" {
		return false
	}

	return true
}

// ValidateCouponCode валидирует код купона (12 цифр)
func ValidateCouponCode(fl validator.FieldLevel) bool {
	code := fl.Field().String()
	if code == "" {
		return true // Позволяем пустое значение, required обрабатывается отдельно
	}

	// Проверяем формат: 12 цифр
	codeRegex := regexp.MustCompile(`^\d{12}$`)
	return codeRegex.MatchString(code)
}

// ValidateImageFormat валидирует формат изображения
func ValidateImageFormat(fl validator.FieldLevel) bool {
	filename := fl.Field().String()
	if filename == "" {
		return true // Позволяем пустое значение
	}

	allowedExtensions := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp"}
	filename = strings.ToLower(filename)

	for _, ext := range allowedExtensions {
		if strings.HasSuffix(filename, ext) {
			return true
		}
	}

	return false
}

// ValidateImageSize валидирует размер мозаики
func ValidateImageSize(fl validator.FieldLevel) bool {
	size := fl.Field().String()
	if size == "" {
		return true // Позволяем пустое значение
	}

	// Допустимые размеры согласно ТЗ
	allowedSizes := []string{
		"21x30", // 21×30 см
		"30x40", // 30×40 см
		"40x40", // 40×40 см
		"40x50", // 40×50 см
		"40x60", // 40×60 см
		"50x70", // 50×70 см
	}

	for _, allowedSize := range allowedSizes {
		if size == allowedSize {
			return true
		}
	}

	return false
}

// ValidateImageStyle валидирует стиль обработки изображения
func ValidateImageStyle(fl validator.FieldLevel) bool {
	style := fl.Field().String()
	if style == "" {
		return true // Позволяем пустое значение
	}

	// Допустимые стили согласно ТЗ
	allowedStyles := []string{
		"grayscale",  // оттенки серого
		"skin_tones", // оттенки телесного
		"pop_art",    // поп-арт
		"max_colors", // максимум цветов
	}

	for _, allowedStyle := range allowedStyles {
		if style == allowedStyle {
			return true
		}
	}

	return false
}
