package preview

import (
	"context"
	"mime/multipart"
)

// PreviewServiceInterface defines methods for preview service
type PreviewServiceInterface interface {
	GeneratePreview(ctx context.Context, file *multipart.FileHeader, size, style string) (map[string]any, error)
	GetPreviewDownloadURL(previewID string) (string, error)
}
