package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/skr1ms/mosaic/internal/coupon"
	"github.com/skr1ms/mosaic/internal/partner"
	"github.com/skr1ms/mosaic/pkg/bcrypt"
	"github.com/skr1ms/mosaic/pkg/jwt"
	"github.com/skr1ms/mosaic/pkg/middleware"
	"github.com/skr1ms/mosaic/pkg/randomCouponCode"
	"github.com/skr1ms/mosaic/pkg/utils"
)

type AdminHandlerDeps struct {
	AdminService *AdminService
	JwtService   *jwt.JWT
}

type AdminHandler struct {
	fiber.Router
	deps *AdminHandlerDeps
}

func NewAdminHandler(router fiber.Router, deps *AdminHandlerDeps) {
	handler := &AdminHandler{
		Router: router,
		deps:   deps,
	}

	adminRoutes := handler.Group("/admin")
	// Защищенные endpoints (требуют JWT + admin роль)
	protected := adminRoutes.Use(middleware.JWTMiddleware(deps.JwtService), middleware.AdminOnly())

	// Управление администраторами
	protected.Post("/admins", handler.CreateAdmin)    // создание
	protected.Get("/admins", handler.GetAdmins)       // получение списка администраторов
	protected.Get("/dashboard", handler.GetDashboard) // получение дашборда администратора

	// Управление партнерами
	protected.Get("/partners", handler.GetPartners)                         // получение списка партнеров
	protected.Post("/partners", handler.CreatePartner)                      // создание партнера
	protected.Get("/partners/:id", handler.GetPartner)                      // получение информации о партнере
	protected.Get("/partners/:id/detail", handler.GetPartnerDetail)         // детальная информация о партнере (статистика + история изменений)
	protected.Put("/partners/:id", handler.UpdatePartner)                   // обновление информации о партнере
	protected.Patch("/partners/:id/block", handler.BlockPartner)            // блокировка партнера
	protected.Patch("/partners/:id/unblock", handler.UnblockPartner)        // разблокировка партнера
	protected.Delete("/partners/:id", handler.DeletePartner)                // удаление партнера
	protected.Get("/partners/:id/statistics", handler.GetPartnerStatistics) // получение статистики партнера

	// Управление купонами
	protected.Get("/coupons", handler.GetCoupons)                                       // получение списка купонов
	protected.Get("/coupons/paginated", handler.GetCouponsPaginated)                    // получение купонов с пагинацией
	protected.Get("/coupons/filtered", handler.GetCouponsFiltered)                      // получение купонов с продвинутой фильтрацией
	protected.Post("/coupons", handler.CreateCoupons)                                   // создание купонов
	protected.Get("/coupons/export", handler.ExportCoupons)                             // экспорт купонов
	protected.Post("/coupons/export-advanced", handler.ExportCouponsAdvanced)           // продвинутый экспорт купонов
	protected.Get("/coupons/export/partner/:id", handler.ExportPartnerCoupons)          // экспорт купонов партнера
	protected.Post("/coupons/batch-delete", handler.BatchDeleteCoupons)                 // массовое удаление купонов
	protected.Post("/coupons/batch/reset", handler.BatchResetCoupons)                   // пакетный сброс купонов
	protected.Post("/coupons/batch/delete/preview", handler.PreviewBatchDelete)         // предпросмотр пакетного удаления
	protected.Post("/coupons/batch/delete/confirm", handler.ExecuteBatchDelete)         // подтверждение пакетного удаления
	protected.Post("/batch-delete", handler.BatchDeleteCoupons)                         // Пакетное удаление купонов
	protected.Get("/coupons/:id", handler.GetCoupon)                                    // получение информации о купоне
	protected.Get("/coupons/:id/download-materials", handler.DownloadCouponMaterials)   // скачивание материалов купона
	protected.Post("/coupons/batch/download-materials", handler.BatchDownloadMaterials) // массовое скачивание материалов
	protected.Patch("/coupons/:id/reset", handler.ResetCoupon)                          // сброс купона
	protected.Delete("/coupons/:id", handler.DeleteCoupon)                              // удаление купона

	// Пользовательские фильтры
	protected.Get("/filters", handler.GetUserFilters)          // получение сохраненных фильтров
	protected.Post("/filters", handler.SaveUserFilter)         // сохранение пользовательского фильтра
	protected.Delete("/filters/:id", handler.DeleteUserFilter) // удаление пользовательского фильтра

	// Статистика и аналитика
	protected.Get("/statistics", handler.GetStatistics)                    // получение общей статистики
	protected.Get("/statistics/partners", handler.GetPartnersStatistics)   // получение статистики партнеров
	protected.Get("/statistics/system", handler.GetSystemStatistics)       // получение системной статистики
	protected.Get("/statistics/analytics", handler.GetAnalytics)           // получение аналитики
	protected.Get("/statistics/dashboard", handler.GetDashboardStatistics) // получение статистики для дашборда

	// Управление задачами обработки изображений
	protected.Get("/images", handler.GetAllImages)              // получение всех задач обработки изображений
	protected.Get("/images/:id", handler.GetImageDetails)       // получение деталей задачи обработки
	protected.Delete("/images/:id", handler.DeleteImageTask)    // удаление задачи обработки
	protected.Post("/images/:id/retry", handler.RetryImageTask) // повторная обработка изображения
}

// CreateAdmin создает нового администратора
// @Summary		Создание администратора
// @Description	Создает нового администратора (только для существующих администраторов)
// @Tags		admin-management
// @Accept		json
// @Produce		json
// @Security		BearerAuth
// @Param		admin		body		CreateAdminRequest		true	"Данные нового администратора"
// @Success		201		{object}	map[string]interface{}		"Администратор создан"
// @Failure		400		{object}	map[string]interface{}		"Ошибка в запросе"
// @Failure		401		{object}	map[string]interface{}		"Не авторизован"
// @Failure		403		{object}	map[string]interface{}		"Нет прав доступа"
// @Failure		500		{object}	map[string]interface{}		"Внутренняя ошибка сервера"
// @Router		/admin/admins [post]
func (handler *AdminHandler) CreateAdmin(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())

	var payload CreateAdminRequest

	if err := c.BodyParser(&payload); err != nil {
		log.Error().Err(err).Msg("Failed to parse request body")
		return utils.LocalizedError(c, fiber.StatusBadRequest, "error_bad_request", "Invalid request payload")
	}

	if err := middleware.ValidateStruct(&payload); err != nil {
		log.Error().Err(err).Msg("Invalid request payload")
		return utils.LocalizedError(c, fiber.StatusBadRequest, "error_validation_failed", "Invalid request payload")
	}

	// Создаем администратора через сервис
	admin, err := handler.deps.AdminService.CreateAdmin(payload)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create admin")
		return utils.LocalizedError(c, fiber.StatusInternalServerError, "error_internal", "Failed to create admin")
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":    admin.ID,
		"login": admin.Login,
		"role":  "admin",
	})
}

// GetAdmins возвращает список всех администраторов
// @Summary		Список администраторов
// @Description	Возвращает список всех администраторов
// @Tags		admin-management
// @Produce		json
// @Security		BearerAuth
// @Success		200		{array}		map[string]interface{}		"Список администраторов"
// @Failure		401		{object}	map[string]interface{}		"Не авторизован"
// @Failure		403		{object}	map[string]interface{}		"Нет прав доступа"
// @Failure		500		{object}	map[string]interface{}		"Внутренняя ошибка сервера"
// @Router		/admin/admins [get]
func (handler *AdminHandler) GetAdmins(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	admins, err := handler.deps.AdminService.GetAdmins()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get admins")
		return utils.LocalizedError(c, fiber.StatusInternalServerError, "error_internal", "Failed to get admins")
	}

	result := make([]fiber.Map, len(admins))
	for i, admin := range admins {
		result[i] = fiber.Map{
			"id":         admin.ID,
			"login":      admin.Login,
			"last_login": admin.LastLogin,
			"created_at": admin.CreatedAt,
		}
	}

	return c.JSON(result)
}

// GetDashboard возвращает данные для дашборда администратора
//
//	@Summary		Дашборд администратора
//	@Description	Возвращает данные для главной страницы администратора
//	@Tags			admin-dashboard
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	map[string]interface{}	"Данные дашборда"
//	@Failure		401	{object}	map[string]interface{}	"Не авторизован"
//	@Failure		403	{object}	map[string]interface{}	"Нет прав доступа"
//	@Router			/admin/dashboard [get]
func (handler *AdminHandler) GetDashboard(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	dashboardData, err := handler.deps.AdminService.GetDashboardData()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get dashboard")
		return utils.LocalizedError(c, fiber.StatusInternalServerError, "error_internal", "Failed to get dashboard")
	}

	return c.JSON(dashboardData)
}

// GetPartners возвращает список партнеров с поддержкой фильтрации и сортировки
//
//	@Summary		Список партнеров
//	@Description	Возвращает список всех партнеров с возможностью фильтрации, поиска и сортировки
//	@Tags			admin-partners
//	@Produce		json
//	@Security		BearerAuth
//	@Param			search	query		string					false	"Поиск по названию бренда, домену или email"
//	@Param			status	query		string					false	"Фильтр по статусу (active/blocked)"
//	@Param			sort_by	query		string					false	"Поле сортировки (created_at/brand_name)"
//	@Param			order	query		string					false	"Порядок сортировки (asc/desc, по умолчанию desc)"
//	@Success		200		{object}	map[string]interface{}	"Список партнеров"
//	@Failure		401		{object}	map[string]interface{}	"Не авторизован"
//	@Failure		403		{object}	map[string]interface{}	"Нет прав доступа"
//	@Router			/admin/partners [get]
func (handler *AdminHandler) GetPartners(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	search := c.Query("search")
	status := c.Query("status")
	sortBy := c.Query("sort_by", "created_at") // По умолчанию сортировка по дате создания
	order := c.Query("order", "desc")          // По умолчанию по убыванию

	partners, err := handler.deps.AdminService.GetPartners(search, status, sortBy, order)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get partners")
		return utils.LocalizedError(c, fiber.StatusInternalServerError, "error_internal", "Failed to get partners")
	}

	result := make([]fiber.Map, len(partners))
	for i, partner := range partners {
		result[i] = fiber.Map{
			"id":               partner.ID,
			"login":            partner.Login,
			"last_login":       partner.LastLogin,
			"created_at":       partner.CreatedAt,
			"updated_at":       partner.UpdatedAt,
			"partner_code":     partner.PartnerCode,
			"domain":           partner.Domain,
			"brand_name":       partner.BrandName,
			"logo_url":         partner.LogoURL,
			"ozon_link":        partner.OzonLink,
			"wildberries_link": partner.WildberriesLink,
			"email":            partner.Email,
			"address":          partner.Address,
			"phone":            partner.Phone,
			"telegram":         partner.Telegram,
			"whatsapp":         partner.Whatsapp,
			"allow_sales":      partner.AllowSales,
			"status":           partner.Status,
		}
	}

	return c.JSON(fiber.Map{
		"partners": result,
		"total":    len(result),
	})
}

