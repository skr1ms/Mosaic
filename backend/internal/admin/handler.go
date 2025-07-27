package admin

import (
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
	"github.com/skr1ms/mosaic/pkg/updatePartnerData"
)

type AdminHandlerDeps struct {
	AdminService *AdminService
	JwtService   *jwt.JWT
	Logger       *zerolog.Logger
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
	protected.Post("/admins", handler.CreateAdmin)    // создание ✔
	protected.Get("/admins", handler.GetAdmins)       // получение списка администраторов ✔
	protected.Get("/dashboard", handler.GetDashboard) // получение дашборда администратора

	// Управление партнерами
	protected.Get("/partners", handler.GetPartners)                         // получение списка партнеров ✔
	protected.Post("/partners", handler.CreatePartner)                      // создание партнера ✔
	protected.Get("/partners/:id", handler.GetPartner)                      // получение информации о партнере ✔
	protected.Put("/partners/:id", handler.UpdatePartner)                   // обновление информации о партнере ✔
	protected.Patch("/partners/:id/block", handler.BlockPartner)            // блокировка партнера ✔
	protected.Patch("/partners/:id/unblock", handler.UnblockPartner)        // разблокировка партнера ✔
	protected.Delete("/partners/:id", handler.DeletePartner)                // удаление партнера ✔
	protected.Get("/partners/:id/statistics", handler.GetPartnerStatistics) // получение статистики партнера ✔

	// Управление купонами
	protected.Get("/coupons", handler.GetCoupons)                              // получение списка купонов ✔
	protected.Get("/coupons/paginated", handler.GetCouponsPaginated)           // получение купонов с пагинацией ✔
	protected.Post("/coupons", handler.CreateCoupons)                          // создание купонов ✔
	protected.Get("/coupons/export", handler.ExportCoupons)                    // экспорт купонов ✔
	protected.Get("/coupons/export/partner/:id", handler.ExportPartnerCoupons) // экспорт купонов партнера ✔
	protected.Post("/coupons/batch-delete", handler.BatchDeleteCoupons)        // массовое удаление купонов ✔
	protected.Get("/coupons/:id", handler.GetCoupon)                           // получение информации о купоне ✔
	protected.Patch("/coupons/:id/reset", handler.ResetCoupon)                 // сброс купона ✔
	protected.Delete("/coupons/:id", handler.DeleteCoupon)                     // удаление купона ✔
	protected.Post("/batch-delete", handler.BatchDeleteCoupons)                // Пакетное удаление купонов ✔

	// Статистика и аналитика
	protected.Get("/statistics", handler.GetStatistics)                    // получение общей статистики ✔
	protected.Get("/statistics/partners", handler.GetPartnersStatistics)   // получение статистики партнеров ✔
	protected.Get("/statistics/system", handler.GetSystemStatistics)       // получение системной статистики ✔
	protected.Get("/statistics/analytics", handler.GetAnalytics)           // получение аналитики ✔
	protected.Get("/statistics/dashboard", handler.GetDashboardStatistics) // получение статистики для дашборда ✔

	// Управление задачами обработки изображений
	protected.Get("/images", handler.GetAllImages)              // получение всех задач обработки изображений ✔
	protected.Get("/images/:id", handler.GetImageDetails)       // получение деталей задачи обработки ✔
	protected.Delete("/images/:id", handler.DeleteImageTask)    // удаление задачи обработки ✔
	protected.Post("/images/:id/retry", handler.RetryImageTask) // повторная обработка изображения ✔

	// Профиль администратора
	protected.Put("/profile/password", handler.ChangePassword) // изменение пароля администратора ✔
}

