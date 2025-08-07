package middleware

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
	"github.com/skr1ms/mosaic/pkg/errors"
	validatepartnerdata "github.com/skr1ms/mosaic/pkg/validatePartnerData"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
	validatepartnerdata.RegisterCustomValidators(validate)
}

// ValidateStruct валидирует структуру и возвращает ошибки валидации
func ValidateStruct(s interface{}) error {
	return validate.Struct(s)
}

// ValidationMiddleware middleware для автоматической валидации JSON payload с асинхронным логированием
func ValidationMiddleware(structType interface{}) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		payload := structType

		// Парсим JSON
		if err := c.BodyParser(payload); err != nil {
			// Асинхронно логируем ошибку парсинга
			go func() {
				logValidationError(c.IP(), c.Get("User-Agent"), c.Path(), "json_parse_error", err.Error(), time.Since(start))
			}()

			return errors.SendError(c, errors.BadRequestError("Invalid JSON format: "+err.Error()))
		}

		// Валидируем структуру
		if err := ValidateStruct(payload); err != nil {
			validationErrors := make([]errors.ValidationFieldError, 0)

			if validationErrs, ok := err.(validator.ValidationErrors); ok {
				for _, validationErr := range validationErrs {
					validationErrors = append(validationErrors, errors.CreateValidationFieldError(
						validationErr.Field(),
						validationErr.Tag(),
						validationErr.Value(),
						getErrorMessage(validationErr),
					))
				}
			}

			// Асинхронно логируем ошибки валидации
			go func() {
				logValidationError(c.IP(), c.Get("User-Agent"), c.Path(), "validation_error", err.Error(), time.Since(start))
			}()

			return errors.SendError(c, errors.ValidationErrorWithFields(validationErrors))
		}

		// Асинхронно логируем успешную валидацию
		go func() {
			logValidationSuccess(c.IP(), c.Get("User-Agent"), c.Path(), time.Since(start))
		}()

		c.Locals("validatedData", payload)
		return c.Next()
	}
}

// getErrorMessage возвращает сообщение об ошибке
func getErrorMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return "Это поле обязательно для заполнения"
	case "email":
		return "Неверный формат email"
	case "url":
		return "Неверный формат URL"
	case "min":
		return "Значение слишком короткое"
	case "max":
		return "Значение слишком длинное"
	case "secure_password":
		return "Пароль должен содержать минимум 8 символов, заглавную букву, цифру и специальный символ"
	case "secure_login":
		return "Логин должен содержать минимум 5 символов и не содержать специальных символов"
	case "international_phone":
		return "Неверный формат номера телефона"
	case "telegram_link":
		return "Неверный формат ссылки на Telegram"
	case "whatsapp_link":
		return "Неверный формат ссылки на WhatsApp"
	case "domain":
		return "Неверный формат доменного имени"
	case "business_email":
		return "Укажите корпоративный email (не gmail, mail.ru и т.д.)"
	case "marketplace_url":
		return "Неверный формат ссылки на маркетплейс"
	case "ozon_url":
		return "Ссылка должна вести на сайт Ozon"
	case "wildberries_url":
		return "Ссылка должна вести на сайт Wildberries"
	case "partner_code":
		return "Код партнера должен быть 4-значным числом от 0001 до 9999"
	case "coupon_code":
		return "Код купона должен содержать 12 цифр"
	case "image_format":
		return "Поддерживаемые форматы: JPG, PNG, GIF, BMP, WebP"
	case "image_size":
		return "Неподдерживаемый размер мозаики"
	case "image_style":
		return "Неподдерживаемый стиль обработки"
	case "oneof":
		return "Недопустимое значение"
	default:
		return "Неверное значение поля"
	}
}

// logValidationError асинхронно логирует ошибки валидации
func logValidationError(ip, userAgent, path, errorType, errorMsg string, duration time.Duration) {
	log.Warn().
		Str("ip", ip).
		Str("user_agent", userAgent).
		Str("path", path).
		Str("error_type", errorType).
		Str("error_message", errorMsg).
		Dur("duration", duration).
		Str("event_type", "validation_error").
		Msg("Validation failed")
}

// logValidationSuccess асинхронно логирует успешную валидацию
func logValidationSuccess(ip, userAgent, path string, duration time.Duration) {
	log.Debug().
		Str("ip", ip).
		Str("user_agent", userAgent).
		Str("path", path).
		Dur("duration", duration).
		Str("event_type", "validation_success").
		Msg("Validation successful")
}
