package payment

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/skr1ms/mosaic/config"
	"github.com/skr1ms/mosaic/internal/coupon"
	"github.com/skr1ms/mosaic/internal/partner"
)

type MockPaymentRepository struct {
	mock.Mock
}

var _ PaymentRepositoryInterface = (*MockPaymentRepository)(nil)

func (m *MockPaymentRepository) CreateOrder(ctx context.Context, order *Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockPaymentRepository) GetOrderByNumber(ctx context.Context, orderNumber string) (*Order, error) {
	args := m.Called(ctx, orderNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Order), args.Error(1)
}

func (m *MockPaymentRepository) GetOrderByID(ctx context.Context, id uuid.UUID) (*Order, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Order), args.Error(1)
}

func (m *MockPaymentRepository) UpdateOrderStatus(ctx context.Context, orderNumber string, status string, alfaBankOrderID *string) error {
	args := m.Called(ctx, orderNumber, status, alfaBankOrderID)
	return args.Error(0)
}

func (m *MockPaymentRepository) UpdateOrderPaymentURL(ctx context.Context, orderNumber string, paymentURL string) error {
	args := m.Called(ctx, orderNumber, paymentURL)
	return args.Error(0)
}

func (m *MockPaymentRepository) UpdateOrderCoupon(ctx context.Context, orderNumber string, couponID uuid.UUID) error {
	args := m.Called(ctx, orderNumber, couponID)
	return args.Error(0)
}

func (m *MockPaymentRepository) GetOrdersByEmail(ctx context.Context, email string, limit int) ([]Order, error) {
	args := m.Called(ctx, email, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]Order), args.Error(1)
}

func (m *MockPaymentRepository) GetOrdersByPartner(ctx context.Context, partnerID uuid.UUID, limit int) ([]Order, error) {
	args := m.Called(ctx, partnerID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]Order), args.Error(1)
}

func (m *MockPaymentRepository) GetOrdersCountByStatus(ctx context.Context, status string) (int, error) {
	args := m.Called(ctx, status)
	return args.Int(0), args.Error(1)
}

func (m *MockPaymentRepository) GetOrdersCountByPartner(ctx context.Context, partnerID uuid.UUID, status string) (int, error) {
	args := m.Called(ctx, partnerID, status)
	return args.Int(0), args.Error(1)
}

type MockCouponRepository struct {
	mock.Mock
}

var _ CouponRepositoryInterface = (*MockCouponRepository)(nil)

func (m *MockCouponRepository) Create(ctx context.Context, coupon *coupon.Coupon) error {
	args := m.Called(ctx, coupon)
	return args.Error(0)
}

func (m *MockCouponRepository) GetByID(ctx context.Context, id uuid.UUID) (*coupon.Coupon, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*coupon.Coupon), args.Error(1)
}

func (m *MockCouponRepository) GetByCode(ctx context.Context, code string) (*coupon.Coupon, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*coupon.Coupon), args.Error(1)
}

func (m *MockCouponRepository) Update(ctx context.Context, coupon *coupon.Coupon) error {
	args := m.Called(ctx, coupon)
	return args.Error(0)
}

func (m *MockCouponRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCouponRepository) GetAll(ctx context.Context) ([]*coupon.Coupon, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*coupon.Coupon), args.Error(1)
}

func (m *MockCouponRepository) GetByPartnerID(ctx context.Context, partnerID uuid.UUID) ([]*coupon.Coupon, error) {
	args := m.Called(ctx, partnerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*coupon.Coupon), args.Error(1)
}

func (m *MockCouponRepository) CodeExists(ctx context.Context, code string) (bool, error) {
	args := m.Called(ctx, code)
	return args.Bool(0), args.Error(1)
}

type MockPartnerRepository struct {
	mock.Mock
}

var _ PartnerRepositoryInterface = (*MockPartnerRepository)(nil)

func (m *MockPartnerRepository) Create(ctx context.Context, partner *partner.Partner) error {
	args := m.Called(ctx, partner)
	return args.Error(0)
}

func (m *MockPartnerRepository) GetByID(ctx context.Context, id uuid.UUID) (*partner.Partner, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*partner.Partner), args.Error(1)
}

func (m *MockPartnerRepository) GetByDomain(ctx context.Context, domain string) (*partner.Partner, error) {
	args := m.Called(ctx, domain)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*partner.Partner), args.Error(1)
}

