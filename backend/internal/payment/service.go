package payment

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/skr1ms/mosaic/config"
	"github.com/skr1ms/mosaic/internal/coupon"
	"github.com/skr1ms/mosaic/internal/partner"
)

type PaymentServiceDeps struct {
	PaymentRepository *PaymentRepository
	CouponRepository  *coupon.CouponRepository
	PartnerRepository *partner.PartnerRepository
	Config            *config.Config
}

type PaymentService struct {
	deps       *PaymentServiceDeps
	alfaClient *AlfaBankClient
}

func NewPaymentService(deps *PaymentServiceDeps) *PaymentService {
	return &PaymentService{
		deps:       deps,
		alfaClient: NewAlfaBankClient(deps.Config),
	}
}

type AlfaBankClient struct {
	config *config.Config
	client *http.Client
}

func NewAlfaBankClient(config *config.Config) *AlfaBankClient {
	return &AlfaBankClient{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *AlfaBankClient) RegisterOrder(ctx context.Context, req *AlfaBankRegisterRequest) (*AlfaBankRegisterResponse, error) {
	data := url.Values{}
	data.Set("userName", c.config.AlphaBankConfig.Username)
	data.Set("password", c.config.AlphaBankConfig.Password)
	data.Set("orderNumber", req.OrderNumber)
	data.Set("amount", strconv.FormatInt(req.Amount, 10))
	data.Set("returnUrl", req.ReturnUrl)

	if req.Currency != "" {
		data.Set("currency", req.Currency)
	} else {
		data.Set("currency", "810") // По умолчанию рубли
	}

	if req.FailUrl != "" {
		data.Set("failUrl", req.FailUrl)
	}
	if req.Description != "" {
		data.Set("description", req.Description)
	}
	if req.Language != "" {
		data.Set("language", req.Language)
	} else {
		data.Set("language", "ru")
	}
	if req.ClientId != "" {
		data.Set("clientId", req.ClientId)
	}
	if req.JsonParams != "" {
		data.Set("jsonParams", req.JsonParams)
	}
	if req.SessionTimeoutSecs > 0 {
		data.Set("sessionTimeoutSecs", strconv.Itoa(req.SessionTimeoutSecs))
	}
	if req.BindingId != "" {
		data.Set("bindingId", req.BindingId)
	}
	if req.Features != "" {
		data.Set("features", req.Features)
	}

	url := c.config.AlphaBankConfig.Url + "/payment/rest/register.do"

	resp, err := c.client.PostForm(url, data)
	if err != nil {
		return nil, fmt.Errorf("error requesting API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	var result AlfaBankRegisterResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	return &result, nil
}

func (c *AlfaBankClient) GetOrderStatus(ctx context.Context, orderID string) (*AlfaBankStatusResponse, error) {
	data := url.Values{}
	data.Set("userName", c.config.AlphaBankConfig.Username)
	data.Set("password", c.config.AlphaBankConfig.Password)
	data.Set("orderId", orderID)
	data.Set("language", "ru")

	resp, err := c.client.PostForm(c.config.AlphaBankConfig.Url+"/payment/rest/getOrderStatus.do", data)
	if err != nil {
		return nil, fmt.Errorf("error requesting status: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading status response: %w", err)
	}

	var result AlfaBankStatusResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, fmt.Errorf("error parsing status response: %w", err)
	}

	return &result, nil
}

// PurchaseCoupon - покупка купона онлайн с оплатой картой
func (s *PaymentService) validateWebhookSignature(notification *PaymentNotificationRequest) bool {
	if s.deps.Config.AlphaBankConfig.WebhookSecret == "" {
		// Если секрет не настроен, пропускаем валидацию (для разработки)
		return true
	}

	// Создаем строку для подписи из основных параметров
	data := fmt.Sprintf("%s;%d;%s;%d;%s",
		notification.OrderNumber,
		notification.OrderStatus,
		notification.AlfaBankOrderID,
		notification.Amount,
		notification.Currency,
	)

	// Вычисляем HMAC-SHA256
	h := hmac.New(sha256.New, []byte(s.deps.Config.AlphaBankConfig.WebhookSecret))
	h.Write([]byte(data))
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	// Сравниваем с полученной подписью
	return strings.EqualFold(expectedSignature, notification.Checksum)
}

func (s *PaymentService) getWebhookURL() string {
	// Приоритет: специальный webhook URL > FrontendURL + путь > fallback
	if s.deps.Config.AlphaBankConfig.WebhookURL != "" {
		return s.deps.Config.AlphaBankConfig.WebhookURL
	}

	baseURL := s.deps.Config.ServerConfig.FrontendURL
	if baseURL == "" {
		baseURL = "https://yourdomain.com" // Замените на ваш реальный домен
	}
	return baseURL + "/api/payment/notification"
}

func (s *PaymentService) PurchaseCoupon(ctx context.Context, req *PurchaseCouponRequest) (*PurchaseCouponResponse, error) {
	// Проверяем, что размер поддерживается
	supportedSizes := map[string]bool{
		Size21x30: true,
		Size30x40: true,
		Size40x40: true,
		Size40x50: true,
		Size40x60: true,
		Size50x70: true,
	}

	if !supportedSizes[req.Size] {
		return &PurchaseCouponResponse{
			Success: false,
			Message: "Unsupported size",
		}, nil
	}

	// Определяем партнера по домену
	var partnerID *uuid.UUID
	if req.Domain != nil && *req.Domain != "" {
		partner, err := s.deps.PartnerRepository.GetByDomain(ctx, *req.Domain)
		if err == nil && partner != nil {
			partnerID = &partner.ID
		}
	}

	// Генерируем уникальный номер заказа
	orderNumber := s.generateOrderNumber()

	// Проверяем, что номер заказа уникален
	for {
		existingOrder, err := s.deps.PaymentRepository.GetOrderByNumber(ctx, orderNumber)
		if err != nil || existingOrder == nil {
			break
		}
		orderNumber = s.generateOrderNumber()
	}

	// Создаем заказ
	order := &Order{
		OrderNumber: orderNumber,
		PartnerID:   partnerID,
		Size:        req.Size,
		Style:       req.Style,
		UserEmail:   req.Email,
		Amount:      int64(FixedPriceRub * 100),
		Currency:    "RUB",
		Status:      OrderStatusCreated,
		ReturnURL:   req.ReturnURL,
		FailURL:     req.FailURL,
		Description: fmt.Sprintf("Purchase of mosaic coupon %s style %s", req.Size, req.Style),
	}

	err := s.deps.PaymentRepository.CreateOrder(ctx, order)
	if err != nil {
		return &PurchaseCouponResponse{
			Success: false,
			Message: "Error creating order",
		}, nil
	}

	// Регистрируем заказ в Альфа-Банке
	language := "ru"
	if req.Language != "" {
		language = req.Language
	}

	// Подготавливаем webhook URL и параметры для jsonParams
	webhookURL := s.getWebhookURL()
	jsonParams := fmt.Sprintf(`{"callbackUrl":"%s"}`, webhookURL)

	alfaReq := &AlfaBankRegisterRequest{
		OrderNumber:        orderNumber,
		Amount:             order.Amount,
		Currency:           "810",
		ReturnUrl:          req.ReturnURL,
		FailUrl:            getStringValue(req.FailURL),
		Description:        order.Description,
		Language:           language,
		ClientId:           req.Email,
		JsonParams:         jsonParams,
		SessionTimeoutSecs: 1200,
	}

	// Логируем данные запроса для отладки
	alfaResp, err := s.alfaClient.RegisterOrder(ctx, alfaReq)
	if err != nil {
		return &PurchaseCouponResponse{
			Success: false,
			Message: fmt.Sprintf("Error creating payment: %v", err),
		}, nil
	}

	if alfaResp.ErrorCode != "" {
		return &PurchaseCouponResponse{
			Success: false,
			Message: fmt.Sprintf("Payment error: %s", alfaResp.ErrorMessage),
		}, nil
	}

	// Обновляем заказ с данными от Альфа-Банка
	err = s.deps.PaymentRepository.UpdateOrderStatus(ctx, orderNumber, OrderStatusPending, &alfaResp.OrderId)
	if err != nil {
		return &PurchaseCouponResponse{
			Success: false,
			Message: "Error updating order",
		}, nil
	}

	err = s.deps.PaymentRepository.UpdateOrderPaymentURL(ctx, orderNumber, alfaResp.FormUrl)
	if err != nil {
		return &PurchaseCouponResponse{
			Success: false,
			Message: "Error updating payment URL",
		}, nil
	}

	return &PurchaseCouponResponse{
		OrderID:    order.ID.String(),
		PaymentURL: alfaResp.FormUrl,
		Success:    true,
	}, nil
}

func (s *PaymentService) generateOrderNumber() string {
	// Используем UUID для гарантии уникальности + timestamp для читаемости
	uuid := uuid.New()
	timestamp := time.Now().Unix()
	return fmt.Sprintf("ORD_%d_%s", timestamp, uuid.String()[:8])
}

func getStringValue(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

// GetOrderStatus - получение статуса заказа
func (s *PaymentService) GetOrderStatus(ctx context.Context, orderNumber string) (*OrderStatusResponse, error) {
	order, err := s.deps.PaymentRepository.GetOrderByNumber(ctx, orderNumber)
	if err != nil {
		return &OrderStatusResponse{
			Success: false,
			Message: "Order not found",
		}, nil
	}

	// Если заказ еще не оплачен, проверяем статус в Альфа-Банке
	if order.Status == OrderStatusPending && order.AlfaBankOrderID != nil {
		alfaResp, err := s.alfaClient.GetOrderStatus(ctx, *order.AlfaBankOrderID)
		if err == nil && alfaResp.ErrorCode == "" {
			// Обрабатываем изменение статуса через webhook логику
			notification := &PaymentNotificationRequest{
				OrderNumber:     orderNumber,
				OrderStatus:     alfaResp.OrderStatus,
				AlfaBankOrderID: *order.AlfaBankOrderID,
			}

			// Обновляем статус заказа
			err = s.ProcessWebhookNotification(ctx, notification)
			if err == nil {
				// Обновляем объект заказа для корректного ответа
				order, _ = s.deps.PaymentRepository.GetOrderByNumber(ctx, orderNumber)
			}
		}
	}

	var couponCode *string
	if order.CouponID != nil {
		// Получаем код купона
		coupon, err := s.deps.CouponRepository.GetByID(ctx, *order.CouponID)
		if err == nil {
			couponCode = &coupon.Code
		}
	}

	return &OrderStatusResponse{
		OrderID:    order.ID.String(),
		Status:     order.Status,
		Size:       order.Size,
		Style:      order.Style,
		Amount:     float64(order.Amount) / 100, // переводим обратно в рубли
		Currency:   order.Currency,
		CouponCode: couponCode,
		Success:    true,
	}, nil
}

// GetAvailableOptions - получение доступных размеров и стилей
func (s *PaymentService) GetAvailableOptions() *AvailableOptionsResponse {
	sizes := []SizeOption{
		{Value: Size21x30, Label: "21×30 см", Description: "Маленький размер", Price: FixedPriceRub},
		{Value: Size30x40, Label: "30×40 см", Description: "Средний размер", Price: FixedPriceRub},
		{Value: Size40x40, Label: "40×40 см", Description: "Квадратный", Price: FixedPriceRub},
		{Value: Size40x50, Label: "40×50 см", Description: "Стандартный", Price: FixedPriceRub},
		{Value: Size40x60, Label: "40×60 см", Description: "Большой", Price: FixedPriceRub},
		{Value: Size50x70, Label: "50×70 см", Description: "Очень большой", Price: FixedPriceRub},
	}

	styles := []StyleOption{
		{Value: StyleGrayscale, Label: "Оттенки серого", Description: "Черно-белое изображение"},
		{Value: StyleSkinTone, Label: "Оттенки телесного", Description: "Портретный стиль"},
		{Value: StylePopArt, Label: "Поп-арт", Description: "Яркие цвета"},
		{Value: StyleMaxColors, Label: "Максимум цветов", Description: "Полная цветовая палитра"},
	}

	return &AvailableOptionsResponse{
		Sizes:  sizes,
		Styles: styles,
	}
}

// ProcessPaymentReturn - обработка возврата от платежной системы
func (s *PaymentService) ProcessPaymentReturn(ctx context.Context, orderNumber string) error {
	order, err := s.deps.PaymentRepository.GetOrderByNumber(ctx, orderNumber)
	if err != nil {
		return fmt.Errorf("order not found: %w", err)
	}

	if order.Status != OrderStatusPending {
		// Заказ уже обработан
		return nil
	}

	// Проверяем статус в Альфа-Банке
	if order.AlfaBankOrderID != nil {
		alfaResp, err := s.alfaClient.GetOrderStatus(ctx, *order.AlfaBankOrderID)
		if err != nil {
			return fmt.Errorf("error checking status in AlfaBank: %w", err)
		}

		// Используем ту же логику, что и в webhook
		notification := &PaymentNotificationRequest{
			OrderNumber:     orderNumber,
			OrderStatus:     alfaResp.OrderStatus,
			AlfaBankOrderID: *order.AlfaBankOrderID,
		}

		err = s.ProcessWebhookNotification(ctx, notification)
		if err != nil {
			return fmt.Errorf("error processing payment status: %w", err)
		}
	}

	return nil
}

func (s *PaymentService) createCouponForOrder(ctx context.Context, order *Order) error {
	var partnerCode string = "0000"

	if order.PartnerID != nil {
		partner, err := s.deps.PartnerRepository.GetByID(ctx, *order.PartnerID)
		if err == nil && partner != nil {
			partnerCode = partner.PartnerCode
		}
	}

	// Используем существующий пакет для генерации уникального кода
	couponCode, err := s.generateUniqueCouponCode(partnerCode)
	if err != nil {
		return fmt.Errorf("error generating coupon code: %w", err)
	}

	// Создаем купон
	now := time.Now()
	newCoupon := &coupon.Coupon{
		Code:          couponCode,
		Size:          order.Size,
		Style:         order.Style,
		Status:        "new",
		IsPurchased:   true,
		PurchaseEmail: &order.UserEmail,
		PurchasedAt:   &now,
	}

	if order.PartnerID != nil {
		newCoupon.PartnerID = *order.PartnerID
	}

	err = s.deps.CouponRepository.Create(ctx, newCoupon)
	if err != nil {
		return fmt.Errorf("error creating coupon: %w", err)
	}

	// Связываем купон с заказом
	err = s.deps.PaymentRepository.UpdateOrderCoupon(ctx, order.OrderNumber, newCoupon.ID)
	if err != nil {
		return fmt.Errorf("error linking coupon to order: %w", err)
	}

	return nil
}

// generateUniqueCouponCode генерирует уникальный код купона
func (s *PaymentService) generateUniqueCouponCode(partnerCode string) (string, error) {
	maxAttempts := 10

	for attempts := 0; attempts < maxAttempts; attempts++ {
		// Генерируем 8 случайных цифр для второй части кода
		randomSuffix := ""
		for i := 0; i < 8; i++ {
			digit, err := rand.Int(rand.Reader, big.NewInt(10))
			if err != nil {
				return "", fmt.Errorf("error generating random digit: %w", err)
			}
			randomSuffix += digit.String()
		}

		// Формируем код в формате XXXX-XXXX-XXXX
		code := fmt.Sprintf("%s-%s-%s",
			partnerCode,
			randomSuffix[:4],
			randomSuffix[4:8])

		// Проверяем уникальность
		exists, err := s.deps.CouponRepository.CodeExists(context.Background(), code)
		if err != nil {
			return "", fmt.Errorf("error checking code uniqueness: %w", err)
		}

		if !exists {
			return code, nil
		}
	}

	return "", fmt.Errorf("failed to generate unique coupon code after %d attempts", maxAttempts)
}

// ProcessWebhookNotification - обработка webhook уведомлений от Альфа-Банка
func (s *PaymentService) ProcessWebhookNotification(ctx context.Context, notification *PaymentNotificationRequest) error {
	// Валидируем подпись webhook'а
	if !s.validateWebhookSignature(notification) {
		return fmt.Errorf("invalid webhook signature")
	}

	// Получаем заказ по номеру
	order, err := s.deps.PaymentRepository.GetOrderByNumber(ctx, notification.OrderNumber)
	if err != nil {
		return fmt.Errorf("order not found: %w", err)
	}

	// Проверяем, что заказ в состоянии ожидания оплаты
	if order.Status != OrderStatusPending {
		// Заказ уже обработан, возвращаем успех
		return nil
	}

	// Обрабатываем статус платежа согласно документации Альфа-Банка
	switch notification.OrderStatus {
	case 0: // Заказ зарегистрирован, но не оплачен
		// Оставляем статус pending
		return nil

	case 1: // Предавторизованная сумма захолдирована (для двухстадийных платежей)
		return nil

	case 2: // Проведена полная авторизация суммы заказа
		// Платеж успешно завершен
		err = s.deps.PaymentRepository.UpdateOrderStatus(ctx, notification.OrderNumber, OrderStatusPaid, &notification.AlfaBankOrderID)
		if err != nil {
			return fmt.Errorf("error updating order status to paid: %w", err)
		}

		// Создаем купон автоматически
		err = s.createCouponForOrder(ctx, order)
		if err != nil {
			return fmt.Errorf("error creating coupon for order: %w", err)
		}

		return nil

	case 3: // Авторизация отменена
		err = s.deps.PaymentRepository.UpdateOrderStatus(ctx, notification.OrderNumber, OrderStatusCancelled, &notification.AlfaBankOrderID)
		if err != nil {
			return fmt.Errorf("error updating order status to cancelled: %w", err)
		}
		return nil

	case 4: // По транзакции была проведена операция возврата
		err = s.deps.PaymentRepository.UpdateOrderStatus(ctx, notification.OrderNumber, OrderStatusFailed, &notification.AlfaBankOrderID)
		if err != nil {
			return fmt.Errorf("error updating order status to failed: %w", err)
		}
		return nil

	case 5: // Инициирована авторизация через ACS банка-эмитента
		// Оставляем статус pending, ждем финального результата
		return nil

	case 6: // Авторизация отклонена
		err = s.deps.PaymentRepository.UpdateOrderStatus(ctx, notification.OrderNumber, OrderStatusFailed, &notification.AlfaBankOrderID)
		if err != nil {
			return fmt.Errorf("error updating order status to failed: %w", err)
		}
		return nil

	default:
		return fmt.Errorf("unknown order status: %d", notification.OrderStatus)
	}
}
