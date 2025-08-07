package payment

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/skr1ms/mosaic/config"
	"github.com/skr1ms/mosaic/internal/coupon"
	"github.com/skr1ms/mosaic/internal/partner"
)

type IntegrationTestSuite struct {
	service  *PaymentService
	mocks    *IntegrationMocks
	testData *IntegrationTestData
}

type IntegrationMocks struct {
	PaymentRepo    *MockPaymentRepository
	CouponRepo     *MockCouponRepository
	PartnerRepo    *MockPartnerRepository
	AlfaBankClient *MockAlfaBankClient
}

type IntegrationTestData struct {
	Order   *Order
	Partner *partner.Partner
	Coupon  *coupon.Coupon
	Config  *config.Config
}

func SetupIntegrationTest(t *testing.T) *IntegrationTestSuite {
	suite := &IntegrationTestSuite{
		mocks: &IntegrationMocks{
			PaymentRepo:    &MockPaymentRepository{},
			CouponRepo:     &MockCouponRepository{},
			PartnerRepo:    &MockPartnerRepository{},
			AlfaBankClient: &MockAlfaBankClient{},
		},
		testData: &IntegrationTestData{
			Order:   createTestOrder(),
			Partner: createTestPartner(),
			Coupon:  createTestCoupon(),
			Config:  createTestConfig(),
		},
	}

	deps := &PaymentServiceDeps{
		PaymentRepository: suite.mocks.PaymentRepo,
		CouponRepository:  suite.mocks.CouponRepo,
		PartnerRepository: suite.mocks.PartnerRepo,
		Config:            suite.testData.Config,
	}

	suite.service = NewPaymentService(deps)
	suite.service.alfaClient = suite.mocks.AlfaBankClient
	return suite
}

