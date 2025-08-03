package image

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/skr1ms/mosaic/internal/coupon"
	"github.com/skr1ms/mosaic/internal/types"
)

type ImageHandlerDeps struct {
	CouponRepository *coupon.CouponRepository
	ImageService     *ImageService
	ImageRepository  *ImageRepository
}

type ImageHandler struct {
	fiber.Router
	deps *ImageHandlerDeps
}

func NewImageProcessingHandler(router fiber.Router, deps *ImageHandlerDeps) {
	handler := &ImageHandler{
		Router: router,
		deps:   deps,
	}

	// API для пользователей (публичные endpoints)
	public := handler.Group("/public")
	public.Post("/upload", handler.UploadImage)                        // Загрузка изображения
	public.Put("/images/:id/edit", handler.EditImage)                  // Редактирование изображения
	public.Put("/images/:id/process", handler.ProcessImage)            // Обработка изображения (выбор стилей)
	public.Post("/images/:id/generate-schema", handler.GenerateSchema) // Создание схемы
	public.Get("/images/:id/status", handler.GetImageStatus)           // Получение статуса обработки

	// API для администрирования (внутренние endpoints)
	admin := handler.Group("/admin")
	admin.Get("/queue", handler.GetQueue)                        // Получение всех задач в очереди
	admin.Get("/queue/:id", handler.GetTaskByID)                 // Получение задачи по ID
	admin.Post("/queue", handler.AddToQueue)                     // Добавление задачи в очередь
	admin.Put("/queue/:id/start", handler.StartProcessing)       // Начало обработки задачи
	admin.Put("/queue/:id/complete", handler.CompleteProcessing) // Завершение обработки задачи
	admin.Put("/queue/:id/fail", handler.FailProcessing)         // Провал обработки задачи
	admin.Put("/queue/:id/retry", handler.RetryTask)             // Повторная попытка обработки задачи
	admin.Delete("/queue/:id", handler.DeleteTask)               // Удаление задачи из очереди
	admin.Get("/statistics", handler.GetStatistics)              // Получение статистики по задачам
	admin.Get("/next", handler.GetNextTask)                      // Получение следующей задачи для обработки
}

// UploadImage загружает изображение пользователя
// @Summary		Загрузка изображения
// @Description	Загружает изображение пользователя для создания алмазной мозаики
// @Tags		public-images
// @Accept		multipart/form-data
// @Produce		json
// @Param		coupon_code	formData	string					true	"12-значный код купона"
// @Param		image		formData	file					true	"Изображение (JPG, PNG, макс. 10MB)"
// @Success		200		{object}	types.ImageUploadResponse	"Изображение успешно загружено"
// @Failure		400		{object}	map[string]string		"Ошибка валидации"
// @Failure		500		{object}	map[string]string		"Внутренняя ошибка сервера"
// @Router		/public/upload [post]
func (h *ImageHandler) UploadImage(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	// Получаем код купона из формы
	couponCode := c.FormValue("coupon_code")
	if len(couponCode) != 12 {
		log.Error().Msg("Invalid coupon code format")
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid coupon code format",
		})
	}

	// Получаем купон по коду
	coupon, err := h.deps.CouponRepository.GetByCode(context.Background(), couponCode)
	if err != nil {
		if err.Error() == "not found" {
			log.Error().Msg("Coupon not found")
			return c.Status(404).JSON(fiber.Map{
				"error": "Coupon not found",
			})
		}
		log.Error().Err(err).Msg("Error finding coupon")
		return c.Status(500).JSON(fiber.Map{
			"error": "Error finding coupon",
		})
	}

	if coupon.Status != "activated" {
		log.Error().Msg("Coupon not activated")
		return c.Status(400).JSON(fiber.Map{
			"error": "Coupon not activated",
		})
	}

	// Получаем загружаемый файл
	file, err := c.FormFile("image")
	if err != nil {
		log.Error().Err(err).Msg("Error getting uploaded image")
		return c.Status(400).JSON(fiber.Map{
			"error": "error getting uploaded image",
		})
	}

	// Загружаем изображение через сервис
	imageRecord, err := h.deps.ImageService.UploadImage(context.Background(), coupon.ID, file, *coupon.UserEmail)
	if err != nil {
		log.Error().Err(err).Msg("Error uploading image")
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(types.ImageUploadResponse{
		Message:     "Image successfully uploaded",
		ImageID:     imageRecord.ID,
		NextStep:    "edit_image",
		CouponSize:  coupon.Size,
		CouponStyle: coupon.Style,
	})
}

