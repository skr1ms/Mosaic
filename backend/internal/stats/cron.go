package stats

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/robfig/cron/v3"
)

type CronService struct {
	Cron             *cron.Cron
	StatsService     StatsServiceInterface
	MetricsCollector MetricsCollectorInterface
}

func NewCronService(statsService *StatsService) *CronService {
	return &CronService{
		Cron:             cron.New(),
		StatsService:     statsService,
		MetricsCollector: NewMetricsCollector(),
	}
}

// Start launches all CRON tasks
func (cs *CronService) Start() error {
	_, err := cs.Cron.AddFunc("*/5 * * * *", cs.updateGeneralStats)
	if err != nil {
		return fmt.Errorf("failed to add general stats update: %w", err)
	}

	_, err = cs.Cron.AddFunc("*/10 * * * *", cs.updatePartnersMetrics)
	if err != nil {
		return fmt.Errorf("failed to add partners metrics update: %w", err)
	}

	_, err = cs.Cron.AddFunc("* * * * *", cs.updateSystemMetrics)
	if err != nil {
		return fmt.Errorf("failed to add system metrics update: %w", err)
	}

	_, err = cs.Cron.AddFunc("0 * * * *", cs.clearStatsCache)
	if err != nil {
		return fmt.Errorf("failed to add stats cache clear: %w", err)
	}

	_, err = cs.Cron.AddFunc("5 0 * * *", cs.aggregateDailyStats)
	if err != nil {
		return fmt.Errorf("failed to add daily stats aggregation: %w", err)
	}

	cs.Cron.Start()
	return nil
}

// Stop stops all CRON tasks
func (cs *CronService) Stop() {
	cs.Cron.Stop()
}

// updateGeneralStats updates general statistics and Prometheus metrics
func (cs *CronService) updateGeneralStats() {
	ctx := context.Background()

	stats, err := cs.StatsService.GetGeneralStats(ctx)
	if err != nil {
		cs.MetricsCollector.IncrementErrors("stats_update", "general")
		return
	}

	cs.MetricsCollector.SetPartnersCount(
		float64(stats.TotalPartnersCount),
		float64(stats.ActivePartnersCount),
	)

}

// updatePartnersMetrics updates partner metrics
func (cs *CronService) updatePartnersMetrics() {
	ctx := context.Background()

	_, err := cs.StatsService.GetAllPartnersStats(ctx)
	if err != nil {
		cs.MetricsCollector.IncrementErrors("stats_update", "partners")
		return
	}

}

// updateSystemMetrics updates system metrics
func (cs *CronService) updateSystemMetrics() {
	ctx := context.Background()

	health, err := cs.StatsService.GetSystemHealth(ctx)
	if err != nil {
		cs.MetricsCollector.IncrementErrors("health_check", "system")
		return
	}

	cs.MetricsCollector.SetImageProcessingQueueSize(float64(health.ImageProcessingQueue))

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	cs.MetricsCollector.SetSystemMetrics(
		float64(m.Alloc),
		0.0,
	)

	realTimeStats, err := cs.StatsService.GetRealTimeStats(ctx)
	if err == nil {
		cs.MetricsCollector.SetActiveUsers(float64(realTimeStats.ActiveUsers))
	}

}

// clearStatsCache clears statistics cache
func (cs *CronService) clearStatsCache() {
	ctx := context.Background()

	_, err := cs.StatsService.GetGeneralStats(ctx)
	if err != nil {
		cs.MetricsCollector.IncrementErrors("stats_update", "general")
	}
}

// aggregateDailyStats aggregates daily statistics
func (cs *CronService) aggregateDailyStats() {
	ctx := context.Background()

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

}