func TestIntegration_FullPurchaseWorkflow(t *testing.T) {
	suite := SetupIntegrationTest(t)

	testPartner := suite.testData.Partner
	domain := testPartner.Domain

	// Шаг 1: Покупка купона
	purchaseReq := &PurchaseCouponRequest{
		Size:      Size40x50,
		Style:     StyleGrayscale,
		Email:     "customer@example.com",
		ReturnURL: "https://test.example.com/return",
		Domain:    &domain,
	}

	alfaResponse := &AlfaBankRegisterResponse{
		OrderId: "alfa_order_123",
		FormUrl: "https://payment.alfabank.ru/payment/order123",
	}

	var actualOrderNumber string

	suite.mocks.PartnerRepo.On("GetByDomain", mock.Anything, domain).Return(testPartner, nil)
	suite.mocks.PaymentRepo.On("GetOrderByNumber", mock.Anything, mock.AnythingOfType("string")).Return(nil, errors.New("not found"))
	suite.mocks.PaymentRepo.On("CreateOrder", mock.Anything, mock.AnythingOfType("*payment.Order")).Run(func(args mock.Arguments) {
		// Захватываем номер заказа из вызова
		if order, ok := args.Get(1).(*Order); ok {
			actualOrderNumber = order.OrderNumber
			t.Logf("Captured order number: %s", actualOrderNumber)
		}
	}).Return(nil)
	suite.mocks.AlfaBankClient.On("RegisterOrder", mock.Anything, mock.AnythingOfType("*payment.AlfaBankRegisterRequest")).Return(alfaResponse, nil)
	suite.mocks.PaymentRepo.On("UpdateOrderStatus", mock.Anything, mock.AnythingOfType("string"), OrderStatusPending, &alfaResponse.OrderId).Return(nil)
	suite.mocks.PaymentRepo.On("UpdateOrderPaymentURL", mock.Anything, mock.AnythingOfType("string"), alfaResponse.FormUrl).Return(nil)

	// Выполняем покупку
	purchaseResult, err := suite.service.PurchaseCoupon(context.Background(), purchaseReq)
	require.NoError(t, err)
	assert.True(t, purchaseResult.Success)
	assert.Equal(t, alfaResponse.FormUrl, purchaseResult.PaymentURL)

	require.NotEmpty(t, actualOrderNumber, "Order number should be captured from CreateOrder call")

	// Шаг 2: Проверка статуса заказа перед оплатой
	pendingOrder := &Order{
		ID:              uuid.New(),
		OrderNumber:     actualOrderNumber,
		Status:          OrderStatusPending,
		AlfaBankOrderID: &alfaResponse.OrderId,
		Amount:          100000,
		Currency:        "RUB",
		Size:            Size40x50,
		Style:           StyleGrayscale,
		UserEmail:       "customer@example.com",
		PartnerID:       &testPartner.ID,
	}

	suite.mocks.PaymentRepo.ExpectedCalls = nil
	suite.mocks.PaymentRepo.On("GetOrderByNumber", mock.Anything, actualOrderNumber).Return(pendingOrder, nil)
	suite.mocks.AlfaBankClient.On("GetOrderStatus", mock.Anything, alfaResponse.OrderId).Return(&AlfaBankStatusResponse{
		OrderStatus: 1,
		ErrorCode:   "",
	}, nil)

	statusResult, err := suite.service.GetOrderStatus(context.Background(), actualOrderNumber)
	require.NoError(t, err)
	assert.True(t, statusResult.Success)
	assert.Equal(t, OrderStatusPending, statusResult.Status) // Шаг 3: Webhook уведомление об успешной оплате
	notification := &PaymentNotificationRequest{
		OrderNumber:     actualOrderNumber,
		OrderStatus:     2,
		AlfaBankOrderID: alfaResponse.OrderId,
		Amount:          pendingOrder.Amount,
		Currency:        pendingOrder.Currency,
	}

	suite.mocks.PaymentRepo.ExpectedCalls = nil
	suite.mocks.PartnerRepo.ExpectedCalls = nil
	suite.mocks.CouponRepo.ExpectedCalls = nil

	suite.mocks.PaymentRepo.On("GetOrderByNumber", mock.Anything, actualOrderNumber).Return(pendingOrder, nil)
	suite.mocks.PaymentRepo.On("UpdateOrderStatus", mock.Anything, actualOrderNumber, OrderStatusPaid, &alfaResponse.OrderId).Return(nil)
	suite.mocks.PartnerRepo.On("GetByID", mock.Anything, testPartner.ID).Return(testPartner, nil)
	suite.mocks.CouponRepo.On("CodeExists", mock.Anything, mock.AnythingOfType("string")).Return(false, nil)
	suite.mocks.CouponRepo.On("Create", mock.Anything, mock.AnythingOfType("*coupon.Coupon")).Return(nil)
	suite.mocks.PaymentRepo.On("UpdateOrderCoupon", mock.Anything, actualOrderNumber, mock.AnythingOfType("uuid.UUID")).Return(nil)

	err = suite.service.ProcessWebhookNotification(context.Background(), notification)
	require.NoError(t, err)

	// Шаг 4: Проверка финального статуса заказа
	testCoupon := suite.testData.Coupon
	paidOrder := &Order{
		ID:              pendingOrder.ID,
		OrderNumber:     actualOrderNumber,
		Status:          OrderStatusPaid,
		AlfaBankOrderID: &alfaResponse.OrderId,
		Amount:          100000,
		Currency:        "RUB",
		Size:            Size40x50,
		Style:           StyleGrayscale,
		UserEmail:       "customer@example.com",
		PartnerID:       &testPartner.ID,
		CouponID:        &testCoupon.ID,
	}

	suite.mocks.PaymentRepo.ExpectedCalls = nil
	suite.mocks.CouponRepo.ExpectedCalls = nil

	suite.mocks.PaymentRepo.On("GetOrderByNumber", mock.Anything, actualOrderNumber).Return(paidOrder, nil)
	suite.mocks.CouponRepo.On("GetByID", mock.Anything, testCoupon.ID).Return(testCoupon, nil)

	finalStatusResult, err := suite.service.GetOrderStatus(context.Background(), actualOrderNumber)
	require.NoError(t, err)
	assert.True(t, finalStatusResult.Success)
	assert.Equal(t, OrderStatusPaid, finalStatusResult.Status)
	assert.NotNil(t, finalStatusResult.CouponCode)
	assert.Equal(t, testCoupon.Code, *finalStatusResult.CouponCode)

	// Проверяем все ожидания
	suite.mocks.PaymentRepo.AssertExpectations(t)
	suite.mocks.CouponRepo.AssertExpectations(t)
	suite.mocks.PartnerRepo.AssertExpectations(t)
	suite.mocks.AlfaBankClient.AssertExpectations(t)
}

