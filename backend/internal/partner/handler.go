package partner

import (
	"github.com/gofiber/fiber/v2"
	"github.com/skr1ms/mosaic/pkg/bcrypt"
	"github.com/skr1ms/mosaic/pkg/jwt"
	"github.com/skr1ms/mosaic/pkg/middleware"
	"gorm.io/gorm"
)

type PartnerHandler struct {
	fiber.Router
	repo       *PartnerRepository
	jwtService *jwt.JWT
}

func NewPartnerHandler(router fiber.Router, db *gorm.DB, jwtService *jwt.JWT) {
	handler := &PartnerHandler{
		Router:     router,
		repo:       NewPartnerRepository(db),
		jwtService: jwtService,
	}

	api := handler.Group("/partner")

	// Публичные endpoints (без JWT)
	api.Post("/login", handler.Login)
	api.Post("/refresh", handler.RefreshToken)

	// Защищенные endpoints (требуют JWT + partner роль)
	protected := api.Use(middleware.JWTMiddleware(jwtService), middleware.PartnerOnly())
	protected.Get("/dashboard", handler.GetDashboard)
	protected.Get("/profile", handler.GetProfile)
	protected.Put("/profile", handler.UpdateProfile)
	protected.Get("/coupons", handler.GetMyCoupons)
	protected.Get("/coupons/:id", handler.GetCouponDetails)
	protected.Get("/statistics", handler.GetMyStatistics)
	protected.Get("/statistics/sales", handler.GetSalesStatistics)
	protected.Get("/statistics/usage", handler.GetUsageStatistics)
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
func (handler *PartnerHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if err := middleware.ValidateStruct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Validation failed",
			"details": err.Error(),
		})
	}

	partner, err := handler.repo.GetByLogin(req.Login)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	if !bcrypt.CheckPassword(req.Password, partner.Password) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	if partner.Status == "blocked" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Account is blocked"})
	}

	tokenPair, err := handler.jwtService.CreateTokenPair(partner.ID, partner.Login, "partner")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate tokens"})
	}

	if err := handler.repo.UpdateLastLogin(partner.ID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update login time"})
	}

	return c.JSON(fiber.Map{
		"message": "Login successful",
		"partner": fiber.Map{
			"id":           partner.ID,
			"login":        partner.Login,
			"partner_code": partner.PartnerCode,
			"brand_name":   partner.BrandName,
			"role":         "partner",
		},
		"access_token":  tokenPair.AccessToken,
		"refresh_token": tokenPair.RefreshToken,
		"expires_in":    tokenPair.ExpiresIn,
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
func (handler *PartnerHandler) RefreshToken(c *fiber.Ctx) error {
	var req RefreshTokenRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	tokenPair, err := handler.jwtService.RefreshTokens(req.RefreshToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid or expired refresh token"})
	}

	return c.JSON(fiber.Map{
		"access_token":  tokenPair.AccessToken,
		"refresh_token": tokenPair.RefreshToken,
		"expires_in":    tokenPair.ExpiresIn,
	})
}

// GetDashboard возвращает данные для дашборда партнера
// @Summary Дашборд партнера
// @Description Возвращает данные для главной страницы партнера
// @Tags partner-dashboard
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Данные дашборда"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 403 {object} map[string]interface{} "Нет прав доступа"
// @Router /partner/dashboard [get]
func (handler *PartnerHandler) GetDashboard(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Partner dashboard"})
}

// GetProfile возвращает профиль партнера
// @Summary Профиль партнера
// @Description Возвращает информацию о профиле текущего партнера
// @Tags partner-profile
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Профиль партнера"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 403 {object} map[string]interface{} "Нет прав доступа"
// @Failure 404 {object} map[string]interface{} "Партнер не найден"
// @Router /partner/profile [get]
func (handler *PartnerHandler) GetProfile(c *fiber.Ctx) error {
	claims, err := jwt.GetClaimsFromFiberContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	partner, err := handler.repo.GetByID(claims.UserID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Partner not found"})
	}

	return c.JSON(fiber.Map{
		"id":           partner.ID,
		"login":        partner.Login,
		"partner_code": partner.PartnerCode,
		"brand_name":   partner.BrandName,
		"domain":       partner.Domain,
		"email":        partner.Email,
		"address":      partner.Address,
		"phone":        partner.Phone,
		"telegram":     partner.Telegram,
		"whatsapp":     partner.Whatsapp,
		"allow_sales":  partner.AllowSales,
		"status":       partner.Status,
		"created_at":   partner.CreatedAt,
		"updated_at":   partner.UpdatedAt,
	})
}

// UpdateProfile обновляет профиль партнера (только для чтения в партнерской панели)
// @Summary Обновление профиля партнера
// @Description Попытка обновления профиля партнера (доступно только администратору)
// @Tags partner-profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Failure 403 {object} map[string]interface{} "Нет прав доступа"
// @Router /partner/profile [put]
func (handler *PartnerHandler) UpdateProfile(c *fiber.Ctx) error {
	return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
		"error": "Partner profile can only be updated by administrator",
	})
}

// GetMyCoupons возвращает купоны партнера
// @Summary Купоны партнера
// @Description Возвращает список купонов текущего партнера
// @Tags partner-coupons
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Список купонов партнера"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 403 {object} map[string]interface{} "Нет прав доступа"
// @Router /partner/coupons [get]
func (handler *PartnerHandler) GetMyCoupons(c *fiber.Ctx) error {
	claims, err := jwt.GetClaimsFromFiberContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	return c.JSON(fiber.Map{
		"message":    "Partner coupons",
		"partner_id": claims.UserID,
	})
}

// GetCouponDetails возвращает детали купона
// @Summary Детали купона
// @Description Возвращает подробную информацию о купоне
// @Tags partner-coupons
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID купона"
// @Success 200 {object} map[string]interface{} "Детали купона"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 403 {object} map[string]interface{} "Нет прав доступа"
// @Router /partner/coupons/{id} [get]
func (handler *PartnerHandler) GetCouponDetails(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Coupon details"})
}

// GetMyStatistics возвращает статистику партнера
// @Summary Статистика партнера
// @Description Возвращает общую статистику текущего партнера
// @Tags partner-statistics
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Статистика партнера"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 403 {object} map[string]interface{} "Нет прав доступа"
// @Router /partner/statistics [get]
func (handler *PartnerHandler) GetMyStatistics(c *fiber.Ctx) error {
	claims, err := jwt.GetClaimsFromFiberContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	return c.JSON(fiber.Map{
		"message":    "Partner statistics",
		"partner_id": claims.UserID,
	})
}

// GetSalesStatistics возвращает статистику продаж партнера
// @Summary Статистика продаж
// @Description Возвращает статистику продаж текущего партнера
// @Tags partner-statistics
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Статистика продаж"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 403 {object} map[string]interface{} "Нет прав доступа"
// @Router /partner/statistics/sales [get]
func (handler *PartnerHandler) GetSalesStatistics(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Sales statistics"})
}

// GetUsageStatistics возвращает статистику использования купонов
// @Summary Статистика использования купонов
// @Description Возвращает статистику использования купонов партнера
// @Tags partner-statistics
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Статистика использования"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 403 {object} map[string]interface{} "Нет прав доступа"
// @Router /partner/statistics/usage [get]
func (handler *PartnerHandler) GetUsageStatistics(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Usage statistics"})
}
