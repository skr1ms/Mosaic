package queue

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/skr1ms/mosaic/pkg/middleware"
)

// ImageTaskQueue specialized queue for image processing
type ImageTaskQueue struct {
	*TaskQueue
}

// NewImageTaskQueue creates queue for image processing tasks
func NewImageTaskQueue(redis *redis.Client, logger *middleware.Logger) *ImageTaskQueue {
	return &ImageTaskQueue{
		TaskQueue: NewTaskQueue("images", redis, logger),
	}
}

// Task types for image processing
const (
	TaskTypeImageProcessing     = "image_processing"
	TaskTypeSchemaGeneration    = "schema_generation"
	TaskTypeEmailSending        = "email_sending"
	TaskTypeImageOptimization   = "image_optimization"
	TaskTypeThumbnailGeneration = "thumbnail_generation"
	TaskTypeAIProcessing        = "ai_processing" // AI processing via Stable Diffusion
	TaskTypeAIPriority          = "ai_priority"   // Priority AI processing
)

// EnqueueImageProcessing adds image processing task
func (q *ImageTaskQueue) EnqueueImageProcessing(imageID uuid.UUID, style string, parameters map[string]any) error {
	payload := map[string]any{
		"image_id":   imageID.String(),
		"style":      style,
		"parameters": parameters,
	}

	return q.Enqueue(TaskTypeImageProcessing, payload, WithPriority(5), WithMaxRetries(3))
}

// EnqueueSchemaGeneration adds schema generation task
func (q *ImageTaskQueue) EnqueueSchemaGeneration(imageID uuid.UUID, couponID uuid.UUID, confirmed bool) error {
	payload := map[string]any{
		"image_id":  imageID.String(),
		"coupon_id": couponID.String(),
		"confirmed": confirmed,
	}

	return q.Enqueue(TaskTypeSchemaGeneration, payload, WithPriority(8), WithMaxRetries(2))
}

// EnqueueEmailSending adds email sending task
func (q *ImageTaskQueue) EnqueueEmailSending(email string, schemaURL string, couponCode string) error {
	payload := map[string]any{
		"email":       email,
		"schema_url":  schemaURL,
		"coupon_code": couponCode,
	}

	return q.Enqueue(TaskTypeEmailSending, payload, WithPriority(3), WithMaxRetries(5))
}

// EnqueueImageOptimization adds image optimization task
func (q *ImageTaskQueue) EnqueueImageOptimization(imageID uuid.UUID, quality int) error {
	payload := map[string]any{
		"image_id": imageID.String(),
		"quality":  quality,
	}

	return q.Enqueue(TaskTypeImageOptimization, payload, WithPriority(2))
}

// EnqueueThumbnailGeneration adds thumbnail generation task
func (q *ImageTaskQueue) EnqueueThumbnailGeneration(imageID uuid.UUID, sizes []string) error {
	payload := map[string]any{
		"image_id": imageID.String(),
		"sizes":    sizes,
	}

	return q.Enqueue(TaskTypeThumbnailGeneration, payload, WithPriority(1))
}

// EnqueueAIProcessing adds AI image processing task via Stable Diffusion
func (q *ImageTaskQueue) EnqueueAIProcessing(
	imageID uuid.UUID,
	userEmail string,
	style string,
	useAI bool,
	parameters map[string]any,
	priority int,
) error {
	payload := map[string]any{
		"image_id":   imageID.String(),
		"user_email": userEmail,
		"style":      style,
		"use_ai":     useAI,
		"parameters": parameters,
		"priority":   priority,
	}

	// Determine priority based on parameters
	calculatedPriority := calculateAIPriority(style, useAI, priority, parameters)

	// AI tasks have high priority by default
	if calculatedPriority < 6 {
		calculatedPriority = 6
	}

	return q.Enqueue(TaskTypeAIProcessing, payload, WithPriority(calculatedPriority), WithMaxRetries(3))
}

// EnqueuePriorityAIProcessing adds priority AI processing task
func (q *ImageTaskQueue) EnqueuePriorityAIProcessing(
	imageID uuid.UUID,
	userEmail string,
	style string,
	useAI bool,
	parameters map[string]any,
) error {
	payload := map[string]any{
		"image_id":   imageID.String(),
		"user_email": userEmail,
		"style":      style,
		"use_ai":     useAI,
		"parameters": parameters,
		"priority":   10, // Maximum priority
	}

	return q.Enqueue(TaskTypeAIPriority, payload, WithPriority(10), WithMaxRetries(3))
}

// GetImageTaskHandlers returns map of handlers for image processing tasks
func GetImageTaskHandlers(imageService *ImageServiceAdapter, emailService *EmailServiceAdapter, logger *middleware.Logger) map[string]TaskHandler {
	return map[string]TaskHandler{
		"process_image": func(ctx context.Context, task *Task) error {
			return handleProcessImage(ctx, task, imageService, logger)
		},
		"generate_schema": func(ctx context.Context, task *Task) error {
			return handleGenerateSchema(ctx, task, imageService, logger)
		},
		"send_schema": func(ctx context.Context, task *Task) error { return handleSendSchema(ctx, task, emailService) },
		"optimize_image": func(ctx context.Context, task *Task) error {
			return handleOptimizeImage(ctx, task, imageService, logger)
		},
		"generate_thumbnails": func(ctx context.Context, task *Task) error {
			return handleGenerateThumbnails(ctx, task, imageService, logger)
		},
		"ai_processing": func(ctx context.Context, task *Task) error {
			return handleAIProcessing(ctx, task, imageService, logger)
		},
		"ai_priority": func(ctx context.Context, task *Task) error { return handleAIPriority(ctx, task, imageService, logger) },
	}
}

