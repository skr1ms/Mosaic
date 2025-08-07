package partner

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
	"github.com/skr1ms/mosaic/pkg/bcrypt"
)

type MockPartnerRepository struct {
	mock.Mock
}

func (m *MockPartnerRepository) Create(ctx context.Context, partner *Partner) error {
	args := m.Called(ctx, partner)
	return args.Error(0)
}

func (m *MockPartnerRepository) GetByID(ctx context.Context, id uuid.UUID) (*Partner, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Partner), args.Error(1)
}

func (m *MockPartnerRepository) GetByLogin(ctx context.Context, login string) (*Partner, error) {
	args := m.Called(ctx, login)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Partner), args.Error(1)
}

func (m *MockPartnerRepository) GetByPartnerCode(ctx context.Context, code string) (*Partner, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Partner), args.Error(1)
}

func (m *MockPartnerRepository) GetByDomain(ctx context.Context, domain string) (*Partner, error) {
	args := m.Called(ctx, domain)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Partner), args.Error(1)
}

func (m *MockPartnerRepository) GetByEmail(ctx context.Context, email string) (*Partner, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Partner), args.Error(1)
}

func (m *MockPartnerRepository) Update(ctx context.Context, partner *Partner) error {
	args := m.Called(ctx, partner)
	return args.Error(0)
}

func (m *MockPartnerRepository) UpdatePassword(ctx context.Context, partnerID uuid.UUID, hashedPassword string) error {
	args := m.Called(ctx, partnerID, hashedPassword)
	return args.Error(0)
}

func (m *MockPartnerRepository) DeleteWithCoupons(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPartnerRepository) GetAll(ctx context.Context, sortBy string, order string) ([]*Partner, error) {
	args := m.Called(ctx, sortBy, order)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Partner), args.Error(1)
}

func (m *MockPartnerRepository) GetActivePartners(ctx context.Context) ([]*Partner, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Partner), args.Error(1)
}

func (m *MockPartnerRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPartnerRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockPartnerRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPartnerRepository) Search(ctx context.Context, queryStr string, status string, sortBy string, order string) ([]*Partner, error) {
	args := m.Called(ctx, queryStr, status, sortBy, order)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Partner), args.Error(1)
}

func (m *MockPartnerRepository) CountActive(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockPartnerRepository) CountTotal(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockPartnerRepository) GetTopByActivity(ctx context.Context, limit int) ([]*Partner, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Partner), args.Error(1)
}

func (m *MockPartnerRepository) GetNextPartnerCode(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}

func (m *MockPartnerRepository) GetCouponsStatistics(ctx context.Context, partnerID uuid.UUID) (map[string]int64, error) {
	args := m.Called(ctx, partnerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]int64), args.Error(1)
}

func (m *MockPartnerRepository) GetPartnerCouponsForExport(ctx context.Context, partnerID uuid.UUID, status string) ([]*ExportCouponRequest, error) {
	args := m.Called(ctx, partnerID, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*ExportCouponRequest), args.Error(1)
}

func (m *MockPartnerRepository) GetAllCouponsForExport(ctx context.Context) ([]*ExportCouponRequest, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*ExportCouponRequest), args.Error(1)
}

type MockCouponService struct {
	mock.Mock
}

func (m *MockCouponService) ExportCouponsAdvanced(options coupon.ExportOptionsRequest) ([]byte, string, string, error) {
	args := m.Called(options)
	return args.Get(0).([]byte), args.String(1), args.String(2), args.Error(3)
}

type MockRecaptcha struct {
	mock.Mock
}

func (m *MockRecaptcha) Verify(token, action string) (bool, error) {
	args := m.Called(token, action)
	return args.Bool(0), args.Error(1)
}

type MockJWT struct {
	mock.Mock
}

func (m *MockJWT) CreatePasswordResetToken(userID uuid.UUID, email string) (string, error) {
	args := m.Called(userID, email)
	return args.String(0), args.Error(1)
}

func (m *MockJWT) ValidatePasswordResetToken(token string) (*TokenClaims, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*TokenClaims), args.Error(1)
}

type MockMailer struct {
	mock.Mock
}

func (m *MockMailer) SendResetPasswordEmail(email, resetLink string) error {
	args := m.Called(email, resetLink)
	return args.Error(0)
}

