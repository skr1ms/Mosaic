package partner

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/skr1ms/mosaic/config"
	"github.com/skr1ms/mosaic/internal/coupon"
	"github.com/skr1ms/mosaic/pkg/email"
	"github.com/skr1ms/mosaic/pkg/jwt"
	"github.com/skr1ms/mosaic/pkg/middleware"
)

type PartnerHandlerDeps struct {
	Config           *config.Config
	PartnerService   *PartnerService
	CouponRepository *coupon.CouponRepository
	JwtService       *jwt.JWT
	MailSender       *email.Mailer
}

type PartnerHandler struct {
	fiber.Router
	deps           *PartnerHandlerDeps
	brandingHelper *middleware.BrandingHelper
}

func NewPartnerHandler(router fiber.Router, deps *PartnerHandlerDeps) {
	handler := &PartnerHandler{
		Router:         router,
		deps:           deps,
		brandingHelper: middleware.NewBrandingHelper(),
	}

	partnerRoutes := router.Group("/partner")

	// Публичные endpoints (без JWT)
	partnerRoutes.Post("/forgot", handler.ForgotPassword) // Запрос на сброс пароля
	partnerRoutes.Post("/reset", handler.ResetPassword)   // Сброс пароля

	// Защищенные endpoints (требуют JWT + partner роль)
	protected := partnerRoutes.Use(middleware.JWTMiddleware(deps.JwtService), middleware.PartnerOnly())
	protected.Get("/dashboard", handler.GetDashboard)                                 // Дашборд партнера
	protected.Get("/profile", handler.GetProfile)                                     // Профиль партнера
	protected.Put("/profile", handler.UpdateProfile)                                  // Обновление профиля партнера (только для чтения в партнерской панели)
	protected.Put("/update/password", handler.UpdatePassword)                         // Обновление пароля партнера
	protected.Get("/coupons", handler.GetMyCoupons)                                   // Купоны партнера
	protected.Get("/coupons/filtered", handler.GetMyCouponsFiltered)                  // Купоны партнера с фильтрацией
	protected.Get("/coupons/:id", handler.GetCouponDetail)                            // Детальная информация о купоне
	protected.Get("/coupons/search/:code", handler.SearchCouponByCode)                // Поиск купона по коду
	protected.Get("/coupons/:id/download-materials", handler.DownloadCouponMaterials) // Скачивание материалов купона
	protected.Get("/coupons/export", handler.ExportCoupons)                           // Экспорт купонов партнера
	protected.Get("/statistics", handler.GetMyStatistics)                             // Статистика партнера
	protected.Get("/statistics/sales", handler.GetSalesStatistics)                    // Статистика продаж
	protected.Get("/statistics/usage", handler.GetUsageStatistics)                    // Статистика использования
}

