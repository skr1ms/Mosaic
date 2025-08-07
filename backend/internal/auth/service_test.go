package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/skr1ms/mosaic/internal/admin"
	"github.com/skr1ms/mosaic/internal/partner"
	"github.com/skr1ms/mosaic/pkg/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockAdminRepository struct {
	mock.Mock
}

var _ AdminRepositoryInterface = (*MockAdminRepository)(nil)

func (m *MockAdminRepository) GetByLogin(login string) (*admin.Admin, error) {
	args := m.Called(login)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*admin.Admin), args.Error(1)
}

func (m *MockAdminRepository) UpdateLastLogin(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

type MockPartnerRepository struct {
	mock.Mock
}

var _ PartnerRepositoryInterface = (*MockPartnerRepository)(nil)

func (m *MockPartnerRepository) GetByLogin(ctx context.Context, login string) (*partner.Partner, error) {
	args := m.Called(ctx, login)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*partner.Partner), args.Error(1)
}

func (m *MockPartnerRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockJwtService struct {
	mock.Mock
}

var _ JwtServiceInterface = (*MockJwtService)(nil)

func (m *MockJwtService) CreateTokenPair(userID uuid.UUID, login, role string) (*jwt.TokenPair, error) {
	args := m.Called(userID, login, role)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*jwt.TokenPair), args.Error(1)
}

func (m *MockJwtService) ValidateRefreshToken(refreshToken string) (*jwt.Claims, error) {
	args := m.Called(refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*jwt.Claims), args.Error(1)
}

func (m *MockJwtService) RefreshTokens(refreshToken string) (*jwt.TokenPair, error) {
	args := m.Called(refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*jwt.TokenPair), args.Error(1)
}

func createTestAdmin() *admin.Admin {
	return &admin.Admin{
		ID:       uuid.New(),
		Login:    "admin123",
		Password: "$2a$10$4MYJR7zzwC2aPqvxPCyXiubzQ.FA0NGn7A07GIN1dFmiqIpzhc1.6",
	}
}

func createTestPartner() *partner.Partner {
	return &partner.Partner{
		ID:          uuid.New(),
		PartnerCode: "0001",
		Login:       "partner123",
		Password:    "$2a$10$4MYJR7zzwC2aPqvxPCyXiubzQ.FA0NGn7A07GIN1dFmiqIpzhc1.6",
		BrandName:   "Test Partner",
		Status:      "active",
	}
}

func createTestTokenPair() *jwt.TokenPair {
	return &jwt.TokenPair{
		AccessToken:  "access_token_123",
		RefreshToken: "refresh_token_456",
		ExpiresIn:    3600,
	}
}

func createTestClaims(role string) *jwt.Claims {
	return &jwt.Claims{
		UserID: uuid.New(),
		Login:  "test@example.com",
		Role:   role,
	}
}

func TestAuthService_AdminLogin_Success(t *testing.T) {
	mockAdminRepo := &MockAdminRepository{}
	mockJwtService := &MockJwtService{}

	testAdmin := createTestAdmin()
	testTokenPair := createTestTokenPair()

	mockAdminRepo.On("GetByLogin", "admin123").Return(testAdmin, nil)
	mockJwtService.On("CreateTokenPair", testAdmin.ID, testAdmin.Login, "admin").Return(testTokenPair, nil)
	mockAdminRepo.On("UpdateLastLogin", testAdmin.ID).Return(nil)

	service := &AuthService{
		deps: &AuthServiceDeps{
			AdminRepository: mockAdminRepo,
			JwtService:      mockJwtService,
		},
	}

	admin, tokens, err := service.AdminLogin("admin123", "password123")

	require.NoError(t, err)
	assert.Equal(t, testAdmin.ID, admin.ID)
	assert.Equal(t, testAdmin.Login, admin.Login)
	assert.Equal(t, testTokenPair.AccessToken, tokens.AccessToken)
	assert.Equal(t, testTokenPair.RefreshToken, tokens.RefreshToken)

	mockAdminRepo.AssertExpectations(t)
	mockJwtService.AssertExpectations(t)
}

