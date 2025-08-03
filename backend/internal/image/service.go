package image

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"time"

	"github.com/disintegration/imaging"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/skr1ms/mosaic/internal/coupon"
	"github.com/skr1ms/mosaic/internal/types"
	"github.com/skr1ms/mosaic/pkg/email"
	"github.com/skr1ms/mosaic/pkg/s3"
	"github.com/skr1ms/mosaic/pkg/stablediffusion"
)

type ImageServiceDeps struct {
	ImageRepository       *ImageRepository
	CouponRepository      *coupon.CouponRepository
	S3Client              *s3.S3Client
	StableDiffusionClient *stablediffusion.StableDiffusionClient
	EmailService          *email.Mailer
}

type ImageService struct {
	deps *ImageServiceDeps
}

func NewImageService(deps *ImageServiceDeps) *ImageService {
	return &ImageService{
		deps: deps,
	}
}

// GetQueue возвращает задачи в очереди с фильтрацией по статусу
func (s *ImageService) GetQueue(status string) ([]*Image, error) {
	var tasks []*Image
	var err error

	if status != "" {
		tasks, err = s.deps.ImageRepository.GetByStatus(context.Background(), status)
	} else {
		tasks, err = s.deps.ImageRepository.GetAll(context.Background())
	}

	if err != nil {
		return nil, fmt.Errorf("failed to fetch queue: %w", err)
	}

	return tasks, nil
}

// UploadImage загружает и сохраняет изображение в S3
func (s *ImageService) UploadImage(ctx context.Context, couponID uuid.UUID, file *multipart.FileHeader, userEmail string) (*Image, error) {
	// Проверяем купон
	coupon, err := s.deps.CouponRepository.GetByID(ctx, couponID)
	if err != nil {
		return nil, fmt.Errorf("coupon not found: %w", err)
	}

	if coupon.Status != "activated" {
		return nil, fmt.Errorf("coupon not activated")
	}

	// Проверяем, что для купона еще не создана задача обработки
	existingImage, err := s.deps.ImageRepository.GetByCouponID(ctx, couponID)
	if err == nil && existingImage != nil {
		return nil, fmt.Errorf("image already uploaded for this coupon")
	}

	// Проверяем тип файла
	if !isValidImageType(file) {
		return nil, fmt.Errorf("invalid image type, supported: JPG, PNG")
	}

	// Проверяем размер файла (макс. 10MB)
	if file.Size > 10<<20 {
		return nil, fmt.Errorf("file too large, maximum size is 10MB")
	}

	// Открываем файл
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Загружаем в S3
	s3Key, err := s.deps.S3Client.UploadFile(ctx, src, file.Size, file.Header.Get("Content-Type"), "originals", couponID)
	if err != nil {
		return nil, fmt.Errorf("failed to upload to S3: %w", err)
	}

	// Создаем запись в базе данных
	imageRecord := &Image{
		CouponID:           couponID,
		OriginalImageS3Key: s3Key,
		UserEmail:          userEmail,
		Status:             "uploaded",
		Priority:           1,
	}

	if err := s.deps.ImageRepository.Create(ctx, imageRecord); err != nil {
		// Если не удалось создать запись, удаляем файл из S3
		s.deps.S3Client.DeleteFile(ctx, s3Key)
		return nil, fmt.Errorf("failed to create image record: %w", err)
	}

	log.Info().
		Str("image_id", imageRecord.ID.String()).
		Str("coupon_id", couponID.String()).
		Str("s3_key", s3Key).
		Msg("Image uploaded successfully")

	return imageRecord, nil
}

// EditImage применяет редактирование к изображению
func (s *ImageService) EditImage(ctx context.Context, imageID uuid.UUID, editParams ImageEditParams) error {
	// Получаем запись об изображении
	imageRecord, err := s.deps.ImageRepository.GetByID(ctx, imageID)
	if err != nil {
		return fmt.Errorf("image not found: %w", err)
	}

	if imageRecord.Status != "uploaded" {
		return fmt.Errorf("image cannot be edited in current status: %s", imageRecord.Status)
	}

	// Скачиваем оригинальное изображение из S3
	originalReader, err := s.deps.S3Client.DownloadFile(ctx, imageRecord.OriginalImageS3Key)
	if err != nil {
		return fmt.Errorf("failed to download original image: %w", err)
	}
	defer originalReader.Close()

	// Декодируем изображение
	img, format, err := image.Decode(originalReader)
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	// Применяем редактирование
	editedImg := s.applyImageEditing(img, editParams)

	// Кодируем обратно в нужный формат
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

	// Загружаем отредактированное изображение в S3
	editedS3Key, err := s.deps.S3Client.UploadFile(ctx, &buf, int64(buf.Len()), "image/jpeg", "edited", imageRecord.CouponID)
	if err != nil {
		return fmt.Errorf("failed to upload edited image: %w", err)
	}

	// Обновляем запись в базе данных
	imageRecord.EditedImageS3Key = &editedS3Key
	imageRecord.Status = "edited"
	if err := s.deps.ImageRepository.Update(ctx, imageRecord); err != nil {
		// Если не удалось обновить запись, удаляем файл из S3
		s.deps.S3Client.DeleteFile(ctx, editedS3Key)
		return fmt.Errorf("failed to update image record: %w", err)
	}

	log.Info().
		Str("image_id", imageID.String()).
		Str("edited_s3_key", editedS3Key).
		Msg("Image edited successfully")

	return nil
}

