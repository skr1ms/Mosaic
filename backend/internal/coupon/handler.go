package coupon

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type CouponHandlerDeps struct {
	CouponRepository *CouponRepository
	Logger           *zerolog.Logger
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
	api.Get("/", handler.GetCoupons)                             // Получение списка купонов с фильтрацией ✔
	api.Get("/:id", handler.GetCouponDetails)                    // Получение деталей купона ✔
	api.Get("/paginated", handler.GetCouponsPaginated)           // Пагинация списка купонов ✔
	api.Get("/:id", handler.GetCouponByID)                       // Получение купона по ID ✔
	api.Get("/code/:code", handler.GetCouponByCode)              // Получение купона по коду ✔
	api.Post("/code/:code/validate", handler.ValidateCoupon)     // Валидация купона по коду ✔
	api.Put("/:id/activate", handler.ActivateCoupon)             // Активация купона ✔
	api.Put("/:id/reset", handler.ResetCoupon)                   // Сброс купона в исходное состояние ✔
	api.Put("/:id/send-schema", handler.SendSchema)              // Отправка схемы купона на email
	api.Put("/:id/purchase", handler.MarkAsPurchased)            // Пометка купона как купленного ✔
	api.Get("/export", handler.ExportCoupons)                    // Экспорт купонов в zip файл
	api.Get("/statistics", handler.GetStatistics)                // Получение статистики по купонам  ✔
	api.Get("/partner/:partner_id", handler.GetCouponsByPartner) // Получение купонов по ID партнера ✔
}

