package public

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/skr1ms/mosaic/internal/image"
	"github.com/skr1ms/mosaic/internal/types"
	"github.com/skr1ms/mosaic/pkg/marketplace"
	"github.com/skr1ms/mosaic/pkg/middleware"
)

type PublicHandlerDeps struct {
	PublicService PublicServiceInterface
	Logger        *middleware.Logger
}

type PublicHandler struct {
	fiber.Router
	deps *PublicHandlerDeps
}

func NewPublicHandler(router fiber.Router, deps *PublicHandlerDeps) *PublicHandler {
	handler := &PublicHandler{
		Router: router,
		deps:   deps,
	}

	// ================================================================
	// PUBLIC API ROUTES: /api/*
	// Access: public (no authentication required)
	// ================================================================
	router.Get("/branding", handler.GetBrandingInfo)                                      // GET /api/branding
	router.Get("/partners/:domain/info", handler.GetPartnerByDomain)                      // GET /api/partners/:domain/info
	router.Get("/partners/:id/articles", handler.GetPartnerArticles)                      // GET /api/partners/:id/articles
	router.Post("/partners/:id/articles/generate-url", handler.GeneratePartnerProductURL) // POST /api/partners/:id/articles/generate-url
	router.Get("/coupons/:code", handler.GetCouponByCode)                                 // GET /api/coupons/:code
	router.Post("/coupons/:code/activate", handler.ActivateCoupon)                        // POST /api/coupons/:code/activate
	router.Post("/coupons/purchase", handler.PurchaseCoupon)                              // POST /api/coupons/purchase
	router.Post("/images/upload", handler.UploadImage)                                    // POST /api/images/upload
	router.Post("/images/:id/edit", handler.EditImage)                                    // POST /api/images/:id/edit
	router.Post("/images/:id/process", handler.ProcessImage)                              // POST /api/images/:id/process
	router.Post("/images/:id/generate-schema", handler.GenerateSchema)                    // POST /api/images/:id/generate-schema
	router.Post("/images/:id/send-email", handler.SendSchemaToEmail)                      // POST /api/images/:id/send-email
	router.Get("/images/:id/preview", handler.GetImagePreview)                            // GET /api/images/:id/preview
	router.Get("/images/:id/status", handler.GetProcessingStatus)                         // GET /api/images/:id/status
	router.Get("/images/:id/download", handler.DownloadSchema)                            // GET /api/images/:id/download
	router.Get("/sizes", handler.GetAvailableSizes)                                       // GET /api/sizes
	router.Get("/styles", handler.GetAvailableStyles)                                     // GET /api/styles
	router.Get("/config/recaptcha", handler.GetRecaptchaSiteKey)                          // GET /api/config/recaptcha
	router.Post("/coupons/:code/reactivate", handler.ReactivateCoupon)                    // POST /api/coupons/:code/reactivate
	router.Post("/images/:id/search-page", handler.SearchSchemaPage)                      // POST /api/images/:id/search-page
	router.Get("/marketplace/status", handler.CheckMarketplaceStatus)                     // GET /api/marketplace/status

	// Public preview routes
	public := handler.Group("/preview")
	public.Post("/generate-variant", handler.GeneratePreview)              // POST /api/preview/generate-variant
	public.Post("/generate-variants", handler.GenerateVariants)            // POST /api/preview/generate-variants
	public.Post("/generate-style-variants", handler.GenerateStyleVariants) // POST /api/preview/generate-style-variants
	public.Post("/generate-ai", handler.GenerateAIPreview)                 // POST /api/preview/generate-ai
	public.Post("/generate-all", handler.GenerateAllPreviews)              // POST /api/preview/generate-all
	public.Get("/:id", handler.GetPreview)                                 // GET /api/preview/:id
	public.Delete("/:id", handler.DeletePreview)                           // DELETE /api/preview/:id
	public.Post("/cleanup-all", handler.CleanupAllPreviews)                // POST /api/preview/cleanup-all (EMERGENCY)

	return handler
}

