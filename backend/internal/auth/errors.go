package auth

import (
	"errors"
	"net/http"
)

// APIError представляет ошибку с HTTP статусом и кодом
type APIError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"-"`
}

func (e APIError) Error() string {
	return e.Message
}

var (
	ErrInvalidRefreshToken = APIError{
		Code:       "INVALID_REFRESH_TOKEN",
		Message:    "Invalid or expired refresh token",
		HTTPStatus: http.StatusUnauthorized,
	}
	ErrInvalidTokenRole = APIError{
		Code:       "INVALID_TOKEN_ROLE",
		Message:    "Invalid token role",
		HTTPStatus: http.StatusForbidden,
	}
	ErrRefreshTokens = APIError{
		Code:       "REFRESH_TOKENS_FAILED",
		Message:    "Failed to refresh tokens",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrInvalidCredentials = APIError{
		Code:       "INVALID_CREDENTIALS",
		Message:    "Invalid credentials",
		HTTPStatus: http.StatusUnauthorized,
	}
	ErrCreateTokenPair = APIError{
		Code:       "TOKEN_CREATION_FAILED",
		Message:    "Failed to create token pair",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrUpdateLastLogin = APIError{
		Code:       "UPDATE_LOGIN_FAILED",
		Message:    "Failed to update last login time",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrPartnerNotFound = APIError{
		Code:       "PARTNER_NOT_FOUND",
		Message:    "Partner not found",
		HTTPStatus: http.StatusNotFound,
	}
	ErrAdminNotFound = APIError{
		Code:       "ADMIN_NOT_FOUND",
		Message:    "Admin not found",
		HTTPStatus: http.StatusNotFound,
	}
	ErrPartnerBlocked = APIError{
		Code:       "PARTNER_BLOCKED",
		Message:    "Partner account is blocked",
		HTTPStatus: http.StatusForbidden,
	}
	ErrInvalidRequestBody = APIError{
		Code:       "INVALID_REQUEST_BODY",
		Message:    "Invalid request body",
		HTTPStatus: http.StatusBadRequest,
	}
	ErrAdminLoginFailed = APIError{
		Code:       "ADMIN_LOGIN_FAILED",
		Message:    "Admin login failed",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrAdminTokenRefreshFailed = APIError{
		Code:       "ADMIN_TOKEN_REFRESH_FAILED",
		Message:    "Admin token refresh failed",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrValidationFailed = APIError{
		Code:       "VALIDATION_FAILED",
		Message:    "Validation failed",
		HTTPStatus: http.StatusBadRequest,
	}
	ErrPartnerLoginFailed = APIError{
		Code:       "PARTNER_LOGIN_FAILED",
		Message:    "Partner login failed",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrPartnerTokenRefreshFailed = APIError{
		Code:       "PARTNER_TOKEN_REFRESH_FAILED",
		Message:    "Partner token refresh failed",
		HTTPStatus: http.StatusInternalServerError,
	}
)

// IsAPIError проверяет, является ли ошибка типом APIError
func IsAPIError(err error) bool {
	var apiErr APIError
	return errors.As(err, &apiErr)
}

// GetAPIError извлекает APIError из ошибки
func GetAPIError(err error) (APIError, bool) {
	var apiErr APIError
	if errors.As(err, &apiErr) {
		return apiErr, true
	}
	return APIError{}, false
}
