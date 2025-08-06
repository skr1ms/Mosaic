package coupon

import (
	"context"
	"database/sql"
	"fmt"
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

// Create создаёт новый купон
func (r *CouponRepository) Create(ctx context.Context, coupon *Coupon) error {
	_, err := r.db.NewInsert().Model(coupon).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to create coupon: %w", err)
	}
	return nil
}

// CreateBatch создаёт множество купонов за одну операцию
func (r *CouponRepository) CreateBatch(ctx context.Context, coupons []*Coupon) error {
	_, err := r.db.NewInsert().Model(&coupons).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to create coupon: %w", err)
	}
	return nil
}

// GetByCode находит купон по коду
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

// GetByID находит купон по ID
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

// GetByPartnerID возвращает купоны партнёра
func (r *CouponRepository) GetByPartnerID(ctx context.Context, partnerID uuid.UUID) ([]*Coupon, error) {
	var coupons []*Coupon
	err := r.db.NewSelect().Model(&coupons).Where("partner_id = ?", partnerID).Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find coupons: %w", err)
	}
	return coupons, nil
}

// GetAll возвращает все купоны
func (r *CouponRepository) GetAll(ctx context.Context) ([]*Coupon, error) {
	var coupons []*Coupon
	err := r.db.NewSelect().Model(&coupons).Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find coupons: %w", err)
	}
	return coupons, nil
}

// Update обновляет купон
func (r *CouponRepository) Update(ctx context.Context, coupon *Coupon) error {
	_, err := r.db.NewUpdate().Model(coupon).WherePK().Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to update coupon status: %w", err)
	}
	return nil
}

// UpdateStatusByPartnerID обновляет статус купонов партнёра
func (r *CouponRepository) UpdateStatusByPartnerID(ctx context.Context, partnerID uuid.UUID, status bool) error {
	_, err := r.db.NewUpdate().Model((*Coupon)(nil)).Set("is_blocked = ?", status).Where("partner_id = ?", partnerID).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to update coupon status: %w", err)
	}
	return nil
}

// ActivateCoupon активирует купон (меняет статус на 'used')
func (r *CouponRepository) ActivateCoupon(ctx context.Context, id uuid.UUID, originalImageURL, previewURL, schemaURL string) error {
	now := time.Now()
	_, err := r.db.NewUpdate().Model((*Coupon)(nil)).
		Set("status = ?", "used").
		Set("used_at = ?", &now).
		Set("original_image_url = ?", originalImageURL).
		Set("preview_url = ?", previewURL).
		Set("schema_url = ?", schemaURL).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to activate coupon: %w", err)
	}
	return nil
}

// SendSchema записывает информацию об отправке схемы на email
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

// MarkAsPurchased помечает купон как купленный онлайн
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

// Delete удаляет купон
func (r *CouponRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.NewDelete().Model((*Coupon)(nil)).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete coupon: %w", err)
	}
	return nil
}

// BatchDelete массово удаляет купоны по списку ID
func (r *CouponRepository) BatchDelete(ctx context.Context, ids []uuid.UUID) (int64, error) {
	res, err := r.db.NewDelete().Model((*Coupon)(nil)).Where("id IN (?)", bun.In(ids)).Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to delete coupons: %w", err)
	}
	rows, _ := res.RowsAffected()
	return rows, nil
}

// Search выполняет поиск купонов по различным критериям
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
	err := query.Order("created_at DESC").Scan(ctx, &coupons)
	if err != nil {
		return nil, fmt.Errorf("failed to find coupons: %w", err)
	}
	return coupons, nil
}

// SearchWithPagination выполняет поиск купонов с пагинацией
func (r *CouponRepository) SearchWithPagination(ctx context.Context, code, status, size, style string, partnerID *uuid.UUID, page, limit int) ([]*Coupon, int, error) {
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

	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find coupons: %w", err)
	}

	offset := (page - 1) * limit
	var coupons []*Coupon
	err = query.Order("created_at DESC").Offset(offset).Limit(limit).Scan(ctx, &coupons)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find coupons: %w", err)
	}
	return coupons, total, nil
}

// GetStatistics возвращает статистику по купонам
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

