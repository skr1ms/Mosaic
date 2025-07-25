package image_processing

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Handler struct {
	fiber.Router
	repo *ImageProcessingRepository
}

func NewImageProcessingHandler(router fiber.Router, db *gorm.DB) {
	handler := &Handler{
		Router: router,
		repo:   NewRepository(db),
	}

	api := handler.Group("/image-processing")
	api.Get("/queue", handler.GetQueue)
	api.Get("/queue/:id", handler.GetTaskByID)
	api.Post("/queue", handler.AddToQueue)
	api.Put("/queue/:id/start", handler.StartProcessing)
	api.Put("/queue/:id/complete", handler.CompleteProcessing)
	api.Put("/queue/:id/fail", handler.FailProcessing)
	api.Put("/queue/:id/retry", handler.RetryTask)
	api.Delete("/queue/:id", handler.DeleteTask)
	api.Get("/statistics", handler.GetStatistics)
	api.Get("/next", handler.GetNextTask)
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
func (h *Handler) GetQueue(c *fiber.Ctx) error {
	status := c.Query("status")
	
	var tasks []*ImageProcessingQueue
	var err error
	
	if status != "" {
		tasks, err = h.repo.GetByStatus(status)
	} else {
		tasks, err = h.repo.GetAll()
	}
	
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch queue"})
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
func (h *Handler) GetTaskByID(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid task ID"})
	}

	task, err := h.repo.GetByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Task not found"})
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
func (h *Handler) AddToQueue(c *fiber.Ctx) error {
	var req AddToQueueRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	task := &ImageProcessingQueue{
		CouponID:          req.CouponID,
		OriginalImagePath: req.OriginalImagePath,
		ProcessingParams:  req.ProcessingParams,
		UserEmail:         req.UserEmail,
		Priority:          req.Priority,
		Status:            "queued",
		RetryCount:        0,
		MaxRetries:        3,
	}

	if err := h.repo.Create(task); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to add task to queue"})
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
func (h *Handler) GetNextTask(c *fiber.Ctx) error {
	task, err := h.repo.GetNextInQueue()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "No tasks in queue"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get next task"})
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
func (h *Handler) StartProcessing(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid task ID"})
	}

	if err := h.repo.StartProcessing(id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to start processing"})
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
func (h *Handler) CompleteProcessing(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid task ID"})
	}

	if err := h.repo.CompleteProcessing(id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to complete processing"})
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
func (h *Handler) FailProcessing(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid task ID"})
	}

	var req FailProcessingRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if err := h.repo.FailProcessing(id, req.ErrorMessage); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to mark as failed"})
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
func (h *Handler) RetryTask(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid task ID"})
	}

	if err := h.repo.RetryTask(id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retry task"})
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
func (h *Handler) DeleteTask(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid task ID"})
	}

	if err := h.repo.Delete(id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete task"})
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
func (h *Handler) GetStatistics(c *fiber.Ctx) error {
	stats, err := h.repo.GetStatistics()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get statistics"})
	}

	return c.JSON(stats)
} 