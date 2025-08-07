package validatepartnerdata

import "github.com/go-playground/validator/v10"

// RegisterCustomValidators регистрирует все кастомные валидаторы
func RegisterCustomValidators(v *validator.Validate) {
	// Базовые валидаторы
	v.RegisterValidation("secure_password", ValidatePassword)
	v.RegisterValidation("secure_login", ValidateLogin)
	v.RegisterValidation("international_phone", ValidateInternationalPhone)
	v.RegisterValidation("telegram_link", ValidateTelegramLink)
	v.RegisterValidation("whatsapp_link", ValidateWhatsappLink)

	// Расширенные валидаторы
	v.RegisterValidation("domain", ValidateDomain)
	v.RegisterValidation("business_email", ValidateBusinessEmail)
	v.RegisterValidation("marketplace_url", ValidateMarketplaceURL)
	v.RegisterValidation("ozon_url", ValidateOzonURL)
	v.RegisterValidation("wildberries_url", ValidateWildberriesURL)
	v.RegisterValidation("partner_code", ValidatePartnerCode)
	v.RegisterValidation("coupon_code", ValidateCouponCode)
	v.RegisterValidation("image_format", ValidateImageFormat)
	v.RegisterValidation("image_size", ValidateImageSize)
	v.RegisterValidation("image_style", ValidateImageStyle)
}