func (m *MockPartnerRepository) Update(ctx context.Context, partner *partner.Partner) error {
	args := m.Called(ctx, partner)
	return args.Error(0)
}

func (m *MockPartnerRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPartnerRepository) GetAll(ctx context.Context, sortBy string, order string) ([]*partner.Partner, error) {
	args := m.Called(ctx, sortBy, order)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*partner.Partner), args.Error(1)
}

type MockAlfaBankClient struct {
	mock.Mock
}

var _ AlfaBankClientInterface = (*MockAlfaBankClient)(nil)

func (m *MockAlfaBankClient) RegisterOrder(ctx context.Context, req *AlfaBankRegisterRequest) (*AlfaBankRegisterResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*AlfaBankRegisterResponse), args.Error(1)
}

func (m *MockAlfaBankClient) GetOrderStatus(ctx context.Context, orderID string) (*AlfaBankStatusResponse, error) {
	args := m.Called(ctx, orderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*AlfaBankStatusResponse), args.Error(1)
}

func createTestConfig() *config.Config {
	return &config.Config{
		AlphaBankConfig: config.AlphaBankConfig{
			Username:      "test_user",
			Password:      "test_password",
			Url:           "https://test.alfabank.ru",
			WebhookSecret: "",
			WebhookURL:    "https://test.example.com/webhook",
		},
		ServerConfig: config.ServerConfig{
			FrontendURL: "https://test.example.com",
		},
	}
}

