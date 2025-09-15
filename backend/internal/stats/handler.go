package stats

import (
	"context"
	"strconv"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/skr1ms/mosaic/pkg/jwt"
	"github.com/skr1ms/mosaic/pkg/middleware"
)

type StatsHandlerDeps struct {
	StatsService StatsServiceInterface
	JwtService   JWTServiceInterface
	Logger       *middleware.Logger
}

type StatsHandler struct {
	fiber.Router
	deps *StatsHandlerDeps
}

func NewStatsHandler(router fiber.Router, deps *StatsHandlerDeps) {
	handler := &StatsHandler{
		Router: router,
		deps:   deps,
	}

	jwtConcrete, ok := deps.JwtService.(*jwt.JWT)
	if !ok {
		panic("JwtService must be *jwt.JWT for middleware")
	}

	// ================================================================
	// ADMIN STATISTICS ROUTES: /api/admin/stats/*
	// Access: admin and main_admin roles only
	// ================================================================
	adminStats := router.Group("/admin/stats")
	adminStats.Use(middleware.JWTMiddleware(jwtConcrete, deps.Logger), middleware.AdminOrMainAdmin())

	adminStats.Get("/general", handler.GetGeneralStats)
	adminStats.Get("/partners", handler.GetAllPartnersStats)
	adminStats.Get("/partners/:partner_id", handler.GetPartnerStats)
	adminStats.Get("/time-series", handler.GetTimeSeriesStats)
	adminStats.Get("/system-health", handler.GetSystemHealth)
	adminStats.Get("/coupons-by-status", handler.GetCouponsByStatus)
	adminStats.Get("/coupons-by-size", handler.GetCouponsBySizes)
	adminStats.Get("/coupons-by-style", handler.GetCouponsByStyles)
	adminStats.Get("/top-partners", handler.GetTopPartners)
	adminStats.Get("/realtime", websocket.New(handler.HandleRealTimeStats))

	// ================================================================
	// PARTNER STATISTICS ROUTES: /api/partner/stats/*
	// Access: authenticated partners only
	// ================================================================
	partnerStats := router.Group("/partner/stats")
	partnerStats.Use(middleware.JWTMiddleware(jwtConcrete, deps.Logger), middleware.PartnerOnly())

	partnerStats.Get("/my", handler.GetMyPartnerStats)
	partnerStats.Get("/my/coupons-by-status", handler.GetMyPartnerCouponsByStatus)
	partnerStats.Get("/my/time-series", handler.GetMyPartnerTimeSeriesStats)
}

// @Summary Get general statistics
// @Description Returns general system statistics including coupon count, partner count, and activation percentage
// @Tags admin-stats
// @Produce json
// @Security BearerAuth
// @Success 200 {object} GeneralStatsResponse "General statistics data"
// @Failure 401 {object} map[string]any "Unauthorized access"
// @Failure 500 {object} map[string]any "Internal server error while getting general statistics"
// @Router /admin/stats/general [get]
func (h *StatsHandler) GetGeneralStats(c *fiber.Ctx) error {
	stats, err := h.deps.StatsService.GetGeneralStats(c.Context())
	if err != nil {
		h.deps.Logger.FromContext(c).Error().
			Err(err).
			Str("handler", "GetGeneralStats").
			Msg("Failed to get general statistics")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get general statistics",
		})
	}

	h.deps.Logger.FromContext(c).Info().
		Str("handler", "GetGeneralStats").
		Msg("General statistics retrieved")

	return c.JSON(stats)
}

