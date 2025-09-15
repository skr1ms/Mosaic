package image

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/skr1ms/mosaic/internal/types"
	"github.com/skr1ms/mosaic/pkg/middleware"
)

type ImageHandlerDeps struct {
	ImageService    ImageServiceInterface
	ImageRepository ImageRepositoryInterface
	Logger          *middleware.Logger
}

type ImageHandler struct {
	fiber.Router
	deps *ImageHandlerDeps
}

func NewImageProcessingHandler(app fiber.Router, deps *ImageHandlerDeps) *ImageHandler {
	handler := &ImageHandler{
		Router: app,
		deps:   deps,
	}

	// ================================================================
	// PUBLIC IMAGE ROUTES: /api/public/*
	// Access: public (no authentication required)
	// ================================================================
	public := handler.Group("/public")
	public.Post("/upload", handler.UploadImage)                                       // POST /api/public/upload
	public.Put("/images/:id/edit", handler.EditImage)                                 // PUT /api/public/images/:id/edit
	public.Put("/images/:id/process", handler.ProcessImage)                           // PUT /api/public/images/:id/process
	public.Post("/images/:id/generate-schema", handler.GenerateSchema)                // POST /api/public/images/:id/generate-schema
	public.Get("/images/:id/status", handler.GetImageStatus)                          // GET /api/public/images/:id/status
	public.Get("/schemas/:schema_uuid/download", handler.DownloadSchemaArchivePublic) // GET /api/public/schemas/:schema_uuid/download

	// ================================================================
	// ADMIN IMAGE ROUTES: /api/admin/*
	// Access: admin and main_admin roles only
	// ================================================================
	admin := handler.Group("/admin")
	admin.Get("/queue", handler.GetQueue)                        // GET /api/admin/queue
	admin.Get("/queue/:id", handler.GetTaskByID)                 // GET /api/admin/queue/:id
	admin.Post("/queue", handler.AddToQueue)                     // POST /api/admin/queue
	admin.Put("/queue/:id/start", handler.StartProcessing)       // PUT /api/admin/queue/:id/start
	admin.Put("/queue/:id/complete", handler.CompleteProcessing) // PUT /api/admin/queue/:id/complete
	admin.Put("/queue/:id/fail", handler.FailProcessing)         // PUT /api/admin/queue/:id/fail
	admin.Put("/queue/:id/retry", handler.RetryTask)             // PUT /api/admin/queue/:id/retry
	admin.Delete("/queue/:id", handler.DeleteTask)               // DELETE /api/admin/queue/:id
	admin.Get("/statistics", handler.GetStatistics)              // GET /api/admin/statistics
	admin.Get("/next", handler.GetNextTask)                      // GET /api/admin/next

	return handler
}

// @Summary Upload image
// @Description Uploads user image for diamond mosaic creation
// @Tags public-images
// @Accept multipart/form-data
// @Produce json
// @Param coupon_code formData string true "12-character coupon code"
// @Param image formData file true "Image file (JPG, PNG, max 10MB)"
// @Success 200 {object} types.ImageUploadResponse "Image successfully uploaded"
// @Failure 400 {object} map[string]string "Validation error - invalid coupon code format or coupon not activated"
// @Failure 404 {object} map[string]string "Coupon not found"
// @Failure 500 {object} map[string]string "Internal server error - failed to find coupon, get uploaded image, or upload image"
// @Router /public/upload [post]
func (handler *ImageHandler) UploadImage(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	couponCode := c.FormValue("coupon_code")
	if len(couponCode) != 12 {
		handler.deps.Logger.FromContext(c).Error().Msg("Invalid coupon code format")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid coupon code format",
		})
	}

	coupon, err := handler.deps.ImageService.GetCouponRepository().GetByCode(ctx, couponCode)
	if err != nil {
		if err.Error() == "not found" {
			handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Coupon not found")
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Coupon not found",
			})
		}
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Error finding coupon")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error finding coupon",
		})
	}

	file, err := c.FormFile("image")
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Error getting uploaded image")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Error getting uploaded image",
		})
	}

	// Use empty email if not set (removed email requirement)
	userEmail := ""
	if coupon.UserEmail != nil && *coupon.UserEmail != "" {
		userEmail = *coupon.UserEmail
	}

	imageRecord, err := handler.deps.ImageService.UploadImage(ctx, coupon.ID, file, userEmail)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Error uploading image")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error uploading image",
		})
	}

	handler.deps.Logger.FromContext(c).Info().Interface("context", map[string]any{
		"image_id":     imageRecord.ID,
		"coupon_code":  couponCode,
		"coupon_size":  coupon.Size,
		"coupon_style": coupon.Style,
	}).Msg("Image uploaded successfully")

	return c.JSON(types.ImageUploadResponse{
		Message:     "Image successfully uploaded",
		ImageID:     imageRecord.ID,
		NextStep:    "edit_image",
		CouponSize:  coupon.Size,
		CouponStyle: coupon.Style,
	})
}

