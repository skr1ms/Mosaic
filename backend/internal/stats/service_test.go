package stats

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/skr1ms/mosaic/internal/partner"
)

type MockCouponRepository struct {
	mock.Mock
}

var _ CouponRepositoryInterface = (*MockCouponRepository)(nil)

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

func (m *MockCouponRepository) GetTimeSeriesData(ctx context.Context, from, to time.Time, period string, partnerID *uuid.UUID) ([]map[string]interface{}, error) {
	args := m.Called(ctx, from, to, period, partnerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}

func (m *MockCouponRepository) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

type MockPartnerRepository struct {
	mock.Mock
}

var _ PartnerRepositoryInterface = (*MockPartnerRepository)(nil)

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

type MockRedisClient struct {
	mock.Mock
}

var _ RedisClientInterface = (*MockRedisClient)(nil)

func (m *MockRedisClient) Get(ctx context.Context, key string) *redis.StringCmd {
	args := m.Called(ctx, key)
	cmd := redis.NewStringCmd(ctx, "get", key)

	if args.Error(1) != nil {
		cmd.SetErr(args.Error(1))
	} else {
		cmd.SetVal(args.String(0))
	}

	return cmd
}

func (m *MockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	args := m.Called(ctx, key, value, expiration)
	cmd := redis.NewStatusCmd(ctx, "set", key, value)

	if args.Error(0) != nil {
		cmd.SetErr(args.Error(0))
	} else {
		cmd.SetVal("OK")
	}

	return cmd
}

func (m *MockRedisClient) Ping(ctx context.Context) *redis.StatusCmd {
	args := m.Called(ctx)
	cmd := redis.NewStatusCmd(ctx, "ping")

	if args.Error(0) != nil {
		cmd.SetErr(args.Error(0))
	} else {
		cmd.SetVal("PONG")
	}

	return cmd
}

func (m *MockRedisClient) LLen(ctx context.Context, key string) *redis.IntCmd {
	args := m.Called(ctx, key)
	cmd := redis.NewIntCmd(ctx, "llen", key)

	if args.Error(1) != nil {
		cmd.SetErr(args.Error(1))
	} else {
		cmd.SetVal(args.Get(0).(int64))
	}

	return cmd
}

func createTestPartner() *partner.Partner {
	return &partner.Partner{
		ID:        uuid.New(),
		BrandName: "Test Partner",
		Domain:    "test.example.com",
		Email:     "test@example.com",
	}
}

func createTestStatsService() (*StatsService, *MockCouponRepository, *MockPartnerRepository, *MockRedisClient) {
	mockCouponRepo := &MockCouponRepository{}
	mockPartnerRepo := &MockPartnerRepository{}
	mockRedisClient := &MockRedisClient{}

	deps := &StatsServiceDeps{
		CouponRepository:  mockCouponRepo,
		PartnerRepository: mockPartnerRepo,
		RedisClient:       mockRedisClient,
	}

	service := NewStatsService(deps)
	return service, mockCouponRepo, mockPartnerRepo, mockRedisClient
}

func TestStatsService_GetGeneralStats_Success(t *testing.T) {
	service, mockCouponRepo, mockPartnerRepo, mockRedisClient := createTestStatsService()
	ctx := context.Background()

	mockRedisClient.On("Get", ctx, "general_stats").Return("", redis.Nil)

	mockCouponRepo.On("CountTotal", ctx).Return(int64(1000), nil)
	mockCouponRepo.On("CountActivated", ctx).Return(int64(750), nil)
	mockPartnerRepo.On("CountActive", ctx).Return(int64(50), nil)
	mockPartnerRepo.On("CountTotal", ctx).Return(int64(60), nil)

	mockRedisClient.On("Set", ctx, "general_stats", mock.AnythingOfType("[]uint8"), 5*time.Minute).Return(nil)

	result, err := service.GetGeneralStats(ctx)

	require.NoError(t, err)
	assert.Equal(t, int64(1000), result.TotalCouponsCreated)
	assert.Equal(t, int64(750), result.TotalCouponsActivated)
	assert.Equal(t, float64(75), result.ActivationRate)
	assert.Equal(t, int64(50), result.ActivePartnersCount)
	assert.Equal(t, int64(60), result.TotalPartnersCount)
	assert.NotEmpty(t, result.LastUpdated)

	mockCouponRepo.AssertExpectations(t)
	mockPartnerRepo.AssertExpectations(t)
	mockRedisClient.AssertExpectations(t)
}

func TestStatsService_GetGeneralStats_FromCache(t *testing.T) {
	service, _, _, mockRedisClient := createTestStatsService()
	ctx := context.Background()

	expectedStats := &GeneralStatsResponse{
		TotalCouponsCreated:   500,
		TotalCouponsActivated: 375,
		ActivationRate:        75.0,
		ActivePartnersCount:   25,
		TotalPartnersCount:    30,
		LastUpdated:           "2023-01-01T00:00:00Z",
	}

	cachedData, _ := json.Marshal(expectedStats)
	mockRedisClient.On("Get", ctx, "general_stats").Return(string(cachedData), nil)

	result, err := service.GetGeneralStats(ctx)

	require.NoError(t, err)
	assert.Equal(t, expectedStats.TotalCouponsCreated, result.TotalCouponsCreated)
	assert.Equal(t, expectedStats.ActivationRate, result.ActivationRate)

	mockRedisClient.AssertExpectations(t)
}

func TestStatsService_GetGeneralStats_DBError(t *testing.T) {
	service, mockCouponRepo, _, mockRedisClient := createTestStatsService()
	ctx := context.Background()

	mockRedisClient.On("Get", ctx, "general_stats").Return("", redis.Nil)
	mockCouponRepo.On("CountTotal", ctx).Return(int64(0), errors.New("database error"))

	result, err := service.GetGeneralStats(ctx)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get total coupons")

	mockCouponRepo.AssertExpectations(t)
	mockRedisClient.AssertExpectations(t)
}

func TestStatsService_GetPartnerStats_Success(t *testing.T) {
	service, mockCouponRepo, mockPartnerRepo, mockRedisClient := createTestStatsService()
	ctx := context.Background()

	testPartner := createTestPartner()
	cacheKey := "partner_stats:" + testPartner.ID.String()

	mockRedisClient.On("Get", ctx, cacheKey).Return("", redis.Nil)
	mockPartnerRepo.On("GetByID", ctx, testPartner.ID).Return(testPartner, nil)
	mockCouponRepo.On("CountByPartner", ctx, testPartner.ID).Return(int64(100), nil)
	mockCouponRepo.On("CountActivatedByPartner", ctx, testPartner.ID).Return(int64(80), nil)
	mockCouponRepo.On("CountBrandedPurchasesByPartner", ctx, testPartner.ID).Return(int64(20), nil)

	lastActivity := time.Now()
	mockCouponRepo.On("GetLastActivityByPartner", ctx, testPartner.ID).Return(&lastActivity, nil)
	mockRedisClient.On("Set", ctx, cacheKey, mock.AnythingOfType("[]uint8"), 10*time.Minute).Return(nil)

	result, err := service.GetPartnerStats(ctx, testPartner.ID)

	require.NoError(t, err)
	assert.Equal(t, testPartner.ID, result.PartnerID)
	assert.Equal(t, testPartner.BrandName, result.PartnerName)
	assert.Equal(t, int64(100), result.TotalCoupons)
	assert.Equal(t, int64(80), result.ActivatedCoupons)
	assert.Equal(t, int64(20), result.UnusedCoupons)
	assert.Equal(t, int64(20), result.BrandedSitePurchases)
	assert.Equal(t, float64(80), result.ActivationRate)
	assert.NotNil(t, result.LastActivity)

	mockCouponRepo.AssertExpectations(t)
	mockPartnerRepo.AssertExpectations(t)
	mockRedisClient.AssertExpectations(t)
}

func TestStatsService_GetPartnerStats_PartnerNotFound(t *testing.T) {
	service, _, mockPartnerRepo, mockRedisClient := createTestStatsService()
	ctx := context.Background()

	partnerID := uuid.New()
	cacheKey := "partner_stats:" + partnerID.String()

	mockRedisClient.On("Get", ctx, cacheKey).Return("", redis.Nil)
	mockPartnerRepo.On("GetByID", ctx, partnerID).Return(nil, errors.New("partner not found"))

	result, err := service.GetPartnerStats(ctx, partnerID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get partner")

	mockPartnerRepo.AssertExpectations(t)
	mockRedisClient.AssertExpectations(t)
}

func TestStatsService_GetAllPartnersStats_Success(t *testing.T) {
	service, mockCouponRepo, mockPartnerRepo, mockRedisClient := createTestStatsService()
	ctx := context.Background()

	partners := []*partner.Partner{createTestPartner(), createTestPartner()}
	mockPartnerRepo.On("GetAll", ctx, "created_at", "desc").Return(partners, nil)

	for _, p := range partners {
		cacheKey := "partner_stats:" + p.ID.String()
		mockRedisClient.On("Get", ctx, cacheKey).Return("", redis.Nil)
		mockPartnerRepo.On("GetByID", ctx, p.ID).Return(p, nil)
		mockCouponRepo.On("CountByPartner", ctx, p.ID).Return(int64(50), nil)
		mockCouponRepo.On("CountActivatedByPartner", ctx, p.ID).Return(int64(40), nil)
		mockCouponRepo.On("CountBrandedPurchasesByPartner", ctx, p.ID).Return(int64(10), nil)
		mockCouponRepo.On("GetLastActivityByPartner", ctx, p.ID).Return(nil, nil)
		mockRedisClient.On("Set", ctx, cacheKey, mock.AnythingOfType("[]uint8"), 10*time.Minute).Return(nil)
	}

	result, err := service.GetAllPartnersStats(ctx)

	require.NoError(t, err)
	assert.Len(t, result.Partners, 2)
	assert.Equal(t, int64(2), result.Total)

	mockPartnerRepo.AssertExpectations(t)
	mockCouponRepo.AssertExpectations(t)
	mockRedisClient.AssertExpectations(t)
}

func TestStatsService_GetTimeSeriesStats_Success(t *testing.T) {
	service, mockCouponRepo, _, _ := createTestStatsService()
	ctx := context.Background()

	dateFrom := "2023-01-01"
	dateTo := "2023-01-02"
	period := "day"

	filters := &StatsFiltersRequest{
		DateFrom: &dateFrom,
		DateTo:   &dateTo,
		Period:   &period,
	}

	rawData := []map[string]interface{}{
		{
			"date":               "2023-01-01",
			"coupons_created":    int64(10),
			"coupons_activated":  int64(8),
			"coupons_purchased":  int64(5),
			"new_partners_count": int64(1),
		},
		{
			"date":               "2023-01-02",
			"coupons_created":    int64(15),
			"coupons_activated":  int64(12),
			"coupons_purchased":  int64(7),
			"new_partners_count": int64(2),
		},
	}

	mockCouponRepo.On("GetTimeSeriesData", ctx, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time"), period, (*uuid.UUID)(nil)).Return(rawData, nil)

	result, err := service.GetTimeSeriesStats(ctx, filters)

	require.NoError(t, err)
	assert.Equal(t, period, result.Period)
	assert.Len(t, result.Data, 2)
	assert.Equal(t, "2023-01-01", result.Data[0].Date)
	assert.Equal(t, int64(10), result.Data[0].CouponsCreated)

	mockCouponRepo.AssertExpectations(t)
}

func TestStatsService_GetSystemHealth_Success(t *testing.T) {
	service, mockCouponRepo, _, mockRedisClient := createTestStatsService()
	ctx := context.Background()

	mockRedisClient.On("Get", ctx, "system_health").Return("", redis.Nil)
	mockCouponRepo.On("HealthCheck", ctx).Return(nil)
	mockRedisClient.On("Ping", ctx).Return(nil)
	mockRedisClient.On("LLen", ctx, "image_processing_queue").Return(int64(5), nil)
	mockRedisClient.On("Set", ctx, "system_health", mock.AnythingOfType("[]uint8"), 1*time.Minute).Return(nil)

	result, err := service.GetSystemHealth(ctx)

	require.NoError(t, err)
	assert.Equal(t, "healthy", result.Status)
	assert.Equal(t, "healthy", result.DatabaseStatus)
	assert.Equal(t, "healthy", result.RedisStatus)
	assert.Equal(t, int64(5), result.ImageProcessingQueue)
	assert.NotEmpty(t, result.LastUpdated)

	mockCouponRepo.AssertExpectations(t)
	mockRedisClient.AssertExpectations(t)
}

func TestStatsService_GetSystemHealth_UnhealthyDatabase(t *testing.T) {
	service, mockCouponRepo, _, mockRedisClient := createTestStatsService()
	ctx := context.Background()

	mockRedisClient.On("Get", ctx, "system_health").Return("", redis.Nil)
	mockCouponRepo.On("HealthCheck", ctx).Return(errors.New("db connection failed"))
	mockRedisClient.On("Ping", ctx).Return(nil)
	mockRedisClient.On("LLen", ctx, "image_processing_queue").Return(int64(0), nil)
	mockRedisClient.On("Set", ctx, "system_health", mock.AnythingOfType("[]uint8"), 1*time.Minute).Return(nil)

	result, err := service.GetSystemHealth(ctx)

	require.NoError(t, err)
	assert.Equal(t, "unhealthy", result.Status)
	assert.Equal(t, "unhealthy", result.DatabaseStatus)
	assert.Equal(t, "healthy", result.RedisStatus)

	mockCouponRepo.AssertExpectations(t)
	mockRedisClient.AssertExpectations(t)
}

func TestStatsService_GetCouponsByStatus_Success(t *testing.T) {
	service, mockCouponRepo, _, _ := createTestStatsService()
	ctx := context.Background()

	statusCounts := map[string]int64{
		"new":       100,
		"activated": 80,
		"used":      60,
		"completed": 40,
	}

	mockCouponRepo.On("GetExtendedStatusCounts", ctx, (*uuid.UUID)(nil)).Return(statusCounts, nil)

	result, err := service.GetCouponsByStatus(ctx, nil)

	require.NoError(t, err)
	assert.Equal(t, int64(100), result.New)
	assert.Equal(t, int64(80), result.Activated)
	assert.Equal(t, int64(60), result.Used)
	assert.Equal(t, int64(40), result.Completed)

	mockCouponRepo.AssertExpectations(t)
}

func TestStatsService_GetCouponsBySize_Success(t *testing.T) {
	service, mockCouponRepo, _, _ := createTestStatsService()
	ctx := context.Background()

	sizeCounts := map[string]int64{
		"21x30": 10,
		"30x40": 20,
		"40x40": 30,
		"40x50": 40,
		"40x60": 50,
		"50x70": 60,
	}

	mockCouponRepo.On("GetSizeCounts", ctx, (*uuid.UUID)(nil)).Return(sizeCounts, nil)

	result, err := service.GetCouponsBySize(ctx, nil)

	require.NoError(t, err)
	assert.Equal(t, int64(10), result.Size21x30)
	assert.Equal(t, int64(20), result.Size30x40)
	assert.Equal(t, int64(60), result.Size50x70)

	mockCouponRepo.AssertExpectations(t)
}

func TestStatsService_GetCouponsByStyle_Success(t *testing.T) {
	service, mockCouponRepo, _, _ := createTestStatsService()
	ctx := context.Background()

	styleCounts := map[string]int64{
		"grayscale":  25,
		"skin_tones": 30,
		"pop_art":    20,
		"max_colors": 25,
	}

	mockCouponRepo.On("GetStyleCounts", ctx, (*uuid.UUID)(nil)).Return(styleCounts, nil)

	result, err := service.GetCouponsByStyle(ctx, nil)

	require.NoError(t, err)
	assert.Equal(t, int64(25), result.Grayscale)
	assert.Equal(t, int64(30), result.SkinTones)
	assert.Equal(t, int64(20), result.PopArt)
	assert.Equal(t, int64(25), result.MaxColors)

	mockCouponRepo.AssertExpectations(t)
}

func TestStatsService_GetTopPartners_Success(t *testing.T) {
	service, mockCouponRepo, mockPartnerRepo, _ := createTestStatsService()
	ctx := context.Background()

	partners := []*partner.Partner{createTestPartner(), createTestPartner()}
	mockPartnerRepo.On("GetAll", ctx, "created_at", "desc").Return(partners, nil)

	for i, p := range partners {
		mockCouponRepo.On("CountByPartner", ctx, p.ID).Return(int64(100-i*50), nil)
		mockCouponRepo.On("CountActivatedByPartner", ctx, p.ID).Return(int64(80-i*40), nil)
	}

	result, err := service.GetTopPartners(ctx, 10)

	require.NoError(t, err)
	assert.Len(t, result.Partners, 2)
	assert.True(t, result.Partners[0].ActivatedCoupons >= result.Partners[1].ActivatedCoupons)

	mockPartnerRepo.AssertExpectations(t)
	mockCouponRepo.AssertExpectations(t)
}

func TestStatsService_GetRealTimeStats_Success(t *testing.T) {
	service, mockCouponRepo, _, _ := createTestStatsService()
	ctx := context.Background()

	fiveMinutesAgo := time.Now().Add(-5 * time.Minute)
	mockCouponRepo.On("CountActivatedInTimeRange", ctx, mock.MatchedBy(func(from time.Time) bool {
		return from.Sub(fiveMinutesAgo) < time.Minute
	}), mock.AnythingOfType("time.Time")).Return(int64(5), nil)

	result, err := service.GetRealTimeStats(ctx)

	require.NoError(t, err)
	assert.Equal(t, int64(23), result.ActiveUsers) // заглушка
	assert.Equal(t, int64(5), result.CouponsActivatedLast5Min)
	assert.Equal(t, int64(7), result.ImagesProcessingNow) // заглушка
	assert.Equal(t, 0.65, result.SystemLoad)              // заглушка
	assert.WithinDuration(t, time.Now(), result.Timestamp, time.Second)

	mockCouponRepo.AssertExpectations(t)
}

func TestStatsService_GetPartnerStats_ZeroCoupons(t *testing.T) {
	service, mockCouponRepo, mockPartnerRepo, mockRedisClient := createTestStatsService()
	ctx := context.Background()

	testPartner := createTestPartner()
	cacheKey := "partner_stats:" + testPartner.ID.String()

	mockRedisClient.On("Get", ctx, cacheKey).Return("", redis.Nil)
	mockPartnerRepo.On("GetByID", ctx, testPartner.ID).Return(testPartner, nil)
	mockCouponRepo.On("CountByPartner", ctx, testPartner.ID).Return(int64(0), nil)
	mockCouponRepo.On("CountActivatedByPartner", ctx, testPartner.ID).Return(int64(0), nil)
	mockCouponRepo.On("CountBrandedPurchasesByPartner", ctx, testPartner.ID).Return(int64(0), nil)
	mockCouponRepo.On("GetLastActivityByPartner", ctx, testPartner.ID).Return(nil, nil)
	mockRedisClient.On("Set", ctx, cacheKey, mock.AnythingOfType("[]uint8"), 10*time.Minute).Return(nil)

	result, err := service.GetPartnerStats(ctx, testPartner.ID)

	require.NoError(t, err)
	assert.Equal(t, int64(0), result.TotalCoupons)
	assert.Equal(t, int64(0), result.ActivatedCoupons)
	assert.Equal(t, float64(0), result.ActivationRate)
	assert.Nil(t, result.LastActivity)

	mockCouponRepo.AssertExpectations(t)
	mockPartnerRepo.AssertExpectations(t)
	mockRedisClient.AssertExpectations(t)
}
