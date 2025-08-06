package coupon

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

type CouponServiceDeps struct {
	CouponRepository *CouponRepository
}

type CouponService struct {
	deps *CouponServiceDeps
}

func NewCouponService(deps *CouponServiceDeps) *CouponService {
	return &CouponService{
		deps: deps,
	}
}

// SearchCoupons ищет купоны с фильтрацией
func (s *CouponService) SearchCoupons(code, status, size, style string, partnerID *uuid.UUID) ([]*Coupon, error) {
	coupons, err := s.deps.CouponRepository.Search(context.Background(), code, status, size, style, partnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to find coupons: %w", err)
	}
	return coupons, nil
}

// GetCouponByID получает купон по ID
func (s *CouponService) GetCouponByID(id uuid.UUID) (*Coupon, error) {
	coupon, err := s.deps.CouponRepository.GetByID(context.Background(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, errors.New("not found")
		}
		return nil, fmt.Errorf("failed to get coupon: %w", err)
	}
	return coupon, nil
}

// GetCouponByCode получает купон по коду
func (s *CouponService) GetCouponByCode(code string) (*Coupon, error) {
	coupon, err := s.deps.CouponRepository.GetByCode(context.Background(), code)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, errors.New("not found")
		}
		return nil, fmt.Errorf("failed to get coupon: %w", err)
	}
	return coupon, nil
}

// ActivateCoupon активирует купон
func (s *CouponService) ActivateCoupon(id uuid.UUID, originalImageURL, previewURL, schemaURL string) error {
	if err := s.deps.CouponRepository.ActivateCoupon(context.Background(), id, originalImageURL, previewURL, schemaURL); err != nil {
		return fmt.Errorf("failed to activate coupon: %w", err)
	}
	return nil
}

// ResetCoupon сбрасывает купон в исходное состояние
func (s *CouponService) ResetCoupon(id uuid.UUID) error {
	if err := s.deps.CouponRepository.ResetCoupon(context.Background(), id); err != nil {
		return fmt.Errorf("failed to reset coupon: %w", err)
	}
	return nil
}

// SendSchema отправляет схему на email
func (s *CouponService) SendSchema(id uuid.UUID, email string) error {
	if err := s.deps.CouponRepository.SendSchema(context.Background(), id, email); err != nil {
		return fmt.Errorf("failed to send schema: %w", err)
	}
	return nil
}

// MarkAsPurchased помечает купон как купленный
func (s *CouponService) MarkAsPurchased(id uuid.UUID, purchaseEmail string) error {
	if err := s.deps.CouponRepository.MarkAsPurchased(context.Background(), id, purchaseEmail); err != nil {
		return fmt.Errorf("failed to mark as purchased: %w", err)
	}
	return nil
}

// GetStatistics возвращает статистику по купонам
func (s *CouponService) GetStatistics(partnerID *uuid.UUID) (map[string]int64, error) {
	stats, err := s.deps.CouponRepository.GetStatistics(context.Background(), partnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get statistics: %w", err)
	}
	return stats, nil
}

// ValidateCoupon проверяет валидность купона
func (s *CouponService) ValidateCoupon(code string) (*CouponValidationResponse, error) {
	coupon, err := s.deps.CouponRepository.GetByCode(context.Background(), code)
	if err != nil {
		return &CouponValidationResponse{
			Valid:   false,
			Message: "Coupon not found",
		}, nil
	}

	if coupon.Status == "used" {
		size := string(coupon.Size)
		style := string(coupon.Style)
		return &CouponValidationResponse{
			Valid:   false,
			Message: "Coupon already used",
			UsedAt:  coupon.UsedAt,
			Size:    &size,
			Style:   &style,
		}, nil
	}

	size := string(coupon.Size)
	style := string(coupon.Style)
	return &CouponValidationResponse{
		Valid:   true,
		Message: "Coupon is valid and ready to use",
		Size:    &size,
		Style:   &style,
	}, nil
}