// GetDashboard возвращает данные для дашборда партнера
//
//	@Summary		Дашборд партнера
//	@Description	Возвращает данные для главной страницы партнера
//	@Tags			partner-dashboard
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	PartnerDashboardResponse	"Данные дашборда"
//	@Failure		401	{object}	map[string]interface{}	"Не авторизован"
//	@Failure		403	{object}	map[string]interface{}	"Нет прав доступа"
//	@Router			/partner/dashboard [get]
func (handler *PartnerHandler) GetDashboard(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())

	// Получаем claims из контекста
	claims, err := jwt.GetClaimsFromFiberContext(c)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get JWT claims")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// Получаем статистику партнера
	stats, err := handler.deps.CouponRepository.GetPartnerStatistics(context.Background(), claims.UserID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get partner statistics")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get statistics"})
	}

	// Получаем последнюю активность
	recentActivity, err := handler.deps.CouponRepository.GetPartnerRecentActivity(context.Background(), claims.UserID, 10)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get recent activity")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get recent activity"})
	}

	// Получаем подсчеты по статусам
	statusCounts, err := handler.deps.CouponRepository.GetExtendedStatusCounts(context.Background(), &claims.UserID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get status counts")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get status counts"})
	}

	// Получаем подсчеты по размерам
	sizeCounts, err := handler.deps.CouponRepository.GetSizeCounts(context.Background(), &claims.UserID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get size counts")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get size counts"})
	}

	// Получаем подсчеты по стилям
	styleCounts, err := handler.deps.CouponRepository.GetStyleCounts(context.Background(), &claims.UserID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get style counts")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get style counts"})
	}

	// Преобразуем данные в ответ
	var recentActivityData []PartnerCouponInfo
	for _, coupon := range recentActivity {
		recentActivityData = append(recentActivityData, PartnerCouponInfo{
			ID:               coupon.ID,
			Code:             coupon.Code,
			Status:           coupon.Status,
			Size:             coupon.Size,
			Style:            coupon.Style,
			CreatedAt:        coupon.CreatedAt,
			ActivatedAt:      coupon.ActivatedAt,
			UsedAt:           coupon.UsedAt,
			CompletedAt:      coupon.CompletedAt,
			UserEmail:        coupon.UserEmail,
			HasOriginalImage: coupon.OriginalImageURL != nil,
			HasPreview:       coupon.PreviewURL != nil,
			HasSchema:        coupon.SchemaURL != nil,
			IsPurchased:      coupon.IsPurchased,
			PurchaseEmail:    coupon.PurchaseEmail,
			PurchasedAt:      coupon.PurchasedAt,
		})
	}

	statistics := PartnerStatistics{
		TotalCoupons:     stats["total_coupons"].(int64),
		ActivatedCoupons: stats["activated_coupons"].(int64),
		UsedCoupons:      stats["used_coupons"].(int64),
		CompletedCoupons: stats["completed_coupons"].(int64),
		PurchasedCoupons: stats["purchased_coupons"].(int64),
	}

	if lastActivity := stats["last_activity"]; lastActivity != nil {
		if activity, ok := lastActivity.(*time.Time); ok {
			statistics.LastActivity = activity
		}
	}

	response := PartnerDashboardResponse{
		Statistics:     statistics,
		RecentActivity: recentActivityData,
		StatusCounts:   statusCounts,
		SizeCounts:     sizeCounts,
		StyleCounts:    styleCounts,
	}

	// Добавляем данные брендинга к ответу
	responseWithBranding := handler.brandingHelper.AddBrandingToResponse(c, map[string]interface{}{
		"statistics":      response.Statistics,
		"recent_activity": response.RecentActivity,
		"status_counts":   response.StatusCounts,
		"size_counts":     response.SizeCounts,
		"style_counts":    response.StyleCounts,
	})

	return c.JSON(responseWithBranding)
}

// GetProfile возвращает профиль партнера
//
//	@Summary		Профиль партнера
//	@Description	Возвращает информацию о профиле текущего партнера
//	@Tags			partner-profile
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	map[string]interface{}	"Профиль партнера"
//	@Failure		401	{object}	map[string]interface{}	"Не авторизован"
//	@Failure		403	{object}	map[string]interface{}	"Нет прав доступа"
//	@Failure		404	{object}	map[string]interface{}	"Партнер не найден"
//	@Router			/partner/profile [get]
func (handler *PartnerHandler) GetProfile(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	// Получаем claims из контекста
	claims, err := jwt.GetClaimsFromFiberContext(c)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get JWT claims")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// Получаем партнера
	partner, err := handler.deps.PartnerService.deps.PartnerRepository.GetByID(context.Background(), claims.UserID)
	if err != nil {
		log.Error().Err(err).Msg("Partner not found")
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Partner not found"})
	}

	// Возвращаем профиль партнера
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
		"allow_sales":  partner.AllowSales,
		"status":       partner.Status,
		"created_at":   partner.CreatedAt,
		"updated_at":   partner.UpdatedAt,
	})
}

// UpdateProfile обновляет профиль партнера (только для чтения в партнерской панели)
//
//	@Summary		Обновление профиля партнера
//	@Description	Попытка обновления профиля партнера (доступно только администратору)
//	@Tags			partner-profile
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Failure		403	{object}	map[string]interface{}	"Нет прав доступа"
//	@Router			/partner/profile [put]
func (handler *PartnerHandler) UpdateProfile(c *fiber.Ctx) error {
	// TODO: Реализовать обновление профиля партнера
	return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
}

