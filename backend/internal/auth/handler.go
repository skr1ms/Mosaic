package auth

import (
	"context"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/skr1ms/mosaic/pkg/jwt"
	"github.com/skr1ms/mosaic/pkg/middleware"
)

type AuthHandlerDeps struct {
	AuthService    AuthServiceInterface
	PartnerService PartnerServiceInterface
	JwtService     JWTServiceInterface
	Logger         *middleware.Logger
}

type AuthHandler struct {
	fiber.Router
	deps *AuthHandlerDeps
}

func NewAuthHandler(router fiber.Router, AuthHandlerDeps *AuthHandlerDeps) {
	handler := &AuthHandler{
		Router: router,
		deps:   AuthHandlerDeps,
	}

	// ================================================================
	// PUBLIC AUTH ROUTES: /api/auth/*
	// Access: public (no authentication required)
	// ================================================================
	handler.Post("/login", handler.UniversalLogin)           // POST /api/auth/login
	handler.Post("/refresh", handler.RefreshTokens)          // POST /api/auth/refresh
	handler.Post("/forgot-password", handler.ForgotPassword) // POST /api/auth/forgot-password
	handler.Post("/reset-password", handler.ResetPassword)   // POST /api/auth/reset-password

	// ================================================================
	// PROTECTED AUTH ROUTES: /api/auth/*
	// Access: authenticated users (admin or partner)
	// ================================================================
	jwtConcrete, _ := handler.deps.JwtService.(*jwt.JWT)
	handler.Post("/change-password", middleware.JWTMiddleware(jwtConcrete, handler.deps.Logger), handler.ChangePassword) // POST /api/auth/change-password
	handler.Post("/change-email", middleware.JWTMiddleware(jwtConcrete, handler.deps.Logger), handler.ChangeEmail)       // POST /api/auth/change-email
}

// @Summary      Universal login
// @Description  Login as admin or partner using credentials
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        credentials body LoginRequest true "Login credentials"
// @Success      200 {object} map[string]any "Successful login"
// @Failure      400 {object} map[string]any "Bad request"
// @Failure      401 {object} map[string]any "Invalid credentials"
// @Failure      403 {object} map[string]any "Account blocked"
// @Failure      500 {object} map[string]any "Internal server error"
// @Router       /auth/login [post]
func (handler *AuthHandler) UniversalLogin(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		handler.deps.Logger.FromContext(c).Warn().
			Err(err).
			Str("handler", "UniversalLogin").
			Msg("Invalid request body")

		errorResponse := fiber.Map{
			"error":      "Invalid request body",
			"request_id": c.Get("X-Request-ID"),
		}
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse)
	}
	if err := middleware.ValidateStruct(&req); err != nil {
		handler.deps.Logger.FromContext(c).Warn().
			Err(err).
			Str("handler", "UniversalLogin").
			Msg("Validation failed")

		errorResponse := fiber.Map{
			"error":      "Validation failed",
			"request_id": c.Get("X-Request-ID"),
		}
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse)
	}

	admin, adminTokens, adminErr := handler.deps.AuthService.AdminLogin(req.Login, req.Password)
	if adminErr == nil && admin != nil {
		handler.deps.Logger.FromContext(c).Info().
			Str("handler", "UniversalLogin").
			Str("user_id", admin.ID.String()).
			Str("role", admin.Role).
			Msg("Admin login successful")

		return c.JSON(fiber.Map{
			"message": "Login successful",
			"user": fiber.Map{
				"id":    admin.ID,
				"login": admin.Login,
				"email": admin.Email,
				"role":  admin.Role,
			},
			"access_token":  adminTokens.AccessToken,
			"refresh_token": adminTokens.RefreshToken,
			"expires_in":    adminTokens.ExpiresIn,
		})
	}

	partner, partnerTokens, partnerErr := handler.deps.AuthService.PartnerLogin(req.Login, req.Password)
	if partnerErr == nil && partner != nil {
		handler.deps.Logger.FromContext(c).Info().
			Str("handler", "UniversalLogin").
			Str("user_id", partner.ID.String()).
			Str("role", "partner").
			Msg("Partner login successful")

		return c.JSON(fiber.Map{
			"message": "Login successful",
			"user": fiber.Map{
				"id":           partner.ID,
				"login":        partner.Login,
				"partner_code": partner.PartnerCode,
				"brand_name":   partner.BrandName,
				"email":        partner.Email,
				"role":         "partner",
			},
			"access_token":  partnerTokens.AccessToken,
			"refresh_token": partnerTokens.RefreshToken,
			"expires_in":    partnerTokens.ExpiresIn,
		})
	}

	handler.deps.Logger.FromContext(c).Warn().
		Str("handler", "UniversalLogin").
		Str("login", req.Login).
		Msg("Invalid credentials")

	errorResponse := fiber.Map{
		"error":      "Invalid credentials",
		"request_id": c.Get("X-Request-ID"),
	}
	return c.Status(fiber.StatusUnauthorized).JSON(errorResponse)
}

