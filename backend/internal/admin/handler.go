package admin

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/skr1ms/mosaic/internal/coupon"
	"github.com/skr1ms/mosaic/internal/partner"
	"github.com/skr1ms/mosaic/pkg/bcrypt"
	"github.com/skr1ms/mosaic/pkg/jwt"
	"github.com/skr1ms/mosaic/pkg/middleware"
	"github.com/skr1ms/mosaic/pkg/utils"
	"gorm.io/gorm"
)

type AdminHandlerDeps struct {
	adminRepository   *AdminRepository
	partnerRepository *partner.PartnerRepository
	couponRepository  *coupon.CouponRepository
	jwtService        *jwt.JWT
}

type AdminHandler struct {
	fiber.Router
	deps *AdminHandlerDeps
}

func NewAdminHandler(router fiber.Router, db *gorm.DB, jwtService *jwt.JWT) {
	handler := &AdminHandler{
		Router: router,
		deps: &AdminHandlerDeps{
			adminRepository:   NewAdminRepository(db),
			partnerRepository: partner.NewPartnerRepository(db),
			couponRepository:  coupon.NewCouponRepository(db),
			jwtService:        jwtService,
		},
	}

	adminRoutes := handler.Group("/admin")
	// Публичные endpoints (без JWT)
	adminRoutes.Post("/login", handler.Login)          // авторизация ✔
	adminRoutes.Post("/refresh", handler.RefreshToken) // обновление токенов ✔

	// Защищенные endpoints (требуют JWT + admin роль)
	protected := adminRoutes.Use(middleware.JWTMiddleware(jwtService), middleware.AdminOnly())

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
	protected.Get("/coupons", handler.GetCoupons)                            // получение списка купонов ✔
	protected.Post("/coupons", handler.CreateCoupons)                        // создание купонов ✔
	protected.Get("/coupons/export", handler.ExportCoupons)                  // экспорт купонов
	protected.Get("/coupons/:id", handler.GetCoupon)                         // получение информации о купоне ✔
	protected.Patch("/coupons/:id/reset", handler.ResetCoupon)               // сброс купона ✔
	protected.Delete("/coupons/:id", handler.DeleteCoupon)                   // удаление купона ✔
	protected.Get("/coupons/:id/materials", handler.DownloadCouponMaterials) // скачивание материалов купона

	// Статистика и аналитика
	protected.Get("/statistics", handler.GetStatistics)                  // получение статистики ✔
	protected.Get("/statistics/partners", handler.GetPartnersStatistics) // получение статистики партнеров ✔

	// Профиль администратора
	protected.Put("/profile/password", handler.ChangePassword) // изменение пароля администратора ✔
}

// Login обрабатывает авторизацию администратора и генерирует JWT токены
// @Summary Авторизация администратора
// @Description Авторизация администратора по логину и паролю
// @Tags admin-auth
// @Accept json
// @Produce json
// @Param credentials body LoginRequest true "Учетные данные для входа"
// @Success 200 {object} map[string]interface{} "Успешная авторизация"
// @Failure 400 {object} map[string]interface{} "Ошибка в запросе"
// @Failure 401 {object} map[string]interface{} "Неверные учетные данные"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /admin/login [post]
func (handler *AdminHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	admin, err := handler.deps.adminRepository.GetByLogin(req.Login)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	if !bcrypt.CheckPassword(req.Password, admin.Password) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	tokenPair, err := handler.deps.jwtService.CreateTokenPair(admin.ID, admin.Login, "admin")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate tokens"})
	}

	if err := handler.deps.adminRepository.UpdateLastLogin(admin.ID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update login time"})
	}

	return c.JSON(fiber.Map{
		"message": "Login successful",
		"admin": fiber.Map{
			"id":    admin.ID,
			"login": admin.Login,
			"role":  "admin",
		},
		"access_token":  tokenPair.AccessToken,
		"refresh_token": tokenPair.RefreshToken,
		"expires_in":    tokenPair.ExpiresIn,
	})
}