func TestAuthService_AdminLogin_AdminNotFound(t *testing.T) {
	mockAdminRepo := &MockAdminRepository{}

	mockAdminRepo.On("GetByLogin", "nonexistent@test.com").Return(nil, errors.New("admin not found"))

	service := &AuthService{
		deps: &AuthServiceDeps{
			AdminRepository: mockAdminRepo,
		},
	}

	admin, tokens, err := service.AdminLogin("nonexistent@test.com", "password123")

	assert.Error(t, err)
	assert.Nil(t, admin)
	assert.Nil(t, tokens)
	assert.Contains(t, err.Error(), "admin not found")

	mockAdminRepo.AssertExpectations(t)
}

func TestAuthService_AdminLogin_TokenGenerationFailed(t *testing.T) {
	mockAdminRepo := &MockAdminRepository{}
	mockJwtService := &MockJwtService{}

	testAdmin := createTestAdmin()

	mockAdminRepo.On("GetByLogin", "admin123").Return(testAdmin, nil)
	mockJwtService.On("CreateTokenPair", testAdmin.ID, testAdmin.Login, "admin").Return(nil, errors.New("token generation failed"))

	service := &AuthService{
		deps: &AuthServiceDeps{
			AdminRepository: mockAdminRepo,
			JwtService:      mockJwtService,
		},
	}

	admin, tokens, err := service.AdminLogin("admin123", "password123")

	assert.Error(t, err)
	assert.Nil(t, admin)
	assert.Nil(t, tokens)
	assert.Contains(t, err.Error(), "failed to create token pair")

	mockAdminRepo.AssertExpectations(t)
	mockJwtService.AssertExpectations(t)
}

func TestAuthService_PartnerLogin_Success(t *testing.T) {
	mockPartnerRepo := &MockPartnerRepository{}
	mockJwtService := &MockJwtService{}

	testPartner := createTestPartner()
	testTokenPair := createTestTokenPair()

	mockPartnerRepo.On("GetByLogin", mock.Anything, "partner@test.com").Return(testPartner, nil)
	mockJwtService.On("CreateTokenPair", testPartner.ID, testPartner.Login, "partner").Return(testTokenPair, nil)
	mockPartnerRepo.On("UpdateLastLogin", mock.Anything, testPartner.ID).Return(nil)

	service := &AuthService{
		deps: &AuthServiceDeps{
			PartnerRepository: mockPartnerRepo,
			JwtService:        mockJwtService,
		},
	}

	partner, tokens, err := service.PartnerLogin("partner@test.com", "password123")

	require.NoError(t, err)
	assert.Equal(t, testPartner.ID, partner.ID)
	assert.Equal(t, testPartner.Login, partner.Login)
	assert.Equal(t, testTokenPair.AccessToken, tokens.AccessToken)

	mockPartnerRepo.AssertExpectations(t)
	mockJwtService.AssertExpectations(t)
}

func TestAuthService_PartnerLogin_PartnerBlocked(t *testing.T) {
	mockPartnerRepo := &MockPartnerRepository{}

	testPartner := createTestPartner()
	testPartner.Status = "blocked"

	mockPartnerRepo.On("GetByLogin", mock.Anything, "partner@test.com").Return(testPartner, nil)

	service := &AuthService{
		deps: &AuthServiceDeps{
			PartnerRepository: mockPartnerRepo,
		},
	}

	partner, tokens, err := service.PartnerLogin("partner@test.com", "password123")

	assert.Error(t, err)
	assert.Nil(t, partner)
	assert.Nil(t, tokens)
	assert.Contains(t, err.Error(), "partner blocked")

	mockPartnerRepo.AssertExpectations(t)
}

func TestAuthService_PartnerLogin_PartnerNotFound(t *testing.T) {
	mockPartnerRepo := &MockPartnerRepository{}

	mockPartnerRepo.On("GetByLogin", mock.Anything, "nonexistent@test.com").Return(nil, errors.New("partner not found"))

	service := &AuthService{
		deps: &AuthServiceDeps{
			PartnerRepository: mockPartnerRepo,
		},
	}

	partner, tokens, err := service.PartnerLogin("nonexistent@test.com", "password123")

	assert.Error(t, err)
	assert.Nil(t, partner)
	assert.Nil(t, tokens)
	assert.Contains(t, err.Error(), "partner not found")

	mockPartnerRepo.AssertExpectations(t)
}

