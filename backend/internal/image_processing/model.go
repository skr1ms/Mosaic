package image_processing

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

// ProcessingParams содержит параметры обработки изображения
type ProcessingParams struct {
	Style    string                 `json:"style" validate:"required,oneof=grayscale skin_tones pop_art max_colors"`
	Settings map[string]interface{} `json:"settings" validate:"required"`
}

// Value реализует интерфейс driver.Valuer для хранения JSON в БД
func (p ProcessingParams) Value() (driver.Value, error) {
	return json.Marshal(p)
}

// Scan реализует интерфейс sql.Scanner для чтения JSON из БД
func (p *ProcessingParams) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("cannot scan non-bytes into ProcessingParams")
	}

	return json.Unmarshal(bytes, p)
}

type ImageProcessingQueue struct {
	ID                uuid.UUID        `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CouponID          uuid.UUID        `gorm:"type:uuid;not null;index:idx_image_coupon_id" json:"coupon_id"`
	OriginalImagePath string           `gorm:"not null;size:255" json:"original_image_path"`
	ProcessingParams  ProcessingParams `gorm:"type:json;not null" json:"processing_params"`
	UserEmail         string           `gorm:"not null;size:255;index:idx_image_user_email" json:"user_email"`
	Status            string           `gorm:"type:processing_status;default:'queued';index:idx_image_status;index:idx_image_queue_order,priority:1;index:idx_image_retry,priority:1" json:"status"`
	Priority          int              `gorm:"default:0;index:idx_image_queue_order,priority:2" json:"priority"`
	StartedAt         *time.Time       `gorm:"index:idx_image_started_at" json:"started_at"`
	CompletedAt       *time.Time       `gorm:"index:idx_image_completed_at" json:"completed_at"`
	ErrorMessage      *string          `gorm:"type:text" json:"error_message"`
	RetryCount        int              `gorm:"default:0;index:idx_image_retry,priority:2" json:"retry_count"`
	MaxRetries        int              `gorm:"default:3" json:"max_retries"`
	CreatedAt         time.Time        `gorm:"index:idx_image_created_at;index:idx_image_queue_order,priority:3" json:"created_at"`
	UpdatedAt         time.Time        `gorm:"index:idx_image_updated_at" json:"updated_at"`
}

// - idx_image_coupon_id: быстрый поиск задач по купону
// - idx_image_status: фильтрация по статусу обработки
// - idx_image_queue_order: составной индекс для очереди (status, priority DESC, created_at ASC)
// - idx_image_retry: составной индекс для поиска задач для повтора (status, retry_count)
// - idx_image_user_email: поиск задач по email пользователя
// - idx_image_started_at: аналитика времени начала обработки
// - idx_image_completed_at: аналитика времени завершения
// - idx_image_created_at: сортировка по дате создания
// - idx_image_updated_at: сортировка по дате обновления
