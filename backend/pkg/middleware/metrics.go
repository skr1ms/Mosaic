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

// MetricsMiddleware middleware для автоматического сбора метрик HTTP запросов с асинхронными операциями
func MetricsMiddleware(metricsCollector MetricsCollector) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Асинхронно логируем начало запроса
		go func() {
			logMetricsStart(c.Method(), c.Path(), c.IP())
		}()

		// Выполняем запрос
		err := c.Next()

		// Собираем данные для метрик
		duration := time.Since(start).Seconds()
		method := c.Method()
		path := c.Route().Path
		status := strconv.Itoa(c.Response().StatusCode())

		// Асинхронно записываем метрики (неблокирующие операции)
		go func() {
			metricsCollector.IncrementHTTPRequests(method, path, status)
			metricsCollector.ObserveHTTPRequestDuration(method, path, duration)

			// Логируем метрики
			logMetricsCompletion(method, path, status, duration)
		}()

		// Если произошла ошибка, асинхронно записываем метрику ошибки
		if err != nil {
			go func() {
				metricsCollector.IncrementErrors("http_error", "request_handler")
				logErrorMetrics(method, path, err.Error())
			}()
		}

		return err
	}
}

// ActiveUsersMiddleware middleware для отслеживания активных пользователей с асинхронными операциями
func ActiveUsersMiddleware(redisClient interface{}) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Получаем IP пользователя
		userIP := c.IP()

		if userIP != "" {
			// Асинхронно добавляем пользователя в Redis set активных пользователей
			go func() {
				// Добавить пользователя в Redis set активных пользователей с TTL 5 минут
				// В реальности здесь будет обращение к Redis:
				// redisKey := "active_users"
				// err := redisClient.SAdd(ctx, redisKey, userIP).Err()
				// if err == nil {
				//     redisClient.Expire(ctx, redisKey, 5*time.Minute)
				// }
				logActiveUser(userIP, c.Get("User-Agent"), c.Path())
			}()
		}

		return c.Next()
	}
}

// logMetricsStart асинхронно логирует начало сбора метрик
func logMetricsStart(method, path, ip string) {
	log.Debug().
		Str("method", method).
		Str("path", path).
		Str("ip", ip).
		Str("event_type", "metrics_start").
		Msg("Metrics collection started")
}

// logMetricsCompletion асинхронно логирует завершение сбора метрик
func logMetricsCompletion(method, path, status string, duration float64) {
	log.Debug().
		Str("method", method).
		Str("path", path).
		Str("status", status).
		Float64("duration_seconds", duration).
		Str("event_type", "metrics_completion").
		Msg("Metrics collection completed")
}

// logErrorMetrics асинхронно логирует ошибки метрик
func logErrorMetrics(method, path, errorMsg string) {
	log.Error().
		Str("method", method).
		Str("path", path).
		Str("error", errorMsg).
		Str("event_type", "metrics_error").
		Msg("Error in metrics collection")
}

// logActiveUser асинхронно логирует активных пользователей
func logActiveUser(ip, userAgent, path string) {
	log.Debug().
		Str("ip", ip).
		Str("user_agent", userAgent).
		Str("path", path).
		Str("event_type", "active_user").
		Msg("Active user tracked")
}