// @Summary Edit image
// @Description Applies cropping, rotation and scaling to uploaded image
// @Tags public-images
// @Accept json
// @Produce json
// @Param id path string true "Image ID (UUID format)"
// @Param params body ImageEditParams true "Edit parameters including crop, rotation and scale settings"
// @Success 200 {object} types.ImageEditResponse "Image successfully edited"
// @Failure 400 {object} map[string]string "Validation error - invalid image ID format or parse error"
// @Failure 404 {object} map[string]string "Image not found"
// @Failure 500 {object} map[string]string "Internal server error - failed to edit image or get image status"
// @Router /public/images/{id}/edit [put]
func (handler *ImageHandler) EditImage(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	imageIDStr := c.Params("id")
	imageID, err := uuid.Parse(imageIDStr)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Invalid image ID format")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid image ID format",
		})
	}

	var editParams ImageEditParams
	if err := c.BodyParser(&editParams); err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Error parsing edit parameters")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Error parsing edit parameters",
		})
	}

	if err := handler.deps.ImageService.EditImage(ctx, imageID, editParams); err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Error editing image")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error editing image",
		})
	}

	status, err := handler.deps.ImageService.GetImageStatus(ctx, imageID)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Error getting image status")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error getting image status",
		})
	}

	previewURL := ""
	if status.EditedURL != nil {
		previewURL = *status.EditedURL
	}

	handler.deps.Logger.FromContext(c).Info().Interface("context", map[string]any{
		"image_id":    imageID,
		"preview_url": previewURL,
	}).Msg("Image edited successfully")

	return c.JSON(types.ImageEditResponse{
		Message:    "Image successfully edited",
		ImageID:    imageID,
		NextStep:   "process_image",
		PreviewURL: previewURL,
	})
}

// @Summary Process image
// @Description Applies selected processing styles to image, including AI processing through Stable Diffusion
// @Tags public-images
// @Accept json
// @Produce json
// @Param id path string true "Image ID (UUID format)"
// @Param params body types.ProcessImageRequest true "Processing parameters including style, AI settings, lighting, contrast, brightness, and saturation"
// @Success 200 {object} types.ProcessImageResponse "Image processing started successfully"
// @Failure 400 {object} map[string]string "Validation error - invalid image ID format or parse error"
// @Failure 404 {object} map[string]string "Image not found"
// @Failure 500 {object} map[string]string "Internal server error - failed to get image status"
// @Router /public/images/{id}/process [put]
func (handler *ImageHandler) ProcessImage(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	imageIDStr := c.Params("id")
	imageID, err := uuid.Parse(imageIDStr)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Invalid image ID format")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid image ID format",
		})
	}

	var processRequest types.ProcessImageRequest
	if err := c.BodyParser(&processRequest); err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Error parsing process parameters")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Error parsing process parameters",
		})
	}

	processParams := ProcessingParams{
		Style:      processRequest.Style,
		UseAI:      processRequest.UseAI,
		Lighting:   processRequest.Lighting,
		Contrast:   processRequest.Contrast,
		Brightness: processRequest.Brightness,
		Saturation: processRequest.Saturation,
	}

	// We start processing in the background with a separate context
	handler.processImageAsync(imageID, &processParams)

	status, err := handler.deps.ImageService.GetImageStatus(ctx, imageID)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Error getting image status")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error getting image status",
		})
	}

	previewURL := ""
	originalURL := ""
	if status.PreviewURL != nil {
		previewURL = *status.PreviewURL
	}
	if status.OriginalURL != nil {
		originalURL = *status.OriginalURL
	}

	handler.deps.Logger.FromContext(c).Info().Interface("context", map[string]any{
		"image_id":     imageID,
		"style":        processRequest.Style,
		"use_ai":       processRequest.UseAI,
		"preview_url":  previewURL,
		"original_url": originalURL,
	}).Msg("Image processing started")

	return c.JSON(types.ProcessImageResponse{
		Message:     "Processing started",
		ImageID:     imageID,
		NextStep:    "generate_schema",
		PreviewURL:  previewURL,
		OriginalURL: originalURL,
	})
}

