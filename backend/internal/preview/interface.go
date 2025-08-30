package preview

import (
	"context"
	"mime/multipart"
)

// PreviewServiceInterface определяет методы для сервиса превью
type PreviewServiceInterface interface {
	GeneratePreview(ctx context.Context, file *multipart.FileHeader, size, style string) (map[string]any, error)
}