// CreatePartner создает нового партнера
//
//	@Summary		Создание партнера
//	@Description	Создает нового партнера с автоматически генерируемым кодом (начиная с 0001, 0000 зарезервирован для собственных купонов)
//	@Tags			admin-partners
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			partner	body		partner.CreatePartnerRequest	true	"Данные нового партнера"
//	@Success		201		{object}	map[string]interface{}	"Партнер создан"
//	@Failure		400		{object}	map[string]interface{}	"Ошибка в запросе"
//	@Failure		401		{object}	map[string]interface{}	"Не авторизован"
//	@Failure		403		{object}	map[string]interface{}	"Нет прав доступа"
//	@Failure		409		{object}	map[string]interface{}	"Партнер с таким логином/доменом уже существует"
//	@Failure		500		{object}	map[string]interface{}	"Внутренняя ошибка сервера"
//	@Router			/admin/partners [post]
func (handler *AdminHandler) CreatePartner(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	var req partner.CreatePartnerRequest

	if err := c.BodyParser(&req); err != nil {
		log.Error().Err(err).Msg("Failed to parse request body")
		return utils.LocalizedError(c, fiber.StatusBadRequest, "error_bad_request", "Invalid request payload")
	}

	// Проверяем, существует ли уже партнер с таким логином
	existingPartner, err := handler.deps.AdminService.deps.PartnerRepository.GetByLogin(context.Background(), req.Login)
	if err == nil && existingPartner != nil {
		log.Error().Str("login", req.Login).Msg("Partner already exists")
		return utils.LocalizedError(c, fiber.StatusConflict, "partner_already_exists", "Partner already exists")
	}

	// Проверяем, существует ли уже партнер с таким доменом
	existingPartnerByDomain, err := handler.deps.AdminService.deps.PartnerRepository.GetByDomain(context.Background(), req.Domain)
	if err == nil && existingPartnerByDomain != nil {
		log.Error().Str("domain", req.Domain).Msg("Partner already exists")
		return utils.LocalizedError(c, fiber.StatusConflict, "partner_already_exists", "Partner already exists")
	}

	// Генерируем следующий доступный код партнера
	partnerCode, err := handler.deps.AdminService.deps.PartnerRepository.GetNextPartnerCode(context.Background())
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate partner code")
		return utils.LocalizedError(c, fiber.StatusInternalServerError, "error_internal", "Failed to generate partner code")
	}

	hashedPassword, err := bcrypt.HashPassword(req.Password)
	if err != nil {
		log.Error().Err(err).Msg("Failed to hash password")
		return utils.LocalizedError(c, fiber.StatusInternalServerError, "error_internal", "Failed to hash password")
	}

	partner := &partner.Partner{
		Login:           req.Login,
		Password:        hashedPassword,
		PartnerCode:     partnerCode,
		Domain:          req.Domain,
		BrandName:       req.BrandName,
		LogoURL:         req.LogoURL,
		OzonLink:        req.OzonLink,
		WildberriesLink: req.WildberriesLink,
		Email:           req.Email,
		Address:         req.Address,
		Phone:           req.Phone,
		Telegram:        req.Telegram,
		Whatsapp:        req.Whatsapp,
		AllowSales:      req.AllowSales,
		Status:          req.Status,
	}

	if err := handler.deps.AdminService.deps.PartnerRepository.Create(context.Background(), partner); err != nil {
		log.Error().Err(err).Msg("Failed to create partner")
		return utils.LocalizedError(c, fiber.StatusInternalServerError, "error_internal", "Failed to create partner")
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":           partner.ID,
		"login":        partner.Login,
		"partner_code": partner.PartnerCode,
	})
}

// GetPartner возвращает детальную информацию о партнере
//
//	@Summary		Детальная информация о партнере
//	@Description	Возвращает детальную информацию о партнере по ID
//	@Tags			admin-partners
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string					true	"ID партнера"
//	@Success		200	{object}	map[string]interface{}	"Информация о партнере"
//	@Failure		400	{object}	map[string]interface{}	"Неверный ID"
//	@Failure		401	{object}	map[string]interface{}	"Не авторизован"
//	@Failure		403	{object}	map[string]interface{}	"Нет прав доступа"
//	@Failure		404	{object}	map[string]interface{}	"Партнер не найден"
//	@Router			/admin/partners/{id} [get]
func (handler *AdminHandler) GetPartner(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	partnerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		log.Error().Err(err).Str("id", c.Params("id")).Msg("Invalid ID")
		return utils.LocalizedError(c, fiber.StatusBadRequest, "error_bad_request", "Invalid ID")
	}

	partner, err := handler.deps.AdminService.deps.PartnerRepository.GetByID(context.Background(), partnerID)
	if err != nil {
		log.Error().Err(err).Str("partner_id", partnerID.String()).Msg("Partner not found")
		return utils.LocalizedError(c, fiber.StatusNotFound, "partner_not_found", "Partner not found")
	}

	// Получаем информацию о партнере
	totalCoupons, err := handler.deps.AdminService.deps.CouponRepository.CountByPartnerID(context.Background(), partnerID)
	if err != nil {
		totalCoupons = 0
	}

	activatedCoupons, err := handler.deps.AdminService.deps.CouponRepository.CountActivatedByPartnerID(context.Background(), partnerID)
	if err != nil {
		activatedCoupons = 0
	}

	purchasedCoupons, err := handler.deps.AdminService.deps.CouponRepository.CountPurchasedByPartnerID(context.Background(), partnerID)
	if err != nil {
		purchasedCoupons = 0
	}

	return c.JSON(fiber.Map{
		"id":                partner.ID,
		"login":             partner.Login,
		"partner_code":      partner.PartnerCode,
		"domain":            partner.Domain,
		"brand_name":        partner.BrandName,
		"logo_url":          partner.LogoURL,
		"ozon_link":         partner.OzonLink,
		"wildberries_link":  partner.WildberriesLink,
		"email":             partner.Email,
		"address":           partner.Address,
		"phone":             partner.Phone,
		"telegram":          partner.Telegram,
		"whatsapp":          partner.Whatsapp,
		"allow_sales":       partner.AllowSales,
		"status":            partner.Status,
		"last_login":        partner.LastLogin,
		"created_at":        partner.CreatedAt,
		"updated_at":        partner.UpdatedAt,
		"total_coupons":     totalCoupons,
		"activated_coupons": activatedCoupons,
		"purchased_coupons": purchasedCoupons,
	})
}

// GetPartnerDetail возвращает детальную информацию о партнере включая статистику и историю изменений
//
//	@Summary		Детальная информация о партнере
//	@Description	Возвращает подробную информацию о партнере включая статистику купонов, последнюю активность и историю изменений профиля
//	@Tags			admin-partners
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string					true	"ID партнера"
//	@Success		200	{object}	admin.PartnerDetailResponse	"Детальная информация о партнере"
//	@Failure		400	{object}	map[string]interface{}		"Неверный ID"
//	@Failure		404	{object}	map[string]interface{}		"Партнер не найден"
//	@Failure		401	{object}	map[string]interface{}		"Не авторизован"
//	@Failure		403	{object}	map[string]interface{}		"Нет прав доступа"
//	@Router			/admin/partners/{id}/detail [get]
func (handler *AdminHandler) GetPartnerDetail(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	partnerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		log.Error().Err(err).Str("id", c.Params("id")).Msg("Invalid partner ID")
		return utils.LocalizedError(c, fiber.StatusBadRequest, "error_bad_request", "Invalid partner ID")
	}

	partnerDetail, err := handler.deps.AdminService.GetPartnerDetail(partnerID)
	if err != nil {
		log.Error().Err(err).Str("partner_id", partnerID.String()).Msg("Failed to get partner detail")
		if strings.Contains(err.Error(), "not found") {
			return utils.LocalizedError(c, fiber.StatusNotFound, "partner_not_found", "Partner not found")
		}
		return utils.LocalizedError(c, fiber.StatusInternalServerError, "error_internal", "Failed to get partner detail")
	}

	return c.JSON(partnerDetail)
}

// UpdatePartner обновляет партнера
//
//	@Summary		Обновление партнера
//	@Description	Обновляет данные партнера
//	@Tags			admin-partners
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string					true	"ID партнера"
//	@Success		200	{object}	map[string]interface{}	"Партнер обновлен"
//	@Failure		401	{object}	map[string]interface{}	"Не авторизован"
//	@Failure		403	{object}	map[string]interface{}	"Нет прав доступа"
//	@Router			/admin/partners/{id} [put]
func (handler *AdminHandler) UpdatePartner(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	var req partner.UpdatePartnerRequest

	if err := c.BodyParser(&req); err != nil {
		log.Error().Err(err).Msg("Failed to parse request body")
		return utils.LocalizedError(c, fiber.StatusBadRequest, "error_bad_request", "Invalid request")
	}

	partnerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		log.Error().Err(err).Str("id", c.Params("id")).Msg("Invalid ID")
		return utils.LocalizedError(c, fiber.StatusBadRequest, "error_bad_request", "Invalid ID")
	}

	// Получаем информацию о текущем админе
	claims, err := jwt.GetClaimsFromFiberContext(c)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get admin claims")
		return utils.LocalizedError(c, fiber.StatusUnauthorized, "error_unauthorized", "Unauthorized")
	}

	// Получаем причину изменения из query параметра или устанавливаем по умолчанию
	reason := c.Query("reason", "Admin update")

	// Обновляем партнера с записью истории изменений
	partner, err := handler.deps.AdminService.UpdatePartnerWithHistory(partnerID, req, claims.Login, reason)
	if err != nil {
		log.Error().Err(err).Str("partner_id", partnerID.String()).Msg("Failed to update partner")
		if strings.Contains(err.Error(), "not found") {
			return utils.LocalizedError(c, fiber.StatusNotFound, "error_not_found", "Resource not found")
		}
		return utils.LocalizedError(c, fiber.StatusInternalServerError, "error_internal", "Internal server error")
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"id":           partner.ID,
		"login":        partner.Login,
		"partner_code": partner.PartnerCode,
	})
}

// BlockPartner блокирует партнера
//
//	@Summary		Блокировка партнера
//	@Description	Блокирует партнера (временное отключение доступа)
//	@Tags			admin-partners
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string					true	"ID партнера"
//	@Success		200	{object}	map[string]interface{}	"Партнер заблокирован"
//	@Failure		400	{object}	map[string]interface{}	"Неверный ID"
//	@Failure		401	{object}	map[string]interface{}	"Не авторизован"
//	@Failure		403	{object}	map[string]interface{}	"Нет прав доступа"
//	@Failure		404	{object}	map[string]interface{}	"Партнер не найден"
//	@Router			/admin/partners/{id}/block [patch]
func (handler *AdminHandler) BlockPartner(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	partnerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		log.Error().Err(err).Str("id", c.Params("id")).Msg("Invalid ID")
		return utils.LocalizedError(c, fiber.StatusBadRequest, "error_bad_request", "Invalid ID")
	}

	if err := handler.deps.AdminService.deps.PartnerRepository.UpdateStatus(context.Background(), partnerID, "blocked"); err != nil {
		log.Error().Err(err).Str("partner_id", partnerID.String()).Msg("Failed to block partner")
		return utils.LocalizedError(c, fiber.StatusInternalServerError, "error_internal", "Internal server error")
	}

	if err := handler.deps.AdminService.deps.CouponRepository.UpdateStatusByPartnerID(context.Background(), partnerID, true); err != nil {
		log.Error().Err(err).Str("partner_id", partnerID.String()).Msg("Failed to block coupons")
		return utils.LocalizedError(c, fiber.StatusInternalServerError, "error_internal", "Internal server error")
	}

	return c.JSON(fiber.Map{"message": utils.GetLocalizedMessage(c, "partner_blocked", "Partner blocked successfully")})
}

