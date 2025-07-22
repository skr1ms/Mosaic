package activation

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Activation struct {
	gorm.Model
	CouponID  uuid.UUID
	UserID    *uuid.UUID
	IPAddress string
}
