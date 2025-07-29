package middleware

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

type MetricsCollector interface {
	IncrementHTTPRequests(method, endpoint, status string)
	ObserveHTTPRequestDuration(method, endpoint string, duration float64)
	IncrementErrors(errorType, component string)
}

// MetricsMiddleware middleware для автоматического сбора метрик HTTP запросов
func MetricsMiddleware(metricsCollector MetricsCollector) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Выполняем запрос
		err := c.Next()

		// Собираем метрики
		duration := time.Since(start).Seconds()
		method := c.Method()
		path := c.Route().Path
		status := strconv.Itoa(c.Response().StatusCode())

		// Записываем метрики
		metricsCollector.IncrementHTTPRequests(method, path, status)
		metricsCollector.ObserveHTTPRequestDuration(method, path, duration)

		// Если произошла ошибка, записываем метрику ошибки
		if err != nil {
			metricsCollector.IncrementErrors("http_error", "request_handler")
		}

		return err
	}
}

// ActiveUsersMiddleware middleware для отслеживания активных пользователей
func ActiveUsersMiddleware(redisClient interface{}) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Здесь можно добавить логику отслеживания активных пользователей
		// Например, записывать IP адреса или user ID в Redis set с TTL

		// Получаем IP пользователя
		userIP := c.IP()
		if userIP != "" {
			// TODO: Добавить пользователя в Redis set активных пользователей
			// с TTL например 5 минут
		}

		return c.Next()
	}
}