// EditImage применяет редактирование к изображению
// @Summary		Редактирование изображения
// @Description	Применяет кадрирование, поворот и масштабирование к загруженному изображению
// @Tags		public-images
// @Accept		json
// @Produce		json
// @Param		id		path		string					true	"ID изображения"
// @Param		params		body		ImageEditParams				true	"Параметры редактирования"
// @Success		200		{object}	types.ImageEditResponse			"Изображение успешно отредактировано"
// @Failure		400		{object}	map[string]string			"Ошибка валидации"
// @Failure		404		{object}	map[string]string			"Изображение не найдено"
// @Failure		500		{object}	map[string]string			"Внутренняя ошибка сервера"
// @Router		/public/images/{id}/edit [put]
func (h *ImageHandler) EditImage(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	// Парсим ID изображения
	imageIDStr := c.Params("id")
	imageID, err := uuid.Parse(imageIDStr)
	if err != nil {
		log.Error().Err(err).Msg("Invalid image ID format")
		return c.Status(400).JSON(fiber.Map{
			"error": "error parsing image ID",
		})
	}

	// Парсим параметры редактирования
	var editParams ImageEditParams
	if err := c.BodyParser(&editParams); err != nil {
		log.Error().Err(err).Msg("Error parsing edit parameters")
		return c.Status(400).JSON(fiber.Map{
			"error": "error parsing edit parameters",
		})
	}

	// Применяем редактирование
	if err := h.deps.ImageService.EditImage(context.Background(), imageID, editParams); err != nil {
		log.Error().Err(err).Msg("Error editing image")
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Получаем статус для возврата URL превью
	status, err := h.deps.ImageService.GetImageStatus(context.Background(), imageID)
	if err != nil {
		log.Error().Err(err).Msg("Error getting image status")
		return c.Status(500).JSON(fiber.Map{
			"error": "error getting image status",
		})
	}

	previewURL := ""
	if status.EditedURL != nil {
		previewURL = *status.EditedURL
	}

	return c.JSON(types.ImageEditResponse{
		Message:    "Image successfully edited",
		ImageID:    imageID,
		NextStep:   "process_image",
		PreviewURL: previewURL,
	})
}

// ProcessImage обрабатывает изображение (применяет стили)
// @Summary		Обработка изображения
// @Description	Применяет выбранные стили обработки к изображению, включая AI обработку через Stable Diffusion
// @Tags		public-images
// @Accept		json
// @Produce		json
// @Param		id		path		string					true	"ID изображения"
// @Param		params		body		types.ProcessImageRequest		true	"Параметры обработки"
// @Success		200		{object}	types.ProcessImageResponse		"Изображение успешно обработано"
// @Failure		400		{object}	map[string]string			"Ошибка валидации"
// @Failure		404		{object}	map[string]string			"Изображение не найдено"
// @Failure		500		{object}	map[string]string			"Внутренняя ошибка сервера"
// @Router		/public/images/{id}/process [put]
func (h *ImageHandler) ProcessImage(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	imageIDStr := c.Params("id")
	imageID, err := uuid.Parse(imageIDStr)
	if err != nil {
		log.Error().Err(err).Msg("Invalid image ID format")
		return c.Status(400).JSON(fiber.Map{
			"error": "error parsing image ID",
		})
	}

	var processRequest types.ProcessImageRequest
	if err := c.BodyParser(&processRequest); err != nil {
		log.Error().Err(err).Msg("Error parsing process parameters")
		return c.Status(400).JSON(fiber.Map{
			"error": "error parsing process parameters",
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

	// Запускаем обработку (асинхронно)
	go func() {
		if err := h.deps.ImageService.ProcessImage(context.Background(), imageID, processParams); err != nil {
			zerolog.Ctx(context.Background()).Error().
				Err(err).
				Str("image_id", imageID.String()).
				Msg("Failed to process image")
		}
	}()

	// Получаем текущий статус
	status, err := h.deps.ImageService.GetImageStatus(context.Background(), imageID)
	if err != nil {
		log.Error().Err(err).Msg("Error getting image status")
		return c.Status(500).JSON(fiber.Map{
			"error": "error getting image status",
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

	return c.JSON(types.ProcessImageResponse{
		Message:     "Processing started",
		ImageID:     imageID,
		NextStep:    "generate_schema",
		PreviewURL:  previewURL,
		OriginalURL: originalURL,
	})
}

// GenerateSchema создает финальную схему алмазной мозаики
// @Summary		Создание схемы мозаики
// @Description	Создает финальную схему алмазной мозаики с инструкциями и цветовой картой
// @Tags		public-images
// @Accept		json
// @Produce		json
// @Param		id		path		string					true	"ID изображения"
// @Param		params		body		types.GenerateSchemaRequest		true	"Подтверждение создания схемы"
// @Success		200		{object}	types.GenerateSchemaResponse		"Схема успешно создана"
// @Failure		400		{object}	map[string]string			"Ошибка валидации"
// @Failure		404		{object}	map[string]string			"Изображение не найдено"
// @Failure		500		{object}	map[string]string			"Внутренняя ошибка сервера"
// @Router		/public/images/{id}/generate-schema [post]
func (h *ImageHandler) GenerateSchema(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	imageIDStr := c.Params("id")
	imageID, err := uuid.Parse(imageIDStr)
	if err != nil {
		log.Error().Err(err).Msg("Invalid image ID format")
		return c.Status(400).JSON(fiber.Map{
			"error": "error parsing image ID",
		})
	}

	var schemaRequest types.GenerateSchemaRequest
	if err := c.BodyParser(&schemaRequest); err != nil {
		log.Error().Err(err).Msg("Error parsing schema request")
		return c.Status(400).JSON(fiber.Map{
			"error": "error parsing schema request",
		})
	}

	if !schemaRequest.Confirmed {
		log.Error().Msg("Schema request not confirmed")
		return c.Status(400).JSON(fiber.Map{
			"error": "schema request not confirmed",
		})
	}

	// Запускаем создание схемы (асинхронно)
	go func() {
		if err := h.deps.ImageService.GenerateSchema(context.Background(), imageID, schemaRequest.Confirmed); err != nil {
			zerolog.Ctx(context.Background()).Error().
				Err(err).
				Str("image_id", imageID.String()).
				Msg("Failed to generate schema")
		}
	}()

	return c.JSON(types.GenerateSchemaResponse{
		Message:   "Schema generation started",
		ImageID:   imageID,
		EmailSent: true, // Email будет отправлен автоматически после создания схемы
	})
}

// GetImageStatus возвращает текущий статус обработки изображения
// @Summary		Статус обработки изображения
// @Description	Возвращает текущий статус обработки изображения и ссылки на файлы
// @Tags		public-images
// @Produce		json
// @Param		id		path		string					true	"ID изображения"
// @Success		200		{object}	types.ImageStatusResponse		"Статус изображения"
// @Failure		400		{object}	map[string]string			"Ошибка валидации"
// @Failure		404		{object}	map[string]string			"Изображение не найдено"
// @Failure		500		{object}	map[string]string			"Внутренняя ошибка сервера"
// @Router		/public/images/{id}/status [get]
func (h *ImageHandler) GetImageStatus(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	imageIDStr := c.Params("id")
	imageID, err := uuid.Parse(imageIDStr)
	if err != nil {
		log.Error().Err(err).Msg("Invalid image ID format")
		return c.Status(400).JSON(fiber.Map{
			"error": "error parsing image ID",
		})
	}

	status, err := h.deps.ImageService.GetImageStatus(context.Background(), imageID)
	if err != nil {
		log.Error().Err(err).Msg("Image not found")
		return c.Status(404).JSON(fiber.Map{
			"error": "image not found",
		})
	}

	return c.JSON(status)
}

// GetQueue возвращает все задачи в очереди
// @Summary		Список задач в очереди
// @Description	Возвращает все задачи в очереди обработки изображений с возможностью фильтрации по статусу
// @Tags		admin-image-processing
// @Produce		json
// @Param		status		query		string					false	"Статус задачи (uploaded, edited, processing, processed, completed, failed)"
// @Success		200		{array}		map[string]interface{}			"Список задач в очереди"
// @Failure		500		{object}	map[string]string			"Внутренняя ошибка сервера"
// @Router		/admin/queue [get]
func (h *ImageHandler) GetQueue(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	status := c.Query("status")

	tasks, err := h.deps.ImageService.GetQueue(status)
	if err != nil {
		log.Error().Err(err).Msg("Error getting queue")
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(tasks)
}

// GetTaskByID возвращает задачу по ID
// @Summary		Получение задачи по ID
// @Description	Возвращает детальную информацию о задаче обработки изображения
// @Tags		admin-image-processing
// @Produce		json
// @Param		id		path		string					true	"ID задачи"
// @Success		200		{object}	map[string]interface{}			"Информация о задаче"
// @Failure		400		{object}	map[string]string			"Неверный ID"
// @Failure		404		{object}	map[string]string			"Задача не найдена"
// @Failure		500		{object}	map[string]string			"Внутренняя ошибка сервера"
// @Router		/admin/queue/{id} [get]
func (h *ImageHandler) GetTaskByID(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		log.Error().Err(err).Msg("Invalid ID format")
		return c.Status(400).JSON(fiber.Map{
			"error": "error parsing ID",
		})
	}

	task, err := h.deps.ImageRepository.GetByID(context.Background(), id)
	if err != nil {
		log.Error().Err(err).Msg("Task not found")
		return c.Status(404).JSON(fiber.Map{
			"error": "task not found",
		})
	}

	return c.JSON(task)
}

// AddToQueue добавляет задачу в очередь обработки
// @Summary		Добавление задачи в очередь
// @Description	Добавляет новую задачу обработки изображения в очередь
// @Tags		admin-image-processing
// @Accept		json
// @Produce		json
// @Param		task		body		AddToQueueRequest		true	"Параметры задачи"
// @Success		201		{object}	map[string]interface{}		"Задача добавлена в очередь"
// @Failure		400		{object}	map[string]string		"Ошибка валидации"
// @Failure		500		{object}	map[string]string		"Внутренняя ошибка сервера"
// @Router		/admin/queue [post]
func (h *ImageHandler) AddToQueue(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	var req AddToQueueRequest
	if err := c.BodyParser(&req); err != nil {
		log.Error().Err(err).Msg("Error parsing request")
		return c.Status(400).JSON(fiber.Map{
			"error": "error parsing request",
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

	if err := h.deps.ImageRepository.Create(context.Background(), task); err != nil {
		log.Error().Err(err).Msg("Error adding task to queue")
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "Task added to queue",
		"task_id": task.ID,
	})
}

// StartProcessing начинает обработку задачи
// @Summary		Начало обработки задачи
// @Description	Помечает задачу как находящуюся в обработке
// @Tags		admin-image-processing
// @Produce		json
// @Param		id		path		string					true	"ID задачи"
// @Success		200		{object}	map[string]interface{}		"Обработка начата"
// @Failure		400		{object}	map[string]string		"Неверный ID"
// @Failure		500		{object}	map[string]string		"Внутренняя ошибка сервера"
// @Router		/admin/queue/{id}/start [put]
func (h *ImageHandler) StartProcessing(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)

	if err != nil {
		log.Error().Err(err).Msg("Invalid ID format")
		return c.Status(400).JSON(fiber.Map{
			"error": "error parsing ID",
		})
	}

	if err := h.deps.ImageRepository.StartProcessing(context.Background(), id); err != nil {
		log.Error().Err(err).Msg("Error starting processing")
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Processing started",
	})
}

// CompleteProcessing завершает обработку задачи
// @Summary		Завершение обработки задачи
// @Description	Помечает задачу как успешно завершённую
// @Tags		admin-image-processing
// @Produce		json
// @Param		id		path		string					true	"ID задачи"
// @Success		200		{object}	map[string]interface{}		"Задача завершена"
// @Failure		400		{object}	map[string]string		"Неверный ID"
// @Failure		500		{object}	map[string]string		"Внутренняя ошибка сервера"
// @Router		/admin/queue/{id}/complete [put]
func (h *ImageHandler) CompleteProcessing(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)

	if err != nil {
		log.Error().Err(err).Msg("Invalid ID format")
		return c.Status(400).JSON(fiber.Map{
			"error": "error parsing ID",
		})
	}

	if err := h.deps.ImageRepository.CompleteProcessing(context.Background(), id); err != nil {
		log.Error().Err(err).Msg("Error completing processing")
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Processing completed",
	})
}

// FailProcessing помечает задачу как неудачную
// @Summary		Провал обработки задачи
// @Description	Помечает задачу как неудачную с указанием причины
// @Tags		admin-image-processing
// @Accept		json
// @Produce		json
// @Param		id		path		string					true	"ID задачи"
// @Param		error		body		FailProcessingRequest			true	"Сообщение об ошибке"
// @Success		200		{object}	map[string]interface{}			"Задача помечена как неудачная"
// @Failure		400		{object}	map[string]string			"Ошибка валидации"
// @Failure		500		{object}	map[string]string			"Внутренняя ошибка сервера"
// @Router		/admin/queue/{id}/fail [put]
func (h *ImageHandler) FailProcessing(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		log.Error().Err(err).Msg("Invalid ID format")
		return c.Status(400).JSON(fiber.Map{
			"error": "error parsing ID",
		})
	}

	var req FailProcessingRequest
	if err := c.BodyParser(&req); err != nil {
		log.Error().Err(err).Msg("Error parsing request")
		return c.Status(400).JSON(fiber.Map{
			"error": "error parsing request",
		})
	}

	if err := h.deps.ImageRepository.FailProcessing(context.Background(), id, req.ErrorMessage); err != nil {
		log.Error().Err(err).Msg("Error failing processing")
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Processing failed",
	})
}

// RetryTask повторяет обработку задачи
// @Summary		Повторная попытка обработки
// @Description	Возвращает неудачную задачу в очередь для повторной обработки
// @Tags		admin-image-processing
// @Produce		json
// @Param		id		path		string					true	"ID задачи"
// @Success		200		{object}	map[string]interface{}		"Задача возвращена в очередь"
// @Failure		400		{object}	map[string]string		"Неверный ID"
// @Failure		500		{object}	map[string]string		"Внутренняя ошибка сервера"
// @Router		/admin/queue/{id}/retry [put]
func (h *ImageHandler) RetryTask(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		log.Error().Err(err).Msg("Invalid ID format")
		return c.Status(400).JSON(fiber.Map{
			"error": "error parsing ID",
		})
	}

	if err := h.deps.ImageRepository.RetryTask(context.Background(), id); err != nil {
		log.Error().Err(err).Msg("Error retrying task")
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Task retried",
	})
}

// DeleteTask удаляет задачу из очереди
// @Summary		Удаление задачи
// @Description	Удаляет задачу из очереди обработки
// @Tags		admin-image-processing
// @Produce		json
// @Param		id		path		string					true	"ID задачи"
// @Success		200		{object}	map[string]interface{}		"Задача удалена"
// @Failure		400		{object}	map[string]string		"Неверный ID"
// @Failure		500		{object}	map[string]string		"Внутренняя ошибка сервера"
// @Router		/admin/queue/{id} [delete]
func (h *ImageHandler) DeleteTask(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		log.Error().Err(err).Msg("Invalid ID format")
		return c.Status(400).JSON(fiber.Map{
			"error": "error parsing ID",
		})
	}

	if err := h.deps.ImageRepository.Delete(context.Background(), id); err != nil {
		log.Error().Err(err).Msg("Error deleting task")
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Task deleted",
	})
}

