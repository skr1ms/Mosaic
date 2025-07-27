package admin

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
	ErrAdminNotFound = APIError{
		Code:       "ADMIN_NOT_FOUND",
		Message:    "Admin not found",
		HTTPStatus: http.StatusNotFound,
	}
	ErrAdminAlreadyExists = APIError{
		Code:       "ADMIN_ALREADY_EXISTS",
		Message:    "Admin with this login already exists",
		HTTPStatus: http.StatusConflict,
	}
	ErrFailedToCreateAdmin = APIError{
		Code:       "ADMIN_CREATION_FAILED",
		Message:    "Failed to create admin",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrFailedToGetAdmins = APIError{
		Code:       "ADMINS_FETCH_FAILED",
		Message:    "Failed to get admins list",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrFailedToChangePassword = APIError{
		Code:       "PASSWORD_CHANGE_FAILED",
		Message:    "Failed to change password",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrInvalidPassword = APIError{
		Code:       "INVALID_PASSWORD",
		Message:    "Invalid current password",
		HTTPStatus: http.StatusBadRequest,
	}
	ErrPasswordHashingFailed = APIError{
		Code:       "PASSWORD_HASHING_FAILED",
		Message:    "Failed to hash password",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrFailedToGetPartners = APIError{
		Code:       "PARTNERS_FETCH_FAILED",
		Message:    "Failed to get partners list",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrFailedToCreatePartner = APIError{
		Code:       "PARTNER_CREATION_FAILED",
		Message:    "Failed to create partner",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrFailedToUpdatePartner = APIError{
		Code:       "PARTNER_UPDATE_FAILED",
		Message:    "Failed to update partner",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrFailedToDeletePartner = APIError{
		Code:       "PARTNER_DELETE_FAILED",
		Message:    "Failed to delete partner",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrFailedToBlockPartner = APIError{
		Code:       "PARTNER_BLOCK_FAILED",
		Message:    "Failed to block partner",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrFailedToUnblockPartner = APIError{
		Code:       "PARTNER_UNBLOCK_FAILED",
		Message:    "Failed to unblock partner",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrPartnerNotFound = APIError{
		Code:       "PARTNER_NOT_FOUND",
		Message:    "Partner not found",
		HTTPStatus: http.StatusNotFound,
	}
	ErrFailedToGetCoupons = APIError{
		Code:       "COUPONS_FETCH_FAILED",
		Message:    "Failed to get coupons",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrFailedToCreateCoupons = APIError{
		Code:       "COUPONS_CREATION_FAILED",
		Message:    "Failed to create coupons",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrFailedToDeleteCoupon = APIError{
		Code:       "COUPON_DELETE_FAILED",
		Message:    "Failed to delete coupon",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrFailedToResetCoupon = APIError{
		Code:       "COUPON_RESET_FAILED",
		Message:    "Failed to reset coupon",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrFailedToBatchDeleteCoupons = APIError{
		Code:       "COUPONS_BATCH_DELETE_FAILED",
		Message:    "Failed to batch delete coupons",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrCouponNotFound = APIError{
		Code:       "COUPON_NOT_FOUND",
		Message:    "Coupon not found",
		HTTPStatus: http.StatusNotFound,
	}
	ErrFailedToGetStatistics = APIError{
		Code:       "STATISTICS_FETCH_FAILED",
		Message:    "Failed to get statistics",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrFailedToGetPartnerStatistics = APIError{
		Code:       "PARTNER_STATISTICS_FETCH_FAILED",
		Message:    "Failed to get partner statistics",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrFailedToGetSystemStatistics = APIError{
		Code:       "SYSTEM_STATISTICS_FETCH_FAILED",
		Message:    "Failed to get system statistics",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrFailedToGetAnalytics = APIError{
		Code:       "ANALYTICS_FETCH_FAILED",
		Message:    "Failed to get analytics",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrFailedToGetDashboard = APIError{
		Code:       "DASHBOARD_FETCH_FAILED",
		Message:    "Failed to get dashboard data",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrFailedToGetImages = APIError{
		Code:       "IMAGES_FETCH_FAILED",
		Message:    "Failed to get images",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrFailedToGetImageDetails = APIError{
		Code:       "IMAGE_DETAILS_FETCH_FAILED",
		Message:    "Failed to get image details",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrFailedToDeleteImageTask = APIError{
		Code:       "IMAGE_TASK_DELETE_FAILED",
		Message:    "Failed to delete image task",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrFailedToRetryImageTask = APIError{
		Code:       "IMAGE_TASK_RETRY_FAILED",
		Message:    "Failed to retry image task",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrImageTaskNotFound = APIError{
		Code:       "IMAGE_TASK_NOT_FOUND",
		Message:    "Image task not found",
		HTTPStatus: http.StatusNotFound,
	}
	ErrInvalidRequest = APIError{
		Code:       "INVALID_REQUEST",
		Message:    "Invalid request data",
		HTTPStatus: http.StatusBadRequest,
	}
	ErrInvalidID = APIError{
		Code:       "INVALID_ID",
		Message:    "Invalid ID format",
		HTTPStatus: http.StatusBadRequest,
	}
	ErrFailedToExportCoupons = APIError{
		Code:       "COUPONS_EXPORT_FAILED",
		Message:    "Failed to export coupons",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrFailedToFindAllAdmins = APIError{
		Code:       "ADMINS_FETCH_FAILED",
		Message:    "Failed to get admins list",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrFailedToFindAdminByLogin = APIError{
		Code:       "ADMIN_FETCH_FAILED",
		Message:    "Failed to get admin by login",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrFailedToFindAdminByID = APIError{
		Code:       "ADMIN_FETCH_FAILED",
		Message:    "Failed to get admin by ID",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrFailedToFindAllPartners = APIError{
		Code:       "PARTNERS_FETCH_FAILED",
		Message:    "Failed to get partners list",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrFailedToFindPartnerByLogin = APIError{
		Code:       "PARTNER_FETCH_FAILED",
		Message:    "Failed to get partner by login",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrFailedToFindPartnerByID = APIError{
		Code:       "PARTNER_FETCH_FAILED",
		Message:    "Failed to get partner by ID",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrFailedToFindAllCoupons = APIError{
		Code:       "COUPONS_FETCH_FAILED",
		Message:    "Failed to get coupons list",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrFailedToFindCouponByID = APIError{
		Code:       "COUPON_FETCH_FAILED",
		Message:    "Failed to get coupon by ID",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrFailedToFindAllImages = APIError{
		Code:       "IMAGES_FETCH_FAILED",
		Message:    "Failed to get images list",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrFailedToFindImageByID = APIError{
		Code:       "IMAGE_FETCH_FAILED",
		Message:    "Failed to get image by ID",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrFailedToFindImageByTaskID = APIError{
		Code:       "IMAGE_FETCH_FAILED",
		Message:    "Failed to get image by task ID",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrFailedToFindImageByCouponID = APIError{
		Code:       "IMAGE_FETCH_FAILED",
		Message:    "Failed to get image by coupon ID",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrFailedToUpdateImage = APIError{
		Code:       "IMAGE_UPDATE_FAILED",
		Message:    "Failed to update image",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrImageNotFound = APIError{
		Code:       "IMAGE_NOT_FOUND",
		Message:    "Image not found",
		HTTPStatus: http.StatusNotFound,
	}
	ErrFailedToDeleteImage = APIError{
		Code:       "IMAGE_DELETE_FAILED",
		Message:    "Failed to delete image",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrFailedToUnblockCoupons = APIError{
		Code:       "COUPONS_UNBLOCK_FAILED",
		Message:    "Failed to unblock coupons",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrFailedToResetCoupons = APIError{
		Code:       "COUPONS_RESET_FAILED",
		Message:    "Failed to reset coupons",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrFailedToDeleteCoupons = APIError{
		Code:       "COUPONS_DELETE_FAILED",
		Message:    "Failed to delete coupons",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrPartnerAlreadyExists = APIError{
		Code:       "PARTNER_ALREADY_EXISTS",
		Message:    "Partner with this login already exists",
		HTTPStatus: http.StatusConflict,
	}
	ErrFailedToGeneratePartnerCode = APIError{
		Code:       "PARTNER_CODE_GENERATION_FAILED",
		Message:    "Failed to generate partner code",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrMaxPartnerCodeReached = APIError{
		Code:       "PARTNER_CODE_MAX_REACHED",
		Message:    "Max partner code reached",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrFailedToBlockCoupons = APIError{
		Code:       "COUPONS_BLOCK_FAILED",
		Message:    "Failed to block coupons",
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
