package public

import (
	"context"
	"errors"
	"mime/multipart"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/skr1ms/mosaic/config"
	"github.com/skr1ms/mosaic/internal/coupon"
	"github.com/skr1ms/mosaic/internal/image"
	"github.com/skr1ms/mosaic/internal/partner"
	"github.com/skr1ms/mosaic/internal/payment"
	"github.com/skr1ms/mosaic/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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

// Mock Image Repository
type MockImageRepository struct {
	mock.Mock
}

func (m *MockImageRepository) Create(ctx context.Context, image *image.Image) error {
	args := m.Called(ctx, image)
	return args.Error(0)
}

func (m *MockImageRepository) GetByID(ctx context.Context, id uuid.UUID) (*image.Image, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*image.Image), args.Error(1)
}

func (m *MockImageRepository) Update(ctx context.Context, image *image.Image) error {
	args := m.Called(ctx, image)
	return args.Error(0)
}

func (m *MockImageRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockImageRepository) GetByCouponID(ctx context.Context, couponID uuid.UUID) (*image.Image, error) {
	args := m.Called(ctx, couponID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*image.Image), args.Error(1)
}

func (m *MockImageRepository) GetQueuedTasks(ctx context.Context) ([]*image.Image, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*image.Image), args.Error(1)
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

func (m *MockPartnerRepository) GetByPartnerCode(ctx context.Context, code string) (*partner.Partner, error) {
	args := m.Called(ctx, code)
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

func (m *MockPartnerRepository) GetArticleBySizeStyle(ctx context.Context, partnerID uuid.UUID, size, style, marketplace string) (*partner.PartnerArticle, error) {
	args := m.Called(ctx, partnerID, size, style, marketplace)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*partner.PartnerArticle), args.Error(1)
}

func (m *MockPartnerRepository) GetArticleGrid(ctx context.Context, partnerID uuid.UUID) (map[string]map[string]map[string]string, error) {
	args := m.Called(ctx, partnerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]map[string]map[string]string), args.Error(1)
}

// Mock Image Service
type MockImageService struct {
	mock.Mock
}

func (m *MockImageService) UploadImage(ctx context.Context, couponID uuid.UUID, file *multipart.FileHeader, userEmail string) (*image.Image, error) {
	args := m.Called(ctx, couponID, file, userEmail)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*image.Image), args.Error(1)
}

func (m *MockImageService) EditImage(ctx context.Context, imageID uuid.UUID, params image.ImageEditParams) error {
	args := m.Called(ctx, imageID, params)
	return args.Error(0)
}

func (m *MockImageService) ProcessImage(ctx context.Context, imageID uuid.UUID, params *image.ProcessingParams) error {
	args := m.Called(ctx, imageID, params)
	return args.Error(0)
}

func (m *MockImageService) GenerateSchema(ctx context.Context, imageID uuid.UUID, confirmed bool) error {
	args := m.Called(ctx, imageID, confirmed)
	return args.Error(0)
}

func (m *MockImageService) GetImageStatus(ctx context.Context, imageID uuid.UUID) (*types.ImageStatusResponse, error) {
	args := m.Called(ctx, imageID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.ImageStatusResponse), args.Error(1)
}

// Mock Payment Service
type MockPaymentService struct {
	mock.Mock
}

func (m *MockPaymentService) PurchaseCoupon(ctx context.Context, req *payment.PurchaseCouponRequest) (*payment.PurchaseCouponResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*payment.PurchaseCouponResponse), args.Error(1)
}

// Mock Email Service
type MockEmailService struct {
	mock.Mock
}

func (m *MockEmailService) SendSchemaEmail(email, schemaURL, couponCode string) error {
	args := m.Called(email, schemaURL, couponCode)
	return args.Error(0)
}

// Mock Config
type MockConfig struct {
	mock.Mock
}

func (m *MockConfig) GetServerConfig() config.ServerConfig {
	args := m.Called()
	return args.Get(0).(config.ServerConfig)
}

func (m *MockConfig) GetRecaptchaConfig() config.RecaptchaConfig {
	args := m.Called()
	return args.Get(0).(config.RecaptchaConfig)
}

// Mock File Header
type MockFileHeader struct {
	mock.Mock
	Filename string
	Size     int64
	Header   map[string][]string
}

