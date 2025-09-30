package coupon

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type CouponServiceDeps struct {
	CouponRepository  CouponRepositoryInterface
	PartnerRepository PartnerRepositoryInterface
	RedisClient       RedisClientInterface
	S3Client          S3Interface
}

type CouponService struct {
	deps *CouponServiceDeps
}

func NewCouponService(deps *CouponServiceDeps) *CouponService {
	return &CouponService{
		deps: deps,
	}
}

// SearchCoupons searches coupons with filtering
func (s *CouponService) SearchCoupons(code, status, size, style string, partnerID *uuid.UUID) ([]*Coupon, error) {
	coupons, err := s.deps.CouponRepository.Search(context.Background(), code, status, size, style, partnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to find coupons: %w", err)
	}
	return coupons, nil
}

// GetCouponByID retrieves coupon by ID
func (s *CouponService) GetCouponByID(id uuid.UUID) (*Coupon, error) {
	coupon, err := s.deps.CouponRepository.GetByID(context.Background(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, fmt.Errorf("not found")
		}
		return nil, fmt.Errorf("failed to get coupon: %w", err)
	}
	return coupon, nil
}

// GetCouponByCode retrieves coupon by code
func (s *CouponService) GetCouponByCode(code string) (*Coupon, error) {
	coupon, err := s.deps.CouponRepository.GetByCode(context.Background(), code)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, fmt.Errorf("not found")
		}
		return nil, fmt.Errorf("failed to get coupon: %w", err)
	}
	return coupon, nil
}

// ActivateCoupon activates coupon (changes status to 'completed')
func (s *CouponService) ActivateCoupon(id uuid.UUID, req ActivateCouponRequest) error {
	if err := s.deps.CouponRepository.ActivateCoupon(context.Background(), id, req); err != nil {
		return fmt.Errorf("failed to activate coupon: %w", err)
	}
	return nil
}

// ResetCoupon resets coupon to initial state
func (s *CouponService) ResetCoupon(id uuid.UUID) error {
	if err := s.deps.CouponRepository.ResetCoupon(context.Background(), id); err != nil {
		return fmt.Errorf("failed to reset coupon: %w", err)
	}
	return nil
}

// SendSchema sends schema to email
func (s *CouponService) SendSchema(id uuid.UUID, email string) error {
	if err := s.deps.CouponRepository.SendSchema(context.Background(), id, email); err != nil {
		return fmt.Errorf("failed to send schema: %w", err)
	}
	return nil
}

// MarkAsPurchased marks coupon as purchased
func (s *CouponService) MarkAsPurchased(id uuid.UUID, purchaseEmail string) error {
	if err := s.deps.CouponRepository.MarkAsPurchased(context.Background(), id, purchaseEmail); err != nil {
		return fmt.Errorf("failed to mark as purchased: %w", err)
	}
	return nil
}

// GetStatistics returns coupon statistics
func (s *CouponService) GetStatistics(partnerID *uuid.UUID) (map[string]int64, error) {
	stats, err := s.deps.CouponRepository.GetStatistics(context.Background(), partnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get statistics: %w", err)
	}
	return stats, nil
}

// ValidateCoupon validates coupon and returns validation response
func (s *CouponService) ValidateCoupon(code string) (*CouponValidationResponse, error) {
	coupon, err := s.deps.CouponRepository.GetByCode(context.Background(), code)
	if err != nil {
		return &CouponValidationResponse{
			Valid:   false,
			Message: "Coupon not found",
		}, nil
	}

	// Get partner information
	partner, err := s.deps.PartnerRepository.GetByID(context.Background(), coupon.PartnerID)
	if err != nil {
		// If failed to get partner, return basic validation
		size := string(coupon.Size)
		style := string(coupon.Style)
		return &CouponValidationResponse{
			Valid:   true,
			Message: "Coupon is valid and ready to use",
			Size:    &size,
			Style:   &style,
		}, nil
	}

	size := string(coupon.Size)
	style := string(coupon.Style)

	response := &CouponValidationResponse{
		Valid:   true,
		Message: "Coupon is valid and ready to use",
		Size:    &size,
		Style:   &style,

		// Partner information for domain validation
		PartnerID:        &partner.ID,
		PartnerCode:      &partner.PartnerCode,
		PartnerDomain:    &partner.Domain,
		PartnerBrandName: &partner.BrandName,
		IsCorrectDomain:  true, // Always true, domain validation is done on frontend
	}

	return response, nil
}

