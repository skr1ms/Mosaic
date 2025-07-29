package payment

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
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

	alfaReq := &AlfaBankRegisterRequest{
		OrderNumber:        orderNumber,
		Amount:             order.Amount,
		Currency:           "810",
		ReturnUrl:          req.ReturnURL,
		FailUrl:            getStringValue(req.FailURL),
		Description:        order.Description,
		Language:           language,
		ClientId:           req.Email,
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
			// Обновляем статус заказа в зависимости от ответа банка
			switch alfaResp.OrderStatus {
			case 2: // Оплачено
				err = s.deps.PaymentRepository.UpdateOrderStatus(ctx, orderNumber, OrderStatusPaid, nil)
				if err == nil {
					order.Status = OrderStatusPaid
					err = s.createCouponForOrder(ctx, order)
					if err != nil {
						return &OrderStatusResponse{
							Success: false,
							Message: fmt.Sprintf("Error creating coupon for order %s: %v", orderNumber, err),
						}, nil
					}
				}
			case 0, 3, 4, 6: // Отклонено/отменено
				err = s.deps.PaymentRepository.UpdateOrderStatus(ctx, orderNumber, OrderStatusFailed, nil)
				if err == nil {
					order.Status = OrderStatusFailed
				}
				// case 1, 5: статусы ожидания - оставляем pending
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
		return fmt.Errorf("order is not in pending status")
	}

	// Проверяем статус в Альфа-Банке
	if order.AlfaBankOrderID != nil {
		alfaResp, err := s.alfaClient.GetOrderStatus(ctx, *order.AlfaBankOrderID)
		if err != nil {
			return fmt.Errorf("error checking status in AlfaBank: %w", err)
		}

		switch alfaResp.OrderStatus {
		case 2: // Оплачено
			err = s.deps.PaymentRepository.UpdateOrderStatus(ctx, orderNumber, OrderStatusPaid, nil)
			if err != nil {
				return fmt.Errorf("error updating order status: %w", err)
			}

			// Создаем купон
			err = s.createCouponForOrder(ctx, order)
			if err != nil {
				return fmt.Errorf("error creating coupon: %w", err)
			}
		case 0: // Заказ отклонен
			err = s.deps.PaymentRepository.UpdateOrderStatus(ctx, orderNumber, OrderStatusFailed, nil)
			if err != nil {
				return fmt.Errorf("error updating order status: %w", err)
			}
		}
	}

	return nil
}

// Приватные методы

func (s *PaymentService) createCouponForOrder(ctx context.Context, order *Order) error {
	// Генерируем код купона согласно ТЗ
	var partnerCode string
	if order.PartnerID != nil {
		partner, err := s.deps.PartnerRepository.GetByID(ctx, *order.PartnerID)
		if err == nil {
			// Генерируем 4-значный код партнера из ID
			partnerCode = fmt.Sprintf("%04d", partner.ID.ID())
		}
	}
	if partnerCode == "" {
		partnerCode = "0000" // Собственные купоны
	}

	couponCode := fmt.Sprintf("%s-%08d", partnerCode, time.Now().UnixNano()%100000000)

	// Создаем купон
	newCoupon := &coupon.Coupon{
		Code:          couponCode,
		Size:          order.Size,
		Style:         order.Style,
		Status:        "new",
		IsPurchased:   true,
		PurchaseEmail: &order.UserEmail,
		PurchasedAt:   &time.Time{},
	}

	if order.PartnerID != nil {
		newCoupon.PartnerID = *order.PartnerID
	}

	*newCoupon.PurchasedAt = time.Now()

	err := s.deps.CouponRepository.Create(ctx, newCoupon)
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