// CreateAdmin создает нового администратора
// @Summary Создание администратора
// @Description Создает нового администратора (только для существующих администраторов)
// @Tags admin-management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param admin body CreateAdminRequest true "Данные нового администратора"
// @Success 201 {object} map[string]interface{} "Администратор создан"
// @Failure 400 {object} map[string]interface{} "Ошибка в запросе"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 403 {object} map[string]interface{} "Нет прав доступа"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /admin/admins [post]
func (handler *AdminHandler) CreateAdmin(c *fiber.Ctx) error {
	var payload CreateAdminRequest

	if err := c.BodyParser(&payload); err != nil {
		handler.deps.Logger.Error().Err(err).Msg(ErrInvalidRequest.Error())
		return c.Status(ErrInvalidRequest.HTTPStatus).JSON(fiber.Map{"error": ErrInvalidRequest.Error()})
	}

	if err := middleware.ValidateStruct(&payload); err != nil {
		handler.deps.Logger.Error().Err(err).Msg(ErrInvalidRequest.Error())
		return c.Status(ErrInvalidRequest.HTTPStatus).JSON(fiber.Map{"error": ErrInvalidRequest.Error()})
	}

	// Создаем администратора через сервис
	admin, err := handler.deps.AdminService.CreateAdmin(payload)
	if err != nil {
		handler.deps.Logger.Error().Err(err).Msg(ErrFailedToCreateAdmin.Error())
		return c.Status(ErrFailedToCreateAdmin.HTTPStatus).JSON(fiber.Map{"error": ErrFailedToCreateAdmin.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":    admin.ID,
		"login": admin.Login,
		"role":  "admin",
	})
}

// GetAdmins возвращает список всех администраторов
// @Summary Список администраторов
// @Description Возвращает список всех администраторов
// @Tags admin-management
// @Produce json
// @Security BearerAuth
// @Success 200 {array} map[string]interface{} "Список администраторов"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 403 {object} map[string]interface{} "Нет прав доступа"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /admin/admins [get]
func (handler *AdminHandler) GetAdmins(c *fiber.Ctx) error {
	admins, err := handler.deps.AdminService.GetAdmins()
	if err != nil {
		handler.deps.Logger.Error().Err(err).Msg(ErrFailedToGetAdmins.Error())
		return c.Status(ErrFailedToGetAdmins.HTTPStatus).JSON(fiber.Map{"error": ErrFailedToGetAdmins.Error()})
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
// @Summary Дашборд администратора
// @Description Возвращает данные для главной страницы администратора
// @Tags admin-dashboard
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Данные дашборда"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 403 {object} map[string]interface{} "Нет прав доступа"
// @Router /admin/dashboard [get]
func (handler *AdminHandler) GetDashboard(c *fiber.Ctx) error {
	dashboardData, err := handler.deps.AdminService.GetDashboardData()
	if err != nil {
		handler.deps.Logger.Error().Err(err).Msg(ErrFailedToGetDashboard.Error())
		return c.Status(ErrFailedToGetDashboard.HTTPStatus).JSON(fiber.Map{"error": ErrFailedToGetDashboard.Error()})
	}

	return c.JSON(dashboardData)
}

// GetPartners возвращает список партнеров с поддержкой фильтрации
// @Summary Список партнеров
// @Description Возвращает список всех партнеров с возможностью фильтрации и поиска
// @Tags admin-partners
// @Produce json
// @Security BearerAuth
// @Param search query string false "Поиск по названию бренда, домену или email"
// @Param status query string false "Фильтр по статусу (active/blocked)"
// @Success 200 {object} map[string]interface{} "Список партнеров"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 403 {object} map[string]interface{} "Нет прав доступа"
// @Router /admin/partners [get]
func (handler *AdminHandler) GetPartners(c *fiber.Ctx) error {
	search := c.Query("search")
	status := c.Query("status")

	partners, err := handler.deps.AdminService.GetPartners(search, status)
	if err != nil {
		handler.deps.Logger.Error().Err(err).Msg(ErrFailedToGetPartners.Error())
		return c.Status(ErrFailedToGetPartners.HTTPStatus).JSON(fiber.Map{"error": ErrFailedToGetPartners.Error()})
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
// @Summary Создание партнера
// @Description Создает нового партнера с автоматически генерируемым кодом (начиная с 0001, 0000 зарезервирован для собственных купонов)
// @Tags admin-partners
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param partner body CreatePartnerRequest true "Данные нового партнера"
// @Success 201 {object} map[string]interface{} "Партнер создан"
// @Failure 400 {object} map[string]interface{} "Ошибка в запросе"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 403 {object} map[string]interface{} "Нет прав доступа"
// @Failure 409 {object} map[string]interface{} "Партнер с таким логином/доменом уже существует"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /admin/partners [post]
func (handler *AdminHandler) CreatePartner(c *fiber.Ctx) error {
	var req partner.CreatePartnerRequest

	if err := c.BodyParser(&req); err != nil {
		handler.deps.Logger.Error().Err(err).Msg(ErrInvalidRequest.Error())
		return c.Status(ErrInvalidRequest.HTTPStatus).JSON(fiber.Map{"error": ErrInvalidRequest.Error()})
	}

	// Проверяем, существует ли уже партнер с таким логином
	existingPartner, err := handler.deps.AdminService.deps.PartnerRepository.GetByLogin(req.Login)
	if err == nil && existingPartner != nil {
		handler.deps.Logger.Error().Str("login", req.Login).Msg("Partner with this login already exists")
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "Partner with this login already exists",
		})
	}

	// Проверяем, существует ли уже партнер с таким доменом
	existingPartnerByDomain, err := handler.deps.AdminService.deps.PartnerRepository.GetByDomain(req.Domain)
	if err == nil && existingPartnerByDomain != nil {
		handler.deps.Logger.Error().Str("domain", req.Domain).Msg("Partner with this domain already exists")
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "Partner with this domain already exists",
		})
	}

	// Генерируем следующий доступный код партнера
	partnerCode, err := handler.deps.AdminService.deps.PartnerRepository.GetNextPartnerCode()
	if err != nil {
		handler.deps.Logger.Error().Err(err).Msg("Failed to generate partner code")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate partner code"})
	}

	hashedPassword, err := bcrypt.HashPassword(req.Password)
	if err != nil {
		handler.deps.Logger.Error().Err(err).Msg(ErrPasswordHashingFailed.Error())
		return c.Status(ErrPasswordHashingFailed.HTTPStatus).JSON(fiber.Map{"error": ErrPasswordHashingFailed.Error()})
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

	if err := handler.deps.AdminService.deps.PartnerRepository.Create(partner); err != nil {
		handler.deps.Logger.Error().Err(err).Msg(ErrFailedToCreatePartner.Error())
		return c.Status(ErrFailedToCreatePartner.HTTPStatus).JSON(fiber.Map{"error": ErrFailedToCreatePartner.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":           partner.ID,
		"login":        partner.Login,
		"partner_code": partner.PartnerCode,
	})
}

// GetPartner возвращает детальную информацию о партнере
// @Summary Детальная информация о партнере
// @Description Возвращает детальную информацию о партнере по ID
// @Tags admin-partners
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID партнера"
// @Success 200 {object} map[string]interface{} "Информация о партнере"
// @Failure 400 {object} map[string]interface{} "Неверный ID"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 403 {object} map[string]interface{} "Нет прав доступа"
// @Failure 404 {object} map[string]interface{} "Партнер не найден"
// @Router /admin/partners/{id} [get]
func (handler *AdminHandler) GetPartner(c *fiber.Ctx) error {
	partnerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		handler.deps.Logger.Error().Err(err).Str("id", c.Params("id")).Msg(ErrInvalidID.Error())
		return c.Status(ErrInvalidID.HTTPStatus).JSON(fiber.Map{"error": ErrInvalidID.Error()})
	}

	partner, err := handler.deps.AdminService.deps.PartnerRepository.GetByID(partnerID)
	if err != nil {
		handler.deps.Logger.Error().Err(err).Str("partner_id", partnerID.String()).Msg(ErrPartnerNotFound.Error())
		return c.Status(ErrPartnerNotFound.HTTPStatus).JSON(fiber.Map{"error": ErrPartnerNotFound.Error()})
	}

	// Получаем информацию о партнере
	totalCoupons, err := handler.deps.AdminService.deps.CouponRepository.CountByPartnerID(partnerID)
	if err != nil {
		totalCoupons = 0
	}

	activatedCoupons, err := handler.deps.AdminService.deps.CouponRepository.CountActivatedByPartnerID(partnerID)
	if err != nil {
		activatedCoupons = 0
	}

	purchasedCoupons, err := handler.deps.AdminService.deps.CouponRepository.CountPurchasedByPartnerID(partnerID)
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

// UpdatePartner обновляет партнера
// @Summary Обновление партнера
// @Description Обновляет данные партнера
// @Tags admin-partners
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID партнера"
// @Success 200 {object} map[string]interface{} "Партнер обновлен"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 403 {object} map[string]interface{} "Нет прав доступа"
// @Router /admin/partners/{id} [put]
func (handler *AdminHandler) UpdatePartner(c *fiber.Ctx) error {
	var req partner.UpdatePartnerRequest

	if err := c.BodyParser(&req); err != nil {
		handler.deps.Logger.Error().Err(err).Msg(ErrInvalidRequest.Error())
		return c.Status(ErrInvalidRequest.HTTPStatus).JSON(fiber.Map{"error": ErrInvalidRequest.Error()})
	}

	partnerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		handler.deps.Logger.Error().Err(err).Str("id", c.Params("id")).Msg(ErrInvalidID.Error())
		return c.Status(ErrInvalidID.HTTPStatus).JSON(fiber.Map{"error": ErrInvalidID.Error()})
	}

	partner, err := handler.deps.AdminService.deps.PartnerRepository.GetByID(partnerID)
	if err != nil {
		handler.deps.Logger.Error().Err(err).Str("partner_id", partnerID.String()).Msg(ErrPartnerNotFound.Error())
		return c.Status(ErrPartnerNotFound.HTTPStatus).JSON(fiber.Map{"error": ErrPartnerNotFound.Error()})
	}

	updatePartnerData.UpdatePartnerData(partner, &req)

	if err := handler.deps.AdminService.deps.PartnerRepository.Update(partner); err != nil {
		handler.deps.Logger.Error().Err(err).Str("partner_id", partnerID.String()).Msg(ErrFailedToUpdatePartner.Error())
		return c.Status(ErrFailedToUpdatePartner.HTTPStatus).JSON(fiber.Map{"error": ErrFailedToUpdatePartner.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"id":           partner.ID,
		"login":        partner.Login,
		"partner_code": partner.PartnerCode,
	})
}

// BlockPartner блокирует партнера
// @Summary Блокировка партнера
// @Description Блокирует партнера (временное отключение доступа)
// @Tags admin-partners
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID партнера"
// @Success 200 {object} map[string]interface{} "Партнер заблокирован"
// @Failure 400 {object} map[string]interface{} "Неверный ID"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 403 {object} map[string]interface{} "Нет прав доступа"
// @Failure 404 {object} map[string]interface{} "Партнер не найден"
// @Router /admin/partners/{id}/block [patch]
func (handler *AdminHandler) BlockPartner(c *fiber.Ctx) error {
	partnerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		handler.deps.Logger.Error().Err(err).Str("id", c.Params("id")).Msg(ErrInvalidID.Error())
		return c.Status(ErrInvalidID.HTTPStatus).JSON(fiber.Map{"error": ErrInvalidID.Error()})
	}

	if err := handler.deps.AdminService.deps.PartnerRepository.UpdateStatus(partnerID, "blocked"); err != nil {
		handler.deps.Logger.Error().Err(err).Str("partner_id", partnerID.String()).Msg(ErrFailedToBlockPartner.Error())
		return c.Status(ErrFailedToBlockPartner.HTTPStatus).JSON(fiber.Map{"error": ErrFailedToBlockPartner.Error()})
	}

	if err := handler.deps.AdminService.deps.CouponRepository.UpdateStatusByPartnerID(partnerID, true); err != nil {
		handler.deps.Logger.Error().Err(err).Str("partner_id", partnerID.String()).Msg("Failed to block coupons")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to block coupons"})
	}

	return c.JSON(fiber.Map{"message": "Partner blocked successfully"})
}

// UnblockPartner разблокирует партнера
// @Summary Разблокировка партнера
// @Description Разблокирует партнера (восстановление доступа)
// @Tags admin-partners
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID партнера"
// @Success 200 {object} map[string]interface{} "Партнер разблокирован"
// @Failure 400 {object} map[string]interface{} "Неверный ID"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 403 {object} map[string]interface{} "Нет прав доступа"
// @Failure 404 {object} map[string]interface{} "Партнер не найден"
// @Router /admin/partners/{id}/unblock [patch]
func (handler *AdminHandler) UnblockPartner(c *fiber.Ctx) error {
	partnerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		handler.deps.Logger.Error().Err(err).Str("id", c.Params("id")).Msg(ErrInvalidID.Error())
		return c.Status(ErrInvalidID.HTTPStatus).JSON(fiber.Map{"error": ErrInvalidID.Error()})
	}

	if err := handler.deps.AdminService.deps.PartnerRepository.UpdateStatus(partnerID, "active"); err != nil {
		handler.deps.Logger.Error().Err(err).Str("partner_id", partnerID.String()).Msg(ErrFailedToUnblockPartner.Error())
		return c.Status(ErrFailedToUnblockPartner.HTTPStatus).JSON(fiber.Map{"error": ErrFailedToUnblockPartner.Error()})
	}

	if err := handler.deps.AdminService.deps.CouponRepository.UpdateStatusByPartnerID(partnerID, false); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to unblock coupons"})
	}

	return c.JSON(fiber.Map{"message": "Partner unblocked successfully"})
}

