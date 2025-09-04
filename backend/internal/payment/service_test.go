package payment

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/skr1ms/mosaic/config"
	"github.com/skr1ms/mosaic/internal/coupon"
	"github.com/skr1ms/mosaic/internal/partner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type AlphaBankConfig = config.AlphaBankConfig
type ServerConfig = config.ServerConfig

// Mock Payment Repository
type MockPaymentRepository struct {
	mock.Mock
}

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

// Mock Coupon Repository
type MockCouponRepository struct {
	mock.Mock
}

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

func (m *MockCouponRepository) FindAvailableCoupon(ctx context.Context, size, style string, partnerID *uuid.UUID) (*coupon.Coupon, error) {
	args := m.Called(ctx, size, style, partnerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*coupon.Coupon), args.Error(1)
}

func (m *MockCouponRepository) MarkAsPurchased(ctx context.Context, id uuid.UUID, purchaseEmail string) error {
	args := m.Called(ctx, id, purchaseEmail)
	return args.Error(0)
}

// Mock Partner Repository
type MockPartnerRepository struct {
	mock.Mock
}

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

func (m *MockPartnerRepository) GetByPartnerCode(ctx context.Context, partnerCode string) (*partner.Partner, error) {
	args := m.Called(ctx, partnerCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*partner.Partner), args.Error(1)
}

// MockEmailService for testing
type MockEmailService struct {
	mock.Mock
}

