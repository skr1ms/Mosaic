package coupon

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
	ErrFailedToFindCoupons = APIError{
		Code:       "COUPONS_FETCH_FAILED",
		Message:    "Failed to find coupons",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrCouponNotBelongsToPartner = APIError{
		Code:       "COUPON_ACCESS_DENIED",
		Message:    "Coupon does not belong to partner",
		HTTPStatus: http.StatusForbidden,
	}
	ErrFailedToActivateCoupon = APIError{
		Code:       "COUPON_ACTIVATION_FAILED",
		Message:    "Failed to activate coupon",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrInvalidCouponID = APIError{
		Code:       "INVALID_COUPON_ID",
		Message:    "Invalid coupon ID",
		HTTPStatus: http.StatusBadRequest,
	}
	ErrFailedToDownloadFile = APIError{
		Code:       "FAILED_TO_DOWNLOAD_FILE",
		Message:    "Failed to download file",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrFailedToCreateFileWriter = APIError{
		Code:       "FAILED_TO_CREATE_FILE_WRITER",
		Message:    "Failed to create file writer",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrFailedToCopyFileToZip = APIError{
		Code:       "FAILED_TO_COPY_FILE_TO_ZIP",
		Message:    "Failed to copy file to zip",
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
