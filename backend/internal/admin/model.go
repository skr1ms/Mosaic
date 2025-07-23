package admin

import (
	"time"

	"github.com/google/uuid"
)

type Admin struct {
	ID        uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Login     string     `gorm:"unique;not null;size:255;index" json:"login"`
	Password  string     `gorm:"not null;size:255" json:"password"`
	LastLogin *time.Time `gorm:"index:idx_admins_last_login" json:"last_login"`
	CreatedAt time.Time  `gorm:"index:idx_admins_created_at" json:"created_at"`
}

// - login уже имеет unique индекс, но добавляем обычный для быстрого поиска
// - idx_admins_last_login: сортировка по последнему входу
// - idx_admins_created_at: сортировка по дате создания
