package partner

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/skr1ms/mosaic/internal/coupon"
	"github.com/skr1ms/mosaic/pkg/jwt"
	"github.com/skr1ms/mosaic/pkg/middleware"
)

type PartnerHandlerDeps struct {
	Config           ConfigInterface
	PartnerService   PartnerServiceInterface
	CouponRepository CouponRepositoryInterface
	JwtService       JWTServiceInterface
	MailSender       MailSenderInterface
	BrandingHelper   BrandingHelperInterface
	Logger           *middleware.Logger
}

type PartnerHandler struct {
	fiber.Router
	deps *PartnerHandlerDeps
}

func NewPartnerHandler(router fiber.Router, deps *PartnerHandlerDeps) {
	handler := &PartnerHandler{
		Router: router,
		deps:   deps,
	}

	// ================================================================
	// PARTNER ROUTES: /api/partner/*
	// Access: authenticated partners only
	// ================================================================
	partnerRoutes := router.Group("/partner")

	jwtConcrete, ok := deps.JwtService.(*jwt.JWT)
	if !ok {
		panic("JwtService must be *jwt.JWT for middleware")
	}

	protected := partnerRoutes.Use(middleware.JWTMiddleware(jwtConcrete, deps.Logger), middleware.PartnerOnly())
	protected.Get("/dashboard", handler.GetDashboard)                                 // GET /api/partner/dashboard
	protected.Get("/profile", handler.GetProfile)                                     // GET /api/partner/profile
	protected.Put("/profile", handler.UpdateProfile)                                  // PUT /api/partner/profile
	protected.Get("/coupons", handler.GetMyCoupons)                                   // GET /api/partner/coupons
	protected.Get("/coupons/export", handler.ExportCoupons)                           // GET /api/partner/coupons/export
	protected.Get("/coupons/:id/download-materials", handler.DownloadCouponMaterials) // GET /api/partner/coupons/:id/download-materials
	protected.Get("/statistics", handler.GetMyStatistics)                             // GET /api/partner/statistics
	protected.Get("/statistics/comparison", handler.GetComparisonStatistics)          // GET /api/partner/statistics/comparison
}

// @Summary Get partner dashboard data
// @Description Returns data for the partner's main dashboard page including coupon statistics and recent activations
// @Tags partner-dashboard
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]any "Dashboard data with coupons and recent activations"
// @Failure 401 {object} map[string]any "Unauthorized: JWT token is missing or invalid"
// @Failure 403 {object} map[string]any "Forbidden: User does not have partner role"
// @Failure 500 {object} map[string]any "Internal server error when retrieving dashboard data"
// @Router /partner/dashboard [get]
func (handler *PartnerHandler) GetDashboard(c *fiber.Ctx) error {
	claims, err := jwt.GetClaimsFromFiberContext(c)
	if err != nil {
		handler.deps.Logger.FromContext(c).Warn().
			Err(err).
			Str("handler", "GetDashboard").
			Msg("Failed to get JWT claims")

		errorResponse := fiber.Map{
			"error":      "Failed to get JWT claims",
			"request_id": c.Get("X-Request-ID"),
		}
		return c.Status(fiber.StatusUnauthorized).JSON(errorResponse)
	}

	stats, err := handler.deps.CouponRepository.GetStatistics(c.UserContext(), &claims.UserID)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().
			Err(err).
			Str("handler", "GetDashboard").
			Str("partner_id", claims.UserID.String()).
			Msg("Failed to get partner coupon statistics")

		errorResponse := fiber.Map{
			"error":      "Failed to get partner coupon statistics",
			"request_id": c.Get("X-Request-ID"),
		}
		if os.Getenv("ENVIRONMENT") == "development" || os.Getenv("ENVIRONMENT") == "dev" {
			errorResponse["details"] = err.Error()
		}
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse)
	}

	recent, err := handler.deps.CouponRepository.GetRecentActivatedByPartner(c.UserContext(), claims.UserID, 10)
	if err != nil {
		recent = []*coupon.Coupon{}
	}

	handler.deps.Logger.FromContext(c).Info().
		Str("handler", "GetDashboard").
		Str("partner_id", claims.UserID.String()).
		Int("recent_count", len(recent)).
		Msg("Partner dashboard retrieved successfully")

	return c.JSON(fiber.Map{
		"message":            "Partner dashboard",
		"coupons":            stats,
		"recent_activations": recent,
	})
}

