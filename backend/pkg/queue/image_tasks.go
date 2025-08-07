package queue

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// ImageTaskQueue специализированная очередь для обработки изображений
type ImageTaskQueue struct {
	*TaskQueue
}

// NewImageTaskQueue создает очередь для задач обработки изображений
func NewImageTaskQueue(redis *redis.Client) *ImageTaskQueue {
	return &ImageTaskQueue{
		TaskQueue: NewTaskQueue(redis, "images"),
	}
}

// Типы задач для обработки изображений
const (
	TaskTypeImageProcessing     = "image_processing"
	TaskTypeSchemaGeneration    = "schema_generation"
	TaskTypeEmailSending        = "email_sending"
	TaskTypeImageOptimization   = "image_optimization"
	TaskTypeThumbnailGeneration = "thumbnail_generation"
)

// EnqueueImageProcessing добавляет задачу обработки изображения
func (q *ImageTaskQueue) EnqueueImageProcessing(imageID uuid.UUID, style string, parameters map[string]interface{}) error {
	payload := map[string]interface{}{
		"image_id":   imageID.String(),
		"style":      style,
		"parameters": parameters,
	}

	return q.Enqueue(TaskTypeImageProcessing, payload, WithPriority(5), WithMaxRetries(3))
}

// EnqueueSchemaGeneration добавляет задачу генерации схемы
func (q *ImageTaskQueue) EnqueueSchemaGeneration(imageID uuid.UUID, couponID uuid.UUID, confirmed bool) error {
	payload := map[string]interface{}{
		"image_id":  imageID.String(),
		"coupon_id": couponID.String(),
		"confirmed": confirmed,
	}

	return q.Enqueue(TaskTypeSchemaGeneration, payload, WithPriority(8), WithMaxRetries(2))
}

// EnqueueEmailSending добавляет задачу отправки email
func (q *ImageTaskQueue) EnqueueEmailSending(email string, schemaURL string, couponCode string) error {
	payload := map[string]interface{}{
		"email":       email,
		"schema_url":  schemaURL,
		"coupon_code": couponCode,
	}

	return q.Enqueue(TaskTypeEmailSending, payload, WithPriority(3), WithMaxRetries(5))
}

// EnqueueImageOptimization добавляет задачу оптимизации изображения
func (q *ImageTaskQueue) EnqueueImageOptimization(imageID uuid.UUID, quality int) error {
	payload := map[string]interface{}{
		"image_id": imageID.String(),
		"quality":  quality,
	}

	return q.Enqueue(TaskTypeImageOptimization, payload, WithPriority(2))
}

// EnqueueThumbnailGeneration добавляет задачу создания превью
func (q *ImageTaskQueue) EnqueueThumbnailGeneration(imageID uuid.UUID, sizes []string) error {
	payload := map[string]interface{}{
		"image_id": imageID.String(),
		"sizes":    sizes,
	}

	return q.Enqueue(TaskTypeThumbnailGeneration, payload, WithPriority(1))
}

// GetImageTaskHandlers возвращает карту обработчиков для задач обработки изображений
func GetImageTaskHandlers(imageService *ImageServiceAdapter, emailService *EmailServiceAdapter) map[string]TaskHandler {
	return map[string]TaskHandler{
		"process_image":       func(ctx context.Context, task *Task) error { return handleProcessImage(ctx, task, imageService) },
		"generate_schema":     func(ctx context.Context, task *Task) error { return handleGenerateSchema(ctx, task, imageService) },
		"send_schema":         func(ctx context.Context, task *Task) error { return handleSendSchema(ctx, task, emailService) },
		"optimize_image":      func(ctx context.Context, task *Task) error { return handleOptimizeImage(ctx, task, imageService) },
		"generate_thumbnails": func(ctx context.Context, task *Task) error { return handleGenerateThumbnails(ctx, task, imageService) },
	}
}

// Интерфейсы для сервисов

type ImageService interface {
	ProcessImageWithStyle(ctx context.Context, imageID uuid.UUID, style string, parameters map[string]interface{}) error
	GenerateSchema(ctx context.Context, imageID uuid.UUID, confirmed bool) error
	OptimizeImage(ctx context.Context, imageID uuid.UUID, quality int) error
	GenerateThumbnails(ctx context.Context, imageID uuid.UUID, sizes []string) error
}

type EmailService interface {
	SendSchema(ctx context.Context, email, schemaURL, couponCode string) error
}

// Конкретные обработчики задач

// handleProcessImage обрабатывает задачу обработки изображения
func handleProcessImage(ctx context.Context, task *Task, imageService *ImageServiceAdapter) error {
	payload := task.Payload

	imageIDStr, ok := payload["image_id"].(string)
	if !ok {
		return fmt.Errorf("invalid image_id")
	}

	imageID, err := uuid.Parse(imageIDStr)
	if err != nil {
		return err
	}

	style, _ := payload["style"].(string)
	parameters, _ := payload["parameters"].(map[string]interface{})

	return imageService.ProcessImageWithStyle(ctx, imageID, style, parameters)
}

// handleGenerateSchema обрабатывает задачу генерации схемы
func handleGenerateSchema(ctx context.Context, task *Task, imageService *ImageServiceAdapter) error {
	payload := task.Payload

	imageIDStr, ok := payload["image_id"].(string)
	if !ok {
		return fmt.Errorf("invalid image_id")
	}

	imageID, err := uuid.Parse(imageIDStr)
	if err != nil {
		return err
	}

	confirmed, _ := payload["confirmed"].(bool)

	return imageService.GenerateSchema(ctx, imageID, confirmed)
}

// handleSendSchema обрабатывает задачу отправки схемы
func handleSendSchema(ctx context.Context, task *Task, emailService *EmailServiceAdapter) error {
	payload := task.Payload

	email, _ := payload["email"].(string)
	schemaURL, _ := payload["schema_url"].(string)
	couponCode, _ := payload["coupon_code"].(string)

	return emailService.SendSchema(ctx, email, schemaURL, couponCode)
}

// handleOptimizeImage обрабатывает задачу оптимизации изображения
func handleOptimizeImage(ctx context.Context, task *Task, imageService *ImageServiceAdapter) error {
	payload := task.Payload

	imageIDStr, ok := payload["image_id"].(string)
	if !ok {
		return fmt.Errorf("invalid image_id")
	}

	imageID, err := uuid.Parse(imageIDStr)
	if err != nil {
		return err
	}

	quality, _ := payload["quality"].(int)

	return imageService.OptimizeImage(ctx, imageID, quality)
}

// handleGenerateThumbnails обрабатывает задачу генерации превью
func handleGenerateThumbnails(ctx context.Context, task *Task, imageService *ImageServiceAdapter) error {
	payload := task.Payload

	imageIDStr, ok := payload["image_id"].(string)
	if !ok {
		return fmt.Errorf("invalid image_id")
	}

	imageID, err := uuid.Parse(imageIDStr)
	if err != nil {
		return err
	}

	sizes, _ := payload["sizes"].([]string)

	return imageService.GenerateThumbnails(ctx, imageID, sizes)
}
