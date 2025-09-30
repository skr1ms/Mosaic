package types

import "github.com/google/uuid"

// Common structures for all modules to avoid circular imports

// UploadImageRequest - image upload request
type UploadImageRequest struct {
	CouponCode string `json:"coupon_code" validate:"required,len=12"` // 12-digit coupon code
}

// EditImageRequest - image editing request
type EditImageRequest struct {
	CropX      int     `json:"crop_x" validate:"min=0"`                // X coordinate of crop start
	CropY      int     `json:"crop_y" validate:"min=0"`                // Y coordinate of crop start
	CropWidth  int     `json:"crop_width" validate:"min=1"`            // Width of crop area
	CropHeight int     `json:"crop_height" validate:"min=1"`           // Height of crop area
	Rotation   int     `json:"rotation" validate:"oneof=0 90 180 270"` // Rotation in degrees
	Scale      float64 `json:"scale" validate:"min=0.1,max=5.0"`       // Scale
}

// ProcessImageRequest - image processing request (style selection)
type ProcessImageRequest struct {
	Style      string  `json:"style" validate:"required,oneof=grayscale skin_tones pop_art max_colors"` // Processing style
	UseAI      bool    `json:"use_ai"`                                                                  // Use AI processing through Stable Diffusion
	Lighting   string  `json:"lighting,omitempty" validate:"omitempty,oneof=sun moon venus"`            // Lighting (sun, moon, venus)
	Contrast   string  `json:"contrast,omitempty" validate:"omitempty,oneof=low high"`                  // Contrast (2 options)
	Brightness float64 `json:"brightness,omitempty" validate:"omitempty,min=-100,max=100"`              // Brightness (-100 to 100)
	Saturation float64 `json:"saturation,omitempty" validate:"omitempty,min=-100,max=100"`              // Saturation (-100 to 100)
}

// GenerateSchemaRequest - schema generation request
type GenerateSchemaRequest struct {
	Confirmed bool `json:"confirmed" validate:"required"`
}

// ImageUploadResponse - response when uploading image
type ImageUploadResponse struct {
	Message     string    `json:"message"`
	ImageID     uuid.UUID `json:"image_id"`
	NextStep    string    `json:"next_step"`
	CouponSize  string    `json:"coupon_size"`
	CouponStyle string    `json:"coupon_style"`
}

// ImageEditResponse - response when editing image
type ImageEditResponse struct {
	Message    string    `json:"message"`
	ImageID    uuid.UUID `json:"image_id"`
	NextStep   string    `json:"next_step"`
	PreviewURL string    `json:"preview_url"`
}

// ProcessImageResponse - response when processing image
type ProcessImageResponse struct {
	Message     string    `json:"message"`
	ImageID     uuid.UUID `json:"image_id"`
	NextStep    string    `json:"next_step"`
	PreviewURL  string    `json:"preview_url"`
	OriginalURL string    `json:"original_url"`
}

// GenerateSchemaResponse - response when creating schema
type GenerateSchemaResponse struct {
	Message    string    `json:"message"`
	ImageID    uuid.UUID `json:"image_id"`
	ZipURL     string    `json:"zip_url"`
	PreviewURL string    `json:"preview_url"`
	EmailSent  bool      `json:"email_sent"`
}

// ImageStatusResponse - response with image processing status
type ImageStatusResponse struct {
	ImageID       uuid.UUID `json:"image_id"`
	Status        string    `json:"status"` // queued, processing, completed, failed
	Message       string    `json:"message"`
	Progress      int       `json:"progress"`
	EstimatedTime *int      `json:"estimated_time"`
	ErrorMessage  *string   `json:"error_message"`
	OriginalURL   *string   `json:"original_url"`
	EditedURL     *string   `json:"edited_url"`
	ProcessedURL  *string   `json:"processed_url"`
	PreviewURL    *string   `json:"preview_url"`
	ZipURL        *string   `json:"zip_url"`
}
