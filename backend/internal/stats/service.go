package stats

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/skr1ms/mosaic/internal/coupon"
	"github.com/skr1ms/mosaic/internal/partner"
)

// StatsServiceDeps содержит зависимости для StatsService
type StatsServiceDeps struct {
	CouponRepository  *coupon.CouponRepository
	PartnerRepository *partner.PartnerRepository
	RedisClient       *redis.Client
}

// StatsService содержит бизнес-логику для статистики
type StatsService struct {
	deps *StatsServiceDeps
}

// NewStatsService создает новый экземпляр StatsService
func NewStatsService(deps *StatsServiceDeps) *StatsService {
	return &StatsService{deps: deps}
}

// GetGeneralStats возвращает общую статистику системы
func (s *StatsService) GetGeneralStats(ctx context.Context) (*GeneralStatsResponse, error) {
	// Пытаемся получить из кэша
	cacheKey := "general_stats"
	cached, err := s.deps.RedisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var stats GeneralStatsResponse
		if json.Unmarshal([]byte(cached), &stats) == nil {
			return &stats, nil
		}
	}

	// Получаем данные из БД
	totalCreated, err := s.deps.CouponRepository.CountTotal(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get total coupons: %w", err)
	}

	totalActivated, err := s.deps.CouponRepository.CountActivated(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get activated coupons: %w", err)
	}

	activePartners, err := s.deps.PartnerRepository.CountActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get active partners: %w", err)
	}

	totalPartners, err := s.deps.PartnerRepository.CountTotal(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get total partners: %w", err)
	}

	var activationRate float64
	if totalCreated > 0 {
		activationRate = float64(totalActivated) / float64(totalCreated) * 100
	}

	stats := &GeneralStatsResponse{
		TotalCouponsCreated:   totalCreated,
		TotalCouponsActivated: totalActivated,
		ActivationRate:        activationRate,
		ActivePartnersCount:   activePartners,
		TotalPartnersCount:    totalPartners,
		LastUpdated:           time.Now().Format(time.RFC3339),
	}

	// Кэшируем на 5 минут
	if data, err := json.Marshal(stats); err == nil {
		s.deps.RedisClient.Set(ctx, cacheKey, data, 5*time.Minute)
	}

	return stats, nil
}

// GetPartnerStats возвращает статистику по конкретному партнеру
func (s *StatsService) GetPartnerStats(ctx context.Context, partnerID uuid.UUID) (*PartnerStatsResponse, error) {
	cacheKey := fmt.Sprintf("partner_stats:%s", partnerID.String())

	// Проверяем кэш
	cached, err := s.deps.RedisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var stats PartnerStatsResponse
		if json.Unmarshal([]byte(cached), &stats) == nil {
			return &stats, nil
		}
	}

	// Получаем партнера
	partner, err := s.deps.PartnerRepository.GetByID(ctx, partnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get partner: %w", err)
	}

	// Получаем статистику купонов
	totalCoupons, err := s.deps.CouponRepository.CountByPartner(ctx, partnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get total coupons for partner: %w", err)
	}

	activatedCoupons, err := s.deps.CouponRepository.CountActivatedByPartner(ctx, partnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get activated coupons for partner: %w", err)
	}

	brandedPurchases, err := s.deps.CouponRepository.CountBrandedPurchasesByPartner(ctx, partnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get branded purchases for partner: %w", err)
	}

	unusedCoupons := totalCoupons - activatedCoupons

	var activationRate float64
	if totalCoupons > 0 {
		activationRate = float64(activatedCoupons) / float64(totalCoupons) * 100
	}

	// Получаем последнюю активность
	lastActivity, _ := s.deps.CouponRepository.GetLastActivityByPartner(ctx, partnerID)
	var lastActivityStr *string
	if lastActivity != nil {
		formatted := lastActivity.Format(time.RFC3339)
		lastActivityStr = &formatted
	}

	stats := &PartnerStatsResponse{
		PartnerID:            partnerID,
		PartnerName:          partner.BrandName,
		TotalCoupons:         totalCoupons,
		ActivatedCoupons:     activatedCoupons,
		UnusedCoupons:        unusedCoupons,
		BrandedSitePurchases: brandedPurchases,
		ActivationRate:       activationRate,
		LastActivity:         lastActivityStr,
	}

	// Кэшируем на 10 минут
	if data, err := json.Marshal(stats); err == nil {
		s.deps.RedisClient.Set(ctx, cacheKey, data, 10*time.Minute)
	}

	return stats, nil
}

