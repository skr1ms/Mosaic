package partner

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/skr1ms/mosaic/internal/coupon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock Repository
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

func (m *MockPartnerRepository) GetByDomain(ctx context.Context, domain string) (*Partner, error) {
	args := m.Called(ctx, domain)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Partner), args.Error(1)
}

func (m *MockPartnerRepository) Update(ctx context.Context, partner *Partner) error {
	args := m.Called(ctx, partner)
	return args.Error(0)
}

func (m *MockPartnerRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPartnerRepository) GetAll(ctx context.Context, search, sortBy string) ([]*Partner, error) {
	args := m.Called(ctx, search, sortBy)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Partner), args.Error(1)
}

func (m *MockPartnerRepository) Block(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPartnerRepository) Unblock(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPartnerRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPartnerRepository) GetPartnerCouponsForExport(ctx context.Context, partnerID uuid.UUID, status string) ([]*ExportCouponRequest, error) {
	args := m.Called(ctx, partnerID, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*ExportCouponRequest), args.Error(1)
}

func (m *MockPartnerRepository) CountActive(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockPartnerRepository) CountTotal(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockPartnerRepository) DeleteWithCoupons(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPartnerRepository) GetActivePartners(ctx context.Context) ([]*Partner, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Partner), args.Error(1)
}

func (m *MockPartnerRepository) GetAllCouponsForExport(ctx context.Context) ([]*ExportCouponRequest, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*ExportCouponRequest), args.Error(1)
}

func (m *MockPartnerRepository) GetByEmail(ctx context.Context, email string) (*Partner, error) {
	args := m.Called(ctx, email)
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

func (m *MockPartnerRepository) GetCouponsStatistics(ctx context.Context, partnerID uuid.UUID) (map[string]int64, error) {
	args := m.Called(ctx, partnerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]int64), args.Error(1)
}

func (m *MockPartnerRepository) GetNextPartnerCode(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}

func (m *MockPartnerRepository) GetTopByActivity(ctx context.Context, limit int) ([]*Partner, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Partner), args.Error(1)
}

func (m *MockPartnerRepository) Search(ctx context.Context, query, status, sortBy, order string) ([]*Partner, error) {
	args := m.Called(ctx, query, status, sortBy, order)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Partner), args.Error(1)
}

func (m *MockPartnerRepository) UpdatePassword(ctx context.Context, id uuid.UUID, hashedPassword string) error {
	args := m.Called(ctx, id, hashedPassword)
	return args.Error(0)
}

func (m *MockPartnerRepository) UpdateEmail(ctx context.Context, id uuid.UUID, email string) error {
	args := m.Called(ctx, id, email)
	return args.Error(0)
}

func (m *MockPartnerRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

type MockCouponRepository struct {
	mock.Mock
}

func (m *MockCouponRepository) GetByPartnerID(ctx context.Context, partnerID uuid.UUID) ([]*coupon.Coupon, error) {
	args := m.Called(ctx, partnerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*coupon.Coupon), args.Error(1)
}

func (m *MockCouponRepository) GetByCode(ctx context.Context, code string) (*coupon.Coupon, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*coupon.Coupon), args.Error(1)
}

func (m *MockCouponRepository) CountByPartnerID(ctx context.Context, partnerID uuid.UUID) (int, error) {
	args := m.Called(ctx, partnerID)
	return args.Int(0), args.Error(1)
}

func (m *MockCouponRepository) GetStatistics(ctx context.Context, partnerID *uuid.UUID) (map[string]int64, error) {
	args := m.Called(ctx, partnerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]int64), args.Error(1)
}

func (m *MockCouponRepository) GetRecentActivatedByPartner(ctx context.Context, partnerID uuid.UUID, limit int) ([]*coupon.Coupon, error) {
	args := m.Called(ctx, partnerID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*coupon.Coupon), args.Error(1)
}

