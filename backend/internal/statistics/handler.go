package statistics

import (
	"github.com/gofiber/fiber/v2"
)

type StatisticsHandler struct {
	fiber.Router
}

func NewStatisticsHandler(router fiber.Router) {
	handler := &StatisticsHandler{
		Router: router,
	}

	api := handler.Group("/statistics")
	api.Get("/", GetAllStatistics)
	api.Get("/partner/:partnerId", GetPartnerStatistics)
	api.Get("/coupon/:couponId", GetCouponStatistics)
	api.Get("/summary", GetSummaryStatistics)
	api.Get("/conversion", GetConversionStatistics)
	api.Get("/revenue", GetRevenueStatistics)
	api.Get("/usage", GetUsageStatistics)
	api.Get("/trends/daily", GetDailyTrends)
	api.Get("/trends/weekly", GetWeeklyTrends)
	api.Get("/trends/monthly", GetMonthlyTrends)
	api.Get("/export", ExportStatistics)
}

// GET /statistics/
func GetAllStatistics(c *fiber.Ctx) error {
	return nil
}

// GET /statistics/partner/:partnerId
func GetPartnerStatistics(c *fiber.Ctx) error {
	return nil
}

// GET /statistics/coupon/:couponId
func GetCouponStatistics(c *fiber.Ctx) error {
	return nil
}

// GET /statistics/summary
func GetSummaryStatistics(c *fiber.Ctx) error {
	return nil
}

// GET /statistics/conversion
func GetConversionStatistics(c *fiber.Ctx) error {
	return nil
}

// GET /statistics/revenue
func GetRevenueStatistics(c *fiber.Ctx) error {
	return nil
}

// GET /statistics/usage
func GetUsageStatistics(c *fiber.Ctx) error {
	return nil
}

// GET /statistics/trends/daily
func GetDailyTrends(c *fiber.Ctx) error {
	return nil
}

// GET /statistics/trends/weekly
func GetWeeklyTrends(c *fiber.Ctx) error {
	return nil
}

// GET /statistics/trends/monthly
func GetMonthlyTrends(c *fiber.Ctx) error {
	return nil
}

// GET /statistics/export
func ExportStatistics(c *fiber.Ctx) error {
	return nil
}
