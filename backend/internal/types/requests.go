package types

import "github.com/google/uuid"

// Общие структуры для всех модулей, чтобы избежать циклических импортов

// UploadImageRequest - запрос загрузки изображения
type UploadImageRequest struct {
	CouponCode string `json:"coupon_code" validate:"required,len=12"` // 12-значный код купона
}

// EditImageRequest - запрос редактирования изображения
type EditImageRequest struct {
	CropX      int     `json:"crop_x" validate:"min=0"`                // X координата начала кадрирования
	CropY      int     `json:"crop_y" validate:"min=0"`                // Y координата начала кадрирования
	CropWidth  int     `json:"crop_width" validate:"min=1"`            // Ширина области кадрирования
	CropHeight int     `json:"crop_height" validate:"min=1"`           // Высота области кадрирования
	Rotation   int     `json:"rotation" validate:"oneof=0 90 180 270"` // Поворот в градусах
	Scale      float64 `json:"scale" validate:"min=0.1,max=5.0"`       // Масштаб
}

// ProcessImageRequest - запрос обработки изображения (выбор стилей)
type ProcessImageRequest struct {
	Style      string  `json:"style" validate:"required,oneof=grayscale skin_tones pop_art max_colors"` // Стиль обработки
	UseAI      bool    `json:"use_ai"`                                                                  // Использовать AI обработку через Stable Diffusion
	Lighting   string  `json:"lighting,omitempty" validate:"omitempty,oneof=sun moon venus"`            // Освещение (солнце, луна, венера)
	Contrast   string  `json:"contrast,omitempty" validate:"omitempty,oneof=low high"`                  // Контрастность (2 варианта)
	Brightness float64 `json:"brightness,omitempty" validate:"omitempty,min=-100,max=100"`              // Яркость (-100 до 100)
	Saturation float64 `json:"saturation,omitempty" validate:"omitempty,min=-100,max=100"`              // Насыщенность (-100 до 100)
}

// GenerateSchemaRequest - запрос создания схемы
type GenerateSchemaRequest struct {
	Confirmed bool `json:"confirmed" validate:"required"` // Подтверждение создания схемы
}

// ImageUploadResponse - ответ при загрузке изображения
type ImageUploadResponse struct {
	Message     string    `json:"message"`
	ImageID     uuid.UUID `json:"image_id"`
	NextStep    string    `json:"next_step"`
	CouponSize  string    `json:"coupon_size"`  // Размер купона (определяется автоматически)
	CouponStyle string    `json:"coupon_style"` // Стиль купона (определяется автоматически)
}

// ImageEditResponse - ответ при редактировании изображения
type ImageEditResponse struct {
	Message    string    `json:"message"`
	ImageID    uuid.UUID `json:"image_id"`
	NextStep   string    `json:"next_step"`
	PreviewURL string    `json:"preview_url"` // URL для просмотра отредактированного изображения
}

// ProcessImageResponse - ответ при обработке изображения
type ProcessImageResponse struct {
	Message     string    `json:"message"`
	ImageID     uuid.UUID `json:"image_id"`
	NextStep    string    `json:"next_step"`
	PreviewURL  string    `json:"preview_url"`  // URL превью обработанного изображения
	OriginalURL string    `json:"original_url"` // URL оригинального изображения для сравнения
}

// GenerateSchemaResponse - ответ при создании схемы
type GenerateSchemaResponse struct {
	Message    string    `json:"message"`
	ImageID    uuid.UUID `json:"image_id"`
	SchemaURL  string    `json:"schema_url"`  // URL для скачивания схемы
	PreviewURL string    `json:"preview_url"` // URL финального превью
	EmailSent  bool      `json:"email_sent"`  // Отправлена ли схема на email
}

// ImageStatusResponse - ответ со статусом обработки изображения
type ImageStatusResponse struct {
	ImageID       uuid.UUID `json:"image_id"`
	Status        string    `json:"status"` // queued, processing, completed, failed
	Message       string    `json:"message"`
	Progress      int       `json:"progress"`       // Прогресс в процентах (0-100)
	EstimatedTime *int      `json:"estimated_time"` // Оценочное время до завершения в секундах
	ErrorMessage  *string   `json:"error_message"`  // Сообщение об ошибке (если есть)
	OriginalURL   *string   `json:"original_url"`   // URL оригинального изображения
	EditedURL     *string   `json:"edited_url"`     // URL отредактированного изображения
	ProcessedURL  *string   `json:"processed_url"`  // URL обработанного изображения
	PreviewURL    *string   `json:"preview_url"`    // URL превью
	SchemaURL     *string   `json:"schema_url"`     // URL схемы (если готова)
}
