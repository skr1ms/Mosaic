package payment

// Request для покупки купона онлайн
type PurchaseCouponRequest struct {
	Size      string  `json:"size" validate:"required,oneof=21x30 30x40 40x40 40x50 40x60 50x70"`
	Style     string  `json:"style" validate:"required,oneof=grayscale skin_tone pop_art max_colors"`
	Email     string  `json:"email" validate:"required,email"`
	ReturnURL string  `json:"return_url" validate:"required,url"`
	FailURL   *string `json:"fail_url,omitempty" validate:"omitempty,url"`
	Language  string  `json:"language,omitempty" validate:"omitempty,oneof=ru en es"`
	Domain    *string `json:"domain,omitempty"` // Домен партнера для White Label
}

// Response для покупки купона
type PurchaseCouponResponse struct {
	OrderID     string `json:"order_id"`
	OrderNumber string `json:"order_number"`
	PaymentURL  string `json:"payment_url"`
	Success     bool   `json:"success"`
	Message     string `json:"message,omitempty"`
	Amount      string `json:"amont"`
}

// Request для получения статуса заказа
type OrderStatusRequest struct {
	OrderNumber string `json:"order_number" validate:"required"`
}

// Response для статуса заказа
type OrderStatusResponse struct {
	OrderID    string  `json:"order_id"`
	Status     string  `json:"status"`
	Size       string  `json:"size,omitempty"`
	Style      string  `json:"style,omitempty"`
	Amount     float64 `json:"amount"`
	Currency   string  `json:"currency"`
	CouponCode *string `json:"coupon_code,omitempty"` // Код купона после успешной оплаты
	Success    bool    `json:"success"`
	Message    string  `json:"message,omitempty"`
}

// Структуры для получения доступных размеров и стилей
type AvailableOptionsResponse struct {
	Sizes  []SizeOption  `json:"sizes"`
	Styles []StyleOption `json:"styles"`
}

type SizeOption struct {
	Value       string  `json:"value"`
	Label       string  `json:"label"`
	Description string  `json:"description"`
	Price       float64 `json:"price"` // Цена в рублях
}

type StyleOption struct {
	Value       string `json:"value"`
	Label       string `json:"label"`
	Description string `json:"description"`
}

// Структуры для работы с Альфа-Банком
type AlfaBankRegisterRequest struct {
	OrderNumber        string `json:"orderNumber"`
	Amount             int64  `json:"amount"`
	Currency           string `json:"currency,omitempty"`
	ReturnUrl          string `json:"returnUrl"`
	FailUrl            string `json:"failUrl,omitempty"`
	Description        string `json:"description,omitempty"`
	Language           string `json:"language,omitempty"`
	ClientId           string `json:"clientId,omitempty"`
	JsonParams         string `json:"jsonParams,omitempty"`
	SessionTimeoutSecs int    `json:"sessionTimeoutSecs,omitempty"`
	BindingId          string `json:"bindingId,omitempty"`
	Features           string `json:"features,omitempty"`
}

type AlfaBankRegisterResponse struct {
	OrderId      string `json:"orderId"`
	FormUrl      string `json:"formUrl"`
	ErrorCode    string `json:"errorCode,omitempty"`
	ErrorMessage string `json:"errorMessage,omitempty"`
}

type AlfaBankStatusResponse struct {
	ErrorCode             string `json:"errorCode"`
	ErrorMessage          string `json:"errorMessage,omitempty"`
	OrderNumber           string `json:"orderNumber"`
	OrderStatus           int    `json:"orderStatus"`
	ActionCode            int    `json:"actionCode"`
	ActionCodeDescription string `json:"actionCodeDescription"`
	Amount                int64  `json:"amount"`
	Currency              string `json:"currency"`
	Date                  int64  `json:"date"`
	Ip                    string `json:"ip"`
	OrderDescription      string `json:"orderDescription"`
}

// Структуры для webhook уведомлений от Альфа-Банка
type PaymentNotificationRequest struct {
	OrderNumber           string `json:"orderNumber" form:"orderNumber"`
	OrderStatus           *int   `json:"orderStatus,omitempty" form:"orderStatus"` // Опционально для тестовой среды
	AlfaBankOrderID       string `json:"orderId,omitempty" form:"orderId"`         // Опционально для тестовой среды
	Amount                int64  `json:"amount" form:"amount"`
	Currency              string `json:"currency" form:"currency"`
	ActionCode            *int   `json:"actionCode,omitempty" form:"actionCode"`                       // Опционально для тестовой среды
	ActionCodeDescription string `json:"actionCodeDescription,omitempty" form:"actionCodeDescription"` // Опционально для тестовой среды
	Date                  int64  `json:"date" form:"date"`
	IP                    string `json:"ip" form:"ip"`
	OrderDescription      string `json:"orderDescription" form:"orderDescription"`
	// Дополнительные поля для валидации
	Checksum string `json:"checksum" form:"checksum"`
}

// Структуры для тестирования интеграции
type TestIntegrationResponse struct {
	Success         bool   `json:"success"`
	Message         string `json:"message,omitempty"`
	OrderNumber     string `json:"order_number,omitempty"`
	AlfaBankOrderID string `json:"alfabank_order_id,omitempty"`
	PaymentURL      string `json:"payment_url,omitempty"`
	TestStatus      string `json:"test_status,omitempty"`
	ConfigStatus    string `json:"config_status,omitempty"`
	ErrorDetails    string `json:"error_details,omitempty"`
}

// Структуры для Swagger документации
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

type SuccessResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}
