package auth

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/skr1ms/mosaic/pkg/middleware"
)

type AuthHandlerDeps struct {
	AuthService *AuthService
	Logger      *zerolog.Logger
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

	handler.Post("/login/admin", handler.AdminLogin)               // Авторизация администратора
	handler.Post("/refresh/admin", handler.RefreshAdminTokens)     // Обновление токена администраторов
	handler.Post("/login/partner", handler.PartnerLogin)           // Авторизация партнера
	handler.Post("/refresh/partner", handler.RefreshPartnerTokens) // Обновление токена партнеров
}

// handleError централизованно обрабатывает ошибки и возвращает соответствующий HTTP ответ
func (handler *AuthHandler) handleError(c *fiber.Ctx, err error) error {
	var apiErr APIError
	if errors.As(err, &apiErr) {
		return c.Status(apiErr.HTTPStatus).JSON(fiber.Map{
			"error": apiErr.Message,
			"code":  apiErr.Code,
		})
	}

	// Для неизвестных ошибок
	handler.deps.Logger.Error().Err(err).Msg("Unexpected error in auth handler")
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"error": "Internal server error",
		"code":  "INTERNAL_ERROR",
	})
}

// Login обрабатывает авторизацию администратора и генерирует JWT токены
// @Summary Авторизация администратора
// @Description Авторизация администратора по логину и паролю
// @Tags admin-auth
// @Accept json
// @Produce json
// @Param credentials body LoginRequest true "Учетные данные для входа"
// @Success 200 {object} map[string]interface{} "Успешная авторизация"
// @Failure 400 {object} map[string]interface{} "Ошибка в запросе"
// @Failure 401 {object} map[string]interface{} "Неверные учетные данные"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /admin/login [post]
func (handler *AuthHandler) AdminLogin(c *fiber.Ctx) error {
	var req LoginRequest

	// Парсим запрос
	if err := c.BodyParser(&req); err != nil {
		handler.deps.Logger.Error().Err(err).Msg("Failed to parse request body")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Авторизуем администратора
	admin, token, err := handler.deps.AuthService.AdminLogin(req.Login, req.Password)
	if err != nil {
		handler.deps.Logger.Error().Err(err).Msg("Admin login failed")
		return handler.handleError(c, err)
	}

	// Возвращаем токены
	return c.JSON(fiber.Map{
		"message": "Login successful",
		"admin": fiber.Map{
			"id":    admin.ID,
			"login": admin.Login,
			"role":  "admin",
		},
		"access_token":  token.AccessToken,
		"refresh_token": token.RefreshToken,
		"expires_in":    token.ExpiresIn,
	})
}

// RefreshToken обновляет токены используя refresh токен
// @Summary Обновление токенов
// @Description Обновляет access и refresh токены используя refresh токен
// @Tags admin-auth
// @Accept json
// @Produce json
// @Param refresh body RefreshTokenRequest true "Refresh токен"
// @Success 200 {object} map[string]interface{} "Новые токены"
// @Failure 400 {object} map[string]interface{} "Ошибка в запросе"
// @Failure 401 {object} map[string]interface{} "Неверный или истекший refresh токен"
// @Router /admin/refresh [post]
func (handler *AuthHandler) RefreshAdminTokens(c *fiber.Ctx) error {
	var req RefreshTokenRequest

	// Парсим запрос
	if err := c.BodyParser(&req); err != nil {
		handler.deps.Logger.Error().Err(err).Msg("Failed to parse request body")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Обновляем токены
	tokens, err := handler.deps.AuthService.RefreshAdminTokens(req.RefreshToken)
	if err != nil {
		handler.deps.Logger.Error().Err(err).Msg("Admin token refresh failed")
		return handler.handleError(c, err)
	}

	// Возвращаем токены
	return c.JSON(fiber.Map{
		"message":       "Tokens refreshed successfully",
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"expires_in":    tokens.ExpiresIn,
	})
}

// Login обрабатывает авторизацию партнера и генерирует JWT токены
// @Summary Авторизация партнера
// @Description Авторизация партнера по логину и паролю
// @Tags partner-auth
// @Accept json
// @Produce json
// @Param credentials body LoginRequest true "Учетные данные для входа"
// @Success 200 {object} map[string]interface{} "Успешная авторизация"
// @Failure 400 {object} map[string]interface{} "Ошибка в запросе"
// @Failure 401 {object} map[string]interface{} "Неверные учетные данные"
// @Failure 403 {object} map[string]interface{} "Аккаунт заблокирован"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /partner/login [post]
func (handler *AuthHandler) PartnerLogin(c *fiber.Ctx) error {
	var req LoginRequest

	// Парсим запрос
	if err := c.BodyParser(&req); err != nil {
		handler.deps.Logger.Error().Err(err).Msg("Failed to parse request body")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Валидируем запрос
	if err := middleware.ValidateStruct(&req); err != nil {
		handler.deps.Logger.Error().Err(err).Msg("Validation failed")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Validation failed",
			"details": err.Error(),
		})
	}

	// Авторизуем партнера
	partner, tokens, err := handler.deps.AuthService.PartnerLogin(req.Login, req.Password)
	if err != nil {
		handler.deps.Logger.Error().Err(err).Msg("Partner login failed")
		return handler.handleError(c, err)
	}

	// Возвращаем токены
	return c.JSON(fiber.Map{
		"message": "Login successful",
		"partner": fiber.Map{
			"id":           partner.ID,
			"login":        partner.Login,
			"partner_code": partner.PartnerCode,
			"brand_name":   partner.BrandName,
			"role":         "partner",
		},
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"expires_in":    tokens.ExpiresIn,
	})
}

// RefreshToken обновляет токены используя refresh токен
// @Summary Обновление токенов партнера
// @Description Обновляет access и refresh токены используя refresh токен
// @Tags partner-auth
// @Accept json
// @Produce json
// @Param refresh body RefreshTokenRequest true "Refresh токен"
// @Success 200 {object} map[string]interface{} "Новые токены"
// @Failure 400 {object} map[string]interface{} "Ошибка в запросе"
// @Failure 401 {object} map[string]interface{} "Неверный или истекший refresh токен"
// @Router /partner/refresh [post]
func (handler *AuthHandler) RefreshPartnerTokens(c *fiber.Ctx) error {
	var req RefreshTokenRequest

	// Парсим запрос
	if err := c.BodyParser(&req); err != nil {
		handler.deps.Logger.Error().Err(err).Msg("Failed to parse request body")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Обновляем токены
	tokenPair, err := handler.deps.AuthService.RefreshPartnerTokens(req.RefreshToken)
	if err != nil {
		handler.deps.Logger.Error().Err(err).Msg("Partner token refresh failed")
		return handler.handleError(c, err)
	}

	// Возвращаем токены
	return c.JSON(fiber.Map{
		"message":       "Tokens refreshed successfully",
		"access_token":  tokenPair.AccessToken,
		"refresh_token": tokenPair.RefreshToken,
		"expires_in":    tokenPair.ExpiresIn,
	})
}
