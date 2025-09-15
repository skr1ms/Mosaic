package coupon

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/skr1ms/mosaic/pkg/middleware"
)

type CouponHandlerDeps struct {
	CouponService CouponServiceInterface
	Logger        *middleware.Logger
}

type CouponHandler struct {
	fiber.Router
	deps *CouponHandlerDeps
}

func NewCouponHandler(router fiber.Router, deps *CouponHandlerDeps) {
	handler := &CouponHandler{
		Router: router,
		deps:   deps,
	}

	// ================================================================
	// COUPON MANAGEMENT ROUTES: /api/coupons/*
	// Access: public (no authentication required)
	// ================================================================
	api := handler.Group("/coupons")
	api.Get("/", handler.GetCoupons)                             // GET /api/coupons/
	api.Get("/paginated", handler.GetCouponsPaginated)           // GET /api/coupons/paginated
	api.Get("/export", handler.ExportCoupons)                    // GET /api/coupons/export
	api.Get("/export-advanced", handler.ExportCouponsAdvanced)   // GET /api/coupons/export-advanced
	api.Get("/statistics", handler.GetStatistics)                // GET /api/coupons/statistics
	api.Get("/partner/:partner_id", handler.GetCouponsByPartner) // GET /api/coupons/partner/:partner_id
	api.Get("/code/:code", handler.GetCouponByCode)              // GET /api/coupons/code/:code

	// Apply rate limiting to coupon validation and activation endpoints
	api.Post("/code/:code/validate", middleware.CouponActivationRateLimiter(deps.Logger), handler.ValidateCoupon)      // POST /api/coupons/code/:code/validate
	api.Get("/:id", handler.GetCouponByID)                                                                             // GET /api/coupons/:id
	api.Put("/:id/activate", middleware.CouponActivationRateLimiter(deps.Logger), handler.ActivateCoupon)              // PUT /api/coupons/:id/activate
	api.Put("/:id/reset", handler.ResetCoupon)                                                                         // PUT /api/coupons/:id/reset
	api.Put("/:id/send-schema", handler.SendSchema)                                                                    // PUT /api/coupons/:id/send-schema
	api.Put("/:id/purchase", middleware.CouponActivationRateLimiter(deps.Logger), handler.MarkAsPurchased)             // PUT /api/coupons/:id/purchase
	api.Get("/:id/download-materials", middleware.CouponActivationRateLimiter(deps.Logger), handler.DownloadMaterials) // GET /api/coupons/:id/download-materials
	api.Post("/batch/reset", handler.BatchResetCoupons)                                                                // POST /api/coupons/batch/reset
	api.Post("/batch/delete/preview", handler.PreviewBatchDelete)                                                      // POST /api/coupons/batch/delete/preview
	api.Post("/batch/delete/confirm", handler.ExecuteBatchDelete)                                                      // POST /api/coupons/batch/delete/confirm
}

// @Summary Get coupons list with filtering
// @Description Returns list of coupons with filtering capabilities by code, status, size, style and partner
// @Tags coupons
// @Produce json
// @Param code query string false "Coupon code for search"
// @Param status query string false "Coupon status (new, used)"
// @Param size query string false "Coupon size"
// @Param style query string false "Coupon style"
// @Param partner_id query string false "Partner ID"
// @Success 200 {array} map[string]any "List of coupons"
// @Failure 500 {object} map[string]any "Internal server error"
// @Router /coupons [get]
func (handler *CouponHandler) GetCoupons(c *fiber.Ctx) error {
	code := c.Query("code")
	status := c.Query("status")
	size := c.Query("size")
	style := c.Query("style")
	partnerIDStr := c.Query("partner_id")

	var partnerID *uuid.UUID
	if partnerIDStr != "" {
		if id, err := uuid.Parse(partnerIDStr); err == nil {
			partnerID = &id
		}
	}

	coupons, err := handler.deps.CouponService.SearchCoupons(code, status, size, style, partnerID)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to search coupons")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to search coupons",
		})
	}

	handler.deps.Logger.FromContext(c).Info().
		Int("count", len(coupons)).
		Str("code", code).
		Str("status", status).
		Str("size", size).
		Str("style", style).
		Str("partner_id", partnerIDStr).
		Msg("Coupons retrieved successfully")

	return c.JSON(coupons)
}

