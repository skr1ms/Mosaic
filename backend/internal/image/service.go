package image

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/csv"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/skr1ms/mosaic/internal/types"
	"github.com/skr1ms/mosaic/pkg/mosaic"
	"github.com/skr1ms/mosaic/pkg/palette"
	"github.com/skr1ms/mosaic/pkg/stableDiffusion"
	"github.com/skr1ms/mosaic/pkg/zip"
)

type ImageServiceDeps struct {
	ImageRepository       ImageRepositoryInterface
	CouponRepository      CouponRepositoryInterface
	S3Client              S3ClientInterface
	StableDiffusionClient StableDiffusionClientInterface
	EmailService          EmailServiceInterface
	ZipService            ZipServiceInterface
	MosaicGenerator       MosaicGeneratorInterface
	PaletteService        *palette.PaletteService
	WorkingDir            string
}

type ImageService struct {
	deps *ImageServiceDeps
}

func NewImageService(deps *ImageServiceDeps) *ImageService {
	s := &ImageService{
		deps: deps,
	}

	go func() {
		s.startRetentionCleaner()
	}()

	return s
}

// Repository access methods
func (s *ImageService) GetCouponRepository() CouponRepositoryInterface {
	return s.deps.CouponRepository
}

func (s *ImageService) GetS3Client() S3ClientInterface {
	return s.deps.S3Client
}

func (s *ImageService) GetImageRepository() ImageRepositoryInterface {
	return s.deps.ImageRepository
}

// GetQueue returns queue tasks with status filtering
func (s *ImageService) GetQueue(status string) ([]*ImageWithPartner, error) {
	var tasks []*ImageWithPartner
	var err error

	if status != "" {
		tasks, err = s.deps.ImageRepository.GetByStatusWithPartner(context.Background(), status)
	} else {
		tasks, err = s.deps.ImageRepository.GetAllWithPartner(context.Background())
	}

	if err != nil {
		return nil, fmt.Errorf("failed to fetch queue: %w", err)
	}

	return tasks, nil
}

// GetQueueWithFilters returns queue tasks with status and date filtering
func (s *ImageService) GetQueueWithFilters(status, dateFrom, dateTo string) ([]*ImageWithPartner, error) {
	tasks, err := s.deps.ImageRepository.GetWithFilters(context.Background(), status, dateFrom, dateTo)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch queue with filters: %w", err)
	}

	return tasks, nil
}

// UploadImage uploads and saves image to S3
func (s *ImageService) UploadImage(ctx context.Context, couponID uuid.UUID, file *multipart.FileHeader, userEmail string) (*Image, error) {
	_, err := s.deps.CouponRepository.GetByID(ctx, couponID)
	if err != nil {
		return nil, fmt.Errorf("coupon not found: %w", err)
	}

	existingImage, err := s.deps.ImageRepository.GetByCouponID(ctx, couponID)
	if err == nil && existingImage != nil {
		// Allow image reload if it's not being processed
		if existingImage.Status == "processing" || existingImage.Status == "completed" {
			return nil, fmt.Errorf("image already uploaded for this coupon and is being processed")
		}
		// Delete old image for reload
		if err := s.deps.ImageRepository.Delete(ctx, existingImage.ID); err != nil {
			log.Warn().Err(err).Str("image_id", existingImage.ID.String()).Msg("Failed to delete old image record")
		}
		s.cleanupLocalFiles(couponID)
	}

	if !isValidImageType(file) {
		return nil, fmt.Errorf("invalid image type, supported: JPG, PNG")
	}

	if file.Size > 15<<20 {
		return nil, fmt.Errorf("file too large, maximum size is 15MB")
	}

	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	uploadsDir := filepath.Join(s.deps.WorkingDir, "uploads", couponID.String())
	if err := os.MkdirAll(uploadsDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create uploads dir: %w", err)
	}

	ext := ".jpg"
	if ct := file.Header.Get("Content-Type"); ct == "image/png" {
		ext = ".png"
	}
	localPath := filepath.Join(uploadsDir, fmt.Sprintf("%d%s", time.Now().Unix(), ext))
	dst, err := os.Create(localPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create local file: %w", err)
	}
	if _, err := io.Copy(dst, src); err != nil {
		dst.Close()
		return nil, fmt.Errorf("failed to write local file: %w", err)
	}
	dst.Close()

	imageRecord := &Image{
		CouponID:           couponID,
		OriginalImageS3Key: "file://" + localPath,
		UserEmail:          userEmail,
		Status:             "uploaded",
		Priority:           1,
	}

	if err := s.deps.ImageRepository.Create(ctx, imageRecord); err != nil {
		return nil, fmt.Errorf("failed to create image record: %w", err)
	}

	log.Info().
		Str("image_id", imageRecord.ID.String()).
		Str("coupon_id", couponID.String()).
		Str("s3_key", imageRecord.OriginalImageS3Key).
		Msg("Image uploaded successfully")

	return imageRecord, nil
}