// ExportCoupons экспортирует купоны в текстовый файл
func (s *CouponService) ExportCoupons(partnerID *uuid.UUID, status, format string) (string, error) {
	coupons, err := s.deps.CouponRepository.Search(context.Background(), "", status, "", "", partnerID)
	if err != nil {
		return "", fmt.Errorf("failed to fetch coupons for export: %w", err)
	}

	var content strings.Builder

	if format == "full" {
		content.WriteString("Code\tPartner ID\tSize\tStyle\tStatus\tCreated At\tUsed At\n")
		for _, coupon := range coupons {
			usedAt := ""
			if coupon.UsedAt != nil {
				usedAt = coupon.UsedAt.Format(time.RFC3339)
			}
			content.WriteString(fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				coupon.Code,
				coupon.PartnerID.String(),
				coupon.Size,
				coupon.Style,
				coupon.Status,
				coupon.CreatedAt.Format(time.RFC3339),
				usedAt,
			))
		}
	} else {
		for _, coupon := range coupons {
			content.WriteString(coupon.Code + "\n")
		}
	}

	return content.String(), nil
}

// DownloadMaterials создает ZIP архив с материалами купона
func (s *CouponService) DownloadMaterials(id uuid.UUID) ([]byte, string, error) {
	coupon, err := s.deps.CouponRepository.GetByID(context.Background(), id)
	if err != nil {
		return nil, "", fmt.Errorf("coupon not found: %w", err)
	}

	if coupon.Status != "used" {
		return nil, "", fmt.Errorf("coupon must be used to download materials")
	}

	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	// Функция для добавления файла по URL в архив
	addFileToZip := func(fileURL, filename string) error {
		if fileURL == "" {
			return nil
		}

		resp, err := http.Get(fileURL)
		if err != nil {
			return fmt.Errorf("failed to download file: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("failed to download file: %w", err)
		}

		fileWriter, err := zipWriter.Create(filename)
		if err != nil {
			return fmt.Errorf("failed to create file writer: %w", err)
		}

		_, err = io.Copy(fileWriter, resp.Body)
		if err != nil {
			return fmt.Errorf("failed to copy file to zip: %w", err)
		}

		return nil
	}

	// Добавляем файлы в архив
	if coupon.OriginalImageURL != nil && *coupon.OriginalImageURL != "" {
		if err := addFileToZip(*coupon.OriginalImageURL, "original_image.jpg"); err != nil {
			return nil, "", fmt.Errorf("failed to add original image to zip: %w", err)
		}
	}

	if coupon.PreviewURL != nil && *coupon.PreviewURL != "" {
		if err := addFileToZip(*coupon.PreviewURL, "preview.jpg"); err != nil {
			return nil, "", fmt.Errorf("failed to add preview to zip: %w", err)
		}
	}

	if coupon.SchemaURL != nil && *coupon.SchemaURL != "" {
		if err := addFileToZip(*coupon.SchemaURL, "schema.pdf"); err != nil {
			return nil, "", fmt.Errorf("failed to add schema to zip: %w", err)
		}
	}

	// Добавляем информационный файл
	infoWriter, err := zipWriter.Create("coupon_info.txt")
	if err == nil {
		infoContent := fmt.Sprintf(`Coupon Information
Code: %s
Size: %s
Style: %s
Created: %s
Used: %s
`,
			coupon.Code,
			coupon.Size,
			coupon.Style,
			coupon.CreatedAt.Format(time.RFC3339),
			coupon.UsedAt.Format(time.RFC3339),
		)
		infoWriter.Write([]byte(infoContent))
	}

	err = zipWriter.Close()
	if err != nil {
		return nil, "", fmt.Errorf("failed to create archive: %w", err)
	}

	filename := fmt.Sprintf("coupon_%s_materials.zip", coupon.Code)
	return buf.Bytes(), filename, nil
}

// SearchCouponsWithPagination возвращает купоны с пагинацией
func (s *CouponService) SearchCouponsWithPagination(code, status, size, style string, partnerID *uuid.UUID, page, limit int) ([]*Coupon, int64, error) {
	// Валидация параметров
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	coupons, total, err := s.deps.CouponRepository.SearchWithPagination(
		context.Background(), code, status, size, style, partnerID, page, limit,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to fetch coupons: %w", err)
	}

	return coupons, int64(total), nil
}

// SearchCouponsByPartner возвращает купоны конкретного партнера
func (s *CouponService) SearchCouponsByPartner(partnerID uuid.UUID, status, size, style string) ([]*Coupon, error) {
	coupons, err := s.deps.CouponRepository.Search(context.Background(), "", status, size, style, &partnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch partner coupons: %w", err)
	}
	return coupons, nil
}

// BatchResetCoupons выполняет пакетный сброс купонов
func (s *CouponService) BatchResetCoupons(couponIDs []string) (*BatchResetResponse, error) {
	response := &BatchResetResponse{
		Success: make([]string, 0),
		Failed:  make([]string, 0),
		Errors:  make([]string, 0),
	}

	// Валидация входных данных
	if len(couponIDs) == 0 {
		return response, errors.New("no coupon IDs provided")
	}

	if len(couponIDs) > 1000 {
		return response, errors.New("too many coupon IDs (maximum 1000)")
	}

	// Конвертируем строковые ID в UUID
	var validIDs []uuid.UUID
	var invalidIDs []string

	for _, idStr := range couponIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			invalidIDs = append(invalidIDs, idStr)
			response.Errors = append(response.Errors, fmt.Sprintf("Invalid UUID format: %s", idStr))
		} else {
			validIDs = append(validIDs, id)
		}
	}

	// Добавляем невалидные ID в неуспешные
	response.Failed = append(response.Failed, invalidIDs...)

	if len(validIDs) == 0 {
		response.FailedCount = len(response.Failed)
		return response, nil
	}

	// Выполняем пакетный сброс
	success, failed, err := s.deps.CouponRepository.BatchReset(context.Background(), validIDs)
	if err != nil {
		// В случае общей ошибки все считаем неуспешными
		for _, id := range validIDs {
			response.Failed = append(response.Failed, id.String())
		}
		response.Errors = append(response.Errors, err.Error())
	} else {
		// Добавляем результаты в ответ
		for _, id := range success {
			response.Success = append(response.Success, id.String())
		}
		for _, id := range failed {
			response.Failed = append(response.Failed, id.String())
		}
	}

	response.SuccessCount = len(response.Success)
	response.FailedCount = len(response.Failed)

	return response, nil
}