// @Summary Get all partners statistics
// @Description Returns statistics for all partners in the system
// @Tags admin-stats
// @Produce json
// @Security BearerAuth
// @Success 200 {object} PartnerListStatsResponse "Partners statistics data"
// @Failure 401 {object} map[string]any "Unauthorized access"
// @Failure 500 {object} map[string]any "Internal server error while getting all partners statistics"
// @Router /admin/stats/partners [get]
func (h *StatsHandler) GetAllPartnersStats(c *fiber.Ctx) error {
	stats, err := h.deps.StatsService.GetAllPartnersStats(c.Context())
	if err != nil {
		h.deps.Logger.FromContext(c).Error().
			Err(err).
			Str("handler", "GetAllPartnersStats").
			Msg("Failed to get partners statistics")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get partners statistics",
		})
	}

	h.deps.Logger.FromContext(c).Info().
		Str("handler", "GetAllPartnersStats").
		Msg("All partners statistics retrieved")

	return c.JSON(stats)
}

// @Summary Get partner statistics
// @Description Returns detailed statistics for a specific partner
// @Tags admin-stats
// @Produce json
// @Security BearerAuth
// @Param partner_id path string true "Partner ID" format(uuid)
// @Success 200 {object} PartnerStatsResponse "Partner statistics data"
// @Failure 400 {object} map[string]any "Invalid partner ID format"
// @Failure 401 {object} map[string]any "Unauthorized access"
// @Failure 404 {object} map[string]any "Partner not found"
// @Failure 500 {object} map[string]any "Internal server error while getting partner statistics"
// @Router /admin/stats/partners/{partner_id} [get]
func (h *StatsHandler) GetPartnerStats(c *fiber.Ctx) error {
	partnerIDStr := c.Params("partner_id")
	partnerID, err := uuid.Parse(partnerIDStr)
	if err != nil {
		h.deps.Logger.FromContext(c).Warn().
			Str("handler", "GetPartnerStats").
			Str("partner_id", partnerIDStr).
			Msg("Invalid partner ID format")

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid partner ID format",
		})
	}

	stats, err := h.deps.StatsService.GetPartnerStats(c.Context(), partnerID)
	if err != nil {
		h.deps.Logger.FromContext(c).Error().
			Err(err).
			Str("handler", "GetPartnerStats").
			Interface("partner_id", partnerID).
			Msg("Failed to get partner statistics")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get partner statistics",
		})
	}

	h.deps.Logger.FromContext(c).Info().
		Str("handler", "GetPartnerStats").
		Interface("partner_id", partnerID).
		Msg("Partner statistics retrieved")

	return c.JSON(stats)
}

// @Summary Get time series statistics
// @Description Returns data for building time-based charts and graphs
// @Tags admin-stats
// @Produce json
// @Security BearerAuth
// @Param partner_id query string false "Partner ID (optional)" format(uuid)
// @Param date_from query string false "Start date (YYYY-MM-DD)" format(date)
// @Param date_to query string false "End date (YYYY-MM-DD)" format(date)
// @Param period query string false "Grouping period" Enums(day, week, month, year)
// @Success 200 {object} TimeSeriesStatsResponse "Time series statistics data"
// @Failure 400 {object} map[string]any "Invalid request parameters format"
// @Failure 401 {object} map[string]any "Unauthorized access"
// @Failure 500 {object} map[string]any "Internal server error while getting time series statistics"
// @Router /admin/stats/time-series [get]
func (h *StatsHandler) GetTimeSeriesStats(c *fiber.Ctx) error {
	filters := &StatsFiltersRequest{}

	if partnerIDStr := c.Query("partner_id"); partnerIDStr != "" {
		partnerID, err := uuid.Parse(partnerIDStr)
		if err != nil {
			h.deps.Logger.FromContext(c).Warn().
				Str("handler", "GetTimeSeriesStats").
				Str("partner_id", partnerIDStr).
				Msg("Invalid partner ID format")

			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid partner ID format",
			})
		}
		filters.PartnerID = &partnerID
	}

	if dateFrom := c.Query("date_from"); dateFrom != "" {
		filters.DateFrom = &dateFrom
	}

	if dateTo := c.Query("date_to"); dateTo != "" {
		filters.DateTo = &dateTo
	}

	if period := c.Query("period"); period != "" {
		filters.Period = &period
	}

	stats, err := h.deps.StatsService.GetTimeSeriesStats(c.Context(), filters)
	if err != nil {
		h.deps.Logger.FromContext(c).Error().
			Err(err).
			Str("handler", "GetTimeSeriesStats").
			Interface("filters", filters).
			Msg("Failed to get time series statistics")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get time series statistics",
		})
	}

	h.deps.Logger.FromContext(c).Info().
		Str("handler", "GetTimeSeriesStats").
		Interface("filters", filters).
		Msg("Time series statistics retrieved")

	return c.JSON(stats)
}

