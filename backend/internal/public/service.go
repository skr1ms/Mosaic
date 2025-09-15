package public

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/google/uuid"
	internalImage "github.com/skr1ms/mosaic/internal/image"
	"github.com/skr1ms/mosaic/internal/payment"
	"github.com/skr1ms/mosaic/internal/types"
	"github.com/skr1ms/mosaic/pkg/goroutine"
	"github.com/skr1ms/mosaic/pkg/marketplace"
)

type PublicServiceDeps struct {
	CouponRepository  CouponRepositoryInterface
	ImageRepository   ImageRepositoryInterface
	PartnerRepository PartnerRepositoryInterface
	ImageService      ImageServiceInterface
	PaymentService    PaymentServiceInterface
	PublicRepository  PublicRepositoryInterface
	EmailService      EmailServiceInterface
	Config            ConfigInterface
	S3Client          S3ClientInterface
	AIClient          AIClientInterface
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

func (s *PublicService) GetPublicRepository() PublicRepositoryInterface {
	return s.deps.PublicRepository
}

func (s *PublicService) GetS3Client() S3ClientInterface {
	return s.deps.S3Client
}

func (s *PublicService) GetEmailService() EmailServiceInterface {
	return s.deps.EmailService
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
			"valid":  true, // Simplified validation - only existence check
		}, nil
	}

	return map[string]any{
		"id":             coupon.ID,
		"code":           coupon.Code,
		"size":           coupon.Size,
		"style":          coupon.Style,
		"status":         coupon.Status,
		"valid":          true, // Simplified validation - only existence check
		"partner_id":     partner.ID,
		"partner_code":   partner.PartnerCode,
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

	// Simplified validation - only coupon existence check
	// Removed status and email validation
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
	// If no coupon provided, treat as preview-only upload
	if couponID == "" {
		// Generate a temporary image ID for preview purposes
		tempImageID := uuid.New()

		return map[string]any{
			"message":      "Изображение успешно загружено для предпросмотра",
			"image_id":     tempImageID,
			"next_step":    "generate_preview",
			"coupon_size":  "30x40",   // Default size for previews
			"coupon_style": "classic", // Default style for previews
			"is_preview":   true,
		}, nil
	}

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
		"is_preview":   false,
	}, nil
}

