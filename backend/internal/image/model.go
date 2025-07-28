package image

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
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

type Image struct {
	bun.BaseModel `bun:"table:images,alias:i"`

	ID                uuid.UUID        `bun:"id,pk,type:uuid,default:gen_random_uuid()" json:"id"`
	CouponID          uuid.UUID        `bun:"coupon_id,type:uuid,notnull" json:"coupon_id"`
	OriginalImagePath string           `bun:"original_image_path,notnull" json:"original_image_path"`
	EditedImagePath   *string          `bun:"edited_image_path" json:"edited_image_path"`           // Путь к отредактированному изображению
	PreviewPath       *string          `bun:"preview_path" json:"preview_path"`                     // Путь к превью изображения
	ResultPath        *string          `bun:"result_path" json:"result_path"`                       // Путь к готовой схеме
	ProcessingParams  ProcessingParams `bun:"processing_params,type:json" json:"processing_params"` // Параметры обработки
	UserEmail         string           `bun:"user_email,notnull" json:"user_email"`
	Status            string           `bun:"status,type:processing_status,default:'queued'" json:"status"`
	Priority          int              `bun:"priority,default:0" json:"priority"`
	StartedAt         *time.Time       `bun:"started_at" json:"started_at"`
	CompletedAt       *time.Time       `bun:"completed_at" json:"completed_at"`
	ErrorMessage      *string          `bun:"error_message,type:text" json:"error_message"`
	RetryCount        int              `bun:"retry_count,default:0" json:"retry_count"`
	MaxRetries        int              `bun:"max_retries,default:3" json:"max_retries"`
	CreatedAt         time.Time        `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt         time.Time        `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`
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

func (i *Image) CreateIndex() string {
	return `
	CREATE INDEX IF NOT EXISTS idx_images_coupon_id ON images(coupon_id);
	CREATE INDEX IF NOT EXISTS idx_images_status ON images(status);
	CREATE INDEX IF NOT EXISTS idx_images_queue_order ON images(status, priority DESC, created_at ASC);
	CREATE INDEX IF NOT EXISTS idx_images_retry ON images(status, retry_count);
	CREATE INDEX IF NOT EXISTS idx_images_user_email ON images(user_email);
	CREATE INDEX IF NOT EXISTS idx_images_started_at ON images(started_at);
	CREATE INDEX IF NOT EXISTS idx_images_completed_at ON images(completed_at);
	CREATE INDEX IF NOT EXISTS idx_images_created_at ON images(created_at);
	CREATE INDEX IF NOT EXISTS idx_images_updated_at ON images(updated_at);
	`
}