// ProcessImage обрабатывает изображение через Stable Diffusion
func (s *ImageService) ProcessImage(ctx context.Context, imageID uuid.UUID, processParams ProcessingParams) error {
	// Получаем запись об изображении
	imageRecord, err := s.deps.ImageRepository.GetByID(ctx, imageID)
	if err != nil {
		return fmt.Errorf("image not found: %w", err)
	}

	if imageRecord.Status != "edited" && imageRecord.Status != "uploaded" {
		return fmt.Errorf("image cannot be processed in current status: %s", imageRecord.Status)
	}

	// Определяем какое изображение использовать для обработки
	sourceS3Key := imageRecord.OriginalImageS3Key
	if imageRecord.EditedImageS3Key != nil {
		sourceS3Key = *imageRecord.EditedImageS3Key
	}

	// Получаем размеры купона
	coupon, err := s.deps.CouponRepository.GetByID(ctx, imageRecord.CouponID)
	if err != nil {
		return fmt.Errorf("failed to get coupon: %w", err)
	}

	width, height := parseCouponSize(coupon.Size)

	// Обновляем статус на "processing"
	imageRecord.Status = "processing"
	imageRecord.ProcessingParams = processParams
	now := time.Now()
	imageRecord.StartedAt = &now
	if err := s.deps.ImageRepository.Update(ctx, imageRecord); err != nil {
		return fmt.Errorf("failed to update status to processing: %w", err)
	}

	// Если AI обработка не требуется, просто создаем превью
	if !processParams.UseAI {
		return s.createPreviewWithoutAI(ctx, imageRecord, sourceS3Key)
	}

	// Скачиваем изображение для AI обработки
	sourceReader, err := s.deps.S3Client.DownloadFile(ctx, sourceS3Key)
	if err != nil {
		return s.markProcessingFailed(ctx, imageRecord, fmt.Errorf("failed to download source image: %w", err))
	}
	defer sourceReader.Close()

	// Читаем изображение в память
	sourceData, err := io.ReadAll(sourceReader)
	if err != nil {
		return s.markProcessingFailed(ctx, imageRecord, fmt.Errorf("failed to read source image: %w", err))
	}

	// Кодируем в base64 для Stable Diffusion
	base64Image := s.deps.StableDiffusionClient.EncodeImageToBase64(sourceData)

	// Подготавливаем запрос к Stable Diffusion
	sdRequest := stablediffusion.ProcessImageRequest{
		ImageBase64: base64Image,
		Style:       stablediffusion.ProcessingStyle(processParams.Style),
		UseAI:       processParams.UseAI,
		Lighting:    stablediffusion.LightingType(processParams.Lighting),
		Contrast:    stablediffusion.ContrastLevel(processParams.Contrast),
		Brightness:  processParams.Brightness,
		Saturation:  processParams.Saturation,
		Width:       width,
		Height:      height,
	}

	// Обрабатываем через Stable Diffusion
	processedBase64, err := s.deps.StableDiffusionClient.ProcessImage(ctx, sdRequest)
	if err != nil {
		return s.markProcessingFailed(ctx, imageRecord, fmt.Errorf("stable diffusion processing failed: %w", err))
	}

	// Декодируем результат
	processedData, err := s.deps.StableDiffusionClient.DecodeBase64Image(processedBase64)
	if err != nil {
		return s.markProcessingFailed(ctx, imageRecord, fmt.Errorf("failed to decode processed image: %w", err))
	}

	// Загружаем обработанное изображение в S3
	processedS3Key, err := s.deps.S3Client.UploadFile(ctx, bytes.NewReader(processedData), int64(len(processedData)), "image/jpeg", "processed", imageRecord.CouponID)
	if err != nil {
		return s.markProcessingFailed(ctx, imageRecord, fmt.Errorf("failed to upload processed image: %w", err))
	}

	// Создаем превью
	previewS3Key, err := s.createPreview(ctx, processedData, imageRecord.CouponID)
	if err != nil {
		return s.markProcessingFailed(ctx, imageRecord, fmt.Errorf("failed to create preview: %w", err))
	}

	// Обновляем запись в базе данных
	imageRecord.ProcessedImageS3Key = &processedS3Key
	imageRecord.PreviewS3Key = &previewS3Key
	imageRecord.Status = "processed"
	if err := s.deps.ImageRepository.Update(ctx, imageRecord); err != nil {
		return fmt.Errorf("failed to update image record: %w", err)
	}

	log.Info().
		Str("image_id", imageID.String()).
		Str("processed_s3_key", processedS3Key).
		Str("preview_s3_key", previewS3Key).
		Msg("Image processed successfully")

	return nil
}

