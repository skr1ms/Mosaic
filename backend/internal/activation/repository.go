package activation

import (
	"gorm.io/gorm"
)

func CreateActivation(db *gorm.DB, activation *Activation) error {
	return db.Create(activation).Error
}
