package activation

import (
	"github.com/gofiber/fiber/v2"
)

type ActivationHandler struct {
	fiber.Router
}

func NewActivationHandler(router fiber.Router) {
	handler := &ActivationHandler{
		Router: router,
	}

	api := handler.Group("/activations")
	api.Get("/", GetAllActivations)
	api.Get("/:id", GetActivationById)
	api.Post("/", CreateActivation)
	api.Get("/coupon/:couponId", GetActivationsByCoupon)
	api.Get("/stats/daily", GetDailyActivationStats)
	api.Get("/stats/weekly", GetWeeklyActivationStats)
	api.Get("/stats/monthly", GetMonthlyActivationStats)
	api.Post("/verify", VerifyActivation)
}

// GET /activations/
func GetAllActivations(c *fiber.Ctx) error {
	return nil
}

// GET /activations/:id
func GetActivationById(c *fiber.Ctx) error {
	return nil
}

// POST /activations/
func CreateActivation(c *fiber.Ctx) error {
	return nil
}

// GET /activations/coupon/:couponId
func GetActivationsByCoupon(c *fiber.Ctx) error {
	return nil
}

// GET /activations/stats/daily
func GetDailyActivationStats(c *fiber.Ctx) error {
	return nil
}

// GET /activations/stats/weekly
func GetWeeklyActivationStats(c *fiber.Ctx) error {
	return nil
}

// GET /activations/stats/monthly
func GetMonthlyActivationStats(c *fiber.Ctx) error {
	return nil
}

// POST /activations/verify
func VerifyActivation(c *fiber.Ctx) error {
	return nil
}