// RefreshToken обновляет токены используя refresh токен
// @Summary Обновление токенов
// @Description Обновляет access и refresh токены используя refresh токен
// @Tags admin-auth
// @Accept json
// @Produce json
// @Param refresh body RefreshTokenRequest true "Refresh токен"
// @Success 200 {object} map[string]interface{} "Новые токены"
// @Failure 400 {object} map[string]interface{} "Ошибка в запросе"
// @Failure 401 {object} map[string]interface{} "Неверный или истекший refresh токен"
// @Router /admin/refresh [post]
func (handler *AdminHandler) RefreshToken(c *fiber.Ctx) error {
	var req RefreshTokenRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	tokenPair, err := handler.deps.jwtService.RefreshTokens(req.RefreshToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid or expired refresh token"})
	}

	return c.JSON(fiber.Map{
		"access_token":  tokenPair.AccessToken,
		"refresh_token": tokenPair.RefreshToken,
		"expires_in":    tokenPair.ExpiresIn,
	})
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
	var req CreateAdminRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	existingAdmin, err := handler.deps.adminRepository.GetByLogin(req.Login)
	if err == nil && existingAdmin != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "Admin with this login already exists",
		})
	}

	hashedPassword, err := bcrypt.HashPassword(req.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to hash password"})
	}

	admin := &Admin{
		Login:    req.Login,
		Password: hashedPassword,
	}

	if err := handler.deps.adminRepository.Create(admin); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create admin"})
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
	admins, err := handler.deps.adminRepository.GetAll()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch admins"})
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
	claims, err := jwt.GetClaimsFromFiberContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	admin, err := handler.deps.adminRepository.GetByID(claims.UserID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get admin info"})
	}

	// Получаем общую статистику
	var totalCoupons int64
	if coupons, err := handler.deps.couponRepository.GetAll(); err == nil {
		totalCoupons = int64(len(coupons))
	}

	var activatedCoupons int64
	if coupons, err := handler.deps.couponRepository.GetFiltered(map[string]interface{}{"status": "used"}); err == nil {
		activatedCoupons = int64(len(coupons))
	}

	var totalPartners int64
	if partners, err := handler.deps.partnerRepository.GetAll(); err == nil {
		totalPartners = int64(len(partners))
	}

	var activePartners int64
	if partners, err := handler.deps.partnerRepository.GetActivePartners(); err == nil {
		activePartners = int64(len(partners))
	}

	// Процент активации купонов
	activationRate := float64(0)
	if totalCoupons > 0 {
		activationRate = (float64(activatedCoupons) / float64(totalCoupons)) * 100
	}

	// Получаем недавнюю активность (последние 5 активированных купонов)
	recentCoupons, _ := handler.deps.couponRepository.GetRecentActivated(5)
	recentActivity := make([]fiber.Map, 0)

	for _, coupon := range recentCoupons {
		partnerName := "Собственный"
		if coupon.PartnerID != uuid.Nil {
			if partner, err := handler.deps.partnerRepository.GetByID(coupon.PartnerID); err == nil {
				partnerName = partner.BrandName
			}
		}

		recentActivity = append(recentActivity, fiber.Map{
			"coupon_code":    coupon.Code,
			"partner_name":   partnerName,
			"activated_at":   coupon.UsedAt,
			"purchase_email": coupon.PurchaseEmail,
		})
	}

	alerts := make([]fiber.Map, 0)

	// Проверяем заблокированных партнеров
	if blockedPartners, err := handler.deps.partnerRepository.Search("", "blocked"); err == nil && len(blockedPartners) > 0 {
		alerts = append(alerts, fiber.Map{
			"type":    "warning",
			"title":   "Заблокированные партнеры",
			"message": fmt.Sprintf("У вас %d заблокированных партнеров", len(blockedPartners)),
			"action":  "/admin/partners?status=blocked",
		})
	}

	// Проверяем низкий процент активации
	if activationRate < 10 && totalCoupons > 100 {
		alerts = append(alerts, fiber.Map{
			"type":    "info",
			"title":   "Низкий процент активации",
			"message": fmt.Sprintf("Процент активации купонов составляет %.1f%%", activationRate),
			"action":  "/admin/statistics",
		})
	}

	return c.JSON(fiber.Map{
		"admin_info": fiber.Map{
			"id":         admin.ID,
			"login":      admin.Login,
			"last_login": admin.LastLogin,
		},
		"statistics": fiber.Map{
			"total_coupons":     totalCoupons,
			"activated_coupons": activatedCoupons,
			"activation_rate":   activationRate,
			"total_partners":    totalPartners,
			"active_partners":   activePartners,
			"unused_coupons":    totalCoupons - activatedCoupons,
		},
		"recent_activity": recentActivity,
		"alerts":          alerts,
		"quick_actions": []fiber.Map{
			{
				"title": "Создать партнера",
				"url":   "/admin/partners",
				"icon":  "user-plus",
			},
			{
				"title": "Сгенерировать купоны",
				"url":   "/admin/coupons",
				"icon":  "ticket",
			},
			{
				"title": "Посмотреть статистику",
				"url":   "/admin/statistics",
				"icon":  "chart-bar",
			},
			{
				"title": "Экспорт купонов",
				"url":   "/admin/coupons/export",
				"icon":  "download",
			},
		},
	})
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

	var partners []*partner.Partner
	var err error

	if search != "" || status != "" {
		partners, err = handler.deps.partnerRepository.Search(search, status)
	} else {
		partners, err = handler.deps.partnerRepository.GetAll()
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch partners"})
	}

	result := make([]fiber.Map, len(partners))
	for i, partner := range partners {
		// Получаем статистику партнера
		totalCoupons, _ := handler.deps.couponRepository.CountByPartnerID(partner.ID)
		activatedCoupons, _ := handler.deps.couponRepository.CountActivatedByPartnerID(partner.ID)

		result[i] = fiber.Map{
			"id":                partner.ID,
			"login":             partner.Login,
			"last_login":        partner.LastLogin,
			"created_at":        partner.CreatedAt,
			"updated_at":        partner.UpdatedAt,
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
			"total_coupons":     totalCoupons,
			"activated_coupons": activatedCoupons,
		}
	}

	return c.JSON(fiber.Map{
		"partners": result,
		"total":    len(result),
	})
}