// UpdatePassword обновляет пароль партнера
//
//	@Summary		Обновление пароля партнера
//	@Description	Обновляет пароль партнера
//	@Tags			partner-profile
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		partner.UpdatePasswordRequest	true	"Новый пароль"
//	@Success		200		{object}	map[string]interface{}	"Пароль обновлен"
//	@Failure		400		{object}	map[string]interface{}	"Ошибка в запросе"
//	@Failure		401		{object}	map[string]interface{}	"Не авторизован"
//	@Failure		404		{object}	map[string]interface{}	"Партнер не найден"
//	@Failure		500		{object}	map[string]interface{}	"Внутренняя ошибка сервера"
//	@Router			/partner/profile/password [put]
func (handler *PartnerHandler) UpdatePassword(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	// Получаем claims из контекста
	claims, err := jwt.GetClaimsFromFiberContext(c)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get JWT claims")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// Парсим запрос
	var req UpdatePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		log.Error().Err(err).Msg("Failed to parse request body")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Bad request"})
	}

	// Валидация запроса
	if err := middleware.ValidateStruct(&req); err != nil {
		log.Error().Err(err).Msg("Validation failed")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Validation failed",
			"details": err.Error(),
		})
	}

	// Обновляем пароль
	err = handler.deps.PartnerService.UpdatePassword(claims.UserID, req.CurrentPassword, req.NewPassword)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update password")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update password"})
	}

	return c.JSON(fiber.Map{"message": "Password updated successfully"})
}

// GetMyCoupons возвращает купоны партнера
//
//	@Summary		Купоны партнера
//	@Description	Возвращает список купонов текущего партнера
//	@Tags			partner-coupons
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	map[string]interface{}	"Список купонов партнера"
//	@Failure		401	{object}	map[string]interface{}	"Не авторизован"
//	@Failure		403	{object}	map[string]interface{}	"Нет прав доступа"
//	@Router			/partner/coupons [get]
func (handler *PartnerHandler) GetMyCoupons(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	// Получаем claims из контекста
	claims, err := jwt.GetClaimsFromFiberContext(c)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get JWT claims")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// Получаем купоны
	coupons, err := handler.deps.CouponRepository.GetByPartnerID(context.Background(), claims.UserID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get coupons")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get coupons"})
	}

	// Возвращаем купоны
	return c.JSON(fiber.Map{
		"message":    "Partner coupons",
		"partner_id": claims.UserID,
		"coupons":    coupons,
	})
}

// ExportCoupons экспортирует купоны партнера в файл
//
//	@Summary		Экспорт купонов партнера
//	@Description	Экспортирует купоны партнера со статусом "new" в формате .txt или .csv
//	@Tags			partner-coupons
//	@Produce		text/plain,text/csv
//	@Security		BearerAuth
//	@Param			format	query		string					false	"Формат файла (txt или csv)"	default(txt)
//	@Success		200		{string}	string					"Файл с купонами"
//	@Failure		401		{object}	map[string]interface{}	"Не авторизован"
//	@Failure		403		{object}	map[string]interface{}	"Нет прав доступа"
//	@Failure		500		{object}	map[string]interface{}	"Внутренняя ошибка сервера"
//	@Router			/partner/coupons/export [get]
func (handler *PartnerHandler) ExportCoupons(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	// Получаем claims из контекста
	claims, err := jwt.GetClaimsFromFiberContext(c)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get JWT claims")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// Получаем формат
	format := strings.ToLower(c.Query("format", "txt"))
	if format != "txt" && format != "csv" {
		format = "txt"
	}

	// Экспортируем купоны используя новый метод
	content, filename, contentType, err := handler.deps.PartnerService.ExportCoupons(claims.UserID, "new", format)
	if err != nil {
		log.Error().Err(err).Msg("Failed to export coupons")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to export coupons"})
	}

	// Устанавливаем заголовки для автоматического скачивания
	c.Set("Content-Type", contentType)
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Set("Cache-Control", "no-cache")

	return c.Send(content)
}