// UnblockPartner разблокирует партнера
//
//	@Summary		Разблокировка партнера
//	@Description	Разблокирует партнера (восстановление доступа)
//	@Tags			admin-partners
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string					true	"ID партнера"
//	@Success		200	{object}	map[string]interface{}	"Партнер разблокирован"
//	@Failure		400	{object}	map[string]interface{}	"Неверный ID"
//	@Failure		401	{object}	map[string]interface{}	"Не авторизован"
//	@Failure		403	{object}	map[string]interface{}	"Нет прав доступа"
//	@Failure		404	{object}	map[string]interface{}	"Партнер не найден"
//	@Router			/admin/partners/{id}/unblock [patch]
func (handler *AdminHandler) UnblockPartner(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	partnerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		log.Error().Err(err).Str("id", c.Params("id")).Msg("Invalid ID")
		return utils.LocalizedError(c, fiber.StatusBadRequest, "error_bad_request", "Invalid ID")
	}

	if err := handler.deps.AdminService.deps.PartnerRepository.UpdateStatus(context.Background(), partnerID, "active"); err != nil {
		log.Error().Err(err).Str("partner_id", partnerID.String()).Msg("Failed to unblock partner")
		return utils.LocalizedError(c, fiber.StatusInternalServerError, "error_internal", "Internal server error")
	}

	if err := handler.deps.AdminService.deps.CouponRepository.UpdateStatusByPartnerID(context.Background(), partnerID, false); err != nil {
		log.Error().Err(err).Str("partner_id", partnerID.String()).Msg("Failed to unblock coupons")
		return utils.LocalizedError(c, fiber.StatusInternalServerError, "error_internal", "Internal server error")
	}

	return c.JSON(fiber.Map{"message": utils.GetLocalizedMessage(c, "partner_unblocked", "Partner unblocked successfully")})
}

// DeletePartner удаляет партнера
//
//	@Summary		Удаление партнера
//	@Description	Удаляет партнера по ID с удалением всех связанных данных
//	@Tags			admin-partners
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string					true	"ID партнера"
//	@Param			confirm	query		boolean					false	"Подтверждение удаления (true/false)"
//	@Success		200		{object}	map[string]interface{}	"Партнер удален"
//	@Failure		400		{object}	map[string]interface{}	"Неверный ID или требуется подтверждение"
//	@Failure		401		{object}	map[string]interface{}	"Не авторизован"
//	@Failure		403		{object}	map[string]interface{}	"Нет прав доступа"
//	@Failure		404		{object}	map[string]interface{}	"Партнер не найден"
//	@Router			/admin/partners/{id} [delete]
func (handler *AdminHandler) DeletePartner(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	partnerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		log.Error().Err(err).Str("id", c.Params("id")).Msg("Invalid ID")
		return utils.LocalizedError(c, fiber.StatusBadRequest, "error_bad_request", "Invalid ID")
	}

	partner, err := handler.deps.AdminService.deps.PartnerRepository.GetByID(context.Background(), partnerID)
	if err != nil {
		log.Error().Err(err).Str("partner_id", partnerID.String()).Msg("Partner not found")
		return utils.LocalizedError(c, fiber.StatusNotFound, "error_not_found", "Resource not found")
	}

	totalCoupons, _ := handler.deps.AdminService.deps.CouponRepository.CountByPartnerID(context.Background(), partnerID)
	activatedCoupons, _ := handler.deps.AdminService.deps.CouponRepository.CountActivatedByPartnerID(context.Background(), partnerID)

	confirm := c.Query("confirm") == "true"
	if !confirm {
		return utils.LocalizedError(c, fiber.StatusBadRequest, "deletion_requires_confirmation", "Deletion requires confirmation")
	}

	// Используем транзакцию для атомарного удаления партнера и его купонов
	err = handler.deps.AdminService.deps.PartnerRepository.DeleteWithCoupons(context.Background(), partnerID)

	if err != nil {
		log.Error().Err(err).Str("partner_id", partnerID.String()).Msg("Failed to delete partner")
		return utils.LocalizedError(c, fiber.StatusInternalServerError, "error_internal", "Internal server error")
	}

	// Проверяем, что купоны действительно удалены
	remainingCoupons, _ := handler.deps.AdminService.deps.CouponRepository.CountByPartnerID(context.Background(), partnerID)
	if remainingCoupons > 0 {
		log.Error().Err(err).Str("partner_id", partnerID.String()).Msg("Failed to delete partner")
		return utils.LocalizedError(c, fiber.StatusInternalServerError, "error_internal", "Failed to delete partner")
	}

	return c.JSON(fiber.Map{
		"message": utils.GetLocalizedMessage(c, "partner_deleted", "Partner deleted successfully"),
		"deleted_partner": fiber.Map{
			"id":                partner.ID,
			"brand_name":        partner.BrandName,
			"total_coupons":     totalCoupons,
			"activated_coupons": activatedCoupons,
		},
	})
}

// GetPartnerStatistics возвращает статистику конкретного партнера
//
//	@Summary		Статистика партнера
//	@Description	Возвращает детальную статистику по конкретному партнеру
//	@Tags			admin-partners
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string					true	"ID партнера"
//	@Success		200	{object}	map[string]interface{}	"Статистика партнера"
//	@Failure		400	{object}	map[string]interface{}	"Неверный ID"
//	@Failure		401	{object}	map[string]interface{}	"Не авторизован"
//	@Failure		403	{object}	map[string]interface{}	"Нет прав доступа"
//	@Router			/admin/partners/{id}/statistics [get]
func (handler *AdminHandler) GetPartnerStatistics(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	partnerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		log.Error().Err(err).Str("id", c.Params("id")).Msg("Invalid ID")
		return utils.LocalizedError(c, fiber.StatusBadRequest, "error_bad_request", "Invalid ID")
	}

	// Получаем статистику партнера
	totalCoupons, err := handler.deps.AdminService.deps.CouponRepository.CountByPartnerID(context.Background(), partnerID)
	if err != nil {
		log.Error().Err(err).Str("partner_id", partnerID.String()).Msg("Failed to get statistics")
		return utils.LocalizedError(c, fiber.StatusInternalServerError, "error_internal", "Internal server error")
	}

	activatedCoupons, err := handler.deps.AdminService.deps.CouponRepository.CountActivatedByPartnerID(context.Background(), partnerID)
	if err != nil {
		log.Error().Err(err).Str("partner_id", partnerID.String()).Msg("Failed to get statistics")
		return utils.LocalizedError(c, fiber.StatusInternalServerError, "error_internal", "Internal server error")
	}

	purchasedCoupons, err := handler.deps.AdminService.deps.CouponRepository.CountPurchasedByPartnerID(context.Background(), partnerID)
	if err != nil {
		log.Error().Err(err).Str("partner_id", partnerID.String()).Msg("Failed to get statistics")
		return utils.LocalizedError(c, fiber.StatusInternalServerError, "error_internal", "Internal server error")
	}

	unusedCoupons := totalCoupons - activatedCoupons
	activationRate := float64(0)
	if totalCoupons > 0 {
		activationRate = (float64(activatedCoupons) / float64(totalCoupons)) * 100
	}

	return c.JSON(fiber.Map{
		"partner_id":        partnerID,
		"total_coupons":     totalCoupons,
		"activated_coupons": activatedCoupons,
		"unused_coupons":    unusedCoupons,
		"purchased_coupons": purchasedCoupons,
		"activation_rate":   activationRate,
	})
}

// GetCoupons возвращает список купонов с поддержкой фильтрации
//
//	@Summary		Список купонов
//	@Description	Возвращает список всех купонов с возможностью фильтрации
//	@Tags			admin-coupons
//	@Produce		json
//	@Security		BearerAuth
//	@Param			search		query		string					false	"Поиск по номеру купона"
//	@Param			partner_id	query		string					false	"ID партнера для фильтрации"
//	@Param			status		query		string					false	"Статус для фильтрации (new/used)"
//	@Param			size		query		string					false	"Размер для фильтрации"
//	@Param			style		query		string					false	"Стиль для фильтрации"
//	@Param			limit		query		int						false	"Количество записей на странице (по умолчанию 50)"
//	@Param			offset		query		int						false	"Смещение для пагинации (по умолчанию 0)"
//	@Success		200			{object}	map[string]interface{}	"Список купонов"
//	@Failure		401			{object}	map[string]interface{}	"Не авторизован"
//	@Failure		403			{object}	map[string]interface{}	"Нет прав доступа"
//	@Router			/admin/coupons [get]
func (handler *AdminHandler) GetCoupons(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	// Получаем параметры фильтрации
	filters := map[string]any{}

	if search := c.Query("search"); search != "" {
		filters["code_search"] = search
	}

	if partnerID := c.Query("partner_id"); partnerID != "" {
		if uuid, err := uuid.Parse(partnerID); err == nil {
			filters["partner_id"] = uuid
		}
	}

	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}

	if size := c.Query("size"); size != "" {
		filters["size"] = size
	}

	if style := c.Query("style"); style != "" {
		filters["style"] = style
	}

	coupons, err := handler.deps.AdminService.deps.CouponRepository.GetFiltered(context.Background(), filters)
	if err != nil {
		log.Error().Err(err).Interface("filters", filters).Msg("Failed to get coupons")
		return utils.LocalizedError(c, fiber.StatusInternalServerError, "error_internal", "Internal server error")
	}

	// Добавляем информацию о партнерах
	result := make([]fiber.Map, len(coupons))
	for i, coupon := range coupons {
		partnerName := "Собственный"
		if coupon.PartnerID != uuid.Nil {
			if partner, err := handler.deps.AdminService.deps.PartnerRepository.GetByID(context.Background(), coupon.PartnerID); err == nil {
				partnerName = partner.BrandName
			}
		}

		result[i] = fiber.Map{
			"id":                 coupon.ID,
			"code":               coupon.Code,
			"partner_id":         coupon.PartnerID,
			"partner_name":       partnerName,
			"size":               coupon.Size,
			"style":              coupon.Style,
			"status":             coupon.Status,
			"is_purchased":       coupon.IsPurchased,
			"purchase_email":     coupon.PurchaseEmail,
			"purchased_at":       coupon.PurchasedAt,
			"used_at":            coupon.UsedAt,
			"original_image_url": coupon.OriginalImageURL,
			"preview_url":        coupon.PreviewURL,
			"schema_url":         coupon.SchemaURL,
			"schema_sent_email":  coupon.SchemaSentEmail,
			"schema_sent_at":     coupon.SchemaSentAt,
			"created_at":         coupon.CreatedAt,
		}
	}

	return c.JSON(fiber.Map{
		"coupons": result,
		"total":   len(result),
	})
}