func (m *MockEmailService) SendCouponPurchaseEmail(to, couponCode, size, style string) error {
	args := m.Called(to, couponCode, size, style)
	return args.Error(0)
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

// Mock Config
type MockConfig struct {
	mock.Mock
}

func (m *MockConfig) GetAlfaBankConfig() AlphaBankConfig {
	args := m.Called()
	return args.Get(0).(AlphaBankConfig)
}

func (m *MockConfig) GetServerConfig() ServerConfig {
	args := m.Called()
	return args.Get(0).(ServerConfig)
}

// Mock Alfa Bank Client
type MockAlfaBankClient struct {
	mock.Mock
}

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

func createTestOrder() *Order {
	return &Order{
		ID:          uuid.New(),
		OrderNumber: "ORD_1234567890_TEST123",
		Size:        "30x40",
		Style:       "grayscale",
		UserEmail:   "test@example.com",
		Amount:      10000,
		Currency:    "RUB",
		Status:      OrderStatusCreated,
		ReturnURL:   "https://example.com/return",
		Description: "Test order",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func createTestPartner() *partner.Partner {
	return &partner.Partner{
		ID:          uuid.New(),
		PartnerCode: "TEST",
		Domain:      "test.example.com",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func createTestCoupon() *coupon.Coupon {
	now := time.Now()
	return &coupon.Coupon{
		ID:            uuid.New(),
		Code:          "TEST-1234-5678",
		Size:          "30x40",
		Style:         "grayscale",
		Status:        "new",
		IsPurchased:   true,
		PurchaseEmail: stringPtr("test@example.com"),
		PurchasedAt:   &now,
		CreatedAt:     now,
	}
}

func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func TestNewPaymentService(t *testing.T) {
	deps := &PaymentServiceDeps{
		PaymentRepository: &MockPaymentRepository{},
		CouponRepository:  &MockCouponRepository{},
		PartnerRepository: &MockPartnerRepository{},
		Config:            &MockConfig{},
		AlfaBankClient:    &MockAlfaBankClient{},
	}

	service := NewPaymentService(deps)

	assert.NotNil(t, service)
	assert.NotNil(t, service.deps)
	assert.Equal(t, deps, service.deps)
}

func TestPaymentService_Structure(t *testing.T) {
	service := &PaymentService{}
	assert.NotNil(t, service)

	service = NewPaymentService(nil)
	assert.NotNil(t, service)
}

func TestPaymentService_ValidateStructure(t *testing.T) {
	tests := []struct {
		name string
		deps *PaymentServiceDeps
	}{
		{
			name: "nil_dependencies",
			deps: nil,
		},
		{
			name: "empty_dependencies",
			deps: &PaymentServiceDeps{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewPaymentService(tt.deps)

			assert.NotNil(t, service)

			if tt.deps != nil {
				assert.Equal(t, tt.deps, service.deps)
			}
		})
	}
}

func TestPaymentService_GenerateOrderNumber(t *testing.T) {
	service := &PaymentService{}

	orderNumber1 := service.GenerateOrderNumber()
	orderNumber2 := service.GenerateOrderNumber()

	assert.NotEmpty(t, orderNumber1)
	assert.NotEmpty(t, orderNumber2)
	assert.NotEqual(t, orderNumber1, orderNumber2)
	assert.Contains(t, orderNumber1, "ORD_")
	assert.Contains(t, orderNumber2, "ORD_")
}

func TestPaymentService_GetAvailableOptions(t *testing.T) {
	service := &PaymentService{}

	options := service.GetAvailableOptions()

	assert.NotNil(t, options)
	assert.Len(t, options.Sizes, 6)
	assert.Len(t, options.Styles, 4)

	expectedSizes := []string{"21x30", "30x40", "40x40", "40x50", "40x60", "50x70"}
	for _, size := range options.Sizes {
		assert.Contains(t, expectedSizes, size.Value)
		assert.Equal(t, FixedPriceRub, size.Price)
	}

	expectedStyles := []string{"grayscale", "skin_tones", "pop_art", "max_colors"}
	for _, style := range options.Styles {
		assert.Contains(t, expectedStyles, style.Value)
	}
}

func TestPaymentService_GetStringValue(t *testing.T) {
	tests := []struct {
		name     string
		ptr      *string
		expected string
	}{
		{
			name:     "nil_pointer",
			ptr:      nil,
			expected: "",
		},
		{
			name:     "valid_string",
			ptr:      stringPtr("test"),
			expected: "test",
		},
		{
			name:     "empty_string",
			ptr:      stringPtr(""),
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getStringValue(tt.ptr)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPaymentService_ValidateWebhookSignature(t *testing.T) {
	tests := []struct {
		name           string
		notification   *PaymentNotificationRequest
		webhookSecret  string
		expectedResult bool
	}{
		{
			name: "valid_signature",
			notification: &PaymentNotificationRequest{
				OrderNumber:     "ORD_123",
				OrderStatus:     intPtr(2),
				AlfaBankOrderID: "ALFA_456",
				Amount:          10000,
				Currency:        "RUB",
				Checksum:        "valid_checksum",
			},
			webhookSecret:  "test_secret",
			expectedResult: false,
		},
		{
			name: "no_webhook_secret",
			notification: &PaymentNotificationRequest{
				OrderNumber:     "ORD_123",
				OrderStatus:     intPtr(2),
				AlfaBankOrderID: "ALFA_456",
				Amount:          10000,
				Currency:        "RUB",
				Checksum:        "any_checksum",
			},
			webhookSecret:  "",
			expectedResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConfig := &MockConfig{}
			mockConfig.On("GetAlfaBankConfig").Return(AlphaBankConfig{
				WebhookSecret: tt.webhookSecret,
			})

			service := &PaymentService{
				deps: &PaymentServiceDeps{
					Config: mockConfig,
				},
			}

			result := service.validateWebhookSignature(tt.notification)
			assert.Equal(t, tt.expectedResult, result)

			mockConfig.AssertExpectations(t)
		})
	}
}

func TestPaymentService_GetWebhookURL(t *testing.T) {
	tests := []struct {
		name           string
		webhookURL     string
		frontendURL    string
		expectedResult string
	}{
		{
			name:           "custom_webhook_url",
			webhookURL:     "https://custom.com/webhook",
			frontendURL:    "https://example.com",
			expectedResult: "https://custom.com/webhook",
		},
		{
			name:           "default_webhook_url",
			webhookURL:     "",
			frontendURL:    "https://example.com",
			expectedResult: "https://example.com/api/payment/notification",
		},
		{
			name:           "fallback_webhook_url",
			webhookURL:     "",
			frontendURL:    "",
			expectedResult: "https://yourdomain.com/api/payment/notification",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConfig := &MockConfig{}
			mockConfig.On("GetAlfaBankConfig").Return(AlphaBankConfig{
				WebhookURL: tt.webhookURL,
			})

			if tt.webhookURL == "" {
				mockConfig.On("GetServerConfig").Return(config.ServerConfig{
					FrontendURL: tt.frontendURL,
				})
			}

			service := &PaymentService{
				deps: &PaymentServiceDeps{
					Config: mockConfig,
				},
			}

			result := service.getWebhookURL()
			assert.Equal(t, tt.expectedResult, result)

			mockConfig.AssertExpectations(t)
		})
	}
}

func TestPaymentService_GetOrderStatus(t *testing.T) {
	tests := []struct {
		name            string
		orderNumber     string
		mockSetup       func(*MockPaymentRepository, *MockCouponRepository, *MockAlfaBankClient)
		expectedError   bool
		expectedSuccess bool
	}{
		{
			name:        "order_not_found",
			orderNumber: "ORD_NOT_FOUND",
			mockSetup: func(repo *MockPaymentRepository, couponRepo *MockCouponRepository, alfaClient *MockAlfaBankClient) {
				repo.On("GetOrderByNumber", mock.Anything, "ORD_NOT_FOUND").Return(nil, errors.New("not found"))
			},
			expectedError:   false,
			expectedSuccess: false,
		},
		{
			name:        "order_found_with_coupon",
			orderNumber: "ORD_FOUND",
			mockSetup: func(repo *MockPaymentRepository, couponRepo *MockCouponRepository, alfaClient *MockAlfaBankClient) {
				order := createTestOrder()
				order.Status = OrderStatusPaid
				couponID := uuid.New()
				order.CouponID = &couponID
				repo.On("GetOrderByNumber", mock.Anything, "ORD_FOUND").Return(order, nil)

				coupon := createTestCoupon()
				coupon.ID = couponID
				couponRepo.On("GetByID", mock.Anything, couponID).Return(coupon, nil)
			},
			expectedError:   false,
			expectedSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockPaymentRepository)
			mockCouponRepo := new(MockCouponRepository)
			mockAlfaClient := new(MockAlfaBankClient)

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo, mockCouponRepo, mockAlfaClient)
			}

			service := &PaymentService{
				deps: &PaymentServiceDeps{
					PaymentRepository: mockRepo,
					CouponRepository:  mockCouponRepo,
					AlfaBankClient:    mockAlfaClient,
				},
			}

			result, err := service.GetOrderStatus(context.Background(), tt.orderNumber)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedSuccess, result.Success)
			}

			mockRepo.AssertExpectations(t)
			mockCouponRepo.AssertExpectations(t)
			mockAlfaClient.AssertExpectations(t)
		})
	}
}