// @Summary Get coupons list with pagination
// @Description Returns paginated list of coupons with filtering capabilities
// @Tags coupons
// @Produce json
// @Param page query int false "Page number (default 1)"
// @Param limit query int false "Items per page (default 20, max 100)"
// @Param code query string false "Coupon code for search"
// @Param status query string false "Coupon status (new, used)"
// @Param size query string false "Coupon size"
// @Param style query string false "Coupon style"
// @Param partner_id query string false "Partner ID"
// @Success 200 {object} map[string]any "Coupons with pagination info"
// @Failure 400 {object} map[string]any "Invalid request parameters"
// @Failure 500 {object} map[string]any "Internal server error"
// @Router /coupons/paginated [get]
func (handler *CouponHandler) GetCouponsPaginated(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	code := c.Query("code")
	status := c.Query("status")
	size := c.Query("size")
	style := c.Query("style")
	partnerIDStr := c.Query("partner_id")

	var partnerID *uuid.UUID
	if partnerIDStr != "" {
		if id, err := uuid.Parse(partnerIDStr); err == nil {
			partnerID = &id
		}
	}

	coupons, total, err := handler.deps.CouponService.SearchCouponsWithPagination(
		code, status, size, style, partnerID, page, limit,
	)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to get coupons with pagination")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get coupons with pagination",
		})
	}

	totalPages := (total + int64(limit) - 1) / int64(limit)
	hasNext := int64(page) < totalPages
	hasPrev := page > 1

	handler.deps.Logger.FromContext(c).Info().
		Int("page", page).
		Int("limit", limit).
		Int64("total", total).
		Int64("total_pages", totalPages).
		Msg("Paginated coupons retrieved successfully")

	return c.JSON(fiber.Map{
		"coupons": coupons,
		"pagination": fiber.Map{
			"current_page": page,
			"per_page":     limit,
			"total":        total,
			"total_pages":  totalPages,
			"has_next":     hasNext,
			"has_previous": hasPrev,
		},
	})
}

// @Summary Export coupons
// @Description Exports coupons to text file with filtering options
// @Tags coupons
// @Produce text/plain
// @Param partner_id query string false "Partner ID for filtering"
// @Param status query string false "Coupon status for export"
// @Param format query string false "Export format: codes (codes only) or full (full information)"
// @Success 200 {string} string "Text file with coupons"
// @Failure 500 {object} map[string]any "Internal server error"
// @Router /coupons/export [get]
func (handler *CouponHandler) ExportCoupons(c *fiber.Ctx) error {
	partnerIDStr := c.Query("partner_id")
	status := c.Query("status")
	format := c.Query("format", "codes")

	var partnerID *uuid.UUID
	if partnerIDStr != "" {
		if id, err := uuid.Parse(partnerIDStr); err == nil {
			partnerID = &id
		}
	}

	content, err := handler.deps.CouponService.ExportCoupons(partnerID, status, format)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to export coupons")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to export coupons",
		})
	}

	filename := fmt.Sprintf("coupons_export_%s.txt", time.Now().Format("20060102_150405"))
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Set("Content-Type", "text/plain")

	handler.deps.Logger.FromContext(c).Info().
		Str("filename", filename).
		Str("format", format).
		Str("partner_id", partnerIDStr).
		Str("status", status).
		Msg("Coupons exported successfully")

	return c.SendString(content)
}

