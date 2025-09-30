package coupon

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type CouponRepository struct {
	db *bun.DB
}

func NewCouponRepository(db *bun.DB) *CouponRepository {
	return &CouponRepository{db: db}
}

func (r *CouponRepository) Create(ctx context.Context, coupon *Coupon) error {
	_, err := r.db.NewInsert().Model(coupon).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to create coupon: %w", err)
	}
	return nil
}

func (r *CouponRepository) CreateBatch(ctx context.Context, coupons []*Coupon) error {
	_, err := r.db.NewInsert().Model(&coupons).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to create coupon: %w", err)
	}
	return nil
}

func (r *CouponRepository) GetByCode(ctx context.Context, code string) (*Coupon, error) {
	coupon := new(Coupon)
	err := r.db.NewSelect().Model(coupon).Where("code = ?", code).Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("not found")
		}
		return nil, fmt.Errorf("failed to find coupon by code: %w", err)
	}
	return coupon, nil
}

func (r *CouponRepository) GetByID(ctx context.Context, id uuid.UUID) (*Coupon, error) {
	coupon := new(Coupon)
	err := r.db.NewSelect().Model(coupon).Where("id = ?", id).Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("not found")
		}
		return nil, fmt.Errorf("failed to find coupon by ID: %w", err)
	}
	return coupon, nil
}

func (r *CouponRepository) GetByPartnerID(ctx context.Context, partnerID uuid.UUID) ([]*Coupon, error) {
	var coupons []*Coupon
	err := r.db.NewSelect().Model(&coupons).Where("partner_id = ?", partnerID).Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find coupons: %w", err)
	}
	return coupons, nil
}

func (r *CouponRepository) GetAll(ctx context.Context) ([]*Coupon, error) {
	var coupons []*Coupon
	err := r.db.NewSelect().Model(&coupons).Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find coupons: %w", err)
	}
	return coupons, nil
}

func (r *CouponRepository) Update(ctx context.Context, coupon *Coupon) error {
	_, err := r.db.NewUpdate().Model(coupon).WherePK().Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to update coupon status: %w", err)
	}
	return nil
}

func (r *CouponRepository) UpdateStatusByPartnerID(ctx context.Context, partnerID uuid.UUID, status bool) error {
	_, err := r.db.NewUpdate().Model((*Coupon)(nil)).Set("is_blocked = ?", status).Where("partner_id = ?", partnerID).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to update coupon status: %w", err)
	}
	return nil
}

// Activates a coupon by setting activation time (simplified activation)
func (r *CouponRepository) ActivateCoupon(ctx context.Context, id uuid.UUID, req ActivateCouponRequest) error {
	now := time.Now()
	query := r.db.NewUpdate().Model((*Coupon)(nil)).
		Set("status = ?", "activated").
		Set("activated_at = ?", &now).
		Where("id = ?", id)

	if req.ZipURL != nil {
		query = query.Set("zip_url = ?", *req.ZipURL)
	}
	if req.PreviewImageURL != nil {
		query = query.Set("preview_image_url = ?", *req.PreviewImageURL)
	}
	if req.SelectedPreviewID != nil {
		query = query.Set("selected_preview_id = ?", *req.SelectedPreviewID)
	}
	if req.StonesCount != nil {
		query = query.Set("stones_count = ?", *req.StonesCount)
	}
	if req.FinalSchemaURL != nil {
		query = query.Set("final_schema_url = ?", *req.FinalSchemaURL)
	}
	if req.PageCount != nil {
		query = query.Set("page_count = ?", *req.PageCount)
	}
	if req.UserEmail != nil {
		query = query.Set("user_email = ?", *req.UserEmail)
	}

	_, err := query.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to activate coupon: %w", err)
	}
	return nil
}

// Records schema email sending information
func (r *CouponRepository) SendSchema(ctx context.Context, id uuid.UUID, email string) error {
	now := time.Now()
	_, err := r.db.NewUpdate().Model((*Coupon)(nil)).
		Set("schema_sent_email = ?", email).
		Set("schema_sent_at = ?", &now).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to send schema: %w", err)
	}
	return nil
}

// Marks a coupon as purchased online
func (r *CouponRepository) MarkAsPurchased(ctx context.Context, id uuid.UUID, purchaseEmail string) error {
	now := time.Now()
	_, err := r.db.NewUpdate().Model((*Coupon)(nil)).
		Set("is_purchased = ?", true).
		Set("purchase_email = ?", purchaseEmail).
		Set("purchased_at = ?", &now).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to mark as purchased: %w", err)
	}
	return nil
}

func (r *CouponRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.NewDelete().Model((*Coupon)(nil)).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete coupon: %w", err)
	}
	return nil
}

func (r *CouponRepository) BatchDelete(ctx context.Context, ids []uuid.UUID) (int64, error) {
	res, err := r.db.NewDelete().Model((*Coupon)(nil)).Where("id IN (?)", bun.In(ids)).Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to delete coupons: %w", err)
	}
	rows, _ := res.RowsAffected()
	return rows, nil
}

func (r *CouponRepository) Search(ctx context.Context, code, status, size, style string, partnerID *uuid.UUID) ([]*Coupon, error) {
	query := r.db.NewSelect().Model((*Coupon)(nil))

	if code != "" {
		query = query.Where("code ILIKE ?", "%"+code+"%")
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if size != "" {
		query = query.Where("size = ?", size)
	}
	if style != "" {
		query = query.Where("style = ?", style)
	}
	if partnerID != nil {
		query = query.Where("partner_id = ?", *partnerID)
	}

	var coupons []*Coupon
	err := query.OrderExpr("created_at DESC").Scan(ctx, &coupons)
	if err != nil {
		return nil, fmt.Errorf("failed to find coupons: %w", err)
	}
	return coupons, nil
}

func (r *CouponRepository) SearchWithPagination(ctx context.Context, code, status, size, style string, partnerID *uuid.UUID, page, limit int, createdFrom, createdTo, usedFrom, usedTo *time.Time, sortBy, sortDir string) ([]*Coupon, int, error) {
	query := r.db.NewSelect().Model((*Coupon)(nil))

	if code != "" {
		// Search by coupon code or partner code
		query = query.Where("(code ILIKE ? OR partner_code ILIKE ?)", "%"+code+"%", "%"+code+"%")
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if size != "" {
		query = query.Where("size = ?", size)
	}
	if style != "" {
		query = query.Where("style = ?", style)
	}
	if partnerID != nil {
		query = query.Where("partner_id = ?", *partnerID)
	}
	if createdFrom != nil {
		query = query.Where("created_at >= ?", *createdFrom)
	}
	if createdTo != nil {
		query = query.Where("created_at <= ?", *createdTo)
	}
	if usedFrom != nil {
		query = query.Where("used_at >= ?", *usedFrom)
	}
	if usedTo != nil {
		query = query.Where("used_at <= ?", *usedTo)
	}

	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find coupons: %w", err)
	}

	orderField := "created_at"
	switch sortBy {
	case "created_at", "used_at", "code":
		orderField = sortBy
	}
	orderDir := "DESC"
	if strings.ToUpper(sortDir) == "ASC" {
		orderDir = "ASC"
	}

	offset := (page - 1) * limit
	var coupons []*Coupon
	err = query.OrderExpr(orderField+" "+orderDir).Offset(offset).Limit(limit).Scan(ctx, &coupons)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find coupons: %w", err)
	}
	return coupons, total, nil
}

