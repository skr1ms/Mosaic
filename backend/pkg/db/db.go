package db

import (
	"log"

	"github.com/skr1ms/mosaic/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Db struct {
	*gorm.DB
}

func NewDB(cfg *config.Config) *Db {
	db, err := gorm.Open(postgres.Open(cfg.DatabaseConfig.URL), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	return &Db{db}
}
