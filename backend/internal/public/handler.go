package public

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/skr1ms/mosaic/internal/types"
	"github.com/skr1ms/mosaic/pkg/utils"
)

type PublicHandlerDeps struct {
	PublicService *PublicService
}

type PublicHandler struct {
	fiber.Router
	deps *PublicHandlerDeps
}

func NewPublicHandler(router fiber.Router, deps *PublicHandlerDeps) {
	handler := &PublicHandler{
		Router: router,
		deps:   deps,
	}

	// Публичные эндпоинты (без авторизации)
	api := handler.Group("/api")

	// Получение информации о партнере по домену (White Label)
	api.Get("/partners/:domain/info", handler.GetPartnerByDomain)

	// Работа с купонами
	api.Get("/coupons/:code", handler.GetCouponByCode)          // проверка валидности купона
	api.Post("/coupons/:code/activate", handler.ActivateCoupon) // активация купона
	api.Post("/coupons/purchase", handler.PurchaseCoupon)       // покупка купона онлайн

	// Работа с изображениями
	api.Post("/images/upload", handler.UploadImage)                 // загрузка изображения
	api.Post("/images/:id/edit", handler.EditImage)                 // редактирование изображения
	api.Post("/images/:id/process", handler.ProcessImage)           // применение стилей обработки
	api.Post("/images/:id/generate-schema", handler.GenerateSchema) // создание схемы мозаики
	api.Get("/images/:id/download", handler.DownloadSchema)         // скачивание схемы
	api.Post("/images/:id/send-email", handler.SendSchemaToEmail)   // отправка схемы на email
	api.Get("/images/:id/preview", handler.GetImagePreview)         // получение превью
	api.Get("/images/:id/status", handler.GetProcessingStatus)      // получение статуса обработки

	// Дополнительные эндпоинты
	api.Get("/sizes", handler.GetAvailableSizes)   // получение доступных размеров
	api.Get("/styles", handler.GetAvailableStyles) // получение доступных стилей
}

// GetPartnerByDomain возвращает информацию о партнере по домену
// @Summary		Информация о партнере по домену
// @Description	Возвращает брендинг и контактную информацию партнера для White Label
// @Tags		public
// @Produce		json
// @Param		domain		path		string					true	"Доменное имя партнера"
// @Success		200		{object}	map[string]interface{}		"Информация о партнере"
// @Failure		404		{object}	map[string]interface{}		"Партнер не найден"
// @Router		/api/partners/{domain}/info [get]
func (handler *PublicHandler) GetPartnerByDomain(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	domain := c.Params("domain")

	result, err := handler.deps.PublicService.deps.PartnerRepository.GetByDomain(context.Background(), domain)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get partner by domain")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get partner by domain",
		})
	}

	return c.JSON(result)
}

// GetCouponByCode возвращает информацию о купоне по коду
// @Summary		Информация о купоне
// @Description	Возвращает информацию о купоне для проверки его валидности
// @Tags		coupons
// @Produce		json
// @Param		code		path		string					true	"Код купона (12 цифр)"
// @Success		200		{object}	map[string]interface{}		"Информация о купоне"
// @Failure		400		{object}	map[string]interface{}		"Неверный формат кода"
// @Failure		404		{object}	map[string]interface{}		"Купон не найден"
// @Router		/api/coupons/{code} [get]
func (handler *PublicHandler) GetCouponByCode(c *fiber.Ctx) error {
	code := c.Params("code")

	result, err := handler.deps.PublicService.GetCouponByCode(code)
	if err != nil {
		return utils.LogAndReturnError(c, err, "Failed to get coupon by code", fiber.StatusInternalServerError, map[string]interface{}{
			"coupon_code": code,
			"handler":     "GetCouponByCode",
		})
	}

	utils.LogSuccess(c, "Successfully retrieved coupon", map[string]interface{}{
		"coupon_code": code,
		"handler":     "GetCouponByCode",
	})

	return c.JSON(result)
}