// @Summary Get system health status
// @Description Returns system health information, service status and performance metrics
// @Tags admin-stats
// @Produce json
// @Security BearerAuth
// @Success 200 {object} SystemHealthResponse "System health data"
// @Failure 401 {object} map[string]any "Unauthorized access"
// @Failure 500 {object} map[string]any "Internal server error while checking system health"
// @Router /admin/stats/system-health [get]
func (h *StatsHandler) GetSystemHealth(c *fiber.Ctx) error {
	health, err := h.deps.StatsService.GetSystemHealth(c.Context())
	if err != nil {
		h.deps.Logger.FromContext(c).Error().
			Err(err).
			Str("handler", "GetSystemHealth").
			Msg("Failed to get system health")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get system health",
		})
	}

	h.deps.Logger.FromContext(c).Info().
		Str("handler", "GetSystemHealth").
		Msg("System health status retrieved")

	return c.JSON(health)
}

// @Summary Get coupons statistics by status
// @Description Returns count of coupons in each status
// @Tags admin-stats
// @Produce json
// @Security BearerAuth
// @Param partner_id query string false "Partner ID (optional)" format(uuid)
// @Success 200 {object} CouponsByStatusResponse "Coupons by status statistics"
// @Failure 400 {object} map[string]any "Invalid partner ID format"
// @Failure 401 {object} map[string]any "Unauthorized access"
// @Failure 500 {object} map[string]any "Internal server error while getting coupons by status statistics"
// @Router /admin/stats/coupons-by-status [get]
func (h *StatsHandler) GetCouponsByStatus(c *fiber.Ctx) error {
	var partnerID *uuid.UUID
	if partnerIDStr := c.Query("partner_id"); partnerIDStr != "" {
		parsed, err := uuid.Parse(partnerIDStr)
		if err != nil {
			h.deps.Logger.FromContext(c).Warn().
				Str("handler", "GetCouponsByStatus").
				Str("partner_id", partnerIDStr).
				Msg("Invalid partner ID format")

			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid partner ID format",
			})
		}
		partnerID = &parsed
	}

	stats, err := h.deps.StatsService.GetCouponsByStatus(c.Context(), partnerID)
	if err != nil {
		h.deps.Logger.FromContext(c).Error().
			Err(err).
			Str("handler", "GetCouponsByStatus").
			Interface("partner_id", partnerID).
			Msg("Failed to get coupons by status statistics")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get coupons by status statistics",
		})
	}

	h.deps.Logger.FromContext(c).Info().
		Str("handler", "GetCouponsByStatus").
		Interface("partner_id", partnerID).
		Msg("Coupons by status statistics retrieved")

	return c.JSON(stats)
}