func createTestOrder() *Order {
	return &Order{
		ID:          uuid.New(),
		OrderNumber: "ORD_1234567890_abcd1234",
		Size:        Size40x50,
		Style:       StyleGrayscale,
		UserEmail:   "test@example.com",
		Amount:      10000,
		Currency:    "RUB",
		Status:      OrderStatusCreated,
		ReturnURL:   "https://test.example.com/return",
		Description: "Test order",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func createTestPartner() *partner.Partner {
	return &partner.Partner{
		ID:             uuid.New(),
		BrandName:      "Test Partner",
		Domain:         "test.example.com",
		PartnerCode:    "TEST",
		LogoURL:        "https://example.com/logo.png",
		Email:          "contact@test.example.com",
		Phone:          "+1234567890",
		Address:        "123 Test Street",
		AllowPurchases: true,
	}
}

func createTestCoupon() *coupon.Coupon {
	return &coupon.Coupon{
		ID:          uuid.New(),
		Code:        "TEST-1234-5678",
		Size:        Size40x50,
		Style:       StyleGrayscale,
		Status:      "new",
		IsPurchased: true,
	}
}

func TestPaymentService_PurchaseCoupon_Success(t *testing.T) {
	mockPaymentRepo := &MockPaymentRepository{}
	mockCouponRepo := &MockCouponRepository{}
	mockPartnerRepo := &MockPartnerRepository{}

	deps := &PaymentServiceDeps{
		PaymentRepository: mockPaymentRepo,
		CouponRepository:  mockCouponRepo,
		PartnerRepository: mockPartnerRepo,
		Config:            createTestConfig(),
	}

	service := NewPaymentService(deps)

	mockAlfaClient := &MockAlfaBankClient{}
	service.alfaClient = mockAlfaClient

	req := &PurchaseCouponRequest{
		Size:      Size40x50,
		Style:     StyleGrayscale,
		Email:     "test@example.com",
		ReturnURL: "https://test.example.com/return",
	}

	alfaResponse := &AlfaBankRegisterResponse{
		OrderId: "alfa_order_123",
		FormUrl: "https://payment.alfabank.ru/payment/order123",
	}

	mockPaymentRepo.On("GetOrderByNumber", mock.Anything, mock.AnythingOfType("string")).Return(nil, errors.New("not found"))
	mockPaymentRepo.On("CreateOrder", mock.Anything, mock.AnythingOfType("*payment.Order")).Return(nil)
	mockAlfaClient.On("RegisterOrder", mock.Anything, mock.AnythingOfType("*payment.AlfaBankRegisterRequest")).Return(alfaResponse, nil)
	mockPaymentRepo.On("UpdateOrderStatus", mock.Anything, mock.AnythingOfType("string"), OrderStatusPending, &alfaResponse.OrderId).Return(nil)
	mockPaymentRepo.On("UpdateOrderPaymentURL", mock.Anything, mock.AnythingOfType("string"), alfaResponse.FormUrl).Return(nil)

	result, err := service.PurchaseCoupon(context.Background(), req)

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, alfaResponse.FormUrl, result.PaymentURL)
	assert.NotEmpty(t, result.OrderID)

	mockPaymentRepo.AssertExpectations(t)
	mockAlfaClient.AssertExpectations(t)
}

func TestPaymentService_PurchaseCoupon_UnsupportedSize(t *testing.T) {
	deps := &PaymentServiceDeps{
		Config: createTestConfig(),
	}
	service := NewPaymentService(deps)

	req := &PurchaseCouponRequest{
		Size:      "invalid_size",
		Style:     StyleGrayscale,
		Email:     "test@example.com",
		ReturnURL: "https://test.example.com/return",
	}

	result, err := service.PurchaseCoupon(context.Background(), req)

	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Equal(t, "Unsupported size", result.Message)
}

func TestPaymentService_PurchaseCoupon_WithPartnerDomain(t *testing.T) {
	mockPaymentRepo := &MockPaymentRepository{}
	mockCouponRepo := &MockCouponRepository{}
	mockPartnerRepo := &MockPartnerRepository{}

	deps := &PaymentServiceDeps{
		PaymentRepository: mockPaymentRepo,
		CouponRepository:  mockCouponRepo,
		PartnerRepository: mockPartnerRepo,
		Config:            createTestConfig(),
	}

	service := NewPaymentService(deps)
	mockAlfaClient := &MockAlfaBankClient{}
	service.alfaClient = mockAlfaClient

	testPartner := createTestPartner()
	domain := "test.example.com"

	req := &PurchaseCouponRequest{
		Size:      Size40x50,
		Style:     StyleGrayscale,
		Email:     "test@example.com",
		ReturnURL: "https://test.example.com/return",
		Domain:    &domain,
	}

	alfaResponse := &AlfaBankRegisterResponse{
		OrderId: "alfa_order_123",
		FormUrl: "https://payment.alfabank.ru/payment/order123",
	}

	mockPartnerRepo.On("GetByDomain", mock.Anything, domain).Return(testPartner, nil)
	mockPaymentRepo.On("GetOrderByNumber", mock.Anything, mock.AnythingOfType("string")).Return(nil, errors.New("not found"))
	mockPaymentRepo.On("CreateOrder", mock.Anything, mock.MatchedBy(func(order *Order) bool {
		return order.PartnerID != nil && *order.PartnerID == testPartner.ID
	})).Return(nil)
	mockAlfaClient.On("RegisterOrder", mock.Anything, mock.AnythingOfType("*payment.AlfaBankRegisterRequest")).Return(alfaResponse, nil)
	mockPaymentRepo.On("UpdateOrderStatus", mock.Anything, mock.AnythingOfType("string"), OrderStatusPending, &alfaResponse.OrderId).Return(nil)
	mockPaymentRepo.On("UpdateOrderPaymentURL", mock.Anything, mock.AnythingOfType("string"), alfaResponse.FormUrl).Return(nil)

	result, err := service.PurchaseCoupon(context.Background(), req)

	require.NoError(t, err)
	assert.True(t, result.Success)

	mockPartnerRepo.AssertExpectations(t)
	mockPaymentRepo.AssertExpectations(t)
	mockAlfaClient.AssertExpectations(t)
}

func TestPaymentService_GetOrderStatus_Success(t *testing.T) {
	mockPaymentRepo := &MockPaymentRepository{}
	mockCouponRepo := &MockCouponRepository{}

	deps := &PaymentServiceDeps{
		PaymentRepository: mockPaymentRepo,
		CouponRepository:  mockCouponRepo,
		Config:            createTestConfig(),
	}

	service := NewPaymentService(deps)

	testOrder := createTestOrder()
	testOrder.Status = OrderStatusPaid
	testCoupon := createTestCoupon()
	testOrder.CouponID = &testCoupon.ID

	mockPaymentRepo.On("GetOrderByNumber", mock.Anything, testOrder.OrderNumber).Return(testOrder, nil)
	mockCouponRepo.On("GetByID", mock.Anything, testCoupon.ID).Return(testCoupon, nil)

	result, err := service.GetOrderStatus(context.Background(), testOrder.OrderNumber)

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, testOrder.ID.String(), result.OrderID)
	assert.Equal(t, OrderStatusPaid, result.Status)
	assert.Equal(t, testCoupon.Code, *result.CouponCode)
	assert.Equal(t, float64(testOrder.Amount)/100, result.Amount)

	mockPaymentRepo.AssertExpectations(t)
	mockCouponRepo.AssertExpectations(t)
}

