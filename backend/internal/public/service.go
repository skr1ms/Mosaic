package public

import (
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/skr1ms/mosaic/internal/coupon"
	"github.com/skr1ms/mosaic/internal/image"
	"github.com/skr1ms/mosaic/internal/partner"
	"github.com/skr1ms/mosaic/internal/types"
	"github.com/skr1ms/mosaic/pkg/randomCouponCode"
)

type PublicService struct {
	CouponRepository  *coupon.CouponRepository
	ImageRepository   *image.ImageRepository
	PartnerRepository *partner.PartnerRepository
	ImageService      *image.ImageService
	Logger            *zerolog.Logger
}

func NewPublicService(couponRepo *coupon.CouponRepository, imageRepo *image.ImageRepository, partnerRepo *partner.PartnerRepository, imageService *image.ImageService, logger *zerolog.Logger) *PublicService {
	return &PublicService{
		CouponRepository:  couponRepo,
		ImageRepository:   imageRepo,
		PartnerRepository: partnerRepo,
		ImageService:      imageService,
		Logger:            logger,
	}
}

// GetPartnerByDomain возвращает публичную информацию о партнере по домену
func (s *PublicService) GetPartnerByDomain(domain string) (map[string]interface{}, error) {
	partner, err := s.PartnerRepository.GetByDomain(domain)
	if err != nil {
		return nil, ErrPartnerNotFound
	}

	// Возвращаем только публичную информацию
	return map[string]interface{}{
		"brand_name":       partner.BrandName,
		"domain":           partner.Domain,
		"logo_url":         partner.LogoURL,
		"ozon_link":        partner.OzonLink,
		"wildberries_link": partner.WildberriesLink,
		"email":            partner.Email,
		"phone":            partner.Phone,
		"address":          partner.Address,
		"telegram_link":    partner.TelegramLink,
		"whatsapp_link":    partner.WhatsappLink,
		"allow_purchases":  partner.AllowPurchases,
	}, nil
}

// GetCouponByCode возвращает информацию о купоне по коду
func (s *PublicService) GetCouponByCode(code string) (map[string]interface{}, error) {
	// Валидация формата кода (12 цифр)
	if len(code) != 12 || !isNumeric(code) {
		return nil, ErrInvalidCouponCode
	}

	coupon, err := s.CouponRepository.GetByCode(code)
	if err != nil {
		return nil, ErrCouponNotFound
	}

	return map[string]interface{}{
		"id":     coupon.ID,
		"code":   coupon.Code,
		"size":   coupon.Size,
		"style":  coupon.Style,
		"status": coupon.Status,
		"valid":  coupon.Status == "new",
	}, nil
}

// ActivateCoupon активирует купон для последующей обработки
func (s *PublicService) ActivateCoupon(code string, req ActivateCouponRequest) (map[string]interface{}, error) {
	coupon, err := s.CouponRepository.GetByCode(code)
	if err != nil {
		return nil, ErrCouponNotFound
	}

	if coupon.Status != "new" {
		return nil, ErrCouponAlreadyUsed
	}

	// Активируем купон
	coupon.Status = "activated"
	coupon.UserEmail = &req.Email
	now := time.Now()
	coupon.ActivatedAt = &now

	if err := s.CouponRepository.Update(coupon); err != nil {
		return nil, ErrFailedToActivateCoupon
	}

	return map[string]interface{}{
		"message":   "Купон успешно активирован",
		"coupon_id": coupon.ID,
		"next_step": "upload_image",
	}, nil
}

// UploadImage загружает изображение для обработки
func (s *PublicService) UploadImage(couponID string, file *multipart.FileHeader) (map[string]interface{}, error) {
	couponUUID, err := uuid.Parse(couponID)
	if err != nil {
		return nil, ErrInvalidCouponID
	}

	// Проверяем купон
	coupon, err := s.CouponRepository.GetByID(couponUUID)
	if err != nil {
		return nil, ErrCouponNotFound
	}

	if coupon.Status != "activated" {
		return nil, ErrCouponNotActivated
	}

	// Проверяем тип файла
	if !isValidImageType(file) {
		return nil, ErrInvalidImageType
	}

	// Проверяем размер файла (макс. 10MB)
	if file.Size > 10<<20 {
		return nil, ErrFileTooLarge
	}

	// Сохраняем файл и создаем задачу на обработку
	imagePath, err := s.saveUploadedFile(file, couponUUID)
	if err != nil {
		return nil, ErrFailedToSaveFile
	}

	// Создаем задачу на обработку
	task := &image.Image{
		CouponID:          couponUUID,
		OriginalImagePath: imagePath,
		UserEmail:         *coupon.UserEmail,
		Status:            "queued",
		Priority:          1, // Высокий приоритет для пользовательских задач
	}

	if err := s.ImageRepository.Create(task); err != nil {
		return nil, ErrFailedToCreateImageTask
	}

	return map[string]interface{}{
		"message":   "Изображение успешно загружено",
		"image_id":  task.ID,
		"next_step": "edit_image",
	}, nil
}

