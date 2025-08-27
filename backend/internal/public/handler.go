package public

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/skr1ms/mosaic/internal/admin"
	"github.com/skr1ms/mosaic/internal/image"
	"github.com/skr1ms/mosaic/internal/types"
	"github.com/skr1ms/mosaic/pkg/goroutine"
	"github.com/skr1ms/mosaic/pkg/middleware"
)

type PublicHandlerDeps struct {
	PublicService PublicServiceInterface
	S3Client      admin.S3ClientInterface
	Logger        *middleware.Logger
}

// PublicHandler обработчик для публичного API
type PublicHandler struct {
	fiber.Router
	deps             *PublicHandlerDeps
	goroutineManager *goroutine.Manager
	schemaPool       *goroutine.WorkerPool
}

func NewPublicHandler(router fiber.Router, deps *PublicHandlerDeps) *PublicHandler {
	handler := &PublicHandler{
		Router: router,
		deps:   deps,
	}

	handler.goroutineManager = goroutine.NewManager(context.Background())
	handler.schemaPool = handler.goroutineManager.NewWorkerPool("schema_generation", 2, 30)

	// ================================================================
	// PUBLIC API ROUTES: /api/*
	// Access: public (no authentication required)
	// ================================================================
	router.Get("/branding", handler.GetBrandingInfo)                   // GET /api/branding
	router.Get("/partners/:domain/info", handler.GetPartnerByDomain)   // GET /api/partners/:domain/info
	router.Get("/coupons/:code", handler.GetCouponByCode)              // GET /api/coupons/:code
	router.Post("/coupons/:code/activate", handler.ActivateCoupon)     // POST /api/coupons/:code/activate
	router.Post("/coupons/purchase", handler.PurchaseCoupon)           // POST /api/coupons/purchase
	router.Post("/images/upload", handler.UploadImage)                 // POST /api/images/upload
	router.Post("/images/:id/edit", handler.EditImage)                 // POST /api/images/:id/edit
	router.Post("/images/:id/process", handler.ProcessImage)           // POST /api/images/:id/process
	router.Post("/images/:id/generate-schema", handler.GenerateSchema) // POST /api/images/:id/generate-schema
	router.Post("/images/:id/send-email", handler.SendSchemaToEmail)   // POST /api/images/:id/send-email
	router.Get("/images/:id/preview", handler.GetImagePreview)         // GET /api/images/:id/preview
	router.Get("/images/:id/status", handler.GetProcessingStatus)      // GET /api/images/:id/status
	router.Get("/images/:id/download", handler.DownloadSchema)         // GET /api/images/:id/download
	router.Get("/sizes", handler.GetAvailableSizes)                    // GET /api/sizes
	router.Get("/styles", handler.GetAvailableStyles)                  // GET /api/styles
	router.Get("/config/recaptcha", handler.GetRecaptchaSiteKey)       // GET /api/config/recaptcha

	return handler
}

// @Summary Get branding information
// @Description Returns branding data (logo, contacts, links) for the current domain
// @Tags public
// @Produce json
// @Success 200 {object} map[string]interface{} "Branding data including logo, contacts, and links"
// @Failure 500 {object} map[string]interface{} "Internal server error when retrieving branding data"
// @Router /api/branding [get]
func (h *PublicHandler) GetBrandingInfo(c *fiber.Ctx) error {
	brandingResponse := middleware.BrandingResponse(c)

	if !brandingResponse["is_default"].(bool) {
		if partner, ok := brandingResponse["partner_id"]; ok && partner != nil {
			brandingData := middleware.GetBrandingFromContext(c)
			if brandingData != nil && brandingData.Partner != nil {
				if brandingData.Partner.LogoURL != "" {
					logoURL := brandingData.Partner.LogoURL
					var marker string
					if strings.Contains(logoURL, "/logos/") {
						marker = "/logos/"
					} else if strings.Contains(logoURL, "/mosaic-logos/") {
						marker = "/mosaic-logos/"
					}

					if marker != "" {
						parts := strings.Split(logoURL, marker)
						if len(parts) > 1 {
							objectKey := parts[1]
							if newURL, err := h.deps.S3Client.GetLogoURL(c.UserContext(), objectKey, 24*time.Hour); err == nil {
								brandingResponse["logo_url"] = newURL
							}
						}
					}
				}
			}
		}
	}

	h.deps.Logger.FromContext(c).Info().
		Str("handler", "GetBrandingInfo").
		Bool("is_default", brandingResponse["is_default"].(bool)).
		Msg("Branding information retrieved successfully")

	return c.JSON(brandingResponse)
}

