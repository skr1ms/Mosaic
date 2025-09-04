package middleware

import (
	"runtime/debug"

	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"

	"github.com/skr1ms/mosaic/pkg/jwt"
)

// Global logger for all middleware (created ONCE)
var globalLogger = NewLogger()

// JWTMiddleware creates JWT middleware for access token validation with async logging
func JWTMiddleware(jwtService *jwt.JWT, logger *Logger) fiber.Handler {
	return jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: jwtService.GetSecretKey()},
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			ip := c.IP()
			userAgent := c.Get("User-Agent")
			path := c.Path()
			errMsg := err.Error()

			logAuthFailure(ip, userAgent, path, errMsg, logger)

			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized: Invalid or expired token",
			})
		},
		SuccessHandler: func(c *fiber.Ctx) error {
			token := c.Locals("user")
			if token == nil {
				ip := c.IP()
				userAgent := c.Get("User-Agent")
				path := c.Path()

				logAuthFailure(ip, userAgent, path, "Token not found in context", logger)

				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": "Unauthorized: Token not found",
				})
			}

			// Extract claims and store them in context for audit middleware
			if claims, err := jwt.GetClaimsFromFiberContext(c); err == nil && claims != nil {
				c.Locals("jwt_claims", claims)
				// Also store user_role for backward compatibility
				c.Locals("user_role", claims.Role)
				c.Locals("user_id", claims.UserID.String())
			}

			ip := c.IP()
			userAgent := c.Get("User-Agent")
			path := c.Path()

			logAuthSuccess(ip, userAgent, path, logger)

			return c.Next()
		},
	})
}

func AdminOnly() fiber.Handler {
	return func(c *fiber.Ctx) error {
		globalLogger.GetZerologLogger().Info().
			Str("middleware", "AdminOnly").
			Str("path", c.Path()).
			Msg("=== ENTERING AdminOnly MIDDLEWARE ===")

		token := c.Locals("user")
		if token == nil {
			ip := c.IP()
			userAgent := c.Get("User-Agent")
			path := c.Path()

			globalLogger.GetZerologLogger().Error().
				Str("middleware", "AdminOnly").
				Str("path", path).
				Msg("Token not found in AdminOnly")

			logAuthorizationFailure(ip, userAgent, path, "Token not found", "admin", globalLogger)

			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized: Token not found",
			})
		}

		claims, err := jwt.GetClaimsFromFiberContext(c)
		if err != nil {
			ip := c.IP()
			userAgent := c.Get("User-Agent")
			path := c.Path()
			errMsg := err.Error()

			globalLogger.GetZerologLogger().Error().
				Str("middleware", "AdminOnly").
				Str("path", path).
				Str("error", errMsg).
				Msg("Failed to get claims in AdminOnly")

			logAuthorizationFailure(ip, userAgent, path, errMsg, "admin", globalLogger)

			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized: Invalid token claims",
			})
		}

		if claims.Role != "admin" {
			ip := c.IP()
			userAgent := c.Get("User-Agent")
			path := c.Path()

			globalLogger.GetZerologLogger().Error().
				Str("middleware", "AdminOnly").
				Str("path", path).
				Str("user_role", claims.Role).
				Msg("Insufficient permissions in AdminOnly")

			logAuthorizationFailure(ip, userAgent, path, "Insufficient permissions", "admin", globalLogger)

			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Forbidden: Admin access required",
			})
		}

		globalLogger.GetZerologLogger().Info().
			Str("middleware", "AdminOnly").
			Str("path", c.Path()).
			Str("user_id", claims.UserID.String()).
			Str("user_role", claims.Role).
			Msg("=== EXITING AdminOnly MIDDLEWARE ===")

		return c.Next()
	}
}

