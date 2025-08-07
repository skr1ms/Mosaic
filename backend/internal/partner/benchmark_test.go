package partner

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/skr1ms/mosaic/internal/coupon"
	"github.com/skr1ms/mosaic/pkg/bcrypt"
)

func BenchmarkPartnerRepository_GetByID(b *testing.B) {
	mockRepo := &MockPartnerRepository{}
	testPartner := createTestPartner()

	mockRepo.On("GetByID", mock.Anything, testPartner.ID).Return(testPartner, nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := mockRepo.GetByID(context.Background(), testPartner.ID)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPartnerRepository_GetByEmail(b *testing.B) {
	mockRepo := &MockPartnerRepository{}
	testPartner := createTestPartner()

	mockRepo.On("GetByEmail", mock.Anything, testPartner.Email).Return(testPartner, nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := mockRepo.GetByEmail(context.Background(), testPartner.Email)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPartnerRepository_GetByDomain(b *testing.B) {
	mockRepo := &MockPartnerRepository{}
	testPartner := createTestPartner()

	mockRepo.On("GetByDomain", mock.Anything, testPartner.Domain).Return(testPartner, nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := mockRepo.GetByDomain(context.Background(), testPartner.Domain)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPartnerRepository_GetAll(b *testing.B) {
	mockRepo := &MockPartnerRepository{}
	partners := []*Partner{createTestPartner(), createTestPartner(), createTestPartner()}

	mockRepo.On("GetAll", mock.Anything, "created_at", "desc").Return(partners, nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := mockRepo.GetAll(context.Background(), "created_at", "desc")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPartnerRepository_Search(b *testing.B) {
	mockRepo := &MockPartnerRepository{}
	partners := []*Partner{createTestPartner()}

	mockRepo.On("Search", mock.Anything, "test", "active", "created_at", "desc").Return(partners, nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := mockRepo.Search(context.Background(), "test", "active", "created_at", "desc")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPartnerRepository_UpdatePassword(b *testing.B) {
	mockRepo := &MockPartnerRepository{}
	partnerID := uuid.New()

	mockRepo.On("UpdatePassword", mock.Anything, partnerID, mock.AnythingOfType("string")).Return(nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		err := mockRepo.UpdatePassword(context.Background(), partnerID, "hashedpassword123")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPartnerRepository_GetCouponsStatistics(b *testing.B) {
	mockRepo := &MockPartnerRepository{}
	partnerID := uuid.New()
	stats := map[string]int64{
		"total":     100,
		"activated": 80,
		"new":       20,
		"purchased": 50,
	}

	mockRepo.On("GetCouponsStatistics", mock.Anything, partnerID).Return(stats, nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := mockRepo.GetCouponsStatistics(context.Background(), partnerID)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCouponService_ExportCouponsAdvanced(b *testing.B) {
	mockService := &MockCouponService{}
	partnerID := uuid.New().String()
	options := coupon.ExportOptionsRequest{
		Format:        coupon.ExportFormatCodes,
		PartnerID:     &partnerID,
		Status:        "new",
		FileFormat:    "txt",
		IncludeHeader: false,
	}

	content := []byte("COUPON001\nCOUPON002\nCOUPON003\n")
	filename := "coupons.txt"
	contentType := "text/plain"

	mockService.On("ExportCouponsAdvanced", options).Return(content, filename, contentType, nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _, _, err := mockService.ExportCouponsAdvanced(options)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJWT_CreatePasswordResetToken(b *testing.B) {
	mockJWT := &MockJWT{}
	userID := uuid.New()
	email := "test@example.com"
	token := "reset-token-123"

	mockJWT.On("CreatePasswordResetToken", userID, email).Return(token, nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := mockJWT.CreatePasswordResetToken(userID, email)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJWT_ValidatePasswordResetToken(b *testing.B) {
	mockJWT := &MockJWT{}
	token := "valid-token"
	claims := &TokenClaims{
		UserID: uuid.New(),
		Login:  "test@example.com",
	}

	mockJWT.On("ValidatePasswordResetToken", token).Return(claims, nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := mockJWT.ValidatePasswordResetToken(token)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMailer_SendResetPasswordEmail(b *testing.B) {
	mockMailer := &MockMailer{}
	email := "test@example.com"
	resetLink := "https://example.com/reset?token=abc123"

	mockMailer.On("SendResetPasswordEmail", email, resetLink).Return(nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		err := mockMailer.SendResetPasswordEmail(email, resetLink)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRecaptcha_Verify(b *testing.B) {
	mockRecaptcha := &MockRecaptcha{}
	token := "recaptcha-token"
	action := "forgot_password"

	mockRecaptcha.On("Verify", token, action).Return(true, nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := mockRecaptcha.Verify(token, action)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBcrypt_HashPassword(b *testing.B) {
	password := "testpassword123"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := bcrypt.HashPassword(password)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBcrypt_CheckPassword(b *testing.B) {
	password := "testpassword123"
	hashedPassword := "$2a$10$N8lBWGvDHwgMcYp380JNHubT7mJMNLfLril5y7K7h0wuEdpk2Uhsq"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		isValid := bcrypt.CheckPassword(password, hashedPassword)
		if !isValid {
			b.Fatal("Password validation failed")
		}
	}
}

func BenchmarkPartnerModel_Creation(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		partner := createTestPartner()
		if partner.ID == uuid.Nil {
			b.Fatal("Invalid partner created")
		}
	}
}

func BenchmarkPartnerRepository_ConcurrentGetByID(b *testing.B) {
	mockRepo := &MockPartnerRepository{}
	testPartner := createTestPartner()

	mockRepo.On("GetByID", mock.Anything, testPartner.ID).Return(testPartner, nil)

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := mockRepo.GetByID(context.Background(), testPartner.ID)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkPartnerRepository_ConcurrentGetByEmail(b *testing.B) {
	mockRepo := &MockPartnerRepository{}
	testPartner := createTestPartner()

	mockRepo.On("GetByEmail", mock.Anything, testPartner.Email).Return(testPartner, nil)

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := mockRepo.GetByEmail(context.Background(), testPartner.Email)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkPartnerRepository_ConcurrentSearch(b *testing.B) {
	mockRepo := &MockPartnerRepository{}
	partners := []*Partner{createTestPartner()}

	mockRepo.On("Search", mock.Anything, "test", "active", "created_at", "desc").Return(partners, nil)

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := mockRepo.Search(context.Background(), "test", "active", "created_at", "desc")
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkPartnerRepository_MemoryAllocation_GetByID(b *testing.B) {
	mockRepo := &MockPartnerRepository{}
	testPartner := createTestPartner()

	mockRepo.On("GetByID", mock.Anything, testPartner.ID).Return(testPartner, nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		partner, err := mockRepo.GetByID(context.Background(), testPartner.ID)
		if err != nil {
			b.Fatal(err)
		}
		_ = partner.BrandName
	}
}

func BenchmarkPartnerRepository_MemoryAllocation_GetAll(b *testing.B) {
	mockRepo := &MockPartnerRepository{}
	partners := make([]*Partner, 100) // Симулируем большой результат
	for i := 0; i < 100; i++ {
		partners[i] = createTestPartner()
	}

	mockRepo.On("GetAll", mock.Anything, "created_at", "desc").Return(partners, nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		result, err := mockRepo.GetAll(context.Background(), "created_at", "desc")
		if err != nil {
			b.Fatal(err)
		}
		_ = len(result)
	}
}

func BenchmarkPartnerRepository_MixedOperations(b *testing.B) {
	mockRepo := &MockPartnerRepository{}
	testPartner := createTestPartner()
	partners := []*Partner{testPartner}

	mockRepo.On("GetByID", mock.Anything, testPartner.ID).Return(testPartner, nil)
	mockRepo.On("GetByEmail", mock.Anything, testPartner.Email).Return(testPartner, nil)
	mockRepo.On("GetByDomain", mock.Anything, testPartner.Domain).Return(testPartner, nil)
	mockRepo.On("GetAll", mock.Anything, "created_at", "desc").Return(partners, nil)
	mockRepo.On("CountActive", mock.Anything).Return(int64(10), nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		switch i % 5 {
		case 0:
			_, err := mockRepo.GetByID(context.Background(), testPartner.ID)
			if err != nil {
				b.Fatal(err)
			}
		case 1:
			_, err := mockRepo.GetByEmail(context.Background(), testPartner.Email)
			if err != nil {
				b.Fatal(err)
			}
		case 2:
			_, err := mockRepo.GetByDomain(context.Background(), testPartner.Domain)
			if err != nil {
				b.Fatal(err)
			}
		case 3:
			_, err := mockRepo.GetAll(context.Background(), "created_at", "desc")
			if err != nil {
				b.Fatal(err)
			}
		case 4:
			_, err := mockRepo.CountActive(context.Background())
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}

func BenchmarkPartnerWorkflow_CompletePasswordReset(b *testing.B) {
	mockRepo := &MockPartnerRepository{}
	mockJWT := &MockJWT{}
	mockMailer := &MockMailer{}

	testPartner := createTestPartner()
	resetToken := "reset-token-123"
	resetLink := "https://example.com/reset?token=" + resetToken

	mockRepo.On("GetByEmail", mock.Anything, testPartner.Email).Return(testPartner, nil)
	mockJWT.On("CreatePasswordResetToken", testPartner.ID, testPartner.Email).Return(resetToken, nil)
	mockMailer.On("SendResetPasswordEmail", testPartner.Email, resetLink).Return(nil)

	claims := &TokenClaims{
		UserID: testPartner.ID,
		Login:  testPartner.Email,
	}
	mockJWT.On("ValidatePasswordResetToken", resetToken).Return(claims, nil)
	mockRepo.On("UpdatePassword", mock.Anything, testPartner.ID, mock.AnythingOfType("string")).Return(nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := mockRepo.GetByEmail(context.Background(), testPartner.Email)
		if err != nil {
			b.Fatal(err)
		}

		_, err = mockJWT.CreatePasswordResetToken(testPartner.ID, testPartner.Email)
		if err != nil {
			b.Fatal(err)
		}

		err = mockMailer.SendResetPasswordEmail(testPartner.Email, resetLink)
		if err != nil {
			b.Fatal(err)
		}

		_, err = mockJWT.ValidatePasswordResetToken(resetToken)
		if err != nil {
			b.Fatal(err)
		}

		_, err = mockRepo.GetByEmail(context.Background(), testPartner.Email)
		if err != nil {
			b.Fatal(err)
		}

		err = mockRepo.UpdatePassword(context.Background(), testPartner.ID, "newhashed123")
		if err != nil {
			b.Fatal(err)
		}
	}
}