// ResetCoupon сбрасывает купон в исходное состояние
func (r *CouponRepository) ResetCoupon(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.NewUpdate().Model((*Coupon)(nil)).
		Set("status = ?", "new").
		Set("used_at = NULL").
		Set("is_blocked = ?", false).
		Set("is_purchased = ?", false).
		Set("purchase_email = NULL").
		Set("purchased_at = NULL").
		Set("original_image_url = NULL").
		Set("preview_url = NULL").
		Set("schema_url = NULL").
		Set("schema_sent_email = NULL").
		Set("schema_sent_at = NULL").
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to reset coupon: %w", err)
	}
	return nil
}

// Reset - алиас для ResetCoupon (для совместимости с handler)
func (r *CouponRepository) Reset(ctx context.Context, id uuid.UUID) error {
	return r.ResetCoupon(ctx, id)
}

// CountByPartnerID возвращает количество купонов партнёра
func (r *CouponRepository) CountByPartnerID(ctx context.Context, partnerID uuid.UUID) (int, error) {
	count, err := r.db.NewSelect().Model((*Coupon)(nil)).Where("partner_id = ?", partnerID).Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count coupons: %w", err)
	}
	return count, nil
}

// CountActivatedByPartnerID возвращает количество активированных купонов партнёра
func (r *CouponRepository) CountActivatedByPartnerID(ctx context.Context, partnerID uuid.UUID) (int, error) {
	count, err := r.db.NewSelect().Model((*Coupon)(nil)).Where("partner_id = ? AND activated_at IS NOT NULL", partnerID).Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count activated coupons: %w", err)
	}
	return count, nil
}

// CountPurchasedByPartnerID возвращает количество купленных онлайн купонов партнёра
func (r *CouponRepository) CountPurchasedByPartnerID(ctx context.Context, partnerID uuid.UUID) (int, error) {
	count, err := r.db.NewSelect().Model((*Coupon)(nil)).Where("partner_id = ? AND is_purchased = ?", partnerID, true).Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count purchased coupons: %w", err)
	}
	return count, nil
}

// GetFiltered возвращает купоны с применением фильтров
func (r *CouponRepository) GetFiltered(ctx context.Context, filters map[string]interface{}) ([]*Coupon, error) {
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

// GetRecentActivated возвращает последние активированные купоны с сортировкой по дате активации и лимитом
func (r *CouponRepository) GetRecentActivated(ctx context.Context, limit int) ([]*Coupon, error) {
	var coupons []*Coupon
	err := r.db.NewSelect().Model(&coupons).
		Where("status = ? AND used_at IS NOT NULL", "used").
		Order("used_at DESC").
		Limit(limit).
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find recent activated coupons: %w", err)
	}
	return coupons, nil
}

// CodeExists проверяет, существует ли купон с данным кодом
func (r *CouponRepository) CodeExists(ctx context.Context, code string) (bool, error) {
	count, err := r.db.NewSelect().Model((*Coupon)(nil)).Where("code = ?", code).Count(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to check code exists: %w", err)
	}
	return count > 0, nil
}

// Методы для статистики

// CountTotal возвращает общее количество купонов
func (r *CouponRepository) CountTotal(ctx context.Context) (int64, error) {
	count, err := r.db.NewSelect().Model((*Coupon)(nil)).Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count total coupons: %w", err)
	}
	return int64(count), nil
}

// CountByStatus возвращает количество купонов по статусу
func (r *CouponRepository) CountByStatus(ctx context.Context, status string) (int64, error) {
	count, err := r.db.NewSelect().Model((*Coupon)(nil)).Where("status = ?", status).Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count coupons by status: %w", err)
	}
	return int64(count), nil
}

// CountByPartner возвращает количество купонов партнера
func (r *CouponRepository) CountByPartner(ctx context.Context, partnerID uuid.UUID) (int64, error) {
	count, err := r.db.NewSelect().Model((*Coupon)(nil)).Where("partner_id = ?", partnerID).Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count coupons by partner: %w", err)
	}
	return int64(count), nil
}