func TestPaymentService_ProcessWebhookNotification(t *testing.T) {
	tests := []struct {
		name          string
		notification  *PaymentNotificationRequest
		mockSetup     func(*MockPaymentRepository, *MockCouponRepository, *MockPartnerRepository, *MockConfig)
		expectedError bool
	}{
		{
			name: "order_not_found",
			notification: &PaymentNotificationRequest{
				OrderNumber: "ORD_NOT_FOUND",
				OrderStatus: intPtr(2),
			},
			mockSetup: func(repo *MockPaymentRepository, couponRepo *MockCouponRepository, partnerRepo *MockPartnerRepository, config *MockConfig) {
				repo.On("GetOrderByNumber", mock.Anything, "ORD_NOT_FOUND").Return(nil, errors.New("not found"))
				config.On("GetAlfaBankConfig").Return(AlphaBankConfig{WebhookSecret: ""})
			},
			expectedError: true,
		},
		{
			name: "order_already_processed",
			notification: &PaymentNotificationRequest{
				OrderNumber:     "ORD_PROCESSED",
				OrderStatus:     intPtr(2),
				AlfaBankOrderID: "ALFA123",
			},
			mockSetup: func(repo *MockPaymentRepository, couponRepo *MockCouponRepository, partnerRepo *MockPartnerRepository, config *MockConfig) {
				order := createTestOrder()
				order.Status = OrderStatusPaid
				repo.On("GetOrderByNumber", mock.Anything, "ORD_PROCESSED").Return(order, nil)
				repo.On("UpdateOrderStatus", mock.Anything, "ORD_PROCESSED", OrderStatusPaid, mock.AnythingOfType("*string")).Return(nil)
				couponRepo.On("CodeExists", mock.Anything, mock.AnythingOfType("string")).Return(false, nil)
				couponRepo.On("Create", mock.Anything, mock.AnythingOfType("*coupon.Coupon")).Return(nil)
				defaultPartner := &partner.Partner{
					ID:          uuid.New(),
					PartnerCode: "0000",
					Domain:      "default.com",
					BrandName:   "Default Brand",
				}
				partnerRepo.On("GetByPartnerCode", mock.Anything, "0000").Return(defaultPartner, nil)
				repo.On("UpdateOrderCoupon", mock.Anything, order.OrderNumber, mock.AnythingOfType("uuid.UUID")).Return(nil)
				config.On("GetAlfaBankConfig").Return(AlphaBankConfig{WebhookSecret: ""})
			},
			expectedError: false,
		},
		{
			name: "payment_successful",
			notification: &PaymentNotificationRequest{
				OrderNumber: "ORD_SUCCESS",
				OrderStatus: intPtr(2),
			},
			mockSetup: func(repo *MockPaymentRepository, couponRepo *MockCouponRepository, partnerRepo *MockPartnerRepository, config *MockConfig) {
				order := createTestOrder()
				order.Status = OrderStatusPending
				repo.On("GetOrderByNumber", mock.Anything, "ORD_SUCCESS").Return(order, nil)
				repo.On("UpdateOrderStatus", mock.Anything, "ORD_SUCCESS", OrderStatusPaid, mock.Anything).Return(nil)
				couponRepo.On("CodeExists", mock.Anything, mock.AnythingOfType("string")).Return(false, nil)
				couponRepo.On("Create", mock.Anything, mock.AnythingOfType("*coupon.Coupon")).Return(nil)
				defaultPartner := &partner.Partner{
					ID:          uuid.New(),
					PartnerCode: "0000",
					Domain:      "default.com",
					BrandName:   "Default Brand",
				}
				partnerRepo.On("GetByPartnerCode", mock.Anything, "0000").Return(defaultPartner, nil)
				repo.On("UpdateOrderCoupon", mock.Anything, order.OrderNumber, mock.AnythingOfType("uuid.UUID")).Return(nil)
				config.On("GetAlfaBankConfig").Return(AlphaBankConfig{WebhookSecret: ""})
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockPaymentRepository)
			mockCouponRepo := new(MockCouponRepository)
			mockPartnerRepo := new(MockPartnerRepository)
			mockConfig := new(MockConfig)

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo, mockCouponRepo, mockPartnerRepo, mockConfig)
			}

			service := &PaymentService{
				deps: &PaymentServiceDeps{
					PaymentRepository: mockRepo,
					CouponRepository:  mockCouponRepo,
					PartnerRepository: mockPartnerRepo,
					Config:            mockConfig,
				},
			}

			err := service.ProcessWebhookNotification(context.Background(), tt.notification)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
			mockCouponRepo.AssertExpectations(t)
			mockPartnerRepo.AssertExpectations(t)
			mockConfig.AssertExpectations(t)
		})
	}
}

