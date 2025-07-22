package coupon

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Coupon struct {
	gorm.Model
	Code         string `gorm:"unique"`
	PartnerID    *uuid.UUID
	Size         string
	Style        string
	Status       string
	UserImageURL *string
	PreviewURL   *string
	SchemaURL    *string
}
