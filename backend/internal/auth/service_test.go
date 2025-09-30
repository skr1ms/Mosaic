package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/skr1ms/mosaic/config"
	"github.com/skr1ms/mosaic/internal/admin"
	"github.com/skr1ms/mosaic/internal/partner"
	"github.com/skr1ms/mosaic/pkg/bcrypt"
	"github.com/skr1ms/mosaic/pkg/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock Partner Repository
type MockPartnerRepository struct {
	mock.Mock
}

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

func (m *MockPartnerRepository) GetByEmail(ctx context.Context, email string) (*partner.Partner, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*partner.Partner), args.Error(1)
}

func (m *MockPartnerRepository) GetByID(ctx context.Context, id uuid.UUID) (*partner.Partner, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*partner.Partner), args.Error(1)
}

func (m *MockPartnerRepository) UpdatePassword(ctx context.Context, id uuid.UUID, hashedPassword string) error {
	args := m.Called(ctx, id, hashedPassword)
	return args.Error(0)
}

func (m *MockPartnerRepository) UpdateEmail(ctx context.Context, id uuid.UUID, email string) error {
	args := m.Called(ctx, id, email)
	return args.Error(0)
}

// Mock Admin Repository
type MockAdminRepository struct {
	mock.Mock
}

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

func (m *MockAdminRepository) GetByEmail(email string) (*admin.Admin, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*admin.Admin), args.Error(1)
}

func (m *MockAdminRepository) GetByID(id uuid.UUID) (*admin.Admin, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*admin.Admin), args.Error(1)
}

func (m *MockAdminRepository) UpdatePassword(id uuid.UUID, hashedPassword string) error {
	args := m.Called(id, hashedPassword)
	return args.Error(0)
}

func (m *MockAdminRepository) UpdateEmail(id uuid.UUID, email string) error {
	args := m.Called(id, email)
	return args.Error(0)
}

// Mock JWT Service
type MockJWTService struct {
	mock.Mock
}

func (m *MockJWTService) CreateTokenPair(userID uuid.UUID, login, role string) (*jwt.TokenPair, error) {
	args := m.Called(userID, login, role)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*jwt.TokenPair), args.Error(1)
}

func (m *MockJWTService) ValidateRefreshToken(refreshToken string) (*jwt.Claims, error) {
	args := m.Called(refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*jwt.Claims), args.Error(1)
}

func (m *MockJWTService) RefreshTokens(refreshToken string) (*jwt.TokenPair, error) {
	args := m.Called(refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*jwt.TokenPair), args.Error(1)
}

func (m *MockJWTService) CreatePasswordResetToken(userID uuid.UUID, email string) (string, error) {
	args := m.Called(userID, email)
	return args.String(0), args.Error(1)
}

func (m *MockJWTService) ValidatePasswordResetToken(token string) (*jwt.Claims, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*jwt.Claims), args.Error(1)
}

// Mock Recaptcha
type MockRecaptcha struct {
	mock.Mock
}

func (m *MockRecaptcha) Verify(token, expectedAction string) (bool, error) {
	args := m.Called(token, expectedAction)
	return args.Bool(0), args.Error(1)
}

// Mock MailSender
type MockMailSender struct {
	mock.Mock
}

func (m *MockMailSender) SendResetPasswordEmail(to, resetLink string) error {
	args := m.Called(to, resetLink)
	return args.Error(0)
}

// Mock Config
type MockConfig struct {
	mock.Mock
}

func (m *MockConfig) GetRecaptchaConfig() config.RecaptchaConfig {
	args := m.Called()
	return args.Get(0).(config.RecaptchaConfig)
}

func (m *MockConfig) GetServerConfig() config.ServerConfig {
	args := m.Called()
	return args.Get(0).(config.ServerConfig)
}