// @Summary Get partner information by domain
// @Description Returns branding and contact information of a partner for White Label implementation
// @Tags public
// @Produce json
// @Param domain path string true "Partner's domain name"
// @Success 200 {object} map[string]interface{} "Partner information including branding and contact details"
// @Failure 404 {object} map[string]interface{} "Partner not found for the specified domain"
// @Failure 500 {object} map[string]interface{} "Internal server error when retrieving partner information"
// @Router /api/partners/{domain}/info [get]
func (h *PublicHandler) GetPartnerByDomain(c *fiber.Ctx) error {
	domain := c.Params("domain")

	result, err := h.deps.PublicService.GetPartnerByDomain(domain)
	if err != nil {
		h.deps.Logger.FromContext(c).Error().
			Err(err).
			Str("handler", "GetPartnerByDomain").
			Str("domain", domain).
			Msg("Failed to get partner by domain")

		errorResponse := fiber.Map{
			"error":      "Failed to get partner by domain",
			"request_id": c.Get("X-Request-ID"),
		}
		if os.Getenv("ENVIRONMENT") == "development" || os.Getenv("ENVIRONMENT") == "dev" {
			errorResponse["details"] = err.Error()
		}
		return c.Status(fiber.StatusNotFound).JSON(errorResponse)
	}

	h.deps.Logger.FromContext(c).Info().
		Str("handler", "GetPartnerByDomain").
		Str("domain", domain).
		Msg("Partner information retrieved successfully")

	return c.JSON(result)
}

// @Summary Get coupon information by code
// @Description Returns coupon information for validation purposes
// @Tags coupons
// @Produce json
// @Param code path string true "Coupon code (12 digits)"
// @Success 200 {object} map[string]interface{} "Coupon information including status, partner, and creation date"
// @Failure 400 {object} map[string]interface{} "Invalid coupon code format"
// @Failure 404 {object} map[string]interface{} "Coupon not found"
// @Failure 500 {object} map[string]interface{} "Internal server error when retrieving coupon"
// @Router /api/coupons/{code} [get]
func (h *PublicHandler) GetCouponByCode(c *fiber.Ctx) error {
	code := c.Params("code")

	result, err := h.deps.PublicService.GetCouponByCode(code)
	if err != nil {
		errStr := err.Error()
		h.deps.Logger.FromContext(c).Error().
			Err(err).
			Str("handler", "GetCouponByCode").
			Str("coupon_code", code).
			Msg("Failed to get coupon by code")

		var errorMsg string
		var statusCode int

		if strings.Contains(errStr, "invalid coupon code") {
			errorMsg = "Invalid coupon code"
			statusCode = fiber.StatusBadRequest
		} else if strings.Contains(errStr, "coupon not found") || strings.Contains(errStr, "not found") {
			errorMsg = "Coupon not found"
			statusCode = fiber.StatusNotFound
		} else {
			errorMsg = "Internal server error when retrieving coupon"
			statusCode = fiber.StatusInternalServerError
		}

		errorResponse := fiber.Map{
			"error":      errorMsg,
			"request_id": c.Get("X-Request-ID"),
		}
		if os.Getenv("ENVIRONMENT") == "development" || os.Getenv("ENVIRONMENT") == "dev" {
			errorResponse["details"] = err.Error()
		}
		return c.Status(statusCode).JSON(errorResponse)
	}

	h.deps.Logger.FromContext(c).Info().
		Str("handler", "GetCouponByCode").
		Str("coupon_code", code).
		Str("status", result["status"].(string)).
		Msg("Coupon information retrieved successfully")

	return c.JSON(result)
}