// EditImage applies editing to image
func (s *ImageService) EditImage(ctx context.Context, imageID uuid.UUID, editParams ImageEditParams) error {
	imageRecord, err := s.deps.ImageRepository.GetByID(ctx, imageID)
	if err != nil {
		return fmt.Errorf("image not found: %w", err)
	}

	if imageRecord.Status != "uploaded" {
		return fmt.Errorf("image cannot be edited in current status: %s", imageRecord.Status)
	}

	originalReader, err := s.openFromStorage(ctx, imageRecord.OriginalImageS3Key)
	if err != nil {
		return fmt.Errorf("failed to download original image: %w", err)
	}
	defer originalReader.Close()

	img, format, err := image.Decode(originalReader)
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	editedImg := s.applyImageEditing(img, editParams)

	var buf bytes.Buffer
	switch format {
	case "jpeg":
		err = jpeg.Encode(&buf, editedImg, &jpeg.Options{Quality: 95})
	case "png":
		err = png.Encode(&buf, editedImg)
	default:
		err = jpeg.Encode(&buf, editedImg, &jpeg.Options{Quality: 95})
	}
	if err != nil {
		return fmt.Errorf("failed to encode edited image: %w", err)
	}

	editedDir := filepath.Join(s.deps.WorkingDir, "edited", imageRecord.CouponID.String())
	if err := os.MkdirAll(editedDir, 0o755); err != nil {
		return fmt.Errorf("failed to create edited dir: %w", err)
	}
	editedPath := filepath.Join(editedDir, fmt.Sprintf("%d.jpg", time.Now().Unix()))
	if err := os.WriteFile(editedPath, buf.Bytes(), 0o644); err != nil {
		return fmt.Errorf("failed to save edited image: %w", err)
	}

	fileEdited := "file://" + editedPath
	imageRecord.EditedImageS3Key = &fileEdited
	imageRecord.Status = "edited"
	if err := s.deps.ImageRepository.Update(ctx, imageRecord); err != nil {
		return fmt.Errorf("failed to update image record: %w", err)
	}

	log.Info().
		Str("image_id", imageID.String()).
		Str("edited_s3_key", fileEdited).
		Msg("Image edited successfully")

	return nil
}