// EditImage applies editing to image (deprecated method, use ImageService)
// Kept for backward compatibility
func (s *PublicService) EditImage(imageID string, req types.EditImageRequest) (map[string]any, error) {
	imageUUID, err := uuid.Parse(imageID)
	if err != nil {
		return nil, fmt.Errorf("invalid image id: %w", err)
	}

	editParams := internalImage.ImageEditParams{
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

	processParams := &internalImage.ProcessingParams{
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
func (s *PublicService) GetImageForDownload(imageID string) (*internalImage.Image, error) {
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

// GetPartnerArticleGrid returns article grid for a partner
func (s *PublicService) GetPartnerArticleGrid(partnerID uuid.UUID) (map[string]any, error) {
	// Check if partner exists
	partner, err := s.deps.PartnerRepository.GetByID(context.Background(), partnerID)
	if err != nil {
		return nil, fmt.Errorf("partner not found: %w", err)
	}

	// Get real partner articles from database
	articleGrid, err := s.deps.PartnerRepository.GetArticleGrid(context.Background(), partnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get article grid: %w", err)
	}

	// If no articles exist, initialize empty structure
	if len(articleGrid) == 0 {
		articleGrid = map[string]map[string]map[string]string{
			"ozon": {
				"grayscale":  make(map[string]string),
				"skin_tones": make(map[string]string),
				"pop_art":    make(map[string]string),
				"max_colors": make(map[string]string),
			},
			"wildberries": {
				"grayscale":  make(map[string]string),
				"skin_tones": make(map[string]string),
				"pop_art":    make(map[string]string),
				"max_colors": make(map[string]string),
			},
		}
	}

	return map[string]any{
		"partner_id":   partner.ID,
		"partner_name": partner.BrandName,
		"article_grid": articleGrid,
	}, nil
}

// GeneratePartnerProductURL generates product URL for partner marketplace
func (s *PublicService) GeneratePartnerProductURL(partnerID uuid.UUID, req GenerateProductURLRequest) (map[string]any, error) {
	// Create marketplace service
	marketplaceRepo := marketplace.NewPartnerRepositoryAdapter(s.deps.PartnerRepository)
	marketplaceService := marketplace.NewService(marketplaceRepo)

	// Convert request to marketplace format
	marketplaceReq := &marketplace.ProductURLRequest{
		PartnerID:   partnerID,
		Marketplace: marketplace.Marketplace(req.Marketplace),
		Size:        req.Size,
		Style:       req.Style,
	}

	response, err := marketplaceService.GenerateProductURL(marketplaceReq)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"url":          response.URL,
		"sku":          response.SKU,
		"has_article":  response.HasArticle,
		"partner_name": response.PartnerName,
		"marketplace":  string(response.Marketplace),
		"size":         response.Size,
		"style":        response.Style,
	}, nil
}

// SearchSchemaPage searches for a specific page in the schema
func (s *PublicService) SearchSchemaPage(ctx context.Context, imageID string, pageNumber int) (*SearchSchemaPageResponse, error) {
	imageUUID, err := uuid.Parse(imageID)
	if err != nil {
		return nil, fmt.Errorf("invalid image ID: %w", err)
	}

	// Get image from database
	img, err := s.deps.ImageRepository.GetByID(ctx, imageUUID)
	if err != nil {
		return nil, fmt.Errorf("image not found: %w", err)
	}

	// Check if schema is ready
	if img.SchemaS3Key == nil {
		return nil, fmt.Errorf("schema not generated yet")
	}

	// Get coupon to check page count
	coupon, err := s.deps.CouponRepository.GetByID(ctx, img.CouponID)
	if err != nil {
		return nil, fmt.Errorf("coupon not found: %w", err)
	}

	// Check if page number is valid
	if pageNumber < 1 || pageNumber > coupon.PageCount {
		return &SearchSchemaPageResponse{
			PageNumber: pageNumber,
			TotalPages: coupon.PageCount,
			Found:      false,
		}, nil
	}

	// Generate URL for specific page
	pageKey := fmt.Sprintf("schemas/%s/page_%d.jpg", imageID, pageNumber)
	pageURL, err := s.deps.S3Client.GetFileURL(ctx, pageKey, 24*time.Hour)
	if err != nil {
		// Try alternative path
		pageKey = fmt.Sprintf("schemas/%s/pages/page_%03d.jpg", imageID, pageNumber)
		pageURL, err = s.deps.S3Client.GetFileURL(ctx, pageKey, 24*time.Hour)
		if err != nil {
			return &SearchSchemaPageResponse{
				PageNumber: pageNumber,
				TotalPages: coupon.PageCount,
				Found:      false,
			}, nil
		}
	}

	return &SearchSchemaPageResponse{
		PageURL:    pageURL,
		PageNumber: pageNumber,
		TotalPages: coupon.PageCount,
		Found:      true,
	}, nil
}

// ReactivateCoupon handles re-access to an already activated coupon
func (s *PublicService) ReactivateCoupon(ctx context.Context, code string) (*ReactivateCouponResponse, error) {
	cleanCode := strings.ReplaceAll(code, "-", "")

	// Get coupon
	coupon, err := s.deps.CouponRepository.GetByCode(ctx, cleanCode)
	if err != nil {
		return nil, fmt.Errorf("coupon not found: %w", err)
	}

	// Check if coupon was activated
	if coupon.Status != "activated" && coupon.Status != "completed" {
		return nil, fmt.Errorf("coupon not activated yet")
	}

	// Get associated image
	img, err := s.deps.ImageRepository.GetByCouponID(ctx, coupon.ID)
	if err != nil {
		return nil, fmt.Errorf("no schema found for this coupon")
	}

	// Get image status
	status, err := s.deps.ImageService.GetImageStatus(ctx, img.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get schema status: %w", err)
	}

	response := &ReactivateCouponResponse{
		ActivatedAt:  coupon.ActivatedAt.Format("02.01.2006"),
		PreviewURL:   "",
		StonesCount:  0,
		ArchiveURL:   "",
		CanDownload:  false,
		CanSendEmail: false,
		PageCount:    coupon.PageCount,
	}

	if status.PreviewURL != nil {
		response.PreviewURL = *status.PreviewURL
	}

	if status.ZipURL != nil {
		response.ArchiveURL = *status.ZipURL
		response.CanDownload = true
		response.CanSendEmail = true
	}

	if coupon.StonesCount != nil {
		response.StonesCount = *coupon.StonesCount
	}

	return response, nil
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
func (s *PublicService) processImageAsync(imageUUID uuid.UUID, processParams *internalImage.ProcessingParams) {
	s.processingPool.SubmitTask(func() {
		if err := s.deps.ImageService.ProcessImage(context.Background(), imageUUID, processParams); err != nil {
			fmt.Printf("Failed to process image: %v, imageUUID: %s\n", err, imageUUID.String())
		}
	})
}

// GeneratePreview generates a single preview with style, lighting and contrast
func (s *PublicService) GeneratePreview(ctx context.Context, file *multipart.FileHeader, size, style, lighting, contrast string) (*PreviewData, error) {
	// Open uploaded file
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	// Decode image
	img, format, err := s.deps.S3Client.Decode(src)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Apply transformations in order
	img = s.resizeImage(img, size)
	img = s.ApplyStyle(img, style)
	img = s.ApplyLighting(img, lighting)
	img = s.ApplyContrast(img, contrast)

	// Convert to bytes
	var buf bytes.Buffer
	switch format {
	case "jpeg", "jpg":
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90})
	case "png":
		err = png.Encode(&buf, img)
	default:
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to encode image: %w", err)
	}

	// Generate unique ID
	previewID := uuid.New()

	// Upload to S3 preview bucket with TTL
	previewKey := fmt.Sprintf("previews/%s.jpg", previewID.String())
	if err := s.deps.S3Client.UploadToPreviewBucket(ctx, previewKey, &buf, int64(buf.Len()), "image/jpeg"); err != nil {
		return nil, fmt.Errorf("failed to upload preview: %w", err)
	}

	// Get public URL for preview
	previewURL := s.deps.S3Client.GetPreviewURL(previewKey)
	if previewURL == "" {
		return nil, fmt.Errorf("failed to get preview URL")
	}

	// Save preview data for database
	previewData := &PreviewData{
		ID:       previewID,
		URL:      previewURL,
		Style:    fmt.Sprintf("%s_%s", style, lighting),
		Contrast: contrast,
		Size:     size,
		S3Key:    previewKey,
	}

	// Save to database
	if err := s.deps.PublicRepository.Create(ctx, previewData); err != nil {
		// If database save fails, cleanup S3
		s.deps.S3Client.SchedulePreviewDeletion(previewKey)
		return nil, fmt.Errorf("failed to save preview to database: %w", err)
	}

	// Schedule automatic deletion after 24 hours (database TTL handles this too)
	s.deps.S3Client.SchedulePreviewDeletion(previewKey)

	return previewData, nil
}

// GenerateStylePreview generates a simple mosaic preview for a specific style
func (s *PublicService) GenerateStylePreview(ctx context.Context, file *multipart.FileHeader, size, style string) (*PreviewData, error) {
	// Open uploaded file
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	// Decode image
	img, format, err := s.deps.S3Client.Decode(src)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Apply transformations for the specific style
	img = s.resizeImage(img, size)
	img = s.ApplyStyle(img, style)
	img = s.ApplyLighting(img, "sun")    // Default lighting for previews
	img = s.ApplyContrast(img, "normal") // Default contrast for previews

	// Convert to bytes
	var buf bytes.Buffer
	switch format {
	case "jpeg", "jpg":
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90})
	case "png":
		err = png.Encode(&buf, img)
	default:
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to encode image: %w", err)
	}

	// Generate unique ID
	previewID := uuid.New()

	// Upload to S3 preview bucket with TTL
	previewKey := fmt.Sprintf("style-previews/%s_%s.jpg", style, previewID.String())
	if err := s.deps.S3Client.UploadToPreviewBucket(ctx, previewKey, &buf, int64(buf.Len()), "image/jpeg"); err != nil {
		return nil, fmt.Errorf("failed to upload preview: %w", err)
	}

	// Get public URL for preview
	previewURL := s.deps.S3Client.GetPreviewURL(previewKey)
	if previewURL == "" {
		return nil, fmt.Errorf("failed to get preview URL")
	}

	// Create preview data for database
	previewData := &PreviewData{
		ID:       previewID,
		URL:      previewURL,
		Style:    style,
		Contrast: "normal", // Default contrast for style previews
		Size:     size,
		S3Key:    previewKey,
	}

	// Save to database
	if err := s.deps.PublicRepository.Create(ctx, previewData); err != nil {
		// If database save fails, cleanup S3
		s.deps.S3Client.SchedulePreviewDeletion(previewKey)
		return nil, fmt.Errorf("failed to save preview to database: %w", err)
	}

	// Schedule automatic deletion after 24 hours (database TTL handles this too)
	s.deps.S3Client.SchedulePreviewDeletion(previewKey)

	return previewData, nil
}

