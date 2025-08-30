package preview

import (
	"github.com/gofiber/fiber/v2"
	"github.com/skr1ms/mosaic/pkg/middleware"
)

type PreviewHandlerDeps struct {
	PreviewService *PreviewService
	Logger         *middleware.Logger
}

type PreviewHandler struct {
	deps *PreviewHandlerDeps
}

func NewPreviewHandler(deps *PreviewHandlerDeps) *PreviewHandler {
	return &PreviewHandler{
		deps: deps,
	}
}

// GeneratePreview генерирует превью мозаики
func (h *PreviewHandler) GeneratePreview(c *fiber.Ctx) error {
	var req PreviewRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	// Валидация входных данных
	if req.Size == "" || req.Style == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Size and style are required",
		})
	}

	// Генерируем превью
	response, err := h.deps.PreviewService.GeneratePreview(c.Context(), &req)
	if err != nil {
		h.deps.Logger.GetZerologLogger().Error().Err(err).Msg("Failed to generate preview")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to generate preview",
			"details": err.Error(),
		})
	}

	// Отправляем ответ
	return c.Status(fiber.StatusOK).JSON(response)
}
