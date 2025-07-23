package coupon

import "github.com/google/uuid"

type CreateCouponsRequest struct {
	Count     int       `json:"count"`
	PartnerID uuid.UUID `json:"partner_id"`
	Size      string    `json:"size"`
	Style     string    `json:"style"`
}
