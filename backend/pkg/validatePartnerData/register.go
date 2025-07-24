package validatepartnerdata

import "github.com/go-playground/validator/v10"

// RegisterCustomValidators регистрирует все кастомные валидаторы
func RegisterCustomValidators(v *validator.Validate) {
	v.RegisterValidation("secure_password", ValidatePassword)
	v.RegisterValidation("secure_login", ValidateLogin)
	v.RegisterValidation("russian_phone", ValidateRussianPhone)
	v.RegisterValidation("telegram_link", ValidateTelegramLink)
	v.RegisterValidation("whatsapp_link", ValidateWhatsappLink)
}
