package payment

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// TestHelpers provides utility functions for testing payment integration
type TestHelpers struct {
	BaseURL string
}

// NewTestHelpers creates a new instance of test helpers
func NewTestHelpers(baseURL string) *TestHelpers {
	return &TestHelpers{
		BaseURL: baseURL,
	}
}

// toStringPtr returns a pointer to a string
func toStringPtr(s string) *string {
	return &s
}

// CreateTestPurchaseRequest creates a valid test purchase request
func (h *TestHelpers) CreateTestPurchaseRequest(email string, size string, style string) *PurchaseCouponRequest {
	return &PurchaseCouponRequest{
		Size:      size,
		Style:     style,
		Email:     email,
		ReturnURL: h.BaseURL + "/payment/success",
		FailURL:   toStringPtr(h.BaseURL + "/payment/fail"),
		Language:  "ru",
	}
}

// CreateTestWebhookRequest creates a test webhook notification request
func (h *TestHelpers) CreateTestWebhookRequest(orderNumber string, status int) *PaymentNotificationRequest {
	return &PaymentNotificationRequest{
		OrderNumber:      orderNumber,
		OrderStatus:      &status,
		AlfaBankOrderID:  fmt.Sprintf("ALFA_%s", uuid.New().String()[:8]),
		Amount:           10000, // 100 rubles in kopecks
		Currency:         "RUB",
		Date:             time.Now().Unix(),
		IP:               "127.0.0.1",
		OrderDescription: fmt.Sprintf("Test order %s", orderNumber),
		Checksum:         "test_checksum_" + orderNumber,
	}
}