// @Summary Advanced coupon export
// @Description Exports coupons in various formats (TXT, CSV, XLSX) with customizable options
// @Tags coupons
// @Accept json
// @Produce application/octet-stream
// @Param request body ExportOptionsRequest true "Export options"
// @Success 200 {string} string "Export file"
// @Failure 400 {object} map[string]any "Invalid request"
// @Failure 500 {object} map[string]any "Internal server error"
// @Router /coupons/export-advanced [get]
func (handler *CouponHandler) ExportCouponsAdvanced(c *fiber.Ctx) error {
	var req ExportOptionsRequest
	if err := c.BodyParser(&req); err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Invalid request body")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Format == "" {
		req.Format = ExportFormatType("codes")
	}

	content, filename, contentType, err := handler.deps.CouponService.ExportCouponsAdvanced(req)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to export coupons")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to export coupons",
		})
	}

	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Set("Content-Type", contentType)

	handler.deps.Logger.FromContext(c).Info().
		Str("filename", filename).
		Str("format", string(req.Format)).
		Str("content_type", contentType).
		Msg("Advanced export completed successfully")

	return c.Send(content)
}

// @Summary Get coupon statistics
// @Description Returns coupon statistics with optional partner filtering
// @Tags coupons
// @Produce json
// @Param partner_id query string false "Partner ID for filtering"
// @Success 200 {object} map[string]any "Coupon statistics"
// @Failure 500 {object} map[string]any "Internal server error"
// @Router /coupons/statistics [get]
func (handler *CouponHandler) GetStatistics(c *fiber.Ctx) error {
	partnerIDStr := c.Query("partner_id")

	var partnerID *uuid.UUID
	if partnerIDStr != "" {
		if id, err := uuid.Parse(partnerIDStr); err == nil {
			partnerID = &id
		}
	}

	stats, err := handler.deps.CouponService.GetStatistics(partnerID)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to get statistics")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get statistics",
		})
	}

	handler.deps.Logger.FromContext(c).Info().
		Str("partner_id", partnerIDStr).
		Msg("Statistics retrieved successfully")

	return c.JSON(stats)
}

// @Summary Get coupons by partner
// @Description Returns all coupons for specific partner with filtering options
// @Tags coupons
// @Produce json
// @Param partner_id path string true "Partner ID"
// @Param status query string false "Coupon status"
// @Param size query string false "Coupon size"
// @Param style query string false "Coupon style"
// @Success 200 {array} map[string]any "Partner coupons"
// @Failure 400 {object} map[string]any "Invalid partner ID"
// @Failure 500 {object} map[string]any "Internal server error"
// @Router /coupons/partner/{partner_id} [get]
func (handler *CouponHandler) GetCouponsByPartner(c *fiber.Ctx) error {
	partnerIDStr := c.Params("partner_id")
	partnerID, err := uuid.Parse(partnerIDStr)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Invalid partner ID")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid partner ID",
		})
	}

	status := c.Query("status")
	size := c.Query("size")
	style := c.Query("style")

	coupons, err := handler.deps.CouponService.SearchCouponsByPartner(partnerID, status, size, style)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to get coupons by partner")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get coupons by partner",
		})
	}

	handler.deps.Logger.FromContext(c).Info().
		Str("partner_id", partnerIDStr).
		Int("count", len(coupons)).
		Str("status", status).
		Str("size", size).
		Str("style", style).
		Msg("Partner coupons retrieved successfully")

	return c.JSON(fiber.Map{
		"partner_id": partnerID,
		"coupons":    coupons,
		"count":      len(coupons),
	})
}

// @Summary Get coupon by code
// @Description Returns detailed coupon information by code
// @Tags coupons
// @Produce json
// @Param code path string true "Coupon code"
// @Success 200 {object} map[string]any "Coupon information"
// @Failure 404 {object} map[string]any "Coupon not found"
// @Failure 500 {object} map[string]any "Internal server error"
// @Router /coupons/code/{code} [get]
func (handler *CouponHandler) GetCouponByCode(c *fiber.Ctx) error {
	code := c.Params("code")

	coupon, err := handler.deps.CouponService.GetCouponByCode(code)
	if err != nil {
		if err.Error() == "not found" {
			handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Coupon not found")
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Coupon not found",
			})
		}
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to get coupon by code")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get coupon by code",
		})
	}

	handler.deps.Logger.FromContext(c).Info().
		Str("code", code).
		Msg("Coupon retrieved by code successfully")

	return c.JSON(coupon)
}