// EditImage применяет редактирование к изображению
func (s *PublicService) EditImage(imageID string, req types.EditImageRequest) (map[string]interface{}, error) {
	imageUUID, err := uuid.Parse(imageID)
	if err != nil {
		return nil, ErrInvalidImageID
	}

	// Получаем задачу
	task, err := s.ImageRepository.GetByID(imageUUID)
	if err != nil {
		return nil, ErrImageNotFound
	}

	// Применяем редактирование через сервис
	if err := s.ImageService.ApplyEditing(task, req); err != nil {
		return nil, ErrFailedToEditImage
	}

	return map[string]interface{}{
		"message":   "Изображение успешно отредактировано",
		"next_step": "choose_style",
	}, nil
}

// ProcessImage применяет стиль обработки к изображению
func (s *PublicService) ProcessImage(imageID string, req types.ProcessImageRequest) (map[string]interface{}, error) {
	imageUUID, err := uuid.Parse(imageID)
	if err != nil {
		return nil, ErrInvalidImageID
	}

	// Получаем задачу
	task, err := s.ImageRepository.GetByID(imageUUID)
	if err != nil {
		return nil, ErrImageNotFound
	}

	// Применяем обработку через сервис
	if err := s.ImageService.ApplyProcessing(task, req); err != nil {
		return nil, ErrFailedToProcessImage
	}

	return map[string]interface{}{
		"message":   "Обработка начата",
		"next_step": "generate_schema",
	}, nil
}

// GenerateSchema создает финальную схему мозаики
func (s *PublicService) GenerateSchema(imageID string, req types.GenerateSchemaRequest) (map[string]interface{}, error) {
	imageUUID, err := uuid.Parse(imageID)
	if err != nil {
		return nil, ErrInvalidImageID
	}

	// Получаем задачу
	task, err := s.ImageRepository.GetByID(imageUUID)
	if err != nil {
		return nil, ErrImageNotFound
	}

	// Генерируем схему через сервис
	schemaPath, err := s.ImageService.GenerateSchema(task, req)
	if err != nil {
		return nil, ErrFailedToGenerateSchema
	}

	// Обновляем купон как использованный
	coupon, _ := s.CouponRepository.GetByID(task.CouponID)
	coupon.Status = "used"
	coupon.SchemaURL = &schemaPath
	completedAt := time.Now()
	coupon.CompletedAt = &completedAt
	s.CouponRepository.Update(coupon)

	return map[string]interface{}{
		"message":    "Схема успешно создана",
		"schema_url": schemaPath,
		"actions":    []string{"download", "send_email"},
	}, nil
}

// GetImagePreview возвращает превью изображения
func (s *PublicService) GetImagePreview(imageID string) (map[string]interface{}, error) {
	imageUUID, err := uuid.Parse(imageID)
	if err != nil {
		return nil, ErrInvalidImageID
	}

	// Получаем задачу
	task, err := s.ImageRepository.GetByID(imageUUID)
	if err != nil {
		return nil, ErrImageNotFound
	}

	var previewURL *string
	if task.PreviewPath != nil {
		previewURL = task.PreviewPath
	}

	return map[string]interface{}{
		"id":           task.ID,
		"status":       task.Status,
		"preview_url":  previewURL,
		"original_url": task.OriginalImagePath,
	}, nil
}

// GetProcessingStatus возвращает статус обработки
func (s *PublicService) GetProcessingStatus(imageID string) (map[string]interface{}, error) {
	imageUUID, err := uuid.Parse(imageID)
	if err != nil {
		return nil, ErrInvalidImageID
	}

	// Получаем задачу
	task, err := s.ImageRepository.GetByID(imageUUID)
	if err != nil {
		return nil, ErrImageNotFound
	}

	progress := 0
	switch task.Status {
	case "queued":
		progress = 10
	case "processing":
		progress = 50
	case "completed":
		progress = 100
	case "failed":
		progress = 0
	}

	return map[string]interface{}{
		"id":           task.ID,
		"status":       task.Status,
		"progress":     progress,
		"error":        task.ErrorMessage,
		"started_at":   task.StartedAt,
		"completed_at": task.CompletedAt,
	}, nil
}

