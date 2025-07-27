package public

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
		Code:       "PARTNER_NOT_FOUND",
		Message:    "Партнер не найден",
		HTTPStatus: http.StatusNotFound,
	}
	ErrCouponNotFound = APIError{
		Code:       "COUPON_NOT_FOUND",
		Message:    "Купон не найден",
		HTTPStatus: http.StatusNotFound,
	}
	ErrInvalidCouponCode = APIError{
		Code:       "INVALID_COUPON_CODE",
		Message:    "Неверный формат кода купона. Код должен содержать 12 цифр",
		HTTPStatus: http.StatusBadRequest,
	}
	ErrCouponAlreadyUsed = APIError{
		Code:       "COUPON_ALREADY_USED",
		Message:    "Купон уже использован",
		HTTPStatus: http.StatusConflict,
	}
	ErrCouponNotActivated = APIError{
		Code:       "COUPON_NOT_ACTIVATED",
		Message:    "Купон не активирован",
		HTTPStatus: http.StatusBadRequest,
	}
	ErrFailedToActivateCoupon = APIError{
		Code:       "FAILED_TO_ACTIVATE_COUPON",
		Message:    "Ошибка при активации купона",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrFailedToCreateCoupon = APIError{
		Code:       "FAILED_TO_CREATE_COUPON",
		Message:    "Ошибка при создании купона",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrFailedToGenerateCouponCode = APIError{
		Code:       "FAILED_TO_GENERATE_COUPON_CODE",
		Message:    "Ошибка при генерации уникального кода купона",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrImageNotFound = APIError{
		Code:       "IMAGE_NOT_FOUND",
		Message:    "Изображение не найдено",
		HTTPStatus: http.StatusNotFound,
	}
	ErrInvalidImageID = APIError{
		Code:       "INVALID_IMAGE_ID",
		Message:    "Неверный формат ID изображения",
		HTTPStatus: http.StatusBadRequest,
	}
	ErrInvalidCouponID = APIError{
		Code:       "INVALID_COUPON_ID",
		Message:    "Неверный формат ID купона",
		HTTPStatus: http.StatusBadRequest,
	}
	ErrCouponIDRequired = APIError{
		Code:       "COUPON_ID_REQUIRED",
		Message:    "ID купона обязателен",
		HTTPStatus: http.StatusBadRequest,
	}
	ErrImageFileRequired = APIError{
		Code:       "IMAGE_FILE_REQUIRED",
		Message:    "Файл изображения обязателен",
		HTTPStatus: http.StatusBadRequest,
	}
	ErrInvalidImageType = APIError{
		Code:       "INVALID_IMAGE_TYPE",
		Message:    "Поддерживаются только файлы JPG и PNG",
		HTTPStatus: http.StatusBadRequest,
	}
	ErrFileTooLarge = APIError{
		Code:       "FILE_TOO_LARGE",
		Message:    "Размер файла не должен превышать 10MB",
		HTTPStatus: http.StatusRequestEntityTooLarge,
	}
	ErrFailedToSaveFile = APIError{
		Code:       "FAILED_TO_SAVE_FILE",
		Message:    "Ошибка при сохранении файла",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrFailedToCreateImageTask = APIError{
		Code:       "FAILED_TO_CREATE_IMAGE_TASK",
		Message:    "Ошибка при создании задачи обработки",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrFailedToEditImage = APIError{
		Code:       "FAILED_TO_EDIT_IMAGE",
		Message:    "Ошибка при редактировании изображения",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrFailedToProcessImage = APIError{
		Code:       "FAILED_TO_PROCESS_IMAGE",
		Message:    "Ошибка при обработке изображения",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrFailedToGenerateSchema = APIError{
		Code:       "FAILED_TO_GENERATE_SCHEMA",
		Message:    "Ошибка при создании схемы",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrSchemaNotReady = APIError{
		Code:       "SCHEMA_NOT_READY",
		Message:    "Схема еще не готова",
		HTTPStatus: http.StatusBadRequest,
	}
	ErrSchemaFileNotFound = APIError{
		Code:       "SCHEMA_FILE_NOT_FOUND",
		Message:    "Файл схемы не найден",
		HTTPStatus: http.StatusNotFound,
	}
	ErrFailedToSendEmail = APIError{
		Code:       "FAILED_TO_SEND_EMAIL",
		Message:    "Ошибка при отправке email",
		HTTPStatus: http.StatusInternalServerError,
	}

	// Ошибки валидации
	ErrInvalidRequest = APIError{
		Code:       "INVALID_REQUEST",
		Message:    "Ошибка в запросе",
		HTTPStatus: http.StatusBadRequest,
	}

	// Общие ошибки
	ErrInternalServerError = APIError{
		Code:       "INTERNAL_SERVER_ERROR",
		Message:    "Внутренняя ошибка сервера",
		HTTPStatus: http.StatusInternalServerError,
	}
)

// NewAPIError создает новую API ошибку
func NewAPIError(code, message string, httpStatus int) APIError {
	return APIError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

// IsAPIError проверяет, является ли ошибка API ошибкой
func IsAPIError(err error) (APIError, bool) {
	var apiErr APIError
	if errors.As(err, &apiErr) {
		return apiErr, true
	}
	return APIError{}, false
}
