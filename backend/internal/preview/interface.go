package preview

import "context"

// PreviewServiceInterface определяет методы для сервиса превью
type PreviewServiceInterface interface {
	GeneratePreview(ctx context.Context, req *PreviewRequest) (*PreviewResponse, error)
}