func MainAdminOnly() fiber.Handler {
	return func(c *fiber.Ctx) error {
		globalLogger.GetZerologLogger().Info().
			Str("middleware", "MainAdminOnly").
			Str("path", c.Path()).
			Msg("=== ENTERING MainAdminOnly MIDDLEWARE ===")

		token := c.Locals("user")
		if token == nil {
			ip := c.IP()
			userAgent := c.Get("User-Agent")
			path := c.Path()

			globalLogger.GetZerologLogger().Error().
				Str("middleware", "MainAdminOnly").
				Str("path", path).
				Msg("Token not found in MainAdminOnly")

			logAuthorizationFailure(ip, userAgent, path, "Token not found", "main_admin", globalLogger)

			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized: Token not found",
			})
		}

		claims, err := jwt.GetClaimsFromFiberContext(c)
		if err != nil {
			ip := c.IP()
			userAgent := c.Get("User-Agent")
			path := c.Path()
			errMsg := err.Error()

			globalLogger.GetZerologLogger().Error().
				Str("middleware", "MainAdminOnly").
				Str("path", path).
				Str("error", errMsg).
				Msg("Failed to get claims in MainAdminOnly")

			logAuthorizationFailure(ip, userAgent, path, errMsg, "main_admin", globalLogger)

			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized: Invalid token claims",
			})
		}

		globalLogger.GetZerologLogger().Info().
			Str("middleware", "MainAdminOnly").
			Str("role", claims.Role).
			Str("path", c.Path()).
			Msg("Checking role in MainAdminOnly middleware")

		if claims.Role != "main_admin" {
			globalLogger.GetZerologLogger().Error().
				Str("middleware", "MainAdminOnly").
				Str("role", claims.Role).
				Str("path", c.Path()).
				Msg("ACCESS DENIED in MainAdminOnly")

			ip := c.IP()
			userAgent := c.Get("User-Agent")
			path := c.Path()

			logAuthorizationFailure(ip, userAgent, path, "insufficient_privileges", "main_admin", globalLogger)

			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Forbidden: Main admin access required",
			})
		}

		globalLogger.GetZerologLogger().Info().
			Str("middleware", "MainAdminOnly").
			Str("role", claims.Role).
			Str("path", c.Path()).
			Msg("ACCESS GRANTED in MainAdminOnly")

		return c.Next()
	}
}

// AdminOrMainAdmin middleware allows access to both admin and main_admin roles
func AdminOrMainAdmin() fiber.Handler {
	return func(c *fiber.Ctx) error {
		globalLogger.GetZerologLogger().Info().
			Str("middleware", "AdminOrMainAdmin").
			Str("path", c.Path()).
			Msg("=== ENTERING AdminOrMainAdmin MIDDLEWARE ===")

		token := c.Locals("user")
		if token == nil {
			ip := c.IP()
			userAgent := c.Get("User-Agent")
			path := c.Path()

			globalLogger.GetZerologLogger().Error().
				Str("middleware", "AdminOrMainAdmin").
				Str("path", path).
				Msg("Token not found in AdminOrMainAdmin")

			logAuthorizationFailure(ip, userAgent, path, "Token not found", "admin_or_main_admin", globalLogger)

			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized: Token not found",
			})
		}

		claims, err := jwt.GetClaimsFromFiberContext(c)
		if err != nil {
			ip := c.IP()
			userAgent := c.Get("User-Agent")
			path := c.Path()
			errMsg := err.Error()

			globalLogger.GetZerologLogger().Error().
				Str("middleware", "AdminOrMainAdmin").
				Str("path", path).
				Str("error", errMsg).
				Msg("Failed to get claims in AdminOrMainAdmin")

			logAuthorizationFailure(ip, userAgent, path, errMsg, "admin_or_main_admin", globalLogger)

			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized: Invalid token claims",
			})
		}

		if claims.Role != "admin" && claims.Role != "main_admin" {
			ip := c.IP()
			userAgent := c.Get("User-Agent")
			path := c.Path()

			globalLogger.GetZerologLogger().Error().
				Str("middleware", "AdminOrMainAdmin").
				Str("path", path).
				Str("user_role", claims.Role).
				Msg("Insufficient permissions in AdminOrMainAdmin")

			logAuthorizationFailure(ip, userAgent, path, "Insufficient permissions", "admin_or_main_admin", globalLogger)

			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Forbidden: Admin or Main Admin access required",
			})
		}

		globalLogger.GetZerologLogger().Info().
			Str("middleware", "AdminOrMainAdmin").
			Str("path", c.Path()).
			Str("user_id", claims.UserID.String()).
			Str("user_role", claims.Role).
			Msg("=== EXITING AdminOrMainAdmin MIDDLEWARE ===")

		return c.Next()
	}
}