// CountByPartnerAndStatus возвращает количество купонов партнера по статусу
func (r *CouponRepository) CountByPartnerAndStatus(ctx context.Context, partnerID uuid.UUID, status string) (int64, error) {
	count, err := r.db.NewSelect().Model((*Coupon)(nil)).
		Where("partner_id = ? AND status = ?", partnerID, status).Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count coupons by partner and status: %w", err)
	}
	return int64(count), nil
}

// CountBrandedPurchasesByPartner возвращает количество купонов, купленных через брендированный сайт партнера
func (r *CouponRepository) CountBrandedPurchasesByPartner(ctx context.Context, partnerID uuid.UUID) (int64, error) {
	count, err := r.db.NewSelect().Model((*Coupon)(nil)).
		Where("partner_id = ? AND is_purchased = ?", partnerID, true).Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count branded purchases by partner: %w", err)
	}
	return int64(count), nil
}

// GetLastActivityByPartner возвращает последнюю активность партнера (последний активированный купон)
func (r *CouponRepository) GetLastActivityByPartner(ctx context.Context, partnerID uuid.UUID) (*time.Time, error) {
	var lastActivity time.Time
	err := r.db.NewSelect().Model((*Coupon)(nil)).
		Column("used_at").
		Where("partner_id = ? AND used_at IS NOT NULL", partnerID).
		Order("used_at DESC").
		Limit(1).
		Scan(ctx, &lastActivity)

	if err != nil {
		return nil, nil // Нет активности
	}
	return &lastActivity, nil
}

// GetTimeSeriesData возвращает данные для временных графиков
func (r *CouponRepository) GetTimeSeriesData(ctx context.Context, dateFrom, dateTo time.Time, period string, partnerID *uuid.UUID) ([]map[string]interface{}, error) {
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

	var results []map[string]interface{}
	for rows.Next() {
		var date string
		var created, activated, purchased, partners int64

		if err := rows.Scan(&date, &created, &activated, &purchased, &partners); err != nil {
			return nil, fmt.Errorf("failed to scan time series data: %w", err)
		}

		results = append(results, map[string]interface{}{
			"date":               date,
			"coupons_created":    created,
			"coupons_activated":  activated,
			"coupons_purchased":  purchased,
			"new_partners_count": partners,
		})
	}

	return results, nil
}

// HealthCheck проверяет соединение с БД
func (r *CouponRepository) HealthCheck(ctx context.Context) error {
	return r.db.Ping()
}

// CountActivatedInTimeRange возвращает количество активированных купонов в временном диапазоне
func (r *CouponRepository) CountActivatedInTimeRange(ctx context.Context, from, to time.Time) (int64, error) {
	count, err := r.db.NewSelect().Model((*Coupon)(nil)).
		Where("activated_at IS NOT NULL AND activated_at >= ? AND activated_at <= ?", from, to).Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count activated coupons in time range: %w", err)
	}
	return int64(count), nil
}

// GetStatusCounts возвращает подсчет купонов по статусам
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

// GetSizeCounts возвращает подсчет купонов по размерам
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

// GetStyleCounts возвращает подсчет купонов по стилям
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

// CountActivated возвращает количество активированных купонов (где activated_at IS NOT NULL)
func (r *CouponRepository) CountActivated(ctx context.Context) (int64, error) {
	count, err := r.db.NewSelect().Model((*Coupon)(nil)).Where("activated_at IS NOT NULL").Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count activated coupons: %w", err)
	}
	return int64(count), nil
}

// CountActivatedByPartner возвращает количество активированных купонов партнера
func (r *CouponRepository) CountActivatedByPartner(ctx context.Context, partnerID uuid.UUID) (int64, error) {
	count, err := r.db.NewSelect().Model((*Coupon)(nil)).
		Where("partner_id = ? AND activated_at IS NOT NULL", partnerID).Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count activated coupons by partner: %w", err)
	}
	return int64(count), nil
}

