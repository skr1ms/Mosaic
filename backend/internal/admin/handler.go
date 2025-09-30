package admin

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/skr1ms/mosaic/internal/coupon"
	"github.com/skr1ms/mosaic/internal/image"
	"github.com/skr1ms/mosaic/internal/partner"

	"github.com/skr1ms/mosaic/pkg/jwt"
	"github.com/skr1ms/mosaic/pkg/marketplace"
	"github.com/skr1ms/mosaic/pkg/middleware"
	"github.com/skr1ms/mosaic/pkg/randomCouponCode"
)

type AdminHandlerDeps struct {
	AdminService AdminServiceInterface
	JwtService   JWTServiceInterface
	Logger       *middleware.Logger
	Config       ConfigInterface
}

type AdminHandler struct {
	fiber.Router
	deps *AdminHandlerDeps
}

func NewAdminHandler(router fiber.Router, deps *AdminHandlerDeps) {
	handler := &AdminHandler{
		Router: router,
		deps:   deps,
	}

	// Base admin group
	adminRoutes := handler.Group("/admin")
	jwtConcrete, ok := handler.deps.JwtService.(*jwt.JWT)
	if !ok {
		panic("JwtService must be *jwt.JWT for middleware")
	}

	// ================================================================
	// MAIN ADMIN ONLY ROUTES: /api/admin/main-admin/*
	// Access: only main_admin role
	// ================================================================
	mainAdminRoutes := adminRoutes.Group("/main-admin")
	mainAdminRoutes.Use(middleware.JWTMiddleware(jwtConcrete, deps.Logger), middleware.MainAdminOnly())

	mainAdminRoutes.Post("/admins", handler.CreateAdmin)       // POST /api/admin/main-admin/admins
	mainAdminRoutes.Get("/admins", handler.GetAdmins)          // GET /api/admin/main-admin/admins
	mainAdminRoutes.Delete("/admins/:id", handler.DeleteAdmin) // DELETE /api/admin/main-admin/admins/:id

	// ================================================================
	// REGULAR ADMIN ROUTES: /api/admin/*
	// Access: admin and main_admin roles
	// ================================================================
	adminRoutes.Use(middleware.JWTMiddleware(jwtConcrete, deps.Logger), middleware.AdminOrMainAdmin())

	// Admin management (partial access)
	adminRoutes.Patch("/admins/:id/password", handler.UpdateAdminPassword) // PATCH /api/admin/admins/:id/password
	adminRoutes.Patch("/admins/:id/email", handler.UpdateAdminEmail)       // PATCH /api/admin/admins/:id/email

	// Dashboard
	adminRoutes.Get("/dashboard", handler.GetDashboard) // GET /api/admin/dashboard

	// Partners management
	adminRoutes.Get("/partners", handler.GetPartners)                         // GET /api/admin/partners
	adminRoutes.Post("/partners", handler.CreatePartner)                      // POST /api/admin/partners
	adminRoutes.Get("/partners/:id", handler.GetPartner)                      // GET /api/admin/partners/:id
	adminRoutes.Post("/partners/:id/logo", handler.UploadPartnerLogo)         // POST /api/admin/partners/:id/logo
	adminRoutes.Get("/partners/:id/detail", handler.GetPartnerDetail)         // GET /api/admin/partners/:id/detail
	adminRoutes.Put("/partners/:id", handler.UpdatePartner)                   // PUT /api/admin/partners/:id
	adminRoutes.Patch("/partners/:id/block", handler.BlockPartner)            // PATCH /api/admin/partners/:id/block
	adminRoutes.Patch("/partners/:id/unblock", handler.UnblockPartner)        // PATCH /api/admin/partners/:id/unblock
	adminRoutes.Delete("/partners/:id", handler.DeletePartner)                // DELETE /api/admin/partners/:id
	adminRoutes.Get("/partners/:id/statistics", handler.GetPartnerStatistics) // GET /api/admin/partners/:id/statistics

	// Nginx management (for CI/CD pipeline)
	adminRoutes.Post("/nginx/deploy", handler.DeployNginxConfig) // POST /api/admin/nginx/deploy

	// Coupons management
	adminRoutes.Get("/coupons", handler.GetCoupons)                                     // GET /api/admin/coupons
	adminRoutes.Get("/coupons/paginated", handler.GetCouponsPaginated)                  // GET /api/admin/coupons/paginated
	adminRoutes.Post("/coupons", handler.CreateCoupons)                                 // POST /api/admin/coupons
	adminRoutes.Post("/coupons/export-advanced", handler.ExportCouponsAdvanced)         // POST /api/admin/coupons/export-advanced
	adminRoutes.Get("/coupons/export/partner/:id", handler.ExportPartnerCoupons)        // GET /api/admin/coupons/export/partner/:id
	adminRoutes.Post("/coupons/batch-delete", handler.BatchDeleteCoupons)               // POST /api/admin/coupons/batch-delete
	adminRoutes.Post("/coupons/batch/reset", handler.BatchResetCoupons)                 // POST /api/admin/coupons/batch/reset
	adminRoutes.Get("/coupons/:id", handler.GetCoupon)                                  // GET /api/admin/coupons/:id
	adminRoutes.Get("/coupons/:id/download-materials", handler.DownloadCouponMaterials) // GET /api/admin/coupons/:id/download-materials
	adminRoutes.Patch("/coupons/:id/reset", handler.ResetCoupon)                        // PATCH /api/admin/coupons/:id/reset
	adminRoutes.Delete("/coupons/:id", handler.DeleteCoupon)                            // DELETE /api/admin/coupons/:id

	// Statistics
	adminRoutes.Get("/statistics", handler.GetStatistics)                  // GET /api/admin/statistics
	adminRoutes.Get("/statistics/partners", handler.GetPartnersStatistics) // GET /api/admin/statistics/partners
	adminRoutes.Get("/statistics/system", handler.GetSystemStatistics)     // GET /api/admin/statistics/system

	// Image management
	adminRoutes.Get("/images", handler.GetAllImages)              // GET /api/admin/images
	adminRoutes.Get("/images/:id", handler.GetImageDetails)       // GET /api/admin/images/:id
	adminRoutes.Delete("/images/:id", handler.DeleteImageTask)    // DELETE /api/admin/images/:id
	adminRoutes.Post("/images/:id/retry", handler.RetryImageTask) // POST /api/admin/images/:id/retry

	// Endpoints for working with partner articles
	adminRoutes.Get("/partners/:id/articles/grid", handler.GetPartnerArticleGrid)       // GET /api/admin/partners/:id/articles/grid
	adminRoutes.Put("/partners/:id/articles/sku", handler.UpdatePartnerArticleSKU)      // PUT /api/admin/partners/:id/articles/sku
	adminRoutes.Post("/partners/:id/articles/generate-url", handler.GenerateProductURL) // POST /api/admin/partners/:id/articles/generate-url
}

// @Summary Create a new admin
// @Description Creates a new admin (only for existing admins)
// @Tags admin-management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param admin body CreateAdminRequest true "New admin data"
// @Success 201 {object} map[string]any "Admin created"
// @Failure 400 {object} map[string]any "Invalid request payload"
// @Failure 401 {object} map[string]any "Unauthorized"
// @Failure 403 {object} map[string]any "Forbidden"
// @Failure 500 {object} map[string]any "Internal server error"
// @Router /admin/main-admin/admins [post]
func (handler *AdminHandler) CreateAdmin(c *fiber.Ctx) error {
	var payload CreateAdminRequest

	if err := c.BodyParser(&payload); err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to parse request body")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to parse request body",
		})
	}
	if err := middleware.ValidateStruct(&payload); err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Invalid request payload")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request payload",
		})
	}
	admin, err := handler.deps.AdminService.CreateAdmin(payload)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to create admin")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create admin",
		})
	}
	handler.deps.Logger.FromContext(c).Info().Interface("context", map[string]any{
		"id":    admin.ID,
		"login": admin.Login,
		"role":  "admin",
	}).Msg("Admin created")
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":    admin.ID,
		"login": admin.Login,
		"role":  "admin",
	})
}

// @Summary Get admin dashboard
// @Description Returns data for admin dashboard
// @Tags admin-dashboard
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]any "Dashboard data"
// @Failure 401 {object} map[string]any "Unauthorized"
// @Failure 403 {object} map[string]any "Forbidden"
// @Failure 500 {object} map[string]any "Internal server error"
// @Router /admin/dashboard [get]
func (handler *AdminHandler) GetDashboard(c *fiber.Ctx) error {
	dashboardData, err := handler.deps.AdminService.GetDashboardData()
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to get dashboard")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get dashboard",
		})
	}
	handler.deps.Logger.FromContext(c).Info().Msg("Dashboard data retrieved")
	return c.JSON(dashboardData)
}