// DeletePartner удаляет партнера
// @Summary Удаление партнера
// @Description Удаляет партнера по ID с удалением всех связанных данных
// @Tags admin-partners
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID партнера"
// @Param confirm query boolean false "Подтверждение удаления (true/false)"
// @Success 200 {object} map[string]interface{} "Партнер удален"
// @Failure 400 {object} map[string]interface{} "Неверный ID или требуется подтверждение"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 403 {object} map[string]interface{} "Нет прав доступа"
// @Failure 404 {object} map[string]interface{} "Партнер не найден"
// @Router /admin/partners/{id} [delete]
func (handler *AdminHandler) DeletePartner(c *fiber.Ctx) error {
	partnerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid partner ID"})
	}

	partner, err := handler.deps.AdminService.deps.PartnerRepository.GetByID(partnerID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Partner not found"})
	}

	totalCoupons, _ := handler.deps.AdminService.deps.CouponRepository.CountByPartnerID(partnerID)
	activatedCoupons, _ := handler.deps.AdminService.deps.CouponRepository.CountActivatedByPartnerID(partnerID)

	confirm := c.Query("confirm") == "true"
	if !confirm {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Deletion requires confirmation",
			"warning": "This action will permanently delete the partner and all associated data",
			"partner_info": fiber.Map{
				"brand_name":        partner.BrandName,
				"total_coupons":     totalCoupons,
				"activated_coupons": activatedCoupons,
			},
			"confirm_url": "/admin/partners/" + partnerID.String() + "?confirm=true",
		})
	}

	// Используем транзакцию для атомарного удаления партнера и его купонов
	err = handler.deps.AdminService.deps.PartnerRepository.DeleteWithCoupons(partnerID)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete partner and associated data"})
	}

	// Проверяем, что купоны действительно удалены
	remainingCoupons, _ := handler.deps.AdminService.deps.CouponRepository.CountByPartnerID(partnerID)
	if remainingCoupons > 0 {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":             "Partner deleted but some coupons remain",
			"remaining_coupons": remainingCoupons,
		})
	}

	return c.JSON(fiber.Map{
		"message": "Partner deleted successfully",
		"deleted_partner": fiber.Map{
			"id":                partner.ID,
			"brand_name":        partner.BrandName,
			"total_coupons":     totalCoupons,
			"activated_coupons": activatedCoupons,
		},
	})
}

