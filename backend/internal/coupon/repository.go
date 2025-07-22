package coupon

import (
	"gorm.io/gorm"
)

func GetAllCoupons(db *gorm.DB) ([]Coupon, error) {
	var coupons []Coupon
	return coupons, db.Find(&coupons).Error
}

func CreateCoupon(db *gorm.DB, coupon *Coupon) error {
	return db.Create(coupon).Error
}