// ProcessImage processes image through Stable Diffusion
func (s *ImageService) ProcessImage(ctx context.Context, imageID uuid.UUID, processParams *ProcessingParams) error {
	imageRecord, err := s.deps.ImageRepository.GetByID(ctx, imageID)
	if err != nil {
		return fmt.Errorf("image not found: %w", err)
	}

	if imageRecord.Status != "edited" && imageRecord.Status != "uploaded" {
		return fmt.Errorf("image cannot be processed in current status: %s", imageRecord.Status)
	}

	sourceS3Key := imageRecord.OriginalImageS3Key
	if imageRecord.EditedImageS3Key != nil {
		sourceS3Key = *imageRecord.EditedImageS3Key
	}

	coupon, err := s.deps.CouponRepository.GetByID(ctx, imageRecord.CouponID)
	if err != nil {
		return fmt.Errorf("failed to get coupon: %w", err)
	}

	width, height := parseCouponSize(coupon.Size)

	imageRecord.Status = "processing"
	imageRecord.ProcessingParams = processParams
	now := time.Now()
	imageRecord.StartedAt = &now
	if err := s.deps.ImageRepository.Update(ctx, imageRecord); err != nil {
		return fmt.Errorf("failed to update status to processing: %w", err)
	}

	if !processParams.UseAI {
		return s.createPreviewWithoutAI(ctx, imageRecord, sourceS3Key)
	}

	if err := s.deps.StableDiffusionClient.CheckHealth(ctx); err != nil {
		log.Warn().
			Err(err).
			Str("image_id", imageID.String()).
			Msg("Stable Diffusion API health check failed, but continuing with processing")
	}

	sourceReader, err := s.openFromStorage(ctx, sourceS3Key)
	if err != nil {
		return s.markProcessingFailed(ctx, imageRecord, fmt.Errorf("failed to download source image: %w", err))
	}
	defer sourceReader.Close()

	sourceData, err := io.ReadAll(sourceReader)
	if err != nil {
		return s.markProcessingFailed(ctx, imageRecord, fmt.Errorf("failed to read source image: %w", err))
	}

	base64Image := s.deps.StableDiffusionClient.EncodeImageToBase64(sourceData)

	sdRequest := stableDiffusion.ProcessImageRequest{
		ImageBase64: base64Image,
		Style:       stableDiffusion.ProcessingStyle(processParams.Style),
		UseAI:       processParams.UseAI,
		Lighting:    stableDiffusion.LightingType(processParams.Lighting),
		Contrast:    stableDiffusion.ContrastLevel(processParams.Contrast),
		Brightness:  processParams.Brightness,
		Saturation:  processParams.Saturation,
		Width:       width,
		Height:      height,
	}

	log.Info().
		Str("image_id", imageID.String()).
		Str("style", processParams.Style).
		Bool("use_ai", processParams.UseAI).
		Str("lighting", processParams.Lighting).
		Str("contrast", processParams.Contrast).
		Float64("brightness", processParams.Brightness).
		Float64("saturation", processParams.Saturation).
		Int("width", width).
		Int("height", height).
		Msg("Starting Stable Diffusion processing")

	processedBase64, err := s.deps.StableDiffusionClient.ProcessImage(ctx, sdRequest)
	if err != nil {
		return s.markProcessingFailed(ctx, imageRecord, fmt.Errorf("stable diffusion processing failed: %w", err))
	}

	processedData, err := s.deps.StableDiffusionClient.DecodeBase64Image(processedBase64)
	if err != nil {
		return s.markProcessingFailed(ctx, imageRecord, fmt.Errorf("failed to decode processed image: %w", err))
	}

	log.Info().
		Str("image_id", imageID.String()).
		Int("processed_data_size", len(processedData)).
		Msg("Stable Diffusion processing completed, uploading result")

	processedDir := filepath.Join(s.deps.WorkingDir, "processed", imageRecord.CouponID.String())
	if err := os.MkdirAll(processedDir, 0o755); err != nil {
		return s.markProcessingFailed(ctx, imageRecord, fmt.Errorf("failed to create processed dir: %w", err))
	}
	processedPath := filepath.Join(processedDir, fmt.Sprintf("%d.jpg", time.Now().Unix()))
	if err := os.WriteFile(processedPath, processedData, 0o644); err != nil {
		return s.markProcessingFailed(ctx, imageRecord, fmt.Errorf("failed to save processed image: %w", err))
	}

	previewS3Key, err := s.createPreview(processedData, imageRecord.CouponID)
	if err != nil {
		return s.markProcessingFailed(ctx, imageRecord, fmt.Errorf("failed to create preview: %w", err))
	}

	processedKey := "file://" + processedPath
	imageRecord.ProcessedImageS3Key = &processedKey
	imageRecord.PreviewS3Key = &previewS3Key
	imageRecord.Status = "processed"
	if err := s.deps.ImageRepository.Update(ctx, imageRecord); err != nil {
		return fmt.Errorf("failed to update image record: %w", err)
	}

	log.Info().
		Str("image_id", imageID.String()).
		Str("processed_s3_key", processedKey).
		Str("preview_s3_key", previewS3Key).
		Msg("Image processed successfully")

	return nil
}

