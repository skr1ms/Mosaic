package utils

import "github.com/skr1ms/mosaic/internal/coupon"

func UpdateCouponData(coupon *coupon.Coupon, req *coupon.UpdateCouponRequest) {
	if req.Status != nil {
		coupon.Status = string(*req.Status)
	}
}