// GetMyStatistics возвращает статистику партнера
//
//	@Summary		Статистика партнера
//	@Description	Возвращает общую статистику текущего партнера
//	@Tags			partner-statistics
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	PartnerStatistics	"Статистика партнера"
//	@Failure		401	{object}	map[string]interface{}	"Не авторизован"
//	@Failure		403	{object}	map[string]interface{}	"Нет прав доступа"
//	@Router			/partner/statistics [get]
func (handler *PartnerHandler) GetMyStatistics(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	claims, err := jwt.GetClaimsFromFiberContext(c)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get JWT claims")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// Получаем детальную статистику партнера
	stats, err := handler.deps.CouponRepository.GetPartnerStatistics(context.Background(), claims.UserID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get partner statistics")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get statistics"})
	}

	statistics := PartnerStatistics{
		TotalCoupons:     stats["total_coupons"].(int64),
		ActivatedCoupons: stats["activated_coupons"].(int64),
		UsedCoupons:      stats["used_coupons"].(int64),
		CompletedCoupons: stats["completed_coupons"].(int64),
		PurchasedCoupons: stats["purchased_coupons"].(int64),
	}

	if lastActivity := stats["last_activity"]; lastActivity != nil {
		if activity, ok := lastActivity.(*time.Time); ok {
			statistics.LastActivity = activity
		}
	}

	return c.JSON(statistics)
}

// GetSalesStatistics возвращает статистику продаж партнера
//
//	@Summary		Статистика продаж
//	@Description	Возвращает статистику продаж текущего партнера
//	@Tags			partner-statistics
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	PartnerSalesStatistics	"Статистика продаж"
//	@Failure		401	{object}	map[string]interface{}	"Не авторизован"
//	@Failure		403	{object}	map[string]interface{}	"Нет прав доступа"
//	@Router			/partner/statistics/sales [get]
func (handler *PartnerHandler) GetSalesStatistics(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	claims, err := jwt.GetClaimsFromFiberContext(c)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get JWT claims")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// Получаем статистику продаж партнера
	salesStats, err := handler.deps.CouponRepository.GetPartnerSalesStatistics(context.Background(), claims.UserID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get partner sales statistics")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get sales statistics"})
	}

	// Преобразуем карты размеров и стилей в слайсы
	var topSizes []SizeStatistic
	if sizes, ok := salesStats["top_sizes"].(map[string]int64); ok {
		for size, count := range sizes {
			topSizes = append(topSizes, SizeStatistic{Size: size, Count: count})
		}
	}

	var topStyles []StyleStatistic
	if styles, ok := salesStats["top_styles"].(map[string]int64); ok {
		for style, count := range styles {
			topStyles = append(topStyles, StyleStatistic{Style: style, Count: count})
		}
	}

	response := PartnerSalesStatistics{
		TotalSales:     salesStats["total_sales"].(int64),
		SalesThisMonth: salesStats["sales_this_month"].(int64),
		SalesThisWeek:  salesStats["sales_this_week"].(int64),
		TopSizes:       topSizes,
		TopStyles:      topStyles,
		// TODO: Добавить временные ряды если нужно
		SalesTimeSeries: []SalesTimePoint{},
	}

	return c.JSON(response)
}

// GetUsageStatistics возвращает статистику использования купонов
//
//	@Summary		Статистика использования купонов
//	@Description	Возвращает статистику использования купонов партнера
//	@Tags			partner-statistics
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	PartnerUsageStatistics	"Статистика использования"
//	@Failure		401	{object}	map[string]interface{}	"Не авторизован"
//	@Failure		403	{object}	map[string]interface{}	"Нет прав доступа"
//	@Router			/partner/statistics/usage [get]
func (handler *PartnerHandler) GetUsageStatistics(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	claims, err := jwt.GetClaimsFromFiberContext(c)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get JWT claims")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// Получаем статистику использования партнера
	usageStats, err := handler.deps.CouponRepository.GetPartnerUsageStatistics(context.Background(), claims.UserID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get partner usage statistics")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get usage statistics"})
	}

	response := PartnerUsageStatistics{
		UsageThisMonth:  usageStats["usage_this_month"].(int64),
		UsageThisWeek:   usageStats["usage_this_week"].(int64),
		ConversionRate:  usageStats["conversion_rate"].(float64),
		CompletionRate:  usageStats["completion_rate"].(float64),
		UsageTimeSeries: []UsageTimePoint{}, // TODO: Добавить временные ряды если нужно
	}

	if avgTime := usageStats["average_time_to_use"]; avgTime != nil {
		if timeVal, ok := avgTime.(*int64); ok {
			response.AverageTimeToUse = timeVal
		}
	}

	return c.JSON(response)
}

