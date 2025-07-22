package partner

import (
	"gorm.io/gorm"
)

func GetAllPartners(db *gorm.DB) ([]Partner, error) {
	var partners []Partner
	return partners, db.Find(&partners).Error
}

func CreatePartner(db *gorm.DB, partner *Partner) error {
	return db.Create(partner).Error
}