func TestPaymentService_ProcessPaymentReturn(t *testing.T) {
	tests := []struct {
		name          string
		orderNumber   string
		mockSetup     func(*MockPaymentRepository, *MockAlfaBankClient, *MockCouponRepository, *MockPartnerRepository, *MockConfig)
		expectedError bool
	}{
		{
			name:        "order_not_found",
			orderNumber: "ORD_NOT_FOUND",
			mockSetup: func(repo *MockPaymentRepository, alfaClient *MockAlfaBankClient, couponRepo *MockCouponRepository, partnerRepo *MockPartnerRepository, config *MockConfig) {
				repo.On("GetOrderByNumber", mock.Anything, "ORD_NOT_FOUND").Return(nil, errors.New("not found"))
			},
			expectedError: true,
		},
		{
			name:        "order_already_processed",
			orderNumber: "ORD_PROCESSED",
			mockSetup: func(repo *MockPaymentRepository, alfaClient *MockAlfaBankClient, couponRepo *MockCouponRepository, partnerRepo *MockPartnerRepository, config *MockConfig) {
				order := createTestOrder()
				order.Status = OrderStatusPaid
				repo.On("GetOrderByNumber", mock.Anything, "ORD_PROCESSED").Return(order, nil)
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockPaymentRepository)
			mockAlfaClient := new(MockAlfaBankClient)
			mockCouponRepo := new(MockCouponRepository)
			mockPartnerRepo := new(MockPartnerRepository)
			mockConfig := new(MockConfig)

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo, mockAlfaClient, mockCouponRepo, mockPartnerRepo, mockConfig)
			}

			service := &PaymentService{
				deps: &PaymentServiceDeps{
					PaymentRepository: mockRepo,
					AlfaBankClient:    mockAlfaClient,
					CouponRepository:  mockCouponRepo,
					PartnerRepository: mockPartnerRepo,
					Config:            mockConfig,
				},
			}

			err := service.ProcessPaymentReturn(context.Background(), tt.orderNumber)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
			mockAlfaClient.AssertExpectations(t)
			mockCouponRepo.AssertExpectations(t)
			mockPartnerRepo.AssertExpectations(t)
			mockConfig.AssertExpectations(t)
		})
	}
}