// @Summary Get branding information
// @Description Returns branding data (logo, contacts, links) for the current domain
// @Tags public
// @Produce json
// @Success 200 {object} map[string]any "Branding data including logo, contacts, and links"
// @Failure 500 {object} map[string]any "Internal server error when retrieving branding data"
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
							if newURL, err := h.deps.PublicService.GetS3Client().GetLogoURL(c.UserContext(), objectKey, 24*time.Hour); err == nil {
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
// @Success 200 {object} map[string]any "Partner information including branding and contact details"
// @Failure 404 {object} map[string]any "Partner not found for the specified domain"
// @Failure 500 {object} map[string]any "Internal server error when retrieving partner information"
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

// @Summary Get partner articles
// @Description Returns articles for a specific partner
// @Tags partners
// @Produce json
// @Param id path string true "Partner ID"
// @Success 200 {array} map[string]any "List of partner articles"
// @Failure 404 {object} map[string]any "Partner not found"
// @Failure 500 {object} map[string]any "Internal server error when retrieving partner articles"
// @Router /api/partners/{id}/articles [get]
func (h *PublicHandler) GetPartnerArticles(c *fiber.Ctx) error {
	partnerID := c.Params("id")

	partnerUUID, err := uuid.Parse(partnerID)
	if err != nil {
		h.deps.Logger.FromContext(c).Warn().
			Err(err).
			Str("handler", "GetPartnerArticles").
			Str("partner_id", partnerID).
			Msg("Invalid partner ID format")

		errorResponse := fiber.Map{
			"error":      "Invalid partner ID format",
			"request_id": c.Get("X-Request-ID"),
		}
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse)
	}

	articles, err := h.deps.PublicService.GetPartnerArticleGrid(partnerUUID)
	if err != nil {
		h.deps.Logger.FromContext(c).Error().
			Err(err).
			Str("handler", "GetPartnerArticles").
			Str("partner_id", partnerID).
			Msg("Failed to get partner articles")

		var errorMsg string
		var statusCode int

		if err.Error() == "partner not found" {
			errorMsg = "Partner not found"
			statusCode = fiber.StatusNotFound
		} else {
			errorMsg = "Failed to get partner articles"
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
		Str("handler", "GetPartnerArticles").
		Str("partner_id", partnerID).
		Int("count", len(articles)).
		Msg("Partner articles retrieved successfully")

	return c.JSON(articles)
}

// @Summary Get coupon information by code
// @Description Returns coupon information for validation purposes
// @Tags coupons
// @Produce json
// @Param code path string true "Coupon code (12 digits)"
// @Success 200 {object} map[string]any "Coupon information including status, partner, and creation date"
// @Failure 400 {object} map[string]any "Invalid coupon code format"
// @Failure 404 {object} map[string]any "Coupon not found"
// @Failure 500 {object} map[string]any "Internal server error when retrieving coupon"
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
// @Success 200 {object} map[string]any "Coupon activated successfully"
// @Failure 400 {object} map[string]any "Invalid request format"
// @Failure 404 {object} map[string]any "Coupon not found"
// @Failure 409 {object} map[string]any "Coupon already used"
// @Failure 500 {object} map[string]any "Internal server error during coupon activation"
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
// @Success 201 {object} map[string]any "Image uploaded successfully"
// @Failure 400 {object} map[string]any "Bad request: missing required fields or invalid data"
// @Failure 413 {object} map[string]any "File too large"
// @Failure 500 {object} map[string]any "Internal server error during image upload"
// @Router /api/images/upload [post]
func (h *PublicHandler) UploadImage(c *fiber.Ctx) error {
	couponID := c.FormValue("coupon_id")
	couponCode := c.FormValue("coupon_code")

	h.deps.Logger.FromContext(c).Info().
		Str("handler", "UploadImage").
		Str("coupon_id", couponID).
		Str("coupon_code", couponCode).
		Str("content_type", c.Get("Content-Type")).
		Msg("Processing upload request")

	file, err := c.FormFile("image")
	if err != nil {
		h.deps.Logger.FromContext(c).Warn().
			Err(err).
			Str("handler", "UploadImage").
			Interface("form_keys", c.Request().MultipartForm).
			Msg("Image file is required - FormFile failed")

		errorResponse := fiber.Map{
			"error":      "Image file is required",
			"request_id": c.Get("X-Request-ID"),
			"details":    err.Error(),
		}
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse)
	}

	// Log file details
	h.deps.Logger.FromContext(c).Info().
		Str("handler", "UploadImage").
		Str("filename", file.Filename).
		Int64("size", file.Size).
		Str("content_type", file.Header.Get("Content-Type")).
		Msg("File received successfully")

	// Handle coupon processing if provided
	if couponID == "" && couponCode != "" {
		cleanCode := strings.TrimSpace(strings.ReplaceAll(couponCode, "-", ""))

		h.deps.Logger.FromContext(c).Info().
			Str("handler", "UploadImage").
			Str("original_coupon_code", couponCode).
			Str("cleaned_coupon_code", cleanCode).
			Int("cleaned_length", len(cleanCode)).
			Msg("Processing coupon code")

		if len(cleanCode) != 12 {
			h.deps.Logger.FromContext(c).Warn().
				Str("handler", "UploadImage").
				Str("original_coupon_code", couponCode).
				Str("cleaned_coupon_code", cleanCode).
				Int("expected_length", 12).
				Int("actual_length", len(cleanCode)).
				Msg("Coupon code must be 12 digits")

			errorResponse := fiber.Map{
				"error":           "Coupon code must be 12 digits",
				"request_id":      c.Get("X-Request-ID"),
				"received_length": len(cleanCode),
				"expected_length": 12,
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
		Interface("image_id", result["image_id"]).
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
// @Success 200 {object} map[string]any "Image edited successfully"
// @Failure 400 {object} map[string]any "Bad request: invalid editing parameters"
// @Failure 404 {object} map[string]any "Image not found"
// @Failure 500 {object} map[string]any "Internal server error during image editing"
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
// @Success 200 {object} map[string]any "Image processing started"
// @Failure 400 {object} map[string]any "Bad request: invalid processing parameters"
// @Failure 404 {object} map[string]any "Image not found"
// @Failure 500 {object} map[string]any "Internal server error during image processing"
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
// @Success 200 {object} map[string]any "Schema generation started"
// @Failure 400 {object} map[string]any "Bad request: invalid schema generation parameters"
// @Failure 404 {object} map[string]any "Image not found"
// @Failure 500 {object} map[string]any "Internal server error during schema generation"
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
// @Failure 400 {object} map[string]any "Invalid image ID format"
// @Failure 404 {object} map[string]any "Image not found"
// @Failure 409 {object} map[string]any "Schema not ready"
// @Failure 500 {object} map[string]any "Internal server error during schema download"
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
// @Success 200 {object} map[string]any "Schema sent successfully"
// @Failure 400 {object} map[string]any "Bad request: invalid email or request format"
// @Failure 404 {object} map[string]any "Image not found"
// @Failure 500 {object} map[string]any "Internal server error during email sending"
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
// @Success 200 {object} map[string]any "Image preview with URL and dimensions"
// @Failure 404 {object} map[string]any "Image not found"
// @Failure 500 {object} map[string]any "Internal server error when retrieving image preview"
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
// @Success 200 {object} map[string]any "Processing status including current step and progress"
// @Failure 404 {object} map[string]any "Image not found"
// @Failure 500 {object} map[string]any "Internal server error when retrieving processing status"
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
// @Success 201 {object} map[string]any "Coupon purchased successfully"
// @Failure 400 {object} map[string]any "Bad request: invalid purchase parameters"
// @Failure 403 {object} map[string]any "Purchase not allowed for this partner"
// @Failure 500 {object} map[string]any "Internal server error during coupon purchase"
// @Router /api/coupons/purchase [post]
func (h *PublicHandler) PurchaseCoupon(c *fiber.Ctx) error {
	if h == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Handler is nil"})
	}
	if h.deps == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Handler not initialized"})
	}
	if h.deps.Logger == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Logger not initialized"})
	}

	h.deps.Logger.FromContext(c).Info().Msg("PurchaseCoupon called")
	if h.deps.PublicService == nil {
		h.deps.Logger.FromContext(c).Error().Msg("PublicService is nil")
		return c.Status(500).JSON(fiber.Map{"error": "PublicService not initialized"})
	}

	h.deps.Logger.FromContext(c).Info().Msg("Getting branding from context")
	branding := middleware.GetBrandingFromContext(c)
	h.deps.Logger.FromContext(c).Info().Interface("branding", branding).Msg("Branding retrieved")
	if branding != nil && !branding.AllowPurchases {
		partnerCode := ""
		domain := ""
		if branding.Partner != nil {
			partnerCode = branding.Partner.PartnerCode
			domain = branding.Partner.Domain
		}
		h.deps.Logger.FromContext(c).Warn().
			Str("handler", "PurchaseCoupon").
			Str("partner_code", partnerCode).
			Str("domain", domain).
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
// @Success 200 {array} map[string]any "Available sizes with dimensions and prices"
// @Failure 500 {object} map[string]any "Internal server error when retrieving available sizes"
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
// @Success 200 {array} map[string]any "Available styles with descriptions and examples"
// @Failure 500 {object} map[string]any "Internal server error when retrieving available styles"
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
// @Failure 500 {object} map[string]any "Internal server error when retrieving reCAPTCHA site key"
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

// @Summary Generate partner product URL
// @Description Generates a marketplace product URL for a specific partner, size, and style combination
// @Tags partners
// @Accept json
// @Produce json
// @Param id path string true "Partner ID"
// @Param request body GenerateProductURLRequest true "URL generation request"
// @Success 200 {object} GenerateProductURLResponse "Generated product URL with details"
// @Failure 400 {object} map[string]any "Invalid request"
// @Failure 404 {object} map[string]any "Partner not found"
// @Failure 500 {object} map[string]any "Internal server error"
// @Router /api/partners/{id}/articles/generate-url [post]
func (h *PublicHandler) GeneratePartnerProductURL(c *fiber.Ctx) error {
	partnerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid partner ID",
		})
	}

	var req GenerateProductURLRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate parameters
	validMarketplaces := map[string]bool{"ozon": true, "wildberries": true}
	validStyles := map[string]bool{"grayscale": true, "skin_tones": true, "pop_art": true, "max_colors": true}
	validSizes := map[string]bool{"21x30": true, "30x40": true, "40x40": true, "40x50": true, "40x60": true, "50x70": true}

	if !validMarketplaces[req.Marketplace] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid marketplace",
		})
	}

	if !validStyles[req.Style] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid style",
		})
	}

	if !validSizes[req.Size] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid size",
		})
	}

	result, err := h.deps.PublicService.GeneratePartnerProductURL(partnerID, req)
	if err != nil {
		if strings.Contains(err.Error(), "partner not found") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Partner not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate product URL",
		})
	}

	h.deps.Logger.FromContext(c).Info().
		Str("handler", "GeneratePartnerProductURL").
		Str("partner_id", partnerID.String()).
		Str("marketplace", req.Marketplace).
		Str("style", req.Style).
		Str("size", req.Size).
		Msg("Product URL generated successfully")

	return c.JSON(result)
}

