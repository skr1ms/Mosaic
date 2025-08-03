package public

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/skr1ms/mosaic/internal/coupon"
	"github.com/skr1ms/mosaic/internal/image"
	"github.com/skr1ms/mosaic/internal/partner"
	"github.com/skr1ms/mosaic/internal/payment"
	"github.com/skr1ms/mosaic/internal/types"
	"github.com/skr1ms/mosaic/pkg/email"
	"github.com/skr1ms/mosaic/pkg/randomCouponCode"
)

type PublicServiceDeps struct {
	CouponRepository  *coupon.CouponRepository
	ImageRepository   *image.ImageRepository
	PartnerRepository *partner.PartnerRepository
	ImageService      *image.ImageService
	PaymentService    *payment.PaymentService
	EmailService      *email.Mailer
}

type PublicService struct {
	deps *PublicServiceDeps
}

func NewPublicService(deps *PublicServiceDeps) *PublicService {
	return &PublicService{
		deps: deps,
	}
}

// GetPartnerByDomain возвращает публичную информацию о партнере по домену
func (s *PublicService) GetPartnerByDomain(domain string) (map[string]interface{}, error) {
	partner, err := s.deps.PartnerRepository.GetByDomain(context.Background(), domain)
	if err != nil {
		return nil, fmt.Errorf("failed to get partner by domain: %w", err)
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
		return nil, fmt.Errorf("invalid coupon code")
	}

	coupon, err := s.deps.CouponRepository.GetByCode(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("coupon not found: %w", err)
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
	coupon, err := s.deps.CouponRepository.GetByCode(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("coupon not found: %w", err)
	}

	if coupon.Status != "new" {
		return nil, fmt.Errorf("coupon already used")
	}

	// Активируем купон
	coupon.Status = "activated"
	coupon.UserEmail = &req.Email
	now := time.Now()
	coupon.ActivatedAt = &now

	if err := s.deps.CouponRepository.Update(context.Background(), coupon); err != nil {
		return nil, fmt.Errorf("failed to activate coupon: %w", err)
	}

	return map[string]interface{}{
		"message":   "Купон успешно активирован",
		"coupon_id": coupon.ID,
		"next_step": "upload_image",
	}, nil
}

// UploadImage загружает изображение для обработки (устаревший метод, используйте ImageService)
// Оставлен для обратной совместимости
func (s *PublicService) UploadImage(couponID string, file *multipart.FileHeader) (map[string]interface{}, error) {
	couponUUID, err := uuid.Parse(couponID)
	if err != nil {
		return nil, fmt.Errorf("invalid coupon id: %w", err)
	}

	// Получаем купон для получения email пользователя
	coupon, err := s.deps.CouponRepository.GetByID(context.Background(), couponUUID)
	if err != nil {
		return nil, fmt.Errorf("coupon not found: %w", err)
	}

	// Используем новый ImageService для загрузки
	imageRecord, err := s.deps.ImageService.UploadImage(context.Background(), couponUUID, file, *coupon.UserEmail)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"message":      "Изображение успешно загружено",
		"image_id":     imageRecord.ID,
		"next_step":    "edit_image",
		"coupon_size":  coupon.Size,
		"coupon_style": coupon.Style,
	}, nil
}