// GenerateSchema creates final diamond art schema
func (s *ImageService) GenerateSchema(ctx context.Context, imageID uuid.UUID, confirmed bool) error {
	if !confirmed {
		return fmt.Errorf("schema generation not confirmed")
	}

	imageRecord, err := s.deps.ImageRepository.GetByID(ctx, imageID)
	if err != nil {
		return fmt.Errorf("image not found: %w", err)
	}

	if imageRecord.Status != "processed" {
		return fmt.Errorf("image must be processed before generating schema")
	}

	sourceS3Key := imageRecord.OriginalImageS3Key
	if imageRecord.ProcessedImageS3Key != nil {
		sourceS3Key = *imageRecord.ProcessedImageS3Key
	} else if imageRecord.EditedImageS3Key != nil {
		sourceS3Key = *imageRecord.EditedImageS3Key
	}

	schemaS3Key, err := s.createSchemaZipArchive(ctx, imageRecord, sourceS3Key)
	if err != nil {
		return s.markProcessingFailed(ctx, imageRecord, fmt.Errorf("failed to create schema ZIP archive: %w", err))
	}

	imageRecord.SchemaS3Key = &schemaS3Key
	imageRecord.Status = "completed"
	now := time.Now()
	imageRecord.CompletedAt = &now
	if err := s.deps.ImageRepository.Update(ctx, imageRecord); err != nil {
		return fmt.Errorf("failed to update image record: %w", err)
	}

	coupon, err := s.deps.CouponRepository.GetByID(ctx, imageRecord.CouponID)
	if err == nil && coupon != nil {
		coupon.Status = "completed"
		coupon.CompletedAt = &now
		if err := s.deps.CouponRepository.Update(ctx, coupon); err != nil {
			log.Error().Err(err).Str("coupon_id", imageRecord.CouponID.String()).Msg("Failed to update coupon status to completed after schema generation")
		}
	}

	log.Info().
		Str("image_id", imageID.String()).
		Str("schema_s3_key", schemaS3Key).
		Msg("Schema generated successfully")

	s.cleanupLocalFilesAsync(imageRecord.CouponID)

	s.sendSchemaEmailAsync(imageRecord, schemaS3Key)

	return nil
}

// GetImageStatus returns image processing status
func (s *ImageService) GetImageStatus(ctx context.Context, imageID uuid.UUID) (*types.ImageStatusResponse, error) {
	imageRecord, err := s.deps.ImageRepository.GetByID(ctx, imageID)
	if err != nil {
		return nil, fmt.Errorf("image not found: %w", err)
	}

	response := &types.ImageStatusResponse{
		ImageID:  imageRecord.ID,
		Status:   imageRecord.Status,
		Message:  s.getStatusMessage(imageRecord.Status),
		Progress: s.calculateProgress(imageRecord.Status),
	}

	// Add error if exists
	if imageRecord.ErrorMessage != nil {
		response.ErrorMessage = imageRecord.ErrorMessage
	}

	// Generate URLs/paths for file access
	if strings.HasPrefix(imageRecord.OriginalImageS3Key, "file://") {
		path := strings.TrimPrefix(imageRecord.OriginalImageS3Key, "file://")
		if u, err := s.buildDataURLFromLocalPath(path); err == nil {
			response.OriginalURL = u
		}
	} else if url, err := s.deps.S3Client.GetFileURL(ctx, imageRecord.OriginalImageS3Key, 24*time.Hour); err == nil {
		response.OriginalURL = &url
	}

	if imageRecord.EditedImageS3Key != nil {
		if strings.HasPrefix(*imageRecord.EditedImageS3Key, "file://") {
			path := strings.TrimPrefix(*imageRecord.EditedImageS3Key, "file://")
			if u, err := s.buildDataURLFromLocalPath(path); err == nil {
				response.EditedURL = u
			}
		} else if url, err := s.deps.S3Client.GetFileURL(ctx, *imageRecord.EditedImageS3Key, 24*time.Hour); err == nil {
			response.EditedURL = &url
		}
	}

	if imageRecord.ProcessedImageS3Key != nil {
		if strings.HasPrefix(*imageRecord.ProcessedImageS3Key, "file://") {
			path := strings.TrimPrefix(*imageRecord.ProcessedImageS3Key, "file://")
			if u, err := s.buildDataURLFromLocalPath(path); err == nil {
				response.ProcessedURL = u
			}
		} else if url, err := s.deps.S3Client.GetFileURL(ctx, *imageRecord.ProcessedImageS3Key, 24*time.Hour); err == nil {
			response.ProcessedURL = &url
		}
	}

	if imageRecord.PreviewS3Key != nil {
		if strings.HasPrefix(*imageRecord.PreviewS3Key, "file://") {
			path := strings.TrimPrefix(*imageRecord.PreviewS3Key, "file://")
			if u, err := s.buildDataURLFromLocalPath(path); err == nil {
				response.PreviewURL = u
			}
		} else if url, err := s.deps.S3Client.GetFileURL(ctx, *imageRecord.PreviewS3Key, 24*time.Hour); err == nil {
			response.PreviewURL = &url
		}
	}

	if imageRecord.SchemaS3Key != nil {
		if url, err := s.deps.S3Client.GetFileURL(ctx, *imageRecord.SchemaS3Key, 24*time.Hour); err == nil {
			response.ZipURL = &url
		}
	}

	return response, nil
}

