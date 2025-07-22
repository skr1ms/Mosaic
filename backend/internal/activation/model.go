package activation

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Activation представляет факт активации купона.
// UserID может быть null для обычных пользователей (анонимная активация).
// UserID заполняется только если купон активировал админ или партнер.
type Activation struct {
	gorm.Model
	CouponID  uuid.UUID
	UserID    *uuid.UUID // null для обычных пользователей, заполняется для админов/партнеров
	IPAddress string
}
