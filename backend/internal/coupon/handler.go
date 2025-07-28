package coupon

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type CouponHandlerDeps struct {
	CouponService *CouponService
	Logger        *zerolog.Logger
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

	api := handler.Group("/coupons")
	api.Get("/", handler.GetCoupons)                             // Получение списка купонов с фильтрацией
	api.Get("/:id", handler.GetCouponDetails)                    // Получение деталей купона
	api.Get("/paginated", handler.GetCouponsPaginated)           // Пагинация списка купонов
	api.Get("/:id", handler.GetCouponByID)                       // Получение купона по ID
	api.Get("/code/:code", handler.GetCouponByCode)              // Получение купона по коду
	api.Post("/code/:code/validate", handler.ValidateCoupon)     // Валидация купона по коду
	api.Put("/:id/activate", handler.ActivateCoupon)             // Активация купона
	api.Put("/:id/reset", handler.ResetCoupon)                   // Сброс купона в исходное состояние
	api.Put("/:id/send-schema", handler.SendSchema)              // Отправка схемы купона на email
	api.Put("/:id/purchase", handler.MarkAsPurchased)            // Пометка купона как купленного
	api.Get("/export", handler.ExportCoupons)                    // Экспорт купонов в zip файл
	api.Get("/statistics", handler.GetStatistics)                // Получение статистики по купонам
	api.Get("/partner/:partner_id", handler.GetCouponsByPartner) // Получение купонов по ID партнера
}

// GetCoupons возвращает список купонов с фильтрацией
// @Summary Список купонов с фильтрацией
// @Description Возвращает список купонов с возможностью фильтрации по коду, статусу, размеру, стилю и партнеру
// @Tags coupons
// @Produce json
// @Param code query string false "Код купона для поиска"
// @Param status query string false "Статус купона (new, used)"
// @Param size query string false "Размер купона"
// @Param style query string false "Стиль купона"
// @Param partner_id query string false "ID партнера"
// @Success 200 {array} map[string]interface{} "Список купонов"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /coupons [get]
func (handler *CouponHandler) GetCoupons(c *fiber.Ctx) error {
	// Получаем параметры запроса
	code := c.Query("code")
	status := c.Query("status")
	size := c.Query("size")
	style := c.Query("style")
	partnerIDStr := c.Query("partner_id")

	// Парсим ID партнера
	var partnerID *uuid.UUID
	if partnerIDStr != "" {
		if id, err := uuid.Parse(partnerIDStr); err == nil {
			partnerID = &id
		}
	}

	// Получаем купоны
	coupons, err := handler.deps.CouponService.SearchCoupons(code, status, size, style, partnerID)
	if err != nil {
		if apiErr, ok := GetAPIError(err); ok {
			return c.Status(apiErr.HTTPStatus).JSON(fiber.Map{"error": apiErr.Error()})
		}
		return c.Status(ErrInternalServerError.HTTPStatus).JSON(fiber.Map{"error": ErrInternalServerError.Error()})
	}

	return c.JSON(coupons)
}

// GetCouponDetails возвращает детали купона
// @Summary Детали купона
// @Description Возвращает подробную информацию о купоне
// @Tags coupons
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID купона"
// @Success 200 {object} map[string]interface{} "Детали купона"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 403 {object} map[string]interface{} "Нет прав доступа"
// @Router /coupons/{id} [get]
func (handler *CouponHandler) GetCouponDetails(c *fiber.Ctx) error {
	handler.deps.Logger.Error().Msg(ErrCouponNotFound.Message)
	return c.Status(ErrCouponNotFound.HTTPStatus).JSON(fiber.Map{"error": ErrCouponNotFound.Error()})
}

// GetCouponByID возвращает купон по ID
// @Summary Получение купона по ID
// @Description Возвращает детальную информацию о купоне по его ID
// @Tags coupons
// @Produce json
// @Param id path string true "ID купона"
// @Success 200 {object} map[string]interface{} "Информация о купоне"
// @Failure 400 {object} map[string]interface{} "Неверный ID купона"
// @Failure 404 {object} map[string]interface{} "Купон не найден"
// @Router /coupons/{id} [get]
func (handler *CouponHandler) GetCouponByID(c *fiber.Ctx) error {
	// Получаем ID купона
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		handler.deps.Logger.Error().Err(err).Msg(ErrInvalidCouponID.Message)
		return c.Status(ErrInvalidCouponID.HTTPStatus).JSON(fiber.Map{"error": ErrInvalidCouponID.Error()})
	}

	// Получаем купон
	coupon, err := handler.deps.CouponService.GetCouponByID(id)
	if err != nil {
		if apiErr, ok := GetAPIError(err); ok {
			return c.Status(apiErr.HTTPStatus).JSON(fiber.Map{"error": apiErr.Error()})
		}
		return c.Status(ErrInternalServerError.HTTPStatus).JSON(fiber.Map{"error": ErrInternalServerError.Error()})
	}

	return c.JSON(coupon)
}

