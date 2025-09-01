package partner

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type PartnerRepository struct {
	db *bun.DB
}

func NewPartnerRepository(db *bun.DB) *PartnerRepository {
	return &PartnerRepository{db: db}
}

func (r *PartnerRepository) Create(ctx context.Context, partner *Partner) error {
	_, err := r.db.NewInsert().Model(partner).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to create partner: %w", err)
	}
	return nil
}

func (r *PartnerRepository) GetByLogin(ctx context.Context, login string) (*Partner, error) {
	partner := new(Partner)
	err := r.db.NewSelect().Model(partner).Where("login = ?", login).Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find partner by login: %w", err)
	}
	return partner, nil
}

func (r *PartnerRepository) GetByPartnerCode(ctx context.Context, code string) (*Partner, error) {
	partner := new(Partner)
	err := r.db.NewSelect().Model(partner).Where("partner_code = ?", code).Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find partner by partner code: %w", err)
	}
	return partner, nil
}

func (r *PartnerRepository) GetByDomain(ctx context.Context, domain string) (*Partner, error) {
	partner := new(Partner)
	err := r.db.NewSelect().Model(partner).Where("domain = ?", domain).Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find partner by domain: %w", err)
	}
	return partner, nil
}

func (r *PartnerRepository) GetByID(ctx context.Context, id uuid.UUID) (*Partner, error) {
	partner := new(Partner)
	err := r.db.NewSelect().Model(partner).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find partner by id: %w", err)
	}
	return partner, nil
}

func (r *PartnerRepository) GetByEmail(ctx context.Context, email string) (*Partner, error) {
	partner := new(Partner)
	err := r.db.NewSelect().Model(partner).Where("email = ?", email).Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find partner by email: %w", err)
	}
	return partner, nil
}

func (r *PartnerRepository) GetAll(ctx context.Context, sortBy string, order string) ([]*Partner, error) {
	var partners []*Partner
	query := r.db.NewSelect().Model(&partners)

	// Add sorting
	if sortBy != "" {
		// Validate sorting field
		switch sortBy {
		case "created_at", "brand_name", "domain", "email", "status":
			if order == "asc" {
				query = query.Order(sortBy + " ASC")
			} else {
				query = query.Order(sortBy + " DESC")
			}
		default:
			// Default sort by creation date
			query = query.Order("created_at DESC")
		}
	} else {
		// Default sort by creation date
		query = query.Order("created_at DESC")
	}

	err := query.Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find all partners: %w", err)
	}
	return partners, nil
}

func (r *PartnerRepository) GetActivePartners(ctx context.Context) ([]*Partner, error) {
	var partners []*Partner
	err := r.db.NewSelect().Model(&partners).Where("status = ?", "active").Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find active partners: %w", err)
	}
	return partners, nil
}

func (r *PartnerRepository) Update(ctx context.Context, partner *Partner) error {
	_, err := r.db.NewUpdate().Model(partner).WherePK().Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to update partner: %w", err)
	}
	return nil
}

func (r *PartnerRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	_, err := r.db.NewUpdate().Model((*Partner)(nil)).Set("last_login = ?", &now).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to update partner last login: %w", err)
	}
	return nil
}

func (r *PartnerRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	_, err := r.db.NewUpdate().Model((*Partner)(nil)).Set("status = ?", status).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to update partner status: %w", err)
	}
	return nil
}

func (r *PartnerRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.NewDelete().Model((*Partner)(nil)).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete partner: %w", err)
	}
	return nil
}

func (r *PartnerRepository) DeleteWithCoupons(ctx context.Context, id uuid.UUID) error {
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		// Delete partner articles first (cascade deletion)
		_, err := tx.NewDelete().Model((*PartnerArticle)(nil)).Where("partner_id = ?", id).Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to delete partner articles: %w", err)
		}

		// Delete partner - coupons and images will be deleted automatically due to CASCADE
		_, err = tx.NewDelete().Model((*Partner)(nil)).Where("id = ?", id).Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to delete partner: %w", err)
		}

		return nil
	})
}

