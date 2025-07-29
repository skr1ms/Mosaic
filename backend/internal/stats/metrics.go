package stats

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/skr1ms/mosaic/pkg/middleware"
)

// Prometheus метрики для мониторинга системы
var (
	// Метрики купонов
	CouponsCreatedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mosaic_coupons_created_total",
			Help: "Total number of coupons created",
		},
		[]string{"partner_id", "size", "style"},
	)

	CouponsActivatedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mosaic_coupons_activated_total",
			Help: "Total number of coupons activated",
		},
		[]string{"partner_id", "size", "style"},
	)

	CouponsPurchasedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mosaic_coupons_purchased_total",
			Help: "Total number of coupons purchased online",
		},
		[]string{"partner_id"},
	)

	// Метрики обработки изображений
	ImageProcessingDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "mosaic_image_processing_duration_seconds",
			Help:    "Duration of image processing operations",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation_type", "status"},
	)

	ImageProcessingQueue = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "mosaic_image_processing_queue_size",
			Help: "Current size of image processing queue",
		},
	)

	// Метрики партнеров
	PartnersTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "mosaic_partners_total",
			Help: "Total number of partners",
		},
	)

	PartnersActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "mosaic_partners_active",
			Help: "Number of active partners",
		},
	)

	// Метрики системы
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mosaic_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "mosaic_http_request_duration_seconds",
			Help:    "Duration of HTTP requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	DatabaseConnectionsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "mosaic_database_connections_active",
			Help: "Number of active database connections",
		},
	)

	RedisConnectionsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "mosaic_redis_connections_active",
			Help: "Number of active Redis connections",
		},
	)

	// Метрики ошибок
	ErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mosaic_errors_total",
			Help: "Total number of errors",
		},
		[]string{"type", "component"},
	)

	// Метрики активных пользователей
	ActiveUsers = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "mosaic_active_users",
			Help: "Number of currently active users",
		},
	)

	// Метрики использования памяти и CPU (опционально)
	MemoryUsageBytes = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "mosaic_memory_usage_bytes",
			Help: "Current memory usage in bytes",
		},
	)

	CPUUsagePercent = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "mosaic_cpu_usage_percent",
			Help: "Current CPU usage percentage",
		},
	)
)

// MetricsCollector обертка для обновления метрик
// Реализует интерфейс middleware.MetricsCollector
type MetricsCollector struct{}

// Проверяем, что MetricsCollector реализует интерфейс
var _ middleware.MetricsCollector = (*MetricsCollector)(nil)

// NewMetricsCollector создает новый экземпляр MetricsCollector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{}
}

// IncrementCouponsCreated увеличивает счетчик созданных купонов
func (m *MetricsCollector) IncrementCouponsCreated(partnerID, size, style string) {
	CouponsCreatedTotal.WithLabelValues(partnerID, size, style).Inc()
}

// IncrementCouponsActivated увеличивает счетчик активированных купонов
func (m *MetricsCollector) IncrementCouponsActivated(partnerID, size, style string) {
	CouponsActivatedTotal.WithLabelValues(partnerID, size, style).Inc()
}

// IncrementCouponsPurchased увеличивает счетчик купленных купонов
func (m *MetricsCollector) IncrementCouponsPurchased(partnerID string) {
	CouponsPurchasedTotal.WithLabelValues(partnerID).Inc()
}

// ObserveImageProcessingDuration записывает время обработки изображения
func (m *MetricsCollector) ObserveImageProcessingDuration(operationType, status string, duration float64) {
	ImageProcessingDuration.WithLabelValues(operationType, status).Observe(duration)
}

// SetImageProcessingQueueSize устанавливает размер очереди обработки
func (m *MetricsCollector) SetImageProcessingQueueSize(size float64) {
	ImageProcessingQueue.Set(size)
}

// SetPartnersCount устанавливает количество партнеров
func (m *MetricsCollector) SetPartnersCount(total, active float64) {
	PartnersTotal.Set(total)
	PartnersActive.Set(active)
}

// IncrementHTTPRequests увеличивает счетчик HTTP запросов
func (m *MetricsCollector) IncrementHTTPRequests(method, endpoint, status string) {
	HTTPRequestsTotal.WithLabelValues(method, endpoint, status).Inc()
}

// ObserveHTTPRequestDuration записывает время обработки HTTP запроса
func (m *MetricsCollector) ObserveHTTPRequestDuration(method, endpoint string, duration float64) {
	HTTPRequestDuration.WithLabelValues(method, endpoint).Observe(duration)
}

// SetDatabaseConnections устанавливает количество соединений с БД
func (m *MetricsCollector) SetDatabaseConnections(count float64) {
	DatabaseConnectionsActive.Set(count)
}

// SetRedisConnections устанавливает количество соединений с Redis
func (m *MetricsCollector) SetRedisConnections(count float64) {
	RedisConnectionsActive.Set(count)
}

// IncrementErrors увеличивает счетчик ошибок
func (m *MetricsCollector) IncrementErrors(errorType, component string) {
	ErrorsTotal.WithLabelValues(errorType, component).Inc()
}

// SetActiveUsers устанавливает количество активных пользователей
func (m *MetricsCollector) SetActiveUsers(count float64) {
	ActiveUsers.Set(count)
}

// SetSystemMetrics устанавливает системные метрики
func (m *MetricsCollector) SetSystemMetrics(memoryBytes, cpuPercent float64) {
	MemoryUsageBytes.Set(memoryBytes)
	CPUUsagePercent.Set(cpuPercent)
}
