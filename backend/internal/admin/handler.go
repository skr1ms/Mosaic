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
func (h *Handler) GetDashboard(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Admin dashboard"})
}

// GetPartners возвращает список партнеров
func (h *Handler) GetPartners(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Partners list"})
}

// CreatePartner создает нового партнера
func (h *Handler) CreatePartner(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Create partner"})
}

// UpdatePartner обновляет партнера
func (h *Handler) UpdatePartner(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Update partner"})
}

// DeletePartner удаляет партнера
func (h *Handler) DeletePartner(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Delete partner"})
}

// GetCoupons возвращает список купонов
func (h *Handler) GetCoupons(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Coupons list"})
}

// CreateCoupon создает новые купоны
func (h *Handler) CreateCoupon(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Create coupons"})
}

// UpdateCoupon обновляет купон
func (h *Handler) UpdateCoupon(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Update coupon"})
}

// DeleteCoupon удаляет купон
func (h *Handler) DeleteCoupon(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Delete coupon"})
}

// GetStatistics возвращает общую статистику
func (h *Handler) GetStatistics(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Statistics"})
}
