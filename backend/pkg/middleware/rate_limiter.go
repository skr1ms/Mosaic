package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/rs/zerolog/log"
	"github.com/skr1ms/mosaic/pkg/errors"
)

type RateLimiterConfig struct {
	Max        int           // Максимальное количество запросов
	Expiration time.Duration // Время окна
	Message    string        // Сообщение при превышении лимита
}

// GeneralRateLimiter общий лимитер для всех эндпоинтов
func GeneralRateLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        100, // 100 запросов
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			// Используем IP адрес как ключ
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			log.Warn().
				Str("ip", c.IP()).
				Str("path", c.Path()).
				Str("method", c.Method()).
				Msg("Rate limit exceeded")

			return errors.SendError(c, errors.RateLimitError("Too many requests. Please try again later."))
		},
	})
}

// AuthRateLimiter строгий лимитер для эндпоинтов авторизации
func AuthRateLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        5, // 5 попыток
		Expiration: 5 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			// Используем IP адрес как ключ
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			log.Warn().
				Str("ip", c.IP()).
				Str("path", c.Path()).
				Str("method", c.Method()).
				Msg("Auth rate limit exceeded")

			return errors.SendError(c, errors.RateLimitError("Too many authentication attempts. Please try again in 5 minutes."))
		},
	})
}

// PaymentRateLimiter лимитер для платежных эндпоинтов
func PaymentRateLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        10, // 10 попыток
		Expiration: 10 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			// Используем IP адрес как ключ
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			log.Warn().
				Str("ip", c.IP()).
				Str("path", c.Path()).
				Str("method", c.Method()).
				Msg("Payment rate limit exceeded")

			return errors.SendError(c, errors.RateLimitError("Too many payment attempts. Please try again in 10 minutes."))
		},
	})
}

// ImageUploadRateLimiter лимитер для загрузки изображений
func ImageUploadRateLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        20, // 20 загрузок
		Expiration: 1 * time.Hour,
		KeyGenerator: func(c *fiber.Ctx) string {
			// Используем IP адрес как ключ
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			log.Warn().
				Str("ip", c.IP()).
				Str("path", c.Path()).
				Str("method", c.Method()).
				Msg("Image upload rate limit exceeded")

			return errors.SendError(c, errors.RateLimitError("Too many image uploads. Please try again in 1 hour."))
		},
	})
}

// PublicAPIRateLimiter лимитер для публичных API эндпоинтов
func PublicAPIRateLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        50, // 50 запросов
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			// Используем IP адрес как ключ
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			log.Warn().
				Str("ip", c.IP()).
				Str("path", c.Path()).
				Str("method", c.Method()).
				Msg("Public API rate limit exceeded")

			return errors.SendError(c, errors.RateLimitError("Too many requests to public API. Please try again later."))
		},
	})
}
