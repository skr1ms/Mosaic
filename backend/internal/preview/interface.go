package preview

import (
	"context"
	"mime/multipart"
)

type PreviewServiceInterface interface {
	GeneratePreview(ctx context.Context, file *multipart.FileHeader, size, style, contrast string) (*PreviewData, error)
	GenerateAIPreview(ctx context.Context, file *multipart.FileHeader, prompt string) (*PreviewData, error)
	ApplyStyle(img interface{}, style string) interface{}
	ApplyContrast(img interface{}, level string) interface{}
}

type PreviewRepositoryInterface interface {
	Create(ctx context.Context, preview *PreviewData) error
	GetByID(ctx context.Context, id string) (*PreviewData, error)
	Delete(ctx context.Context, id string) error
	GetByUserSession(ctx context.Context, sessionID string) ([]*PreviewData, error)
	CleanupExpired(ctx context.Context) error
}

type PreviewHandlerDependencies struct {
	PreviewService PreviewServiceInterface
	Logger         interface{}
}