func TestIntegration_PaymentReturnWorkflow(t *testing.T) {
	suite := SetupIntegrationTest(t)

	testOrder := &Order{
		ID:          uuid.New(),
		OrderNumber: "TEST-ORDER-RETURN-123",
		Status:      OrderStatusPending,
		Amount:      100000,
		Currency:    "RUB",
		Size:        Size30x40,
		Style:       StyleGrayscale,
		UserEmail:   "customer@example.com",
	}
	alfaBankOrderID := "alfa_order_123"
	testOrder.AlfaBankOrderID = &alfaBankOrderID

	alfaStatusResponse := &AlfaBankStatusResponse{
		OrderNumber: testOrder.OrderNumber,
		OrderStatus: 2,
		Amount:      testOrder.Amount,
		Currency:    testOrder.Currency,
	}

	suite.mocks.PaymentRepo.On("GetOrderByNumber", mock.Anything, testOrder.OrderNumber).Return(testOrder, nil)
	suite.mocks.AlfaBankClient.On("GetOrderStatus", mock.Anything, alfaBankOrderID).Return(alfaStatusResponse, nil)
	suite.mocks.PaymentRepo.On("UpdateOrderStatus", mock.Anything, testOrder.OrderNumber, OrderStatusPaid, &alfaBankOrderID).Return(nil)

	suite.mocks.CouponRepo.On("CodeExists", mock.Anything, mock.AnythingOfType("string")).Return(false, nil)
	suite.mocks.CouponRepo.On("Create", mock.Anything, mock.AnythingOfType("*coupon.Coupon")).Return(nil)
	suite.mocks.PaymentRepo.On("UpdateOrderCoupon", mock.Anything, testOrder.OrderNumber, mock.AnythingOfType("uuid.UUID")).Return(nil)

	err := suite.service.ProcessPaymentReturn(context.Background(), testOrder.OrderNumber)

	require.NoError(t, err)
	suite.mocks.PaymentRepo.AssertExpectations(t)
	suite.mocks.CouponRepo.AssertExpectations(t)
	suite.mocks.PartnerRepo.AssertExpectations(t)
	suite.mocks.AlfaBankClient.AssertExpectations(t)
}

func TestIntegration_MultipleWebhookNotifications(t *testing.T) {
	testOrder := createTestOrder()
	testOrder.Status = OrderStatusPending
	alfaBankOrderID := "alfa_order_123"

	testCases := []struct {
		name               string
		orderStatus        int
		expectedStatus     string
		shouldCreateCoupon bool
	}{
		{
			name:               "Pre-authorization",
			orderStatus:        1,
			expectedStatus:     OrderStatusPending,
			shouldCreateCoupon: false,
		},
		{
			name:               "Successful payment",
			orderStatus:        2,
			expectedStatus:     OrderStatusPaid,
			shouldCreateCoupon: true,
		},
		{
			name:               "Payment cancelled",
			orderStatus:        3,
			expectedStatus:     OrderStatusCancelled,
			shouldCreateCoupon: false,
		},
		{
			name:               "Payment declined",
			orderStatus:        6,
			expectedStatus:     OrderStatusFailed,
			shouldCreateCoupon: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testSuite := SetupIntegrationTest(t)
			order := createTestOrder()
			order.Status = OrderStatusPending

			notification := &PaymentNotificationRequest{
				OrderNumber:     order.OrderNumber,
				OrderStatus:     tc.orderStatus,
				AlfaBankOrderID: alfaBankOrderID,
				Amount:          order.Amount,
				Currency:        order.Currency,
			}

			testSuite.mocks.PaymentRepo.On("GetOrderByNumber", mock.Anything, order.OrderNumber).Return(order, nil)

			if tc.expectedStatus != OrderStatusPending {
				testSuite.mocks.PaymentRepo.On("UpdateOrderStatus", mock.Anything, order.OrderNumber, tc.expectedStatus, &alfaBankOrderID).Return(nil)
			}

			if tc.shouldCreateCoupon {
				testSuite.mocks.CouponRepo.On("CodeExists", mock.Anything, mock.AnythingOfType("string")).Return(false, nil)
				testSuite.mocks.CouponRepo.On("Create", mock.Anything, mock.AnythingOfType("*coupon.Coupon")).Return(nil)
				testSuite.mocks.PaymentRepo.On("UpdateOrderCoupon", mock.Anything, order.OrderNumber, mock.AnythingOfType("uuid.UUID")).Return(nil)
			}

			err := testSuite.service.ProcessWebhookNotification(context.Background(), notification)
			assert.NoError(t, err)

			testSuite.mocks.PaymentRepo.AssertExpectations(t)
			if tc.shouldCreateCoupon {
				testSuite.mocks.CouponRepo.AssertExpectations(t)
			}
		})
	}
}

