package preview

import (
	"context"
	"fmt"
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

// GeneratePreview создает превью мозаики используя существующий функционал
func (s *PreviewService) GeneratePreview(ctx context.Context, req *PreviewRequest) (*PreviewResponse, error) {
	previewID := uuid.New()

	// Создаем временную директорию для превью
	tempDir := filepath.Join(s.deps.WorkingDir, "preview", previewID.String())
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Получаем путь к палитре для выбранного стиля
	palettePath, err := s.deps.PaletteService.GetPalettePath(palette.Style(req.Style))
	if err != nil {
		return nil, fmt.Errorf("failed to get palette path: %w", err)
	}

	// Определяем размеры мозаики на основе выбранного размера
	stonesX, stonesY := s.getSizeDimensions(req.Size)

	// Создаем запрос для генерации мозаики
	generationReq := &mosaic.GenerationRequest{
		ImagePath:   "", // Пустой путь для генерации превью без изображения
		StonesX:     stonesX,
		StonesY:     stonesY,
		StoneSizeMM: 4.0, // Стандартный размер камня
		DPI:         300,
		PreviewDPI:  150,       // DPI для превью
		SchemeDPI:   0,         // Не генерируем схему
		Mode:        "preview", // Режим только для превью
		Style:       req.Style,
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

	// Загружаем превью в S3
	file, err := os.Open(result.PreviewPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open preview file: %w", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	// Загружаем в бакет превью
	objectKey, err := s.deps.S3Client.UploadPreviewFile(
		ctx,
		file,
		fileInfo.Size(),
		"image/png",
		"previews",
		previewID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to upload preview to S3: %w", err)
	}

	// Получаем публичную ссылку на превью
	previewURL, err := s.deps.S3Client.GetFileURL(ctx, objectKey, 24*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("failed to get preview URL: %w", err)
	}

	// Очищаем временные файлы
	s.deps.MosaicGenerator.Cleanup(result)

	return &PreviewResponse{
		PreviewID:   previewID.String(),
		PreviewURL:  previewURL,
		Size:        req.Size,
		Style:       req.Style,
		GeneratedAt: time.Now().Format(time.RFC3339),
	}, nil
}

// getSizeDimensions возвращает размеры мозаики в камнях для выбранного размера
func (s *PreviewService) getSizeDimensions(size string) (int, int) {
	switch size {
	case "20x20":
		return 20, 20
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