// @Summary Activate coupon
// @Description Activates a coupon and prepares it for image upload
// @Tags coupons
// @Accept json
// @Produce json
// @Param code path string true "Coupon code"
// @Success 200 {object} map[string]interface{} "Coupon activated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request format"
// @Failure 404 {object} map[string]interface{} "Coupon not found"
// @Failure 409 {object} map[string]interface{} "Coupon already used"
// @Failure 500 {object} map[string]interface{} "Internal server error during coupon activation"
// @Router /api/coupons/{code}/activate [post]
func (h *PublicHandler) ActivateCoupon(c *fiber.Ctx) error {
	code := c.Params("code")

	result, err := h.deps.PublicService.ActivateCoupon(code)
	if err != nil {
		h.deps.Logger.FromContext(c).Error().
			Err(err).
			Str("handler", "ActivateCoupon").
			Str("coupon_code", code).
			Msg("Failed to activate coupon")

		var errorMsg string
		var statusCode int

		if err.Error() == "coupon not found" {
			errorMsg = "Coupon not found"
			statusCode = fiber.StatusNotFound
		} else if err.Error() == "coupon already used" {
			errorMsg = "Coupon already used"
			statusCode = fiber.StatusConflict
		} else {
			errorMsg = "Failed to activate coupon"
			statusCode = fiber.StatusInternalServerError
		}

		errorResponse := fiber.Map{
			"error":      errorMsg,
			"request_id": c.Get("X-Request-ID"),
		}
		if os.Getenv("ENVIRONMENT") == "development" || os.Getenv("ENVIRONMENT") == "dev" {
			errorResponse["details"] = err.Error()
		}
		return c.Status(statusCode).JSON(errorResponse)
	}

	h.deps.Logger.FromContext(c).Info().
		Str("handler", "ActivateCoupon").
		Str("coupon_code", code).
		Msg("Coupon activated successfully")

	return c.JSON(result)
}

// @Summary Upload image for processing
// @Description Uploads user's image for mosaic pattern creation
// @Tags images
// @Accept multipart/form-data
// @Produce json
// @Param coupon_id formData string false "ID of activated coupon (UUID)"
// @Param coupon_code formData string false "Coupon code (12 digits) - alternative to coupon_id"
// @Param image formData file true "Image file (JPG, PNG)"
// @Success 201 {object} map[string]interface{} "Image uploaded successfully"
// @Failure 400 {object} map[string]interface{} "Bad request: missing required fields or invalid data"
// @Failure 413 {object} map[string]interface{} "File too large"
// @Failure 500 {object} map[string]interface{} "Internal server error during image upload"
// @Router /api/images/upload [post]
func (h *PublicHandler) UploadImage(c *fiber.Ctx) error {
	couponID := c.FormValue("coupon_id")
	couponCode := c.FormValue("coupon_code")

	if couponID == "" && couponCode == "" {
		h.deps.Logger.FromContext(c).Warn().
			Str("handler", "UploadImage").
			Msg("Coupon ID or Code is required")

		errorResponse := fiber.Map{
			"error":      "Coupon ID or Code is required",
			"request_id": c.Get("X-Request-ID"),
		}
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse)
	}

	file, err := c.FormFile("image")
	if err != nil {
		h.deps.Logger.FromContext(c).Warn().
			Err(err).
			Str("handler", "UploadImage").
			Msg("Image file is required")

		errorResponse := fiber.Map{
			"error":      "Image file is required",
			"request_id": c.Get("X-Request-ID"),
		}
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse)
	}

	if couponID == "" && couponCode != "" {
		cleanCode := strings.TrimSpace(strings.ReplaceAll(couponCode, "-", ""))
		if len(cleanCode) != 12 {
			h.deps.Logger.FromContext(c).Warn().
				Str("handler", "UploadImage").
				Str("coupon_code", cleanCode).
				Int("length", len(cleanCode)).
				Msg("Coupon code must be 12 digits")

			errorResponse := fiber.Map{
				"error":      "Coupon code must be 12 digits",
				"request_id": c.Get("X-Request-ID"),
			}
			return c.Status(fiber.StatusBadRequest).JSON(errorResponse)
		}

		coupon, err := h.deps.PublicService.GetCouponRepository().GetByCode(context.Background(), cleanCode)
		if err != nil {
			h.deps.Logger.FromContext(c).Error().
				Err(err).
				Str("handler", "UploadImage").
				Str("coupon_code", cleanCode).
				Msg("Coupon not found")

			errorResponse := fiber.Map{
				"error":      "Coupon not found",
				"request_id": c.Get("X-Request-ID"),
			}
			return c.Status(fiber.StatusNotFound).JSON(errorResponse)
		}
		couponID = coupon.ID.String()
	}

	result, err := h.deps.PublicService.UploadImage(couponID, file)
	if err != nil {
		h.deps.Logger.FromContext(c).Error().
			Err(err).
			Str("handler", "UploadImage").
			Str("coupon_id", couponID).
			Msg("Failed to upload image")

		errorResponse := fiber.Map{
			"error":      "Failed to upload image",
			"request_id": c.Get("X-Request-ID"),
		}
		if os.Getenv("ENVIRONMENT") == "development" || os.Getenv("ENVIRONMENT") == "dev" {
			errorResponse["details"] = err.Error()
		}
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse)
	}

	h.deps.Logger.FromContext(c).Info().
		Str("handler", "UploadImage").
		Str("coupon_id", couponID).
		Str("image_id", result["image_id"].(string)).
		Str("filename", file.Filename).
		Int("size", int(file.Size)).
		Msg("Image uploaded successfully")

	return c.Status(fiber.StatusCreated).JSON(result)
}

