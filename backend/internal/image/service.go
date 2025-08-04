package image

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"math"
	"mime/multipart"
	"sort"
	"time"

	"github.com/disintegration/imaging"
	"github.com/google/uuid"
	"github.com/jung-kurt/gofpdf"
	"github.com/rs/zerolog/log"
	"github.com/skr1ms/mosaic/internal/coupon"
	"github.com/skr1ms/mosaic/internal/types"
	"github.com/skr1ms/mosaic/pkg/email"
	"github.com/skr1ms/mosaic/pkg/s3"
	"github.com/skr1ms/mosaic/pkg/stablediffusion"
	"github.com/skr1ms/mosaic/pkg/zip"
)

type ImageServiceDeps struct {
	ImageRepository       *ImageRepository
	CouponRepository      *coupon.CouponRepository
	S3Client              *s3.S3Client
	StableDiffusionClient *stablediffusion.StableDiffusionClient
	EmailService          *email.Mailer
	ZipService            *zip.ZipService
}

type ImageService struct {
	deps *ImageServiceDeps
}

// MosaicColor представляет цвет в палитре алмазной мозаики
type MosaicColor struct {
	R, G, B uint8
	Code    string // Код цвета (например, "001", "002")
	Name    string // Название цвета (например, "Белый", "Черный")
	HexCode string // HEX код цвета
	Count   int    // Количество камней этого цвета
}