// generateSchemaAsync generates a circuit asynchronously
func (h *PublicHandler) generateSchemaAsync(imageUUID uuid.UUID, confirmed bool, task *image.Image) {
	go func() {
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
	}()
}

// GeneratePreview generates a preview with specified style
func (h *PublicHandler) GeneratePreview(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Parse multipart form
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid form data",
		})
	}

	// Get image file
	files := form.File["image"]
	if len(files) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Image file is required",
		})
	}

	file := files[0]

	// Get parameters
	size := c.FormValue("size", "30x40")
	style := c.FormValue("style", "grayscale")
	lighting := c.FormValue("lighting", "sun")
	contrastLevel := c.FormValue("contrast_level", "normal")

	// Generate preview ID
	previewID := uuid.New().String()

	// Process image
	previewData, err := h.deps.PublicService.GeneratePreview(ctx, file, size, style, lighting, contrastLevel)
	if err != nil {
		h.deps.Logger.FromContext(c).Error().
			Err(err).
			Str("handler", "GeneratePreview").
			Msg("Failed to generate preview")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate preview",
		})
	}

	return c.JSON(fiber.Map{
		"preview_id":  previewID,
		"preview_url": previewData.URL,
		"style":       style,
		"size":        size,
		"contrast":    contrastLevel,
	})
}

