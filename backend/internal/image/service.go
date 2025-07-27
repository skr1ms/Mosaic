package image

import (
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/skr1ms/mosaic/internal/coupon"
	"github.com/skr1ms/mosaic/internal/types"
)

type ImageService struct {
	ImageRepository  *ImageRepository
	CouponRepository *coupon.CouponRepository
	Logger           *zerolog.Logger
}

func NewImageService(repo *ImageRepository, couponRepo *coupon.CouponRepository, logger *zerolog.Logger) *ImageService {
	return &ImageService{
		ImageRepository:  repo,
		CouponRepository: couponRepo,
		Logger:           logger,
	}
}

func (s *ImageService) GetQueue(status string) ([]*Image, error) {
	var tasks []*Image
	var err error

	if status != "" {
		tasks, err = s.ImageRepository.GetByStatus(status)
	} else {
		tasks, err = s.ImageRepository.GetAll()
	}

	if err != nil {
		s.Logger.Error().Err(err).Msg(ErrFailedToFetchQueue.Error())
		return nil, ErrFailedToFetchQueue
	}

	return tasks, nil
}

func (s *ImageService) AddToQueue(couponID uuid.UUID) error {
	// Получаем купон
	coupon, err := s.CouponRepository.GetByID(couponID)
	if err != nil {
		s.Logger.Error().Err(err).Msg(ErrCouponNotFound.Error())
		return ErrCouponNotFound
	}

	// Если купон уже в очереди, то не добавляем в очередь
	existingTask, err := s.ImageRepository.GetByCouponID(couponID)
	if err == nil && existingTask != nil {
		s.Logger.Error().Err(err).Msg(ErrCouponAlreadyInQueue.Error())
		return ErrCouponAlreadyInQueue
	}

	// Если купон уже использован, то не добавляем в очередь
	if coupon.Status == "used" && coupon.PreviewURL != nil && *coupon.PreviewURL != "" {
		s.Logger.Error().Err(err).Msg(ErrCouponAlreadyProcessed.Error())
		return ErrCouponAlreadyProcessed
	}

	return nil
}

// ApplyEditing применяет редактирование к изображению
func (s *ImageService) ApplyEditing(task *Image, req types.EditImageRequest) error {
	s.Logger.Info().
		Str("task_id", task.ID.String()).
		Interface("edit_params", req).
		Msg("Applying image editing")

	// TODO:
	// Здесь будет логика обработки изображения:
	// 1. Кадрирование по координатам CropX, CropY, CropWidth, CropHeight
	// 2. Поворот на угол Rotation
	// 3. Масштабирование на Scale
	// 4. Сохранение отредактированного изображения

	// Генерируем путь для отредактированного изображения
	editedPath := filepath.Join("uploads", "edited", task.CouponID.String(), "edited.jpg")
	task.EditedImagePath = &editedPath

	// Обновляем задачу в базе данных
	if err := s.ImageRepository.Update(task); err != nil {
		s.Logger.Error().Err(err).Msg("Failed to update task with edited image path")
		return err
	}

	s.Logger.Info().
		Str("task_id", task.ID.String()).
		Str("edited_path", editedPath).
		Msg("Image editing completed")

	return nil
}

// ApplyProcessing применяет стиль обработки к изображению
func (s *ImageService) ApplyProcessing(task *Image, req types.ProcessImageRequest) error {
	s.Logger.Info().
		Str("task_id", task.ID.String()).
		Interface("processing_params", req).
		Msg("Applying image processing")

	// Сохраняем параметры обработки
	task.ProcessingParams = ProcessingParams{
		Style: req.Style,
		Settings: map[string]interface{}{
			"use_ai":     req.UseAI,
			"lighting":   req.Lighting,
			"contrast":   req.Contrast,
			"brightness": req.Brightness,
			"saturation": req.Saturation,
			"settings":   req.Settings,
		},
	}

	// TODO:
	// Здесь будет логика обработки изображения:
	// 1. Применение стиля (grayscale, skin_tones, pop_art, max_colors)
	// 2. AI обработка если UseAI = true
	// 3. Настройка освещения, контрастности, яркости, насыщенности
	// 4. Создание превью

	// Генерируем пути для превью
	previewPath := filepath.Join("uploads", "previews", task.CouponID.String(), "preview.jpg")
	task.PreviewPath = &previewPath

	// Обновляем статус задачи
	task.Status = "processing"
	now := time.Now()
	task.StartedAt = &now

	// Обновляем задачу в базе данных
	if err := s.ImageRepository.Update(task); err != nil {
		s.Logger.Error().Err(err).Msg("Failed to update task with processing params")
		return err
	}

	s.Logger.Info().
		Str("task_id", task.ID.String()).
		Str("preview_path", previewPath).
		Msg("Image processing started")

	return nil
}

// GenerateSchema создает финальную схему мозаики
func (s *ImageService) GenerateSchema(task *Image, req types.GenerateSchemaRequest) (string, error) {
	s.Logger.Info().
		Str("task_id", task.ID.String()).
		Bool("confirmed", req.Confirmed).
		Msg("Generating mosaic schema")

	if !req.Confirmed {
		return "", ErrSchemaNotConfirmed
	}

	// TODO:
	// Здесь будет логика создания схемы:
	// 1. Конвертация изображения в схему алмазной мозаики
	// 2. Создание PDF с инструкциями
	// 3. Создание файла с цветовой схемой
	// 4. Упаковка в ZIP архив

	// Генерируем путь для готовой схемы
	schemaPath := filepath.Join("uploads", "schemas", task.CouponID.String(), "schema.zip")
	task.ResultPath = &schemaPath

	// Обновляем статус задачи как завершенной
	task.Status = "completed"
	now := time.Now()
	task.CompletedAt = &now

	// Обновляем задачу в базе данных
	if err := s.ImageRepository.Update(task); err != nil {
		s.Logger.Error().Err(err).Msg("Failed to update task with schema path")
		return "", err
	}

	s.Logger.Info().
		Str("task_id", task.ID.String()).
		Str("schema_path", schemaPath).
		Msg("Schema generation completed")

	return schemaPath, nil
}

// SaveUploadedFile сохраняет загруженный файл
func (s *ImageService) SaveUploadedFile(filename string, couponID uuid.UUID, fileData []byte) (string, error) {
	// Создаем путь для сохранения файла
	uploadPath := filepath.Join("uploads", "originals", couponID.String(), filename)

	// ToDO:
	// Здесь будет логика сохранения файла на диск
	// 1. Создание директории если не существует
	// 2. Сохранение файла
	// 3. Проверка целостности

	s.Logger.Info().
		Str("coupon_id", couponID.String()).
		Str("upload_path", uploadPath).
		Msg("File uploaded successfully")

	return uploadPath, nil
}