// PartnerOnly middleware checks that user is partner
func PartnerOnly() fiber.Handler {
	return func(c *fiber.Ctx) error {
		globalLogger.GetZerologLogger().Info().
			Str("middleware", "PartnerOnly").
			Str("path", c.Path()).
			Msg("=== ENTERING PartnerOnly MIDDLEWARE ===")

		token := c.Locals("user")
		if token == nil {
			ip := c.IP()
			userAgent := c.Get("User-Agent")
			path := c.Path()

			globalLogger.GetZerologLogger().Error().
				Str("middleware", "PartnerOnly").
				Str("path", path).
				Msg("Token not found in PartnerOnly")

			logAuthorizationFailure(ip, userAgent, path, "Token not found", "partner", globalLogger)

			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized: Token not found",
			})
		}

		claims, err := jwt.GetClaimsFromFiberContext(c)
		if err != nil {
			ip := c.IP()
			userAgent := c.Get("User-Agent")
			path := c.Path()
			errMsg := err.Error()

			globalLogger.GetZerologLogger().Error().
				Str("middleware", "PartnerOnly").
				Str("path", path).
				Str("error", errMsg).
				Msg("Failed to get claims in PartnerOnly")

			logAuthorizationFailure(ip, userAgent, path, errMsg, "partner", globalLogger)

			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized: Invalid token claims",
			})
		}

		if claims.Role != "partner" {
			ip := c.IP()
			userAgent := c.Get("User-Agent")
			path := c.Path()

			globalLogger.GetZerologLogger().Error().
				Str("middleware", "PartnerOnly").
				Str("path", path).
				Str("user_role", claims.Role).
				Msg("Insufficient permissions in PartnerOnly")

			logAuthorizationFailure(ip, userAgent, path, "Insufficient permissions", "partner", globalLogger)

			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Forbidden: Partner access required",
			})
		}

		globalLogger.GetZerologLogger().Info().
			Str("middleware", "PartnerOnly").
			Str("path", c.Path()).
			Str("user_id", claims.UserID.String()).
			Str("user_role", claims.Role).
			Msg("=== EXITING PartnerOnly MIDDLEWARE ===")

		return c.Next()
	}
}

// AdminOrPartner middleware allows access to admins and partners
func AdminOrPartner(logger *Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Locals("user")
		if token == nil {
			ip := c.IP()
			userAgent := c.Get("User-Agent")
			path := c.Path()

			logAuthorizationFailure(ip, userAgent, path, "Token not found", "admin_or_partner", logger)

			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized: Token not found",
			})
		}

		claims, err := jwt.GetClaimsFromFiberContext(c)
		if err != nil {
			ip := c.IP()
			userAgent := c.Get("User-Agent")
			path := c.Path()
			errMsg := err.Error()

			logAuthorizationFailure(ip, userAgent, path, errMsg, "admin_or_partner", logger)

			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized: Invalid token claims",
			})
		}

		if claims.Role != "admin" && claims.Role != "partner" {
			ip := c.IP()
			userAgent := c.Get("User-Agent")
			path := c.Path()

			logAuthorizationFailure(ip, userAgent, path, "insufficient_privileges", "admin_or_partner", logger)

			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Forbidden: Admin or Partner access required",
			})
		}

		ip := c.IP()
		userAgent := c.Get("User-Agent")
		path := c.Path()
		role := claims.Role

		logAuthorizationSuccess(ip, userAgent, path, role, "admin_or_partner", logger)

		return c.Next()
	}
}

// logAuthFailure logs authentication failure
func logAuthFailure(ip, userAgent, path, errMsg string, logger *Logger) {
	logger.GetZerologLogger().Warn().
		Str("ip", ip).
		Str("user_agent", userAgent).
		Str("path", path).
		Str("error", errMsg).
		Msg("Authentication failure")
}

// logAuthSuccess logs successful authentication
func logAuthSuccess(ip, userAgent, path string, logger *Logger) {
	logger.GetZerologLogger().Info().
		Str("ip", ip).
		Str("user_agent", userAgent).
		Str("path", path).
		Msg("Authentication success")
}

// logAuthorizationFailure logs authorization failure
func logAuthorizationFailure(ip, userAgent, path, errMsg, requiredRole string, logger *Logger) {
	logger.GetZerologLogger().Warn().
		Str("ip", ip).
		Str("user_agent", userAgent).
		Str("path", path).
		Str("error", errMsg).
		Str("required_role", requiredRole).
		Msg("Authorization failure")
}

// logAuthorizationSuccess logs successful authorization
func logAuthorizationSuccess(ip, userAgent, path, userRole, requiredRole string, logger *Logger) {
	logger.GetZerologLogger().Debug().
		Str("ip", ip).
		Str("user_agent", userAgent).
		Str("path", path).
		Str("user_role", userRole).
		Str("required_role", requiredRole).
		Str("event_type", "authorization_success").
		Str("stack_trace", string(debug.Stack())).
		Msg("Authorization successful")
}