func TestPaymentService_CreateCouponForOrder(t *testing.T) {
	tests := []struct {
		name          string
		order         *Order
		mockSetup     func(*MockCouponRepository, *MockPartnerRepository, *MockPaymentRepository)
		expectedError bool
	}{
		{
			name: "successful_coupon_creation",
			order: func() *Order {
				order := createTestOrder()
				partnerID := uuid.New()
				order.PartnerID = &partnerID
				return order
			}(),
			mockSetup: func(couponRepo *MockCouponRepository, partnerRepo *MockPartnerRepository, paymentRepo *MockPaymentRepository) {
				couponRepo.On("CodeExists", mock.Anything, mock.AnythingOfType("string")).Return(false, nil)
				couponRepo.On("Create", mock.Anything, mock.AnythingOfType("*coupon.Coupon")).Return(nil)
				paymentRepo.On("UpdateOrderCoupon", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("uuid.UUID")).Return(nil)
			},
			expectedError: false,
		},
		{
			name: "coupon_creation_failed",
			order: func() *Order {
				order := createTestOrder()
				partnerID := uuid.New()
				order.PartnerID = &partnerID
				return order
			}(),
			mockSetup: func(couponRepo *MockCouponRepository, partnerRepo *MockPartnerRepository, paymentRepo *MockPaymentRepository) {
				couponRepo.On("CodeExists", mock.Anything, mock.AnythingOfType("string")).Return(false, nil)
				couponRepo.On("Create", mock.Anything, mock.AnythingOfType("*coupon.Coupon")).Return(errors.New("create failed"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCouponRepo := new(MockCouponRepository)
			mockPartnerRepo := new(MockPartnerRepository)
			mockPaymentRepo := new(MockPaymentRepository)

			if tt.mockSetup != nil {
				tt.mockSetup(mockCouponRepo, mockPartnerRepo, mockPaymentRepo)
			}

			service := &PaymentService{
				deps: &PaymentServiceDeps{
					CouponRepository:  mockCouponRepo,
					PartnerRepository: mockPartnerRepo,
					PaymentRepository: mockPaymentRepo,
				},
			}

			err := service.createCouponForOrder(context.Background(), tt.order)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockCouponRepo.AssertExpectations(t)
			mockPartnerRepo.AssertExpectations(t)
			mockPaymentRepo.AssertExpectations(t)
		})
	}
}

func TestPaymentService_Constants(t *testing.T) {
	assert.Equal(t, "created", OrderStatusCreated)
	assert.Equal(t, "pending", OrderStatusPending)
	assert.Equal(t, "paid", OrderStatusPaid)
	assert.Equal(t, "failed", OrderStatusFailed)
	assert.Equal(t, "cancelled", OrderStatusCancelled)

	assert.Equal(t, 100.0, FixedPriceRub)

	assert.Equal(t, "21x30", Size21x30)
	assert.Equal(t, "30x40", Size30x40)
	assert.Equal(t, "40x40", Size40x40)
	assert.Equal(t, "40x50", Size40x50)
	assert.Equal(t, "40x60", Size40x60)
	assert.Equal(t, "50x70", Size50x70)

	assert.Equal(t, "grayscale", StyleGrayscale)
	assert.Equal(t, "skin_tones", StyleSkinTone)
	assert.Equal(t, "pop_art", StylePopArt)
	assert.Equal(t, "max_colors", StyleMaxColors)
}

func TestPaymentService_TestAlfaBankIntegration(t *testing.T) {
	tests := []struct {
		name            string
		request         *AlfaBankRegisterRequest
		mockSetup       func(*MockConfig, *MockAlfaBankClient)
		expectedError   bool
		expectedSuccess bool
	}{
		{
			name: "missing_configuration",
			request: &AlfaBankRegisterRequest{
				OrderNumber: "TEST_ORDER",
				Amount:      10000,
				Currency:    "810",
			},
			mockSetup: func(config *MockConfig, alfaClient *MockAlfaBankClient) {
				config.On("GetAlfaBankConfig").Return(AlphaBankConfig{
					Username: "",
					Password: "",
					Url:      "",
				})
			},
			expectedError:   false,
			expectedSuccess: false,
		},
		{
			name: "api_connection_failed",
			request: &AlfaBankRegisterRequest{
				OrderNumber: "TEST_ORDER",
				Amount:      10000,
				Currency:    "810",
			},
			mockSetup: func(config *MockConfig, alfaClient *MockAlfaBankClient) {
				config.On("GetAlfaBankConfig").Return(AlphaBankConfig{
					Username: "test_user",
					Password: "test_pass",
					Url:      "https://test.alfabank.ru",
				})
				alfaClient.On("RegisterOrder", mock.Anything, mock.Anything).Return(nil, errors.New("connection failed"))
			},
			expectedError:   false,
			expectedSuccess: false,
		},
		{
			name: "api_returned_error",
			request: &AlfaBankRegisterRequest{
				OrderNumber: "TEST_ORDER",
				Amount:      10000,
				Currency:    "810",
			},
			mockSetup: func(config *MockConfig, alfaClient *MockAlfaBankClient) {
				config.On("GetAlfaBankConfig").Return(AlphaBankConfig{
					Username: "test_user",
					Password: "test_pass",
					Url:      "https://test.alfabank.ru",
				})
				alfaClient.On("RegisterOrder", mock.Anything, mock.Anything).Return(&AlfaBankRegisterResponse{
					ErrorCode:    "ERROR_001",
					ErrorMessage: "Invalid parameters",
				}, nil)
			},
			expectedError:   false,
			expectedSuccess: false,
		},
		{
			name: "successful_integration",
			request: &AlfaBankRegisterRequest{
				OrderNumber: "TEST_ORDER",
				Amount:      10000,
				Currency:    "810",
			},
			mockSetup: func(config *MockConfig, alfaClient *MockAlfaBankClient) {
				config.On("GetAlfaBankConfig").Return(AlphaBankConfig{
					Username: "r-photo_doyoupaint-api",
					Password: "r-photo_doyoupaint*?1",
					Url:      "https://pay.alfabank.ru",
				})
				alfaClient.On("RegisterOrder", mock.Anything, mock.Anything).Return(&AlfaBankRegisterResponse{
					OrderId:   "ALFA_ORDER_123",
					FormUrl:   "https://pay.alfabank.ru/payment/form",
					ErrorCode: "",
				}, nil)
			},
			expectedError:   false,
			expectedSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConfig := new(MockConfig)
			mockAlfaClient := new(MockAlfaBankClient)

			if tt.mockSetup != nil {
				tt.mockSetup(mockConfig, mockAlfaClient)
			}

			service := &PaymentService{
				deps: &PaymentServiceDeps{
					Config:         mockConfig,
					AlfaBankClient: mockAlfaClient,
				},
			}

			result, err := service.TestAlfaBankIntegration(context.Background(), tt.request)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedSuccess, result.Success)
				assert.Equal(t, tt.request.OrderNumber, result.OrderNumber)

				if tt.expectedSuccess {
					assert.NotEmpty(t, result.AlfaBankOrderID)
					assert.NotEmpty(t, result.PaymentURL)
					assert.Contains(t, result.TestStatus, "successful")
				} else {
					assert.NotEmpty(t, result.Message)
					if result.ConfigStatus != "" {
						assert.Contains(t, result.ConfigStatus, "Configuration")
					}
				}
			}

			mockConfig.AssertExpectations(t)
			mockAlfaClient.AssertExpectations(t)
		})
	}
}