// EditImage применяет редактирование к изображению (устаревший метод, используйте ImageService)
// Оставлен для обратной совместимости
func (s *PublicService) EditImage(imageID string, req types.EditImageRequest) (map[string]interface{}, error) {
	imageUUID, err := uuid.Parse(imageID)
	if err != nil {
		return nil, fmt.Errorf("invalid image id: %w", err)
	}

	// Конвертируем в новый формат параметров
	editParams := image.ImageEditParams{
		CropX:      req.CropX,
		CropY:      req.CropY,
		CropWidth:  req.CropWidth,
		CropHeight: req.CropHeight,
		Rotation:   req.Rotation,
		Scale:      req.Scale,
	}

	// Используем новый ImageService для редактирования
	if err := s.deps.ImageService.EditImage(context.Background(), imageUUID, editParams); err != nil {
		return nil, fmt.Errorf("failed to edit image: %w", err)
	}

	// Получаем статус для возврата URL превью
	status, err := s.deps.ImageService.GetImageStatus(context.Background(), imageUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get image status: %w", err)
	}

	previewURL := ""
	if status.EditedURL != nil {
		previewURL = *status.EditedURL
	}

	return map[string]interface{}{
		"message":     "Изображение успешно отредактировано",
		"next_step":   "choose_style",
		"preview_url": previewURL,
	}, nil
}

// ProcessImage применяет стиль обработки к изображению (устаревший метод, используйте ImageService)
// Оставлен для обратной совместимости
func (s *PublicService) ProcessImage(imageID string, req types.ProcessImageRequest) (map[string]interface{}, error) {
	imageUUID, err := uuid.Parse(imageID)
	if err != nil {
		return nil, fmt.Errorf("invalid image id: %w", err)
	}

	// Конвертируем в новый формат параметров
	processParams := image.ProcessingParams{
		Style:      req.Style,
		UseAI:      req.UseAI,
		Lighting:   req.Lighting,
		Contrast:   req.Contrast,
		Brightness: req.Brightness,
		Saturation: req.Saturation,
	}

	// Запускаем обработку асинхронно
	go func() {
		if err := s.deps.ImageService.ProcessImage(context.Background(), imageUUID, processParams); err != nil {
			// Логируем ошибку, но не прерываем выполнение
			fmt.Printf("Failed to process image %s: %v\n", imageUUID.String(), err)
		}
	}()

	// Получаем статус для возврата URL
	status, err := s.deps.ImageService.GetImageStatus(context.Background(), imageUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get image status: %w", err)
	}

	previewURL := ""
	originalURL := ""
	if status.PreviewURL != nil {
		previewURL = *status.PreviewURL
	}
	if status.OriginalURL != nil {
		originalURL = *status.OriginalURL
	}

	return map[string]interface{}{
		"message":      "Обработка запущена",
		"next_step":    "generate_schema",
		"preview_url":  previewURL,
		"original_url": originalURL,
	}, nil
}

// GetImagePreview возвращает превью изображения
func (s *PublicService) GetImagePreview(imageID string) (map[string]interface{}, error) {
	imageUUID, err := uuid.Parse(imageID)
	if err != nil {
		return nil, fmt.Errorf("invalid image id: %w", err)
	}

	// Получаем задачу
	task, err := s.deps.ImageRepository.GetByID(context.Background(), imageUUID)
	if err != nil {
		return nil, fmt.Errorf("image not found: %w", err)
	}

	// Используем новый ImageService для получения статуса с URL
	status, err := s.deps.ImageService.GetImageStatus(context.Background(), imageUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get image status: %w", err)
	}

	return map[string]interface{}{
		"id":           task.ID,
		"status":       task.Status,
		"preview_url":  status.PreviewURL,
		"original_url": status.OriginalURL,
	}, nil
}

// GetProcessingStatus возвращает статус обработки
func (s *PublicService) GetProcessingStatus(imageID string) (map[string]interface{}, error) {
	imageUUID, err := uuid.Parse(imageID)
	if err != nil {
		return nil, fmt.Errorf("invalid image id: %w", err)
	}

	// Получаем задачу
	task, err := s.deps.ImageRepository.GetByID(context.Background(), imageUUID)
	if err != nil {
		return nil, fmt.Errorf("image not found: %w", err)
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
		return nil, fmt.Errorf("invalid image id: %w", err)
	}

	// Получаем задачу
	task, err := s.deps.ImageRepository.GetByID(context.Background(), imageUUID)
	if err != nil {
		return nil, fmt.Errorf("image not found: %w", err)
	}

	if task.Status != "completed" {
		return nil, fmt.Errorf("schema not ready")
	}

	// Проверяем наличие результата через ImageService
	status, err := s.deps.ImageService.GetImageStatus(context.Background(), imageUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get image status: %w", err)
	}

	if status.SchemaURL == nil {
		return nil, fmt.Errorf("schema file not found")
	}

	return task, nil
}

