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

	"github.com/rs/zerolog/log"
)

// Attachment представляет вложение email
type Attachment struct {
	Name        string // Имя файла
	ContentType string // MIME тип
	Data        []byte // Данные файла
	Size        int64  // Размер файла в байтах
	Inline      bool   // Встроенное изображение (для HTML)
	ContentID   string // Content-ID для встроенных изображений
}

// AttachmentService управляет вложениями email
type AttachmentService struct {
	maxFileSize  int64 // Максимальный размер файла в байтах (по умолчанию 10MB)
	allowedTypes map[string]bool
	httpClient   *http.Client
}

// NewAttachmentService создает новый сервис вложений
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

// CreateAttachmentFromFile создает вложение из файла
func (as *AttachmentService) CreateAttachmentFromFile(filePath string, data []byte) (*Attachment, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("file data is empty")
	}

	if int64(len(data)) > as.maxFileSize {
		return nil, fmt.Errorf("file size %d exceeds maximum allowed size %d", len(data), as.maxFileSize)
	}

	contentType := as.detectContentType(data, filePath)

	if !as.allowedTypes[contentType] {
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

	log.Info().
		Str("file", fileName).
		Str("type", contentType).
		Int64("size", attachment.Size).
		Msg("Created email attachment")

	return attachment, nil
}

// CreateAttachmentFromURL создает вложение из URL
func (as *AttachmentService) CreateAttachmentFromURL(ctx context.Context, url, fileName string) (*Attachment, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := as.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download file: status %d", resp.StatusCode)
	}

	limitedReader := io.LimitReader(resp.Body, as.maxFileSize+1)
	data, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read file data: %w", err)
	}

	if int64(len(data)) > as.maxFileSize {
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

// CreateInlineAttachment создает встроенное вложение (для изображений в HTML)
func (as *AttachmentService) CreateInlineAttachment(data []byte, fileName, contentID string) (*Attachment, error) {
	attachment, err := as.CreateAttachmentFromFile(fileName, data)
	if err != nil {
		return nil, err
	}

	// Проверяем, что это изображение
	if !strings.HasPrefix(attachment.ContentType, "image/") {
		return nil, fmt.Errorf("inline attachments must be images, got %s", attachment.ContentType)
	}

	attachment.Inline = true
	attachment.ContentID = contentID

	return attachment, nil
}

// ValidateAttachment проверяет вложение на корректность
func (as *AttachmentService) ValidateAttachment(attachment *Attachment) error {
	if attachment == nil {
		return fmt.Errorf("attachment is nil")
	}

	if attachment.Name == "" {
		return fmt.Errorf("attachment name is required")
	}

	if len(attachment.Data) == 0 {
		return fmt.Errorf("attachment data is empty")
	}

	if attachment.Size > as.maxFileSize {
		return fmt.Errorf("attachment size %d exceeds maximum %d", attachment.Size, as.maxFileSize)
	}

	if !as.allowedTypes[attachment.ContentType] {
		return fmt.Errorf("attachment type %s is not allowed", attachment.ContentType)
	}

	return nil
}

// GetTotalSize возвращает общий размер всех вложений
func (as *AttachmentService) GetTotalSize(attachments []*Attachment) int64 {
	var total int64
	for _, attachment := range attachments {
		if attachment != nil {
			total += attachment.Size
		}
	}
	return total
}

// FormatAttachmentForMIME форматирует вложение для MIME сообщения
func (as *AttachmentService) FormatAttachmentForMIME(attachment *Attachment) string {
	if attachment == nil {
		return ""
	}

	// Кодируем данные в base64
	encodedData := base64.StdEncoding.EncodeToString(attachment.Data)

	// Разбиваем на строки по 76 символов (стандарт MIME)
	lines := as.splitBase64(encodedData, 76)

	var mimeData strings.Builder

	if attachment.Inline {
		// Встроенное вложение
		mimeData.WriteString(fmt.Sprintf("Content-Type: %s\r\n", attachment.ContentType))
		mimeData.WriteString("Content-Transfer-Encoding: base64\r\n")
		mimeData.WriteString(fmt.Sprintf("Content-ID: <%s>\r\n", attachment.ContentID))
		mimeData.WriteString(fmt.Sprintf("Content-Disposition: inline; filename=\"%s\"\r\n", attachment.Name))
	} else {
		// Обычное вложение
		mimeData.WriteString(fmt.Sprintf("Content-Type: %s\r\n", attachment.ContentType))
		mimeData.WriteString("Content-Transfer-Encoding: base64\r\n")
		mimeData.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n", attachment.Name))
	}

	mimeData.WriteString("\r\n")
	mimeData.WriteString(strings.Join(lines, "\r\n"))
	mimeData.WriteString("\r\n")

	return mimeData.String()
}

// detectContentType определяет MIME тип файла
func (as *AttachmentService) detectContentType(data []byte, fileName string) string {
	// Сначала пытаемся определить по содержимому
	contentType := http.DetectContentType(data)

	// Если не удалось определить или получили generic тип, пытаемся по расширению
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
			// Пытаемся через mime пакет
			mimeType := mime.TypeByExtension(ext)
			if mimeType != "" {
				return mimeType
			}
		}
	}

	return contentType
}

// splitBase64 разбивает base64 строку на строки заданной длины
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

// SetMaxFileSize устанавливает максимальный размер файла
func (as *AttachmentService) SetMaxFileSize(size int64) {
	as.maxFileSize = size
}

// AddAllowedType добавляет разрешенный MIME тип
func (as *AttachmentService) AddAllowedType(mimeType string) {
	as.allowedTypes[mimeType] = true
}

// RemoveAllowedType удаляет разрешенный MIME тип
func (as *AttachmentService) RemoveAllowedType(mimeType string) {
	delete(as.allowedTypes, mimeType)
}

// GetAllowedTypes возвращает список разрешенных MIME типов
func (as *AttachmentService) GetAllowedTypes() []string {
	var types []string
	for mimeType := range as.allowedTypes {
		types = append(types, mimeType)
	}
	return types
}