// @Summary Get partner profile
// @Description Returns information about the current partner's profile
// @Tags partner-profile
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]any "Partner profile data"
// @Failure 401 {object} map[string]any "Unauthorized: JWT token is missing or invalid"
// @Failure 403 {object} map[string]any "Forbidden: User does not have partner role"
// @Failure 404 {object} map[string]any "Partner not found in the database"
// @Router /partner/profile [get]
func (handler *PartnerHandler) GetProfile(c *fiber.Ctx) error {
	claims, err := jwt.GetClaimsFromFiberContext(c)
	if err != nil {
		handler.deps.Logger.FromContext(c).Warn().
			Err(err).
			Str("handler", "GetProfile").
			Msg("Failed to get JWT claims")

		errorResponse := fiber.Map{
			"error":      "Failed to get JWT claims",
			"request_id": c.Get("X-Request-ID"),
		}
		return c.Status(fiber.StatusUnauthorized).JSON(errorResponse)
	}

	partner, err := handler.deps.PartnerService.GetPartnerRepository().GetByID(context.Background(), claims.UserID)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().
			Err(err).
			Str("handler", "GetProfile").
			Str("partner_id", claims.UserID.String()).
			Msg("Partner not found")

		errorResponse := fiber.Map{
			"error":      "Partner not found",
			"request_id": c.Get("X-Request-ID"),
		}
		return c.Status(fiber.StatusNotFound).JSON(errorResponse)
	}

	handler.deps.Logger.FromContext(c).Info().
		Str("handler", "GetProfile").
		Str("partner_id", partner.ID.String()).
		Str("brand_name", partner.BrandName).
		Msg("Partner profile retrieved successfully")

	return c.JSON(fiber.Map{
		"id":           partner.ID,
		"login":        partner.Login,
		"partner_code": partner.PartnerCode,
		"brand_name":   partner.BrandName,
		"domain":       partner.Domain,
		"email":        partner.Email,
		"address":      partner.Address,
		"phone":        partner.Phone,
		"telegram":     partner.Telegram,
		"whatsapp":     partner.Whatsapp,
		"brand_colors": partner.BrandColors,
		"allow_sales":  partner.AllowSales,
		"status":       partner.Status,
		"created_at":   partner.CreatedAt,
		"updated_at":   partner.UpdatedAt,
	})
}

// @Summary Update partner profile
// @Description Attempts to update partner profile (read-only in partner panel, available only to administrator)
// @Tags partner-profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Failure 403 {object} map[string]any "Forbidden: Profile is read-only, contact administrator"
// @Router /partner/profile [put]
func (handler *PartnerHandler) UpdateProfile(c *fiber.Ctx) error {
	handler.deps.Logger.FromContext(c).Warn().
		Str("handler", "UpdateProfile").
		Msg("Profile is read-only; contact administrator")

	errorResponse := fiber.Map{
		"error":      "Profile is read-only; contact administrator",
		"request_id": c.Get("X-Request-ID"),
	}
	return c.Status(fiber.StatusForbidden).JSON(errorResponse)
}

