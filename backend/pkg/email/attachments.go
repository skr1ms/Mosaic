package email

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/skr1ms/mosaic/pkg/middleware"
)

var logger = middleware.NewLogger()

// Attachment represents email attachment
type Attachment struct {
	Name        string
	ContentType string
	Data        []byte
	Size        int64
	Inline      bool
	ContentID   string
}

// AttachmentService manages email attachments
type AttachmentService struct {
	maxFileSize  int64
	allowedTypes map[string]bool
	httpClient   *http.Client
}

// NewAttachmentService creates new attachment service
func NewAttachmentService() *AttachmentService {
	return &AttachmentService{
		maxFileSize: 10 * 1024 * 1024, // 10MB
		allowedTypes: map[string]bool{
			"application/pdf":          true,
			"image/jpeg":               true,
			"image/jpg":                true,
			"image/png":                true,
			"image/gif":                true,
			"application/zip":          true,
			"text/plain":               true,
			"text/csv":                 true,
			"application/vnd.ms-excel": true,
			"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": true,
		},
		httpClient: &http.Client{},
	}
}

func (as *AttachmentService) CreateAttachmentFromFile(filePath string, data []byte) (*Attachment, error) {
	if len(data) == 0 {
		logger.GetZerologLogger().Error().Str("file_path", filePath).Msg("File data is empty")
		return nil, fmt.Errorf("file data is empty")
	}

	if int64(len(data)) > as.maxFileSize {
		logger.GetZerologLogger().Error().Int("file_size", len(data)).Int64("max_size", as.maxFileSize).Str("file_path", filePath).Msg("File size exceeds maximum allowed size")
		return nil, fmt.Errorf("file size %d exceeds maximum allowed size %d", len(data), as.maxFileSize)
	}

	contentType := as.detectContentType(data, filePath)

	if !as.allowedTypes[contentType] {
		logger.GetZerologLogger().Error().Str("content_type", contentType).Str("file_path", filePath).Msg("File type not allowed")
		return nil, fmt.Errorf("file type %s is not allowed", contentType)
	}

	fileName := filepath.Base(filePath)

	attachment := &Attachment{
		Name:        fileName,
		ContentType: contentType,
		Data:        data,
		Size:        int64(len(data)),
		Inline:      false,
	}

	logger.GetZerologLogger().Info().
		Str("file", fileName).
		Str("type", contentType).
		Int64("size", attachment.Size).
		Msg("Created email attachment")

	return attachment, nil
}

func (as *AttachmentService) CreateAttachmentFromURL(ctx context.Context, url, fileName string) (*Attachment, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		logger.GetZerologLogger().Error().Err(err).Str("url", url).Msg("Failed to create HTTP request for attachment download")
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := as.httpClient.Do(req)
	if err != nil {
		logger.GetZerologLogger().Error().Err(err).Str("url", url).Msg("Failed to download file from URL")
		return nil, fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.GetZerologLogger().Error().Int("status_code", resp.StatusCode).Str("url", url).Msg("Failed to download file from URL - non-OK status")
		return nil, fmt.Errorf("failed to download file: status %d", resp.StatusCode)
	}

	limitedReader := io.LimitReader(resp.Body, as.maxFileSize+1)
	data, err := io.ReadAll(limitedReader)
	if err != nil {
		logger.GetZerologLogger().Error().Err(err).Str("url", url).Msg("Failed to read file data from URL")
		return nil, fmt.Errorf("failed to read file data: %w", err)
	}

	if int64(len(data)) > as.maxFileSize {
		logger.GetZerologLogger().Error().Int("file_size", len(data)).Int64("max_size", as.maxFileSize).Str("url", url).Msg("File size from URL exceeds maximum allowed size")
		return nil, fmt.Errorf("file size exceeds maximum allowed size %d", as.maxFileSize)
	}

	if fileName == "" {
		fileName = filepath.Base(url)
		if fileName == "." || fileName == "/" {
			fileName = "attachment"
		}
	}

	return as.CreateAttachmentFromFile(fileName, data)
}

func (as *AttachmentService) CreateInlineAttachment(data []byte, fileName, contentID string) (*Attachment, error) {
	attachment, err := as.CreateAttachmentFromFile(fileName, data)
	if err != nil {
		return nil, err
	}

	if !strings.HasPrefix(attachment.ContentType, "image/") {
		logger.GetZerologLogger().Error().Str("content_type", attachment.ContentType).Str("file_name", fileName).Msg("Inline attachments must be images")
		return nil, fmt.Errorf("inline attachments must be images, got %s", attachment.ContentType)
	}

	attachment.Inline = true
	attachment.ContentID = contentID

	return attachment, nil
}