func TestPaymentService_GetOrderStatus_NotFound(t *testing.T) {
	mockPaymentRepo := &MockPaymentRepository{}

	deps := &PaymentServiceDeps{
		PaymentRepository: mockPaymentRepo,
		Config:            createTestConfig(),
	}

	service := NewPaymentService(deps)

	mockPaymentRepo.On("GetOrderByNumber", mock.Anything, "nonexistent").Return(nil, errors.New("not found"))

	result, err := service.GetOrderStatus(context.Background(), "nonexistent")

	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Equal(t, "Order not found", result.Message)

	mockPaymentRepo.AssertExpectations(t)
}

func TestPaymentService_GetAvailableOptions(t *testing.T) {
	deps := &PaymentServiceDeps{
		Config: createTestConfig(),
	}
	service := NewPaymentService(deps)

	options := service.GetAvailableOptions()

	require.NotNil(t, options)
	assert.Len(t, options.Sizes, 6)
	assert.Len(t, options.Styles, 4)

	sizeValues := make([]string, len(options.Sizes))
	for i, size := range options.Sizes {
		sizeValues[i] = size.Value
		assert.Equal(t, FixedPriceRub, size.Price)
		assert.NotEmpty(t, size.Label)
		assert.NotEmpty(t, size.Description)
	}

	expectedSizes := []string{Size21x30, Size30x40, Size40x40, Size40x50, Size40x60, Size50x70}
	for _, expectedSize := range expectedSizes {
		assert.Contains(t, sizeValues, expectedSize)
	}

	styleValues := make([]string, len(options.Styles))
	for i, style := range options.Styles {
		styleValues[i] = style.Value
		assert.NotEmpty(t, style.Label)
		assert.NotEmpty(t, style.Description)
	}

	expectedStyles := []string{StyleGrayscale, StyleSkinTone, StylePopArt, StyleMaxColors}
	for _, expectedStyle := range expectedStyles {
		assert.Contains(t, styleValues, expectedStyle)
	}
}

func TestPaymentService_ProcessWebhookNotification_SuccessfulPayment(t *testing.T) {
	mockPaymentRepo := &MockPaymentRepository{}
	mockCouponRepo := &MockCouponRepository{}
	mockPartnerRepo := &MockPartnerRepository{}

	deps := &PaymentServiceDeps{
		PaymentRepository: mockPaymentRepo,
		CouponRepository:  mockCouponRepo,
		PartnerRepository: mockPartnerRepo,
		Config:            createTestConfig(),
	}

	service := NewPaymentService(deps)

	testOrder := createTestOrder()
	testOrder.Status = OrderStatusPending
	alfaBankOrderID := "alfa_order_123"
	testOrder.AlfaBankOrderID = &alfaBankOrderID

	notification := &PaymentNotificationRequest{
		OrderNumber:     testOrder.OrderNumber,
		OrderStatus:     2,
		AlfaBankOrderID: alfaBankOrderID,
		Amount:          testOrder.Amount,
		Currency:        testOrder.Currency,
	}

	mockPaymentRepo.On("GetOrderByNumber", mock.Anything, testOrder.OrderNumber).Return(testOrder, nil)
	mockPaymentRepo.On("UpdateOrderStatus", mock.Anything, testOrder.OrderNumber, OrderStatusPaid, &alfaBankOrderID).Return(nil)
	mockCouponRepo.On("CodeExists", mock.Anything, mock.AnythingOfType("string")).Return(false, nil)
	mockCouponRepo.On("Create", mock.Anything, mock.AnythingOfType("*coupon.Coupon")).Return(nil)
	mockPaymentRepo.On("UpdateOrderCoupon", mock.Anything, testOrder.OrderNumber, mock.AnythingOfType("uuid.UUID")).Return(nil)

	err := service.ProcessWebhookNotification(context.Background(), notification)

	require.NoError(t, err)
	mockPaymentRepo.AssertExpectations(t)
	mockCouponRepo.AssertExpectations(t)
}