func (r *CouponRepository) GetStatistics(ctx context.Context, partnerID *uuid.UUID) (map[string]int64, error) {
	stats := make(map[string]int64)

	baseQuery := r.db.NewSelect().Model((*Coupon)(nil))
	if partnerID != nil {
		baseQuery = baseQuery.Where("partner_id = ?", *partnerID)
	}

	count, err := baseQuery.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get statistics: %w", err)
	}
	stats["total"] = int64(count)

	usedQuery := r.db.NewSelect().Model((*Coupon)(nil)).Where("status = ?", "used")
	if partnerID != nil {
		usedQuery = usedQuery.Where("partner_id = ?", *partnerID)
	}
	count, err = usedQuery.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get statistics: %w", err)
	}
	stats["used"] = int64(count)

	newQuery := r.db.NewSelect().Model((*Coupon)(nil)).Where("status = ?", "new")
	if partnerID != nil {
		newQuery = newQuery.Where("partner_id = ?", *partnerID)
	}
	count, err = newQuery.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get statistics: %w", err)
	}
	stats["new"] = int64(count)

	purchasedQuery := r.db.NewSelect().Model((*Coupon)(nil)).Where("is_purchased = ?", true)
	if partnerID != nil {
		purchasedQuery = purchasedQuery.Where("partner_id = ?", *partnerID)
	}
	count, err = purchasedQuery.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get statistics: %w", err)
	}
	stats["purchased"] = int64(count)

	return stats, nil
}

// Resets a coupon to its original state
func (r *CouponRepository) ResetCoupon(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.NewUpdate().Model((*Coupon)(nil)).
		Set("status = ?", "new").
		Set("used_at = NULL").
		Set("is_blocked = ?", false).
		Set("is_purchased = ?", false).
		Set("purchase_email = NULL").
		Set("purchased_at = NULL").
		Set("zip_url = NULL").
		Set("schema_sent_email = NULL").
		Set("schema_sent_at = NULL").
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to reset coupon: %w", err)
	}
	return nil
}

// Alias for ResetCoupon (for handler compatibility)
func (r *CouponRepository) Reset(ctx context.Context, id uuid.UUID) error {
	return r.ResetCoupon(ctx, id)
}

func (r *CouponRepository) CountByPartnerID(ctx context.Context, partnerID uuid.UUID) (int, error) {
	count, err := r.db.NewSelect().Model((*Coupon)(nil)).Where("partner_id = ?", partnerID).Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count coupons: %w", err)
	}
	return count, nil
}

func (r *CouponRepository) CountActivatedByPartnerID(ctx context.Context, partnerID uuid.UUID) (int, error) {
	count, err := r.db.NewSelect().Model((*Coupon)(nil)).Where("partner_id = ? AND activated_at IS NOT NULL", partnerID).Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count activated coupons: %w", err)
	}
	return count, nil
}

// Counts partner created coupons in date range (created_at)
func (r *CouponRepository) CountCreatedByPartnerInRange(ctx context.Context, partnerID uuid.UUID, from, to *time.Time) (int, error) {
	query := r.db.NewSelect().Model((*Coupon)(nil)).Where("partner_id = ?", partnerID)
	if from != nil {
		query = query.Where("created_at >= ?", *from)
	}
	if to != nil {
		query = query.Where("created_at <= ?", *to)
	}
	count, err := query.Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count created coupons in range: %w", err)
	}
	return count, nil
}

// Counts partner activated coupons in date range (activated_at)
func (r *CouponRepository) CountActivatedByPartnerInRange(ctx context.Context, partnerID uuid.UUID, from, to *time.Time) (int, error) {
	query := r.db.NewSelect().Model((*Coupon)(nil)).Where("partner_id = ? AND activated_at IS NOT NULL", partnerID)
	if from != nil {
		query = query.Where("activated_at >= ?", *from)
	}
	if to != nil {
		query = query.Where("activated_at <= ?", *to)
	}
	count, err := query.Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count activated coupons in range: %w", err)
	}
	return count, nil
}

// Counts partner purchased coupons in date range (purchased_at)
func (r *CouponRepository) CountPurchasedByPartnerInRange(ctx context.Context, partnerID uuid.UUID, from, to *time.Time) (int, error) {
	query := r.db.NewSelect().Model((*Coupon)(nil)).Where("partner_id = ? AND is_purchased = ?", partnerID, true)
	if from != nil {
		query = query.Where("purchased_at >= ?", *from)
	}
	if to != nil {
		query = query.Where("purchased_at <= ?", *to)
	}
	count, err := query.Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count purchased coupons in range: %w", err)
	}
	return count, nil
}

func (r *CouponRepository) CountPurchasedByPartnerID(ctx context.Context, partnerID uuid.UUID) (int, error) {
	count, err := r.db.NewSelect().Model((*Coupon)(nil)).Where("partner_id = ? AND is_purchased = ?", partnerID, true).Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count purchased coupons: %w", err)
	}
	return count, nil
}

func (r *CouponRepository) GetFiltered(ctx context.Context, filters map[string]any) ([]*Coupon, error) {
	query := r.db.NewSelect().Model((*Coupon)(nil))

	for key, value := range filters {
		switch key {
		case "partner_id":
			query = query.Where("partner_id = ?", value)
		case "status":
			query = query.Where("status = ?", value)
		case "size":
			query = query.Where("size = ?", value)
		case "style":
			query = query.Where("style = ?", value)
		case "is_purchased":
			query = query.Where("is_purchased = ?", value)
		case "code_search":
			query = query.Where("code ILIKE ?", "%"+value.(string)+"%")
		}
	}

	var coupons []*Coupon
	err := query.Order("created_at DESC").Scan(ctx, &coupons)
	if err != nil {
		return nil, fmt.Errorf("failed to find coupons: %w", err)
	}
	return coupons, nil
}

func (r *CouponRepository) GetRecentActivated(ctx context.Context, limit int) ([]*Coupon, error) {
	var coupons []*Coupon
	err := r.db.NewSelect().Model(&coupons).
		Where("(status = ? OR status = ?) AND (used_at IS NOT NULL OR completed_at IS NOT NULL)", "used", "completed").
		OrderExpr("COALESCE(used_at, completed_at) DESC").
		Limit(limit).
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find recent activated coupons: %w", err)
	}
	return coupons, nil
}