// ForgotPassword отправляет email с ссылкой для сброса пароля
//
//	@Summary		Запрос сброса пароля партнера
//	@Description	Отправляет email с ссылкой для сброса пароля партнера
//	@Tags			partner-auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		partner.ForgotPasswordRequest	true	"Email и captcha токен"
//	@Success		200		{object}	map[string]interface{}	"Email отправлен"
//	@Failure		400		{object}	map[string]interface{}	"Ошибка в запросе"
//	@Failure		404		{object}	map[string]interface{}	"Партнер не найден"
//	@Failure		500		{object}	map[string]interface{}	"Внутренняя ошибка сервера"
//	@Router			/partner/forgot-password [post]
//
// ForgotPassword обрабатывает запрос на сброс пароля
func (handler *PartnerHandler) ForgotPassword(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	var reqPayload ForgotPasswordRequest
	if err := c.BodyParser(&reqPayload); err != nil {
		log.Error().Err(err).Msg("Failed to parse request body")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Bad request",
		})
	}

	if err := middleware.ValidateStruct(&reqPayload); err != nil {
		log.Error().Err(err).Msg("Validation failed")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Validation failed",
			"details": err.Error(),
		})
	}

	err := handler.deps.PartnerService.ForgotPassword(context.Background(), reqPayload.Email /*captcha*/)
	if err != nil {
		log.Error().Err(err).Str("email", reqPayload.Email).Msg("Failed to send forgot password email")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to send forgot password email"})
	}

	return c.JSON(fiber.Map{
		"message": "If email exists, email sent",
	})
}

// ResetPassword сбрасывает пароль по токену
//
//	@Summary		Сброс пароля партнера
//	@Description	Сбрасывает пароль партнера используя токен из email
//	@Tags			partner-auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		partner.ResetPasswordRequest	true	"Токен сброса и новый пароль"
//	@Success		200		{object}	map[string]interface{}	"Пароль изменен"
//	@Failure		400		{object}	map[string]interface{}	"Ошибка в запросе"
//	@Failure		401		{object}	map[string]interface{}	"Неверный или истекший токен"
//	@Failure		404		{object}	map[string]interface{}	"Партнер не найден"
//	@Failure		500		{object}	map[string]interface{}	"Внутренняя ошибка сервера"
//	@Router			/partner/reset-password [post]
func (handler *PartnerHandler) ResetPassword(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	var req ResetPasswordRequest

	if err := c.BodyParser(&req); err != nil {
		log.Error().Err(err).Msg("Failed to parse request body")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Bad request"})
	}

	if err := middleware.ValidateStruct(&req); err != nil {
		log.Error().Err(err).Msg("Validation failed")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Validation failed",
			"details": err.Error(),
		})
	}

	err := handler.deps.PartnerService.ResetPassword(context.Background(), req.Token, req.NewPassword)
	if err != nil {
		log.Error().Err(err).Msg("Failed to reset password")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to reset password"})
	}

	return c.JSON(fiber.Map{
		"message": "Password has been reset successfully",
	})
}

