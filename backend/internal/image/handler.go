package image

import (
	"context"
	"database/sql"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/skr1ms/mosaic/internal/coupon"
)

type ImageHandlerDeps struct {
	CouponRepository *coupon.CouponRepository
	ImageService     *ImageService
	ImageRepository  *ImageRepository
	Logger           *zerolog.Logger
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

	api := handler.Group("/image-processing")
	api.Get("/queue", handler.GetQueue)                        // Получение всех задач в очереди
	api.Get("/queue/:id", handler.GetTaskByID)                 // Получение задачи по ID
	api.Post("/queue", handler.AddToQueue)                     // Добавление задачи в очередь
	api.Put("/queue/:id/start", handler.StartProcessing)       // Начало обработки задачи
	api.Put("/queue/:id/complete", handler.CompleteProcessing) // Завершение обработки задачи
	api.Put("/queue/:id/fail", handler.FailProcessing)         // Провал обработки задачи
	api.Put("/queue/:id/retry", handler.RetryTask)             // Повторная попытка обработки задачи
	api.Delete("/queue/:id", handler.DeleteTask)               // Удаление задачи из очереди
	api.Get("/statistics", handler.GetStatistics)              // Получение статистики по задачам
	api.Get("/next", handler.GetNextTask)                      // Получение следующей задачи для обработки
}

// GetQueue возвращает все задачи в очереди
// @Summary Список задач в очереди
// @Description Возвращает все задачи в очереди обработки изображений с возможностью фильтрации по статусу
// @Tags image-processing
// @Produce json
// @Param status query string false "Статус задачи (queued, processing, completed, failed)"
// @Success 200 {array} map[string]interface{} "Список задач в очереди"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /image-processing/queue [get]
func (handler *ImageHandler) GetQueue(c *fiber.Ctx) error {
	status := c.Query("status")

	tasks, err := handler.deps.ImageService.GetQueue(status)
	if err != nil {
		handler.deps.Logger.Error().Err(err).Msg(ErrFailedToFetchQueue.Message)
		return c.Status(ErrFailedToFetchQueue.HTTPStatus).JSON(fiber.Map{"error": ErrFailedToFetchQueue.Error()})
	}

	return c.JSON(tasks)
}

// GetTaskByID возвращает задачу по ID
// @Summary Получение задачи по ID
// @Description Возвращает детальную информацию о задаче обработки изображения по ID
// @Tags image-processing
// @Produce json
// @Param id path string true "ID задачи"
// @Success 200 {object} map[string]interface{} "Информация о задаче"
// @Failure 400 {object} map[string]interface{} "Неверный ID задачи"
// @Failure 404 {object} map[string]interface{} "Задача не найдена"
// @Router /image-processing/queue/{id} [get]
func (handler *ImageHandler) GetTaskByID(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		handler.deps.Logger.Error().Err(err).Msg(ErrInvalidTaskID.Error())
		return c.Status(ErrInvalidTaskID.HTTPStatus).JSON(fiber.Map{"error": ErrInvalidTaskID.Error()})
	}

	// Получаем задачу по ID
	task, err := handler.deps.ImageRepository.GetByID(context.Background(), id)
	if err != nil {
		handler.deps.Logger.Error().Err(err).Msg(ErrTaskNotFound.Error())
		return c.Status(ErrTaskNotFound.HTTPStatus).JSON(fiber.Map{"error": ErrTaskNotFound.Error()})
	}

	return c.JSON(task)
}

