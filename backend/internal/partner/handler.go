package partner

import (
	"github.com/gofiber/fiber/v2"
	"github.com/skr1ms/mosaic/pkg/middleware"
	"github.com/skr1ms/mosaic/pkg/utils"
	"gorm.io/gorm"
)

type PartnerHandler struct {
	fiber.Router
	repo       *PartnerRepository
	jwtService *utils.JWT
}

func NewPartnerHandler(router fiber.Router, db *gorm.DB, jwtService *utils.JWT) {
	handler := &PartnerHandler{
		Router:     router,
		repo:       NewRepository(db),
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
func (h *PartnerHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	partner, err := h.repo.GetByLogin(req.Login)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	if !utils.CheckPassword(req.Password, partner.Password) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	if partner.Status == "blocked" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Account is blocked"})
	}

	tokenPair, err := h.jwtService.CreateTokenPair(partner.ID, partner.Login, "partner")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate tokens"})
	}

	if err := h.repo.UpdateLastLogin(partner.ID); err != nil {
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
func (h *PartnerHandler) RefreshToken(c *fiber.Ctx) error {
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

// GetDashboard возвращает данные для дашборда партнера
func (h *PartnerHandler) GetDashboard(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Partner dashboard"})
}

// GetProfile возвращает профиль партнера
func (h *PartnerHandler) GetProfile(c *fiber.Ctx) error {
	claims, err := utils.GetClaimsFromFiberContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	partner, err := h.repo.GetByID(claims.UserID)
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
func (h *PartnerHandler) UpdateProfile(c *fiber.Ctx) error {
	return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
		"error": "Partner profile can only be updated by administrator",
	})
}

// GetMyCoupons возвращает купоны партнера
func (h *PartnerHandler) GetMyCoupons(c *fiber.Ctx) error {
	claims, err := utils.GetClaimsFromFiberContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	return c.JSON(fiber.Map{
		"message":    "Partner coupons",
		"partner_id": claims.UserID,
	})
}

// GetCouponDetails возвращает детали купона
func (h *PartnerHandler) GetCouponDetails(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Coupon details"})
}

// GetMyStatistics возвращает статистику партнера
func (h *PartnerHandler) GetMyStatistics(c *fiber.Ctx) error {
	claims, err := utils.GetClaimsFromFiberContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	return c.JSON(fiber.Map{
		"message":    "Partner statistics",
		"partner_id": claims.UserID,
	})
}

// GetSalesStatistics возвращает статистику продаж партнера
func (h *PartnerHandler) GetSalesStatistics(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Sales statistics"})
}

// GetUsageStatistics возвращает статистику использования купонов
func (h *PartnerHandler) GetUsageStatistics(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Usage statistics"})
}
