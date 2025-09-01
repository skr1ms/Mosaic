package preview

import (
	"context"
	"mime/multipart"
)

// PreviewServiceInterface defines methods for preview service
type PreviewServiceInterface interface {
	GeneratePreview(ctx context.Context, file *multipart.FileHeader, size, style string, useAI bool) (map[string]any, error)
	GetPreviewDownloadURL(previewID string) (string, error)
	GetPreviewData(previewID string) ([]byte, error)
}