// ActivateCoupon активирует купон для последующей обработки
//
//	@Summary		Активация купона
//	@Description	Активирует купон и подготавливает его для загрузки изображения
//	@Tags			coupons
//	@Accept			json
//	@Produce		json
//	@Param			code	path		string					true	"Код купона"
//	@Param			request	body		public.ActivateCouponRequest	true	"Данные для активации"
//	@Success		200		{object}	map[string]interface{}	"Купон активирован"
//	@Failure		400		{object}	map[string]interface{}	"Ошибка в запросе"
//	@Failure		404		{object}	map[string]interface{}	"Купон не найден"
//	@Failure		409		{object}	map[string]interface{}	"Купон уже использован"
//	@Router			/api/coupons/{code}/activate [post]
func (handler *PublicHandler) ActivateCoupon(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	code := c.Params("code")

	var req ActivateCouponRequest
	if err := c.BodyParser(&req); err != nil {
		log.Error().Err(err).Msg("Failed to parse request body")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Bad request",
		})
	}

	result, err := handler.deps.PublicService.ActivateCoupon(code, req)
	if err != nil {
		log.Error().Err(err).Msg("Failed to activate coupon")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to activate coupon",
		})
	}

	return c.JSON(result)
}

// UploadImage загружает изображение для обработки
//
//	@Summary		Загрузка изображения
//	@Description	Загружает изображение пользователя для создания схемы мозаики
//	@Tags			images
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			coupon_id	formData	string					true	"ID активированного купона"
//	@Param			image		formData	file					true	"Файл изображения (JPG, PNG)"
//	@Success		201			{object}	map[string]interface{}	"Изображение загружено"
//	@Failure		400			{object}	map[string]interface{}	"Ошибка в запросе"
//	@Failure		413			{object}	map[string]interface{}	"Файл слишком большой"
//	@Router			/api/images/upload [post]
func (handler *PublicHandler) UploadImage(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	couponID := c.FormValue("coupon_id")
	if couponID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Coupon ID is required",
		})
	}

	// Получаем файл
	file, err := c.FormFile("image")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Image file is required",
		})
	}

	result, err := handler.deps.PublicService.UploadImage(couponID, file)
	if err != nil {
		log.Error().Err(err).Msg("Failed to upload image")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to upload image",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(result)
}

// EditImage применяет редактирование к изображению
//
//	@Summary		Редактирование изображения
//	@Description	Применяет кадрирование, поворот и масштабирование к изображению
//	@Tags			images
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string					true	"ID изображения"
//	@Param			request	body		types.EditImageRequest		true	"Параметры редактирования"
//	@Success		200		{object}	map[string]interface{}	"Изображение отредактировано"
//	@Failure		400		{object}	map[string]interface{}	"Ошибка в запросе"
//	@Failure		404		{object}	map[string]interface{}	"Изображение не найдено"
//	@Router			/api/images/{id}/edit [post]
func (handler *PublicHandler) EditImage(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	imageID := c.Params("id")

	var req types.EditImageRequest
	if err := c.BodyParser(&req); err != nil {
		log.Error().Err(err).Msg("Failed to parse request body")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Bad request",
		})
	}

	result, err := handler.deps.PublicService.EditImage(imageID, req)
	if err != nil {
		log.Error().Err(err).Msg("Failed to edit image")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to edit image",
		})
	}

	return c.JSON(result)
}

// ProcessImage применяет стиль обработки к изображению
//
//	@Summary		Обработка изображения
//	@Description	Применяет выбранный стиль обработки к изображению
//	@Tags			images
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string					true	"ID изображения"
//	@Param			request	body		types.ProcessImageRequest		true	"Параметры обработки"
//	@Success		200		{object}	map[string]interface{}	"Обработка начата"
//	@Failure		400		{object}	map[string]interface{}	"Ошибка в запросе"
//	@Failure		404		{object}	map[string]interface{}	"Изображение не найдено"
//	@Router			/api/images/{id}/process [post]
func (handler *PublicHandler) ProcessImage(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	imageID := c.Params("id")

	var req types.ProcessImageRequest
	if err := c.BodyParser(&req); err != nil {
		log.Error().Err(err).Msg("Failed to parse request body")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Bad request",
		})
	}

	result, err := handler.deps.PublicService.ProcessImage(imageID, req)
	if err != nil {
		log.Error().Err(err).Msg("Failed to process image")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to process image",
		})
	}

	return c.JSON(result)
}