// @Summary Get partner's coupons
// @Description Returns a list of coupons for the current partner with filtering and pagination
// @Tags partner-coupons
// @Produce json
// @Security BearerAuth
// @Param code query string false "Coupon code for search"
// @Param status query string false "Coupon status (new, used, completed)"
// @Param size query string false "Coupon size"
// @Param style query string false "Coupon style"
// @Param created_from query string false "Creation date from (RFC3339 format)"
// @Param created_to query string false "Creation date to (RFC3339 format)"
// @Param used_from query string false "Usage date from (RFC3339 format)"
// @Param used_to query string false "Usage date to (RFC3339 format)"
// @Param sort_by query string false "Field to sort by" Enums(created_at,updated_at,used_at) default(created_at)
// @Param sort_order query string false "Sort order" Enums(asc,desc) default(desc)
// @Param page query integer false "Page number" default(1)
// @Param limit query integer false "Number of items per page" default(50) minimum(1) maximum(100)
// @Success 200 {object} map[string]any "List of partner coupons with pagination info"
// @Failure 401 {object} map[string]any "Unauthorized: JWT token is missing or invalid"
// @Failure 403 {object} map[string]any "Forbidden: User does not have partner role"
// @Failure 500 {object} map[string]any "Internal server error when retrieving coupons"
// @Router /partner/coupons [get]
func (handler *PartnerHandler) GetMyCoupons(c *fiber.Ctx) error {
	claims, err := jwt.GetClaimsFromFiberContext(c)
	if err != nil {
		handler.deps.Logger.FromContext(c).Warn().
			Err(err).
			Str("handler", "GetMyCoupons").
			Msg("Failed to get JWT claims")

		errorResponse := fiber.Map{
			"error":      "Failed to get JWT claims",
			"request_id": c.Get("X-Request-ID"),
		}
		return c.Status(fiber.StatusUnauthorized).JSON(errorResponse)
	}

	code := c.Query("code", "")
	status := c.Query("status", "")
	size := c.Query("size", "")
	style := c.Query("style", "")
	sortBy := c.Query("sort_by", "created_at")
	sortOrder := c.Query("sort_order", "desc")
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 50)

	var createdFromPtr, createdToPtr, usedFromPtr, usedToPtr *time.Time
	if v := c.Query("created_from"); v != "" {
		if t, e := time.Parse(time.RFC3339, v); e == nil {
			createdFromPtr = &t
		}
	}
	if v := c.Query("created_to"); v != "" {
		if t, e := time.Parse(time.RFC3339, v); e == nil {
			createdToPtr = &t
		}
	}
	if v := c.Query("used_from"); v != "" {
		if t, e := time.Parse(time.RFC3339, v); e == nil {
			usedFromPtr = &t
		}
	}
	if v := c.Query("used_to"); v != "" {
		if t, e := time.Parse(time.RFC3339, v); e == nil {
			usedToPtr = &t
		}
	}

	coupons, total, err := handler.deps.CouponRepository.SearchPartnerCoupons(c.UserContext(), claims.UserID, code, status, size, style, createdFromPtr, createdToPtr, usedFromPtr, usedToPtr, sortBy, sortOrder, page, limit)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().
			Err(err).
			Str("handler", "GetMyCoupons").
			Str("partner_id", claims.UserID.String()).
			Msg("Failed to get coupons")

		errorResponse := fiber.Map{
			"error":      "Failed to get coupons",
			"request_id": c.Get("X-Request-ID"),
		}
		if os.Getenv("ENVIRONMENT") == "development" || os.Getenv("ENVIRONMENT") == "dev" {
			errorResponse["details"] = err.Error()
		}
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse)
	}

	handler.deps.Logger.FromContext(c).Info().
		Str("handler", "GetMyCoupons").
		Str("partner_id", claims.UserID.String()).
		Int("total", total).
		Int("page", page).
		Int("limit", limit).
		Msg("Partner coupons retrieved successfully")

	return c.JSON(fiber.Map{
		"message":    "Partner coupons retrieved",
		"partner_id": claims.UserID,
		"total":      total,
		"page":       page,
		"limit":      limit,
		"coupons":    coupons,
	})
}

