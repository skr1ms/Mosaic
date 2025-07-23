package image_processing

import "github.com/google/uuid"

type AddToQueueRequest struct {
	CouponID          uuid.UUID        `json:"coupon_id"`
	OriginalImagePath string           `json:"original_image_path"`
	ProcessingParams  ProcessingParams `json:"processing_params"`
	UserEmail         string           `json:"user_email"`
	Priority          int              `json:"priority"`
}

type FailProcessingRequest struct {
	ErrorMessage string `json:"error_message"`
}