// GetAllPartnersStats возвращает статистику по всем партнерам
func (s *StatsService) GetAllPartnersStats(ctx context.Context) (*PartnerListStatsResponse, error) {
	partners, err := s.deps.PartnerRepository.GetAll(ctx, "created_at", "desc")
	if err != nil {
		return nil, fmt.Errorf("failed to get partners: %w", err)
	}

	var partnerStats []PartnerStatsResponse
	for _, p := range partners {
		stats, err := s.GetPartnerStats(ctx, p.ID)
		if err != nil {
			continue
		}
		partnerStats = append(partnerStats, *stats)
	}

	return &PartnerListStatsResponse{
		Partners: partnerStats,
		Total:    int64(len(partnerStats)),
	}, nil
}

// GetTimeSeriesStats возвращает временную статистику для графиков
func (s *StatsService) GetTimeSeriesStats(ctx context.Context, filters *StatsFiltersRequest) (*TimeSeriesStatsResponse, error) {
	period := "day"
	if filters.Period != nil {
		period = *filters.Period
	}

	dateFrom := time.Now().AddDate(0, 0, -30) // последние 30 дней по умолчанию
	if filters.DateFrom != nil {
		if parsed, err := time.Parse("2006-01-02", *filters.DateFrom); err == nil {
			dateFrom = parsed
		}
	}

	dateTo := time.Now()
	if filters.DateTo != nil {
		if parsed, err := time.Parse("2006-01-02", *filters.DateTo); err == nil {
			dateTo = parsed
		}
	}

	rawData, err := s.deps.CouponRepository.GetTimeSeriesData(ctx, dateFrom, dateTo, period, filters.PartnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get time series data: %w", err)
	}

	// Преобразуем сырые данные в нужный формат
	var data []TimeSeriesStatsPoint
	for _, item := range rawData {
		point := TimeSeriesStatsPoint{
			Date:             item["date"].(string),
			CouponsCreated:   item["coupons_created"].(int64),
			CouponsActivated: item["coupons_activated"].(int64),
			CouponsPurchased: item["coupons_purchased"].(int64),
			NewPartnersCount: item["new_partners_count"].(int64),
		}
		data = append(data, point)
	}

	return &TimeSeriesStatsResponse{
		Period: period,
		Data:   data,
	}, nil
}

// GetSystemHealth возвращает состояние системы
func (s *StatsService) GetSystemHealth(ctx context.Context) (*SystemHealthResponse, error) {
	cacheKey := "system_health"

	// Проверяем кэш (обновляем каждую минуту)
	cached, err := s.deps.RedisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var health SystemHealthResponse
		if json.Unmarshal([]byte(cached), &health) == nil {
			return &health, nil
		}
	}

	// Проверяем статус БД
	dbStatus := "healthy"
	if err := s.deps.CouponRepository.HealthCheck(ctx); err != nil {
		dbStatus = "unhealthy"
	}

	// Проверяем статус Redis
	redisStatus := "healthy"
	if err := s.deps.RedisClient.Ping(ctx).Err(); err != nil {
		redisStatus = "unhealthy"
	}

	// Общий статус системы
	overallStatus := "healthy"
	if dbStatus != "healthy" || redisStatus != "healthy" {
		overallStatus = "unhealthy"
	}

	// Проверяем очередь обработки изображений
	imageProcessingQueue := int64(0)
	if queueSize, err := s.deps.RedisClient.LLen(ctx, "image_processing_queue").Result(); err == nil {
		imageProcessingQueue = queueSize
	}

	health := &SystemHealthResponse{
		Status:                overallStatus,
		DatabaseStatus:        dbStatus,
		RedisStatus:           redisStatus,
		ImageProcessingQueue:  imageProcessingQueue,
		AverageProcessingTime: 45.6,          // заглушка, пока что
		ErrorRate:             0.02,          // заглушка, пока что
		Uptime:                "72h 15m 30s", // заглушка, пока что
		LastUpdated:           time.Now().Format(time.RFC3339),
	}

	// Кэшируем на 1 минуту
	if data, err := json.Marshal(health); err == nil {
		s.deps.RedisClient.Set(ctx, cacheKey, data, 1*time.Minute)
	}

	return health, nil
}

