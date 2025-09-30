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

func NewLogger() *Logger {
	logger := zerolog.New(os.Stdout).With().
		Timestamp().
		Str("service", "mosaic-api").
		Caller().
		Logger()

	env := os.Getenv("ENVIRONMENT")
	switch env {
	case "development", "dev":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		logger = logger.Output(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		})
	case "production", "prod":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	return &Logger{logger: logger}
}

func (l *Logger) GetZerologLogger() *zerolog.Logger {
	return &l.logger
}

func generateRequestID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID if crypto fails
		fallbackBytes := make([]byte, 8)
		timestamp := time.Now().UnixNano()
		for i := 0; i < 8; i++ {
			fallbackBytes[i] = byte(timestamp >> (i * 8))
		}
		return hex.EncodeToString(fallbackBytes)
	}
	return hex.EncodeToString(bytes)
}

func (l *Logger) RequestIDMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
			c.Set("X-Request-ID", requestID)
		}

		c.Set("X-Request-ID", requestID)

		loggerWithRequestID := l.logger.With().Str("request_id", requestID).Logger()
		c.SetUserContext(loggerWithRequestID.WithContext(c.UserContext()))

		return c.Next()
	}
}

func (l *Logger) ErrorHandler() fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		code := fiber.StatusInternalServerError
		message := "Internal Server Error"

		if e, ok := err.(*fiber.Error); ok {
			code = e.Code
			message = e.Message
		}

		logger := l.FromContext(c)

		logEvent := logger.Error()
		if code >= 500 {
			logEvent = logger.Error()
		} else if code >= 400 {
			logEvent = logger.Warn()
		}

		logEvent.
			Err(err).
			Int("status_code", code).
			Str("method", c.Method()).
			Str("path", c.Path()).
			Str("query", c.Request().URI().QueryArgs().String()).
			Str("ip", c.IP()).
			Str("user_agent", c.Get("User-Agent")).
			Str("referer", c.Get("Referer")).
			Msg("HTTP Error")

		errorResponse := fiber.Map{
			"error":      message,
			"request_id": c.Get("X-Request-ID"),
		}

		if os.Getenv("ENVIRONMENT") == "development" {
			errorResponse["details"] = err.Error()
		}

		return c.Status(code).JSON(errorResponse)
	}
}

func (l *Logger) CombinedMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
			c.Set("X-Request-ID", requestID)
		}

		loggerWithRequestID := l.logger.With().Str("request_id", requestID).Logger()
		c.SetUserContext(loggerWithRequestID.WithContext(c.UserContext()))

		loggerWithRequestID.Debug().
			Str("method", c.Method()).
			Str("path", c.Path()).
			Str("query", c.Request().URI().QueryArgs().String()).
			Str("ip", c.IP()).
			Str("user_agent", c.Get("User-Agent")).
			Str("referer", c.Get("Referer")).
			Int("content_length", len(c.Body())).
			Msg("Request started")

		err := c.Next()

		duration := time.Since(start)
		status := c.Response().StatusCode()

		logEvent := loggerWithRequestID.Info()
		if status >= 500 {
			logEvent = loggerWithRequestID.Error()
		} else if status >= 400 {
			logEvent = loggerWithRequestID.Warn()
		}

		logEvent.
			Str("method", c.Method()).
			Str("path", c.Path()).
			Str("query", c.Request().URI().QueryArgs().String()).
			Int("status", status).
			Dur("duration", duration).
			Int("response_size", len(c.Response().Body())).
			Str("ip", c.IP()).
			Msg("Request completed")

		// Log slow requests (> 1s)
		if duration > time.Second {
			loggerWithRequestID.Warn().
				Str("method", c.Method()).
				Str("path", c.Path()).
				Dur("duration", duration).
				Msg("Slow request detected")
		}

		return err
	}
}

func (l *Logger) FromContext(c *fiber.Ctx) *zerolog.Logger {
	if logger := zerolog.Ctx(c.UserContext()); logger != nil {
		return logger
	}
	return &l.logger
}

// HealthCheckMiddleware logs only errors for health check endpoints
func (l *Logger) HealthCheckMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Path() == "/health" || c.Path() == "/healthz" || c.Path() == "/ping" {
			err := c.Next()
			if err != nil {
				l.logger.Error().
					Err(err).
					Str("path", c.Path()).
					Msg("Health check failed")
			}
			return err
		}
		return c.Next()
	}
}

// SkipLoggingMiddleware excludes specified paths from logging
func (l *Logger) SkipLoggingMiddleware(skipPaths ...string) fiber.Handler {
	skipMap := make(map[string]bool)
	for _, path := range skipPaths {
		skipMap[path] = true
	}

	return func(c *fiber.Ctx) error {
		if skipMap[c.Path()] {
			return c.Next()
		}
		return l.CombinedMiddleware()(c)
	}
}