func TestIntegration_PartnerSpecificOrders(t *testing.T) {
	partner1 := createTestPartner()
	partner1.Domain = "partner1.example.com"
	partner1.PartnerCode = "PRT1"

	partner2 := createTestPartner()
	partner2.ID = uuid.New()
	partner2.Domain = "partner2.example.com"
	partner2.PartnerCode = "PRT2"

	testCases := []struct {
		name    string
		partner *partner.Partner
		domain  string
	}{
		{
			name:    "Partner 1 purchase",
			partner: partner1,
			domain:  partner1.Domain,
		},
		{
			name:    "Partner 2 purchase",
			partner: partner2,
			domain:  partner2.Domain,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testSuite := SetupIntegrationTest(t)

			req := &PurchaseCouponRequest{
				Size:      Size40x50,
				Style:     StyleGrayscale,
				Email:     "test@example.com",
				ReturnURL: "https://test.example.com/return",
				Domain:    &tc.domain,
			}

			alfaResponse := &AlfaBankRegisterResponse{
				OrderId: "alfa_order_" + tc.partner.PartnerCode,
				FormUrl: "https://payment.alfabank.ru/payment/order" + tc.partner.PartnerCode,
			}

			testSuite.mocks.PartnerRepo.On("GetByDomain", mock.Anything, tc.domain).Return(tc.partner, nil)
			testSuite.mocks.PaymentRepo.On("GetOrderByNumber", mock.Anything, mock.AnythingOfType("string")).Return(nil, errors.New("not found"))
			testSuite.mocks.PaymentRepo.On("CreateOrder", mock.Anything, mock.MatchedBy(func(order *Order) bool {
				return order.PartnerID != nil && *order.PartnerID == tc.partner.ID
			})).Return(nil)
			testSuite.mocks.AlfaBankClient.On("RegisterOrder", mock.Anything, mock.AnythingOfType("*payment.AlfaBankRegisterRequest")).Return(alfaResponse, nil)
			testSuite.mocks.PaymentRepo.On("UpdateOrderStatus", mock.Anything, mock.AnythingOfType("string"), OrderStatusPending, &alfaResponse.OrderId).Return(nil)
			testSuite.mocks.PaymentRepo.On("UpdateOrderPaymentURL", mock.Anything, mock.AnythingOfType("string"), alfaResponse.FormUrl).Return(nil)

			result, err := testSuite.service.PurchaseCoupon(context.Background(), req)

			require.NoError(t, err)
			assert.True(t, result.Success)
			assert.Equal(t, alfaResponse.FormUrl, result.PaymentURL)

			testSuite.mocks.PaymentRepo.AssertExpectations(t)
			testSuite.mocks.PartnerRepo.AssertExpectations(t)
			testSuite.mocks.AlfaBankClient.AssertExpectations(t)
		})
	}
}