// GetCouponsByStatus возвращает статистику купонов по статусам
func (s *StatsService) GetCouponsByStatus(ctx context.Context, partnerID *uuid.UUID) (*CouponsByStatusResponse, error) {
	statusCounts, err := s.deps.CouponRepository.GetExtendedStatusCounts(ctx, partnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get coupons by status: %w", err)
	}

	return &CouponsByStatusResponse{
		New:       statusCounts["new"],
		Activated: statusCounts["activated"],
		Used:      statusCounts["used"],
		Completed: statusCounts["completed"],
	}, nil
}

// GetCouponsBySize возвращает статистику купонов по размерам
func (s *StatsService) GetCouponsBySize(ctx context.Context, partnerID *uuid.UUID) (*CouponsBySizeResponse, error) {
	sizeCounts, err := s.deps.CouponRepository.GetSizeCounts(ctx, partnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get coupons by size: %w", err)
	}

	return &CouponsBySizeResponse{
		Size21x30: sizeCounts["21x30"],
		Size30x40: sizeCounts["30x40"],
		Size40x40: sizeCounts["40x40"],
		Size40x50: sizeCounts["40x50"],
		Size40x60: sizeCounts["40x60"],
		Size50x70: sizeCounts["50x70"],
	}, nil
}

// GetCouponsByStyle возвращает статистику купонов по стилям
func (s *StatsService) GetCouponsByStyle(ctx context.Context, partnerID *uuid.UUID) (*CouponsByStyleResponse, error) {
	styleCounts, err := s.deps.CouponRepository.GetStyleCounts(ctx, partnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get coupons by style: %w", err)
	}

	return &CouponsByStyleResponse{
		Grayscale: styleCounts["grayscale"],
		SkinTones: styleCounts["skin_tones"],
		PopArt:    styleCounts["pop_art"],
		MaxColors: styleCounts["max_colors"],
	}, nil
}

// GetTopPartners возвращает топ партнеров по активности
func (s *StatsService) GetTopPartners(ctx context.Context, limit int, sortBy ...string) (*TopPartnersResponse, error) {
	if limit <= 0 {
		limit = 10
	}

	partners, err := s.deps.PartnerRepository.GetAll(ctx, "created_at", "desc")
	if err != nil {
		return nil, fmt.Errorf("failed to get partners: %w", err)
	}

	var topPartners []TopPartnerItem
	for _, partner := range partners {
		totalCoupons, _ := s.deps.CouponRepository.CountByPartner(ctx, partner.ID)
		activatedCoupons, _ := s.deps.CouponRepository.CountActivatedByPartner(ctx, partner.ID)

		var activationRate float64
		if totalCoupons > 0 {
			activationRate = float64(activatedCoupons) / float64(totalCoupons) * 100
		}

		topPartners = append(topPartners, TopPartnerItem{
			PartnerID:        partner.ID,
			PartnerName:      partner.BrandName,
			ActivatedCoupons: activatedCoupons,
			TotalCoupons:     totalCoupons,
			ActivationRate:   activationRate,
		})
	}

	// Простая сортировка по активированным купонам (убывание)
	for i := 0; i < len(topPartners); i++ {
		for j := i + 1; j < len(topPartners); j++ {
			if topPartners[i].ActivatedCoupons < topPartners[j].ActivatedCoupons {
				topPartners[i], topPartners[j] = topPartners[j], topPartners[i]
			}
		}
	}

	// Обрезаем до нужного лимита
	if len(topPartners) > limit {
		topPartners = topPartners[:limit]
	}

	return &TopPartnersResponse{
		Partners: topPartners,
	}, nil
}

// GetRealTimeStats возвращает статистику в реальном времени
func (s *StatsService) GetRealTimeStats(ctx context.Context) (*RealTimeStatsResponse, error) {
	// Активные пользователи (заглушка), пока что
	activeUsers := int64(23)

	// Купоны активированные за последние 5 минут
	fiveMinutesAgo := time.Now().Add(-5 * time.Minute)
	couponsLast5Min, err := s.deps.CouponRepository.CountActivatedInTimeRange(ctx, fiveMinutesAgo, time.Now())
	if err != nil {
		couponsLast5Min = 0
	}

	// Изображения в обработке (заглушка), пока что
	imagesProcessing := int64(7)

	// Нагрузка системы (заглушка), пока что
	systemLoad := 0.65

	return &RealTimeStatsResponse{
		Timestamp:                time.Now(),
		ActiveUsers:              activeUsers,
		CouponsActivatedLast5Min: couponsLast5Min,
		ImagesProcessingNow:      imagesProcessing,
		SystemLoad:               systemLoad,
	}, nil
}
