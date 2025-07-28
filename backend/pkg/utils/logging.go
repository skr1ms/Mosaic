package utils

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
)

// LogAndReturnError логирует ошибку с контекстом и возвращает HTTP ответ
func LogAndReturnError(c *fiber.Ctx, err error, message string, statusCode int, fields map[string]interface{}) error {
	log := zerolog.Ctx(c.UserContext())

	event := log.Error().
		Err(err).
		Str("method", c.Method()).
		Str("path", c.Path()).
		Str("ip", c.IP()).
		Str("user_agent", c.Get("User-Agent"))

	for key, value := range fields {
		switch v := value.(type) {
		case string:
			event = event.Str(key, v)
		case int:
			event = event.Int(key, v)
		case bool:
			event = event.Bool(key, v)
		default:
			event = event.Interface(key, v)
		}
	}

	event.Msg(message)

	return c.Status(statusCode).JSON(fiber.Map{
		"error": message,
	})
}

// LogSuccess логирует успешное выполнение операции
func LogSuccess(c *fiber.Ctx, message string, fields map[string]interface{}) {
	log := zerolog.Ctx(c.UserContext())

	event := log.Info().
		Str("method", c.Method()).
		Str("path", c.Path()).
		Str("ip", c.IP())

	for key, value := range fields {
		switch v := value.(type) {
		case string:
			event = event.Str(key, v)
		case int:
			event = event.Int(key, v)
		case bool:
			event = event.Bool(key, v)
		default:
			event = event.Interface(key, v)
		}
	}

	event.Msg(message)
}