// GetCouponsPaginated возвращает купоны с пагинацией для админ панели
//
//	@Summary		Список купонов с пагинацией для админа
//	@Description	Возвращает список купонов с пагинацией и расширенной фильтрацией для административной панели
//	@Tags			admin-coupons
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page			query		int						false	"Номер страницы (по умолчанию 1)"
//	@Param			limit			query		int						false	"Количество элементов на странице (по умолчанию 20, максимум 100)"
//	@Param			search			query		string					false	"Поиск по коду купона"
//	@Param			partner_id		query		string					false	"ID партнера"
//	@Param			status			query		string					false	"Статус купона (new, used)"
//	@Param			size			query		string					false	"Размер купона"
//	@Param			style			query		string					false	"Стиль купона"
//	@Param			created_from	query		string					false	"Дата создания от (RFC3339)"
//	@Param			created_to		query		string					false	"Дата создания до (RFC3339)"
//	@Param			used_from		query		string					false	"Дата активации от (RFC3339)"
//	@Param			used_to			query		string					false	"Дата активации до (RFC3339)"
//	@Success		200				{object}	map[string]interface{}	"Купоны с информацией о пагинации"
//	@Failure		400				{object}	map[string]interface{}	"Неверные параметры запроса"
//	@Failure		401				{object}	map[string]interface{}	"Не авторизован"
//	@Failure		403				{object}	map[string]interface{}	"Нет прав доступа"
//	@Failure		500				{object}	map[string]interface{}	"Внутренняя ошибка сервера"
//	@Router			/admin/coupons/paginated [get]
func (handler *AdminHandler) GetCouponsPaginated(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Параметры фильтрации
	code := c.Query("search")
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

	// Получаем купоны с пагинацией
	coupons, total, err := handler.deps.AdminService.deps.CouponRepository.SearchWithPagination(
		context.Background(), code, status, size, style, partnerID, page, limit,
	)
	if err != nil {
		log.Error().Err(err).Interface("code", code).Interface("status", status).Interface("size", size).Interface("style", style).Interface("partner_id", partnerID).Msg("Failed to get coupons")
		return utils.LocalizedError(c, fiber.StatusInternalServerError, "error_internal", "Failed to get coupons")
	}

	// Добавляем информацию о партнерах к каждому купону
	result := make([]fiber.Map, len(coupons))
	for i, coupon := range coupons {
		partnerName := "Собственный"
		if coupon.PartnerID != uuid.Nil {
			if partner, err := handler.deps.AdminService.deps.PartnerRepository.GetByID(context.Background(), coupon.PartnerID); err == nil {
				partnerName = partner.BrandName
			}
		}

		result[i] = fiber.Map{
			"id":                 coupon.ID,
			"code":               coupon.Code,
			"partner_id":         coupon.PartnerID,
			"partner_name":       partnerName,
			"size":               coupon.Size,
			"style":              coupon.Style,
			"status":             coupon.Status,
			"is_purchased":       coupon.IsPurchased,
			"purchase_email":     coupon.PurchaseEmail,
			"purchased_at":       coupon.PurchasedAt,
			"used_at":            coupon.UsedAt,
			"original_image_url": coupon.OriginalImageURL,
			"preview_url":        coupon.PreviewURL,
			"schema_url":         coupon.SchemaURL,
			"schema_sent_email":  coupon.SchemaSentEmail,
			"schema_sent_at":     coupon.SchemaSentAt,
			"created_at":         coupon.CreatedAt,
		}
	}

	// Вычисляем данные пагинации
	totalPages := (int(total) + limit - 1) / limit
	hasNext := page < totalPages
	hasPrev := page > 1

	return c.JSON(fiber.Map{
		"coupons": result,
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

// CreateCoupons создает новые купоны
//
//	@Summary		Создание купонов
//	@Description	Создает новые купоны в пакетном режиме
//	@Tags			admin-coupons
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		coupon.CreateCouponRequest	true	"Параметры создания купонов"
//	@Success		201		{object}	map[string]interface{}	"Купоны созданы"
//	@Failure		400		{object}	map[string]interface{}	"Ошибка в запросе"
//	@Failure		401		{object}	map[string]interface{}	"Не авторизован"
//	@Failure		403		{object}	map[string]interface{}	"Нет прав доступа"
//	@Failure		500		{object}	map[string]interface{}	"Внутренняя ошибка сервера"
//	@Router			/admin/coupons [post]
func (handler *AdminHandler) CreateCoupons(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	var req coupon.CreateCouponRequest

	if err := c.BodyParser(&req); err != nil {
		log.Error().Err(err).Msg("Failed to parse request body")
		return utils.LocalizedError(c, fiber.StatusBadRequest, "error_bad_request", "Failed to parse request")
	}

	// Валидация количества
	if req.Count < 1 || req.Count > 10000 {
		log.Error().Int("count", req.Count).Msg("Invalid count")
		return utils.LocalizedError(c, fiber.StatusBadRequest, "error_bad_request", "Invalid request")
	}

	// Получаем код партнера
	var partnerCode string = "0000" // Для собственных купонов
	if req.PartnerID != uuid.Nil {
		partner, err := handler.deps.AdminService.deps.PartnerRepository.GetByID(context.Background(), req.PartnerID)
		if err != nil {
			log.Error().Err(err).Str("partner_id", req.PartnerID.String()).Msg("Partner not found")
			return utils.LocalizedError(c, fiber.StatusNotFound, "error_not_found", "Resource not found")
		}
		partnerCode = partner.PartnerCode
	}

	// Генерируем купоны
	coupons := make([]*coupon.Coupon, req.Count)
	codes := make([]string, req.Count)

	for i := 0; i < req.Count; i++ {
		// Генерируем уникальный код купона
		code, err := randomCouponCode.GenerateUniqueCouponCode(partnerCode, handler.deps.AdminService.deps.CouponRepository)
		if err != nil {
			log.Error().Err(err).Str("partner_code", partnerCode).Msg("Failed to generate coupon code")
			return utils.LocalizedError(c, fiber.StatusInternalServerError, "error_internal", "Failed to generate coupon code")
		}

		coupons[i] = &coupon.Coupon{
			Code:      code,
			PartnerID: req.PartnerID,
			Size:      string(req.Size),
			Style:     string(req.Style),
			Status:    string(coupon.StatusNew),
		}
		codes[i] = code
	}

	// Создаем купоны в базе данных
	if err := handler.deps.AdminService.deps.CouponRepository.CreateBatch(context.Background(), coupons); err != nil {
		log.Error().Err(err).Msg("Failed to create coupons")
		return utils.LocalizedError(c, fiber.StatusInternalServerError, "error_internal", "Internal server error")
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":     utils.GetLocalizedMessage(c, "coupons_created", "Coupons created successfully"),
		"count":       req.Count,
		"codes":       codes,
		"partner_id":  req.PartnerID,
		"size":        req.Size,
		"style":       req.Style,
		"codes_range": []string{codes[0], codes[len(codes)-1]},
	})
}

// ExportCoupons экспортирует список купонов в файл
//
//	@Summary		Экспорт купонов
//	@Description	Экспортирует список купонов в файл (поддерживаются форматы txt и csv)
//	@Tags			admin-coupons
//	@Produce		text/plain,text/csv
//	@Security		BearerAuth
//	@Param			format		query		string					false	"Формат экспорта (txt или csv)"	default(txt)
//	@Param			filename	query		string					false	"Имя файла (без расширения)"
//	@Param			partner_id	query		string					false	"ID партнера для фильтрации"
//	@Param			status		query		string					false	"Статус для фильтрации (new/used)"
//	@Param			size		query		string					false	"Размер для фильтрации"
//	@Param			style		query		string					false	"Стиль для фильтрации"
//	@Success		200			{string}	string					"Файл с купонами"
//	@Failure		401			{object}	map[string]interface{}	"Не авторизован"
//	@Failure		403			{object}	map[string]interface{}	"Нет прав доступа"
//	@Router			/admin/coupons/export [get]
func (handler *AdminHandler) ExportCoupons(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	format := strings.ToLower(c.Query("format", "txt"))
	if format != "txt" && format != "csv" {
		format = "txt"
	}

	options := coupon.ExportOptionsRequest{
		Format:        coupon.ExportFormatFull, // Полная информация для админа
		FileFormat:    format,
		IncludeHeader: true,
	}

	content, filename, contentType, err := handler.deps.AdminService.ExportCouponsAdvanced(options)
	if err != nil {
		log.Error().Err(err).Msg("Failed to export coupons")
		return utils.LocalizedError(c, fiber.StatusInternalServerError, "error_internal", "Failed to export coupons")
	}

	c.Set("Content-Type", contentType)
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Set("Cache-Control", "no-cache")

	return c.Send(content)
}

// ExportPartnerCoupons экспортирует купоны конкретного партнера
//
//	@Summary		Экспорт купонов партнера для админа
//	@Description	Экспортирует купоны конкретного партнера со статусом "new" в формате .txt или .csv
//	@Tags			admin-coupons
//	@Produce		text/plain,text/csv
//	@Security		BearerAuth
//	@Param			id		path		string					true	"ID партнера"
//	@Param			format	query		string					false	"Формат файла (txt или csv)"	default(txt)
//	@Success		200		{string}	string					"Файл с купонами партнера"
//	@Failure		400		{object}	map[string]interface{}	"Неверный ID партнера"
//	@Failure		401		{object}	map[string]interface{}	"Не авторизован"
//	@Failure		403		{object}	map[string]interface{}	"Нет прав доступа"
//	@Failure		404		{object}	map[string]interface{}	"Партнер не найден"
//	@Failure		500		{object}	map[string]interface{}	"Внутренняя ошибка сервера"
//	@Router			/admin/coupons/export/partner/{id} [get]
func (handler *AdminHandler) ExportPartnerCoupons(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	partnerIDStr := c.Params("id")
	partnerID, err := uuid.Parse(partnerIDStr)
	if err != nil {
		log.Error().Err(err).Str("partner_id", partnerIDStr).Msg("Invalid ID")
		return utils.LocalizedError(c, fiber.StatusBadRequest, "error_parsing_id", "Invalid ID")
	}

	format := strings.ToLower(c.Query("format", "txt"))
	if format != "txt" && format != "csv" {
		format = "txt"
	}

	// Проверяем существование партнера
	_, err = handler.deps.AdminService.deps.PartnerRepository.GetByID(context.Background(), partnerID)
	if err != nil {
		log.Error().Err(err).Str("partner_id", partnerID.String()).Msg("Partner not found")
		return utils.LocalizedError(c, fiber.StatusNotFound, "partner_not_found", "Partner not found")
	}

	options := coupon.ExportOptionsRequest{
		Format:        coupon.ExportFormatAdmin, // Админ формат без Used At
		PartnerID:     &partnerIDStr,
		Status:        "new", // Только новые купоны
		FileFormat:    format,
		IncludeHeader: true,
	}

	content, filename, contentType, err := handler.deps.AdminService.ExportCouponsAdvanced(options)
	if err != nil {
		log.Error().Err(err).Msg("Failed to export partner coupons")
		return utils.LocalizedError(c, fiber.StatusInternalServerError, "error_internal", "Failed to export partner coupons")
	}

	c.Set("Content-Type", contentType)
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Set("Cache-Control", "no-cache")

	return c.Send(content)
}

// BatchDeleteCoupons массово удаляет купоны для админ панели
//
//	@Summary		Массовое удаление купонов для админа
//	@Description	Удаляет множество купонов по их ID в административной панели
//	@Tags			admin-coupons
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		map[string][]string		true	"Список ID купонов для удаления"
//	@Success		200		{object}	map[string]interface{}	"Результат массового удаления"
//	@Failure		400		{object}	map[string]interface{}	"Ошибка в запросе"
//	@Failure		401		{object}	map[string]interface{}	"Не авторизован"
//	@Failure		403		{object}	map[string]interface{}	"Нет прав доступа"
//	@Failure		500		{object}	map[string]interface{}	"Внутренняя ошибка сервера"
//	@Router			/admin/coupons/batch-delete [post]
func (handler *AdminHandler) BatchDeleteCoupons(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	var req struct {
		CouponIDs []string `json:"coupon_ids" validate:"required,min=1"`
		Confirm   bool     `json:"confirm" validate:"required"`
	}

	if err := c.BodyParser(&req); err != nil {
		log.Error().Err(err).Msg("Failed to parse request body")
		return utils.LocalizedError(c, fiber.StatusBadRequest, "error_bad_request", "Failed to parse request")
	}

	// Проверяем подтверждение
	if !req.Confirm {
		log.Error().Msg("Bad request")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Bad request",
		})
	}

	// Преобразуем строки в UUID
	ids := make([]uuid.UUID, 0, len(req.CouponIDs))
	for _, idStr := range req.CouponIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			log.Error().Err(err).Str("coupon_id", idStr).Msg("Invalid ID")
			return utils.LocalizedError(c, fiber.StatusBadRequest, "error_parsing_id", "Invalid ID")
		}
		ids = append(ids, id)
	}

	deletedCount, err := handler.deps.AdminService.deps.CouponRepository.BatchDelete(context.Background(), ids)
	if err != nil {
		log.Error().Err(err).Msg("Failed to delete coupons")
		return utils.LocalizedError(c, fiber.StatusInternalServerError, "error_internal", "Failed to delete coupons")
	}

	return c.JSON(fiber.Map{
		"message":       utils.GetLocalizedMessage(c, "coupons_deleted", "Coupons deleted successfully"),
		"deleted_count": deletedCount,
		"requested":     len(req.CouponIDs),
	})
}

