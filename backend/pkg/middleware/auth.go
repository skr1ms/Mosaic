package middleware

import (
	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
	"github.com/skr1ms/mosaic/pkg/jwt"
)

// JWTMiddleware создает JWT middleware для проверки access токенов с асинхронным логированием
func JWTMiddleware(jwtService *jwt.JWT) fiber.Handler {
	return jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: jwtService.GetSecretKey()},
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			// Асинхронно логируем неудачные попытки аутентификации
			go func() {
				logAuthFailure(c.IP(), c.Get("User-Agent"), c.Path(), err.Error())
			}()

			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized: Invalid or expired token",
			})
		},
		SuccessHandler: func(c *fiber.Ctx) error {
			token := c.Locals("user")
			if token == nil {
				// Асинхронно логируем отсутствие токена
				go func() {
					logAuthFailure(c.IP(), c.Get("User-Agent"), c.Path(), "Token not found in context")
				}()

				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": "Unauthorized: Token not found",
				})
			}

			// Асинхронно логируем успешную аутентификацию
			go func() {
				logAuthSuccess(c.IP(), c.Get("User-Agent"), c.Path())
			}()

			return c.Next()
		},
	})
}

// AdminOnly middleware проверяет что пользователь - админ
func AdminOnly() fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Locals("user")
		if token == nil {
			// Асинхронно логируем неудачу авторизации
			go func() {
				logAuthorizationFailure(c.IP(), c.Get("User-Agent"), c.Path(), "Token not found", "admin")
			}()

			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized: Token not found",
			})
		}

		claims, err := jwt.GetClaimsFromFiberContext(c)
		if err != nil {
			// Асинхронно логируем ошибку получения claims
			go func() {
				logAuthorizationFailure(c.IP(), c.Get("User-Agent"), c.Path(), err.Error(), "admin")
			}()

			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized: Invalid token claims",
			})
		}

		if claims.Role != "admin" {
			// Асинхронно логируем права
			go func() {
				logAuthorizationFailure(c.IP(), c.Get("User-Agent"), c.Path(), "insufficient_privileges", "admin")
			}()

			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Forbidden: Admin access required",
			})
		}

		// Асинхронно логируем успешную авторизацию
		go func() {
			logAuthorizationSuccess(c.IP(), c.Get("User-Agent"), c.Path(), claims.Role, "admin")
		}()

		return c.Next()
	}
}

// PartnerOnly middleware проверяет что пользователь - партнер
func PartnerOnly() fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Locals("user")
		if token == nil {
			// Асинхронно логируем неудачу авторизации
			go func() {
				logAuthorizationFailure(c.IP(), c.Get("User-Agent"), c.Path(), "Token not found", "partner")
			}()

			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized: Token not found",
			})
		}

		claims, err := jwt.GetClaimsFromFiberContext(c)
		if err != nil {
			// Асинхронно логируем ошибку получения claims
			go func() {
				logAuthorizationFailure(c.IP(), c.Get("User-Agent"), c.Path(), err.Error(), "partner")
			}()

			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized: Invalid token claims",
			})
		}

		if claims.Role != "partner" {
			// Асинхронно логируем права
			go func() {
				logAuthorizationFailure(c.IP(), c.Get("User-Agent"), c.Path(), "insufficient_privileges", "partner")
			}()

			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Forbidden: Partner access required",
			})
		}

		// Асинхронно логируем успешную авторизацию
		go func() {
			logAuthorizationSuccess(c.IP(), c.Get("User-Agent"), c.Path(), claims.Role, "partner")
		}()

		return c.Next()
	}
}

// AdminOrPartner middleware разрешает доступ админам и партнерам
func AdminOrPartner() fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Locals("user")
		if token == nil {
			// Асинхронно логируем неудачу авторизации
			go func() {
				logAuthorizationFailure(c.IP(), c.Get("User-Agent"), c.Path(), "Token not found", "admin_or_partner")
			}()

			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized: Token not found",
			})
		}

		claims, err := jwt.GetClaimsFromFiberContext(c)
		if err != nil {
			// Асинхронно логируем ошибку получения claims
			go func() {
				logAuthorizationFailure(c.IP(), c.Get("User-Agent"), c.Path(), err.Error(), "admin_or_partner")
			}()

			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized: Invalid token claims",
			})
		}

		if claims.Role != "admin" && claims.Role != "partner" {
			// Асинхронно логируем права
			go func() {
				logAuthorizationFailure(c.IP(), c.Get("User-Agent"), c.Path(), "insufficient_privileges", "admin_or_partner")
			}()

			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Forbidden: Admin or Partner access required",
			})
		}

		// Асинхронно логируем успешную авторизацию
		go func() {
			logAuthorizationSuccess(c.IP(), c.Get("User-Agent"), c.Path(), claims.Role, "admin_or_partner")
		}()

		return c.Next()
	}
}

// logAuthFailure асинхронно логирует неудачные попытки аутентификации
func logAuthFailure(ip, userAgent, path, reason string) {
	log.Warn().
		Str("ip", ip).
		Str("user_agent", userAgent).
		Str("path", path).
		Str("reason", reason).
		Str("event_type", "auth_failure").
		Msg("Authentication failed")
}

// logAuthSuccess асинхронно логирует успешную аутентификацию
func logAuthSuccess(ip, userAgent, path string) {
	log.Info().
		Str("ip", ip).
		Str("user_agent", userAgent).
		Str("path", path).
		Str("event_type", "auth_success").
		Msg("Authentication successful")
}

// logAuthorizationFailure асинхронно логирует неудачи авторизации
func logAuthorizationFailure(ip, userAgent, path, reason, requiredRole string) {
	log.Warn().
		Str("ip", ip).
		Str("user_agent", userAgent).
		Str("path", path).
		Str("reason", reason).
		Str("required_role", requiredRole).
		Str("event_type", "authorization_failure").
		Msg("Authorization failed")
}

// logAuthorizationSuccess асинхронно логирует успешную авторизацию
func logAuthorizationSuccess(ip, userAgent, path, userRole, requiredRole string) {
	log.Debug().
		Str("ip", ip).
		Str("user_agent", userAgent).
		Str("path", path).
		Str("user_role", userRole).
		Str("required_role", requiredRole).
		Str("event_type", "authorization_success").
		Msg("Authorization successful")
}
