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
	ErrInvalidRequestBody = APIError{
		Code:       "INVALID_REQUEST_BODY",
		Message:    "Invalid request body",
		HTTPStatus: http.StatusBadRequest,
	}
	ErrFailedToResetCoupon = APIError{
		Code:       "FAILED_TO_RESET_COUPON",
		Message:    "Failed to reset coupon",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrFailedToMarkAsPurchased = APIError{
		Code:       "FAILED_TO_MARK_AS_PURCHASED",
		Message:    "Failed to mark as purchased",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrFailedToSendSchema = APIError{
		Code:       "FAILED_TO_SEND_SCHEMA",
		Message:    "Failed to send schema",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrFailedToGetStatistics = APIError{
		Code:       "FAILED_TO_GET_STATISTICS",
		Message:    "Failed to get statistics",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrCouponNotFound = APIError{
		Code:       "COUPON_NOT_FOUND",
		Message:    "Coupon not found",
		HTTPStatus: http.StatusNotFound,
	}
	ErrCouponAlreadyUsed = APIError{
		Code:       "COUPON_ALREADY_USED",
		Message:    "Coupon already used",
		HTTPStatus: http.StatusBadRequest,
	}
	ErrFailedToFetchCouponsForExport = APIError{
		Code:       "FAILED_TO_FETCH_COUPONS_FOR_EXPORT",
		Message:    "Failed to fetch coupons for export",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrCouponMustBeUsedToDownloadMaterials = APIError{
		Code:       "COUPON_MUST_BE_USED_TO_DOWNLOAD_MATERIALS",
		Message:    "Coupon must be used to download materials",
		HTTPStatus: http.StatusBadRequest,
	}
	ErrFailedToCreateArchive = APIError{
		Code:       "FAILED_TO_CREATE_ARCHIVE",
		Message:    "Failed to create archive",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrFailedToFetchCoupons = APIError{
		Code:       "FAILED_TO_FETCH_COUPONS",
		Message:    "Failed to fetch coupons",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrInvalidPartnerID = APIError{
		Code:       "INVALID_PARTNER_ID",
		Message:    "Invalid partner ID",
		HTTPStatus: http.StatusBadRequest,
	}
	ErrFailedToFetchPartnerCoupons = APIError{
		Code:       "FAILED_TO_FETCH_PARTNER_COUPONS",
		Message:    "Failed to fetch partner coupons",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrFailedToFindCouponsByPartnerID = APIError{
		Code:       "FAILED_TO_FIND_COUPONS_BY_PARTNER_ID",
		Message:    "Failed to find coupons by partner ID",
		HTTPStatus: http.StatusNotFound,
	}
	ErrFailedToFindAllCoupons = APIError{
		Code:       "FAILED_TO_FIND_ALL_COUPONS",
		Message:    "Failed to find all coupons",
		HTTPStatus: http.StatusNotFound,
	}
	ErrFailedToFindCouponsWithPagination = APIError{
		Code:       "FAILED_TO_FIND_COUPONS_WITH_PAGINATION",
		Message:    "Failed to find coupons with pagination",
		HTTPStatus: http.StatusNotFound,
	}
	ErrFailedToFindRecentActivatedCoupons = APIError{
		Code:       "FAILED_TO_FIND_RECENT_ACTIVATED_COUPONS",
		Message:    "Failed to find recent activated coupons",
		HTTPStatus: http.StatusNotFound,
	}
	ErrFailedToCheckCodeExists = APIError{
		Code:       "FAILED_TO_CHECK_CODE_EXISTS",
		Message:    "Failed to check code exists",
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
