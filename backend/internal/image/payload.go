package image

import "github.com/google/uuid"

type AddToQueueRequest struct {
	CouponID          uuid.UUID         `json:"coupon_id" validate:"required"`
	OriginalImagePath string            `json:"original_image_path" validate:"required"`
	ProcessingParams  *ProcessingParams `json:"processing_params" validate:"required"`
	UserEmail         string            `json:"user_email" validate:"required,email"`
	Priority          int               `json:"priority" validate:"min=1,max=10"`
}

type FailProcessingRequest struct {
	ErrorMessage string `json:"error_message" validate:"required,min=1,max=500"`
}
