package queue

import (
	"context"

	"github.com/google/uuid"
	"github.com/skr1ms/mosaic/internal/image"
	"github.com/skr1ms/mosaic/pkg/email"
)

// ImageServiceAdapter адаптер для совместимости с queue.ImageService
type ImageServiceAdapter struct {
	imageService *image.ImageService
}

// NewImageServiceAdapter создает новый адаптер
func NewImageServiceAdapter(imageService *image.ImageService) *ImageServiceAdapter {
	return &ImageServiceAdapter{
		imageService: imageService,
	}
}

// ProcessImageWithStyle обрабатывает изображение с заданным стилем
func (a *ImageServiceAdapter) ProcessImageWithStyle(ctx context.Context, imageID uuid.UUID, style string, parameters map[string]interface{}) error {
	// Конвертируем parameters в ProcessingParams
	processParams := image.ProcessingParams{
		Style: style,
	}

	// Извлекаем параметры из map
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
	if settings, ok := parameters["settings"].(map[string]interface{}); ok {
		processParams.Settings = settings
	}

	return a.imageService.ProcessImage(ctx, imageID, processParams)
}

// GenerateSchema генерирует схему мозаики
func (a *ImageServiceAdapter) GenerateSchema(ctx context.Context, imageID uuid.UUID, confirmed bool) error {
	return a.imageService.GenerateSchema(ctx, imageID, confirmed)
}

// OptimizeImage оптимизирует изображение (метод-заглушка, можно реализовать позже)
func (a *ImageServiceAdapter) OptimizeImage(ctx context.Context, imageID uuid.UUID, quality int) error {
	// Пока что возвращаем nil, можно добавить реализацию позже
	return nil
}

// GenerateThumbnails генерирует превью изображений (метод-заглушка, можно реализовать позже)
func (a *ImageServiceAdapter) GenerateThumbnails(ctx context.Context, imageID uuid.UUID, sizes []string) error {
	// Пока что возвращаем nil, можно добавить реализацию позже
	return nil
}

// EmailServiceAdapter адаптер для совместимости с queue.EmailService
type EmailServiceAdapter struct {
	mailer *email.Mailer
}

// NewEmailServiceAdapter создает новый адаптер
func NewEmailServiceAdapter(mailer *email.Mailer) *EmailServiceAdapter {
	return &EmailServiceAdapter{
		mailer: mailer,
	}
}

// SendSchema отправляет схему по email
func (a *EmailServiceAdapter) SendSchema(ctx context.Context, email, schemaURL, couponCode string) error {
	return a.mailer.SendSchemaEmail(email, schemaURL, couponCode)
}
