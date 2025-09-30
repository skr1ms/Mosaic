package admin

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/skr1ms/mosaic/internal/coupon"
	"github.com/skr1ms/mosaic/internal/image"
	"github.com/skr1ms/mosaic/internal/partner"
	"github.com/skr1ms/mosaic/pkg/bcrypt"
	"github.com/skr1ms/mosaic/pkg/gitlab"
	"github.com/skr1ms/mosaic/pkg/randomCouponCode"
	"github.com/skr1ms/mosaic/pkg/updatePartnerData"
	validateData "github.com/skr1ms/mosaic/pkg/validateData"
)

type AdminServiceDeps struct {
	AdminRepository   AdminRepositoryInterface
	PartnerRepository PartnerRepositoryInterface
	CouponRepository  CouponRepositoryInterface
	ImageRepository   ImageRepositoryInterface
	S3Client          S3ClientInterface
	RedisClient       RedisClientInterface
	GitLabClient      *gitlab.Client
}

type AdminService struct {
	deps *AdminServiceDeps
}

func NewAdminService(deps *AdminServiceDeps) *AdminService {
	return &AdminService{
		deps: deps,
	}
}

// Repository access methods
func (s *AdminService) GetPartnerRepository() PartnerRepositoryInterface {
	return s.deps.PartnerRepository
}

func (s *AdminService) GetCouponRepository() CouponRepositoryInterface {
	return s.deps.CouponRepository
}

func (s *AdminService) GetImageRepository() ImageRepositoryInterface {
	return s.deps.ImageRepository
}

func (s *AdminService) GetS3Client() S3ClientInterface {
	return s.deps.S3Client
}

// CreateAdmin creates new admin user with unique login/email validation and password hashing
func (s *AdminService) CreateAdmin(req CreateAdminRequest) (*Admin, error) {
	existingAdmin, err := s.deps.AdminRepository.GetByLogin(req.Login)
	if err == nil && existingAdmin != nil {
		return nil, fmt.Errorf("admin already exists: %w", err)
	}

	if req.Email != "" {
		if existingByEmail, errByEmail := s.deps.AdminRepository.GetByEmail(req.Email); errByEmail == nil && existingByEmail != nil {
			return nil, fmt.Errorf("admin already exists: %w", errByEmail)
		}
	}

	hashedPassword, err := bcrypt.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	admin := &Admin{
		Login:    req.Login,
		Email:    req.Email,
		Password: hashedPassword,
		Role:     "admin",
	}

	if err := s.deps.AdminRepository.Create(admin); err != nil {
		return nil, fmt.Errorf("failed to create admin: %w", err)
	}

	return admin, nil
}

// GetDashboardData aggregates statistics for admin dashboard including coupons, partners, and image processing
func (s *AdminService) GetDashboardData() (map[string]any, error) {
	result := make(map[string]any)

	allCoupons, err := s.deps.CouponRepository.GetAll(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to find all coupons: %w", err)
	}

	totalCoupons := len(allCoupons)
	usedCoupons := 0
	purchasedCoupons := 0
	for _, c := range allCoupons {
		if c.Status == "used" || c.Status == "completed" {
			usedCoupons++
		}
		if c.IsPurchased {
			purchasedCoupons++
		}
	}

	result["coupons"] = map[string]int{
		"total":     totalCoupons,
		"used":      usedCoupons,
		"purchased": purchasedCoupons,
		"new":       totalCoupons - usedCoupons,
	}

	allPartners, err := s.deps.PartnerRepository.GetAll(context.Background(), "created_at", "desc")
	if err != nil {
		return nil, fmt.Errorf("failed to find all partners: %w", err)
	}

	activePartners := 0
	for _, p := range allPartners {
		if p.Status == "active" {
			activePartners++
		}
	}

	result["partners"] = map[string]int{
		"total":  len(allPartners),
		"active": activePartners,
	}

	recentCoupons, err := s.deps.CouponRepository.GetRecentActivated(context.Background(), 10)
	if err == nil {
		result["recent_activations"] = recentCoupons
	}

	allImages, err := s.deps.ImageRepository.GetAll(context.Background())
	if err == nil {
		processingImages := 0
		completedImages := 0
		failedImages := 0

		for _, img := range allImages {
			switch img.Status {
			case "processing":
				processingImages++
			case "completed":
				completedImages++
			case "failed":
				failedImages++
			}
		}

		result["image_processing"] = map[string]int{
			"total":      len(allImages),
			"processing": processingImages,
			"completed":  completedImages,
			"failed":     failedImages,
		}
	}

	return result, nil
}

// GetPartners retrieves partners with optional filtering and sorting
func (s *AdminService) GetPartners(search, status, sortBy, order string) ([]*partner.Partner, error) {
	var partners []*partner.Partner
	var err error

	if search == "" && status == "" {
		partners, err = s.deps.PartnerRepository.GetAll(context.Background(), sortBy, order)
	} else {
		partners, err = s.deps.PartnerRepository.Search(context.Background(), search, status, sortBy, order)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to find all partners: %w", err)
	}

	return partners, nil
}