func (s *ImageService) applyImageEditing(img image.Image, params ImageEditParams) image.Image {
	// Apply scaling
	if params.Scale != 1.0 {
		newWidth := int(float64(img.Bounds().Dx()) * params.Scale)
		newHeight := int(float64(img.Bounds().Dy()) * params.Scale)
		img = imaging.Resize(img, newWidth, newHeight, imaging.Lanczos)
	}

	// Apply rotation
	switch params.Rotation {
	case 90:
		img = imaging.Rotate90(img)
	case 180:
		img = imaging.Rotate180(img)
	case 270:
		img = imaging.Rotate270(img)
	}

	// Apply cropping
	if params.CropWidth > 0 && params.CropHeight > 0 {
		cropRect := image.Rect(params.CropX, params.CropY, params.CropX+params.CropWidth, params.CropY+params.CropHeight)
		img = imaging.Crop(img, cropRect)
	}

	return img
}

func (s *ImageService) createPreviewWithoutAI(ctx context.Context, imageRecord *Image, sourceS3Key string) error {
	sourceReader, err := s.openFromStorage(ctx, sourceS3Key)
	if err != nil {
		return s.markProcessingFailed(ctx, imageRecord, fmt.Errorf("failed to download source image: %w", err))
	}
	defer sourceReader.Close()

	sourceData, err := io.ReadAll(sourceReader)
	if err != nil {
		return s.markProcessingFailed(ctx, imageRecord, fmt.Errorf("failed to read source image: %w", err))
	}

	previewS3Key, err := s.createPreview(sourceData, imageRecord.CouponID)
	if err != nil {
		return s.markProcessingFailed(ctx, imageRecord, fmt.Errorf("failed to create preview: %w", err))
	}

	imageRecord.PreviewS3Key = &previewS3Key
	imageRecord.Status = "processed"
	if err := s.deps.ImageRepository.Update(ctx, imageRecord); err != nil {
		return fmt.Errorf("failed to update image record: %w", err)
	}

	return nil
}

func (s *ImageService) createPreview(imageData []byte, couponID uuid.UUID) (string, error) {
	// Decode image
	img, _, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return "", fmt.Errorf("failed to decode image for preview: %w", err)
	}

	// Create preview with size 400x300
	preview := imaging.Resize(img, 400, 300, imaging.Lanczos)

	// Encode to JPEG
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, preview, &jpeg.Options{Quality: 85}); err != nil {
		return "", fmt.Errorf("failed to encode preview: %w", err)
	}

	// Save locally
	previewsDir := filepath.Join(s.deps.WorkingDir, "previews", couponID.String())
	if err := os.MkdirAll(previewsDir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create previews dir: %w", err)
	}
	previewPath := filepath.Join(previewsDir, fmt.Sprintf("%d.jpg", time.Now().Unix()))
	if err := os.WriteFile(previewPath, buf.Bytes(), 0o644); err != nil {
		return "", fmt.Errorf("failed to save preview: %w", err)
	}
	return "file://" + previewPath, nil
}

// createSchemaZipArchive creates ZIP archive with diamond art schema files
func (s *ImageService) createSchemaZipArchive(ctx context.Context, imageRecord *Image, sourceS3Key string) (string, error) {
	files := []zip.FileData{}

	originalReader, err := s.openFromStorage(ctx, sourceS3Key)
	if err != nil {
		return "", fmt.Errorf("failed to download original image: %w", err)
	}
	defer originalReader.Close()

	originalData, err := io.ReadAll(originalReader)
	if err != nil {
		return "", fmt.Errorf("failed to read original image data: %w", err)
	}

	files = append(files, zip.FileData{
		Name:    "original.jpg",
		Content: bytes.NewReader(originalData),
		Size:    int64(len(originalData)),
	})

	mosaicFiles, _, err := s.generateMosaicFiles(ctx, sourceS3Key, imageRecord)
	if err != nil {
		return "", fmt.Errorf("failed to generate mosaic files: %w", err)
	}

	for i := range mosaicFiles {
		file := mosaicFiles[i]
		if file.Name == "mosaic_preview.png" {
			file.Name = "preview.png"
		}
		if file.Name == "mosaic_legend.csv" {
			continue
		}
		files = append(files, file)
	}

	zipBuffer, err := s.deps.ZipService.CreateSchemaArchive(imageRecord.ID, files)
	if err != nil {
		return "", fmt.Errorf("failed to create ZIP archive: %w", err)
	}

	uploadedKey, err := s.deps.S3Client.UploadFile(ctx, zipBuffer, int64(zipBuffer.Len()), "application/zip", "schemas", imageRecord.ID)
	if err != nil {
		return "", fmt.Errorf("failed to upload ZIP archive to S3: %w", err)
	}

	log.Info().
		Str("image_id", imageRecord.ID.String()).
		Str("zip_s3_key", uploadedKey).
		Int("files_count", len(files)).
		Msg("Schema ZIP archive created successfully")

	return uploadedKey, nil
}