// @Summary Export partner's coupons
// @Description Exports partner's coupons with status "new" in .txt or .csv format
// @Tags partner-coupons
// @Produce text/plain,text/csv
// @Security BearerAuth
// @Param format query string false "File format (txt or csv)" Enums(txt,csv) default(csv)
// @Param status query string false "Coupon status to export (new, used, completed, or 'all' for all statuses)" default(new)
// @Success 200 {string} string "File with coupons as attachment"
// @Failure 400 {object} map[string]any "Bad request: Invalid parameters"
// @Failure 401 {object} map[string]any "Unauthorized: JWT token is missing or invalid"
// @Failure 403 {object} map[string]any "Forbidden: User does not have partner role"
// @Failure 500 {object} map[string]any "Internal server error during export process"
// @Router /partner/coupons/export [get]
func (handler *PartnerHandler) ExportCoupons(c *fiber.Ctx) error {
	claims, err := jwt.GetClaimsFromFiberContext(c)
	if err != nil {
		handler.deps.Logger.FromContext(c).Warn().
			Err(err).
			Str("handler", "ExportCoupons").
			Msg("Failed to get JWT claims")

		errorResponse := fiber.Map{
			"error":      "Failed to get JWT claims",
			"request_id": c.Get("X-Request-ID"),
		}
		return c.Status(fiber.StatusUnauthorized).JSON(errorResponse)
	}

	format := strings.ToLower(c.Query("format", "csv"))
	if format != "txt" && format != "csv" {
		format = "csv"
	}

	status := strings.ToLower(strings.TrimSpace(c.Query("status", "")))
	if status == "all" {
		status = ""
	}

	content, filename, contentType, exportErr := handler.deps.PartnerService.ExportCoupons(claims.UserID, status, format)
	if exportErr != nil {
		handler.deps.Logger.FromContext(c).Error().
			Err(exportErr).
			Str("handler", "ExportCoupons").
			Str("partner_id", claims.UserID.String()).
			Str("format", format).
			Str("status", status).
			Msg("Failed to export coupons")

		errorResponse := fiber.Map{
			"error":      "Failed to export coupons",
			"request_id": c.Get("X-Request-ID"),
		}
		if os.Getenv("ENVIRONMENT") == "development" || os.Getenv("ENVIRONMENT") == "dev" {
			errorResponse["details"] = exportErr.Error()
		}
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse)
	}

	c.Set("Content-Type", contentType)
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Set("Cache-Control", "no-cache")

	handler.deps.Logger.FromContext(c).Info().
		Str("handler", "ExportCoupons").
		Str("partner_id", claims.UserID.String()).
		Str("format", format).
		Str("status", status).
		Str("filename", filename).
		Int("size", len(content)).
		Msg("Coupons exported successfully")

	return c.Send(content)
}

// @Summary Get partner statistics
// @Description Returns general statistics for the current partner including coupon counts by status
// @Tags partner-statistics
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]any "Partner statistics with coupon counts"
// @Failure 401 {object} map[string]any "Unauthorized: JWT token is missing or invalid"
// @Failure 403 {object} map[string]any "Forbidden: User does not have partner role"
// @Failure 500 {object} map[string]any "Internal server error when retrieving statistics"
// @Router /partner/statistics [get]
func (handler *PartnerHandler) GetMyStatistics(c *fiber.Ctx) error {
	claims, err := jwt.GetClaimsFromFiberContext(c)
	if err != nil {
		handler.deps.Logger.FromContext(c).Warn().
			Err(err).
			Str("handler", "GetMyStatistics").
			Msg("Failed to get JWT claims")

		errorResponse := fiber.Map{
			"error":      "Failed to get JWT claims",
			"request_id": c.Get("X-Request-ID"),
		}
		return c.Status(fiber.StatusUnauthorized).JSON(errorResponse)
	}

	stats, err := handler.deps.CouponRepository.GetStatistics(c.UserContext(), &claims.UserID)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().
			Err(err).
			Str("handler", "GetMyStatistics").
			Str("partner_id", claims.UserID.String()).
			Msg("Failed to load partner statistics")

		errorResponse := fiber.Map{
			"error":      "Failed to load partner statistics",
			"request_id": c.Get("X-Request-ID"),
		}
		if os.Getenv("ENVIRONMENT") == "development" || os.Getenv("ENVIRONMENT") == "dev" {
			errorResponse["details"] = err.Error()
		}
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse)
	}

	handler.deps.Logger.FromContext(c).Info().
		Str("handler", "GetMyStatistics").
		Str("partner_id", claims.UserID.String()).
		Interface("statistics", stats).
		Msg("Partner statistics retrieved successfully")

	return c.JSON(fiber.Map{
		"message": "Partner statistics",
		"coupons": stats,
	})
}

