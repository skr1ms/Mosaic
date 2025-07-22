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
	api.Get("/", HandleGetAllActivations)
	api.Get("/:id", HandleGetActivationById)
	api.Post("/", HandleCreateActivation)
	api.Get("/coupon/:couponId", HandleGetActivationsByCoupon)
	api.Get("/stats/daily", HandleGetDailyActivationStats)
	api.Get("/stats/weekly", HandleGetWeeklyActivationStats)
	api.Get("/stats/monthly", HandleGetMonthlyActivationStats)
	api.Post("/verify", HandleVerifyActivation)
}

// GET /activations/
func HandleGetAllActivations(c *fiber.Ctx) error {
	return nil
}

// GET /activations/:id
func HandleGetActivationById(c *fiber.Ctx) error {
	return nil
}

// POST /activations/
func HandleCreateActivation(c *fiber.Ctx) error {
	return nil
}

// GET /activations/coupon/:couponId
func HandleGetActivationsByCoupon(c *fiber.Ctx) error {
	return nil
}

// GET /activations/stats/daily
func HandleGetDailyActivationStats(c *fiber.Ctx) error {
	return nil
}

// GET /activations/stats/weekly
func HandleGetWeeklyActivationStats(c *fiber.Ctx) error {
	return nil
}

// GET /activations/stats/monthly
func HandleGetMonthlyActivationStats(c *fiber.Ctx) error {
	return nil
}

// POST /activations/verify
func HandleVerifyActivation(c *fiber.Ctx) error {
	return nil
}