// GenerateVariants generates multiple style variants
func (h *PublicHandler) GenerateVariants(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Parse multipart form
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid form data",
		})
	}

	// Get image file
	files := form.File["image"]
	if len(files) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Image file is required",
		})
	}

	file := files[0]
	size := c.FormValue("size", "30x40")

	// Define style variants
	variants := []struct {
		Style    string `json:"style"`
		Contrast string `json:"contrast"`
		Label    string `json:"label"`
	}{
		{"venus", "soft", "Венера (мягкий контраст)"},
		{"venus", "strong", "Венера (сильный контраст)"},
		{"sun", "soft", "Солнце (мягкий контраст)"},
		{"sun", "strong", "Солнце (сильный контраст)"},
		{"moon", "soft", "Луна (мягкий контраст)"},
		{"moon", "strong", "Луна (сильный контраст)"},
		{"mars", "soft", "Марс (мягкий контраст)"},
		{"mars", "strong", "Марс (сильный контраст)"},
	}

	var previews []fiber.Map

	for _, variant := range variants {
		// For variants, use style as lighting and keep grayscale as base style
		previewData, err := h.deps.PublicService.GeneratePreview(ctx, file, size, "grayscale", variant.Style, variant.Contrast)
		if err != nil {
			h.deps.Logger.FromContext(c).
				Error().
				Err(err).
				Str("handler", "GenerateVariants").
				Str("style", variant.Style).
				Msg("Failed to generate variant")
			continue
		}

		previews = append(previews, fiber.Map{
			"style":       variant.Style,
			"contrast":    variant.Contrast,
			"label":       variant.Label,
			"preview_url": previewData.URL,
		})
	}

	return c.JSON(fiber.Map{
		"previews": previews,
		"total":    len(previews),
	})
}