func (s *PublicService) GenerateAIPreview(ctx context.Context, file *multipart.FileHeader, prompt string) (*PreviewData, error) {
	// Open uploaded file
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	// Read file content
	_, err = io.ReadAll(src)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// For now, return a placeholder (implement AI generation later)
	previewID := uuid.New()
	s3Key := fmt.Sprintf("ai-previews/%s.jpg", previewID.String())
	previewURL := s.deps.S3Client.GetPreviewURL(s3Key)

	previewData := &PreviewData{
		ID:       previewID,
		URL:      previewURL,
		Style:    "ai_generated",
		Contrast: "normal",
		Size:     "unknown",
		S3Key:    s3Key,
	}

	// Save to database
	if err := s.deps.PublicRepository.Create(ctx, previewData); err != nil {
		return nil, fmt.Errorf("failed to save AI preview to database: %w", err)
	}

	return previewData, nil
}

// ApplyStyle applies the main processing style to the image
func (s *PublicService) ApplyStyle(img image.Image, style string) image.Image {
	switch style {
	case "grayscale":
		// Convert to grayscale
		return imaging.Grayscale(img)
	case "skin_tones":
		// Optimize for skin tones (portraits)
		img = imaging.AdjustSaturation(img, -20)
		return s.applyColorFilter(img, color.RGBA{255, 220, 200, 20})
	case "pop_art":
		// High saturation and contrast for pop art effect
		img = imaging.AdjustSaturation(img, 50)
		return imaging.AdjustContrast(img, 30)
	case "max_colors":
		// Maximum color range
		img = imaging.AdjustSaturation(img, 20)
		return imaging.AdjustGamma(img, 1.2)
	default:
		return img
	}
}