// @Summary Get comparison statistics with other partners
// @Description Returns aggregated comparison with other partners - top partners by used and purchased coupons
// @Tags partner-statistics
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]any "Comparison statistics with other partners"
// @Failure 401 {object} map[string]any "Unauthorized: JWT token is missing or invalid"
// @Failure 403 {object} map[string]any "Forbidden: User does not have partner role"
// @Failure 500 {object} map[string]any "Internal server error when retrieving comparison statistics"
// @Router /partner/statistics/comparison [get]
func (handler *PartnerHandler) GetComparisonStatistics(c *fiber.Ctx) error {
	claims, err := jwt.GetClaimsFromFiberContext(c)
	if err != nil {
		handler.deps.Logger.FromContext(c).Warn().
			Err(err).
			Str("handler", "GetComparisonStatistics").
			Msg("Failed to get JWT claims")

		errorResponse := fiber.Map{
			"error":      "Failed to get JWT claims",
			"request_id": c.Get("X-Request-ID"),
		}
		return c.Status(fiber.StatusUnauthorized).JSON(errorResponse)
	}

	result, err := handler.deps.PartnerService.GetComparisonStatistics(c.UserContext(), claims.UserID)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().
			Err(err).
			Str("handler", "GetComparisonStatistics").
			Str("partner_id", claims.UserID.String()).
			Msg("Failed to get comparison statistics")

		errorResponse := fiber.Map{
			"error":      "Failed to get comparison statistics",
			"request_id": c.Get("X-Request-ID"),
		}
		if os.Getenv("ENVIRONMENT") == "development" || os.Getenv("ENVIRONMENT") == "dev" {
			errorResponse["details"] = err.Error()
		}
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse)
	}

	handler.deps.Logger.FromContext(c).Info().
		Str("handler", "GetComparisonStatistics").
		Str("partner_id", claims.UserID.String()).
		Msg("Comparison statistics retrieved successfully")

	return c.JSON(result)
}