func (r *CouponRepository) GetRecentActivatedByPartner(ctx context.Context, partnerID uuid.UUID, limit int) ([]*Coupon, error) {
	var coupons []*Coupon
	err := r.db.NewSelect().Model(&coupons).
		Where("partner_id = ? AND (activated_at IS NOT NULL OR used_at IS NOT NULL)", partnerID).
		OrderExpr("COALESCE(used_at, activated_at) DESC").
		Limit(limit).
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find recent activated coupons by partner: %w", err)
	}
	return coupons, nil
}

func (r *CouponRepository) SearchPartnerCoupons(
	ctx context.Context,
	partnerID uuid.UUID,
	code, status, size, style string,
	createdFrom, createdTo, usedFrom, usedTo *time.Time,
	sortBy, sortOrder string,
	page, limit int,
) ([]*Coupon, int, error) {
	query := r.db.NewSelect().Model((*Coupon)(nil)).Where("partner_id = ?", partnerID)

	if code != "" {
		query = query.Where("code ILIKE ?", "%"+code+"%")
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if size != "" {
		query = query.Where("size = ?", size)
	}
	if style != "" {
		query = query.Where("style = ?", style)
	}

	if createdFrom != nil {
		query = query.Where("created_at >= ?", *createdFrom)
	}
	if createdTo != nil {
		query = query.Where("created_at <= ?", *createdTo)
	}
	if usedFrom != nil {
		query = query.Where("used_at IS NOT NULL AND used_at >= ?", *usedFrom)
	}
	if usedTo != nil {
		query = query.Where("used_at IS NOT NULL AND used_at <= ?", *usedTo)
	}

	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count partner coupons: %w", err)
	}

	order := "DESC"
	if sortOrder == "asc" || sortOrder == "ASC" {
		order = "ASC"
	}
	switch sortBy {
	case "used_at":
		query = query.Order("used_at " + order)
	case "code":
		query = query.Order("code " + order)
	case "status":
		query = query.Order("status " + order)
	case "created_at":
		fallthrough
	default:
		query = query.Order("created_at " + order)
	}

	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	offset := (page - 1) * limit

	var coupons []*Coupon
	err = query.Offset(offset).Limit(limit).Scan(ctx, &coupons)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find partner coupons: %w", err)
	}
	return coupons, total, nil
}

func (r *CouponRepository) CodeExists(ctx context.Context, code string) (bool, error) {
	count, err := r.db.NewSelect().Model((*Coupon)(nil)).Where("code = ?", code).Count(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to check code exists: %w", err)
	}
	return count > 0, nil
}

func (r *CouponRepository) CountTotal(ctx context.Context) (int64, error) {
	count, err := r.db.NewSelect().Model((*Coupon)(nil)).Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count total coupons: %w", err)
	}
	return int64(count), nil
}

func (r *CouponRepository) CountByStatus(ctx context.Context, status string) (int64, error) {
	count, err := r.db.NewSelect().Model((*Coupon)(nil)).Where("status = ?", status).Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count coupons by status: %w", err)
	}
	return int64(count), nil
}

func (r *CouponRepository) CountByPartner(ctx context.Context, partnerID uuid.UUID) (int64, error) {
	count, err := r.db.NewSelect().Model((*Coupon)(nil)).Where("partner_id = ?", partnerID).Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count coupons by partner: %w", err)
	}
	return int64(count), nil
}

func (r *CouponRepository) CountByPartnerAndStatus(ctx context.Context, partnerID uuid.UUID, status string) (int64, error) {
	count, err := r.db.NewSelect().Model((*Coupon)(nil)).
		Where("partner_id = ? AND status = ?", partnerID, status).Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count coupons by partner and status: %w", err)
	}
	return int64(count), nil
}

// Returns coupons purchased through partner branded site
func (r *CouponRepository) CountBrandedPurchasesByPartner(ctx context.Context, partnerID uuid.UUID) (int64, error) {
	count, err := r.db.NewSelect().Model((*Coupon)(nil)).
		Where("partner_id = ? AND is_purchased = ?", partnerID, true).Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count branded purchases by partner: %w", err)
	}
	return int64(count), nil
}

// Returns partner last activity (last activated coupon)
func (r *CouponRepository) GetLastActivityByPartner(ctx context.Context, partnerID uuid.UUID) (*time.Time, error) {
	var lastActivity time.Time
	err := r.db.NewSelect().Model((*Coupon)(nil)).
		Column("used_at").
		Where("partner_id = ? AND used_at IS NOT NULL", partnerID).
		Order("used_at DESC").
		Limit(1).
		Scan(ctx, &lastActivity)

	if err != nil {
		return nil, nil
	}
	return &lastActivity, nil
}

// Returns data for time series charts
func (r *CouponRepository) GetTimeSeriesData(ctx context.Context, dateFrom, dateTo time.Time, period string, partnerID *uuid.UUID) ([]map[string]any, error) {
	var dateFormat string
	switch period {
	case "day":
		dateFormat = "DATE(created_at)"
	case "week":
		dateFormat = "DATE_TRUNC('week', created_at)"
	case "month":
		dateFormat = "DATE_TRUNC('month', created_at)"
	case "year":
		dateFormat = "DATE_TRUNC('year', created_at)"
	default:
		dateFormat = "DATE(created_at)"
	}

	query := r.db.NewSelect().
		ColumnExpr(dateFormat+" as date").
		ColumnExpr("COUNT(*) as coupons_created").
		ColumnExpr("COUNT(CASE WHEN activated_at IS NOT NULL THEN 1 END) as coupons_activated").
		ColumnExpr("COUNT(CASE WHEN is_purchased = true THEN 1 END) as coupons_purchased").
		ColumnExpr("0 as new_partners_count").
		Model((*Coupon)(nil)).
		Where("created_at >= ? AND created_at <= ?", dateFrom, dateTo).
		Group("date").
		Order("date")

	if partnerID != nil {
		query = query.Where("partner_id = ?", *partnerID)
	}

	rows, err := query.Rows(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get time series data: %w", err)
	}
	defer rows.Close()

	var results []map[string]any
	for rows.Next() {
		var date string
		var created, activated, purchased, partners int64

		if err := rows.Scan(&date, &created, &activated, &purchased, &partners); err != nil {
			return nil, fmt.Errorf("failed to scan time series data: %w", err)
		}

		results = append(results, map[string]any{
			"date":               date,
			"coupons_created":    created,
			"coupons_activated":  activated,
			"coupons_purchased":  purchased,
			"new_partners_count": partners,
		})
	}

	return results, nil
}

func (r *CouponRepository) HealthCheck(ctx context.Context) error {
	return r.db.Ping()
}

