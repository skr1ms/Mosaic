package public

import (
	"context"
	"fmt"
	"mime/multipart"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/skr1ms/mosaic/internal/image"
	"github.com/skr1ms/mosaic/internal/payment"
	"github.com/skr1ms/mosaic/internal/types"
	"github.com/skr1ms/mosaic/pkg/goroutine"
)

type PublicServiceDeps struct {
	CouponRepository  CouponRepositoryInterface
	ImageRepository   ImageRepositoryInterface
	PartnerRepository PartnerRepositoryInterface
	ImageService      ImageServiceInterface
	PaymentService    PaymentServiceInterface
	EmailService      EmailServiceInterface
	Config            ConfigInterface
	RecaptchaSiteKey  string
}

type PublicService struct {
	deps             *PublicServiceDeps
	goroutineManager *goroutine.Manager
	processingPool   *goroutine.WorkerPool
}

func NewPublicService(deps *PublicServiceDeps) *PublicService {
	s := &PublicService{
		deps: deps,
	}

	s.goroutineManager = goroutine.NewManager(context.Background())
	s.processingPool = s.goroutineManager.NewWorkerPool("public_processing", 3, 50)

	return s
}

// Repository access methods
func (s *PublicService) GetCouponRepository() CouponRepositoryInterface {
	return s.deps.CouponRepository
}

func (s *PublicService) GetImageRepository() ImageRepositoryInterface {
	return s.deps.ImageRepository
}

func (s *PublicService) GetPartnerRepository() PartnerRepositoryInterface {
	return s.deps.PartnerRepository
}

func (s *PublicService) GetImageService() ImageServiceInterface {
	return s.deps.ImageService
}

func (s *PublicService) GetPaymentService() PaymentServiceInterface {
	return s.deps.PaymentService
}