// @Summary Edit image
// @Description Applies cropping, rotation and scaling to the image
// @Tags images
// @Accept json
// @Produce json
// @Param id path string true "Image ID"
// @Param request body types.EditImageRequest true "Editing parameters"
// @Success 200 {object} map[string]interface{} "Image edited successfully"
// @Failure 400 {object} map[string]interface{} "Bad request: invalid editing parameters"
// @Failure 404 {object} map[string]interface{} "Image not found"
// @Failure 500 {object} map[string]interface{} "Internal server error during image editing"
// @Router /api/images/{id}/edit [post]
func (h *PublicHandler) EditImage(c *fiber.Ctx) error {
	imageID := c.Params("id")

	var req types.EditImageRequest
	if err := c.BodyParser(&req); err != nil {
		h.deps.Logger.FromContext(c).Warn().
			Err(err).
			Str("handler", "EditImage").
			Str("image_id", imageID).
			Msg("Failed to parse request body")

		errorResponse := fiber.Map{
			"error":      "Failed to parse request body",
			"request_id": c.Get("X-Request-ID"),
		}
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse)
	}

	result, err := h.deps.PublicService.EditImage(imageID, req)
	if err != nil {
		h.deps.Logger.FromContext(c).Error().
			Err(err).
			Str("handler", "EditImage").
			Str("image_id", imageID).
			Msg("Failed to edit image")

		var errorMsg string
		var statusCode int

		if err.Error() == "image not found" {
			errorMsg = "Image not found"
			statusCode = fiber.StatusNotFound
		} else {
			errorMsg = "Failed to edit image"
			statusCode = fiber.StatusInternalServerError
		}

		errorResponse := fiber.Map{
			"error":      errorMsg,
			"request_id": c.Get("X-Request-ID"),
		}
		if os.Getenv("ENVIRONMENT") == "development" || os.Getenv("ENVIRONMENT") == "dev" {
			errorResponse["details"] = err.Error()
		}
		return c.Status(statusCode).JSON(errorResponse)
	}

	h.deps.Logger.FromContext(c).Info().
		Str("handler", "EditImage").
		Str("image_id", imageID).
		Interface("params", req).
		Msg("Image edited successfully")

	return c.JSON(result)
}

// @Summary Process image
// @Description Applies selected processing style to the image
// @Tags images
// @Accept json
// @Produce json
// @Param id path string true "Image ID"
// @Param request body types.ProcessImageRequest true "Processing parameters"
// @Success 200 {object} map[string]interface{} "Image processing started"
// @Failure 400 {object} map[string]interface{} "Bad request: invalid processing parameters"
// @Failure 404 {object} map[string]interface{} "Image not found"
// @Failure 500 {object} map[string]interface{} "Internal server error during image processing"
// @Router /api/images/{id}/process [post]
func (h *PublicHandler) ProcessImage(c *fiber.Ctx) error {
	imageID := c.Params("id")

	var req types.ProcessImageRequest
	if err := c.BodyParser(&req); err != nil {
		h.deps.Logger.FromContext(c).Warn().
			Err(err).
			Str("handler", "ProcessImage").
			Str("image_id", imageID).
			Msg("Failed to parse request body")

		errorResponse := fiber.Map{
			"error":      "Failed to parse request body",
			"request_id": c.Get("X-Request-ID"),
		}
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse)
	}

	result, err := h.deps.PublicService.ProcessImage(imageID, req)
	if err != nil {
		h.deps.Logger.FromContext(c).Error().
			Err(err).
			Str("handler", "ProcessImage").
			Str("image_id", imageID).
			Msg("Failed to process image")

		var errorMsg string
		var statusCode int

		if err.Error() == "image not found" {
			errorMsg = "Image not found"
			statusCode = fiber.StatusNotFound
		} else {
			errorMsg = "Failed to process image"
			statusCode = fiber.StatusInternalServerError
		}

		errorResponse := fiber.Map{
			"error":      errorMsg,
			"request_id": c.Get("X-REQUEST-ID"),
		}
		if os.Getenv("ENVIRONMENT") == "development" || os.Getenv("ENVIRONMENT") == "dev" {
			errorResponse["details"] = err.Error()
		}
		return c.Status(statusCode).JSON(errorResponse)
	}

	h.deps.Logger.FromContext(c).Info().
		Str("handler", "ProcessImage").
		Str("image_id", imageID).
		Str("style", req.Style).
		Msg("Image processed successfully")

	return c.JSON(result)
}

