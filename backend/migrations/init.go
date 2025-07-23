package migrations

import (
	"github.com/skr1ms/mosaic/config"
	"github.com/skr1ms/mosaic/internal/admin"
	"github.com/skr1ms/mosaic/internal/coupon"
	"github.com/skr1ms/mosaic/internal/image_processing"
	"github.com/skr1ms/mosaic/internal/partner"
	"github.com/skr1ms/mosaic/pkg/db"
)

func Init(cfg *config.Config) {
	database := db.NewDB(cfg)
	err := database.AutoMigrate(
		&partner.Partner{},
		&admin.Admin{},
		&coupon.Coupon{},
		&image_processing.ImageProcessingQueue{},
	)
	if err != nil {
		panic("Failed to migrate database: " + err.Error())
	}
}
