package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
)

type Logger struct {
	logger zerolog.Logger
}

// NewLogger создает новый экземпляр Logger с настройками по умолчанию
func NewLogger() *Logger {
	logger := zerolog.New(os.Stdout).With().
		Timestamp().
		Str("service", "mosaic-api").
		Caller().
		Logger()

	// Устанавливаем уровень логирования в зависимости от окружения
	if os.Getenv("APP_MODE") == "development" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		logger = logger.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	return &Logger{logger: logger}
}

// GetZerologLogger возвращает zerolog.Logger для использования в других пакетах
func (l *Logger) GetZerologLogger() *zerolog.Logger {
	return &l.logger
}

// generateRequestID создает уникальный ID для запроса
func generateRequestID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// RequestLoggingMiddleware возвращает middleware для логирования запросов с асинхронными операциями
func (l *Logger) RequestLoggingMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Асинхронно собираем метрики 
		go func() {
			l.logRequestMetrics(c.Method(), c.Path(), c.IP(), c.Get("User-Agent"))
		}()

		err := c.Next()

		// Асинхронно логируем завершение запроса
		go func() {
			duration := time.Since(start)
			l.logRequestCompletion(c.Method(), c.Path(), c.Response().StatusCode(), duration, c.Get("X-Request-ID"))
		}()

		return err
	}
}

// RequestIDMiddleware добавляет request_id в контекст
func (l *Logger) RequestIDMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
			c.Set("X-Request-ID", requestID)
		}

		// Добавляем логгер с request_id в контекст
		loggerWithRequestID := l.logger.With().Str("request_id", requestID).Logger()
		c.SetUserContext(loggerWithRequestID.WithContext(c.UserContext()))

		return c.Next()
	}
}

// ErrorHandler возвращает обработчик ошибок с логированием
func (l *Logger) ErrorHandler() fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		code := fiber.StatusInternalServerError
		if e, ok := err.(*fiber.Error); ok {
			code = e.Code
		}

		// Логируем все HTTP ошибки
		l.logger.Error().
			Err(err).
			Int("status_code", code).
			Str("method", c.Method()).
			Str("path", c.Path()).
			Str("ip", c.IP()).
			Msg("HTTP Error")

		return c.Status(code).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
}

// logRequestMetrics асинхронно логирует метрики запроса
func (l *Logger) logRequestMetrics(method, path, ip, userAgent string) {
	l.logger.Debug().
		Str("method", method).
		Str("path", path).
		Str("ip", ip).
		Str("user_agent", userAgent).
		Msg("Request metrics collected")
}

// logRequestCompletion асинхронно логирует завершение запроса
func (l *Logger) logRequestCompletion(method, path string, status int, duration time.Duration, requestID string) {
	l.logger.Info().
		Str("method", method).
		Str("path", path).
		Int("status", status).
		Dur("duration", duration).
		Str("request_id", requestID).
		Msg("Request completed")
}

// CombinedMiddleware объединяет логирование и request ID в один middleware для лучшей производительности
func (l *Logger) CombinedMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Генерируем или получаем request ID
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
			c.Set("X-Request-ID", requestID)
		}

		// Добавляем логгер с request_id в контекст
		loggerWithRequestID := l.logger.With().Str("request_id", requestID).Logger()
		c.SetUserContext(loggerWithRequestID.WithContext(c.UserContext()))

		// Асинхронно собираем начальные метрики
		go func() {
			l.logRequestStart(c.Method(), c.Path(), c.IP(), c.Get("User-Agent"), requestID)
		}()

		// Обрабатываем запрос
		err := c.Next()

		// Асинхронно логируем завершение
		go func() {
			duration := time.Since(start)
			l.logRequestCompletion(c.Method(), c.Path(), c.Response().StatusCode(), duration, requestID)
		}()

		return err
	}
}

// logRequestStart асинхронно логирует начало запроса
func (l *Logger) logRequestStart(method, path, ip, userAgent, requestID string) {
	l.logger.Debug().
		Str("method", method).
		Str("path", path).
		Str("ip", ip).
		Str("user_agent", userAgent).
		Str("request_id", requestID).
		Msg("Request started")
}

// AnalyticsMiddleware собирает аналитику асинхронно
func (l *Logger) AnalyticsMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Собираем данные синхронно
		analyticsData := map[string]interface{}{
			"method":     c.Method(),
			"path":       c.Path(),
			"ip":         c.IP(),
			"user_agent": c.Get("User-Agent"),
			"timestamp":  time.Now(),
			"request_id": c.Get("X-Request-ID"),
		}

		// Асинхронно отправляем в аналитику
		go func() {
			l.sendToAnalytics(analyticsData)
		}()

		// Асинхронно обновляем метрики
		go func() {
			l.updateMetrics(c.Method(), c.Path())
		}()

		return c.Next()
	}
}

// sendToAnalytics асинхронно отправляет данные в аналитику
func (l *Logger) sendToAnalytics(data map[string]interface{}) {
	// Здесь можно отправить данные в аналитическую систему
	l.logger.Debug().
		Interface("analytics_data", data).
		Msg("Analytics data collected")
}

// updateMetrics асинхронно обновляет метрики
func (l *Logger) updateMetrics(method, path string) {
	// Здесь можно обновить счетчики метрик
	l.logger.Debug().
		Str("method", method).
		Str("path", path).
		Msg("Metrics updated")
}