// GenerateSchema создает финальную схему алмазной мозаики
func (s *ImageService) GenerateSchema(ctx context.Context, imageID uuid.UUID, confirmed bool) error {
	if !confirmed {
		return fmt.Errorf("schema generation not confirmed")
	}

	// Получаем запись об изображении
	imageRecord, err := s.deps.ImageRepository.GetByID(ctx, imageID)
	if err != nil {
		return fmt.Errorf("image not found: %w", err)
	}

	if imageRecord.Status != "processed" {
		return fmt.Errorf("image must be processed before generating schema")
	}

	// Определяем источник для создания схемы
	sourceS3Key := imageRecord.OriginalImageS3Key
	if imageRecord.ProcessedImageS3Key != nil {
		sourceS3Key = *imageRecord.ProcessedImageS3Key
	} else if imageRecord.EditedImageS3Key != nil {
		sourceS3Key = *imageRecord.EditedImageS3Key
	}

	// TODO: Здесь будет логика создания схемы алмазной мозаики
	// 1. Скачать финальное изображение
	// 2. Конвертировать в схему мозаики
	// 3. Создать PDF с инструкциями
	// 4. Создать файл с цветовой схемой
	// 5. Упаковать в ZIP архив
	// 6. Загрузить в S3

	// Временная заглушка - копируем обработанное изображение как схему
	schemaS3Key, err := s.createTempSchema(ctx, sourceS3Key, imageRecord.CouponID)
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	// Обновляем запись в базе данных
	imageRecord.SchemaS3Key = &schemaS3Key
	imageRecord.Status = "completed"
	now := time.Now()
	imageRecord.CompletedAt = &now
	if err := s.deps.ImageRepository.Update(ctx, imageRecord); err != nil {
		return fmt.Errorf("failed to update image record: %w", err)
	}

	log.Info().
		Str("image_id", imageID.String()).
		Str("schema_s3_key", schemaS3Key).
		Msg("Schema generated successfully")

	// Отправляем email пользователю с готовой схемой
	go func() {
		// Получаем информацию о купоне для email пользователя
		coupon, err := s.deps.CouponRepository.GetByID(context.Background(), imageRecord.CouponID)
		if err != nil {
			log.Error().Err(err).Str("coupon_id", imageRecord.CouponID.String()).Msg("Failed to get coupon for email sending")
			return
		}

		if coupon.UserEmail == nil || *coupon.UserEmail == "" {
			log.Warn().Str("coupon_id", imageRecord.CouponID.String()).Msg("No email address for coupon, skipping email sending")
			return
		}

		// Генерируем presigned URL для скачивания схемы
		schemaURL, err := s.deps.S3Client.GetFileURL(context.Background(), schemaS3Key, 30*24*time.Hour) // 30 дней
		if err != nil {
			log.Error().Err(err).Str("schema_s3_key", schemaS3Key).Msg("Failed to generate presigned URL for schema")
			return
		}

		// Отправляем email
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

	return nil
}

// GetImageStatus возвращает статус обработки изображения
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

	// Добавляем ошибку если есть
	if imageRecord.ErrorMessage != nil {
		response.ErrorMessage = imageRecord.ErrorMessage
	}

	// Генерируем URL для доступа к файлам
	urlExpiry := 24 * time.Hour

	if imageRecord.OriginalImageS3Key != "" {
		if url, err := s.deps.S3Client.GetFileURL(ctx, imageRecord.OriginalImageS3Key, urlExpiry); err == nil {
			response.OriginalURL = &url
		}
	}

	if imageRecord.EditedImageS3Key != nil {
		if url, err := s.deps.S3Client.GetFileURL(ctx, *imageRecord.EditedImageS3Key, urlExpiry); err == nil {
			response.EditedURL = &url
		}
	}

	if imageRecord.ProcessedImageS3Key != nil {
		if url, err := s.deps.S3Client.GetFileURL(ctx, *imageRecord.ProcessedImageS3Key, urlExpiry); err == nil {
			response.ProcessedURL = &url
		}
	}

	if imageRecord.PreviewS3Key != nil {
		if url, err := s.deps.S3Client.GetFileURL(ctx, *imageRecord.PreviewS3Key, urlExpiry); err == nil {
			response.PreviewURL = &url
		}
	}

	if imageRecord.SchemaS3Key != nil {
		if url, err := s.deps.S3Client.GetFileURL(ctx, *imageRecord.SchemaS3Key, urlExpiry); err == nil {
			response.SchemaURL = &url
		}
	}

	return response, nil
}