// GetPartnerStatistics возвращает статистику конкретного партнера
// @Summary Статистика партнера
// @Description Возвращает детальную статистику по конкретному партнеру
// @Tags admin-partners
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID партнера"
// @Success 200 {object} map[string]interface{} "Статистика партнера"
// @Failure 400 {object} map[string]interface{} "Неверный ID"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 403 {object} map[string]interface{} "Нет прав доступа"
// @Router /admin/partners/{id}/statistics [get]
func (handler *AdminHandler) GetPartnerStatistics(c *fiber.Ctx) error {
	partnerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid partner ID"})
	}

	// Получаем статистику партнера
	totalCoupons, err := handler.deps.AdminService.deps.CouponRepository.CountByPartnerID(partnerID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get statistics"})
	}

	activatedCoupons, err := handler.deps.AdminService.deps.CouponRepository.CountActivatedByPartnerID(partnerID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get statistics"})
	}

	purchasedCoupons, err := handler.deps.AdminService.deps.CouponRepository.CountPurchasedByPartnerID(partnerID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get statistics"})
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
// @Summary Список купонов
// @Description Возвращает список всех купонов с возможностью фильтрации
// @Tags admin-coupons
// @Produce json
// @Security BearerAuth
// @Param search query string false "Поиск по номеру купона"
// @Param partner_id query string false "ID партнера для фильтрации"
// @Param status query string false "Статус для фильтрации (new/used)"
// @Param size query string false "Размер для фильтрации"
// @Param style query string false "Стиль для фильтрации"
// @Param limit query int false "Количество записей на странице (по умолчанию 50)"
// @Param offset query int false "Смещение для пагинации (по умолчанию 0)"
// @Success 200 {object} map[string]interface{} "Список купонов"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 403 {object} map[string]interface{} "Нет прав доступа"
// @Router /admin/coupons [get]
func (handler *AdminHandler) GetCoupons(c *fiber.Ctx) error {
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

	coupons, err := handler.deps.AdminService.deps.CouponRepository.GetFiltered(filters)
	if err != nil {
		handler.deps.Logger.Error().Err(err).Interface("filters", filters).Msg(ErrFailedToGetCoupons.Error())
		return c.Status(ErrFailedToGetCoupons.HTTPStatus).JSON(fiber.Map{"error": ErrFailedToGetCoupons.Error()})
	}

	// Добавляем информацию о партнерах
	result := make([]fiber.Map, len(coupons))
	for i, coupon := range coupons {
		partnerName := "Собственный"
		if coupon.PartnerID != uuid.Nil {
			if partner, err := handler.deps.AdminService.deps.PartnerRepository.GetByID(coupon.PartnerID); err == nil {
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
// @Summary Список купонов с пагинацией для админа
// @Description Возвращает список купонов с пагинацией и расширенной фильтрацией для административной панели
// @Tags admin-coupons
// @Produce json
// @Security BearerAuth
// @Param page query int false "Номер страницы (по умолчанию 1)"
// @Param limit query int false "Количество элементов на странице (по умолчанию 20, максимум 100)"
// @Param search query string false "Поиск по коду купона"
// @Param partner_id query string false "ID партнера"
// @Param status query string false "Статус купона (new, used)"
// @Param size query string false "Размер купона"
// @Param style query string false "Стиль купона"
// @Param created_from query string false "Дата создания от (RFC3339)"
// @Param created_to query string false "Дата создания до (RFC3339)"
// @Param used_from query string false "Дата активации от (RFC3339)"
// @Param used_to query string false "Дата активации до (RFC3339)"
// @Success 200 {object} map[string]interface{} "Купоны с информацией о пагинации"
// @Failure 400 {object} map[string]interface{} "Неверные параметры запроса"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 403 {object} map[string]interface{} "Нет прав доступа"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /admin/coupons/paginated [get]
func (handler *AdminHandler) GetCouponsPaginated(c *fiber.Ctx) error {
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
		code, status, size, style, partnerID, page, limit,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch coupons",
		})
	}

	// Добавляем информацию о партнерах к каждому купону
	result := make([]fiber.Map, len(coupons))
	for i, coupon := range coupons {
		partnerName := "Собственный"
		if coupon.PartnerID != uuid.Nil {
			if partner, err := handler.deps.AdminService.deps.PartnerRepository.GetByID(coupon.PartnerID); err == nil {
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
	totalPages := (total + int64(limit) - 1) / int64(limit)
	hasNext := int64(page) < totalPages
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
// @Summary Создание купонов
// @Description Создает новые купоны в пакетном режиме
// @Tags admin-coupons
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateCouponsRequest true "Параметры создания купонов"
// @Success 201 {object} map[string]interface{} "Купоны созданы"
// @Failure 400 {object} map[string]interface{} "Ошибка в запросе"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 403 {object} map[string]interface{} "Нет прав доступа"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /admin/coupons [post]
func (handler *AdminHandler) CreateCoupons(c *fiber.Ctx) error {
	var req coupon.CreateCouponRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	// Валидация количества
	if req.Count < 1 || req.Count > 10000 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Count must be between 1 and 10000"})
	}

	// Получаем код партнера
	var partnerCode string = "0000" // Для собственных купонов
	if req.PartnerID != uuid.Nil {
		partner, err := handler.deps.AdminService.deps.PartnerRepository.GetByID(req.PartnerID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Partner not found"})
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
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Ошибка при генерации уникального кода купона: " + err.Error(),
			})
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
	if err := handler.deps.AdminService.deps.CouponRepository.CreateBatch(coupons); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create coupons"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":     "Coupons created successfully",
		"count":       req.Count,
		"codes":       codes,
		"partner_id":  req.PartnerID,
		"size":        req.Size,
		"style":       req.Style,
		"codes_range": []string{codes[0], codes[len(codes)-1]},
	})
}

// ExportCoupons экспортирует список купонов в файл
// @Summary Экспорт купонов
// @Description Экспортирует список купонов в файл (поддерживаются форматы txt и csv)
// @Tags admin-coupons
// @Produce text/plain,text/csv
// @Security BearerAuth
// @Param format query string false "Формат экспорта (txt или csv)" default(txt)
// @Param filename query string false "Имя файла (без расширения)"
// @Param partner_id query string false "ID партнера для фильтрации"
// @Param status query string false "Статус для фильтрации (new/used)"
// @Param size query string false "Размер для фильтрации"
// @Param style query string false "Стиль для фильтрации"
// @Success 200 {string} string "Файл с купонами"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 403 {object} map[string]interface{} "Нет прав доступа"
// @Router /admin/coupons/export [get]
func (handler *AdminHandler) ExportCoupons(c *fiber.Ctx) error {
	format := strings.ToLower(c.Query("format", "txt"))
	if format != "txt" && format != "csv" {
		format = "txt"
	}

	// Получаем все купоны с данными партнеров
	coupons, err := handler.deps.AdminService.deps.PartnerRepository.GetAllCouponsForExport()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch coupons",
		})
	}

	if len(coupons) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "No coupons found for export",
		})
	}

	var content strings.Builder
	filename := fmt.Sprintf("all_coupons_%s.%s", time.Now().Format("20060102_150405"), format)

	if format == "csv" {
		// CSV формат с полными данными для админа и группировкой по партнерам
		content.WriteString("Coupon Code,Partner ID,Partner Status,Coupon Status,Size,Style,Brand Name,Email,Created At,Used At\n")

		var currentPartnerID uuid.UUID
		for _, coupon := range coupons {
			// Если это новый партнер, добавляем разделительную строку
			if coupon.PartnerID != currentPartnerID {
				currentPartnerID = coupon.PartnerID
				if content.Len() > len("Coupon Code,Partner ID,Partner Status,Coupon Status,Size,Style,Brand Name,Email,Created At,Used At\n") {
					// Добавляем пустую строку между партнерами (кроме первого)
					content.WriteString(",,,,,,,,,,\n")
				}
				// Добавляем строку с информацией о партнере
				content.WriteString(fmt.Sprintf("=== PARTNER: %s ===,,%s,,,,%s,%s,,\n",
					coupon.BrandName,
					coupon.PartnerStatus,
					coupon.BrandName,
					coupon.Email,
				))
			}

			usedAt := ""
			if coupon.UsedAt != nil {
				usedAt = coupon.UsedAt.Format("2006-01-02 15:04:05")
			}
			content.WriteString(fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s,%s,%s\n",
				coupon.CouponCode,
				coupon.PartnerID.String(),
				coupon.PartnerStatus,
				coupon.CouponStatus,
				coupon.Size,
				coupon.Style,
				coupon.BrandName,
				coupon.Email,
				coupon.CreatedAt.Format("2006-01-02 15:04:05"),
				usedAt,
			))
		}
		c.Set("Content-Type", "text/csv")
	} else {
		// TXT формат с группировкой по партнерам
		content.WriteString("All Coupons Export\n")
		content.WriteString("==================\n\n")

		var currentPartnerID uuid.UUID
		for _, coupon := range coupons {
			// Если это новый партнер, добавляем заголовок
			if coupon.PartnerID != currentPartnerID {
				currentPartnerID = coupon.PartnerID
				content.WriteString(fmt.Sprintf("\n=== Partner: %s ===\n", coupon.BrandName))
				content.WriteString(fmt.Sprintf("Partner ID: %s\n", coupon.PartnerID.String()))
				content.WriteString(fmt.Sprintf("Email: %s\n", coupon.Email))
				content.WriteString(fmt.Sprintf("Status: %s\n\n", coupon.PartnerStatus))
			}

			content.WriteString(fmt.Sprintf("Code: %s\n", coupon.CouponCode))
			content.WriteString(fmt.Sprintf("Coupon Status: %s\n", coupon.CouponStatus))
			content.WriteString(fmt.Sprintf("Size: %s\n", coupon.Size))
			content.WriteString(fmt.Sprintf("Style: %s\n", coupon.Style))
			content.WriteString(fmt.Sprintf("Created: %s\n", coupon.CreatedAt.Format("2006-01-02 15:04:05")))
			if coupon.UsedAt != nil {
				content.WriteString(fmt.Sprintf("Used: %s\n", coupon.UsedAt.Format("2006-01-02 15:04:05")))
			}
			content.WriteString("---\n")
		}
		c.Set("Content-Type", "text/plain")
	}

	// Устанавливаем заголовки для автоматического скачивания
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Set("Cache-Control", "no-cache")

	return c.SendString(content.String())
}