func TestIntegration_ErrorHandling(t *testing.T) {
	testCases := []struct {
		name          string
		setupMocks    func(*IntegrationTestSuite)
		action        func(*IntegrationTestSuite) error
		expectedError string
	}{
		{
			name: "AlfaBank registration failure",
			setupMocks: func(s *IntegrationTestSuite) {
				s.mocks.PaymentRepo.On("GetOrderByNumber", mock.Anything, mock.AnythingOfType("string")).Return(nil, errors.New("not found"))
				s.mocks.PaymentRepo.On("CreateOrder", mock.Anything, mock.AnythingOfType("*payment.Order")).Return(nil)
				s.mocks.AlfaBankClient.On("RegisterOrder", mock.Anything, mock.AnythingOfType("*payment.AlfaBankRegisterRequest")).Return(nil, errors.New("bank error"))
			},
			action: func(s *IntegrationTestSuite) error {
				req := &PurchaseCouponRequest{
					Size:      Size40x50,
					Style:     StyleGrayscale,
					Email:     "test@example.com",
					ReturnURL: "https://test.example.com/return",
				}
				_, err := s.service.PurchaseCoupon(context.Background(), req)
				return err
			},
			expectedError: "",
		},
		{
			name: "Database order creation failure",
			setupMocks: func(s *IntegrationTestSuite) {
				s.mocks.PaymentRepo.On("GetOrderByNumber", mock.Anything, mock.AnythingOfType("string")).Return(nil, errors.New("not found"))
				s.mocks.PaymentRepo.On("CreateOrder", mock.Anything, mock.AnythingOfType("*payment.Order")).Return(errors.New("db error"))
			},
			action: func(s *IntegrationTestSuite) error {
				req := &PurchaseCouponRequest{
					Size:      Size40x50,
					Style:     StyleGrayscale,
					Email:     "test@example.com",
					ReturnURL: "https://test.example.com/return",
				}
				_, err := s.service.PurchaseCoupon(context.Background(), req)
				return err
			},
			expectedError: "",
		},
		{
			name: "Coupon creation failure during webhook",
			setupMocks: func(s *IntegrationTestSuite) {
				order := createTestOrder()
				order.Status = OrderStatusPending
				s.mocks.PaymentRepo.On("GetOrderByNumber", mock.Anything, order.OrderNumber).Return(order, nil)
				s.mocks.PaymentRepo.On("UpdateOrderStatus", mock.Anything, order.OrderNumber, OrderStatusPaid, mock.AnythingOfType("*string")).Return(nil)
				s.mocks.CouponRepo.On("CodeExists", mock.Anything, mock.AnythingOfType("string")).Return(false, nil)
				s.mocks.CouponRepo.On("Create", mock.Anything, mock.AnythingOfType("*coupon.Coupon")).Return(errors.New("coupon creation failed"))
			},
			action: func(s *IntegrationTestSuite) error {
				notification := &PaymentNotificationRequest{
					OrderNumber:     "ORD_1234567890_abcd1234",
					OrderStatus:     2,
					AlfaBankOrderID: "alfa_123",
				}
				return s.service.ProcessWebhookNotification(context.Background(), notification)
			},
			expectedError: "coupon creation failed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testSuite := SetupIntegrationTest(t)
			tc.setupMocks(testSuite)

			err := tc.action(testSuite)

			if tc.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			}
		})
	}
}

