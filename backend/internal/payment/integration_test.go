package payment

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/skr1ms/mosaic/pkg/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Integration test suite for payment handlers
type PaymentIntegrationTestSuite struct {
	app              *fiber.App
	paymentService   *MockPaymentService
	couponRepository *MockCouponRepository
	logger           *middleware.Logger
}

// Mock Payment Service for handler testing
type MockPaymentService struct {
	mock.Mock
}

func (m *MockPaymentService) PurchaseCoupon(ctx context.Context, req *PurchaseCouponRequest) (*PurchaseCouponResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*PurchaseCouponResponse), args.Error(1)
}

func (m *MockPaymentService) GetOrderStatus(ctx context.Context, orderNumber string) (*OrderStatusResponse, error) {
	args := m.Called(ctx, orderNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*OrderStatusResponse), args.Error(1)
}

func (m *MockPaymentService) GenerateOrderNumber() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockPaymentService) GetAvailableOptions() *AvailableOptionsResponse {
	args := m.Called()
	return args.Get(0).(*AvailableOptionsResponse)
}

func (m *MockPaymentService) ProcessPaymentReturn(ctx context.Context, orderNumber string) error {
	args := m.Called(ctx, orderNumber)
	return args.Error(0)
}

func (m *MockPaymentService) ProcessWebhookNotification(ctx context.Context, notification *PaymentNotificationRequest) error {
	args := m.Called(ctx, notification)
	return args.Error(0)
}

func (m *MockPaymentService) TestAlfaBankIntegration(ctx context.Context, req *AlfaBankRegisterRequest) (*TestIntegrationResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*TestIntegrationResponse), args.Error(1)
}

func setupTestSuite() *PaymentIntegrationTestSuite {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	mockPaymentService := new(MockPaymentService)
	mockCouponRepository := new(MockCouponRepository)
	mockLogger := &middleware.Logger{}

	// Setup payment handler
	api := app.Group("/api")
	deps := &PaymentHandlerDeps{
		PaymentService:   mockPaymentService,
		CouponRepository: mockCouponRepository,
		Logger:           mockLogger,
	}
	NewPaymentHandler(api, deps)

	return &PaymentIntegrationTestSuite{
		app:              app,
		paymentService:   mockPaymentService,
		couponRepository: mockCouponRepository,
		logger:           mockLogger,
	}
}