func (r *CouponRepository) CountActivatedInTimeRange(ctx context.Context, from, to time.Time) (int64, error) {
	count, err := r.db.NewSelect().Model((*Coupon)(nil)).
		Where("activated_at IS NOT NULL AND activated_at >= ? AND activated_at <= ?", from, to).Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count activated coupons in time range: %w", err)
	}
	return int64(count), nil
}

// FindAvailableCoupon finds an available (unused, not purchased) coupon with specified size and style
func (r *CouponRepository) FindAvailableCoupon(ctx context.Context, size, style string, partnerID *uuid.UUID) (*Coupon, error) {
	query := r.db.NewSelect().Model((*Coupon)(nil)).
		Where("size = ? AND style = ? AND status = 'new' AND is_purchased = false", size, style)

	if partnerID != nil {
		query = query.Where("partner_id = ?", *partnerID)
	} else {
		query = query.Where("partner_id = (SELECT id FROM partners WHERE partner_code = '0000' LIMIT 1)")
	}

	query = query.Order("created_at ASC").Limit(1)

	coupon := new(Coupon)
	err := query.Scan(ctx, coupon)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no available coupons found for size %s and style %s", size, style)
		}
		return nil, fmt.Errorf("failed to find available coupon: %w", err)
	}

	return coupon, nil
}

func (r *CouponRepository) GetStatusCounts(ctx context.Context, partnerID *uuid.UUID) (map[string]int64, error) {
	query := r.db.NewSelect().
		ColumnExpr("status").
		ColumnExpr("COUNT(*) as count").
		Model((*Coupon)(nil)).
		Group("status")

	if partnerID != nil {
		query = query.Where("partner_id = ?", *partnerID)
	}

	rows, err := query.Rows(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get status counts: %w", err)
	}
	defer rows.Close()

	statusCounts := make(map[string]int64)
	for rows.Next() {
		var status string
		var count int64
		if err := rows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("failed to scan status counts: %w", err)
		}
		statusCounts[status] = count
	}

	return statusCounts, nil
}

func (r *CouponRepository) GetSizeCounts(ctx context.Context, partnerID *uuid.UUID) (map[string]int64, error) {
	query := r.db.NewSelect().
		ColumnExpr("size").
		ColumnExpr("COUNT(*) as count").
		Model((*Coupon)(nil)).
		Group("size")

	if partnerID != nil {
		query = query.Where("partner_id = ?", *partnerID)
	}

	rows, err := query.Rows(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get size counts: %w", err)
	}
	defer rows.Close()

	sizeCounts := make(map[string]int64)
	for rows.Next() {
		var size string
		var count int64
		if err := rows.Scan(&size, &count); err != nil {
			return nil, fmt.Errorf("failed to scan size counts: %w", err)
		}
		sizeCounts[size] = count
	}

	return sizeCounts, nil
}

func (r *CouponRepository) GetStyleCounts(ctx context.Context, partnerID *uuid.UUID) (map[string]int64, error) {
	query := r.db.NewSelect().
		ColumnExpr("style").
		ColumnExpr("COUNT(*) as count").
		Model((*Coupon)(nil)).
		Group("style")

	if partnerID != nil {
		query = query.Where("partner_id = ?", *partnerID)
	}

	rows, err := query.Rows(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get style counts: %w", err)
	}
	defer rows.Close()

	styleCounts := make(map[string]int64)
	for rows.Next() {
		var style string
		var count int64
		if err := rows.Scan(&style, &count); err != nil {
			return nil, fmt.Errorf("failed to scan style counts: %w", err)
		}
		styleCounts[style] = count
	}

	return styleCounts, nil
}

func (r *CouponRepository) GetTopActivatedByPartner(ctx context.Context, limit int) ([]PartnerCount, error) {
	if limit <= 0 {
		limit = 10
	}
	rows, err := r.db.NewSelect().
		TableExpr("coupons AS c").
		Join("JOIN partners AS p ON p.id = c.partner_id").
		ColumnExpr("c.partner_id").
		ColumnExpr("p.partner_code").
		ColumnExpr("p.brand_name").
		ColumnExpr("COUNT(*) AS count").
		Where("c.status IN ('used','completed') OR c.activated_at IS NOT NULL").
		GroupExpr("c.partner_id, p.partner_code, p.brand_name").
		OrderExpr("count DESC").
		Limit(limit).
		Rows(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get top activated by partner: %w", err)
	}
	defer rows.Close()
	var result []PartnerCount
	for rows.Next() {
		var pc PartnerCount
		if err := rows.Scan(&pc.PartnerID, &pc.PartnerCode, &pc.BrandName, &pc.Count); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		result = append(result, pc)
	}
	return result, nil
}

func (r *CouponRepository) GetTopPurchasedByPartner(ctx context.Context, limit int) ([]PartnerCount, error) {
	if limit <= 0 {
		limit = 10
	}
	rows, err := r.db.NewSelect().
		TableExpr("coupons AS c").
		Join("JOIN partners AS p ON p.id = c.partner_id").
		ColumnExpr("c.partner_id").
		ColumnExpr("p.partner_code").
		ColumnExpr("p.brand_name").
		ColumnExpr("COUNT(*) AS count").
		Where("c.is_purchased = TRUE").
		GroupExpr("c.partner_id, p.partner_code, p.brand_name").
		OrderExpr("count DESC").
		Limit(limit).
		Rows(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get top purchased by partner: %w", err)
	}
	defer rows.Close()
	var result []PartnerCount
	for rows.Next() {
		var pc PartnerCount
		if err := rows.Scan(&pc.PartnerID, &pc.PartnerCode, &pc.BrandName, &pc.Count); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		result = append(result, pc)
	}
	return result, nil
}

func (r *CouponRepository) CountActivated(ctx context.Context) (int64, error) {
	count, err := r.db.NewSelect().Model((*Coupon)(nil)).Where("activated_at IS NOT NULL").Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count activated coupons: %w", err)
	}
	return int64(count), nil
}

func (r *CouponRepository) CountActivatedByPartner(ctx context.Context, partnerID uuid.UUID) (int64, error) {
	count, err := r.db.NewSelect().Model((*Coupon)(nil)).
		Where("partner_id = ? AND activated_at IS NOT NULL", partnerID).Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count activated coupons by partner: %w", err)
	}
	return int64(count), nil
}

