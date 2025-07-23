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
func (h *Handler) GetStatistics(c *fiber.Ctx) error {
	stats, err := h.repo.GetStatistics()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get statistics"})
	}

	return c.JSON(stats)
} 