// GetPartnerDetail retrieves comprehensive partner information including statistics and profile change history
func (s *AdminService) GetPartnerDetail(partnerID uuid.UUID) (*PartnerDetailResponse, error) {

	partner, err := s.deps.PartnerRepository.GetByID(context.Background(), partnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get partner: %w", err)
	}

	totalCoupons, err := s.deps.CouponRepository.CountByPartnerID(context.Background(), partnerID)
	if err != nil {
		totalCoupons = 0
	}

	activatedCoupons, err := s.deps.CouponRepository.CountActivatedByPartnerID(context.Background(), partnerID)
	if err != nil {
		activatedCoupons = 0
	}

	unusedCoupons := totalCoupons - activatedCoupons

	lastActivity, err := s.deps.CouponRepository.GetLastActivityByPartner(context.Background(), partnerID)
	if err != nil {
		lastActivity = nil
	}

	profileChangesLogs, err := s.deps.AdminRepository.GetProfileChangesByPartnerID(partnerID)
	if err != nil {
		profileChangesLogs = []*ProfileChangeLog{}
	}

	profileChanges := make([]ProfileChange, len(profileChangesLogs))
	for i, changeLog := range profileChangesLogs {
		profileChanges[i] = ProfileChange{
			ID:        changeLog.ID,
			PartnerID: changeLog.PartnerID,
			Field:     changeLog.Field,
			OldValue:  changeLog.OldValue,
			NewValue:  changeLog.NewValue,
			ChangedBy: changeLog.ChangedBy,
			ChangedAt: changeLog.ChangedAt,
			Reason:    changeLog.Reason,
		}
	}

	response := &PartnerDetailResponse{
		ID:               partner.ID,
		Login:            partner.Login,
		BrandName:        partner.BrandName,
		Domain:           partner.Domain,
		Email:            partner.Email,
		Phone:            partner.Phone,
		TelegramLink:     partner.TelegramLink,
		WhatsappLink:     partner.WhatsappLink,
		WildberriesLink:  partner.WildberriesLink,
		OzonLink:         partner.OzonLink,
		BrandColors:      partner.BrandColors,
		Status:           partner.Status,
		CreatedAt:        partner.CreatedAt,
		LastLogin:        partner.LastLogin,
		TotalCoupons:     totalCoupons,
		ActivatedCoupons: activatedCoupons,
		UnusedCoupons:    unusedCoupons,
		LastActivity:     lastActivity,
		ProfileChanges:   profileChanges,
	}

	return response, nil
}

// CreatePartner creates new partner with unique login/domain validation and processed contact links
func (s *AdminService) CreatePartner(req partner.CreatePartnerRequest) (*partner.Partner, error) {
	if _, err := s.deps.PartnerRepository.GetByLogin(context.Background(), req.Login); err == nil {
		return nil, fmt.Errorf("partner already exists: %w", err)
	}

	if _, err := s.deps.PartnerRepository.GetByDomain(context.Background(), req.Domain); err == nil {
		return nil, fmt.Errorf("partner already exists: %w", err)
	}

	partnerCode, err := s.deps.PartnerRepository.GetNextPartnerCode(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to generate partner code: %w", err)
	}

	hashedPassword, err := bcrypt.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	status := req.Status
	if status == "" {
		status = "active"
	}
	telegramLink := validateData.ProcessTeleramLink(req.Telegram)
	whatsappLink := validateData.ProcessWhatsappLink(req.Whatsapp)

	newPartner := &partner.Partner{
		PartnerCode:     partnerCode,
		Login:           req.Login,
		Password:        hashedPassword,
		Domain:          req.Domain,
		BrandName:       req.BrandName,
		LogoURL:         req.LogoURL,
		OzonLink:        req.OzonLink,
		WildberriesLink: req.WildberriesLink,
		Email:           req.Email,
		Address:         req.Address,
		Phone:           req.Phone,
		Telegram:        req.Telegram,
		Whatsapp:        whatsappLink,
		TelegramLink:    telegramLink,
		WhatsappLink:    whatsappLink,
		BrandColors:     req.BrandColors,
		AllowSales:      req.AllowSales,
		AllowPurchases:  req.AllowPurchases,
		Status:          status,
	}

	if err := s.deps.PartnerRepository.Create(context.Background(), newPartner); err != nil {
		return nil, fmt.Errorf("failed to create partner: %w", err)
	}

	if err := s.deps.PartnerRepository.InitializeArticleGrid(context.Background(), newPartner.ID); err != nil {
	} else {
	}

	// Trigger nginx configuration update immediately for new domain
	go func() {
		if err := s.triggerNginxConfigUpdate(); err != nil {
			// Log error but don't fail the partner creation
			fmt.Printf("Warning: Failed to update nginx config after partner creation: %v\n", err)
		}
	}()

	// Trigger CI/CD pipeline for domain update if GitLab client is available
	if s.deps.GitLabClient != nil {
		go func() {
			_, err := s.deps.GitLabClient.TriggerDomainUpdateWithDetails("main", "add", "", newPartner.Domain)
			if err != nil {
				// Log error but don't fail the partner creation
				return
			}
		}()
	}

	return newPartner, nil
}