func (as *AttachmentService) ValidateAttachment(attachment *Attachment) error {
	if attachment == nil {
		logger.GetZerologLogger().Error().Msg("Attachment is nil during validation")
		return fmt.Errorf("attachment is nil")
	}

	if attachment.Name == "" {
		logger.GetZerologLogger().Error().Msg("Attachment name is required during validation")
		return fmt.Errorf("attachment name is required")
	}

	if len(attachment.Data) == 0 {
		logger.GetZerologLogger().Error().Str("attachment_name", attachment.Name).Msg("Attachment data is empty during validation")
		return fmt.Errorf("attachment data is empty")
	}

	if attachment.Size > as.maxFileSize {
		logger.GetZerologLogger().Error().Int64("attachment_size", attachment.Size).Int64("max_size", as.maxFileSize).Str("attachment_name", attachment.Name).Msg("Attachment size exceeds maximum during validation")
		return fmt.Errorf("attachment size %d exceeds maximum %d", attachment.Size, as.maxFileSize)
	}

	if !as.allowedTypes[attachment.ContentType] {
		logger.GetZerologLogger().Error().Str("content_type", attachment.ContentType).Str("attachment_name", attachment.Name).Msg("Attachment type not allowed during validation")
		return fmt.Errorf("attachment type %s is not allowed", attachment.ContentType)
	}

	return nil
}

// GetTotalSize returns total size of all attachments
func (as *AttachmentService) GetTotalSize(attachments []*Attachment) int64 {
	var total int64
	for _, attachment := range attachments {
		if attachment != nil {
			total += attachment.Size
		}
	}
	return total
}

// FormatAttachmentForMIME formats attachment for MIME message
func (as *AttachmentService) FormatAttachmentForMIME(attachment *Attachment) string {
	if attachment == nil {
		return ""
	}

	encodedData := base64.StdEncoding.EncodeToString(attachment.Data)
	lines := as.splitBase64(encodedData, 76)

	var mimeData strings.Builder

	if attachment.Inline {
		mimeData.WriteString(fmt.Sprintf("Content-Type: %s\r\n", attachment.ContentType))
		mimeData.WriteString("Content-Transfer-Encoding: base64\r\n")
		mimeData.WriteString(fmt.Sprintf("Content-ID: <%s>\r\n", attachment.ContentID))
		mimeData.WriteString(fmt.Sprintf("Content-Disposition: inline; filename=\"%s\"\r\n", attachment.Name))
	} else {
		mimeData.WriteString(fmt.Sprintf("Content-Type: %s\r\n", attachment.ContentType))
		mimeData.WriteString("Content-Transfer-Encoding: base64\r\n")
		mimeData.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n", attachment.Name))
	}

	mimeData.WriteString("\r\n")
	mimeData.WriteString(strings.Join(lines, "\r\n"))
	mimeData.WriteString("\r\n")

	return mimeData.String()
}

// detectContentType determines file MIME type
func (as *AttachmentService) detectContentType(data []byte, fileName string) string {
	contentType := http.DetectContentType(data)

	if contentType == "application/octet-stream" || contentType == "text/plain; charset=utf-8" {
		ext := strings.ToLower(filepath.Ext(fileName))
		switch ext {
		case ".pdf":
			return "application/pdf"
		case ".jpg", ".jpeg":
			return "image/jpeg"
		case ".png":
			return "image/png"
		case ".gif":
			return "image/gif"
		case ".zip":
			return "application/zip"
		case ".csv":
			return "text/csv"
		case ".txt":
			return "text/plain"
		case ".xls":
			return "application/vnd.ms-excel"
		case ".xlsx":
			return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
		default:
			mimeType := mime.TypeByExtension(ext)
			if mimeType != "" {
				return mimeType
			}
		}
	}

	return contentType
}

// splitBase64 splits base64 string into lines of specified length
func (as *AttachmentService) splitBase64(data string, lineLength int) []string {
	var lines []string
	for i := 0; i < len(data); i += lineLength {
		end := i + lineLength
		if end > len(data) {
			end = len(data)
		}
		lines = append(lines, data[i:end])
	}
	return lines
}

// SetMaxFileSize sets maximum file size
func (as *AttachmentService) SetMaxFileSize(size int64) {
	as.maxFileSize = size
}

// AddAllowedType adds allowed MIME type
func (as *AttachmentService) AddAllowedType(mimeType string) {
	as.allowedTypes[mimeType] = true
}

// RemoveAllowedType removes allowed MIME type
func (as *AttachmentService) RemoveAllowedType(mimeType string) {
	delete(as.allowedTypes, mimeType)
}

// GetAllowedTypes returns list of allowed MIME types
func (as *AttachmentService) GetAllowedTypes() []string {
	var types []string
	for mimeType := range as.allowedTypes {
		types = append(types, mimeType)
	}
	return types
}