// @Summary Download coupon materials
// @Description Downloads archive with materials of a redeemed partner's coupon (original, preview, scheme)
// @Tags partner-coupons
// @Produce application/zip
// @Security BearerAuth
// @Param id path string true "Coupon ID"
// @Success 200 {string} string "ZIP archive with materials as attachment"
// @Failure 400 {object} map[string]any "Bad request: Invalid coupon ID format"
// @Failure 401 {object} map[string]any "Unauthorized: JWT token is missing or invalid"
// @Failure 403 {object} map[string]any "Forbidden: Coupon does not belong to partner"
// @Failure 404 {object} map[string]any "Coupon not found or materials not available"
// @Failure 500 {object} map[string]any "Internal server error during download process"
// @Router /partner/coupons/{id}/download-materials [get]
func (handler *PartnerHandler) DownloadCouponMaterials(c *fiber.Ctx) error {
	claims, err := jwt.GetClaimsFromFiberContext(c)
	if err != nil {
		handler.deps.Logger.FromContext(c).Warn().
			Err(err).
			Str("handler", "DownloadCouponMaterials").
			Msg("Failed to get JWT claims")

		errorResponse := fiber.Map{
			"error":      "Failed to get JWT claims",
			"request_id": c.Get("X-Request-ID"),
		}
		return c.Status(fiber.StatusUnauthorized).JSON(errorResponse)
	}

	idStr := c.Params("id")
	if strings.TrimSpace(idStr) == "" {
		handler.deps.Logger.FromContext(c).Warn().
			Str("handler", "DownloadCouponMaterials").
			Msg("Invalid coupon ID")

		errorResponse := fiber.Map{
			"error":      "Invalid coupon ID",
			"request_id": c.Get("X-Request-ID"),
		}
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse)
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		handler.deps.Logger.FromContext(c).Warn().
			Err(err).
			Str("handler", "DownloadCouponMaterials").
			Str("coupon_id", idStr).
			Msg("Invalid coupon ID format")

		errorResponse := fiber.Map{
			"error":      "Invalid coupon ID format",
			"request_id": c.Get("X-Request-ID"),
		}
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse)
	}

	coupon, err := handler.deps.CouponRepository.GetByID(c.UserContext(), id)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().
			Err(err).
			Str("handler", "DownloadCouponMaterials").
			Str("coupon_id", idStr).
			Msg("Coupon not found")

		errorResponse := fiber.Map{
			"error":      "Coupon not found",
			"request_id": c.Get("X-Request-ID"),
		}
		return c.Status(fiber.StatusNotFound).JSON(errorResponse)
	}

	if coupon.PartnerID != claims.UserID {
		handler.deps.Logger.FromContext(c).Warn().
			Str("handler", "DownloadCouponMaterials").
			Str("coupon_id", idStr).
			Str("partner_id", claims.UserID.String()).
			Msg("Access denied")

		errorResponse := fiber.Map{
			"error":      "Access denied",
			"request_id": c.Get("X-Request-ID"),
		}
		return c.Status(fiber.StatusForbidden).JSON(errorResponse)
	}

	if coupon.Status != "used" && coupon.Status != "completed" {
		handler.deps.Logger.FromContext(c).Warn().
			Str("handler", "DownloadCouponMaterials").
			Str("coupon_id", idStr).
			Str("status", coupon.Status).
			Msg("Coupon must be used or completed to download materials")

		errorResponse := fiber.Map{
			"error":      "Coupon must be used or completed to download materials",
			"request_id": c.Get("X-Request-ID"),
		}
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse)
	}

	if coupon.ZipURL == nil || *coupon.ZipURL == "" {
		handler.deps.Logger.FromContext(c).Warn().
			Str("handler", "DownloadCouponMaterials").
			Str("coupon_id", idStr).
			Msg("Schema materials not available")

		errorResponse := fiber.Map{
			"error":      "Schema materials not available",
			"request_id": c.Get("X-Request-ID"),
		}
		return c.Status(fiber.StatusNotFound).JSON(errorResponse)
	}

	archiveData, filename, downloadErr := handler.deps.PartnerService.DownloadCouponMaterials(coupon.ID)
	if downloadErr != nil {
		handler.deps.Logger.FromContext(c).Error().
			Err(downloadErr).
			Str("handler", "DownloadCouponMaterials").
			Str("coupon_id", idStr).
			Msg("Failed to download materials")

		errorResponse := fiber.Map{
			"error":      "Failed to download materials",
			"request_id": c.Get("X-Request-ID"),
		}
		if os.Getenv("ENVIRONMENT") == "development" || os.Getenv("ENVIRONMENT") == "dev" {
			errorResponse["details"] = downloadErr.Error()
		}
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse)
	}

	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Set("Content-Type", "application/zip")

	handler.deps.Logger.FromContext(c).Info().
		Str("handler", "DownloadCouponMaterials").
		Str("coupon_id", idStr).
		Str("filename", filename).
		Int("size", len(archiveData)).
		Msg("Coupon materials downloaded successfully")

	return c.Send(archiveData)
}