// @Summary Validate coupon
// @Description Validates coupon existence and availability for activation
// @Tags coupons
// @Produce json
// @Param code path string true "Coupon code"
// @Success 200 {object} map[string]any "Coupon validation status"
// @Failure 404 {object} map[string]any "Coupon not found"
// @Failure 500 {object} map[string]any "Internal server error"
// @Router /coupons/code/{code}/validate [post]
func (handler *CouponHandler) ValidateCoupon(c *fiber.Ctx) error {
	code := c.Params("code")

	validationResult, err := handler.deps.CouponService.ValidateCoupon(code)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to validate coupon")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to validate coupon",
		})
	}

	if !validationResult.Valid {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Coupon not found")
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Coupon not found",
		})
	}

	handler.deps.Logger.FromContext(c).Info().
		Str("code", code).
		Bool("valid", validationResult.Valid).
		Msg("Coupon validated successfully")

	return c.JSON(validationResult)
}

// @Summary Get coupon by ID
// @Description Returns detailed coupon information by ID
// @Tags coupons
// @Produce json
// @Param id path string true "Coupon ID"
// @Success 200 {object} map[string]any "Coupon information"
// @Failure 400 {object} map[string]any "Invalid coupon ID"
// @Failure 404 {object} map[string]any "Coupon not found"
// @Failure 500 {object} map[string]any "Internal server error"
// @Router /coupons/{id} [get]
func (handler *CouponHandler) GetCouponByID(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Invalid coupon ID")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid coupon ID",
		})
	}

	coupon, err := handler.deps.CouponService.GetCouponByID(id)
	if err != nil {
		if err.Error() == "not found" {
			handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Coupon not found")
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Coupon not found",
			})
		}
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to get coupon by ID")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get coupon by ID",
		})
	}

	handler.deps.Logger.FromContext(c).Info().
		Str("coupon_id", idStr).
		Msg("Coupon retrieved by ID successfully")

	return c.JSON(coupon)
}

// @Summary Activate coupon
// @Description Activates coupon by changing status to 'used' and adding image links
// @Tags coupons
// @Accept json
// @Produce json
// @Param id path string true "Coupon ID"
// @Param request body ActivateCouponRequest true "Image links"
// @Success 200 {object} map[string]any "Coupon activated"
// @Failure 400 {object} map[string]any "Invalid coupon ID or request body"
// @Failure 500 {object} map[string]any "Internal server error"
// @Router /coupons/{id}/activate [put]
func (handler *CouponHandler) ActivateCoupon(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Invalid coupon ID")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid coupon ID",
		})
	}

	var req ActivateCouponRequest
	if err := c.BodyParser(&req); err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to parse request body")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to parse request body",
		})
	}

	if err := handler.deps.CouponService.ActivateCoupon(id, req); err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to activate coupon")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to activate coupon",
		})
	}

	logger := handler.deps.Logger.FromContext(c).Info().
		Str("coupon_id", idStr)

	if req.ZipURL != nil {
		logger = logger.Str("zip_url", *req.ZipURL)
	}
	if req.PreviewImageURL != nil {
		logger = logger.Str("preview_image_url", *req.PreviewImageURL)
	}
	if req.SelectedPreviewID != nil {
		logger = logger.Str("selected_preview_id", *req.SelectedPreviewID)
	}
	if req.UserEmail != nil {
		logger = logger.Str("user_email", *req.UserEmail)
	}

	logger.Msg("Coupon activated successfully")

	return c.JSON(fiber.Map{"message": "Coupon activated successfully"})
}