// @Summary Generate mosaic schema
// @Description Creates final mosaic schema from processed image
// @Tags images
// @Accept json
// @Produce json
// @Param id path string true "Image ID"
// @Param request body types.GenerateSchemaRequest true "Schema generation parameters"
// @Success 200 {object} map[string]interface{} "Schema generation started"
// @Failure 400 {object} map[string]interface{} "Bad request: invalid schema generation parameters"
// @Failure 404 {object} map[string]interface{} "Image not found"
// @Failure 500 {object} map[string]interface{} "Internal server error during schema generation"
// @Router /api/images/{id}/generate-schema [post]
func (h *PublicHandler) GenerateSchema(c *fiber.Ctx) error {
	imageID := c.Params("id")

	imageUUID, err := uuid.Parse(imageID)
	if err != nil {
		h.deps.Logger.FromContext(c).Warn().
			Err(err).
			Str("handler", "GenerateSchema").
			Str("image_id", imageID).
			Msg("Invalid image ID format")

		errorResponse := fiber.Map{
			"error":      "Invalid image ID format",
			"request_id": c.Get("X-REQUEST-ID"),
		}
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse)
	}

	var req types.GenerateSchemaRequest
	if err := c.BodyParser(&req); err != nil {
		h.deps.Logger.FromContext(c).Warn().
			Err(err).
			Str("handler", "GenerateSchema").
			Str("image_id", imageID).
			Msg("Failed to parse request body")

		errorResponse := fiber.Map{
			"error":      "Failed to parse request body",
			"request_id": c.Get("X-REQUEST-ID"),
		}
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse)
	}

	task, err := h.deps.PublicService.GetImageRepository().GetByID(context.Background(), imageUUID)
	if err != nil {
		h.deps.Logger.FromContext(c).Error().
			Err(err).
			Str("handler", "GenerateSchema").
			Str("image_id", imageID).
			Msg("Image not found")

		errorResponse := fiber.Map{
			"error":      "Image not found",
			"request_id": c.Get("X-REQUEST-ID"),
		}
		return c.Status(fiber.StatusNotFound).JSON(errorResponse)
	}

	var previewURL *string
	if status, err := h.deps.PublicService.GetImageService().GetImageStatus(c.UserContext(), imageUUID); err == nil {
		previewURL = status.PreviewURL
	}

	h.generateSchemaAsync(imageUUID, req.Confirmed, task)

	h.deps.Logger.FromContext(c).Info().
		Str("handler", "GenerateSchema").
		Str("image_id", imageID).
		Str("coupon_id", task.CouponID.String()).
		Msg("Schema generation started")

	return c.JSON(fiber.Map{
		"message":     "Schema generation started",
		"actions":     []string{"download"},
		"email_sent":  true,
		"schema_uuid": imageUUID.String(),
		"preview_url": previewURL,
	})
}