// Returns coupon count by extended statuses
func (r *CouponRepository) GetExtendedStatusCounts(ctx context.Context, partnerID *uuid.UUID) (map[string]int64, error) {
	baseQuery := r.db.NewSelect().Model((*Coupon)(nil))
	if partnerID != nil {
		baseQuery = baseQuery.Where("partner_id = ?", *partnerID)
	}

	statusCounts := make(map[string]int64)

	// Count new (status = 'new' AND activated_at IS NULL)
	newCount, err := baseQuery.
		Where("status = ? AND activated_at IS NULL", "new").
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count new coupons: %w", err)
	}
	statusCounts["new"] = int64(newCount)

	// Count activated (activated_at IS NOT NULL AND used_at IS NULL)
	activatedCount, err := baseQuery.
		Where("activated_at IS NOT NULL AND used_at IS NULL").
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count activated coupons: %w", err)
	}
	statusCounts["activated"] = int64(activatedCount)

	// Count used (status = 'used' OR used_at IS NOT NULL)
	usedCount, err := baseQuery.
		Where("status = ? OR used_at IS NOT NULL", "used").
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count used coupons: %w", err)
	}
	statusCounts["used"] = int64(usedCount)

	// Count completed (completed_at IS NOT NULL)
	completedCount, err := baseQuery.
		Where("completed_at IS NOT NULL").
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count completed coupons: %w", err)
	}
	statusCounts["completed"] = int64(completedCount)

	return statusCounts, nil
}

func (r *CouponRepository) GetCouponsWithAdvancedFilter(ctx context.Context, filter CouponFilterRequest) ([]*CouponInfo, int, error) {
	query := r.db.NewSelect().
		Model((*Coupon)(nil)).
		TableExpr("coupons AS coupon").
		Join("LEFT JOIN partners p ON p.id = coupon.partner_id")

	// Apply filters
	if filter.PartnerID != nil {
		query = query.Where("coupon.partner_id = ?", *filter.PartnerID)
	}

	if filter.Status != "" {
		query = query.Where("coupon.status = ?", filter.Status)
	}

	if filter.Size != "" {
		query = query.Where("coupon.size = ?", filter.Size)
	}

	if filter.Style != "" {
		query = query.Where("coupon.style = ?", filter.Style)
	}

	if filter.Search != "" {
		query = query.Where("coupon.code ILIKE ?", "%"+filter.Search+"%")
	}

	// Date filters
	if filter.CreatedFrom != nil {
		query = query.Where("coupon.created_at >= ?", *filter.CreatedFrom)
	}

	if filter.CreatedTo != nil {
		query = query.Where("coupon.created_at <= ?", *filter.CreatedTo)
	}

	if filter.ActivatedFrom != nil {
		query = query.Where("coupon.activated_at >= ?", *filter.ActivatedFrom)
	}

	if filter.ActivatedTo != nil {
		query = query.Where("coupon.activated_at <= ?", *filter.ActivatedTo)
	}

	// Count total records for pagination
	totalQuery := query.Clone()
	total, err := totalQuery.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count coupons: %w", err)
	}

	// Sorting
	sortBy := "coupon.created_at"
	order := "DESC"

	if filter.SortBy != "" {
		switch filter.SortBy {
		case "created_at":
			sortBy = "coupon.created_at"
		case "activated_at":
			sortBy = "coupon.activated_at"
		case "code":
			sortBy = "coupon.code"
		case "partner_name":
			sortBy = "p.brand_name"
		default:
			sortBy = "coupon.created_at"
		}
	}

	if filter.Order == "asc" {
		order = "ASC"
	}

	query = query.Order(sortBy + " " + order)

	// Pagination
	pageSize := 20
	if filter.PageSize > 0 && filter.PageSize <= 100 {
		pageSize = filter.PageSize
	}

	page := 1
	if filter.Page > 0 {
		page = filter.Page
	}

	offset := (page - 1) * pageSize
	query = query.Offset(offset).Limit(pageSize)

	rows, err := query.
		Column("coupon.*", "p.brand_name").
		Rows(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get coupons: %w", err)
	}
	defer rows.Close()

	var coupons []*CouponInfo
	for rows.Next() {
		var coupon Coupon
		var partnerName string

		err := r.db.ScanRow(ctx, rows, &coupon, &partnerName)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan coupon row: %w", err)
		}

		coupons = append(coupons, &CouponInfo{
			ID:          coupon.ID,
			Code:        coupon.Code,
			PartnerID:   coupon.PartnerID,
			PartnerName: partnerName,
			Status:      coupon.Status,
			Size:        coupon.Size,
			Style:       coupon.Style,
			CreatedAt:   coupon.CreatedAt,
			ActivatedAt: coupon.ActivatedAt,
			UsedAt:      coupon.UsedAt,
		})
	}

	return coupons, total, nil
}