func (s *ImageService) markProcessingFailed(ctx context.Context, imageRecord *Image, err error) error {
	imageRecord.Status = "failed"
	errorMsg := err.Error()
	imageRecord.ErrorMessage = &errorMsg
	imageRecord.RetryCount++

	if updateErr := s.deps.ImageRepository.Update(ctx, imageRecord); updateErr != nil {
		log.Error().Err(updateErr).Msg("Failed to update image record with error status")
	}

	log.Error().
		Err(err).
		Str("image_id", imageRecord.ID.String()).
		Msg("Image processing failed")

	return err
}

func (s *ImageService) getStatusMessage(status string) string {
	switch status {
	case "uploaded":
		return "Изображение загружено, готово к редактированию"
	case "edited":
		return "Изображение отредактировано, готово к обработке"
	case "processing":
		return "Изображение обрабатывается..."
	case "processed":
		return "Изображение обработано, готово к созданию схемы"
	case "completed":
		return "Схема алмазной мозаики создана"
	case "failed":
		return "Произошла ошибка при обработке"
	default:
		return "Неизвестный статус"
	}
}

func (s *ImageService) calculateProgress(status string) int {
	switch status {
	case "uploaded":
		return 20
	case "edited":
		return 40
	case "processing":
		return 60
	case "processed":
		return 80
	case "completed":
		return 100
	case "failed":
		return 0
	default:
		return 0
	}
}

func isValidImageType(file *multipart.FileHeader) bool {
	contentType := file.Header.Get("Content-Type")
	return contentType == "image/jpeg" || contentType == "image/png"
}

func parseCouponSize(size string) (width, height int) {
	if size == "" {
		size = "30x40" // Default value
	}

	switch size {
	case "21x30":
		return 840, 1200
	case "30x40":
		return 1200, 1600
	case "40x40":
		return 1600, 1600
	case "40x50":
		return 1600, 2000
	case "40x60":
		return 1600, 2400
	case "50x70":
		return 2000, 2800
	default:
		return 1200, 1600
	}
}