// @Summary      Refresh tokens
// @Description  Refresh access and refresh tokens using a refresh token (admin or partner)
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        refresh body RefreshTokenRequest true "Refresh token"
// @Success      200 {object} map[string]any "New tokens"
// @Failure      400 {object} map[string]any "Bad request"
// @Failure      401 {object} map[string]any "Invalid or expired refresh token"
// @Failure      500 {object} map[string]any "Internal server error"
// @Router       /auth/refresh [post]
func (handler *AuthHandler) RefreshTokens(c *fiber.Ctx) error {
	var req RefreshTokenRequest
	if err := c.BodyParser(&req); err != nil {
		handler.deps.Logger.FromContext(c).Warn().
			Err(err).
			Str("handler", "RefreshTokens").
			Msg("Invalid request body")

		errorResponse := fiber.Map{
			"error":      "Invalid request body",
			"request_id": c.Get("X-Request-ID"),
		}
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse)
	}

	claims, err := handler.deps.JwtService.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		handler.deps.Logger.FromContext(c).Warn().
			Err(err).
			Str("handler", "RefreshTokens").
			Msg("Invalid refresh token")

		errorResponse := fiber.Map{
			"error":      "Invalid refresh token",
			"request_id": c.Get("X-Request-ID"),
		}
		return c.Status(fiber.StatusUnauthorized).JSON(errorResponse)
	}

	var tokens *jwt.TokenPair
	switch claims.Role {
	case "admin", "main_admin", "partner":
		tokens, err = handler.deps.AuthService.RefreshTokens(req.RefreshToken)
	default:
		handler.deps.Logger.FromContext(c).Warn().
			Str("handler", "RefreshTokens").
			Str("role", claims.Role).
			Msg("Invalid user role")

		errorResponse := fiber.Map{
			"error":      "Invalid user role",
			"request_id": c.Get("X-Request-ID"),
		}
		return c.Status(fiber.StatusUnauthorized).JSON(errorResponse)
	}

	if err != nil {
		handler.deps.Logger.FromContext(c).Error().
			Err(err).
			Str("handler", "RefreshTokens").
			Msg("Invalid refresh token")

		errorResponse := fiber.Map{
			"error":      "Invalid refresh token",
			"request_id": c.Get("X-Request-ID"),
		}
		if os.Getenv("ENVIRONMENT") == "development" || os.Getenv("ENVIRONMENT") == "dev" {
			errorResponse["details"] = err.Error()
		}
		return c.Status(fiber.StatusUnauthorized).JSON(errorResponse)
	}

	handler.deps.Logger.FromContext(c).Info().
		Str("handler", "RefreshTokens").
		Str("user_id", claims.UserID.String()).
		Str("role", claims.Role).
		Msg("Tokens refreshed successfully")

	return c.JSON(fiber.Map{
		"message":       "Tokens refreshed successfully",
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"expires_in":    tokens.ExpiresIn,
	})
}

// @Summary      Forgot password request
// @Description  Sends an email with a password reset link
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body ForgotPasswordRequest true "Verification data and captcha"
// @Success      200 {object} map[string]any "Email sent"
// @Failure      400 {object} map[string]any "Bad request"
// @Failure      404 {object} map[string]any "User not found"
// @Failure      500 {object} map[string]any "Internal server error"
// @Router       /auth/forgot-password [post]
func (handler *AuthHandler) ForgotPassword(c *fiber.Ctx) error {
	var reqPayload ForgotPasswordRequest
	if err := c.BodyParser(&reqPayload); err != nil {
		handler.deps.Logger.FromContext(c).Warn().
			Err(err).
			Str("handler", "ForgotPassword").
			Msg("Bad request")

		errorResponse := fiber.Map{
			"error":      "Bad request",
			"request_id": c.Get("X-Request-ID"),
		}
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse)
	}
	if err := middleware.ValidateStruct(&reqPayload); err != nil {
		handler.deps.Logger.FromContext(c).Warn().
			Err(err).
			Str("handler", "ForgotPassword").
			Msg("Validation failed")

		errorResponse := fiber.Map{
			"error":      "Validation failed",
			"request_id": c.Get("X-Request-ID"),
		}
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse)
	}

	err := handler.deps.AuthService.ForgotPassword(
		context.Background(),
		reqPayload.Login,
		reqPayload.Email,
		reqPayload.Captcha,
	)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().
			Err(err).
			Str("handler", "ForgotPassword").
			Str("login", reqPayload.Login).
			Str("email", reqPayload.Email).
			Msg("Failed to send forgot password email")

		errorResponse := fiber.Map{
			"error":      "Failed to send forgot password email",
			"request_id": c.Get("X-Request-ID"),
		}
		if os.Getenv("ENVIRONMENT") == "development" || os.Getenv("ENVIRONMENT") == "dev" {
			errorResponse["details"] = err.Error()
		}
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse)
	}

	handler.deps.Logger.FromContext(c).Info().
		Str("handler", "ForgotPassword").
		Str("login", reqPayload.Login).
		Str("email", reqPayload.Email).
		Msg("Forgot password email sent")

	return c.JSON(fiber.Map{"message": "If email exists, email sent"})
}

