package middleware

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

type MetricsCollector interface {
	IncrementHTTPRequests(method, endpoint, status string)
	ObserveHTTPRequestDuration(method, endpoint string, duration float64)
	IncrementErrors(errorType, component string)
}

// MetricsMiddleware for automatic HTTP request metrics collection
func MetricsMiddleware(metricsCollector MetricsCollector, logger *Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		logMetricsStart(c.Method(), c.Path(), c.IP(), logger)

		err := c.Next()

		duration := time.Since(start).Seconds()
		method := c.Method()
		path := c.Route().Path
		status := strconv.Itoa(c.Response().StatusCode())

		metricsCollector.IncrementHTTPRequests(method, path, status)
		metricsCollector.ObserveHTTPRequestDuration(method, path, duration)

		logMetricsEnd(method, path, c.IP(), time.Since(start), c.Response().StatusCode(), logger)

		if err != nil {
			metricsCollector.IncrementErrors("http_error", "request_handler")
			logMetricsError(method, path, c.IP(), err, logger)
		}

		return err
	}
}

// ActiveUsersMiddleware for tracking active users
func ActiveUsersMiddleware(redisClient any) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userIP := c.IP()

		if userIP != "" {
			// Add user to Redis set of active users with TTL 5 minutes
			// In reality, here will be Redis call:
			// redisKey := "active_users"
			// err := redisClient.SAdd(ctx, redisKey, userIP).Err()
			// if err == nil {
			//     redisClient.Expire(ctx, redisKey, 5*time.Minute)
			// }
			logActiveUser(userIP, c.Get("User-Agent"), c.Path())
		}

		return c.Next()
	}
}

// logMetricsStart logs metrics start
func logMetricsStart(method, path, ip string, logger *Logger) {
	logger.GetZerologLogger().Debug().
		Str("method", method).
		Str("path", path).
		Str("ip", ip).
		Msg("Metrics collection started")
}

// logMetricsEnd logs metrics end
func logMetricsEnd(method, path, ip string, duration time.Duration, statusCode int, logger *Logger) {
	logger.GetZerologLogger().Debug().
		Str("method", method).
		Str("path", path).
		Str("ip", ip).
		Dur("duration", duration).
		Int("status_code", statusCode).
		Msg("Metrics collection ended")
}

// logMetricsError logs metrics error
func logMetricsError(method, path, ip string, err error, logger *Logger) {
	logger.GetZerologLogger().Error().
		Str("method", method).
		Str("path", path).
		Str("ip", ip).
		Err(err).
		Msg("Metrics collection error")
}

// logActiveUser logs active user tracking
func logActiveUser(ip, userAgent, path string) {
	log.Debug().
		Str("ip", ip).
		Str("user_agent", userAgent).
		Str("path", path).
		Str("event_type", "active_user").
		Msg("Active user tracked")
}