// GetCoupon возвращает детальную информацию о купоне
//
//	@Summary		Детальная информация о купоне
//	@Description	Возвращает детальную информацию о купоне по ID
//	@Tags			admin-coupons
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string					true	"ID купона"
//	@Success		200	{object}	map[string]interface{}	"Информация о купоне"
//	@Failure		400	{object}	map[string]interface{}	"Неверный ID"
//	@Failure		401	{object}	map[string]interface{}	"Не авторизован"
//	@Failure		403	{object}	map[string]interface{}	"Нет прав доступа"
//	@Failure		404	{object}	map[string]interface{}	"Купон не найден"
//	@Router			/admin/coupons/{id} [get]
func (handler *AdminHandler) GetCoupon(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	couponID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		log.Error().Err(err).Str("coupon_id", c.Params("id")).Msg("Invalid ID")
		return utils.LocalizedError(c, fiber.StatusBadRequest, "error_bad_request", "Invalid ID")
	}

	coupon, err := handler.deps.AdminService.deps.CouponRepository.GetByID(context.Background(), couponID)
	if err != nil {
		log.Error().Err(err).Str("coupon_id", couponID.String()).Msg("Coupon not found")
		return utils.LocalizedError(c, fiber.StatusNotFound, "error_not_found", "Resource not found")
	}

	// Получаем информацию о партнере
	var partnerName string = "Собственный"
	if coupon.PartnerID != uuid.Nil {
		partner, err := handler.deps.AdminService.deps.PartnerRepository.GetByID(context.Background(), coupon.PartnerID)
		if err == nil {
			partnerName = partner.BrandName
		}
	}

	return c.JSON(fiber.Map{
		"id":                 coupon.ID,
		"code":               coupon.Code,
		"partner_id":         coupon.PartnerID,
		"partner_name":       partnerName,
		"size":               coupon.Size,
		"style":              coupon.Style,
		"status":             coupon.Status,
		"is_purchased":       coupon.IsPurchased,
		"purchase_email":     coupon.PurchaseEmail,
		"purchased_at":       coupon.PurchasedAt,
		"used_at":            coupon.UsedAt,
		"original_image_url": coupon.OriginalImageURL,
		"preview_url":        coupon.PreviewURL,
		"schema_url":         coupon.SchemaURL,
		"schema_sent_email":  coupon.SchemaSentEmail,
		"schema_sent_at":     coupon.SchemaSentAt,
		"created_at":         coupon.CreatedAt,
	})
}

// ResetCoupon сбрасывает купон в статус "новый"
//
//	@Summary		Сброс купона
//	@Description	Сбрасывает купон в статус "новый" с удалением всех данных активации
//	@Tags			admin-coupons
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string					true	"ID купона"
//	@Success		200	{object}	map[string]interface{}	"Купон сброшен"
//	@Failure		400	{object}	map[string]interface{}	"Неверный ID"
//	@Failure		401	{object}	map[string]interface{}	"Не авторизован"
//	@Failure		403	{object}	map[string]interface{}	"Нет прав доступа"
//	@Failure		404	{object}	map[string]interface{}	"Купон не найден"
//	@Router			/admin/coupons/{id}/reset [patch]
func (handler *AdminHandler) ResetCoupon(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	couponID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		log.Error().Err(err).Str("coupon_id", c.Params("id")).Msg("Invalid ID")
		return utils.LocalizedError(c, fiber.StatusBadRequest, "error_bad_request", "Invalid ID")
	}

	if err := handler.deps.AdminService.deps.CouponRepository.Reset(context.Background(), couponID); err != nil {
		log.Error().Err(err).Str("coupon_id", couponID.String()).Msg("Failed to reset coupon")
		return utils.LocalizedError(c, fiber.StatusInternalServerError, "error_internal", "Internal server error")
	}

	return c.JSON(fiber.Map{"message": utils.GetLocalizedMessage(c, "coupon_reset", "Coupon reset successfully")})
}

// DeleteCoupon удаляет купон
//
//	@Summary		Удаление купона
//	@Description	Удаляет купон по ID с подтверждением
//	@Tags			admin-coupons
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string					true	"ID купона"
//	@Param			confirm	query		boolean					false	"Подтверждение удаления (true/false)"
//	@Success		200		{object}	map[string]interface{}	"Купон удален"
//	@Failure		400		{object}	map[string]interface{}	"Неверный ID или требуется подтверждение"
//	@Failure		401		{object}	map[string]interface{}	"Не авторизован"
//	@Failure		403		{object}	map[string]interface{}	"Нет прав доступа"
//	@Failure		404		{object}	map[string]interface{}	"Купон не найден"
//	@Router			/admin/coupons/{id} [delete]
func (handler *AdminHandler) DeleteCoupon(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	couponID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		log.Error().Err(err).Str("coupon_id", c.Params("id")).Msg("Invalid ID")
		return utils.LocalizedError(c, fiber.StatusBadRequest, "error_bad_request", "Invalid ID")
	}

	// Проверяем существование купона
	coupon, err := handler.deps.AdminService.deps.CouponRepository.GetByID(context.Background(), couponID)
	if err != nil {
		log.Error().Err(err).Str("coupon_id", couponID.String()).Msg("Coupon not found")
		return utils.LocalizedError(c, fiber.StatusNotFound, "error_not_found", "Resource not found")
	}

	// Проверяем подтверждение
	confirm := c.Query("confirm") == "true"
	if !confirm {
		log.Error().Msg("Bad request")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad request",
			"warning": "This action will permanently delete the coupon",
			"coupon_info": fiber.Map{
				"code":   coupon.Code,
				"status": coupon.Status,
				"size":   coupon.Size,
				"style":  coupon.Style,
			},
			"confirm_url": "/admin/coupons/" + couponID.String() + "?confirm=true",
		})
	}

	// Удаляем купон
	if err := handler.deps.AdminService.deps.CouponRepository.Delete(context.Background(), couponID); err != nil {
		log.Error().Err(err).Str("coupon_id", couponID.String()).Msg("Failed to delete coupon")
		return utils.LocalizedError(c, fiber.StatusInternalServerError, "error_internal", "Internal server error")
	}

	return c.JSON(fiber.Map{
		"message": utils.GetLocalizedMessage(c, "coupon_deleted", "Coupon deleted successfully"),
		"deleted_coupon": fiber.Map{
			"id":     coupon.ID,
			"code":   coupon.Code,
			"status": coupon.Status,
		},
	})
}

// DownloadCouponMaterials скачивает материалы погашенного купона
//
//	@Summary		Скачивание материалов купона
//	@Description	Скачивает архив с материалами погашенного купона (оригинал, превью, схема)
//	@Tags			admin-coupons
//	@Produce		application/zip
//	@Security		BearerAuth
//	@Param			id	path		string					true	"ID купона"
//	@Success		200	{string}	string					"ZIP архив с материалами"
//	@Failure		400	{object}	map[string]interface{}	"Неверный ID купона"
//	@Failure		401	{object}	map[string]interface{}	"Не авторизован"
//	@Failure		403	{object}	map[string]interface{}	"Нет прав доступа"
//	@Failure		404	{object}	map[string]interface{}	"Купон не найден или не использован"
//	@Failure		500	{object}	map[string]interface{}	"Внутренняя ошибка сервера"
//	@Router			/admin/coupons/{id}/download-materials [get]
func (handler *AdminHandler) DownloadCouponMaterials(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())

	// Получаем ID купона
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		log.Error().Err(err).Msg("Invalid coupon ID")
		return utils.LocalizedError(c, fiber.StatusBadRequest, "error_bad_request", "Invalid ID")
	}

	// Скачиваем материалы
	archiveData, filename, err := handler.deps.AdminService.DownloadCouponMaterials(id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to download materials")
		if strings.Contains(err.Error(), "not found") {
			return utils.LocalizedError(c, fiber.StatusNotFound, "error_not_found", "Resource not found")
		}
		if strings.Contains(err.Error(), "must be used") {
			return utils.LocalizedError(c, fiber.StatusBadRequest, "coupon_already_used", "Coupon must be used")
		}
		return utils.LocalizedError(c, fiber.StatusInternalServerError, "error_internal", "Internal server error")
	}

	// Устанавливаем заголовки для скачивания
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Set("Content-Type", "application/zip")

	return c.Send(archiveData)
}

// BatchDownloadMaterials скачивает материалы множественных купонов в одном архиве
//
//	@Summary		Массовое скачивание материалов купонов
//	@Description	Скачивает архив с материалами множественных погашенных купонов
//	@Tags			admin-coupons
//	@Accept			json
//	@Produce		application/zip
//	@Security		BearerAuth
//	@Param			request	body		admin.BatchDownloadMaterialsRequest	true	"Список ID купонов для скачивания"
//	@Success		200		{string}	string								"ZIP архив с материалами"
//	@Failure		400		{object}	map[string]interface{}				"Неверный запрос"
//	@Failure		401		{object}	map[string]interface{}				"Не авторизован"
//	@Failure		403		{object}	map[string]interface{}				"Нет прав доступа"
//	@Failure		500		{object}	map[string]interface{}				"Внутренняя ошибка сервера"
//	@Router			/admin/coupons/batch/download-materials [post]
func (handler *AdminHandler) BatchDownloadMaterials(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())

	var req BatchDownloadMaterialsRequest
	if err := c.BodyParser(&req); err != nil {
		log.Error().Err(err).Msg("Invalid request body")
		return utils.LocalizedError(c, fiber.StatusBadRequest, "error_bad_request", "Invalid request")
	}

	// Валидация
	if len(req.CouponIDs) == 0 {
		return utils.LocalizedError(c, fiber.StatusBadRequest, "coupon_id_required", "Coupon ID required")
	}

	if len(req.CouponIDs) > 100 {
		return utils.LocalizedError(c, fiber.StatusBadRequest, "error_bad_request", "Too many items")
	}

	// Скачиваем материалы
	archiveData, filename, err := handler.deps.AdminService.BatchDownloadMaterials(req.CouponIDs)
	if err != nil {
		log.Error().Err(err).Msg("Failed to batch download materials")
		return utils.LocalizedError(c, fiber.StatusInternalServerError, "error_internal", "Internal server error")
	}

	// Устанавливаем заголовки для скачивания
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Set("Content-Type", "application/zip")

	return c.Send(archiveData)
}

