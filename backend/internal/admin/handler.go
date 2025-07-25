package admin

import (
	"github.com/gofiber/fiber/v2"
	"github.com/skr1ms/mosaic/pkg/middleware"
	"github.com/skr1ms/mosaic/pkg/utils"
	"gorm.io/gorm"
)

type Handler struct {
	fiber.Router
	repo       *AdminRepository
	jwtService *utils.JWT
}

func NewAdminHandler(router fiber.Router, db *gorm.DB, jwtService *utils.JWT) {
	handler := &Handler{
		Router:     router,
		repo:       NewRepository(db),
		jwtService: jwtService,
	}

	api := handler.Group("/admin")

	// Публичные endpoints (без JWT)
	api.Post("/login", handler.Login)
	api.Post("/refresh", handler.RefreshToken)

	// Защищенные endpoints (требуют JWT + admin роль)
	protected := api.Use(middleware.JWTMiddleware(jwtService), middleware.AdminOnly())
	protected.Post("/admins", handler.CreateAdmin)
	protected.Get("/admins", handler.GetAdmins)
	protected.Get("/dashboard", handler.GetDashboard)
	protected.Get("/partners", handler.GetPartners)
	protected.Post("/partners", handler.CreatePartner)
	protected.Put("/partners/:id", handler.UpdatePartner)
	protected.Delete("/partners/:id", handler.DeletePartner)
	protected.Get("/coupons", handler.GetCoupons)
	protected.Post("/coupons", handler.CreateCoupon)
	protected.Put("/coupons/:id", handler.UpdateCoupon)
	protected.Delete("/coupons/:id", handler.DeleteCoupon)
	protected.Get("/statistics", handler.GetStatistics)
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
func (h *Handler) Login(c *fiber.Ctx) error {
	var req LoginRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	admin, err := h.repo.GetByLogin(req.Login)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	if !utils.CheckPassword(req.Password, admin.Password) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	tokenPair, err := h.jwtService.CreateTokenPair(admin.ID, admin.Login, "admin")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate tokens"})
	}

	if err := h.repo.UpdateLastLogin(admin.ID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update login time"})
	}

	return c.JSON(fiber.Map{
		"message": "Login successful",
		"admin": fiber.Map{
			"id":    admin.ID,
			"login": admin.Login,
			"role":  "admin",
		},
		"access_token":  tokenPair.AccessToken,
		"refresh_token": tokenPair.RefreshToken,
		"expires_in":    tokenPair.ExpiresIn,
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
func (h *Handler) RefreshToken(c *fiber.Ctx) error {
	var req RefreshTokenRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	tokenPair, err := h.jwtService.RefreshTokens(req.RefreshToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid or expired refresh token"})
	}

	return c.JSON(fiber.Map{
		"access_token":  tokenPair.AccessToken,
		"refresh_token": tokenPair.RefreshToken,
		"expires_in":    tokenPair.ExpiresIn,
	})
}

// CreateAdmin создает нового администратора
// @Summary Создание администратора
// @Description Создает нового администратора (только для существующих администраторов)
// @Tags admin-management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param admin body CreateAdminRequest true "Данные нового администратора"
// @Success 201 {object} map[string]interface{} "Администратор создан"
// @Failure 400 {object} map[string]interface{} "Ошибка в запросе"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 403 {object} map[string]interface{} "Нет прав доступа"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /admin/admins [post]
func (h *Handler) CreateAdmin(c *fiber.Ctx) error {
	var req CreateAdminRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to hash password"})
	}

	admin := &Admin{
		Login:    req.Login,
		Password: hashedPassword,
	}

	if err := h.repo.Create(admin); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create admin"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":    admin.ID,
		"login": admin.Login,
		"role":  "admin",
	})
}

