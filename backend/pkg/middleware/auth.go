package middleware

import (
	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/skr1ms/mosaic/pkg/utils"
)

// JWTMiddleware создает JWT middleware для проверки access токенов
func JWTMiddleware(jwtService *utils.JWT) fiber.Handler {
	return jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: jwtService.GetSecretKey()},
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized: Invalid or expired token",
			})
		},
		SuccessHandler: func(c *fiber.Ctx) error {
			token := c.Locals("user")
			if token == nil {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": "Unauthorized: Token not found",
				})
			}
			return c.Next()
		},
	})
}

// AdminOnly middleware проверяет что пользователь - админ
func AdminOnly() fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Locals("user")
		if token == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized: Token not found",
			})
		}

		claims, err := utils.GetClaimsFromFiberContext(c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized: Invalid token claims",
			})
		}

		if claims.Role != "admin" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Forbidden: Admin access required",
			})
		}

		return c.Next()
	}
}

// PartnerOnly middleware проверяет что пользователь - партнер
func PartnerOnly() fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Locals("user")
		if token == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized: Token not found",
			})
		}

		claims, err := utils.GetClaimsFromFiberContext(c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized: Invalid token claims",
			})
		}

		if claims.Role != "partner" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Forbidden: Partner access required",
			})
		}

		return c.Next()
	}
}

// AdminOrPartner middleware разрешает доступ админам и партнерам
func AdminOrPartner() fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Locals("user")
		if token == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized: Token not found",
			})
		}

		claims, err := utils.GetClaimsFromFiberContext(c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized: Invalid token claims",
			})
		}

		if claims.Role != "admin" && claims.Role != "partner" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Forbidden: Admin or Partner access required",
			})
		}

		return c.Next()
	}
}