func (m *MockCouponRepository) SearchPartnerCoupons(ctx context.Context, partnerID uuid.UUID, code, status, size, style string, createdFrom, createdTo, usedFrom, usedTo *time.Time, sortBy, sortOrder string, page, limit int) ([]*coupon.Coupon, int, error) {
	args := m.Called(ctx, partnerID, code, status, size, style, createdFrom, createdTo, usedFrom, usedTo, sortBy, sortOrder, page, limit)
	return args.Get(0).([]*coupon.Coupon), args.Int(1), args.Error(2)
}

func (m *MockCouponRepository) GetTopActivatedByPartner(ctx context.Context, limit int) ([]coupon.PartnerCount, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]coupon.PartnerCount), args.Error(1)
}

func (m *MockCouponRepository) GetTopPurchasedByPartner(ctx context.Context, limit int) ([]coupon.PartnerCount, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]coupon.PartnerCount), args.Error(1)
}

func (m *MockCouponRepository) CountActivatedByPartner(ctx context.Context, partnerID uuid.UUID) (int64, error) {
	args := m.Called(ctx, partnerID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCouponRepository) CountBrandedPurchasesByPartner(ctx context.Context, partnerID uuid.UUID) (int64, error) {
	args := m.Called(ctx, partnerID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCouponRepository) GetByID(ctx context.Context, id uuid.UUID) (*coupon.Coupon, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*coupon.Coupon), args.Error(1)
}

// Mock Coupon
type Coupon struct {
	ID   uuid.UUID `json:"id"`
	Code string    `json:"code"`
}

// Mock Coupon Statistics
type CouponStatistics struct {
	Total      int64 `json:"total"`
	Activated  int64 `json:"activated"`
	Used       int64 `json:"used"`
	Processing int64 `json:"processing"`
}

func createTestPartner() *Partner {
	return &Partner{
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

func createTestExportCoupon() *ExportCouponRequest {
	return &ExportCouponRequest{
		CouponCode:    "TEST-1234-5678",
		PartnerID:     uuid.New(),
		PartnerStatus: "active",
		CouponStatus:  "new",
		Size:          "30x40",
		Style:         "grayscale",
		BrandName:     "Test Brand",
		Email:         "test@example.com",
		CreatedAt:     time.Now(),
	}
}

func TestNewPartnerService(t *testing.T) {
	deps := &PartnerServiceDeps{
		PartnerRepository: &MockPartnerRepository{},
		CouponRepository:  &MockCouponRepository{},
	}

	service := NewPartnerService(deps)

	assert.NotNil(t, service)
	assert.NotNil(t, service.deps)
	assert.Equal(t, deps, service.deps)
}

func TestPartnerService_GetPartnerRepository(t *testing.T) {
	mockRepo := &MockPartnerRepository{}
	deps := &PartnerServiceDeps{
		PartnerRepository: mockRepo,
	}

	service := NewPartnerService(deps)

	result := service.GetPartnerRepository()
	assert.Equal(t, mockRepo, result)
}

func TestPartnerService_GetCouponRepository(t *testing.T) {
	mockRepo := &MockCouponRepository{}
	deps := &PartnerServiceDeps{
		CouponRepository: mockRepo,
	}

	service := NewPartnerService(deps)

	result := service.GetCouponRepository()
	assert.Equal(t, mockRepo, result)
}

func TestPartnerService_PartnerLogin(t *testing.T) {
	tests := []struct {
		name          string
		login         string
		password      string
		expectedError bool
	}{
		{
			name:          "login_not_implemented",
			login:         "partner@example.com",
			password:      "password123",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps := &PartnerServiceDeps{
				PartnerRepository: &MockPartnerRepository{},
			}
			service := NewPartnerService(deps)

			partner, tokenPair, err := service.PartnerLogin(tt.login, tt.password)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, partner)
				assert.Nil(t, tokenPair)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, partner)
				assert.NotNil(t, tokenPair)
			}
		})
	}
}