// UpdatePartnerWithHistory updates partner data and logs all changes for audit purposes
func (s *AdminService) UpdatePartnerWithHistory(partnerID uuid.UUID, req partner.UpdatePartnerRequest, adminLogin string, reason string) (*partner.Partner, error) {

	oldPartner, err := s.deps.PartnerRepository.GetByID(context.Background(), partnerID)
	if err != nil {
		return nil, fmt.Errorf("partner not found: %w", err)
	}

	// Save old domain for CI/CD trigger check
	oldDomain := oldPartner.Domain

	oldValues := map[string]string{
		"brand_name":       oldPartner.BrandName,
		"domain":           oldPartner.Domain,
		"email":            oldPartner.Email,
		"phone":            oldPartner.Phone,
		"telegram":         oldPartner.Telegram,
		"whatsapp":         oldPartner.Whatsapp,
		"telegram_link":    oldPartner.TelegramLink,
		"whatsapp_link":    oldPartner.WhatsappLink,
		"ozon_link":        oldPartner.OzonLink,
		"wildberries_link": oldPartner.WildberriesLink,
		"address":          oldPartner.Address,
		"logo_url":         oldPartner.LogoURL,
		"allow_sales":      fmt.Sprintf("%t", oldPartner.AllowSales),
	}

	updatePartnerData.UpdatePartnerData(oldPartner, &req)

	if err := s.deps.PartnerRepository.Update(context.Background(), oldPartner); err != nil {
		return nil, fmt.Errorf("failed to update partner: %w", err)
	}

	newValues := map[string]string{
		"brand_name":       oldPartner.BrandName,
		"domain":           oldPartner.Domain,
		"email":            oldPartner.Email,
		"phone":            oldPartner.Phone,
		"telegram":         oldPartner.Telegram,
		"whatsapp":         oldPartner.Whatsapp,
		"telegram_link":    oldPartner.TelegramLink,
		"whatsapp_link":    oldPartner.WhatsappLink,
		"ozon_link":        oldPartner.OzonLink,
		"wildberries_link": oldPartner.WildberriesLink,
		"address":          oldPartner.Address,
		"logo_url":         oldPartner.LogoURL,
		"allow_sales":      fmt.Sprintf("%t", oldPartner.AllowSales),
	}

	for field, oldValue := range oldValues {
		newValue := newValues[field]
		if oldValue != newValue {
			changeLog := &ProfileChangeLog{
				PartnerID: partnerID,
				Field:     field,
				OldValue:  oldValue,
				NewValue:  newValue,
				ChangedBy: adminLogin,
				Reason:    reason,
			}

			if err := s.deps.AdminRepository.CreateProfileChangeLog(changeLog); err != nil {
				return nil, fmt.Errorf("failed to create changelog for field %s: %w", field, err)
			}
		}
	}

	// Trigger nginx configuration update immediately if domain was changed
	if req.Domain != nil && *req.Domain != oldDomain {
		go func() {
			if err := s.triggerNginxConfigUpdate(); err != nil {
				// Log error but don't fail the partner update
				fmt.Printf("Warning: Failed to update nginx config after partner domain change: %v\n", err)
			}
		}()
	}

	// Trigger CI/CD pipeline for domain update only if domain was actually changed
	if req.Domain != nil && *req.Domain != oldDomain {
		if s.deps.GitLabClient != nil {
			go func() {
				_, err := s.deps.GitLabClient.TriggerDomainUpdateWithDetails("main", "update", oldDomain, *req.Domain)
				if err != nil {
					// Log error but don't fail the partner update
					return
				}
			}()
		}
	}

	return oldPartner, nil
}

// GetPartner retrieves partner by ID
func (s *AdminService) GetPartner(id uuid.UUID) (*partner.Partner, error) {
	p, err := s.deps.PartnerRepository.GetByID(context.Background(), id)
	if err != nil {
		return nil, fmt.Errorf("partner not found: %w", err)
	}
	return p, nil
}

// BlockPartner blocks partner and all associated coupons
func (s *AdminService) BlockPartner(id uuid.UUID) error {
	p, err := s.deps.PartnerRepository.GetByID(context.Background(), id)
	if err != nil {
		return fmt.Errorf("partner not found: %w", err)
	}

	p.Status = "blocked"
	if err := s.deps.PartnerRepository.Update(context.Background(), p); err != nil {
		return fmt.Errorf("failed to block partner: %w", err)
	}

	if err := s.deps.CouponRepository.UpdateStatusByPartnerID(context.Background(), id, true); err != nil {
		return fmt.Errorf("failed to block coupons: %w", err)
	}

	return nil
}

// UnblockPartner unblocks partner and all associated coupons
func (s *AdminService) UnblockPartner(id uuid.UUID) error {
	p, err := s.deps.PartnerRepository.GetByID(context.Background(), id)
	if err != nil {
		return fmt.Errorf("partner not found: %w", err)
	}

	p.Status = "active"
	if err := s.deps.PartnerRepository.Update(context.Background(), p); err != nil {
		return fmt.Errorf("failed to unblock partner: %w", err)
	}

	if err := s.deps.CouponRepository.UpdateStatusByPartnerID(context.Background(), id, false); err != nil {
		return fmt.Errorf("failed to unblock coupons: %w", err)
	}

	return nil
}

