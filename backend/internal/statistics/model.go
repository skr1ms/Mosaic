package statistics

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Statistics struct {
	gorm.Model
	PartnerID        *uuid.UUID
	TotalCoupons     int
	RedeemedCoupons  int
	PurchasedCoupons int
}
