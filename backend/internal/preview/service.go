package preview

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/skr1ms/mosaic/pkg/mosaic"
	"github.com/skr1ms/mosaic/pkg/palette"
	"github.com/skr1ms/mosaic/pkg/s3"
)

type PreviewServiceDeps struct {
	S3Client        *s3.S3Client
	MosaicGenerator *mosaic.MosaicGenerator
	PaletteService  *palette.PaletteService
	WorkingDir      string
}

type PreviewService struct {
	deps *PreviewServiceDeps
}

func NewPreviewService(deps *PreviewServiceDeps) *PreviewService {
	return &PreviewService{
		deps: deps,
	}
}

// GeneratePreview creates a mosaic preview from uploaded file
func (s *PreviewService) GeneratePreview(ctx context.Context, file *multipart.FileHeader, size, style string) (map[string]any, error) {
	previewID := uuid.New()

	// Create temporary directory for preview
	tempDir := filepath.Join(s.deps.WorkingDir, "preview", previewID.String())
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Save uploaded file to temporary directory
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Determine file extension
	ext := ".jpg"
	if file.Header.Get("Content-Type") == "image/png" {
		ext = ".png"
	}

	tempImagePath := filepath.Join(tempDir, "input"+ext)
	dst, err := os.Create(tempImagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return nil, fmt.Errorf("failed to copy file: %w", err)
	}

	// Get palette path for selected style
	palettePath, err := s.deps.PaletteService.GetPalettePath(palette.Style(style))
	if err != nil {
		return nil, fmt.Errorf("failed to get palette path: %w", err)
	}

	// Determine mosaic dimensions based on selected size (same logic as ImageService)
	width, height := s.getSizeDimensions(size)
	stonesX := width / 4 // Convert pixels to stones count
	stonesY := height / 4

	// Create request for mosaic generation
	generationReq := &mosaic.GenerationRequest{
		ImagePath:   tempImagePath,
		StonesX:     stonesX,
		StonesY:     stonesY,
		StoneSizeMM: 2.52,   // Standard stone size
		DPI:         150,    // Same as working schema generation
		PreviewDPI:  120,    // Same as working schema generation
		SchemeDPI:   150,    // Same as working schema generation
		Mode:        "both", // Generate both preview and schema (same as ImageService)
		Style:       s.mapStyleToMosaicStyle(style),
		WithLegend:  true, // Same as working schema generation
		Threads:     4,
		PalettePath: palettePath,
	}

	// Generate mosaic (same as working schema generation)
	result, err := s.deps.MosaicGenerator.Generate(ctx, generationReq)
	if err != nil {
		return nil, fmt.Errorf("failed to generate mosaic preview: %w", err)
	}

	// Check that preview was created
	if result.PreviewPath == "" {
		return nil, fmt.Errorf("preview file not generated")
	}

	// Upload preview to S3 (to preview-images bucket)
	previewFile, err := os.Open(result.PreviewPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open preview file: %w", err)
	}
	defer previewFile.Close()

	fileInfo, err := previewFile.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	// Create file key in format: previews/{uuid}.png
	objectKey := fmt.Sprintf("previews/%s.png", previewID.String())

	// Upload to preview-images bucket
	err = s.deps.S3Client.UploadToPreviewBucket(ctx, objectKey, previewFile, fileInfo.Size(), "image/png")
	if err != nil {
		return nil, fmt.Errorf("failed to upload preview to S3: %w", err)
	}

	// Get public URL for preview
	previewURL := s.deps.S3Client.GetPreviewURL(objectKey)
	return map[string]any{
		"preview_id":   previewID.String(),
		"preview_url":  previewURL,
		"size":         size,
		"style":        style,
		"generated_at": time.Now().Format(time.RFC3339),
	}, nil
}

// GetPreviewDownloadURL returns the download URL for a preview by ID
func (s *PreviewService) GetPreviewDownloadURL(previewID string) (string, error) {
	// Extract UUID from filename (assuming format: previews/uuid.png)
	objectKey := fmt.Sprintf("previews/%s.png", previewID)

	// Get URL from S3 client
	downloadURL := s.deps.S3Client.GetPreviewURL(objectKey)
	if downloadURL == "" {
		return "", fmt.Errorf("failed to get preview download URL for key: %s", objectKey)
	}

	return downloadURL, nil
}

// getSizeDimensions returns width and height in pixels for the given size (same as ImageService)
func (s *PreviewService) getSizeDimensions(size string) (int, int) {
	if size == "" {
		size = "30x40" // Default value
	}

	switch size {
	case "21x30":
		return 840, 1200
	case "30x40":
		return 1200, 1600
	case "40x40":
		return 1600, 1600
	case "40x50":
		return 1600, 2000
	case "40x60":
		return 1600, 2400
	case "50x70":
		return 2000, 2800
	default:
		return 1200, 1600
	}
}

// mapStyleToMosaicStyle converts API style to mosaic style
func (s *PreviewService) mapStyleToMosaicStyle(style string) string {
	switch style {
	case "grayscale":
		return "grayscale"
	case "skin_tones":
		return "skin_tones"
	case "pop_art":
		return "pop_art"
	case "max_colors":
		return "max_colors"
	default:
		return "max_colors"
	}
}