// @Summary Get admins list
// @Description Returns list of all administrators in the system
// @Tags admin-management
// @Produce json
// @Security BearerAuth
// @Success 200 {array} map[string]any "List of administrators"
// @Failure 401 {object} map[string]any "Unauthorized"
// @Failure 403 {object} map[string]any "Forbidden"
// @Failure 500 {object} map[string]any "Internal server error"
// @Router /admin/main-admin/admins [get]
func (handler *AdminHandler) GetAdmins(c *fiber.Ctx) error {
	admins, err := handler.deps.AdminService.GetAdmins()
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to get admins")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get admins",
		})
	}
	handler.deps.Logger.FromContext(c).Info().Interface("context", map[string]any{"count": len(admins)}).Msg("Admins list retrieved")
	return c.JSON(admins)
}

// @Summary Delete admin
// @Description Deletes administrator by ID
// @Tags admin-management
// @Produce json
// @Security BearerAuth
// @Param id path string true "Admin ID"
// @Success 200 {object} map[string]any "Admin deleted"
// @Failure 400 {object} map[string]any "Invalid admin ID"
// @Failure 401 {object} map[string]any "Unauthorized"
// @Failure 403 {object} map[string]any "Forbidden"
// @Failure 404 {object} map[string]any "Admin not found"
// @Failure 500 {object} map[string]any "Internal server error"
// @Router /admin/main-admin/admins/{id} [delete]
func (handler *AdminHandler) DeleteAdmin(c *fiber.Ctx) error {
	idStr := c.Params("id")
	if idStr == "" {
		handler.deps.Logger.FromContext(c).Error().Msg("Admin ID is required")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Admin ID is required",
		})
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Invalid admin ID format")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid admin ID format",
		})
	}

	if err := handler.deps.AdminService.DeleteAdmin(id); err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to delete admin")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete admin",
		})
	}

	handler.deps.Logger.FromContext(c).Info().Interface("context", map[string]any{"admin_id": id}).Msg("Admin deleted")
	return c.JSON(fiber.Map{
		"message": "Admin deleted successfully",
	})
}

// @Summary Update admin password
// @Description Updates administrator password by ID
// @Tags admin-management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Admin ID"
// @Param request body UpdateAdminPasswordRequest true "New password"
// @Success 200 {object} map[string]any "Password updated"
// @Failure 400 {object} map[string]any "Invalid request or password too short"
// @Failure 401 {object} map[string]any "Unauthorized"
// @Failure 403 {object} map[string]any "Forbidden"
// @Failure 404 {object} map[string]any "Admin not found"
// @Failure 500 {object} map[string]any "Internal server error"
// @Router /admin/admins/{id}/password [patch]
func (handler *AdminHandler) UpdateAdminPassword(c *fiber.Ctx) error {
	idStr := c.Params("id")
	if idStr == "" {
		handler.deps.Logger.FromContext(c).Error().Msg("Admin ID is required")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Admin ID is required",
		})
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Invalid admin ID format")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid admin ID format",
		})
	}

	claims, err := jwt.GetClaimsFromFiberContext(c)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to get JWT claims")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Failed to get JWT claims",
		})
	}

	if claims.Role != "main_admin" && claims.UserID != id {
		handler.deps.Logger.FromContext(c).Error().Msg("You can only change your own password")
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You can only change your own password",
		})
	}

	var req UpdateAdminPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to parse request body")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to parse request body",
		})
	}

	if req.Password == "" || len(req.Password) < 6 {
		handler.deps.Logger.FromContext(c).Error().Msg("Password must be at least 6 characters long")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Password must be at least 6 characters long",
		})
	}

	if err := handler.deps.AdminService.UpdateAdminPassword(id, req.Password); err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to update admin password")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update admin password",
		})
	}

	handler.deps.Logger.FromContext(c).Info().Interface("context", map[string]any{"admin_id": id}).Msg("Admin password updated")
	return c.JSON(fiber.Map{
		"message": "Admin password updated successfully",
	})
}

// @Summary Update admin email
// @Description Updates administrator email by ID
// @Tags admin-management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Admin ID"
// @Param request body UpdateAdminEmailRequest true "New email"
// @Success 200 {object} map[string]any "Email updated"
// @Failure 400 {object} map[string]any "Invalid request or email format"
// @Failure 401 {object} map[string]any "Unauthorized"
// @Failure 403 {object} map[string]any "Forbidden"
// @Failure 404 {object} map[string]any "Admin not found"
// @Failure 500 {object} map[string]any "Internal server error"
// @Router /admin/admins/{id}/email [patch]
func (handler *AdminHandler) UpdateAdminEmail(c *fiber.Ctx) error {
	idStr := c.Params("id")
	if idStr == "" {
		handler.deps.Logger.FromContext(c).Error().Msg("Admin ID is required")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Admin ID is required",
		})
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Invalid admin ID format")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid admin ID format",
		})
	}

	claims, err := jwt.GetClaimsFromFiberContext(c)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to get JWT claims")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Failed to get JWT claims",
		})
	}

	if claims.Role != "main_admin" && claims.UserID != id {
		handler.deps.Logger.FromContext(c).Error().Msg("You can only change your own email")
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You can only change your own email",
		})
	}

	var req UpdateAdminEmailRequest
	if err := c.BodyParser(&req); err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to parse request body")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to parse request body",
		})
	}

	if req.Email != "" {
		emailRegex := regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
		if !emailRegex.MatchString(req.Email) {
			handler.deps.Logger.FromContext(c).Error().Msg("Invalid email format")
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid email format",
			})
		}
	}

	if err := handler.deps.AdminService.UpdateAdminEmail(id, req.Email); err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to update admin email")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update admin email",
		})
	}

	handler.deps.Logger.FromContext(c).Info().Interface("context", map[string]any{"admin_id": id}).Msg("Admin email updated")
	return c.JSON(fiber.Map{
		"message": "Admin email updated successfully",
	})
}

// @Summary Get partners list with filtering and sorting
// @Description Returns list of all partners with filtering, search and sorting
// @Tags admin-partners
// @Produce json
// @Security BearerAuth
// @Param search query string false "Search by brand name, domain or email"
// @Param status query string false "Filter by status (active/blocked)"
// @Param sort_by query string false "Sort field (created_at/brand_name)"
// @Param order query string false "Sort order (asc/desc, default desc)"
// @Success 200 {object} map[string]any "Partners list"
// @Failure 401 {object} map[string]any "Unauthorized"
// @Failure 403 {object} map[string]any "Forbidden"
// @Failure 500 {object} map[string]any "Internal server error"
// @Router /admin/partners [get]
func (handler *AdminHandler) GetPartners(c *fiber.Ctx) error {
	search := c.Query("search")
	status := c.Query("status")
	sortBy := c.Query("sort_by", "created_at")
	order := c.Query("order", "desc")
	partners, err := handler.deps.AdminService.GetPartners(search, status, sortBy, order)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to get partners")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get partners",
		})
	}
	result := make([]fiber.Map, len(partners))
	for i, partner := range partners {
		result[i] = fiber.Map{
			"id":               partner.ID,
			"login":            partner.Login,
			"last_login":       partner.LastLogin,
			"created_at":       partner.CreatedAt,
			"updated_at":       partner.UpdatedAt,
			"partner_code":     partner.PartnerCode,
			"domain":           partner.Domain,
			"brand_name":       partner.BrandName,
			"logo_url":         partner.LogoURL,
			"ozon_link":        partner.OzonLink,
			"wildberries_link": partner.WildberriesLink,
			"email":            partner.Email,
			"address":          partner.Address,
			"phone":            partner.Phone,
			"telegram":         partner.Telegram,
			"whatsapp":         partner.Whatsapp,
			"allow_sales":      partner.AllowSales,
			"brand_colors":     partner.BrandColors,
			"status":           partner.Status,
		}
	}
	handler.deps.Logger.FromContext(c).Info().Interface("context", map[string]any{"count": len(result)}).Msg("Partners list retrieved")
	return c.JSON(fiber.Map{
		"partners": result,
		"total":    len(result),
	})
}

// @Summary Create a new partner
// @Description Creates a new partner with an auto-generated code (starting from 0001, 0000 is reserved)
// @Tags admin-partners
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param partner body partner.CreatePartnerRequest true "New partner data"
// @Success 201 {object} map[string]any "Partner created"
// @Failure 400 {object} map[string]any "Invalid request payload"
// @Failure 401 {object} map[string]any "Unauthorized"
// @Failure 403 {object} map[string]any "Forbidden"
// @Failure 409 {object} map[string]any "Partner with this login/domain already exists"
// @Failure 500 {object} map[string]any "Internal server error"
// @Router /admin/partners [post]
func (handler *AdminHandler) CreatePartner(c *fiber.Ctx) error {
	var req partner.CreatePartnerRequest
	if err := c.BodyParser(&req); err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to parse request body")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to parse request body",
		})
	}
	existingPartner, err := handler.deps.AdminService.GetPartnerRepository().GetByLogin(context.Background(), req.Login)
	if err == nil && existingPartner != nil {
		handler.deps.Logger.FromContext(c).Error().Msg("Partner with this login already exists")
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "Partner with this login already exists",
		})
	}
	existingPartnerByDomain, err := handler.deps.AdminService.GetPartnerRepository().GetByDomain(context.Background(), req.Domain)
	if err == nil && existingPartnerByDomain != nil {
		handler.deps.Logger.FromContext(c).Error().Msg("Partner with this domain already exists")
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "Partner with this domain already exists",
		})
	}
	partner, err := handler.deps.AdminService.CreatePartner(req)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to create partner")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create partner",
		})
	}

	handler.deps.Logger.FromContext(c).Info().Interface("context", map[string]any{
		"id":           partner.ID,
		"login":        partner.Login,
		"partner_code": partner.PartnerCode,
	}).Msg("Partner created")
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":           partner.ID,
		"login":        partner.Login,
		"partner_code": partner.PartnerCode,
	})
}

