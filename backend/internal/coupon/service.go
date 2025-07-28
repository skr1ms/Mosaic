package coupon

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type CouponServiceDeps struct {
	CouponRepository *CouponRepository
	Logger           *zerolog.Logger
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
		s.deps.Logger.Error().Err(err).Msg(ErrFailedToFindCoupons.Error())
		return nil, ErrFailedToFindCoupons
	}
	return coupons, nil
}

// GetCouponByID получает купон по ID
func (s *CouponService) GetCouponByID(id uuid.UUID) (*Coupon, error) {
	coupon, err := s.deps.CouponRepository.GetByID(context.Background(), id)
	if err != nil {
		s.deps.Logger.Error().Err(err).Msg(ErrCouponNotFound.Message)
		return nil, ErrCouponNotFound
	}
	return coupon, nil
}

// GetCouponByCode получает купон по коду
func (s *CouponService) GetCouponByCode(code string) (*Coupon, error) {
	coupon, err := s.deps.CouponRepository.GetByCode(context.Background(), code)
	if err != nil {
		s.deps.Logger.Error().Err(err).Msg(ErrCouponNotFound.Message)
		return nil, ErrCouponNotFound
	}
	return coupon, nil
}

// ActivateCoupon активирует купон
func (s *CouponService) ActivateCoupon(id uuid.UUID, originalImageURL, previewURL, schemaURL string) error {
	if err := s.deps.CouponRepository.ActivateCoupon(context.Background(), id, originalImageURL, previewURL, schemaURL); err != nil {
		s.deps.Logger.Error().Err(err).Msg(ErrFailedToActivateCoupon.Message)
		return ErrFailedToActivateCoupon
	}
	return nil
}

// ResetCoupon сбрасывает купон в исходное состояние
func (s *CouponService) ResetCoupon(id uuid.UUID) error {
	if err := s.deps.CouponRepository.ResetCoupon(context.Background(), id); err != nil {
		s.deps.Logger.Error().Err(err).Msg(ErrFailedToResetCoupon.Message)
		return ErrFailedToResetCoupon
	}
	return nil
}

// SendSchema отправляет схему на email
func (s *CouponService) SendSchema(id uuid.UUID, email string) error {
	if err := s.deps.CouponRepository.SendSchema(context.Background(), id, email); err != nil {
		s.deps.Logger.Error().Err(err).Msg(ErrFailedToSendSchema.Message)
		return ErrFailedToSendSchema
	}
	return nil
}

// MarkAsPurchased помечает купон как купленный
func (s *CouponService) MarkAsPurchased(id uuid.UUID, purchaseEmail string) error {
	if err := s.deps.CouponRepository.MarkAsPurchased(context.Background(), id, purchaseEmail); err != nil {
		s.deps.Logger.Error().Err(err).Msg(ErrFailedToMarkAsPurchased.Message)
		return ErrFailedToMarkAsPurchased
	}
	return nil
}

// GetStatistics возвращает статистику по купонам
func (s *CouponService) GetStatistics(partnerID *uuid.UUID) (map[string]int64, error) {
	stats, err := s.deps.CouponRepository.GetStatistics(context.Background(), partnerID)
	if err != nil {
		s.deps.Logger.Error().Err(err).Msg(ErrFailedToGetStatistics.Message)
		return nil, ErrFailedToGetStatistics
	}
	return stats, nil
}

// ValidateCoupon проверяет валидность купона
func (s *CouponService) ValidateCoupon(code string) (*CouponValidationResponse, error) {
	coupon, err := s.deps.CouponRepository.GetByCode(context.Background(), code)
	if err != nil {
		s.deps.Logger.Error().Err(err).Msg(ErrCouponNotFound.Message)
		return &CouponValidationResponse{
			Valid:   false,
			Message: ErrCouponNotFound.Message,
		}, nil
	}

	if coupon.Status == "used" {
		s.deps.Logger.Error().Msg(ErrCouponAlreadyUsed.Message)
		size := string(coupon.Size)
		style := string(coupon.Style)
		return &CouponValidationResponse{
			Valid:   false,
			Message: ErrCouponAlreadyUsed.Message,
			UsedAt:  coupon.UsedAt,
			Size:    &size,
			Style:   &style,
		}, nil
	}

	s.deps.Logger.Info().Msg("Coupon is valid and ready to use")
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
		s.deps.Logger.Error().Err(err).Msg(ErrFailedToFetchCouponsForExport.Message)
		return "", ErrFailedToFetchCouponsForExport
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
		s.deps.Logger.Error().Err(err).Msg(ErrCouponNotFound.Message)
		return nil, "", ErrCouponNotFound
	}

	if coupon.Status != "used" {
		s.deps.Logger.Error().Msg(ErrCouponMustBeUsedToDownloadMaterials.Message)
		return nil, "", ErrCouponMustBeUsedToDownloadMaterials
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
			s.deps.Logger.Error().Err(err).Msg(ErrFailedToDownloadFile.Message)
			return ErrFailedToDownloadFile
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			s.deps.Logger.Error().Msg(ErrFailedToDownloadFile.Message)
			return ErrFailedToDownloadFile
		}

		fileWriter, err := zipWriter.Create(filename)
		if err != nil {
			s.deps.Logger.Error().Err(err).Msg(ErrFailedToCreateFileWriter.Message)
			return ErrFailedToCreateFileWriter
		}

		_, err = io.Copy(fileWriter, resp.Body)
		if err != nil {
			s.deps.Logger.Error().Err(err).Msg(ErrFailedToCopyFileToZip.Message)
			return ErrFailedToCopyFileToZip
		}

		return nil
	}

	// Добавляем файлы в архив
	if coupon.OriginalImageURL != nil && *coupon.OriginalImageURL != "" {
		if err := addFileToZip(*coupon.OriginalImageURL, "original_image.jpg"); err != nil {
			s.deps.Logger.Error().Err(err).Msg("Error adding original image to zip")
		}
	}

	if coupon.PreviewURL != nil && *coupon.PreviewURL != "" {
		if err := addFileToZip(*coupon.PreviewURL, "preview.jpg"); err != nil {
			s.deps.Logger.Error().Err(err).Msg("Error adding preview to zip")
		}
	}

	if coupon.SchemaURL != nil && *coupon.SchemaURL != "" {
		if err := addFileToZip(*coupon.SchemaURL, "schema.pdf"); err != nil {
			s.deps.Logger.Error().Err(err).Msg("Error adding schema to zip")
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
		s.deps.Logger.Error().Err(err).Msg(ErrFailedToCreateArchive.Message)
		return nil, "", ErrFailedToCreateArchive
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
		s.deps.Logger.Error().Err(err).Msg(ErrFailedToFetchCoupons.Message)
		return nil, 0, ErrFailedToFetchCoupons
	}

	return coupons, int64(total), nil
}

// SearchCouponsByPartner возвращает купоны конкретного партнера
func (s *CouponService) SearchCouponsByPartner(partnerID uuid.UUID, status, size, style string) ([]*Coupon, error) {
	coupons, err := s.deps.CouponRepository.Search(context.Background(), "", status, size, style, &partnerID)
	if err != nil {
		s.deps.Logger.Error().Err(err).Msg(ErrFailedToFetchPartnerCoupons.Message)
		return nil, ErrFailedToFetchPartnerCoupons
	}
	return coupons, nil
}
