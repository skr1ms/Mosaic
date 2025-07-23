package admin

import (
	"time"

	"github.com/google/uuid"
)

type Admin struct {
	ID        uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Login     string     `gorm:"unique;not null;size:255" json:"login"`
	Password  string     `gorm:"not null;size:255" json:"password"`
	LastLogin *time.Time `json:"last_login"`
	CreatedAt time.Time  `json:"created_at"`
}