// @Summary Generate mosaic schema
// @Description Creates final diamond mosaic schema with instructions and color map
// @Tags public-images
// @Accept json
// @Produce json
// @Param id path string true "Image ID (UUID format)"
// @Param params body types.GenerateSchemaRequest true "Schema generation confirmation - must be true to proceed"
// @Success 200 {object} types.GenerateSchemaResponse "Schema generation started successfully"
// @Failure 400 {object} map[string]string "Validation error - invalid image ID format, parse error, or schema request not confirmed"
// @Failure 404 {object} map[string]string "Image not found"
// @Failure 500 {object} map[string]string "Internal server error - failed to generate schema"
// @Router /public/images/{id}/generate-schema [post]
func (handler *ImageHandler) GenerateSchema(c *fiber.Ctx) error {
	imageIDStr := c.Params("id")
	imageID, err := uuid.Parse(imageIDStr)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Invalid image ID format")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid image ID format",
		})
	}

	var schemaRequest types.GenerateSchemaRequest
	if err := c.BodyParser(&schemaRequest); err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Error parsing schema request")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Error parsing schema request",
		})
	}

	if !schemaRequest.Confirmed {
		handler.deps.Logger.FromContext(c).Error().Msg("Schema request not confirmed")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Schema request not confirmed",
		})
	}

	// Starting the scheme generation in the background with a separate context
	handler.generateSchemaAsync(imageID, schemaRequest.Confirmed)

	handler.deps.Logger.FromContext(c).Info().Interface("context", map[string]any{
		"image_id":  imageID,
		"confirmed": schemaRequest.Confirmed,
	}).Msg("Schema generation started")

	return c.JSON(types.GenerateSchemaResponse{
		Message:   "Schema generation started",
		ImageID:   imageID,
		EmailSent: true,
	})
}

// @Summary Get image processing status
// @Description Returns current image processing status and file links
// @Tags public-images
// @Produce json
// @Param id path string true "Image ID (UUID format)"
// @Success 200 {object} types.ImageStatusResponse "Image status retrieved successfully"
// @Failure 400 {object} map[string]string "Validation error - invalid image ID format"
// @Failure 404 {object} map[string]string "Image not found"
// @Failure 500 {object} map[string]string "Internal server error - failed to get image status"
// @Router /public/images/{id}/status [get]
func (handler *ImageHandler) GetImageStatus(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	imageIDStr := c.Params("id")
	imageID, err := uuid.Parse(imageIDStr)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Invalid image ID format")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid image ID format",
		})
	}

	status, err := handler.deps.ImageService.GetImageStatus(ctx, imageID)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Image not found")
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Image not found",
		})
	}

	handler.deps.Logger.FromContext(c).Info().Interface("context", map[string]any{
		"image_id": imageID,
		"status":   status.Status,
	}).Msg("Image status retrieved")

	return c.JSON(status)
}

// downloadSchemaArchive - a common function for downloading the schema archive
func (handler *ImageHandler) downloadSchemaArchive(c *fiber.Ctx, isPublic bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	schemaUUID := c.Params("schema_uuid")
	if schemaUUID == "" {
		handler.deps.Logger.FromContext(c).Error().Interface("context", map[string]any{"schema_uuid": schemaUUID}).Msg("Schema UUID is required")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Schema UUID is required",
		})
	}

	parsedUUID, err := uuid.Parse(schemaUUID)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Interface("context", map[string]any{"schema_uuid": schemaUUID}).Msg("Invalid UUID format")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid UUID format",
		})
	}

	imageRecord, err := handler.deps.ImageRepository.GetByID(ctx, parsedUUID)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Interface("context", map[string]any{"schema_uuid": schemaUUID}).Msg("Schema not found")
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Schema not found",
		})
	}

	if imageRecord.Status != "completed" || imageRecord.SchemaS3Key == nil {
		handler.deps.Logger.FromContext(c).Error().Interface("context", map[string]any{"schema_uuid": schemaUUID}).Msg("Schema is not ready for download")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Schema is not ready for download",
		})
	}

	downloadURL, err := handler.deps.ImageService.GetS3Client().GetFileURL(ctx, *imageRecord.SchemaS3Key, 1*time.Hour)
	if err != nil {
		log.Error().
			Err(err).
			Str("schema_s3_key", *imageRecord.SchemaS3Key).
			Msg("Failed to generate download URL")
		handler.deps.Logger.FromContext(c).Error().Err(err).Interface("context", map[string]any{"schema_uuid": schemaUUID, "s3_key": *imageRecord.SchemaS3Key}).Msg("Failed to generate download URL")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate download URL",
		})
	}

	logContext := map[string]any{
		"schema_uuid": schemaUUID,
		"s3_key":      *imageRecord.SchemaS3Key,
		"is_public":   isPublic,
	}

	handler.deps.Logger.FromContext(c).Info().Interface("context", logContext).Msg("Schema archive download initiated")

	return c.Redirect(downloadURL, fiber.StatusFound)
}