func (r *CouponRepository) BatchReset(ctx context.Context, ids []uuid.UUID) ([]uuid.UUID, []uuid.UUID, error) {
	var success []uuid.UUID
	var failed []uuid.UUID

	// Check which coupons exist and can be reset
	var existingCoupons []*Coupon
	err := r.db.NewSelect().Model(&existingCoupons).Where("id IN (?)", bun.In(ids)).Scan(ctx)
	if err != nil {
		return nil, ids, fmt.Errorf("failed to check existing coupons: %w", err)
	}

	// Create map for existence check
	existingMap := make(map[uuid.UUID]*Coupon)
	for _, coupon := range existingCoupons {
		existingMap[coupon.ID] = coupon
	}

	// Separate existing and non-existing
	var validIDs []uuid.UUID
	for _, id := range ids {
		if _, exists := existingMap[id]; exists {
			validIDs = append(validIDs, id)
		} else {
			failed = append(failed, id)
		}
	}

	if len(validIDs) == 0 {
		return success, failed, nil
	}

	// Perform batch reset
	result, err := r.db.NewUpdate().Model((*Coupon)(nil)).
		Set("status = ?", "new").
		Set("used_at = NULL").
		Set("is_blocked = ?", false).
		Set("is_purchased = ?", false).
		Set("purchase_email = NULL").
		Set("purchased_at = NULL").
		Set("zip_url = NULL").
		Set("schema_sent_email = NULL").
		Set("schema_sent_at = NULL").
		Set("activated_at = NULL").
		Set("user_email = NULL").
		Set("completed_at = NULL").
		Where("id IN (?)", bun.In(validIDs)).
		Exec(ctx)

	if err != nil {
		return success, ids, fmt.Errorf("failed to batch reset coupons: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()

	// If all coupons successfully reset
	if rowsAffected == int64(len(validIDs)) {
		success = validIDs
	} else {
		// If not all coupons reset, check which ones
		var resetCoupons []*Coupon
		err := r.db.NewSelect().Model(&resetCoupons).
			Column("id").
			Where("id IN (?) AND status = ? AND used_at IS NULL", bun.In(validIDs), "new").
			Scan(ctx)

		if err == nil {
			resetMap := make(map[uuid.UUID]bool)
			for _, coupon := range resetCoupons {
				resetMap[coupon.ID] = true
				success = append(success, coupon.ID)
			}

			// Consider others as failed
			for _, id := range validIDs {
				if !resetMap[id] {
					failed = append(failed, id)
				}
			}
		} else {
			// In case of verification error, consider all as failed
			failed = append(failed, validIDs...)
		}
	}

	return success, failed, nil
}

// Returns coupon information for deletion preview
func (r *CouponRepository) GetCouponsForDeletion(ctx context.Context, ids []uuid.UUID) ([]*CouponDeletePreview, error) {
	var previews []*CouponDeletePreview

	rows, err := r.db.NewSelect().
		Model((*Coupon)(nil)).
		TableExpr("coupons AS coupon").
		Column("coupon.id", "coupon.code", "coupon.status", "coupon.created_at", "coupon.used_at", "p.brand_name").
		Join("LEFT JOIN partners p ON p.id = coupon.partner_id").
		Where("coupon.id IN (?)", bun.In(ids)).
		Order("coupon.created_at DESC").
		Rows(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get coupons for deletion preview: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var preview CouponDeletePreview
		var id uuid.UUID
		var brandName sql.NullString
		var usedAt sql.NullTime

		err := rows.Scan(&id, &preview.Code, &preview.Status, &preview.CreatedAt, &usedAt, &brandName)
		if err != nil {
			return nil, fmt.Errorf("failed to scan coupon preview: %w", err)
		}

		preview.ID = id.String()
		if brandName.Valid {
			preview.PartnerName = brandName.String
		} else {
			preview.PartnerName = "Unknown"
		}
		if usedAt.Valid {
			preview.UsedAt = &usedAt.Time
		}

		previews = append(previews, &preview)
	}

	return previews, nil
}

// Returns coupons for export with extended information
func (r *CouponRepository) GetCouponsForExport(ctx context.Context, options ExportOptionsRequest) (any, error) {
	query := r.db.NewSelect().Model((*Coupon)(nil)).TableExpr("coupons AS coupon")

	// Apply filters
	if options.PartnerID != nil {
		pid := strings.TrimSpace(*options.PartnerID)
		// If it's UUID - filter by partner_id
		if parsed, err := uuid.Parse(pid); err == nil {
			query = query.Where("coupon.partner_id = ?", parsed)
		} else {
			// Otherwise treat as partner_code (including '0000'): filter via EXISTS to avoid conflicts with other JOINs by alias p
			query = query.Where("EXISTS (SELECT 1 FROM partners p2 WHERE p2.id = coupon.partner_id AND p2.partner_code = ?)", pid)
		}
	}

	if len(options.PartnerCodes) > 0 {
		// Filter by multiple partner codes
		query = query.Where("EXISTS (SELECT 1 FROM partners p3 WHERE p3.id = coupon.partner_id AND p3.partner_code IN (?))", bun.In(options.PartnerCodes))
	}

	if options.Status != "" {
		query = query.Where("coupon.status = ?", options.Status)
	}

	if options.Size != "" {
		query = query.Where("coupon.size = ?", options.Size)
	}

	if options.Style != "" {
		query = query.Where("coupon.style = ?", options.Style)
	}

	// Date filters
	if options.CreatedFrom != nil {
		query = query.Where("coupon.created_at >= ?", *options.CreatedFrom)
	}

	if options.CreatedTo != nil {
		query = query.Where("coupon.created_at <= ?", *options.CreatedTo)
	}

	if options.ActivatedFrom != nil {
		query = query.Where("coupon.activated_at >= ?", *options.ActivatedFrom)
	}

	if options.ActivatedTo != nil {
		query = query.Where("coupon.activated_at <= ?", *options.ActivatedTo)
	}

	switch options.Format {
	case ExportFormatType("codes"):
		var codes []string
		err := query.Distinct().Column("code").Order("coupon.partner_id ASC").Order("coupon.created_at DESC").Scan(ctx, &codes)
		if err != nil {
			return nil, fmt.Errorf("failed to export coupon codes: %w", err)
		}
		return codes, nil

	case ExportFormatType("basic"):
		var exports []BasicExportRow
		err := query.Distinct().Column("code", "status", "size", "style", "created_at").
			Order("coupon.partner_id ASC").Order("coupon.created_at DESC").Scan(ctx, &exports)
		if err != nil {
			return nil, fmt.Errorf("failed to export basic coupons: %w", err)
		}
		return exports, nil

	case ExportFormatType("partner"):

		rows, err := query.
			Column("coupon.code", "coupon.status", "coupon.size", "coupon.style",
				"coupon.created_at", "p.status", "coupon.activated_at", "coupon.used_at").
			Join("LEFT JOIN partners p ON p.id = coupon.partner_id").
			Distinct().Order("coupon.partner_id ASC").Order("coupon.created_at DESC").
			Rows(ctx)

		if err != nil {
			return nil, fmt.Errorf("failed to query partner coupons for export: %w", err)
		}
		defer rows.Close()

		var exports []PartnerExportRow
		for rows.Next() {
			var export PartnerExportRow
			var partnerStatus sql.NullString

			var activatedAt, usedAt sql.NullTime
			err := rows.Scan(&export.Code, &export.CouponStatus, &export.Size, &export.Style,
				&export.CreatedAt, &partnerStatus, &activatedAt, &usedAt)
			if err != nil {
				return nil, fmt.Errorf("failed to scan partner export row: %w", err)
			}

			if partnerStatus.Valid {
				export.PartnerStatus = partnerStatus.String
			} else {
				export.PartnerStatus = "unknown"
			}

			// Normalize coupon status by time fields
			if usedAt.Valid {
				// if explicitly completed keep it; otherwise consider used
				if export.CouponStatus != "completed" {
					export.CouponStatus = "used"
				}
			} else if activatedAt.Valid {
				// activated but not yet redeemed
				if export.CouponStatus == "new" || export.CouponStatus == "" {
					export.CouponStatus = "activated"
				}
			}

			exports = append(exports, export)
		}
		return exports, nil

	case ExportFormatType("admin"):
		// if CreatedFrom/To etc. are explicitly specified, use them; otherwise limit to last 10,000 to protect against accidental huge export
		// (soft limit for admin format)

		rows, err := query.
			Column("coupon.code", "coupon.partner_id", "coupon.status", "coupon.size", "coupon.style",
				"coupon.created_at", "p.status", "p.brand_name", "p.email", "coupon.activated_at", "coupon.used_at").
			Join("LEFT JOIN partners p ON p.id = coupon.partner_id").
			Distinct().Order("coupon.partner_id ASC").Order("coupon.created_at DESC").
			Limit(10000).
			Rows(ctx)

		if err != nil {
			return nil, fmt.Errorf("failed to query admin coupons for export: %w", err)
		}
		defer rows.Close()

		var exports []AdminExportRow
		for rows.Next() {
			var export AdminExportRow
			var partnerStatus, brandName, email sql.NullString

			var activatedAt, usedAt sql.NullTime
			err := rows.Scan(&export.Code, &export.PartnerID, &export.CouponStatus, &export.Size, &export.Style,
				&export.CreatedAt, &partnerStatus, &brandName, &email, &activatedAt, &usedAt)
			if err != nil {
				return nil, fmt.Errorf("failed to scan admin export row: %w", err)
			}

			if partnerStatus.Valid {
				export.PartnerStatus = partnerStatus.String
			} else {
				export.PartnerStatus = "unknown"
			}

			if brandName.Valid {
				export.BrandName = brandName.String
			} else {
				export.BrandName = "Unknown"
			}

			if email.Valid {
				export.Email = email.String
			} else {
				export.Email = "unknown"
			}

			// Normalize coupon status by time fields
			if usedAt.Valid {
				if export.CouponStatus != "completed" {
					export.CouponStatus = "used"
				}
			} else if activatedAt.Valid {
				if export.CouponStatus == "new" || export.CouponStatus == "" {
					export.CouponStatus = "activated"
				}
			}

			exports = append(exports, export)
		}
		return exports, nil

	case ExportFormatType("activity"):

		rows, err := query.
			Column("coupon.*", "p.brand_name").
			Join("LEFT JOIN partners p ON p.id = coupon.partner_id").
			Distinct().Order("coupon.partner_id ASC").Order("coupon.created_at DESC").
			Rows(ctx)

		if err != nil {
			return nil, fmt.Errorf("failed to query activity coupons for export: %w", err)
		}
		defer rows.Close()

		var exports []ActivityExportRow
		for rows.Next() {
			var coupon Coupon
			var brandName sql.NullString

			err := r.db.ScanRow(ctx, rows, &coupon, &brandName)
			if err != nil {
				return nil, fmt.Errorf("failed to scan activity export row: %w", err)
			}

			export := ActivityExportRow{
				Code:            coupon.Code,
				Status:          coupon.Status,
				Size:            coupon.Size,
				Style:           coupon.Style,
				CreatedAt:       coupon.CreatedAt,
				ActivatedAt:     coupon.ActivatedAt,
				UsedAt:          coupon.UsedAt,
				CompletedAt:     coupon.CompletedAt,
				UserEmail:       coupon.UserEmail,
				PurchaseEmail:   coupon.PurchaseEmail,
				PurchasedAt:     coupon.PurchasedAt,
				SchemaSentEmail: coupon.SchemaSentEmail,
				SchemaSentAt:    coupon.SchemaSentAt,
			}

			if brandName.Valid {
				export.PartnerName = brandName.String
			} else {
				export.PartnerName = "Unknown"
			}

			exports = append(exports, export)
		}
		return exports, nil

	default: // ExportFormatFull

		rows, err := query.
			Column("coupon.code", "coupon.partner_id", "coupon.status", "coupon.size", "coupon.style",
				"coupon.created_at", "coupon.used_at", "p.status", "p.brand_name", "p.email").
			Join("LEFT JOIN partners p ON p.id = coupon.partner_id").
			Distinct().Order("coupon.partner_id ASC").Order("coupon.created_at DESC").
			Rows(ctx)

		if err != nil {
			return nil, fmt.Errorf("failed to query full coupons for export: %w", err)
		}
		defer rows.Close()

		var exports []FullExportRow
		for rows.Next() {
			var export FullExportRow
			var partnerStatus, brandName, email sql.NullString

			err := rows.Scan(&export.Code, &export.PartnerID, &export.CouponStatus, &export.Size, &export.Style,
				&export.CreatedAt, &export.UsedAt, &partnerStatus, &brandName, &email)
			if err != nil {
				return nil, fmt.Errorf("failed to scan full export row: %w", err)
			}

			if partnerStatus.Valid {
				export.PartnerStatus = partnerStatus.String
			} else {
				export.PartnerStatus = "unknown"
			}

			if brandName.Valid {
				export.BrandName = brandName.String
			} else {
				export.BrandName = "Unknown"
			}

			if email.Valid {
				export.Email = email.String
			} else {
				export.Email = "unknown"
			}

			exports = append(exports, export)
		}
		return exports, nil
	}
}

func (r *CouponRepository) GetPartnerCouponsWithFilter(ctx context.Context, partnerID uuid.UUID, filters map[string]any, page, limit int, sortBy, order string) ([]*Coupon, int, error) {
	query := r.db.NewSelect().Model((*Coupon)(nil)).Where("partner_id = ?", partnerID)

	// Apply filters
	for key, value := range filters {
		switch key {
		case "status":
			query = query.Where("status = ?", value)
		case "size":
			query = query.Where("size = ?", value)
		case "style":
			query = query.Where("style = ?", value)
		case "search":
			query = query.Where("code ILIKE ?", "%"+value.(string)+"%")
		case "created_from":
			query = query.Where("created_at >= ?", value)
		case "created_to":
			query = query.Where("created_at <= ?", value)
		case "activated_from":
			query = query.Where("activated_at >= ?", value)
		case "activated_to":
			query = query.Where("activated_at <= ?", value)
		}
	}

	// Count total
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count partner coupons: %w", err)
	}

	// Sorting
	sortColumn := "created_at"
	sortOrder := "DESC"

	switch sortBy {
	case "created_at", "activated_at", "used_at", "code", "status":
		sortColumn = sortBy
	}

	if order == "asc" {
		sortOrder = "ASC"
	}

	query = query.Order(sortColumn + " " + sortOrder)

	// Pagination
	offset := (page - 1) * limit
	query = query.Offset(offset).Limit(limit)

	var coupons []*Coupon
	err = query.Scan(ctx, &coupons)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get partner coupons: %w", err)
	}

	return coupons, total, nil
}