// PreviewBatchDelete возвращает предпросмотр пакетного удаления
func (s *CouponService) PreviewBatchDelete(couponIDs []string) (*BatchDeletePreviewResponse, error) {
	// Валидация входных данных
	if len(couponIDs) == 0 {
		return nil, errors.New("no coupon IDs provided")
	}

	if len(couponIDs) > 1000 {
		return nil, errors.New("too many coupon IDs (maximum 1000)")
	}

	// Конвертируем строковые ID в UUID
	var validIDs []uuid.UUID
	for _, idStr := range couponIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			return nil, fmt.Errorf("invalid UUID format: %s", idStr)
		}
		validIDs = append(validIDs, id)
	}

	// Получаем информацию о купонах
	previews, err := s.deps.CouponRepository.GetCouponsForDeletion(context.Background(), validIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get coupons preview: %w", err)
	}

	// Подсчитываем статистику
	usedCount := 0
	newCount := 0
	for _, preview := range previews {
		switch preview.Status {
		case "used":
			usedCount++
		case "new":
			newCount++
		}
	}

	// Генерируем ключ подтверждения (временный токен)
	confirmationKey := uuid.New().String()
	expiresAt := time.Now().Add(15 * time.Minute) // Ключ действителен 15 минут

	response := &BatchDeletePreviewResponse{
		TotalCount:      len(previews),
		UsedCount:       usedCount,
		NewCount:        newCount,
		Coupons:         previews,
		ConfirmationKey: confirmationKey,
		ExpiresAt:       expiresAt,
	}

	// TODO: Сохранить ключ подтверждения в Redis или кэше с TTL 15 минут
	// Пока что просто возвращаем, в продакшене нужно использовать Redis

	return response, nil
}