// ExportCoupons exports coupons to text file
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

// DownloadMaterials downloads ZIP archive with coupon materials
func (s *CouponService) DownloadMaterials(id uuid.UUID) ([]byte, string, error) {
	coupon, err := s.deps.CouponRepository.GetByID(context.Background(), id)
	if err != nil {
		return nil, "", fmt.Errorf("coupon not found: %w", err)
	}

	if coupon.Status != "used" && coupon.Status != "completed" {
		return nil, "", fmt.Errorf("coupon must be used or completed to download materials")
	}

	if coupon.ZipURL == nil || *coupon.ZipURL == "" {
		return nil, "", fmt.Errorf("schema archive URL not available")
	}

	resp, err := http.Get(*coupon.ZipURL)
	if err != nil {
		return nil, "", fmt.Errorf("failed to download schema archive: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("failed to download schema archive: status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read schema archive: %w", err)
	}

	filename := fmt.Sprintf("coupon_%s_materials.zip", coupon.Code)
	return data, filename, nil
}

// SearchCouponsWithPagination returns coupons with pagination
func (s *CouponService) SearchCouponsWithPagination(code, status, size, style string, partnerID *uuid.UUID, page, limit int) ([]*Coupon, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	coupons, total, err := s.deps.CouponRepository.SearchWithPagination(
		context.Background(), code, status, size, style, partnerID, page, limit,
		nil, nil, nil, nil, "", "",
	)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to fetch coupons: %w", err)
	}

	return coupons, int64(total), nil
}

// SearchCouponsByPartner returns coupons for specific partner
func (s *CouponService) SearchCouponsByPartner(partnerID uuid.UUID, status, size, style string) ([]*Coupon, error) {
	coupons, err := s.deps.CouponRepository.Search(context.Background(), "", status, size, style, &partnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch partner coupons: %w", err)
	}
	return coupons, nil
}

// BatchResetCoupons performs batch reset of coupons
func (s *CouponService) BatchResetCoupons(couponIDs []string) (*BatchResetResponse, error) {
	response := &BatchResetResponse{
		Success: make([]string, 0),
		Failed:  make([]string, 0),
		Errors:  make([]string, 0),
	}

	if len(couponIDs) == 0 {
		return response, fmt.Errorf("no coupon IDs provided")
	}

	if len(couponIDs) > 1000 {
		return response, fmt.Errorf("too many coupon IDs (maximum 1000)")
	}

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

	response.Failed = append(response.Failed, invalidIDs...)

	if len(validIDs) == 0 {
		response.FailedCount = len(response.Failed)
		return response, nil
	}

	success, failed, err := s.deps.CouponRepository.BatchReset(context.Background(), validIDs)
	if err != nil {
		for _, id := range validIDs {
			response.Failed = append(response.Failed, id.String())
		}
		response.Errors = append(response.Errors, err.Error())
	} else {
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

// PreviewBatchDelete returns preview for batch deletion
func (s *CouponService) PreviewBatchDelete(couponIDs []string) (*BatchDeletePreviewResponse, error) {
	if len(couponIDs) == 0 {
		return nil, fmt.Errorf("no coupon IDs provided")
	}

	if len(couponIDs) > 1000 {
		return nil, fmt.Errorf("too many coupon IDs (maximum 1000)")
	}

	var validIDs []uuid.UUID
	for _, idStr := range couponIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			return nil, fmt.Errorf("invalid UUID format: %s", idStr)
		}
		validIDs = append(validIDs, id)
	}

	previews, err := s.deps.CouponRepository.GetCouponsForDeletion(context.Background(), validIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get coupons preview: %w", err)
	}

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

	confirmationKey := uuid.New().String()
	expiresAt := time.Now().Add(15 * time.Minute)

	response := &BatchDeletePreviewResponse{
		TotalCount:      len(previews),
		UsedCount:       usedCount,
		NewCount:        newCount,
		Coupons:         previews,
		ConfirmationKey: confirmationKey,
		ExpiresAt:       expiresAt,
	}

	confirmationData := map[string]any{
		"action":     "batch_delete",
		"coupon_ids": couponIDs,
		"created_at": time.Now(),
	}

	dataJSON, err := json.Marshal(confirmationData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal confirmation data: %w", err)
	}

	redisKey := fmt.Sprintf("confirmation:%s", confirmationKey)
	err = s.deps.RedisClient.Set(context.Background(), redisKey, dataJSON, 15*time.Minute).Err()
	if err != nil {
		return nil, fmt.Errorf("failed to save confirmation key to Redis: %w", err)
	}

	return response, nil
}