// GetMyCouponsFiltered возвращает купоны партнера с фильтрацией и пагинацией
//
//	@Summary		Купоны партнера с фильтрацией
//	@Description	Возвращает список купонов партнера с возможностью фильтрации и пагинации
//	@Tags			partner-coupons
//	@Produce		json
//	@Security		BearerAuth
//	@Param			status			query		string	false	"Статус купона (new, activated, used, completed)"
//	@Param			size			query		string	false	"Размер купона"
//	@Param			style			query		string	false	"Стиль купона"
//	@Param			search			query		string	false	"Поиск по коду купона"
//	@Param			created_from	query		string	false	"Дата создания от (YYYY-MM-DD)"
//	@Param			created_to		query		string	false	"Дата создания до (YYYY-MM-DD)"
//	@Param			activated_from	query		string	false	"Дата активации от (YYYY-MM-DD)"
//	@Param			activated_to	query		string	false	"Дата активации до (YYYY-MM-DD)"
//	@Param			sort_by			query		string	false	"Поле сортировки (created_at, activated_at, used_at, code, status)"
//	@Param			order			query		string	false	"Порядок сортировки (asc, desc)"
//	@Param			page			query		int		false	"Номер страницы"
//	@Param			page_size		query		int		false	"Размер страницы"
//	@Success		200	{object}	PartnerCouponsResponse	"Список купонов партнера"
//	@Failure		401	{object}	map[string]interface{}	"Не авторизован"
//	@Failure		403	{object}	map[string]interface{}	"Нет прав доступа"
//	@Failure		400	{object}	map[string]interface{}	"Ошибка валидации"
//	@Router			/partner/coupons/filtered [get]
func (handler *PartnerHandler) GetMyCouponsFiltered(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())

	// Получаем claims из контекста
	claims, err := jwt.GetClaimsFromFiberContext(c)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get JWT claims")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// Парсим параметры фильтрации
	var filters PartnerCouponFilterRequest
	if err := c.QueryParser(&filters); err != nil {
		log.Error().Err(err).Msg("Failed to parse filter parameters")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid filter parameters"})
	}

	// Валидация
	if err := middleware.ValidateStruct(&filters); err != nil {
		log.Error().Err(err).Msg("Filter validation failed")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Validation failed",
			"details": err.Error(),
		})
	}

	// Устанавливаем значения по умолчанию
	page := 1
	if filters.Page > 0 {
		page = filters.Page
	}

	pageSize := 20
	if filters.PageSize > 0 && filters.PageSize <= 100 {
		pageSize = filters.PageSize
	}

	sortBy := "created_at"
	if filters.SortBy != "" {
		sortBy = filters.SortBy
	}

	order := "desc"
	if filters.Order == "asc" {
		order = "asc"
	}

	// Создаем карту фильтров
	filterMap := make(map[string]interface{})
	if filters.Status != "" {
		filterMap["status"] = filters.Status
	}
	if filters.Size != "" {
		filterMap["size"] = filters.Size
	}
	if filters.Style != "" {
		filterMap["style"] = filters.Style
	}
	if filters.Search != "" {
		filterMap["search"] = filters.Search
	}
	if filters.CreatedFrom != nil {
		filterMap["created_from"] = *filters.CreatedFrom
	}
	if filters.CreatedTo != nil {
		filterMap["created_to"] = *filters.CreatedTo
	}
	if filters.ActivatedFrom != nil {
		filterMap["activated_from"] = *filters.ActivatedFrom
	}
	if filters.ActivatedTo != nil {
		filterMap["activated_to"] = *filters.ActivatedTo
	}

	// Получаем купоны с фильтрацией
	coupons, total, err := handler.deps.CouponRepository.GetPartnerCouponsWithFilter(
		context.Background(), claims.UserID, filterMap, page, pageSize, sortBy, order)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get filtered coupons")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get coupons"})
	}

	// Преобразуем в PartnerCouponInfo
	var couponInfos []PartnerCouponInfo
	for _, coupon := range coupons {
		couponInfos = append(couponInfos, PartnerCouponInfo{
			ID:               coupon.ID,
			Code:             coupon.Code,
			Status:           coupon.Status,
			Size:             coupon.Size,
			Style:            coupon.Style,
			CreatedAt:        coupon.CreatedAt,
			ActivatedAt:      coupon.ActivatedAt,
			UsedAt:           coupon.UsedAt,
			CompletedAt:      coupon.CompletedAt,
			UserEmail:        coupon.UserEmail,
			HasOriginalImage: coupon.OriginalImageURL != nil,
			HasPreview:       coupon.PreviewURL != nil,
			HasSchema:        coupon.SchemaURL != nil,
			IsPurchased:      coupon.IsPurchased,
			PurchaseEmail:    coupon.PurchaseEmail,
			PurchasedAt:      coupon.PurchasedAt,
		})
	}

	// Вычисляем количество страниц
	pages := (total + pageSize - 1) / pageSize

	response := PartnerCouponsResponse{
		Coupons: couponInfos,
		Total:   total,
		Page:    page,
		Limit:   pageSize,
		Pages:   pages,
	}

	return c.JSON(response)
}

