package stats

import (
	"context"
	"strconv"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/skr1ms/mosaic/pkg/middleware"
	"github.com/skr1ms/mosaic/pkg/utils"
)

type StatsHandlerDeps struct {
	StatsService *StatsService
}

type StatsHandler struct {
	fiber.Router
	deps *StatsHandlerDeps
}

// NewStatsHandler создает новый экземпляр StatsHandler
func NewStatsHandler(router fiber.Router, deps *StatsHandlerDeps) {
	handler := &StatsHandler{
		Router: router,
		deps:   deps,
	}

	// Группа для административных API статистики
	adminStats := router.Group("/admin/stats")
	adminStats.Use(middleware.AdminOnly()) // Требует админ авторизации

	// Основные endpoints статистики
	adminStats.Get("/general", handler.GetGeneralStats)
	adminStats.Get("/partners", handler.GetAllPartnersStats)
	adminStats.Get("/partners/:partner_id", handler.GetPartnerStats)
	adminStats.Get("/time-series", handler.GetTimeSeriesStats)
	adminStats.Get("/system-health", handler.GetSystemHealth)
	adminStats.Get("/coupons-by-status", handler.GetCouponsByStatus)
	adminStats.Get("/coupons-by-size", handler.GetCouponsBySizes)
	adminStats.Get("/coupons-by-style", handler.GetCouponsByStyles)
	adminStats.Get("/top-partners", handler.GetTopPartners)

	// WebSocket для real-time статистики
	adminStats.Get("/realtime", websocket.New(handler.HandleRealTimeStats))

	// Группа для партнерских API статистики
	partnerStats := router.Group("/partner/stats")
	partnerStats.Use(middleware.PartnerOnly()) // Требует партнер авторизации

	// Endpoints для партнеров (ограниченный функционал)
	partnerStats.Get("/my", handler.GetMyPartnerStats)
	partnerStats.Get("/my/coupons-by-status", handler.GetMyPartnerCouponsByStatus)
	partnerStats.Get("/my/time-series", handler.GetMyPartnerTimeSeriesStats)

}

// GetGeneralStats возвращает общую статистику системы
// @Summary Получить общую статистику
// @Description Возвращает общую статистику системы: количество купонов, партнеров, процент активации
// @Tags Admin Stats
// @Accept json
// @Produce json
// @Success 200 {object} GeneralStatsResponse
// @Failure 500 {object} map[string]string
// @Router /admin/stats/general [get]
func (h *StatsHandler) GetGeneralStats(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	stats, err := h.deps.StatsService.GetGeneralStats(c.Context())
	if err != nil {
		log.Error().Err(err).Msg("failed to get general stats")
		return utils.LocalizedError(c, fiber.StatusInternalServerError, "error_internal", "Failed to get general statistics")
	}

	return c.JSON(stats)
}

// GetPartnerStats возвращает статистику по конкретному партнеру
// @Summary Получить статистику партнера
// @Description Возвращает детальную статистику по конкретному партнеру
// @Tags Admin Stats
// @Accept json
// @Produce json
// @Param partner_id path string true "ID партнера"
// @Success 200 {object} PartnerStatsResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/stats/partners/{partner_id} [get]
func (h *StatsHandler) GetPartnerStats(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	partnerIDStr := c.Params("partner_id")
	partnerID, err := uuid.Parse(partnerIDStr)
	if err != nil {
		return utils.LocalizedError(c, fiber.StatusBadRequest, "partner_not_found", "Invalid partner ID")
	}

	stats, err := h.deps.StatsService.GetPartnerStats(c.Context(), partnerID)
	if err != nil {
		log.Error().Err(err).Str("partner_id", partnerID.String()).Msg("failed to get partner stats")
		return utils.LocalizedError(c, fiber.StatusInternalServerError, "error_internal", "Failed to get partner statistics")
	}

	return c.JSON(stats)
}

// GetAllPartnersStats возвращает статистику по всем партнерам
// @Summary Получить статистику всех партнеров
// @Description Возвращает статистику по всем партнерам в системе
// @Tags Admin Stats
// @Accept json
// @Produce json
// @Success 200 {object} PartnerListStatsResponse
// @Failure 500 {object} map[string]string
// @Router /admin/stats/partners [get]
func (h *StatsHandler) GetAllPartnersStats(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	stats, err := h.deps.StatsService.GetAllPartnersStats(c.Context())
	if err != nil {
		log.Error().Err(err).Msg("failed to get all partners stats")
		return utils.LocalizedError(c, fiber.StatusInternalServerError, "error_internal", "Failed to get partners statistics")
	}

	return c.JSON(stats)
}