// @Summary Get partner details
// @Description Returns partner details by ID
// @Tags admin-partners
// @Produce json
// @Security BearerAuth
// @Param id path string true "Partner ID"
// @Success 200 {object} map[string]any "Partner info"
// @Failure 400 {object} map[string]any "Invalid ID"
// @Failure 401 {object} map[string]any "Unauthorized"
// @Failure 403 {object} map[string]any "Forbidden"
// @Failure 404 {object} map[string]any "Partner not found"
// @Failure 500 {object} map[string]any "Internal server error"
// @Router /admin/partners/{id} [get]
func (handler *AdminHandler) GetPartner(c *fiber.Ctx) error {
	partnerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Invalid partner ID")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid partner ID",
		})
	}
	partner, err := handler.deps.AdminService.GetPartnerRepository().GetByID(context.Background(), partnerID)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Partner not found")
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Partner not found",
		})
	}
	totalCoupons, err := handler.deps.AdminService.GetCouponRepository().CountByPartnerID(context.Background(), partnerID)
	if err != nil {
		totalCoupons = 0
	}
	activatedCoupons, err := handler.deps.AdminService.GetCouponRepository().CountActivatedByPartnerID(context.Background(), partnerID)
	if err != nil {
		activatedCoupons = 0
	}
	purchasedCoupons, err := handler.deps.AdminService.GetCouponRepository().CountPurchasedByPartnerID(context.Background(), partnerID)
	if err != nil {
		purchasedCoupons = 0
	}
	handler.deps.Logger.FromContext(c).Info().Interface("context", map[string]any{"partner_id": partner.ID}).Msg("Partner details retrieved")
	return c.JSON(fiber.Map{
		"id":                partner.ID,
		"login":             partner.Login,
		"partner_code":      partner.PartnerCode,
		"domain":            partner.Domain,
		"brand_name":        partner.BrandName,
		"logo_url":          partner.LogoURL,
		"ozon_link":         partner.OzonLink,
		"wildberries_link":  partner.WildberriesLink,
		"email":             partner.Email,
		"address":           partner.Address,
		"phone":             partner.Phone,
		"telegram":          partner.Telegram,
		"whatsapp":          partner.Whatsapp,
		"allow_sales":       partner.AllowSales,
		"brand_colors":      partner.BrandColors,
		"status":            partner.Status,
		"last_login":        partner.LastLogin,
		"created_at":        partner.CreatedAt,
		"updated_at":        partner.UpdatedAt,
		"total_coupons":     totalCoupons,
		"activated_coupons": activatedCoupons,
		"purchased_coupons": purchasedCoupons,
	})
}

// @Summary Upload partner logo
// @Description Accepts multipart/form-data with logo file and saves it to S3 logos bucket
// @Tags admin-partners
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param id path string true "Partner ID"
// @Param logo formData file true "Logo file (image/*)"
// @Success 200 {object} map[string]any "Logo URL"
// @Failure 400 {object} map[string]any "Bad request"
// @Failure 401 {object} map[string]any "Unauthorized"
// @Failure 403 {object} map[string]any "Forbidden"
// @Failure 404 {object} map[string]any "Partner not found"
// @Failure 500 {object} map[string]any "Internal server error"
// @Router /admin/partners/{id}/logo [post]
func (handler *AdminHandler) UploadPartnerLogo(c *fiber.Ctx) error {
	partnerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Invalid partner ID")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid partner ID",
		})
	}
	p, err := handler.deps.AdminService.GetPartnerRepository().GetByID(context.Background(), partnerID)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Partner not found")
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Partner not found",
		})
	}
	fileHeader, err := c.FormFile("logo")
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Logo file is required")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Logo file is required",
		})
	}
	f, err := fileHeader.Open()
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to open logo file")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to open logo file",
		})
	}
	defer f.Close()
	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/png"
	}
	key, err := handler.deps.AdminService.GetS3Client().UploadLogo(c.UserContext(), f, fileHeader.Size, contentType, p.ID.String())
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to upload logo")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to upload logo",
		})
	}
	url, err := handler.deps.AdminService.GetS3Client().GetLogoURL(c.UserContext(), key, 24*time.Hour)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to get logo URL")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get logo URL",
		})
	}
	p.LogoURL = url
	if err := handler.deps.AdminService.GetPartnerRepository().Update(context.Background(), p); err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to update partner")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update partner",
		})
	}
	handler.deps.Logger.FromContext(c).Info().Interface("context", map[string]any{"partner_id": p.ID, "logo_key": key}).Msg("Partner logo uploaded")
	return c.JSON(fiber.Map{
		"logo_url": url,
		"logo_key": key,
	})
}

// @Summary Get partner detail (with stats and history)
// @Description Returns detailed partner info including coupon stats, last activity, and profile history
// @Tags admin-partners
// @Produce json
// @Security BearerAuth
// @Param id path string true "Partner ID"
// @Success 200 {object} map[string]any "Partner detail info"
// @Failure 400 {object} map[string]any "Invalid partner ID"
// @Failure 404 {object} map[string]any "Partner not found"
// @Failure 401 {object} map[string]any "Unauthorized"
// @Failure 403 {object} map[string]any "Forbidden"
// @Failure 500 {object} map[string]any "Internal server error"
// @Router /admin/partners/{id}/detail [get]
func (handler *AdminHandler) GetPartnerDetail(c *fiber.Ctx) error {
	partnerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Invalid partner ID")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid partner ID",
		})
	}
	partnerDetail, err := handler.deps.AdminService.GetPartnerDetail(partnerID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Partner not found")
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Partner not found",
			})
		}
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to get partner detail")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get partner detail",
		})
	}
	handler.deps.Logger.FromContext(c).Info().Interface("context", map[string]any{"partner_id": partnerID}).Msg("Partner detail retrieved")
	return c.JSON(partnerDetail)
}

// @Summary Update partner
// @Description Updates partner data
// @Tags admin-partners
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Partner ID"
// @Success 200 {object} map[string]any "Partner updated"
// @Failure 401 {object} map[string]any "Unauthorized"
// @Failure 403 {object} map[string]any "Forbidden"
// @Failure 404 {object} map[string]any "Partner not found"
// @Failure 500 {object} map[string]any "Internal server error"
// @Router /admin/partners/{id} [put]
func (handler *AdminHandler) UpdatePartner(c *fiber.Ctx) error {
	var req partner.UpdatePartnerRequest
	if err := c.BodyParser(&req); err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to parse request body")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to parse request body",
		})
	}
	partnerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Invalid partner ID")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid partner ID",
		})
	}
	claims, err := jwt.GetClaimsFromFiberContext(c)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to get admin claims")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Failed to get admin claims",
		})
	}

	reason := c.Query("reason", "Admin update")
	partner, err := handler.deps.AdminService.UpdatePartnerWithHistory(partnerID, req, claims.Login, reason)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Partner not found")
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Partner not found",
			})
		}
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to update partner")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update partner",
		})
	}

	// CI/CD pipeline is now triggered automatically in UpdatePartner method when domain changes
	handler.deps.Logger.FromContext(c).Info().Interface("context", map[string]any{"partner_id": partner.ID}).Msg("Partner updated")
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"id":           partner.ID,
		"login":        partner.Login,
		"partner_code": partner.PartnerCode,
	})
}

// @Summary Block partner
// @Description Blocks partner (temporarily disables access)
// @Tags admin-partners
// @Produce json
// @Security BearerAuth
// @Param id path string true "Partner ID"
// @Success 200 {object} map[string]any "Partner blocked"
// @Failure 400 {object} map[string]any "Invalid ID"
// @Failure 401 {object} map[string]any "Unauthorized"
// @Failure 403 {object} map[string]any "Forbidden"
// @Failure 404 {object} map[string]any "Partner not found"
// @Failure 500 {object} map[string]any "Internal server error"
// @Router /admin/partners/{id}/block [patch]
func (handler *AdminHandler) BlockPartner(c *fiber.Ctx) error {
	partnerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Invalid partner ID")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid partner ID",
		})
	}
	if err := handler.deps.AdminService.GetPartnerRepository().UpdateStatus(context.Background(), partnerID, "blocked"); err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to block partner")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to block partner",
		})
	}
	if err := handler.deps.AdminService.GetCouponRepository().UpdateStatusByPartnerID(context.Background(), partnerID, true); err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to block coupons")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to block coupons",
		})
	}
	handler.deps.Logger.FromContext(c).Info().Interface("context", map[string]any{"partner_id": partnerID}).Msg("Partner blocked")
	return c.JSON(fiber.Map{"message": "Partner blocked successfully"})
}