func TestPaymentService_ProcessWebhookNotification_FailedPayment(t *testing.T) {
	mockPaymentRepo := &MockPaymentRepository{}

	deps := &PaymentServiceDeps{
		PaymentRepository: mockPaymentRepo,
		Config:            createTestConfig(),
	}

	service := NewPaymentService(deps)

	testOrder := createTestOrder()
	testOrder.Status = OrderStatusPending
	alfaBankOrderID := "alfa_order_123"

	notification := &PaymentNotificationRequest{
		OrderNumber:     testOrder.OrderNumber,
		OrderStatus:     6,
		AlfaBankOrderID: alfaBankOrderID,
	}

	mockPaymentRepo.On("GetOrderByNumber", mock.Anything, testOrder.OrderNumber).Return(testOrder, nil)
	mockPaymentRepo.On("UpdateOrderStatus", mock.Anything, testOrder.OrderNumber, OrderStatusFailed, &alfaBankOrderID).Return(nil)

	err := service.ProcessWebhookNotification(context.Background(), notification)

	require.NoError(t, err)
	mockPaymentRepo.AssertExpectations(t)
}

func TestPaymentService_ProcessWebhookNotification_OrderNotFound(t *testing.T) {
	mockPaymentRepo := &MockPaymentRepository{}

	deps := &PaymentServiceDeps{
		PaymentRepository: mockPaymentRepo,
		Config:            createTestConfig(),
	}

	service := NewPaymentService(deps)

	notification := &PaymentNotificationRequest{
		OrderNumber:     "nonexistent",
		OrderStatus:     2,
		AlfaBankOrderID: "alfa_order_123",
	}

	mockPaymentRepo.On("GetOrderByNumber", mock.Anything, "nonexistent").Return(nil, errors.New("not found"))

	err := service.ProcessWebhookNotification(context.Background(), notification)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "order not found")
	mockPaymentRepo.AssertExpectations(t)
}

func TestPaymentService_ProcessPaymentReturn_Success(t *testing.T) {
	mockPaymentRepo := &MockPaymentRepository{}
	mockCouponRepo := &MockCouponRepository{}
	mockPartnerRepo := &MockPartnerRepository{}

	deps := &PaymentServiceDeps{
		PaymentRepository: mockPaymentRepo,
		CouponRepository:  mockCouponRepo,
		PartnerRepository: mockPartnerRepo,
		Config:            createTestConfig(),
	}

	service := NewPaymentService(deps)
	mockAlfaClient := &MockAlfaBankClient{}
	service.alfaClient = mockAlfaClient

	testOrder := createTestOrder()
	testOrder.Status = OrderStatusPending
	alfaBankOrderID := "alfa_order_123"
	testOrder.AlfaBankOrderID = &alfaBankOrderID

	alfaStatusResponse := &AlfaBankStatusResponse{
		OrderNumber: testOrder.OrderNumber,
		OrderStatus: 2,
	}

	mockPaymentRepo.On("GetOrderByNumber", mock.Anything, testOrder.OrderNumber).Return(testOrder, nil)
	mockAlfaClient.On("GetOrderStatus", mock.Anything, alfaBankOrderID).Return(alfaStatusResponse, nil)
	mockPaymentRepo.On("UpdateOrderStatus", mock.Anything, testOrder.OrderNumber, OrderStatusPaid, &alfaBankOrderID).Return(nil)

	mockCouponRepo.On("CodeExists", mock.Anything, mock.AnythingOfType("string")).Return(false, nil)
	mockCouponRepo.On("Create", mock.Anything, mock.AnythingOfType("*coupon.Coupon")).Return(nil)
	mockPaymentRepo.On("UpdateOrderCoupon", mock.Anything, testOrder.OrderNumber, mock.AnythingOfType("uuid.UUID")).Return(nil)

	err := service.ProcessPaymentReturn(context.Background(), testOrder.OrderNumber)

	require.NoError(t, err)
	mockPaymentRepo.AssertExpectations(t)
	mockCouponRepo.AssertExpectations(t)
	mockPartnerRepo.AssertExpectations(t)
	mockAlfaClient.AssertExpectations(t)
}

func TestPaymentService_generateUniqueCouponCode(t *testing.T) {
	mockCouponRepo := &MockCouponRepository{}

	deps := &PaymentServiceDeps{
		CouponRepository: mockCouponRepo,
		Config:           createTestConfig(),
	}

	service := NewPaymentService(deps)

	partnerCode := "TEST"

	mockCouponRepo.On("CodeExists", mock.Anything, mock.AnythingOfType("string")).Return(true, nil).Once()
	mockCouponRepo.On("CodeExists", mock.Anything, mock.AnythingOfType("string")).Return(false, nil).Once()

	code, err := service.generateUniqueCouponCode(partnerCode)

	require.NoError(t, err)
	assert.NotEmpty(t, code)
	assert.Contains(t, code, partnerCode)
	assert.Contains(t, code, "-")

	mockCouponRepo.AssertExpectations(t)
}