func TestAuthService_PartnerLogin(t *testing.T) {
	tests := []struct {
		name          string
		login         string
		password      string
		mockSetup     func(*MockPartnerRepository, *MockAdminRepository, *MockJWTService)
		expectedError bool
	}{
		{
			name:     "successful_partner_login",
			login:    "partner@example.com",
			password: "password123",
			mockSetup: func(partnerRepo *MockPartnerRepository, adminRepo *MockAdminRepository, jwtService *MockJWTService) {
				hashedPassword, _ := bcrypt.HashPassword("password123")
				partnerData := &partner.Partner{
					ID:        uuid.New(),
					Login:     "partner@example.com",
					Password:  hashedPassword,
					BrandName: "Test Partner",
				}
				partnerRepo.On("GetByLogin", mock.Anything, "partner@example.com").Return(partnerData, nil)
				partnerRepo.On("UpdateLastLogin", mock.Anything, partnerData.ID).Return(nil)

				tokenPair := &jwt.TokenPair{
					AccessToken:  "access_token",
					RefreshToken: "refresh_token",
				}
				jwtService.On("CreateTokenPair", partnerData.ID, partnerData.Login, "partner").Return(tokenPair, nil)
			},
			expectedError: false,
		},
		{
			name:     "partner_not_found",
			login:    "nonexistent@example.com",
			password: "password123",
			mockSetup: func(partnerRepo *MockPartnerRepository, adminRepo *MockAdminRepository, jwtService *MockJWTService) {
				partnerRepo.On("GetByLogin", mock.Anything, "nonexistent@example.com").Return(nil, errors.New("not found"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPartnerRepo := new(MockPartnerRepository)
			mockAdminRepo := new(MockAdminRepository)
			mockJWTService := new(MockJWTService)

			if tt.mockSetup != nil {
				tt.mockSetup(mockPartnerRepo, mockAdminRepo, mockJWTService)
			}

			deps := &AuthServiceDeps{
				PartnerRepository: mockPartnerRepo,
				AdminRepository:   mockAdminRepo,
				JwtService:        mockJWTService,
			}
			service := NewAuthService(deps)

			_, _, err := service.PartnerLogin(tt.login, tt.password)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockPartnerRepo.AssertExpectations(t)
			mockJWTService.AssertExpectations(t)
		})
	}
}

func TestAuthService_AdminLogin(t *testing.T) {
	tests := []struct {
		name          string
		login         string
		password      string
		mockSetup     func(*MockPartnerRepository, *MockAdminRepository, *MockJWTService)
		expectedError bool
	}{
		{
			name:     "successful_admin_login",
			login:    "admin@example.com",
			password: "adminpass",
			mockSetup: func(partnerRepo *MockPartnerRepository, adminRepo *MockAdminRepository, jwtService *MockJWTService) {
				hashedPassword, _ := bcrypt.HashPassword("adminpass")
				adminData := &admin.Admin{
					ID:       uuid.New(),
					Login:    "admin@example.com",
					Password: hashedPassword,
					Role:     "admin",
				}
				adminRepo.On("GetByLogin", "admin@example.com").Return(adminData, nil)
				adminRepo.On("UpdateLastLogin", adminData.ID).Return(nil)

				tokenPair := &jwt.TokenPair{
					AccessToken:  "admin_token",
					RefreshToken: "admin_refresh",
				}
				jwtService.On("CreateTokenPair", adminData.ID, adminData.Login, "admin").Return(tokenPair, nil)
			},
			expectedError: false,
		},
		{
			name:     "admin_not_found",
			login:    "nonexistent@admin.com",
			password: "password",
			mockSetup: func(partnerRepo *MockPartnerRepository, adminRepo *MockAdminRepository, jwtService *MockJWTService) {
				adminRepo.On("GetByLogin", "nonexistent@admin.com").Return(nil, errors.New("not found"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPartnerRepo := new(MockPartnerRepository)
			mockAdminRepo := new(MockAdminRepository)
			mockJWTService := new(MockJWTService)

			if tt.mockSetup != nil {
				tt.mockSetup(mockPartnerRepo, mockAdminRepo, mockJWTService)
			}

			deps := &AuthServiceDeps{
				PartnerRepository: mockPartnerRepo,
				AdminRepository:   mockAdminRepo,
				JwtService:        mockJWTService,
			}
			service := NewAuthService(deps)

			_, _, err := service.AdminLogin(tt.login, tt.password)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockAdminRepo.AssertExpectations(t)
			mockJWTService.AssertExpectations(t)
		})
	}
}

func TestAuthService_RefreshTokens(t *testing.T) {
	tests := []struct {
		name          string
		refreshToken  string
		mockSetup     func(*MockPartnerRepository, *MockAdminRepository, *MockJWTService)
		expectedError bool
	}{
		{
			name:         "successful_refresh",
			refreshToken: "valid_refresh_token",
			mockSetup: func(partnerRepo *MockPartnerRepository, adminRepo *MockAdminRepository, jwtService *MockJWTService) {
				claims := &jwt.Claims{
					UserID: uuid.New(),
					Role:   "partner",
				}
				jwtService.On("ValidateRefreshToken", "valid_refresh_token").Return(claims, nil)

				tokenPair := &jwt.TokenPair{
					AccessToken:  "new_access_token",
					RefreshToken: "new_refresh_token",
				}
				jwtService.On("RefreshTokens", "valid_refresh_token").Return(tokenPair, nil)
			},
			expectedError: false,
		},
		{
			name:         "invalid_refresh_token",
			refreshToken: "invalid_token",
			mockSetup: func(partnerRepo *MockPartnerRepository, adminRepo *MockAdminRepository, jwtService *MockJWTService) {
				jwtService.On("ValidateRefreshToken", "invalid_token").Return(nil, errors.New("invalid token"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPartnerRepo := new(MockPartnerRepository)
			mockAdminRepo := new(MockAdminRepository)
			mockJWTService := new(MockJWTService)

			if tt.mockSetup != nil {
				tt.mockSetup(mockPartnerRepo, mockAdminRepo, mockJWTService)
			}

			deps := &AuthServiceDeps{
				PartnerRepository: mockPartnerRepo,
				AdminRepository:   mockAdminRepo,
				JwtService:        mockJWTService,
			}
			service := NewAuthService(deps)

			result, err := service.RefreshTokens(tt.refreshToken)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockJWTService.AssertExpectations(t)
		})
	}
}

func TestAuthService_ForgotPassword(t *testing.T) {
	tests := []struct {
		name          string
		login         string
		email         string
		captcha       string
		mockSetup     func(*MockPartnerRepository, *MockAdminRepository, *MockJWTService, *MockRecaptcha, *MockMailSender, *MockConfig)
		expectedError bool
	}{
		{
			name:    "successful_partner_forgot_password",
			login:   "partner@example.com",
			email:   "partner@example.com",
			captcha: "valid_captcha",
			mockSetup: func(partnerRepo *MockPartnerRepository, adminRepo *MockAdminRepository, jwtService *MockJWTService, recaptcha *MockRecaptcha, mailSender *MockMailSender, mockConfig *MockConfig) {
				recaptcha.On("Verify", "valid_captcha", "forgot_password").Return(true, nil)

				recaptchaConfig := config.RecaptchaConfig{Environment: "test"}
				serverConfig := config.ServerConfig{FrontendURL: "http://localhost:3000"}
				mockConfig.On("GetRecaptchaConfig").Return(recaptchaConfig)
				mockConfig.On("GetServerConfig").Return(serverConfig)

				adminRepo.On("GetByLogin", "partner@example.com").Return(nil, errors.New("admin not found"))

				partnerData := &partner.Partner{
					ID:     uuid.New(),
					Login:  "partner@example.com",
					Email:  "partner@example.com",
					Status: "active",
				}
				partnerRepo.On("GetByLogin", mock.Anything, "partner@example.com").Return(partnerData, nil)

				jwtService.On("CreatePasswordResetToken", partnerData.ID, "partner@example.com").Return("reset_token", nil)

				mailSender.On("SendResetPasswordEmail", "partner@example.com", mock.AnythingOfType("string")).Return(nil)
			},
			expectedError: false,
		},
		{
			name:    "invalid_captcha",
			login:   "partner@example.com",
			email:   "partner@example.com",
			captcha: "invalid_captcha",
			mockSetup: func(partnerRepo *MockPartnerRepository, adminRepo *MockAdminRepository, jwtService *MockJWTService, recaptcha *MockRecaptcha, mailSender *MockMailSender, mockConfig *MockConfig) {
				recaptcha.On("Verify", "invalid_captcha", "forgot_password").Return(false, nil)
				recaptchaConfig := config.RecaptchaConfig{Environment: "test"}
				mockConfig.On("GetRecaptchaConfig").Return(recaptchaConfig)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPartnerRepo := new(MockPartnerRepository)
			mockAdminRepo := new(MockAdminRepository)
			mockJWTService := new(MockJWTService)
			mockRecaptcha := new(MockRecaptcha)
			mockMailSender := new(MockMailSender)
			mockConfig := new(MockConfig)

			if tt.mockSetup != nil {
				tt.mockSetup(mockPartnerRepo, mockAdminRepo, mockJWTService, mockRecaptcha, mockMailSender, mockConfig)
			}

			deps := &AuthServiceDeps{
				PartnerRepository: mockPartnerRepo,
				AdminRepository:   mockAdminRepo,
				JwtService:        mockJWTService,
				Recaptcha:         mockRecaptcha,
				MailSender:        mockMailSender,
				Config:            mockConfig,
			}
			service := NewAuthService(deps)

			err := service.ForgotPassword(context.Background(), tt.login, tt.email, tt.captcha)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockPartnerRepo.AssertExpectations(t)
			mockAdminRepo.AssertExpectations(t)
			mockJWTService.AssertExpectations(t)
			mockRecaptcha.AssertExpectations(t)
			mockMailSender.AssertExpectations(t)
			mockConfig.AssertExpectations(t)
		})
	}
}

func TestAuthService_ResetPassword(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name          string
		token         string
		newPassword   string
		mockSetup     func(*MockPartnerRepository, *MockAdminRepository, *MockJWTService)
		expectedError bool
	}{
		{
			name:        "successful_partner_reset_password",
			token:       "valid_reset_token",
			newPassword: "newpassword123",
			mockSetup: func(partnerRepo *MockPartnerRepository, adminRepo *MockAdminRepository, jwtService *MockJWTService) {
				claims := &jwt.Claims{
					UserID: userID,
					Role:   "partner",
				}
				jwtService.On("ValidatePasswordResetToken", "valid_reset_token").Return(claims, nil)

				adminRepo.On("GetByID", userID).Return(nil, errors.New("admin not found"))

				partnerData := &partner.Partner{
					ID:     userID,
					Login:  "partner@example.com",
					Status: "active",
				}
				partnerRepo.On("GetByID", mock.Anything, userID).Return(partnerData, nil)
				partnerRepo.On("UpdatePassword", mock.Anything, userID, mock.AnythingOfType("string")).Return(nil)
			},
			expectedError: false,
		},
		{
			name:        "invalid_token",
			token:       "invalid_token",
			newPassword: "newpassword123",
			mockSetup: func(partnerRepo *MockPartnerRepository, adminRepo *MockAdminRepository, jwtService *MockJWTService) {
				jwtService.On("ValidatePasswordResetToken", "invalid_token").Return(nil, errors.New("invalid token"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPartnerRepo := new(MockPartnerRepository)
			mockAdminRepo := new(MockAdminRepository)
			mockJWTService := new(MockJWTService)

			if tt.mockSetup != nil {
				tt.mockSetup(mockPartnerRepo, mockAdminRepo, mockJWTService)
			}

			deps := &AuthServiceDeps{
				PartnerRepository: mockPartnerRepo,
				AdminRepository:   mockAdminRepo,
				JwtService:        mockJWTService,
			}
			service := NewAuthService(deps)

			err := service.ResetPassword(context.Background(), tt.token, tt.newPassword)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockPartnerRepo.AssertExpectations(t)
			mockAdminRepo.AssertExpectations(t)
			mockJWTService.AssertExpectations(t)
		})
	}
}

func TestAuthService_ChangePassword(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name            string
		userID          uuid.UUID
		userRole        string
		currentPassword string
		newPassword     string
		mockSetup       func(*MockPartnerRepository, *MockAdminRepository)
		expectedError   bool
	}{
		{
			name:            "successful_partner_password_change",
			userID:          userID,
			userRole:        "partner",
			currentPassword: "currentpass123",
			newPassword:     "newpass123",
			mockSetup: func(partnerRepo *MockPartnerRepository, adminRepo *MockAdminRepository) {
				hashedPassword, _ := bcrypt.HashPassword("currentpass123")
				partnerData := &partner.Partner{
					ID:       userID,
					Password: hashedPassword,
				}
				partnerRepo.On("GetByID", mock.Anything, userID).Return(partnerData, nil)
				partnerRepo.On("UpdatePassword", mock.Anything, userID, mock.AnythingOfType("string")).Return(nil)
			},
			expectedError: false,
		},
		{
			name:            "user_not_found",
			userID:          userID,
			userRole:        "partner",
			currentPassword: "currentpass123",
			newPassword:     "newpass123",
			mockSetup: func(partnerRepo *MockPartnerRepository, adminRepo *MockAdminRepository) {
				partnerRepo.On("GetByID", mock.Anything, userID).Return(nil, errors.New("not found"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPartnerRepo := new(MockPartnerRepository)
			mockAdminRepo := new(MockAdminRepository)

			if tt.mockSetup != nil {
				tt.mockSetup(mockPartnerRepo, mockAdminRepo)
			}

			deps := &AuthServiceDeps{
				PartnerRepository: mockPartnerRepo,
				AdminRepository:   mockAdminRepo,
			}
			service := NewAuthService(deps)

			err := service.ChangePassword(tt.userID, tt.userRole, tt.currentPassword, tt.newPassword)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockPartnerRepo.AssertExpectations(t)
			mockAdminRepo.AssertExpectations(t)
		})
	}
}

func TestAuthService_ChangeAdminEmail(t *testing.T) {
	adminID := uuid.New()

	tests := []struct {
		name            string
		adminID         uuid.UUID
		currentPassword string
		newEmail        string
		mockSetup       func(*MockAdminRepository)
		expectedError   bool
	}{
		{
			name:            "successful_email_change",
			adminID:         adminID,
			currentPassword: "currentpass123",
			newEmail:        "newemail@example.com",
			mockSetup: func(adminRepo *MockAdminRepository) {
				hashedPassword, _ := bcrypt.HashPassword("currentpass123")
				adminData := &admin.Admin{
					ID:       adminID,
					Password: hashedPassword,
					Email:    "oldemail@example.com",
				}
				adminRepo.On("GetByID", adminID).Return(adminData, nil)
				adminRepo.On("GetByEmail", "newemail@example.com").Return(nil, errors.New("not found"))
				adminRepo.On("UpdateEmail", adminID, "newemail@example.com").Return(nil)
			},
			expectedError: false,
		},
		{
			name:            "admin_not_found",
			adminID:         adminID,
			currentPassword: "currentpass123",
			newEmail:        "newemail@example.com",
			mockSetup: func(adminRepo *MockAdminRepository) {
				adminRepo.On("GetByID", adminID).Return(nil, errors.New("not found"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAdminRepo := new(MockAdminRepository)

			if tt.mockSetup != nil {
				tt.mockSetup(mockAdminRepo)
			}

			deps := &AuthServiceDeps{
				AdminRepository: mockAdminRepo,
			}
			service := NewAuthService(deps)

			err := service.ChangeAdminEmail(tt.adminID, tt.currentPassword, tt.newEmail)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockAdminRepo.AssertExpectations(t)
		})
	}
}
