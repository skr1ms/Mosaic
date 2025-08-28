package image

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type ProcessingParams struct {
	Style      string         `json:"style" validate:"required,oneof=grayscale skin_tones pop_art max_colors"`
	UseAI      bool           `json:"use_ai"`
	Lighting   string         `json:"lighting,omitempty" validate:"omitempty,oneof=sun moon venus"`
	Contrast   string         `json:"contrast,omitempty" validate:"omitempty,oneof=low high"`
	Brightness float64        `json:"brightness,omitempty" validate:"omitempty,min=-100,max=100"`
	Saturation float64        `json:"saturation,omitempty" validate:"omitempty,min=-100,max=100"`
	Settings   map[string]any `json:"settings,omitempty"`
}

// Value implements driver.Valuer interface to convert ProcessingParams to database value
func (p *ProcessingParams) Value() (driver.Value, error) {
	if p == nil {
		return nil, nil
	}
	b, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

// Scan implements sql.Scanner interface to convert database value to ProcessingParams
func (p *ProcessingParams) Scan(value any) error {
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

	ID                  uuid.UUID         `bun:"id,pk,type:uuid,default:gen_random_uuid()" json:"id"`
	CouponID            uuid.UUID         `bun:"coupon_id,type:uuid,notnull" json:"coupon_id"`
	OriginalImageS3Key  string            `bun:"original_image_s3_key,notnull" json:"original_image_s3_key"`
	EditedImageS3Key    *string           `bun:"edited_image_s3_key" json:"edited_image_s3_key"`
	ProcessedImageS3Key *string           `bun:"processed_image_s3_key" json:"processed_image_s3_key"`
	PreviewS3Key        *string           `bun:"preview_s3_key" json:"preview_s3_key"`
	SchemaS3Key         *string           `bun:"schema_s3_key" json:"schema_s3_key"`
	ProcessingParams    *ProcessingParams `bun:"processing_params,type:json" json:"processing_params"`
	UserEmail           string            `bun:"user_email,notnull" json:"user_email"`
	Status              string            `bun:"status,type:processing_status,default:'queued'" json:"status"`
	Priority            int               `bun:"priority,default:0" json:"priority"`
	StartedAt           *time.Time        `bun:"started_at" json:"started_at"`
	CompletedAt         *time.Time        `bun:"completed_at" json:"completed_at"`
	ErrorMessage        *string           `bun:"error_message,type:text" json:"error_message"`
	RetryCount          int               `bun:"retry_count,default:0" json:"retry_count"`
	MaxRetries          int               `bun:"max_retries,default:3" json:"max_retries"`
	CreatedAt           time.Time         `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt           time.Time         `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`
}

type ImageWithPartner struct {
	*Image
	PartnerID   uuid.UUID `bun:"partner_id" json:"partner_id"`
	PartnerCode string    `bun:"partner_code" json:"partner_code"`
}

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

type ImageEditParams struct {
	CropX      int     `json:"crop_x" validate:"min=0"`
	CropY      int     `json:"crop_y" validate:"min=0"`
	CropWidth  int     `json:"crop_width" validate:"min=1"`
	CropHeight int     `json:"crop_height" validate:"min=1"`
	Rotation   int     `json:"rotation" validate:"oneof=0 90 180 270"`
	Scale      float64 `json:"scale" validate:"min=0.1,max=5.0"`
}