func TestPaymentService_generateOrderNumber(t *testing.T) {
	deps := &PaymentServiceDeps{
		Config: createTestConfig(),
	}
	service := NewPaymentService(deps)

	orderNumber := service.generateOrderNumber()

	assert.NotEmpty(t, orderNumber)
	assert.Contains(t, orderNumber, "ORD_")
	assert.True(t, len(orderNumber) > 10)
}

func TestPaymentService_createCouponForOrder_Success(t *testing.T) {
	mockPaymentRepo := &MockPaymentRepository{}
	mockCouponRepo := &MockCouponRepository{}
	mockPartnerRepo := &MockPartnerRepository{}

	deps := &PaymentServiceDeps{
		PaymentRepository: mockPaymentRepo,
		CouponRepository:  mockCouponRepo,
		PartnerRepository: mockPartnerRepo,
		Config:            createTestConfig(),
	}

	service := NewPaymentService(deps)

	testOrder := createTestOrder()
	testPartner := createTestPartner()
	testOrder.PartnerID = &testPartner.ID

	mockPartnerRepo.On("GetByID", mock.Anything, testPartner.ID).Return(testPartner, nil)
	mockCouponRepo.On("CodeExists", mock.Anything, mock.AnythingOfType("string")).Return(false, nil)
	mockCouponRepo.On("Create", mock.Anything, mock.MatchedBy(func(c *coupon.Coupon) bool {
		return c.Size == testOrder.Size &&
			c.Style == testOrder.Style &&
			c.IsPurchased == true &&
			c.PurchaseEmail != nil &&
			*c.PurchaseEmail == testOrder.UserEmail
	})).Return(nil)
	mockPaymentRepo.On("UpdateOrderCoupon", mock.Anything, testOrder.OrderNumber, mock.AnythingOfType("uuid.UUID")).Return(nil)

	err := service.createCouponForOrder(context.Background(), testOrder)

	require.NoError(t, err)
	mockPartnerRepo.AssertExpectations(t)
	mockCouponRepo.AssertExpectations(t)
	mockPaymentRepo.AssertExpectations(t)
}

func TestPaymentService_validateWebhookSignature(t *testing.T) {
	config := createTestConfig()
	config.AlphaBankConfig.WebhookSecret = ""

	deps := &PaymentServiceDeps{
		Config: config,
	}
	service := NewPaymentService(deps)

	notification := &PaymentNotificationRequest{
		OrderNumber:     "test_order",
		OrderStatus:     2,
		AlfaBankOrderID: "alfa_123",
		Amount:          10000,
		Currency:        "RUB",
		Checksum:        "invalid_checksum",
	}

	isValid := service.validateWebhookSignature(notification)
	assert.True(t, isValid)
}

func TestPaymentService_getWebhookURL(t *testing.T) {
	testCases := []struct {
		name        string
		setupConfig func() *config.Config
		expectedURL string
	}{
		{
			name: "with webhook URL configured",
			setupConfig: func() *config.Config {
				cfg := createTestConfig()
				cfg.AlphaBankConfig.WebhookURL = "https://custom.webhook.url/payment"
				return cfg
			},
			expectedURL: "https://custom.webhook.url/payment",
		},
		{
			name: "with frontend URL",
			setupConfig: func() *config.Config {
				cfg := createTestConfig()
				cfg.AlphaBankConfig.WebhookURL = ""
				cfg.ServerConfig.FrontendURL = "https://frontend.example.com"
				return cfg
			},
			expectedURL: "https://frontend.example.com/api/payment/notification",
		},
		{
			name: "fallback URL",
			setupConfig: func() *config.Config {
				cfg := createTestConfig()
				cfg.AlphaBankConfig.WebhookURL = ""
				cfg.ServerConfig.FrontendURL = ""
				return cfg
			},
			expectedURL: "https://yourdomain.com/api/payment/notification",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			deps := &PaymentServiceDeps{
				Config: tc.setupConfig(),
			}
			service := NewPaymentService(deps)

			url := service.getWebhookURL()
			assert.Equal(t, tc.expectedURL, url)
		})
	}
}