// TestPurchaseCoupon performs a test coupon purchase
func (h *TestHelpers) TestPurchaseCoupon(req *PurchaseCouponRequest) (*PurchaseCouponResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(h.BaseURL+"/api/payment/purchase", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	var response PurchaseCouponResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

// TestGetOrderStatus performs a test order status check
func (h *TestHelpers) TestGetOrderStatus(orderNumber string) (*OrderStatusResponse, error) {
	resp, err := http.Get(h.BaseURL + "/api/payment/orders/" + orderNumber + "/status")
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	var response OrderStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

// TestIntegration performs a test of Alfa Bank integration
func (h *TestHelpers) TestIntegration() (*TestIntegrationResponse, error) {
	resp, err := http.Get(h.BaseURL + "/api/payment/test-integration")
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	var response TestIntegrationResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

// TestWebhook sends a test webhook notification
func (h *TestHelpers) TestWebhook(req *PaymentNotificationRequest) (map[string]any, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(h.BaseURL+"/api/payment/notification", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	var response map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return response, nil
}

// GetAvailableOptions gets available payment options
func (h *TestHelpers) GetAvailableOptions() (*AvailableOptionsResponse, error) {
	resp, err := http.Get(h.BaseURL + "/api/payment/options")
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	var response AvailableOptionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

// SimulatePaymentReturn simulates a payment return from Alfa Bank
func (h *TestHelpers) SimulatePaymentReturn(orderNumber string) error {
	resp, err := http.Get(h.BaseURL + "/api/payment/return?orderNumber=" + orderNumber)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// TestScenarios provides common test scenarios
type TestScenarios struct {
	helpers *TestHelpers
}

// NewTestScenarios creates a new test scenarios instance
func NewTestScenarios(baseURL string) *TestScenarios {
	return &TestScenarios{
		helpers: NewTestHelpers(baseURL),
	}
}

// EndToEndSuccessfulPayment simulates a complete successful payment flow
func (ts *TestScenarios) EndToEndSuccessfulPayment(customerEmail string) (*TestScenarioResult, error) {
	result := &TestScenarioResult{
		Steps: []TestStep{},
	}

	// Step 1: Get available options
	result.AddStep("get_options", "Getting available payment options")
	options, err := ts.helpers.GetAvailableOptions()
	if err != nil {
		result.SetError("get_options", err)
		return result, err
	}
	result.SetSuccess("get_options", fmt.Sprintf("Found %d sizes and %d styles", len(options.Sizes), len(options.Styles)))

	// Step 2: Create purchase request
	result.AddStep("create_purchase", "Creating coupon purchase")
	purchaseReq := ts.helpers.CreateTestPurchaseRequest(customerEmail, options.Sizes[0].Value, options.Styles[0].Value)

	// Step 3: Purchase coupon
	result.AddStep("purchase_coupon", "Processing coupon purchase")
	purchaseResp, err := ts.helpers.TestPurchaseCoupon(purchaseReq)
	if err != nil {
		result.SetError("purchase_coupon", err)
		return result, err
	}
	if !purchaseResp.Success {
		err = fmt.Errorf("purchase failed: %s", purchaseResp.Message)
		result.SetError("purchase_coupon", err)
		return result, err
	}
	result.SetSuccess("purchase_coupon", fmt.Sprintf("Order created: %s", purchaseResp.OrderNumber))
	result.OrderNumber = purchaseResp.OrderNumber

	// Step 4: Check initial order status
	result.AddStep("check_initial_status", "Checking initial order status")
	statusResp, err := ts.helpers.TestGetOrderStatus(result.OrderNumber)
	if err != nil {
		result.SetError("check_initial_status", err)
		return result, err
	}
	result.SetSuccess("check_initial_status", fmt.Sprintf("Order status: %s", statusResp.Status))

	// Step 5: Simulate webhook notification (payment successful)
	result.AddStep("webhook_notification", "Simulating payment webhook")
	webhookReq := ts.helpers.CreateTestWebhookRequest(result.OrderNumber, 2) // Status 2 = paid
	_, err = ts.helpers.TestWebhook(webhookReq)
	if err != nil {
		result.SetError("webhook_notification", err)
		return result, err
	}
	result.SetSuccess("webhook_notification", "Webhook processed successfully")

	// Step 6: Check final order status
	result.AddStep("check_final_status", "Checking final order status")
	finalStatus, err := ts.helpers.TestGetOrderStatus(result.OrderNumber)
	if err != nil {
		result.SetError("check_final_status", err)
		return result, err
	}

	if finalStatus.CouponCode != nil && *finalStatus.CouponCode != "" {
		result.SetSuccess("check_final_status", fmt.Sprintf("Order paid, coupon generated: %s", *finalStatus.CouponCode))
		result.CouponCode = *finalStatus.CouponCode
	} else {
		result.SetSuccess("check_final_status", "Order processed successfully")
	}

	result.Success = true
	return result, nil
}

// TestIntegrationScenario tests the Alfa Bank integration endpoint
func (ts *TestScenarios) TestIntegrationScenario() (*TestScenarioResult, error) {
	result := &TestScenarioResult{
		Steps: []TestStep{},
	}

	result.AddStep("test_integration", "Testing Alfa Bank integration")
	integrationResp, err := ts.helpers.TestIntegration()
	if err != nil {
		result.SetError("test_integration", err)
		return result, err
	}

	if integrationResp.Success {
		result.SetSuccess("test_integration", fmt.Sprintf("Integration successful: %s", integrationResp.Message))
		result.Success = true
	} else {
		result.SetError("test_integration", fmt.Errorf("integration failed: %s", integrationResp.Message))
	}

	return result, nil
}

// TestScenarioResult represents the result of a test scenario
type TestScenarioResult struct {
	Success     bool       `json:"success"`
	OrderNumber string     `json:"order_number,omitempty"`
	CouponCode  string     `json:"coupon_code,omitempty"`
	Steps       []TestStep `json:"steps"`
}

// TestStep represents a single step in a test scenario
type TestStep struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Success     bool      `json:"success"`
	Message     string    `json:"message"`
	Timestamp   time.Time `json:"timestamp"`
}

// AddStep adds a new test step
func (r *TestScenarioResult) AddStep(name, description string) {
	r.Steps = append(r.Steps, TestStep{
		Name:        name,
		Description: description,
		Success:     false,
		Timestamp:   time.Now(),
	})
}

// SetSuccess marks a step as successful
func (r *TestScenarioResult) SetSuccess(name, message string) {
	for i := range r.Steps {
		if r.Steps[i].Name == name {
			r.Steps[i].Success = true
			r.Steps[i].Message = message
			break
		}
	}
}

// SetError marks a step as failed
func (r *TestScenarioResult) SetError(name string, err error) {
	for i := range r.Steps {
		if r.Steps[i].Name == name {
			r.Steps[i].Success = false
			r.Steps[i].Message = err.Error()
			break
		}
	}
}

// GetSummary returns a summary of the test scenario
func (r *TestScenarioResult) GetSummary() string {
	successCount := 0
	for _, step := range r.Steps {
		if step.Success {
			successCount++
		}
	}

	return fmt.Sprintf("Test completed: %d/%d steps successful", successCount, len(r.Steps))
}
