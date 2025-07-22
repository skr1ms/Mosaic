package coupon

import (
	"github.com/gofiber/fiber/v2"
)

type CouponHandler struct {
	fiber.Router
}

func NewCouponHandler(router fiber.Router) {
	handler := &CouponHandler{
		Router: router,
	}

	api := handler.Group("/coupons")
	api.Get("/", HandleGetAllCoupons)
	api.Get("/:id", HandleGetCouponById)
	api.Get("/code/:code", HandleGetCouponByCode)
	api.Post("/", HandleCreateCoupon)
	api.Put("/:id", HandleUpdateCoupon)
	api.Delete("/:id", HandleDeleteCoupon)
	api.Post("/:id/activate", HandleActivateCoupon)
	api.Get("/:id/status", HandleGetCouponStatus)
	api.Post("/:id/validate", HandleValidateCoupon)
	api.Get("/partner/:partnerId", HandleGetCouponsByPartner)
	api.Get("/:id/qr-code", HandleGenerateQRCode)
}

// GET /coupons/
func HandleGetAllCoupons(c *fiber.Ctx) error {
	return nil
}

// GET /coupons/:id
func HandleGetCouponById(c *fiber.Ctx) error {
	return nil
}

// GET /coupons/code/:code
func HandleGetCouponByCode(c *fiber.Ctx) error {
	return nil
}

// POST /coupons/
func HandleCreateCoupon(c *fiber.Ctx) error {
	return nil
}

// PUT /coupons/:id
func HandleUpdateCoupon(c *fiber.Ctx) error {
	return nil
}

// DELETE /coupons/:id
func HandleDeleteCoupon(c *fiber.Ctx) error {
	return nil
}

// POST /coupons/:id/activate
func HandleActivateCoupon(c *fiber.Ctx) error {
	return nil
}

// GET /coupons/:id/status
func HandleGetCouponStatus(c *fiber.Ctx) error {
	return nil
}

// POST /coupons/:id/validate
func HandleValidateCoupon(c *fiber.Ctx) error {
	return nil
}

// GET /coupons/partner/:partnerId
func HandleGetCouponsByPartner(c *fiber.Ctx) error {
	return nil
}

// GET /coupons/:id/qr-code
func HandleGenerateQRCode(c *fiber.Ctx) error {
	return nil
}