// GenerateStyleVariants generates previews for all 4 main styles
func (h *PublicHandler) GenerateStyleVariants(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Parse multipart form
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid form data",
		})
	}

	// Get image file
	files := form.File["image"]
	if len(files) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Image file is required",
		})
	}

	file := files[0]
	size := c.FormValue("size", "30x40")

	// Define the 4 main styles
	styles := []struct {
		Key   string `json:"key"`
		Label string `json:"label"`
	}{
		{"grayscale", "Grayscale"},
		{"skin_tones", "Skin Tones"},
		{"pop_art", "Pop Art"},
		{"max_colors", "Maximum Colors"},
	}

	type previewResult struct {
		Style string    `json:"style"`
		Label string    `json:"label"`
		URL   string    `json:"preview_url"`
		ID    uuid.UUID `json:"preview_id"`
		Error error     `json:"-"`
	}

	resultChan := make(chan previewResult, len(styles))

	// Launch all 4 style generations in parallel
	for _, style := range styles {
		style := style // Capture loop variable
		go func() {
			previewData, err := h.deps.PublicService.GenerateStylePreview(ctx, file, size, style.Key)
			resultChan <- previewResult{
				Style: style.Key,
				Label: style.Label,
				URL: func() string {
					if previewData != nil {
						return previewData.URL
					}
					return ""
				}(),
				ID: func() uuid.UUID {
					if previewData != nil {
						return previewData.ID
					}
					return uuid.Nil
				}(),
				Error: err,
			}
		}()
	}

	// Collect results from all 4 goroutines
	var previews []fiber.Map
	for i := 0; i < len(styles); i++ {
		result := <-resultChan
		if result.Error != nil {
			h.deps.Logger.FromContext(c).
				Error().
				Err(result.Error).
				Str("handler", "GenerateStyleVariants").
				Str("style", result.Style).
				Msg("Failed to generate style variant")
			continue
		}

		previews = append(previews, fiber.Map{
			"style":       result.Style,
			"label":       result.Label,
			"preview_url": result.URL,
			"preview_id":  result.ID,
		})
	}

	return c.JSON(fiber.Map{
		"previews": previews,
		"total":    len(previews),
		"size":     size,
	})
}

// GenerateAIPreview generates AI-enhanced previews using Stable Diffusion
func (h *PublicHandler) GenerateAIPreview(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Parse multipart form
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid form data",
		})
	}

	// Get image file
	files := form.File["image"]
	if len(files) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Image file is required",
		})
	}

	file := files[0]

	// Get parameters
	prompt := c.FormValue("prompt", "artistic mosaic style, high quality, detailed")
	numVariants := 2 // Generate 2 AI variants by default

	var aiPreviews []fiber.Map

	for i := 0; i < numVariants; i++ {
		// Generate AI preview
		aiPreviewData, err := h.deps.PublicService.GenerateAIPreview(ctx, file, prompt)
		if err != nil {
			continue
		}

		aiPreviews = append(aiPreviews, fiber.Map{
			"preview_url": aiPreviewData.URL,
			"prompt":      prompt,
			"variant":     i + 1,
		})
	}

	return c.JSON(fiber.Map{
		"ai_previews": aiPreviews,
		"total":       len(aiPreviews),
	})
}

