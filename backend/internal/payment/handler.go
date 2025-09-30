package payment

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/skr1ms/mosaic/pkg/middleware"
)

type PaymentHandlerDeps struct {
	PaymentService   PaymentServiceInterface
	CouponRepository CouponRepositoryInterface
	Logger           *middleware.Logger
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

	// ================================================================
	// PAYMENT ROUTES: /api/payment/*
	// Access: public (no authentication required)
	// ================================================================
	paymentGroup := router.Group("/payment")
	paymentGroup.Post("/purchase", handler.PurchaseCoupon)                  // POST /api/payment/purchase
	paymentGroup.Get("/orders/:orderNumber/status", handler.GetOrderStatus) // GET /api/payment/orders/:orderNumber/status
	paymentGroup.Get("/options", handler.GetAvailableOptions)               // GET /api/payment/options
	paymentGroup.Get("/return", handler.PaymentReturn)                      // GET /api/payment/return
	paymentGroup.Post("/notification", handler.PaymentNotification)         // POST /api/payment/notification
	paymentGroup.Get("/notification", handler.PaymentNotificationGet)       // GET /api/payment/notification
	paymentGroup.Get("/test-integration", handler.TestIntegration)          // GET /api/payment/test-integration
}

// @Summary Purchase coupon
// @Description Process coupon purchase and create payment order
// @Tags payment
// @Accept json
// @Produce json
// @Param request body PurchaseCouponRequest true "Purchase request data"
// @Success 200 {object} PurchaseCouponResponse
// @Failure 400 {object} map[string]any "Invalid request data"
// @Failure 500 {object} map[string]any "Failed to create order"
// @Router /payment/purchase [post]
func (h *PaymentHandler) PurchaseCoupon(c *fiber.Ctx) error {
	var req PurchaseCouponRequest
	if err := c.BodyParser(&req); err != nil {
		h.deps.Logger.FromContext(c).Warn().
			Err(err).
			Str("handler", "PurchaseCoupon").
			Msg("Failed to parse request body")

		errorResponse := fiber.Map{
			"error":      "Failed to parse request body",
			"request_id": c.Get("X-Request-ID"),
		}
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse)
	}

	domain := c.Get("Host")
	if domain != "" {
		req.Domain = &domain
	}

	response, err := h.deps.PaymentService.PurchaseCoupon(c.Context(), &req)
	if err != nil {
		h.deps.Logger.FromContext(c).Error().
			Err(err).
			Str("handler", "PurchaseCoupon").
			Msg("Failed to create order")

		errorResponse := fiber.Map{
			"error":      "Failed to create order",
			"request_id": c.Get("X-Request-ID"),
		}
		if os.Getenv("ENVIRONMENT") == "development" || os.Getenv("ENVIRONMENT") == "dev" {
			errorResponse["details"] = err.Error()
		}
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse)
	}

	h.deps.Logger.FromContext(c).Info().
		Str("handler", "PurchaseCoupon").
		Str("order_number", response.OrderNumber).
		Str("amount", response.Amount).
		Msg("Coupon purchased successfully")

	return c.JSON(response)
}

// @Summary Get order status
// @Description Retrieve payment order status by order number
// @Tags payment
// @Produce json
// @Param orderNumber path string true "Order number"
// @Success 200 {object} OrderStatusResponse
// @Failure 400 {object} map[string]any "Order number is required"
// @Failure 500 {object} map[string]any "Failed to get order status"
// @Router /payment/orders/{orderNumber}/status [get]
func (h *PaymentHandler) GetOrderStatus(c *fiber.Ctx) error {
	orderNumber := strings.TrimSpace(c.Params("orderNumber"))
	if orderNumber == "" {
		h.deps.Logger.FromContext(c).Warn().
			Str("handler", "GetOrderStatus").
			Msg("Order number is required")

		errorResponse := fiber.Map{
			"error":      "Order number is required",
			"request_id": c.Get("X-Request-ID"),
		}
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse)
	}

	response, err := h.deps.PaymentService.GetOrderStatus(c.Context(), orderNumber)
	if err != nil {
		h.deps.Logger.FromContext(c).Error().
			Err(err).
			Str("handler", "GetOrderStatus").
			Str("order_number", orderNumber).
			Msg("Failed to get order status")

		errorResponse := fiber.Map{
			"error":      "Failed to get order status",
			"request_id": c.Get("X-Request-ID"),
		}
		if os.Getenv("ENVIRONMENT") == "development" || os.Getenv("ENVIRONMENT") == "dev" {
			errorResponse["details"] = err.Error()
		}
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse)
	}

	h.deps.Logger.FromContext(c).Info().
		Str("handler", "GetOrderStatus").
		Str("order_number", orderNumber).
		Str("status", response.Status).
		Msg("Order status retrieved successfully")

	return c.JSON(response)
}