// GetExtendedStatusCounts возвращает подсчет купонов по расширенным статусам
func (r *CouponRepository) GetExtendedStatusCounts(ctx context.Context, partnerID *uuid.UUID) (map[string]int64, error) {
	baseQuery := r.db.NewSelect().Model((*Coupon)(nil))
	if partnerID != nil {
		baseQuery = baseQuery.Where("partner_id = ?", *partnerID)
	}

	statusCounts := make(map[string]int64)

	// Подсчет new (status = 'new' AND activated_at IS NULL)
	newCount, err := baseQuery.
		Where("status = ? AND activated_at IS NULL", "new").
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count new coupons: %w", err)
	}
	statusCounts["new"] = int64(newCount)

	// Подсчет activated (activated_at IS NOT NULL AND used_at IS NULL)
	activatedCount, err := baseQuery.
		Where("activated_at IS NOT NULL AND used_at IS NULL").
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count activated coupons: %w", err)
	}
	statusCounts["activated"] = int64(activatedCount)

	// Подсчет used (status = 'used' OR used_at IS NOT NULL)
	usedCount, err := baseQuery.
		Where("status = ? OR used_at IS NOT NULL", "used").
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count used coupons: %w", err)
	}
	statusCounts["used"] = int64(usedCount)

	// Подсчет completed (completed_at IS NOT NULL)
	completedCount, err := baseQuery.
		Where("completed_at IS NOT NULL").
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count completed coupons: %w", err)
	}
	statusCounts["completed"] = int64(completedCount)

	return statusCounts, nil
}

// GetCouponsWithAdvancedFilter возвращает купоны с продвинутой фильтрацией и пагинацией
func (r *CouponRepository) GetCouponsWithAdvancedFilter(ctx context.Context, filter CouponFilterRequest) ([]*CouponInfo, int, error) {
	// Базовый запрос с JOIN для получения имени партнера
	query := r.db.NewSelect().
		Model((*Coupon)(nil)).
		Join("LEFT JOIN partners p ON p.id = coupon.partner_id")

	// Применяем фильтры
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

	// Фильтры по датам
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

	// Подсчет общего количества записей для пагинации
	totalQuery := query.Clone()
	total, err := totalQuery.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count coupons: %w", err)
	}

	// Сортировка
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

	// Пагинация
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

// BatchReset сбрасывает купоны пакетно
func (r *CouponRepository) BatchReset(ctx context.Context, ids []uuid.UUID) ([]uuid.UUID, []uuid.UUID, error) {
	var success []uuid.UUID
	var failed []uuid.UUID

	// Проверяем какие купоны существуют и можно ли их сбросить
	var existingCoupons []*Coupon
	err := r.db.NewSelect().Model(&existingCoupons).Where("id IN (?)", bun.In(ids)).Scan(ctx)
	if err != nil {
		return nil, ids, fmt.Errorf("failed to check existing coupons: %w", err)
	}

	// Создаем мапу для проверки существования
	existingMap := make(map[uuid.UUID]*Coupon)
	for _, coupon := range existingCoupons {
		existingMap[coupon.ID] = coupon
	}

	// Разделяем на существующие и несуществующие
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

	// Выполняем пакетный сброс
	result, err := r.db.NewUpdate().Model((*Coupon)(nil)).
		Set("status = ?", "new").
		Set("used_at = NULL").
		Set("is_blocked = ?", false).
		Set("is_purchased = ?", false).
		Set("purchase_email = NULL").
		Set("purchased_at = NULL").
		Set("original_image_url = NULL").
		Set("preview_url = NULL").
		Set("schema_url = NULL").
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

	// Если все купоны успешно сброшены
	if rowsAffected == int64(len(validIDs)) {
		success = validIDs
	} else {
		// Если не все купоны сброшены, проверяем какие именно
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

			// Остальные считаем неуспешными
			for _, id := range validIDs {
				if !resetMap[id] {
					failed = append(failed, id)
				}
			}
		} else {
			// В случае ошибки проверки считаем все неуспешными
			failed = append(failed, validIDs...)
		}
	}

	return success, failed, nil
}

