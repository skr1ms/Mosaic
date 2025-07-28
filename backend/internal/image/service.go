package image

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/skr1ms/mosaic/internal/coupon"
	"github.com/skr1ms/mosaic/internal/types"
)

type ImageServiceDeps struct {
	ImageRepository  *ImageRepository
	CouponRepository *coupon.CouponRepository
}

type ImageService struct {
	deps *ImageServiceDeps
}

func NewImageService(deps *ImageServiceDeps) *ImageService {
	return &ImageService{
		deps: deps,
	}
}

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

func (s *ImageService) AddToQueue(couponID uuid.UUID) error {
	// Получаем купон
	coupon, err := s.deps.CouponRepository.GetByID(context.Background(), couponID)
	if err != nil {
		return fmt.Errorf("coupon not found: %w", err)
	}

	// Если купон уже в очереди, то не добавляем в очередь
	existingTask, err := s.deps.ImageRepository.GetByCouponID(context.Background(), couponID)
	if err == nil && existingTask != nil {
		return fmt.Errorf("coupon already in queue: %w", err)
	}

	// Если купон уже использован, то не добавляем в очередь
	if coupon.Status == "used" && coupon.PreviewURL != nil && *coupon.PreviewURL != "" {
		return fmt.Errorf("coupon already processed: %w", err)
	}

	return nil
}

// ApplyEditing применяет редактирование к изображению
func (s *ImageService) ApplyEditing(task *Image, req types.EditImageRequest) error {
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
	if err := s.deps.ImageRepository.Update(context.Background(), task); err != nil {
		return fmt.Errorf("failed to update task with edited image path: %w", err)
	}

	return nil
}

// ApplyProcessing применяет стиль обработки к изображению
func (s *ImageService) ApplyProcessing(task *Image, req types.ProcessImageRequest) error {
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
	if err := s.deps.ImageRepository.Update(context.Background(), task); err != nil {
		return fmt.Errorf("failed to update task with processing params: %w", err)
	}

	return nil
}

// GenerateSchema создает финальную схему мозаики
func (s *ImageService) GenerateSchema(task *Image, req types.GenerateSchemaRequest) (string, error) {
	if !req.Confirmed {
		return "", fmt.Errorf("schema not confirmed")
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
	if err := s.deps.ImageRepository.Update(context.Background(), task); err != nil {
		return "", fmt.Errorf("failed to update task with schema path: %w", err)
	}

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

	return uploadPath, nil
}