// AddToQueue добавляет новую задачу в очередь обработки
// @Summary Добавление задачи в очередь
// @Description Добавляет новую задачу обработки изображения в очередь
// @Tags image-processing
// @Accept json
// @Produce json
// @Param request body AddToQueueRequest true "Параметры задачи обработки"
// @Success 201 {object} map[string]interface{} "Задача добавлена в очередь"
// @Failure 400 {object} map[string]interface{} "Ошибка в запросе"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /image-processing/queue [post]
func (handler *ImageHandler) AddToQueue(c *fiber.Ctx) error {
	var req AddToQueueRequest

	if err := c.BodyParser(&req); err != nil {
		handler.deps.Logger.Error().Err(err).Msg(ErrBadRequest.Error())
		return c.Status(ErrBadRequest.HTTPStatus).JSON(fiber.Map{"error": ErrBadRequest.Error()})
	}

	// Добавить логику для добавления задачи в очередь
	if err := handler.deps.ImageService.AddToQueue(req.CouponID); err != nil {
		handler.deps.Logger.Error().Err(err).Msg(ErrFailedToAddToQueue.Error())
		return c.Status(ErrFailedToAddToQueue.HTTPStatus).JSON(fiber.Map{"error": ErrFailedToAddToQueue.Error()})
	}

	task := &Image{
		CouponID:          req.CouponID,
		OriginalImagePath: req.OriginalImagePath,
		ProcessingParams:  req.ProcessingParams,
		UserEmail:         req.UserEmail,
		Priority:          req.Priority,
		Status:            "queued",
		RetryCount:        0,
		MaxRetries:        3,
	}

	// Добавить логику для создания задачи
	if err := handler.deps.ImageRepository.Create(context.Background(), task); err != nil {
		handler.deps.Logger.Error().Err(err).Msg(ErrFailedToCreateTask.Error())
		return c.Status(ErrFailedToCreateTask.HTTPStatus).JSON(fiber.Map{"error": ErrFailedToCreateTask.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(task)
}

// GetNextTask возвращает следующую задачу для обработки
// @Summary Получение следующей задачи
// @Description Возвращает следующую задачу в очереди для обработки (приоритетный порядок)
// @Tags image-processing
// @Produce json
// @Success 200 {object} map[string]interface{} "Следующая задача для обработки"
// @Failure 404 {object} map[string]interface{} "Нет задач в очереди"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /image-processing/next [get]
func (handler *ImageHandler) GetNextTask(c *fiber.Ctx) error {
	// Получаем следующую задачу в очереди
	task, err := handler.deps.ImageRepository.GetNextInQueue(context.Background())
	if err != nil {
		if err == sql.ErrNoRows {
			handler.deps.Logger.Info().Msg("No tasks in queue")
			return c.Status(ErrNoTasksInQueue.HTTPStatus).JSON(fiber.Map{"message": ErrNoTasksInQueue.Error()})
		}
		handler.deps.Logger.Error().Err(err).Msg(ErrFailedToGetNextTask.Error())
		return c.Status(ErrFailedToGetNextTask.HTTPStatus).JSON(fiber.Map{"error": ErrFailedToGetNextTask.Error()})
	}

	return c.JSON(task)
}

// StartProcessing помечает задачу как обрабатываемую
// @Summary Начать обработку задачи
// @Description Помечает задачу как находящуюся в процессе обработки
// @Tags image-processing
// @Produce json
// @Param id path string true "ID задачи"
// @Success 200 {object} map[string]interface{} "Обработка началась"
// @Failure 400 {object} map[string]interface{} "Неверный ID задачи"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /image-processing/queue/{id}/start [put]
func (handler *ImageHandler) StartProcessing(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		handler.deps.Logger.Error().Err(err).Msg(ErrInvalidTaskID.Error())
		return c.Status(ErrInvalidTaskID.HTTPStatus).JSON(fiber.Map{"error": ErrInvalidTaskID.Error()})
	}

	// Начинаем обработку задачи: обновляем статус и время начала
	if err := handler.deps.ImageRepository.StartProcessing(context.Background(), id); err != nil {
		handler.deps.Logger.Error().Err(err).Msg(ErrFailedToStartProcessing.Error())
		return c.Status(ErrFailedToStartProcessing.HTTPStatus).JSON(fiber.Map{"error": ErrFailedToStartProcessing.Error()})
	}

	return c.JSON(fiber.Map{"message": "Processing started"})
}

// CompleteProcessing помечает задачу как завершенную
// @Summary Завершить обработку задачи
// @Description Помечает задачу как успешно завершенную
// @Tags image-processing
// @Produce json
// @Param id path string true "ID задачи"
// @Success 200 {object} map[string]interface{} "Обработка завершена"
// @Failure 400 {object} map[string]interface{} "Неверный ID задачи"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /image-processing/queue/{id}/complete [put]
func (handler *ImageHandler) CompleteProcessing(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		handler.deps.Logger.Error().Err(err).Msg(ErrInvalidTaskID.Error())
		return c.Status(ErrInvalidTaskID.HTTPStatus).JSON(fiber.Map{"error": ErrInvalidTaskID.Error()})
	}

	// Завершаем обработку задачи: обновляем статус и время завершения
	if err := handler.deps.ImageRepository.CompleteProcessing(context.Background(), id); err != nil {
		handler.deps.Logger.Error().Err(err).Msg(ErrFailedToCompleteProcessing.Error())
		return c.Status(ErrFailedToCompleteProcessing.HTTPStatus).JSON(fiber.Map{"error": ErrFailedToCompleteProcessing.Error()})
	}

	return c.JSON(fiber.Map{"message": "Processing completed"})
}

// FailProcessing помечает задачу как неудачную
// @Summary Пометить задачу как неудачную
// @Description Помечает задачу как неудачную с указанием причины ошибки
// @Tags image-processing
// @Accept json
// @Produce json
// @Param id path string true "ID задачи"
// @Param request body FailProcessingRequest true "Сообщение об ошибке"
// @Success 200 {object} map[string]interface{} "Задача помечена как неудачная"
// @Failure 400 {object} map[string]interface{} "Неверный ID задачи"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /image-processing/queue/{id}/fail [put]
func (handler *ImageHandler) FailProcessing(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		handler.deps.Logger.Error().Err(err).Msg(ErrInvalidTaskID.Error())
		return c.Status(ErrInvalidTaskID.HTTPStatus).JSON(fiber.Map{"error": ErrInvalidTaskID.Error()})
	}

	var req FailProcessingRequest

	if err := c.BodyParser(&req); err != nil {
		handler.deps.Logger.Error().Err(err).Msg(ErrBadRequest.Error())
		return c.Status(ErrBadRequest.HTTPStatus).JSON(fiber.Map{"error": ErrBadRequest.Error()})
	}

	// Помечаем задачу как неудачную с сообщением об ошибке
	if err := handler.deps.ImageRepository.FailProcessing(context.Background(), id, req.ErrorMessage); err != nil {
		handler.deps.Logger.Error().Err(err).Msg(ErrFailedToMarkAsFailed.Error())
		return c.Status(ErrFailedToMarkAsFailed.HTTPStatus).JSON(fiber.Map{"error": ErrFailedToMarkAsFailed.Error()})
	}

	return c.JSON(fiber.Map{"message": "Task marked as failed"})
}