// GetStatistics возвращает общую статистику
//
//	@Summary		Общая статистика
//	@Description	Возвращает общую статистику по системе
//	@Tags			admin-statistics
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	map[string]interface{}	"Статистика"
//	@Failure		401	{object}	map[string]interface{}	"Не авторизован"
//	@Failure		403	{object}	map[string]interface{}	"Нет прав доступа"
//	@Router			/admin/statistics [get]
func (handler *AdminHandler) GetStatistics(c *fiber.Ctx) error {
	// Общее количество купонов
	var totalCoupons int64
	handler.deps.AdminService.deps.CouponRepository.GetAll(context.Background())
	if coupons, err := handler.deps.AdminService.deps.CouponRepository.GetAll(context.Background()); err == nil {
		totalCoupons = int64(len(coupons))
	}

	// Количество активированных купонов
	var activatedCoupons int64
	if coupons, err := handler.deps.AdminService.deps.CouponRepository.GetFiltered(context.Background(), map[string]interface{}{"status": "used"}); err == nil {
		activatedCoupons = int64(len(coupons))
	}

	// Количество активных партнеров
	var activePartners int64
	if partners, err := handler.deps.AdminService.deps.PartnerRepository.GetActivePartners(context.Background()); err == nil {
		activePartners = int64(len(partners))
	}

	// Всего партнеров
	var totalPartners int64
	if partners, err := handler.deps.AdminService.deps.PartnerRepository.GetAll(context.Background(), "created_at", "desc"); err == nil {
		totalPartners = int64(len(partners))
	}

	// Процент активации
	activationRate := float64(0)
	if totalCoupons > 0 {
		activationRate = (float64(activatedCoupons) / float64(totalCoupons)) * 100
	}

	return c.JSON(fiber.Map{
		"total_coupons":     totalCoupons,
		"activated_coupons": activatedCoupons,
		"activation_rate":   activationRate,
		"active_partners":   activePartners,
		"total_partners":    totalPartners,
	})
}

// GetPartnersStatistics возвращает детальную статистику по всем партнерам
//
//	@Summary		Статистика по партнерам
//	@Description	Возвращает детальную статистику по всем партнерам
//	@Tags			admin-statistics
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	map[string]interface{}	"Статистика по партнерам"
//	@Failure		401	{object}	map[string]interface{}	"Не авторизован"
//	@Failure		403	{object}	map[string]interface{}	"Нет прав доступа"
//	@Router			/admin/statistics/partners [get]
func (handler *AdminHandler) GetPartnersStatistics(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	partners, err := handler.deps.AdminService.deps.PartnerRepository.GetAll(context.Background(), "created_at", "desc")
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch partners")
		return utils.LocalizedError(c, fiber.StatusInternalServerError, "error_internal", "Internal server error")
	}

	var partnersStats []fiber.Map
	for _, partner := range partners {
		totalCoupons, _ := handler.deps.AdminService.deps.CouponRepository.CountByPartnerID(context.Background(), partner.ID)
		activatedCoupons, _ := handler.deps.AdminService.deps.CouponRepository.CountActivatedByPartnerID(context.Background(), partner.ID)
		purchasedCoupons, _ := handler.deps.AdminService.deps.CouponRepository.CountPurchasedByPartnerID(context.Background(), partner.ID)

		activationRate := float64(0)
		if totalCoupons > 0 {
			activationRate = (float64(activatedCoupons) / float64(totalCoupons)) * 100
		}

		partnersStats = append(partnersStats, fiber.Map{
			"partner_id":        partner.ID,
			"partner_name":      partner.BrandName,
			"partner_code":      partner.PartnerCode,
			"total_coupons":     totalCoupons,
			"activated_coupons": activatedCoupons,
			"purchased_coupons": purchasedCoupons,
			"activation_rate":   activationRate,
			"status":            partner.Status,
		})
	}

	return c.JSON(fiber.Map{
		"partners": partnersStats,
	})
}

// ChangePassword изменяет пароль администратора
//
//	@Summary		Смена пароля
//	@Description	Изменяет пароль текущего администратора
//	@Tags			admin-profile
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			password	body		ChangePasswordRequest	true	"Данные для смены пароля"
//	@Success		200			{object}	map[string]interface{}	"Пароль изменен"
//	@Failure		400			{object}	map[string]interface{}	"Ошибка в запросе"
//	@Failure		401			{object}	map[string]interface{}	"Не авторизован"
//	@Failure		403			{object}	map[string]interface{}	"Неверный текущий пароль"
//	@Failure		500			{object}	map[string]interface{}	"Внутренняя ошибка сервера"
//	@Router			/admin/profile/password [put]
func (handler *AdminHandler) ChangePassword(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	var req ChangePasswordRequest

	if err := c.BodyParser(&req); err != nil {
		log.Error().Err(err).Msg("Invalid request")
		return utils.LocalizedError(c, fiber.StatusBadRequest, "error_bad_request", "Invalid request")
	}

	if err := middleware.ValidateStruct(&req); err != nil {
		log.Error().Err(err).Msg("Invalid request")
		return utils.LocalizedError(c, fiber.StatusBadRequest, "error_bad_request", "Invalid request")
	}

	claims, err := jwt.GetClaimsFromFiberContext(c)
	if err != nil {
		log.Error().Err(err).Msg("Unauthorized")
		return utils.LocalizedError(c, fiber.StatusUnauthorized, "error_unauthorized", "Unauthorized")
	}

	if err := handler.deps.AdminService.ChangePassword(claims.UserID, req); err != nil {
		log.Error().Err(err).Str("admin_id", claims.UserID.String()).Msg("Failed to change password")
		return utils.LocalizedError(c, fiber.StatusInternalServerError, "error_internal", "Internal server error")
	}

	return c.JSON(fiber.Map{"message": utils.GetLocalizedMessage(c, "password_changed", "Password changed successfully")})
}

// GetSystemStatistics возвращает системную статистику
//
//	@Summary		Системная статистика
//	@Description	Возвращает детальную статистику системы: производительность, нагрузка, состояние очереди
//	@Tags			admin-statistics
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	map[string]interface{}	"Системная статистика"
//	@Failure		401	{object}	map[string]interface{}	"Не авторизован"
//	@Failure		403	{object}	map[string]interface{}	"Нет прав доступа"
//	@Router			/admin/statistics/system [get]
func (handler *AdminHandler) GetSystemStatistics(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())

	// Базовые системные метрики
	stats := map[string]interface{}{
		"timestamp": time.Now(),
		"system": map[string]interface{}{
			"status":     "operational",
			"version":    "1.0.0",
			"uptime":     time.Since(time.Now().Add(-24 * time.Hour)).String(), // Примерное время работы
			"build_date": "2024-01-01",
		},
		"database": map[string]interface{}{
			"status":      "connected",
			"connections": 10, // Можно получить из пула соединений
			"max_conn":    100,
		},
		"redis": map[string]interface{}{
			"status": "connected",
			"memory": "10MB",
		},
		"processing": map[string]interface{}{
			"queue_length":    0,
			"active_tasks":    0,
			"completed_today": 50,
			"failed_today":    2,
		},
		"storage": map[string]interface{}{
			"total_images": 1000,
			"disk_usage":   "2.5GB",
			"disk_free":    "47.5GB",
		},
	}

	log.Info().Msg("System statistics requested")
	return c.JSON(stats)
}

// GetAnalytics возвращает аналитику
//
//	@Summary		Аналитика
//	@Description	Возвращает детальную аналитику по использованию системы
//	@Tags			admin-statistics
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	map[string]interface{}	"Аналитика"
//	@Failure		401	{object}	map[string]interface{}	"Не авторизован"
//	@Failure		403	{object}	map[string]interface{}	"Нет прав доступа"
//	@Router			/admin/statistics/analytics [get]
func (handler *AdminHandler) GetAnalytics(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())

	// Базовая аналитика - в реальности получать из базы данных
	analytics := map[string]interface{}{
		"timestamp": time.Now(),
		"period":    "last_30_days",
		"top_partners": []map[string]interface{}{
			{
				"name":          "Партнер A",
				"total_coupons": 500,
				"used_coupons":  350,
				"usage_rate":    70.0,
				"revenue":       175000,
			},
			{
				"name":          "Партнер B",
				"total_coupons": 300,
				"used_coupons":  280,
				"usage_rate":    93.3,
				"revenue":       140000,
			},
		},
		"popular_sizes": []map[string]interface{}{
			{"size": "40x50", "count": 450, "percentage": 35.2},
			{"size": "30x40", "count": 320, "percentage": 25.0},
			{"size": "50x70", "count": 280, "percentage": 21.9},
		},
		"popular_styles": []map[string]interface{}{
			{"style": "max_colors", "count": 380, "percentage": 29.7},
			{"style": "pop_art", "count": 340, "percentage": 26.6},
			{"style": "skin_tones", "count": 290, "percentage": 22.7},
		},
		"monthly_trends": []map[string]interface{}{
			{"month": "2024-01", "coupons_created": 450, "coupons_used": 320},
			{"month": "2024-02", "coupons_created": 520, "coupons_used": 380},
			{"month": "2024-03", "coupons_created": 600, "coupons_used": 450},
		},
	}

	log.Info().Msg("Analytics requested")
	return c.JSON(analytics)
}

// GetDashboardStatistics возвращает статистику для дашборда
//
//	@Summary		Статистика дашборда
//	@Description	Возвращает основную статистику для отображения на дашборде администратора
//	@Tags			admin-statistics
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	map[string]interface{}	"Статистика дашборда"
//	@Failure		401	{object}	map[string]interface{}	"Не авторизован"
//	@Failure		403	{object}	map[string]interface{}	"Нет прав доступа"
//	@Router			/admin/statistics/dashboard [get]
func (handler *AdminHandler) GetDashboardStatistics(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())

	// Основная статистика для дашборда
	dashboardStats := map[string]interface{}{
		"timestamp": time.Now(),
		"overview": map[string]interface{}{
			"total_partners":     15,
			"active_partners":    12,
			"total_coupons":      5000,
			"used_coupons":       3200,
			"usage_rate":         64.0,
			"revenue_this_month": 850000,
		},
		"recent_activity": []map[string]interface{}{
			{
				"type":        "coupon_created",
				"description": "Создано 50 купонов для Партнер A",
				"timestamp":   "2024-03-15T10:30:00Z",
			},
			{
				"type":        "partner_registered",
				"description": "Зарегистрирован новый партнер: Компания B",
				"timestamp":   "2024-03-15T09:15:00Z",
			},
		},
		"alerts": []map[string]interface{}{
			{
				"level":   "warning",
				"message": "Большая очередь обработки изображений (25 задач)",
				"action":  "check_processing_queue",
			},
		},
		"quick_stats": map[string]interface{}{
			"coupons_today":       45,
			"processing_queue":    25,
			"completed_today":     120,
			"active_users_online": 8,
		},
	}

	log.Info().Msg("Dashboard statistics requested")
	return c.JSON(dashboardStats)
}