// @Summary      Reset password
// @Description  Resets password using a token from email
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body ResetPasswordRequest true "Reset token and new password"
// @Success      200 {object} map[string]any "Password changed"
// @Failure      400 {object} map[string]any "Bad request"
// @Failure      401 {object} map[string]any "Invalid or expired token"
// @Failure      404 {object} map[string]any "User not found"
// @Failure      500 {object} map[string]any "Internal server error"
// @Router       /auth/reset-password [post]
func (handler *AuthHandler) ResetPassword(c *fiber.Ctx) error {
	var req ResetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		handler.deps.Logger.FromContext(c).Warn().
			Err(err).
			Str("handler", "ResetPassword").
			Msg("Bad request")

		errorResponse := fiber.Map{
			"error":      "Bad request",
			"request_id": c.Get("X-Request-ID"),
		}
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse)
	}
	if err := middleware.ValidateStruct(&req); err != nil {
		handler.deps.Logger.FromContext(c).Warn().
			Err(err).
			Str("handler", "ResetPassword").
			Msg("Validation failed")

		errorResponse := fiber.Map{
			"error":      "Validation failed",
			"request_id": c.Get("X-Request-ID"),
		}
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse)
	}

	err := handler.deps.AuthService.ResetPassword(context.Background(), req.Token, req.NewPassword)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().
			Err(err).
			Str("handler", "ResetPassword").
			Msg("Failed to reset password")

		errorResponse := fiber.Map{
			"error":      "Failed to reset password",
			"request_id": c.Get("X-Request-ID"),
		}
		if os.Getenv("ENVIRONMENT") == "development" || os.Getenv("ENVIRONMENT") == "dev" {
			errorResponse["details"] = err.Error()
		}
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse)
	}

	handler.deps.Logger.FromContext(c).Info().
		Str("handler", "ResetPassword").
		Msg("Password has been reset successfully")

	return c.JSON(fiber.Map{"message": "Password has been reset successfully"})
}

// @Summary      Change password
// @Description  Changes the password of the current user (admin or partner)
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body ChangePasswordRequest true "Current and new password"
// @Success      200 {object} map[string]any "Password changed"
// @Failure      400 {object} map[string]any "Bad request"
// @Failure      401 {object} map[string]any "Invalid current password"
// @Failure      500 {object} map[string]any "Internal server error"
// @Router       /auth/change-password [post]
// @Security     BearerAuth
func (handler *AuthHandler) ChangePassword(c *fiber.Ctx) error {
	var req ChangePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		handler.deps.Logger.FromContext(c).Warn().
			Err(err).
			Str("handler", "ChangePassword").
			Msg("Bad request")

		errorResponse := fiber.Map{
			"error":      "Bad request",
			"request_id": c.Get("X-Request-ID"),
		}
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse)
	}
	if err := middleware.ValidateStruct(&req); err != nil {
		handler.deps.Logger.FromContext(c).Warn().
			Err(err).
			Str("handler", "ChangePassword").
			Msg("Validation failed")

		errorResponse := fiber.Map{
			"error":      "Validation failed",
			"request_id": c.Get("X-Request-ID"),
		}
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse)
	}

	claims, err := jwt.GetClaimsFromFiberContext(c)
	if err != nil {
		handler.deps.Logger.FromContext(c).Warn().
			Err(err).
			Str("handler", "ChangePassword").
			Msg("Unauthorized")

		errorResponse := fiber.Map{
			"error":      "Unauthorized",
			"request_id": c.Get("X-Request-ID"),
		}
		return c.Status(fiber.StatusUnauthorized).JSON(errorResponse)
	}

	err = handler.deps.AuthService.ChangePassword(claims.UserID, claims.Role, req.CurrentPassword, req.NewPassword)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().
			Err(err).
			Str("handler", "ChangePassword").
			Str("user_id", claims.UserID.String()).
			Str("role", claims.Role).
			Msg("Failed to change password")

		if err.Error() == "invalid current password" {
			errorResponse := fiber.Map{
				"error":      "Invalid current password",
				"request_id": c.Get("X-Request-ID"),
			}
			return c.Status(fiber.StatusUnauthorized).JSON(errorResponse)
		}

		errorResponse := fiber.Map{
			"error":      "Failed to change password",
			"request_id": c.Get("X-Request-ID"),
		}
		if os.Getenv("ENVIRONMENT") == "development" || os.Getenv("ENVIRONMENT") == "dev" {
			errorResponse["details"] = err.Error()
		}
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse)
	}

	handler.deps.Logger.FromContext(c).Info().
		Str("handler", "ChangePassword").
		Str("user_id", claims.UserID.String()).
		Str("role", claims.Role).
		Msg("Password has been changed successfully")

	return c.JSON(fiber.Map{"message": "Password has been changed successfully"})
}