// GetPreview retrieves a preview by ID
func (h *PublicHandler) GetPreview(c *fiber.Ctx) error {
	previewIDStr := c.Params("id")
	if previewIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Preview ID is required",
		})
	}

	// Parse UUID
	previewID, err := uuid.Parse(previewIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid preview ID format",
		})
	}

	// Get preview from database
	preview, err := h.deps.PublicService.GetPublicRepository().GetByID(c.Context(), previewID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Preview not found",
		})
	}

	return c.JSON(preview)
}

// DeletePreview deletes a preview by ID
func (h *PublicHandler) DeletePreview(c *fiber.Ctx) error {
	previewIDStr := c.Params("id")
	if previewIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Preview ID is required",
		})
	}

	// Parse UUID
	previewID, err := uuid.Parse(previewIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid preview ID format",
		})
	}

	// Get preview first to get S3 key
	preview, err := h.deps.PublicService.GetPublicRepository().GetByID(c.Context(), previewID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Preview not found",
		})
	}

	// Delete from database
	if err := h.deps.PublicService.GetPublicRepository().Delete(c.Context(), previewID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete preview from database",
		})
	}

	// Schedule S3 deletion (async)
	if preview.S3Key != "" {
		h.deps.PublicService.GetS3Client().SchedulePreviewDeletion(preview.S3Key)
	}

	return c.JSON(fiber.Map{
		"message": "Preview deleted successfully",
	})
}

// GenerateAllPreviews generates all 8 base previews + optional 1 AI preview
// @Summary Generate all preview variants
// @Description Generates 8 base previews (4 styles × 2 contrasts) and optionally 1 AI preview
// @Tags preview
// @Accept multipart/form-data
// @Produce json
// @Param image_id formData string false "Image ID (optional for preview uploads)"
// @Param size formData string true "Size (e.g., 30x40)"
// @Param use_ai formData bool false "Generate AI preview"
// @Param image formData file false "Image file (required if image_id not found in DB)"
// @Success 200 {object} GenerateAllPreviewsResponse "All previews generated successfully"
// @Failure 400 {object} map[string]any "Invalid request"
// @Failure 500 {object} map[string]any "Internal server error"
// @Router /api/preview/generate-all [post]
func (h *PublicHandler) GenerateAllPreviews(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Get form parameters
	imageID := c.FormValue("image_id")
	size := c.FormValue("size", "30x40")
	useAI := c.FormValue("use_ai") == "true"

	h.deps.Logger.FromContext(c).Info().
		Str("handler", "GenerateAllPreviews").
		Str("image_id", imageID).
		Str("size", size).
		Bool("use_ai", useAI).
		Msg("Starting preview generation")

	// Try method 1: Use existing image from database
	if imageID != "" {
		result, err := h.deps.PublicService.GenerateAllPreviews(ctx, imageID, size, useAI)
		if err == nil {
			h.deps.Logger.FromContext(c).Info().
				Str("handler", "GenerateAllPreviews").
				Str("image_id", imageID).
				Int("total_previews", result.Total).
				Bool("use_ai", useAI).
				Msg("All previews generated successfully from database")
			return c.JSON(result)
		}

		// Log the error but continue to method 2
		h.deps.Logger.FromContext(c).Warn().
			Err(err).
			Str("handler", "GenerateAllPreviews").
			Str("image_id", imageID).
			Msg("Failed to generate from database, trying file upload method")
	}

	// Method 2: Use uploaded file (fallback for preview uploads)
	file, err := c.FormFile("image")
	if err != nil || file == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Image file is required when image_id is not found in database",
		})
	}

	h.deps.Logger.FromContext(c).Info().
		Str("handler", "GenerateAllPreviews").
		Str("filename", file.Filename).
		Int64("file_size", file.Size).
		Msg("Using uploaded file for preview generation")

	// Generate previews from uploaded file
	result, err := h.deps.PublicService.GenerateAllPreviewsFromFile(ctx, file, size, useAI)
	if err != nil {
		h.deps.Logger.FromContext(c).Error().
			Err(err).
			Str("handler", "GenerateAllPreviews").
			Str("filename", file.Filename).
			Msg("Failed to generate all previews from file")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to generate previews",
			"details": err.Error(),
		})
	}

	h.deps.Logger.FromContext(c).Info().
		Str("handler", "GenerateAllPreviews").
		Str("filename", file.Filename).
		Int("total_previews", result.Total).
		Bool("use_ai", useAI).
		Msg("All previews generated successfully from uploaded file")

	return c.JSON(result)
}