// @Summary Unblock partner
// @Description Unblocks partner (restores access)
// @Tags admin-partners
// @Produce json
// @Security BearerAuth
// @Param id path string true "Partner ID"
// @Success 200 {object} map[string]any "Partner unblocked"
// @Failure 400 {object} map[string]any "Invalid ID"
// @Failure 401 {object} map[string]any "Unauthorized"
// @Failure 403 {object} map[string]any "Forbidden"
// @Failure 404 {object} map[string]any "Partner not found"
// @Failure 500 {object} map[string]any "Internal server error"
// @Router /admin/partners/{id}/unblock [patch]
func (handler *AdminHandler) UnblockPartner(c *fiber.Ctx) error {
	partnerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Invalid partner ID")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid partner ID",
		})
	}
	if err := handler.deps.AdminService.GetPartnerRepository().UpdateStatus(context.Background(), partnerID, "active"); err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to unblock partner")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to unblock partner",
		})
	}
	if err := handler.deps.AdminService.GetCouponRepository().UpdateStatusByPartnerID(context.Background(), partnerID, false); err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to unblock coupons")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to unblock coupons",
		})
	}
	handler.deps.Logger.FromContext(c).Info().Interface("context", map[string]any{"partner_id": partnerID}).Msg("Partner unblocked")
	return c.JSON(fiber.Map{"message": "Partner unblocked successfully"})
}

// @Summary Delete partner
// @Description Deletes partner by ID with all related data
// @Tags admin-partners
// @Produce json
// @Security BearerAuth
// @Param id path string true "Partner ID"
// @Param confirm query boolean false "Delete confirmation (true/false)"
// @Success 200 {object} map[string]any "Partner deleted"
// @Failure 400 {object} map[string]any "Invalid ID or confirmation required"
// @Failure 401 {object} map[string]any "Unauthorized"
// @Failure 403 {object} map[string]any "Forbidden"
// @Failure 404 {object} map[string]any "Partner not found"
// @Failure 500 {object} map[string]any "Internal server error"
// @Router /admin/partners/{id} [delete]
func (handler *AdminHandler) DeletePartner(c *fiber.Ctx) error {
	partnerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Invalid partner ID")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid partner ID",
		})
	}
	partner, err := handler.deps.AdminService.GetPartnerRepository().GetByID(context.Background(), partnerID)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Partner not found")
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Partner not found",
		})
	}
	totalCoupons, _ := handler.deps.AdminService.GetCouponRepository().CountByPartnerID(context.Background(), partnerID)
	activatedCoupons, _ := handler.deps.AdminService.GetCouponRepository().CountActivatedByPartnerID(context.Background(), partnerID)
	confirm := c.Query("confirm") == "true"
	if !confirm {
		handler.deps.Logger.FromContext(c).Error().Msg("Deletion requires confirmation")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Deletion requires confirmation"})
	}
	err = handler.deps.AdminService.DeletePartner(partnerID)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to delete partner")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete partner",
		})
	}
	remainingCoupons, _ := handler.deps.AdminService.GetCouponRepository().CountByPartnerID(context.Background(), partnerID)
	if remainingCoupons > 0 {
		handler.deps.Logger.FromContext(c).Error().Msg("Failed to delete partner")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete partner"})
	}
	handler.deps.Logger.FromContext(c).Info().Interface("context", map[string]any{"partner_id": partner.ID}).Msg("Partner deleted")
	return c.JSON(fiber.Map{
		"deleted_partner": fiber.Map{
			"id":                partner.ID,
			"brand_name":        partner.BrandName,
			"total_coupons":     totalCoupons,
			"activated_coupons": activatedCoupons,
		},
	})
}

// @Summary Get partner statistics
// @Description Returns detailed statistics for a specific partner
// @Tags admin-partners
// @Produce json
// @Security BearerAuth
// @Param id path string true "Partner ID"
// @Success 200 {object} map[string]any "Partner statistics"
// @Failure 400 {object} map[string]any "Invalid ID"
// @Failure 401 {object} map[string]any "Unauthorized"
// @Failure 403 {object} map[string]any "Forbidden"
// @Failure 404 {object} map[string]any "Partner not found"
// @Failure 500 {object} map[string]any "Internal server error"
// @Router /admin/partners/{id}/statistics [get]
func (handler *AdminHandler) GetPartnerStatistics(c *fiber.Ctx) error {
	partnerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Invalid partner ID")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid partner ID",
		})
	}
	totalCoupons, err := handler.deps.AdminService.GetCouponRepository().CountByPartnerID(context.Background(), partnerID)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to get statistics")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get statistics",
		})
	}
	activatedCoupons, err := handler.deps.AdminService.GetCouponRepository().CountActivatedByPartnerID(context.Background(), partnerID)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to get statistics")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get statistics",
		})
	}
	purchasedCoupons, err := handler.deps.AdminService.GetCouponRepository().CountPurchasedByPartnerID(context.Background(), partnerID)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to get statistics")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get statistics",
		})
	}
	unusedCoupons := totalCoupons - activatedCoupons
	activationRate := float64(0)
	if totalCoupons > 0 {
		activationRate = (float64(activatedCoupons) / float64(totalCoupons)) * 100
	}
	handler.deps.Logger.FromContext(c).Info().Interface("context", map[string]any{"partner_id": partnerID}).Msg("Partner statistics retrieved")
	return c.JSON(fiber.Map{
		"partner_id":        partnerID,
		"total_coupons":     totalCoupons,
		"activated_coupons": activatedCoupons,
		"unused_coupons":    unusedCoupons,
		"purchased_coupons": purchasedCoupons,
		"activation_rate":   activationRate,
	})
}

// @Summary Get coupons list with filtering
// @Description Returns list of all coupons with filtering capabilities by code, status, size, style and partner
// @Tags admin-coupons
// @Produce json
// @Security BearerAuth
// @Param search query string false "Search by coupon code"
// @Param partner_id query string false "Partner ID for filtering"
// @Param status query string false "Status for filtering (new/used)"
// @Param size query string false "Size for filtering"
// @Param style query string false "Style for filtering"
// @Param limit query int false "Number of records per page (default 50)"
// @Param offset query int false "Offset for pagination (default 0)"
// @Success 200 {object} map[string]any "Coupons list"
// @Failure 401 {object} map[string]any "Unauthorized"
// @Failure 403 {object} map[string]any "Forbidden"
// @Failure 500 {object} map[string]any "Internal server error"
// @Router /admin/coupons [get]
func (handler *AdminHandler) GetCoupons(c *fiber.Ctx) error {
	filters := map[string]any{}

	if search := c.Query("search"); search != "" {
		filters["code_search"] = search
	}

	if partnerID := c.Query("partner_id"); partnerID != "" {
		if uuid, err := uuid.Parse(partnerID); err == nil {
			filters["partner_id"] = uuid
		}
	}

	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}

	if size := c.Query("size"); size != "" {
		filters["size"] = size
	}

	if style := c.Query("style"); style != "" {
		filters["style"] = style
	}

	coupons, err := handler.deps.AdminService.GetCouponRepository().GetFiltered(context.Background(), filters)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to get coupons")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get coupons",
		})
	}

	result := make([]fiber.Map, len(coupons))
	for i, coupon := range coupons {
		partnerName := "Own"
		isBlocked := coupon.IsBlocked
		if coupon.PartnerID != uuid.Nil {
			if partner, err := handler.deps.AdminService.GetPartnerRepository().GetByID(context.Background(), coupon.PartnerID); err == nil {
				partnerName = partner.BrandName
				if partner.Status == "blocked" {
					isBlocked = true
				}
			}
		}

		result[i] = fiber.Map{
			"id":                coupon.ID,
			"code":              coupon.Code,
			"partner_id":        coupon.PartnerID,
			"partner_name":      partnerName,
			"size":              coupon.Size,
			"style":             coupon.Style,
			"status":            coupon.Status,
			"is_blocked":        isBlocked,
			"is_purchased":      coupon.IsPurchased,
			"purchase_email":    coupon.PurchaseEmail,
			"purchased_at":      coupon.PurchasedAt,
			"used_at":           coupon.UsedAt,
			"zip_url":           coupon.ZipURL,
			"schema_sent_email": coupon.SchemaSentEmail,
			"schema_sent_at":    coupon.SchemaSentAt,
			"created_at":        coupon.CreatedAt,
		}
	}

	handler.deps.Logger.FromContext(c).Info().Interface("context", map[string]any{"count": len(result)}).Msg("Coupons list retrieved")
	return c.JSON(fiber.Map{
		"coupons": result,
		"total":   len(result),
	})
}