// @Summary      Change admin email
// @Description  Changes the email of the current admin (admin only)
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body ChangeEmailRequest true "Password and new email"
// @Success      200 {object} map[string]any "Email changed successfully"
// @Failure      400 {object} map[string]any "Validation error"
// @Failure      401 {object} map[string]any "Invalid password"
// @Failure      403 {object} map[string]any "Admins only"
// @Failure      409 {object} map[string]any "Email already in use"
// @Failure      500 {object} map[string]any "Internal server error"
// @Router       /auth/change-email [post]
// @Security     BearerAuth
func (handler *AuthHandler) ChangeEmail(c *fiber.Ctx) error {
	var req ChangeEmailRequest
	if err := c.BodyParser(&req); err != nil {
		handler.deps.Logger.FromContext(c).Warn().
			Err(err).
			Str("handler", "ChangeEmail").
			Msg("Invalid request payload")

		errorResponse := fiber.Map{
			"error":      "Invalid request payload",
			"request_id": c.Get("X-Request-ID"),
		}
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse)
	}
	if err := middleware.ValidateStruct(&req); err != nil {
		handler.deps.Logger.FromContext(c).Warn().
			Err(err).
			Str("handler", "ChangeEmail").
			Msg("Invalid request payload")

		errorResponse := fiber.Map{
			"error":      "Invalid request payload",
			"request_id": c.Get("X-Request-ID"),
		}
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse)
	}

	claims, err := jwt.GetClaimsFromFiberContext(c)
	if err != nil {
		handler.deps.Logger.FromContext(c).Warn().
			Err(err).
			Str("handler", "ChangeEmail").
			Msg("Unauthorized")

		errorResponse := fiber.Map{
			"error":      "Unauthorized",
			"request_id": c.Get("X-Request-ID"),
		}
		return c.Status(fiber.StatusUnauthorized).JSON(errorResponse)
	}

	if claims.Role != "admin" && claims.Role != "main_admin" && claims.Role != "partner" {
		handler.deps.Logger.FromContext(c).Warn().
			Str("handler", "ChangeEmail").
			Str("user_id", claims.UserID.String()).
			Str("role", claims.Role).
			Msg("Email change is only available for admins and partners")

		errorResponse := fiber.Map{
			"error":      "Email change is only available for admins and partners",
			"request_id": c.Get("X-Request-ID"),
		}
		return c.Status(fiber.StatusForbidden).JSON(errorResponse)
	}

	switch claims.Role {
	case "admin", "main_admin":
		err = handler.deps.AuthService.ChangeAdminEmail(claims.UserID, req.Password, req.NewEmail)
	case "partner":
		err = handler.deps.AuthService.ChangePartnerEmail(claims.UserID, req.Password, req.NewEmail)
	}

	if err != nil {
		handler.deps.Logger.FromContext(c).Error().
			Err(err).
			Str("handler", "ChangeEmail").
			Str("user_id", claims.UserID.String()).
			Str("role", claims.Role).
			Str("new_email", req.NewEmail).
			Msg("Failed to update email")

		if err.Error() == "invalid password" {
			errorResponse := fiber.Map{
				"error":      "Invalid password",
				"request_id": c.Get("X-Request-ID"),
			}
			return c.Status(fiber.StatusUnauthorized).JSON(errorResponse)
		}
		if strings.Contains(err.Error(), "already in use") {
			errorResponse := fiber.Map{
				"error":      "Email already in use",
				"request_id": c.Get("X-Request-ID"),
			}
			return c.Status(fiber.StatusConflict).JSON(errorResponse)
		}

		errorResponse := fiber.Map{
			"error":      "Failed to update email",
			"request_id": c.Get("X-Request-ID"),
		}
		if os.Getenv("ENVIRONMENT") == "development" || os.Getenv("ENVIRONMENT") == "dev" {
			errorResponse["details"] = err.Error()
		}
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse)
	}

	handler.deps.Logger.FromContext(c).Info().
		Str("handler", "ChangeEmail").
		Str("user_id", claims.UserID.String()).
		Str("new_email", req.NewEmail).
		Msg("Email has been changed successfully")

	return c.JSON(fiber.Map{"message": "Email has been changed successfully"})
}