// handleError централизованно обрабатывает ошибки и возвращает соответствующий HTTP ответ
func (handler *CouponHandler) handleError(c *fiber.Ctx, err error) error {
	var apiErr APIError
	if errors.As(err, &apiErr) {
		handler.deps.Logger.Error().Err(err).Msg(apiErr.Error())
		return c.Status(apiErr.HTTPStatus).JSON(fiber.Map{
			"error": apiErr.Message,
			"code":  apiErr.Code,
		})
	}

	// Для неизвестных ошибок
	handler.deps.Logger.Error().Err(err).Msg("Internal server error")
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"error": "Internal server error",
		"code":  "INTERNAL_ERROR",
	})
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
	coupons, err := handler.deps.CouponRepository.Search(code, status, size, style, partnerID)
	if err != nil {
		handler.deps.Logger.Error().Err(err).Msg(ErrFailedToFindCoupons.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": ErrFailedToFindCoupons.Error()})
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
	handler.deps.Logger.Error().Msg("Coupon details")
	return c.JSON(fiber.Map{"message": "Coupon details"})
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
		handler.deps.Logger.Error().Err(err).Msg("Invalid coupon ID")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid coupon ID"})
	}

	// Получаем купон
	coupon, err := handler.deps.CouponRepository.GetByID(id)
	if err != nil {
		handler.deps.Logger.Error().Err(err).Msg("Coupon not found")
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Coupon not found"})
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
	coupon, err := handler.deps.CouponRepository.GetByCode(code)
	if err != nil {
		handler.deps.Logger.Error().Err(err).Msg("Coupon not found")
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Coupon not found"})
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": ErrInvalidCouponID.Error()})
	}

	var req ActivateCouponRequest

	// Парсим запрос
	if err := c.BodyParser(&req); err != nil {
		handler.deps.Logger.Error().Err(err).Msg(err.Error())
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	// Активируем купон
	if err := handler.deps.CouponRepository.ActivateCoupon(id, *req.OriginalImageURL, *req.PreviewURL, *req.SchemaURL); err != nil {
		handler.deps.Logger.Error().Err(err).Msg(err.Error())
		return handler.handleError(c, err)
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
		handler.deps.Logger.Error().Err(err).Msg("Invalid coupon ID")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid coupon ID"})
	}

	// Сбрасываем купон
	if err := handler.deps.CouponRepository.ResetCoupon(id); err != nil {
		handler.deps.Logger.Error().Err(err).Msg("Failed to reset coupon")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to reset coupon"})
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
		handler.deps.Logger.Error().Err(err).Msg("Invalid coupon ID")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid coupon ID"})
	}

	var req SendSchemaRequest

	// Парсим запрос
	if err := c.BodyParser(&req); err != nil {
		handler.deps.Logger.Error().Err(err).Msg(err.Error())
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	// Отправляем схему
	if err := handler.deps.CouponRepository.SendSchema(id, req.Email); err != nil {
		handler.deps.Logger.Error().Err(err).Msg("Failed to send schema")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to send schema"})
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
		handler.deps.Logger.Error().Err(err).Msg("Invalid coupon ID")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid coupon ID"})
	}

	var req MarkAsPurchasedRequest

	// Парсим запрос
	if err := c.BodyParser(&req); err != nil {
		handler.deps.Logger.Error().Err(err).Msg(err.Error())
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	// Помечаем купон как купленный
	if err := handler.deps.CouponRepository.MarkAsPurchased(id, req.PurchaseEmail); err != nil {
		handler.deps.Logger.Error().Err(err).Msg("Failed to mark as purchased")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to mark as purchased"})
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
	stats, err := handler.deps.CouponRepository.GetStatistics(partnerID)
	if err != nil {
		handler.deps.Logger.Error().Err(err).Msg("Failed to get statistics")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get statistics"})
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

	// Получаем купон
	coupon, err := handler.deps.CouponRepository.GetByCode(code)
	if err != nil {
		handler.deps.Logger.Error().Err(err).Msg("Coupon not found")
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"valid":   false,
			"message": "Coupon not found",
		})
	}

	// Проверяем статус купона
	if coupon.Status == "used" {
		handler.deps.Logger.Error().Msg("Coupon already used")
		return c.JSON(fiber.Map{
			"valid":   false,
			"message": "Coupon already used",
			"used_at": coupon.UsedAt,
			"size":    coupon.Size,
			"style":   coupon.Style,
		})
	}

	// Возвращаем результат
	handler.deps.Logger.Error().Msg("Coupon is valid and ready to use")
	return c.JSON(fiber.Map{
		"valid":   true,
		"message": "Coupon is valid and ready to use",
		"size":    coupon.Size,
		"style":   coupon.Style,
	})
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

	// Получаем купоны
	coupons, err := handler.deps.CouponRepository.Search("", status, "", "", partnerID)
	if err != nil {
		handler.deps.Logger.Error().Err(err).Msg("Failed to fetch coupons for export")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch coupons for export",
		})
	}

	// Создаем строитль для формирования вывода
	var content strings.Builder

	// Если формат "full", то выводим полную информацию о купонах
	if format == "full" {
		// Полная информация о купонах
		content.WriteString("Code\tPartner ID\tSize\tStyle\tStatus\tCreated At\tUsed At\n")
		for _, coupon := range coupons {
			usedAt := ""
			if coupon.UsedAt != nil {
				usedAt = coupon.UsedAt.Format(time.RFC3339)
			}
			content.WriteString(fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				coupon.Code,
				coupon.PartnerID.String(),
				coupon.Size,
				coupon.Style,
				coupon.Status,
				coupon.CreatedAt.Format(time.RFC3339),
				usedAt,
			))
		}
	} else {
		// Только коды купонов
		for _, coupon := range coupons {
			content.WriteString(coupon.Code + "\n")
		}
	}

	// Устанавливаем заголовки для скачивания файла
	// Формируем имя файла
	filename := fmt.Sprintf("coupons_export_%s.txt", time.Now().Format("20060102_150405"))
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Set("Content-Type", "text/plain")

	return c.SendString(content.String())
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
		handler.deps.Logger.Error().Err(err).Msg("Invalid coupon ID")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid coupon ID"})
	}

	// Получаем купон
	coupon, err := handler.deps.CouponRepository.GetByID(id)
	if err != nil {
		handler.deps.Logger.Error().Err(err).Msg("Coupon not found")
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Coupon not found"})
	}

	// Проверяем статус купона
	if coupon.Status != "used" {
		handler.deps.Logger.Error().Msg("Coupon must be used to download materials")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Coupon must be used to download materials",
		})
	}

	// Создаем ZIP архив в памяти
	// Создаем буфер для хранения архива
	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	// Функция для добавления файла по URL в архив
	addFileToZip := func(fileURL, filename string) error {
		if fileURL == "" {
			return nil // Пропускаем пустые URL
		}

		resp, err := http.Get(fileURL)
		if err != nil {
			handler.deps.Logger.Error().Err(err).Msg("Failed to download file from " + fileURL)
			return ErrFailedToDownloadFile
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			handler.deps.Logger.Error().Msg("Failed to download file from " + fileURL + ": status " + strconv.Itoa(resp.StatusCode))
			return ErrFailedToDownloadFile
		}

		fileWriter, err := zipWriter.Create(filename)
		if err != nil {
			handler.deps.Logger.Error().Err(err).Msg("Failed to create file writer")
			return ErrFailedToCreateFileWriter
		}

		_, err = io.Copy(fileWriter, resp.Body)
		if err != nil {
			handler.deps.Logger.Error().Err(err).Msg("Failed to copy file to zip")
			return ErrFailedToCopyFileToZip
		}

		return nil
	}

	// Добавляем файлы в архив
	if coupon.OriginalImageURL != nil && *coupon.OriginalImageURL != "" {
		if err := addFileToZip(*coupon.OriginalImageURL, "original_image.jpg"); err != nil {
			// Логируем ошибку, но продолжаем
			fmt.Printf("Error adding original image to zip: %v\n", err)
		}
	}

	if coupon.PreviewURL != nil && *coupon.PreviewURL != "" {
		if err := addFileToZip(*coupon.PreviewURL, "preview.jpg"); err != nil {
			fmt.Printf("Error adding preview to zip: %v\n", err)
		}
	}

	if coupon.SchemaURL != nil && *coupon.SchemaURL != "" {
		if err := addFileToZip(*coupon.SchemaURL, "schema.pdf"); err != nil {
			fmt.Printf("Error adding schema to zip: %v\n", err)
		}
	}

	// Добавляем информационный файл
	infoWriter, err := zipWriter.Create("coupon_info.txt")
	if err == nil {
		infoContent := fmt.Sprintf(`Coupon Information
Code: %s
Size: %s
Style: %s
Created: %s
Used: %s
`,
			coupon.Code,
			coupon.Size,
			coupon.Style,
			coupon.CreatedAt.Format(time.RFC3339),
			coupon.UsedAt.Format(time.RFC3339),
		)
		infoWriter.Write([]byte(infoContent))
	}

	// Закрываем архив
	err = zipWriter.Close()
	if err != nil {
		handler.deps.Logger.Error().Err(err).Msg("Failed to create archive")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create archive",
		})
	}

	// Устанавливаем заголовки для скачивания
	filename := fmt.Sprintf("coupon_%s_materials.zip", coupon.Code)
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Set("Content-Type", "application/zip")

	return c.Send(buf.Bytes())
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
	coupons, total, err := handler.deps.CouponRepository.SearchWithPagination(
		code, status, size, style, partnerID, page, limit,
	)
	if err != nil {
		handler.deps.Logger.Error().Err(err).Msg("Failed to fetch coupons")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch coupons",
		})
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
		handler.deps.Logger.Error().Err(err).Msg("Invalid partner ID")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid partner ID",
		})
	}

	// Дополнительные фильтры
	status := c.Query("status")
	size := c.Query("size")
	style := c.Query("style")

	// Получаем купоны
	coupons, err := handler.deps.CouponRepository.Search("", status, size, style, &partnerID)
	if err != nil {
		handler.deps.Logger.Error().Err(err).Msg("Failed to fetch partner coupons")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch partner coupons",
		})
	}

	return c.JSON(fiber.Map{
		"partner_id": partnerID,
		"coupons":    coupons,
		"count":      len(coupons),
	})
}
