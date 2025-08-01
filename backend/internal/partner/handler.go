package partner

import (
	"context"
	"fmt"
	"strings"

	"github.com/skr1ms/mosaic/pkg/middleware"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/skr1ms/mosaic/config"
	"github.com/skr1ms/mosaic/internal/coupon"
	"github.com/skr1ms/mosaic/pkg/email"
	"github.com/skr1ms/mosaic/pkg/jwt"
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
	deps *PartnerHandlerDeps
}

func NewPartnerHandler(router fiber.Router, deps *PartnerHandlerDeps) {
	handler := &PartnerHandler{
		Router: router,
		deps:   deps,
	}

	partnerRoutes := router.Group("/partner")

	// Публичные endpoints (без JWT)
	partnerRoutes.Post("/forgot", handler.ForgotPassword) // Запрос на сброс пароля
	partnerRoutes.Post("/reset", handler.ResetPassword)   // Сброс пароля

	// Защищенные endpoints (требуют JWT + partner роль)
	protected := partnerRoutes.Use(middleware.JWTMiddleware(deps.JwtService), middleware.PartnerOnly())
	protected.Get("/dashboard", handler.GetDashboard)              // Дашборд партнера
	protected.Get("/profile", handler.GetProfile)                  // Профиль партнера
	protected.Put("/profile", handler.UpdateProfile)               // Обновление профиля партнера (только для чтения в партнерской панели)
	protected.Put("/update/password", handler.UpdatePassword)      // Обновление пароля партнера
	protected.Get("/coupons", handler.GetMyCoupons)                // Купоны партнера
	protected.Get("/coupons/export", handler.ExportCoupons)        // Экспорт купонов партнера
	protected.Get("/statistics", handler.GetMyStatistics)          // Статистика партнера
	protected.Get("/statistics/sales", handler.GetSalesStatistics) // Статистика продаж
	protected.Get("/statistics/usage", handler.GetUsageStatistics) // Статистика использования
}

// GetDashboard возвращает данные для дашборда партнера
//
//	@Summary		Дашборд партнера
//	@Description	Возвращает данные для главной страницы партнера
//	@Tags			partner-dashboard
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	map[string]interface{}	"Данные дашборда"
//	@Failure		401	{object}	map[string]interface{}	"Не авторизован"
//	@Failure		403	{object}	map[string]interface{}	"Нет прав доступа"
//	@Router			/partner/dashboard [get]
func (handler *PartnerHandler) GetDashboard(c *fiber.Ctx) error {
	// TODO: Реализовать дашборд партнера
	return c.JSON(fiber.Map{"message": "Partner dashboard"})
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

	// Экспортируем купоны
	content, filename, err := handler.deps.PartnerService.ExportCoupons(claims.UserID, "new", format)
	if err != nil {
		log.Error().Err(err).Msg("Failed to export coupons")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to export coupons"})
	}

	// Устанавливаем заголовки для автоматического скачивания
	c.Set("Content-Type", format)
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Set("Cache-Control", "no-cache")

	return c.SendString(content)
}

// GetMyStatistics возвращает статистику партнера
//
//	@Summary		Статистика партнера
//	@Description	Возвращает общую статистику текущего партнера
//	@Tags			partner-statistics
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	map[string]interface{}	"Статистика партнера"
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

	return c.JSON(fiber.Map{
		"message":    "Partner statistics",
		"partner_id": claims.UserID,
	})
}

// GetSalesStatistics возвращает статистику продаж партнера
//
//	@Summary		Статистика продаж
//	@Description	Возвращает статистику продаж текущего партнера
//	@Tags			partner-statistics
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	map[string]interface{}	"Статистика продаж"
//	@Failure		401	{object}	map[string]interface{}	"Не авторизован"
//	@Failure		403	{object}	map[string]interface{}	"Нет прав доступа"
//	@Router			/partner/statistics/sales [get]
func (handler *PartnerHandler) GetSalesStatistics(c *fiber.Ctx) error {
	// TODO: Реализовать статистику продаж купонов
	return c.JSON(fiber.Map{"message": "Sales statistics"})
}

// GetUsageStatistics возвращает статистику использования купонов
//
//	@Summary		Статистика использования купонов
//	@Description	Возвращает статистику использования купонов партнера
//	@Tags			partner-statistics
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	map[string]interface{}	"Статистика использования"
//	@Failure		401	{object}	map[string]interface{}	"Не авторизован"
//	@Failure		403	{object}	map[string]interface{}	"Нет прав доступа"
//	@Router			/partner/statistics/usage [get]
func (handler *PartnerHandler) GetUsageStatistics(c *fiber.Ctx) error {
	// TODO: Реализовать 
	return c.JSON(fiber.Map{"message": "Usage statistics"})
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
