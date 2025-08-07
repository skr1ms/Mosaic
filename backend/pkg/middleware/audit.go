package middleware

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

type AuditEvent struct {
	UserID       string                 `json:"user_id,omitempty"`
	UserType     string                 `json:"user_type"` // admin, partner, public
	Action       string                 `json:"action"`
	Resource     string                 `json:"resource"`
	ResourceID   string                 `json:"resource_id,omitempty"`
	Details      map[string]interface{} `json:"details,omitempty"`
	IPAddress    string                 `json:"ip_address"`
	UserAgent    string                 `json:"user_agent"`
	Timestamp    time.Time              `json:"timestamp"`
	Success      bool                   `json:"success"`
	ErrorMessage string                 `json:"error_message,omitempty"`
}

// AuditLogger middleware для логирования критических действий
func AuditLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Проверяем, нужно ли логировать данный эндпоинт
		if !shouldAudit(c.Path(), c.Method()) {
			return c.Next()
		}

		start := time.Now()
		err := c.Next()
		duration := time.Since(start)

		// Создаем событие аудита
		event := createAuditEvent(c, err, duration)

		// Логируем событие
		logAuditEvent(event)

		return err
	}
}

// shouldAudit определяет, нужно ли логировать данный запрос
func shouldAudit(path, method string) bool {
	// Логируем все изменяющие операции (POST, PUT, DELETE, PATCH)
	if method != "GET" && method != "HEAD" && method != "OPTIONS" {
		return true
	}

	// Логируем критические GET запросы
	criticalPaths := []string{
		"/api/admin",
		"/api/partners",
		"/api/coupons",
		"/api/payments",
		"/api/auth",
	}

	for _, criticalPath := range criticalPaths {
		if strings.HasPrefix(path, criticalPath) {
			return true
		}
	}

	return false
}

// createAuditEvent создает событие аудита на основе запроса
func createAuditEvent(c *fiber.Ctx, err error, duration time.Duration) AuditEvent {
	event := AuditEvent{
		Action:    determineAction(c.Path(), c.Method()),
		Resource:  determineResource(c.Path()),
		IPAddress: c.IP(),
		UserAgent: c.Get("User-Agent"),
		Timestamp: time.Now(),
		Success:   err == nil && c.Response().StatusCode() < 400,
		Details: map[string]interface{}{
			"method":   c.Method(),
			"path":     c.Path(),
			"status":   c.Response().StatusCode(),
			"duration": duration.Milliseconds(),
			"query":    string(c.Request().URI().QueryString()),
		},
	}

	// Добавляем информацию об ошибке, если есть
	if err != nil {
		event.ErrorMessage = err.Error()
	} else if c.Response().StatusCode() >= 400 {
		event.ErrorMessage = "HTTP error: " + strconv.Itoa(c.Response().StatusCode())
	}

	// Извлекаем информацию о пользователе из контекста
	event.UserType, event.UserID = extractUserInfo(c)

	// Извлекаем ID ресурса из пути
	event.ResourceID = extractResourceID(c.Path())

	// Добавляем дополнительные детали для критических операций
	addCriticalDetails(&event, c)

	return event
}

// determineAction определяет действие на основе пути и метода
func determineAction(path, method string) string {
	switch method {
	case "POST":
		if strings.Contains(path, "login") {
			return "LOGIN"
		} else if strings.Contains(path, "register") {
			return "REGISTER"
		} else if strings.Contains(path, "activate") {
			return "ACTIVATE"
		} else if strings.Contains(path, "purchase") {
			return "PURCHASE"
		}
		return "CREATE"
	case "PUT", "PATCH":
		return "UPDATE"
	case "DELETE":
		return "DELETE"
	case "GET":
		if strings.Contains(path, "admin") {
			return "VIEW_ADMIN"
		}
		return "VIEW"
	default:
		return method
	}
}