// Service interfaces

type ImageService interface {
	ProcessImageWithStyle(ctx context.Context, imageID uuid.UUID, style string, parameters map[string]any) error
	GenerateSchema(ctx context.Context, imageID uuid.UUID, confirmed bool) error
	OptimizeImage(ctx context.Context, imageID uuid.UUID, quality int) error
	GenerateThumbnails(ctx context.Context, imageID uuid.UUID, sizes []string) error
	ProcessImageWithAI(ctx context.Context, imageID uuid.UUID, style string, useAI bool, parameters map[string]any) error
}

type EmailService interface {
	SendSchema(ctx context.Context, email, schemaURL, couponCode string) error
}

// Concrete task handlers

func handleProcessImage(ctx context.Context, task *Task, imageService *ImageServiceAdapter, logger *middleware.Logger) error {
	payload := task.Payload

	imageIDStr, ok := payload["image_id"].(string)
	if !ok {
		logger.GetZerologLogger().Error().Interface("payload", payload).Str("task_type", task.Type).Msg("Invalid image_id in process image task")
		return fmt.Errorf("invalid image_id")
	}

	imageID, err := uuid.Parse(imageIDStr)
	if err != nil {
		return err
	}

	style, _ := payload["style"].(string)
	parameters, _ := payload["parameters"].(map[string]any)

	return imageService.ProcessImageWithStyle(ctx, imageID, style, parameters)
}

func handleGenerateSchema(ctx context.Context, task *Task, imageService *ImageServiceAdapter, logger *middleware.Logger) error {
	payload := task.Payload

	imageIDStr, ok := payload["image_id"].(string)
	if !ok {
		logger.GetZerologLogger().Error().Interface("payload", payload).Str("task_type", task.Type).Msg("Invalid image_id in generate schema task")
		return fmt.Errorf("invalid image_id")
	}

	imageID, err := uuid.Parse(imageIDStr)
	if err != nil {
		return err
	}

	confirmed, _ := payload["confirmed"].(bool)

	return imageService.GenerateSchema(ctx, imageID, confirmed)
}

func handleSendSchema(ctx context.Context, task *Task, emailService *EmailServiceAdapter) error {
	payload := task.Payload

	email, _ := payload["email"].(string)
	schemaURL, _ := payload["schema_url"].(string)
	couponCode, _ := payload["coupon_code"].(string)

	return emailService.SendSchema(ctx, email, schemaURL, couponCode)
}

func handleOptimizeImage(ctx context.Context, task *Task, imageService *ImageServiceAdapter, logger *middleware.Logger) error {
	payload := task.Payload

	imageIDStr, ok := payload["image_id"].(string)
	if !ok {
		logger.GetZerologLogger().Error().Interface("payload", payload).Str("task_type", task.Type).Msg("Invalid image_id in optimize image task")
		return fmt.Errorf("invalid image_id")
	}

	imageID, err := uuid.Parse(imageIDStr)
	if err != nil {
		return err
	}

	quality, _ := payload["quality"].(int)

	return imageService.OptimizeImage(ctx, imageID, quality)
}

func handleGenerateThumbnails(ctx context.Context, task *Task, imageService *ImageServiceAdapter, logger *middleware.Logger) error {
	payload := task.Payload

	imageIDStr, ok := payload["image_id"].(string)
	if !ok {
		logger.GetZerologLogger().Error().Interface("payload", payload).Str("task_type", task.Type).Msg("Invalid image_id in generate thumbnails task")
		return fmt.Errorf("invalid image_id")
	}

	imageID, err := uuid.Parse(imageIDStr)
	if err != nil {
		return err
	}

	sizes, _ := payload["sizes"].([]string)

	return imageService.GenerateThumbnails(ctx, imageID, sizes)
}

func handleAIProcessing(ctx context.Context, task *Task, imageService *ImageServiceAdapter, logger *middleware.Logger) error {
	payload := task.Payload

	imageIDStr, ok := payload["image_id"].(string)
	if !ok {
		logger.GetZerologLogger().Error().Interface("payload", payload).Str("task_type", task.Type).Msg("Invalid image_id in AI processing task")
		return fmt.Errorf("invalid image_id")
	}

	imageID, err := uuid.Parse(imageIDStr)
	if err != nil {
		return err
	}

	style, _ := payload["style"].(string)
	parameters, _ := payload["parameters"].(map[string]any)

	return imageService.ProcessImageWithStyle(ctx, imageID, style, parameters)
}

// handleAIPriority processes priority AI processing task
func handleAIPriority(ctx context.Context, task *Task, imageService *ImageServiceAdapter, logger *middleware.Logger) error {
	return handleAIProcessing(ctx, task, imageService, logger)
}

// calculateAIPriority calculates priority for AI task
func calculateAIPriority(style string, useAI bool, basePriority int, parameters map[string]any) int {
	if !useAI {
		return basePriority
	}

	priority := basePriority

	// Increase priority for certain styles
	switch style {
	case "pop_art", "max_colors":
		priority += 2
	case "grayscale", "skin_tones":
		priority += 1
	}

	// Increase priority for AI tasks
	priority += 3

	// Check additional parameters
	if lighting, ok := parameters["lighting"].(string); ok && lighting != "" {
		priority += 1
	}
	if contrast, ok := parameters["contrast"].(string); ok && contrast != "" {
		priority += 1
	}

	// Limit priority
	if priority > 10 {
		priority = 10
	}
	if priority < 1 {
		priority = 1
	}

	return priority
}
