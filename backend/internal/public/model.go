package public

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type PreviewData struct {
	bun.BaseModel `bun:"table:previews,alias:p"`

	ID        uuid.UUID  `bun:"id,pk,type:uuid,default:gen_random_uuid()" json:"id"`
	URL       string     `bun:"url,notnull" json:"url"`
	Style     string     `bun:"style,notnull" json:"style"`
	Contrast  string     `bun:"contrast,notnull" json:"contrast"`
	Size      string     `bun:"size,notnull" json:"size"`
	ImageID   *uuid.UUID `bun:"image_id,type:uuid" json:"image_id,omitempty"`
	S3Key     string     `bun:"s3_key,notnull" json:"s3_key"`
	CreatedAt time.Time  `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt time.Time  `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`
	ExpiresAt *time.Time `bun:"expires_at" json:"expires_at,omitempty"` // TTL for cleanup
}
