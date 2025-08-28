package middleware

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/skr1ms/mosaic/pkg/jwt"
)

var auditLogger = NewLogger()

type AuditEvent struct {
	UserID       string         `json:"user_id,omitempty"`
	UserType     string         `json:"user_type"` // admin, partner, public
	Action       string         `json:"action"`
	Resource     string         `json:"resource"`
	ResourceID   string         `json:"resource_id,omitempty"`
	Details      map[string]any `json:"details,omitempty"`
	IPAddress    string         `json:"ip_address"`
	UserAgent    string         `json:"user_agent"`
	Timestamp    time.Time      `json:"timestamp"`
	Success      bool           `json:"success"`
	ErrorMessage string         `json:"error_message,omitempty"`
}

// AuditLogger middleware for logging critical actions
func AuditLogger(logger *Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if !shouldAudit(c.Path(), c.Method()) {
			return c.Next()
		}

		start := time.Now()
		err := c.Next()
		duration := time.Since(start)

		event := createAuditEvent(c, err, duration)

		logAuditEvent(event)

		return err
	}
}

// shouldAudit determines if request should be logged
func shouldAudit(path, method string) bool {
	if method != "GET" && method != "HEAD" && method != "OPTIONS" {
		return true
	}

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

// createAuditEvent creates audit event based on request
func createAuditEvent(c *fiber.Ctx, err error, duration time.Duration) AuditEvent {
	event := AuditEvent{
		Action:    determineAction(c.Path(), c.Method()),
		Resource:  determineResource(c.Path()),
		IPAddress: c.IP(),
		UserAgent: c.Get("User-Agent"),
		Timestamp: time.Now(),
		Success:   err == nil && c.Response().StatusCode() < 400,

		Details: map[string]any{
			"method":   c.Method(),
			"path":     c.Path(),
			"status":   c.Response().StatusCode(),
			"duration": duration.Milliseconds(),
			"query":    string(c.Request().URI().QueryString()),
		},
	}

	if err != nil {
		event.ErrorMessage = err.Error()
	} else if c.Response().StatusCode() >= 400 {
		event.ErrorMessage = "HTTP error: " + strconv.Itoa(c.Response().StatusCode())
	}

	event.UserType, event.UserID = extractUserInfo(c)

	event.ResourceID = extractResourceID(c.Path())

	addCriticalDetails(&event, c)

	return event
}

// determineAction defines an action based on the path and method
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

// determineResource defines a resource based on the path
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

// extractUserInfo extracts user information from the context
func extractUserInfo(c *fiber.Ctx) (userType, userID string) {
	// Try to extract information from JWT claims stored in context
	if claims := c.Locals("jwt_claims"); claims != nil {
		// First try to handle jwt.Claims struct type (new format)
		if jwtClaims, ok := claims.(*jwt.Claims); ok && jwtClaims != nil {
			userType = jwtClaims.Role
			userID = jwtClaims.UserID.String()
			auditLogger.GetZerologLogger().Info().Str("user_type", userType).Str("user_id", userID).Msg("JWT claims extracted from context in audit middleware (struct format)")
			return userType, userID
		}
		// Fallback to map format for backward compatibility
		if claimsMap, ok := claims.(map[string]any); ok {
			if role, exists := claimsMap["role"]; exists {
				userType = fmt.Sprintf("%v", role)
			}
			if userIDVal, exists := claimsMap["user_id"]; exists {
				userID = fmt.Sprintf("%v", userIDVal)
			}
			auditLogger.GetZerologLogger().Info().Str("user_type", userType).Str("user_id", userID).Msg("JWT claims extracted from context in audit middleware (map format)")
			return userType, userID
		}
		// Unknown claims format
		auditLogger.GetZerologLogger().Warn().Interface("claims_type", fmt.Sprintf("%T", claims)).Msg("JWT claims found but unsupported type in audit middleware")
	} else {
		auditLogger.GetZerologLogger().Debug().Msg("No JWT claims found in context in audit middleware")
	}

	// If it was not possible to extract from the token, we determine along the way
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

// extractResourceID retrieves the resource ID from the path
func extractResourceID(path string) string {
	segments := strings.Split(path, "/")
	for _, segment := range segments {
		if len(segment) == 36 && strings.Count(segment, "-") == 4 {
			return segment
		}
		if len(segment) == 12 && isNumeric(segment) {
			return segment
		}
	}
	return ""
}

// addCriticalDetails adds additional details for critical operations
func addCriticalDetails(event *AuditEvent, c *fiber.Ctx) {
	if event.Resource == "COUPON" && strings.Contains(c.Path(), "/api/coupons/") {
		segments := strings.Split(c.Path(), "/")
		for _, segment := range segments {
			if len(segment) == 12 && isNumeric(segment) {
				event.Details["coupon_code"] = segment
				break
			}
		}
	}

	// For payment transactions, we add the amount (from the body)
	if event.Resource == "PAYMENT" && c.Method() == "POST" {
		if bodyBytes := c.Body(); len(bodyBytes) > 0 {
			var paymentData map[string]any
			if err := json.Unmarshal(bodyBytes, &paymentData); err == nil {
				if amount, exists := paymentData["amount"]; exists {
					event.Details["amount"] = amount
				}
			}
		}
	}

	// For authorization operations, we add branding information
	if event.Resource == "AUTH" {
		if branding := GetBrandingFromContext(c); branding != nil {
			if branding.Partner != nil {
				event.Details["partner_code"] = branding.Partner.PartnerCode
				event.Details["domain"] = branding.Partner.Domain
			}
		}
	}
}

// logAuditEvent logs audit event
func logAuditEvent(event AuditEvent) {

	// Determine log level based on success
	var logEvent *zerolog.Event
	if event.Success {
		logEvent = auditLogger.GetZerologLogger().Info()
	} else {
		logEvent = auditLogger.GetZerologLogger().Warn()
	}

	// Add common fields
	logEvent = logEvent.
		Str("action", event.Action).
		Str("resource", event.Resource).
		Str("ip_address", event.IPAddress).
		Str("user_agent", event.UserAgent).
		Time("timestamp", event.Timestamp).
		Bool("success", event.Success)

	// Add optional fields
	if event.UserID != "" {
		logEvent = logEvent.Str("user_id", event.UserID)
	}
	if event.UserType != "" {
		logEvent = logEvent.Str("user_type", event.UserType)
	}
	if event.ResourceID != "" {
		logEvent = logEvent.Str("resource_id", event.ResourceID)
	}
	if event.ErrorMessage != "" {
		logEvent = logEvent.Str("error_message", event.ErrorMessage)
	}

	// Add details as JSON
	if event.Details != nil {
		if detailsJSON, err := json.Marshal(event.Details); err == nil {
			logEvent = logEvent.RawJSON("details", detailsJSON)
		}
	}

	logEvent.Msg("Audit event")
}

// isNumeric checks if the string consists of only numbers
func isNumeric(s string) bool {
	for _, char := range s {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}
