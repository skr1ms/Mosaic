package middleware

import (
	"time"

	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/gofiber/fiber/v2/middleware/limiter"

	"github.com/skr1ms/mosaic/pkg/errors"
)

type RateLimiterConfig struct {
	Max        int           // Maximum number of requests
	Expiration time.Duration // Time window
	Message    string        // Message when limit exceeded
}

// GeneralRateLimiter general limiter for all endpoints
func GeneralRateLimiter(logger *Logger) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        2000,            // increased limit
		Expiration: 5 * time.Minute, // increased timeout
		Next: func(c *fiber.Ctx) bool {
			p := c.Path()
			if p == "/favicon.ico" || strings.HasPrefix(p, "/static/") || strings.HasPrefix(p, "/swagger") {
				return true
			}
			if p == "/api/chat/unread-count" || p == "/api/chat/users" {
				return true
			}
			if strings.HasPrefix(p, "/api/public/attachments") {
				return true
			}
			return false
		},
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			logger.GetZerologLogger().Warn().
				Str("ip", c.IP()).
				Str("path", c.Path()).
				Str("method", c.Method()).
				Msg("Rate limit exceeded")

			return errors.SendError(c, errors.RateLimitError("Too many requests. Please try again later."))
		},
	})
}

// AuthRateLimiter strict limiter for authorization endpoints
func AuthRateLimiter(logger *Logger) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        10,               // increased to 10 attempts
		Expiration: 30 * time.Minute, // increased timeout
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			logger.GetZerologLogger().Warn().
				Str("ip", c.IP()).
				Str("path", c.Path()).
				Str("method", c.Method()).
				Msg("Auth rate limit exceeded")

			return errors.SendError(c, errors.RateLimitError("Too many authentication attempts. Please try again in 10 minutes."))
		},
	})
}

// PaymentRateLimiter limiter for payment endpoints
func PaymentRateLimiter(logger *Logger) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        20,               // increased to 20 attempts
		Expiration: 30 * time.Minute, // increased timeout
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			logger.GetZerologLogger().Warn().
				Str("ip", c.IP()).
				Str("path", c.Path()).
				Str("method", c.Method()).
				Msg("Payment rate limit exceeded")

			return errors.SendError(c, errors.RateLimitError("Too many payment attempts. Please try again in 15 minutes."))
		},
	})
}

// ImageUploadRateLimiter limiter for image uploads
func ImageUploadRateLimiter(logger *Logger) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        50,            // increased to 50 uploads
		Expiration: 4 * time.Hour, // increased timeout
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			logger.GetZerologLogger().Warn().
				Str("ip", c.IP()).
				Str("path", c.Path()).
				Str("method", c.Method()).
				Msg("Image upload rate limit exceeded")

			return errors.SendError(c, errors.RateLimitError("Too many image uploads. Please try again in 2 hours."))
		},
	})
}

// PublicAPIRateLimiter limiter for public API endpoints
func PublicAPIRateLimiter(logger *Logger) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        100,             // increased to 100 requests
		Expiration: 5 * time.Minute, // increased timeout
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			logger.GetZerologLogger().Warn().
				Str("ip", c.IP()).
				Str("path", c.Path()).
				Str("method", c.Method()).
				Msg("Public API rate limit exceeded")

			return errors.SendError(c, errors.RateLimitError("Too many requests to public API. Please try again later."))
		},
	})
}

// CouponActivationRateLimiter strict limiter for coupon activation endpoints
func CouponActivationRateLimiter(logger *Logger) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        10,               // Maximum 10 coupon activation attempts
		Expiration: 60 * time.Minute, // Ban for 60 minutes after limit exceeded
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			logger.GetZerologLogger().Warn().
				Str("ip", c.IP()).
				Str("path", c.Path()).
				Str("method", c.Method()).
				Msg("Coupon activation rate limit exceeded")

			return errors.SendError(c, errors.RateLimitError("Too many coupon activation attempts. Please try again in 30 minutes."))
		},
	})
}
