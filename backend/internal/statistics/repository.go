package statistics

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func GetStatistics(db *gorm.DB, partnerID *uuid.UUID) (*Statistics, error) {
	var stats Statistics
	return &stats, db.Where("partner_id = ?", partnerID).First(&stats).Error
}