func TestPartnerService_DeletePartnerWithCoupons(t *testing.T) {
	tests := []struct {
		name          string
		partnerID     uuid.UUID
		mockSetup     func(*MockPartnerRepository)
		expectedError bool
	}{
		{
			name:      "successful_deletion",
			partnerID: uuid.New(),
			mockSetup: func(repo *MockPartnerRepository) {
				repo.On("DeleteWithCoupons", mock.Anything, mock.Anything).Return(nil)
			},
			expectedError: false,
		},
		{
			name:      "deletion_failed",
			partnerID: uuid.New(),
			mockSetup: func(repo *MockPartnerRepository) {
				repo.On("DeleteWithCoupons", mock.Anything, mock.Anything).Return(errors.New("deletion failed"))
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

			deps := &PartnerServiceDeps{
				PartnerRepository: mockRepo,
			}
			service := NewPartnerService(deps)

			err := service.DeletePartnerWithCoupons(context.Background(), tt.partnerID)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestPartnerService_ExportCoupons(t *testing.T) {
	tests := []struct {
		name          string
		partnerID     uuid.UUID
		status        string
		format        string
		mockSetup     func(*MockPartnerRepository)
		expectedError bool
	}{
		{
			name:      "successful_export",
			partnerID: uuid.New(),
			status:    "active",
			format:    "csv",
			mockSetup: func(repo *MockPartnerRepository) {
				coupons := []*ExportCouponRequest{
					createTestExportCoupon(),
					createTestExportCoupon(),
				}
				repo.On("GetPartnerCouponsForExport", mock.Anything, mock.Anything, "active").Return(coupons, nil)
			},
			expectedError: false,
		},
		{
			name:      "no_coupons_found",
			partnerID: uuid.New(),
			status:    "active",
			format:    "csv",
			mockSetup: func(repo *MockPartnerRepository) {
				repo.On("GetPartnerCouponsForExport", mock.Anything, mock.Anything, "active").Return([]*ExportCouponRequest{}, nil)
			},
			expectedError: true,
		},
		{
			name:      "repository_error",
			partnerID: uuid.New(),
			status:    "active",
			format:    "csv",
			mockSetup: func(repo *MockPartnerRepository) {
				repo.On("GetPartnerCouponsForExport", mock.Anything, mock.Anything, "active").Return(nil, errors.New("database error"))
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

			deps := &PartnerServiceDeps{
				PartnerRepository: mockRepo,
			}
			service := NewPartnerService(deps)

			content, filename, contentType, err := service.ExportCoupons(tt.partnerID, tt.status, tt.format)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Empty(t, filename)
				assert.Empty(t, content)
				assert.Empty(t, contentType)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, filename)
				assert.NotEmpty(t, content)
				assert.NotEmpty(t, contentType)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestPartnerService_GetComparisonStatistics(t *testing.T) {
	tests := []struct {
		name          string
		partnerID     uuid.UUID
		mockSetup     func(*MockPartnerRepository, *MockCouponRepository)
		expectedError bool
	}{
		{
			name:      "successful_statistics",
			partnerID: uuid.New(),
			mockSetup: func(partnerRepo *MockPartnerRepository, couponRepo *MockCouponRepository) {
				partner := createTestPartner()
				partner.ID = uuid.New()
				partnerRepo.On("GetByID", mock.Anything, mock.Anything).Return(partner, nil)

				couponRepo.On("GetTopActivatedByPartner", mock.Anything, 100).Return([]coupon.PartnerCount{}, nil)
				couponRepo.On("GetTopPurchasedByPartner", mock.Anything, 100).Return([]coupon.PartnerCount{}, nil)
				couponRepo.On("CountActivatedByPartner", mock.Anything, mock.Anything).Return(int64(10), nil)
				couponRepo.On("CountBrandedPurchasesByPartner", mock.Anything, mock.Anything).Return(int64(5), nil)
			},
			expectedError: false,
		},
		{
			name:      "repository_errors_handled",
			partnerID: uuid.New(),
			mockSetup: func(partnerRepo *MockPartnerRepository, couponRepo *MockCouponRepository) {
				partner := createTestPartner()
				partner.ID = uuid.New()
				partnerRepo.On("GetByID", mock.Anything, mock.Anything).Return(partner, nil)

				couponRepo.On("GetTopActivatedByPartner", mock.Anything, 100).Return(nil, errors.New("error"))
				couponRepo.On("GetTopPurchasedByPartner", mock.Anything, 100).Return(nil, errors.New("error"))
				couponRepo.On("CountActivatedByPartner", mock.Anything, mock.Anything).Return(int64(0), errors.New("error"))
				couponRepo.On("CountBrandedPurchasesByPartner", mock.Anything, mock.Anything).Return(int64(0), errors.New("error"))
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPartnerRepo := new(MockPartnerRepository)
			mockCouponRepo := new(MockCouponRepository)

			if tt.mockSetup != nil {
				tt.mockSetup(mockPartnerRepo, mockCouponRepo)
			}

			deps := &PartnerServiceDeps{
				PartnerRepository: mockPartnerRepo,
				CouponRepository:  mockCouponRepo,
			}
			service := NewPartnerService(deps)

			result, err := service.GetComparisonStatistics(context.Background(), tt.partnerID)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Contains(t, result, "used")
				assert.Contains(t, result, "purchased")
				assert.Contains(t, result, "me")
			}

			mockPartnerRepo.AssertExpectations(t)
			mockCouponRepo.AssertExpectations(t)
		})
	}
}

func TestPartnerService_EdgeCases(t *testing.T) {
	t.Run("nil_dependencies", func(t *testing.T) {
		service := NewPartnerService(nil)
		assert.NotNil(t, service)
		assert.Nil(t, service.deps)
	})

	t.Run("empty_dependencies", func(t *testing.T) {
		deps := &PartnerServiceDeps{}
		service := NewPartnerService(deps)
		assert.NotNil(t, service)
		assert.Equal(t, deps, service.deps)
	})
}

func TestPartnerService_ExportCoupons_Formats(t *testing.T) {
	partnerID := uuid.New()

	tests := []struct {
		name          string
		status        string
		expectedError bool
	}{
		{
			name:          "empty_status",
			status:        "",
			expectedError: false,
		},
		{
			name:          "specific_status",
			status:        "active",
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockPartnerRepository)
			coupons := []*ExportCouponRequest{
				createTestExportCoupon(),
			}
			mockRepo.On("GetPartnerCouponsForExport", mock.Anything, partnerID, tt.status).Return(coupons, nil)

			deps := &PartnerServiceDeps{
				PartnerRepository: mockRepo,
			}
			service := NewPartnerService(deps)

			content, filename, contentType, err := service.ExportCoupons(partnerID, tt.status, "csv")

			assert.NoError(t, err)
			assert.NotEmpty(t, content)
			assert.NotEmpty(t, filename)
			assert.Contains(t, contentType, "text/csv")
			assert.Contains(t, filename, ".csv")

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestPartnerService_Constants(t *testing.T) {
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

func TestPartnerService_ExportCouponRequest_Structure(t *testing.T) {
	coupon := createTestExportCoupon()

	assert.NotEmpty(t, coupon.CouponCode)
	assert.NotZero(t, coupon.PartnerID)
	assert.NotEmpty(t, coupon.PartnerStatus)
	assert.NotEmpty(t, coupon.CouponStatus)
	assert.NotEmpty(t, coupon.Size)
	assert.NotEmpty(t, coupon.Style)
	assert.NotEmpty(t, coupon.BrandName)
	assert.NotEmpty(t, coupon.Email)
	assert.NotZero(t, coupon.CreatedAt)
}