// @Summary Get coupons statistics by size
// @Description Returns count of coupons by their sizes
// @Tags admin-stats
// @Produce json
// @Security BearerAuth
// @Param partner_id query string false "Partner ID (optional)" format(uuid)
// @Success 200 {object} CouponsBySizeResponse "Coupons by size statistics"
// @Failure 400 {object} map[string]any "Invalid partner ID format"
// @Failure 401 {object} map[string]any "Unauthorized access"
// @Failure 500 {object} map[string]any "Internal server error while getting coupons by size statistics"
// @Router /admin/stats/coupons-by-size [get]
func (h *StatsHandler) GetCouponsBySizes(c *fiber.Ctx) error {
	var partnerID *uuid.UUID
	if partnerIDStr := c.Query("partner_id"); partnerIDStr != "" {
		parsed, err := uuid.Parse(partnerIDStr)
		if err != nil {
			h.deps.Logger.FromContext(c).Warn().
				Str("handler", "GetCouponsBySizes").
				Str("partner_id", partnerIDStr).
				Msg("Invalid partner ID format")

			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid partner ID format",
			})
		}
		partnerID = &parsed
	}

	stats, err := h.deps.StatsService.GetCouponsBySize(c.Context(), partnerID)
	if err != nil {
		h.deps.Logger.FromContext(c).Error().
			Err(err).
			Str("handler", "GetCouponsBySizes").
			Interface("partner_id", partnerID).
			Msg("Failed to get coupons by size statistics")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get coupons by size statistics",
		})
	}

	h.deps.Logger.FromContext(c).Info().
		Str("handler", "GetCouponsBySizes").
		Interface("partner_id", partnerID).
		Msg("Coupons by size statistics retrieved")

	return c.JSON(stats)
}

// @Summary Get coupons statistics by style
// @Description Returns count of coupons by their processing styles
// @Tags admin-stats
// @Produce json
// @Security BearerAuth
// @Param partner_id query string false "Partner ID (optional)" format(uuid)
// @Success 200 {object} CouponsByStyleResponse "Coupons by style statistics"
// @Failure 400 {object} map[string]any "Invalid partner ID format"
// @Failure 401 {object} map[string]any "Unauthorized access"
// @Failure 500 {object} map[string]any "Internal server error while getting coupons by style statistics"
// @Router /admin/stats/coupons-by-style [get]
func (h *StatsHandler) GetCouponsByStyles(c *fiber.Ctx) error {
	var partnerID *uuid.UUID
	if partnerIDStr := c.Query("partner_id"); partnerIDStr != "" {
		parsed, err := uuid.Parse(partnerIDStr)
		if err != nil {
			h.deps.Logger.FromContext(c).Warn().
				Str("handler", "GetCouponsByStyles").
				Str("partner_id", partnerIDStr).
				Msg("Invalid partner ID format")

			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid partner ID format",
			})
		}
		partnerID = &parsed
	}

	stats, err := h.deps.StatsService.GetCouponsByStyle(c.Context(), partnerID)
	if err != nil {
		h.deps.Logger.FromContext(c).Error().
			Err(err).
			Str("handler", "GetCouponsByStyles").
			Interface("partner_id", partnerID).
			Msg("Failed to get coupons by style statistics")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get coupons by style statistics",
		})
	}

	h.deps.Logger.FromContext(c).Info().
		Str("handler", "GetCouponsByStyles").
		Interface("partner_id", partnerID).
		Msg("Coupons by style statistics retrieved")

	return c.JSON(stats)
}

// @Summary Get top partners by activity
// @Description Returns list of top partners by activity metrics
// @Tags admin-stats
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Number of partners to return (default: 10)" minimum(1) maximum(100)
// @Success 200 {object} TopPartnersResponse "Top partners statistics"
// @Failure 400 {object} map[string]any "Invalid limit parameter"
// @Failure 401 {object} map[string]any "Unauthorized access"
// @Failure 500 {object} map[string]any "Internal server error while getting top partners"
// @Router /admin/stats/top-partners [get]
func (h *StatsHandler) GetTopPartners(c *fiber.Ctx) error {
	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		parsed, err := strconv.Atoi(limitStr)
		if err != nil {
			h.deps.Logger.FromContext(c).Warn().
				Str("handler", "GetTopPartners").
				Str("limit", limitStr).
				Msg("Invalid limit parameter format")

			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid limit parameter format",
			})
		}
		if parsed <= 0 {
			h.deps.Logger.FromContext(c).Warn().
				Str("handler", "GetTopPartners").
				Int("limit", parsed).
				Msg("Limit must be greater than 0")

			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Limit must be greater than 0",
			})
		}
		limit = parsed
	}

	stats, err := h.deps.StatsService.GetTopPartners(c.Context(), limit)
	if err != nil {
		h.deps.Logger.FromContext(c).Error().
			Err(err).
			Str("handler", "GetTopPartners").
			Int("limit", limit).
			Msg("Failed to get top partners")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get top partners",
		})
	}

	h.deps.Logger.FromContext(c).Info().
		Str("handler", "GetTopPartners").
		Int("limit", limit).
		Msg("Top partners statistics retrieved")

	return c.JSON(stats)
}

