package coupon

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Handler struct {
	fiber.Router
	repo *CouponRepository
}

func NewCouponHandler(router fiber.Router, db *gorm.DB) {
	handler := &Handler{
		Router: router,
		repo:   NewRepository(db),
	}

	api := handler.Group("/coupons")
	api.Get("/", handler.GetCoupons)
	api.Get("/:id", handler.GetCouponByID)
	api.Post("/", handler.CreateCoupons)
	api.Get("/code/:code", handler.GetCouponByCode)
	api.Put("/:id/activate", handler.ActivateCoupon)
	api.Put("/:id/reset", handler.ResetCoupon)
	api.Put("/:id/send-schema", handler.SendSchema)
	api.Put("/:id/purchase", handler.MarkAsPurchased)
	api.Delete("/:id", handler.DeleteCoupon)
	api.Get("/statistics", handler.GetStatistics)
}

// GetCoupons возвращает список купонов с фильтрацией
func (h *Handler) GetCoupons(c *fiber.Ctx) error {
	code := c.Query("code")
	status := c.Query("status")
	size := c.Query("size")
	style := c.Query("style")
	partnerIDStr := c.Query("partner_id")

	var partnerID *uuid.UUID
	if partnerIDStr != "" {
		if id, err := uuid.Parse(partnerIDStr); err == nil {
			partnerID = &id
		}
	}

	coupons, err := h.repo.Search(code, status, size, style, partnerID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch coupons"})
	}

	return c.JSON(coupons)
}

// GetCouponByID возвращает купон по ID
func (h *Handler) GetCouponByID(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid coupon ID"})
	}

	coupon, err := h.repo.GetByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Coupon not found"})
	}

	return c.JSON(coupon)
}

// GetCouponByCode возвращает купон по коду
func (h *Handler) GetCouponByCode(c *fiber.Ctx) error {
	code := c.Params("code")

	coupon, err := h.repo.GetByCode(code)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Coupon not found"})
	}

	return c.JSON(coupon)
}

// CreateCoupons создает новые купоны
func (h *Handler) CreateCoupons(c *fiber.Ctx) error {
	var req CreateCouponsRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if req.Count <= 0 || req.Count > 10000 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Count must be between 1 and 10000"})
	}

	// Генерируем купоны
	coupons, err := h.generateCoupons(req.Count, req.PartnerID, req.Size, req.Style)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate coupons"})
	}

	// Сохраняем купоны в базу данных
	if err := h.repo.CreateBatch(coupons); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save coupons"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Coupons created successfully",
		"count":   len(coupons),
		"coupons": coupons,
	})
}

// ActivateCoupon активирует купон
func (h *Handler) ActivateCoupon(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid coupon ID"})
	}

	var req struct {
		OriginalImageURL string `json:"original_image_url"`
		PreviewURL       string `json:"preview_url"`
		SchemaURL        string `json:"schema_url"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if err := h.repo.ActivateCoupon(id, req.OriginalImageURL, req.PreviewURL, req.SchemaURL); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to activate coupon"})
	}

	return c.JSON(fiber.Map{"message": "Coupon activated successfully"})
}

// ResetCoupon сбрасывает купон в исходное состояние
func (h *Handler) ResetCoupon(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid coupon ID"})
	}

	if err := h.repo.ResetCoupon(id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to reset coupon"})
	}

	return c.JSON(fiber.Map{"message": "Coupon reset successfully"})
}

// SendSchema отправляет схему на email
func (h *Handler) SendSchema(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid coupon ID"})
	}

	var req struct {
		Email string `json:"email"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if err := h.repo.SendSchema(id, req.Email); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to send schema"})
	}

	return c.JSON(fiber.Map{"message": "Schema sent successfully"})
}

// MarkAsPurchased помечает купон как купленный
func (h *Handler) MarkAsPurchased(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid coupon ID"})
	}

	var req struct {
		PurchaseEmail string `json:"purchase_email"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if err := h.repo.MarkAsPurchased(id, req.PurchaseEmail); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to mark as purchased"})
	}

	return c.JSON(fiber.Map{"message": "Coupon marked as purchased"})
}

// DeleteCoupon удаляет купон
func (h *Handler) DeleteCoupon(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid coupon ID"})
	}

	if err := h.repo.Delete(id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete coupon"})
	}

	return c.JSON(fiber.Map{"message": "Coupon deleted successfully"})
}

// GetStatistics возвращает статистику по купонам
func (h *Handler) GetStatistics(c *fiber.Ctx) error {
	partnerIDStr := c.Query("partner_id")

	var partnerID *uuid.UUID
	if partnerIDStr != "" {
		if id, err := uuid.Parse(partnerIDStr); err == nil {
			partnerID = &id
		}
	}

	stats, err := h.repo.GetStatistics(partnerID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get statistics"})
	}

	return c.JSON(stats)
}

// generateCoupons генерирует указанное количество купонов
func (h *Handler) generateCoupons(count int, partnerID uuid.UUID, size, style string) ([]*Coupon, error) {
	// Получаем код партнера (первые 4 цифры)
	partnerCode := "0000"

	coupons := make([]*Coupon, 0, count)

	for i := 0; i < count; i++ {
		// Генерируем последние 8 цифр
		randomPart, err := h.generateRandomCode(8)
		if err != nil {
			return nil, err
		}

		// Формируем полный код: XXXX-XXXX-XXXX
		code := fmt.Sprintf("%s-%s-%s",
			partnerCode,
			randomPart[:4],
			randomPart[4:8])

		coupon := &Coupon{
			Code:      code,
			PartnerID: partnerID,
			Size:      size,
			Style:     style,
			Status:    "new",
		}

		coupons = append(coupons, coupon)
	}

	return coupons, nil
}

// generateRandomCode генерирует случайный числовой код указанной длины
func (h *Handler) generateRandomCode(length int) (string, error) {
	code := ""
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		code += strconv.Itoa(int(n.Int64()))
	}
	return code, nil
}
