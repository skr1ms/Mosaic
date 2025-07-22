package user

import (
	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	fiber.Router
}

func NewUserHandler(router fiber.Router) {
	handler := &UserHandler{
		Router: router,
	}

	api := handler.Group("/user")
	api.Post("/activate-coupon", ActivateCoupon)
	api.Post("/upload-image", UploadImage)
	api.Post("/process-image", ProcessImage)
	api.Get("/download-schema/:couponId", DownloadSchema)
	api.Get("/white-label-config", GetWhiteLabelConfig)
}

// POST /user/activate-coupon
func ActivateCoupon(c *fiber.Ctx) error {
	return nil
}

// POST /user/upload-image
func UploadImage(c *fiber.Ctx) error {
	return nil
}

// POST /user/process-image
func ProcessImage(c *fiber.Ctx) error {
	return nil
}

// GET /user/download-schema/:couponId
func DownloadSchema(c *fiber.Ctx) error {
	return nil
}

// GET /user/white-label-config
func GetWhiteLabelConfig(c *fiber.Ctx) error {
	return nil
}