// GetPartnerByDomain returns public partner information by domain
func (s *PublicService) GetPartnerByDomain(domain string) (map[string]any, error) {
	partner, err := s.deps.PartnerRepository.GetByDomain(context.Background(), domain)
	if err != nil {
		return nil, fmt.Errorf("failed to get partner by domain: %w", err)
	}

	return map[string]any{
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

// GetCouponByCode returns coupon information by code
func (s *PublicService) GetCouponByCode(code string) (map[string]any, error) {
	cleanCode := strings.ReplaceAll(code, "-", "")
	if len(cleanCode) != 12 || !isNumeric(cleanCode) {
		return nil, fmt.Errorf("invalid coupon code: length=%d, isNumeric=%v", len(cleanCode), isNumeric(cleanCode))
	}

	coupon, err := s.deps.CouponRepository.GetByCode(context.Background(), cleanCode)
	if err != nil {
		return nil, fmt.Errorf("coupon not found: %w", err)
	}

	// Get partner information for the coupon
	partner, err := s.deps.PartnerRepository.GetByID(context.Background(), coupon.PartnerID)
	if err != nil {
		// If partner not found, continue without partner info
		return map[string]any{
			"id":     coupon.ID,
			"code":   coupon.Code,
			"size":   coupon.Size,
			"style":  coupon.Style,
			"status": coupon.Status,
			"valid":  coupon.Status == "new",
		}, nil
	}

	return map[string]any{
		"id":           coupon.ID,
		"code":         coupon.Code,
		"size":         coupon.Size,
		"style":        coupon.Style,
		"status":       coupon.Status,
		"valid":        coupon.Status == "new",
		"partner_id":   partner.ID,
		"partner_code": partner.PartnerCode,
		"partner_domain": partner.Domain,
	}, nil
}

// ActivateCoupon activates coupon for subsequent processing
func (s *PublicService) ActivateCoupon(code string) (map[string]any, error) {
	cleanCode := strings.ReplaceAll(code, "-", "")

	coupon, err := s.deps.CouponRepository.GetByCode(context.Background(), cleanCode)
	if err != nil {
		return nil, fmt.Errorf("coupon not found: %w", err)
	}

	if coupon.Status == "used" || coupon.Status == "activated" || coupon.Status == "completed" {
		return nil, fmt.Errorf("coupon already used, activated or completed")
	}

	coupon.Status = "activated"
	now := time.Now()
	coupon.ActivatedAt = &now

	if err := s.deps.CouponRepository.Update(context.Background(), coupon); err != nil {
		return nil, fmt.Errorf("failed to activate coupon: %w", err)
	}

	return map[string]any{
		"message":   "Купон успешно активирован",
		"coupon_id": coupon.ID,
		"next_step": "upload_image",
	}, nil
}

// UploadImage uploads image for processing (uses ImageService)
func (s *PublicService) UploadImage(couponID string, file *multipart.FileHeader) (map[string]any, error) {
	couponUUID, err := uuid.Parse(couponID)
	if err != nil {
		return nil, fmt.Errorf("invalid coupon id: %w", err)
	}

	coupon, err := s.deps.CouponRepository.GetByID(context.Background(), couponUUID)
	if err != nil {
		return nil, fmt.Errorf("coupon not found: %w", err)
	}

	// Use empty email if not set (removed email requirement)
	userEmail := ""
	if coupon.UserEmail != nil {
		userEmail = *coupon.UserEmail
	}

	imageRecord, err := s.deps.ImageService.UploadImage(context.Background(), couponUUID, file, userEmail)
	if err != nil {
		return nil, fmt.Errorf("failed to upload image: %w", err)
	}

	return map[string]any{
		"message":      "Изображение успешно загружено",
		"image_id":     imageRecord.ID,
		"next_step":    "edit_image",
		"coupon_size":  coupon.Size,
		"coupon_style": coupon.Style,
	}, nil
}

// EditImage applies editing to image (deprecated method, use ImageService)
// Kept for backward compatibility
func (s *PublicService) EditImage(imageID string, req types.EditImageRequest) (map[string]any, error) {
	imageUUID, err := uuid.Parse(imageID)
	if err != nil {
		return nil, fmt.Errorf("invalid image id: %w", err)
	}

	editParams := image.ImageEditParams{
		CropX:      req.CropX,
		CropY:      req.CropY,
		CropWidth:  req.CropWidth,
		CropHeight: req.CropHeight,
		Rotation:   req.Rotation,
		Scale:      req.Scale,
	}

	if err := s.deps.ImageService.EditImage(context.Background(), imageUUID, editParams); err != nil {
		return nil, fmt.Errorf("failed to edit image: %w", err)
	}

	status, err := s.deps.ImageService.GetImageStatus(context.Background(), imageUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get image status: %w", err)
	}

	previewURL := ""
	if status.EditedURL != nil {
		previewURL = *status.EditedURL
	}

	return map[string]any{
		"message":     "Изображение успешно отредактировано",
		"next_step":   "choose_style",
		"preview_url": previewURL,
	}, nil
}

// ProcessImage applies processing style to image
func (s *PublicService) ProcessImage(imageID string, req types.ProcessImageRequest) (map[string]any, error) {
	imageUUID, err := uuid.Parse(imageID)
	if err != nil {
		return nil, fmt.Errorf("invalid image id: %w", err)
	}

	processParams := &image.ProcessingParams{
		Style:      req.Style,
		UseAI:      req.UseAI,
		Lighting:   req.Lighting,
		Contrast:   req.Contrast,
		Brightness: req.Brightness,
		Saturation: req.Saturation,
		Settings:   make(map[string]any),
	}

	s.processImageAsync(imageUUID, processParams)

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

	return map[string]any{
		"message":      "Обработка запущена",
		"next_step":    "generate_schema",
		"preview_url":  previewURL,
		"original_url": originalURL,
	}, nil
}

// GetImagePreview returns image preview
func (s *PublicService) GetImagePreview(imageID string) (map[string]any, error) {
	imageUUID, err := uuid.Parse(imageID)
	if err != nil {
		return nil, fmt.Errorf("invalid image id: %w", err)
	}

	task, err := s.deps.ImageRepository.GetByID(context.Background(), imageUUID)
	if err != nil {
		return nil, fmt.Errorf("image not found: %w", err)
	}

	status, err := s.deps.ImageService.GetImageStatus(context.Background(), imageUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get image status: %w", err)
	}

	return map[string]any{
		"id":           task.ID,
		"status":       task.Status,
		"preview_url":  status.PreviewURL,
		"original_url": status.OriginalURL,
	}, nil
}

// GetProcessingStatus returns processing status
func (s *PublicService) GetProcessingStatus(imageID string) (map[string]any, error) {
	imageUUID, err := uuid.Parse(imageID)
	if err != nil {
		return nil, fmt.Errorf("invalid image id: %w", err)
	}

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

	return map[string]any{
		"id":           task.ID,
		"status":       task.Status,
		"progress":     progress,
		"error":        task.ErrorMessage,
		"started_at":   task.StartedAt,
		"completed_at": task.CompletedAt,
	}, nil
}

// GetImageForDownload returns task for download
func (s *PublicService) GetImageForDownload(imageID string) (*image.Image, error) {
	imageUUID, err := uuid.Parse(imageID)
	if err != nil {
		return nil, fmt.Errorf("invalid image id: %w", err)
	}

	task, err := s.deps.ImageRepository.GetByID(context.Background(), imageUUID)
	if err != nil {
		return nil, fmt.Errorf("image not found: %w", err)
	}

	if task.Status != "completed" {
		return nil, fmt.Errorf("schema not ready")
	}

	status, err := s.deps.ImageService.GetImageStatus(context.Background(), imageUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get image status: %w", err)
	}

	if status.ZipURL == nil {
		return nil, fmt.Errorf("schema file not found")
	}

	return task, nil
}

// SendSchemaToEmail sends schema to email
func (s *PublicService) SendSchemaToEmail(imageID string, req SendEmailRequest) (map[string]any, error) {
	imageUUID, err := uuid.Parse(imageID)
	if err != nil {
		return nil, fmt.Errorf("invalid image id: %w", err)
	}

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

	coupon, err := s.deps.CouponRepository.GetByID(context.Background(), task.CouponID)
	if err != nil {
		return nil, fmt.Errorf("coupon not found: %w", err)
	}

	schemaURL, err := s.deps.ImageService.GetImageStatus(context.Background(), imageUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get schema URL: %w", err)
	}

	if schemaURL.ZipURL == nil {
		return nil, fmt.Errorf("schema URL not available")
	}

	err = s.deps.EmailService.SendSchemaEmail(req.Email, *schemaURL.ZipURL, coupon.Code)
	if err != nil {
		return nil, fmt.Errorf("failed to send email: %w", err)
	}

	return map[string]any{
		"message": "Schema successfully sent to email",
	}, nil
}

// PurchaseCoupon purchases new coupon online (coupon creation happens after payment in PaymentService)
func (s *PublicService) PurchaseCoupon(req PurchaseCouponRequest) (map[string]any, error) {
	if s.deps == nil {
		return nil, fmt.Errorf("service dependencies are not initialized")
	}
	if s.deps.PaymentService == nil {
		return nil, fmt.Errorf("payment service is not initialized")
	}
	if s.deps.Config == nil {
		return nil, fmt.Errorf("config is not initialized")
	}

	serverConfig := s.deps.Config.GetServerConfig()

	paymentReq := &payment.PurchaseCouponRequest{
		Size:      req.Size,
		Style:     req.Style,
		Email:     req.Email,
		ReturnURL: serverConfig.PaymentSuccessURL,
		Language:  "ru",
	}

	response, err := s.deps.PaymentService.PurchaseCoupon(context.Background(), paymentReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment order: %w", err)
	}

	if !response.Success {
		return map[string]any{
			"success": false,
			"message": response.Message,
		}, nil
	}

	return map[string]any{
		"success":     true,
		"order_id":    response.OrderID,
		"payment_url": response.PaymentURL,
		"message":     "Заказ создан, переходите по ссылке для оплаты",
	}, nil
}

// GetAvailableSizes returns available sizes
func (s *PublicService) GetAvailableSizes() []map[string]any {
	return []map[string]any{
		{"size": "21x30", "title": "21×30 см", "price": int(payment.FixedPriceRub)},
		{"size": "30x40", "title": "30×40 см", "price": int(payment.FixedPriceRub)},
		{"size": "40x40", "title": "40×40 см", "price": int(payment.FixedPriceRub)},
		{"size": "40x50", "title": "40×50 см", "price": int(payment.FixedPriceRub)},
		{"size": "40x60", "title": "40×60 см", "price": int(payment.FixedPriceRub)},
		{"size": "50x70", "title": "50×70 см", "price": int(payment.FixedPriceRub)},
	}
}

// GetAvailableStyles returns available styles
func (s *PublicService) GetAvailableStyles() []map[string]any {
	return []map[string]any{
		{"style": "grayscale", "title": "Оттенки серого", "description": "Классическая обработка в оттенках серого"},
		{"style": "skin_tones", "title": "Оттенки телесного", "description": "Подходит для портретов"},
		{"style": "pop_art", "title": "Поп-арт", "description": "Яркие насыщенные цвета"},
		{"style": "max_colors", "title": "Максимум цветов", "description": "Максимальное количество оттенков"},
	}
}

// GetRecaptchaSiteKey returns public reCAPTCHA site key
func (s *PublicService) GetRecaptchaSiteKey() string {
	return s.deps.Config.GetRecaptchaConfig().SiteKey
}

// Close releases service resources
func (s *PublicService) Close() error {
	if s.goroutineManager != nil {
		return s.goroutineManager.Close()
	}
	return nil
}

// GetMetrics returns service metrics
func (s *PublicService) GetMetrics() goroutine.Metrics {
	if s.goroutineManager != nil {
		return s.goroutineManager.GetMetrics()
	}
	return goroutine.Metrics{}
}

func isNumeric(s string) bool {
	if len(s) == 0 {
		return false
	}

	cleanCode := strings.ReplaceAll(s, "-", "")

	if len(cleanCode) != 12 {
		return false
	}

	for _, char := range cleanCode {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}

// processImageAsync runs image processing asynchronously
func (s *PublicService) processImageAsync(imageUUID uuid.UUID, processParams *image.ProcessingParams) {
	s.processingPool.SubmitTask(func() {
		if err := s.deps.ImageService.ProcessImage(context.Background(), imageUUID, processParams); err != nil {
			fmt.Printf("Failed to process image: %v, imageUUID: %s\n", err, imageUUID.String())
		}
	})
}