// DeletePartner permanently deletes partner and all associated coupons, articles, and images
func (s *AdminService) DeletePartner(id uuid.UUID) error {
	// Get partner info before deletion to trigger cleanup
	partner, err := s.deps.PartnerRepository.GetByID(context.Background(), id)
	if err != nil {
		return fmt.Errorf("partner not found: %w", err)
	}

	// Delete partner from database (with cascade deletion of coupons, articles, images)
	if err := s.deps.PartnerRepository.DeleteWithCoupons(context.Background(), id); err != nil {
		return fmt.Errorf("failed to delete partner: %w", err)
	}

	// Trigger nginx configuration update immediately for domain cleanup
	if partner.Domain != "" {
		go func() {
			if err := s.triggerNginxConfigUpdate(); err != nil {
				// Log error but don't fail the partner deletion
				fmt.Printf("Warning: Failed to update nginx config after partner deletion: %v\n", err)
			}
		}()
	}

	// Trigger CI/CD pipeline for domain cleanup if GitLab client is available
	if s.deps.GitLabClient != nil && partner.Domain != "" {
		go func() {
			_, err := s.deps.GitLabClient.TriggerDomainUpdateWithDetails("main", "delete", partner.Domain, "")
			if err != nil {
				// Log error but don't fail the partner deletion
				return
			}
		}()
	}

	return nil
}

// GetPartnerStatistics retrieves coupon statistics for specific partner
func (s *AdminService) GetPartnerStatistics(partnerID uuid.UUID) (map[string]any, error) {
	if _, err := s.deps.PartnerRepository.GetByID(context.Background(), partnerID); err != nil {
		return nil, fmt.Errorf("partner not found: %w", err)
	}

	stats, err := s.deps.CouponRepository.GetStatistics(context.Background(), &partnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get coupon statistics: %w", err)
	}

	return map[string]any{
		"coupon_statistics": stats,
	}, nil
}

// GetCoupons retrieves coupons with optional filtering
func (s *AdminService) GetCoupons(code, status, size, style string, partnerID *uuid.UUID) ([]*coupon.Coupon, error) {
	coupons, err := s.deps.CouponRepository.Search(context.Background(), code, status, size, style, partnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get coupons: %w", err)
	}
	return coupons, nil
}

// GetCouponsPaginated retrieves coupons with pagination and optional filtering
func (s *AdminService) GetCouponsPaginated(code, status, size, style string, partnerID *uuid.UUID, page, limit int) ([]*coupon.Coupon, int64, error) {
	coupons, total, err := s.deps.CouponRepository.SearchWithPagination(context.Background(), code, status, size, style, partnerID, page, limit, nil, nil, nil, nil, "", "")
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get coupons: %w", err)
	}
	return coupons, int64(total), nil
}

// CreateCoupons creates multiple coupons with unique codes for specified partner or own partner
func (s *AdminService) CreateCoupons(req coupon.CreateCouponRequest) ([]*coupon.Coupon, error) {
	partnerCode := "0000"
	effectivePartnerID := req.PartnerID

	if req.PartnerID != uuid.Nil {
		partner, err := s.deps.PartnerRepository.GetByID(context.Background(), req.PartnerID)
		if err != nil {
			return nil, fmt.Errorf("partner not found: %w", err)
		}
		partnerCode = partner.PartnerCode
	} else {
		own, err := s.deps.PartnerRepository.GetByPartnerCode(context.Background(), "0000")
		if err != nil {
			return nil, fmt.Errorf("own partner (0000) not found: %w", err)
		}
		effectivePartnerID = own.ID
	}

	coupons := make([]*coupon.Coupon, 0, req.Count)

	for i := 0; i < req.Count; i++ {
		code, err := randomCouponCode.GenerateUniqueCouponCode(partnerCode, s.deps.CouponRepository)
		if err != nil {
			return nil, fmt.Errorf("failed to create coupons: %w", err)
		}

		coupons = append(coupons, &coupon.Coupon{
			Code:      code,
			PartnerID: effectivePartnerID,
			Size:      string(req.Size),
			Style:     string(req.Style),
			Status:    string(coupon.StatusNew),
		})
	}

	if err := s.deps.CouponRepository.CreateBatch(context.Background(), coupons); err != nil {
		return nil, fmt.Errorf("failed to create coupons: %w", err)
	}

	return coupons, nil
}