// CreatePartner создает нового партнера
// @Summary Создание партнера
// @Description Создает нового партнера
// @Tags admin-partners
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 201 {object} map[string]interface{} "Партнер создан"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 403 {object} map[string]interface{} "Нет прав доступа"
// @Router /admin/partners [post]
func (handler *AdminHandler) CreatePartner(c *fiber.Ctx) error {
	var req partner.CreatePartnerRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	// Проверяем, существует ли уже партнер с таким логином
	existingPartner, err := handler.deps.partnerRepository.GetByLogin(req.Login)
	if err == nil && existingPartner != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "Partner with this login already exists",
		})
	}

	// Проверяем, существует ли уже партнер с таким кодом
	existingPartnerByCode, err := handler.deps.partnerRepository.GetByPartnerCode(req.PartnerCode)
	if err == nil && existingPartnerByCode != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "Partner with this partner code already exists",
		})
	}

	// Проверяем, существует ли уже партнер с таким доменом
	existingPartnerByDomain, err := handler.deps.partnerRepository.GetByDomain(req.Domain)
	if err == nil && existingPartnerByDomain != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "Partner with this domain already exists",
		})
	}

	hashedPassword, err := bcrypt.HashPassword(req.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to hash password"})
	}

	partner := &partner.Partner{
		Login:           req.Login,
		Password:        hashedPassword,
		PartnerCode:     req.PartnerCode,
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

	if err := handler.deps.partnerRepository.Create(partner); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create partner"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":           partner.ID,
		"login":        partner.Login,
		"partner_code": partner.PartnerCode,
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	partnerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get partner"})
	}

	partner, err := handler.deps.partnerRepository.GetByID(partnerID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get partner"})
	}

	utils.UpdatePartnerData(partner, &req)

	if err := handler.deps.partnerRepository.Update(partner); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update partner"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"id":           partner.ID,
		"login":        partner.Login,
		"partner_code": partner.PartnerCode,
	})
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

	partner, err := handler.deps.partnerRepository.GetByID(partnerID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Partner not found"})
	}

	totalCoupons, _ := handler.deps.couponRepository.CountByPartnerID(partnerID)
	activatedCoupons, _ := handler.deps.couponRepository.CountActivatedByPartnerID(partnerID)

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

	if err := handler.deps.partnerRepository.Delete(partnerID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete partner"})
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

	coupons, err := handler.deps.couponRepository.GetFiltered(filters)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch coupons"})
	}

	// Добавляем информацию о партнерах
	result := make([]fiber.Map, len(coupons))
	for i, coupon := range coupons {
		partnerName := "Собственный"
		if coupon.PartnerID != uuid.Nil {
			if partner, err := handler.deps.partnerRepository.GetByID(coupon.PartnerID); err == nil {
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
	var partnerCode int16 = 0 // Для собственных купонов
	if req.PartnerID != uuid.Nil {
		partner, err := handler.deps.partnerRepository.GetByID(req.PartnerID)
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
		code := utils.GenerateCouponCode(partnerCode)

		// Проверяем уникальность
		for {
			if _, err := handler.deps.couponRepository.GetByCode(code); err != nil {
				break
			}
			code = utils.GenerateCouponCode(partnerCode)
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
	if err := handler.deps.couponRepository.CreateBatch(coupons); err != nil {
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
	coupon, err := handler.deps.couponRepository.GetByID(couponID)
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
	if err := handler.deps.couponRepository.Delete(couponID); err != nil {
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
	handler.deps.couponRepository.GetAll()
	if coupons, err := handler.deps.couponRepository.GetAll(); err == nil {
		totalCoupons = int64(len(coupons))
	}

	// Количество активированных купонов
	var activatedCoupons int64
	if coupons, err := handler.deps.couponRepository.GetFiltered(map[string]interface{}{"status": "used"}); err == nil {
		activatedCoupons = int64(len(coupons))
	}

	// Количество активных партнеров
	var activePartners int64
	if partners, err := handler.deps.partnerRepository.GetActivePartners(); err == nil {
		activePartners = int64(len(partners))
	}

	// Всего партнеров
	var totalPartners int64
	if partners, err := handler.deps.partnerRepository.GetAll(); err == nil {
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	claims, err := jwt.GetClaimsFromFiberContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	adminUUID := claims.UserID

	admin, err := handler.deps.adminRepository.GetByID(adminUUID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get admin"})
	}

	if !bcrypt.CheckPassword(req.CurrentPassword, admin.Password) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Current password is incorrect"})
	}

	hashedPassword, err := bcrypt.HashPassword(req.NewPassword)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to hash password"})
	}

	if err := handler.deps.adminRepository.UpdatePassword(adminUUID, hashedPassword); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update password"})
	}

	return c.JSON(fiber.Map{"message": "Password changed successfully"})
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid partner ID"})
	}

	partner, err := handler.deps.partnerRepository.GetByID(partnerID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Partner not found"})
	}

	// Получаем информацию о партнере
	totalCoupons, err := handler.deps.couponRepository.CountByPartnerID(partnerID)
	if err != nil {
		totalCoupons = 0
	}

	activatedCoupons, err := handler.deps.couponRepository.CountActivatedByPartnerID(partnerID)
	if err != nil {
		activatedCoupons = 0
	}

	purchasedCoupons, err := handler.deps.couponRepository.CountPurchasedByPartnerID(partnerID)
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid partner ID"})
	}

	if err := handler.deps.partnerRepository.UpdateStatus(partnerID, "blocked"); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to block partner"})
	}

	if err := handler.deps.couponRepository.UpdateStatusByPartnerID(partnerID, true); err != nil {
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid partner ID"})
	}

	if err := handler.deps.partnerRepository.UpdateStatus(partnerID, "active"); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to unblock partner"})
	}

	if err := handler.deps.couponRepository.UpdateStatusByPartnerID(partnerID, false); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to unblock coupons"})
	}

	return c.JSON(fiber.Map{"message": "Partner unblocked successfully"})
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
	totalCoupons, err := handler.deps.couponRepository.CountByPartnerID(partnerID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get statistics"})
	}

	activatedCoupons, err := handler.deps.couponRepository.CountActivatedByPartnerID(partnerID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get statistics"})
	}

	purchasedCoupons, err := handler.deps.couponRepository.CountPurchasedByPartnerID(partnerID)
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

	coupon, err := handler.deps.couponRepository.GetByID(couponID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Coupon not found"})
	}

	// Получаем информацию о партнере
	var partnerName string = "Собственный"
	if coupon.PartnerID != uuid.Nil {
		partner, err := handler.deps.partnerRepository.GetByID(coupon.PartnerID)
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

	if err := handler.deps.couponRepository.Reset(couponID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to reset coupon"})
	}

	return c.JSON(fiber.Map{"message": "Coupon reset successfully"})
}