// ApplyLighting applies lighting effects (sun, moon, venus)
func (s *PublicService) ApplyLighting(img image.Image, lighting string) image.Image {
	switch lighting {
	case "sun":
		// Солнце - яркие жёлтые тона
		return s.applyColorFilter(img, color.RGBA{255, 255, 100, 30})
	case "moon":
		// Луна - холодные синие тона
		return s.applyColorFilter(img, color.RGBA{150, 150, 255, 40})
	case "venus":
		// Венера - тёплые оттенки
		return s.applyColorFilter(img, color.RGBA{255, 200, 150, 50})
	case "mars":
		// Марс - красные тона (дополнительно к ТЗ)
		return s.applyColorFilter(img, color.RGBA{255, 100, 100, 50})
	default:
		return img
	}
}

func (s *PublicService) ApplyContrast(img image.Image, level string) image.Image {
	switch level {
	case "soft":
		return imaging.AdjustContrast(img, -10)
	case "strong":
		return imaging.AdjustContrast(img, 30)
	case "normal":
		return img
	default:
		return img
	}
}

func (s *PublicService) applyColorFilter(img image.Image, filterColor color.RGBA) image.Image {
	bounds := img.Bounds()
	filtered := image.NewRGBA(bounds)

	// Create overlay with filter color
	overlay := image.NewUniform(filterColor)

	// Draw original image
	draw.Draw(filtered, bounds, img, bounds.Min, draw.Src)

	// Apply color filter with transparency
	draw.DrawMask(filtered, bounds, overlay, bounds.Min, nil, bounds.Min, draw.Over)

	return filtered
}