func TestIntegration_ConcurrentOrders(t *testing.T) {
	suite := SetupIntegrationTest(t)

	numGoroutines := 5
	results := make(chan error, numGoroutines)

	alfaResponse := &AlfaBankRegisterResponse{
		OrderId: "alfa_order_concurrent",
		FormUrl: "https://payment.alfabank.ru/payment/order_concurrent",
	}

	suite.mocks.PaymentRepo.On("GetOrderByNumber", mock.Anything, mock.AnythingOfType("string")).Return(nil, errors.New("not found")).Times(numGoroutines)
	suite.mocks.PaymentRepo.On("CreateOrder", mock.Anything, mock.AnythingOfType("*payment.Order")).Return(nil).Times(numGoroutines)
	suite.mocks.AlfaBankClient.On("RegisterOrder", mock.Anything, mock.AnythingOfType("*payment.AlfaBankRegisterRequest")).Return(alfaResponse, nil).Times(numGoroutines)
	suite.mocks.PaymentRepo.On("UpdateOrderStatus", mock.Anything, mock.AnythingOfType("string"), OrderStatusPending, &alfaResponse.OrderId).Return(nil).Times(numGoroutines)
	suite.mocks.PaymentRepo.On("UpdateOrderPaymentURL", mock.Anything, mock.AnythingOfType("string"), alfaResponse.FormUrl).Return(nil).Times(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			req := &PurchaseCouponRequest{
				Size:      Size40x50,
				Style:     StyleGrayscale,
				Email:     "test" + string(rune('0'+index)) + "@example.com",
				ReturnURL: "https://test.example.com/return",
			}

			_, err := suite.service.PurchaseCoupon(context.Background(), req)
			results <- err
		}(i)
	}

	for i := 0; i < numGoroutines; i++ {
		err := <-results
		assert.NoError(t, err)
	}

	suite.mocks.PaymentRepo.AssertExpectations(t)
	suite.mocks.AlfaBankClient.AssertExpectations(t)
}

func TestIntegration_OptionsValidation(t *testing.T) {
	deps := &PaymentServiceDeps{
		Config: createTestConfig(),
	}
	service := NewPaymentService(deps)

	options := service.GetAvailableOptions()

	require.NotNil(t, options)
	assert.Len(t, options.Sizes, 6)
	assert.Len(t, options.Styles, 4)

	for _, size := range options.Sizes {
		assert.Equal(t, FixedPriceRub, size.Price)
		assert.NotEmpty(t, size.Value)
		assert.NotEmpty(t, size.Label)
		assert.NotEmpty(t, size.Description)
	}

	for _, style := range options.Styles {
		assert.NotEmpty(t, style.Value)
		assert.NotEmpty(t, style.Label)
		assert.NotEmpty(t, style.Description)
	}

	supportedSizes := []string{Size21x30, Size30x40, Size40x40, Size40x50, Size40x60, Size50x70}
	for _, expectedSize := range supportedSizes {
		found := false
		for _, size := range options.Sizes {
			if size.Value == expectedSize {
				found = true
				break
			}
		}
		assert.True(t, found, "Size %s should be supported", expectedSize)
	}
}

func BenchmarkIntegration_PurchaseWorkflow(b *testing.B) {
	suite := &IntegrationTestSuite{
		mocks: &IntegrationMocks{
			PaymentRepo:    &MockPaymentRepository{},
			CouponRepo:     &MockCouponRepository{},
			PartnerRepo:    &MockPartnerRepository{},
			AlfaBankClient: &MockAlfaBankClient{},
		},
		testData: &IntegrationTestData{
			Config: createTestConfig(),
		},
	}

	deps := &PaymentServiceDeps{
		PaymentRepository: suite.mocks.PaymentRepo,
		CouponRepository:  suite.mocks.CouponRepo,
		PartnerRepository: suite.mocks.PartnerRepo,
		Config:            suite.testData.Config,
	}

	service := NewPaymentService(deps)
	service.alfaClient = suite.mocks.AlfaBankClient

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

	suite.mocks.PaymentRepo.On("GetOrderByNumber", mock.Anything, mock.AnythingOfType("string")).Return(nil, errors.New("not found"))
	suite.mocks.PaymentRepo.On("CreateOrder", mock.Anything, mock.AnythingOfType("*payment.Order")).Return(nil)
	suite.mocks.AlfaBankClient.On("RegisterOrder", mock.Anything, mock.AnythingOfType("*payment.AlfaBankRegisterRequest")).Return(alfaResponse, nil)
	suite.mocks.PaymentRepo.On("UpdateOrderStatus", mock.Anything, mock.AnythingOfType("string"), OrderStatusPending, &alfaResponse.OrderId).Return(nil)
	suite.mocks.PaymentRepo.On("UpdateOrderPaymentURL", mock.Anything, mock.AnythingOfType("string"), alfaResponse.FormUrl).Return(nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := service.PurchaseCoupon(context.Background(), req)
		if err != nil {
			b.Fatal(err)
		}
	}
}