// @Summary Reset coupon
// @Description Resets coupon to initial state (status 'new')
// @Tags coupons
// @Produce json
// @Param id path string true "Coupon ID"
// @Success 200 {object} map[string]any "Coupon reset"
// @Failure 400 {object} map[string]any "Invalid coupon ID"
// @Failure 500 {object} map[string]any "Internal server error"
// @Router /coupons/{id}/reset [put]
func (handler *CouponHandler) ResetCoupon(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Invalid coupon ID")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid coupon ID",
		})
	}

	if err := handler.deps.CouponService.ResetCoupon(id); err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to reset coupon")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to reset coupon",
		})
	}

	handler.deps.Logger.FromContext(c).Info().
		Str("coupon_id", idStr).
		Msg("Coupon reset successfully")

	return c.JSON(fiber.Map{"message": "Coupon reset successfully"})
}

// @Summary Send schema to email
// @Description Sends coupon schema to specified email address
// @Tags coupons
// @Accept json
// @Produce json
// @Param id path string true "Coupon ID"
// @Param request body SendSchemaRequest true "Email for sending"
// @Success 200 {object} map[string]any "Schema sent"
// @Failure 400 {object} map[string]any "Invalid coupon ID or request body"
// @Failure 500 {object} map[string]any "Internal server error"
// @Router /coupons/{id}/send-schema [put]
func (handler *CouponHandler) SendSchema(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Invalid coupon ID")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid coupon ID",
		})
	}

	var req SendSchemaRequest
	if err := c.BodyParser(&req); err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to parse request body")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to parse request body",
		})
	}

	if err := handler.deps.CouponService.SendSchema(id, req.Email); err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to send schema")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to send schema",
		})
	}

	handler.deps.Logger.FromContext(c).Info().
		Str("coupon_id", idStr).
		Str("email", req.Email).
		Msg("Schema sent successfully")

	return c.JSON(fiber.Map{"message": "Schema sent successfully"})
}

// @Summary Mark coupon as purchased
// @Description Marks coupon as purchased online with buyer's email
// @Tags coupons
// @Accept json
// @Produce json
// @Param id path string true "Coupon ID"
// @Param request body MarkAsPurchasedRequest true "Buyer's email"
// @Success 200 {object} map[string]any "Coupon marked as purchased"
// @Failure 400 {object} map[string]any "Invalid coupon ID or request body"
// @Failure 500 {object} map[string]any "Internal server error"
// @Router /coupons/{id}/purchase [put]
func (handler *CouponHandler) MarkAsPurchased(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Invalid coupon ID")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid coupon ID",
		})
	}

	var req MarkAsPurchasedRequest
	if err := c.BodyParser(&req); err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to parse request body")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to parse request body",
		})
	}

	if err := handler.deps.CouponService.MarkAsPurchased(id, req.PurchaseEmail); err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to mark as purchased")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to mark as purchased",
		})
	}

	handler.deps.Logger.FromContext(c).Info().
		Str("coupon_id", idStr).
		Str("purchase_email", req.PurchaseEmail).
		Msg("Coupon marked as purchased successfully")

	return c.JSON(fiber.Map{"message": "Coupon marked as purchased successfully"})
}

// @Summary Download coupon materials
// @Description Downloads archive with redeemed coupon materials (original, preview, schema)
// @Tags coupons
// @Produce application/zip
// @Param id path string true "Coupon ID"
// @Success 200 {string} string "ZIP archive with materials"
// @Failure 400 {object} map[string]any "Invalid coupon ID"
// @Failure 404 {object} map[string]any "Coupon not found or not used"
// @Failure 500 {object} map[string]any "Internal server error"
// @Router /coupons/{id}/download-materials [get]
func (handler *CouponHandler) DownloadMaterials(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Invalid coupon ID")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid coupon ID",
		})
	}

	archiveData, filename, err := handler.deps.CouponService.DownloadMaterials(id)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to download materials")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to download materials",
		})
	}

	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Set("Content-Type", "application/zip")

	handler.deps.Logger.FromContext(c).Info().
		Str("coupon_id", idStr).
		Str("filename", filename).
		Msg("Materials downloaded successfully")

	return c.Send(archiveData)
}