// ExportCoupons экспортирует список купонов в текстовый файл
// @Summary Экспорт купонов
// @Description Экспортирует список купонов в текстовый файл
// @Tags admin-coupons
// @Produce text/plain
// @Security BearerAuth
// @Param partner_id query string false "ID партнера для фильтрации"
// @Param status query string false "Статус для фильтрации (new/used)"
// @Param size query string false "Размер для фильтрации"
// @Param style query string false "Стиль для фильтрации"
// @Success 200 {string} string "Текстовый файл с купонами"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 403 {object} map[string]interface{} "Нет прав доступа"
// @Router /admin/coupons/export [get]
func (handler *AdminHandler) ExportCoupons(c *fiber.Ctx) error {
	// Получаем параметры фильтрации
	filters := map[string]interface{}{}

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

	coupons, err := handler.deps.couponRepository.GetFiltered(filters)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch coupons"})
	}

	// Генерируем текстовый файл
	var content string
	for _, coupon := range coupons {
		content += coupon.Code + "\n"
	}

	c.Set("Content-Type", "text/plain")
	c.Set("Content-Disposition", "attachment; filename=coupons.txt")

	return c.SendString(content)
}

// DownloadCouponMaterials скачивает материалы погашенного купона
// @Summary Скачивание материалов купона
// @Description Скачивает материалы погашенного купона (оригинал, превью, схема)
// @Tags admin-coupons
// @Produce application/zip
// @Security BearerAuth
// @Param id path string true "ID купона"
// @Success 200 {string} binary "ZIP архив с материалами"
// @Failure 400 {object} map[string]interface{} "Неверный ID или купон не погашен"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 403 {object} map[string]interface{} "Нет прав доступа"
// @Failure 404 {object} map[string]interface{} "Купон не найден"
// @Router /admin/coupons/{id}/materials [get]
func (handler *AdminHandler) DownloadCouponMaterials(c *fiber.Ctx) error {
	couponID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid coupon ID"})
	}

	coupon, err := handler.deps.couponRepository.GetByID(couponID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Coupon not found"})
	}

	if coupon.Status != "used" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Coupon is not used"})
	}

	// TODO: Здесь должна быть логика создания ZIP архива с материалами
	// Пока возвращаем заглушку
	return c.JSON(fiber.Map{
		"message": "Materials download functionality not implemented yet",
		"files": []string{
			"original_image.jpg",
			"preview.jpg",
			"schema.pdf",
		},
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
	partners, err := handler.deps.partnerRepository.GetAll()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch partners"})
	}

	var partnersStats []fiber.Map
	for _, partner := range partners {
		totalCoupons, _ := handler.deps.couponRepository.CountByPartnerID(partner.ID)
		activatedCoupons, _ := handler.deps.couponRepository.CountActivatedByPartnerID(partner.ID)
		purchasedCoupons, _ := handler.deps.couponRepository.CountPurchasedByPartnerID(partner.ID)

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