// generateMosaicFiles generates mosaic files using Python script
func (s *ImageService) generateMosaicFiles(ctx context.Context, sourceS3Key string, imageRecord *Image) ([]zip.FileData, string, error) {
	coupon, err := s.deps.CouponRepository.GetByID(ctx, imageRecord.CouponID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get coupon: %w", err)
	}

	imageReader, err := s.openFromStorage(ctx, sourceS3Key)
	if err != nil {
		return nil, "", fmt.Errorf("failed to download source image: %w", err)
	}
	defer imageReader.Close()

	tempImageFile, err := os.CreateTemp("", "mosaic_input_*.jpg")
	if err != nil {
		return nil, "", fmt.Errorf("failed to create temp image file: %w", err)
	}
	defer os.Remove(tempImageFile.Name())
	defer tempImageFile.Close()

	_, err = io.Copy(tempImageFile, imageReader)
	if err != nil {
		return nil, "", fmt.Errorf("failed to copy image to temp file: %w", err)
	}
	tempImageFile.Close()

	if coupon.Size == "" {
		coupon.Size = "30x40"
	}
	if coupon.Style == "" {
		coupon.Style = "max_colors"
	}

	stonesX, stonesY := parseCouponSize(coupon.Size)

	stonesX = stonesX / 4
	stonesY = stonesY / 4

	paletteStyle := s.mapCouponStyleToPaletteStyle(coupon.Style)
	palettePath, err := s.deps.PaletteService.GetPalettePath(paletteStyle)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get palette path for style %s: %w", coupon.Style, err)
	}

	log.Info().
		Str("coupon_style", coupon.Style).
		Str("palette_style", string(paletteStyle)).
		Str("palette_path", palettePath).
		Msg("Using palette for mosaic generation")

	req := &mosaic.GenerationRequest{
		ImagePath:   tempImageFile.Name(),
		StonesX:     stonesX,
		StonesY:     stonesY,
		StoneSizeMM: 2.52,
		DPI:         150,
		PreviewDPI:  120,
		SchemeDPI:   150,
		Mode:        "both",
		Style:       s.mapCouponStyleToMosaicStyle(coupon.Style),
		WithLegend:  true,
		Threads:     4,
		PalettePath: palettePath,
	}

	result, err := s.deps.MosaicGenerator.Generate(ctx, req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate mosaic: %w", err)
	}

	var files []zip.FileData

	if result.PreviewPath != "" {
		previewData, err := os.ReadFile(result.PreviewPath)
		if err != nil {
			return nil, "", fmt.Errorf("failed to read preview file: %w", err)
		}
		files = append(files, zip.FileData{
			Name:    "mosaic_preview.png",
			Content: bytes.NewReader(previewData),
			Size:    int64(len(previewData)),
		})
		log.Info().
			Str("image_id", imageRecord.ID.String()).
			Str("preview_path", result.PreviewPath).
			Int("preview_size", len(previewData)).
			Msg("Added preview file to archive")
	}

	if result.SchemePath != "" {
		schemeData, err := os.ReadFile(result.SchemePath)
		if err != nil {
			return nil, "", fmt.Errorf("failed to read scheme file: %w", err)
		}
		files = append(files, zip.FileData{
			Name:    "mosaic_scheme.png",
			Content: bytes.NewReader(schemeData),
			Size:    int64(len(schemeData)),
		})
		log.Info().
			Str("image_id", imageRecord.ID.String()).
			Str("scheme_path", result.SchemePath).
			Int("scheme_size", len(schemeData)).
			Msg("Added scheme file to archive")
	} else {
		log.Warn().
			Str("image_id", imageRecord.ID.String()).
			Msg("No scheme path in result")
	}

	if result.LegendPath != "" {
		legendData, err := os.ReadFile(result.LegendPath)
		if err != nil {
			return nil, "", fmt.Errorf("failed to read legend file: %w", err)
		}
		files = append(files, zip.FileData{
			Name:    "mosaic_legend.csv",
			Content: bytes.NewReader(legendData),
			Size:    int64(len(legendData)),
		})

		// Parse CSV to get total stones count
		stonesCount, err := s.parseStonesCountFromCSV(result.LegendPath)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to parse stones count from CSV")
		} else {
			// Update coupon with stones count
			coupon.StonesCount = &stonesCount
			if err := s.deps.CouponRepository.Update(ctx, coupon); err != nil {
				log.Error().Err(err).
					Str("coupon_id", coupon.ID.String()).
					Int("stones_count", stonesCount).
					Msg("Failed to update coupon with stones count")
			} else {
				log.Info().
					Str("coupon_id", coupon.ID.String()).
					Int("stones_count", stonesCount).
					Msg("Updated coupon with stones count")
			}
		}
	}

	log.Info().
		Str("image_id", imageRecord.ID.String()).
		Int("files_count", len(files)).
		Msg("Mosaic files generated successfully")

	return files, result.SchemaUUID, nil
}

