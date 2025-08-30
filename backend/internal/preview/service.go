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

	// Determine mosaic dimensions based on selected size
	stonesX, stonesY := s.getSizeDimensions(size)

	// Create request for mosaic generation
	generationReq := &mosaic.GenerationRequest{
		ImagePath:   tempImagePath,
		StonesX:     stonesX,
		StonesY:     stonesY,
		StoneSizeMM: 2.52, // Standard stone size
		DPI:         150,
		PreviewDPI:  120,       // DPI for preview
		SchemeDPI:   0,         // Don't generate scheme
		Mode:        "preview", // Preview mode only
		Style:       s.mapStyleToMosaicStyle(style),
		WithLegend:  false, // No legend needed for preview
		Threads:     4,
		PalettePath: palettePath,
	}

	// Generate mosaic (preview only)
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

// getSizeDimensions returns mosaic dimensions in stones for selected size
func (s *PreviewService) getSizeDimensions(size string) (int, int) {
	switch size {
	case "21x30":
		return 21, 30
	case "30x40":
		return 30, 40
	case "40x40":
		return 40, 40
	case "40x50":
		return 40, 50
	case "40x60":
		return 40, 60
	case "50x70":
		return 50, 70
	default:
		return 30, 40 // Default
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