// MosaicGrid представляет сетку алмазной мозаики
type MosaicGrid struct {
	Width  int           // Ширина в камнях
	Height int           // Высота в камнях
	Colors []MosaicColor // Палитра цветов
	Grid   [][]string    // Сетка с кодами цветов
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

	// Создаем ZIP-архив схемы алмазной мозаики
	schemaS3Key, err := s.createSchemaZipArchive(ctx, imageRecord, sourceS3Key)
	if err != nil {
		return fmt.Errorf("failed to create schema ZIP archive: %w", err)
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

// createSchemaZipArchive создает ZIP-архив с файлами схемы алмазной мозаики
func (s *ImageService) createSchemaZipArchive(ctx context.Context, imageRecord *Image, sourceS3Key string) (string, error) {
	// Скачиваем необходимые файлы из S3
	files := []zip.FileData{}

	// 1. Оригинальное изображение
	originalReader, err := s.deps.S3Client.DownloadFile(ctx, imageRecord.OriginalImageS3Key)
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

	// 2. Превью изображения (если есть)
	if imageRecord.PreviewS3Key != nil {
		previewReader, err := s.deps.S3Client.DownloadFile(ctx, *imageRecord.PreviewS3Key)
		if err == nil {
			defer previewReader.Close()
			previewData, err := io.ReadAll(previewReader)
			if err == nil {
				files = append(files, zip.FileData{
					Name:    "preview.jpg",
					Content: bytes.NewReader(previewData),
					Size:    int64(len(previewData)),
				})
			} else {
				log.Warn().Err(err).Msg("Failed to read preview image data for ZIP archive")
			}
		} else {
			log.Warn().Err(err).Msg("Failed to download preview image for ZIP archive")
		}
	}

	// 3. Создаем PDF схему (временно используем обработанное изображение)
	schemaData, err := s.createMosaicSchemaPDF(ctx, sourceS3Key, imageRecord)
	if err != nil {
		return "", fmt.Errorf("failed to create mosaic schema PDF: %w", err)
	}
	files = append(files, zip.FileData{
		Name:    "schema.pdf",
		Content: bytes.NewReader(schemaData),
		Size:    int64(len(schemaData)),
	})

	// Создаем ZIP-архив с использованием ID изображения
	zipBuffer, err := s.deps.ZipService.CreateSchemaArchive(imageRecord.ID, files)
	if err != nil {
		return "", fmt.Errorf("failed to create ZIP archive: %w", err)
	}

	// Загружаем ZIP-архив в S3 с именем по ID изображения
	zipS3Key := fmt.Sprintf("schemas/%s.zip", imageRecord.ID.String())
	_, err = s.deps.S3Client.UploadFile(ctx, zipBuffer, int64(zipBuffer.Len()), "application/zip", zipS3Key, imageRecord.ID)
	if err != nil {
		return "", fmt.Errorf("failed to upload ZIP archive to S3: %w", err)
	}

	log.Info().
		Str("image_id", imageRecord.ID.String()).
		Str("zip_s3_key", zipS3Key).
		Int("files_count", len(files)).
		Msg("Schema ZIP archive created successfully")

	return zipS3Key, nil
}

// createMosaicSchemaPDF создает PDF файл с схемой алмазной мозаики
func (s *ImageService) createMosaicSchemaPDF(ctx context.Context, sourceS3Key string, imageRecord *Image) ([]byte, error) {
	// Получаем информацию о купоне для определения размеров и стиля
	coupon, err := s.deps.CouponRepository.GetByID(ctx, imageRecord.CouponID)
	if err != nil {
		return nil, fmt.Errorf("failed to get coupon: %w", err)
	}

	// Скачиваем исходное изображение
	imageReader, err := s.deps.S3Client.DownloadFile(ctx, sourceS3Key)
	if err != nil {
		return nil, fmt.Errorf("failed to download source image: %w", err)
	}
	defer imageReader.Close()

	imageData, err := io.ReadAll(imageReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read source image data: %w", err)
	}

	// Декодируем изображение
	img, _, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Создаем сетку алмазной мозаики
	mosaicGrid, err := s.createMosaicGrid(img, coupon.Size, coupon.Style)
	if err != nil {
		return nil, fmt.Errorf("failed to create mosaic grid: %w", err)
	}

	// Создаем PDF документ
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Генерируем PDF с схемой
	err = s.generateMosaicPDF(pdf, mosaicGrid, coupon.Size, coupon.Code)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	// Сохраняем PDF в буфер
	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to output PDF: %w", err)
	}

	log.Info().
		Str("image_id", imageRecord.ID.String()).
		Str("source_s3_key", sourceS3Key).
		Str("coupon_size", coupon.Size).
		Str("coupon_style", coupon.Style).
		Int("colors_count", len(mosaicGrid.Colors)).
		Msg("Mosaic schema PDF created successfully")

	return buf.Bytes(), nil
}

// createMosaicGrid создает сетку алмазной мозаики из изображения
func (s *ImageService) createMosaicGrid(img image.Image, size, style string) (*MosaicGrid, error) {
	// Определяем размеры сетки в камнях (примерно 2-3 камня на сантиметр)
	gridWidth, gridHeight := s.calculateGridDimensions(size)

	// Изменяем размер изображения под сетку
	resizedImg := imaging.Resize(img, gridWidth, gridHeight, imaging.Lanczos)

	// Создаем цветовую палитру на основе стиля
	palette := s.createColorPalette(style)

	// Создаем сетку
	grid := make([][]string, gridHeight)
	colorCounts := make(map[string]int)

	for y := 0; y < gridHeight; y++ {
		grid[y] = make([]string, gridWidth)
		for x := 0; x < gridWidth; x++ {
			// Получаем цвет пикселя
			pixelColor := resizedImg.At(x, y)
			r, g, b, _ := pixelColor.RGBA()

			// Находим ближайший цвет в палитре
			closestColor := s.findClosestColor(uint8(r>>8), uint8(g>>8), uint8(b>>8), palette)
			grid[y][x] = closestColor.Code
			colorCounts[closestColor.Code]++
		}
	}

	// Обновляем количество камней для каждого цвета
	for i := range palette {
		palette[i].Count = colorCounts[palette[i].Code]
	}

	// Фильтруем неиспользуемые цвета
	usedColors := []MosaicColor{}
	for _, color := range palette {
		if color.Count > 0 {
			usedColors = append(usedColors, color)
		}
	}

	// Сортируем по количеству использования (по убыванию)
	sort.Slice(usedColors, func(i, j int) bool {
		return usedColors[i].Count > usedColors[j].Count
	})

	return &MosaicGrid{
		Width:  gridWidth,
		Height: gridHeight,
		Colors: usedColors,
		Grid:   grid,
	}, nil
}

// calculateGridDimensions вычисляет размеры сетки в камнях
func (s *ImageService) calculateGridDimensions(size string) (width, height int) {
	// Примерно 2.5 камня на сантиметр
	const stonesPerCm = 2.5

	switch size {
	case "21x30":
		return int(math.Round(21 * stonesPerCm)), int(math.Round(30 * stonesPerCm))
	case "30x40":
		return int(math.Round(30 * stonesPerCm)), int(math.Round(40 * stonesPerCm))
	case "40x40":
		return int(math.Round(40 * stonesPerCm)), int(math.Round(40 * stonesPerCm))
	case "40x50":
		return int(math.Round(40 * stonesPerCm)), int(math.Round(50 * stonesPerCm))
	case "40x60":
		return int(math.Round(40 * stonesPerCm)), int(math.Round(60 * stonesPerCm))
	case "50x70":
		return int(math.Round(50 * stonesPerCm)), int(math.Round(70 * stonesPerCm))
	default:
		return int(math.Round(30 * stonesPerCm)), int(math.Round(40 * stonesPerCm))
	}
}

// createColorPalette создает цветовую палитру на основе стиля
func (s *ImageService) createColorPalette(style string) []MosaicColor {
	switch style {
	case "gray":
		return s.createGrayScalePalette()
	case "flesh":
		return s.createFleshTonePalette()
	case "popart":
		return s.createPopArtPalette()
	case "max_colors":
		return s.createMaxColorsPalette()
	default:
		return s.createMaxColorsPalette()
	}
}

// createGrayScalePalette создает палитру оттенков серого
func (s *ImageService) createGrayScalePalette() []MosaicColor {
	colors := []MosaicColor{}
	for i := 0; i < 16; i++ {
		gray := uint8(i * 255 / 15)
		colors = append(colors, MosaicColor{
			R:       gray,
			G:       gray,
			B:       gray,
			Code:    fmt.Sprintf("G%02d", i+1),
			Name:    fmt.Sprintf("Серый %d", i+1),
			HexCode: fmt.Sprintf("#%02X%02X%02X", gray, gray, gray),
		})
	}
	return colors
}

// createFleshTonePalette создает палитру телесных оттенков
func (s *ImageService) createFleshTonePalette() []MosaicColor {
	return []MosaicColor{
		{R: 255, G: 239, B: 213, Code: "F01", Name: "Очень светлый", HexCode: "#FFEFD5"},
		{R: 255, G: 228, B: 196, Code: "F02", Name: "Светлый", HexCode: "#FFE4C4"},
		{R: 255, G: 218, B: 185, Code: "F03", Name: "Светло-персиковый", HexCode: "#FFDAB9"},
		{R: 255, G: 160, B: 122, Code: "F04", Name: "Персиковый", HexCode: "#FFA07A"},
		{R: 238, G: 203, B: 173, Code: "F05", Name: "Светло-коричневый", HexCode: "#EECBAD"},
		{R: 222, G: 184, B: 135, Code: "F06", Name: "Бежевый", HexCode: "#DEB887"},
		{R: 205, G: 133, B: 63, Code: "F07", Name: "Коричневый", HexCode: "#CD853F"},
		{R: 160, G: 82, B: 45, Code: "F08", Name: "Темно-коричневый", HexCode: "#A0522D"},
		{R: 139, G: 69, B: 19, Code: "F09", Name: "Очень темный", HexCode: "#8B4513"},
		{R: 255, G: 192, B: 203, Code: "F10", Name: "Розоватый", HexCode: "#FFC0CB"},
	}
}

// createPopArtPalette создает яркую палитру в стиле поп-арт
func (s *ImageService) createPopArtPalette() []MosaicColor {
	return []MosaicColor{
		{R: 255, G: 0, B: 0, Code: "P01", Name: "Красный", HexCode: "#FF0000"},
		{R: 0, G: 255, B: 0, Code: "P02", Name: "Зеленый", HexCode: "#00FF00"},
		{R: 0, G: 0, B: 255, Code: "P03", Name: "Синий", HexCode: "#0000FF"},
		{R: 255, G: 255, B: 0, Code: "P04", Name: "Желтый", HexCode: "#FFFF00"},
		{R: 255, G: 0, B: 255, Code: "P05", Name: "Пурпурный", HexCode: "#FF00FF"},
		{R: 0, G: 255, B: 255, Code: "P06", Name: "Циан", HexCode: "#00FFFF"},
		{R: 255, G: 165, B: 0, Code: "P07", Name: "Оранжевый", HexCode: "#FFA500"},
		{R: 128, G: 0, B: 128, Code: "P08", Name: "Фиолетовый", HexCode: "#800080"},
		{R: 255, G: 192, B: 203, Code: "P09", Name: "Розовый", HexCode: "#FFC0CB"},
		{R: 0, G: 128, B: 0, Code: "P10", Name: "Темно-зеленый", HexCode: "#008000"},
		{R: 0, G: 0, B: 0, Code: "P11", Name: "Черный", HexCode: "#000000"},
		{R: 255, G: 255, B: 255, Code: "P12", Name: "Белый", HexCode: "#FFFFFF"},
	}
}

// createMaxColorsPalette создает максимальную цветовую палитру
func (s *ImageService) createMaxColorsPalette() []MosaicColor {
	colors := []MosaicColor{}

	// Основные цвета
	basicColors := []struct {
		r, g, b uint8
		name    string
	}{
		{255, 255, 255, "Белый"},
		{0, 0, 0, "Черный"},
		{255, 0, 0, "Красный"},
		{0, 255, 0, "Зеленый"},
		{0, 0, 255, "Синий"},
		{255, 255, 0, "Желтый"},
		{255, 0, 255, "Пурпурный"},
		{0, 255, 255, "Циан"},
		{255, 165, 0, "Оранжевый"},
		{128, 0, 128, "Фиолетовый"},
		{255, 192, 203, "Розовый"},
		{165, 42, 42, "Коричневый"},
		{128, 128, 128, "Серый"},
		{192, 192, 192, "Светло-серый"},
		{64, 64, 64, "Темно-серый"},
	}

	for i, color := range basicColors {
		colors = append(colors, MosaicColor{
			R:       color.r,
			G:       color.g,
			B:       color.b,
			Code:    fmt.Sprintf("M%02d", i+1),
			Name:    color.name,
			HexCode: fmt.Sprintf("#%02X%02X%02X", color.r, color.g, color.b),
		})
	}

	// Добавляем дополнительные оттенки
	additionalColors := []struct {
		r, g, b uint8
		name    string
	}{
		{139, 69, 19, "Темно-коричневый"},
		{255, 228, 196, "Бисквитный"},
		{255, 218, 185, "Персиковый"},
		{240, 230, 140, "Хаки"},
		{0, 128, 128, "Бирюзовый"},
		{128, 0, 0, "Темно-красный"},
		{0, 128, 0, "Темно-зеленый"},
		{0, 0, 128, "Темно-синий"},
		{255, 20, 147, "Глубокий розовый"},
		{75, 0, 130, "Индиго"},
		{255, 215, 0, "Золотой"},
		{192, 192, 192, "Серебряный"},
		{128, 128, 0, "Оливковый"},
		{0, 128, 128, "Морской волны"},
		{128, 0, 128, "Сливовый"},
	}

	for i, color := range additionalColors {
		colors = append(colors, MosaicColor{
			R:       color.r,
			G:       color.g,
			B:       color.b,
			Code:    fmt.Sprintf("A%02d", i+1),
			Name:    color.name,
			HexCode: fmt.Sprintf("#%02X%02X%02X", color.r, color.g, color.b),
		})
	}

	return colors
}

// findClosestColor находит ближайший цвет в палитре
func (s *ImageService) findClosestColor(r, g, b uint8, palette []MosaicColor) MosaicColor {
	var closestColor MosaicColor
	minDistance := math.MaxFloat64

	for _, color := range palette {
		// Вычисляем расстояние в цветовом пространстве RGB
		dr := float64(r) - float64(color.R)
		dg := float64(g) - float64(color.G)
		db := float64(b) - float64(color.B)
		distance := math.Sqrt(dr*dr + dg*dg + db*db)

		if distance < minDistance {
			minDistance = distance
			closestColor = color
		}
	}

	return closestColor
}

// generateMosaicPDF генерирует PDF документ с схемой мозаики
func (s *ImageService) generateMosaicPDF(pdf *gofpdf.Fpdf, grid *MosaicGrid, size, couponCode string) error {
	// Заголовок
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, "Схема алмазной мозаики")
	pdf.Ln(12)

	// Информация о схеме
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(0, 8, fmt.Sprintf("Размер: %s см", size))
	pdf.Ln(6)
	pdf.Cell(0, 8, fmt.Sprintf("Код купона: %s", couponCode))
	pdf.Ln(6)
	pdf.Cell(0, 8, fmt.Sprintf("Размер сетки: %d x %d камней", grid.Width, grid.Height))
	pdf.Ln(6)
	pdf.Cell(0, 8, fmt.Sprintf("Количество цветов: %d", len(grid.Colors)))
	pdf.Ln(10)

	// Легенда цветов
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 8, "Легенда цветов:")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 10)

	// Таблица с цветами
	colWidth := 35.0
	rowHeight := 6.0

	// Заголовки таблицы
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(colWidth, rowHeight, "Код")
	pdf.Cell(colWidth, rowHeight, "Название")
	pdf.Cell(colWidth, rowHeight, "HEX")
	pdf.Cell(colWidth, rowHeight, "Количество")
	pdf.Cell(colWidth, rowHeight, "Цвет")
	pdf.Ln(rowHeight)

	pdf.SetFont("Arial", "", 9)

	for _, color := range grid.Colors {
		// Данные цвета
		pdf.Cell(colWidth, rowHeight, color.Code)
		pdf.Cell(colWidth, rowHeight, color.Name)
		pdf.Cell(colWidth, rowHeight, color.HexCode)
		pdf.Cell(colWidth, rowHeight, fmt.Sprintf("%d", color.Count))

		// Цветной квадрат
		x, y := pdf.GetXY()
		pdf.SetFillColor(int(color.R), int(color.G), int(color.B))
		pdf.Rect(x, y, colWidth, rowHeight, "F")
		pdf.SetXY(x+colWidth, y)

		pdf.Ln(rowHeight)
	}

	// Добавляем новую страницу для схемы
	pdf.AddPage()

	// Заголовок схемы
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 8, "Схема выкладки:")
	pdf.Ln(12)

	// Инструкции
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 6, "1. Найдите код цвета в легенде")
	pdf.Ln(5)
	pdf.Cell(0, 6, "2. Найдите соответствующие камни")
	pdf.Ln(5)
	pdf.Cell(0, 6, "3. Разместите камни согласно схеме")
	pdf.Ln(5)
	pdf.Cell(0, 6, "4. Работайте участками для удобства")
	pdf.Ln(10)

	// Мини-схема (упрощенная версия для демонстрации)
	s.drawMiniGrid(pdf, grid)

	return nil
}