// ExecuteBatchDelete executes batch deletion with confirmation
func (s *CouponService) ExecuteBatchDelete(request BatchDeleteConfirmRequest) (*BatchDeleteResponse, error) {
	response := &BatchDeleteResponse{
		Deleted: make([]string, 0),
		Failed:  make([]string, 0),
		Errors:  make([]string, 0),
	}

	if len(request.CouponIDs) == 0 {
		return response, fmt.Errorf("no coupon IDs provided")
	}

	if len(request.CouponIDs) > 1000 {
		return response, fmt.Errorf("too many coupon IDs (maximum 1000)")
	}

	redisKey := fmt.Sprintf("confirmation:%s", request.ConfirmationKey)
	confirmationDataJSON, err := s.deps.RedisClient.Get(context.Background(), redisKey).Result()
	if err != nil {
		if err == redis.Nil {
			return response, fmt.Errorf("invalid or expired confirmation key")
		}
		return response, fmt.Errorf("failed to check confirmation key: %w", err)
	}

	var confirmationData map[string]any
	if err := json.Unmarshal([]byte(confirmationDataJSON), &confirmationData); err != nil {
		return response, fmt.Errorf("invalid confirmation data")
	}

	if action, ok := confirmationData["action"].(string); !ok || action != "batch_delete" {
		return response, fmt.Errorf("invalid confirmation action")
	}

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

	response.Failed = append(response.Failed, invalidIDs...)

	if len(validIDs) == 0 {
		response.FailedCount = len(response.Failed)
		return response, nil
	}

	var codesBeforeDeletion []string
	coupons, err := s.deps.CouponRepository.GetCouponsForDeletion(context.Background(), validIDs)
	if err == nil {
		for _, coupon := range coupons {
			codesBeforeDeletion = append(codesBeforeDeletion, coupon.Code)
		}
	}

	deletedCount, err := s.deps.CouponRepository.BatchDelete(context.Background(), validIDs)
	if err != nil {
		for _, id := range validIDs {
			response.Failed = append(response.Failed, id.String())
		}
		response.Errors = append(response.Errors, err.Error())
	} else {
		if deletedCount == int64(len(validIDs)) {
			response.Deleted = codesBeforeDeletion
		} else {
			deletedItems := int(deletedCount)
			if deletedItems > len(codesBeforeDeletion) {
				deletedItems = len(codesBeforeDeletion)
			}

			response.Deleted = codesBeforeDeletion[:deletedItems]

			for i := deletedItems; i < len(codesBeforeDeletion); i++ {
				response.Failed = append(response.Failed, codesBeforeDeletion[i])
			}
		}
	}

	response.DeletedCount = len(response.Deleted)
	response.FailedCount = len(response.Failed)

	redisKey = fmt.Sprintf("confirmation:%s", request.ConfirmationKey)
	if err := s.deps.RedisClient.Del(context.Background(), redisKey).Err(); err != nil {
		// log.Error().Err(err).Msg("Failed to delete confirmation key from Redis")
	}

	return response, nil
}