func (r *PartnerRepository) Search(ctx context.Context, queryStr string, status string, sortBy string, order string) ([]*Partner, error) {
	query := r.db.NewSelect().Model((*Partner)(nil))

	if queryStr != "" {
		query = query.Where("brand_name ILIKE ? OR domain ILIKE ? OR email ILIKE ?", "%"+queryStr+"%", "%"+queryStr+"%", "%"+queryStr+"%")
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Add sorting
	if sortBy != "" {
		// Validate sorting field
		switch sortBy {
		case "created_at", "brand_name", "domain", "email", "status":
			if order == "asc" {
				query = query.Order(sortBy + " ASC")
			} else {
				query = query.Order(sortBy + " DESC")
			}
		default:
			// Default sort by creation date
			query = query.Order("created_at DESC")
		}
	} else {
		// Default sort by creation date
		query = query.Order("created_at DESC")
	}

	var partners []*Partner
	err := query.Scan(ctx, &partners)
	if err != nil {
		return nil, fmt.Errorf("failed to find partners: %w", err)
	}
	return partners, nil
}

func (r *PartnerRepository) GetPartnerCouponsForExport(ctx context.Context, partnerID uuid.UUID, status string) ([]*ExportCouponRequest, error) {
	var coupons []*ExportCouponRequest
	query := `
        SELECT 
			c.code as coupon_code,
			c.partner_id,
			p.status as partner_status,
			c.status as coupon_status,
			c.size,
			c.style,
			p.brand_name,
			p.email,
			c.created_at,
            c.activated_at,
			c.used_at
		FROM coupons c
		JOIN partners p ON c.partner_id = p.id
		WHERE c.partner_id = ?`
	args := []any{partnerID}
	if status != "" {
		query += " AND c.status = ?"
		args = append(args, status)
	}
	query += " ORDER BY c.created_at DESC"
	err := r.db.NewRaw(query, args...).Scan(ctx, &coupons)
	if err != nil {
		return nil, fmt.Errorf("failed to find partner coupons for export: %w", err)
	}
	return coupons, nil
}

func (r *PartnerRepository) GetAllCouponsForExport(ctx context.Context) ([]*ExportCouponRequest, error) {
	var coupons []*ExportCouponRequest
	query := `
        SELECT 
			c.code as coupon_code,
			c.partner_id,
			p.status as partner_status,
			c.status as coupon_status,
			c.size,
			c.style,
			p.brand_name,
			p.email,
			c.created_at,
            c.activated_at,
			c.used_at
		FROM coupons c
		JOIN partners p ON c.partner_id = p.id
		ORDER BY p.id, c.created_at DESC`
	err := r.db.NewRaw(query).Scan(ctx, &coupons)
	if err != nil {
		return nil, fmt.Errorf("failed to get all coupons for export: %w", err)
	}
	return coupons, nil
}

func (r *PartnerRepository) GetCouponsStatistics(ctx context.Context, partnerID uuid.UUID) (map[string]int64, error) {
	stats := make(map[string]int64)

	var count int
	var err error

	count, err = r.db.NewSelect().Model(map[string]any{}).Table("coupons").Where("partner_id = ?", partnerID).Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get coupons statistics: %w", err)
	}
	stats["total"] = int64(count)

	count, err = r.db.NewSelect().Model(map[string]any{}).Table("coupons").Where("partner_id = ? AND status = ?", partnerID, "used").Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get coupons statistics: %w", err)
	}
	stats["activated"] = int64(count)

	count, err = r.db.NewSelect().Model(map[string]any{}).Table("coupons").Where("partner_id = ? AND status = ?", partnerID, "new").Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get coupons statistics: %w", err)
	}
	stats["new"] = int64(count)

	count, err = r.db.NewSelect().Model(map[string]any{}).Table("coupons").Where("partner_id = ? AND is_purchased = ?", partnerID, true).Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get coupons statistics: %w", err)
	}
	stats["purchased"] = int64(count)

	return stats, nil
}

func (r *PartnerRepository) UpdatePassword(ctx context.Context, partnerID uuid.UUID, hashedPassword string) error {
	_, err := r.db.NewUpdate().Model((*Partner)(nil)).Set("password = ?", hashedPassword).Where("id = ?", partnerID).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to update partner password: %w", err)
	}
	return nil
}

func (r *PartnerRepository) UpdateEmail(ctx context.Context, partnerID uuid.UUID, email string) error {
	_, err := r.db.NewUpdate().Model((*Partner)(nil)).Set("email = ?", email).Where("id = ?", partnerID).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to update partner email: %w", err)
	}
	return nil
}

func (r *PartnerRepository) GetNextPartnerCode(ctx context.Context) (string, error) {
	var maxCode string
	err := r.db.NewSelect().Model((*Partner)(nil)).ColumnExpr("COALESCE(MAX(CAST(partner_code AS INTEGER)), 0)").Scan(ctx, &maxCode)
	if err != nil {
		return "", fmt.Errorf("failed to get partner code: %w", err)
	}
	var nextCode int
	if maxCode == "" || maxCode == "0" {
		nextCode = 1
	} else {
		if _, err := fmt.Sscanf(maxCode, "%d", &nextCode); err != nil {
			return "", fmt.Errorf("failed to get partner code: %w", err)
		}
		nextCode++
	}
	return fmt.Sprintf("%04d", nextCode), nil
}

