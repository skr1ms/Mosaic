package errors

import (
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

type ProblemDetail struct {
	Type       string         `json:"type"`
	Title      string         `json:"title"`
	Status     int            `json:"status"`
	Detail     string         `json:"detail"`
	Instance   string         `json:"instance"`
	Extensions map[string]any `json:"extensions,omitempty"`
}

// Error implements error interface
func (p *ProblemDetail) Error() string {
	return p.Detail
}

// ErrorType defines error types
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

// GetBaseURI returns base URI for error types
func GetBaseURI() string {
	return "https://api.mosaic.com/errors/"
}

// NewProblemDetail creates new ProblemDetail
func NewProblemDetail(errorType ErrorType, title, detail string, status int) *ProblemDetail {
	return &ProblemDetail{
		Type:   GetBaseURI() + string(errorType),
		Title:  title,
		Status: status,
		Detail: detail,
	}
}

// WithInstance adds instance URI
func (p *ProblemDetail) WithInstance(instance string) *ProblemDetail {
	p.Instance = instance
	return p
}

// WithExtension adds additional field
func (p *ProblemDetail) WithExtension(key string, value any) *ProblemDetail {
	if p.Extensions == nil {
		p.Extensions = make(map[string]any)
	}
	p.Extensions[key] = value
	return p
}

// SendError sends standardized error
func SendError(c *fiber.Ctx, problem *ProblemDetail) error {
	problem.WithExtension("timestamp", time.Now().UTC().Format(time.RFC3339))

	if requestID := c.Get("X-Request-ID"); requestID != "" {
		problem.WithExtension("request_id", requestID)
	}

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

	c.Set("Content-Type", "application/problem+json")

	return c.Status(problem.Status).JSON(problem)
}

// Predefined errors
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
	log.Error().Str("detail", detail).Msg("Creating internal server error")
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
	log.Error().Str("detail", detail).Msg("Creating payment error")
	return NewProblemDetail(
		ErrorTypePayment,
		"Payment Error",
		detail,
		http.StatusPaymentRequired,
	)
}

// ValidationErrorWithFields creates validation error with field details
func ValidationErrorWithFields(fields []ValidationFieldError) *ProblemDetail {
	problem := ValidationError("One or more validation errors occurred")
	problem.WithExtension("validation_errors", fields)
	return problem
}

// ValidationFieldError represents field validation error
type ValidationFieldError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Value   any    `json:"value"`
	Message string `json:"message"`
}

// CreateValidationFieldError creates field validation error
func CreateValidationFieldError(field, tag string, value any, message string) ValidationFieldError {
	return ValidationFieldError{
		Field:   field,
		Tag:     tag,
		Value:   value,
		Message: message,
	}
}

func ErrorHandler() fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		if problem, ok := err.(*ProblemDetail); ok {
			return SendError(c, problem)
		}

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

		log.Error().Err(err).Msg("Unexpected error occurred")
		problem := InternalServerError("An unexpected error occurred")
		problem.WithExtension("original_error", err.Error())

		return SendError(c, problem)
	}
}