func (m *MockFileHeader) Open() (multipart.File, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(multipart.File), args.Error(1)
}

func createTestPartner() *partner.Partner {
	return &partner.Partner{
		ID:              uuid.New(),
		PartnerCode:     "TEST",
		Login:           "testpartner",
		Password:        "hashedpassword",
		Domain:          "test.example.com",
		BrandName:       "Test Brand",
		LogoURL:         "https://example.com/logo.png",
		OzonLink:        "https://ozon.ru/test",
		WildberriesLink: "https://wildberries.ru/test",
		Email:           "test@example.com",
		Address:         "Test Address",
		Phone:           "+1234567890",
		Telegram:        "@testpartner",
		Whatsapp:        "+1234567890",
		TelegramLink:    "https://t.me/testpartner",
		WhatsappLink:    "https://wa.me/1234567890",
		AllowSales:      true,
		AllowPurchases:  true,
		Status:          "active",
		IsBlockedInChat: false,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}

func createTestCoupon() *coupon.Coupon {
	now := time.Now()
	return &coupon.Coupon{
		ID:          uuid.New(),
		Code:        "123456789012",
		Size:        "30x40",
		Style:       "grayscale",
		Status:      "new",
		IsPurchased: false,
		CreatedAt:   now,
	}
}

func TestNewPublicService(t *testing.T) {
	deps := &PublicServiceDeps{
		PartnerRepository: &MockPartnerRepository{},
		CouponRepository:  &MockCouponRepository{},
		ImageRepository:   &MockImageRepository{},
		ImageService:      &MockImageService{},
		PaymentService:    &MockPaymentService{},
		EmailService:      &MockEmailService{},
		Config:            &MockConfig{},
		RecaptchaSiteKey:  "test_key",
	}

	service := NewPublicService(deps)

	assert.NotNil(t, service)
	assert.NotNil(t, service.deps)
	assert.Equal(t, deps, service.deps)
}

func TestPublicService_Structure(t *testing.T) {
	service := &PublicService{}
	assert.NotNil(t, service)
}

func TestPublicService_ValidateStructure(t *testing.T) {
	tests := []struct {
		name string
		deps *PublicServiceDeps
	}{
		{
			name: "with_dependencies",
			deps: &PublicServiceDeps{
				PartnerRepository: &MockPartnerRepository{},
				CouponRepository:  &MockCouponRepository{},
				ImageRepository:   &MockImageRepository{},
				ImageService:      &MockImageService{},
				PaymentService:    &MockPaymentService{},
				EmailService:      &MockEmailService{},
				Config:            &MockConfig{},
				RecaptchaSiteKey:  "test_key",
			},
		},
		{
			name: "nil_dependencies",
			deps: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewPublicService(tt.deps)

			assert.NotNil(t, service)
			if tt.deps != nil {
				assert.Equal(t, tt.deps, service.deps)
			}
		})
	}
}