// GetCouponsForDeletion возвращает информацию о купонах для предпросмотра удаления
func (r *CouponRepository) GetCouponsForDeletion(ctx context.Context, ids []uuid.UUID) ([]*CouponDeletePreview, error) {
	var previews []*CouponDeletePreview

	rows, err := r.db.NewSelect().
		Model((*Coupon)(nil)).
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

// GetCouponsForExport возвращает купоны для экспорта с расширенной информацией
func (r *CouponRepository) GetCouponsForExport(ctx context.Context, options ExportOptionsRequest) (interface{}, error) {
	query := r.db.NewSelect().Model((*Coupon)(nil))

	// Применяем фильтры
	if options.PartnerID != nil {
		partnerID, err := uuid.Parse(*options.PartnerID)
		if err == nil {
			query = query.Where("coupon.partner_id = ?", partnerID)
		}
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

	// Фильтры по датам
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
		err := query.Column("code").Order("coupon.partner_id ASC, coupon.created_at DESC").Scan(ctx, &codes)
		return codes, err

	case ExportFormatType("basic"):
		type BasicExport struct {
			Code      string    `json:"code"`
			Status    string    `json:"status"`
			Size      string    `json:"size"`
			Style     string    `json:"style"`
			CreatedAt time.Time `json:"created_at"`
		}
		var exports []BasicExport
		err := query.Column("code", "status", "size", "style", "created_at").
			Order("coupon.partner_id ASC, coupon.created_at DESC").Scan(ctx, &exports)
		return exports, err

	case ExportFormatType("partner"):
		// Для партнеров: Coupon Code, Partner Status, Coupon Status, Size, Style, Created At
		type PartnerExport struct {
			Code          string    `json:"code"`
			PartnerStatus string    `json:"partner_status"`
			CouponStatus  string    `json:"coupon_status"`
			Size          string    `json:"size"`
			Style         string    `json:"style"`
			CreatedAt     time.Time `json:"created_at"`
		}

		rows, err := query.
			Column("coupon.code", "coupon.status", "coupon.size", "coupon.style",
				"coupon.created_at", "p.status").
			Join("LEFT JOIN partners p ON p.id = coupon.partner_id").
			Order("coupon.partner_id ASC, coupon.created_at DESC").
			Rows(ctx)

		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var exports []PartnerExport
		for rows.Next() {
			var export PartnerExport
			var partnerStatus sql.NullString

			err := rows.Scan(&export.Code, &export.CouponStatus, &export.Size, &export.Style,
				&export.CreatedAt, &partnerStatus)
			if err != nil {
				return nil, err
			}

			if partnerStatus.Valid {
				export.PartnerStatus = partnerStatus.String
			} else {
				export.PartnerStatus = "unknown"
			}

			exports = append(exports, export)
		}
		return exports, nil

	case ExportFormatType("admin"):
		// Для админа (новые купоны): Coupon Code, Partner ID, Partner Status, Coupon Status, Size, Style, Brand Name, Email, Created At
		type AdminExport struct {
			Code          string    `json:"code"`
			PartnerID     string    `json:"partner_id"`
			PartnerStatus string    `json:"partner_status"`
			CouponStatus  string    `json:"coupon_status"`
			Size          string    `json:"size"`
			Style         string    `json:"style"`
			BrandName     string    `json:"brand_name"`
			Email         string    `json:"email"`
			CreatedAt     time.Time `json:"created_at"`
		}

		rows, err := query.
			Column("coupon.code", "coupon.partner_id", "coupon.status", "coupon.size", "coupon.style",
				"coupon.created_at", "p.status", "p.brand_name", "p.email").
			Join("LEFT JOIN partners p ON p.id = coupon.partner_id").
			Order("coupon.partner_id ASC, coupon.created_at DESC").
			Rows(ctx)

		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var exports []AdminExport
		for rows.Next() {
			var export AdminExport
			var partnerStatus, brandName, email sql.NullString

			err := rows.Scan(&export.Code, &export.PartnerID, &export.CouponStatus, &export.Size, &export.Style,
				&export.CreatedAt, &partnerStatus, &brandName, &email)
			if err != nil {
				return nil, err
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

	case ExportFormatType("activity"):
		type ActivityExport struct {
			Code            string     `json:"code"`
			PartnerName     string     `json:"partner_name"`
			Status          string     `json:"status"`
			Size            string     `json:"size"`
			Style           string     `json:"style"`
			CreatedAt       time.Time  `json:"created_at"`
			ActivatedAt     *time.Time `json:"activated_at"`
			UsedAt          *time.Time `json:"used_at"`
			CompletedAt     *time.Time `json:"completed_at"`
			UserEmail       *string    `json:"user_email"`
			PurchaseEmail   *string    `json:"purchase_email"`
			PurchasedAt     *time.Time `json:"purchased_at"`
			SchemaSentEmail *string    `json:"schema_sent_email"`
			SchemaSentAt    *time.Time `json:"schema_sent_at"`
		}

		rows, err := query.
			Column("coupon.*", "p.brand_name").
			Join("LEFT JOIN partners p ON p.id = coupon.partner_id").
			Order("coupon.partner_id ASC, coupon.created_at DESC").
			Rows(ctx)

		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var exports []ActivityExport
		for rows.Next() {
			var coupon Coupon
			var brandName sql.NullString

			err := r.db.ScanRow(ctx, rows, &coupon, &brandName)
			if err != nil {
				return nil, err
			}

			export := ActivityExport{
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
		// Для админа (все купоны): Coupon Code, Partner ID, Partner Status, Coupon Status, Size, Style, Brand Name, Email, Created At, Used At
		type FullExport struct {
			Code          string     `json:"code"`
			PartnerID     string     `json:"partner_id"`
			PartnerStatus string     `json:"partner_status"`
			CouponStatus  string     `json:"coupon_status"`
			Size          string     `json:"size"`
			Style         string     `json:"style"`
			BrandName     string     `json:"brand_name"`
			Email         string     `json:"email"`
			CreatedAt     time.Time  `json:"created_at"`
			UsedAt        *time.Time `json:"used_at"`
		}

		rows, err := query.
			Column("coupon.code", "coupon.partner_id", "coupon.status", "coupon.size", "coupon.style",
				"coupon.created_at", "coupon.used_at", "p.status", "p.brand_name", "p.email").
			Join("LEFT JOIN partners p ON p.id = coupon.partner_id").
			Order("coupon.partner_id ASC, coupon.created_at DESC").
			Rows(ctx)

		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var exports []FullExport
		for rows.Next() {
			var export FullExport
			var partnerStatus, brandName, email sql.NullString

			err := rows.Scan(&export.Code, &export.PartnerID, &export.CouponStatus, &export.Size, &export.Style,
				&export.CreatedAt, &export.UsedAt, &partnerStatus, &brandName, &email)
			if err != nil {
				return nil, err
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

// GetPartnerCouponsWithFilter возвращает купоны партнера с фильтрацией и пагинацией
func (r *CouponRepository) GetPartnerCouponsWithFilter(ctx context.Context, partnerID uuid.UUID, filters map[string]interface{}, page, limit int, sortBy, order string) ([]*Coupon, int, error) {
	query := r.db.NewSelect().Model((*Coupon)(nil)).Where("partner_id = ?", partnerID)

	// Применяем фильтры
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

	// Подсчет общего количества
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count partner coupons: %w", err)
	}

	// Сортировка
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

	// Пагинация
	offset := (page - 1) * limit
	query = query.Offset(offset).Limit(limit)

	var coupons []*Coupon
	err = query.Scan(ctx, &coupons)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get partner coupons: %w", err)
	}

	return coupons, total, nil
}

// GetPartnerCouponByCode возвращает купон партнера по коду
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

// GetPartnerCouponDetail возвращает детальную информацию о купоне партнера
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

// GetPartnerRecentActivity возвращает последние активированные купоны партнера
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

// GetPartnerStatistics возвращает детальную статистику партнера
func (r *CouponRepository) GetPartnerStatistics(ctx context.Context, partnerID uuid.UUID) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Общее количество купонов
	totalCount, err := r.db.NewSelect().Model((*Coupon)(nil)).
		Where("partner_id = ?", partnerID).
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count total coupons: %w", err)
	}
	stats["total_coupons"] = int64(totalCount)

	// Активированные купоны
	activatedCount, err := r.db.NewSelect().Model((*Coupon)(nil)).
		Where("partner_id = ? AND activated_at IS NOT NULL", partnerID).
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count activated coupons: %w", err)
	}
	stats["activated_coupons"] = int64(activatedCount)

	// Использованные купоны
	usedCount, err := r.db.NewSelect().Model((*Coupon)(nil)).
		Where("partner_id = ? AND used_at IS NOT NULL", partnerID).
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count used coupons: %w", err)
	}
	stats["used_coupons"] = int64(usedCount)

	// Завершенные купоны
	completedCount, err := r.db.NewSelect().Model((*Coupon)(nil)).
		Where("partner_id = ? AND completed_at IS NOT NULL", partnerID).
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count completed coupons: %w", err)
	}
	stats["completed_coupons"] = int64(completedCount)

	// Купленные онлайн купоны
	purchasedCount, err := r.db.NewSelect().Model((*Coupon)(nil)).
		Where("partner_id = ? AND is_purchased = true", partnerID).
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count purchased coupons: %w", err)
	}
	stats["purchased_coupons"] = int64(purchasedCount)

	// Последняя активность
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

// GetPartnerSalesStatistics возвращает статистику продаж партнера
func (r *CouponRepository) GetPartnerSalesStatistics(ctx context.Context, partnerID uuid.UUID) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Общие продажи (купленные онлайн)
	totalSales, err := r.db.NewSelect().Model((*Coupon)(nil)).
		Where("partner_id = ? AND is_purchased = true", partnerID).
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count total sales: %w", err)
	}
	stats["total_sales"] = int64(totalSales)

	// Продажи в этом месяце
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	salesThisMonth, err := r.db.NewSelect().Model((*Coupon)(nil)).
		Where("partner_id = ? AND is_purchased = true AND purchased_at >= ?", partnerID, startOfMonth).
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count sales this month: %w", err)
	}
	stats["sales_this_month"] = int64(salesThisMonth)

	// Продажи на этой неделе
	startOfWeek := now.AddDate(0, 0, -int(now.Weekday()))
	startOfWeek = time.Date(startOfWeek.Year(), startOfWeek.Month(), startOfWeek.Day(), 0, 0, 0, 0, startOfWeek.Location())

	salesThisWeek, err := r.db.NewSelect().Model((*Coupon)(nil)).
		Where("partner_id = ? AND is_purchased = true AND purchased_at >= ?", partnerID, startOfWeek).
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count sales this week: %w", err)
	}
	stats["sales_this_week"] = int64(salesThisWeek)

	// Статистика по размерам
	sizeCounts, err := r.GetSizeCounts(ctx, &partnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get size counts: %w", err)
	}
	stats["top_sizes"] = sizeCounts

	// Статистика по стилям
	styleCounts, err := r.GetStyleCounts(ctx, &partnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get style counts: %w", err)
	}
	stats["top_styles"] = styleCounts

	return stats, nil
}

