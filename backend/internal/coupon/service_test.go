package coupon

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}

// Mock Repository
type MockCouponRepository struct {
	mock.Mock
}

type MockPartnerRepository struct {
	mock.Mock
}

func (m *MockPartnerRepository) GetByID(ctx context.Context, id uuid.UUID) (*Partner, error) {
	args := m.Called(ctx, id)
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

func (m *MockCouponRepository) Create(ctx context.Context, coupon *Coupon) error {
	args := m.Called(ctx, coupon)
	return args.Error(0)
}

func (m *MockCouponRepository) CreateBatch(ctx context.Context, coupons []*Coupon) error {
	args := m.Called(ctx, coupons)
	return args.Error(0)
}

func (m *MockCouponRepository) GetByID(ctx context.Context, id uuid.UUID) (*Coupon, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Coupon), args.Error(1)
}

func (m *MockCouponRepository) GetByCode(ctx context.Context, code string) (*Coupon, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Coupon), args.Error(1)
}

func (m *MockCouponRepository) GetByPartnerID(ctx context.Context, partnerID uuid.UUID) ([]*Coupon, error) {
	args := m.Called(ctx, partnerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Coupon), args.Error(1)
}

func (m *MockCouponRepository) GetAll(ctx context.Context) ([]*Coupon, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Coupon), args.Error(1)
}

func (m *MockCouponRepository) Update(ctx context.Context, coupon *Coupon) error {
	args := m.Called(ctx, coupon)
	return args.Error(0)
}

func (m *MockCouponRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCouponRepository) BatchDelete(ctx context.Context, ids []uuid.UUID) (int64, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCouponRepository) Search(ctx context.Context, code, status, size, style string, partnerID *uuid.UUID) ([]*Coupon, error) {
	args := m.Called(ctx, code, status, size, style, partnerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Coupon), args.Error(1)
}

func (m *MockCouponRepository) SearchWithPagination(ctx context.Context, code, status, size, style string, partnerID *uuid.UUID, page, limit int, createdFrom, createdTo, usedFrom, usedTo *time.Time, sortBy, sortDir string) ([]*Coupon, int, error) {
	args := m.Called(ctx, code, status, size, style, partnerID, page, limit, createdFrom, createdTo, usedFrom, usedTo, sortBy, sortDir)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*Coupon), args.Int(1), args.Error(2)
}

func (m *MockCouponRepository) GetFiltered(ctx context.Context, filters map[string]any) ([]*Coupon, error) {
	args := m.Called(ctx, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Coupon), args.Error(1)
}

func (m *MockCouponRepository) GetPartnerCouponsWithFilter(ctx context.Context, partnerID uuid.UUID, filters map[string]any, page, limit int, sortBy, order string) ([]*Coupon, int, error) {
	args := m.Called(ctx, partnerID, filters, page, limit, sortBy, order)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*Coupon), args.Int(1), args.Error(2)
}

func (m *MockCouponRepository) GetPartnerCouponByCode(ctx context.Context, partnerID uuid.UUID, code string) (*Coupon, error) {
	args := m.Called(ctx, partnerID, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Coupon), args.Error(1)
}

func (m *MockCouponRepository) GetPartnerCouponDetail(ctx context.Context, partnerID uuid.UUID, couponID uuid.UUID) (*Coupon, error) {
	args := m.Called(ctx, partnerID, couponID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Coupon), args.Error(1)
}

func (m *MockCouponRepository) GetPartnerRecentActivity(ctx context.Context, partnerID uuid.UUID, limit int) ([]*Coupon, error) {
	args := m.Called(ctx, partnerID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Coupon), args.Error(1)
}

func (m *MockCouponRepository) UpdateStatusByPartnerID(ctx context.Context, partnerID uuid.UUID, status bool) error {
	args := m.Called(ctx, partnerID, status)
	return args.Error(0)
}

func (m *MockCouponRepository) ActivateCoupon(ctx context.Context, id uuid.UUID, req ActivateCouponRequest) error {
	args := m.Called(ctx, id, req)
	return args.Error(0)
}

func (m *MockCouponRepository) SendSchema(ctx context.Context, id uuid.UUID, email string) error {
	args := m.Called(ctx, id, email)
	return args.Error(0)
}

func (m *MockCouponRepository) MarkAsPurchased(ctx context.Context, id uuid.UUID, purchaseEmail string) error {
	args := m.Called(ctx, id, purchaseEmail)
	return args.Error(0)
}