func TestAuthService_RefreshAdminTokens_Success(t *testing.T) {
	mockJwtService := &MockJwtService{}

	testClaims := createTestClaims("admin")
	testTokenPair := createTestTokenPair()

	mockJwtService.On("ValidateRefreshToken", "valid_refresh_token").Return(testClaims, nil)
	mockJwtService.On("RefreshTokens", "valid_refresh_token").Return(testTokenPair, nil)

	service := &AuthService{
		deps: &AuthServiceDeps{
			JwtService: mockJwtService,
		},
	}

	tokens, err := service.RefreshAdminTokens("valid_refresh_token")

	require.NoError(t, err)
	assert.Equal(t, testTokenPair.AccessToken, tokens.AccessToken)
	assert.Equal(t, testTokenPair.RefreshToken, tokens.RefreshToken)

	mockJwtService.AssertExpectations(t)
}

func TestAuthService_RefreshAdminTokens_InvalidToken(t *testing.T) {
	mockJwtService := &MockJwtService{}

	mockJwtService.On("ValidateRefreshToken", "invalid_token").Return(nil, errors.New("invalid token"))

	service := &AuthService{
		deps: &AuthServiceDeps{
			JwtService: mockJwtService,
		},
	}

	tokens, err := service.RefreshAdminTokens("invalid_token")

	assert.Error(t, err)
	assert.Nil(t, tokens)
	assert.Contains(t, err.Error(), "invalid refresh token")

	mockJwtService.AssertExpectations(t)
}

func TestAuthService_RefreshAdminTokens_WrongRole(t *testing.T) {
	mockJwtService := &MockJwtService{}

	testClaims := createTestClaims("partner")

	mockJwtService.On("ValidateRefreshToken", "partner_refresh_token").Return(testClaims, nil)

	service := &AuthService{
		deps: &AuthServiceDeps{
			JwtService: mockJwtService,
		},
	}

	tokens, err := service.RefreshAdminTokens("partner_refresh_token")

	assert.Error(t, err)
	assert.Nil(t, tokens)
	assert.Contains(t, err.Error(), "invalid token role")

	mockJwtService.AssertExpectations(t)
}

func TestAuthService_RefreshPartnerTokens_Success(t *testing.T) {
	mockJwtService := &MockJwtService{}

	testClaims := createTestClaims("partner")
	testTokenPair := createTestTokenPair()

	mockJwtService.On("ValidateRefreshToken", "valid_refresh_token").Return(testClaims, nil)
	mockJwtService.On("RefreshTokens", "valid_refresh_token").Return(testTokenPair, nil)

	service := &AuthService{
		deps: &AuthServiceDeps{
			JwtService: mockJwtService,
		},
	}

	tokens, err := service.RefreshPartnerTokens("valid_refresh_token")

	require.NoError(t, err)
	assert.Equal(t, testTokenPair.AccessToken, tokens.AccessToken)
	assert.Equal(t, testTokenPair.RefreshToken, tokens.RefreshToken)

	mockJwtService.AssertExpectations(t)
}

func TestAuthService_RefreshPartnerTokens_WrongRole(t *testing.T) {
	mockJwtService := &MockJwtService{}

	testClaims := createTestClaims("admin")

	mockJwtService.On("ValidateRefreshToken", "admin_refresh_token").Return(testClaims, nil)

	service := &AuthService{
		deps: &AuthServiceDeps{
			JwtService: mockJwtService,
		},
	}

	tokens, err := service.RefreshPartnerTokens("admin_refresh_token")

	assert.Error(t, err)
	assert.Nil(t, tokens)
	assert.Contains(t, err.Error(), "invalid token role")

	mockJwtService.AssertExpectations(t)
}

