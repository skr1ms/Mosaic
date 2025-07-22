package user

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Email        string    `gorm:"unique"`
	PasswordHash string
	Role         string
	PartnerID    *uuid.UUID
}