// GetCouponByCode возвращает купон по коду
// @Summary Получение купона по коду
// @Description Возвращает детальную информацию о купоне по его коду
// @Tags coupons
// @Produce json
// @Param code path string true "Код купона"
// @Success 200 {object} map[string]interface{} "Информация о купоне"
// @Failure 404 {object} map[string]interface{} "Купон не найден"
// @Router /coupons/code/{code} [get]
func (handler *CouponHandler) GetCouponByCode(c *fiber.Ctx) error {
	// Получаем код купона
	code := c.Params("code")

	// Получаем купон
	coupon, err := handler.deps.CouponService.GetCouponByCode(code)
	if err != nil {
		if apiErr, ok := GetAPIError(err); ok {
			return c.Status(apiErr.HTTPStatus).JSON(fiber.Map{"error": apiErr.Error()})
		}
		handler.deps.Logger.Error().Err(err).Msg(ErrInternalServerError.Error())
		return c.Status(ErrInternalServerError.HTTPStatus).JSON(fiber.Map{"error": ErrInternalServerError.Error()})
	}

	return c.JSON(coupon)
}

// ActivateCoupon активирует купон
// @Summary Активация купона
// @Description Активирует купон, изменяя его статус на 'used' и добавляя ссылки на изображения
// @Tags coupons
// @Accept json
// @Produce json
// @Param id path string true "ID купона"
// @Param request body ActivateCouponRequest true "Ссылки на изображения"
// @Success 200 {object} map[string]interface{} "Купон активирован"
// @Failure 400 {object} map[string]interface{} "Неверный ID купона"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /coupons/{id}/activate [put]
func (handler *CouponHandler) ActivateCoupon(c *fiber.Ctx) error {
	// Получаем ID купона
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		handler.deps.Logger.Error().Err(err).Msg(ErrInvalidCouponID.Error())
		return c.Status(ErrInvalidCouponID.HTTPStatus).JSON(fiber.Map{"error": ErrInvalidCouponID.Error()})
	}

	var req ActivateCouponRequest

	// Парсим запрос
	if err := c.BodyParser(&req); err != nil {
		handler.deps.Logger.Error().Err(err).Msg(ErrInvalidRequestBody.Message)
		return c.Status(ErrInvalidRequestBody.HTTPStatus).JSON(fiber.Map{"error": ErrInvalidRequestBody.Error()})
	}

	// Активируем купон
	if err := handler.deps.CouponService.ActivateCoupon(id, *req.OriginalImageURL, *req.PreviewURL, *req.SchemaURL); err != nil {
		if apiErr, ok := GetAPIError(err); ok {
			return c.Status(apiErr.HTTPStatus).JSON(fiber.Map{"error": apiErr.Error()})
		}
		handler.deps.Logger.Error().Err(err).Msg(ErrInternalServerError.Error())
		return c.Status(ErrInternalServerError.HTTPStatus).JSON(fiber.Map{"error": ErrInternalServerError.Error()})
	}

	return c.JSON(fiber.Map{"message": "Coupon activated successfully"})
}

// ResetCoupon сбрасывает купон в исходное состояние
// @Summary Сброс купона
// @Description Сбрасывает купон в исходное состояние (статус 'new')
// @Tags coupons
// @Produce json
// @Param id path string true "ID купона"
// @Success 200 {object} map[string]interface{} "Купон сброшен"
// @Failure 400 {object} map[string]interface{} "Неверный ID купона"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /coupons/{id}/reset [put]
func (handler *CouponHandler) ResetCoupon(c *fiber.Ctx) error {
	// Получаем ID купона
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		handler.deps.Logger.Error().Err(err).Msg(ErrInvalidCouponID.Message)
		return c.Status(ErrInvalidCouponID.HTTPStatus).JSON(fiber.Map{"error": ErrInvalidCouponID.Error()})
	}

	// Сбрасываем купон
	if err := handler.deps.CouponService.ResetCoupon(id); err != nil {
		if apiErr, ok := GetAPIError(err); ok {
			return c.Status(apiErr.HTTPStatus).JSON(fiber.Map{"error": apiErr.Error()})
		}
		handler.deps.Logger.Error().Err(err).Msg(ErrInternalServerError.Error())
		return c.Status(ErrInternalServerError.HTTPStatus).JSON(fiber.Map{"error": ErrInternalServerError.Error()})
	}

	return c.JSON(fiber.Map{"message": "Coupon reset successfully"})
}