// GenerateAllPreviews generates all 8 base previews + optional 2 AI previews
func (s *PublicService) GenerateAllPreviews(ctx context.Context, imageID string, size string, useAI bool) (*GenerateAllPreviewsResponse, error) {
	imageUUID, err := uuid.Parse(imageID)
	if err != nil {
		return nil, fmt.Errorf("invalid image ID: %w", err)
	}

	// Get original image from database
	img, err := s.deps.ImageRepository.GetByID(ctx, imageUUID)
	if err != nil {
		return nil, fmt.Errorf("image not found: %w", err)
	}

	// Download original image from S3
	imageData, err := s.deps.S3Client.DownloadFile(ctx, img.OriginalImageS3Key)
	if err != nil {
		return nil, fmt.Errorf("failed to download image: %w", err)
	}
	defer imageData.Close()

	// Decode image
	originalImg, format, err := s.deps.S3Client.Decode(imageData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	var previews []PreviewInfo

	// Define combinations according to TZ:
	// 4 styles × 2 contrasts = 8 base previews
	styles := []string{"venus", "sun", "moon", "mars"}
	contrasts := []struct {
		Value string
		Label string
	}{
		{"soft", "мягкий контраст"},
		{"strong", "сильный контраст"},
	}

	styleLabels := map[string]string{
		"venus": "Венера",
		"sun":   "Солнце",
		"moon":  "Луна",
		"mars":  "Марс",
	}

	// Generate 8 base previews
	for _, style := range styles {
		for _, contrast := range contrasts {
			// Process image
			processedImg := s.resizeImage(originalImg, size)
			processedImg = s.ApplyLighting(processedImg, style)
			processedImg = s.ApplyContrast(processedImg, contrast.Value)

			// Encode image
			var buf bytes.Buffer
			switch format {
			case "jpeg", "jpg":
				jpeg.Encode(&buf, processedImg, &jpeg.Options{Quality: 90})
			case "png":
				png.Encode(&buf, processedImg)
			default:
				jpeg.Encode(&buf, processedImg, &jpeg.Options{Quality: 90})
			}

			// Upload to S3 preview bucket with TTL
			previewID := uuid.New().String()
			previewKey := fmt.Sprintf("previews/%s/%s_%s_%s.jpg", imageID, style, contrast.Value, previewID[:8])
			if err := s.deps.S3Client.UploadToPreviewBucket(ctx, previewKey, &buf, int64(buf.Len()), "image/jpeg"); err != nil {
				continue // Skip failed uploads
			}

			// Schedule automatic deletion after 30 minutes
			s.deps.S3Client.SchedulePreviewDeletion(previewKey)

			// Get URL for preview
			previewURL := s.deps.S3Client.GetPreviewURL(previewKey)

			previews = append(previews, PreviewInfo{
				ID:       previewID,
				URL:      previewURL,
				Style:    style,
				Contrast: contrast.Value,
				Label:    fmt.Sprintf("%s (%s)", styleLabels[style], contrast.Label),
				IsAI:     false,
			})
		}
	}

	// Generate AI previews if requested
	if useAI && s.deps.AIClient != nil {
		for i := 0; i < 2; i++ {
			prompt := "artistic mosaic style, high quality, detailed"
			if i == 1 {
				prompt = "enhanced artistic style, vibrant colors, professional quality"
			}

			// Call AI service (simplified)
			// NOTE: This is a placeholder - actual implementation depends on StableDiffusion API structure
			aiResult := ""
			if false { // Disabled for now, implement when SD client is properly configured
				_ = prompt
				// aiResult, err = s.deps.AIClient.ProcessImage(ctx, req)
			}
			if aiResult != "" {
				previewID := uuid.New().String()
				previews = append(previews, PreviewInfo{
					ID:       previewID,
					URL:      aiResult,
					Style:    "ai",
					Contrast: "normal",
					Label:    fmt.Sprintf("AI вариант %d", i+1),
					IsAI:     true,
				})
			}
		}
	}

	return &GenerateAllPreviewsResponse{
		Previews: previews,
		Total:    len(previews),
		ImageID:  imageID,
	}, nil
}

func (s *PublicService) resizeImage(img image.Image, size string) image.Image {
	var width, height int

	switch size {
	case "21x30":
		width, height = 210, 300
	case "30x40":
		width, height = 300, 400
	case "40x40":
		width, height = 400, 400
	case "40x50":
		width, height = 400, 500
	case "40x60":
		width, height = 400, 600
	case "50x70":
		width, height = 500, 700
	default:
		width, height = 300, 400
	}

	// Resize maintaining aspect ratio
	return imaging.Fit(img, width, height, imaging.Lanczos)
}