// SendSchemaToEmail отправляет схему на email
func (s *PublicService) SendSchemaToEmail(imageID string, req SendEmailRequest) (map[string]interface{}, error) {
	imageUUID, err := uuid.Parse(imageID)
	if err != nil {
		return nil, fmt.Errorf("invalid image id: %w", err)
	}

	// Получаем задачу
	task, err := s.deps.ImageRepository.GetByID(context.Background(), imageUUID)
	if err != nil {
		return nil, fmt.Errorf("image not found: %w", err)
	}

	if task.Status != "completed" {
		return nil, fmt.Errorf("schema not ready")
	}

	if task.SchemaS3Key == nil {
		return nil, fmt.Errorf("schema file not found")
	}

	// Получаем информацию о купоне для кода купона
	coupon, err := s.deps.CouponRepository.GetByID(context.Background(), task.CouponID)
	if err != nil {
		return nil, fmt.Errorf("coupon not found: %w", err)
	}

	// Генерируем presigned URL для скачивания схемы
	schemaURL, err := s.deps.ImageService.GetImageStatus(context.Background(), imageUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get schema URL: %w", err)
	}

	if schemaURL.SchemaURL == nil {
		return nil, fmt.Errorf("schema URL not available")
	}

	// Отправляем email
	err = s.deps.EmailService.SendSchemaEmail(req.Email, *schemaURL.SchemaURL, coupon.Code)
	if err != nil {
		return nil, fmt.Errorf("failed to send email: %w", err)
	}

	return map[string]interface{}{
		"message": "Schema successfully sent to email",
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
	code, err := randomCouponCode.GenerateUniqueCouponCode("0000", s.deps.CouponRepository)
	if err != nil {
		return nil, fmt.Errorf("failed to generate coupon code: %w", err)
	}
	newCoupon.Code = code

	if err := s.deps.CouponRepository.Create(context.Background(), newCoupon); err != nil {
		return nil, fmt.Errorf("failed to create coupon: %w", err)
	}

	// Используем PaymentService для создания заказа и получения URL оплаты
	paymentReq := &payment.PurchaseCouponRequest{
		Size:      req.Size,
		Style:     req.Style,
		Email:     req.Email,
		ReturnURL: "http://localhost:3000/payment/success", // TODO: получать из конфига
		Language:  "ru",
	}

	response, err := s.deps.PaymentService.PurchaseCoupon(context.Background(), paymentReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment order: %w", err)
	}

	if !response.Success {
		return map[string]interface{}{
			"success": false,
			"message": response.Message,
		}, nil
	}

	return map[string]interface{}{
		"success":     true,
		"order_id":    response.OrderID,
		"payment_url": response.PaymentURL,
		"message":     "Заказ создан, переходите по ссылке для оплаты",
	}, nil
}

// GetAvailableSizes возвращает доступные размеры
func (s *PublicService) GetAvailableSizes() []map[string]interface{} {
	return []map[string]interface{}{
		{"size": "21x30", "title": "21×30 см", "price": int(payment.FixedPriceRub)},
		{"size": "30x40", "title": "30×40 см", "price": int(payment.FixedPriceRub)},
		{"size": "40x40", "title": "40×40 см", "price": int(payment.FixedPriceRub)},
		{"size": "40x50", "title": "40×50 см", "price": int(payment.FixedPriceRub)},
		{"size": "40x60", "title": "40×60 см", "price": int(payment.FixedPriceRub)},
		{"size": "50x70", "title": "50×70 см", "price": int(payment.FixedPriceRub)},
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