// @Summary Get coupons list with pagination for admin
// @Description Returns list of coupons with pagination and extended filtering for administrative panel
// @Tags admin-coupons
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number (default 1)"
// @Param limit query int false "Number of items per page (default 20, max 100)"
// @Param search query string false "Search by coupon code"
// @Param partner_id query string false "Partner ID"
// @Param status query string false "Coupon status (new, used)"
// @Param size query string false "Coupon size"
// @Param style query string false "Coupon style"
// @Param created_from query string false "Creation date from (RFC3339)"
// @Param created_to query string false "Creation date to (RFC3339)"
// @Param used_from query string false "Usage date from (RFC3339)"
// @Param used_to query string false "Usage date to (RFC3339)"
// @Success 200 {object} map[string]any "Coupons with pagination info"
// @Failure 400 {object} map[string]any "Invalid request parameters"
// @Failure 401 {object} map[string]any "Unauthorized"
// @Failure 403 {object} map[string]any "Forbidden"
// @Failure 500 {object} map[string]any "Internal server error"
// @Router /admin/coupons/paginated [get]
func (handler *AdminHandler) GetCouponsPaginated(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	code := c.Query("search")
	status := c.Query("status")
	size := c.Query("size")
	style := c.Query("style")
	partnerIDStr := c.Query("partner_id")
	createdFromStr := c.Query("created_from")
	createdToStr := c.Query("created_to")
	usedFromStr := c.Query("used_from")
	usedToStr := c.Query("used_to")
	sortBy := c.Query("sort_by")
	sortDir := c.Query("sort_dir")

	var partnerID *uuid.UUID
	if partnerIDStr != "" {
		if partnerIDStr == "0000" {
			id := uuid.Nil
			partnerID = &id
		} else if id, err := uuid.Parse(partnerIDStr); err == nil {
			partnerID = &id
		}
	}

	var createdFromPtr, createdToPtr, usedFromPtr, usedToPtr *time.Time
	if createdFromStr != "" {
		if t, err := time.Parse(time.RFC3339, createdFromStr); err == nil {
			createdFromPtr = &t
		}
	}
	if createdToStr != "" {
		if t, err := time.Parse(time.RFC3339, createdToStr); err == nil {
			createdToPtr = &t
		}
	}
	if usedFromStr != "" {
		if t, err := time.Parse(time.RFC3339, usedFromStr); err == nil {
			usedFromPtr = &t
		}
	}
	if usedToStr != "" {
		if t, err := time.Parse(time.RFC3339, usedToStr); err == nil {
			usedToPtr = &t
		}
	}

	coupons, total, err := handler.deps.AdminService.GetCouponRepository().SearchWithPagination(
		context.Background(), code, status, size, style, partnerID, page, limit,
		createdFromPtr, createdToPtr, usedFromPtr, usedToPtr, sortBy, sortDir,
	)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to get coupons")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get coupons",
		})
	}

	result := make([]fiber.Map, len(coupons))
	for i, coupon := range coupons {
		partnerName := "Own"
		if coupon.PartnerID != uuid.Nil {
			if partner, err := handler.deps.AdminService.GetPartnerRepository().GetByID(context.Background(), coupon.PartnerID); err == nil {
				partnerName = partner.BrandName
			}
		}

		result[i] = fiber.Map{
			"id":                coupon.ID,
			"code":              coupon.Code,
			"partner_id":        coupon.PartnerID,
			"partner_name":      partnerName,
			"size":              coupon.Size,
			"style":             coupon.Style,
			"status":            coupon.Status,
			"is_purchased":      coupon.IsPurchased,
			"purchase_email":    coupon.PurchaseEmail,
			"purchased_at":      coupon.PurchasedAt,
			"used_at":           coupon.UsedAt,
			"zip_url":           coupon.ZipURL,
			"schema_sent_email": coupon.SchemaSentEmail,
			"schema_sent_at":    coupon.SchemaSentAt,
			"created_at":        coupon.CreatedAt,
		}
	}

	totalPages := (int(total) + limit - 1) / limit
	hasNext := page < totalPages
	hasPrev := page > 1

	handler.deps.Logger.FromContext(c).Info().Interface("context", map[string]any{
		"page":  page,
		"limit": limit,
		"total": total,
	}).Msg("Coupons paginated list retrieved")
	return c.JSON(fiber.Map{
		"coupons": result,
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

// @Summary Create coupons
// @Description Creates new coupons in batch mode
// @Tags admin-coupons
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body coupon.CreateCouponRequest true "Coupon creation parameters"
// @Success 201 {object} map[string]any "Coupons created"
// @Failure 400 {object} map[string]any "Request error"
// @Failure 401 {object} map[string]any "Unauthorized"
// @Failure 403 {object} map[string]any "Forbidden"
// @Failure 500 {object} map[string]any "Internal server error"
// @Router /admin/coupons [post]
func (handler *AdminHandler) CreateCoupons(c *fiber.Ctx) error {
	var req coupon.CreateCouponRequest

	if err := c.BodyParser(&req); err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to parse request body")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to parse request body",
		})
	}

	if req.Count < 1 || req.Count > 10000 {
		handler.deps.Logger.FromContext(c).Error().Msg("Invalid request")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	var partnerCode string = "0000"
	effectivePartnerID := req.PartnerID
	if req.PartnerID != uuid.Nil {
		partner, err := handler.deps.AdminService.GetPartnerRepository().GetByID(context.Background(), req.PartnerID)
		if err != nil {
			handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Resource not found")
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Resource not found",
			})
		}
		partnerCode = partner.PartnerCode
	} else {
		own, err := handler.deps.AdminService.GetPartnerRepository().GetByPartnerCode(context.Background(), "0000")
		if err != nil {
			handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to get own partner")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to get own partner",
			})
		}
		effectivePartnerID = own.ID
	}

	coupons := make([]*coupon.Coupon, 0, req.Count)
	codes := make([]string, 0, req.Count)

	for i := 0; i < req.Count; i++ {
		code, err := randomCouponCode.GenerateUniqueCouponCode(partnerCode, handler.deps.AdminService.GetCouponRepository())
		if err != nil {
			handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to generate coupon code")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to generate coupon code",
			})
		}

		coupons = append(coupons, &coupon.Coupon{
			Code:      code,
			PartnerID: effectivePartnerID,
			Size:      string(req.Size),
			Style:     string(req.Style),
			Status:    string(coupon.StatusNew),
		})
		codes = append(codes, code)
	}

	if err := handler.deps.AdminService.GetCouponRepository().CreateBatch(context.Background(), coupons); err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to create coupons")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create coupons",
		})
	}

	handler.deps.Logger.FromContext(c).Info().Interface("context", map[string]any{
		"count":      req.Count,
		"partner_id": req.PartnerID,
		"size":       req.Size,
		"style":      req.Style,
	}).Msg("Coupons created")
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":     "Coupons created successfully",
		"count":       req.Count,
		"codes":       codes,
		"partner_id":  req.PartnerID,
		"size":        req.Size,
		"style":       req.Style,
		"codes_range": []string{codes[0], codes[len(codes)-1]},
	})
}

// @Summary Export partner coupons for admin
// @Description Exports coupons of specific partner with status "new" in .txt or .csv format
// @Tags admin-coupons
// @Produce text/plain,text/csv
// @Security BearerAuth
// @Param id path string true "Partner ID"
// @Param format query string false "File format (txt or csv)" default(txt)
// @Success 200 {string} string "Partner coupons file"
// @Failure 400 {object} map[string]any "Invalid partner ID"
// @Failure 401 {object} map[string]any "Unauthorized"
// @Failure 403 {object} map[string]any "Forbidden"
// @Failure 404 {object} map[string]any "Partner not found"
// @Failure 500 {object} map[string]any "Internal server error"
// @Router /admin/coupons/export/partner/{id} [get]
func (handler *AdminHandler) ExportPartnerCoupons(c *fiber.Ctx) error {
	partnerIDStr := c.Params("id")
	partnerID, err := uuid.Parse(partnerIDStr)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Invalid ID")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid ID",
		})
	}

	format := strings.ToLower(c.Query("format", "txt"))
	if format != "txt" && format != "csv" {
		format = "txt"
	}

	_, err = handler.deps.AdminService.GetPartnerRepository().GetByID(context.Background(), partnerID)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Partner not found")
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Partner not found",
		})
	}

	options := coupon.ExportOptionsRequest{
		Format:        coupon.ExportFormatAdmin,
		PartnerID:     &partnerIDStr,
		Status:        "new",
		FileFormat:    format,
		IncludeHeader: true,
	}

	content, filename, contentType, err := handler.deps.AdminService.ExportCouponsAdvanced(options)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to export partner coupons")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to export partner coupons",
		})
	}

	c.Set("Content-Type", contentType)
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Set("Cache-Control", "no-cache")

	handler.deps.Logger.FromContext(c).Info().Interface("context", map[string]any{
		"partner_id": partnerID,
		"format":     format,
		"filename":   filename,
	}).Msg("Partner coupons exported")
	return c.Send(content)
}