// ExecuteBatchDelete выполняет пакетное удаление с подтверждением
func (s *CouponService) ExecuteBatchDelete(request BatchDeleteConfirmRequest) (*BatchDeleteResponse, error) {
	response := &BatchDeleteResponse{
		Deleted: make([]string, 0),
		Failed:  make([]string, 0),
		Errors:  make([]string, 0),
	}

	// Валидация входных данных
	if len(request.CouponIDs) == 0 {
		return response, errors.New("no coupon IDs provided")
	}

	if len(request.CouponIDs) > 1000 {
		return response, errors.New("too many coupon IDs (maximum 1000)")
	}

	// TODO: Проверить ключ подтверждения в Redis
	// Пока что делаем простую валидацию формата
	if _, err := uuid.Parse(request.ConfirmationKey); err != nil {
		return response, errors.New("invalid confirmation key")
	}

	// Конвертируем строковые ID в UUID
	var validIDs []uuid.UUID
	var invalidIDs []string

	for _, idStr := range request.CouponIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			invalidIDs = append(invalidIDs, idStr)
			response.Errors = append(response.Errors, fmt.Sprintf("Invalid UUID format: %s", idStr))
		} else {
			validIDs = append(validIDs, id)
		}
	}

	// Добавляем невалидные ID в неуспешные
	response.Failed = append(response.Failed, invalidIDs...)

	if len(validIDs) == 0 {
		response.FailedCount = len(response.Failed)
		return response, nil
	}

	// Получаем коды купонов перед удалением
	var codesBeforeDeletion []string
	coupons, err := s.deps.CouponRepository.GetCouponsForDeletion(context.Background(), validIDs)
	if err == nil {
		for _, coupon := range coupons {
			codesBeforeDeletion = append(codesBeforeDeletion, coupon.Code)
		}
	}

	// Выполняем пакетное удаление
	deletedCount, err := s.deps.CouponRepository.BatchDelete(context.Background(), validIDs)
	if err != nil {
		// В случае ошибки все считаем неуспешными
		for _, id := range validIDs {
			response.Failed = append(response.Failed, id.String())
		}
		response.Errors = append(response.Errors, err.Error())
	} else {
		// Если удаление прошло успешно
		if deletedCount == int64(len(validIDs)) {
			// Все купоны удалены успешно
			response.Deleted = codesBeforeDeletion
		} else {
			// Частично удалено - нужно проверить какие именно
			// В простой реализации считаем, что удалилось столько, сколько вернул БД
			deletedItems := int(deletedCount)
			if deletedItems > len(codesBeforeDeletion) {
				deletedItems = len(codesBeforeDeletion)
			}

			response.Deleted = codesBeforeDeletion[:deletedItems]

			// Остальные добавляем в неуспешные
			for i := deletedItems; i < len(codesBeforeDeletion); i++ {
				response.Failed = append(response.Failed, codesBeforeDeletion[i])
			}
		}
	}

	response.DeletedCount = len(response.Deleted)
	response.FailedCount = len(response.Failed)

	// TODO: Удалить ключ подтверждения из Redis после использования

	return response, nil
}

// ExportCouponsAdvanced экспортирует купоны с настраиваемыми форматами
func (s *CouponService) ExportCouponsAdvanced(options ExportOptionsRequest) ([]byte, string, string, error) {
	// Получаем данные для экспорта
	data, err := s.deps.CouponRepository.GetCouponsForExport(context.Background(), options)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to get export data: %w", err)
	}

	// Генерируем имя файла
	timestamp := time.Now().Format("20060102_150405")
	var filename string
	var contentType string

	switch options.FileFormat {
	case "csv":
		content, err := s.generateCSV(data, options)
		if err != nil {
			return nil, "", "", fmt.Errorf("failed to generate CSV: %w", err)
		}
		filename = fmt.Sprintf("coupons_export_%s_%s.csv", options.Format, timestamp)
		contentType = "text/csv"
		return content, filename, contentType, nil

	default:
		content, err := s.generateTXT(data, options)
		if err != nil {
			return nil, "", "", fmt.Errorf("failed to generate TXT: %w", err)
		}
		filename = fmt.Sprintf("coupons_export_%s_%s.txt", options.Format, timestamp)
		contentType = "text/plain"
		return content, filename, contentType, nil
	}
}