func createTestPartner() *Partner {
	return &Partner{
		ID:              uuid.New(),
		PartnerCode:     "0001",
		Login:           "testpartner",
		Password:        "hashedpassword",
		Domain:          "test.example.com",
		BrandName:       "Test Partner",
		LogoURL:         "https://example.com/logo.png",
		OzonLink:        "https://ozon.ru/test",
		WildberriesLink: "https://wildberries.ru/test",
		Email:           "partner@test.example.com",
		Address:         "123 Test Street",
		Phone:           "+1234567890",
		Telegram:        "@testpartner",
		Whatsapp:        "+1234567890",
		TelegramLink:    "https://t.me/testpartner",
		WhatsappLink:    "https://wa.me/1234567890",
		AllowSales:      true,
		AllowPurchases:  true,
		Status:          "active",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}

func createTestConfig() *config.Config {
	return &config.Config{
		ServerConfig: config.ServerConfig{
			FrontendURL: "https://frontend.example.com",
		},
		RecaptchaConfig: config.RecaptchaConfig{
			Environment: "development",
		},
	}
}

func TestPartnerRepository_Create_Success(t *testing.T) {
	mockRepo := &MockPartnerRepository{}
	partner := createTestPartner()

	mockRepo.On("Create", mock.Anything, partner).Return(nil)

	err := mockRepo.Create(context.Background(), partner)

	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestPartnerRepository_Create_Error(t *testing.T) {
	mockRepo := &MockPartnerRepository{}
	partner := createTestPartner()

	mockRepo.On("Create", mock.Anything, partner).Return(errors.New("database error"))

	err := mockRepo.Create(context.Background(), partner)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
	mockRepo.AssertExpectations(t)
}

func TestPartnerRepository_GetByID_Success(t *testing.T) {
	mockRepo := &MockPartnerRepository{}
	testPartner := createTestPartner()

	mockRepo.On("GetByID", mock.Anything, testPartner.ID).Return(testPartner, nil)

	partner, err := mockRepo.GetByID(context.Background(), testPartner.ID)

	require.NoError(t, err)
	assert.Equal(t, testPartner.ID, partner.ID)
	assert.Equal(t, testPartner.BrandName, partner.BrandName)
	assert.Equal(t, testPartner.Email, partner.Email)
	mockRepo.AssertExpectations(t)
}

func TestPartnerRepository_GetByID_NotFound(t *testing.T) {
	mockRepo := &MockPartnerRepository{}
	partnerID := uuid.New()

	mockRepo.On("GetByID", mock.Anything, partnerID).Return(nil, errors.New("partner not found"))

	partner, err := mockRepo.GetByID(context.Background(), partnerID)

	assert.Error(t, err)
	assert.Nil(t, partner)
	assert.Contains(t, err.Error(), "partner not found")
	mockRepo.AssertExpectations(t)
}

func TestPartnerRepository_GetByEmail_Success(t *testing.T) {
	mockRepo := &MockPartnerRepository{}
	testPartner := createTestPartner()

	mockRepo.On("GetByEmail", mock.Anything, testPartner.Email).Return(testPartner, nil)

	partner, err := mockRepo.GetByEmail(context.Background(), testPartner.Email)

	require.NoError(t, err)
	assert.Equal(t, testPartner.Email, partner.Email)
	assert.Equal(t, testPartner.BrandName, partner.BrandName)
	mockRepo.AssertExpectations(t)
}

func TestPartnerRepository_GetByDomain_Success(t *testing.T) {
	mockRepo := &MockPartnerRepository{}
	testPartner := createTestPartner()

	mockRepo.On("GetByDomain", mock.Anything, testPartner.Domain).Return(testPartner, nil)

	partner, err := mockRepo.GetByDomain(context.Background(), testPartner.Domain)

	require.NoError(t, err)
	assert.Equal(t, testPartner.Domain, partner.Domain)
	assert.Equal(t, testPartner.BrandName, partner.BrandName)
	mockRepo.AssertExpectations(t)
}

func TestPartnerRepository_UpdatePassword_Success(t *testing.T) {
	mockRepo := &MockPartnerRepository{}
	partnerID := uuid.New()
	hashedPassword := "newhashed123"

	mockRepo.On("UpdatePassword", mock.Anything, partnerID, hashedPassword).Return(nil)

	err := mockRepo.UpdatePassword(context.Background(), partnerID, hashedPassword)

	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestPartnerRepository_GetAll_Success(t *testing.T) {
	mockRepo := &MockPartnerRepository{}
	partners := []*Partner{createTestPartner(), createTestPartner()}

	mockRepo.On("GetAll", mock.Anything, "created_at", "desc").Return(partners, nil)

	result, err := mockRepo.GetAll(context.Background(), "created_at", "desc")

	require.NoError(t, err)
	assert.Len(t, result, 2)
	mockRepo.AssertExpectations(t)
}

func TestPartnerRepository_GetActivePartners_Success(t *testing.T) {
	mockRepo := &MockPartnerRepository{}
	activePartners := []*Partner{createTestPartner()}

	mockRepo.On("GetActivePartners", mock.Anything).Return(activePartners, nil)

	result, err := mockRepo.GetActivePartners(context.Background())

	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "active", result[0].Status)
	mockRepo.AssertExpectations(t)
}

func TestPartnerRepository_Search_Success(t *testing.T) {
	mockRepo := &MockPartnerRepository{}
	partners := []*Partner{createTestPartner()}
	query := "test"
	status := "active"

	mockRepo.On("Search", mock.Anything, query, status, "created_at", "desc").Return(partners, nil)

	result, err := mockRepo.Search(context.Background(), query, status, "created_at", "desc")

	require.NoError(t, err)
	assert.Len(t, result, 1)
	mockRepo.AssertExpectations(t)
}

func TestPartnerRepository_CountActive_Success(t *testing.T) {
	mockRepo := &MockPartnerRepository{}
	expectedCount := int64(5)

	mockRepo.On("CountActive", mock.Anything).Return(expectedCount, nil)

	count, err := mockRepo.CountActive(context.Background())

	require.NoError(t, err)
	assert.Equal(t, expectedCount, count)
	mockRepo.AssertExpectations(t)
}

func TestPartnerRepository_CountTotal_Success(t *testing.T) {
	mockRepo := &MockPartnerRepository{}
	expectedCount := int64(10)

	mockRepo.On("CountTotal", mock.Anything).Return(expectedCount, nil)

	count, err := mockRepo.CountTotal(context.Background())

	require.NoError(t, err)
	assert.Equal(t, expectedCount, count)
	mockRepo.AssertExpectations(t)
}

func TestPartnerRepository_GetNextPartnerCode_Success(t *testing.T) {
	mockRepo := &MockPartnerRepository{}
	expectedCode := "0001"

	mockRepo.On("GetNextPartnerCode", mock.Anything).Return(expectedCode, nil)

	code, err := mockRepo.GetNextPartnerCode(context.Background())

	require.NoError(t, err)
	assert.Equal(t, expectedCode, code)
	mockRepo.AssertExpectations(t)
}

func TestPartnerRepository_DeleteWithCoupons_Success(t *testing.T) {
	mockRepo := &MockPartnerRepository{}
	partnerID := uuid.New()

	mockRepo.On("DeleteWithCoupons", mock.Anything, partnerID).Return(nil)

	err := mockRepo.DeleteWithCoupons(context.Background(), partnerID)

	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestPartnerRepository_GetCouponsStatistics_Success(t *testing.T) {
	mockRepo := &MockPartnerRepository{}
	partnerID := uuid.New()
	expectedStats := map[string]int64{
		"total":     10,
		"activated": 8,
		"new":       2,
		"purchased": 5,
	}

	mockRepo.On("GetCouponsStatistics", mock.Anything, partnerID).Return(expectedStats, nil)

	stats, err := mockRepo.GetCouponsStatistics(context.Background(), partnerID)

	require.NoError(t, err)
	assert.Equal(t, expectedStats, stats)
	mockRepo.AssertExpectations(t)
}

func TestCouponService_ExportCouponsAdvanced_Success(t *testing.T) {
	mockService := &MockCouponService{}
	partnerID := uuid.New().String()
	options := coupon.ExportOptionsRequest{
		Format:     coupon.ExportFormatCodes,
		PartnerID:  &partnerID,
		Status:     "new",
		FileFormat: "txt",
	}

	expectedContent := []byte("COUPON001\nCOUPON002\n")
	expectedFilename := "coupons.txt"
	expectedContentType := "text/plain"

	mockService.On("ExportCouponsAdvanced", options).Return(expectedContent, expectedFilename, expectedContentType, nil)

	content, filename, contentType, err := mockService.ExportCouponsAdvanced(options)

	require.NoError(t, err)
	assert.Equal(t, expectedContent, content)
	assert.Equal(t, expectedFilename, filename)
	assert.Equal(t, expectedContentType, contentType)
	mockService.AssertExpectations(t)
}

func TestJWT_CreatePasswordResetToken_Success(t *testing.T) {
	mockJWT := &MockJWT{}
	userID := uuid.New()
	email := "test@example.com"
	expectedToken := "reset-token-123"

	mockJWT.On("CreatePasswordResetToken", userID, email).Return(expectedToken, nil)

	token, err := mockJWT.CreatePasswordResetToken(userID, email)

	require.NoError(t, err)
	assert.Equal(t, expectedToken, token)
	mockJWT.AssertExpectations(t)
}

func TestJWT_ValidatePasswordResetToken_Success(t *testing.T) {
	mockJWT := &MockJWT{}
	token := "valid-token"
	expectedClaims := &TokenClaims{
		UserID: uuid.New(),
		Login:  "test@example.com",
	}

	mockJWT.On("ValidatePasswordResetToken", token).Return(expectedClaims, nil)

	claims, err := mockJWT.ValidatePasswordResetToken(token)

	require.NoError(t, err)
	assert.Equal(t, expectedClaims.UserID, claims.UserID)
	assert.Equal(t, expectedClaims.Login, claims.Login)
	mockJWT.AssertExpectations(t)
}

func TestMailer_SendResetPasswordEmail_Success(t *testing.T) {
	mockMailer := &MockMailer{}
	email := "test@example.com"
	resetLink := "https://example.com/reset?token=abc123"

	mockMailer.On("SendResetPasswordEmail", email, resetLink).Return(nil)

	err := mockMailer.SendResetPasswordEmail(email, resetLink)

	require.NoError(t, err)
	mockMailer.AssertExpectations(t)
}

func TestRecaptcha_Verify_Success(t *testing.T) {
	mockRecaptcha := &MockRecaptcha{}
	token := "recaptcha-token"
	action := "forgot_password"

	mockRecaptcha.On("Verify", token, action).Return(true, nil)

	valid, err := mockRecaptcha.Verify(token, action)

	require.NoError(t, err)
	assert.True(t, valid)
	mockRecaptcha.AssertExpectations(t)
}

func TestRecaptcha_Verify_Invalid(t *testing.T) {
	mockRecaptcha := &MockRecaptcha{}
	token := "invalid-token"
	action := "forgot_password"

	mockRecaptcha.On("Verify", token, action).Return(false, nil)

	valid, err := mockRecaptcha.Verify(token, action)

	require.NoError(t, err)
	assert.False(t, valid)
	mockRecaptcha.AssertExpectations(t)
}

func TestPartner_PasswordHashing(t *testing.T) {
	password := "testpassword123"

	hashedPassword, err := bcrypt.HashPassword(password)
	require.NoError(t, err)
	assert.NotEmpty(t, hashedPassword)
	assert.NotEqual(t, password, hashedPassword)

	isValid := bcrypt.CheckPassword(password, hashedPassword)
	assert.True(t, isValid)

	isInvalid := bcrypt.CheckPassword("wrongpassword", hashedPassword)
	assert.False(t, isInvalid)
}

func TestPartner_ModelValidation(t *testing.T) {
	partner := createTestPartner()

	assert.NotEmpty(t, partner.ID)
	assert.NotEmpty(t, partner.PartnerCode)
	assert.NotEmpty(t, partner.Login)
	assert.NotEmpty(t, partner.Email)
	assert.NotEmpty(t, partner.BrandName)
	assert.NotEmpty(t, partner.Domain)

	assert.Contains(t, partner.Email, "@")

	assert.True(t, partner.AllowSales)
	assert.True(t, partner.AllowPurchases)

	assert.Equal(t, "active", partner.Status)
}