// @Summary Batch delete coupons for admin
// @Description Deletes multiple coupons by their IDs in administrative panel
// @Tags admin-coupons
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body map[string][]string true "List of coupon IDs to delete"
// @Success 200 {object} map[string]any "Batch deletion result"
// @Failure 400 {object} map[string]any "Request error"
// @Failure 401 {object} map[string]any "Unauthorized"
// @Failure 403 {object} map[string]any "Forbidden"
// @Failure 500 {object} map[string]any "Internal server error"
// @Router /admin/coupons/batch-delete [post]
func (handler *AdminHandler) BatchDeleteCoupons(c *fiber.Ctx) error {
	var req struct {
		CouponIDs []string `json:"coupon_ids" validate:"required,min=1"`
		Confirm   bool     `json:"confirm" validate:"required"`
	}

	if err := c.BodyParser(&req); err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to parse request body")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to parse request body",
		})
	}

	if !req.Confirm {
		handler.deps.Logger.FromContext(c).Error().Msg("Bad request")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Bad request"})
	}

	ids := make([]uuid.UUID, 0, len(req.CouponIDs))
	for _, idStr := range req.CouponIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Invalid ID")
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid ID",
			})
		}
		ids = append(ids, id)
	}

	deletedCount, err := handler.deps.AdminService.GetCouponRepository().BatchDelete(context.Background(), ids)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to delete coupons")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete coupons",
		})
	}

	handler.deps.Logger.FromContext(c).Info().Interface("context", map[string]any{
		"deleted_count": deletedCount,
		"requested":     len(req.CouponIDs),
	}).Msg("Coupons batch deleted")
	return c.JSON(fiber.Map{
		"message":       "Coupons deleted successfully",
		"deleted_count": deletedCount,
		"requested":     len(req.CouponIDs),
	})
}

// @Summary Get coupon details
// @Description Returns detailed information about coupon by ID
// @Tags admin-coupons
// @Produce json
// @Security BearerAuth
// @Param id path string true "Coupon ID"
// @Success 200 {object} map[string]any "Coupon information"
// @Failure 400 {object} map[string]any "Invalid ID"
// @Failure 401 {object} map[string]any "Unauthorized"
// @Failure 403 {object} map[string]any "Forbidden"
// @Failure 404 {object} map[string]any "Coupon not found"
// @Failure 500 {object} map[string]any "Internal server error"
// @Router /admin/coupons/{id} [get]
func (handler *AdminHandler) GetCoupon(c *fiber.Ctx) error {
	couponID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Invalid ID")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid ID",
		})
	}

	coupon, err := handler.deps.AdminService.GetCouponRepository().GetByID(context.Background(), couponID)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Resource not found")
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Resource not found",
		})
	}

	var partnerName string = "Own"
	if coupon.PartnerID != uuid.Nil {
		partner, err := handler.deps.AdminService.GetPartnerRepository().GetByID(context.Background(), coupon.PartnerID)
		if err == nil {
			partnerName = partner.BrandName
		}
	}

	handler.deps.Logger.FromContext(c).Info().Interface("context", map[string]any{"coupon_id": coupon.ID}).Msg("Coupon details retrieved")
	return c.JSON(fiber.Map{
		"id":                coupon.ID,
		"code":              coupon.Code,
		"partner_id":        coupon.PartnerID,
		"partner_name":      partnerName,
		"size":              coupon.Size,
		"style":             coupon.Style,
		"status":            coupon.Status,
		"is_purchased":      coupon.IsPurchased,
		"purchase_email":    coupon.PurchaseEmail,
		"purchased_at":      coupon.PurchasedAt,
		"used_at":           coupon.UsedAt,
		"zip_url":           coupon.ZipURL,
		"schema_sent_email": coupon.SchemaSentEmail,
		"schema_sent_at":    coupon.SchemaSentAt,
		"created_at":        coupon.CreatedAt,
	})
}

// @Summary Reset coupon
// @Description Resets coupon to "new" status with removal of all activation data
// @Tags admin-coupons
// @Produce json
// @Security BearerAuth
// @Param id path string true "Coupon ID"
// @Success 200 {object} map[string]any "Coupon reset"
// @Failure 400 {object} map[string]any "Invalid ID"
// @Failure 401 {object} map[string]any "Unauthorized"
// @Failure 403 {object} map[string]any "Forbidden"
// @Failure 404 {object} map[string]any "Coupon not found"
// @Failure 500 {object} map[string]any "Internal server error"
// @Router /admin/coupons/{id}/reset [patch]
func (handler *AdminHandler) ResetCoupon(c *fiber.Ctx) error {
	couponID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Invalid ID")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid ID",
		})
	}

	if err := handler.deps.AdminService.ResetCoupon(couponID); err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to reset coupon")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to reset coupon",
		})
	}

	handler.deps.Logger.FromContext(c).Info().Interface("context", map[string]any{"coupon_id": couponID}).Msg("Coupon reset")
	return c.JSON(fiber.Map{"message": "Coupon reset successfully"})
}

// @Summary Delete coupon
// @Description Deletes coupon by ID with confirmation
// @Tags admin-coupons
// @Produce json
// @Security BearerAuth
// @Param id path string true "Coupon ID"
// @Param confirm query boolean false "Delete confirmation (true/false)"
// @Success 200 {object} map[string]any "Coupon deleted"
// @Failure 400 {object} map[string]any "Invalid ID or confirmation required"
// @Failure 401 {object} map[string]any "Unauthorized"
// @Failure 403 {object} map[string]any "Forbidden"
// @Failure 404 {object} map[string]any "Coupon not found"
// @Failure 500 {object} map[string]any "Internal server error"
// @Router /admin/coupons/{id} [delete]
func (handler *AdminHandler) DeleteCoupon(c *fiber.Ctx) error {
	couponID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Invalid ID")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid ID",
		})
	}

	coupon, err := handler.deps.AdminService.GetCouponRepository().GetByID(context.Background(), couponID)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Resource not found")
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Resource not found",
		})
	}

	confirm := c.Query("confirm") == "true"
	if !confirm {
		handler.deps.Logger.FromContext(c).Error().Msg("Deletion requires confirmation")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Deletion requires confirmation"})
	}

	if err := handler.deps.AdminService.DeleteCoupon(couponID); err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to delete coupon")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete coupon",
		})
	}

	handler.deps.Logger.FromContext(c).Info().Interface("context", map[string]any{"coupon_id": couponID}).Msg("Coupon deleted")
	return c.JSON(fiber.Map{
		"message": "Coupon deleted successfully",
		"deleted_coupon": fiber.Map{
			"id":     coupon.ID,
			"code":   coupon.Code,
			"status": coupon.Status,
		},
	})
}

// @Summary Download coupon materials
// @Description Downloads archive with materials of used coupon (original, preview, schema)
// @Tags admin-coupons
// @Produce application/zip
// @Security BearerAuth
// @Param id path string true "Coupon ID"
// @Success 200 {string} string "ZIP archive with materials"
// @Failure 400 {object} map[string]any "Invalid coupon ID"
// @Failure 401 {object} map[string]any "Unauthorized"
// @Failure 403 {object} map[string]any "Forbidden"
// @Failure 404 {object} map[string]any "Coupon not found or not used"
// @Failure 500 {object} map[string]any "Internal server error"
// @Router /admin/coupons/{id}/download-materials [get]
func (handler *AdminHandler) DownloadCouponMaterials(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Invalid coupon ID")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid coupon ID",
		})
	}

	archiveData, filename, err := handler.deps.AdminService.DownloadCouponMaterials(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Resource not found")
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Resource not found",
			})
		}
		if strings.Contains(err.Error(), "must be used or completed") {
			handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Coupon must be used or completed to download materials")
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Coupon must be used or completed to download materials",
			})
		}
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to download materials")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to download materials",
		})
	}

	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Set("Content-Type", "application/zip")

	handler.deps.Logger.FromContext(c).Info().Interface("context", map[string]any{"coupon_id": id}).Msg("Coupon materials downloaded")
	return c.Send(archiveData)
}

