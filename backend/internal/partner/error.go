package partner

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
	ErrPartnerNotFound = APIError{
		Code:       "USER_NOT_FOUND",
		Message:    "Partner not found",
		HTTPStatus: http.StatusNotFound,
	}
	ErrInvalidCaptcha = APIError{
		Code:       "INVALID_CAPTCHA",
		Message:    "Invalid captcha",
		HTTPStatus: http.StatusBadRequest,
	}
	ErrFailedToCreateToken = APIError{
		Code:       "TOKEN_CREATION_FAILED",
		Message:    "Failed to create authentication token",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrFailedToSendEmail = APIError{
		Code:       "EMAIL_SEND_FAILED",
		Message:    "Failed to send email",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrFailedToHashPassword = APIError{
		Code:       "PASSWORD_HASH_FAILED",
		Message:    "Failed to process password",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrFailedToUpdatePassword = APIError{
		Code:       "PASSWORD_UPDATE_FAILED",
		Message:    "Failed to update password",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrInvalidToken = APIError{
		Code:       "INVALID_TOKEN",
		Message:    "Invalid or expired token",
		HTTPStatus: http.StatusUnauthorized,
	}
	ErrBadRequest = APIError{
		Code:       "BAD_REQUEST",
		Message:    "Invalid request data",
		HTTPStatus: http.StatusBadRequest,
	}
	ErrValidationFailed = APIError{
		Code:       "VALIDATION_FAILED",
		Message:    "Validation failed",
		HTTPStatus: http.StatusBadRequest,
	}
	ErrFailedToFindPartnerByEmail = APIError{
		Code:       "PARTNER_NOT_FOUND_BY_EMAIL",
		Message:    "Partner not found",
		HTTPStatus: http.StatusNotFound,
	}
	ErrFailedToCreatePartner = APIError{
		Code:       "PARTNER_CREATION_FAILED",
		Message:    "Failed to create partner",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrFailedToFindPartnerByID = APIError{
		Code:       "PARTNER_NOT_FOUND_BY_ID",
		Message:    "Partner not found",
		HTTPStatus: http.StatusNotFound,
	}
	ErrPartnerStatusNotActive = APIError{
		Code:       "PARTNER_INACTIVE",
		Message:    "Partner account is not active",
		HTTPStatus: http.StatusForbidden,
	}
	ErrFailedToFetchCoupons = APIError{
		Code:       "COUPONS_FETCH_FAILED",
		Message:    "Failed to fetch coupons",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrNoCouponsFound = APIError{
		Code:       "NO_COUPONS_FOUND",
		Message:    "No coupons found",
		HTTPStatus: http.StatusNotFound,
	}
	ErrFailedToGetJwtClaims = APIError{
		Code:       "JWT_CLAIMS_FAILED",
		Message:    "Failed to get authentication claims",
		HTTPStatus: http.StatusUnauthorized,
	}
	ErrCurrentPasswordIsIncorrect = APIError{
		Code:       "INCORRECT_PASSWORD",
		Message:    "Current password is incorrect",
		HTTPStatus: http.StatusBadRequest,
	}
	ErrFailedToGetCoupons = APIError{
		Code:       "COUPONS_FETCH_FAILED",
		Message:    "Failed to fetch coupons",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrFailedToExportCoupons = APIError{
		Code:       "COUPONS_EXPORT_FAILED",
		Message:    "Failed to export coupons",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrFailedToResetPassword = APIError{
		Code:       "PASSWORD_RESET_FAILED",
		Message:    "Failed to reset password",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrFailedToSendForgotPasswordEmail = APIError{
		Code:       "FORGOT_PASSWORD_EMAIL_FAILED",
		Message:    "Failed to send forgot password email",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrFailedToDeleteCoupons = APIError{
		Code:       "FAILED_TO_DELETE_COUPONS",
		Message:    "Failed to delete coupons",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrFailedToDeletePartner = APIError{
		Code:       "FAILED_TO_DELETE_PARTNER",
		Message:    "Failed to delete partner",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrFailedToFindPartnerByPartnerCode = APIError{
		Code:       "FAILED_TO_FIND_PARTNER_BY_PARTNER_CODE",
		Message:    "Failed to find partner by partner code",
		HTTPStatus: http.StatusNotFound,
	}
	ErrFailedToFindPartnerByLogin = APIError{
		Code:       "FAILED_TO_FIND_PARTNER_BY_LOGIN",
		Message:    "Failed to find partner by login",
		HTTPStatus: http.StatusNotFound,
	}
	ErrFailedToFindPartnerByDomain = APIError{
		Code:       "FAILED_TO_FIND_PARTNER_BY_DOMAIN",
		Message:    "Failed to find partner by domain",
		HTTPStatus: http.StatusNotFound,
	}
	ErrFailedToFindAllPartners = APIError{
		Code:       "FAILED_TO_FIND_ALL_PARTNERS",
		Message:    "Failed to find all partners",
		HTTPStatus: http.StatusNotFound,
	}
	ErrFailedToFindActivePartners = APIError{
		Code:       "FAILED_TO_FIND_ACTIVE_PARTNERS",
		Message:    "Failed to find active partners",
		HTTPStatus: http.StatusNotFound,
	}
	ErrFailedToFindPartners = APIError{
		Code:       "FAILED_TO_FIND_PARTNERS",
		Message:    "Failed to find partners",
		HTTPStatus: http.StatusNotFound,
	}
	ErrFailedToFindPartnerCouponsForExport = APIError{
		Code:       "FAILED_TO_FIND_PARTNER_COUPONS_FOR_EXPORT",
		Message:    "Failed to find partner coupons for export",
		HTTPStatus: http.StatusNotFound,
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
