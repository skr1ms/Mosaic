package preview

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/skr1ms/mosaic/pkg/middleware"
)

type PreviewHandlerDeps struct {
	PreviewService    PreviewServiceInterface
	PreviewRepository PreviewRepositoryInterface
	Logger            *middleware.Logger
}

type PreviewHandler struct {
	fiber.Router
	deps *PreviewHandlerDeps
}

func NewPreviewHandler(app fiber.Router, deps *PreviewHandlerDeps) *PreviewHandler {
	handler := &PreviewHandler{
		Router: app,
		deps:   deps,
	}

	// Public preview routes
	public := handler.Group("/public")
	public.Post("/generate", handler.GeneratePreview)
	public.Post("/generate-variants", handler.GenerateVariants)
	public.Post("/generate-ai", handler.GenerateAIPreview)
	public.Get("/:id", handler.GetPreview)
	public.Delete("/:id", handler.DeletePreview)

	return handler
}

// GeneratePreview generates a preview with specified style
func (h *PreviewHandler) GeneratePreview(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Parse multipart form
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid form data",
		})
	}

	// Get image file
	files := form.File["image"]
	if len(files) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Image file is required",
		})
	}

	file := files[0]
	
	// Get parameters
	size := c.FormValue("size", "30x40")
	style := c.FormValue("style", "natural")
	contrastLevel := c.FormValue("contrast_level", "normal")

	// Generate preview ID
	previewID := uuid.New().String()

	// Process image
	previewData, err := h.deps.PreviewService.GeneratePreview(ctx, file, size, style, contrastLevel)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate preview")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate preview",
		})
	}

	return c.JSON(fiber.Map{
		"preview_id":  previewID,
		"preview_url": previewData.URL,
		"style":       style,
		"size":        size,
		"contrast":    contrastLevel,
	})
}

// GenerateVariants generates multiple style variants
func (h *PreviewHandler) GenerateVariants(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Parse multipart form
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid form data",
		})
	}

	// Get image file
	files := form.File["image"]
	if len(files) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Image file is required",
		})
	}

	file := files[0]
	size := c.FormValue("size", "30x40")

	// Define style variants
	variants := []struct {
		Style    string `json:"style"`
		Contrast string `json:"contrast"`
		Label    string `json:"label"`
	}{
		{"venus", "soft", "Венера (мягкий контраст)"},
		{"venus", "strong", "Венера (сильный контраст)"},
		{"sun", "soft", "Солнце (мягкий контраст)"},
		{"sun", "strong", "Солнце (сильный контраст)"},
		{"moon", "soft", "Луна (мягкий контраст)"},
		{"moon", "strong", "Луна (сильный контраст)"},
		{"mars", "soft", "Марс (мягкий контраст)"},
		{"mars", "strong", "Марс (сильный контраст)"},
	}

	var previews []fiber.Map

	for _, variant := range variants {
		previewData, err := h.deps.PreviewService.GeneratePreview(ctx, file, size, variant.Style, variant.Contrast)
		if err != nil {
			log.Error().Err(err).Str("style", variant.Style).Msg("Failed to generate variant")
			continue
		}

		previews = append(previews, fiber.Map{
			"style":       variant.Style,
			"contrast":    variant.Contrast,
			"label":       variant.Label,
			"preview_url": previewData.URL,
		})
	}

	return c.JSON(fiber.Map{
		"previews": previews,
		"total":    len(previews),
	})
}

// GenerateAIPreview generates AI-enhanced previews using Stable Diffusion
func (h *PreviewHandler) GenerateAIPreview(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Parse multipart form
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid form data",
		})
	}

	// Get image file
	files := form.File["image"]
	if len(files) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Image file is required",
		})
	}

	file := files[0]
	
	// Get parameters
	prompt := c.FormValue("prompt", "artistic mosaic style, high quality, detailed")
	numVariants := 2 // Generate 2 AI variants by default

	var aiPreviews []fiber.Map

	for i := 0; i < numVariants; i++ {
		// Generate AI preview
		aiPreviewData, err := h.deps.PreviewService.GenerateAIPreview(ctx, file, prompt)
		if err != nil {
			log.Error().Err(err).Int("variant", i+1).Msg("Failed to generate AI preview")
			continue
		}

		aiPreviews = append(aiPreviews, fiber.Map{
			"preview_url": aiPreviewData.URL,
			"prompt":      prompt,
			"variant":     i + 1,
		})
	}

	return c.JSON(fiber.Map{
		"ai_previews": aiPreviews,
		"total":       len(aiPreviews),
	})
}

// GetPreview retrieves a preview by ID
func (h *PreviewHandler) GetPreview(c *fiber.Ctx) error {
	previewID := c.Params("id")
	if previewID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Preview ID is required",
		})
	}

	preview, err := h.deps.PreviewRepository.GetByID(context.Background(), previewID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Preview not found",
		})
	}

	return c.JSON(preview)
}

// DeletePreview deletes a preview by ID
func (h *PreviewHandler) DeletePreview(c *fiber.Ctx) error {
	previewID := c.Params("id")
	if previewID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Preview ID is required",
		})
	}

	err := h.deps.PreviewRepository.Delete(context.Background(), previewID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to delete preview: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Preview deleted successfully",
	})
}