package middleware

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/skr1ms/mosaic/pkg/errors"
	validatepartnerdata "github.com/skr1ms/mosaic/pkg/validateData"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
	validatepartnerdata.RegisterCustomValidators(validate)
}

// ValidateStruct validates structure and returns validation errors
func ValidateStruct(s any) error {
	return validate.Struct(s)
}

// ValidationMiddleware middleware for automatic JSON payload validation with async logging
type ValidationMiddleware struct {
	structType any
	logger     *Logger
}

// NewValidationMiddleware creates new validation middleware
func NewValidationMiddleware(structType any, logger *Logger) *ValidationMiddleware {
	middleware := &ValidationMiddleware{
		structType: structType,
		logger:     logger,
	}

	return middleware
}

// ValidateStruct validates structure and returns validation errors
func (vm *ValidationMiddleware) ValidateStruct(s any) error {
	return validate.Struct(s)
}

// ValidationMiddlewareHandler creates middleware for automatic JSON payload validation
func (vm *ValidationMiddleware) ValidationMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		payload := vm.structType

		if err := c.BodyParser(payload); err != nil {
			vm.logValidationErrorAsync(c.IP(), c.Get("User-Agent"), c.Path(), "json_parse_error", err.Error(), time.Since(start))

			return errors.SendError(c, errors.BadRequestError("Invalid JSON format: "+err.Error()))
		}

		if err := vm.ValidateStruct(payload); err != nil {
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

			vm.logValidationErrorAsync(c.IP(), c.Get("User-Agent"), c.Path(), "validation_error", err.Error(), time.Since(start))

			return errors.SendError(c, errors.ValidationErrorWithFields(validationErrors))
		}

		vm.logValidationSuccessAsync(c.IP(), c.Get("User-Agent"), c.Path(), time.Since(start))

		c.Locals("validatedData", payload)
		return c.Next()
	}
}

// getErrorMessage returns error message
func getErrorMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email format"
	case "url":
		return "Invalid URL format"
	case "min":
		return "Value is too short"
	case "max":
		return "Value is too long"
	case "secure_password":
		return "Password must contain at least 8 characters, uppercase letter, number and special character"
	case "secure_login":
		return "Login must contain at least 5 characters and no special characters"
	case "international_phone":
		return "Invalid phone number format"
	case "telegram_link":
		return "Invalid Telegram link format"
	case "whatsapp_link":
		return "Invalid WhatsApp link format"
	case "domain":
		return "Invalid domain name format"
	case "business_email":
		return "Please provide corporate email (not gmail, mail.ru, etc.)"
	case "marketplace_url":
		return "Invalid marketplace link format"
	case "ozon_url":
		return "Link must lead to Ozon website"
	case "wildberries_url":
		return "Link must lead to Wildberries website"
	case "partner_code":
		return "Partner code must be 4-digit number from 0001 to 9999"
	case "coupon_code":
		return "Coupon code must contain 12 digits"
	case "image_format":
		return "Supported formats: JPG, PNG, GIF, BMP, WebP"
	case "image_size":
		return "Unsupported mosaic size"
	case "image_style":
		return "Unsupported processing style"
	case "oneof":
		return "Invalid value"
	default:
		return "Invalid field value"
	}
}

// logValidationError asynchronously logs validation errors
func logValidationError(ip, userAgent, path, errorType, errorMsg string, duration time.Duration, logger *Logger) {
	logger.GetZerologLogger().Warn().
		Str("ip", ip).
		Str("user_agent", userAgent).
		Str("path", path).
		Str("error_type", errorType).
		Str("error_message", errorMsg).
		Dur("duration", duration).
		Str("event_type", "validation_error").
		Msg("Validation failed")
}

// logValidationSuccess asynchronously logs successful validation
func logValidationSuccess(ip, userAgent, path string, duration time.Duration, logger *Logger) {
	logger.GetZerologLogger().Debug().
		Str("ip", ip).
		Str("user_agent", userAgent).
		Str("path", path).
		Dur("duration", duration).
		Str("event_type", "validation_success").
		Msg("Validation successful")
}

// logValidationErrorAsync logs validation error asynchronously
func (vm *ValidationMiddleware) logValidationErrorAsync(ip, userAgent, path, errorType, errorMsg string, duration time.Duration) {
	go func() {
		logValidationError(ip, userAgent, path, errorType, errorMsg, duration, vm.logger)
	}()
}

// logValidationSuccessAsync logs successful validation asynchronously
func (vm *ValidationMiddleware) logValidationSuccessAsync(ip, userAgent, path string, duration time.Duration) {
	go func() {
		logValidationSuccess(ip, userAgent, path, duration, vm.logger)
	}()
}
