package stats

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/skr1ms/mosaic/pkg/middleware"
)

// Prometheus metrics for system monitoring
var (
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

	ErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mosaic_errors_total",
			Help: "Total number of errors",
		},
		[]string{"type", "component"},
	)

	ActiveUsers = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "mosaic_active_users",
			Help: "Number of currently active users",
		},
	)

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

// MetricsCollector wrapper for updating metrics
type MetricsCollector struct{}

var _ middleware.MetricsCollector = (*MetricsCollector)(nil)

// NewMetricsCollector creates new metrics collector instance
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{}
}

// IncrementCouponsCreated increments created coupons counter
func (m *MetricsCollector) IncrementCouponsCreated(partnerID, size, style string) {
	CouponsCreatedTotal.WithLabelValues(partnerID, size, style).Inc()
}

// IncrementCouponsActivated increments activated coupons counter
func (m *MetricsCollector) IncrementCouponsActivated(partnerID, size, style string) {
	CouponsActivatedTotal.WithLabelValues(partnerID, size, style).Inc()
}

// IncrementCouponsPurchased increments purchased coupons counter
func (m *MetricsCollector) IncrementCouponsPurchased(partnerID string) {
	CouponsPurchasedTotal.WithLabelValues(partnerID).Inc()
}

// ObserveImageProcessingDuration records image processing duration
func (m *MetricsCollector) ObserveImageProcessingDuration(operationType, status string, duration float64) {
	ImageProcessingDuration.WithLabelValues(operationType, status).Observe(duration)
}

// SetImageProcessingQueueSize sets image processing queue size
func (m *MetricsCollector) SetImageProcessingQueueSize(size float64) {
	ImageProcessingQueue.Set(size)
}

// SetPartnersCount sets partners count
func (m *MetricsCollector) SetPartnersCount(total, active float64) {
	PartnersTotal.Set(total)
	PartnersActive.Set(active)
}

// IncrementHTTPRequests increments HTTP requests counter
func (m *MetricsCollector) IncrementHTTPRequests(method, endpoint, status string) {
	HTTPRequestsTotal.WithLabelValues(method, endpoint, status).Inc()
}

// ObserveHTTPRequestDuration records HTTP request duration
func (m *MetricsCollector) ObserveHTTPRequestDuration(method, endpoint string, duration float64) {
	HTTPRequestDuration.WithLabelValues(method, endpoint).Observe(duration)
}

// SetDatabaseConnections sets database connections count
func (m *MetricsCollector) SetDatabaseConnections(count float64) {
	DatabaseConnectionsActive.Set(count)
}

// SetRedisConnections sets Redis connections count
func (m *MetricsCollector) SetRedisConnections(count float64) {
	RedisConnectionsActive.Set(count)
}

// IncrementErrors increments errors counter
func (m *MetricsCollector) IncrementErrors(errorType, component string) {
	ErrorsTotal.WithLabelValues(errorType, component).Inc()
}

// SetActiveUsers sets active users count
func (m *MetricsCollector) SetActiveUsers(count float64) {
	ActiveUsers.Set(count)
}

// SetSystemMetrics sets system metrics
func (m *MetricsCollector) SetSystemMetrics(memoryBytes, cpuPercent float64) {
	MemoryUsageBytes.Set(memoryBytes)
	CPUUsagePercent.Set(cpuPercent)
}