// SendSchema отправляет схему на email
// @Summary Отправка схемы на email
// @Description Отправляет схему купона на указанный email адрес
// @Tags coupons
// @Accept json
// @Produce json
// @Param id path string true "ID купона"
// @Param request body SendSchemaRequest true "Email для отправки"
// @Success 200 {object} map[string]interface{} "Схема отправлена"
// @Failure 400 {object} map[string]interface{} "Неверный ID купона"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /coupons/{id}/send-schema [put]
func (handler *CouponHandler) SendSchema(c *fiber.Ctx) error {
	// Получаем ID купона
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		handler.deps.Logger.Error().Err(err).Msg(ErrInvalidCouponID.Message)
		return c.Status(ErrInvalidCouponID.HTTPStatus).JSON(fiber.Map{"error": ErrInvalidCouponID.Error()})
	}

	var req SendSchemaRequest

	// Парсим запрос
	if err := c.BodyParser(&req); err != nil {
		handler.deps.Logger.Error().Err(err).Msg(ErrInvalidRequestBody.Message)
		return c.Status(ErrInvalidRequestBody.HTTPStatus).JSON(fiber.Map{"error": ErrInvalidRequestBody.Error()})
	}

	// Отправляем схему
	if err := handler.deps.CouponService.SendSchema(id, req.Email); err != nil {
		if apiErr, ok := GetAPIError(err); ok {
			return c.Status(apiErr.HTTPStatus).JSON(fiber.Map{"error": apiErr.Error()})
		}
		handler.deps.Logger.Error().Err(err).Msg(ErrInternalServerError.Error())
		return c.Status(ErrInternalServerError.HTTPStatus).JSON(fiber.Map{"error": ErrInternalServerError.Error()})
	}

	return c.JSON(fiber.Map{"message": "Schema sent successfully"})
}

// MarkAsPurchased помечает купон как купленный
// @Summary Пометка купона как купленного
// @Description Помечает купон как купленный онлайн с указанием email покупателя
// @Tags coupons
// @Accept json
// @Produce json
// @Param id path string true "ID купона"
// @Param request body MarkAsPurchasedRequest true "Email покупателя"
// @Success 200 {object} map[string]interface{} "Купон помечен как купленный"
// @Failure 400 {object} map[string]interface{} "Неверный ID купона"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /coupons/{id}/purchase [put]
func (handler *CouponHandler) MarkAsPurchased(c *fiber.Ctx) error {
	// Получаем ID купона
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		handler.deps.Logger.Error().Err(err).Msg(ErrInvalidCouponID.Message)
		return c.Status(ErrInvalidCouponID.HTTPStatus).JSON(fiber.Map{"error": ErrInvalidCouponID.Error()})
	}

	var req MarkAsPurchasedRequest

	// Парсим запрос
	if err := c.BodyParser(&req); err != nil {
		handler.deps.Logger.Error().Err(err).Msg(ErrInvalidRequestBody.Message)
		return c.Status(ErrInvalidRequestBody.HTTPStatus).JSON(fiber.Map{"error": ErrInvalidRequestBody.Error()})
	}

	// Помечаем купон как купленный
	if err := handler.deps.CouponService.MarkAsPurchased(id, req.PurchaseEmail); err != nil {
		if apiErr, ok := GetAPIError(err); ok {
			return c.Status(apiErr.HTTPStatus).JSON(fiber.Map{"error": apiErr.Error()})
		}
		handler.deps.Logger.Error().Err(err).Msg(ErrInternalServerError.Error())
		return c.Status(ErrInternalServerError.HTTPStatus).JSON(fiber.Map{"error": ErrInternalServerError.Error()})
	}

	return c.JSON(fiber.Map{"message": "Coupon marked as purchased"})
}

