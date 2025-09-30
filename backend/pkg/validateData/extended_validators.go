package validatedata

import (
	"net"
	"net/url"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

// ValidateDomain validates domain name
func ValidateDomain(fl validator.FieldLevel) bool {
	domain := fl.Field().String()
	if domain == "" {
		return true
	}

	if strings.HasPrefix(domain, "http://") || strings.HasPrefix(domain, "https://") {
		return false
	}

	host, _, err := net.SplitHostPort(domain)
	if err != nil {
		host = domain
	}

	domainRegex := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*\.[a-zA-Z0-9]{1,}$`)
	if !domainRegex.MatchString(host) {
		return false
	}

	if len(host) > 253 {
		return false
	}

	parts := strings.Split(host, ".")
	for _, part := range parts {
		if len(part) > 63 {
			return false
		}
	}

	return strings.Contains(host, ".")
}

// ValidateBusinessEmail validates business email (not common domains)
func ValidateBusinessEmail(fl validator.FieldLevel) bool {
	email := fl.Field().String()
	if email == "" {
		return true
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return false
	}

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

// ValidateMarketplaceURL validates marketplace URL
func ValidateMarketplaceURL(fl validator.FieldLevel) bool {
	urlStr := fl.Field().String()
	if urlStr == "" {
		return true
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return false
	}

	if parsedURL.Host == "" {
		return false
	}

	return true
}

// ValidateOzonURL validates the link to Ozon
func ValidateOzonURL(fl validator.FieldLevel) bool {
	urlStr := fl.Field().String()
	if urlStr == "" {
		return true
	}

	if !ValidateMarketplaceURL(fl) {
		return false
	}

	parsedURL, _ := url.Parse(urlStr)
	host := strings.ToLower(parsedURL.Host)

	return strings.Contains(host, "ozon") || strings.Contains(host, "ozon.ru")
}

// ValidateWildberriesURL validates the link to Wildberries
func ValidateWildberriesURL(fl validator.FieldLevel) bool {
	urlStr := fl.Field().String()
	if urlStr == "" {
		return true
	}

	if !ValidateMarketplaceURL(fl) {
		return false
	}

	parsedURL, _ := url.Parse(urlStr)
	host := strings.ToLower(parsedURL.Host)

	return strings.Contains(host, "wildberries") || strings.Contains(host, "wb.ru")
}

// ValidatePartnerCode validates the partner's code (4 digits, 0001-9999)
func ValidatePartnerCode(fl validator.FieldLevel) bool {
	code := fl.Field().String()
	if code == "" {
		return true // Allowing an empty value, required is handled separately
	}

	codeRegex := regexp.MustCompile(`^\d{4}$`)
	if !codeRegex.MatchString(code) {
		return false
	}

	if code == "0000" {
		return false
	}

	return true
}

// ValidateCouponCode validates the coupon code (12 digits)
func ValidateCouponCode(fl validator.FieldLevel) bool {
	code := fl.Field().String()
	if code == "" {
		return true
	}

	codeRegex := regexp.MustCompile(`^\d{12}$`)
	return codeRegex.MatchString(code)
}

// ValidateImageFormat validates the image format
func ValidateImageFormat(fl validator.FieldLevel) bool {
	filename := fl.Field().String()
	if filename == "" {
		return true
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

// ValidateImageSize validates the size of the mosaic
func ValidateImageSize(fl validator.FieldLevel) bool {
	size := fl.Field().String()
	if size == "" {
		return true
	}

	allowedSizes := []string{
		"21x30",
		"30x40",
		"40x40",
		"40x50",
		"40x60",
		"50x70",
	}

	for _, allowedSize := range allowedSizes {
		if size == allowedSize {
			return true
		}
	}

	return false
}

// ValidateImageStyle validates the image processing style
func ValidateImageStyle(fl validator.FieldLevel) bool {
	style := fl.Field().String()
	if style == "" {
		return true
	}

	allowedStyles := []string{
		"grayscale",
		"skin_tones",
		"pop_art",
		"max_colors",
	}

	for _, allowedStyle := range allowedStyles {
		if style == allowedStyle {
			return true
		}
	}

	return false
}

// ValidateHexColor validates the HEX color code in the #RRGGBB or #RGB format (case is not important)
func ValidateHexColor(fl validator.FieldLevel) bool {
	value := strings.TrimSpace(fl.Field().String())
	if value == "" {
		return true
	}
	re := regexp.MustCompile(`^#(?:[0-9a-fA-F]{3}){1,2}$`)
	return re.MatchString(value)
}