// @Summary Get available payment options
// @Description Retrieve available payment methods and configurations
// @Tags payment
// @Produce json
// @Success 200 {object} map[string]any
// @Failure 500 {object} map[string]any "Failed to get payment options"
// @Router /payment/options [get]
func (h *PaymentHandler) GetAvailableOptions(c *fiber.Ctx) error {
	response := h.deps.PaymentService.GetAvailableOptions()

	h.deps.Logger.FromContext(c).Info().
		Str("handler", "GetAvailableOptions").
		Int("options_count", len(response.Sizes)+len(response.Styles)).
		Msg("Payment options retrieved successfully")

	return c.JSON(response)
}

// @Summary Process payment return
// @Description Handle payment return callback from payment gateway
// @Tags payment
// @Produce json
// @Param orderNumber query string true "Order number"
// @Success 302 "Redirect to success page"
// @Failure 400 {object} map[string]any "Order number is required"
// @Failure 500 {object} map[string]any "Failed to process payment return"
// @Router /payment/return [get]
func (h *PaymentHandler) PaymentReturn(c *fiber.Ctx) error {
	orderNumber := c.Query("orderNumber")
	if orderNumber == "" {
		h.deps.Logger.FromContext(c).Warn().
			Str("handler", "PaymentReturn").
			Msg("Order number is required")

		errorResponse := fiber.Map{
			"error":      "Order number is required",
			"request_id": c.Get("X-Request-ID"),
		}
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse)
	}

	err := h.deps.PaymentService.ProcessPaymentReturn(c.Context(), orderNumber)
	if err != nil {
		h.deps.Logger.FromContext(c).Error().
			Err(err).
			Str("handler", "PaymentReturn").
			Str("order_number", orderNumber).
			Msg("Failed to process payment return")

		errorResponse := fiber.Map{
			"error":      "Failed to process payment return",
			"request_id": c.Get("X-Request-ID"),
		}
		if os.Getenv("ENVIRONMENT") == "development" || os.Getenv("ENVIRONMENT") == "dev" {
			errorResponse["details"] = err.Error()
		}
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse)
	}

	h.deps.Logger.FromContext(c).Info().
		Str("handler", "PaymentReturn").
		Str("order_number", orderNumber).
		Str("redirect_to", "/payment/success?order="+orderNumber).
		Msg("Payment return processed successfully")

	return c.Redirect("/payment/success?order=" + orderNumber)
}

// @Summary Process payment notification
// @Description Handle payment webhook notification from payment gateway
// @Tags payment
// @Accept json
// @Produce json
// @Param notification body PaymentNotificationRequest true "Notification data"
// @Success 200 {object} map[string]any "Notification processed successfully"
// @Failure 400 {object} map[string]any "Invalid notification data"
// @Failure 500 {object} map[string]any "Failed to process notification"
// @Router /payment/notification [post]
func (h *PaymentHandler) PaymentNotification(c *fiber.Ctx) error {
	var notification PaymentNotificationRequest

	if err := c.BodyParser(&notification); err != nil {
		h.deps.Logger.FromContext(c).Warn().
			Err(err).
			Str("handler", "PaymentNotification").
			Msg("Failed to parse webhook notification")

		errorResponse := fiber.Map{
			"error":      "Failed to parse webhook notification",
			"request_id": c.Get("X-Request-ID"),
		}
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse)
	}

	orderStatus := -1
	if notification.OrderStatus != nil {
		orderStatus = *notification.OrderStatus
	}

	err := h.deps.PaymentService.ProcessWebhookNotification(c.Context(), &notification)
	if err != nil {
		h.deps.Logger.FromContext(c).Error().
			Err(err).
			Str("handler", "PaymentNotification").
			Str("order_number", notification.OrderNumber).
			Msg("Failed to process webhook notification")

		errorResponse := fiber.Map{
			"error":      "Failed to process webhook notification",
			"request_id": c.Get("X-Request-ID"),
		}
		if os.Getenv("ENVIRONMENT") == "development" || os.Getenv("ENVIRONMENT") == "dev" {
			errorResponse["details"] = err.Error()
		}
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse)
	}

	h.deps.Logger.FromContext(c).Info().
		Str("handler", "PaymentNotification").
		Str("order_number", notification.OrderNumber).
		Int("order_status", orderStatus).
		Msg("Payment notification processed successfully")

	return c.JSON(fiber.Map{
		"success": true,
	})
}