func TestPaymentHandler_PurchaseCoupon_Integration(t *testing.T) {
	tests := []struct {
		name               string
		requestBody        map[string]any
		mockSetup          func(*MockPaymentService)
		expectedStatusCode int
		expectedSuccess    bool
	}{
		{
			name: "successful_purchase",
			requestBody: map[string]any{
				"size":       "40x50",
				"style":      "max_colors",
				"email":      "test@example.com",
				"return_url": "https://example.com/success",
				"fail_url":   "https://example.com/fail",
				"language":   "ru",
			},
			mockSetup: func(service *MockPaymentService) {
				service.On("PurchaseCoupon", mock.Anything, mock.Anything).Return(&PurchaseCouponResponse{
					OrderID:     uuid.New().String(),
					OrderNumber: "ORD_TEST_123",
					PaymentURL:  "https://pay.alfabank.ru/payment/form",
					Success:     true,
					Amount:      "100.00",
				}, nil)
			},
			expectedStatusCode: 200,
			expectedSuccess:    true,
		},
		{
			name: "invalid_request_body",
			requestBody: map[string]any{
				"size": "INVALID_SIZE",
			},
			mockSetup: func(service *MockPaymentService) {
				// Handler will call service even for invalid data, so we need to mock it
				service.On("PurchaseCoupon", mock.Anything, mock.Anything).Return(&PurchaseCouponResponse{
					Success: false,
					Message: "Invalid size value",
				}, nil)
			},
			expectedStatusCode: 200,
			expectedSuccess:    false,
		},
		{
			name: "missing_required_fields",
			requestBody: map[string]any{
				"size": "40x50",
				// Missing email and return_url
			},
			mockSetup: func(service *MockPaymentService) {
				// Handler will call service even for missing fields, so we need to mock it
				service.On("PurchaseCoupon", mock.Anything, mock.Anything).Return(&PurchaseCouponResponse{
					Success: false,
					Message: "Missing required fields",
				}, nil)
			},
			expectedStatusCode: 200,
			expectedSuccess:    false,
		},
		{
			name: "service_error",
			requestBody: map[string]any{
				"size":       "40x50",
				"style":      "max_colors",
				"email":      "test@example.com",
				"return_url": "https://example.com/success",
			},
			mockSetup: func(service *MockPaymentService) {
				service.On("PurchaseCoupon", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("internal service error"))
			},
			expectedStatusCode: 500,
			expectedSuccess:    false,
		},
		{
			name: "invalid_style_value",
			requestBody: map[string]any{
				"size":       "40x50",
				"style":      "invalid_style",
				"email":      "test@example.com",
				"return_url": "https://example.com/success",
			},
			mockSetup: func(service *MockPaymentService) {
				// Handler will call service even for invalid style, so we need to mock it
				service.On("PurchaseCoupon", mock.Anything, mock.Anything).Return(&PurchaseCouponResponse{
					Success: false,
					Message: "Invalid style value",
				}, nil)
			},
			expectedStatusCode: 200,
			expectedSuccess:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suite := setupTestSuite()

			if tt.mockSetup != nil {
				tt.mockSetup(suite.paymentService)
			}

			// Create request
			bodyJSON, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/api/payment/purchase", bytes.NewBuffer(bodyJSON))
			req.Header.Set("Content-Type", "application/json")

			// Execute request
			resp, _ := suite.app.Test(req, -1)

			// Assertions
			assert.Equal(t, tt.expectedStatusCode, resp.StatusCode)

			if tt.expectedStatusCode == 200 {
				var response PurchaseCouponResponse
				json.NewDecoder(resp.Body).Decode(&response)
				assert.Equal(t, tt.expectedSuccess, response.Success)

				if tt.expectedSuccess {
					assert.NotEmpty(t, response.OrderID)
					assert.NotEmpty(t, response.PaymentURL)
					assert.NotEmpty(t, response.Amount)
				}
			}

			suite.paymentService.AssertExpectations(t)
		})
	}
}

func TestPaymentHandler_GetOrderStatus_Integration(t *testing.T) {
	tests := []struct {
		name               string
		orderNumber        string
		mockSetup          func(*MockPaymentService)
		expectedStatusCode int
		expectedSuccess    bool
	}{
		{
			name:        "successful_status_check",
			orderNumber: "ORD_VALID_123",
			mockSetup: func(service *MockPaymentService) {
				service.On("GetOrderStatus", mock.Anything, "ORD_VALID_123").Return(&OrderStatusResponse{
					OrderID:    uuid.New().String(),
					Status:     OrderStatusPaid,
					Size:       "40x50",
					Style:      "max_colors",
					Amount:     100.0,
					Currency:   "RUB",
					CouponCode: stringPtr("COUPON-123-456"),
					Success:    true,
				}, nil)
			},
			expectedStatusCode: 200,
			expectedSuccess:    true,
		},
		{
			name:        "order_not_found",
			orderNumber: "ORD_NOT_FOUND",
			mockSetup: func(service *MockPaymentService) {
				service.On("GetOrderStatus", mock.Anything, "ORD_NOT_FOUND").Return(&OrderStatusResponse{
					Success: false,
					Message: "Order not found",
				}, nil)
			},
			expectedStatusCode: 200,
			expectedSuccess:    false,
		},
		{
			name:        "invalid_order_number",
			orderNumber: "INVALID_ORDER",
			mockSetup: func(service *MockPaymentService) {
				// Handler will call service with this value, expect error
				service.On("GetOrderStatus", mock.Anything, "INVALID_ORDER").Return(nil, errors.New("order not found"))
			},
			expectedStatusCode: 500, // Service error will return 500
			expectedSuccess:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suite := setupTestSuite()

			if tt.mockSetup != nil {
				tt.mockSetup(suite.paymentService)
			}

			// Create request
			url := fmt.Sprintf("/api/payment/orders/%s/status", tt.orderNumber)
			req := httptest.NewRequest("GET", url, nil)

			// Execute request
			resp, _ := suite.app.Test(req, -1)

			// Assertions
			assert.Equal(t, tt.expectedStatusCode, resp.StatusCode)

			if tt.expectedStatusCode == 200 {
				var response OrderStatusResponse
				json.NewDecoder(resp.Body).Decode(&response)
				assert.Equal(t, tt.expectedSuccess, response.Success)
			}

			suite.paymentService.AssertExpectations(t)
		})
	}
}