func (m *MockCouponRepository) ResetCoupon(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCouponRepository) Reset(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCouponRepository) BatchReset(ctx context.Context, ids []uuid.UUID) ([]uuid.UUID, []uuid.UUID, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).([]uuid.UUID), args.Get(1).([]uuid.UUID), args.Error(2)
}

// Other methods implementation...
func (m *MockCouponRepository) CountByPartnerID(ctx context.Context, partnerID uuid.UUID) (int, error) {
	args := m.Called(ctx, partnerID)
	return args.Int(0), args.Error(1)
}

func (m *MockCouponRepository) CountActivatedByPartnerID(ctx context.Context, partnerID uuid.UUID) (int, error) {
	args := m.Called(ctx, partnerID)
	return args.Int(0), args.Error(1)
}

func (m *MockCouponRepository) CountPurchasedByPartnerID(ctx context.Context, partnerID uuid.UUID) (int, error) {
	args := m.Called(ctx, partnerID)
	return args.Int(0), args.Error(1)
}

func (m *MockCouponRepository) CountTotal(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCouponRepository) CountByStatus(ctx context.Context, status string) (int64, error) {
	args := m.Called(ctx, status)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCouponRepository) CountByPartner(ctx context.Context, partnerID uuid.UUID) (int64, error) {
	args := m.Called(ctx, partnerID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCouponRepository) CountByPartnerAndStatus(ctx context.Context, partnerID uuid.UUID, status string) (int64, error) {
	args := m.Called(ctx, partnerID, status)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCouponRepository) CountActivated(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
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

func (m *MockCouponRepository) CountActivatedInTimeRange(ctx context.Context, from, to time.Time) (int64, error) {
	args := m.Called(ctx, from, to)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCouponRepository) GetStatistics(ctx context.Context, partnerID *uuid.UUID) (map[string]int64, error) {
	args := m.Called(ctx, partnerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]int64), args.Error(1)
}

func (m *MockCouponRepository) GetPartnerStatistics(ctx context.Context, partnerID uuid.UUID) (map[string]any, error) {
	args := m.Called(ctx, partnerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]any), args.Error(1)
}

func (m *MockCouponRepository) GetPartnerSalesStatistics(ctx context.Context, partnerID uuid.UUID) (map[string]any, error) {
	args := m.Called(ctx, partnerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]any), args.Error(1)
}

func (m *MockCouponRepository) GetPartnerUsageStatistics(ctx context.Context, partnerID uuid.UUID) (map[string]any, error) {
	args := m.Called(ctx, partnerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]any), args.Error(1)
}

func (m *MockCouponRepository) GetStatusCounts(ctx context.Context, partnerID *uuid.UUID) (map[string]int64, error) {
	args := m.Called(ctx, partnerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]int64), args.Error(1)
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

func (m *MockCouponRepository) GetTopActivatedByPartner(ctx context.Context, limit int) ([]PartnerCount, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]PartnerCount), args.Error(1)
}

func (m *MockCouponRepository) GetTopPurchasedByPartner(ctx context.Context, limit int) ([]PartnerCount, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]PartnerCount), args.Error(1)
}

func (m *MockCouponRepository) GetLastActivityByPartner(ctx context.Context, partnerID uuid.UUID) (*time.Time, error) {
	args := m.Called(ctx, partnerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*time.Time), args.Error(1)
}

func (m *MockCouponRepository) GetTimeSeriesData(ctx context.Context, from, to time.Time, period string, partnerID *uuid.UUID) ([]map[string]any, error) {
	args := m.Called(ctx, from, to, period, partnerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]map[string]any), args.Error(1)
}

func (m *MockCouponRepository) GetRecentActivated(ctx context.Context, limit int) ([]*Coupon, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Coupon), args.Error(1)
}

func (m *MockCouponRepository) CodeExists(ctx context.Context, code string) (bool, error) {
	args := m.Called(ctx, code)
	return args.Bool(0), args.Error(1)
}

func (m *MockCouponRepository) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockCouponRepository) GetCouponsWithAdvancedFilter(ctx context.Context, filter CouponFilterRequest) ([]*CouponInfo, int, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*CouponInfo), args.Int(1), args.Error(2)
}