func (s *ImageService) applyImageEditing(img image.Image, params ImageEditParams) image.Image {
	// Применяем масштабирование
	if params.Scale != 1.0 {
		newWidth := int(float64(img.Bounds().Dx()) * params.Scale)
		newHeight := int(float64(img.Bounds().Dy()) * params.Scale)
		img = imaging.Resize(img, newWidth, newHeight, imaging.Lanczos)
	}

	// Применяем поворот
	switch params.Rotation {
	case 90:
		img = imaging.Rotate90(img)
	case 180:
		img = imaging.Rotate180(img)
	case 270:
		img = imaging.Rotate270(img)
	}

	// Применяем кадрирование
	if params.CropWidth > 0 && params.CropHeight > 0 {
		cropRect := image.Rect(params.CropX, params.CropY, params.CropX+params.CropWidth, params.CropY+params.CropHeight)
		img = imaging.Crop(img, cropRect)
	}

	return img
}

func (s *ImageService) createPreviewWithoutAI(ctx context.Context, imageRecord *Image, sourceS3Key string) error {
	// Скачиваем изображение
	sourceReader, err := s.deps.S3Client.DownloadFile(ctx, sourceS3Key)
	if err != nil {
		return s.markProcessingFailed(ctx, imageRecord, fmt.Errorf("failed to download source image: %w", err))
	}
	defer sourceReader.Close()

	sourceData, err := io.ReadAll(sourceReader)
	if err != nil {
		return s.markProcessingFailed(ctx, imageRecord, fmt.Errorf("failed to read source image: %w", err))
	}

	// Создаем превью
	previewS3Key, err := s.createPreview(ctx, sourceData, imageRecord.CouponID)
	if err != nil {
		return s.markProcessingFailed(ctx, imageRecord, fmt.Errorf("failed to create preview: %w", err))
	}

	// Обновляем запись
	imageRecord.PreviewS3Key = &previewS3Key
	imageRecord.Status = "processed"
	if err := s.deps.ImageRepository.Update(ctx, imageRecord); err != nil {
		return fmt.Errorf("failed to update image record: %w", err)
	}

	return nil
}

func (s *ImageService) createPreview(ctx context.Context, imageData []byte, couponID uuid.UUID) (string, error) {
	// Декодируем изображение
	img, _, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return "", fmt.Errorf("failed to decode image for preview: %w", err)
	}

	// Создаем превью размером 400x300
	preview := imaging.Resize(img, 400, 300, imaging.Lanczos)

	// Кодируем в JPEG
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, preview, &jpeg.Options{Quality: 85}); err != nil {
		return "", fmt.Errorf("failed to encode preview: %w", err)
	}

	// Загружаем в S3
	previewS3Key, err := s.deps.S3Client.UploadFile(ctx, &buf, int64(buf.Len()), "image/jpeg", "previews", couponID)
	if err != nil {
		return "", fmt.Errorf("failed to upload preview: %w", err)
	}

	return previewS3Key, nil
}

func (s *ImageService) createTempSchema(ctx context.Context, sourceS3Key string, couponID uuid.UUID) (string, error) {
	// Временная реализация - копируем файл в папку schemas
	destKey := fmt.Sprintf("schemas/%s_schema.jpg", couponID.String())
	if err := s.deps.S3Client.CopyFile(ctx, sourceS3Key, destKey); err != nil {
		return "", err
	}
	return destKey, nil
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
	switch size {
	case "21x30":
		return 840, 1200 // 21×30 см в пикселях (40 пикселей на см)
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
		return 1200, 1600 // По умолчанию 30x40
	}
}