// @Summary Download schema
// @Description Downloads the generated mosaic schema
// @Tags images
// @Produce application/octet-stream
// @Param id path string true "Image ID"
// @Success 200 {file} file "Schema file"
// @Failure 400 {object} map[string]interface{} "Invalid image ID format"
// @Failure 404 {object} map[string]interface{} "Image not found"
// @Failure 409 {object} map[string]interface{} "Schema not ready"
// @Failure 500 {object} map[string]interface{} "Internal server error during schema download"
// @Router /api/images/{id}/download [get]
func (h *PublicHandler) DownloadSchema(c *fiber.Ctx) error {
	imageID := c.Params("id")

	imageUUID, err := uuid.Parse(imageID)
	if err != nil {
		h.deps.Logger.FromContext(c).Warn().
			Err(err).
			Str("handler", "DownloadSchema").
			Str("image_id", imageID).
			Msg("Invalid image ID")

		errorResponse := fiber.Map{
			"error":      "Invalid image ID",
			"request_id": c.Get("X-REQUEST-ID"),
		}
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse)
	}

	status, err := h.deps.PublicService.GetImageService().GetImageStatus(c.UserContext(), imageUUID)
	if err != nil {
		h.deps.Logger.FromContext(c).Error().
			Err(err).
			Str("handler", "DownloadSchema").
			Str("image_id", imageID).
			Msg("Image not found")

		errorResponse := fiber.Map{
			"error":      "Image not found",
			"request_id": c.Get("X-REQUEST-ID"),
		}
		return c.Status(fiber.StatusNotFound).JSON(errorResponse)
	}

	if status.Status != "completed" || status.ZipURL == nil {
		h.deps.Logger.FromContext(c).Warn().
			Str("handler", "DownloadSchema").
			Str("image_id", imageID).
			Str("status", status.Status).
			Msg("Schema not ready")

		errorResponse := fiber.Map{
			"error":      "Schema not ready",
			"request_id": c.Get("X-REQUEST-ID"),
		}
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse)
	}

	h.deps.Logger.FromContext(c).Info().
		Str("handler", "DownloadSchema").
		Str("image_id", imageID).
		Msg("Schema downloaded successfully")

	return c.Redirect(*status.ZipURL)
}

// @Summary Send schema to email
// @Description Sends the generated mosaic schema to the specified email
// @Tags images
// @Accept json
// @Produce json
// @Param id path string true "Image ID"
// @Param request body public.SendEmailRequest true "Email address for sending"
// @Success 200 {object} map[string]interface{} "Schema sent successfully"
// @Failure 400 {object} map[string]interface{} "Bad request: invalid email or request format"
// @Failure 404 {object} map[string]interface{} "Image not found"
// @Failure 500 {object} map[string]interface{} "Internal server error during email sending"
// @Router /api/images/{id}/send-email [post]
func (h *PublicHandler) SendSchemaToEmail(c *fiber.Ctx) error {
	imageID := c.Params("id")

	var req SendEmailRequest
	if err := c.BodyParser(&req); err != nil {
		h.deps.Logger.FromContext(c).Warn().
			Err(err).
			Str("handler", "SendSchemaToEmail").
			Str("image_id", imageID).
			Msg("Failed to parse request body")

		errorResponse := fiber.Map{
			"error":      "Failed to parse request body",
			"request_id": c.Get("X-REQUEST-ID"),
		}
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse)
	}

	result, err := h.deps.PublicService.SendSchemaToEmail(imageID, req)
	if err != nil {
		h.deps.Logger.FromContext(c).Error().
			Err(err).
			Str("handler", "SendSchemaToEmail").
			Str("image_id", imageID).
			Str("email", req.Email).
			Msg("Failed to send schema to email")

		errorResponse := fiber.Map{
			"error":      "Failed to send schema to email",
			"request_id": c.Get("X-REQUEST-ID"),
		}
		if os.Getenv("ENVIRONMENT") == "development" || os.Getenv("ENVIRONMENT") == "dev" {
			errorResponse["details"] = err.Error()
		}
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse)
	}

	h.deps.Logger.FromContext(c).Info().
		Str("handler", "SendSchemaToEmail").
		Str("image_id", imageID).
		Str("email", req.Email).
		Msg("Schema sent to email successfully")

	return c.JSON(result)
}

// @Summary Get image preview
// @Description Returns the preview of the processed image
// @Tags images
// @Produce json
// @Param id path string true "Image ID"
// @Success 200 {object} map[string]interface{} "Image preview with URL and dimensions"
// @Failure 404 {object} map[string]interface{} "Image not found"
// @Failure 500 {object} map[string]interface{} "Internal server error when retrieving image preview"
// @Router /api/images/{id}/preview [get]
func (h *PublicHandler) GetImagePreview(c *fiber.Ctx) error {
	imageID := c.Params("id")

	result, err := h.deps.PublicService.GetImagePreview(imageID)
	if err != nil {
		h.deps.Logger.FromContext(c).Error().
			Err(err).
			Str("handler", "GetImagePreview").
			Str("image_id", imageID).
			Msg("Failed to get image preview")

		var errorMsg string
		var statusCode int

		if err.Error() == "image not found" {
			errorMsg = "Image not found"
			statusCode = fiber.StatusNotFound
		} else {
			errorMsg = "Failed to get image preview"
			statusCode = fiber.StatusInternalServerError
		}

		errorResponse := fiber.Map{
			"error":      errorMsg,
			"request_id": c.Get("X-REQUEST-ID"),
		}
		if os.Getenv("ENVIRONMENT") == "development" || os.Getenv("ENVIRONMENT") == "dev" {
			errorResponse["details"] = err.Error()
		}
		return c.Status(statusCode).JSON(errorResponse)
	}

	h.deps.Logger.FromContext(c).Info().
		Str("handler", "GetImagePreview").
		Str("image_id", imageID).
		Msg("Image preview retrieved successfully")

	return c.JSON(result)
}