// GetAdmins возвращает список всех администраторов
// @Summary Список администраторов
// @Description Возвращает список всех администраторов
// @Tags admin-management
// @Produce json
// @Security BearerAuth
// @Success 200 {array} map[string]interface{} "Список администраторов"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 403 {object} map[string]interface{} "Нет прав доступа"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /admin/admins [get]
func (h *Handler) GetAdmins(c *fiber.Ctx) error {
	admins, err := h.repo.GetAll()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch admins"})
	}

	result := make([]fiber.Map, len(admins))
	for i, admin := range admins {
		result[i] = fiber.Map{
			"id":         admin.ID,
			"login":      admin.Login,
			"last_login": admin.LastLogin,
			"created_at": admin.CreatedAt,
		}
	}

	return c.JSON(result)
}

// GetDashboard возвращает данные для дашборда администратора
// @Summary Дашборд администратора
// @Description Возвращает данные для главной страницы администратора
// @Tags admin-dashboard
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Данные дашборда"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 403 {object} map[string]interface{} "Нет прав доступа"
// @Router /admin/dashboard [get]
func (h *Handler) GetDashboard(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Admin dashboard"})
}

// GetPartners возвращает список партнеров
// @Summary Список партнеров
// @Description Возвращает список всех партнеров
// @Tags admin-partners
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Список партнеров"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 403 {object} map[string]interface{} "Нет прав доступа"
// @Router /admin/partners [get]
func (h *Handler) GetPartners(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Partners list"})
}

// CreatePartner создает нового партнера
// @Summary Создание партнера
// @Description Создает нового партнера
// @Tags admin-partners
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 201 {object} map[string]interface{} "Партнер создан"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 403 {object} map[string]interface{} "Нет прав доступа"
// @Router /admin/partners [post]
func (h *Handler) CreatePartner(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Create partner"})
}

// UpdatePartner обновляет партнера
// @Summary Обновление партнера
// @Description Обновляет данные партнера
// @Tags admin-partners
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID партнера"
// @Success 200 {object} map[string]interface{} "Партнер обновлен"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 403 {object} map[string]interface{} "Нет прав доступа"
// @Router /admin/partners/{id} [put]
func (h *Handler) UpdatePartner(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Update partner"})
}

// DeletePartner удаляет партнера
// @Summary Удаление партнера
// @Description Удаляет партнера по ID
// @Tags admin-partners
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID партнера"
// @Success 200 {object} map[string]interface{} "Партнер удален"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 403 {object} map[string]interface{} "Нет прав доступа"
// @Router /admin/partners/{id} [delete]
func (h *Handler) DeletePartner(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Delete partner"})
}

// GetCoupons возвращает список купонов
// @Summary Список купонов
// @Description Возвращает список всех купонов
// @Tags admin-coupons
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Список купонов"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 403 {object} map[string]interface{} "Нет прав доступа"
// @Router /admin/coupons [get]
func (h *Handler) GetCoupons(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Coupons list"})
}

// CreateCoupon создает новые купоны
// @Summary Создание купонов
// @Description Создает новые купоны
// @Tags admin-coupons
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 201 {object} map[string]interface{} "Купоны созданы"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 403 {object} map[string]interface{} "Нет прав доступа"
// @Router /admin/coupons [post]
func (h *Handler) CreateCoupon(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Create coupons"})
}

// UpdateCoupon обновляет купон
// @Summary Обновление купона
// @Description Обновляет данные купона
// @Tags admin-coupons
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID купона"
// @Success 200 {object} map[string]interface{} "Купон обновлен"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 403 {object} map[string]interface{} "Нет прав доступа"
// @Router /admin/coupons/{id} [put]
func (h *Handler) UpdateCoupon(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Update coupon"})
}

// DeleteCoupon удаляет купон
// @Summary Удаление купона
// @Description Удаляет купон по ID
// @Tags admin-coupons
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID купона"
// @Success 200 {object} map[string]interface{} "Купон удален"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 403 {object} map[string]interface{} "Нет прав доступа"
// @Router /admin/coupons/{id} [delete]
func (h *Handler) DeleteCoupon(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Delete coupon"})
}

// GetStatistics возвращает общую статистику
// @Summary Общая статистика
// @Description Возвращает общую статистику по системе
// @Tags admin-statistics
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Статистика"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 403 {object} map[string]interface{} "Нет прав доступа"
// @Router /admin/statistics [get]
func (h *Handler) GetStatistics(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Statistics"})
}