func TestPaymentService_PurchaseCoupon_Comprehensive(t *testing.T) {
	tests := []struct {
		name            string
		request         *PurchaseCouponRequest
		mockSetup       func(*MockPaymentRepository, *MockPartnerRepository, *MockAlfaBankClient, *MockConfig)
		expectedError   bool
		expectedSuccess bool
	}{
		{
			name: "unsupported_size",
			request: &PurchaseCouponRequest{
				Size:      "INVALID_SIZE",
				Style:     "max_colors",
				Email:     "test@example.com",
				ReturnURL: "https://example.com/return",
			},
			mockSetup: func(paymentRepo *MockPaymentRepository, partnerRepo *MockPartnerRepository, alfaClient *MockAlfaBankClient, config *MockConfig) {
				// No mocks needed for unsupported size - it returns early
			},
			expectedError:   false,
			expectedSuccess: false,
		},
		{
			name: "successful_purchase_without_partner",
			request: &PurchaseCouponRequest{
				Size:      "40x50",
				Style:     "max_colors",
				Email:     "test@example.com",
				ReturnURL: "https://example.com/return",
				FailURL:   stringPtr("https://example.com/fail"),
				Language:  "ru",
			},
			mockSetup: func(paymentRepo *MockPaymentRepository, partnerRepo *MockPartnerRepository, alfaClient *MockAlfaBankClient, config *MockConfig) {
				// Mock order creation
				paymentRepo.On("GetOrderByNumber", mock.Anything, mock.Anything).Return(nil, errors.New("not found"))
				paymentRepo.On("CreateOrder", mock.Anything, mock.Anything).Return(nil)
				paymentRepo.On("UpdateOrderStatus", mock.Anything, mock.Anything, OrderStatusPending, mock.Anything).Return(nil)
				paymentRepo.On("UpdateOrderPaymentURL", mock.Anything, mock.Anything, mock.Anything).Return(nil)

				// Mock Alfa Bank response
				alfaClient.On("RegisterOrder", mock.Anything, mock.Anything).Return(&AlfaBankRegisterResponse{
					OrderId:   "ALFA_ORDER_123",
					FormUrl:   "https://pay.alfabank.ru/payment/form",
					ErrorCode: "",
				}, nil)

				// Mock config
				config.On("GetAlfaBankConfig").Return(AlphaBankConfig{
					WebhookURL: "https://test.com/webhook",
				})
				config.On("GetServerConfig").Return(ServerConfig{
					FrontendURL: "https://test.com",
				}).Maybe()
			},
			expectedError:   false,
			expectedSuccess: true,
		},
		{
			name: "successful_purchase_with_partner",
			request: &PurchaseCouponRequest{
				Size:      "30x40",
				Style:     "grayscale",
				Email:     "partner@example.com",
				ReturnURL: "https://partner.com/return",
				Domain:    stringPtr("partner.example.com"),
			},
			mockSetup: func(paymentRepo *MockPaymentRepository, partnerRepo *MockPartnerRepository, alfaClient *MockAlfaBankClient, config *MockConfig) {
				// Mock partner lookup
				partner := createTestPartner()
				partner.Domain = "partner.example.com"
				partnerRepo.On("GetByDomain", mock.Anything, "partner.example.com").Return(partner, nil)

				// Mock order creation
				paymentRepo.On("GetOrderByNumber", mock.Anything, mock.Anything).Return(nil, errors.New("not found"))
				paymentRepo.On("CreateOrder", mock.Anything, mock.Anything).Return(nil)
				paymentRepo.On("UpdateOrderStatus", mock.Anything, mock.Anything, OrderStatusPending, mock.Anything).Return(nil)
				paymentRepo.On("UpdateOrderPaymentURL", mock.Anything, mock.Anything, mock.Anything).Return(nil)

				// Mock Alfa Bank response
				alfaClient.On("RegisterOrder", mock.Anything, mock.Anything).Return(&AlfaBankRegisterResponse{
					OrderId:   "ALFA_ORDER_456",
					FormUrl:   "https://pay.alfabank.ru/payment/form2",
					ErrorCode: "",
				}, nil)

				// Mock config
				config.On("GetAlfaBankConfig").Return(AlphaBankConfig{
					WebhookURL: "https://partner.com/webhook",
				})
				config.On("GetServerConfig").Return(ServerConfig{
					FrontendURL: "https://partner.com",
				}).Maybe()
			},
			expectedError:   false,
			expectedSuccess: true,
		},
		{
			name: "alfa_bank_api_error",
			request: &PurchaseCouponRequest{
				Size:      "40x50",
				Style:     "pop_art",
				Email:     "test@example.com",
				ReturnURL: "https://example.com/return",
			},
			mockSetup: func(paymentRepo *MockPaymentRepository, partnerRepo *MockPartnerRepository, alfaClient *MockAlfaBankClient, config *MockConfig) {
				// Mock order creation
				paymentRepo.On("GetOrderByNumber", mock.Anything, mock.Anything).Return(nil, errors.New("not found"))
				paymentRepo.On("CreateOrder", mock.Anything, mock.Anything).Return(nil)

				// Mock Alfa Bank error
				alfaClient.On("RegisterOrder", mock.Anything, mock.Anything).Return(nil, errors.New("API connection failed"))

				// Mock config
				config.On("GetAlfaBankConfig").Return(AlphaBankConfig{
					WebhookURL: "https://test.com/webhook",
				})
				config.On("GetServerConfig").Return(ServerConfig{
					FrontendURL: "https://test.com",
				}).Maybe()
			},
			expectedError:   false,
			expectedSuccess: false,
		},
		{
			name: "alfa_bank_returns_error_code",
			request: &PurchaseCouponRequest{
				Size:      "50x70",
				Style:     "skin_tones",
				Email:     "test@example.com",
				ReturnURL: "https://example.com/return",
			},
			mockSetup: func(paymentRepo *MockPaymentRepository, partnerRepo *MockPartnerRepository, alfaClient *MockAlfaBankClient, config *MockConfig) {
				// Mock order creation
				paymentRepo.On("GetOrderByNumber", mock.Anything, mock.Anything).Return(nil, errors.New("not found"))
				paymentRepo.On("CreateOrder", mock.Anything, mock.Anything).Return(nil)

				// Mock Alfa Bank error response
				alfaClient.On("RegisterOrder", mock.Anything, mock.Anything).Return(&AlfaBankRegisterResponse{
					ErrorCode:    "INVALID_MERCHANT",
					ErrorMessage: "Merchant not found",
				}, nil)

				// Mock config
				config.On("GetAlfaBankConfig").Return(AlphaBankConfig{
					WebhookURL: "https://test.com/webhook",
				})
				config.On("GetServerConfig").Return(ServerConfig{
					FrontendURL: "https://test.com",
				}).Maybe()
			},
			expectedError:   false,
			expectedSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPaymentRepo := new(MockPaymentRepository)
			mockPartnerRepo := new(MockPartnerRepository)
			mockAlfaClient := new(MockAlfaBankClient)
			mockConfig := new(MockConfig)
			mockCouponRepo := new(MockCouponRepository)

			if tt.mockSetup != nil {
				tt.mockSetup(mockPaymentRepo, mockPartnerRepo, mockAlfaClient, mockConfig)
			}

			service := &PaymentService{
				deps: &PaymentServiceDeps{
					PaymentRepository: mockPaymentRepo,
					PartnerRepository: mockPartnerRepo,
					AlfaBankClient:    mockAlfaClient,
					Config:            mockConfig,
					CouponRepository:  mockCouponRepo,
				},
			}

			result, err := service.PurchaseCoupon(context.Background(), tt.request)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedSuccess, result.Success)

				if tt.expectedSuccess {
					assert.NotEmpty(t, result.OrderID)
					assert.NotEmpty(t, result.PaymentURL)
					assert.Empty(t, result.Message)
				} else {
					assert.NotEmpty(t, result.Message)
				}
			}

			mockPaymentRepo.AssertExpectations(t)
			mockPartnerRepo.AssertExpectations(t)
			mockAlfaClient.AssertExpectations(t)
			mockConfig.AssertExpectations(t)
			mockCouponRepo.AssertExpectations(t)
		})
	}
}

