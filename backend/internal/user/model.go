package user

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User представляет служебные аккаунты системы (админы и партнеры).
// Обычные пользователи НЕ создают записи в этой таблице и работают анонимно через купоны.
type User struct {
	gorm.Model
	Email        string `gorm:"unique"`
	PasswordHash string
	Role         string     // "admin" или "partner"
	PartnerID    *uuid.UUID // Заполняется только для партнеров
}