// @Summary Download schema archive (public)
// @Description Allows user to download ZIP archive with mosaic schema files
// @Tags public-images
// @Produce application/zip
// @Param schema_uuid path string true "Schema UUID (UUID format)"
// @Success 200 {file} binary "ZIP archive with schema files"
// @Failure 400 {object} map[string]string "Validation error - schema UUID required, invalid UUID format, or schema not ready for download"
// @Failure 404 {object} map[string]string "Schema not found"
// @Failure 500 {object} map[string]string "Internal server error - failed to generate download URL"
// @Router /public/schemas/{schema_uuid}/download [get]
func (handler *ImageHandler) DownloadSchemaArchivePublic(c *fiber.Ctx) error {
	return handler.downloadSchemaArchive(c, true)
}

// @Summary Get processing queue
// @Description Returns all tasks in image processing queue with optional status filtering
// @Tags admin-image-processing
// @Produce json
// @Param status query string false "Task status filter (uploaded, edited, processing, processed, completed, failed)"
// @Param date_from query string false "Start date filter (YYYY-MM-DD)"
// @Param date_to query string false "End date filter (YYYY-MM-DD)"
// @Success 200 {array} map[string]any "List of tasks in processing queue"
// @Failure 500 {object} map[string]string "Internal server error - failed to get queue"
// @Router /admin/queue [get]
func (handler *ImageHandler) GetQueue(c *fiber.Ctx) error {
	status := c.Query("status")
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")

	tasks, err := handler.deps.ImageService.GetQueueWithFilters(status, dateFrom, dateTo)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Error getting queue")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error getting queue",
		})
	}

	handler.deps.Logger.FromContext(c).Info().Interface("context", map[string]any{
		"status":    status,
		"date_from": dateFrom,
		"date_to":   dateTo,
		"count":     len(tasks),
	}).Msg("Processing queue retrieved")

	return c.JSON(tasks)
}

// @Summary Get task by ID
// @Description Returns detailed information about image processing task
// @Tags admin-image-processing
// @Produce json
// @Param id path string true "Task ID (UUID format)"
// @Success 200 {object} map[string]any "Task information with status details"
// @Failure 400 {object} map[string]string "Validation error - invalid task ID format"
// @Failure 404 {object} map[string]string "Task not found"
// @Failure 500 {object} map[string]string "Internal server error - failed to get task"
// @Router /admin/queue/{id} [get]
func (handler *ImageHandler) GetTaskByID(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Invalid ID format")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid ID format",
		})
	}

	task, err := handler.deps.ImageRepository.GetByID(ctx, id)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Task not found")
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Task not found",
		})
	}

	status, err := handler.deps.ImageService.GetImageStatus(ctx, id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get image status")
		return c.JSON(task)
	}

	result := map[string]any{
		"id":            task.ID,
		"coupon_id":     task.CouponID,
		"user_email":    task.UserEmail,
		"status":        task.Status,
		"priority":      task.Priority,
		"created_at":    task.CreatedAt,
		"started_at":    task.StartedAt,
		"completed_at":  task.CompletedAt,
		"error_message": task.ErrorMessage,
		"progress":      status.Progress,
		"message":       status.Message,
		"original_url":  status.OriginalURL,
		"edited_url":    status.EditedURL,
		"processed_url": status.ProcessedURL,
		"preview_url":   status.PreviewURL,
		"zip_url":       status.ZipURL,
	}

	handler.deps.Logger.FromContext(c).Info().Interface("context", map[string]any{
		"task_id": id,
		"status":  task.Status,
	}).Msg("Task retrieved by ID")

	return c.JSON(result)
}