// parseStonesCountFromCSV parses the legend CSV file and returns total stones count
func (s *ImageService) parseStonesCountFromCSV(csvPath string) (int, error) {
	file, err := os.Open(csvPath)
	if err != nil {
		return 0, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ';' // The Python script uses semicolon as delimiter

	// Read all records
	records, err := reader.ReadAll()
	if err != nil {
		return 0, fmt.Errorf("failed to read CSV: %w", err)
	}

	// Skip header row and sum up all Count values
	totalStones := 0
	for i, record := range records {
		if i == 0 { // Skip header
			continue
		}

		// Count is in the 3rd column (index 2)
		if len(record) > 2 {
			count, err := strconv.Atoi(record[2])
			if err != nil {
				log.Warn().Err(err).Str("count_value", record[2]).Msg("Failed to parse count value")
				continue
			}
			totalStones += count
		}
	}

	return totalStones, nil
}

// mapCouponStyleToMosaicStyle converts coupon style to mosaic style
func (s *ImageService) mapCouponStyleToMosaicStyle(couponStyle string) string {
	if couponStyle == "" {
		couponStyle = "max_colors"
	}

	switch couponStyle {
	case "grayscale":
		return "soft"
	case "skin_tones":
		return "soft"
	case "pop_art":
		return "contrast"
	case "max_colors":
		return "glossy-dark"
	default:
		return "default"
	}
}

// mapCouponStyleToPaletteStyle converts coupon style to palette style
func (s *ImageService) mapCouponStyleToPaletteStyle(couponStyle string) palette.Style {
	if couponStyle == "" {
		couponStyle = "max_colors"
	}

	switch couponStyle {
	case "grayscale":
		return palette.StyleGrayscale
	case "skin_tones":
		return palette.StyleSkinTones
	case "pop_art":
		return palette.StylePopArt
	case "max_colors":
		return palette.StyleMaxColors
	default:
		return palette.StyleMaxColors
	}
}

// cleanupLocalFilesAsync removes local files asynchronously
func (s *ImageService) cleanupLocalFilesAsync(couponID uuid.UUID) {
	go func() {
		s.cleanupLocalFiles(couponID)
	}()
}

// sendSchemaEmailAsync sends email to user with ready schema asynchronously
func (s *ImageService) sendSchemaEmailAsync(imageRecord *Image, schemaS3Key string) {
	go func() {
		coupon, err := s.deps.CouponRepository.GetByID(context.Background(), imageRecord.CouponID)
		if err != nil {
			log.Error().Err(err).Str("coupon_id", imageRecord.CouponID.String()).Msg("Failed to get coupon for email sending")
			return
		}

		if coupon.UserEmail == nil || *coupon.UserEmail == "" {
			log.Warn().Str("coupon_id", imageRecord.CouponID.String()).Msg("No email address for coupon, skipping email sending")
			return
		}

		schemaURL, err := s.deps.S3Client.GetFileURL(context.Background(), schemaS3Key, 7*24*time.Hour)
		if err != nil {
			log.Error().Err(err).Str("schema_s3_key", schemaS3Key).Msg("Failed to generate presigned URL for schema")
			return
		}

		err = s.deps.EmailService.SendSchemaEmail(*coupon.UserEmail, schemaURL, coupon.Code)
		if err != nil {
			log.Error().Err(err).
				Str("email", *coupon.UserEmail).
				Str("coupon_code", coupon.Code).
				Msg("Failed to send schema email")
		} else {
			log.Info().
				Str("email", *coupon.UserEmail).
				Str("coupon_code", coupon.Code).
				Msg("Schema email sent successfully")
		}
	}()
}

// openFromStorage opens file from local path (file://) or downloads from S3
func (s *ImageService) openFromStorage(ctx context.Context, key string) (io.ReadCloser, error) {
	if strings.HasPrefix(key, "file://") {
		path := strings.TrimPrefix(key, "file://")
		f, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("failed to open local file: %w", err)
		}
		return f, nil
	}
	return s.deps.S3Client.DownloadFile(ctx, key)
}

// cleanupLocalFiles removes local temporary files for coupon (uploads/edited/processed/previews)
func (s *ImageService) cleanupLocalFiles(couponID uuid.UUID) {
	base := filepath.Join(s.deps.WorkingDir)
	dirs := []string{
		filepath.Join(base, "uploads", couponID.String()),
		filepath.Join(base, "edited", couponID.String()),
		filepath.Join(base, "processed", couponID.String()),
		filepath.Join(base, "previews", couponID.String()),
	}
	for _, d := range dirs {
		_ = os.RemoveAll(d)
	}
}

// startRetentionCleaner cleans working directories hourly, removing files older than retention
func (s *ImageService) startRetentionCleaner() error {
	retentionHours := 24
	if v := os.Getenv("IMAGE_LOCAL_RETENTION_HOURS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			retentionHours = n
		}
	}
	retention := time.Duration(retentionHours) * time.Hour
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := s.cleanupOldFiles(retention); err != nil {
				log.Error().Err(err).Msg("Failed to cleanup old files")
			}
		case <-time.After(1 * time.Hour):
		}
	}
}

// cleanupOldFiles removes old files
func (s *ImageService) cleanupOldFiles(retention time.Duration) error {
	return filepath.Walk(filepath.Join(s.deps.WorkingDir), func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil {
			return nil
		}
		if info.IsDir() {
			if time.Since(info.ModTime()) > retention {
				_ = os.Remove(path)
			}
			return nil
		}
		if time.Since(info.ModTime()) > retention {
			_ = os.Remove(path)
		}
		return nil
	})
}

// buildDataURLFromLocalPath reads file and returns data URL for browser embedding
func (s *ImageService) buildDataURLFromLocalPath(path string) (*string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	ext := strings.ToLower(filepath.Ext(path))
	mime := "image/jpeg"
	if ext == ".png" {
		mime = "image/png"
	}
	enc := base64.StdEncoding.EncodeToString(b)
	u := fmt.Sprintf("data:%s;base64,%s", mime, enc)
	return &u, nil
}
