package image

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
	ErrImageNotFound = APIError{
		Code:       "IMAGE_NOT_FOUND",
		Message:    "Image not found",
		HTTPStatus: http.StatusNotFound,
	}
	ErrImageUploadFailed = APIError{
		Code:       "IMAGE_UPLOAD_FAILED",
		Message:    "Failed to upload image",	
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrImageProcessingFailed = APIError{
		Code:       "IMAGE_PROCESSING_FAILED",
		Message:    "Failed to process image",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrFailedToFetchQueue = APIError{
		Code:       "QUEUE_FETCH_FAILED",
		Message:    "Failed to fetch image queue",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrCouponNotFound = APIError{
		Code:       "COUPON_NOT_FOUND",
		Message:    "Coupon not found",
		HTTPStatus: http.StatusNotFound,
	}
	ErrCouponAlreadyInQueue = APIError{
		Code:       "COUPON_ALREADY_IN_QUEUE",
		Message:    "Coupon already in queue",
		HTTPStatus: http.StatusConflict,
	}
	ErrCouponAlreadyProcessed = APIError{
		Code:       "COUPON_ALREADY_PROCESSED",
		Message:    "Coupon already processed",
		HTTPStatus: http.StatusConflict,
	}
	ErrFailedToAddToQueue = APIError{
		Code:       "FAILED_TO_ADD_TO_QUEUE",
		Message:    "Failed to add task to queue",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrFailedToCreateTask = APIError{
		Code:       "FAILED_TO_CREATE_TASK",
		Message:    "Failed to create task",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrNoTasksInQueue = APIError{
		Code:       "NO_TASKS_IN_QUEUE",
		Message:    "No tasks in queue",
		HTTPStatus: http.StatusNotFound,
	}
	ErrFailedToGetNextTask = APIError{
		Code:       "FAILED_TO_GET_NEXT_TASK",
		Message:    "Failed to get next task",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrInvalidTaskID = APIError{
		Code:       "INVALID_TASK_ID",
		Message:    "Invalid task ID",
		HTTPStatus: http.StatusBadRequest,
	}
	ErrFailedToStartProcessing = APIError{
		Code:       "FAILED_TO_START_PROCESSING",
		Message:    "Failed to start processing",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrFailedToCompleteProcessing = APIError{
		Code:       "FAILED_TO_COMPLETE_PROCESSING",
		Message:    "Failed to complete processing",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrFailedToMarkAsFailed = APIError{
		Code:       "FAILED_TO_MARK_AS_FAILED",
		Message:    "Failed to mark as failed",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrTaskNotFound = APIError{
		Code:       "TASK_NOT_FOUND",
		Message:    "Task not found",
		HTTPStatus: http.StatusNotFound,
	}
	ErrBadRequest = APIError{
		Code:       "BAD_REQUEST",
		Message:    "Bad request",
		HTTPStatus: http.StatusBadRequest,
	}
	ErrFailedToRetryTask = APIError{
		Code:       "FAILED_TO_RETRY_TASK",
		Message:    "Failed to retry task",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrFailedToDeleteTask = APIError{
		Code:       "FAILED_TO_DELETE_TASK",
		Message:    "Failed to delete task",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrFailedToGetStatistics = APIError{
		Code:       "FAILED_TO_GET_STATISTICS",
		Message:    "Failed to get statistics",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrSchemaNotConfirmed = APIError{
		Code:       "SCHEMA_NOT_CONFIRMED",
		Message:    "Schema generation not confirmed",
		HTTPStatus: http.StatusBadRequest,
	}
	ErrSchemaGenerationFailed = APIError{
		Code:       "SCHEMA_GENERATION_FAILED",
		Message:    "Failed to generate schema",
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