// determineResource определяет ресурс на основе пути
func determineResource(path string) string {
	if strings.Contains(path, "/api/admin") {
		return "ADMIN"
	} else if strings.Contains(path, "/api/partners") {
		return "PARTNER"
	} else if strings.Contains(path, "/api/coupons") {
		return "COUPON"
	} else if strings.Contains(path, "/api/images") {
		return "IMAGE"
	} else if strings.Contains(path, "/api/payments") {
		return "PAYMENT"
	} else if strings.Contains(path, "/api/auth") {
		return "AUTH"
	}
	return "UNKNOWN"
}

// extractUserInfo извлекает информацию о пользователе из контекста
func extractUserInfo(c *fiber.Ctx) (userType, userID string) {
	// Пытаемся извлечь информацию из JWT токена
	if claims := c.Locals("claims"); claims != nil {
		if claimsMap, ok := claims.(map[string]interface{}); ok {
			if role, exists := claimsMap["role"]; exists {
				userType = role.(string)
			}
			if id, exists := claimsMap["sub"]; exists {
				userID = id.(string)
			}
		}
	}

	// Если не удалось извлечь из токена, определяем по пути
	if userType == "" {
		if strings.Contains(c.Path(), "/api/admin") {
			userType = "admin"
		} else if strings.Contains(c.Path(), "/api/partner") {
			userType = "partner"
		} else {
			userType = "public"
		}
	}

	return userType, userID
}

// extractResourceID извлекает ID ресурса из пути
func extractResourceID(path string) string {
	// Простая логика извлечения UUID из пути
	segments := strings.Split(path, "/")
	for _, segment := range segments {
		// Проверяем, похож ли сегмент на UUID (36 символов с тире)
		if len(segment) == 36 && strings.Count(segment, "-") == 4 {
			return segment
		}
		// Проверяем, похож ли сегмент на код купона (12 цифр)
		if len(segment) == 12 && isNumeric(segment) {
			return segment
		}
	}
	return ""
}

// addCriticalDetails добавляет дополнительные детали для критических операций
func addCriticalDetails(event *AuditEvent, c *fiber.Ctx) {
	// Для операций с купонами добавляем код купона
	if event.Resource == "COUPON" && strings.Contains(c.Path(), "/api/coupons/") {
		segments := strings.Split(c.Path(), "/")
		for _, segment := range segments {
			if len(segment) == 12 && isNumeric(segment) {
				event.Details["coupon_code"] = segment
				break
			}
		}
	}

	// Для платежных операций добавляем сумму (из body)
	if event.Resource == "PAYMENT" && c.Method() == "POST" {
		if bodyBytes := c.Body(); len(bodyBytes) > 0 {
			var paymentData map[string]interface{}
			if err := json.Unmarshal(bodyBytes, &paymentData); err == nil {
				if amount, exists := paymentData["amount"]; exists {
					event.Details["amount"] = amount
				}
			}
		}
	}

	// Для операций авторизации добавляем брендинг информацию
	if event.Resource == "AUTH" {
		if branding := GetBrandingFromContext(c); branding != nil {
			if branding.Partner != nil {
				event.Details["partner_code"] = branding.Partner.PartnerCode
				event.Details["domain"] = branding.Partner.Domain
			}
		}
	}
}

// logAuditEvent записывает событие аудита в лог
func logAuditEvent(event AuditEvent) {
	logEvent := log.Info()

	if !event.Success {
		logEvent = log.Warn()
	}

	logEvent.
		Str("audit_event", "true").
		Str("user_type", event.UserType).
		Str("user_id", event.UserID).
		Str("action", event.Action).
		Str("resource", event.Resource).
		Str("resource_id", event.ResourceID).
		Str("ip_address", event.IPAddress).
		Bool("success", event.Success).
		Str("error_message", event.ErrorMessage).
		Interface("details", event.Details).
		Time("timestamp", event.Timestamp).
		Msg("Audit event")
}

// isNumeric проверяет, состоит ли строка только из цифр
func isNumeric(s string) bool {
	for _, char := range s {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}