// ExportCoupons exports coupons data to CSV format with filtering options
func (s *AdminService) ExportCoupons(code, status, size, style string, partnerID *uuid.UUID) (string, error) {
	coupons, err := s.deps.CouponRepository.Search(context.Background(), code, status, size, style, partnerID)
	if err != nil {
		return "", fmt.Errorf("failed to export coupons: %w", err)
	}

	if len(coupons) == 0 {
		return "", fmt.Errorf("no coupons found for export")
	}

	var csvBuilder strings.Builder
	csvBuilder.WriteString("Code,Partner ID,Size,Style,Status,Is Purchased,Purchase Email,Created At,Used At\n")

	for _, c := range coupons {
		usedAt := ""
		if c.UsedAt != nil {
			usedAt = c.UsedAt.Format("2006-01-02 15:04:05")
		}

		purchaseEmail := ""
		if c.PurchaseEmail != nil {
			purchaseEmail = *c.PurchaseEmail
		}

		csvBuilder.WriteString(fmt.Sprintf("%s,%s,%s,%s,%s,%t,%s,%s,%s\n",
			c.Code,
			c.PartnerID.String(),
			c.Size,
			c.Style,
			c.Status,
			c.IsPurchased,
			purchaseEmail,
			c.CreatedAt.Format("2006-01-02 15:04:05"),
			usedAt,
		))
	}

	return csvBuilder.String(), nil
}

// ExportPartnerCoupons exports all coupons for specific partner to CSV format
func (s *AdminService) ExportPartnerCoupons(partnerID uuid.UUID) (string, error) {
	return s.ExportCoupons("", "", "", "", &partnerID)
}

// BatchDeleteCoupons deletes multiple coupons and returns count of deleted items
func (s *AdminService) BatchDeleteCoupons(couponIDs []uuid.UUID) (int64, error) {
	if len(couponIDs) == 0 {
		return 0, fmt.Errorf("bad request")
	}

	deleted, err := s.deps.CouponRepository.BatchDelete(context.Background(), couponIDs)
	if err != nil {
		return 0, fmt.Errorf("failed to delete coupons: %w", err)
	}

	return deleted, nil
}

// GetCoupon retrieves coupon by ID
func (s *AdminService) GetCoupon(id uuid.UUID) (*coupon.Coupon, error) {
	c, err := s.deps.CouponRepository.GetByID(context.Background(), id)
	if err != nil {
		return nil, fmt.Errorf("coupon not found: %w", err)
	}
	return c, nil
}