// ReactivateCoupon handles re-access to an already activated coupon
// @Summary Reactivate coupon
// @Description Provides access to already activated coupon data and schema
// @Tags coupons
// @Accept json
// @Produce json
// @Param code path string true "Coupon code"
// @Param request body ReactivateCouponRequest true "Reactivation request"
// @Success 200 {object} ReactivateCouponResponse "Coupon data retrieved"
// @Failure 400 {object} map[string]any "Invalid request"
// @Failure 404 {object} map[string]any "Coupon not found"
// @Failure 500 {object} map[string]any "Internal server error"
// @Router /api/coupons/{code}/reactivate [post]
func (h *PublicHandler) ReactivateCoupon(c *fiber.Ctx) error {
	code := c.Params("code")

	var req ReactivateCouponRequest
	if err := c.BodyParser(&req); err != nil {
		req.Code = code // Use path parameter if body is empty
	}

	if req.Code == "" {
		req.Code = code
	}

	ctx := context.Background()
	result, err := h.deps.PublicService.ReactivateCoupon(ctx, req.Code)
	if err != nil {
		h.deps.Logger.FromContext(c).Error().
			Err(err).
			Str("handler", "ReactivateCoupon").
			Str("code", req.Code).
			Msg("Failed to reactivate coupon")

		var statusCode int
		if strings.Contains(err.Error(), "not found") {
			statusCode = fiber.StatusNotFound
		} else if strings.Contains(err.Error(), "not activated") {
			statusCode = fiber.StatusBadRequest
		} else {
			statusCode = fiber.StatusInternalServerError
		}

		return c.Status(statusCode).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Send email if requested
	if req.Email != "" && result.CanSendEmail && result.ArchiveURL != "" {
		if coupon, err := h.deps.PublicService.GetCouponRepository().GetByCode(ctx, req.Code); err == nil {
			h.deps.PublicService.GetEmailService().SendSchemaEmail(req.Email, result.ArchiveURL, coupon.Code)
		}
	}

	h.deps.Logger.FromContext(c).Info().
		Str("handler", "ReactivateCoupon").
		Str("code", req.Code).
		Msg("Coupon reactivated successfully")

	return c.JSON(result)
}

// SearchSchemaPage searches for a specific page in the schema
// @Summary Search schema page
// @Description Finds and returns URL for a specific page number in the schema
// @Tags images
// @Accept json
// @Produce json
// @Param id path string true "Image ID"
// @Param request body SearchSchemaPageRequest true "Page search request"
// @Success 200 {object} SearchSchemaPageResponse "Page found"
// @Failure 400 {object} map[string]any "Invalid request"
// @Failure 404 {object} map[string]any "Page not found"
// @Failure 500 {object} map[string]any "Internal server error"
// @Router /api/images/{id}/search-page [post]
func (h *PublicHandler) SearchSchemaPage(c *fiber.Ctx) error {
	imageID := c.Params("id")

	var req SearchSchemaPageRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.ImageID == "" {
		req.ImageID = imageID
	}

	if err := middleware.ValidateStruct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Validation failed",
			"details": err.Error(),
		})
	}

	ctx := context.Background()
	result, err := h.deps.PublicService.SearchSchemaPage(ctx, req.ImageID, req.PageNumber)
	if err != nil {
		h.deps.Logger.FromContext(c).Error().
			Err(err).
			Str("handler", "SearchSchemaPage").
			Str("image_id", req.ImageID).
			Int("page_number", req.PageNumber).
			Msg("Failed to search schema page")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to search page",
			"details": err.Error(),
		})
	}

	h.deps.Logger.FromContext(c).Info().
		Str("handler", "SearchSchemaPage").
		Str("image_id", req.ImageID).
		Int("page_number", req.PageNumber).
		Bool("found", result.Found).
		Msg("Schema page search completed")

	return c.JSON(result)
}

