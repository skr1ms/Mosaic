package public

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
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
	"github.com/skr1ms/mosaic/pkg/marketplace"
	"github.com/skr1ms/mosaic/pkg/stableDiffusion"
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
	RedisClient       RedisClientInterface
	RecaptchaSiteKey  string
}

type PublicService struct {
	deps *PublicServiceDeps
}

func NewPublicService(deps *PublicServiceDeps) *PublicService {
	s := &PublicService{
		deps: deps,
	}

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

	partner, err := s.deps.PartnerRepository.GetByID(context.Background(), coupon.PartnerID)
	if err != nil {
		return map[string]any{
			"id":     coupon.ID,
			"code":   coupon.Code,
			"size":   coupon.Size,
			"style":  coupon.Style,
			"status": coupon.Status,
			"valid":  true,
		}, nil
	}

	return map[string]any{
		"id":             coupon.ID,
		"code":           coupon.Code,
		"size":           coupon.Size,
		"style":          coupon.Style,
		"status":         coupon.Status,
		"valid":          true,
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
	if couponID == "" {
		return map[string]any{
			"message":      "Изображение успешно загружено для предпросмотра",
			"image_id":     nil,
			"next_step":    "generate_preview",
			"coupon_size":  "30x40",
			"coupon_style": "classic",
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

// EditImage applies editing to image
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
	go func() {
		if err := s.deps.ImageService.ProcessImage(context.Background(), imageUUID, processParams); err != nil {
			fmt.Printf("Failed to process image: %v, imageUUID: %s\n", err, imageUUID.String())
		}
	}()
}

// GeneratePreview generates a single preview with style, lighting and contrast
func (s *PublicService) GeneratePreview(ctx context.Context, file *multipart.FileHeader, size, style, lighting, contrast string) (*PreviewData, error) {
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	fileContent, err := io.ReadAll(src)
	if err != nil {
		src.Close()
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	src.Close()

	fileHash := fmt.Sprintf("%x", sha256.Sum256(fileContent))
	cacheKey := fmt.Sprintf("preview:%s:%s:%s:%s:%s", fileHash[:16], size, style, lighting, contrast)

	if s.deps.RedisClient != nil {
		cachedData := s.deps.RedisClient.Get(ctx, cacheKey)
		if cachedData.Err() == nil {
			var previewData PreviewData
			if err := json.Unmarshal([]byte(cachedData.Val()), &previewData); err == nil {
				return &previewData, nil
			}
		}
	}

	previewHash := fmt.Sprintf("%s_%s_%s", size, fmt.Sprintf("%s_%s", style, lighting), contrast)
	previewID := uuid.NewSHA1(uuid.NameSpaceURL, []byte(previewHash))

	existingPreview, err := s.deps.PublicRepository.GetByID(ctx, previewID)
	if err == nil && existingPreview != nil {
		if s.deps.RedisClient != nil {
			previewJSON, _ := json.Marshal(existingPreview)
			s.deps.RedisClient.Set(ctx, cacheKey, previewJSON, 2*time.Hour)
		}
		return existingPreview, nil
	}

	src, err = file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	img, format, err := s.deps.S3Client.Decode(src)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	img = s.resizeImage(img, size)
	img = s.ApplyStyle(img, style)
	img = s.ApplyLighting(img, lighting)
	img = s.ApplyContrast(img, contrast)

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

	previewKey := fmt.Sprintf("previews/%s.jpg", previewID.String())
	if err := s.deps.S3Client.UploadToPreviewBucket(ctx, previewKey, &buf, int64(buf.Len()), "image/jpeg"); err != nil {
		return nil, fmt.Errorf("failed to upload preview: %w", err)
	}

	previewURL := s.deps.S3Client.GetPreviewURL(previewKey)
	if previewURL == "" {
		return nil, fmt.Errorf("failed to get preview URL")
	}

	previewData := &PreviewData{
		ID:       previewID,
		URL:      previewURL,
		Style:    fmt.Sprintf("%s_%s", style, lighting),
		Contrast: contrast,
		Size:     size,
		ImageID:  nil,
		S3Key:    previewKey,
	}

	if err := s.deps.PublicRepository.Create(ctx, previewData); err != nil {
		// If database save fails, cleanup S3
		s.deps.S3Client.SchedulePreviewDeletion(previewKey)
		return nil, fmt.Errorf("failed to save preview to database: %w", err)
	}

	s.deps.S3Client.SchedulePreviewDeletion(previewKey)

	if s.deps.RedisClient != nil {
		previewJSON, err := json.Marshal(previewData)
		if err == nil {
			s.deps.RedisClient.Set(ctx, cacheKey, previewJSON, 2*time.Hour)
		}
	}

	return previewData, nil
}

// GenerateStylePreview generates a simple mosaic preview for a specific style
func (s *PublicService) GenerateStylePreview(ctx context.Context, file *multipart.FileHeader, size, style string) (*PreviewData, error) {
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	fileContent, err := io.ReadAll(src)
	if err != nil {
		src.Close()
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	src.Close()

	fileHash := fmt.Sprintf("%x", sha256.Sum256(fileContent))
	previewHash := fmt.Sprintf("style_%s_%s_%s", fileHash[:16], style, size)
	previewID := uuid.NewSHA1(uuid.NameSpaceURL, []byte(previewHash))

	cacheKey := fmt.Sprintf("style_preview:%s:%s:%s", fileHash[:16], style, size)

	if s.deps.RedisClient != nil {
		cachedData := s.deps.RedisClient.Get(ctx, cacheKey)
		if cachedData.Err() == nil {
			var previewData PreviewData
			if err := json.Unmarshal([]byte(cachedData.Val()), &previewData); err == nil {
				return &previewData, nil
			}
		}
	}

	existingPreview, err := s.deps.PublicRepository.GetByID(ctx, previewID)
	if err == nil && existingPreview != nil {
		if s.deps.RedisClient != nil {
			previewJSON, _ := json.Marshal(existingPreview)
			s.deps.RedisClient.Set(ctx, cacheKey, previewJSON, 2*time.Hour)
		}
		return existingPreview, nil
	}

	src, err = file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	img, format, err := s.deps.S3Client.Decode(src)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	img = s.resizeImage(img, size)
	img = s.ApplyStyle(img, style)
	img = s.ApplyLighting(img, "sun")
	img = s.ApplyContrast(img, "normal")

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

	previewKey := fmt.Sprintf("style-previews/%s_%s.jpg", style, previewID.String())
	if err := s.deps.S3Client.UploadToPreviewBucket(ctx, previewKey, &buf, int64(buf.Len()), "image/jpeg"); err != nil {
		return nil, fmt.Errorf("failed to upload preview: %w", err)
	}

	previewURL := s.deps.S3Client.GetPreviewURL(previewKey)
	if previewURL == "" {
		return nil, fmt.Errorf("failed to get preview URL")
	}

	previewData := &PreviewData{
		ID:       previewID,
		URL:      previewURL,
		Style:    style,
		Contrast: "normal",
		Size:     size,
		ImageID:  nil,
		S3Key:    previewKey,
	}

	if err := s.deps.PublicRepository.Create(ctx, previewData); err != nil {
		s.deps.S3Client.SchedulePreviewDeletion(previewKey)
		return nil, fmt.Errorf("failed to save preview to database: %w", err)
	}

	s.deps.S3Client.SchedulePreviewDeletion(previewKey)

	if s.deps.RedisClient != nil {
		previewJSON, err := json.Marshal(previewData)
		if err == nil {
			s.deps.RedisClient.Set(ctx, cacheKey, previewJSON, 2*time.Hour)
		}
	}

	return previewData, nil
}

func (s *PublicService) GenerateAIPreview(ctx context.Context, file *multipart.FileHeader, prompt string) (*PreviewData, error) {
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	_, err = io.ReadAll(src)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

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

	if err := s.deps.PublicRepository.Create(ctx, previewData); err != nil {
		return nil, fmt.Errorf("failed to save AI preview to database: %w", err)
	}

	return previewData, nil
}

// ApplyStyle applies the main processing style to the image
func (s *PublicService) ApplyStyle(img image.Image, style string) image.Image {
	switch style {
	case "grayscale":
		return imaging.Grayscale(img)
	case "skin_tones":
		img = imaging.AdjustSaturation(img, -20)
		return s.applyColorFilter(img, color.RGBA{255, 220, 200, 20})
	case "pop_art":
		img = imaging.AdjustSaturation(img, 50)
		return imaging.AdjustContrast(img, 30)
	case "max_colors":
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
		return s.applyColorFilter(img, color.RGBA{255, 255, 100, 30})
	case "moon":
		return s.applyColorFilter(img, color.RGBA{150, 150, 255, 40})
	case "venus":
		return s.applyColorFilter(img, color.RGBA{255, 200, 150, 50})
	case "mars":
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

	overlay := image.NewUniform(filterColor)

	draw.Draw(filtered, bounds, img, bounds.Min, draw.Src)

	draw.DrawMask(filtered, bounds, overlay, bounds.Min, nil, bounds.Min, draw.Over)

	return filtered
}

// GenerateAllPreviews generates all 8 base previews + optional 1 AI preview
func (s *PublicService) GenerateAllPreviews(ctx context.Context, imageID string, size string, useAI bool) (*GenerateAllPreviewsResponse, error) {
	imageUUID, err := uuid.Parse(imageID)
	if err != nil {
		return nil, fmt.Errorf("invalid image ID: %w", err)
	}

	img, err := s.deps.ImageRepository.GetByID(ctx, imageUUID)
	if err != nil {
		return nil, fmt.Errorf("image not found: %w", err)
	}

	imageData, err := s.deps.S3Client.DownloadFile(ctx, img.OriginalImageS3Key)
	if err != nil {
		return nil, fmt.Errorf("failed to download image: %w", err)
	}
	defer imageData.Close()

	originalImg, format, err := s.deps.S3Client.Decode(imageData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	var previews []PreviewInfo

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

	type previewTask struct {
		style    string
		contrast struct {
			Value string
			Label string
		}
	}

	var tasks []previewTask
	for _, style := range styles {
		for _, contrast := range contrasts {
			tasks = append(tasks, previewTask{
				style:    style,
				contrast: contrast,
			})
		}
	}

	resultChan := make(chan PreviewInfo, len(tasks))
	errorChan := make(chan error, len(tasks))

	for _, task := range tasks {
		go func() {
			processedImg := s.resizeImage(originalImg, size)
			processedImg = s.ApplyLighting(processedImg, task.style)
			processedImg = s.ApplyContrast(processedImg, task.contrast.Value)

			var buf bytes.Buffer
			switch format {
			case "jpeg", "jpg":
				jpeg.Encode(&buf, processedImg, &jpeg.Options{Quality: 90})
			case "png":
				png.Encode(&buf, processedImg)
			default:
				jpeg.Encode(&buf, processedImg, &jpeg.Options{Quality: 90})
			}

			previewHash := fmt.Sprintf("all_%s_%s_%s_%s", imageID, task.style, task.contrast.Value, size)
			previewID := uuid.NewSHA1(uuid.NameSpaceURL, []byte(previewHash))
			previewKey := fmt.Sprintf("previews/%s/%s_%s_%s.jpg", imageID, task.style, task.contrast.Value, previewID.String()[:8])

			if err := s.deps.S3Client.UploadToPreviewBucket(ctx, previewKey, &buf, int64(buf.Len()), "image/jpeg"); err != nil {
				errorChan <- err
				return
			}

			s.deps.S3Client.SchedulePreviewDeletion(previewKey)

			previewURL := s.deps.S3Client.GetPreviewURL(previewKey)

			resultChan <- PreviewInfo{
				ID:       previewID.String(),
				URL:      previewURL,
				Style:    task.style,
				Contrast: task.contrast.Value,
				Label:    fmt.Sprintf("%s (%s)", styleLabels[task.style], task.contrast.Label),
				IsAI:     false,
			}
		}()
	}

	totalExpected := len(tasks)
	if useAI && s.deps.AIClient != nil {
		totalExpected++
	}

	for i := 0; i < totalExpected; i++ {
		select {
		case preview := <-resultChan:
			previews = append(previews, preview)
		case err := <-errorChan:
			_ = err
		case <-ctx.Done():
			return nil, fmt.Errorf("preview generation timeout")
		}
	}

	if useAI && s.deps.AIClient != nil {
		var buf bytes.Buffer
		switch format {
		case "jpeg", "jpg":
			jpeg.Encode(&buf, originalImg, &jpeg.Options{Quality: 90})
		case "png":
			png.Encode(&buf, originalImg)
		default:
			jpeg.Encode(&buf, originalImg, &jpeg.Options{Quality: 90})
		}

		aiClient, ok := s.deps.AIClient.(*stableDiffusion.StableDiffusionClient)
		if !ok {
			return &GenerateAllPreviewsResponse{
				Previews: previews,
				Total:    len(previews),
				ImageID:  imageID,
			}, fmt.Errorf("AI client is not StableDiffusionClient")
		}
		base64Image := aiClient.EncodeImageToBase64(buf.Bytes())

		width, height := s.parseSize(size)

		aiRequest := stableDiffusion.ProcessImageRequest{
			ImageBase64: base64Image,
			Style:       stableDiffusion.StyleMaxColors,
			UseAI:       true,
			Lighting:    stableDiffusion.LightingSun,
			Contrast:    stableDiffusion.ContrastHigh,
			Brightness:  0.0,
			Saturation:  0.0,
			Width:       width,
			Height:      height,
		}

		go func() {
			aiResultBase64, err := s.deps.AIClient.ProcessImage(ctx, aiRequest)
			if err != nil {
				errorChan <- err
				return
			}

			aiClient, ok := s.deps.AIClient.(*stableDiffusion.StableDiffusionClient)
			if !ok {
				errorChan <- fmt.Errorf("AI client is not StableDiffusionClient")
				return
			}
			aiResultData, err := aiClient.DecodeBase64Image(aiResultBase64)
			if err != nil {
				errorChan <- err
				return
			}

			previewHash := fmt.Sprintf("ai_%s_%s", imageID, size)
			previewID := uuid.NewSHA1(uuid.NameSpaceURL, []byte(previewHash))
			previewKey := fmt.Sprintf("previews/%s/ai_%s.jpg", imageID, previewID.String()[:8])

			aiDataReader := bytes.NewReader(aiResultData)
			if err := s.deps.S3Client.UploadToPreviewBucket(ctx, previewKey, aiDataReader, int64(len(aiResultData)), "image/jpeg"); err != nil {
				errorChan <- err
				return
			}

			s.deps.S3Client.SchedulePreviewDeletion(previewKey)

			previewURL := s.deps.S3Client.GetPreviewURL(previewKey)

			resultChan <- PreviewInfo{
				ID:       previewID.String(),
				URL:      previewURL,
				Style:    "ai",
				Contrast: "enhanced",
				Label:    "AI Enhanced Mosaic",
				IsAI:     true,
			}
		}()
	}

	return &GenerateAllPreviewsResponse{
		Previews: previews,
		Total:    len(previews),
		ImageID:  imageID,
	}, nil
}

// GenerateAllPreviewsFromFile generates all 8 base previews + optional 1 AI preview directly from uploaded file
func (s *PublicService) GenerateAllPreviewsFromFile(ctx context.Context, file *multipart.FileHeader, size string, useAI bool) (*GenerateAllPreviewsResponse, error) {
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	originalImg, format, err := s.deps.S3Client.Decode(src)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	sessionHash := fmt.Sprintf("session_%s_%s", file.Filename, size)
	sessionID := uuid.NewSHA1(uuid.NameSpaceURL, []byte(sessionHash)).String()

	var previews []PreviewInfo

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

	type previewTask struct {
		style    string
		contrast struct {
			Value string
			Label string
		}
	}

	var tasks []previewTask
	for _, style := range styles {
		for _, contrast := range contrasts {
			tasks = append(tasks, previewTask{
				style:    style,
				contrast: contrast,
			})
		}
	}

	resultChan := make(chan PreviewInfo, len(tasks))
	errorChan := make(chan error, len(tasks))

	for _, task := range tasks {
		task := task
		go func() {
			processedImg := s.resizeImage(originalImg, size)
			processedImg = s.ApplyLighting(processedImg, task.style)
			processedImg = s.ApplyContrast(processedImg, task.contrast.Value)

			var buf bytes.Buffer
			switch format {
			case "jpeg", "jpg":
				jpeg.Encode(&buf, processedImg, &jpeg.Options{Quality: 90})
			case "png":
				png.Encode(&buf, processedImg)
			default:
				jpeg.Encode(&buf, processedImg, &jpeg.Options{Quality: 90})
			}

			previewHash := fmt.Sprintf("file_%s_%s_%s_%s", sessionID[:8], task.style, task.contrast.Value, size)
			previewID := uuid.NewSHA1(uuid.NameSpaceURL, []byte(previewHash))
			previewKey := fmt.Sprintf("previews/%s/%s_%s_%s.jpg", sessionID[:8], task.style, task.contrast.Value, previewID.String()[:8])

			if err := s.deps.S3Client.UploadToPreviewBucket(ctx, previewKey, &buf, int64(buf.Len()), "image/jpeg"); err != nil {
				errorChan <- err
				return
			}

			s.deps.S3Client.SchedulePreviewDeletion(previewKey)

			previewURL := s.deps.S3Client.GetPreviewURL(previewKey)

			resultChan <- PreviewInfo{
				ID:       previewID.String(),
				URL:      previewURL,
				Style:    task.style,
				Contrast: task.contrast.Value,
				Label:    fmt.Sprintf("%s (%s)", styleLabels[task.style], task.contrast.Label),
				IsAI:     false,
			}
		}()
	}

	totalExpected := len(tasks)
	if useAI && s.deps.AIClient != nil {
		totalExpected++
	}

	for i := 0; i < totalExpected; i++ {
		select {
		case preview := <-resultChan:
			previews = append(previews, preview)
		case err := <-errorChan:
			_ = err
		case <-ctx.Done():
			return nil, fmt.Errorf("preview generation timeout")
		}
	}

	if useAI && s.deps.AIClient != nil {
		aiClient, ok := s.deps.AIClient.(*stableDiffusion.StableDiffusionClient)
		if !ok {
		} else {
			var buf bytes.Buffer
			switch format {
			case "jpeg", "jpg":
				jpeg.Encode(&buf, originalImg, &jpeg.Options{Quality: 90})
			case "png":
				png.Encode(&buf, originalImg)
			default:
				jpeg.Encode(&buf, originalImg, &jpeg.Options{Quality: 90})
			}

			base64Image := aiClient.EncodeImageToBase64(buf.Bytes())

			width, height := s.parseSize(size)

			aiRequest := stableDiffusion.ProcessImageRequest{
				ImageBase64: base64Image,
				Style:       stableDiffusion.StyleMaxColors,
				UseAI:       true,
				Lighting:    stableDiffusion.LightingSun,
				Contrast:    stableDiffusion.ContrastHigh,
				Brightness:  0.0,
				Saturation:  0.0,
				Width:       width,
				Height:      height,
			}

			go func() {
				aiResultBase64, err := s.deps.AIClient.ProcessImage(ctx, aiRequest)
				if err != nil {
					errorChan <- err
					return
				}

				aiResultData, err := aiClient.DecodeBase64Image(aiResultBase64)
				if err != nil {
					errorChan <- err
					return
				}

				previewHash := fmt.Sprintf("ai_file_%s_%s", sessionID[:8], size)
				previewID := uuid.NewSHA1(uuid.NameSpaceURL, []byte(previewHash))
				previewKey := fmt.Sprintf("previews/%s/ai_%s.jpg", sessionID[:8], previewID.String()[:8])

				aiDataReader := bytes.NewReader(aiResultData)
				if err := s.deps.S3Client.UploadToPreviewBucket(ctx, previewKey, aiDataReader, int64(len(aiResultData)), "image/jpeg"); err != nil {
					errorChan <- err
					return
				}

				s.deps.S3Client.SchedulePreviewDeletion(previewKey)

				previewURL := s.deps.S3Client.GetPreviewURL(previewKey)

				resultChan <- PreviewInfo{
					ID:       previewID.String(),
					URL:      previewURL,
					Style:    "ai",
					Contrast: "enhanced",
					Label:    "AI Enhanced Mosaic",
					IsAI:     true,
				}
			}()
		}
	}

	return &GenerateAllPreviewsResponse{
		Previews: previews,
		Total:    len(previews),
		ImageID:  sessionID,
	}, nil
}

func (s *PublicService) parseSize(size string) (int, int) {
	switch size {
	case "21x30":
		return 210, 300
	case "30x40":
		return 300, 400
	case "40x40":
		return 400, 400
	case "40x50":
		return 400, 500
	case "40x60":
		return 400, 600
	case "50x70":
		return 500, 700
	default:
		return 300, 400
	}
}

func (s *PublicService) resizeImage(img image.Image, size string) image.Image {
	width, height := s.parseSize(size)
	return imaging.Fit(img, width, height, imaging.Lanczos)
}