// GenerateSchema создает финальную схему мозаики
//
//	@Summary		Создание схемы мозаики
//	@Description	Создает финальную схему алмазной мозаики
//	@Tags			images
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string					true	"ID изображения"
//	@Param			request	body		types.GenerateSchemaRequest	true	"Параметры генерации"
//	@Success		200		{object}	map[string]interface{}	"Схема создана"
//	@Failure		400		{object}	map[string]interface{}	"Ошибка в запросе"
//	@Failure		404		{object}	map[string]interface{}	"Изображение не найдено"
//	@Router			/api/images/{id}/generate-schema [post]
func (handler *PublicHandler) GenerateSchema(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	imageID := c.Params("id")

	// Парсим ID изображения
	imageUUID, err := uuid.Parse(imageID)
	if err != nil {
		log.Error().Err(err).Msg("Invalid image ID format")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid image ID format",
		})
	}

	var req types.GenerateSchemaRequest
	if err := c.BodyParser(&req); err != nil {
		log.Error().Err(err).Msg("Failed to parse request body")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Bad request",
		})
	}

	// Получаем задачу для получения CouponID (для обновления купона после создания схемы)
	task, err := handler.deps.PublicService.deps.ImageRepository.GetByID(context.Background(), imageUUID)
	if err != nil {
		log.Error().Err(err).Msg("Image not found")
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Image not found",
		})
	}

	// Запускаем создание схемы асинхронно через ImageService
	go func() {
		if err := handler.deps.PublicService.deps.ImageService.GenerateSchema(context.Background(), imageUUID, req.Confirmed); err != nil {
			log.Error().Err(err).Str("image_id", imageUUID.String()).Msg("Failed to generate schema")
			return
		}

		// Обновляем купон как использованный после успешного создания схемы
		if coupon, err := handler.deps.PublicService.deps.CouponRepository.GetByID(context.Background(), task.CouponID); err == nil {
			coupon.Status = "used"
			// Получаем актуальный статус изображения для URL схемы
			if status, err := handler.deps.PublicService.deps.ImageService.GetImageStatus(context.Background(), imageUUID); err == nil && status.SchemaURL != nil {
				coupon.SchemaURL = status.SchemaURL
			}
			completedAt := time.Now()
			coupon.CompletedAt = &completedAt
			handler.deps.PublicService.deps.CouponRepository.Update(context.Background(), coupon)
		}
	}()

	return c.JSON(fiber.Map{
		"message":    "Schema generation started",
		"actions":    []string{"download"},
		"email_sent": true, // Email будет отправлен автоматически после создания схемы
	})
}

// DownloadSchema позволяет скачать готовую схему
//
//	@Summary		Скачивание схемы
//	@Description	Скачивает готовую схему мозаики
//	@Tags			images
//	@Produce		application/octet-stream
//	@Param			id	path		string					true	"ID изображения"
//	@Success		200	{file}		file					"Файл схемы"
//	@Failure		404	{object}	map[string]interface{}	"Схема не найдена"
//	@Router			/api/images/{id}/download [get]
func (handler *PublicHandler) DownloadSchema(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	imageID := c.Params("id")

	// Парсим ID изображения
	imageUUID, err := uuid.Parse(imageID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid image ID",
		})
	}

	// Получаем статус изображения для получения URL схемы
	status, err := handler.deps.PublicService.deps.ImageService.GetImageStatus(c.UserContext(), imageUUID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get image status")
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Image not found",
		})
	}

	if status.Status != "completed" || status.SchemaURL == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Schema not ready",
		})
	}

	// Перенаправляем на URL схемы (presigned URL от S3)
	return c.Redirect(*status.SchemaURL)
}

