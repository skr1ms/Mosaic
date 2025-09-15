package stats

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/skr1ms/mosaic/internal/partner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock Coupon Repository
type MockCouponRepository struct {
	mock.Mock
}

func (m *MockCouponRepository) CountTotal(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCouponRepository) CountActivated(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCouponRepository) CountByPartner(ctx context.Context, partnerID uuid.UUID) (int64, error) {
	args := m.Called(ctx, partnerID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCouponRepository) CountActivatedByPartner(ctx context.Context, partnerID uuid.UUID) (int64, error) {
	args := m.Called(ctx, partnerID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCouponRepository) CountBrandedPurchasesByPartner(ctx context.Context, partnerID uuid.UUID) (int64, error) {
	args := m.Called(ctx, partnerID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCouponRepository) GetLastActivityByPartner(ctx context.Context, partnerID uuid.UUID) (*time.Time, error) {
	args := m.Called(ctx, partnerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*time.Time), args.Error(1)
}

func (m *MockCouponRepository) CountActivatedInTimeRange(ctx context.Context, from, to time.Time) (int64, error) {
	args := m.Called(ctx, from, to)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCouponRepository) GetExtendedStatusCounts(ctx context.Context, partnerID *uuid.UUID) (map[string]int64, error) {
	args := m.Called(ctx, partnerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]int64), args.Error(1)
}

func (m *MockCouponRepository) GetSizeCounts(ctx context.Context, partnerID *uuid.UUID) (map[string]int64, error) {
	args := m.Called(ctx, partnerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]int64), args.Error(1)
}

func (m *MockCouponRepository) GetStyleCounts(ctx context.Context, partnerID *uuid.UUID) (map[string]int64, error) {
	args := m.Called(ctx, partnerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]int64), args.Error(1)
}

func (m *MockCouponRepository) GetTimeSeriesData(ctx context.Context, from, to time.Time, period string, partnerID *uuid.UUID) ([]map[string]any, error) {
	args := m.Called(ctx, from, to, period, partnerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]map[string]any), args.Error(1)
}

func (m *MockCouponRepository) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Mock Partner Repository
type MockPartnerRepository struct {
	mock.Mock
}

func (m *MockPartnerRepository) CountActive(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockPartnerRepository) CountTotal(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockPartnerRepository) GetByID(ctx context.Context, id uuid.UUID) (*partner.Partner, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*partner.Partner), args.Error(1)
}

func (m *MockPartnerRepository) GetAll(ctx context.Context, sortBy string, order string) ([]*partner.Partner, error) {
	args := m.Called(ctx, sortBy, order)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*partner.Partner), args.Error(1)
}

// Mock Redis Client
type MockRedisClient struct {
	mock.Mock
}

func (m *MockRedisClient) Get(ctx context.Context, key string) *redis.StringCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.StringCmd)
}

func (m *MockRedisClient) Set(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd {
	args := m.Called(ctx, key, value, expiration)
	return args.Get(0).(*redis.StatusCmd)
}

func (m *MockRedisClient) Ping(ctx context.Context) *redis.StatusCmd {
	args := m.Called(ctx)
	return args.Get(0).(*redis.StatusCmd)
}

func (m *MockRedisClient) LLen(ctx context.Context, key string) *redis.IntCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.IntCmd)
}

func createTestPartner() *partner.Partner {
	return &partner.Partner{
		ID:        uuid.New(),
		BrandName: "Test Partner",
		Domain:    "test.example.com",
		Status:    "active",
	}
}

func createTestTimeSeriesData() []map[string]any {
	return []map[string]any{
		{
			"date":               "2024-01-01",
			"coupons_created":    int64(10),
			"coupons_activated":  int64(8),
			"coupons_purchased":  int64(2),
			"new_partners_count": int64(1),
		},
		{
			"date":               "2024-01-02",
			"coupons_created":    int64(15),
			"coupons_activated":  int64(12),
			"coupons_purchased":  int64(3),
			"new_partners_count": int64(0),
		},
	}
}