// GetImageForDownload возвращает задачу для скачивания
func (s *PublicService) GetImageForDownload(imageID string) (*image.Image, error) {
	imageUUID, err := uuid.Parse(imageID)
	if err != nil {
		return nil, ErrInvalidImageID
	}

	// Получаем задачу
	task, err := s.ImageRepository.GetByID(imageUUID)
	if err != nil {
		return nil, ErrImageNotFound
	}

	if task.Status != "completed" {
		return nil, ErrSchemaNotReady
	}

	// Проверяем наличие результата
	if task.ResultPath == nil {
		return nil, ErrSchemaFileNotFound
	}

	return task, nil
}

// SendSchemaToEmail отправляет схему на email
func (s *PublicService) SendSchemaToEmail(imageID string, req SendEmailRequest) (map[string]interface{}, error) {
	imageUUID, err := uuid.Parse(imageID)
	if err != nil {
		return nil, ErrInvalidImageID
	}

	// Получаем задачу
	task, err := s.ImageRepository.GetByID(imageUUID)
	if err != nil {
		return nil, ErrImageNotFound
	}

	if task.Status != "completed" {
		return nil, ErrSchemaNotReady
	}

	// TODO: Отправляем email через сервис
	// if err := s.EmailService.SendSchemaEmail(req.Email, task); err != nil {
	// 	return nil, ErrFailedToSendEmail
	// }

	return map[string]interface{}{
		"message": "Схема успешно отправлена на email",
	}, nil
}

// PurchaseCoupon покупает новый купон онлайн
func (s *PublicService) PurchaseCoupon(req PurchaseCouponRequest) (map[string]interface{}, error) {
	// Создаем новый купон
	newCoupon := &coupon.Coupon{
		Size:      req.Size,
		Style:     req.Style,
		Status:    "purchased",
		UserEmail: &req.Email,
	}

	// Генерируем код купона для собственных купонов (код партнера 0000)
	code, err := randomCouponCode.GenerateUniqueCouponCode("0000", s.CouponRepository)
	if err != nil {
		s.Logger.Error().Err(err).Msg("Failed to generate unique coupon code")
		return nil, ErrFailedToGenerateCouponCode
	}
	newCoupon.Code = code

	if err := s.CouponRepository.Create(newCoupon); err != nil {
		return nil, ErrFailedToCreateCoupon
	}

	// TODO: Интеграция с платежной системой

	return map[string]interface{}{
		"message":     "Купон успешно куплен",
		"coupon_code": newCoupon.Code,
		"coupon_id":   newCoupon.ID,
	}, nil
}

// GetAvailableSizes возвращает доступные размеры
func (s *PublicService) GetAvailableSizes() []map[string]interface{} {
	return []map[string]interface{}{
		{"size": "21x30", "title": "21×30 см", "price": 1500},
		{"size": "30x40", "title": "30×40 см", "price": 2000},
		{"size": "40x40", "title": "40×40 см", "price": 2200},
		{"size": "40x50", "title": "40×50 см", "price": 2500},
		{"size": "40x60", "title": "40×60 см", "price": 2800},
		{"size": "50x70", "title": "50×70 см", "price": 3500},
	}
}

// GetAvailableStyles возвращает доступные стили
func (s *PublicService) GetAvailableStyles() []map[string]interface{} {
	return []map[string]interface{}{
		{"style": "grayscale", "title": "Оттенки серого", "description": "Классическая обработка в оттенках серого"},
		{"style": "skin_tones", "title": "Оттенки телесного", "description": "Подходит для портретов"},
		{"style": "pop_art", "title": "Поп-арт", "description": "Яркие насыщенные цвета"},
		{"style": "max_colors", "title": "Максимум цветов", "description": "Максимальное количество оттенков"},
	}
}

// Вспомогательные функции

func isNumeric(s string) bool {
	for _, char := range s {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}

func isValidImageType(file *multipart.FileHeader) bool {
	ext := strings.ToLower(filepath.Ext(file.Filename))
	return ext == ".jpg" || ext == ".jpeg" || ext == ".png"
}

func (s *PublicService) saveUploadedFile(file *multipart.FileHeader, couponID uuid.UUID) (string, error) {
	// TODO: Реализовать сохранение файла
	// Создать директорию uploads/images/[couponID]/
	// Сохранить файл с оригинальным именем
	// Вернуть путь к сохраненному файлу

	filename := file.Filename
	path := filepath.Join("uploads", "images", couponID.String(), filename)

	// Реализация сохранения файла будет в image service
	return path, nil
}