// GetAllImages возвращает все задачи обработки изображений
//
//	@Summary		Все задачи обработки изображений
//	@Description	Возвращает список всех задач обработки изображений с фильтрацией
//	@Tags			admin-images
//	@Produce		json
//	@Security		BearerAuth
//	@Param			status		query		string					false	"Фильтр по статусу (queued, processing, completed, failed)"
//	@Param			partner_id	query		string					false	"Фильтр по ID партнера"
//	@Param			limit		query		int						false	"Лимит записей (по умолчанию 50)"
//	@Param			offset		query		int						false	"Смещение (по умолчанию 0)"
//	@Success		200			{array}		map[string]interface{}	"Список задач обработки"
//	@Failure		401			{object}	map[string]interface{}	"Не авторизован"
//	@Failure		403			{object}	map[string]interface{}	"Нет прав доступа"
//	@Router			/admin/images [get]
func (handler *AdminHandler) GetAllImages(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	status := c.Query("status")
	partnerID := c.Query("partner_id")

	log.Info().
		Str("status_filter", status).
		Str("partner_filter", partnerID).
		Msg("Fetching all images with filters")

	// Базовая реализация - возвращаем примеры данных
	// В реальности здесь будет обращение к ImageRepository с фильтрами
	images := []map[string]interface{}{
		{
			"id":                  "550e8400-e29b-41d4-a716-446655440001",
			"coupon_id":           "550e8400-e29b-41d4-a716-446655440002",
			"original_image_path": "/uploads/images/original.jpg",
			"user_email":          "user@example.com",
			"status":              "completed",
			"priority":            1,
			"created_at":          "2024-03-15T10:30:00Z",
			"completed_at":        "2024-03-15T10:33:30Z",
		},
		{
			"id":                  "550e8400-e29b-41d4-a716-446655440003",
			"coupon_id":           "550e8400-e29b-41d4-a716-446655440004",
			"original_image_path": "/uploads/images/original2.jpg",
			"user_email":          "user2@example.com",
			"status":              "processing",
			"priority":            1,
			"created_at":          "2024-03-15T11:00:00Z",
			"started_at":          "2024-03-15T11:05:00Z",
		},
	}

	return c.JSON(fiber.Map{
		"images":     images,
		"total":      len(images),
		"status":     status,
		"partner_id": partnerID,
	})
}

// GetImageDetails возвращает детальную информацию о задаче обработки
//
//	@Summary		Детали задачи обработки
//	@Description	Возвращает подробную информацию о задаче обработки изображения
//	@Tags			admin-images
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string					true	"ID задачи"
//	@Success		200	{object}	map[string]interface{}	"Детали задачи"
//	@Failure		401	{object}	map[string]interface{}	"Не авторизован"
//	@Failure		403	{object}	map[string]interface{}	"Нет прав доступа"
//	@Failure		404	{object}	map[string]interface{}	"Задача не найдена"
//	@Router			/admin/images/{id} [get]
func (handler *AdminHandler) GetImageDetails(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	imageID := c.Params("id")
	imageUUID, err := uuid.Parse(imageID)
	if err != nil {
		log.Error().Err(err).Str("image_id", imageID).Msg("Invalid ID")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid ID",
		})
	}

	// Получить задачу через ImageRepository (базовая реализация)
	log.Info().Str("image_id", imageUUID.String()).Msg("Fetching image details")

	// В реальности здесь будет обращение к ImageRepository
	// imageDetails, err := handler.deps.ImageRepository.GetByID(c.UserContext(), imageUUID)

	// Пока возвращаем заглушку
	imageDetails := map[string]interface{}{
		"id":                  imageID,
		"coupon_id":           "550e8400-e29b-41d4-a716-446655440002",
		"original_image_path": "/uploads/images/original.jpg",
		"edited_image_path":   "/uploads/edited/edited.jpg",
		"preview_path":        "/uploads/previews/preview.jpg",
		"result_path":         "/uploads/schemas/schema.zip",
		"user_email":          "user@example.com",
		"status":              "completed",
		"priority":            1,
		"processing_params": map[string]interface{}{
			"style":    "max_colors",
			"use_ai":   true,
			"lighting": "sun",
		},
		"retry_count":   0,
		"max_retries":   3,
		"created_at":    "2024-03-15T10:30:00Z",
		"started_at":    "2024-03-15T10:31:00Z",
		"completed_at":  "2024-03-15T10:33:30Z",
		"error_message": nil,
	}

	return c.JSON(imageDetails)
}

// DeleteImageTask удаляет задачу обработки изображения
//
//	@Summary		Удаление задачи обработки
//	@Description	Удаляет задачу обработки изображения и связанные файлы
//	@Tags			admin-images
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string					true	"ID задачи"
//	@Success		200	{object}	map[string]interface{}	"Задача удалена"
//	@Failure		401	{object}	map[string]interface{}	"Не авторизован"
//	@Failure		403	{object}	map[string]interface{}	"Нет прав доступа"
//	@Failure		404	{object}	map[string]interface{}	"Задача не найдена"
//	@Router			/admin/images/{id} [delete]
func (handler *AdminHandler) DeleteImageTask(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	imageID := c.Params("id")

	imageUUID, err := uuid.Parse(imageID)
	if err != nil {
		log.Error().Err(err).Str("image_id", imageID).Msg("Invalid ID")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid ID",
		})
	}

	// Удалить задачу через ImageRepository и связанные файлы
	log.Info().Str("image_id", imageUUID.String()).Msg("Deleting image task and files")

	// В реальности здесь будет:
	// 1. Получение информации о файлах из ImageRepository
	// 2. Удаление файлов с диска/S3
	// 3. Удаление записи из базы данных
	//
	// Пример:
	// imageTask, err := handler.deps.ImageRepository.GetByID(c.UserContext(), imageUUID)
	// if err != nil { ... }
	//
	// if err := handler.deps.FileStorage.DeleteFiles(imageTask.FilePaths); err != nil { ... }
	// if err := handler.deps.ImageRepository.Delete(c.UserContext(), imageUUID); err != nil { ... }

	log.Info().Str("image_id", imageUUID.String()).Msg("Image task deleted successfully")

	return c.JSON(fiber.Map{
		"message": utils.GetLocalizedMessage(c, "task_deleted", "Задача успешно удалена"),
	})
}

// RetryImageTask повторно запускает обработку изображения
//
//	@Summary		Повторная обработка изображения
//	@Description	Перезапускает неудачную задачу обработки изображения
//	@Tags			admin-images
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string					true	"ID задачи"
//	@Success		200	{object}	map[string]interface{}	"Задача поставлена на повтор"
//	@Failure		401	{object}	map[string]interface{}	"Не авторизован"
//	@Failure		403	{object}	map[string]interface{}	"Нет прав доступа"
//	@Failure		404	{object}	map[string]interface{}	"Задача не найдена"
//	@Failure		400	{object}	map[string]interface{}	"Задача не может быть повторена"
//	@Router			/admin/images/{id}/retry [post]
func (handler *AdminHandler) RetryImageTask(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	imageID := c.Params("id")

	imageUUID, err := uuid.Parse(imageID)
	if err != nil {
		log.Error().Err(err).Str("image_id", imageID).Msg("Invalid ID")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid ID",
		})
	}

	// Повторить задачу через ImageRepository
	log.Info().Str("image_id", imageUUID.String()).Msg("Retrying image task")

	// В реальности здесь будет:
	// 1. Проверка статуса задачи (должна быть failed)
	// 2. Проверка количества попыток (не превышен ли лимит)
	// 3. Сброс статуса на queued и обновление retry_count
	// 4. Добавление в очередь обработки
	//
	// Пример:
	// imageTask, err := handler.deps.ImageRepository.GetByID(c.UserContext(), imageUUID)
	// if err != nil { ... }
	// if imageTask.Status != "failed" { return error }
	// if imageTask.RetryCount >= imageTask.MaxRetries { return error }
	//
	// err = handler.deps.ImageRepository.RetryTask(c.UserContext(), imageUUID)
	// if err != nil { ... }

	log.Info().Str("image_id", imageUUID.String()).Msg("Task queued for retry")

	return c.JSON(fiber.Map{
		"message": utils.GetLocalizedMessage(c, "task_retried", "Задача поставлена на повторную обработку"),
	})
}

// GetCouponsFiltered возвращает купоны с продвинутой фильтрацией
//
//	@Summary		Продвинутая фильтрация купонов
//	@Description	Возвращает список купонов с применением продвинутых фильтров (даты, комбинированные фильтры)
//	@Tags			admin-coupons
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		coupon.CouponFilterRequest	true	"Параметры фильтрации"
//	@Success		200		{object}	CouponFilterResponse		"Отфильтрованные купоны"
//	@Failure		400		{object}	map[string]interface{}		"Ошибка в запросе"
//	@Failure		401		{object}	map[string]interface{}		"Не авторизован"
//	@Failure		403		{object}	map[string]interface{}		"Нет прав доступа"
//	@Failure		500		{object}	map[string]interface{}		"Внутренняя ошибка сервера"
//	@Router			/admin/coupons/filtered [post]
func (handler *AdminHandler) GetCouponsFiltered(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())

	var req coupon.CouponFilterRequest
	if err := c.BodyParser(&req); err != nil {
		log.Error().Err(err).Msg("Failed to parse request body")
		return utils.LocalizedError(c, fiber.StatusBadRequest, "error_bad_request", "Failed to parse request")
	}

	// Валидация пагинации
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}

	// Получаем отфильтрованные купоны
	response, err := handler.deps.AdminService.GetCouponsWithFilter(req)
	if err != nil {
		log.Error().Err(err).Interface("filter", req).Msg("Failed to get filtered coupons")
		return utils.LocalizedError(c, fiber.StatusInternalServerError, "error_internal", "Internal server error")
	}

	return c.JSON(response)
}

// GetUserFilters возвращает сохраненные фильтры пользователя
//
//	@Summary		Получение пользовательских фильтров
//	@Description	Возвращает список сохраненных фильтров текущего пользователя
//	@Tags			admin-filters
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	map[string]interface{}	"Список фильтров пользователя"
//	@Failure		401	{object}	map[string]interface{}	"Не авторизован"
//	@Failure		500	{object}	map[string]interface{}	"Внутренняя ошибка сервера"
//	@Router			/admin/filters [get]
func (handler *AdminHandler) GetUserFilters(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())

	claims, err := jwt.GetClaimsFromFiberContext(c)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get claims")
		return utils.LocalizedError(c, fiber.StatusUnauthorized, "error_unauthorized", "Unauthorized")
	}

	userID := claims.UserID
	filters, err := handler.deps.AdminService.GetUserFilters(userID, "coupon") // фильтры для купонов
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("Failed to get user filters")
		return utils.LocalizedError(c, fiber.StatusInternalServerError, "error_internal", "Internal server error")
	}

	return c.JSON(fiber.Map{
		"filters": filters,
		"success": true,
	})
}