// GetStatistics возвращает статистику по купонам
// @Summary Статистика по купонам
// @Description Возвращает статистику по купонам с возможностью фильтрации по партнеру
// @Tags coupons
// @Produce json
// @Param partner_id query string false "ID партнера для фильтрации"
// @Success 200 {object} map[string]interface{} "Статистика по купонам"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /coupons/statistics [get]
func (handler *CouponHandler) GetStatistics(c *fiber.Ctx) error {
	// Получаем ID партнера
	partnerIDStr := c.Query("partner_id")

	// Парсим ID партнера
	var partnerID *uuid.UUID
	if partnerIDStr != "" {
		if id, err := uuid.Parse(partnerIDStr); err == nil {
			partnerID = &id
		}
	}

	// Получаем статистику
	stats, err := handler.deps.CouponService.GetStatistics(partnerID)
	if err != nil {
		if apiErr, ok := GetAPIError(err); ok {
			return c.Status(apiErr.HTTPStatus).JSON(fiber.Map{"error": apiErr.Error()})
		}
		handler.deps.Logger.Error().Err(err).Msg(ErrInternalServerError.Error())
		return c.Status(ErrInternalServerError.HTTPStatus).JSON(fiber.Map{"error": ErrInternalServerError.Error()})
	}

	return c.JSON(stats)
}

// ValidateCoupon проверяет валидность купона без получения полной информации
// @Summary Валидация купона
// @Description Проверяет существование и доступность купона для активации
// @Tags coupons
// @Produce json
// @Param code path string true "Код купона"
// @Success 200 {object} map[string]interface{} "Информация о статусе купона"
// @Failure 404 {object} map[string]interface{} "Купон не найден"
// @Router /coupons/code/{code}/validate [post]
func (handler *CouponHandler) ValidateCoupon(c *fiber.Ctx) error {
	// Получаем код купона
	code := c.Params("code")

	// Валидируем купон
	validationResult, err := handler.deps.CouponService.ValidateCoupon(code)
	if err != nil {
		if apiErr, ok := GetAPIError(err); ok {
			return c.Status(apiErr.HTTPStatus).JSON(fiber.Map{"error": apiErr.Error()})
		}
		handler.deps.Logger.Error().Err(err).Msg(ErrInternalServerError.Error())
		return c.Status(ErrInternalServerError.HTTPStatus).JSON(fiber.Map{"error": ErrInternalServerError.Error()})
	}

	if !validationResult.Valid {
		return c.Status(ErrCouponNotFound.HTTPStatus).JSON(validationResult)
	}

	return c.JSON(validationResult)
}

// ExportCoupons экспортирует список купонов в текстовый файл
// @Summary Экспорт купонов
// @Description Экспортирует купоны в текстовый файл с фильтрацией
// @Tags coupons
// @Produce text/plain
// @Param partner_id query string false "ID партнера для фильтрации"
// @Param status query string false "Статус купонов для экспорта"
// @Param format query string false "Формат экспорта: codes (только коды) или full (полная информация)"
// @Success 200 {string} string "Текстовый файл с купонами"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /coupons/export [get]
func (handler *CouponHandler) ExportCoupons(c *fiber.Ctx) error {
	// Получаем параметры запроса
	partnerIDStr := c.Query("partner_id")
	status := c.Query("status")
	format := c.Query("format", "codes")

	// Парсим ID партнера
	var partnerID *uuid.UUID
	if partnerIDStr != "" {
		if id, err := uuid.Parse(partnerIDStr); err == nil {
			partnerID = &id
		}
	}

	// Экспортируем купоны
	content, err := handler.deps.CouponService.ExportCoupons(partnerID, status, format)
	if err != nil {
		if apiErr, ok := GetAPIError(err); ok {
			return c.Status(apiErr.HTTPStatus).JSON(fiber.Map{"error": apiErr.Error()})
		}
		handler.deps.Logger.Error().Err(err).Msg(ErrInternalServerError.Error())
		return c.Status(ErrInternalServerError.HTTPStatus).JSON(fiber.Map{"error": ErrInternalServerError.Error()})
	}

	// Устанавливаем заголовки для скачивания файла
	filename := fmt.Sprintf("coupons_export_%s.txt", time.Now().Format("20060102_150405"))
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Set("Content-Type", "text/plain")

	return c.SendString(content)
}

