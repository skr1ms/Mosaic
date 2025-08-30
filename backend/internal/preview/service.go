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

// GeneratePreview создает превью мозаики из загруженного файла
func (s *PreviewService) GeneratePreview(ctx context.Context, file *multipart.FileHeader, size, style string) (map[string]any, error) {
	previewID := uuid.New()

	// Создаем временную директорию для превью
	tempDir := filepath.Join(s.deps.WorkingDir, "preview", previewID.String())
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Сохраняем загруженный файл во временную директорию
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Определяем расширение файла
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

	// Получаем путь к палитре для выбранного стиля
	palettePath, err := s.deps.PaletteService.GetPalettePath(palette.Style(style))
	if err != nil {
		return nil, fmt.Errorf("failed to get palette path: %w", err)
	}

	// Определяем размеры мозаики на основе выбранного размера
	stonesX, stonesY := s.getSizeDimensions(size)

	// Создаем запрос для генерации мозаики
	generationReq := &mosaic.GenerationRequest{
		ImagePath:   tempImagePath,
		StonesX:     stonesX,
		StonesY:     stonesY,
		StoneSizeMM: 2.52, // Стандартный размер камня
		DPI:         150,
		PreviewDPI:  120,       // DPI для превью
		SchemeDPI:   0,         // Не генерируем схему
		Mode:        "preview", // Режим только для превью
		Style:       s.mapStyleToMosaicStyle(style),
		WithLegend:  false, // Не нужна легенда для превью
		Threads:     4,
		PalettePath: palettePath,
	}

	// Генерируем мозаику (только превью)
	result, err := s.deps.MosaicGenerator.Generate(ctx, generationReq)
	if err != nil {
		return nil, fmt.Errorf("failed to generate mosaic preview: %w", err)
	}

	// Проверяем, что превью было создано
	if result.PreviewPath == "" {
		return nil, fmt.Errorf("preview file not generated")
	}

	// Загружаем превью в S3 (в бакет preview-images)
	previewFile, err := os.Open(result.PreviewPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open preview file: %w", err)
	}
	defer previewFile.Close()

	fileInfo, err := previewFile.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	// Создаем ключ для файла в формате: previews/{uuid}.png
	objectKey := fmt.Sprintf("previews/%s.png", previewID.String())

	// Загружаем в бакет preview-images
	err = s.deps.S3Client.UploadToPreviewBucket(ctx, objectKey, previewFile, fileInfo.Size(), "image/png")
	if err != nil {
		return nil, fmt.Errorf("failed to upload preview to S3: %w", err)
	}

	// Получаем публичную ссылку на превью
	previewURL := s.deps.S3Client.GetPreviewURL(objectKey)

	return map[string]any{
		"preview_id":   previewID.String(),
		"preview_url":  previewURL,
		"size":         size,
		"style":        style,
		"generated_at": time.Now().Format(time.RFC3339),
	}, nil
}

// getSizeDimensions возвращает размеры мозаики в камнях для выбранного размера
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
		return 30, 40 // По умолчанию
	}
}

// mapStyleToMosaicStyle конвертирует стиль API в стиль мозаики
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