// generateTXT создает текстовый файл для экспорта
func (s *CouponService) generateTXT(data interface{}, options ExportOptionsRequest) ([]byte, error) {
	var content strings.Builder

	switch options.Format {
	case ExportFormatCodes:
		codes, ok := data.([]string)
		if !ok {
			return nil, errors.New("invalid data format for codes export")
		}

		if options.IncludeHeader {
			content.WriteString("Coupon Codes\n")
			content.WriteString("============\n\n")
		}

		// Для партнеров (один партнер) просто выводим коды
		if options.PartnerID != nil {
			for _, code := range codes {
				content.WriteString(code + "\n")
			}
		} else {
			// Для админа группируем по партнерам (по первым 4 символам кода)
			groupedCodes := make(map[string][]string)

			for _, code := range codes {
				if len(code) >= 4 {
					prefix := code[:4]
					groupedCodes[prefix] = append(groupedCodes[prefix], code)
				} else {
					groupedCodes["UNKN"] = append(groupedCodes["UNKN"], code)
				}
			}

			// Выводим группами
			for prefix, partnerCodes := range groupedCodes {
				if len(partnerCodes) > 0 {
					content.WriteString(fmt.Sprintf("Partner %s:\n", prefix))
					for _, code := range partnerCodes {
						content.WriteString(code + "\n")
					}
					content.WriteString("\n") // Пустая строка между партнерами
				}
			}
		}

	case ExportFormatPartner:
		// Партнер формат: Coupon Code, Partner Status, Coupon Status, Size, Style, Created At
		type PartnerExport struct {
			Code          string    `json:"code"`
			PartnerStatus string    `json:"partner_status"`
			CouponStatus  string    `json:"coupon_status"`
			Size          string    `json:"size"`
			Style         string    `json:"style"`
			CreatedAt     time.Time `json:"created_at"`
		}

		exports, ok := data.([]PartnerExport)
		if !ok {
			return nil, errors.New("invalid data format for partner export")
		}

		if options.IncludeHeader {
			content.WriteString("Partner Coupons Export\n")
			content.WriteString("======================\n\n")
		}

		for _, export := range exports {
			content.WriteString(fmt.Sprintf("Code: %s\n", export.Code))
			content.WriteString(fmt.Sprintf("Partner Status: %s\n", export.PartnerStatus))
			content.WriteString(fmt.Sprintf("Coupon Status: %s\n", export.CouponStatus))
			content.WriteString(fmt.Sprintf("Size: %s\n", export.Size))
			content.WriteString(fmt.Sprintf("Style: %s\n", export.Style))
			content.WriteString(fmt.Sprintf("Created: %s\n", export.CreatedAt.Format("2006-01-02 15:04:05")))
			content.WriteString("---\n")
		}

	case ExportFormatAdmin:
		// Админ формат: Coupon Code, Partner ID, Partner Status, Coupon Status, Size, Style, Brand Name, Email, Created At
		type AdminExport struct {
			Code          string    `json:"code"`
			PartnerID     string    `json:"partner_id"`
			PartnerStatus string    `json:"partner_status"`
			CouponStatus  string    `json:"coupon_status"`
			Size          string    `json:"size"`
			Style         string    `json:"style"`
			BrandName     string    `json:"brand_name"`
			Email         string    `json:"email"`
			CreatedAt     time.Time `json:"created_at"`
		}

		exports, ok := data.([]AdminExport)
		if !ok {
			return nil, errors.New("invalid data format for admin export")
		}

		if options.IncludeHeader {
			content.WriteString("Admin Coupons Export\n")
			content.WriteString("====================\n\n")
		}

		for _, export := range exports {
			content.WriteString(fmt.Sprintf("Code: %s\n", export.Code))
			content.WriteString(fmt.Sprintf("Partner ID: %s\n", export.PartnerID))
			content.WriteString(fmt.Sprintf("Partner Status: %s\n", export.PartnerStatus))
			content.WriteString(fmt.Sprintf("Coupon Status: %s\n", export.CouponStatus))
			content.WriteString(fmt.Sprintf("Size: %s\n", export.Size))
			content.WriteString(fmt.Sprintf("Style: %s\n", export.Style))
			content.WriteString(fmt.Sprintf("Brand Name: %s\n", export.BrandName))
			content.WriteString(fmt.Sprintf("Email: %s\n", export.Email))
			content.WriteString(fmt.Sprintf("Created: %s\n", export.CreatedAt.Format("2006-01-02 15:04:05")))
			content.WriteString("---\n")
		}

	case ExportFormatFull:
		// Полный формат: Coupon Code, Partner ID, Partner Status, Coupon Status, Size, Style, Brand Name, Email, Created At, Used At
		type FullExport struct {
			Code          string     `json:"code"`
			PartnerID     string     `json:"partner_id"`
			PartnerStatus string     `json:"partner_status"`
			CouponStatus  string     `json:"coupon_status"`
			Size          string     `json:"size"`
			Style         string     `json:"style"`
			BrandName     string     `json:"brand_name"`
			Email         string     `json:"email"`
			CreatedAt     time.Time  `json:"created_at"`
			UsedAt        *time.Time `json:"used_at"`
		}

		exports, ok := data.([]FullExport)
		if !ok {
			return nil, errors.New("invalid data format for full export")
		}

		if options.IncludeHeader {
			content.WriteString("Full Coupons Export\n")
			content.WriteString("===================\n\n")
		}

		for _, export := range exports {
			content.WriteString(fmt.Sprintf("Code: %s\n", export.Code))
			content.WriteString(fmt.Sprintf("Partner ID: %s\n", export.PartnerID))
			content.WriteString(fmt.Sprintf("Partner Status: %s\n", export.PartnerStatus))
			content.WriteString(fmt.Sprintf("Coupon Status: %s\n", export.CouponStatus))
			content.WriteString(fmt.Sprintf("Size: %s\n", export.Size))
			content.WriteString(fmt.Sprintf("Style: %s\n", export.Style))
			content.WriteString(fmt.Sprintf("Brand Name: %s\n", export.BrandName))
			content.WriteString(fmt.Sprintf("Email: %s\n", export.Email))
			content.WriteString(fmt.Sprintf("Created: %s\n", export.CreatedAt.Format("2006-01-02 15:04:05")))
			if export.UsedAt != nil {
				content.WriteString(fmt.Sprintf("Used: %s\n", export.UsedAt.Format("2006-01-02 15:04:05")))
			}
			content.WriteString("---\n")
		}

	default:
		// Для других форматов используем JSON
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal data: %w", err)
		}
		content.Write(jsonData)
	}

	return []byte(content.String()), nil
}