// Statistics methods

func (r *PartnerRepository) CountActive(ctx context.Context) (int64, error) {
	count, err := r.db.NewSelect().Model((*Partner)(nil)).Where("status = ?", "active").Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count active partners: %w", err)
	}
	return int64(count), nil
}

func (r *PartnerRepository) CountTotal(ctx context.Context) (int64, error) {
	count, err := r.db.NewSelect().Model((*Partner)(nil)).Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count total partners: %w", err)
	}
	return int64(count), nil
}

func (r *PartnerRepository) GetTopByActivity(ctx context.Context, limit int) ([]*Partner, error) {
	var partners []*Partner
	err := r.db.NewSelect().
		Model(&partners).
		Where("status = ?", "active").
		Order("created_at DESC").
		Limit(limit).
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get top partners: %w", err)
	}
	return partners, nil
}

// InitializeArticleGrid creates empty article grid for partner
func (r *PartnerRepository) InitializeArticleGrid(ctx context.Context, partnerID uuid.UUID) error {
	var articles []PartnerArticle

	// Create 48 cells (4 styles × 6 sizes × 2 marketplaces)
	for _, marketplace := range Marketplaces {
		for _, style := range AvailableStyles {
			for _, size := range AvailableSizes {
				articles = append(articles, PartnerArticle{
					PartnerID:   partnerID,
					Size:        size,
					Style:       style,
					Marketplace: marketplace,
					SKU:         "", // empty article
					IsActive:    true,
				})
			}
		}
	}

	// Insert all cells
	_, err := r.db.NewInsert().Model(&articles).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize article grid: %w", err)
	}

	return nil
}

// GetArticleGrid gets article grid for partner
func (r *PartnerRepository) GetArticleGrid(ctx context.Context, partnerID uuid.UUID) (map[string]map[string]map[string]string, error) {
	var articles []PartnerArticle

	err := r.db.NewSelect().
		Model(&articles).
		Where("partner_id = ? AND is_active = true", partnerID).
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get article grid: %w", err)
	}

	// Structure: marketplace -> style -> size -> sku
	grid := make(map[string]map[string]map[string]string)

	for _, article := range articles {
		if grid[article.Marketplace] == nil {
			grid[article.Marketplace] = make(map[string]map[string]string)
		}
		if grid[article.Marketplace][article.Style] == nil {
			grid[article.Marketplace][article.Style] = make(map[string]string)
		}
		grid[article.Marketplace][article.Style][article.Size] = article.SKU
	}

	return grid, nil
}

// UpdateArticleSKU updates article in grid cell
func (r *PartnerRepository) UpdateArticleSKU(ctx context.Context, partnerID uuid.UUID, size, style, marketplace, sku string) error {
	_, err := r.db.NewUpdate().
		Model((*PartnerArticle)(nil)).
		Set("sku = ?, updated_at = CURRENT_TIMESTAMP", sku).
		Where("partner_id = ? AND size = ? AND style = ? AND marketplace = ?",
			partnerID, size, style, marketplace).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to update article SKU: %w", err)
	}

	return nil
}

// GetArticleBySizeStyle gets article by size, style and marketplace
func (r *PartnerRepository) GetArticleBySizeStyle(ctx context.Context, partnerID uuid.UUID, size, style, marketplace string) (*PartnerArticle, error) {
	article := new(PartnerArticle)

	err := r.db.NewSelect().
		Model(article).
		Where("partner_id = ? AND size = ? AND style = ? AND marketplace = ? AND is_active = true",
			partnerID, size, style, marketplace).
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get article: %w", err)
	}

	return article, nil
}

// DeleteArticleGrid deletes entire article grid for partner
func (r *PartnerRepository) DeleteArticleGrid(ctx context.Context, partnerID uuid.UUID) error {
	_, err := r.db.NewDelete().
		Model((*PartnerArticle)(nil)).
		Where("partner_id = ?", partnerID).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to delete article grid: %w", err)
	}

	return nil
}

// GetAllArticlesByPartner gets all partner articles
func (r *PartnerRepository) GetAllArticlesByPartner(ctx context.Context, partnerID uuid.UUID) ([]*PartnerArticle, error) {
	var articles []*PartnerArticle

	err := r.db.NewSelect().
		Model(&articles).
		Where("partner_id = ? AND is_active = true", partnerID).
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get partner articles: %w", err)
	}

	return articles, nil
}
