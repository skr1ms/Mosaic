package auth

import (
	"context"
	"testing"

	"github.com/skr1ms/mosaic/internal/admin"
	"github.com/skr1ms/mosaic/pkg/jwt"
	"github.com/stretchr/testify/mock"
)

func BenchmarkAuthService_AdminLogin_Success(b *testing.B) {
	mockAdminRepo := &MockAdminRepository{}
	mockJwtService := &MockJwtService{}

	testAdmin := createTestAdmin()
	testTokenPair := createTestTokenPair()

	mockAdminRepo.On("GetByLogin", "admin123").Return(testAdmin, nil)
	mockJwtService.On("CreateTokenPair", testAdmin.ID, testAdmin.Login, "admin").Return(testTokenPair, nil)
	mockAdminRepo.On("UpdateLastLogin", testAdmin.ID).Return(nil)

	service := NewAuthService(&AuthServiceDeps{
		AdminRepository: mockAdminRepo,
		JwtService:      mockJwtService,
	})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _, err := service.AdminLogin("admin123", "password123")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAuthService_PartnerLogin_Success(b *testing.B) {
	mockPartnerRepo := &MockPartnerRepository{}
	mockJwtService := &MockJwtService{}

	testPartner := createTestPartner()
	testTokenPair := createTestTokenPair()

	mockPartnerRepo.On("GetByLogin", mock.Anything, "partner123").Return(testPartner, nil)
	mockJwtService.On("CreateTokenPair", testPartner.ID, testPartner.Login, "partner").Return(testTokenPair, nil)
	mockPartnerRepo.On("UpdateLastLogin", mock.Anything, testPartner.ID).Return(nil)

	service := NewAuthService(&AuthServiceDeps{
		PartnerRepository: mockPartnerRepo,
		JwtService:        mockJwtService,
	})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _, err := service.PartnerLogin("partner123", "password123")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAuthService_RefreshAdminTokens_Success(b *testing.B) {
	mockJwtService := &MockJwtService{}

	testClaims := createTestClaims("admin")
	testTokenPair := createTestTokenPair()

	mockJwtService.On("ValidateRefreshToken", "valid_refresh_token").Return(testClaims, nil)
	mockJwtService.On("RefreshTokens", "valid_refresh_token").Return(testTokenPair, nil)

	service := NewAuthService(&AuthServiceDeps{
		JwtService: mockJwtService,
	})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := service.RefreshAdminTokens("valid_refresh_token")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAuthService_RefreshPartnerTokens_Success(b *testing.B) {
	mockJwtService := &MockJwtService{}

	testClaims := createTestClaims("partner")
	testTokenPair := createTestTokenPair()

	mockJwtService.On("ValidateRefreshToken", "valid_refresh_token").Return(testClaims, nil)
	mockJwtService.On("RefreshTokens", "valid_refresh_token").Return(testTokenPair, nil)

	service := NewAuthService(&AuthServiceDeps{
		JwtService: mockJwtService,
	})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := service.RefreshPartnerTokens("valid_refresh_token")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAuthService_ConcurrentAdminLogins(b *testing.B) {
	mockAdminRepo := &MockAdminRepository{}
	mockJwtService := &MockJwtService{}

	testAdmin := createTestAdmin()
	testTokenPair := createTestTokenPair()

	mockAdminRepo.On("GetByLogin", "admin123").Return(testAdmin, nil)
	mockJwtService.On("CreateTokenPair", testAdmin.ID, testAdmin.Login, "admin").Return(testTokenPair, nil)
	mockAdminRepo.On("UpdateLastLogin", testAdmin.ID).Return(nil)

	service := NewAuthService(&AuthServiceDeps{
		AdminRepository: mockAdminRepo,
		JwtService:      mockJwtService,
	})

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _, err := service.AdminLogin("admin123", "password123")
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkAuthService_ConcurrentPartnerLogins(b *testing.B) {
	mockPartnerRepo := &MockPartnerRepository{}
	mockJwtService := &MockJwtService{}

	testPartner := createTestPartner()
	testTokenPair := createTestTokenPair()

	mockPartnerRepo.On("GetByLogin", mock.Anything, "partner123").Return(testPartner, nil)
	mockJwtService.On("CreateTokenPair", testPartner.ID, testPartner.Login, "partner").Return(testTokenPair, nil)
	mockPartnerRepo.On("UpdateLastLogin", mock.Anything, testPartner.ID).Return(nil)

	service := NewAuthService(&AuthServiceDeps{
		PartnerRepository: mockPartnerRepo,
		JwtService:        mockJwtService,
	})

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _, err := service.PartnerLogin("partner123", "password123")
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkAuthService_MixedOperations(b *testing.B) {
	mockAdminRepo := &MockAdminRepository{}
	mockPartnerRepo := &MockPartnerRepository{}
	mockJwtService := &MockJwtService{}

	testAdmin := createTestAdmin()
	testPartner := createTestPartner()
	testTokenPair := createTestTokenPair()
	testClaims := createTestClaims("admin")

	mockAdminRepo.On("GetByLogin", "admin123").Return(testAdmin, nil)
	mockAdminRepo.On("UpdateLastLogin", testAdmin.ID).Return(nil)
	mockPartnerRepo.On("GetByLogin", mock.Anything, "partner123").Return(testPartner, nil)
	mockPartnerRepo.On("UpdateLastLogin", mock.Anything, testPartner.ID).Return(nil)
	mockJwtService.On("CreateTokenPair", mock.Anything, mock.Anything, mock.Anything).Return(testTokenPair, nil)
	mockJwtService.On("ValidateRefreshToken", "refresh_token").Return(testClaims, nil)
	mockJwtService.On("RefreshTokens", "refresh_token").Return(testTokenPair, nil)

	service := NewAuthService(&AuthServiceDeps{
		AdminRepository:   mockAdminRepo,
		PartnerRepository: mockPartnerRepo,
		JwtService:        mockJwtService,
	})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		switch i % 4 {
		case 0:
			_, _, err := service.AdminLogin("admin123", "password123")
			if err != nil {
				b.Fatal(err)
			}
		case 1:
			_, _, err := service.PartnerLogin("partner123", "password123")
			if err != nil {
				b.Fatal(err)
			}
		case 2:
			_, err := service.RefreshAdminTokens("refresh_token")
			if err != nil {
				b.Fatal(err)
			}
		case 3:
			_, err := service.RefreshPartnerTokens("refresh_token")
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}

func BenchmarkAuthService_MemoryAllocation_AdminLogin(b *testing.B) {
	mockAdminRepo := &MockAdminRepository{}
	mockJwtService := &MockJwtService{}

	testAdmin := createTestAdmin()
	testTokenPair := createTestTokenPair()

	mockAdminRepo.On("GetByLogin", "admin123").Return(testAdmin, nil)
	mockJwtService.On("CreateTokenPair", testAdmin.ID, testAdmin.Login, "admin").Return(testTokenPair, nil)
	mockAdminRepo.On("UpdateLastLogin", testAdmin.ID).Return(nil)

	service := NewAuthService(&AuthServiceDeps{
		AdminRepository: mockAdminRepo,
		JwtService:      mockJwtService,
	})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		admin, tokens, err := service.AdminLogin("admin123", "password123")
		if err != nil {
			b.Fatal(err)
		}
		_ = admin.ID
		_ = tokens.AccessToken
	}
}

func BenchmarkAuthService_MemoryAllocation_RefreshTokens(b *testing.B) {
	mockJwtService := &MockJwtService{}

	testClaims := createTestClaims("admin")
	testTokenPair := createTestTokenPair()

	mockJwtService.On("ValidateRefreshToken", "valid_refresh_token").Return(testClaims, nil)
	mockJwtService.On("RefreshTokens", "valid_refresh_token").Return(testTokenPair, nil)

	service := NewAuthService(&AuthServiceDeps{
		JwtService: mockJwtService,
	})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		tokens, err := service.RefreshAdminTokens("valid_refresh_token")
		if err != nil {
			b.Fatal(err)
		}
		_ = tokens.AccessToken
		_ = tokens.RefreshToken
	}
}

func BenchmarkAuthService_ErrorPath_AdminNotFound(b *testing.B) {
	mockAdminRepo := &MockAdminRepository{}

	mockAdminRepo.On("GetByLogin", "nonexistent@test.com").Return(nil, context.DeadlineExceeded)

	service := NewAuthService(&AuthServiceDeps{
		AdminRepository: mockAdminRepo,
	})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _, err := service.AdminLogin("nonexistent@test.com", "password123")
		if err == nil {
			b.Fatal("Expected an error")
		}
	}
}

func BenchmarkAuthService_ErrorPath_InvalidRefreshToken(b *testing.B) {
	mockJwtService := &MockJwtService{}

	mockJwtService.On("ValidateRefreshToken", "invalid_token").Return(nil, context.DeadlineExceeded)

	service := NewAuthService(&AuthServiceDeps{
		JwtService: mockJwtService,
	})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := service.RefreshAdminTokens("invalid_token")
		if err == nil {
			b.Fatal("Expected an error")
		}
	}
}

func BenchmarkAuthService_LargeScale_AdminLogin(b *testing.B) {
	mockAdminRepo := &MockAdminRepository{}
	mockJwtService := &MockJwtService{}

	numAdmins := 1000
	admins := make([]*admin.Admin, numAdmins)
	tokenPairs := make([]*jwt.TokenPair, numAdmins)

	for i := 0; i < numAdmins; i++ {
		admins[i] = createTestAdmin()
		tokenPairs[i] = createTestTokenPair()

		login := admins[i].Login
		mockAdminRepo.On("GetByLogin", login).Return(admins[i], nil)
		mockJwtService.On("CreateTokenPair", admins[i].ID, admins[i].Login, "admin").Return(tokenPairs[i], nil)
		mockAdminRepo.On("UpdateLastLogin", admins[i].ID).Return(nil)
	}

	service := NewAuthService(&AuthServiceDeps{
		AdminRepository: mockAdminRepo,
		JwtService:      mockJwtService,
	})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		adminIndex := i % numAdmins
		_, _, err := service.AdminLogin(admins[adminIndex].Login, "password123")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAuthService_StressTest_ConcurrentMixedOperations(b *testing.B) {
	mockAdminRepo := &MockAdminRepository{}
	mockPartnerRepo := &MockPartnerRepository{}
	mockJwtService := &MockJwtService{}

	testAdmin := createTestAdmin()
	testPartner := createTestPartner()
	testTokenPair := createTestTokenPair()
	testClaims := createTestClaims("admin")

	mockAdminRepo.On("GetByLogin", mock.Anything).Return(testAdmin, nil)
	mockAdminRepo.On("UpdateLastLogin", mock.Anything).Return(nil)
	mockPartnerRepo.On("GetByLogin", mock.Anything, mock.Anything).Return(testPartner, nil)
	mockPartnerRepo.On("UpdateLastLogin", mock.Anything, mock.Anything).Return(nil)
	mockJwtService.On("CreateTokenPair", mock.Anything, mock.Anything, mock.Anything).Return(testTokenPair, nil)
	mockJwtService.On("ValidateRefreshToken", mock.Anything).Return(testClaims, nil)
	mockJwtService.On("RefreshTokens", mock.Anything).Return(testTokenPair, nil)

	service := NewAuthService(&AuthServiceDeps{
		AdminRepository:   mockAdminRepo,
		PartnerRepository: mockPartnerRepo,
		JwtService:        mockJwtService,
	})

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			switch i % 6 {
			case 0:
				_, _, err := service.AdminLogin("admin123", "password123")
				if err != nil {
					b.Fatal(err)
				}
			case 1:
				_, _, err := service.PartnerLogin("partner123", "password123")
				if err != nil {
					b.Fatal(err)
				}
			case 2:
				_, err := service.RefreshAdminTokens("refresh_token")
				if err != nil {
					b.Fatal(err)
				}
			case 3:
				_, err := service.RefreshPartnerTokens("refresh_token")
				if err != nil {
					b.Fatal(err)
				}
			case 4:
				_, _, _ = service.AdminLogin("", "")
			case 5:
				_, _ = service.RefreshAdminTokens("")
			}
			i++
		}
	})
}
