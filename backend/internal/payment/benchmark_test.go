package payment

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
)

func BenchmarkPaymentService_PurchaseCoupon(b *testing.B) {
	mockPaymentRepo := &MockPaymentRepository{}
	mockCouponRepo := &MockCouponRepository{}
	mockPartnerRepo := &MockPartnerRepository{}
	mockAlfaClient := &MockAlfaBankClient{}

	deps := &PaymentServiceDeps{
		PaymentRepository: mockPaymentRepo,
		CouponRepository:  mockCouponRepo,
		PartnerRepository: mockPartnerRepo,
		Config:            createTestConfig(),
	}

	service := NewPaymentService(deps)
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

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := service.PurchaseCoupon(context.Background(), req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPaymentService_GetOrderStatus(b *testing.B) {
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

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := service.GetOrderStatus(context.Background(), testOrder.OrderNumber)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPaymentService_GetAvailableOptions(b *testing.B) {
	deps := &PaymentServiceDeps{
		Config: createTestConfig(),
	}
	service := NewPaymentService(deps)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		options := service.GetAvailableOptions()
		if len(options.Sizes) != 6 || len(options.Styles) != 4 {
			b.Fatalf("Unexpected options structure: %d sizes, %d styles", len(options.Sizes), len(options.Styles))
		}
	}
}

func BenchmarkPaymentService_ProcessWebhookNotification(b *testing.B) {
	mockPaymentRepo := &MockPaymentRepository{}
	mockCouponRepo := &MockCouponRepository{}

	deps := &PaymentServiceDeps{
		PaymentRepository: mockPaymentRepo,
		CouponRepository:  mockCouponRepo,
		Config:            createTestConfig(),
	}

	service := NewPaymentService(deps)

	testOrder := createTestOrder()
	testOrder.Status = OrderStatusPending
	alfaBankOrderID := "alfa_order_123"

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

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		err := service.ProcessWebhookNotification(context.Background(), notification)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPaymentService_generateOrderNumber(b *testing.B) {
	deps := &PaymentServiceDeps{
		Config: createTestConfig(),
	}
	service := NewPaymentService(deps)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		orderNumber := service.generateOrderNumber()
		if len(orderNumber) < 10 {
			b.Fatalf("Generated order number too short: %s", orderNumber)
		}
	}
}

func BenchmarkPaymentService_generateUniqueCouponCode(b *testing.B) {
	mockCouponRepo := &MockCouponRepository{}

	deps := &PaymentServiceDeps{
		CouponRepository: mockCouponRepo,
		Config:           createTestConfig(),
	}

	service := NewPaymentService(deps)

	mockCouponRepo.On("CodeExists", mock.Anything, mock.AnythingOfType("string")).Return(false, nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		code, err := service.generateUniqueCouponCode("TEST")
		if err != nil {
			b.Fatal(err)
		}
		if len(code) < 10 {
			b.Fatalf("Generated coupon code too short: %s", code)
		}
	}
}

func BenchmarkPaymentService_validateWebhookSignature(b *testing.B) {
	deps := &PaymentServiceDeps{
		Config: createTestConfig(),
	}
	service := NewPaymentService(deps)

	notification := &PaymentNotificationRequest{
		OrderNumber:     "test_order",
		OrderStatus:     2,
		AlfaBankOrderID: "alfa_123",
		Amount:          10000,
		Currency:        "RUB",
		Checksum:        "test_checksum",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = service.validateWebhookSignature(notification)
	}
}

func BenchmarkPaymentService_createCouponForOrder(b *testing.B) {
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
	mockCouponRepo.On("Create", mock.Anything, mock.AnythingOfType("*coupon.Coupon")).Return(nil)
	mockPaymentRepo.On("UpdateOrderCoupon", mock.Anything, testOrder.OrderNumber, mock.AnythingOfType("uuid.UUID")).Return(nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		err := service.createCouponForOrder(context.Background(), testOrder)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPaymentService_ConcurrentPurchases(b *testing.B) {
	mockPaymentRepo := &MockPaymentRepository{}
	mockCouponRepo := &MockCouponRepository{}
	mockPartnerRepo := &MockPartnerRepository{}
	mockAlfaClient := &MockAlfaBankClient{}

	deps := &PaymentServiceDeps{
		PaymentRepository: mockPaymentRepo,
		CouponRepository:  mockCouponRepo,
		PartnerRepository: mockPartnerRepo,
		Config:            createTestConfig(),
	}

	service := NewPaymentService(deps)
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

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := service.PurchaseCoupon(context.Background(), req)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkPaymentService_MixedOperations(b *testing.B) {
	mockPaymentRepo := &MockPaymentRepository{}
	mockCouponRepo := &MockCouponRepository{}
	mockPartnerRepo := &MockPartnerRepository{}
	mockAlfaClient := &MockAlfaBankClient{}

	deps := &PaymentServiceDeps{
		PaymentRepository: mockPaymentRepo,
		CouponRepository:  mockCouponRepo,
		PartnerRepository: mockPartnerRepo,
		Config:            createTestConfig(),
	}

	service := NewPaymentService(deps)
	service.alfaClient = mockAlfaClient

	testOrder := createTestOrder()
	testOrder.Status = OrderStatusPaid
	testCoupon := createTestCoupon()
	testOrder.CouponID = &testCoupon.ID

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

	mockPaymentRepo.On("GetOrderByNumber", mock.Anything, mock.AnythingOfType("string")).Return(nil, errors.New("not found")).Maybe()
	mockPaymentRepo.On("GetOrderByNumber", mock.Anything, testOrder.OrderNumber).Return(testOrder, nil).Maybe()
	mockPaymentRepo.On("CreateOrder", mock.Anything, mock.AnythingOfType("*payment.Order")).Return(nil).Maybe()
	mockAlfaClient.On("RegisterOrder", mock.Anything, mock.AnythingOfType("*payment.AlfaBankRegisterRequest")).Return(alfaResponse, nil).Maybe()
	mockPaymentRepo.On("UpdateOrderStatus", mock.Anything, mock.AnythingOfType("string"), OrderStatusPending, &alfaResponse.OrderId).Return(nil).Maybe()
	mockPaymentRepo.On("UpdateOrderPaymentURL", mock.Anything, mock.AnythingOfType("string"), alfaResponse.FormUrl).Return(nil).Maybe()
	mockCouponRepo.On("GetByID", mock.Anything, testCoupon.ID).Return(testCoupon, nil).Maybe()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		switch i % 3 {
		case 0:
			_, err := service.PurchaseCoupon(context.Background(), req)
			if err != nil {
				b.Fatal(err)
			}
		case 1:
			_, err := service.GetOrderStatus(context.Background(), testOrder.OrderNumber)
			if err != nil {
				b.Fatal(err)
			}
		case 2:
			options := service.GetAvailableOptions()
			if len(options.Sizes) == 0 {
				b.Fatal("No sizes returned")
			}
		}
	}
}

func BenchmarkPaymentService_MemoryAllocation_PurchaseCoupon(b *testing.B) {
	mockPaymentRepo := &MockPaymentRepository{}
	mockCouponRepo := &MockCouponRepository{}
	mockPartnerRepo := &MockPartnerRepository{}
	mockAlfaClient := &MockAlfaBankClient{}

	deps := &PaymentServiceDeps{
		PaymentRepository: mockPaymentRepo,
		CouponRepository:  mockCouponRepo,
		PartnerRepository: mockPartnerRepo,
		Config:            createTestConfig(),
	}

	service := NewPaymentService(deps)
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

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		result, err := service.PurchaseCoupon(context.Background(), req)
		if err != nil {
			b.Fatal(err)
		}
		_ = result.Success
		_ = result.PaymentURL
	}
}

func BenchmarkPaymentService_MemoryAllocation_GetOptions(b *testing.B) {
	deps := &PaymentServiceDeps{
		Config: createTestConfig(),
	}
	service := NewPaymentService(deps)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		options := service.GetAvailableOptions()
		_ = options.Sizes[0].Value
		_ = options.Styles[0].Value
	}
}

func BenchmarkPaymentService_MemoryAllocation_OrderNumber(b *testing.B) {
	deps := &PaymentServiceDeps{
		Config: createTestConfig(),
	}
	service := NewPaymentService(deps)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		orderNumber := service.generateOrderNumber()
		_ = orderNumber
	}
}