// SendSchemaToEmail отправляет схему на email
//
//	@Summary		Отправка схемы на email
//	@Description	Отправляет готовую схему мозаики на указанный email
//	@Tags			images
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string					true	"ID изображения"
//	@Param			request	body		public.SendEmailRequest		true	"Email для отправки"
//	@Success		200		{object}	map[string]interface{}	"Схема отправлена"
//	@Failure		400		{object}	map[string]interface{}	"Ошибка в запросе"
//	@Failure		404		{object}	map[string]interface{}	"Схема не найдена"
//	@Router			/api/images/{id}/send-email [post]
func (handler *PublicHandler) SendSchemaToEmail(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	imageID := c.Params("id")

	var req SendEmailRequest
	if err := c.BodyParser(&req); err != nil {
		log.Error().Err(err).Msg("Failed to parse request body")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Bad request",
		})
	}

	result, err := handler.deps.PublicService.SendSchemaToEmail(imageID, req)
	if err != nil {
		log.Error().Err(err).Msg("Failed to send schema to email")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to send schema to email",
		})
	}

	return c.JSON(result)
}

// GetImagePreview возвращает превью изображения
//
//	@Summary		Превью изображения
//	@Description	Возвращает превью обработанного изображения
//	@Tags			images
//	@Produce		json
//	@Param			id	path		string					true	"ID изображения"
//	@Success		200	{object}	map[string]interface{}	"Превью изображения"
//	@Failure		404	{object}	map[string]interface{}	"Изображение не найдено"
//	@Router			/api/images/{id}/preview [get]
func (handler *PublicHandler) GetImagePreview(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	imageID := c.Params("id")

	result, err := handler.deps.PublicService.GetImagePreview(imageID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get image preview")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get image preview",
		})
	}

	return c.JSON(result)
}

// GetProcessingStatus возвращает статус обработки
//
//	@Summary		Статус обработки
//	@Description	Возвращает текущий статус обработки изображения
//	@Tags			images
//	@Produce		json
//	@Param			id	path		string					true	"ID изображения"
//	@Success		200	{object}	map[string]interface{}	"Статус обработки"
//	@Failure		404	{object}	map[string]interface{}	"Изображение не найдено"
//	@Router			/api/images/{id}/status [get]
func (handler *PublicHandler) GetProcessingStatus(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	imageID := c.Params("id")

	result, err := handler.deps.PublicService.GetProcessingStatus(imageID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get processing status")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get processing status",
		})
	}

	return c.JSON(result)
}

// PurchaseCoupon покупает новый купон онлайн
//
//	@Summary		Покупка купона
//	@Description	Покупает новый купон с оплатой картой
//	@Tags			coupons
//	@Accept			json
//	@Produce		json
//	@Param			request	body		public.PurchaseCouponRequest	true	"Параметры покупки"
//	@Success		201		{object}	map[string]interface{}	"Купон куплен"
//	@Failure		400		{object}	map[string]interface{}	"Ошибка в запросе"
//	@Router			/api/coupons/purchase [post]
func (handler *PublicHandler) PurchaseCoupon(c *fiber.Ctx) error {
	log := zerolog.Ctx(c.UserContext())
	var req PurchaseCouponRequest
	if err := c.BodyParser(&req); err != nil {
		log.Error().Err(err).Msg("Failed to parse request body")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Bad request",
		})
	}

	result, err := handler.deps.PublicService.PurchaseCoupon(req)
	if err != nil {
		log.Error().Err(err).Msg("Failed to purchase coupon")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to purchase coupon",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(result)
}

// GetAvailableSizes возвращает доступные размеры
//
//	@Summary		Доступные размеры
//	@Description	Возвращает список доступных размеров мозаики
//	@Tags			public
//	@Produce		json
//	@Success		200	{array}	map[string]interface{}	"Доступные размеры"
//	@Router			/api/sizes [get]
func (handler *PublicHandler) GetAvailableSizes(c *fiber.Ctx) error {
	// TODO: Реализовать получение доступных размеров
	sizes := handler.deps.PublicService.GetAvailableSizes()
	return c.JSON(sizes)
}

// GetAvailableStyles возвращает доступные стили
//
//	@Summary		Доступные стили
//	@Description	Возвращает список доступных стилей обработки
//	@Tags			public
//	@Produce		json
//	@Success		200	{array}	map[string]interface{}	"Доступные стили"
//	@Router			/api/styles [get]
func (handler *PublicHandler) GetAvailableStyles(c *fiber.Ctx) error {
	// TODO: Реализовать получение доступных стилей
	styles := handler.deps.PublicService.GetAvailableStyles()
	return c.JSON(styles)
}