// @Summary Batch reset coupons
// @Description Resets multiple coupons to initial state
// @Tags coupons
// @Accept json
// @Produce json
// @Param request body BatchResetRequest true "List of coupon IDs to reset"
// @Success 200 {object} BatchResetResponse "Batch reset result"
// @Failure 400 {object} map[string]any "Invalid request"
// @Failure 500 {object} map[string]any "Internal server error"
// @Router /coupons/batch/reset [post]
func (handler *CouponHandler) BatchResetCoupons(c *fiber.Ctx) error {
	var req BatchResetRequest
	if err := c.BodyParser(&req); err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Invalid request body")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if len(req.CouponIDs) == 0 {
		handler.deps.Logger.FromContext(c).Error().Msg("Coupon ID required")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Coupon ID required",
		})
	}

	if len(req.CouponIDs) > 1000 {
		handler.deps.Logger.FromContext(c).Error().Msg("Too many items")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Too many items",
		})
	}

	response, err := handler.deps.CouponService.BatchResetCoupons(req.CouponIDs)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to batch reset coupons")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to batch reset coupons",
		})
	}

	handler.deps.Logger.FromContext(c).Info().
		Int("total_requested", len(req.CouponIDs)).
		Int("succsess_count", response.SuccessCount).
		Int("failed_count", len(response.Errors)).
		Msg("Batch reset completed successfully")

	return c.JSON(response)
}

// @Summary Batch delete preview
// @Description Shows information about coupons before deletion and generates confirmation key
// @Tags coupons
// @Accept json
// @Produce json
// @Param request body BatchDeleteRequest true "List of coupon IDs for deletion"
// @Success 200 {object} BatchDeletePreviewResponse "Delete preview"
// @Failure 400 {object} map[string]any "Invalid request"
// @Failure 500 {object} map[string]any "Internal server error"
// @Router /coupons/batch/delete/preview [post]
func (handler *CouponHandler) PreviewBatchDelete(c *fiber.Ctx) error {
	var req BatchDeleteRequest
	if err := c.BodyParser(&req); err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Invalid request body")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if len(req.CouponIDs) == 0 {
		handler.deps.Logger.FromContext(c).Error().Msg("Coupon ID required")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Coupon ID required",
		})
	}

	if len(req.CouponIDs) > 1000 {
		handler.deps.Logger.FromContext(c).Error().Msg("Too many items")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Too many items",
		})
	}

	response, err := handler.deps.CouponService.PreviewBatchDelete(req.CouponIDs)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to get batch delete preview")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get batch delete preview",
		})
	}

	handler.deps.Logger.FromContext(c).Info().
		Int("total_requested", len(req.CouponIDs)).
		Int("found_count", len(response.Coupons)).
		Msg("Batch delete preview generated successfully")

	return c.JSON(response)
}

// @Summary Confirmed batch delete
// @Description Deletes coupons after confirmation with key
// @Tags coupons
// @Accept json
// @Produce json
// @Param request body BatchDeleteConfirmRequest true "Delete confirmation with key"
// @Success 200 {object} BatchDeleteResponse "Delete result"
// @Failure 400 {object} map[string]any "Invalid request"
// @Failure 500 {object} map[string]any "Internal server error"
// @Router /coupons/batch/delete/confirm [post]
func (handler *CouponHandler) ExecuteBatchDelete(c *fiber.Ctx) error {
	var req BatchDeleteConfirmRequest
	if err := c.BodyParser(&req); err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Invalid request body")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if len(req.CouponIDs) == 0 {
		handler.deps.Logger.FromContext(c).Error().Msg("Coupon ID required")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Coupon ID required",
		})
	}

	if req.ConfirmationKey == "" {
		handler.deps.Logger.FromContext(c).Error().Msg("Confirmation key required")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Confirmation key required",
		})
	}

	response, err := handler.deps.CouponService.ExecuteBatchDelete(req)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to execute batch delete")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to execute batch delete",
		})
	}

	handler.deps.Logger.FromContext(c).Info().
		Int("total_requested", len(req.CouponIDs)).
		Int("deleted_count", response.DeletedCount).
		Int("failed_count", len(response.Errors)).
		Msg("Batch delete executed successfully")

	return c.JSON(response)
}