// @Summary Add task to queue
// @Description Adds new image processing task to queue
// @Tags admin-image-processing
// @Accept json
// @Produce json
// @Param task body AddToQueueRequest true "Task parameters including coupon ID, image path, processing params, user email and priority"
// @Success 201 {object} map[string]any "Task added to queue successfully"
// @Failure 400 {object} map[string]string "Validation error - failed to parse request"
// @Failure 500 {object} map[string]string "Internal server error - failed to add task to queue"
// @Router /admin/queue [post]
func (handler *ImageHandler) AddToQueue(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var req AddToQueueRequest
	if err := c.BodyParser(&req); err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Error parsing request")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Error parsing request",
		})
	}

	task := &Image{
		CouponID:           req.CouponID,
		OriginalImageS3Key: req.OriginalImagePath,
		ProcessingParams:   req.ProcessingParams,
		UserEmail:          req.UserEmail,
		Status:             "queued",
		Priority:           req.Priority,
	}

	if err := handler.deps.ImageRepository.Create(ctx, task); err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Error adding task to queue")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error adding task to queue",
		})
	}

	handler.deps.Logger.FromContext(c).Info().Interface("context", map[string]any{
		"task_id":    task.ID,
		"coupon_id":  req.CouponID,
		"user_email": req.UserEmail,
		"priority":   req.Priority,
	}).Msg("Task added to queue")

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Task added to queue",
		"task_id": task.ID,
	})
}

// @Summary Start task processing
// @Description Marks task as being processed
// @Tags admin-image-processing
// @Produce json
// @Param id path string true "Task ID (UUID format)"
// @Success 200 {object} map[string]any "Processing started successfully"
// @Failure 400 {object} map[string]string "Validation error - invalid task ID format"
// @Failure 500 {object} map[string]string "Internal server error - failed to start processing"
// @Router /admin/queue/{id}/start [put]
func (handler *ImageHandler) StartProcessing(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Invalid ID format")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid ID format",
		})
	}

	if err := handler.deps.ImageRepository.StartProcessing(ctx, id); err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Error starting processing")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error starting processing",
		})
	}

	handler.deps.Logger.FromContext(c).Info().Interface("context", map[string]any{
		"task_id": id,
	}).Msg("Task processing started")

	return c.JSON(fiber.Map{
		"message": "Processing started",
	})
}

// @Summary Complete task processing
// @Description Marks task as successfully completed
// @Tags admin-image-processing
// @Produce json
// @Param id path string true "Task ID (UUID format)"
// @Success 200 {object} map[string]any "Task completed successfully"
// @Failure 400 {object} map[string]string "Validation error - invalid task ID format"
// @Failure 500 {object} map[string]string "Internal server error - failed to complete processing"
// @Router /admin/queue/{id}/complete [put]
func (handler *ImageHandler) CompleteProcessing(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Invalid ID format")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid ID format",
		})
	}

	if err := handler.deps.ImageRepository.CompleteProcessing(ctx, id); err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Error completing processing")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error completing processing",
		})
	}

	handler.deps.Logger.FromContext(c).Info().Interface("context", map[string]any{
		"task_id": id,
	}).Msg("Task processing completed")

	return c.JSON(fiber.Map{
		"message": "Processing completed",
	})
}

// @Summary Fail task processing
// @Description Marks task as failed with error reason
// @Tags admin-image-processing
// @Accept json
// @Produce json
// @Param id path string true "Task ID (UUID format)"
// @Param error body FailProcessingRequest true "Error message details"
// @Success 200 {object} map[string]any "Task marked as failed successfully"
// @Failure 400 {object} map[string]string "Validation error - invalid task ID format or failed to parse request"
// @Failure 500 {object} map[string]string "Internal server error - failed to mark task as failed"
// @Router /admin/queue/{id}/fail [put]
func (handler *ImageHandler) FailProcessing(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Invalid ID format")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid ID format",
		})
	}

	var req FailProcessingRequest
	if err := c.BodyParser(&req); err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Error parsing request")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Error parsing request",
		})
	}

	if err := handler.deps.ImageRepository.FailProcessing(ctx, id, req.ErrorMessage); err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Error failing processing")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error failing processing",
		})
	}

	handler.deps.Logger.FromContext(c).Info().Interface("context", map[string]any{
		"task_id":       id,
		"error_message": req.ErrorMessage,
	}).Msg("Task processing failed")

	return c.JSON(fiber.Map{
		"message": "Processing failed",
	})
}