func TestPaymentService_OrderNumberGeneration(t *testing.T) {
	service := &PaymentService{}

	// Test uniqueness of generated order numbers
	generated := make(map[string]bool)
	for i := 0; i < 100; i++ {
		orderNumber := service.GenerateOrderNumber()
		assert.NotEmpty(t, orderNumber)
		assert.Contains(t, orderNumber, "ORD_")
		assert.False(t, generated[orderNumber], "Order number %s was generated twice", orderNumber)
		generated[orderNumber] = true
	}
}

func TestPaymentService_WebhookValidation(t *testing.T) {
	tests := []struct {
		name           string
		notification   *PaymentNotificationRequest
		webhookSecret  string
		expectedResult bool
		description    string
	}{
		{
			name: "empty_webhook_secret_should_pass",
			notification: &PaymentNotificationRequest{
				OrderNumber:     "ORD_123",
				OrderStatus:     intPtr(2),
				AlfaBankOrderID: "ALFA_456",
				Amount:          10000,
				Currency:        "RUB",
				Checksum:        "any_checksum",
			},
			webhookSecret:  "",
			expectedResult: true,
			description:    "When webhook secret is empty, validation should pass",
		},
		{
			name: "nil_order_status_handling",
			notification: &PaymentNotificationRequest{
				OrderNumber:     "ORD_123",
				OrderStatus:     nil,
				AlfaBankOrderID: "ALFA_456",
				Amount:          10000,
				Currency:        "RUB",
				Checksum:        "test_checksum",
			},
			webhookSecret:  "test_secret",
			expectedResult: false,
			description:    "Should handle nil order status gracefully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConfig := &MockConfig{}
			mockConfig.On("GetAlfaBankConfig").Return(AlphaBankConfig{
				WebhookSecret: tt.webhookSecret,
			})

			service := &PaymentService{
				deps: &PaymentServiceDeps{
					Config: mockConfig,
				},
			}

			result := service.validateWebhookSignature(tt.notification)
			assert.Equal(t, tt.expectedResult, result, tt.description)

			mockConfig.AssertExpectations(t)
		})
	}
}

