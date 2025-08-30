package preview

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/skr1ms/mosaic/pkg/errors"
	"github.com/skr1ms/mosaic/pkg/middleware"
)

type PreviewHandler struct {
	deps *PreviewHandlerDependencies
}

type PreviewHandlerDependencies struct {
	PreviewService PreviewServiceInterface
	Logger         *middleware.Logger
}

func NewPreviewHandler(api fiber.Router, deps *PreviewHandlerDependencies) {
	handler := &PreviewHandler{
		deps: deps,
	}

	// Регистрируем роуты
	previewGroup := api.Group("/preview")
	previewGroup.Post("/generate", handler.GeneratePreview)
}

// @Summary Generate mosaic preview
// @Description Generates a preview of how the uploaded image will look as a mosaic
// @Tags preview
// @Accept multipart/form-data
// @Produce json
// @Param image formData file true "Image file (JPG, PNG)"
// @Param size formData string true "Mosaic size (21x30, 30x40, 40x40, 40x50, 40x60, 50x70)"
// @Param style formData string true "Mosaic style (grayscale, skin_tones, pop_art, max_colors)"
// @Success 200 {object} map[string]any "Preview generated successfully"
// @Failure 400 {object} map[string]any "Bad request: missing required fields or invalid data"
// @Failure 413 {object} map[string]any "File too large"
// @Failure 500 {object} map[string]any "Internal server error during preview generation"
// @Router /api/preview/generate [post]
func (h *PreviewHandler) GeneratePreview(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Get form values
	size := c.FormValue("size")
	style := c.FormValue("style")

	if size == "" || style == "" {
		h.deps.Logger.FromContext(c).Warn().
			Str("handler", "GeneratePreview").
			Str("size", size).
			Str("style", style).
			Msg("Size and style are required")

		return errors.SendError(c, errors.ValidationError("Size and style are required"))
	}

	// Validate size
	validSizes := []string{"21x30", "30x40", "40x40", "40x50", "40x60", "50x70"}
	isValidSize := false
	for _, validSize := range validSizes {
		if size == validSize {
			isValidSize = true
			break
		}
	}
	if !isValidSize {
		h.deps.Logger.FromContext(c).Warn().
			Str("handler", "GeneratePreview").
			Str("size", size).
			Msg("Invalid size")

		return errors.SendError(c, errors.ValidationError("Invalid size. Valid sizes: 21x30, 30x40, 40x40, 40x50, 40x60, 50x70"))
	}

	// Validate style
	validStyles := []string{"grayscale", "skin_tones", "pop_art", "max_colors"}
	isValidStyle := false
	for _, validStyle := range validStyles {
		if style == validStyle {
			isValidStyle = true
			break
		}
	}
	if !isValidStyle {
		h.deps.Logger.FromContext(c).Warn().
			Str("handler", "GeneratePreview").
			Str("style", style).
			Msg("Invalid style")

		return errors.SendError(c, errors.ValidationError("Invalid style. Valid styles: grayscale, skin_tones, pop_art, max_colors"))
	}

	// Get uploaded file
	file, err := c.FormFile("image")
	if err != nil {
		h.deps.Logger.FromContext(c).Warn().
			Err(err).
			Str("handler", "GeneratePreview").
			Msg("Image file is required")

		return errors.SendError(c, errors.ValidationError("Image file is required"))
	}

	// Validate file type
	if file.Header.Get("Content-Type") != "image/jpeg" &&
		file.Header.Get("Content-Type") != "image/jpg" &&
		file.Header.Get("Content-Type") != "image/png" {
		h.deps.Logger.FromContext(c).Warn().
			Str("handler", "GeneratePreview").
			Str("content_type", file.Header.Get("Content-Type")).
			Msg("Invalid file type")

		return errors.SendError(c, errors.ValidationError("Invalid file type. Only JPG and PNG are supported"))
	}

	// Validate file size (max 15MB)
	if file.Size > 15<<20 {
		h.deps.Logger.FromContext(c).Warn().
			Str("handler", "GeneratePreview").
			Int64("file_size", file.Size).
			Msg("File too large")

		return errors.SendError(c, errors.ValidationError("File too large. Maximum size is 15MB"))
	}

	h.deps.Logger.FromContext(c).Info().
		Str("handler", "GeneratePreview").
		Str("size", size).
		Str("style", style).
		Str("filename", file.Filename).
		Int64("file_size", file.Size).
		Msg("Starting preview generation")

	// Generate preview
	result, err := h.deps.PreviewService.GeneratePreview(ctx, file, size, style)
	if err != nil {
		h.deps.Logger.FromContext(c).Error().
			Err(err).
			Str("handler", "GeneratePreview").
			Str("size", size).
			Str("style", style).
			Msg("Failed to generate preview")

		return errors.SendError(c, errors.InternalServerError("Failed to generate preview"))
	}

	h.deps.Logger.FromContext(c).Info().
		Str("handler", "GeneratePreview").
		Str("preview_url", result["preview_url"].(string)).
		Msg("Preview generated successfully")

	return c.JSON(result)
}