// @Summary Get processing status
// @Description Returns the current processing status of the image
// @Tags images
// @Produce json
// @Param id path string true "Image ID"
// @Success 200 {object} map[string]interface{} "Processing status including current step and progress"
// @Failure 404 {object} map[string]interface{} "Image not found"
// @Failure 500 {object} map[string]interface{} "Internal server error when retrieving processing status"
// @Router /api/images/{id}/status [get]
func (h *PublicHandler) GetProcessingStatus(c *fiber.Ctx) error {
	imageID := c.Params("id")

	result, err := h.deps.PublicService.GetProcessingStatus(imageID)
	if err != nil {
		h.deps.Logger.FromContext(c).Error().
			Err(err).
			Str("handler", "GetProcessingStatus").
			Str("image_id", imageID).
			Msg("Failed to get processing status")

		var errorMsg string
		var statusCode int

		if err.Error() == "image not found" {
			errorMsg = "Image not found"
			statusCode = fiber.StatusNotFound
		} else {
			errorMsg = "Failed to get processing status"
			statusCode = fiber.StatusInternalServerError
		}

		errorResponse := fiber.Map{
			"error":      errorMsg,
			"request_id": c.Get("X-REQUEST-ID"),
		}
		if os.Getenv("ENVIRONMENT") == "development" || os.Getenv("ENVIRONMENT") == "dev" {
			errorResponse["details"] = err.Error()
		}
		return c.Status(statusCode).JSON(errorResponse)
	}

	h.deps.Logger.FromContext(c).Info().
		Str("handler", "GetProcessingStatus").
		Str("image_id", imageID).
		Str("status", result["status"].(string)).
		Msg("Processing status retrieved successfully")

	return c.JSON(result)
}

// @Summary Purchase coupon
// @Description Purchases a new coupon with card payment
// @Tags coupons
// @Accept json
// @Produce json
// @Param request body public.PurchaseCouponRequest true "Purchase parameters"
// @Success 201 {object} map[string]interface{} "Coupon purchased successfully"
// @Failure 400 {object} map[string]interface{} "Bad request: invalid purchase parameters"
// @Failure 403 {object} map[string]interface{} "Purchase not allowed for this partner"
// @Failure 500 {object} map[string]interface{} "Internal server error during coupon purchase"
// @Router /api/coupons/purchase [post]
func (h *PublicHandler) PurchaseCoupon(c *fiber.Ctx) error {
	branding := middleware.GetBrandingFromContext(c)
	if branding != nil && !branding.AllowPurchases {
		h.deps.Logger.FromContext(c).Warn().
			Str("handler", "PurchaseCoupon").
			Str("partner_code", branding.Partner.PartnerCode).
			Str("domain", branding.Partner.Domain).
			Msg("Purchase not allowed for this partner")

		errorResponse := fiber.Map{
			"error":      "Purchase not allowed for this partner",
			"request_id": c.Get("X-REQUEST-ID"),
		}
		return c.Status(fiber.StatusForbidden).JSON(errorResponse)
	}

	var req PurchaseCouponRequest
	if err := c.BodyParser(&req); err != nil {
		h.deps.Logger.FromContext(c).Warn().
			Err(err).
			Str("handler", "PurchaseCoupon").
			Msg("Failed to parse request body")

		errorResponse := fiber.Map{
			"error":      "Failed to parse request body",
			"request_id": c.Get("X-REQUEST-ID"),
		}
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse)
	}

	result, err := h.deps.PublicService.PurchaseCoupon(req)
	if err != nil {
		h.deps.Logger.FromContext(c).Error().
			Err(err).
			Str("handler", "PurchaseCoupon").
			Msg("Failed to purchase coupon")
			
		errorResponse := fiber.Map{
			"error":      "Failed to purchase coupon",
			"request_id": c.Get("X-REQUEST-ID"),
		}
		if os.Getenv("ENVIRONMENT") == "development" || os.Getenv("ENVIRONMENT") == "dev" {
			errorResponse["details"] = err.Error()
		}
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse)
	}

	resultWithBranding := middleware.BrandingResponse(c)
	resultWithBranding["purchase_result"] = result

	h.deps.Logger.FromContext(c).Info().
		Str("handler", "PurchaseCoupon").
		Str("size", req.Size).
		Str("style", req.Style).
		Int("amount", int(req.Amount)).
		Msg("Coupon purchased successfully")

	return c.Status(fiber.StatusCreated).JSON(resultWithBranding)
}

