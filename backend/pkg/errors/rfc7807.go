package errors

import (
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

type ProblemDetail struct {
	Type       string                 `json:"type"`                 // URI идентифицирующий тип проблемы
	Title      string                 `json:"title"`                // Краткое описание проблемы
	Status     int                    `json:"status"`               // HTTP статус код
	Detail     string                 `json:"detail"`               // Детальное описание проблемы
	Instance   string                 `json:"instance"`             // URI идентифицирующий конкретное вхождение проблемы
	Extensions map[string]interface{} `json:"extensions,omitempty"` // Дополнительные поля
}

// Error реализует интерфейс error
func (p *ProblemDetail) Error() string {
	return p.Detail
}

// ErrorType определяет типы ошибок
type ErrorType string

const (
	ErrorTypeValidation     ErrorType = "validation-error"
	ErrorTypeAuthentication ErrorType = "authentication-error"
	ErrorTypeAuthorization  ErrorType = "authorization-error"
	ErrorTypeNotFound       ErrorType = "not-found-error"
	ErrorTypeConflict       ErrorType = "conflict-error"
	ErrorTypeRateLimit      ErrorType = "rate-limit-error"
	ErrorTypeInternal       ErrorType = "internal-server-error"
	ErrorTypeBadRequest     ErrorType = "bad-request-error"
	ErrorTypePayment        ErrorType = "payment-error"
)

// GetBaseURI возвращает базовый URI для типов ошибок
func GetBaseURI() string {
	return "https://api.mosaic.com/errors/"
}

// NewProblemDetail создает новый ProblemDetail
func NewProblemDetail(errorType ErrorType, title, detail string, status int) *ProblemDetail {
	return &ProblemDetail{
		Type:   GetBaseURI() + string(errorType),
		Title:  title,
		Status: status,
		Detail: detail,
	}
}

// WithInstance добавляет instance URI
func (p *ProblemDetail) WithInstance(instance string) *ProblemDetail {
	p.Instance = instance
	return p
}

// WithExtension добавляет дополнительное поле
func (p *ProblemDetail) WithExtension(key string, value interface{}) *ProblemDetail {
	if p.Extensions == nil {
		p.Extensions = make(map[string]interface{})
	}
	p.Extensions[key] = value
	return p
}

// SendError отправляет стандартизированную ошибку
func SendError(c *fiber.Ctx, problem *ProblemDetail) error {
	// Добавляем timestamp
	problem.WithExtension("timestamp", time.Now().UTC().Format(time.RFC3339))

	// Добавляем request ID если есть
	if requestID := c.Get("X-Request-ID"); requestID != "" {
		problem.WithExtension("request_id", requestID)
	}

	// Логируем ошибку
	logLevel := log.Warn()
	if problem.Status >= 500 {
		logLevel = log.Error()
	}

	logLevel.
		Int("status", problem.Status).
		Str("type", problem.Type).
		Str("title", problem.Title).
		Str("detail", problem.Detail).
		Str("ip", c.IP()).
		Str("path", c.Path()).
		Str("method", c.Method()).
		Interface("extensions", problem.Extensions).
		Msg("Error response sent")

	// Устанавливаем Content-Type согласно RFC 7807
	c.Set("Content-Type", "application/problem+json")

	return c.Status(problem.Status).JSON(problem)
}

// Предопределенные ошибки

func ValidationError(detail string) *ProblemDetail {
	return NewProblemDetail(
		ErrorTypeValidation,
		"Validation Failed",
		detail,
		http.StatusBadRequest,
	)
}

func AuthenticationError(detail string) *ProblemDetail {
	return NewProblemDetail(
		ErrorTypeAuthentication,
		"Authentication Required",
		detail,
		http.StatusUnauthorized,
	)
}

func AuthorizationError(detail string) *ProblemDetail {
	return NewProblemDetail(
		ErrorTypeAuthorization,
		"Insufficient Permissions",
		detail,
		http.StatusForbidden,
	)
}

func NotFoundError(resource string) *ProblemDetail {
	return NewProblemDetail(
		ErrorTypeNotFound,
		"Resource Not Found",
		resource+" not found",
		http.StatusNotFound,
	)
}

func ConflictError(detail string) *ProblemDetail {
	return NewProblemDetail(
		ErrorTypeConflict,
		"Resource Conflict",
		detail,
		http.StatusConflict,
	)
}

func RateLimitError(detail string) *ProblemDetail {
	return NewProblemDetail(
		ErrorTypeRateLimit,
		"Rate Limit Exceeded",
		detail,
		http.StatusTooManyRequests,
	)
}

func InternalServerError(detail string) *ProblemDetail {
	return NewProblemDetail(
		ErrorTypeInternal,
		"Internal Server Error",
		detail,
		http.StatusInternalServerError,
	)
}

func BadRequestError(detail string) *ProblemDetail {
	return NewProblemDetail(
		ErrorTypeBadRequest,
		"Bad Request",
		detail,
		http.StatusBadRequest,
	)
}

func PaymentError(detail string) *ProblemDetail {
	return NewProblemDetail(
		ErrorTypePayment,
		"Payment Error",
		detail,
		http.StatusPaymentRequired,
	)
}

// ValidationErrorWithFields создает ошибку валидации с деталями полей
func ValidationErrorWithFields(fields []ValidationFieldError) *ProblemDetail {
	problem := ValidationError("One or more validation errors occurred")
	problem.WithExtension("validation_errors", fields)
	return problem
}

// ValidationFieldError представляет ошибку валидации поля
type ValidationFieldError struct {
	Field   string      `json:"field"`
	Tag     string      `json:"tag"`
	Value   interface{} `json:"value"`
	Message string      `json:"message"`
}

// CreateValidationFieldError создает ошибку валидации поля
func CreateValidationFieldError(field, tag string, value interface{}, message string) ValidationFieldError {
	return ValidationFieldError{
		Field:   field,
		Tag:     tag,
		Value:   value,
		Message: message,
	}
}

// ErrorHandler middleware для обработки ошибок
func ErrorHandler() fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		// Если это уже ProblemDetail, отправляем как есть
		if problem, ok := err.(*ProblemDetail); ok {
			return SendError(c, problem)
		}

		// Если это fiber.Error, конвертируем
		if fiberErr, ok := err.(*fiber.Error); ok {
			var errorType ErrorType
			switch fiberErr.Code {
			case fiber.StatusBadRequest:
				errorType = ErrorTypeBadRequest
			case fiber.StatusUnauthorized:
				errorType = ErrorTypeAuthentication
			case fiber.StatusForbidden:
				errorType = ErrorTypeAuthorization
			case fiber.StatusNotFound:
				errorType = ErrorTypeNotFound
			case fiber.StatusConflict:
				errorType = ErrorTypeConflict
			case fiber.StatusTooManyRequests:
				errorType = ErrorTypeRateLimit
			default:
				errorType = ErrorTypeInternal
			}

			problem := NewProblemDetail(
				errorType,
				http.StatusText(fiberErr.Code),
				fiberErr.Message,
				fiberErr.Code,
			)

			return SendError(c, problem)
		}

		// Для всех других ошибок создаем internal server error
		problem := InternalServerError("An unexpected error occurred")
		problem.WithExtension("original_error", err.Error())

		return SendError(c, problem)
	}
}
