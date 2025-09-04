package preview

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"

	"github.com/disintegration/imaging"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type PreviewServiceInterface interface {
	GeneratePreview(ctx context.Context, file *multipart.FileHeader, size, style, contrast string) (*PreviewData, error)
	GenerateAIPreview(ctx context.Context, file *multipart.FileHeader, prompt string) (*PreviewData, error)
	ApplyStyle(img image.Image, style string) image.Image
	ApplyContrast(img image.Image, level string) image.Image
}

type PreviewService struct {
	repository PreviewRepositoryInterface
	s3Client   interface{} // Add your S3 client interface
	aiClient   interface{} // Add your Stable Diffusion client interface
}

type PreviewData struct {
	ID       string
	URL      string
	Style    string
	Contrast string
	Size     string
}

func NewPreviewService(repo PreviewRepositoryInterface, s3 interface{}, ai interface{}) *PreviewService {
	return &PreviewService{
		repository: repo,
		s3Client:   s3,
		aiClient:   ai,
	}
}

func (s *PreviewService) GeneratePreview(ctx context.Context, file *multipart.FileHeader, size, style, contrast string) (*PreviewData, error) {
	// Open uploaded file
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	// Decode image
	img, format, err := image.Decode(src)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Apply size transformation
	img = s.resizeImage(img, size)

	// Apply style
	img = s.ApplyStyle(img, style)

	// Apply contrast
	img = s.ApplyContrast(img, contrast)

	// Convert to bytes
	var buf bytes.Buffer
	switch format {
	case "jpeg", "jpg":
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90})
	case "png":
		err = png.Encode(&buf, img)
	default:
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to encode image: %w", err)
	}

	// Generate unique ID
	previewID := uuid.New().String()

	// Upload to S3 (simplified - implement according to your S3 client)
	previewURL := fmt.Sprintf("/previews/%s.jpg", previewID)

	// Save preview data
	previewData := &PreviewData{
		ID:       previewID,
		URL:      previewURL,
		Style:    style,
		Contrast: contrast,
		Size:     size,
	}

	// Store in repository
	if err := s.repository.Create(ctx, previewData); err != nil {
		return nil, fmt.Errorf("failed to save preview: %w", err)
	}

	return previewData, nil
}

func (s *PreviewService) GenerateAIPreview(ctx context.Context, file *multipart.FileHeader, prompt string) (*PreviewData, error) {
	// Open uploaded file
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	// Read file content
	fileContent, err := io.ReadAll(src)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Call AI service (simplified - implement according to your AI client)
	// This would typically call Stable Diffusion API
	log.Info().Str("prompt", prompt).Msg("Generating AI preview")

	// For now, return a placeholder
	previewID := uuid.New().String()
	previewURL := fmt.Sprintf("/ai-previews/%s.jpg", previewID)

	previewData := &PreviewData{
		ID:    previewID,
		URL:   previewURL,
		Style: "ai-generated",
	}

	// Store in repository
	if err := s.repository.Create(ctx, previewData); err != nil {
		return nil, fmt.Errorf("failed to save AI preview: %w", err)
	}

	return previewData, nil
}

func (s *PreviewService) ApplyStyle(img image.Image, style string) image.Image {
	switch style {
	case "venus":
		// Apply warm tones
		return s.applyColorFilter(img, color.RGBA{255, 200, 150, 50})
	case "sun":
		// Apply bright yellow tones
		return s.applyColorFilter(img, color.RGBA{255, 255, 100, 30})
	case "moon":
		// Apply cool blue tones
		return s.applyColorFilter(img, color.RGBA{150, 150, 255, 40})
	case "mars":
		// Apply red tones
		return s.applyColorFilter(img, color.RGBA{255, 100, 100, 50})
	case "vintage":
		// Apply sepia effect
		return imaging.Sepia(img, 30)
	case "monochrome":
		// Convert to grayscale
		return imaging.Grayscale(img)
	case "enhanced":
		// Increase saturation and contrast
		img = imaging.AdjustSaturation(img, 30)
		return imaging.AdjustContrast(img, 20)
	default:
		return img
	}
}

func (s *PreviewService) ApplyContrast(img image.Image, level string) image.Image {
	switch level {
	case "soft":
		return imaging.AdjustContrast(img, -10)
	case "strong":
		return imaging.AdjustContrast(img, 30)
	case "normal":
		return img
	default:
		return img
	}
}

func (s *PreviewService) applyColorFilter(img image.Image, filterColor color.RGBA) image.Image {
	bounds := img.Bounds()
	filtered := image.NewRGBA(bounds)
	
	// Create overlay with filter color
	overlay := image.NewUniform(filterColor)
	
	// Draw original image
	draw.Draw(filtered, bounds, img, bounds.Min, draw.Src)
	
	// Apply color filter with transparency
	draw.DrawMask(filtered, bounds, overlay, bounds.Min, nil, bounds.Min, draw.Over)
	
	return filtered
}

func (s *PreviewService) resizeImage(img image.Image, size string) image.Image {
	var width, height int

	switch size {
	case "21x30":
		width, height = 210, 300
	case "30x40":
		width, height = 300, 400
	case "40x40":
		width, height = 400, 400
	case "40x50":
		width, height = 400, 500
	case "40x60":
		width, height = 400, 600
	case "50x70":
		width, height = 500, 700
	default:
		width, height = 300, 400
	}

	// Resize maintaining aspect ratio
	return imaging.Fit(img, width, height, imaging.Lanczos)
}