// GetStatistics возвращает статистику по задачам
// @Summary		Статистика обработки
// @Description	Возвращает статистику по обработке изображений
// @Tags		admin-image-processing
// @Produce		json
// @Success		200		{object}	map[string]interface{}		"Статистика"
// @Failure		500		{object}	map[string]string		"Внутренняя ошибка сервера"
// @Router		/admin/statistics [get]
func (h *ImageHandler) GetStatistics(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	stats, err := h.deps.ImageRepository.GetStatistics(context.Background())
	if err != nil {
		log.Error().Err(err).Msg("Error getting statistics")
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(stats)
}

// GetNextTask возвращает следующую задачу для обработки
// @Summary		Следующая задача для обработки
// @Description	Возвращает следующую задачу в очереди для обработки
// @Tags		admin-image-processing
// @Produce		json
// @Success		200		{object}	map[string]interface{}		"Следующая задача"
// @Failure		404		{object}	map[string]string		"Нет задач в очереди"
// @Failure		500		{object}	map[string]string		"Внутренняя ошибка сервера"
// @Router		/admin/next [get]
func (h *ImageHandler) GetNextTask(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	task, err := h.deps.ImageRepository.GetNextInQueue(context.Background())
	if err != nil {
		if err.Error() == "no tasks in queue" {
			return c.Status(404).JSON(fiber.Map{
				"message": "No tasks in queue",
			})
		}
		log.Error().Err(err).Msg("Error getting next task")
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(task)
}