// @Summary Process payment notification via GET
// @Description Handle payment webhook notification from payment gateway via GET request with query parameters
// @Tags payment
// @Produce json
// @Param date query string false "Payment date"
// @Param amount query string false "Payment amount"
// @Param orderNumber query string true "Order number"
// @Param status query string false "Payment status"
// @Param checksum query string false "Checksum for verification"
// @Success 200 {object} map[string]any "Notification processed successfully"
// @Failure 400 {object} map[string]any "Invalid notification data"
// @Failure 500 {object} map[string]any "Failed to process notification"
// @Router /payment/notification [get]
func (h *PaymentHandler) PaymentNotificationGet(c *fiber.Ctx) error {
	// Parse query parameters into notification struct
	var notification PaymentNotificationRequest

	// Parse basic fields
	notification.OrderNumber = c.Query("orderNumber")
	notification.Currency = c.Query("currency")
	notification.OrderDescription = c.Query("orderDescription")
	notification.IP = c.Query("ip")
	notification.Checksum = c.Query("checksum")

	// Parse amount (convert from string to int64)
	if amountStr := c.Query("amount"); amountStr != "" {
		if amount, err := strconv.ParseInt(amountStr, 10, 64); err == nil {
			notification.Amount = amount
		}
	}

	// Parse date (convert from string to int64)
	if dateStr := c.Query("date"); dateStr != "" {
		if date, err := strconv.ParseInt(dateStr, 10, 64); err == nil {
			notification.Date = date
		}
	}

	// Parse optional orderStatus
	if statusStr := c.Query("status"); statusStr != "" {
		if status, err := strconv.Atoi(statusStr); err == nil {
			notification.OrderStatus = &status
		}
	}

	// Log all received parameters for debugging
	h.deps.Logger.FromContext(c).Info().
		Str("handler", "PaymentNotificationGet").
		Str("order_number", notification.OrderNumber).
		Int64("amount", notification.Amount).
		Str("currency", notification.Currency).
		Str("checksum", notification.Checksum).
		Interface("all_query_params", c.Queries()).
		Msg("Received payment notification via GET")

	if notification.OrderNumber == "" {
		h.deps.Logger.FromContext(c).Warn().
			Str("handler", "PaymentNotificationGet").
			Msg("Order number is required")

		errorResponse := fiber.Map{
			"error":      "Order number is required",
			"request_id": c.Get("X-Request-ID"),
		}
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse)
	}

	err := h.deps.PaymentService.ProcessWebhookNotification(c.Context(), &notification)
	if err != nil {
		h.deps.Logger.FromContext(c).Error().
			Err(err).
			Str("handler", "PaymentNotificationGet").
			Str("order_number", notification.OrderNumber).
			Msg("Failed to process webhook notification")

		errorResponse := fiber.Map{
			"error":      "Failed to process webhook notification",
			"request_id": c.Get("X-Request-ID"),
		}
		if os.Getenv("ENVIRONMENT") == "development" || os.Getenv("ENVIRONMENT") == "dev" {
			errorResponse["details"] = err.Error()
		}
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse)
	}

	h.deps.Logger.FromContext(c).Info().
		Str("handler", "PaymentNotificationGet").
		Str("order_number", notification.OrderNumber).
		Msg("Payment notification via GET processed successfully")

	return c.JSON(fiber.Map{
		"success": true,
	})
}

// @Summary Test Alfa Bank integration
// @Description Test integration with Alfa Bank API using current configuration
// @Tags payment
// @Produce json
// @Success 200 {object} map[string]any "Integration test results"
// @Failure 500 {object} map[string]any "Integration test failed"
// @Router /payment/test-integration [get]
func (h *PaymentHandler) TestIntegration(c *fiber.Ctx) error {
	// Create a test order to verify integration
	testOrderNumber := fmt.Sprintf("TEST_%d_%s", time.Now().Unix(), uuid.New().String()[:8])

	// Test registration with minimal data
	alfaReq := &AlfaBankRegisterRequest{
		OrderNumber: testOrderNumber,
		Amount:      10000, // 100 rubles in kopecks
		Currency:    "810", // RUB
		ReturnUrl:   "https://photo.doyoupaint.com/payment/test-return",
		FailUrl:     "https://photo.doyoupaint.com/payment/test-fail",
		Description: "Test integration order",
		Language:    "ru",
	}

	response, err := h.deps.PaymentService.TestAlfaBankIntegration(c.Context(), alfaReq)
	if err != nil {
		h.deps.Logger.FromContext(c).Error().
			Err(err).
			Str("handler", "TestIntegration").
			Msg("Integration test failed")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Integration test failed",
			"details": err.Error(),
		})
	}

	h.deps.Logger.FromContext(c).Info().
		Str("handler", "TestIntegration").
		Str("test_order_number", testOrderNumber).
		Msg("Integration test completed successfully")

	return c.JSON(response)
}