func (r *CouponRepository) GetPartnerCouponByCode(ctx context.Context, partnerID uuid.UUID, code string) (*Coupon, error) {
	coupon := new(Coupon)
	err := r.db.NewSelect().Model(coupon).
		Where("partner_id = ? AND code = ?", partnerID, code).
		Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("coupon not found")
		}
		return nil, fmt.Errorf("failed to find partner coupon by code: %w", err)
	}
	return coupon, nil
}

func (r *CouponRepository) GetPartnerCouponDetail(ctx context.Context, partnerID uuid.UUID, couponID uuid.UUID) (*Coupon, error) {
	coupon := new(Coupon)
	err := r.db.NewSelect().Model(coupon).
		Where("partner_id = ? AND id = ?", partnerID, couponID).
		Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("coupon not found")
		}
		return nil, fmt.Errorf("failed to find partner coupon detail: %w", err)
	}
	return coupon, nil
}

func (r *CouponRepository) GetPartnerRecentActivity(ctx context.Context, partnerID uuid.UUID, limit int) ([]*Coupon, error) {
	var coupons []*Coupon
	err := r.db.NewSelect().Model(&coupons).
		Where("partner_id = ? AND activated_at IS NOT NULL", partnerID).
		Order("activated_at DESC").
		Limit(limit).
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get partner recent activity: %w", err)
	}
	return coupons, nil
}

