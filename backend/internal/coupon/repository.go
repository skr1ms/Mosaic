package coupon

import (
	"context"
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
		return nil, fmt.Errorf("failed to find coupon by code: %w", err)
	}
	return coupon, nil
}

// GetByID находит купон по ID
func (r *CouponRepository) GetByID(ctx context.Context, id uuid.UUID) (*Coupon, error) {
	coupon := new(Coupon)
	err := r.db.NewSelect().Model(coupon).Where("id = ?", id).Scan(ctx)
	if err != nil {
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
	count, err := r.db.NewSelect().Model((*Coupon)(nil)).Where("partner_id = ? AND status = ?", partnerID, "used").Count(ctx)
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
