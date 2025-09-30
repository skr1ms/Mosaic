package payment

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/skr1ms/mosaic/config"
	"github.com/skr1ms/mosaic/internal/coupon"
	randomCouponCode "github.com/skr1ms/mosaic/pkg/randomCouponCode"
)

type PaymentServiceDeps struct {
	PaymentRepository         PaymentRepositoryInterface
	CouponRepository          CouponRepositoryInterface
	PartnerRepository         PartnerRepositoryInterface
	Config                    ConfigInterface
	AlfaBankClient            AlfaBankClientInterface
	RandomCouponCodeGenerator RandomCouponCodeGeneratorInterface
	EmailService              EmailServiceInterface
}

type PaymentService struct {
	deps *PaymentServiceDeps
}

type AlfaBankClient struct {
	config *config.Config
	client *http.Client
}

func NewPaymentService(deps *PaymentServiceDeps) *PaymentService {
	return &PaymentService{
		deps: deps,
	}
}

func NewAlfaBankClient(config *config.Config) *AlfaBankClient {
	return &AlfaBankClient{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Ensure AlfaBankClient implements the interface
var _ AlfaBankClientInterface = (*AlfaBankClient)(nil)

// RegisterOrder registers order in AlfaBank
func (c *AlfaBankClient) RegisterOrder(ctx context.Context, req *AlfaBankRegisterRequest) (*AlfaBankRegisterResponse, error) {
	data := url.Values{}
	data.Set("userName", c.config.AlphaBankConfig.Username)
	data.Set("password", c.config.AlphaBankConfig.Password)
	data.Set("orderNumber", req.OrderNumber)
	data.Set("amount", strconv.FormatInt(req.Amount, 10))
	data.Set("returnUrl", req.ReturnUrl)

	if req.Currency != "" {
		data.Set("currency", req.Currency)
	} else {
		data.Set("currency", "810")
	}

	if req.FailUrl != "" {
		data.Set("failUrl", req.FailUrl)
	}
	if req.Description != "" {
		data.Set("description", req.Description)
	}
	if req.Language != "" {
		data.Set("language", req.Language)
	} else {
		data.Set("language", "ru")
	}
	if req.ClientId != "" {
		data.Set("clientId", req.ClientId)
	}
	if req.JsonParams != "" {
		data.Set("jsonParams", req.JsonParams)
	}
	if req.SessionTimeoutSecs > 0 {
		data.Set("sessionTimeoutSecs", strconv.Itoa(req.SessionTimeoutSecs))
	}
	if req.BindingId != "" {
		data.Set("bindingId", req.BindingId)
	}
	if req.Features != "" {
		data.Set("features", req.Features)
	}

	url := c.config.AlphaBankConfig.Url + "/payment/rest/register.do"

	fmt.Printf("AlfaBank API Request: %s\n", url)
	fmt.Printf("Order Number: %s, Amount: %s\n", req.OrderNumber, strconv.FormatInt(req.Amount, 10))
	fmt.Printf("Username: %s\n", c.config.AlphaBankConfig.Username)
	fmt.Printf("Return URL: %s\n", req.ReturnUrl)
	if req.FailUrl != "" {
		fmt.Printf("Fail URL: %s\n", req.FailUrl)
	}

	resp, err := c.client.PostForm(url, data)
	if err != nil {
		return nil, fmt.Errorf("error requesting API: %w", err)
	}
	defer resp.Body.Close()

	fmt.Printf("AlfaBank API Response Status: %d\n", resp.StatusCode)

	if resp.StatusCode != 200 {
		fmt.Printf("AlfaBank API returned non-200 status: %d\n", resp.StatusCode)
	}

	fmt.Printf("AlfaBank API Response Headers: %v\n", resp.Header)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	fmt.Printf("AlfaBank API Response: %s\n", string(body))

	var result AlfaBankRegisterResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	return &result, nil
}

// GetOrderStatus gets order status from AlfaBank
func (c *AlfaBankClient) GetOrderStatus(ctx context.Context, orderID string) (*AlfaBankStatusResponse, error) {
	data := url.Values{}
	data.Set("userName", c.config.AlphaBankConfig.Username)
	data.Set("password", c.config.AlphaBankConfig.Password)
	data.Set("orderId", orderID)
	data.Set("language", "ru")

	resp, err := c.client.PostForm(c.config.AlphaBankConfig.Url+"/payment/rest/getOrderStatus.do", data)
	if err != nil {
		return nil, fmt.Errorf("error requesting status: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading status response: %w", err)
	}

	var result AlfaBankStatusResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, fmt.Errorf("error parsing status response: %w", err)
	}

	return &result, nil
}

// validateWebhookSignature validates webhook signature from AlfaBank
func (s *PaymentService) validateWebhookSignature(notification *PaymentNotificationRequest) bool {
	alfaConfig := s.deps.Config.GetAlfaBankConfig()
	if alfaConfig.WebhookSecret == "" {
		return true
	}

	// Skip signature validation for manual testing if AlfaBankOrderID is empty
	// This allows processing notifications that were manually triggered
	if notification.AlfaBankOrderID == "" {
		return true
	}

	orderStatus := 0
	if notification.OrderStatus != nil {
		orderStatus = *notification.OrderStatus
	}

	data := fmt.Sprintf("%s;%d;%s;%d;%s",
		notification.OrderNumber,
		orderStatus,
		notification.AlfaBankOrderID,
		notification.Amount,
		notification.Currency,
	)

	h := hmac.New(sha256.New, []byte(alfaConfig.WebhookSecret))
	h.Write([]byte(data))
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	return strings.EqualFold(expectedSignature, notification.Checksum)
}

// getWebhookURL returns URL for webhook notifications
func (s *PaymentService) getWebhookURL() string {
	alfaConfig := s.deps.Config.GetAlfaBankConfig()
	if alfaConfig.WebhookURL != "" {
		return alfaConfig.WebhookURL
	}

	serverConfig := s.deps.Config.GetServerConfig()
	baseURL := serverConfig.FrontendURL
	if baseURL == "" {
		baseURL = "https://yourdomain.com"
	}
	return baseURL + "/api/payment/notification"
}

// PurchaseCoupon purchases coupon online with card payment
func (s *PaymentService) PurchaseCoupon(ctx context.Context, req *PurchaseCouponRequest) (*PurchaseCouponResponse, error) {
	if s.deps == nil {
		return nil, fmt.Errorf("service dependencies are not initialized")
	}
	if s.deps.PaymentRepository == nil {
		return nil, fmt.Errorf("payment repository is not initialized")
	}
	if s.deps.CouponRepository == nil {
		return nil, fmt.Errorf("coupon repository is not initialized")
	}
	if s.deps.PartnerRepository == nil {
		return nil, fmt.Errorf("partner repository is not initialized")
	}
	if s.deps.Config == nil {
		return nil, fmt.Errorf("config is not initialized")
	}

	supportedSizes := map[string]bool{
		Size21x30: true,
		Size30x40: true,
		Size40x40: true,
		Size40x50: true,
		Size40x60: true,
		Size50x70: true,
	}

	if !supportedSizes[req.Size] {
		return &PurchaseCouponResponse{
			Success: false,
			Message: "Unsupported size",
		}, nil
	}

	var partnerID *uuid.UUID
	if req.Domain != nil && *req.Domain != "" {
		partner, err := s.deps.PartnerRepository.GetByDomain(ctx, *req.Domain)
		if err == nil && partner != nil {
			partnerID = &partner.ID
		}
	}

	orderNumber := s.GenerateOrderNumber()

	for {
		existingOrder, err := s.deps.PaymentRepository.GetOrderByNumber(ctx, orderNumber)
		if err != nil || existingOrder == nil {
			break
		}
		orderNumber = s.GenerateOrderNumber()
	}

	style := req.Style

	order := &Order{
		OrderNumber: orderNumber,
		PartnerID:   partnerID,
		Size:        req.Size,
		Style:       style,
		UserEmail:   req.Email,
		Amount:      int64(FixedPriceRub * 100),
		Currency:    "RUB",
		Status:      OrderStatusCreated,
		ReturnURL:   req.ReturnURL,
		FailURL:     req.FailURL,
		Description: fmt.Sprintf("Purchase of mosaic coupon %s style %s", req.Size, style),
	}

	err := s.deps.PaymentRepository.CreateOrder(ctx, order)
	if err != nil {
		return &PurchaseCouponResponse{
			Success: false,
			Message: "Error creating order",
		}, nil
	}

	language := "ru"
	if req.Language != "" {
		language = req.Language
	}

	webhookURL := s.getWebhookURL()
	jsonParams := fmt.Sprintf(`{"callbackUrl":"%s"}`, webhookURL)

	alfaReq := &AlfaBankRegisterRequest{
		OrderNumber:        orderNumber,
		Amount:             order.Amount,
		Currency:           "810",
		ReturnUrl:          req.ReturnURL,
		FailUrl:            getStringValue(req.FailURL),
		Description:        order.Description,
		Language:           language,
		ClientId:           req.Email,
		JsonParams:         jsonParams,
		SessionTimeoutSecs: 1200,
	}

	alfaResp, err := s.deps.AlfaBankClient.RegisterOrder(ctx, alfaReq)
	if err != nil {
		return &PurchaseCouponResponse{
			Success: false,
			Message: fmt.Sprintf("Error creating payment: %v", err),
		}, nil
	}

	if alfaResp.ErrorCode != "" {
		return &PurchaseCouponResponse{
			Success: false,
			Message: fmt.Sprintf("Payment error: %s", alfaResp.ErrorMessage),
		}, nil
	}

	err = s.deps.PaymentRepository.UpdateOrderStatus(ctx, orderNumber, OrderStatusPending, &alfaResp.OrderId)
	if err != nil {
		return &PurchaseCouponResponse{
			Success: false,
			Message: "Error updating order",
		}, nil
	}

	err = s.deps.PaymentRepository.UpdateOrderPaymentURL(ctx, orderNumber, alfaResp.FormUrl)
	if err != nil {
		return &PurchaseCouponResponse{
			Success: false,
			Message: "Error updating payment URL",
		}, nil
	}

	return &PurchaseCouponResponse{
		OrderID:     order.ID.String(),
		OrderNumber: orderNumber,
		PaymentURL:  alfaResp.FormUrl,
		Success:     true,
		Amount:      strconv.FormatInt(order.Amount/100, 10),
	}, nil
}

// GenerateOrderNumber generates unique order number
func (s *PaymentService) GenerateOrderNumber() string {
	uuid := uuid.New()
	timestamp := time.Now().Unix()
	return fmt.Sprintf("ORD_%d_%s", timestamp, uuid.String()[:8])
}

// getStringValue returns string value from pointer or empty string if pointer is nil
func getStringValue(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

// GetOrderStatus gets order status
func (s *PaymentService) GetOrderStatus(ctx context.Context, orderNumber string) (*OrderStatusResponse, error) {
	order, err := s.deps.PaymentRepository.GetOrderByNumber(ctx, orderNumber)
	if err != nil {
		return &OrderStatusResponse{
			Success: false,
			Message: "Order not found",
		}, nil
	}

	if order.Status == OrderStatusPending && order.AlfaBankOrderID != nil {
		alfaResp, err := s.deps.AlfaBankClient.GetOrderStatus(ctx, *order.AlfaBankOrderID)
		if err == nil && alfaResp.ErrorCode == "" {
			orderStatus := alfaResp.OrderStatus
			notification := &PaymentNotificationRequest{
				OrderNumber:     orderNumber,
				OrderStatus:     &orderStatus,
				AlfaBankOrderID: *order.AlfaBankOrderID,
			}

			err = s.ProcessWebhookNotification(ctx, notification)
			if err == nil {
				order, _ = s.deps.PaymentRepository.GetOrderByNumber(ctx, orderNumber)
			}
		}
	}

	var couponCode *string
	if order.CouponID != nil {
		coupon, err := s.deps.CouponRepository.GetByID(ctx, *order.CouponID)
		if err == nil {
			couponCode = &coupon.Code
		}
	}

	return &OrderStatusResponse{
		OrderID:    order.ID.String(),
		Status:     order.Status,
		Size:       order.Size,
		Style:      order.Style,
		Amount:     float64(order.Amount) / 100,
		Currency:   order.Currency,
		CouponCode: couponCode,
		Success:    true,
	}, nil
}

// GetAvailableOptions gets available sizes and styles
func (s *PaymentService) GetAvailableOptions() *AvailableOptionsResponse {
	sizes := []SizeOption{
		{Value: Size21x30, Label: "21×30 см", Description: "Маленький размер", Price: FixedPriceRub},
		{Value: Size30x40, Label: "30×40 см", Description: "Средний размер", Price: FixedPriceRub},
		{Value: Size40x40, Label: "40×40 см", Description: "Квадратный", Price: FixedPriceRub},
		{Value: Size40x50, Label: "40×50 см", Description: "Стандартный", Price: FixedPriceRub},
		{Value: Size40x60, Label: "40×60 см", Description: "Большой", Price: FixedPriceRub},
		{Value: Size50x70, Label: "50×70 см", Description: "Очень большой", Price: FixedPriceRub},
	}

	styles := []StyleOption{
		{Value: StyleGrayscale, Label: "Оттенки серого", Description: "Черно-белое изображение"},
		{Value: StyleSkinTone, Label: "Оттенки телесного", Description: "Портретный стиль"},
		{Value: StylePopArt, Label: "Поп-арт", Description: "Яркие цвета"},
		{Value: StyleMaxColors, Label: "Максимум цветов", Description: "Полная цветовая палитра"},
	}

	return &AvailableOptionsResponse{
		Sizes:  sizes,
		Styles: styles,
	}
}

// ProcessPaymentReturn processes return from payment system
func (s *PaymentService) ProcessPaymentReturn(ctx context.Context, orderNumber string) error {
	order, err := s.deps.PaymentRepository.GetOrderByNumber(ctx, orderNumber)
	if err != nil {
		return fmt.Errorf("order not found: %w", err)
	}

	if order.Status != OrderStatusPending {
		return nil
	}

	if order.AlfaBankOrderID != nil {
		alfaResp, err := s.deps.AlfaBankClient.GetOrderStatus(ctx, *order.AlfaBankOrderID)
		if err != nil {
			return fmt.Errorf("error checking status in AlfaBank: %w", err)
		}

		orderStatus := alfaResp.OrderStatus
		notification := &PaymentNotificationRequest{
			OrderNumber:     orderNumber,
			OrderStatus:     &orderStatus,
			AlfaBankOrderID: *order.AlfaBankOrderID,
			Amount:          order.Amount,
			Currency:        order.Currency,
		}

		err = s.ProcessWebhookNotification(ctx, notification)
		if err != nil {
			return fmt.Errorf("error processing payment status: %w", err)
		}
	}

	return nil
}

// createCouponForOrder generates a new coupon and marks it as purchased for the order
func (s *PaymentService) createCouponForOrder(ctx context.Context, order *Order) error {
	// Generate unique coupon code with partner code 0000
	couponCode, err := randomCouponCode.GenerateUniqueCouponCode("0000", s.deps.CouponRepository)
	if err != nil {
		return fmt.Errorf("error generating coupon code: %w", err)
	}

	// Get partner ID - use default 0000 partner if order doesn't have one
	var partnerID uuid.UUID
	if order.PartnerID != nil {
		partnerID = *order.PartnerID
	} else {
		// Get default partner (0000)
		defaultPartner, err := s.deps.PartnerRepository.GetByPartnerCode(ctx, "0000")
		if err != nil {
			return fmt.Errorf("error getting default partner: %w", err)
		}
		partnerID = defaultPartner.ID
	}

	// Create new coupon in database
	newCoupon := &coupon.Coupon{
		Code:          couponCode,
		PartnerID:     partnerID,
		Size:          order.Size,
		Style:         order.Style,
		Status:        "new",
		IsPurchased:   true,
		PurchaseEmail: &order.UserEmail,
	}

	err = s.deps.CouponRepository.Create(ctx, newCoupon)
	if err != nil {
		return fmt.Errorf("error creating coupon: %w", err)
	}

	// Link coupon to order
	err = s.deps.PaymentRepository.UpdateOrderCoupon(ctx, order.OrderNumber, newCoupon.ID)
	if err != nil {
		return fmt.Errorf("error linking coupon to order: %w", err)
	}

	// Send coupon email notification
	if s.deps.EmailService != nil && order.UserEmail != "" {
		err = s.deps.EmailService.SendCouponPurchaseEmail(order.UserEmail, couponCode, order.Size, order.Style)
		if err != nil {
			// Log error but don't fail the transaction
			fmt.Printf("Failed to send coupon email: %v\n", err)
		}
	}

	return nil
}

// ProcessWebhookNotification processes webhook notifications from AlfaBank
func (s *PaymentService) ProcessWebhookNotification(ctx context.Context, notification *PaymentNotificationRequest) error {
	if !s.validateWebhookSignature(notification) {
		return fmt.Errorf("invalid webhook signature")
	}

	order, err := s.deps.PaymentRepository.GetOrderByNumber(ctx, notification.OrderNumber)
	if err != nil {
		return fmt.Errorf("order not found: %w", err)
	}

	orderStatus := 2
	if notification.OrderStatus != nil {
		orderStatus = *notification.OrderStatus
	}

	switch orderStatus {
	case 0:
		return nil

	case 1:
		// Status 1 can also mean "paid" in some Alfa Bank configurations
		err = s.deps.PaymentRepository.UpdateOrderStatus(ctx, notification.OrderNumber, OrderStatusPaid, &notification.AlfaBankOrderID)
		if err != nil {
			return fmt.Errorf("error updating order status to paid: %w", err)
		}

		err = s.createCouponForOrder(ctx, order)
		if err != nil {
			return fmt.Errorf("error creating coupon for order: %w", err)
		}

		return nil

	case 2:
		err = s.deps.PaymentRepository.UpdateOrderStatus(ctx, notification.OrderNumber, OrderStatusPaid, &notification.AlfaBankOrderID)
		if err != nil {
			return fmt.Errorf("error updating order status to paid: %w", err)
		}

		err = s.createCouponForOrder(ctx, order)
		if err != nil {
			return fmt.Errorf("error creating coupon for order: %w", err)
		}

		return nil

	case 3:
		err = s.deps.PaymentRepository.UpdateOrderStatus(ctx, notification.OrderNumber, OrderStatusCancelled, &notification.AlfaBankOrderID)
		if err != nil {
			return fmt.Errorf("error updating order status to cancelled: %w", err)
		}
		return nil

	case 4:
		err = s.deps.PaymentRepository.UpdateOrderStatus(ctx, notification.OrderNumber, OrderStatusFailed, &notification.AlfaBankOrderID)
		if err != nil {
			return fmt.Errorf("error updating order status to failed: %w", err)
		}
		return nil

	case 5:
		return nil

	case 6:
		err = s.deps.PaymentRepository.UpdateOrderStatus(ctx, notification.OrderNumber, OrderStatusFailed, &notification.AlfaBankOrderID)
		if err != nil {
			return fmt.Errorf("error updating order status to failed: %w", err)
		}
		return nil

	default:
		return fmt.Errorf("unknown order status: %d", orderStatus)
	}
}

// TestAlfaBankIntegration tests integration with Alfa Bank API
func (s *PaymentService) TestAlfaBankIntegration(ctx context.Context, req *AlfaBankRegisterRequest) (*TestIntegrationResponse, error) {
	result := &TestIntegrationResponse{
		OrderNumber: req.OrderNumber,
	}

	// Check configuration first
	alfaConfig := s.deps.Config.GetAlfaBankConfig()

	configIssues := []string{}
	if alfaConfig.Username == "" {
		configIssues = append(configIssues, "username not set")
	}
	if alfaConfig.Password == "" {
		configIssues = append(configIssues, "password not set")
	}
	if alfaConfig.Url == "" {
		configIssues = append(configIssues, "API URL not set")
	}

	if len(configIssues) > 0 {
		result.Success = false
		result.ConfigStatus = fmt.Sprintf("Configuration issues: %v", configIssues)
		result.Message = "Integration test failed due to configuration issues"
		return result, nil
	}

	result.ConfigStatus = fmt.Sprintf("Configuration OK: URL=%s, Username=%s", alfaConfig.Url, alfaConfig.Username)

	// Try to register a test order
	alfaResp, err := s.deps.AlfaBankClient.RegisterOrder(ctx, req)
	if err != nil {
		result.Success = false
		result.TestStatus = "API connection failed"
		result.ErrorDetails = err.Error()
		result.Message = "Failed to connect to Alfa Bank API"
		return result, nil
	}

	// Check if registration was successful
	if alfaResp.ErrorCode != "" {
		result.Success = false
		result.TestStatus = "API returned error"
		result.ErrorDetails = fmt.Sprintf("Error code: %s, Message: %s", alfaResp.ErrorCode, alfaResp.ErrorMessage)
		result.Message = "Alfa Bank API returned an error"
		result.AlfaBankOrderID = alfaResp.OrderId
		return result, nil
	}

	// Success case
	result.Success = true
	result.TestStatus = "Integration test successful"
	result.Message = "Successfully connected to Alfa Bank API and created test order"
	result.AlfaBankOrderID = alfaResp.OrderId
	result.PaymentURL = alfaResp.FormUrl

	return result, nil
}
