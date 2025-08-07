package payment

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/skr1ms/mosaic/internal/coupon"
	"github.com/skr1ms/mosaic/pkg/utils"
)

type PaymentHandlerDeps struct {
	PaymentService   *PaymentService
	CouponRepository *coupon.CouponRepository
}

type PaymentHandler struct {
	fiber.Router
	deps *PaymentHandlerDeps
}

func NewPaymentHandler(router fiber.Router, deps *PaymentHandlerDeps) {
	handler := &PaymentHandler{
		Router: router,
		deps:   deps,
	}

	// Публичные маршруты для покупки купонов
	paymentGroup := router.Group("/payment")

	// Покупка купона онлайн
	paymentGroup.Post("/purchase", handler.PurchaseCoupon)

	// Получение статуса заказа
	paymentGroup.Get("/orders/:orderNumber/status", handler.GetOrderStatus)

	// Получение доступных размеров и стилей
	paymentGroup.Get("/options", handler.GetAvailableOptions)

	// Обработка возврата от платежной системы
	paymentGroup.Get("/return", handler.PaymentReturn)
	paymentGroup.Post("/notification", handler.PaymentNotification)
}

// PurchaseCoupon - покупка купона онлайн согласно
// @Summary Покупка купона онлайн
// @Description Создание заказа на покупку купона с оплатой картой через Альфа-Банк
// @Tags payment
// @Accept json
// @Produce json
// @Param request body PurchaseCouponRequest true "Данные для покупки купона"
// @Success 200 {object} PurchaseCouponResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /payment/purchase [post]
func (h *PaymentHandler) PurchaseCoupon(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	var req PurchaseCouponRequest
	if err := c.BodyParser(&req); err != nil {
		log.Error().Err(err).Msg("Error parsing request body")
		return utils.LocalizedError(c, fiber.StatusBadRequest, "error_bad_request", "Error parsing request body")
	}

	// Получаем домен из заголовка или параметра
	domain := c.Get("Host")
	if domain != "" {
		req.Domain = &domain
	}

	response, err := h.deps.PaymentService.PurchaseCoupon(c.Context(), &req)
	if err != nil {
		log.Error().Err(err).Msg("Error creating order")
		return utils.LocalizedError(c, fiber.StatusInternalServerError, "error_internal", "Error creating order")
	}

	return c.JSON(response)
}

// GetOrderStatus - получение статуса заказа
// @Summary Получение статуса заказа
// @Description Проверка статуса оплаты заказа
// @Tags payment
// @Produce json
// @Param orderNumber path string true "Номер заказа"
// @Success 200 {object} OrderStatusResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /payment/orders/{orderNumber}/status [get]
func (h *PaymentHandler) GetOrderStatus(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	orderNumber := c.Params("orderNumber")
	if orderNumber == "" {
		log.Error().Msg("Order number is not specified")
		return utils.LocalizedError(c, fiber.StatusBadRequest, "order_number_required", "Order number is not specified")
	}

	response, err := h.deps.PaymentService.GetOrderStatus(c.Context(), orderNumber)
	if err != nil {
		log.Error().Err(err).Msg("Error getting order status")
		return utils.LocalizedError(c, fiber.StatusInternalServerError, "error_internal", "Error getting order status")
	}

	return c.JSON(response)
}

// GetAvailableOptions - получение доступных размеров и стилей
// @Summary Получение доступных опций
// @Description Получение списка доступных размеров и стилей мозаики с ценами
// @Tags payment
// @Produce json
// @Success 200 {object} AvailableOptionsResponse
// @Router /payment/options [get]
func (h *PaymentHandler) GetAvailableOptions(c *fiber.Ctx) error {
	response := h.deps.PaymentService.GetAvailableOptions()
	return c.JSON(response)
}

// PaymentReturn - обработка возврата от платежной системы
func (h *PaymentHandler) PaymentReturn(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	orderNumber := c.Query("orderNumber")
	if orderNumber == "" {
		log.Error().Msg("Order number is not specified")
		return utils.LocalizedError(c, fiber.StatusBadRequest, "order_number_required", "Order number is not specified")
	}

	err := h.deps.PaymentService.ProcessPaymentReturn(c.Context(), orderNumber)
	if err != nil {
		log.Error().Err(err).Msg("Error processing payment return")
		return utils.LocalizedError(c, fiber.StatusInternalServerError, "error_internal", "Error processing payment")
	}

	// Перенаправляем на страницу успеха
	return c.Redirect("/payment/success?order=" + orderNumber)
}

// PaymentNotification - обработка уведомлений от платежной системы
// @Summary Обработка webhook уведомлений от Альфа-Банка
// @Description Обрабатывает уведомления о смене статуса заказа от платежной системы Альфа-Банк
// @Tags payment
// @Accept json,application/x-www-form-urlencoded
// @Produce json
// @Param notification body PaymentNotificationRequest true "Данные уведомления от Альфа-Банка"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /payment/notification [post]
func (h *PaymentHandler) PaymentNotification(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())

	// Получаем данные webhook'а
	var notification PaymentNotificationRequest
	if err := c.BodyParser(&notification); err != nil {
		log.Error().Err(err).Msg("Error parsing webhook notification")
		return utils.LocalizedError(c, fiber.StatusBadRequest, "error_parsing_parameters", "Invalid request format")
	}

	// Логируем получение webhook
	log.Info().
		Str("order_number", notification.OrderNumber).
		Int("order_status", notification.OrderStatus).
		Msg("Received payment webhook notification")

	// Обрабатываем уведомление
	err := h.deps.PaymentService.ProcessWebhookNotification(c.Context(), &notification)
	if err != nil {
		log.Error().Err(err).
			Str("order_number", notification.OrderNumber).
			Msg("Error processing webhook notification")
		return utils.LocalizedError(c, fiber.StatusInternalServerError, "error_internal", "Error processing notification")
	}

	log.Info().
		Str("order_number", notification.OrderNumber).
		Msg("Successfully processed payment webhook")

	// Альфа-Банк ожидает статус 200 для подтверждения получения webhook'а
	return c.JSON(fiber.Map{
		"success": true,
	})
}