// ResetCoupon resets coupon to initial state and cleans up associated S3 files
func (s *AdminService) ResetCoupon(id uuid.UUID) error {
	var errs []error

	if s.deps.S3Client != nil {
		image, err := s.deps.ImageRepository.GetByCouponID(context.Background(), id)
		if err != nil {
			errs = append(errs, fmt.Errorf("get image for coupon %s: %w", id.String(), err))
		} else if image != nil && image.SchemaS3Key != nil && *image.SchemaS3Key != "" {
			if err := s.deps.S3Client.DeleteFile(context.Background(), *image.SchemaS3Key); err != nil {
				errs = append(errs, fmt.Errorf("delete S3 file %s: %w", *image.SchemaS3Key, err))
			}
		}
	}

	if err := s.deps.CouponRepository.Reset(context.Background(), id); err != nil {
		if len(errs) > 0 {
			return fmt.Errorf("failed to reset coupon %s (additional errors: %v): %w", id.String(), errs, err)
		}
		return fmt.Errorf("failed to reset coupon %s: %w", id.String(), err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("coupon %s reset successfully, but with warnings: %v", id.String(), errs)
	}

	return nil
}

// DeleteCoupon permanently deletes coupon and associated S3 files
func (s *AdminService) DeleteCoupon(id uuid.UUID) error {
	var s3Err error
	var s3Key string

	if s.deps.S3Client != nil {
		image, err := s.deps.ImageRepository.GetByCouponID(context.Background(), id)
		if err == nil && image != nil && image.SchemaS3Key != nil && *image.SchemaS3Key != "" {
			s3Key = *image.SchemaS3Key
			s3Err = s.deps.S3Client.DeleteFile(context.Background(), s3Key)
		}
	}

	if err := s.deps.CouponRepository.Delete(context.Background(), id); err != nil {
		if s3Err != nil {
			return fmt.Errorf("failed to delete coupon %s (S3 error: %v): %w", id.String(), s3Err, err)
		}
		return fmt.Errorf("failed to delete coupon %s: %w", id.String(), err)
	}

	if s3Err != nil {
		return fmt.Errorf("coupon %s deleted successfully, but failed to delete S3 file %s: %w", id.String(), s3Key, s3Err)
	}

	return nil
}

// GetStatistics retrieves overall coupon statistics
func (s *AdminService) GetStatistics() (map[string]any, error) {
	stats, err := s.deps.CouponRepository.GetStatistics(context.Background(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get statistics: %w", err)
	}

	return map[string]any{
		"coupon_statistics": stats,
	}, nil
}

// GetPartnersStatistics retrieves coupon statistics for all partners
func (s *AdminService) GetPartnersStatistics() (map[string]any, error) {
	partners, err := s.deps.PartnerRepository.GetAll(context.Background(), "created_at", "desc")
	if err != nil {
		return nil, fmt.Errorf("failed to get partners: %w", err)
	}

	result := make(map[string]any)
	partnerStats := make([]map[string]any, 0, len(partners))

	for _, p := range partners {
		stats, err := s.deps.CouponRepository.GetStatistics(context.Background(), &p.ID)
		if err != nil {
			continue
		}

		partnerStats = append(partnerStats, map[string]any{
			"partner_id":   p.ID,
			"brand_name":   p.BrandName,
			"partner_code": p.PartnerCode,
			"status":       p.Status,
			"statistics":   stats,
		})
	}

	result["partners"] = partnerStats
	return result, nil
}

// GetSystemStatistics aggregates comprehensive system statistics including coupons, partners, and image processing
func (s *AdminService) GetSystemStatistics() (map[string]any, error) {
	result := make(map[string]any)

	couponStats, err := s.deps.CouponRepository.GetStatistics(context.Background(), nil)
	if err == nil {
		result["coupons"] = couponStats
	}

	partners, err := s.deps.PartnerRepository.GetAll(context.Background(), "created_at", "desc")
	if err == nil {
		activePartners := 0
		for _, p := range partners {
			if p.Status == "active" {
				activePartners++
			}
		}
		result["partners"] = map[string]any{
			"total":  len(partners),
			"active": activePartners,
		}
	}

	images, err := s.deps.ImageRepository.GetAll(context.Background())
	if err == nil {
		processingImages := 0
		completedImages := 0
		failedImages := 0

		for _, img := range images {
			switch img.Status {
			case "processing":
				processingImages++
			case "completed":
				completedImages++
			case "failed":
				failedImages++
			}
		}

		result["image_processing"] = map[string]any{
			"total":      len(images),
			"processing": processingImages,
			"completed":  completedImages,
			"failed":     failedImages,
		}
	}

	return result, nil
}

// GetAnalytics provides detailed analytics data including daily creation trends and distribution statistics
func (s *AdminService) GetAnalytics() (map[string]any, error) {
	result := make(map[string]any)

	allCoupons, err := s.deps.CouponRepository.GetAll(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get coupons: %w", err)
	}

	dailyStats := make(map[string]int)
	sizeStats := make(map[string]int)
	styleStats := make(map[string]int)

	for _, c := range allCoupons {
		day := c.CreatedAt.Format("2006-01-02")
		dailyStats[day]++

		sizeStats[c.Size]++

		styleStats[c.Style]++
	}

	result["daily_creation"] = dailyStats
	result["size_distribution"] = sizeStats
	result["style_distribution"] = styleStats

	return result, nil
}

// GetDashboardStatistics returns dashboard data (alias for GetDashboardData)
func (s *AdminService) GetDashboardStatistics() (map[string]any, error) {
	return s.GetDashboardData()
}

// GetAllImages retrieves all image processing tasks
func (s *AdminService) GetAllImages() ([]*image.Image, error) {
	images, err := s.deps.ImageRepository.GetAll(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get images: %w", err)
	}
	return images, nil
}

// GetImageDetails retrieves image processing task by ID
func (s *AdminService) GetImageDetails(id uuid.UUID) (*image.Image, error) {
	task, err := s.deps.ImageRepository.GetByID(context.Background(), id)
	if err != nil {
		return nil, fmt.Errorf("image not found: %w", err)
	}
	return task, nil
}

// DeleteImageTask permanently deletes image processing task
func (s *AdminService) DeleteImageTask(id uuid.UUID) error {
	if err := s.deps.ImageRepository.Delete(context.Background(), id); err != nil {
		return fmt.Errorf("failed to delete image: %w", err)
	}
	return nil
}

// RetryImageTask resets image task to queued status for reprocessing
func (s *AdminService) RetryImageTask(id uuid.UUID) error {
	task, err := s.deps.ImageRepository.GetByID(context.Background(), id)
	if err != nil {
		return fmt.Errorf("image not found: %w", err)
	}

	task.Status = "queued"
	task.ErrorMessage = nil
	task.StartedAt = nil
	task.CompletedAt = nil

	if err := s.deps.ImageRepository.Update(context.Background(), task); err != nil {
		return fmt.Errorf("failed to update image: %w", err)
	}

	return nil
}

// BatchResetCoupons resets multiple coupons and cleans up associated S3 files
func (s *AdminService) BatchResetCoupons(couponIDs []string) (*coupon.BatchResetResponse, error) {
	if s.deps.S3Client != nil {
		for _, idStr := range couponIDs {
			if id, err := uuid.Parse(idStr); err == nil {
				if image, err := s.deps.ImageRepository.GetByCouponID(context.Background(), id); err == nil {
					if image.SchemaS3Key != nil && *image.SchemaS3Key != "" {
						s.deps.S3Client.DeleteFile(context.Background(), *image.SchemaS3Key)
					}
				}
			}
		}
	}

	couponService := coupon.NewCouponService(&coupon.CouponServiceDeps{
		CouponRepository: s.deps.CouponRepository.(coupon.CouponRepositoryInterface),
		RedisClient:      s.deps.RedisClient,
		S3Client:         s.deps.S3Client,
	})

	response, err := couponService.BatchResetCoupons(couponIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to batch reset coupons: %w", err)
	}

	return response, nil
}

// PreviewBatchDelete shows preview of coupons that will be deleted and cleans up S3 files
func (s *AdminService) PreviewBatchDelete(couponIDs []string) (*coupon.BatchDeletePreviewResponse, error) {
	if s.deps.S3Client != nil {
		for _, idStr := range couponIDs {
			if id, err := uuid.Parse(idStr); err == nil {
				if image, err := s.deps.ImageRepository.GetByCouponID(context.Background(), id); err == nil {
					if image.SchemaS3Key != nil && *image.SchemaS3Key != "" {
						s.deps.S3Client.DeleteFile(context.Background(), *image.SchemaS3Key)
					}
				}
			}
		}
	}

	couponService := coupon.NewCouponService(&coupon.CouponServiceDeps{
		CouponRepository: s.deps.CouponRepository.(coupon.CouponRepositoryInterface),
		RedisClient:      s.deps.RedisClient,
		S3Client:         s.deps.S3Client,
	})

	response, err := couponService.PreviewBatchDelete(couponIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get batch delete preview: %w", err)
	}

	return response, nil
}

// ExecuteBatchDelete permanently deletes multiple coupons and associated S3 files
func (s *AdminService) ExecuteBatchDelete(req coupon.BatchDeleteConfirmRequest) (*coupon.BatchDeleteResponse, error) {
	if s.deps.S3Client != nil {
		for _, idStr := range req.CouponIDs {
			if id, err := uuid.Parse(idStr); err == nil {
				if image, err := s.deps.ImageRepository.GetByCouponID(context.Background(), id); err == nil {
					if image.SchemaS3Key != nil && *image.SchemaS3Key != "" {
						s.deps.S3Client.DeleteFile(context.Background(), *image.SchemaS3Key)
					}
				}
			}
		}
	}

	couponService := coupon.NewCouponService(&coupon.CouponServiceDeps{
		CouponRepository: s.deps.CouponRepository.(coupon.CouponRepositoryInterface),
		RedisClient:      s.deps.RedisClient,
		S3Client:         s.deps.S3Client,
	})

	response, err := couponService.ExecuteBatchDelete(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute batch delete: %w", err)
	}

	return response, nil
}

// ExportCouponsAdvanced exports coupons with advanced options and multiple format support
func (s *AdminService) ExportCouponsAdvanced(options coupon.ExportOptionsRequest) ([]byte, string, string, error) {
	couponService := coupon.NewCouponService(&coupon.CouponServiceDeps{
		CouponRepository: s.deps.CouponRepository.(coupon.CouponRepositoryInterface),
		RedisClient:      s.deps.RedisClient,
		S3Client:         s.deps.S3Client,
	})

	content, filename, contentType, err := couponService.ExportCouponsAdvanced(options)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to export coupons advanced: %w", err)
	}

	return content, filename, contentType, nil
}

// DownloadCouponMaterials downloads ZIP archive with coupon materials from S3
func (s *AdminService) DownloadCouponMaterials(id uuid.UUID) ([]byte, string, error) {
	coupon, err := s.deps.CouponRepository.GetByID(context.Background(), id)
	if err != nil {
		return nil, "", fmt.Errorf("coupon not found: %w", err)
	}

	if coupon.Status != "used" && coupon.Status != "completed" {
		return nil, "", fmt.Errorf("coupon must be used or completed to download materials")
	}

	imageRecord, err := s.deps.ImageRepository.GetByCouponID(context.Background(), id)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get image for coupon: %w", err)
	}

	if imageRecord.SchemaS3Key == nil {
		return nil, "", fmt.Errorf("no schema ZIP archive found for coupon")
	}

	reader, err := s.deps.S3Client.DownloadFile(context.Background(), *imageRecord.SchemaS3Key)
	if err != nil {
		return nil, "", fmt.Errorf("failed to download ZIP archive: %w", err)
	}
	defer reader.Close()

	archiveData := make([]byte, 0)
	// Read ZIP archive data
	buffer := make([]byte, 32*1024)
	for {
		n, err := reader.Read(buffer)
		if n > 0 {
			archiveData = append(archiveData, buffer[:n]...)
		}
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return nil, "", fmt.Errorf("failed to read ZIP archive: %w", err)
		}
	}

	filename := fmt.Sprintf("coupon_%s_materials.zip", coupon.Code)

	return archiveData, filename, nil
}

