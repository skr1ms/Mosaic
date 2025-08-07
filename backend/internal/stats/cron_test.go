package stats

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/skr1ms/mosaic/internal/partner"
)

func TestCronService_Start_Success(t *testing.T) {
	service, _, _, _ := createTestStatsService()
	cronService := NewCronService(service)

	err := cronService.Start()
	require.NoError(t, err)

	assert.NotNil(t, cronService.Cron)

	cronService.Stop()
}

func TestCronService_updateGeneralStats(t *testing.T) {
	service, mockCouponRepo, mockPartnerRepo, mockRedisClient := createTestStatsService()
	cronService := NewCronService(service)

	mockRedisClient.On("Get", mock.Anything, "general_stats").Return("", redis.Nil)

	mockCouponRepo.On("CountTotal", mock.Anything).Return(int64(1000), nil)
	mockCouponRepo.On("CountActivated", mock.Anything).Return(int64(750), nil)
	mockPartnerRepo.On("CountActive", mock.Anything).Return(int64(50), nil)
	mockPartnerRepo.On("CountTotal", mock.Anything).Return(int64(60), nil)

	mockRedisClient.On("Set", mock.Anything, "general_stats", mock.AnythingOfType("[]uint8"), 5*time.Minute).Return(nil)

	cronService.updateGeneralStats()

	mockCouponRepo.AssertExpectations(t)
	mockPartnerRepo.AssertExpectations(t)
	mockRedisClient.AssertExpectations(t)
}

func TestCronService_updatePartnersMetrics(t *testing.T) {
	service, _, mockPartnerRepo, _ := createTestStatsService()
	cronService := NewCronService(service)

	mockPartnerRepo.On("GetAll", mock.Anything, "created_at", "desc").Return([]*partner.Partner{}, nil)

	cronService.updatePartnersMetrics()

	mockPartnerRepo.AssertExpectations(t)
}

func TestCronService_updateSystemMetrics(t *testing.T) {
	service, mockCouponRepo, _, mockRedisClient := createTestStatsService()
	cronService := NewCronService(service)

	mockRedisClient.On("Get", mock.Anything, "system_health").Return("", redis.Nil)
	mockCouponRepo.On("HealthCheck", mock.Anything).Return(nil)
	mockRedisClient.On("Ping", mock.Anything).Return(nil)
	mockRedisClient.On("LLen", mock.Anything, "image_processing_queue").Return(int64(5), nil)
	mockRedisClient.On("Set", mock.Anything, "system_health", mock.AnythingOfType("[]uint8"), 1*time.Minute).Return(nil)

	mockCouponRepo.On("CountActivatedInTimeRange", mock.Anything, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).Return(int64(5), nil)

	cronService.updateSystemMetrics()

	mockCouponRepo.AssertExpectations(t)
	mockRedisClient.AssertExpectations(t)
}

func TestCronService_clearStatsCache(t *testing.T) {
	service, mockCouponRepo, mockPartnerRepo, mockRedisClient := createTestStatsService()
	cronService := NewCronService(service)

	mockRedisClient.On("Get", mock.Anything, "general_stats").Return("", redis.Nil)
	mockCouponRepo.On("CountTotal", mock.Anything).Return(int64(1000), nil)
	mockCouponRepo.On("CountActivated", mock.Anything).Return(int64(750), nil)
	mockPartnerRepo.On("CountActive", mock.Anything).Return(int64(50), nil)
	mockPartnerRepo.On("CountTotal", mock.Anything).Return(int64(60), nil)
	mockRedisClient.On("Set", mock.Anything, "general_stats", mock.AnythingOfType("[]uint8"), 5*time.Minute).Return(nil)

	cronService.clearStatsCache()

	mockCouponRepo.AssertExpectations(t)
	mockPartnerRepo.AssertExpectations(t)
	mockRedisClient.AssertExpectations(t)
}

func TestCronService_aggregateDailyStats(t *testing.T) {
	service, mockCouponRepo, _, _ := createTestStatsService()
	cronService := NewCronService(service)

	rawData := []map[string]interface{}{
		{
			"date":               "2023-01-01",
			"coupons_created":    int64(10),
			"coupons_activated":  int64(8),
			"coupons_purchased":  int64(5),
			"new_partners_count": int64(1),
		},
	}
	mockCouponRepo.On("GetTimeSeriesData", mock.Anything, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time"), "day", (*uuid.UUID)(nil)).Return(rawData, nil)

	cronService.aggregateDailyStats()

	mockCouponRepo.AssertExpectations(t)
}

func TestNewCronService(t *testing.T) {
	service, _, _, _ := createTestStatsService()
	cronService := NewCronService(service)

	assert.NotNil(t, cronService)
	assert.NotNil(t, cronService.Cron)
	assert.NotNil(t, cronService.StatsService)
	assert.NotNil(t, cronService.MetricsCollector)
}

func TestCronService_Stop(t *testing.T) {
	service, _, _, _ := createTestStatsService()
	cronService := NewCronService(service)

	err := cronService.Start()
	require.NoError(t, err)

	assert.NotPanics(t, func() {
		cronService.Stop()
	})
}

func TestMetricsCollector_Implementation(t *testing.T) {
	collector := NewMetricsCollector()

	assert.NotPanics(t, func() {
		collector.IncrementCouponsCreated("partner-1", "40x50", "grayscale")
		collector.IncrementCouponsActivated("partner-1", "40x50", "grayscale")
		collector.IncrementCouponsPurchased("partner-1")
		collector.ObserveImageProcessingDuration("upload", "success", 1.5)
		collector.SetImageProcessingQueueSize(10)
		collector.SetPartnersCount(100, 80)
		collector.IncrementHTTPRequests("GET", "/stats", "200")
		collector.ObserveHTTPRequestDuration("GET", "/stats", 0.5)
		collector.SetDatabaseConnections(10)
		collector.SetRedisConnections(5)
		collector.IncrementErrors("database", "connection")
		collector.SetActiveUsers(25)
		collector.SetSystemMetrics(1024*1024*100, 45.5)
	})
}

func TestCronService_updateGeneralStats_Error(t *testing.T) {
	service, mockCouponRepo, _, mockRedisClient := createTestStatsService()
	cronService := NewCronService(service)

	mockRedisClient.On("Get", mock.Anything, "general_stats").Return("", nil)

	mockCouponRepo.On("CountTotal", mock.Anything).Return(int64(0), assert.AnError)

	assert.NotPanics(t, func() {
		cronService.updateGeneralStats()
	})

	mockCouponRepo.AssertExpectations(t)
	mockRedisClient.AssertExpectations(t)
}

func TestCronService_updateSystemMetrics_Error(t *testing.T) {
	service, mockCouponRepo, _, mockRedisClient := createTestStatsService()
	cronService := NewCronService(service)

	mockRedisClient.On("Get", mock.Anything, "system_health").Return("", redis.Nil)
	mockCouponRepo.On("HealthCheck", mock.Anything).Return(assert.AnError)
	mockRedisClient.On("Ping", mock.Anything).Return(nil)
	mockRedisClient.On("LLen", mock.Anything, "image_processing_queue").Return(int64(0), nil)
	mockRedisClient.On("Set", mock.Anything, "system_health", mock.AnythingOfType("[]uint8"), 1*time.Minute).Return(nil)
	mockCouponRepo.On("CountActivatedInTimeRange", mock.Anything, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).Return(int64(5), nil)

	assert.NotPanics(t, func() {
		cronService.updateSystemMetrics()
	})

	mockCouponRepo.AssertExpectations(t)
	mockRedisClient.AssertExpectations(t)
}