// generateCSV создает CSV файл для экспорта
func (s *CouponService) generateCSV(data interface{}, options ExportOptionsRequest) ([]byte, error) {
	var buffer bytes.Buffer
	writer := csv.NewWriter(&buffer)

	delimiter := ","
	if options.Delimiter != "" {
		delimiter = options.Delimiter
	}
	writer.Comma = rune(delimiter[0])

	defer writer.Flush()

	switch options.Format {
	case ExportFormatCodes:
		codes, ok := data.([]string)
		if !ok {
			return nil, errors.New("invalid data format for codes export")
		}

		if options.IncludeHeader {
			writer.Write([]string{"Coupon Code"})
		}

		// Для партнеров (один партнер) просто выводим коды
		if options.PartnerID != nil {
			for _, code := range codes {
				writer.Write([]string{code})
			}
		} else {
			// Для админа группируем по партнерам (по первым 4 символам кода)
			groupedCodes := make(map[string][]string)

			for _, code := range codes {
				if len(code) >= 4 {
					prefix := code[:4]
					groupedCodes[prefix] = append(groupedCodes[prefix], code)
				} else {
					groupedCodes["UNKN"] = append(groupedCodes["UNKN"], code)
				}
			}

			// Выводим группами
			for prefix, partnerCodes := range groupedCodes {
				if len(partnerCodes) > 0 {
					// Добавляем заголовок партнера
					writer.Write([]string{fmt.Sprintf("=== Partner %s ===", prefix)})

					for _, code := range partnerCodes {
						writer.Write([]string{code})
					}

					// Пустая строка между партнерами
					writer.Write([]string{""})
				}
			}
		}

	case ExportFormatBasic:
		exports, ok := data.([]struct {
			Code      string    `json:"code"`
			Status    string    `json:"status"`
			Size      string    `json:"size"`
			Style     string    `json:"style"`
			CreatedAt time.Time `json:"created_at"`
		})
		if !ok {
			return nil, errors.New("invalid data format for basic export")
		}

		if options.IncludeHeader {
			writer.Write([]string{"Code", "Status", "Size", "Style", "Created At"})
		}

		for _, export := range exports {
			writer.Write([]string{
				export.Code,
				export.Status,
				export.Size,
				export.Style,
				export.CreatedAt.Format(time.RFC3339),
			})
		}

	case ExportFormatAdmin:
		// Админ формат: Coupon Code, Partner ID, Partner Status, Coupon Status, Size, Style, Brand Name, Email, Created At
		type AdminExport struct {
			Code          string    `json:"code"`
			PartnerID     string    `json:"partner_id"`
			PartnerStatus string    `json:"partner_status"`
			CouponStatus  string    `json:"coupon_status"`
			Size          string    `json:"size"`
			Style         string    `json:"style"`
			BrandName     string    `json:"brand_name"`
			Email         string    `json:"email"`
			CreatedAt     time.Time `json:"created_at"`
		}

		exports, ok := data.([]AdminExport)
		if !ok {
			return nil, errors.New("invalid data format for admin export")
		}

		if options.IncludeHeader {
			writer.Write([]string{"Coupon Code", "Partner ID", "Partner Status", "Coupon Status", "Size", "Style", "Brand Name", "Email", "Created At"})
		}

		for _, export := range exports {
			writer.Write([]string{
				export.Code,
				export.PartnerID,
				export.PartnerStatus,
				export.CouponStatus,
				export.Size,
				export.Style,
				export.BrandName,
				export.Email,
				export.CreatedAt.Format("2006-01-02 15:04:05"),
			})
		}

	case ExportFormatPartner:
		// Партнер формат: Coupon Code, Partner Status, Coupon Status, Size, Style, Created At
		type PartnerExport struct {
			Code          string    `json:"code"`
			PartnerStatus string    `json:"partner_status"`
			CouponStatus  string    `json:"coupon_status"`
			Size          string    `json:"size"`
			Style         string    `json:"style"`
			CreatedAt     time.Time `json:"created_at"`
		}

		exports, ok := data.([]PartnerExport)
		if !ok {
			return nil, errors.New("invalid data format for partner export")
		}

		if options.IncludeHeader {
			writer.Write([]string{"Coupon Code", "Partner Status", "Coupon Status", "Size", "Style", "Created At"})
		}

		for _, export := range exports {
			writer.Write([]string{
				export.Code,
				export.PartnerStatus,
				export.CouponStatus,
				export.Size,
				export.Style,
				export.CreatedAt.Format("2006-01-02 15:04:05"),
			})
		}

	case ExportFormatFull:
		// Полный формат: Coupon Code, Partner ID, Partner Status, Coupon Status, Size, Style, Brand Name, Email, Created At, Used At
		type FullExport struct {
			Code          string     `json:"code"`
			PartnerID     string     `json:"partner_id"`
			PartnerStatus string     `json:"partner_status"`
			CouponStatus  string     `json:"coupon_status"`
			Size          string     `json:"size"`
			Style         string     `json:"style"`
			BrandName     string     `json:"brand_name"`
			Email         string     `json:"email"`
			CreatedAt     time.Time  `json:"created_at"`
			UsedAt        *time.Time `json:"used_at"`
		}

		exports, ok := data.([]FullExport)
		if !ok {
			return nil, errors.New("invalid data format for full export")
		}

		if options.IncludeHeader {
			writer.Write([]string{"Coupon Code", "Partner ID", "Partner Status", "Coupon Status", "Size", "Style", "Brand Name", "Email", "Created At", "Used At"})
		}

		for _, export := range exports {
			usedAt := ""
			if export.UsedAt != nil {
				usedAt = export.UsedAt.Format("2006-01-02 15:04:05")
			}

			writer.Write([]string{
				export.Code,
				export.PartnerID,
				export.PartnerStatus,
				export.CouponStatus,
				export.Size,
				export.Style,
				export.BrandName,
				export.Email,
				export.CreatedAt.Format("2006-01-02 15:04:05"),
				usedAt,
			})
		}

	default:
		// Для сложных форматов используем рефлексию для общей обработки
		return s.generateCSVGeneric(data, options)
	}

	return buffer.Bytes(), nil
}

// generateCSVGeneric универсальный генератор CSV через рефлексию
func (s *CouponService) generateCSVGeneric(data interface{}, options ExportOptionsRequest) ([]byte, error) {
	// Этот метод использует рефлексию для обработки сложных структур
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var buffer bytes.Buffer
	writer := csv.NewWriter(&buffer)
	defer writer.Flush()

	if options.IncludeHeader {
		writer.Write([]string{"JSON Data"})
	}
	writer.Write([]string{string(jsonData)})

	return buffer.Bytes(), nil
}