// BatchDownloadMaterials downloads materials for multiple coupons and creates a single ZIP archive
func (s *AdminService) BatchDownloadMaterials(couponIDs []uuid.UUID) ([]byte, string, error) {

	if len(couponIDs) == 0 {
		return nil, "", fmt.Errorf("no coupon IDs provided")
	}

	if len(couponIDs) > 100 {
		return nil, "", fmt.Errorf("too many coupons (maximum 100)")
	}

	validCoupons := make([]*coupon.Coupon, 0)
	couponArchives := make(map[string][]byte)

	for _, couponID := range couponIDs {
		couponData, err := s.deps.CouponRepository.GetByID(context.Background(), couponID)
		if err != nil {
			continue
		}

		if couponData.Status != "used" && couponData.Status != "completed" {
			continue
		}

		archiveData, _, err := s.DownloadCouponMaterials(couponID)
		if err != nil {
			continue
		}

		validCoupons = append(validCoupons, couponData)
		couponArchives[couponData.Code] = archiveData
	}

	if len(validCoupons) == 0 {
		return nil, "", fmt.Errorf("no valid coupons found for download")
	}

	return s.createBatchArchive(couponArchives)
}

// createBatchArchive creates a ZIP archive containing multiple coupon material files
func (s *AdminService) createBatchArchive(couponArchives map[string][]byte) ([]byte, string, error) {
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	for couponCode, archiveData := range couponArchives {
		fileName := fmt.Sprintf("coupon_%s_materials.zip", couponCode)
		fileWriter, err := zipWriter.Create(fileName)
		if err != nil {
			zipWriter.Close()
			return nil, "", fmt.Errorf("failed to create file %s in batch archive: %w", fileName, err)
		}

		_, err = fileWriter.Write(archiveData)
		if err != nil {
			zipWriter.Close()
			return nil, "", fmt.Errorf("failed to write data for %s: %w", fileName, err)
		}
	}

	err := zipWriter.Close()
	if err != nil {
		return nil, "", fmt.Errorf("failed to close batch archive: %w", err)
	}

	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("batch_coupon_materials_%s.zip", timestamp)

	return buf.Bytes(), filename, nil
}