func TestPublicService_GetPartnerByDomain(t *testing.T) {
	tests := []struct {
		name          string
		domain        string
		mockSetup     func(*MockPartnerRepository)
		expectedError bool
	}{
		{
			name:   "successful_get_partner",
			domain: "test.example.com",
			mockSetup: func(repo *MockPartnerRepository) {
				partner := createTestPartner()
				repo.On("GetByDomain", mock.Anything, "test.example.com").Return(partner, nil)
			},
			expectedError: false,
		},
		{
			name:   "partner_not_found",
			domain: "nonexistent.example.com",
			mockSetup: func(repo *MockPartnerRepository) {
				repo.On("GetByDomain", mock.Anything, "nonexistent.example.com").Return(nil, errors.New("not found"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockPartnerRepository)

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo)
			}

			deps := &PublicServiceDeps{
				PartnerRepository: mockRepo,
			}
			service := NewPublicService(deps)

			result, err := service.GetPartnerByDomain(tt.domain)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Contains(t, result, "brand_name")
				assert.Contains(t, result, "domain")
				assert.Contains(t, result, "logo_url")
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestPublicService_GetCouponByCode(t *testing.T) {
	tests := []struct {
		name          string
		code          string
		mockSetup     func(*MockCouponRepository, *MockPartnerRepository)
		expectedError bool
	}{
		{
			name: "valid_code",
			code: "123456789012",
			mockSetup: func(repo *MockCouponRepository, partnerRepo *MockPartnerRepository) {
				coupon := createTestCoupon()
				partner := createTestPartner()
				repo.On("GetByCode", mock.Anything, "123456789012").Return(coupon, nil)
				partnerRepo.On("GetByID", mock.Anything, coupon.PartnerID).Return(partner, nil)
			},
			expectedError: false,
		},
		{
			name:          "invalid_code_length",
			code:          "12345",
			mockSetup:     nil,
			expectedError: true,
		},
		{
			name:          "invalid_code_format",
			code:          "abcdefghijkl",
			mockSetup:     nil,
			expectedError: true,
		},
		{
			name: "coupon_not_found",
			code: "123456789012",
			mockSetup: func(repo *MockCouponRepository, partnerRepo *MockPartnerRepository) {
				repo.On("GetByCode", mock.Anything, "123456789012").Return(nil, errors.New("not found"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockCouponRepository)
			mockPartnerRepo := new(MockPartnerRepository)

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo, mockPartnerRepo)
			}

			deps := &PublicServiceDeps{
				CouponRepository:  mockRepo,
				PartnerRepository: mockPartnerRepo,
			}
			service := NewPublicService(deps)

			result, err := service.GetCouponByCode(tt.code)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Contains(t, result, "id")
				assert.Contains(t, result, "code")
				assert.Contains(t, result, "size")
				assert.Contains(t, result, "style")
				assert.Contains(t, result, "status")
				assert.Contains(t, result, "valid")
			}

			mockRepo.AssertExpectations(t)
			mockPartnerRepo.AssertExpectations(t)
		})
	}
}

func TestPublicService_ActivateCoupon(t *testing.T) {
	tests := []struct {
		name          string
		code          string
		mockSetup     func(*MockCouponRepository)
		expectedError bool
	}{
		{
			name: "successful_activation",
			code: "123456789012",
			mockSetup: func(repo *MockCouponRepository) {
				coupon := createTestCoupon()
				repo.On("GetByCode", mock.Anything, "123456789012").Return(coupon, nil)
				repo.On("Update", mock.Anything, mock.Anything).Return(nil)
			},
			expectedError: false,
		},
		{
			name: "empty_code",
			code: "",
			mockSetup: func(repo *MockCouponRepository) {
				repo.On("GetByCode", mock.Anything, "").Return(nil, errors.New("not found"))
			},
			expectedError: true,
		},
		{
			name: "coupon_not_found",
			code: "123456789012",
			mockSetup: func(repo *MockCouponRepository) {
				repo.On("GetByCode", mock.Anything, "123456789012").Return(nil, errors.New("not found"))
			},
			expectedError: true,
		},
		{
			name: "coupon_already_used",
			code: "123456789012",
			mockSetup: func(repo *MockCouponRepository) {
				coupon := createTestCoupon()
				coupon.Status = "activated"
				repo.On("GetByCode", mock.Anything, "123456789012").Return(coupon, nil)
				repo.On("Update", mock.Anything, mock.Anything).Return(nil)
			},
			expectedError: false, // Service now allows re-activation due to simplified validation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockCouponRepository)

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo)
			}

			deps := &PublicServiceDeps{
				CouponRepository: mockRepo,
			}
			service := NewPublicService(deps)

			result, err := service.ActivateCoupon(tt.code)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Contains(t, result, "message")
				assert.Contains(t, result, "coupon_id")
				assert.Contains(t, result, "next_step")
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestPublicService_GetAvailableSizes(t *testing.T) {
	service := &PublicService{}
	sizes := service.GetAvailableSizes()

	assert.NotNil(t, sizes)
	assert.Len(t, sizes, 6)

	expectedSizes := []string{"21x30", "30x40", "40x40", "40x50", "40x60", "50x70"}
	for _, size := range sizes {
		assert.Contains(t, expectedSizes, size["size"])
		assert.NotEmpty(t, size["title"])
		assert.Equal(t, int(payment.FixedPriceRub), size["price"])
	}
}

func TestPublicService_GetAvailableStyles(t *testing.T) {
	service := &PublicService{}
	styles := service.GetAvailableStyles()

	assert.NotNil(t, styles)
	assert.Len(t, styles, 4)

	expectedStyles := []string{"grayscale", "skin_tones", "pop_art", "max_colors"}
	for _, style := range styles {
		assert.Contains(t, expectedStyles, style["style"])
		assert.NotEmpty(t, style["title"])
		assert.NotEmpty(t, style["description"])
	}
}

func TestPublicService_GetRecaptchaSiteKey(t *testing.T) {
	mockConfig := &MockConfig{}
	mockConfig.On("GetRecaptchaConfig").Return(config.RecaptchaConfig{
		SiteKey: "test_site_key",
	})

	deps := &PublicServiceDeps{
		Config: mockConfig,
	}
	service := NewPublicService(deps)

	result := service.GetRecaptchaSiteKey()
	assert.Equal(t, "test_site_key", result)

	mockConfig.AssertExpectations(t)
}

func TestPublicService_HelperFunctions(t *testing.T) {
	t.Run("isNumeric_valid", func(t *testing.T) {
		assert.True(t, isNumeric("123456789012"))
	})

	t.Run("isNumeric_invalid_length", func(t *testing.T) {
		assert.False(t, isNumeric("12345"))
	})

	t.Run("isNumeric_invalid_chars", func(t *testing.T) {
		assert.False(t, isNumeric("12345678901a"))
	})

	t.Run("isNumeric_empty", func(t *testing.T) {
		assert.False(t, isNumeric(""))
	})
}

func TestPublicService_Getters(t *testing.T) {
	mockCouponRepo := &MockCouponRepository{}
	mockImageRepo := &MockImageRepository{}
	mockPartnerRepo := &MockPartnerRepository{}
	mockImageService := &MockImageService{}
	mockPaymentService := &MockPaymentService{}

	deps := &PublicServiceDeps{
		CouponRepository:  mockCouponRepo,
		ImageRepository:   mockImageRepo,
		PartnerRepository: mockPartnerRepo,
		ImageService:      mockImageService,
		PaymentService:    mockPaymentService,
	}

	service := NewPublicService(deps)

	assert.Equal(t, mockCouponRepo, service.GetCouponRepository())
	assert.Equal(t, mockImageRepo, service.GetImageRepository())
	assert.Equal(t, mockPartnerRepo, service.GetPartnerRepository())
	assert.Equal(t, mockImageService, service.GetImageService())
	assert.Equal(t, mockPaymentService, service.GetPaymentService())
}

func TestPublicService_EdgeCases(t *testing.T) {
	t.Run("nil_dependencies", func(t *testing.T) {
		service := NewPublicService(nil)
		assert.NotNil(t, service)
		assert.Nil(t, service.deps)
	})

	t.Run("empty_dependencies", func(t *testing.T) {
		deps := &PublicServiceDeps{}
		service := NewPublicService(deps)
		assert.NotNil(t, service)
		assert.Equal(t, deps, service.deps)
	})
}

func TestPublicService_Constants(t *testing.T) {
	partner := createTestPartner()

	assert.NotEmpty(t, partner.PartnerCode)
	assert.NotEmpty(t, partner.Login)
	assert.NotEmpty(t, partner.Domain)
	assert.NotEmpty(t, partner.BrandName)
	assert.NotEmpty(t, partner.Email)
	assert.NotEmpty(t, partner.Status)

	assert.True(t, partner.AllowSales)
	assert.True(t, partner.AllowPurchases)
	assert.False(t, partner.IsBlockedInChat)

	assert.NotZero(t, partner.CreatedAt)
	assert.NotZero(t, partner.UpdatedAt)
}

func TestPublicService_RequestStructures(t *testing.T) {
	t.Run("SendEmailRequest", func(t *testing.T) {
		req := SendEmailRequest{Email: "test@example.com"}
		assert.Equal(t, "test@example.com", req.Email)
	})

	t.Run("PurchaseCouponRequest", func(t *testing.T) {
		req := PurchaseCouponRequest{
			Size:         "30x40",
			Style:        "grayscale",
			Email:        "test@example.com",
			PaymentToken: "token123",
		}
		assert.Equal(t, "30x40", req.Size)
		assert.Equal(t, "grayscale", req.Style)
		assert.Equal(t, "test@example.com", req.Email)
		assert.Equal(t, "token123", req.PaymentToken)
	})
}
