package partner

import (
	"gorm.io/gorm"
)

type Partner struct {
	gorm.Model
	Domain          string `gorm:"unique"`
	BrandName       string
	LogoURL         string
	OzonLink        string
	WildberriesLink string
	Email           string
	Address         string
	Phone           string
	Telegram        string
	Whatsapp        string
	AllowSales      bool
	Status          string
}