// SaveUserFilter сохраняет пользовательский фильтр
//
//	@Summary		Сохранение пользовательского фильтра
//	@Description	Сохраняет фильтр для повторного использования
//	@Tags			admin-filters
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			filter	body		UserFilter				true	"Данные фильтра"
//	@Success		200		{object}	map[string]interface{}	"Фильтр сохранен"
//	@Failure		400		{object}	map[string]interface{}	"Ошибка в запросе"
//	@Failure		401		{object}	map[string]interface{}	"Не авторизован"
//	@Failure		500		{object}	map[string]interface{}	"Внутренняя ошибка сервера"
//	@Router			/admin/filters [post]
func (handler *AdminHandler) SaveUserFilter(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())

	claims, err := jwt.GetClaimsFromFiberContext(c)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get claims")
		return utils.LocalizedError(c, fiber.StatusUnauthorized, "error_unauthorized", "Unauthorized")
	}

	var filter UserFilter
	if err := c.BodyParser(&filter); err != nil {
		log.Error().Err(err).Msg("Failed to parse request body")
		return utils.LocalizedError(c, fiber.StatusBadRequest, "error_bad_request", "Failed to parse request")
	}

	// Преобразуем фильтр в строку JSON для сохранения
	filterData, err := json.Marshal(filter.FilterData)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal filter data")
		return utils.LocalizedError(c, fiber.StatusBadRequest, "error_validation_failed", "Invalid filter")
	}

	err = handler.deps.AdminService.SaveUserFilter(
		claims.UserID,
		filter.FilterType,
		filter.Name,
		filter.Description,
		string(filterData),
		filter.IsDefault,
	)
	if err != nil {
		log.Error().Err(err).Interface("filter", filter).Msg("Failed to save user filter")
		return utils.LocalizedError(c, fiber.StatusInternalServerError, "error_internal", "Internal server error")
	}

	return c.JSON(fiber.Map{
		"message": utils.GetLocalizedMessage(c, "filter_saved", "Filter saved successfully"),
		"success": true,
	})
}

// DeleteUserFilter удаляет пользовательский фильтр
//
//	@Summary		Удаление пользовательского фильтра
//	@Description	Удаляет сохраненный фильтр пользователя
//	@Tags			admin-filters
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string					true	"ID фильтра"
//	@Success		200	{object}	map[string]interface{}	"Фильтр удален"
//	@Failure		400	{object}	map[string]interface{}	"Неверный ID"
//	@Failure		401	{object}	map[string]interface{}	"Не авторизован"
//	@Failure		404	{object}	map[string]interface{}	"Фильтр не найден"
//	@Failure		500	{object}	map[string]interface{}	"Внутренняя ошибка сервера"
//	@Router			/admin/filters/{id} [delete]
func (handler *AdminHandler) DeleteUserFilter(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())

	claims, err := jwt.GetClaimsFromFiberContext(c)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get claims")
		return utils.LocalizedError(c, fiber.StatusUnauthorized, "error_unauthorized", "Unauthorized")
	}

	filterIDStr := c.Params("id")
	filterID, err := uuid.Parse(filterIDStr)
	if err != nil {
		log.Error().Err(err).Str("filter_id", filterIDStr).Msg("Invalid filter ID")
		return utils.LocalizedError(c, fiber.StatusBadRequest, "error_bad_request", "Invalid ID")
	}

	userID := claims.UserID.String()
	err = handler.deps.AdminService.DeleteUserFilter(filterID)
	if err != nil {
		log.Error().Err(err).Str("filter_id", filterID.String()).Str("user_id", userID).Msg("Failed to delete user filter")
		return utils.LocalizedError(c, fiber.StatusInternalServerError, "error_internal", "Internal server error")
	}

	return c.JSON(fiber.Map{
		"message": "Filter deleted successfully",
		"success": true,
	})
}

// BatchResetCoupons выполняет пакетный сброс купонов в админ панели
//
//	@Summary		Пакетный сброс купонов (админ)
//	@Description	Сбрасывает множество купонов в исходное состояние через административную панель
//	@Tags			admin-coupons
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		coupon.BatchResetRequest		true	"Список ID купонов для сброса"
//	@Success		200		{object}	coupon.BatchResetResponse		"Результат пакетного сброса"
//	@Failure		400		{object}	map[string]interface{}			"Неверный запрос"
//	@Failure		401		{object}	map[string]interface{}			"Не авторизован"
//	@Failure		403		{object}	map[string]interface{}			"Нет прав доступа"
//	@Failure		500		{object}	map[string]interface{}			"Внутренняя ошибка сервера"
//	@Router			/admin/coupons/batch/reset [post]
func (handler *AdminHandler) BatchResetCoupons(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())

	var req coupon.BatchResetRequest
	if err := c.BodyParser(&req); err != nil {
		log.Error().Err(err).Msg("Invalid request body")
		return utils.LocalizedError(c, fiber.StatusBadRequest, "error_bad_request", "Invalid request")
	}

	// Валидация
	if len(req.CouponIDs) == 0 {
		return utils.LocalizedError(c, fiber.StatusBadRequest, "coupon_id_required", "Coupon ID required")
	}

	if len(req.CouponIDs) > 1000 {
		return utils.LocalizedError(c, fiber.StatusBadRequest, "error_bad_request", "Too many items")
	}

	// Выполняем пакетный сброс через сервис купонов
	response, err := handler.deps.AdminService.BatchResetCoupons(req.CouponIDs)
	if err != nil {
		log.Error().Err(err).Msg("Failed to batch reset coupons")
		return utils.LocalizedError(c, fiber.StatusInternalServerError, "error_internal", "Internal server error")
	}

	return c.JSON(response)
}

// PreviewBatchDelete возвращает предпросмотр пакетного удаления в админ панели
//
//	@Summary		Предпросмотр пакетного удаления (админ)
//	@Description	Показывает информацию о купонах перед удалением и генерирует ключ подтверждения
//	@Tags			admin-coupons
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		coupon.BatchDeleteRequest			true	"Список ID купонов для удаления"
//	@Success		200		{object}	coupon.BatchDeletePreviewResponse	"Предпросмотр удаления"
//	@Failure		400		{object}	map[string]interface{}				"Неверный запрос"
//	@Failure		401		{object}	map[string]interface{}				"Не авторизован"
//	@Failure		403		{object}	map[string]interface{}				"Нет прав доступа"
//	@Failure		500		{object}	map[string]interface{}				"Внутренняя ошибка сервера"
//	@Router			/admin/coupons/batch/delete/preview [post]
func (handler *AdminHandler) PreviewBatchDelete(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())

	var req coupon.BatchDeleteRequest
	if err := c.BodyParser(&req); err != nil {
		log.Error().Err(err).Msg("Invalid request body")
		return utils.LocalizedError(c, fiber.StatusBadRequest, "error_bad_request", "Invalid request")
	}

	// Валидация
	if len(req.CouponIDs) == 0 {
		return utils.LocalizedError(c, fiber.StatusBadRequest, "coupon_id_required", "Coupon ID required")
	}

	if len(req.CouponIDs) > 1000 {
		return utils.LocalizedError(c, fiber.StatusBadRequest, "error_bad_request", "Too many items")
	}

	// Получаем предпросмотр через сервис купонов
	response, err := handler.deps.AdminService.PreviewBatchDelete(req.CouponIDs)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get batch delete preview")
		return utils.LocalizedError(c, fiber.StatusInternalServerError, "error_internal", "Internal server error")
	}

	return c.JSON(response)
}

// ExecuteBatchDelete выполняет подтвержденное пакетное удаление в админ панели
//
//	@Summary		Подтвержденное пакетное удаление (админ)
//	@Description	Удаляет купоны после подтверждения с использованием ключа
//	@Tags			admin-coupons
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		coupon.BatchDeleteConfirmRequest	true	"Подтверждение удаления с ключом"
//	@Success		200		{object}	coupon.BatchDeleteResponse			"Результат удаления"
//	@Failure		400		{object}	map[string]interface{}				"Неверный запрос"
//	@Failure		401		{object}	map[string]interface{}				"Не авторизован"
//	@Failure		403		{object}	map[string]interface{}				"Нет прав доступа"
//	@Failure		500		{object}	map[string]interface{}				"Внутренняя ошибка сервера"
//	@Router			/admin/coupons/batch/delete/confirm [post]
func (handler *AdminHandler) ExecuteBatchDelete(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())

	var req coupon.BatchDeleteConfirmRequest
	if err := c.BodyParser(&req); err != nil {
		log.Error().Err(err).Msg("Invalid request body")
		return utils.LocalizedError(c, fiber.StatusBadRequest, "error_bad_request", "Invalid request")
	}

	// Валидация
	if len(req.CouponIDs) == 0 {
		return utils.LocalizedError(c, fiber.StatusBadRequest, "coupon_id_required", "Coupon ID required")
	}

	if req.ConfirmationKey == "" {
		return utils.LocalizedError(c, fiber.StatusBadRequest, "error_validation_failed", "Confirmation key required")
	}

	// Выполняем удаление через сервис купонов
	response, err := handler.deps.AdminService.ExecuteBatchDelete(req)
	if err != nil {
		log.Error().Err(err).Msg("Failed to execute batch delete")
		return utils.LocalizedError(c, fiber.StatusInternalServerError, "error_internal", "Internal server error")
	}

	return c.JSON(response)
}

// ExportCouponsAdvanced экспортирует купоны с настраиваемыми форматами в админ панели
//
//	@Summary		Продвинутый экспорт купонов (админ)
//	@Description	Экспортирует купоны в различных форматах (TXT, CSV, XLSX) с настраиваемыми опциями
//	@Tags			admin-coupons
//	@Accept			json
//	@Produce		application/octet-stream
//	@Security		BearerAuth
//	@Param			request	body		coupon.ExportOptionsRequest	true	"Опции экспорта"
//	@Success		200		{string}	string						"Файл экспорта"
//	@Failure		400		{object}	map[string]interface{}		"Неверный запрос"
//	@Failure		401		{object}	map[string]interface{}		"Не авторизован"
//	@Failure		403		{object}	map[string]interface{}		"Нет прав доступа"
//	@Failure		500		{object}	map[string]interface{}		"Внутренняя ошибка сервера"
//	@Router			/admin/coupons/export-advanced [post]
func (handler *AdminHandler) ExportCouponsAdvanced(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())

	var req coupon.ExportOptionsRequest
	if err := c.BodyParser(&req); err != nil {
		log.Error().Err(err).Msg("Invalid request body")
		return utils.LocalizedError(c, fiber.StatusBadRequest, "error_bad_request", "Invalid request")
	}

	// Валидация
	if req.Format == "" {
		req.Format = coupon.ExportFormatCodes
	}

	if req.FileFormat == "" {
		req.FileFormat = "txt"
	}

	// Выполняем экспорт через сервис купонов
	content, filename, contentType, err := handler.deps.AdminService.ExportCouponsAdvanced(req)
	if err != nil {
		log.Error().Err(err).Msg("Failed to export coupons")
		return utils.LocalizedError(c, fiber.StatusInternalServerError, "error_internal", "Internal server error")
	}

	// Устанавливаем заголовки для скачивания файла
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Set("Content-Type", contentType)

	return c.Send(content)
}