// ExportPartnerCoupons экспортирует купоны конкретного партнера
// @Summary Экспорт купонов партнера для админа
// @Description Экспортирует купоны конкретного партнера со статусом "new" в формате .txt или .csv
// @Tags admin-coupons
// @Produce text/plain,text/csv
// @Security BearerAuth
// @Param id path string true "ID партнера"
// @Param format query string false "Формат файла (txt или csv)" default(txt)
// @Success 200 {string} string "Файл с купонами партнера"
// @Failure 400 {object} map[string]interface{} "Неверный ID партнера"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 403 {object} map[string]interface{} "Нет прав доступа"
// @Failure 404 {object} map[string]interface{} "Партнер не найден"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /admin/coupons/export/partner/{id} [get]
func (handler *AdminHandler) ExportPartnerCoupons(c *fiber.Ctx) error {
	partnerIDStr := c.Params("id")
	partnerID, err := uuid.Parse(partnerIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid partner ID",
		})
	}

	format := strings.ToLower(c.Query("format", "txt"))
	if format != "txt" && format != "csv" {
		format = "txt"
	}

	// Проверяем существование партнера
	partner, err := handler.deps.AdminService.deps.PartnerRepository.GetByID(partnerID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Partner not found",
		})
	}

	// Получаем купоны партнера со статусом "new"
	coupons, err := handler.deps.AdminService.deps.PartnerRepository.GetPartnerCouponsForExport(partnerID, "new")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch partner coupons",
		})
	}

	if len(coupons) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "No new coupons found for this partner",
		})
	}

	// Генерируем содержимое файла
	var content strings.Builder
	filename := fmt.Sprintf("partner_%s_coupons_%s.%s",
		partner.BrandName,
		time.Now().Format("20060102_150405"),
		format)

	if format == "csv" {
		// CSV формат с расширенными данными для админа
		content.WriteString("Coupon Code,Partner ID,Partner Status,Coupon Status,Size,Style,Brand Name,Email,Created At\n")
		for _, coupon := range coupons {
			content.WriteString(fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s,%s\n",
				coupon.CouponCode,
				coupon.PartnerID.String(),
				coupon.PartnerStatus,
				coupon.CouponStatus,
				coupon.Size,
				coupon.Style,
				coupon.BrandName,
				coupon.Email,
				coupon.CreatedAt.Format("2006-01-02 15:04:05"),
			))
		}
		c.Set("Content-Type", "text/csv")
	} else {
		// TXT формат с расширенными данными для админа
		content.WriteString(fmt.Sprintf("Partner Coupons Export - %s\n", partner.BrandName))
		content.WriteString("=====================================\n\n")
		content.WriteString(fmt.Sprintf("Partner ID: %s\n", partnerID.String()))
		content.WriteString(fmt.Sprintf("Brand Name: %s\n", partner.BrandName))
		content.WriteString(fmt.Sprintf("Email: %s\n", partner.Email))
		content.WriteString(fmt.Sprintf("Partner Status: %s\n\n", partner.Status))

		for _, coupon := range coupons {
			content.WriteString(fmt.Sprintf("Code: %s\n", coupon.CouponCode))
			content.WriteString(fmt.Sprintf("Coupon Status: %s\n", coupon.CouponStatus))
			content.WriteString(fmt.Sprintf("Size: %s\n", coupon.Size))
			content.WriteString(fmt.Sprintf("Style: %s\n", coupon.Style))
			content.WriteString(fmt.Sprintf("Created: %s\n", coupon.CreatedAt.Format("2006-01-02 15:04:05")))
			content.WriteString("---\n")
		}
		c.Set("Content-Type", "text/plain")
	}

	// Устанавливаем заголовки для автоматического скачивания
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Set("Cache-Control", "no-cache")

	return c.SendString(content.String())
}

