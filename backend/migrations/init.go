package migrations

import (
	"github.com/skr1ms/mosaic/config"
	"github.com/skr1ms/mosaic/internal/activation"
	"github.com/skr1ms/mosaic/internal/coupon"
	"github.com/skr1ms/mosaic/internal/partner"
	"github.com/skr1ms/mosaic/pkg/db"
)

func Init(cfg *config.Config) {
	database := db.NewDB(cfg)
	err := database.AutoMigrate(&partner.Partner{}, &coupon.Coupon{}, &activation.Activation{})
	if err != nil {
		panic("Failed to migrate database: " + err.Error())
	}
}