// GetCouponDetail возвращает детальную информацию о купоне партнера
//
//	@Summary		Детальная информация о купоне
//	@Description	Возвращает детальную информацию о купоне партнера
//	@Tags			partner-coupons
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"ID купона"
//	@Success		200	{object}	PartnerCouponDetail	"Детальная информация о купоне"
//	@Failure		401	{object}	map[string]interface{}	"Не авторизован"
//	@Failure		403	{object}	map[string]interface{}	"Нет прав доступа"
//	@Failure		404	{object}	map[string]interface{}	"Купон не найден"
//	@Router			/partner/coupons/{id} [get]
func (handler *PartnerHandler) GetCouponDetail(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())

	// Получаем claims из контекста
	claims, err := jwt.GetClaimsFromFiberContext(c)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get JWT claims")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// Получаем ID купона из параметров
	couponIDStr := c.Params("id")
	if couponIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Coupon ID is required"})
	}

	// Конвертируем ID в UUID
	couponID, err := uuid.Parse(couponIDStr)
	if err != nil {
		log.Error().Err(err).Str("coupon_id", couponIDStr).Msg("Invalid coupon ID format")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid coupon ID format"})
	}

	// Получаем детальную информацию о купоне
	coupon, err := handler.deps.CouponRepository.GetPartnerCouponDetail(context.Background(), claims.UserID, couponID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Coupon not found"})
		}
		log.Error().Err(err).Msg("Failed to get coupon detail")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get coupon detail"})
	}

	// Формируем детальный ответ
	detail := PartnerCouponDetail{
		ID:                  coupon.ID,
		Code:                coupon.Code,
		Status:              coupon.Status,
		Size:                coupon.Size,
		Style:               coupon.Style,
		CreatedAt:           coupon.CreatedAt,
		ActivatedAt:         coupon.ActivatedAt,
		UsedAt:              coupon.UsedAt,
		CompletedAt:         coupon.CompletedAt,
		UserEmail:           coupon.UserEmail,
		IsPurchased:         coupon.IsPurchased,
		PurchaseEmail:       coupon.PurchaseEmail,
		PurchasedAt:         coupon.PurchasedAt,
		OriginalImageURL:    coupon.OriginalImageURL,
		PreviewURL:          coupon.PreviewURL,
		SchemaURL:           coupon.SchemaURL,
		SchemaSentEmail:     coupon.SchemaSentEmail,
		SchemaSentAt:        coupon.SchemaSentAt,
		CanDownloadMaterial: coupon.Status == "used" && coupon.UsedAt != nil, // Можно скачивать только для использованных купонов
	}

	return c.JSON(detail)
}

