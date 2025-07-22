package partner

import (
	"github.com/gofiber/fiber/v2"
)

type PartnerHandler struct {
	fiber.Router
}

func NewPartnerHandler(router fiber.Router) {
	handler := &PartnerHandler{
		Router: router,
	}

	api := handler.Group("/partner")
	api.Get("/dashboard", GetDashboard)
	api.Get("/profile", GetProfile)
	api.Put("/profile", UpdateProfile)
	api.Get("/coupons", GetMyCoupons)
	api.Post("/coupons/generate", GenerateCoupons)
	api.Get("/coupons/:id", GetCouponDetails)
	api.Put("/coupons/:id", UpdateCoupon)
	api.Delete("/coupons/:id", DeleteCoupon)
	api.Get("/statistics", GetMyStatistics)
	api.Get("/statistics/sales", GetSalesStatistics)
	api.Get("/statistics/usage", GetUsageStatistics)
	api.Post("/white-label/config", UpdateWhiteLabelConfig)
	api.Get("/white-label/config", GetWhiteLabelConfig)
}

// GET /partner/dashboard
func GetDashboard(c *fiber.Ctx) error {
	return nil
}

// GET /partner/profile
func GetProfile(c *fiber.Ctx) error {
	return nil
}

// PUT /partner/profile
func UpdateProfile(c *fiber.Ctx) error {
	return nil
}

// GET /partner/coupons
func GetMyCoupons(c *fiber.Ctx) error {
	return nil
}

// POST /partner/coupons/generate
func GenerateCoupons(c *fiber.Ctx) error {
	return nil
}

// GET /partner/coupons/:id
func GetCouponDetails(c *fiber.Ctx) error {
	return nil
}

// PUT /partner/coupons/:id
func UpdateCoupon(c *fiber.Ctx) error {
	return nil
}

// DELETE /partner/coupons/:id
func DeleteCoupon(c *fiber.Ctx) error {
	return nil
}

// GET /partner/statistics
func GetMyStatistics(c *fiber.Ctx) error {
	return nil
}

// GET /partner/statistics/sales
func GetSalesStatistics(c *fiber.Ctx) error {
	return nil
}

// GET /partner/statistics/usage
func GetUsageStatistics(c *fiber.Ctx) error {
	return nil
}

// POST /partner/white-label/config
func UpdateWhiteLabelConfig(c *fiber.Ctx) error {
	return nil
}

// GET /partner/white-label/config
func GetWhiteLabelConfig(c *fiber.Ctx) error {
	return nil
}
