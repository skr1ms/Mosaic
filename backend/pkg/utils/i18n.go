package utils

import (
	"github.com/gofiber/fiber/v2"
)

// GetLocalizedMessage получает локализованное сообщение из контекста Fiber
func GetLocalizedMessage(c *fiber.Ctx, key string, fallback string) string {
	i18n := c.Locals("i18n")
	if i18n != nil {
		if translate, ok := i18n.(func(string, ...interface{}) string); ok {
			return translate(key)
		}
	}
	return fallback
}

// LocalizedError возвращает локализованную ошибку в JSON формате
func LocalizedError(c *fiber.Ctx, statusCode int, key string, fallback string) error {
	message := GetLocalizedMessage(c, key, fallback)
	return c.Status(statusCode).JSON(fiber.Map{
		"error": message,
	})
}

// LocalizedSuccess возвращает локализованный успешный ответ
func LocalizedSuccess(c *fiber.Ctx, key string, fallback string, data interface{}) error {
	message := GetLocalizedMessage(c, key, fallback)
	return c.JSON(fiber.Map{
		"message": message,
		"data":    data,
	})
}