// SearchCouponByCode поиск купона партнера по коду
//
//	@Summary		Поиск купона по коду
//	@Description	Поиск купона партнера по коду
//	@Tags			partner-coupons
//	@Produce		json
//	@Security		BearerAuth
//	@Param			code	path		string	true	"Код купона"
//	@Success		200	{object}	PartnerCouponDetail	"Информация о найденном купоне"
//	@Failure		401	{object}	map[string]interface{}	"Не авторизован"
//	@Failure		403	{object}	map[string]interface{}	"Нет прав доступа"
//	@Failure		404	{object}	map[string]interface{}	"Купон не найден"
//	@Router			/partner/coupons/search/{code} [get]
func (handler *PartnerHandler) SearchCouponByCode(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())

	// Получаем claims из контекста
	claims, err := jwt.GetClaimsFromFiberContext(c)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get JWT claims")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// Получаем код купона из параметров
	code := c.Params("code")
	if code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Coupon code is required"})
	}

	// Ищем купон по коду у данного партнера
	coupon, err := handler.deps.CouponRepository.GetPartnerCouponByCode(context.Background(), claims.UserID, code)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Coupon not found"})
		}
		log.Error().Err(err).Msg("Failed to search coupon by code")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to search coupon"})
	}

	// Формируем детальный ответ
	detail := PartnerCouponDetail{
		ID:                  coupon.ID,
		Code:                coupon.Code,
		Status:              coupon.Status,
		Size:                coupon.Size,
		Style:               coupon.Style,
		CreatedAt:           coupon.CreatedAt,
		ActivatedAt:         coupon.ActivatedAt,
		UsedAt:              coupon.UsedAt,
		CompletedAt:         coupon.CompletedAt,
		UserEmail:           coupon.UserEmail,
		IsPurchased:         coupon.IsPurchased,
		PurchaseEmail:       coupon.PurchaseEmail,
		PurchasedAt:         coupon.PurchasedAt,
		OriginalImageURL:    coupon.OriginalImageURL,
		PreviewURL:          coupon.PreviewURL,
		SchemaURL:           coupon.SchemaURL,
		SchemaSentEmail:     coupon.SchemaSentEmail,
		SchemaSentAt:        coupon.SchemaSentAt,
		CanDownloadMaterial: coupon.Status == "used" && coupon.UsedAt != nil,
	}

	return c.JSON(detail)
}

// DownloadCouponMaterials скачивание материалов купона (оригинал, превью, схема)
//
//	@Summary		Скачивание материалов купона
//	@Description	Скачивание ZIP-архива с материалами купона (доступно только для использованных купонов)
//	@Tags			partner-coupons
//	@Produce		application/zip
//	@Security		BearerAuth
//	@Param			id	path		string	true	"ID купона"
//	@Success		200	{string}	string	"ZIP-архив с материалами"
//	@Failure		401	{object}	map[string]interface{}	"Не авторизован"
//	@Failure		403	{object}	map[string]interface{}	"Нет прав доступа"
//	@Failure		404	{object}	map[string]interface{}	"Купон не найден"
//	@Failure		400	{object}	map[string]interface{}	"Нельзя скачать материалы"
//	@Router			/partner/coupons/{id}/download-materials [get]
func (handler *PartnerHandler) DownloadCouponMaterials(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())

	// Получаем claims из контекста
	claims, err := jwt.GetClaimsFromFiberContext(c)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get JWT claims")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// Получаем ID купона из параметров
	couponIDStr := c.Params("id")
	if couponIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Coupon ID is required"})
	}

	// Конвертируем ID в UUID
	couponID, err := uuid.Parse(couponIDStr)
	if err != nil {
		log.Error().Err(err).Str("coupon_id", couponIDStr).Msg("Invalid coupon ID format")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid coupon ID format"})
	}

	// Получаем информацию о купоне
	coupon, err := handler.deps.CouponRepository.GetPartnerCouponDetail(context.Background(), claims.UserID, couponID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Coupon not found"})
		}
		log.Error().Err(err).Msg("Failed to get coupon detail")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get coupon detail"})
	}

	// Проверяем, что купон использован и есть материалы для скачивания
	if coupon.Status != "used" || coupon.UsedAt == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Materials can only be downloaded for used coupons",
		})
	}

	if coupon.OriginalImageURL == nil && coupon.PreviewURL == nil && coupon.SchemaURL == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No materials available for download",
		})
	}

	// TODO: Реализовать создание ZIP-архива с материалами
	// Здесь должна быть логика:
	// 1. Скачивание файлов с S3 по URL'ам
	// 2. Создание ZIP-архива
	// 3. Отправка архива клиенту

	// Пока возвращаем информацию о доступных материалах
	materials := make(map[string]interface{})
	if coupon.OriginalImageURL != nil {
		materials["original_image"] = *coupon.OriginalImageURL
	}
	if coupon.PreviewURL != nil {
		materials["preview"] = *coupon.PreviewURL
	}
	if coupon.SchemaURL != nil {
		materials["schema"] = *coupon.SchemaURL
	}

	return c.JSON(fiber.Map{
		"message":     "Download materials endpoint - implementation needed",
		"coupon_code": coupon.Code,
		"materials":   materials,
		"note":        "This endpoint needs implementation for actual file download from S3 and ZIP creation",
	})
}