func TestPaymentHandler_TestIntegration_Integration(t *testing.T) {
	tests := []struct {
		name               string
		mockSetup          func(*MockPaymentService)
		expectedStatusCode int
		expectedSuccess    bool
	}{
		{
			name: "successful_integration_test",
			mockSetup: func(service *MockPaymentService) {
				service.On("TestAlfaBankIntegration", mock.Anything, mock.Anything).Return(&TestIntegrationResponse{
					Success:         true,
					Message:         "Successfully connected to Alfa Bank API and created test order",
					OrderNumber:     "TEST_123456789_abc",
					AlfaBankOrderID: "ALFA_ORDER_123",
					PaymentURL:      "https://pay.alfabank.ru/payment/form",
					TestStatus:      "Integration test successful",
					ConfigStatus:    "Configuration OK: URL=https://pay.alfabank.ru, Username=r-photo_doyoupaint-api",
				}, nil)
			},
			expectedStatusCode: 200,
			expectedSuccess:    true,
		},
		{
			name: "configuration_error",
			mockSetup: func(service *MockPaymentService) {
				service.On("TestAlfaBankIntegration", mock.Anything, mock.Anything).Return(&TestIntegrationResponse{
					Success:      false,
					Message:      "Integration test failed due to configuration issues",
					ConfigStatus: "Configuration issues: [username not set, password not set]",
					TestStatus:   "Configuration check failed",
				}, nil)
			},
			expectedStatusCode: 200,
			expectedSuccess:    false,
		},
		{
			name: "api_connection_error",
			mockSetup: func(service *MockPaymentService) {
				service.On("TestAlfaBankIntegration", mock.Anything, mock.Anything).Return(&TestIntegrationResponse{
					Success:      false,
					Message:      "Failed to connect to Alfa Bank API",
					TestStatus:   "API connection failed",
					ErrorDetails: "connection timeout",
				}, nil)
			},
			expectedStatusCode: 200,
			expectedSuccess:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suite := setupTestSuite()

			if tt.mockSetup != nil {
				tt.mockSetup(suite.paymentService)
			}

			// Create request
			req := httptest.NewRequest("GET", "/api/payment/test-integration", nil)

			// Execute request
			resp, _ := suite.app.Test(req, -1)

			// Assertions
			assert.Equal(t, tt.expectedStatusCode, resp.StatusCode)

			var response TestIntegrationResponse
			json.NewDecoder(resp.Body).Decode(&response)
			assert.Equal(t, tt.expectedSuccess, response.Success)

			if tt.expectedSuccess {
				assert.NotEmpty(t, response.OrderNumber)
				assert.NotEmpty(t, response.PaymentURL)
				assert.Contains(t, response.TestStatus, "successful")
			} else {
				assert.NotEmpty(t, response.Message)
			}

			suite.paymentService.AssertExpectations(t)
		})
	}
}

