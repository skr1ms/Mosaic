package public

import (
	"testing"

	"github.com/skr1ms/mosaic/internal/payment"
	"github.com/stretchr/testify/mock"
)


func BenchmarkPublicService_GetCouponByCode_Success(b *testing.B) {
	mockCouponRepo := &MockCouponRepository{}
	deps := &PublicServiceDeps{
		CouponRepository: mockCouponRepo,
	}
	service := NewPublicService(deps)

	testCoupon := createTestCoupon()
	mockCouponRepo.On("GetByCode", mock.Anything, testCoupon.Code).Return(testCoupon, nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := service.GetCouponByCode(testCoupon.Code)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPublicService_GetCouponByCode_ValidationOnly(b *testing.B) {
	deps := &PublicServiceDeps{}
	service := NewPublicService(deps)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := service.GetCouponByCode("invalid")
		if err == nil {
			b.Fatal("Expected validation error")
		}
	}
}

func BenchmarkPublicService_GetAvailableSizes(b *testing.B) {
	service := NewPublicService(&PublicServiceDeps{})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		sizes := service.GetAvailableSizes()
		if len(sizes) != 6 {
			b.Fatalf("Expected 6 sizes, got %d", len(sizes))
		}
	}
}

func BenchmarkPublicService_GetAvailableStyles(b *testing.B) {
	service := NewPublicService(&PublicServiceDeps{})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		styles := service.GetAvailableStyles()
		if len(styles) != 4 {
			b.Fatalf("Expected 4 styles, got %d", len(styles))
		}
	}
}

func BenchmarkPublicService_ActivateCoupon(b *testing.B) {
	mockCouponRepo := &MockCouponRepository{}
	deps := &PublicServiceDeps{
		CouponRepository: mockCouponRepo,
	}
	service := NewPublicService(deps)

	testCoupon := createTestCoupon()
	req := ActivateCouponRequest{Email: "test@example.com"}

	mockCouponRepo.On("GetByCode", mock.Anything, testCoupon.Code).Return(testCoupon, nil)
	mockCouponRepo.On("Update", mock.Anything, mock.AnythingOfType("*coupon.Coupon")).Return(nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := service.ActivateCoupon(testCoupon.Code, req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPublicService_PurchaseCoupon(b *testing.B) {
	mockCouponRepo := &MockCouponRepository{}
	mockPaymentService := &MockPaymentService{}
	deps := &PublicServiceDeps{
		CouponRepository: mockCouponRepo,
		PaymentService:   mockPaymentService,
		Config:           createTestConfig(),
	}
	service := NewPublicService(deps)

	req := PurchaseCouponRequest{
		Size:         "40x50",
		Style:        "grayscale",
		Email:        "test@example.com",
		PaymentToken: "test-token",
	}

	paymentResponse := &payment.PurchaseCouponResponse{
		Success:    true,
		OrderID:    "order123",
		PaymentURL: "https://payment.example.com/order123",
		Message:    "Success",
	}

	mockCouponRepo.On("CodeExists", mock.Anything, mock.AnythingOfType("string")).Return(false, nil)
	mockCouponRepo.On("Create", mock.Anything, mock.AnythingOfType("*coupon.Coupon")).Return(nil)
	mockPaymentService.On("PurchaseCoupon", mock.Anything, mock.AnythingOfType("*payment.PurchaseCouponRequest")).Return(paymentResponse, nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := service.PurchaseCoupon(req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkIsNumeric(b *testing.B) {
	testCases := []struct {
		name  string
		input string
	}{
		{"Valid12Digits", "123456789012"},
		{"Invalid_TooShort", "12345"},
		{"Invalid_WithLetters", "12345a789012"},
		{"Invalid_Empty", ""},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = isNumeric(tc.input)
			}
		})
	}
}


func BenchmarkPublicService_ConcurrentGetCoupon(b *testing.B) {
	mockCouponRepo := &MockCouponRepository{}
	deps := &PublicServiceDeps{
		CouponRepository: mockCouponRepo,
	}
	service := NewPublicService(deps)

	testCoupon := createTestCoupon()
	mockCouponRepo.On("GetByCode", mock.Anything, testCoupon.Code).Return(testCoupon, nil)

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := service.GetCouponByCode(testCoupon.Code)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkPublicService_MixedOperations(b *testing.B) {
	mockCouponRepo := &MockCouponRepository{}
	mockPartnerRepo := &MockPartnerRepository{}
	deps := &PublicServiceDeps{
		CouponRepository:  mockCouponRepo,
		PartnerRepository: mockPartnerRepo,
	}
	service := NewPublicService(deps)

	testCoupon := createTestCoupon()
	testPartner := createTestPartner()

	mockCouponRepo.On("GetByCode", mock.Anything, testCoupon.Code).Return(testCoupon, nil)
	mockPartnerRepo.On("GetByDomain", mock.Anything, testPartner.Domain).Return(testPartner, nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		switch i % 4 {
		case 0:
			_, err := service.GetCouponByCode(testCoupon.Code)
			if err != nil {
				b.Fatal(err)
			}
		case 1:
			_, err := service.GetPartnerByDomain(testPartner.Domain)
			if err != nil {
				b.Fatal(err)
			}
		case 2:
			sizes := service.GetAvailableSizes()
			if len(sizes) == 0 {
				b.Fatal("No sizes")
			}
		case 3:
			styles := service.GetAvailableStyles()
			if len(styles) == 0 {
				b.Fatal("No styles")
			}
		}
	}
}

func BenchmarkPublicService_MemoryAllocation_GetCoupon(b *testing.B) {
	mockCouponRepo := &MockCouponRepository{}
	deps := &PublicServiceDeps{
		CouponRepository: mockCouponRepo,
	}
	service := NewPublicService(deps)

	testCoupon := createTestCoupon()
	mockCouponRepo.On("GetByCode", mock.Anything, testCoupon.Code).Return(testCoupon, nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		result, err := service.GetCouponByCode(testCoupon.Code)
		if err != nil {
			b.Fatal(err)
		}
		_ = result["id"]
	}
}

func BenchmarkPublicService_MemoryAllocation_GetSizes(b *testing.B) {
	service := NewPublicService(&PublicServiceDeps{})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		sizes := service.GetAvailableSizes()
		_ = sizes[0]["size"]
	}
}

func BenchmarkPublicService_MemoryAllocation_GetStyles(b *testing.B) {
	service := NewPublicService(&PublicServiceDeps{})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		styles := service.GetAvailableStyles()
		_ = styles[0]["style"]
	}
}
