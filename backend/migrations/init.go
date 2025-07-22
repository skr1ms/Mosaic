package migrations

import (
	"github.com/skr1ms/mosaic/config"
	"github.com/skr1ms/mosaic/internal/user"
	"github.com/skr1ms/mosaic/pkg/db"
)

func Init(cfg *config.Config) {
	database := db.NewDB(cfg)
	err := database.AutoMigrate(&user.User{})
	if err != nil {
		panic("Failed to migrate database: " + err.Error())
	}
}
