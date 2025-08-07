package stats

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/skr1ms/mosaic/internal/partner"
)

func BenchmarkStatsService_GetGeneralStats_Success(b *testing.B) {
	service, mockCouponRepo, mockPartnerRepo, mockRedisClient := createTestStatsService()
	ctx := context.Background()

	mockRedisClient.On("Get", ctx, "general_stats").Return("", nil)

	mockCouponRepo.On("CountTotal", ctx).Return(int64(1000), nil)
	mockCouponRepo.On("CountActivated", ctx).Return(int64(750), nil)
	mockPartnerRepo.On("CountActive", ctx).Return(int64(50), nil)
	mockPartnerRepo.On("CountTotal", ctx).Return(int64(60), nil)

	mockRedisClient.On("Set", ctx, "general_stats", mock.AnythingOfType("[]uint8"), 5*time.Minute).Return(nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := service.GetGeneralStats(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStatsService_GetGeneralStats_FromCache(b *testing.B) {
	service, _, _, mockRedisClient := createTestStatsService()
	ctx := context.Background()

	cachedData := `{"total_coupons_created":1000,"total_coupons_activated":750,"activation_rate":75.0,"active_partners_count":50,"total_partners_count":60,"last_updated":"2023-01-01T00:00:00Z"}`
	mockRedisClient.On("Get", ctx, "general_stats").Return(cachedData, nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := service.GetGeneralStats(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStatsService_GetPartnerStats_Success(b *testing.B) {
	service, mockCouponRepo, mockPartnerRepo, mockRedisClient := createTestStatsService()
	ctx := context.Background()

	testPartner := createTestPartner()
	cacheKey := "partner_stats:" + testPartner.ID.String()

	mockRedisClient.On("Get", ctx, cacheKey).Return("", nil)
	mockPartnerRepo.On("GetByID", ctx, testPartner.ID).Return(testPartner, nil)
	mockCouponRepo.On("CountByPartner", ctx, testPartner.ID).Return(int64(100), nil)
	mockCouponRepo.On("CountActivatedByPartner", ctx, testPartner.ID).Return(int64(80), nil)
	mockCouponRepo.On("CountBrandedPurchasesByPartner", ctx, testPartner.ID).Return(int64(20), nil)

	lastActivity := time.Now()
	mockCouponRepo.On("GetLastActivityByPartner", ctx, testPartner.ID).Return(&lastActivity, nil)
	mockRedisClient.On("Set", ctx, cacheKey, mock.AnythingOfType("[]uint8"), 10*time.Minute).Return(nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := service.GetPartnerStats(ctx, testPartner.ID)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStatsService_GetAllPartnersStats(b *testing.B) {
	service, mockCouponRepo, mockPartnerRepo, mockRedisClient := createTestStatsService()
	ctx := context.Background()

	numPartners := 10
	partners := make([]*partner.Partner, numPartners)
	for i := 0; i < numPartners; i++ {
		partners[i] = createTestPartner()
	}

	mockPartnerRepo.On("GetAll", ctx, "created_at", "desc").Return(partners, nil)

	for _, p := range partners {
		cacheKey := "partner_stats:" + p.ID.String()
		mockRedisClient.On("Get", ctx, cacheKey).Return("", nil)
		mockPartnerRepo.On("GetByID", ctx, p.ID).Return(p, nil)
		mockCouponRepo.On("CountByPartner", ctx, p.ID).Return(int64(50), nil)
		mockCouponRepo.On("CountActivatedByPartner", ctx, p.ID).Return(int64(40), nil)
		mockCouponRepo.On("CountBrandedPurchasesByPartner", ctx, p.ID).Return(int64(10), nil)
		mockCouponRepo.On("GetLastActivityByPartner", ctx, p.ID).Return(nil, nil)
		mockRedisClient.On("Set", ctx, cacheKey, mock.AnythingOfType("[]uint8"), 10*time.Minute).Return(nil)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := service.GetAllPartnersStats(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStatsService_GetTimeSeriesStats(b *testing.B) {
	service, mockCouponRepo, _, _ := createTestStatsService()
	ctx := context.Background()

	dateFrom := "2023-01-01"
	dateTo := "2023-01-07"
	period := "day"

	filters := &StatsFiltersRequest{
		DateFrom: &dateFrom,
		DateTo:   &dateTo,
		Period:   &period,
	}

	numDays := 7
	rawData := make([]map[string]interface{}, numDays)
	for i := 0; i < numDays; i++ {
		rawData[i] = map[string]interface{}{
			"date":               "2023-01-" + string(rune('0'+1+i)),
			"coupons_created":    int64(100 + i*10),
			"coupons_activated":  int64(80 + i*8),
			"coupons_purchased":  int64(50 + i*5),
			"new_partners_count": int64(1 + i%2),
		}
	}

	mockCouponRepo.On("GetTimeSeriesData", ctx, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time"), period, (*uuid.UUID)(nil)).Return(rawData, nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := service.GetTimeSeriesStats(ctx, filters)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStatsService_GetSystemHealth(b *testing.B) {
	service, mockCouponRepo, _, mockRedisClient := createTestStatsService()
	ctx := context.Background()

	mockRedisClient.On("Get", ctx, "system_health").Return("", nil)
	mockCouponRepo.On("HealthCheck", ctx).Return(nil)
	mockRedisClient.On("Ping", ctx).Return(nil)
	mockRedisClient.On("LLen", ctx, "image_processing_queue").Return(int64(5), nil)
	mockRedisClient.On("Set", ctx, "system_health", mock.AnythingOfType("[]uint8"), 1*time.Minute).Return(nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := service.GetSystemHealth(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStatsService_GetCouponsByStatus(b *testing.B) {
	service, mockCouponRepo, _, _ := createTestStatsService()
	ctx := context.Background()

	statusCounts := map[string]int64{
		"new":       100,
		"activated": 80,
		"used":      60,
		"completed": 40,
	}

	mockCouponRepo.On("GetExtendedStatusCounts", ctx, (*uuid.UUID)(nil)).Return(statusCounts, nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := service.GetCouponsByStatus(ctx, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStatsService_GetTopPartners(b *testing.B) {
	service, mockCouponRepo, mockPartnerRepo, _ := createTestStatsService()
	ctx := context.Background()

	numPartners := 100
	partners := make([]*partner.Partner, numPartners)
	for i := 0; i < numPartners; i++ {
		partners[i] = createTestPartner()
	}

	mockPartnerRepo.On("GetAll", ctx, "created_at", "desc").Return(partners, nil)

	for i, p := range partners {
		mockCouponRepo.On("CountByPartner", ctx, p.ID).Return(int64(100-i), nil)
		mockCouponRepo.On("CountActivatedByPartner", ctx, p.ID).Return(int64(80-i), nil)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := service.GetTopPartners(ctx, 10)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStatsService_GetRealTimeStats(b *testing.B) {
	service, mockCouponRepo, _, _ := createTestStatsService()
	ctx := context.Background()

	mockCouponRepo.On("CountActivatedInTimeRange", ctx, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).Return(int64(5), nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := service.GetRealTimeStats(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStatsService_ConcurrentGetGeneralStats(b *testing.B) {
	service, mockCouponRepo, mockPartnerRepo, mockRedisClient := createTestStatsService()
	ctx := context.Background()

	mockRedisClient.On("Get", ctx, "general_stats").Return("", nil)

	mockCouponRepo.On("CountTotal", ctx).Return(int64(1000), nil)
	mockCouponRepo.On("CountActivated", ctx).Return(int64(750), nil)
	mockPartnerRepo.On("CountActive", ctx).Return(int64(50), nil)
	mockPartnerRepo.On("CountTotal", ctx).Return(int64(60), nil)

	mockRedisClient.On("Set", ctx, "general_stats", mock.AnythingOfType("[]uint8"), 5*time.Minute).Return(nil)

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := service.GetGeneralStats(ctx)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkStatsService_MixedOperations(b *testing.B) {
	service, mockCouponRepo, mockPartnerRepo, mockRedisClient := createTestStatsService()
	ctx := context.Background()

	testPartner := createTestPartner()

	mockRedisClient.On("Get", ctx, "general_stats").Return("", nil)
	mockCouponRepo.On("CountTotal", ctx).Return(int64(1000), nil)
	mockCouponRepo.On("CountActivated", ctx).Return(int64(750), nil)
	mockPartnerRepo.On("CountActive", ctx).Return(int64(50), nil)
	mockPartnerRepo.On("CountTotal", ctx).Return(int64(60), nil)
	mockRedisClient.On("Set", ctx, "general_stats", mock.AnythingOfType("[]uint8"), 5*time.Minute).Return(nil)

	cacheKey := "partner_stats:" + testPartner.ID.String()
	mockRedisClient.On("Get", ctx, cacheKey).Return("", nil)
	mockPartnerRepo.On("GetByID", ctx, testPartner.ID).Return(testPartner, nil)
	mockCouponRepo.On("CountByPartner", ctx, testPartner.ID).Return(int64(100), nil)
	mockCouponRepo.On("CountActivatedByPartner", ctx, testPartner.ID).Return(int64(80), nil)
	mockCouponRepo.On("CountBrandedPurchasesByPartner", ctx, testPartner.ID).Return(int64(20), nil)
	mockCouponRepo.On("GetLastActivityByPartner", ctx, testPartner.ID).Return(nil, nil)
	mockRedisClient.On("Set", ctx, cacheKey, mock.AnythingOfType("[]uint8"), 10*time.Minute).Return(nil)

	statusCounts := map[string]int64{
		"new":       100,
		"activated": 80,
		"used":      60,
		"completed": 40,
	}
	mockCouponRepo.On("GetExtendedStatusCounts", ctx, (*uuid.UUID)(nil)).Return(statusCounts, nil)

	mockCouponRepo.On("CountActivatedInTimeRange", ctx, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).Return(int64(5), nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		switch i % 4 {
		case 0:
			_, err := service.GetGeneralStats(ctx)
			if err != nil {
				b.Fatal(err)
			}
		case 1:
			_, err := service.GetPartnerStats(ctx, testPartner.ID)
			if err != nil {
				b.Fatal(err)
			}
		case 2:
			_, err := service.GetCouponsByStatus(ctx, nil)
			if err != nil {
				b.Fatal(err)
			}
		case 3:
			_, err := service.GetRealTimeStats(ctx)
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}

func BenchmarkStatsService_MemoryAllocation_GetGeneralStats(b *testing.B) {
	service, mockCouponRepo, mockPartnerRepo, mockRedisClient := createTestStatsService()
	ctx := context.Background()

	mockRedisClient.On("Get", ctx, "general_stats").Return("", nil)
	mockCouponRepo.On("CountTotal", ctx).Return(int64(1000), nil)
	mockCouponRepo.On("CountActivated", ctx).Return(int64(750), nil)
	mockPartnerRepo.On("CountActive", ctx).Return(int64(50), nil)
	mockPartnerRepo.On("CountTotal", ctx).Return(int64(60), nil)
	mockRedisClient.On("Set", ctx, "general_stats", mock.AnythingOfType("[]uint8"), 5*time.Minute).Return(nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		result, err := service.GetGeneralStats(ctx)
		if err != nil {
			b.Fatal(err)
		}
		_ = result.TotalCouponsCreated
	}
}

func BenchmarkStatsService_MemoryAllocation_GetTimeSeriesStats(b *testing.B) {
	service, mockCouponRepo, _, _ := createTestStatsService()
	ctx := context.Background()

	rawData := []map[string]interface{}{
		{
			"date":               "2023-01-01",
			"coupons_created":    int64(10),
			"coupons_activated":  int64(8),
			"coupons_purchased":  int64(5),
			"new_partners_count": int64(1),
		},
	}

	filters := &StatsFiltersRequest{
		Period: stringPtr("day"),
	}

	mockCouponRepo.On("GetTimeSeriesData", ctx, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time"), "day", (*uuid.UUID)(nil)).Return(rawData, nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		result, err := service.GetTimeSeriesStats(ctx, filters)
		if err != nil {
			b.Fatal(err)
		}
		_ = result.Data[0].Date
	}
}

func BenchmarkStatsService_LargeDataSet_GetTimeSeriesStats(b *testing.B) {
	service, mockCouponRepo, _, _ := createTestStatsService()
	ctx := context.Background()

	numDays := 365
	rawData := make([]map[string]interface{}, numDays)
	for i := 0; i < numDays; i++ {
		rawData[i] = map[string]interface{}{
			"date":               "2023-01-01",
			"coupons_created":    int64(100 + i),
			"coupons_activated":  int64(80 + i),
			"coupons_purchased":  int64(50 + i),
			"new_partners_count": int64(1 + i%10),
		}
	}

	filters := &StatsFiltersRequest{
		Period: stringPtr("day"),
	}

	mockCouponRepo.On("GetTimeSeriesData", ctx, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time"), "day", (*uuid.UUID)(nil)).Return(rawData, nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		result, err := service.GetTimeSeriesStats(ctx, filters)
		if err != nil {
			b.Fatal(err)
		}
		_ = len(result.Data)
	}
}

func stringPtr(s string) *string {
	return &s
}

func BenchmarkMetricsCollector_IncrementCouponsCreated(b *testing.B) {
	collector := NewMetricsCollector()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		collector.IncrementCouponsCreated("partner-1", "40x50", "grayscale")
	}
}

func BenchmarkMetricsCollector_ObserveImageProcessingDuration(b *testing.B) {
	collector := NewMetricsCollector()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		collector.ObserveImageProcessingDuration("upload", "success", 1.5)
	}
}

func BenchmarkMetricsCollector_SetSystemMetrics(b *testing.B) {
	collector := NewMetricsCollector()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		collector.SetSystemMetrics(1024*1024*100, 45.5)
	}
}

func BenchmarkMetricsCollector_ConcurrentOperations(b *testing.B) {
	collector := NewMetricsCollector()

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			switch i % 4 {
			case 0:
				collector.IncrementCouponsCreated("partner-1", "40x50", "grayscale")
			case 1:
				collector.IncrementCouponsActivated("partner-1", "40x50", "grayscale")
			case 2:
				collector.ObserveImageProcessingDuration("process", "success", 2.5)
			case 3:
				collector.SetActiveUsers(100)
			}
			i++
		}
	})
}