// @Summary Batch reset coupons (admin)
// @Description Resets multiple coupons to original state through administrative panel
// @Tags admin-coupons
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body coupon.BatchResetRequest true "List of coupon IDs to reset"
// @Success 200 {object} coupon.BatchResetResponse "Batch reset result"
// @Failure 400 {object} map[string]any "Invalid request"
// @Failure 401 {object} map[string]any "Unauthorized"
// @Failure 403 {object} map[string]any "Forbidden"
// @Failure 500 {object} map[string]any "Internal server error"
// @Router /admin/coupons/batch/reset [post]
func (handler *AdminHandler) BatchResetCoupons(c *fiber.Ctx) error {
	var req coupon.BatchResetRequest
	if err := c.BodyParser(&req); err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Invalid request body")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if len(req.CouponIDs) == 0 {
		handler.deps.Logger.FromContext(c).Error().Msg("Coupon ID required")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Coupon ID required"})
	}

	if len(req.CouponIDs) > 1000 {
		handler.deps.Logger.FromContext(c).Error().Msg("Too many items")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Too many items"})
	}

	response, err := handler.deps.AdminService.BatchResetCoupons(req.CouponIDs)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to batch reset coupons")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to batch reset coupons",
		})
	}

	handler.deps.Logger.FromContext(c).Info().Interface("context", map[string]any{"count": len(req.CouponIDs)}).Msg("Coupons batch reset")
	return c.JSON(response)
}

// @Summary Advanced export coupons (admin)
// @Description Exports coupons in various formats (TXT, CSV, XLSX) with configurable options
// @Tags admin-coupons
// @Accept json
// @Produce application/octet-stream
// @Security BearerAuth
// @Param request body coupon.ExportOptionsRequest true "Export options"
// @Success 200 {string} string "Export file"
// @Failure 400 {object} map[string]any "Invalid request"
// @Failure 401 {object} map[string]any "Unauthorized"
// @Failure 403 {object} map[string]any "Forbidden"
// @Failure 500 {object} map[string]any "Internal server error"
// @Router /admin/coupons/export-advanced [post]
func (handler *AdminHandler) ExportCouponsAdvanced(c *fiber.Ctx) error {
	var req coupon.ExportOptionsRequest
	if err := c.BodyParser(&req); err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Invalid request body")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Format == "" {
		req.Format = coupon.ExportFormatAdmin
	}

	if req.FileFormat == "" {
		req.FileFormat = "csv"
	}

	content, filename, contentType, err := handler.deps.AdminService.ExportCouponsAdvanced(req)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to export coupons")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to export coupons",
		})
	}

	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Set("Content-Type", contentType)

	handler.deps.Logger.FromContext(c).Info().Interface("context", map[string]any{
		"format":      req.Format,
		"file_format": req.FileFormat,
		"filename":    filename,
	}).Msg("Coupons exported")
	return c.Send(content)
}

// @Summary Get general statistics
// @Description Returns general system statistics
// @Tags admin-statistics
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]any "Statistics"
// @Failure 401 {object} map[string]any "Unauthorized"
// @Failure 403 {object} map[string]any "Forbidden"
// @Failure 500 {object} map[string]any "Internal server error"
// @Router /admin/statistics [get]
func (handler *AdminHandler) GetStatistics(c *fiber.Ctx) error {
	stats, err := handler.deps.AdminService.GetStatistics()
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to get statistics")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get statistics",
		})
	}
	handler.deps.Logger.FromContext(c).Info().Msg("General statistics retrieved")
	return c.JSON(stats)
}

// @Summary Get partners statistics
// @Description Returns detailed statistics for all partners
// @Tags admin-statistics
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]any "Partners statistics"
// @Failure 401 {object} map[string]any "Unauthorized"
// @Failure 403 {object} map[string]any "Forbidden"
// @Failure 500 {object} map[string]any "Internal server error"
// @Router /admin/statistics/partners [get]
func (handler *AdminHandler) GetPartnersStatistics(c *fiber.Ctx) error {
	dateFromStr := strings.TrimSpace(c.Query("date_from"))
	dateToStr := strings.TrimSpace(c.Query("date_to"))
	partnerStatus := strings.TrimSpace(c.Query("status"))
	search := strings.TrimSpace(c.Query("search"))
	partnerCode := strings.TrimSpace(c.Query("partner_code"))

	partners, err := handler.deps.AdminService.GetPartnerRepository().GetAll(context.Background(), "created_at", "desc")
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to fetch partners")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch partners",
		})
	}

	type partnerView struct {
		ID          uuid.UUID
		BrandName   string
		PartnerCode string
		Status      string
	}
	var filteredPartners []partnerView
	for _, p := range partners {
		if partnerStatus != "" && !strings.EqualFold(p.Status, partnerStatus) {
			continue
		}
		if partnerCode != "" && p.PartnerCode != partnerCode {
			continue
		}
		if search != "" && !strings.Contains(strings.ToLower(p.BrandName+" "+p.PartnerCode), strings.ToLower(search)) {
			continue
		}
		filteredPartners = append(filteredPartners, partnerView{ID: p.ID, BrandName: p.BrandName, PartnerCode: p.PartnerCode, Status: p.Status})
	}

	var dateFromPtr, dateToPtr *time.Time
	if dateFromStr != "" {
		if t, err := time.Parse(time.RFC3339, dateFromStr); err == nil {
			dateFromPtr = &t
		}
	}
	if dateToStr != "" {
		if t, err := time.Parse(time.RFC3339, dateToStr); err == nil {
			dateToPtr = &t
		}
	}

	var partnersStats []fiber.Map
	for _, partner := range filteredPartners {
		totalCoupons, _ := handler.deps.AdminService.GetCouponRepository().CountCreatedByPartnerInRange(context.Background(), partner.ID, dateFromPtr, dateToPtr)
		activatedCoupons, _ := handler.deps.AdminService.GetCouponRepository().CountActivatedByPartnerInRange(context.Background(), partner.ID, dateFromPtr, dateToPtr)
		purchasedCoupons, _ := handler.deps.AdminService.GetCouponRepository().CountPurchasedByPartnerInRange(context.Background(), partner.ID, dateFromPtr, dateToPtr)

		activationRate := float64(0)
		if totalCoupons > 0 {
			activationRate = (float64(activatedCoupons) / float64(totalCoupons)) * 100
		}

		partnersStats = append(partnersStats, fiber.Map{
			"partner_id":   partner.ID,
			"brand_name":   partner.BrandName,
			"partner_code": partner.PartnerCode,
			"statistics": fiber.Map{
				"total":     totalCoupons,
				"used":      activatedCoupons,
				"purchased": purchasedCoupons,
			},
			"activation_rate": activationRate,
			"status":          partner.Status,
		})
	}

	handler.deps.Logger.FromContext(c).Info().Interface("context", map[string]any{"count": len(partnersStats)}).Msg("Partners statistics retrieved")
	return c.JSON(fiber.Map{
		"partners": partnersStats,
	})
}

// @Summary Get system statistics
// @Description Returns detailed system statistics: performance, load, queue status
// @Tags admin-statistics
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]any "System statistics"
// @Failure 401 {object} map[string]any "Unauthorized"
// @Failure 403 {object} map[string]any "Forbidden"
// @Failure 500 {object} map[string]any "Internal server error"
// @Router /admin/statistics/system [get]
func (handler *AdminHandler) GetSystemStatistics(c *fiber.Ctx) error {
	stats, err := handler.deps.AdminService.GetSystemStatistics()
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to get system statistics")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get system statistics",
		})
	}
	handler.deps.Logger.FromContext(c).Info().Msg("System statistics retrieved")
	return c.JSON(stats)
}

