package stats

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
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

	totalActivated, err := s.deps.CouponRepository.CountByStatus(ctx, "activated")
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

	activatedCoupons, err := s.deps.CouponRepository.CountByPartnerAndStatus(ctx, partnerID, "activated")
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
	partners, err := s.deps.PartnerRepository.GetAll(ctx)
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
		return nil, fmt.Errorf("database health check failed: %w", err)
	}

	// Проверяем статус Redis
	redisStatus := "healthy"
	if err := s.deps.RedisClient.Ping(ctx).Err(); err != nil {
		redisStatus = "unhealthy"
		return nil, fmt.Errorf("redis health check failed: %w", err)
	}

	// Получаем метрики обработки изображений
	imageQueueSize, _ := s.getImageProcessingQueueSize(ctx)
	avgProcessingTime, _ := s.getAverageProcessingTime(ctx)
	errorRate, _ := s.getErrorRate(ctx)

	// Определяем общий статус
	status := "healthy"
	if dbStatus != "healthy" || redisStatus != "healthy" || errorRate > 5.0 {
		status = "critical"
	} else if imageQueueSize > 100 || avgProcessingTime > 60.0 {
		status = "warning"
	}

	health := &SystemHealthResponse{
		Status:                status,
		DatabaseStatus:        dbStatus,
		RedisStatus:           redisStatus,
		ImageProcessingQueue:  imageQueueSize,
		AverageProcessingTime: avgProcessingTime,
		ErrorRate:             errorRate,
		Uptime:                s.getUptime(),
		LastUpdated:           time.Now().Format(time.RFC3339),
	}

	// Кэшируем на 1 минуту
	if data, err := json.Marshal(health); err == nil {
		s.deps.RedisClient.Set(ctx, cacheKey, data, 1*time.Minute)
	}

	return health, nil
}

// GetRealTimeStats возвращает real-time статистику
func (s *StatsService) GetRealTimeStats(ctx context.Context) (*RealTimeStatsResponse, error) {
	// Получаем активных пользователей (из Redis)
	activeUsers, _ := s.getActiveUsersCount(ctx)

	// Купоны активированные за последние 5 минут
	couponsLast5Min, _ := s.deps.CouponRepository.CountActivatedInTimeRange(ctx, time.Now().Add(-5*time.Minute), time.Now())

	// Изображения в обработке
	imagesProcessing, _ := s.getImageProcessingQueueSize(ctx)

	// Загрузка системы (можно получить из Prometheus метрик)
	systemLoad, _ := s.getSystemLoad(ctx)

	return &RealTimeStatsResponse{
		Timestamp:                time.Now(),
		ActiveUsers:              activeUsers,
		CouponsActivatedLast5Min: couponsLast5Min,
		ImagesProcessingNow:      imagesProcessing,
		SystemLoad:               systemLoad,
	}, nil
}

// GetCouponsByStatus возвращает статистику купонов по статусам
func (s *StatsService) GetCouponsByStatus(ctx context.Context, partnerID *uuid.UUID) (*CouponsByStatusResponse, error) {
	cacheKey := "coupons_by_status"
	if partnerID != nil {
		cacheKey += ":" + partnerID.String()
	}

	// Проверяем кэш
	cached, err := s.deps.RedisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var stats CouponsByStatusResponse
		if json.Unmarshal([]byte(cached), &stats) == nil {
			return &stats, nil
		}
	}

	statusCounts, err := s.deps.CouponRepository.GetStatusCounts(ctx, partnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get status counts: %w", err)
	}

	stats := &CouponsByStatusResponse{
		New:       statusCounts["new"],
		Activated: statusCounts["activated"],
		Used:      statusCounts["used"],
		Completed: statusCounts["completed"],
	}

	// Кэшируем на 5 минут
	if data, err := json.Marshal(stats); err == nil {
		s.deps.RedisClient.Set(ctx, cacheKey, data, 5*time.Minute)
	}

	return stats, nil
}

// GetTopPartners возвращает топ партнеров по активности
func (s *StatsService) GetTopPartners(ctx context.Context, limit int) (*TopPartnersResponse, error) {
	if limit <= 0 {
		limit = 10
	}

	partners, err := s.deps.PartnerRepository.GetTopByActivity(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get top partners: %w", err)
	}

	var topPartners []TopPartnerItem
	for _, p := range partners {
		stats, err := s.GetPartnerStats(ctx, p.ID)
		if err != nil {
			continue
		}

		topPartners = append(topPartners, TopPartnerItem{
			PartnerID:        p.ID,
			PartnerName:      p.BrandName,
			ActivatedCoupons: stats.ActivatedCoupons,
			TotalCoupons:     stats.TotalCoupons,
			ActivationRate:   stats.ActivationRate,
		})
	}

	return &TopPartnersResponse{
		Partners: topPartners,
	}, nil
}

// Вспомогательные методы

func (s *StatsService) getImageProcessingQueueSize(ctx context.Context) (int64, error) {
	// Получаем размер очереди из Redis
	queueSize, err := s.deps.RedisClient.LLen(ctx, "image_processing_queue").Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get image processing queue size: %w", err)
	}
	return queueSize, nil
}

func (s *StatsService) getAverageProcessingTime(ctx context.Context) (float64, error) {
	// Получаем среднее время обработки из метрик
	avgTimeStr, err := s.deps.RedisClient.Get(ctx, "avg_processing_time").Result()
	if err != nil {
		return 0, nil // Возвращаем 0 если метрика не найдена
	}

	avgTime, err := strconv.ParseFloat(avgTimeStr, 64)
	if err != nil {
		return 0, nil
	}

	return avgTime, nil
}

func (s *StatsService) getErrorRate(ctx context.Context) (float64, error) {
	// Получаем процент ошибок из метрик
	errorRateStr, err := s.deps.RedisClient.Get(ctx, "error_rate").Result()
	if err != nil {
		return 0, nil
	}

	errorRate, err := strconv.ParseFloat(errorRateStr, 64)
	if err != nil {
		return 0, nil
	}

	return errorRate, nil
}

func (s *StatsService) getUptime() string {
	// Можно получить из переменной окружения или файла
	// Пока возвращаем заглушку
	return "99.9%"
}

func (s *StatsService) getActiveUsersCount(ctx context.Context) (int64, error) {
	// Подсчитываем активных пользователей из Redis (например, по сессиям)
	activeCount, err := s.deps.RedisClient.SCard(ctx, "active_users").Result()
	if err != nil {
		return 0, nil
	}
	return activeCount, nil
}

func (s *StatsService) getSystemLoad(ctx context.Context) (float64, error) {
	// Получаем загрузку системы из метрик
	loadStr, err := s.deps.RedisClient.Get(ctx, "system_load").Result()
	if err != nil {
		return 0, nil
	}

	load, err := strconv.ParseFloat(loadStr, 64)
	if err != nil {
		return 0, nil
	}

	return load, nil
}