func (r *CouponRepository) GetPartnerStatistics(ctx context.Context, partnerID uuid.UUID) (map[string]any, error) {
	stats := make(map[string]any)

	// Total coupon count
	totalCount, err := r.db.NewSelect().Model((*Coupon)(nil)).
		Where("partner_id = ?", partnerID).
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count total coupons: %w", err)
	}
	stats["total_coupons"] = int64(totalCount)

	// Activated coupons
	activatedCount, err := r.db.NewSelect().Model((*Coupon)(nil)).
		Where("partner_id = ? AND activated_at IS NOT NULL", partnerID).
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count activated coupons: %w", err)
	}
	stats["activated_coupons"] = int64(activatedCount)

	// Used coupons
	usedCount, err := r.db.NewSelect().Model((*Coupon)(nil)).
		Where("partner_id = ? AND used_at IS NOT NULL", partnerID).
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count used coupons: %w", err)
	}
	stats["used_coupons"] = int64(usedCount)

	// Completed coupons
	completedCount, err := r.db.NewSelect().Model((*Coupon)(nil)).
		Where("partner_id = ? AND completed_at IS NOT NULL", partnerID).
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count completed coupons: %w", err)
	}
	stats["completed_coupons"] = int64(completedCount)

	// Online purchased coupons
	purchasedCount, err := r.db.NewSelect().Model((*Coupon)(nil)).
		Where("partner_id = ? AND is_purchased = true", partnerID).
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count purchased coupons: %w", err)
	}
	stats["purchased_coupons"] = int64(purchasedCount)

	// Last activity
	var lastActivity time.Time
	err = r.db.NewSelect().Model((*Coupon)(nil)).
		Column("activated_at").
		Where("partner_id = ? AND activated_at IS NOT NULL", partnerID).
		Order("activated_at DESC").
		Limit(1).
		Scan(ctx, &lastActivity)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get last activity: %w", err)
	}
	if err != sql.ErrNoRows {
		stats["last_activity"] = &lastActivity
	} else {
		stats["last_activity"] = nil
	}

	return stats, nil
}

func (r *CouponRepository) GetPartnerSalesStatistics(ctx context.Context, partnerID uuid.UUID) (map[string]any, error) {
	stats := make(map[string]any)

	// Total sales (online purchased)
	totalSales, err := r.db.NewSelect().Model((*Coupon)(nil)).
		Where("partner_id = ? AND is_purchased = true", partnerID).
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count total sales: %w", err)
	}
	stats["total_sales"] = int64(totalSales)

	// Sales this month
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	salesThisMonth, err := r.db.NewSelect().Model((*Coupon)(nil)).
		Where("partner_id = ? AND is_purchased = true AND purchased_at >= ?", partnerID, startOfMonth).
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count sales this month: %w", err)
	}
	stats["sales_this_month"] = int64(salesThisMonth)

	// Sales this week
	startOfWeek := now.AddDate(0, 0, -int(now.Weekday()))
	startOfWeek = time.Date(startOfWeek.Year(), startOfWeek.Month(), startOfWeek.Day(), 0, 0, 0, 0, startOfWeek.Location())

	salesThisWeek, err := r.db.NewSelect().Model((*Coupon)(nil)).
		Where("partner_id = ? AND is_purchased = true AND purchased_at >= ?", partnerID, startOfWeek).
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count sales this week: %w", err)
	}
	stats["sales_this_week"] = int64(salesThisWeek)

	// Size statistics
	sizeCounts, err := r.GetSizeCounts(ctx, &partnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get size counts: %w", err)
	}
	stats["top_sizes"] = sizeCounts

	// Style statistics
	styleCounts, err := r.GetStyleCounts(ctx, &partnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get style counts: %w", err)
	}
	stats["top_styles"] = styleCounts

	return stats, nil
}

func (r *CouponRepository) GetPartnerUsageStatistics(ctx context.Context, partnerID uuid.UUID) (map[string]any, error) {
	stats := make(map[string]any)

	// Usage this month
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	usageThisMonth, err := r.db.NewSelect().Model((*Coupon)(nil)).
		Where("partner_id = ? AND used_at IS NOT NULL AND used_at >= ?", partnerID, startOfMonth).
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count usage this month: %w", err)
	}
	stats["usage_this_month"] = int64(usageThisMonth)

	// Usage this week
	startOfWeek := now.AddDate(0, 0, -int(now.Weekday()))
	startOfWeek = time.Date(startOfWeek.Year(), startOfWeek.Month(), startOfWeek.Day(), 0, 0, 0, 0, startOfWeek.Location())

	usageThisWeek, err := r.db.NewSelect().Model((*Coupon)(nil)).
		Where("partner_id = ? AND used_at IS NOT NULL AND used_at >= ?", partnerID, startOfWeek).
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count usage this week: %w", err)
	}
	stats["usage_this_week"] = int64(usageThisWeek)

	// General statistics for coefficient calculation
	totalCoupons, err := r.db.NewSelect().Model((*Coupon)(nil)).
		Where("partner_id = ?", partnerID).
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count total coupons: %w", err)
	}

	activatedCoupons, err := r.db.NewSelect().Model((*Coupon)(nil)).
		Where("partner_id = ? AND activated_at IS NOT NULL", partnerID).
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count activated coupons: %w", err)
	}

	completedCoupons, err := r.db.NewSelect().Model((*Coupon)(nil)).
		Where("partner_id = ? AND completed_at IS NOT NULL", partnerID).
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count completed coupons: %w", err)
	}

	// Conversion rate (activated from total)
	var conversionRate float64
	if totalCoupons > 0 {
		conversionRate = float64(activatedCoupons) / float64(totalCoupons) * 100
	}
	stats["conversion_rate"] = conversionRate

	// Completion rate (completed from activated)
	var completionRate float64
	if activatedCoupons > 0 {
		completionRate = float64(completedCoupons) / float64(activatedCoupons) * 100
	}
	stats["completion_rate"] = completionRate

	// Average time from creation to usage
	var avgTimeToUse sql.NullFloat64
	err = r.db.NewSelect().Model((*Coupon)(nil)).
		ColumnExpr("AVG(EXTRACT(EPOCH FROM (used_at - created_at))/3600)").
		Where("partner_id = ? AND used_at IS NOT NULL", partnerID).
		Scan(ctx, &avgTimeToUse)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate average time to use: %w", err)
	}
	if avgTimeToUse.Valid {
		hours := int64(avgTimeToUse.Float64)
		stats["average_time_to_use"] = &hours
	} else {
		stats["average_time_to_use"] = nil
	}

	return stats, nil
}