// RetryTask повторяет неудачную задачу
// @Summary Повторить неудачную задачу
// @Description Возвращает неудачную задачу обратно в очередь для повторной обработки
// @Tags image-processing
// @Produce json
// @Param id path string true "ID задачи"
// @Success 200 {object} map[string]interface{} "Задача поставлена на повтор"
// @Failure 400 {object} map[string]interface{} "Неверный ID задачи"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /image-processing/queue/{id}/retry [put]
func (handler *ImageHandler) RetryTask(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		handler.deps.Logger.Error().Err(err).Msg(ErrInvalidTaskID.Error())
		return c.Status(ErrInvalidTaskID.HTTPStatus).JSON(fiber.Map{"error": ErrInvalidTaskID.Error()})
	}

	// Ставим задачу на повторную обработку
	if err := handler.deps.ImageRepository.RetryTask(context.Background(), id); err != nil {
		handler.deps.Logger.Error().Err(err).Msg(ErrFailedToRetryTask.Error())
		return c.Status(ErrFailedToRetryTask.HTTPStatus).JSON(fiber.Map{"error": ErrFailedToRetryTask.Error()})
	}

	return c.JSON(fiber.Map{"message": "Task queued for retry"})
}

// DeleteTask удаляет задачу из очереди
// @Summary Удаление задачи
// @Description Удаляет задачу из очереди обработки изображений
// @Tags image-processing
// @Produce json
// @Param id path string true "ID задачи"
// @Success 200 {object} map[string]interface{} "Задача удалена"
// @Failure 400 {object} map[string]interface{} "Неверный ID задачи"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /image-processing/queue/{id} [delete]
func (handler *ImageHandler) DeleteTask(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		handler.deps.Logger.Error().Err(err).Msg(ErrInvalidTaskID.Error())
		return c.Status(ErrInvalidTaskID.HTTPStatus).JSON(fiber.Map{"error": ErrInvalidTaskID.Error()})
	}

	// Удаляем задачу из базы данных
	if err := handler.deps.ImageRepository.Delete(context.Background(), id); err != nil {
		handler.deps.Logger.Error().Err(err).Msg(ErrFailedToDeleteTask.Error())
		return c.Status(ErrFailedToDeleteTask.HTTPStatus).JSON(fiber.Map{"error": ErrFailedToDeleteTask.Error()})
	}

	return c.JSON(fiber.Map{"message": "Task deleted successfully"})
}

// GetStatistics возвращает статистику по обработке изображений
// @Summary Статистика обработки изображений
// @Description Возвращает статистику по задачам обработки изображений
// @Tags image-processing
// @Produce json
// @Success 200 {object} map[string]interface{} "Статистика обработки"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /image-processing/statistics [get]
func (handler *ImageHandler) GetStatistics(c *fiber.Ctx) error {
	stats, err := handler.deps.ImageRepository.GetStatistics(context.Background())
	if err != nil {
		handler.deps.Logger.Error().Err(err).Msg(ErrFailedToGetStatistics.Error())
		return c.Status(ErrFailedToGetStatistics.HTTPStatus).JSON(fiber.Map{"error": ErrFailedToGetStatistics.Error()})
	}

	return c.JSON(stats)
}