// DownloadMaterials скачивает материалы погашенного купона
// @Summary Скачивание материалов купона
// @Description Скачивает архив с материалами погашенного купона (оригинал, превью, схема)
// @Tags coupons
// @Produce application/zip
// @Param id path string true "ID купона"
// @Success 200 {string} string "ZIP архив с материалами"
// @Failure 400 {object} map[string]interface{} "Неверный ID купона"
// @Failure 404 {object} map[string]interface{} "Купон не найден или не использован"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /coupons/{id}/download-materials [get]
func (handler *CouponHandler) DownloadMaterials(c *fiber.Ctx) error {
	// Получаем ID купона
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		handler.deps.Logger.Error().Err(err).Msg(ErrInvalidCouponID.Message)
		return c.Status(ErrInvalidCouponID.HTTPStatus).JSON(fiber.Map{"error": ErrInvalidCouponID.Error()})
	}

	// Скачиваем материалы
	archiveData, filename, err := handler.deps.CouponService.DownloadMaterials(id)
	if err != nil {
		if apiErr, ok := GetAPIError(err); ok {
			return c.Status(apiErr.HTTPStatus).JSON(fiber.Map{"error": apiErr.Error()})
		}
		handler.deps.Logger.Error().Err(err).Msg(ErrInternalServerError.Error())
		return c.Status(ErrInternalServerError.HTTPStatus).JSON(fiber.Map{"error": ErrInternalServerError.Error()})
	}

	// Устанавливаем заголовки для скачивания
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Set("Content-Type", "application/zip")

	return c.Send(archiveData)
}

// GetCouponsPaginated возвращает купоны с пагинацией
// @Summary Список купонов с пагинацией
// @Description Возвращает список купонов с пагинацией и фильтрацией
// @Tags coupons
// @Produce json
// @Param page query int false "Номер страницы (по умолчанию 1)"
// @Param limit query int false "Количество элементов на странице (по умолчанию 20, максимум 100)"
// @Param code query string false "Код купона для поиска"
// @Param status query string false "Статус купона (new, used)"
// @Param size query string false "Размер купона"
// @Param style query string false "Стиль купона"
// @Param partner_id query string false "ID партнера"
// @Success 200 {object} map[string]interface{} "Купоны с информацией о пагинации"
// @Failure 400 {object} map[string]interface{} "Неверные параметры запроса"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /coupons/paginated [get]
func (handler *CouponHandler) GetCouponsPaginated(c *fiber.Ctx) error {
	// Получаем параметры пагинации
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)

	// Валидация параметров
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Параметры фильтрации
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

	// Получаем купоны с пагинацией
	coupons, total, err := handler.deps.CouponService.SearchCouponsWithPagination(
		code, status, size, style, partnerID, page, limit,
	)
	if err != nil {
		if apiErr, ok := GetAPIError(err); ok {
			return c.Status(apiErr.HTTPStatus).JSON(fiber.Map{"error": apiErr.Error()})
		}
		handler.deps.Logger.Error().Err(err).Msg(ErrInternalServerError.Error())
		return c.Status(ErrInternalServerError.HTTPStatus).JSON(fiber.Map{"error": ErrInternalServerError.Error()})
	}

	// Вычисляем данные пагинации
	totalPages := (total + int64(limit) - 1) / int64(limit)
	hasNext := int64(page) < totalPages
	hasPrev := page > 1

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

// GetCouponsByPartner возвращает купоны конкретного партнера
// @Summary Купоны партнера
// @Description Возвращает все купоны конкретного партнера с возможностью фильтрации
// @Tags coupons
// @Produce json
// @Param partner_id path string true "ID партнера"
// @Param status query string false "Статус купонов"
// @Param size query string false "Размер купонов"
// @Param style query string false "Стиль купонов"
// @Success 200 {array} map[string]interface{} "Купоны партнера"
// @Failure 400 {object} map[string]interface{} "Неверный ID партнера"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /coupons/partner/{partner_id} [get]
func (handler *CouponHandler) GetCouponsByPartner(c *fiber.Ctx) error {
	partnerIDStr := c.Params("partner_id")
	partnerID, err := uuid.Parse(partnerIDStr)
	if err != nil {
		handler.deps.Logger.Error().Err(err).Msg(ErrInvalidPartnerID.Message)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": ErrInvalidPartnerID.Message,
		})
	}

	// Дополнительные фильтры
	status := c.Query("status")
	size := c.Query("size")
	style := c.Query("style")

	// Получаем купоны
	coupons, err := handler.deps.CouponService.SearchCouponsByPartner(partnerID, status, size, style)
	if err != nil {
		if apiErr, ok := GetAPIError(err); ok {
			return c.Status(apiErr.HTTPStatus).JSON(fiber.Map{"error": apiErr.Error()})
		}
		handler.deps.Logger.Error().Err(err).Msg(ErrInternalServerError.Error())
		return c.Status(ErrInternalServerError.HTTPStatus).JSON(fiber.Map{"error": ErrInternalServerError.Error()})
	}

	return c.JSON(fiber.Map{
		"partner_id": partnerID,
		"coupons":    coupons,
		"count":      len(coupons),
	})
}