// CheckMarketplaceStatus checks product availability on marketplaces
// @Summary Check marketplace status
// @Description Checks if a product is available on Ozon or Wildberries
// @Tags marketplace
// @Accept json
// @Produce json
// @Param request body MarketplaceStatusRequest true "Marketplace check request"
// @Success 200 {object} MarketplaceStatusResponse "Status retrieved"
// @Failure 400 {object} map[string]any "Invalid request"
// @Failure 500 {object} map[string]any "Internal server error"
// @Router /api/marketplace/status [get]
func (h *PublicHandler) CheckMarketplaceStatus(c *fiber.Ctx) error {
	marketplaceStr := c.Query("marketplace")
	partnerIDStr := c.Query("partner_id")
	size := c.Query("size")
	style := c.Query("style")
	sku := c.Query("sku")

	if marketplaceStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Marketplace is required",
		})
	}

	// Create marketplace service
	marketplaceRepo := marketplace.NewPartnerRepositoryAdapter(h.deps.PublicService.GetPartnerRepository())
	marketplaceService := marketplace.NewService(marketplaceRepo)

	// Parse partner ID if provided
	var partnerID *uuid.UUID
	if partnerIDStr != "" {
		if pid, err := uuid.Parse(partnerIDStr); err == nil {
			partnerID = &pid
		}
	}

	// Create availability request
	availabilityReq := &marketplace.ProductAvailabilityRequest{
		PartnerID:   partnerID,
		Marketplace: marketplace.Marketplace(marketplaceStr),
		Size:        size,
		Style:       style,
		SKU:         sku,
	}

	// Check availability using marketplace service
	response, err := marketplaceService.CheckProductAvailability(availabilityReq)
	if err != nil {
		if strings.Contains(err.Error(), "validation failed") {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to check marketplace status",
		})
	}

	// Convert to public response format
	publicResponse := MarketplaceStatusResponse{
		Marketplace: string(response.Marketplace),
		SKU:         response.SKU,
		Available:   response.Available,
		ProductURL:  response.ProductURL,
	}

	h.deps.Logger.FromContext(c).Info().
		Str("handler", "CheckMarketplaceStatus").
		Str("marketplace", marketplaceStr).
		Str("partner_id", partnerIDStr).
		Str("size", size).
		Str("style", style).
		Str("sku", publicResponse.SKU).
		Bool("available", publicResponse.Available).
		Msg("Marketplace status checked")

	return c.JSON(publicResponse)
}

// CleanupAllPreviews emergency endpoint to cleanup all old preview files from MinIO
// @Summary Emergency cleanup of old preview files
// @Description Mass deletes all preview files older than 1 hour from MinIO (EMERGENCY USE ONLY)
// @Tags preview
// @Produce json
// @Success 200 {object} map[string]any "Cleanup completed successfully"
// @Failure 500 {object} map[string]any "Internal server error during cleanup"
// @Router /api/preview/cleanup-all [post]
func (h *PublicHandler) CleanupAllPreviews(c *fiber.Ctx) error {
	h.deps.Logger.FromContext(c).Warn().
		Str("handler", "CleanupAllPreviews").
		Str("ip", c.IP()).
		Msg("EMERGENCY: Preview cleanup requested")

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second) // 5 minutes timeout
	defer cancel()

	err := h.deps.PublicService.GetS3Client().CleanupAllPreviews(ctx)
	if err != nil {
		h.deps.Logger.FromContext(c).Error().
			Err(err).
			Str("handler", "CleanupAllPreviews").
			Msg("Failed to cleanup previews")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":      "Failed to cleanup previews",
			"request_id": c.Get("X-Request-ID"),
		})
	}

	h.deps.Logger.FromContext(c).Info().
		Str("handler", "CleanupAllPreviews").
		Msg("Preview cleanup completed successfully")

	return c.JSON(fiber.Map{
		"message": "Preview cleanup completed successfully",
		"note":    "All preview files older than 1 hour have been deleted",
	})
}