// BatchDeleteCoupons массово удаляет купоны для админ панели
// @Summary Массовое удаление купонов для админа
// @Description Удаляет множество купонов по их ID в административной панели
// @Tags admin-coupons
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body map[string][]string true "Список ID купонов для удаления"
// @Success 200 {object} map[string]interface{} "Результат массового удаления"
// @Failure 400 {object} map[string]interface{} "Ошибка в запросе"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 403 {object} map[string]interface{} "Нет прав доступа"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /admin/coupons/batch-delete [post]
func (handler *AdminHandler) BatchDeleteCoupons(c *fiber.Ctx) error {
	var req struct {
		CouponIDs []string `json:"coupon_ids" validate:"required,min=1"`
		Confirm   bool     `json:"confirm" validate:"required"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	// Проверяем подтверждение
	if !req.Confirm {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Confirmation required for batch deletion",
		})
	}

	// Преобразуем строки в UUID
	ids := make([]uuid.UUID, 0, len(req.CouponIDs))
	for _, idStr := range req.CouponIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("Invalid UUID: %s", idStr),
			})
		}
		ids = append(ids, id)
	}

	deletedCount, err := handler.deps.AdminService.deps.CouponRepository.BatchDelete(ids)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete coupons",
		})
	}

	return c.JSON(fiber.Map{
		"message":       "Coupons deleted successfully",
		"deleted_count": deletedCount,
		"requested":     len(req.CouponIDs),
	})
}

// GetCoupon возвращает детальную информацию о купоне
// @Summary Детальная информация о купоне
// @Description Возвращает детальную информацию о купоне по ID
// @Tags admin-coupons
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID купона"
// @Success 200 {object} map[string]interface{} "Информация о купоне"
// @Failure 400 {object} map[string]interface{} "Неверный ID"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 403 {object} map[string]interface{} "Нет прав доступа"
// @Failure 404 {object} map[string]interface{} "Купон не найден"
// @Router /admin/coupons/{id} [get]
func (handler *AdminHandler) GetCoupon(c *fiber.Ctx) error {
	couponID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid coupon ID"})
	}

	coupon, err := handler.deps.AdminService.deps.CouponRepository.GetByID(couponID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Coupon not found"})
	}

	// Получаем информацию о партнере
	var partnerName string = "Собственный"
	if coupon.PartnerID != uuid.Nil {
		partner, err := handler.deps.AdminService.deps.PartnerRepository.GetByID(coupon.PartnerID)
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
// @Summary Сброс купона
// @Description Сбрасывает купон в статус "новый" с удалением всех данных активации
// @Tags admin-coupons
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID купона"
// @Success 200 {object} map[string]interface{} "Купон сброшен"
// @Failure 400 {object} map[string]interface{} "Неверный ID"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 403 {object} map[string]interface{} "Нет прав доступа"
// @Failure 404 {object} map[string]interface{} "Купон не найден"
// @Router /admin/coupons/{id}/reset [patch]
func (handler *AdminHandler) ResetCoupon(c *fiber.Ctx) error {
	couponID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid coupon ID"})
	}

	if err := handler.deps.AdminService.deps.CouponRepository.Reset(couponID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to reset coupon"})
	}

	return c.JSON(fiber.Map{"message": "Coupon reset successfully"})
}

// DeleteCoupon удаляет купон
// @Summary Удаление купона
// @Description Удаляет купон по ID с подтверждением
// @Tags admin-coupons
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID купона"
// @Param confirm query boolean false "Подтверждение удаления (true/false)"
// @Success 200 {object} map[string]interface{} "Купон удален"
// @Failure 400 {object} map[string]interface{} "Неверный ID или требуется подтверждение"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 403 {object} map[string]interface{} "Нет прав доступа"
// @Failure 404 {object} map[string]interface{} "Купон не найден"
// @Router /admin/coupons/{id} [delete]
func (handler *AdminHandler) DeleteCoupon(c *fiber.Ctx) error {
	couponID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid coupon ID"})
	}

	// Проверяем существование купона
	coupon, err := handler.deps.AdminService.deps.CouponRepository.GetByID(couponID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Coupon not found"})
	}

	// Проверяем подтверждение
	confirm := c.Query("confirm") == "true"
	if !confirm {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Deletion requires confirmation",
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
	if err := handler.deps.AdminService.deps.CouponRepository.Delete(couponID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete coupon"})
	}

	return c.JSON(fiber.Map{
		"message": "Coupon deleted successfully",
		"deleted_coupon": fiber.Map{
			"id":     coupon.ID,
			"code":   coupon.Code,
			"status": coupon.Status,
		},
	})
}

// GetStatistics возвращает общую статистику
// @Summary Общая статистика
// @Description Возвращает общую статистику по системе
// @Tags admin-statistics
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Статистика"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 403 {object} map[string]interface{} "Нет прав доступа"
// @Router /admin/statistics [get]
func (handler *AdminHandler) GetStatistics(c *fiber.Ctx) error {
	// Общее количество купонов
	var totalCoupons int64
	handler.deps.AdminService.deps.CouponRepository.GetAll()
	if coupons, err := handler.deps.AdminService.deps.CouponRepository.GetAll(); err == nil {
		totalCoupons = int64(len(coupons))
	}

	// Количество активированных купонов
	var activatedCoupons int64
	if coupons, err := handler.deps.AdminService.deps.CouponRepository.GetFiltered(map[string]interface{}{"status": "used"}); err == nil {
		activatedCoupons = int64(len(coupons))
	}

	// Количество активных партнеров
	var activePartners int64
	if partners, err := handler.deps.AdminService.deps.PartnerRepository.GetActivePartners(); err == nil {
		activePartners = int64(len(partners))
	}

	// Всего партнеров
	var totalPartners int64
	if partners, err := handler.deps.AdminService.deps.PartnerRepository.GetAll(); err == nil {
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
// @Summary Статистика по партнерам
// @Description Возвращает детальную статистику по всем партнерам
// @Tags admin-statistics
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Статистика по партнерам"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 403 {object} map[string]interface{} "Нет прав доступа"
// @Router /admin/statistics/partners [get]
func (handler *AdminHandler) GetPartnersStatistics(c *fiber.Ctx) error {
	partners, err := handler.deps.AdminService.deps.PartnerRepository.GetAll()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch partners"})
	}

	var partnersStats []fiber.Map
	for _, partner := range partners {
		totalCoupons, _ := handler.deps.AdminService.deps.CouponRepository.CountByPartnerID(partner.ID)
		activatedCoupons, _ := handler.deps.AdminService.deps.CouponRepository.CountActivatedByPartnerID(partner.ID)
		purchasedCoupons, _ := handler.deps.AdminService.deps.CouponRepository.CountPurchasedByPartnerID(partner.ID)

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
// @Summary Смена пароля
// @Description Изменяет пароль текущего администратора
// @Tags admin-profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param password body ChangePasswordRequest true "Данные для смены пароля"
// @Success 200 {object} map[string]interface{} "Пароль изменен"
// @Failure 400 {object} map[string]interface{} "Ошибка в запросе"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 403 {object} map[string]interface{} "Неверный текущий пароль"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /admin/profile/password [put]
func (handler *AdminHandler) ChangePassword(c *fiber.Ctx) error {
	var req ChangePasswordRequest

	if err := c.BodyParser(&req); err != nil {
		handler.deps.Logger.Error().Err(err).Msg(ErrInvalidRequest.Error())
		return c.Status(ErrInvalidRequest.HTTPStatus).JSON(fiber.Map{"error": ErrInvalidRequest.Error()})
	}

	if err := middleware.ValidateStruct(&req); err != nil {
		handler.deps.Logger.Error().Err(err).Msg(ErrInvalidRequest.Error())
		return c.Status(ErrInvalidRequest.HTTPStatus).JSON(fiber.Map{"error": ErrInvalidRequest.Error()})
	}

	claims, err := jwt.GetClaimsFromFiberContext(c)
	if err != nil {
		handler.deps.Logger.Error().Err(err).Msg("Unauthorized - failed to get claims")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	if err := handler.deps.AdminService.ChangePassword(claims.UserID, req); err != nil {
		if IsAPIError(err) {
			apiErr, _ := GetAPIError(err)
			handler.deps.Logger.Error().Err(err).Str("admin_id", claims.UserID.String()).Msg(apiErr.Error())
			return c.Status(apiErr.HTTPStatus).JSON(fiber.Map{"error": apiErr.Error()})
		}
		handler.deps.Logger.Error().Err(err).Str("admin_id", claims.UserID.String()).Msg(ErrFailedToChangePassword.Error())
		return c.Status(ErrFailedToChangePassword.HTTPStatus).JSON(fiber.Map{"error": ErrFailedToChangePassword.Error()})
	}

	return c.JSON(fiber.Map{"message": "Password changed successfully"})
}

// GetSystemStatistics возвращает системную статистику
// @Summary Системная статистика
// @Description Возвращает детальную статистику системы: производительность, нагрузка, состояние очереди
// @Tags admin-statistics
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Системная статистика"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 403 {object} map[string]interface{} "Нет прав доступа"
// @Router /admin/statistics/system [get]
func (handler *AdminHandler) GetSystemStatistics(c *fiber.Ctx) error {
	// Здесь будет логика получения системной статистики:
	// - Состояние очереди обработки изображений
	// - Производительность системы
	// - Использование ресурсов
	// - Состояние базы данных

	systemStats := map[string]interface{}{
		"queue_status": map[string]interface{}{
			"total_tasks":      150,
			"queued_tasks":     25,
			"processing_tasks": 5,
			"completed_tasks":  120,
			"failed_tasks":     0,
		},
		"system_health": map[string]interface{}{
			"status":            "healthy",
			"uptime_hours":      168,
			"memory_usage_mb":   512,
			"cpu_usage_percent": 25,
		},
		"database": map[string]interface{}{
			"status":           "connected",
			"response_time_ms": 15,
		},
		"processing_performance": map[string]interface{}{
			"avg_processing_time_minutes": 3.5,
			"success_rate_percent":        99.2,
		},
	}

	return c.JSON(systemStats)
}

// GetAnalytics возвращает аналитику
// @Summary Аналитика
// @Description Возвращает детальную аналитику по использованию системы
// @Tags admin-statistics
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Аналитика"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 403 {object} map[string]interface{} "Нет прав доступа"
// @Router /admin/statistics/analytics [get]
func (handler *AdminHandler) GetAnalytics(c *fiber.Ctx) error {
	// Здесь будет логика получения аналитики:
	// - Топ партнеры
	// - Тренды использования
	// - Популярные размеры и стили
	// - Региональная статистика

	analytics := map[string]interface{}{
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

	return c.JSON(analytics)
}

// GetDashboardStatistics возвращает статистику для дашборда
// @Summary Статистика дашборда
// @Description Возвращает основную статистику для отображения на дашборде администратора
// @Tags admin-statistics
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Статистика дашборда"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 403 {object} map[string]interface{} "Нет прав доступа"
// @Router /admin/statistics/dashboard [get]
func (handler *AdminHandler) GetDashboardStatistics(c *fiber.Ctx) error {
	// Здесь будет логика получения статистики для дашборда:
	// - Ключевые метрики
	// - Последние активности
	// - Уведомления и предупреждения

	dashboardStats := map[string]interface{}{
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

	return c.JSON(dashboardStats)
}

// GetAllImages возвращает все задачи обработки изображений
// @Summary Все задачи обработки изображений
// @Description Возвращает список всех задач обработки изображений с фильтрацией
// @Tags admin-images
// @Produce json
// @Security BearerAuth
// @Param status query string false "Фильтр по статусу (queued, processing, completed, failed)"
// @Param partner_id query string false "Фильтр по ID партнера"
// @Param limit query int false "Лимит записей (по умолчанию 50)"
// @Param offset query int false "Смещение (по умолчанию 0)"
// @Success 200 {array} map[string]interface{} "Список задач обработки"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 403 {object} map[string]interface{} "Нет прав доступа"
// @Router /admin/images [get]
func (handler *AdminHandler) GetAllImages(c *fiber.Ctx) error {
	status := c.Query("status")
	partnerID := c.Query("partner_id")

	// TODO: Реализовать фильтрацию и пагинацию через ImageRepository
	// Пока возвращаем заглушку

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
// @Summary Детали задачи обработки
// @Description Возвращает подробную информацию о задаче обработки изображения
// @Tags admin-images
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID задачи"
// @Success 200 {object} map[string]interface{} "Детали задачи"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 403 {object} map[string]interface{} "Нет прав доступа"
// @Failure 404 {object} map[string]interface{} "Задача не найдена"
// @Router /admin/images/{id} [get]
func (handler *AdminHandler) GetImageDetails(c *fiber.Ctx) error {
	imageID := c.Params("id")

	imageUUID, err := uuid.Parse(imageID)
	if err != nil {
		handler.deps.Logger.Error().Err(err).Str("image_id", imageID).Msg(ErrInvalidID.Error())
		return c.Status(ErrInvalidID.HTTPStatus).JSON(fiber.Map{
			"error": ErrInvalidID.Error(),
		})
	}

	// TODO: Получить задачу через ImageRepository
	_ = imageUUID

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
// @Summary Удаление задачи обработки
// @Description Удаляет задачу обработки изображения и связанные файлы
// @Tags admin-images
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID задачи"
// @Success 200 {object} map[string]interface{} "Задача удалена"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 403 {object} map[string]interface{} "Нет прав доступа"
// @Failure 404 {object} map[string]interface{} "Задача не найдена"
// @Router /admin/images/{id} [delete]
func (handler *AdminHandler) DeleteImageTask(c *fiber.Ctx) error {
	imageID := c.Params("id")

	imageUUID, err := uuid.Parse(imageID)
	if err != nil {
		handler.deps.Logger.Error().Err(err).Str("image_id", imageID).Msg(ErrInvalidID.Error())
		return c.Status(ErrInvalidID.HTTPStatus).JSON(fiber.Map{
			"error": ErrInvalidID.Error(),
		})
	}

	// TODO: Удалить задачу через ImageRepository
	// TODO: Удалить связанные файлы с диска
	_ = imageUUID

	return c.JSON(fiber.Map{
		"message": "Задача успешно удалена",
	})
}

// RetryImageTask повторно запускает обработку изображения
// @Summary Повторная обработка изображения
// @Description Перезапускает неудачную задачу обработки изображения
// @Tags admin-images
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID задачи"
// @Success 200 {object} map[string]interface{} "Задача поставлена на повтор"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 403 {object} map[string]interface{} "Нет прав доступа"
// @Failure 404 {object} map[string]interface{} "Задача не найдена"
// @Failure 400 {object} map[string]interface{} "Задача не может быть повторена"
// @Router /admin/images/{id}/retry [post]
func (handler *AdminHandler) RetryImageTask(c *fiber.Ctx) error {
	imageID := c.Params("id")

	imageUUID, err := uuid.Parse(imageID)
	if err != nil {
		handler.deps.Logger.Error().Err(err).Str("image_id", imageID).Msg(ErrInvalidID.Error())
		return c.Status(ErrInvalidID.HTTPStatus).JSON(fiber.Map{
			"error": ErrInvalidID.Error(),
		})
	}

	// TODO: Повторить задачу через ImageRepository
	_ = imageUUID

	return c.JSON(fiber.Map{
		"message": "Задача поставлена на повторную обработку",
	})
}
