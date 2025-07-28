package errors

import "fmt"

// AppError представляет ошибку приложения с контекстом
type AppError struct {
	Code     string                 `json:"code"`
	Message  string                 `json:"message"`
	Details  map[string]interface{} `json:"details,omitempty"`
	Internal error                  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Internal != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Internal)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Internal
}

// NewAppError создает новую ошибку приложения
func NewAppError(code, message string, internal error) *AppError {
	return &AppError{
		Code:     code,
		Message:  message,
		Internal: internal,
		Details:  make(map[string]interface{}),
	}
}

// WithDetails добавляет детали к ошибке
func (e *AppError) WithDetails(details map[string]interface{}) *AppError {
	for k, v := range details {
		e.Details[k] = v
	}
	return e
}

// Предопределенные коды ошибок
const (
	CodeCouponNotFound    = "COUPON_NOT_FOUND"
	CodeCouponAlreadyUsed = "COUPON_ALREADY_USED"
	CodeImageNotFound     = "IMAGE_NOT_FOUND"
	CodePartnerNotFound   = "PARTNER_NOT_FOUND"
	CodeInvalidInput      = "INVALID_INPUT"
	CodeInternalError     = "INTERNAL_ERROR"
	CodeUnauthorized      = "UNAUTHORIZED"
	CodeForbidden         = "FORBIDDEN"
)

// Convenience functions для часто используемых ошибок
func CouponNotFound(err error) *AppError {
	return NewAppError(CodeCouponNotFound, "Coupon not found", err)
}

func CouponAlreadyUsed(err error) *AppError {
	return NewAppError(CodeCouponAlreadyUsed, "Coupon already used", err)
}

func ImageNotFound(err error) *AppError {
	return NewAppError(CodeImageNotFound, "Image not found", err)
}

func PartnerNotFound(err error) *AppError {
	return NewAppError(CodePartnerNotFound, "Partner not found", err)
}

func InvalidInput(message string, err error) *AppError {
	return NewAppError(CodeInvalidInput, message, err)
}

func InternalError(err error) *AppError {
	return NewAppError(CodeInternalError, "Internal server error", err)
}