// drawMiniGrid рисует упрощенную схему сетки
func (s *ImageService) drawMiniGrid(pdf *gofpdf.Fpdf, grid *MosaicGrid) {
	// Масштабируем сетку для отображения в PDF
	maxWidth := 180.0  // Максимальная ширина в мм
	maxHeight := 200.0 // Максимальная высота в мм

	scaleX := maxWidth / float64(grid.Width)
	scaleY := maxHeight / float64(grid.Height)
	scale := math.Min(scaleX, scaleY)

	// Если камни слишком мелкие, показываем только часть схемы
	cellSize := scale
	if cellSize < 1.0 {
		cellSize = 1.0
		pdf.SetFont("Arial", "", 10)
		pdf.Cell(0, 6, "Примечание: Схема показана в уменьшенном масштабе")
		pdf.Ln(8)
	}

	// Рисуем только центральную часть, если схема слишком большая
	startX := 0
	startY := 0
	drawWidth := grid.Width
	drawHeight := grid.Height

	if cellSize == 1.0 {
		// Показываем центральную часть
		maxCells := int(math.Min(maxWidth, maxHeight))
		if grid.Width > maxCells {
			startX = (grid.Width - maxCells) / 2
			drawWidth = maxCells
		}
		if grid.Height > maxCells {
			startY = (grid.Height - maxCells) / 2
			drawHeight = maxCells
		}
	}

	x0, y0 := pdf.GetXY()

	// Создаем карту цветов для быстрого поиска
	colorMap := make(map[string]MosaicColor)
	for _, color := range grid.Colors {
		colorMap[color.Code] = color
	}

	// Рисуем сетку
	for y := 0; y < drawHeight; y++ {
		for x := 0; x < drawWidth; x++ {
			gridX := startX + x
			gridY := startY + y

			if gridY < len(grid.Grid) && gridX < len(grid.Grid[gridY]) {
				colorCode := grid.Grid[gridY][gridX]
				if color, exists := colorMap[colorCode]; exists {
					pdf.SetFillColor(int(color.R), int(color.G), int(color.B))
					pdf.Rect(x0+float64(x)*cellSize, y0+float64(y)*cellSize, cellSize, cellSize, "F")
				}
			}
		}
	}

	// Добавляем сетку для лучшей видимости (каждые 10 камней)
	if cellSize > 2.0 {
		pdf.SetDrawColor(128, 128, 128)
		pdf.SetLineWidth(0.1)

		for i := 0; i <= drawWidth; i += 10 {
			x := x0 + float64(i)*cellSize
			pdf.Line(x, y0, x, y0+float64(drawHeight)*cellSize)
		}

		for i := 0; i <= drawHeight; i += 10 {
			y := y0 + float64(i)*cellSize
			pdf.Line(x0, y, x0+float64(drawWidth)*cellSize, y)
		}
	}
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