// ExportCouponsAdvanced exports coupons with configurable formats
func (s *CouponService) ExportCouponsAdvanced(options ExportOptionsRequest) ([]byte, string, string, error) {
	data, err := s.deps.CouponRepository.GetCouponsForExport(context.Background(), options)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to get export data: %w", err)
	}

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

	case "txt":
		content, err := s.generateTXT(data, options)
		if err != nil {
			return nil, "", "", fmt.Errorf("failed to generate TXT: %w", err)
		}
		filename = fmt.Sprintf("coupons_export_%s_%s.txt", options.Format, timestamp)
		contentType = "text/plain"
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

// generateTXT creates text file for export
func (s *CouponService) generateTXT(data any, options ExportOptionsRequest) ([]byte, error) {
	var content strings.Builder

	switch options.Format {
	case ExportFormatCodes:
		codes, ok := data.([]string)
		if !ok {
			return nil, fmt.Errorf("invalid data format for codes export")
		}

		if options.IncludeHeader {
			content.WriteString("Coupon Codes\n")
			content.WriteString("============\n\n")
		}

		if options.PartnerID != nil {
			for _, code := range codes {
				content.WriteString(code + "\n")
			}
		} else {
			groupedCodes := make(map[string][]string)

			for _, code := range codes {
				if len(code) >= 4 {
					prefix := code[:4]
					groupedCodes[prefix] = append(groupedCodes[prefix], code)
				} else {
					groupedCodes["UNKN"] = append(groupedCodes["UNKN"], code)
				}
			}

			for prefix, partnerCodes := range groupedCodes {
				if len(partnerCodes) > 0 {
					content.WriteString(fmt.Sprintf("Partner %s:\n", prefix))
					for _, code := range partnerCodes {
						content.WriteString(code + "\n")
					}
					content.WriteString("\n")
				}
			}
		}

	case ExportFormatPartner:
		exports, ok := data.([]PartnerExportRow)
		if !ok {
			return nil, fmt.Errorf("invalid data format for partner export")
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
		exports, ok := data.([]AdminExportRow)
		if !ok {
			return nil, fmt.Errorf("invalid data format for admin export")
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
		exports, ok := data.([]FullExportRow)
		if !ok {
			return nil, fmt.Errorf("invalid data format for full export")
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
		// For other formats use JSON
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal data: %w", err)
		}
		content.Write(jsonData)
	}

	out := content.String()
	if out == "" && options.IncludeHeader {
		// Minimal protection: if nothing found by filters but headers were expected — return empty file with newline
		out = "\n"
	}
	return []byte(out), nil
}

// generateCSV creates CSV file for export
func (s *CouponService) generateCSV(data any, options ExportOptionsRequest) ([]byte, error) {
	var buffer bytes.Buffer
	// Write BOM for better Excel support with CSV
	if options.FileFormat == "csv" {
		buffer.Write([]byte{0xEF, 0xBB, 0xBF})
	}
	writer := csv.NewWriter(&buffer)

	delimiter := ","
	if options.Delimiter != "" {
		delimiter = options.Delimiter
	}
	writer.Comma = rune(delimiter[0])

	switch options.Format {
	case ExportFormatCodes:
		codes, ok := data.([]string)
		if !ok {
			return nil, fmt.Errorf("invalid data format for codes export")
		}

		if options.IncludeHeader {
			writer.Write([]string{"Coupon Code"})
		}

		if options.PartnerID != nil {
			for _, code := range codes {
				writer.Write([]string{code})
			}
		} else {
			groupedCodes := make(map[string][]string)

			for _, code := range codes {
				if len(code) >= 4 {
					prefix := code[:4]
					groupedCodes[prefix] = append(groupedCodes[prefix], code)
				} else {
					groupedCodes["UNKN"] = append(groupedCodes["UNKN"], code)
				}
			}

			keys := make([]string, 0, len(groupedCodes))
			for k := range groupedCodes {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			firstGroup := true
			for _, prefix := range keys {
				partnerCodes := groupedCodes[prefix]
				if len(partnerCodes) > 0 {
					if !firstGroup {
						writer.Write([]string{""})
					}
					firstGroup = false
					writer.Write([]string{fmt.Sprintf("=== Partner %s ===", prefix)})

					for _, code := range partnerCodes {
						writer.Write([]string{code})
					}
				}
			}
		}

	case ExportFormatBasic:
		exports, ok := data.([]BasicExportRow)
		if !ok {
			return nil, fmt.Errorf("invalid data format for basic export")
		}

		if options.IncludeHeader {
			writer.Write([]string{"Code", "Status", "Size", "Style", "Created At"})
		}

		lastPrefix := ""
		for _, export := range exports {
			prefix := ""
			if len(export.Code) >= 4 {
				prefix = export.Code[:4]
			}
			if lastPrefix != "" && prefix != lastPrefix {
				// empty line between partner groups
				writer.Write([]string{""})
			}
			writer.Write([]string{
				export.Code,
				export.Status,
				export.Size,
				export.Style,
				export.CreatedAt.Format(time.RFC3339),
			})
			lastPrefix = prefix
		}

	case ExportFormatAdmin:
		exports, ok := data.([]AdminExportRow)
		if !ok {
			return nil, fmt.Errorf("invalid data format for admin export")
		}

		if options.IncludeHeader {
			writer.Write([]string{"Coupon Code", "Partner ID", "Partner Status", "Coupon Status", "Size", "Style", "Brand Name", "Email", "Created At"})
		}

		lastPartner := ""
		for _, export := range exports {
			if lastPartner != "" && export.PartnerID != lastPartner {
				writer.Write([]string{""})
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
			})
			lastPartner = export.PartnerID
		}

	case ExportFormatPartner:
		exports, ok := data.([]PartnerExportRow)
		if !ok {
			return nil, fmt.Errorf("invalid data format for partner export")
		}

		if options.IncludeHeader {
			writer.Write([]string{"Coupon Code", "Partner Status", "Coupon Status", "Size", "Style", "Created At"})
		}

		lastPrefix := ""
		for _, export := range exports {
			prefix := ""
			if len(export.Code) >= 4 {
				prefix = export.Code[:4]
			}
			if lastPrefix != "" && prefix != lastPrefix {
				writer.Write([]string{""})
			}
			writer.Write([]string{
				export.Code,
				export.PartnerStatus,
				export.CouponStatus,
				export.Size,
				export.Style,
				export.CreatedAt.Format("2006-01-02 15:04:05"),
			})
			lastPrefix = prefix
		}

	case ExportFormatFull:
		exports, ok := data.([]FullExportRow)
		if !ok {
			return nil, fmt.Errorf("invalid data format for full export")
		}

		if options.IncludeHeader {
			writer.Write([]string{"Coupon Code", "Partner ID", "Partner Status", "Coupon Status", "Size", "Style", "Brand Name", "Email", "Created At", "Used At"})
		}

		lastPartner := ""
		for _, export := range exports {
			usedAt := ""
			if export.UsedAt != nil {
				usedAt = export.UsedAt.Format("2006-01-02 15:04:05")
			}
			if lastPartner != "" && export.PartnerID != lastPartner {
				writer.Write([]string{""})
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
			lastPartner = export.PartnerID
		}

	default:
		// For complex formats use reflection for general processing
		return s.generateCSVGeneric(data, options)
	}

	// Explicitly flush buffer before reading Bytes
	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("failed to flush csv writer: %w", err)
	}
	return buffer.Bytes(), nil
}

// generateCSVGeneric universal CSV generator using reflection
func (s *CouponService) generateCSVGeneric(data any, options ExportOptionsRequest) ([]byte, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var buffer bytes.Buffer
	writer := csv.NewWriter(&buffer)

	if options.IncludeHeader {
		writer.Write([]string{"JSON Data"})
	}
	writer.Write([]string{string(jsonData)})

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("failed to flush csv writer: %w", err)
	}
	return buffer.Bytes(), nil
}