// @Summary Get all image processing tasks
// @Description Returns list of all image processing tasks with filtering
// @Tags admin-images
// @Produce json
// @Security BearerAuth
// @Param status query string false "Filter by status (queued, processing, completed, failed)"
// @Param partner_id query string false "Filter by partner ID"
// @Param limit query int false "Record limit (default 50)"
// @Param offset query int false "Offset (default 0)"
// @Success 200 {array} map[string]any "Image processing tasks list"
// @Failure 401 {object} map[string]any "Unauthorized"
// @Failure 403 {object} map[string]any "Forbidden"
// @Failure 500 {object} map[string]any "Internal server error"
// @Router /admin/images [get]
func (handler *AdminHandler) GetAllImages(c *fiber.Ctx) error {
	status := strings.TrimSpace(c.Query("status"))
	_ = c.Query("partner_id")
	limit := c.QueryInt("limit", 50)
	offset := c.QueryInt("offset", 0)
	if limit < 1 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}

	var (
		tasks []*image.Image
		err   error
	)
	if status != "" {
		tasks, err = handler.deps.AdminService.GetImageRepository().GetByStatus(c.UserContext(), status)
	} else {
		tasks, err = handler.deps.AdminService.GetImageRepository().GetAll(c.UserContext())
	}
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to get images")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get images",
		})
	}

	total := len(tasks)
	start := offset
	if start > total {
		start = total
	}
	end := offset + limit
	if end > total {
		end = total
	}
	paged := tasks[start:end]

	result := make([]fiber.Map, 0, len(paged))
	for _, t := range paged {
		result = append(result, fiber.Map{
			"id":                     t.ID,
			"coupon_id":              t.CouponID,
			"original_image_s3_key":  t.OriginalImageS3Key,
			"edited_image_s3_key":    t.EditedImageS3Key,
			"preview_s3_key":         t.PreviewS3Key,
			"processed_image_s3_key": t.ProcessedImageS3Key,
			"processing_params":      t.ProcessingParams,
			"user_email":             t.UserEmail,
			"status":                 t.Status,
			"priority":               t.Priority,
			"retry_count":            t.RetryCount,
			"max_retries":            t.MaxRetries,
			"created_at":             t.CreatedAt,
			"started_at":             t.StartedAt,
			"completed_at":           t.CompletedAt,
			"error_message":          t.ErrorMessage,
		})
	}

	handler.deps.Logger.FromContext(c).Info().Interface("context", map[string]any{
		"total":  total,
		"limit":  limit,
		"offset": offset,
	}).Msg("Image tasks retrieved")
	return c.JSON(fiber.Map{
		"images": result,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// @Summary Get image task details
// @Description Returns detailed information about image processing task
// @Tags admin-images
// @Produce json
// @Security BearerAuth
// @Param id path string true "Task ID"
// @Success 200 {object} map[string]any "Task details"
// @Failure 401 {object} map[string]any "Unauthorized"
// @Failure 403 {object} map[string]any "Forbidden"
// @Failure 404 {object} map[string]any "Task not found"
// @Failure 500 {object} map[string]any "Internal server error"
// @Router /admin/images/{id} [get]
func (handler *AdminHandler) GetImageDetails(c *fiber.Ctx) error {
	imageID := c.Params("id")
	imageUUID, err := uuid.Parse(imageID)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Invalid ID")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid ID",
		})
	}

	task, err := handler.deps.AdminService.GetImageDetails(imageUUID)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Image task not found")
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Image task not found",
		})
	}

	handler.deps.Logger.FromContext(c).Info().Interface("context", map[string]any{"task_id": imageUUID}).Msg("Image task details retrieved")
	return c.JSON(task)
}

// @Summary Delete image processing task
// @Description Deletes image processing task and related files
// @Tags admin-images
// @Produce json
// @Security BearerAuth
// @Param id path string true "Task ID"
// @Success 200 {object} map[string]any "Task deleted"
// @Failure 401 {object} map[string]any "Unauthorized"
// @Failure 403 {object} map[string]any "Forbidden"
// @Failure 404 {object} map[string]any "Task not found"
// @Failure 500 {object} map[string]any "Internal server error"
// @Router /admin/images/{id} [delete]
func (handler *AdminHandler) DeleteImageTask(c *fiber.Ctx) error {
	imageID := c.Params("id")

	imageUUID, err := uuid.Parse(imageID)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Invalid ID")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid ID",
		})
	}

	if err := handler.deps.AdminService.DeleteImageTask(imageUUID); err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Image task not found")
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Image task not found",
		})
	}

	handler.deps.Logger.FromContext(c).Info().Interface("context", map[string]any{"task_id": imageUUID}).Msg("Image task deleted")
	return c.JSON(fiber.Map{"message": "Image task deleted"})
}

// @Summary Retry image processing
// @Description Restarts failed image processing task
// @Tags admin-images
// @Produce json
// @Security BearerAuth
// @Param id path string true "Task ID"
// @Success 200 {object} map[string]any "Task queued for retry"
// @Failure 401 {object} map[string]any "Unauthorized"
// @Failure 403 {object} map[string]any "Forbidden"
// @Failure 404 {object} map[string]any "Task not found"
// @Failure 400 {object} map[string]any "Task cannot be retried"
// @Failure 500 {object} map[string]any "Internal server error"
// @Router /admin/images/{id}/retry [post]
func (handler *AdminHandler) RetryImageTask(c *fiber.Ctx) error {
	imageID := c.Params("id")

	imageUUID, err := uuid.Parse(imageID)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Invalid ID")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid ID",
		})
	}

	if err := handler.deps.AdminService.RetryImageTask(imageUUID); err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Image task not found")
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Image task not found",
		})
	}

	handler.deps.Logger.FromContext(c).Info().Interface("context", map[string]any{"task_id": imageUUID}).Msg("Image task queued for retry")
	return c.JSON(fiber.Map{"message": "Image task queued for retry"})
}

// @Summary Force nginx update
// @Description Forces generation and update of nginx configuration via CI/CD pipeline
// @Tags admin-domains
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]any "Nginx configuration updated"
// @Failure 401 {object} map[string]any "Unauthorized"
// @Failure 403 {object} map[string]any "Forbidden"
// @Failure 500 {object} map[string]any "Internal server error"
// @Router /admin/nginx/deploy [post]
func (handler *AdminHandler) DeployNginxConfig(c *fiber.Ctx) error {
	err := handler.deps.AdminService.DeployNginxConfig()
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to update nginx config")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update nginx config",
		})
	}

	handler.deps.Logger.FromContext(c).Info().Msg("Nginx configuration updated")
	return c.JSON(fiber.Map{"message": "Nginx configuration updated successfully"})
}

// @Summary Get partner article grid
// @Description Gets the complete article grid for a specific partner with all SKUs organized by marketplace, style, and size
// @Tags admin-partners
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Partner ID"
// @Success 200 {object} map[string]map[string]map[string]string "Article grid (marketplace -> style -> size -> sku)"
// @Failure 400 {object} map[string]any "Invalid partner ID"
// @Failure 500 {object} map[string]any "Failed to get article grid"
// @Router /admin/partners/{id}/articles/grid [get]
func (handler *AdminHandler) GetPartnerArticleGrid(c *fiber.Ctx) error {
	partnerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid partner ID",
		})
	}

	grid, err := handler.deps.AdminService.GetPartnerRepository().GetArticleGrid(c.Context(), partnerID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get article grid",
		})
	}

	return c.JSON(grid)
}

// @Summary Update partner article SKU
// @Description Updates a specific SKU in the partner's article grid for a given marketplace, style, and size combination
// @Tags admin-partners
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Partner ID"
// @Param request body object{marketplace=string,style=string,size=string,sku=string} true "SKU update request"
// @Success 200 {object} map[string]string "SKU updated successfully"
// @Failure 400 {object} map[string]any "Invalid request"
// @Failure 500 {object} map[string]any "Failed to update SKU"
// @Router /admin/partners/{id}/articles/sku [put]
func (handler *AdminHandler) UpdatePartnerArticleSKU(c *fiber.Ctx) error {
	partnerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid partner ID",
		})
	}

	var req struct {
		Marketplace string `json:"marketplace"`
		Style       string `json:"style"`
		Size        string `json:"size"`
		SKU         string `json:"sku"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	err = handler.deps.AdminService.GetPartnerRepository().UpdateArticleSKU(
		c.Context(), partnerID, req.Size, req.Style, req.Marketplace, req.SKU)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update article SKU",
		})
	}

	return c.JSON(fiber.Map{
		"message": "SKU updated successfully",
	})
}

// @Summary Generate product URL
// @Description Generates a product URL for a specific marketplace, style, and size combination using partner's SKU and templates
// @Tags admin-partners
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Partner ID"
// @Param request body GenerateProductURLRequest true "URL generation request"
// @Success 200 {object} GenerateProductURLResponse "Generated product URL with details"
// @Failure 400 {object} map[string]any "Invalid request"
// @Failure 404 {object} map[string]any "Partner not found"
// @Failure 500 {object} map[string]any "Internal server error"
// @Router /admin/partners/{id}/articles/generate-url [post]
func (handler *AdminHandler) GenerateProductURL(c *fiber.Ctx) error {
	partnerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid partner ID",
		})
	}

	var req GenerateProductURLRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Create marketplace service
	marketplaceRepo := marketplace.NewPartnerRepositoryAdapter(handler.deps.AdminService.GetPartnerRepository())
	marketplaceService := marketplace.NewService(marketplaceRepo)

	// Convert request to marketplace format
	marketplaceReq := &marketplace.ProductURLRequest{
		PartnerID:   partnerID,
		Marketplace: marketplace.Marketplace(req.Marketplace),
		Size:        req.Size,
		Style:       req.Style,
	}

	// Generate URL using marketplace service
	response, err := marketplaceService.GenerateProductURL(marketplaceReq)
	if err != nil {
		if strings.Contains(err.Error(), "partner not found") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Partner not found",
			})
		}
		if strings.Contains(err.Error(), "validation failed") {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate product URL",
		})
	}

	handler.deps.Logger.FromContext(c).Info().Interface("context", map[string]any{
		"partner_id":  partnerID,
		"marketplace": req.Marketplace,
		"size":        req.Size,
		"style":       req.Style,
		"has_sku":     response.HasArticle,
		"url_length":  len(response.URL),
	}).Msg("Product URL generated")

	return c.JSON(GenerateProductURLResponse{
		URL:         response.URL,
		SKU:         response.SKU,
		HasArticle:  response.HasArticle,
		PartnerName: response.PartnerName,
		Marketplace: string(response.Marketplace),
		Size:        response.Size,
		Style:       response.Style,
	})
}
