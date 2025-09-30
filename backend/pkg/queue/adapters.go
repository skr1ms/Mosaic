package queue

import (
	"context"

	"github.com/google/uuid"
	"github.com/skr1ms/mosaic/internal/image"
	"github.com/skr1ms/mosaic/pkg/email"
)

// ImageServiceAdapter adapter for compatibility with queue.ImageService
type ImageServiceAdapter struct {
	imageService *image.ImageService
}

// NewImageServiceAdapter creates new adapter
func NewImageServiceAdapter(imageService *image.ImageService) *ImageServiceAdapter {
	return &ImageServiceAdapter{
		imageService: imageService,
	}
}

// ProcessImageWithStyle processes image with given style
func (a *ImageServiceAdapter) ProcessImageWithStyle(ctx context.Context, imageID uuid.UUID, style string, parameters map[string]any) error {
	processParams := image.ProcessingParams{
		Style: style,
	}

	if contrast, ok := parameters["contrast"].(string); ok {
		processParams.Contrast = contrast
	}
	if brightness, ok := parameters["brightness"].(float64); ok {
		processParams.Brightness = brightness
	}
	if saturation, ok := parameters["saturation"].(float64); ok {
		processParams.Saturation = saturation
	}
	if useAI, ok := parameters["use_ai"].(bool); ok {
		processParams.UseAI = useAI
	}
	if lighting, ok := parameters["lighting"].(string); ok {
		processParams.Lighting = lighting
	}
	if settings, ok := parameters["settings"].(map[string]any); ok {
		processParams.Settings = settings
	}

	return a.imageService.ProcessImage(ctx, imageID, &processParams)
}

// ProcessImageWithAI processes image using AI (Stable Diffusion)
func (a *ImageServiceAdapter) ProcessImageWithAI(ctx context.Context, imageID uuid.UUID, style string, useAI bool, parameters map[string]any) error {
	processParams := image.ProcessingParams{
		Style: style,
		UseAI: useAI,
	}

	if contrast, ok := parameters["contrast"].(string); ok {
		processParams.Contrast = contrast
	}
	if brightness, ok := parameters["brightness"].(float64); ok {
		processParams.Brightness = brightness
	}
	if saturation, ok := parameters["saturation"].(float64); ok {
		processParams.Saturation = saturation
	}
	if lighting, ok := parameters["lighting"].(string); ok {
		processParams.Lighting = lighting
	}
	if settings, ok := parameters["settings"].(map[string]any); ok {
		processParams.Settings = settings
	}

	// Log processing image with AI

	return a.imageService.ProcessImage(ctx, imageID, &processParams)
}

// GenerateSchema generates mosaic schema
func (a *ImageServiceAdapter) GenerateSchema(ctx context.Context, imageID uuid.UUID, confirmed bool) error {
	return a.imageService.GenerateSchema(ctx, imageID, confirmed)
}

// OptimizeImage optimizes image (stub method, can be implemented later)
func (a *ImageServiceAdapter) OptimizeImage(ctx context.Context, imageID uuid.UUID, quality int) error {
	return nil
}

// GenerateThumbnails generates image thumbnails (stub method, can be implemented later)
func (a *ImageServiceAdapter) GenerateThumbnails(ctx context.Context, imageID uuid.UUID, sizes []string) error {
	// For now return nil, can add implementation later
	return nil
}

// EmailServiceAdapter adapter for compatibility with queue.EmailService
type EmailServiceAdapter struct {
	mailer *email.Mailer
}

// NewEmailServiceAdapter creates new adapter
func NewEmailServiceAdapter(mailer *email.Mailer) *EmailServiceAdapter {
	return &EmailServiceAdapter{
		mailer: mailer,
	}
}

// SendSchema sends schema via email
func (a *EmailServiceAdapter) SendSchema(ctx context.Context, email, schemaURL, couponCode string) error {
	return a.mailer.SendSchemaEmail(email, schemaURL, couponCode)
}
