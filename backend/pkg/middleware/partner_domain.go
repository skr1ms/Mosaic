package middleware

import (
	"context"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PartnerDomainMiddleware identifies partner by domain and adds partner ID to context
func PartnerDomainMiddleware(db *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get host from request
		host := c.Hostname()

		// Remove port if present
		if idx := strings.Index(host, ":"); idx != -1 {
			host = host[:idx]
		}

		// Skip for main domains
		if host == "photo.doyoupaint.com" || host == "adm.doyoupaint.com" ||
			strings.Contains(host, "localhost") || strings.Contains(host, "127.0.0.1") {
			return c.Next()
		}

		// Check if this is a partner domain
		var partnerID uuid.UUID
		var brandName string
		err := db.QueryRow(context.Background(),
			`SELECT id, brand_name FROM partners 
			 WHERE domain = $1 AND status = 'active' 
			 LIMIT 1`,
			host).Scan(&partnerID, &brandName)

		if err == nil {
			// Add partner info to context
			c.Locals("partner_id", partnerID)
			c.Locals("partner_domain", host)
			c.Locals("partner_brand", brandName)

			// Add header for backend processing
			c.Set("X-Partner-Domain", host)
			c.Set("X-Partner-ID", partnerID.String())
		}

		return c.Next()
	}
}

// GetPartnerDomainFromContext retrieves partner domain information from context
func GetPartnerDomainFromContext(c *fiber.Ctx) (partnerID uuid.UUID, domain string, brandName string, isPartner bool) {
	if id, ok := c.Locals("partner_id").(uuid.UUID); ok {
		partnerID = id
		isPartner = true
	}

	if d, ok := c.Locals("partner_domain").(string); ok {
		domain = d
	}

	if b, ok := c.Locals("partner_brand").(string); ok {
		brandName = b
	}

	return
}