// @Summary Get real-time statistics via WebSocket
// @Description Provides real-time statistics updates via WebSocket connection
// @Tags admin-stats
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 101 {string} string "Switching to WebSocket protocol"
// @Failure 400 {object} map[string]any "WebSocket upgrade failed"
// @Failure 401 {object} map[string]any "Unauthorized access"
// @Router /admin/stats/realtime [get]
func (h *StatsHandler) HandleRealTimeStats(c *websocket.Conn) {
	// WebSocket doesn't have direct access to fiber.Ctx
	// Use common logger but without request_id
	logger := h.deps.Logger.GetZerologLogger().With().
		Str("handler", "HandleRealTimeStats").
		Logger()

	defer c.Close()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		stats, err := h.deps.StatsService.GetRealTimeStats(context.Background())
		if err != nil {
			logger.Error().Err(err).Msg("failed to get real-time stats")
			continue
		}

		if err := c.WriteJSON(stats); err != nil {
			logger.Error().Err(err).Msg("failed to send real-time stats via WebSocket")
			return
		}
	}
}

// @Summary Get my partner statistics
// @Description Returns statistics for the current authenticated partner
// @Tags partner-stats
// @Produce json
// @Security BearerAuth
// @Success 200 {object} PartnerStatsResponse "Partner statistics data"
// @Failure 401 {object} map[string]any "Partner not authenticated"
// @Failure 500 {object} map[string]any "Internal server error while getting partner statistics"
// @Router /partner/stats/my [get]
func (h *StatsHandler) GetMyPartnerStats(c *fiber.Ctx) error {
	partnerID := c.Locals("partner_id")
	if partnerID == nil {
		h.deps.Logger.FromContext(c).Warn().
			Str("handler", "GetMyPartnerStats").
			Msg("Partner not authenticated")

		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Partner not authenticated",
		})
	}

	uuidPartnerID, ok := partnerID.(uuid.UUID)
	if !ok {
		h.deps.Logger.FromContext(c).Error().
			Str("handler", "GetMyPartnerStats").
			Interface("partner_id_type", partnerID).
			Msg("Invalid partner ID type")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Invalid partner ID type",
		})
	}

	stats, err := h.deps.StatsService.GetPartnerStats(c.Context(), uuidPartnerID)
	if err != nil {
		h.deps.Logger.FromContext(c).Error().
			Err(err).
			Str("handler", "GetMyPartnerStats").
			Interface("partner_id", uuidPartnerID).
			Msg("Failed to get partner statistics")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get partner statistics",
		})
	}

	h.deps.Logger.FromContext(c).Info().
		Str("handler", "GetMyPartnerStats").
		Interface("partner_id", uuidPartnerID).
		Msg("My partner statistics retrieved")

	return c.JSON(stats)
}