func (m *MockCouponRepository) GetCouponsForDeletion(ctx context.Context, ids []uuid.UUID) ([]*CouponDeletePreview, error) {
	args := m.Called(ctx, ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*CouponDeletePreview), args.Error(1)
}

func (m *MockCouponRepository) GetCouponsForExport(ctx context.Context, options ExportOptionsRequest) (any, error) {
	args := m.Called(ctx, options)
	return args.Get(0), args.Error(1)
}

func (m *MockCouponRepository) FindAvailableCoupon(ctx context.Context, size, style string, partnerID *uuid.UUID) (*Coupon, error) {
	args := m.Called(ctx, size, style, partnerID)
	return args.Get(0).(*Coupon), args.Error(1)
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

func (m *MockRedisClient) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	args := m.Called(ctx, keys)
	return args.Get(0).(*redis.IntCmd)
}

func TestCouponService_GetCouponByCode(t *testing.T) {
	tests := []struct {
		name           string
		code           string
		mockSetup      func(*MockCouponRepository, *MockRedisClient)
		expectedResult *Coupon
		expectedError  bool
	}{
		{
			name: "successful_get_coupon",
			code: "123456789012",
			mockSetup: func(repo *MockCouponRepository, redis *MockRedisClient) {
				coupon := &Coupon{
					ID:   uuid.New(),
					Code: "123456789012",
					Size: "40x50",
				}
				repo.On("GetByCode", mock.Anything, "123456789012").Return(coupon, nil)
			},
			expectedResult: &Coupon{
				Code: "123456789012",
				Size: "40x50",
			},
			expectedError: false,
		},
		{
			name: "coupon_not_found",
			code: "nonexistent",
			mockSetup: func(repo *MockCouponRepository, redis *MockRedisClient) {
				repo.On("GetByCode", mock.Anything, "nonexistent").Return(nil, errors.New("not found"))
			},
			expectedResult: nil,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockCouponRepository)
			mockRedis := new(MockRedisClient)

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo, mockRedis)
			}

			deps := &CouponServiceDeps{
				CouponRepository: mockRepo,
				RedisClient:      mockRedis,
			}
			service := NewCouponService(deps)

			result, err := service.GetCouponByCode(tt.code)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedResult.Code, result.Code)
				assert.Equal(t, tt.expectedResult.Size, result.Size)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestCouponService_ActivateCoupon(t *testing.T) {
	tests := []struct {
		name          string
		couponID      uuid.UUID
		req           ActivateCouponRequest
		mockSetup     func(*MockCouponRepository, *MockRedisClient)
		expectedError bool
	}{
		{
			name:     "successful_activation",
			couponID: uuid.New(),
			req: ActivateCouponRequest{
				ZipURL: stringPtr("http://example.com/materials.zip"),
			},
			mockSetup: func(repo *MockCouponRepository, redis *MockRedisClient) {
				repo.On("ActivateCoupon", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("coupon.ActivateCouponRequest")).Return(nil)
			},
			expectedError: false,
		},
		{
			name:     "repository_error",
			couponID: uuid.New(),
			req: ActivateCouponRequest{
				ZipURL: stringPtr("http://example.com/materials.zip"),
			},
			mockSetup: func(repo *MockCouponRepository, redis *MockRedisClient) {
				repo.On("ActivateCoupon", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("coupon.ActivateCouponRequest")).Return(errors.New("database error"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockCouponRepository)
			mockRedis := new(MockRedisClient)

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo, mockRedis)
			}

			deps := &CouponServiceDeps{
				CouponRepository: mockRepo,
				RedisClient:      mockRedis,
			}
			service := NewCouponService(deps)

			err := service.ActivateCoupon(tt.couponID, tt.req)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestCouponService_GetStatistics(t *testing.T) {
	tests := []struct {
		name          string
		partnerID     *uuid.UUID
		mockSetup     func(*MockCouponRepository, *MockRedisClient)
		expectedError bool
	}{
		{
			name:      "successful_statistics",
			partnerID: nil,
			mockSetup: func(repo *MockCouponRepository, redis *MockRedisClient) {
				stats := map[string]int64{
					"total":  100,
					"active": 80,
					"used":   20,
				}
				repo.On("GetStatistics", mock.Anything, (*uuid.UUID)(nil)).Return(stats, nil)
			},
			expectedError: false,
		},
		{
			name:      "repository_error",
			partnerID: nil,
			mockSetup: func(repo *MockCouponRepository, redis *MockRedisClient) {
				repo.On("GetStatistics", mock.Anything, (*uuid.UUID)(nil)).Return(nil, errors.New("database error"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockCouponRepository)
			mockRedis := new(MockRedisClient)

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo, mockRedis)
			}

			deps := &CouponServiceDeps{
				CouponRepository: mockRepo,
				RedisClient:      mockRedis,
			}
			service := NewCouponService(deps)

			result, err := service.GetStatistics(tt.partnerID)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestCouponService_SearchCoupons(t *testing.T) {
	tests := []struct {
		name          string
		code          string
		status        string
		size          string
		style         string
		partnerID     *uuid.UUID
		mockSetup     func(*MockCouponRepository, *MockRedisClient)
		expectedError bool
	}{
		{
			name:      "successful_search",
			code:      "123",
			status:    "active",
			size:      "40x50",
			style:     "diamond",
			partnerID: nil,
			mockSetup: func(repo *MockCouponRepository, redis *MockRedisClient) {
				coupons := []*Coupon{
					{ID: uuid.New(), Code: "123456789012", Size: "40x50", Style: "diamond", Status: "active"},
				}
				repo.On("Search", mock.Anything, "123", "active", "40x50", "diamond", (*uuid.UUID)(nil)).Return(coupons, nil)
			},
			expectedError: false,
		},
		{
			name:      "repository_error",
			code:      "123",
			status:    "active",
			size:      "40x50",
			style:     "diamond",
			partnerID: nil,
			mockSetup: func(repo *MockCouponRepository, redis *MockRedisClient) {
				repo.On("Search", mock.Anything, "123", "active", "40x50", "diamond", (*uuid.UUID)(nil)).Return(nil, errors.New("database error"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockCouponRepository)
			mockRedis := new(MockRedisClient)

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo, mockRedis)
			}

			deps := &CouponServiceDeps{
				CouponRepository: mockRepo,
				RedisClient:      mockRedis,
			}
			service := NewCouponService(deps)

			result, err := service.SearchCoupons(tt.code, tt.status, tt.size, tt.style, tt.partnerID)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestCouponService_GetCouponByID(t *testing.T) {
	tests := []struct {
		name          string
		couponID      uuid.UUID
		mockSetup     func(*MockCouponRepository, *MockRedisClient)
		expectedError bool
	}{
		{
			name:     "successful_get_by_id",
			couponID: uuid.New(),
			mockSetup: func(repo *MockCouponRepository, redis *MockRedisClient) {
				coupon := &Coupon{
					ID:     uuid.New(),
					Code:   "123456789012",
					Size:   "40x50",
					Status: "active",
				}
				repo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(coupon, nil)
			},
			expectedError: false,
		},
		{
			name:     "coupon_not_found",
			couponID: uuid.New(),
			mockSetup: func(repo *MockCouponRepository, redis *MockRedisClient) {
				repo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil, errors.New("not found"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockCouponRepository)
			mockRedis := new(MockRedisClient)

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo, mockRedis)
			}

			deps := &CouponServiceDeps{
				CouponRepository: mockRepo,
				RedisClient:      mockRedis,
			}
			service := NewCouponService(deps)

			result, err := service.GetCouponByID(tt.couponID)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestCouponService_ResetCoupon(t *testing.T) {
	tests := []struct {
		name          string
		couponID      uuid.UUID
		mockSetup     func(*MockCouponRepository, *MockRedisClient)
		expectedError bool
	}{
		{
			name:     "successful_reset",
			couponID: uuid.New(),
			mockSetup: func(repo *MockCouponRepository, redis *MockRedisClient) {
				repo.On("ResetCoupon", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil)
			},
			expectedError: false,
		},
		{
			name:     "repository_error",
			couponID: uuid.New(),
			mockSetup: func(repo *MockCouponRepository, redis *MockRedisClient) {
				repo.On("ResetCoupon", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(errors.New("database error"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockCouponRepository)
			mockRedis := new(MockRedisClient)

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo, mockRedis)
			}

			deps := &CouponServiceDeps{
				CouponRepository: mockRepo,
				RedisClient:      mockRedis,
			}
			service := NewCouponService(deps)

			err := service.ResetCoupon(tt.couponID)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestCouponService_SendSchema(t *testing.T) {
	tests := []struct {
		name          string
		couponID      uuid.UUID
		email         string
		mockSetup     func(*MockCouponRepository, *MockRedisClient)
		expectedError bool
	}{
		{
			name:     "successful_send_schema",
			couponID: uuid.New(),
			email:    "test@example.com",
			mockSetup: func(repo *MockCouponRepository, redis *MockRedisClient) {
				repo.On("SendSchema", mock.Anything, mock.AnythingOfType("uuid.UUID"), "test@example.com").Return(nil)
			},
			expectedError: false,
		},
		{
			name:     "repository_error",
			couponID: uuid.New(),
			email:    "test@example.com",
			mockSetup: func(repo *MockCouponRepository, redis *MockRedisClient) {
				repo.On("SendSchema", mock.Anything, mock.AnythingOfType("uuid.UUID"), "test@example.com").Return(errors.New("database error"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockCouponRepository)
			mockRedis := new(MockRedisClient)

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo, mockRedis)
			}

			deps := &CouponServiceDeps{
				CouponRepository: mockRepo,
				RedisClient:      mockRedis,
			}
			service := NewCouponService(deps)

			err := service.SendSchema(tt.couponID, tt.email)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestCouponService_MarkAsPurchased(t *testing.T) {
	tests := []struct {
		name          string
		couponID      uuid.UUID
		email         string
		mockSetup     func(*MockCouponRepository, *MockRedisClient)
		expectedError bool
	}{
		{
			name:     "successful_mark_as_purchased",
			couponID: uuid.New(),
			email:    "buyer@example.com",
			mockSetup: func(repo *MockCouponRepository, redis *MockRedisClient) {
				repo.On("MarkAsPurchased", mock.Anything, mock.AnythingOfType("uuid.UUID"), "buyer@example.com").Return(nil)
			},
			expectedError: false,
		},
		{
			name:     "repository_error",
			couponID: uuid.New(),
			email:    "buyer@example.com",
			mockSetup: func(repo *MockCouponRepository, redis *MockRedisClient) {
				repo.On("MarkAsPurchased", mock.Anything, mock.AnythingOfType("uuid.UUID"), "buyer@example.com").Return(errors.New("database error"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockCouponRepository)
			mockRedis := new(MockRedisClient)

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo, mockRedis)
			}

			deps := &CouponServiceDeps{
				CouponRepository: mockRepo,
				RedisClient:      mockRedis,
			}
			service := NewCouponService(deps)

			err := service.MarkAsPurchased(tt.couponID, tt.email)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestCouponService_ValidateCoupon(t *testing.T) {
	tests := []struct {
		name          string
		code          string
		mockSetup     func(*MockCouponRepository, *MockRedisClient, *MockPartnerRepository)
		expectedError bool
		expectedValid bool
	}{
		{
			name: "valid_coupon",
			code: "123456789012",
			mockSetup: func(repo *MockCouponRepository, redis *MockRedisClient, partnerRepo *MockPartnerRepository) {
				coupon := &Coupon{
					ID:        uuid.New(),
					Code:      "123456789012",
					Status:    "active",
					Size:      "40x50",
					Style:     "diamond",
					PartnerID: uuid.New(),
				}
				partner := &Partner{
					ID:          coupon.PartnerID,
					PartnerCode: "TEST_PARTNER",
					Domain:      "example.com",
					BrandName:   "Test Brand",
				}
				repo.On("GetByCode", mock.Anything, "123456789012").Return(coupon, nil)
				partnerRepo.On("GetByID", mock.Anything, coupon.PartnerID).Return(partner, nil)
			},
			expectedError: false,
			expectedValid: true,
		},
		{
			name: "invalid_coupon_not_found",
			code: "nonexistent",
			mockSetup: func(repo *MockCouponRepository, redis *MockRedisClient, partnerRepo *MockPartnerRepository) {
				repo.On("GetByCode", mock.Anything, "nonexistent").Return(nil, errors.New("not found"))
			},
			expectedError: false,
			expectedValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockCouponRepository)
			mockRedis := new(MockRedisClient)
			mockPartnerRepo := new(MockPartnerRepository)

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo, mockRedis, mockPartnerRepo)
			}

			deps := &CouponServiceDeps{
				CouponRepository:  mockRepo,
				RedisClient:       mockRedis,
				PartnerRepository: mockPartnerRepo,
			}
			service := NewCouponService(deps)

			result, err := service.ValidateCoupon(tt.code)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedValid, result.Valid)
			}

			mockRepo.AssertExpectations(t)
			mockPartnerRepo.AssertExpectations(t)
		})
	}
}

func TestCouponService_SearchCouponsWithPagination(t *testing.T) {
	tests := []struct {
		name          string
		code          string
		status        string
		size          string
		style         string
		partnerID     *uuid.UUID
		page          int
		limit         int
		mockSetup     func(*MockCouponRepository, *MockRedisClient)
		expectedError bool
	}{
		{
			name:      "successful_search_with_pagination",
			code:      "123",
			status:    "active",
			size:      "40x50",
			style:     "diamond",
			partnerID: nil,
			page:      1,
			limit:     10,
			mockSetup: func(repo *MockCouponRepository, redis *MockRedisClient) {
				coupons := []*Coupon{
					{ID: uuid.New(), Code: "123456789012", Size: "40x50", Style: "diamond", Status: "active"},
				}
				repo.On("SearchWithPagination", mock.Anything, "123", "active", "40x50", "diamond", (*uuid.UUID)(nil), 1, 10, (*time.Time)(nil), (*time.Time)(nil), (*time.Time)(nil), (*time.Time)(nil), "", "").Return(coupons, 1, nil)
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockCouponRepository)
			mockRedis := new(MockRedisClient)

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo, mockRedis)
			}

			deps := &CouponServiceDeps{
				CouponRepository: mockRepo,
				RedisClient:      mockRedis,
			}
			service := NewCouponService(deps)

			result, total, err := service.SearchCouponsWithPagination(tt.code, tt.status, tt.size, tt.style, tt.partnerID, tt.page, tt.limit)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Equal(t, int64(0), total)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Greater(t, total, int64(0))
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestCouponService_SearchCouponsByPartner(t *testing.T) {
	tests := []struct {
		name          string
		partnerID     uuid.UUID
		status        string
		size          string
		style         string
		mockSetup     func(*MockCouponRepository, *MockRedisClient)
		expectedError bool
	}{
		{
			name:      "successful_search_by_partner",
			partnerID: uuid.New(),
			status:    "active",
			size:      "40x50",
			style:     "diamond",
			mockSetup: func(repo *MockCouponRepository, redis *MockRedisClient) {
				coupons := []*Coupon{
					{ID: uuid.New(), Code: "123456789012", Size: "40x50", Style: "diamond", Status: "active"},
				}
				repo.On("Search", mock.Anything, "", "active", "40x50", "diamond", mock.AnythingOfType("*uuid.UUID")).Return(coupons, nil)
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockCouponRepository)
			mockRedis := new(MockRedisClient)

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo, mockRedis)
			}

			deps := &CouponServiceDeps{
				CouponRepository: mockRepo,
				RedisClient:      mockRedis,
			}
			service := NewCouponService(deps)

			result, err := service.SearchCouponsByPartner(tt.partnerID, tt.status, tt.size, tt.style)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestCouponService_BatchResetCoupons(t *testing.T) {
	tests := []struct {
		name          string
		couponIDs     []string
		mockSetup     func(*MockCouponRepository, *MockRedisClient)
		expectedError bool
	}{
		{
			name:      "successful_batch_reset",
			couponIDs: []string{uuid.New().String(), uuid.New().String()},
			mockSetup: func(repo *MockCouponRepository, redis *MockRedisClient) {
				successIDs := []uuid.UUID{uuid.New(), uuid.New()}
				failedIDs := []uuid.UUID{}
				repo.On("BatchReset", mock.Anything, mock.AnythingOfType("[]uuid.UUID")).Return(successIDs, failedIDs, nil)
			},
			expectedError: false,
		},
		{
			name:          "invalid_uuid",
			couponIDs:     []string{"invalid-uuid"},
			mockSetup:     func(repo *MockCouponRepository, redis *MockRedisClient) {},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockCouponRepository)
			mockRedis := new(MockRedisClient)

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo, mockRedis)
			}

			deps := &CouponServiceDeps{
				CouponRepository: mockRepo,
				RedisClient:      mockRedis,
			}
			service := NewCouponService(deps)

			result, err := service.BatchResetCoupons(tt.couponIDs)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.name == "invalid_uuid" {
					assert.Greater(t, result.FailedCount, 0)
				}
			}

			if tt.name != "invalid_uuid" {
				mockRepo.AssertExpectations(t)
			}
		})
	}
}
