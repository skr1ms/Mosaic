package stats

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/robfig/cron/v3"
)

// CronService сервис для периодических задач статистики
type CronService struct {
	Cron             *cron.Cron
	StatsService     *StatsService
	MetricsCollector  *MetricsCollector
}

// NewCronService создает новый экземпляр CronService
func NewCronService(statsService *StatsService) *CronService {
	return &CronService{
		Cron:             cron.New(),
		StatsService:     statsService,
		MetricsCollector: NewMetricsCollector(),
	}
}

// Start запускает все CRON задачи
func (cs *CronService) Start() error {
	// Обновление общей статистики каждые 5 минут
	_, err := cs.Cron.AddFunc("*/5 * * * *", cs.updateGeneralStats)
	if err != nil {
		return fmt.Errorf("failed to add general stats update: %w", err)
	}

	// Обновление метрик партнеров каждые 10 минут
	_, err = cs.Cron.AddFunc("*/10 * * * *", cs.updatePartnersMetrics)
	if err != nil {
		return fmt.Errorf("failed to add partners metrics update: %w", err)
	}

	// Обновление системных метрик каждую минуту
	_, err = cs.Cron.AddFunc("* * * * *", cs.updateSystemMetrics)
	if err != nil {
		return fmt.Errorf("failed to add system metrics update: %w", err)
	}

	// Очистка кэша статистики каждый час
	_, err = cs.Cron.AddFunc("0 * * * *", cs.clearStatsCache)
	if err != nil {
		return fmt.Errorf("failed to add stats cache clear: %w", err)
	}

	// Агрегация дневной статистики каждый день в 00:05
	_, err = cs.Cron.AddFunc("5 0 * * *", cs.aggregateDailyStats)
	if err != nil {
		return fmt.Errorf("failed to add daily stats aggregation: %w", err)
	}

	cs.Cron.Start()
	return nil
}

// Stop останавливает все CRON задачи
func (cs *CronService) Stop() {
	cs.Cron.Stop()
}

// updateGeneralStats обновляет общую статистику и Prometheus метрики
func (cs *CronService) updateGeneralStats() {
	ctx := context.Background()

	stats, err := cs.StatsService.GetGeneralStats(ctx)
	if err != nil {
		cs.MetricsCollector.IncrementErrors("stats_update", "general")
		return
	}

	// Обновляем Prometheus метрики
	cs.MetricsCollector.SetPartnersCount(
		float64(stats.TotalPartnersCount),
		float64(stats.ActivePartnersCount),
	)

}

// updatePartnersMetrics обновляет метрики партнеров
func (cs *CronService) updatePartnersMetrics() {
	ctx := context.Background()

	_, err := cs.StatsService.GetAllPartnersStats(ctx)
	if err != nil {
		cs.MetricsCollector.IncrementErrors("stats_update", "partners")
		return
	}

}

// updateSystemMetrics обновляет системные метрики
func (cs *CronService) updateSystemMetrics() {
	ctx := context.Background()

	// Получаем информацию о состоянии системы
	health, err := cs.StatsService.GetSystemHealth(ctx)
	if err != nil {
		cs.MetricsCollector.IncrementErrors("health_check", "system")
		return
	}

	// Обновляем метрики очереди обработки изображений
	cs.MetricsCollector.SetImageProcessingQueueSize(float64(health.ImageProcessingQueue))

	// Получаем статистику использования памяти
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Обновляем системные метрики
	cs.MetricsCollector.SetSystemMetrics(
		float64(m.Alloc), // Память в байтах
		0.0,              // CPU процент (можно добавить реальный расчет)
	)

	// Получаем количество активных пользователей
	realTimeStats, err := cs.StatsService.GetRealTimeStats(ctx)
	if err == nil {
		cs.MetricsCollector.SetActiveUsers(float64(realTimeStats.ActiveUsers))
	}

}

// clearStatsCache очищает кэш статистики
func (cs *CronService) clearStatsCache() {
	// Здесь можно добавить логику очистки устаревшего кэша
	// Например, удаление ключей старше определенного времени

	ctx := context.Background()

	// Принудительно обновляем основную статистику
	_, err := cs.StatsService.GetGeneralStats(ctx)
	if err != nil {
		cs.MetricsCollector.IncrementErrors("stats_update", "general")
	}
}

// aggregateDailyStats агрегирует дневную статистику
func (cs *CronService) aggregateDailyStats() {
	ctx := context.Background()

	// Получаем статистику за вчерашний день
	yesterday := time.Now().AddDate(0, 0, -1)
	filters := &StatsFiltersRequest{
		DateFrom: &[]string{yesterday.Format("2006-01-02")}[0],
		DateTo:   &[]string{yesterday.Format("2006-01-02")}[0],
		Period:   &[]string{"day"}[0],
	}

	_, err := cs.StatsService.GetTimeSeriesStats(ctx, filters)
	if err != nil {
		cs.MetricsCollector.IncrementErrors("aggregation", "daily_stats")
		return
	}

	// Здесь можно добавить логику сохранения агрегированной статистики
	// в отдельную таблицу для исторических данных
}