// @Summary Get my partner coupons statistics by status
// @Description Returns coupons statistics by status for the current authenticated partner
// @Tags partner-stats
// @Produce json
// @Security BearerAuth
// @Success 200 {object} CouponsByStatusResponse "Partner coupons by status statistics"
// @Failure 401 {object} map[string]any "Partner not authenticated"
// @Failure 500 {object} map[string]any "Internal server error while getting coupons by status statistics"
// @Router /partner/stats/my/coupons-by-status [get]
func (h *StatsHandler) GetMyPartnerCouponsByStatus(c *fiber.Ctx) error {
	partnerID := c.Locals("partner_id")
	if partnerID == nil {
		h.deps.Logger.FromContext(c).Warn().
			Str("handler", "GetMyPartnerCouponsByStatus").
			Msg("Partner not authenticated")

		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Partner not authenticated",
		})
	}

	uuidPartnerID, ok := partnerID.(uuid.UUID)
	if !ok {
		h.deps.Logger.FromContext(c).Error().
			Str("handler", "GetMyPartnerCouponsByStatus").
			Interface("partner_id_type", partnerID).
			Msg("Invalid partner ID type")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Invalid partner ID type",
		})
	}

	stats, err := h.deps.StatsService.GetCouponsByStatus(c.Context(), &uuidPartnerID)
	if err != nil {
		h.deps.Logger.FromContext(c).Error().
			Err(err).
			Str("handler", "GetMyPartnerCouponsByStatus").
			Interface("partner_id", uuidPartnerID).
			Msg("Failed to get coupons by status statistics")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get coupons by status statistics",
		})
	}

	h.deps.Logger.FromContext(c).Info().
		Str("handler", "GetMyPartnerCouponsByStatus").
		Interface("partner_id", uuidPartnerID).
		Msg("My partner coupons by status statistics retrieved")

	return c.JSON(stats)
}

// @Summary Get my partner time series statistics
// @Description Returns time series statistics for the current authenticated partner
// @Tags partner-stats
// @Produce json
// @Security BearerAuth
// @Param date_from query string false "Start date (YYYY-MM-DD)" format(date)
// @Param date_to query string false "End date (YYYY-MM-DD)" format(date)
// @Param period query string false "Grouping period" Enums(day, week, month)
// @Success 200 {object} TimeSeriesStatsResponse "Partner time series statistics"
// @Failure 401 {object} map[string]any "Partner not authenticated"
// @Failure 500 {object} map[string]any "Internal server error while getting time series statistics"
// @Router /partner/stats/my/time-series [get]
func (h *StatsHandler) GetMyPartnerTimeSeriesStats(c *fiber.Ctx) error {
	partnerID := c.Locals("partner_id")
	if partnerID == nil {
		h.deps.Logger.FromContext(c).Warn().
			Str("handler", "GetMyPartnerTimeSeriesStats").
			Msg("Partner not authenticated")

		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Partner not authenticated",
		})
	}

	uuidPartnerID, ok := partnerID.(uuid.UUID)
	if !ok {
		h.deps.Logger.FromContext(c).Error().
			Str("handler", "GetMyPartnerTimeSeriesStats").
			Interface("partner_id_type", partnerID).
			Msg("Invalid partner ID type")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Invalid partner ID type",
		})
	}

	filters := &StatsFiltersRequest{
		PartnerID: &uuidPartnerID,
	}

	if dateFrom := c.Query("date_from"); dateFrom != "" {
		filters.DateFrom = &dateFrom
	}

	if dateTo := c.Query("date_to"); dateTo != "" {
		filters.DateTo = &dateTo
	}

	if period := c.Query("period"); period != "" {
		filters.Period = &period
	}

	stats, err := h.deps.StatsService.GetTimeSeriesStats(c.Context(), filters)
	if err != nil {
		h.deps.Logger.FromContext(c).Error().
			Err(err).
			Str("handler", "GetMyPartnerTimeSeriesStats").
			Interface("filters", filters).
			Msg("Failed to get time series statistics")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get time series statistics",
		})
	}

	h.deps.Logger.FromContext(c).Info().
		Str("handler", "GetMyPartnerTimeSeriesStats").
		Interface("partner_id", uuidPartnerID).
		Interface("filters", filters).
		Msg("My partner time series statistics retrieved")

	return c.JSON(stats)
}