// @Summary Get available sizes
// @Description Returns list of available mosaic sizes
// @Tags public
// @Produce json
// @Success 200 {array} map[string]interface{} "Available sizes with dimensions and prices"
// @Failure 500 {object} map[string]interface{} "Internal server error when retrieving available sizes"
// @Router /api/sizes [get]
func (h *PublicHandler) GetAvailableSizes(c *fiber.Ctx) error {
	sizes := h.deps.PublicService.GetAvailableSizes()

	h.deps.Logger.FromContext(c).Info().
		Str("handler", "GetAvailableSizes").
		Int("count", len(sizes)).
		Msg("Available sizes retrieved successfully")

	return c.JSON(sizes)
}

// @Summary Get available styles
// @Description Returns list of available processing styles
// @Tags public
// @Produce json
// @Success 200 {array} map[string]interface{} "Available styles with descriptions and examples"
// @Failure 500 {object} map[string]interface{} "Internal server error when retrieving available styles"
// @Router /api/styles [get]
func (h *PublicHandler) GetAvailableStyles(c *fiber.Ctx) error {
	styles := h.deps.PublicService.GetAvailableStyles()

	logger := h.deps.Logger.FromContext(c)
	logger.Info().
		Str("handler", "GetAvailableStyles").
		Int("count", len(styles)).
		Msg("Available styles retrieved successfully")

	return c.JSON(styles)
}

// @Summary Get reCAPTCHA site key
// @Description Returns the public reCAPTCHA v3 site key for frontend
// @Tags public
// @Produce json
// @Success 200 {object} map[string]string "reCAPTCHA site key"
// @Failure 500 {object} map[string]interface{} "Internal server error when retrieving reCAPTCHA site key"
// @Router /api/config/recaptcha [get]
func (h *PublicHandler) GetRecaptchaSiteKey(c *fiber.Ctx) error {
	siteKey := h.deps.PublicService.GetRecaptchaSiteKey()

	logger := h.deps.Logger.FromContext(c)
	logger.Info().
		Str("handler", "GetRecaptchaSiteKey").
		Bool("has_key", siteKey != "").
		Msg("reCAPTCHA site key retrieved successfully")

	return c.JSON(fiber.Map{"site_key": siteKey})
}

// generateSchemaAsync generates a circuit asynchronously
func (h *PublicHandler) generateSchemaAsync(imageUUID uuid.UUID, confirmed bool, task *image.Image) {
	h.schemaPool.SubmitTask(func() {
		if err := h.deps.PublicService.GetImageService().GenerateSchema(context.Background(), imageUUID, confirmed); err != nil {
			return
		}

		if coupon, err := h.deps.PublicService.GetCouponRepository().GetByID(context.Background(), task.CouponID); err == nil {
			coupon.Status = "completed"
			if status, err := h.deps.PublicService.GetImageService().GetImageStatus(context.Background(), imageUUID); err == nil && status.ZipURL != nil {
				coupon.ZipURL = status.ZipURL
			}
			completedAt := time.Now()
			coupon.CompletedAt = &completedAt
			h.deps.PublicService.GetCouponRepository().Update(context.Background(), coupon)
		}
	})
}

// Close frees up handler resources
func (h *PublicHandler) Close() error {
	if h.goroutineManager != nil {
		return h.goroutineManager.Close()
	}
	return nil
}

// GetMetrics returns the metrics of the handler's operation
func (h *PublicHandler) GetMetrics() goroutine.Metrics {
	if h.goroutineManager != nil {
		return h.goroutineManager.GetMetrics()
	}
	return goroutine.Metrics{}
}