// GetPartnerUsageStatistics возвращает статистику использования купонов партнера
func (r *CouponRepository) GetPartnerUsageStatistics(ctx context.Context, partnerID uuid.UUID) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Использование в этом месяце
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	usageThisMonth, err := r.db.NewSelect().Model((*Coupon)(nil)).
		Where("partner_id = ? AND used_at IS NOT NULL AND used_at >= ?", partnerID, startOfMonth).
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count usage this month: %w", err)
	}
	stats["usage_this_month"] = int64(usageThisMonth)

	// Использование на этой неделе
	startOfWeek := now.AddDate(0, 0, -int(now.Weekday()))
	startOfWeek = time.Date(startOfWeek.Year(), startOfWeek.Month(), startOfWeek.Day(), 0, 0, 0, 0, startOfWeek.Location())

	usageThisWeek, err := r.db.NewSelect().Model((*Coupon)(nil)).
		Where("partner_id = ? AND used_at IS NOT NULL AND used_at >= ?", partnerID, startOfWeek).
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count usage this week: %w", err)
	}
	stats["usage_this_week"] = int64(usageThisWeek)

	// Общая статистика для расчета коэффициентов
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

	// Коэффициент конверсии (активированные от общего числа)
	var conversionRate float64
	if totalCoupons > 0 {
		conversionRate = float64(activatedCoupons) / float64(totalCoupons) * 100
	}
	stats["conversion_rate"] = conversionRate

	// Коэффициент завершения (завершенные от активированных)
	var completionRate float64
	if activatedCoupons > 0 {
		completionRate = float64(completedCoupons) / float64(activatedCoupons) * 100
	}
	stats["completion_rate"] = completionRate

	// Среднее время от создания до использования
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