// GetTimeSeriesStats возвращает временную статистику для графиков
// @Summary Получить временную статистику
// @Description Возвращает данные для построения временных графиков
// @Tags Admin Stats
// @Accept json
// @Produce json
// @Param partner_id query string false "ID партнера (опционально)"
// @Param date_from query string false "Дата начала (YYYY-MM-DD)"
// @Param date_to query string false "Дата окончания (YYYY-MM-DD)"
// @Param period query string false "Период группировки (day, week, month, year)"
// @Success 200 {object} TimeSeriesStatsResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/stats/time-series [get]
func (h *StatsHandler) GetTimeSeriesStats(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	filters := &StatsFiltersRequest{}

	// Парсим query параметры
	if partnerIDStr := c.Query("partner_id"); partnerIDStr != "" {
		partnerID, err := uuid.Parse(partnerIDStr)
		if err != nil {
			return utils.LocalizedError(c, fiber.StatusBadRequest, "partner_not_found", "Invalid partner ID")
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
		log.Error().Err(err).Msg("failed to get time series stats")
		return utils.LocalizedError(c, fiber.StatusInternalServerError, "error_internal", "Failed to get time series statistics")
	}

	return c.JSON(stats)
}

// GetSystemHealth возвращает состояние системы
// @Summary Получить состояние системы
// @Description Возвращает информацию о здоровье системы, статусе сервисов и метриках производительности
// @Tags Admin Stats
// @Accept json
// @Produce json
// @Success 200 {object} SystemHealthResponse
// @Failure 500 {object} map[string]string
// @Router /admin/stats/system-health [get]
func (h *StatsHandler) GetSystemHealth(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	health, err := h.deps.StatsService.GetSystemHealth(c.Context())
	if err != nil {
		log.Error().Err(err).Msg("failed to get system health")
		return utils.LocalizedError(c, fiber.StatusInternalServerError, "error_internal", "Failed to get system health")
	}

	return c.JSON(health)
}

// GetCouponsByStatus возвращает статистику купонов по статусам
// @Summary Получить статистику купонов по статусам
// @Description Возвращает количество купонов в каждом статусе
// @Tags Admin Stats
// @Accept json
// @Produce json
// @Param partner_id query string false "ID партнера (опционально)"
// @Success 200 {object} CouponsByStatusResponse
// @Failure 500 {object} map[string]string
// @Router /admin/stats/coupons-by-status [get]
func (h *StatsHandler) GetCouponsByStatus(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	var partnerID *uuid.UUID
	if partnerIDStr := c.Query("partner_id"); partnerIDStr != "" {
		if parsed, err := uuid.Parse(partnerIDStr); err == nil {
			partnerID = &parsed
		}
	}

	stats, err := h.deps.StatsService.GetCouponsByStatus(c.Context(), partnerID)
	if err != nil {
		log.Error().Err(err).Msg("failed to get coupons by status stats")
		return utils.LocalizedError(c, fiber.StatusInternalServerError, "error_internal", "Failed to get coupons by status statistics")
	}

	return c.JSON(stats)
}

// GetCouponsBySizes возвращает статистику купонов по размерам
// @Summary Получить статистику купонов по размерам
// @Description Возвращает количество купонов по размерам
// @Tags Admin Stats
// @Accept json
// @Produce json
// @Param partner_id query string false "ID партнера (опционально)"
// @Success 200 {object} CouponsBySizeResponse
// @Router /admin/stats/coupons-by-size [get]
func (h *StatsHandler) GetCouponsBySizes(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	var partnerID *uuid.UUID
	if partnerIDStr := c.Query("partner_id"); partnerIDStr != "" {
		if parsed, err := uuid.Parse(partnerIDStr); err == nil {
			partnerID = &parsed
		}
	}

	stats, err := h.deps.StatsService.GetCouponsBySize(c.Context(), partnerID)
	if err != nil {
		log.Error().Err(err).Msg("failed to get coupons by size stats")
		return utils.LocalizedError(c, fiber.StatusInternalServerError, "error_internal", "Failed to get coupons by size statistics")
	}

	return c.JSON(stats)
}

// GetCouponsByStyles возвращает статистику купонов по стилям
// @Summary Получить статистику купонов по стилям
// @Description Возвращает количество купонов по стилям обработки
// @Tags Admin Stats
// @Accept json
// @Produce json
// @Param partner_id query string false "ID партнера (опционально)"
// @Success 200 {object} CouponsByStyleResponse
// @Router /admin/stats/coupons-by-style [get]
func (h *StatsHandler) GetCouponsByStyles(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	var partnerID *uuid.UUID
	if partnerIDStr := c.Query("partner_id"); partnerIDStr != "" {
		if parsed, err := uuid.Parse(partnerIDStr); err == nil {
			partnerID = &parsed
		}
	}

	stats, err := h.deps.StatsService.GetCouponsByStyle(c.Context(), partnerID)
	if err != nil {
		log.Error().Err(err).Msg("failed to get coupons by style stats")
		return utils.LocalizedError(c, fiber.StatusInternalServerError, "error_internal", "Failed to get coupons by style statistics")
	}

	return c.JSON(stats)
}

// GetTopPartners возвращает топ партнеров по активности
// @Summary Получить топ партнеров
// @Description Возвращает список топ партнеров по активности
// @Tags Admin Stats
// @Accept json
// @Produce json
// @Param limit query int false "Количество партнеров (по умолчанию 10)"
// @Success 200 {object} TopPartnersResponse
// @Failure 500 {object} map[string]string
// @Router /admin/stats/top-partners [get]
func (h *StatsHandler) GetTopPartners(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	stats, err := h.deps.StatsService.GetTopPartners(c.Context(), limit)
	if err != nil {
		log.Error().Err(err).Msg("failed to get top partners")
		return utils.LocalizedError(c, fiber.StatusInternalServerError, "error_internal", "Failed to get top partners")
	}

	return c.JSON(stats)
}

// HandleRealTimeStats обработчик WebSocket для real-time статистики
func (h *StatsHandler) HandleRealTimeStats(c *websocket.Conn) {
	log := zerolog.Ctx(context.Background()).With().Str("handler", "HandleRealTimeStats").Logger()
	defer c.Close()

	ticker := time.NewTicker(5 * time.Second) // Обновления каждые 5 секунд
	defer ticker.Stop()

	for range ticker.C {
		stats, err := h.deps.StatsService.GetRealTimeStats(context.Background())
		if err != nil {
			log.Error().Err(err).Msg("failed to get real-time stats")
			continue
		}

		if err := c.WriteJSON(stats); err != nil {
			log.Error().Err(err).Msg("failed to send real-time stats")
			return
		}
	}
}

// Партнерские endpoints (ограниченный функционал)

// GetMyPartnerStats возвращает статистику текущего партнера
// @Summary Получить мою статистику (партнер)
// @Description Возвращает статистику текущего авторизованного партнера
// @Tags Partner Stats
// @Accept json
// @Produce json
// @Success 200 {object} PartnerStatsResponse
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /partner/stats/my [get]
func (h *StatsHandler) GetMyPartnerStats(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	// Получаем ID партнера из middleware (предполагается, что middleware установил его)
	partnerID := c.Locals("partner_id")
	if partnerID == nil {
		return utils.LocalizedError(c, fiber.StatusUnauthorized, "error_unauthorized", "Partner not authenticated")
	}

	stats, err := h.deps.StatsService.GetPartnerStats(c.Context(), partnerID.(uuid.UUID))
	if err != nil {
		log.Error().Err(err).Str("partner_id", partnerID.(uuid.UUID).String()).Msg("failed to get partner stats")
		return utils.LocalizedError(c, fiber.StatusInternalServerError, "error_internal", "Failed to get partner statistics")
	}

	return c.JSON(stats)
}

// GetMyPartnerCouponsByStatus возвращает статистику купонов партнера по статусам
// @Summary Получить мою статистику купонов по статусам (партнер)
// @Description Возвращает статистику купонов текущего партнера по статусам
// @Tags Partner Stats
// @Accept json
// @Produce json
// @Success 200 {object} CouponsByStatusResponse
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /partner/stats/my/coupons-by-status [get]
func (h *StatsHandler) GetMyPartnerCouponsByStatus(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	partnerID := c.Locals("partner_id")
	if partnerID == nil {
		return utils.LocalizedError(c, fiber.StatusUnauthorized, "error_unauthorized", "Partner not authenticated")
	}

	partnerUUID := partnerID.(uuid.UUID)
	stats, err := h.deps.StatsService.GetCouponsByStatus(c.Context(), &partnerUUID)
	if err != nil {
		log.Error().Err(err).Str("partner_id", partnerUUID.String()).Msg("failed to get partner coupons by status")
		return utils.LocalizedError(c, fiber.StatusInternalServerError, "error_internal", "Failed to get coupons by status statistics")
	}

	return c.JSON(stats)
}

// GetMyPartnerTimeSeriesStats возвращает временную статистику партнера
// @Summary Получить мою временную статистику (партнер)
// @Description Возвращает временную статистику текущего партнера
// @Tags Partner Stats
// @Accept json
// @Produce json
// @Param date_from query string false "Дата начала (YYYY-MM-DD)"
// @Param date_to query string false "Дата окончания (YYYY-MM-DD)"
// @Param period query string false "Период группировки (day, week, month)"
// @Success 200 {object} TimeSeriesStatsResponse
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /partner/stats/my/time-series [get]
func (h *StatsHandler) GetMyPartnerTimeSeriesStats(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	partnerID := c.Locals("partner_id")
	if partnerID == nil {
		return utils.LocalizedError(c, fiber.StatusUnauthorized, "error_unauthorized", "Partner not authenticated")
	}

	partnerUUID := partnerID.(uuid.UUID)
	filters := &StatsFiltersRequest{
		PartnerID: &partnerUUID,
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
		log.Error().Err(err).Str("partner_id", partnerUUID.String()).Msg("failed to get partner time series stats")
		return utils.LocalizedError(c, fiber.StatusInternalServerError, "error_internal", "Failed to get time series statistics")
	}

	return c.JSON(stats)
}
