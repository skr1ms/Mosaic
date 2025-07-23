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
// @Summary Список купонов с фильтрацией
// @Description Возвращает список купонов с возможностью фильтрации по коду, статусу, размеру, стилю и партнеру
// @Tags coupons
// @Produce json
// @Param code query string false "Код купона для поиска"
// @Param status query string false "Статус купона (new, used)"
// @Param size query string false "Размер купона"
// @Param style query string false "Стиль купона"
// @Param partner_id query string false "ID партнера"
// @Success 200 {array} map[string]interface{} "Список купонов"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /coupons [get]
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
// @Summary Получение купона по ID
// @Description Возвращает детальную информацию о купоне по его ID
// @Tags coupons
// @Produce json
// @Param id path string true "ID купона"
// @Success 200 {object} map[string]interface{} "Информация о купоне"
// @Failure 400 {object} map[string]interface{} "Неверный ID купона"
// @Failure 404 {object} map[string]interface{} "Купон не найден"
// @Router /coupons/{id} [get]
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
// @Summary Получение купона по коду
// @Description Возвращает детальную информацию о купоне по его коду
// @Tags coupons
// @Produce json
// @Param code path string true "Код купона"
// @Success 200 {object} map[string]interface{} "Информация о купоне"
// @Failure 404 {object} map[string]interface{} "Купон не найден"
// @Router /coupons/code/{code} [get]
func (h *Handler) GetCouponByCode(c *fiber.Ctx) error {
	code := c.Params("code")

	coupon, err := h.repo.GetByCode(code)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Coupon not found"})
	}

	return c.JSON(coupon)
}

// CreateCoupons создает новые купоны
// @Summary Создание купонов
// @Description Создает указанное количество новых купонов для партнера
// @Tags coupons
// @Accept json
// @Produce json
// @Param request body CreateCouponsRequest true "Параметры для создания купонов"
// @Success 201 {object} map[string]interface{} "Купоны созданы успешно"
// @Failure 400 {object} map[string]interface{} "Ошибка в запросе"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /coupons [post]
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
// @Summary Активация купона
// @Description Активирует купон, изменяя его статус на 'used' и добавляя ссылки на изображения
// @Tags coupons
// @Accept json
// @Produce json
// @Param id path string true "ID купона"
// @Param request body map[string]string true "Ссылки на изображения"
// @Success 200 {object} map[string]interface{} "Купон активирован"
// @Failure 400 {object} map[string]interface{} "Неверный ID купона"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /coupons/{id}/activate [put]
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
// @Summary Сброс купона
// @Description Сбрасывает купон в исходное состояние (статус 'new')
// @Tags coupons
// @Produce json
// @Param id path string true "ID купона"
// @Success 200 {object} map[string]interface{} "Купон сброшен"
// @Failure 400 {object} map[string]interface{} "Неверный ID купона"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /coupons/{id}/reset [put]
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
// @Summary Отправка схемы на email
// @Description Отправляет схему купона на указанный email адрес
// @Tags coupons
// @Accept json
// @Produce json
// @Param id path string true "ID купона"
// @Param request body map[string]string true "Email для отправки"
// @Success 200 {object} map[string]interface{} "Схема отправлена"
// @Failure 400 {object} map[string]interface{} "Неверный ID купона"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /coupons/{id}/send-schema [put]
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
// @Summary Пометка купона как купленного
// @Description Помечает купон как купленный онлайн с указанием email покупателя
// @Tags coupons
// @Accept json
// @Produce json
// @Param id path string true "ID купона"
// @Param request body map[string]string true "Email покупателя"
// @Success 200 {object} map[string]interface{} "Купон помечен как купленный"
// @Failure 400 {object} map[string]interface{} "Неверный ID купона"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /coupons/{id}/purchase [put]
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
// @Summary Удаление купона
// @Description Удаляет купон по ID
// @Tags coupons
// @Produce json
// @Param id path string true "ID купона"
// @Success 200 {object} map[string]interface{} "Купон удален"
// @Failure 400 {object} map[string]interface{} "Неверный ID купона"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /coupons/{id} [delete]
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
// @Summary Статистика по купонам
// @Description Возвращает статистику по купонам с возможностью фильтрации по партнеру
// @Tags coupons
// @Produce json
// @Param partner_id query string false "ID партнера для фильтрации"
// @Success 200 {object} map[string]interface{} "Статистика по купонам"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /coupons/statistics [get]
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