func TestAuthService_RefreshPartnerTokens_TokenRefreshFailed(t *testing.T) {
	mockJwtService := &MockJwtService{}

	testClaims := createTestClaims("partner")

	mockJwtService.On("ValidateRefreshToken", "valid_refresh_token").Return(testClaims, nil)
	mockJwtService.On("RefreshTokens", "valid_refresh_token").Return(nil, errors.New("refresh failed"))

	service := &AuthService{
		deps: &AuthServiceDeps{
			JwtService: mockJwtService,
		},
	}

	tokens, err := service.RefreshPartnerTokens("valid_refresh_token")

	assert.Error(t, err)
	assert.Nil(t, tokens)
	assert.Contains(t, err.Error(), "failed to refresh tokens")

	mockJwtService.AssertExpectations(t)
}

func TestLoginRequest_Validation(t *testing.T) {
	validRequest := LoginRequest{
		Login:    "user@example.com",
		Password: "password123",
	}

	assert.Equal(t, "user@example.com", validRequest.Login)
	assert.Equal(t, "password123", validRequest.Password)
}

func TestRefreshTokenRequest_Validation(t *testing.T) {
	validRequest := RefreshTokenRequest{
		RefreshToken: "refresh_token_123",
	}

	assert.Equal(t, "refresh_token_123", validRequest.RefreshToken)
}

func TestAuthService_AdminLogin_UpdateLastLoginFailed(t *testing.T) {
	mockAdminRepo := &MockAdminRepository{}
	mockJwtService := &MockJwtService{}

	testAdmin := createTestAdmin()
	testTokenPair := createTestTokenPair()

	mockAdminRepo.On("GetByLogin", "admin123").Return(testAdmin, nil)
	mockJwtService.On("CreateTokenPair", testAdmin.ID, testAdmin.Login, "admin").Return(testTokenPair, nil)
	mockAdminRepo.On("UpdateLastLogin", testAdmin.ID).Return(errors.New("update failed"))

	service := &AuthService{
		deps: &AuthServiceDeps{
			AdminRepository: mockAdminRepo,
			JwtService:      mockJwtService,
		},
	}

	admin, tokens, err := service.AdminLogin("admin123", "password123")

	assert.Error(t, err)
	assert.Nil(t, admin)
	assert.Nil(t, tokens)
	assert.Contains(t, err.Error(), "failed to update last login")

	mockAdminRepo.AssertExpectations(t)
	mockJwtService.AssertExpectations(t)
}

func TestAuthService_PartnerLogin_UpdateLastLoginFailed(t *testing.T) {
	mockPartnerRepo := &MockPartnerRepository{}
	mockJwtService := &MockJwtService{}

	testPartner := createTestPartner()
	testTokenPair := createTestTokenPair()

	mockPartnerRepo.On("GetByLogin", mock.Anything, "partner@test.com").Return(testPartner, nil)
	mockJwtService.On("CreateTokenPair", testPartner.ID, testPartner.Login, "partner").Return(testTokenPair, nil)
	mockPartnerRepo.On("UpdateLastLogin", mock.Anything, testPartner.ID).Return(errors.New("update failed"))

	service := &AuthService{
		deps: &AuthServiceDeps{
			PartnerRepository: mockPartnerRepo,
			JwtService:        mockJwtService,
		},
	}

	partner, tokens, err := service.PartnerLogin("partner@test.com", "password123")

	assert.Error(t, err)
	assert.Nil(t, partner)
	assert.Nil(t, tokens)
	assert.Contains(t, err.Error(), "failed to update last login")

	mockPartnerRepo.AssertExpectations(t)
	mockJwtService.AssertExpectations(t)
}

