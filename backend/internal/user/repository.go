package user

import (
	"gorm.io/gorm"
)

func GetUserByEmail(db *gorm.DB, email string) (*User, error) {
	var user User
	return &user, db.Where("email = ?", email).First(&user).Error
}
func CreateUser(db *gorm.DB, user *User) error {
	return db.Create(user).Error
}
