package logger

import (
	"crypto/rand"
	"encoding/hex"
	"os"

	"github.com/gofiber/contrib/fiberzerolog"
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

// RequestLoggingMiddleware возвращает middleware для логирования запросов
func (l *Logger) RequestLoggingMiddleware() fiber.Handler {
	return fiberzerolog.New(fiberzerolog.Config{
		Logger: &l.logger,
		Fields: []string{"ip", "method", "url", "status", "latency", "user_agent"},
	})
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