func createTestStatusCounts() map[string]int64 {
	return map[string]int64{
		"new":       100,
		"activated": 80,
		"used":      60,
		"completed": 40,
	}
}

func createTestSizeCounts() map[string]int64 {
	return map[string]int64{
		"21x30": 50,
		"30x40": 80,
		"40x40": 60,
		"40x50": 70,
		"40x60": 40,
		"50x70": 30,
	}
}

func createTestStyleCounts() map[string]int64 {
	return map[string]int64{
		"grayscale":  120,
		"skin_tones": 90,
		"pop_art":    70,
		"max_colors": 60,
	}
}

func TestNewStatsService(t *testing.T) {
	deps := &StatsServiceDeps{
		CouponRepository:  &MockCouponRepository{},
		PartnerRepository: &MockPartnerRepository{},
		RedisClient:       &MockRedisClient{},
	}

	service := NewStatsService(deps)

	assert.NotNil(t, service)
	assert.NotNil(t, service.deps)
	assert.Equal(t, deps, service.deps)
}

func TestStatsService_Structure(t *testing.T) {
	service := &StatsService{}
	assert.NotNil(t, service)

	service = NewStatsService(nil)
	assert.NotNil(t, service)
}

func TestStatsService_ValidateStructure(t *testing.T) {
	tests := []struct {
		name string
		deps *StatsServiceDeps
	}{
		{
			name: "nil_dependencies",
			deps: nil,
		},
		{
			name: "empty_dependencies",
			deps: &StatsServiceDeps{},
		},
		{
			name: "with_dependencies",
			deps: &StatsServiceDeps{
				CouponRepository:  &MockCouponRepository{},
				PartnerRepository: &MockPartnerRepository{},
				RedisClient:       &MockRedisClient{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewStatsService(tt.deps)

			assert.NotNil(t, service)

			if tt.deps != nil {
				assert.Equal(t, tt.deps, service.deps)
			}
		})
	}
}

func TestStatsService_GetGeneralStats(t *testing.T) {
	tests := []struct {
		name          string
		mockSetup     func(*MockCouponRepository, *MockPartnerRepository, *MockRedisClient)
		expectedError bool
	}{
		{
			name: "successful_get_stats",
			mockSetup: func(couponRepo *MockCouponRepository, partnerRepo *MockPartnerRepository, redisClient *MockRedisClient) {
				redisClient.On("Get", mock.Anything, "general_stats").Return(&redis.StringCmd{})
				couponRepo.On("CountTotal", mock.Anything).Return(int64(1000), nil)
				couponRepo.On("CountActivated", mock.Anything).Return(int64(800), nil)
				partnerRepo.On("CountActive", mock.Anything).Return(int64(50), nil)
				partnerRepo.On("CountTotal", mock.Anything).Return(int64(100), nil)
				redisClient.On("Set", mock.Anything, "general_stats", mock.Anything, 5*time.Minute).Return(&redis.StatusCmd{})
			},
			expectedError: false,
		},
		{
			name: "coupon_count_error",
			mockSetup: func(couponRepo *MockCouponRepository, partnerRepo *MockPartnerRepository, redisClient *MockRedisClient) {
				redisClient.On("Get", mock.Anything, "general_stats").Return(&redis.StringCmd{})
				couponRepo.On("CountTotal", mock.Anything).Return(int64(0), errors.New("database error"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCouponRepo := &MockCouponRepository{}
			mockPartnerRepo := &MockPartnerRepository{}
			mockRedisClient := &MockRedisClient{}

			if tt.mockSetup != nil {
				tt.mockSetup(mockCouponRepo, mockPartnerRepo, mockRedisClient)
			}

			deps := &StatsServiceDeps{
				CouponRepository:  mockCouponRepo,
				PartnerRepository: mockPartnerRepo,
				RedisClient:       mockRedisClient,
			}
			service := NewStatsService(deps)

			result, err := service.GetGeneralStats(context.Background())

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, int64(1000), result.TotalCouponsCreated)
				assert.Equal(t, int64(800), result.TotalCouponsActivated)
				assert.Equal(t, float64(80.0), result.ActivationRate)
				assert.Equal(t, int64(50), result.ActivePartnersCount)
				assert.Equal(t, int64(100), result.TotalPartnersCount)
				assert.NotEmpty(t, result.LastUpdated)
			}

			mockCouponRepo.AssertExpectations(t)
			mockPartnerRepo.AssertExpectations(t)
			mockRedisClient.AssertExpectations(t)
		})
	}
}

func TestStatsService_GetPartnerStats(t *testing.T) {
	partnerID := uuid.New()
	partner := createTestPartner()

	tests := []struct {
		name          string
		mockSetup     func(*MockCouponRepository, *MockPartnerRepository, *MockRedisClient)
		expectedError bool
	}{
		{
			name: "successful_get_partner_stats",
			mockSetup: func(couponRepo *MockCouponRepository, partnerRepo *MockPartnerRepository, redisClient *MockRedisClient) {
				redisClient.On("Get", mock.Anything, mock.Anything).Return(&redis.StringCmd{})
				partnerRepo.On("GetByID", mock.Anything, partnerID).Return(partner, nil)
				couponRepo.On("CountByPartner", mock.Anything, partnerID).Return(int64(100), nil)
				couponRepo.On("CountActivatedByPartner", mock.Anything, partnerID).Return(int64(80), nil)
				couponRepo.On("CountBrandedPurchasesByPartner", mock.Anything, partnerID).Return(int64(20), nil)
				couponRepo.On("GetLastActivityByPartner", mock.Anything, partnerID).Return(nil, nil)
				redisClient.On("Set", mock.Anything, mock.Anything, mock.Anything, 10*time.Minute).Return(&redis.StatusCmd{})
			},
			expectedError: false,
		},
		{
			name: "partner_not_found",
			mockSetup: func(couponRepo *MockCouponRepository, partnerRepo *MockPartnerRepository, redisClient *MockRedisClient) {
				redisClient.On("Get", mock.Anything, mock.Anything).Return(&redis.StringCmd{})
				partnerRepo.On("GetByID", mock.Anything, partnerID).Return(nil, errors.New("not found"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCouponRepo := &MockCouponRepository{}
			mockPartnerRepo := &MockPartnerRepository{}
			mockRedisClient := &MockRedisClient{}

			if tt.mockSetup != nil {
				tt.mockSetup(mockCouponRepo, mockPartnerRepo, mockRedisClient)
			}

			deps := &StatsServiceDeps{
				CouponRepository:  mockCouponRepo,
				PartnerRepository: mockPartnerRepo,
				RedisClient:       mockRedisClient,
			}
			service := NewStatsService(deps)

			result, err := service.GetPartnerStats(context.Background(), partnerID)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, partnerID, result.PartnerID)
				assert.Equal(t, partner.BrandName, result.PartnerName)
				assert.Equal(t, int64(100), result.TotalCoupons)
				assert.Equal(t, int64(80), result.ActivatedCoupons)
				assert.Equal(t, int64(20), result.UnusedCoupons)
				assert.Equal(t, int64(20), result.BrandedSitePurchases)
				assert.Equal(t, float64(80.0), result.ActivationRate)
			}

			mockCouponRepo.AssertExpectations(t)
			mockPartnerRepo.AssertExpectations(t)
			mockRedisClient.AssertExpectations(t)
		})
	}
}

func TestStatsService_GetAllPartnersStats(t *testing.T) {
	partners := []*partner.Partner{
		createTestPartner(),
		createTestPartner(),
	}

	tests := []struct {
		name          string
		mockSetup     func(*MockPartnerRepository, *MockCouponRepository, *MockRedisClient)
		expectedError bool
	}{
		{
			name: "successful_get_all_partners_stats",
			mockSetup: func(partnerRepo *MockPartnerRepository, couponRepo *MockCouponRepository, redisClient *MockRedisClient) {
				partnerRepo.On("GetAll", mock.Anything, "created_at", "desc").Return(partners, nil)
				redisClient.On("Get", mock.Anything, mock.Anything).Return(&redis.StringCmd{})
				partnerRepo.On("GetByID", mock.Anything, mock.Anything).Return(partners[0], nil)
				couponRepo.On("CountByPartner", mock.Anything, mock.Anything).Return(int64(100), nil)
				couponRepo.On("CountActivatedByPartner", mock.Anything, mock.Anything).Return(int64(80), nil)
				couponRepo.On("CountBrandedPurchasesByPartner", mock.Anything, mock.Anything).Return(int64(20), nil)
				couponRepo.On("GetLastActivityByPartner", mock.Anything, mock.Anything).Return(nil, nil)
				redisClient.On("Set", mock.Anything, mock.Anything, mock.Anything, 10*time.Minute).Return(&redis.StatusCmd{})
			},
			expectedError: false,
		},
		{
			name: "failed_to_get_partners",
			mockSetup: func(partnerRepo *MockPartnerRepository, couponRepo *MockCouponRepository, redisClient *MockRedisClient) {
				partnerRepo.On("GetAll", mock.Anything, "created_at", "desc").Return(nil, errors.New("database error"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPartnerRepo := &MockPartnerRepository{}
			mockCouponRepo := &MockCouponRepository{}
			mockRedisClient := &MockRedisClient{}

			if tt.mockSetup != nil {
				tt.mockSetup(mockPartnerRepo, mockCouponRepo, mockRedisClient)
			}

			deps := &StatsServiceDeps{
				PartnerRepository: mockPartnerRepo,
				CouponRepository:  mockCouponRepo,
				RedisClient:       mockRedisClient,
			}
			service := NewStatsService(deps)

			result, err := service.GetAllPartnersStats(context.Background())

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Len(t, result.Partners, 2)
				assert.Equal(t, int64(2), result.Total)
			}

			mockPartnerRepo.AssertExpectations(t)
			mockCouponRepo.AssertExpectations(t)
			mockRedisClient.AssertExpectations(t)
		})
	}
}

func TestStatsService_GetTimeSeriesStats(t *testing.T) {
	filters := &StatsFiltersRequest{
		Period:   stringPtr("day"),
		DateFrom: stringPtr("2024-01-01"),
		DateTo:   stringPtr("2024-01-02"),
	}

	tests := []struct {
		name          string
		mockSetup     func(*MockCouponRepository)
		expectedError bool
	}{
		{
			name: "successful_get_time_series_stats",
			mockSetup: func(couponRepo *MockCouponRepository) {
				data := createTestTimeSeriesData()
				couponRepo.On("GetTimeSeriesData", mock.Anything, mock.Anything, mock.Anything, "day", (*uuid.UUID)(nil)).Return(data, nil)
			},
			expectedError: false,
		},
		{
			name: "failed_to_get_time_series_data",
			mockSetup: func(couponRepo *MockCouponRepository) {
				couponRepo.On("GetTimeSeriesData", mock.Anything, mock.Anything, mock.Anything, "day", (*uuid.UUID)(nil)).Return(nil, errors.New("database error"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCouponRepo := &MockCouponRepository{}

			if tt.mockSetup != nil {
				tt.mockSetup(mockCouponRepo)
			}

			deps := &StatsServiceDeps{
				CouponRepository: mockCouponRepo,
			}
			service := NewStatsService(deps)

			result, err := service.GetTimeSeriesStats(context.Background(), filters)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, "day", result.Period)
				assert.Len(t, result.Data, 2)
			}

			mockCouponRepo.AssertExpectations(t)
		})
	}
}

func TestStatsService_GetSystemHealth(t *testing.T) {
	tests := []struct {
		name          string
		mockSetup     func(*MockCouponRepository, *MockRedisClient)
		expectedError bool
	}{
		{
			name: "healthy_system",
			mockSetup: func(couponRepo *MockCouponRepository, redisClient *MockRedisClient) {
				redisClient.On("Get", mock.Anything, "system_health").Return(&redis.StringCmd{})
				couponRepo.On("HealthCheck", mock.Anything).Return(nil)
				redisClient.On("Ping", mock.Anything).Return(&redis.StatusCmd{})
				redisClient.On("LLen", mock.Anything, "image_processing_queue").Return(&redis.IntCmd{})
				redisClient.On("Set", mock.Anything, "system_health", mock.Anything, 1*time.Minute).Return(&redis.StatusCmd{})
			},
			expectedError: false,
		},
		{
			name: "unhealthy_database",
			mockSetup: func(couponRepo *MockCouponRepository, redisClient *MockRedisClient) {
				redisClient.On("Get", mock.Anything, "system_health").Return(&redis.StringCmd{})
				couponRepo.On("HealthCheck", mock.Anything).Return(errors.New("database error"))
				redisClient.On("Ping", mock.Anything).Return(&redis.StatusCmd{})
				redisClient.On("LLen", mock.Anything, "image_processing_queue").Return(&redis.IntCmd{})
				redisClient.On("Set", mock.Anything, "system_health", mock.Anything, 1*time.Minute).Return(&redis.StatusCmd{})
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCouponRepo := &MockCouponRepository{}
			mockRedisClient := &MockRedisClient{}

			if tt.mockSetup != nil {
				tt.mockSetup(mockCouponRepo, mockRedisClient)
			}

			deps := &StatsServiceDeps{
				CouponRepository: mockCouponRepo,
				RedisClient:      mockRedisClient,
			}
			service := NewStatsService(deps)

			result, err := service.GetSystemHealth(context.Background())

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.NotEmpty(t, result.Status)
			assert.NotEmpty(t, result.DatabaseStatus)
			assert.NotEmpty(t, result.RedisStatus)
			assert.NotZero(t, result.AverageProcessingTime)
			assert.NotZero(t, result.ErrorRate)
			assert.NotEmpty(t, result.Uptime)
			assert.NotEmpty(t, result.LastUpdated)

			mockCouponRepo.AssertExpectations(t)
			mockRedisClient.AssertExpectations(t)
		})
	}
}

func TestStatsService_GetCouponsByStatus(t *testing.T) {
	tests := []struct {
		name          string
		partnerID     *uuid.UUID
		mockSetup     func(*MockCouponRepository)
		expectedError bool
	}{
		{
			name:      "successful_get_coupons_by_status",
			partnerID: nil,
			mockSetup: func(couponRepo *MockCouponRepository) {
				statusCounts := createTestStatusCounts()
				couponRepo.On("GetExtendedStatusCounts", mock.Anything, (*uuid.UUID)(nil)).Return(statusCounts, nil)
			},
			expectedError: false,
		},
		{
			name:      "failed_to_get_status_counts",
			partnerID: nil,
			mockSetup: func(couponRepo *MockCouponRepository) {
				couponRepo.On("GetExtendedStatusCounts", mock.Anything, (*uuid.UUID)(nil)).Return(nil, errors.New("database error"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCouponRepo := &MockCouponRepository{}

			if tt.mockSetup != nil {
				tt.mockSetup(mockCouponRepo)
			}

			deps := &StatsServiceDeps{
				CouponRepository: mockCouponRepo,
			}
			service := NewStatsService(deps)

			result, err := service.GetCouponsByStatus(context.Background(), tt.partnerID)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, int64(100), result.New)
				assert.Equal(t, int64(80), result.Activated)
				assert.Equal(t, int64(60), result.Used)
				assert.Equal(t, int64(40), result.Completed)
			}

			mockCouponRepo.AssertExpectations(t)
		})
	}
}

func TestStatsService_GetCouponsBySize(t *testing.T) {
	tests := []struct {
		name          string
		partnerID     *uuid.UUID
		mockSetup     func(*MockCouponRepository)
		expectedError bool
	}{
		{
			name:      "successful_get_coupons_by_size",
			partnerID: nil,
			mockSetup: func(couponRepo *MockCouponRepository) {
				sizeCounts := createTestSizeCounts()
				couponRepo.On("GetSizeCounts", mock.Anything, (*uuid.UUID)(nil)).Return(sizeCounts, nil)
			},
			expectedError: false,
		},
		{
			name:      "failed_to_get_size_counts",
			partnerID: nil,
			mockSetup: func(couponRepo *MockCouponRepository) {
				couponRepo.On("GetSizeCounts", mock.Anything, (*uuid.UUID)(nil)).Return(nil, errors.New("database error"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCouponRepo := &MockCouponRepository{}

			if tt.mockSetup != nil {
				tt.mockSetup(mockCouponRepo)
			}

			deps := &StatsServiceDeps{
				CouponRepository: mockCouponRepo,
			}
			service := NewStatsService(deps)

			result, err := service.GetCouponsBySize(context.Background(), tt.partnerID)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, int64(50), result.Size21x30)
				assert.Equal(t, int64(80), result.Size30x40)
				assert.Equal(t, int64(60), result.Size40x40)
				assert.Equal(t, int64(70), result.Size40x50)
				assert.Equal(t, int64(40), result.Size40x60)
				assert.Equal(t, int64(30), result.Size50x70)
			}

			mockCouponRepo.AssertExpectations(t)
		})
	}
}

func TestStatsService_GetCouponsByStyle(t *testing.T) {
	tests := []struct {
		name          string
		partnerID     *uuid.UUID
		mockSetup     func(*MockCouponRepository)
		expectedError bool
	}{
		{
			name:      "successful_get_coupons_by_style",
			partnerID: nil,
			mockSetup: func(couponRepo *MockCouponRepository) {
				styleCounts := createTestStyleCounts()
				couponRepo.On("GetStyleCounts", mock.Anything, (*uuid.UUID)(nil)).Return(styleCounts, nil)
			},
			expectedError: false,
		},
		{
			name:      "failed_to_get_style_counts",
			partnerID: nil,
			mockSetup: func(couponRepo *MockCouponRepository) {
				couponRepo.On("GetStyleCounts", mock.Anything, (*uuid.UUID)(nil)).Return(nil, errors.New("database error"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCouponRepo := &MockCouponRepository{}

			if tt.mockSetup != nil {
				tt.mockSetup(mockCouponRepo)
			}

			deps := &StatsServiceDeps{
				CouponRepository: mockCouponRepo,
			}
			service := NewStatsService(deps)

			result, err := service.GetCouponsByStyle(context.Background(), tt.partnerID)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, int64(120), result.Grayscale)
				assert.Equal(t, int64(90), result.SkinTones)
				assert.Equal(t, int64(70), result.PopArt)
				assert.Equal(t, int64(60), result.MaxColors)
			}

			mockCouponRepo.AssertExpectations(t)
		})
	}
}

func TestStatsService_GetTopPartners(t *testing.T) {
	partners := []*partner.Partner{
		createTestPartner(),
		createTestPartner(),
	}

	tests := []struct {
		name          string
		limit         int
		mockSetup     func(*MockPartnerRepository, *MockCouponRepository)
		expectedError bool
	}{
		{
			name:  "successful_get_top_partners",
			limit: 5,
			mockSetup: func(partnerRepo *MockPartnerRepository, couponRepo *MockCouponRepository) {
				partnerRepo.On("GetAll", mock.Anything, "created_at", "desc").Return(partners, nil)
				couponRepo.On("CountByPartner", mock.Anything, mock.Anything).Return(int64(100), nil)
				couponRepo.On("CountActivatedByPartner", mock.Anything, mock.Anything).Return(int64(80), nil)
			},
			expectedError: false,
		},
		{
			name:  "default_limit",
			limit: 0,
			mockSetup: func(partnerRepo *MockPartnerRepository, couponRepo *MockCouponRepository) {
				partnerRepo.On("GetAll", mock.Anything, "created_at", "desc").Return(partners, nil)
				couponRepo.On("CountByPartner", mock.Anything, mock.Anything).Return(int64(100), nil)
				couponRepo.On("CountActivatedByPartner", mock.Anything, mock.Anything).Return(int64(80), nil)
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPartnerRepo := &MockPartnerRepository{}
			mockCouponRepo := &MockCouponRepository{}

			if tt.mockSetup != nil {
				tt.mockSetup(mockPartnerRepo, mockCouponRepo)
			}

			deps := &StatsServiceDeps{
				PartnerRepository: mockPartnerRepo,
				CouponRepository:  mockCouponRepo,
			}
			service := NewStatsService(deps)

			result, err := service.GetTopPartners(context.Background(), tt.limit)

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.NotEmpty(t, result.Partners)

			mockPartnerRepo.AssertExpectations(t)
			mockCouponRepo.AssertExpectations(t)
		})
	}
}

func TestStatsService_GetRealTimeStats(t *testing.T) {
	tests := []struct {
		name          string
		mockSetup     func(*MockCouponRepository)
		expectedError bool
	}{
		{
			name: "successful_get_real_time_stats",
			mockSetup: func(couponRepo *MockCouponRepository) {
				couponRepo.On("CountActivatedInTimeRange", mock.Anything, mock.Anything, mock.Anything).Return(int64(5), nil)
			},
			expectedError: false,
		},
		{
			name: "failed_to_get_coupons_count",
			mockSetup: func(couponRepo *MockCouponRepository) {
				couponRepo.On("CountActivatedInTimeRange", mock.Anything, mock.Anything, mock.Anything).Return(int64(0), errors.New("database error"))
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCouponRepo := &MockCouponRepository{}

			if tt.mockSetup != nil {
				tt.mockSetup(mockCouponRepo)
			}

			deps := &StatsServiceDeps{
				CouponRepository: mockCouponRepo,
			}
			service := NewStatsService(deps)

			result, err := service.GetRealTimeStats(context.Background())

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.NotZero(t, result.Timestamp)
			assert.NotZero(t, result.ActiveUsers)
			assert.NotZero(t, result.ImagesProcessingNow)
			assert.NotZero(t, result.SystemLoad)

			mockCouponRepo.AssertExpectations(t)
		})
	}
}

func TestStatsService_EdgeCases(t *testing.T) {
	t.Run("nil_dependencies", func(t *testing.T) {
		service := NewStatsService(nil)
		assert.NotNil(t, service)
		assert.Nil(t, service.deps)
	})

	t.Run("empty_dependencies", func(t *testing.T) {
		deps := &StatsServiceDeps{}
		service := NewStatsService(deps)
		assert.NotNil(t, service)
		assert.Equal(t, deps, service.deps)
	})
}

func TestStatsService_Constants(t *testing.T) {
	partner := createTestPartner()

	assert.NotEmpty(t, partner.ID)
	assert.NotEmpty(t, partner.BrandName)
	assert.NotEmpty(t, partner.Domain)
	assert.NotEmpty(t, partner.Status)
}

func TestStatsService_RequestStructures(t *testing.T) {
	t.Run("StatsFiltersRequest", func(t *testing.T) {
		req := &StatsFiltersRequest{
			PartnerID: &uuid.UUID{},
			DateFrom:  stringPtr("2024-01-01"),
			DateTo:    stringPtr("2024-01-02"),
			Period:    stringPtr("day"),
		}
		assert.NotNil(t, req.PartnerID)
		assert.Equal(t, "2024-01-01", *req.DateFrom)
		assert.Equal(t, "2024-01-02", *req.DateTo)
		assert.Equal(t, "day", *req.Period)
	})

	t.Run("TimeSeriesStatsPoint", func(t *testing.T) {
		point := TimeSeriesStatsPoint{
			Date:             "2024-01-01",
			CouponsCreated:   10,
			CouponsActivated: 8,
			CouponsPurchased: 2,
			NewPartnersCount: 1,
		}
		assert.Equal(t, "2024-01-01", point.Date)
		assert.Equal(t, int64(10), point.CouponsCreated)
		assert.Equal(t, int64(8), point.CouponsActivated)
		assert.Equal(t, int64(2), point.CouponsPurchased)
		assert.Equal(t, int64(1), point.NewPartnersCount)
	})
}

func stringPtr(s string) *string {
	return &s
}