func TestPaymentHandler_PaymentNotification_Integration(t *testing.T) {
	tests := []struct {
		name               string
		requestBody        map[string]any
		mockSetup          func(*MockPaymentService)
		expectedStatusCode int
		expectedSuccess    bool
	}{
		{
			name: "successful_webhook_processing",
			requestBody: map[string]any{
				"orderNumber": "ORD_WEBHOOK_123",
				"orderStatus": 2,
				"orderId":     "ALFA_ORDER_456",
				"amount":      10000,
				"currency":    "RUB",
				"checksum":    "valid_checksum",
			},
			mockSetup: func(service *MockPaymentService) {
				service.On("ProcessWebhookNotification", mock.Anything, mock.Anything).Return(nil)
			},
			expectedStatusCode: 200,
			expectedSuccess:    true,
		},
		{
			name:        "invalid_request_body",
			requestBody: map[string]any{},
			mockSetup: func(service *MockPaymentService) {
				service.On("ProcessWebhookNotification", mock.Anything, mock.Anything).Return(fmt.Errorf("invalid signature"))
			},
			expectedStatusCode: 500,
			expectedSuccess:    false,
		},
		{
			name: "webhook_processing_error",
			requestBody: map[string]any{
				"orderNumber": "ORD_ERROR_123",
				"orderStatus": 2,
			},
			mockSetup: func(service *MockPaymentService) {
				service.On("ProcessWebhookNotification", mock.Anything, mock.Anything).Return(fmt.Errorf("order not found"))
			},
			expectedStatusCode: 500,
			expectedSuccess:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suite := setupTestSuite()

			if tt.mockSetup != nil {
				tt.mockSetup(suite.paymentService)
			}

			// Create request
			bodyJSON, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/api/payment/notification", bytes.NewBuffer(bodyJSON))
			req.Header.Set("Content-Type", "application/json")

			// Execute request
			resp, _ := suite.app.Test(req, -1)

			// Assertions
			assert.Equal(t, tt.expectedStatusCode, resp.StatusCode)

			if tt.expectedStatusCode == 200 {
				var response map[string]any
				json.NewDecoder(resp.Body).Decode(&response)
				success, exists := response["success"]
				assert.True(t, exists)
				assert.Equal(t, tt.expectedSuccess, success)
			}

			suite.paymentService.AssertExpectations(t)
		})
	}
}

func TestPaymentHandler_PaymentReturn_Integration(t *testing.T) {
	tests := []struct {
		name               string
		orderNumber        string
		mockSetup          func(*MockPaymentService)
		expectedStatusCode int
		expectedLocation   string
	}{
		{
			name:        "successful_return",
			orderNumber: "ORD_RETURN_123",
			mockSetup: func(service *MockPaymentService) {
				service.On("ProcessPaymentReturn", mock.Anything, "ORD_RETURN_123").Return(nil)
			},
			expectedStatusCode: 302,
			expectedLocation:   "/payment/success?order=ORD_RETURN_123",
		},
		{
			name:               "missing_order_number",
			orderNumber:        "",
			mockSetup:          nil,
			expectedStatusCode: 400,
			expectedLocation:   "",
		},
		{
			name:        "processing_error",
			orderNumber: "ORD_ERROR_123",
			mockSetup: func(service *MockPaymentService) {
				service.On("ProcessPaymentReturn", mock.Anything, "ORD_ERROR_123").Return(fmt.Errorf("order not found"))
			},
			expectedStatusCode: 500,
			expectedLocation:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suite := setupTestSuite()

			if tt.mockSetup != nil {
				tt.mockSetup(suite.paymentService)
			}

			// Create request
			url := "/api/payment/return"
			if tt.orderNumber != "" {
				url += "?orderNumber=" + tt.orderNumber
			}
			req := httptest.NewRequest("GET", url, nil)

			// Execute request
			resp, _ := suite.app.Test(req, -1)

			// Assertions
			assert.Equal(t, tt.expectedStatusCode, resp.StatusCode)

			if tt.expectedStatusCode == 302 {
				location := resp.Header.Get("Location")
				assert.Equal(t, tt.expectedLocation, location)
			}

			suite.paymentService.AssertExpectations(t)
		})
	}
}

func TestPaymentHandler_GetAvailableOptions_Integration(t *testing.T) {
	t.Run("successful_options_retrieval", func(t *testing.T) {
		suite := setupTestSuite()

		suite.paymentService.On("GetAvailableOptions").Return(&AvailableOptionsResponse{
			Sizes: []SizeOption{
				{Value: "40x50", Label: "40×50 см", Description: "Стандартный", Price: 100.0},
			},
			Styles: []StyleOption{
				{Value: "max_colors", Label: "Максимум цветов", Description: "Полная цветовая палитра"},
			},
		})

		// Create request
		req := httptest.NewRequest("GET", "/api/payment/options", nil)

		// Execute request
		resp, _ := suite.app.Test(req, -1)

		// Assertions
		assert.Equal(t, 200, resp.StatusCode)

		var response AvailableOptionsResponse
		json.NewDecoder(resp.Body).Decode(&response)
		assert.Len(t, response.Sizes, 1)
		assert.Len(t, response.Styles, 1)
		assert.Equal(t, "40x50", response.Sizes[0].Value)
		assert.Equal(t, 100.0, response.Sizes[0].Price)

		suite.paymentService.AssertExpectations(t)
	})
}