func TestPaymentService_CompletePaymentFlow(t *testing.T) {
	t.Run("end_to_end_successful_payment", func(t *testing.T) {
		// Setup mocks
		mockPaymentRepo := new(MockPaymentRepository)
		mockCouponRepo := new(MockCouponRepository)
		mockPartnerRepo := new(MockPartnerRepository)
		mockAlfaClient := new(MockAlfaBankClient)
		mockConfig := new(MockConfig)

		// Purchase request
		request := &PurchaseCouponRequest{
			Size:      "40x50",
			Style:     "max_colors",
			Email:     "customer@example.com",
			ReturnURL: "https://shop.com/success",
			FailURL:   stringPtr("https://shop.com/fail"),
		}

		// Mock successful purchase - order doesn't exist initially
		mockPaymentRepo.On("GetOrderByNumber", mock.Anything, mock.Anything).Return(nil, errors.New("not found")).Once()

		// Mock CreateOrder to return a specific order
		testOrder := &Order{
			ID:          uuid.New(),
			OrderNumber: "ORD_1234567890_TEST123", // Fixed order number for testing
			Size:        "40x50",
			Style:       "max_colors",
			UserEmail:   "customer@example.com",
			Amount:      10000,
			Currency:    "RUB",
			Status:      OrderStatusCreated,
			ReturnURL:   "https://shop.com/success",
			FailURL:     stringPtr("https://shop.com/fail"),
		}
		mockPaymentRepo.On("CreateOrder", mock.Anything, mock.MatchedBy(func(order *Order) bool {
			// Copy the generated order number to our test order
			testOrder.OrderNumber = order.OrderNumber
			return true
		})).Return(nil).Run(func(args mock.Arguments) {
			// Store the created order for later use
			createdOrder := args.Get(1).(*Order)
			testOrder.OrderNumber = createdOrder.OrderNumber
		})

		mockPaymentRepo.On("UpdateOrderStatus", mock.Anything, mock.Anything, OrderStatusPending, mock.Anything).Return(nil)
		mockPaymentRepo.On("UpdateOrderPaymentURL", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		mockAlfaClient.On("RegisterOrder", mock.Anything, mock.Anything).Return(&AlfaBankRegisterResponse{
			OrderId: "ALFA_SUCCESS_123",
			FormUrl: "https://pay.alfabank.ru/payment/form",
		}, nil)

		mockConfig.On("GetAlfaBankConfig").Return(AlphaBankConfig{
			WebhookURL: "https://shop.com/webhook",
		})
		mockConfig.On("GetServerConfig").Return(ServerConfig{
			FrontendURL: "https://shop.com",
		}).Maybe()

		service := &PaymentService{
			deps: &PaymentServiceDeps{
				PaymentRepository: mockPaymentRepo,
				CouponRepository:  mockCouponRepo,
				PartnerRepository: mockPartnerRepo,
				AlfaBankClient:    mockAlfaClient,
				Config:            mockConfig,
			},
		}

		// 1. Purchase coupon
		purchaseResult, err := service.PurchaseCoupon(context.Background(), request)
		assert.NoError(t, err)
		assert.True(t, purchaseResult.Success)
		assert.NotEmpty(t, purchaseResult.PaymentURL)

		// Now mock webhook processing after payment
		// Use the same order that was created in PurchaseCoupon
		testOrder.Status = OrderStatusPending
		testOrder.AlfaBankOrderID = stringPtr("ALFA_SUCCESS_123")

		mockPaymentRepo.On("GetOrderByNumber", mock.Anything, testOrder.OrderNumber).Return(testOrder, nil)
		mockPaymentRepo.On("UpdateOrderStatus", mock.Anything, mock.Anything, OrderStatusPaid, mock.Anything).Return(nil)

		// Mock coupon generation and creation
		mockCouponRepo.On("CodeExists", mock.Anything, mock.AnythingOfType("string")).Return(false, nil)
		mockCouponRepo.On("Create", mock.Anything, mock.AnythingOfType("*coupon.Coupon")).Return(nil)
		defaultPartner := &partner.Partner{
			ID:          uuid.New(),
			PartnerCode: "0000",
			Domain:      "default.com",
			BrandName:   "Default Brand",
		}
		mockPartnerRepo.On("GetByPartnerCode", mock.Anything, "0000").Return(defaultPartner, nil)
		mockPaymentRepo.On("UpdateOrderCoupon", mock.Anything, mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil)

		// 2. Process webhook notification (payment successful)
		notification := &PaymentNotificationRequest{
			OrderNumber:     testOrder.OrderNumber,
			OrderStatus:     intPtr(2), // Paid
			AlfaBankOrderID: "ALFA_SUCCESS_123",
			Amount:          testOrder.Amount,
			Currency:        testOrder.Currency,
		}

		err = service.ProcessWebhookNotification(context.Background(), notification)
		assert.NoError(t, err)

		// Verify all mocks were called as expected
		mockPaymentRepo.AssertExpectations(t)
		mockCouponRepo.AssertExpectations(t)
		mockPartnerRepo.AssertExpectations(t)
		mockAlfaClient.AssertExpectations(t)
		mockConfig.AssertExpectations(t)
	})
}