// @Summary Retry task processing
// @Description Returns failed task back to processing queue
// @Tags admin-image-processing
// @Produce json
// @Param id path string true "Task ID (UUID format)"
// @Success 200 {object} map[string]any "Task returned to queue successfully"
// @Failure 400 {object} map[string]string "Validation error - invalid task ID format"
// @Failure 500 {object} map[string]string "Internal server error - failed to retry task"
// @Router /admin/queue/{id}/retry [put]
func (handler *ImageHandler) RetryTask(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Invalid ID format")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid ID format",
		})
	}

	if err := handler.deps.ImageRepository.RetryTask(ctx, id); err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Error retrying task")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error retrying task",
		})
	}

	handler.deps.Logger.FromContext(c).Info().Interface("context", map[string]any{
		"task_id": id,
	}).Msg("Task retried")

	return c.JSON(fiber.Map{
		"message": "Task retried",
	})
}

// @Summary Delete task
// @Description Removes task from processing queue
// @Tags admin-image-processing
// @Produce json
// @Param id path string true "Task ID (UUID format)"
// @Success 200 {object} map[string]any "Task deleted successfully"
// @Failure 400 {object} map[string]string "Validation error - invalid task ID format"
// @Failure 500 {object} map[string]string "Internal server error - failed to delete task"
// @Router /admin/queue/{id} [delete]
func (handler *ImageHandler) DeleteTask(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Interface("context", map[string]any{"task_id": idStr}).Msg("Invalid ID format")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid ID format",
		})
	}

	if err := handler.deps.ImageRepository.Delete(ctx, id); err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Interface("context", map[string]any{"task_id": id}).Msg("Error deleting task")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error deleting task",
		})
	}

	handler.deps.Logger.FromContext(c).Info().Interface("context", map[string]any{"task_id": id}).Msg("Task deleted")

	return c.JSON(fiber.Map{
		"message": "Task deleted",
	})
}

// @Summary Get processing statistics
// @Description Returns image processing statistics
// @Tags admin-image-processing
// @Produce json
// @Success 200 {object} map[string]any "Processing statistics"
// @Failure 500 {object} map[string]string "Internal server error - failed to get statistics"
// @Router /admin/statistics [get]
func (handler *ImageHandler) GetStatistics(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stats, err := handler.deps.ImageRepository.GetStatistics(ctx)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Error getting statistics")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error getting statistics",
		})
	}

	handler.deps.Logger.FromContext(c).Info().Msg("Processing statistics retrieved")
	return c.JSON(stats)
}

// @Summary Get next task for processing
// @Description Returns next task in queue for processing
// @Tags admin-image-processing
// @Produce json
// @Success 200 {object} map[string]any "Next task for processing"
// @Failure 404 {object} map[string]string "No tasks in queue"
// @Failure 500 {object} map[string]string "Internal server error - failed to get next task"
// @Router /admin/next [get]
func (handler *ImageHandler) GetNextTask(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	task, err := handler.deps.ImageRepository.GetNextInQueue(ctx)
	if err != nil {
		if err.Error() == "no tasks in queue" {
			handler.deps.Logger.FromContext(c).Error().Err(err).Msg("No tasks in queue")
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "No tasks in queue",
			})
		}
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Error getting next task")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error getting next task",
		})
	}

	handler.deps.Logger.FromContext(c).Info().Msg("Next task for processing retrieved")
	return c.JSON(task)
}

// We start processing in the background with a separate context
func (handler *ImageHandler) processImageAsync(imageID uuid.UUID, processParams *ProcessingParams) {
	go func() {
		bgCtx, bgCancel := context.WithTimeout(context.Background(), 2*time.Hour)
		defer bgCancel()

		logger := zerolog.Ctx(bgCtx).With().
			Str("image_id", imageID.String()).
			Str("style", processParams.Style).
			Bool("use_ai", processParams.UseAI).
			Logger()

		if err := handler.deps.ImageService.ProcessImage(bgCtx, imageID, processParams); err != nil {
			logger.Error().
				Err(err).
				Msg("Failed to process image")
		} else {
			logger.Info().Msg("Image processing completed successfully")
		}
	}()
}

// Starting the scheme generation in the background with a separate context
func (handler *ImageHandler) generateSchemaAsync(imageID uuid.UUID, confirmed bool) {
	go func() {
		bgCtx, bgCancel := context.WithTimeout(context.Background(), 1*time.Hour)
		defer bgCancel()

		logger := zerolog.Ctx(bgCtx).With().
			Str("image_id", imageID.String()).
			Bool("confirmed", confirmed).
			Logger()

		if err := handler.deps.ImageService.GenerateSchema(bgCtx, imageID, confirmed); err != nil {
			logger.Error().
				Err(err).
				Msg("Failed to generate schema")
		} else {
			logger.Info().Msg("Schema generation completed successfully")
		}
	}()
}