// GetActivePartnersWithDomains retrieves all active partners that have domain names configured
func (s *AdminService) GetActivePartnersWithDomains() ([]partner.Partner, error) {
	partners, err := s.deps.PartnerRepository.GetAll(context.Background(), "created_at", "desc")
	if err != nil {
		return nil, fmt.Errorf("failed to get partners: %w", err)
	}

	activePartners := make([]partner.Partner, 0)
	for _, p := range partners {
		if p.Status == "active" && p.Domain != "" {
			activePartners = append(activePartners, *p)
		}
	}

	return activePartners, nil
}

// GetAdmins retrieves all admin users
func (s *AdminService) GetAdmins() ([]*Admin, error) {
	admins, err := s.deps.AdminRepository.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get admins: %w", err)
	}

	return admins, nil
}

// DeleteAdmin permanently deletes admin user
func (s *AdminService) DeleteAdmin(id uuid.UUID) error {
	admin, err := s.deps.AdminRepository.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get admin: %w", err)
	}

	if admin == nil {
		return fmt.Errorf("admin not found")
	}

	if err := s.deps.AdminRepository.Delete(id); err != nil {
		return fmt.Errorf("failed to delete admin: %w", err)
	}

	return nil
}

// UpdateAdminPassword updates admin password with bcrypt hashing
func (s *AdminService) UpdateAdminPassword(id uuid.UUID, newPassword string) error {
	admin, err := s.deps.AdminRepository.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get admin: %w", err)
	}

	if admin == nil {
		return fmt.Errorf("admin not found")
	}

	hashedPassword, err := bcrypt.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	if err := s.deps.AdminRepository.UpdatePassword(id, hashedPassword); err != nil {
		return fmt.Errorf("failed to update admin password: %w", err)
	}

	return nil
}

// UpdateAdminEmail updates admin email with uniqueness validation
func (s *AdminService) UpdateAdminEmail(id uuid.UUID, newEmail string) error {
	admin, err := s.deps.AdminRepository.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get admin: %w", err)
	}

	if admin == nil {
		return fmt.Errorf("admin not found")
	}

	if newEmail != "" {
		if existingByEmail, errByEmail := s.deps.AdminRepository.GetByEmail(newEmail); errByEmail == nil && existingByEmail != nil && existingByEmail.ID != id {
			return fmt.Errorf("email already exists")
		}
	}

	if err := s.deps.AdminRepository.UpdateEmail(id, newEmail); err != nil {
		return fmt.Errorf("failed to update admin email: %w", err)
	}

	return nil
}

// triggerNginxConfigUpdate triggers immediate nginx configuration update
// This method calls the nginx config generator script to update configuration
// for newly added partner domains without waiting for CI/CD pipeline
func (s *AdminService) triggerNginxConfigUpdate() error {
	return s.executeNginxConfigGenerator(true)
}

// executeNginxConfigGenerator calls the enhanced nginx config generator script
// reloadNginx parameter controls whether nginx should be reloaded after config generation
func (s *AdminService) executeNginxConfigGenerator(reloadNginx bool) error {
	// Path to the enhanced nginx config generator script
	scriptPath := "../../scripts/generate-nginx-config.sh"

	// Check if script exists
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return fmt.Errorf("nginx config generator script not found: %s", scriptPath)
	}

	// Execute the script using template method
	cmd := exec.Command("/bin/bash", scriptPath, "template")

	// Set environment variables for the script
	reloadValue := "false"
	if reloadNginx {
		reloadValue = "true"
	}

	cmd.Env = append(os.Environ(),
		fmt.Sprintf("RELOAD_NGINX=%s", reloadValue),
	)

	// Execute the command
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to execute nginx config generator: %w, output: %s", err, string(output))
	}

	return nil
}

// DeployNginxConfig generates and deploys nginx configuration
// This method is used by the admin API endpoint for manual nginx config deployment
func (s *AdminService) DeployNginxConfig() error {
	return s.executeNginxConfigGenerator(false) // Don't reload nginx from API call, let CI/CD handle it
}