func TestAuthService_ConcurrentAdminLogins(t *testing.T) {
	mockAdminRepo := &MockAdminRepository{}
	mockJwtService := &MockJwtService{}

	testAdmin := createTestAdmin()
	testTokenPair := createTestTokenPair()

	mockAdminRepo.On("GetByLogin", "admin123").Return(testAdmin, nil).Times(5)
	mockJwtService.On("CreateTokenPair", testAdmin.ID, testAdmin.Login, "admin").Return(testTokenPair, nil).Times(5)
	mockAdminRepo.On("UpdateLastLogin", testAdmin.ID).Return(nil).Times(5)

	service := &AuthService{
		deps: &AuthServiceDeps{
			AdminRepository: mockAdminRepo,
			JwtService:      mockJwtService,
		},
	}

	numGoroutines := 5
	results := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			admin, tokens, err := service.AdminLogin("admin123", "password123")
			if err != nil {
				results <- err
				return
			}
			if admin == nil || tokens == nil {
				results <- errors.New("nil result")
				return
			}
			results <- nil
		}()
	}

	for i := 0; i < numGoroutines; i++ {
		err := <-results
		assert.NoError(t, err)
	}

	mockAdminRepo.AssertExpectations(t)
	mockJwtService.AssertExpectations(t)
}

func TestAuthService_EdgeCases(t *testing.T) {
	t.Run("Empty login credentials", func(t *testing.T) {
		mockAdminRepo := &MockAdminRepository{}

		mockAdminRepo.On("GetByLogin", "").Return(nil, errors.New("admin not found"))

		service := &AuthService{
			deps: &AuthServiceDeps{
				AdminRepository: mockAdminRepo,
			},
		}

		admin, tokens, err := service.AdminLogin("", "password")

		assert.Error(t, err)
		assert.Nil(t, admin)
		assert.Nil(t, tokens)

		mockAdminRepo.AssertExpectations(t)
	})

	t.Run("Empty password", func(t *testing.T) {
		mockAdminRepo := &MockAdminRepository{}

		testAdmin := createTestAdmin()
		mockAdminRepo.On("GetByLogin", "admin123").Return(testAdmin, nil)

		service := &AuthService{
			deps: &AuthServiceDeps{
				AdminRepository: mockAdminRepo,
			},
		}

		admin, tokens, err := service.AdminLogin("admin123", "")

		assert.Error(t, err)
		assert.Nil(t, admin)
		assert.Nil(t, tokens)
		assert.Contains(t, err.Error(), "invalid credentials")

		mockAdminRepo.AssertExpectations(t)
	})

	t.Run("Empty refresh token", func(t *testing.T) {
		mockJwtService := &MockJwtService{}

		mockJwtService.On("ValidateRefreshToken", "").Return(nil, errors.New("invalid token"))

		service := &AuthService{
			deps: &AuthServiceDeps{
				JwtService: mockJwtService,
			},
		}

		tokens, err := service.RefreshAdminTokens("")

		assert.Error(t, err)
		assert.Nil(t, tokens)

		mockJwtService.AssertExpectations(t)
	})
}

func TestAuthService_NilDependencies(t *testing.T) {
	service := &AuthService{
		deps: &AuthServiceDeps{},
	}

	t.Run("Admin login with nil AdminRepository", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Log("Expected panic due to nil dependency")
			}
		}()

		_, _, err := service.AdminLogin("admin123", "password123")
		if err == nil {
			t.Error("Expected an error due to nil dependency")
		}
	})
}

func TestAuthService_LongRunningOperations(t *testing.T) {
	mockAdminRepo := &MockAdminRepository{}
	mockJwtService := &MockJwtService{}

	testAdmin := createTestAdmin()
	testTokenPair := createTestTokenPair()

	mockAdminRepo.On("GetByLogin", "admin123").Return(testAdmin, nil).After(100 * time.Millisecond)
	mockJwtService.On("CreateTokenPair", testAdmin.ID, testAdmin.Login, "admin").Return(testTokenPair, nil)
	mockAdminRepo.On("UpdateLastLogin", testAdmin.ID).Return(nil)

	service := &AuthService{
		deps: &AuthServiceDeps{
			AdminRepository: mockAdminRepo,
			JwtService:      mockJwtService,
		},
	}

	start := time.Now()
	admin, tokens, err := service.AdminLogin("admin123", "password123")
	duration := time.Since(start)

	require.NoError(t, err)
	assert.NotNil(t, admin)
	assert.NotNil(t, tokens)
	assert.GreaterOrEqual(t, duration, 100*time.Millisecond)

	mockAdminRepo.AssertExpectations(t)
	mockJwtService.AssertExpectations(t)
}
