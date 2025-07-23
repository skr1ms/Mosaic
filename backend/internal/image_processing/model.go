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
	Style    string                 `json:"style"`
	Settings map[string]interface{} `json:"settings"`
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
	CouponID          uuid.UUID        `gorm:"type:uuid;not null" json:"coupon_id"`
	OriginalImagePath string           `gorm:"not null;size:255" json:"original_image_path"`
	ProcessingParams  ProcessingParams `gorm:"type:json;not null" json:"processing_params"`
	UserEmail         string           `gorm:"not null;size:255" json:"user_email"`
	Status            string           `gorm:"type:enum('queued','processing','completed','failed');default:'queued'" json:"status"`
	Priority          int              `gorm:"default:0" json:"priority"`
	StartedAt         *time.Time       `json:"started_at"`
	CompletedAt       *time.Time       `json:"completed_at"`
	ErrorMessage      *string          `gorm:"type:text" json:"error_message"`
	RetryCount        int              `gorm:"default:0" json:"retry_count"`
	MaxRetries        int              `gorm:"default:3" json:"max_retries"`
	CreatedAt         time.Time        `json:"created_at"`
	UpdatedAt         time.Time        `json:"updated_at"`
}
