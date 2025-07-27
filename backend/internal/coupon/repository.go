package coupon

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CouponRepository struct {
	db *gorm.DB
}

func NewCouponRepository(db *gorm.DB) *CouponRepository {
	return &CouponRepository{db: db}
}

// Create создает новый купон
func (r *CouponRepository) Create(coupon *Coupon) error {
	return r.db.Create(coupon).Error
}

// CreateBatch создает множество купонов за одну операцию
func (r *CouponRepository) CreateBatch(coupons []*Coupon) error {
	return r.db.CreateInBatches(coupons, 100).Error
}

// GetByCode находит купон по коду
func (r *CouponRepository) GetByCode(code string) (*Coupon, error) {
	var coupon Coupon
	err := r.db.Where("code = ?", code).First(&coupon).Error
	if err != nil {
		return nil, err
	}
	return &coupon, nil
}

// GetByID находит купон по ID
func (r *CouponRepository) GetByID(id uuid.UUID) (*Coupon, error) {
	var coupon Coupon
	err := r.db.Where("id = ?", id).First(&coupon).Error
	if err != nil {
		return nil, err
	}
	return &coupon, nil
}

// GetByPartnerID возвращает купоны партнера
func (r *CouponRepository) GetByPartnerID(partnerID uuid.UUID) ([]*Coupon, error) {
	var coupons []*Coupon
	err := r.db.Where("partner_id = ?", partnerID).Find(&coupons).Error
	if err != nil {
		return nil, ErrFailedToFindCouponsByPartnerID
	}
	return coupons, nil
}

// GetAll возвращает все купоны
func (r *CouponRepository) GetAll() ([]*Coupon, error) {
	var coupons []*Coupon
	err := r.db.Find(&coupons).Error
	if err != nil {
		return nil, ErrFailedToFindAllCoupons
	}
	return coupons, nil
}

// Update обновляет купон
func (r *CouponRepository) Update(coupon *Coupon) error {
	return r.db.Save(coupon).Error
}

// UpdateStatusByPartnerID обновляет статус купонов партнера
func (r *CouponRepository) UpdateStatusByPartnerID(partnerID uuid.UUID, status bool) error {
	return r.db.Model(&Coupon{}).Where("partner_id = ?", partnerID).Update("is_blocked", status).Error
}

// ActivateCoupon активирует купон (меняет статус на 'used')
func (r *CouponRepository) ActivateCoupon(id uuid.UUID, originalImageURL, previewURL, schemaURL string) error {
	now := time.Now()
	return r.db.Model(&Coupon{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":             "used",
		"used_at":            &now,
		"original_image_url": originalImageURL,
		"preview_url":        previewURL,
		"schema_url":         schemaURL,
	}).Error
}

// SendSchema записывает информацию об отправке схемы на email
func (r *CouponRepository) SendSchema(id uuid.UUID, email string) error {
	now := time.Now()
	return r.db.Model(&Coupon{}).Where("id = ?", id).Updates(map[string]interface{}{
		"schema_sent_email": email,
		"schema_sent_at":    &now,
	}).Error
}

// MarkAsPurchased помечает купон как купленный онлайн
func (r *CouponRepository) MarkAsPurchased(id uuid.UUID, purchaseEmail string) error {
	now := time.Now()
	return r.db.Model(&Coupon{}).Where("id = ?", id).Updates(map[string]interface{}{
		"is_purchased":   true,
		"purchase_email": purchaseEmail,
		"purchased_at":   &now,
	}).Error
}

// Delete удаляет купон
func (r *CouponRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&Coupon{}, id).Error
}

// BatchDelete массово удаляет купоны по списку ID
func (r *CouponRepository) BatchDelete(ids []uuid.UUID) (int64, error) {
	result := r.db.Where("id IN ?", ids).Delete(&Coupon{})
	return result.RowsAffected, result.Error
}

// Search выполняет поиск купонов по различным критериям
func (r *CouponRepository) Search(code, status, size, style string, partnerID *uuid.UUID) ([]*Coupon, error) {
	query := r.db.Model(&Coupon{})

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
	err := query.Order("created_at DESC").Find(&coupons).Error
	if err != nil {
		return nil, ErrFailedToFindCoupons
	}
	return coupons, nil
}

// SearchWithPagination выполняет поиск купонов с пагинацией
func (r *CouponRepository) SearchWithPagination(code, status, size, style string, partnerID *uuid.UUID, page, limit int) ([]*Coupon, int64, error) {
	query := r.db.Model(&Coupon{})

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

	// Подсчитываем общее количество
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Применяем пагинацию
	offset := (page - 1) * limit
	var coupons []*Coupon
	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&coupons).Error

	if err != nil {
		return nil, 0, ErrFailedToFindCoupons
	}
	return coupons, total, nil
}

// GetStatistics возвращает статистику по купонам
func (r *CouponRepository) GetStatistics(partnerID *uuid.UUID) (map[string]int64, error) {
	stats := make(map[string]int64)

	baseQuery := r.db.Model(&Coupon{})
	if partnerID != nil {
		baseQuery = baseQuery.Where("partner_id = ?", *partnerID)
	}

	var total, used, new, purchased int64

	// Общее количество купонов
	baseQuery.Count(&total)
	stats["total"] = total

	// Активированные купоны
	usedQuery := r.db.Model(&Coupon{}).Where("status = ?", "used")
	if partnerID != nil {
		usedQuery = usedQuery.Where("partner_id = ?", *partnerID)
	}
	usedQuery.Count(&used)
	stats["used"] = used

	// Новые купоны
	newQuery := r.db.Model(&Coupon{}).Where("status = ?", "new")
	if partnerID != nil {
		newQuery = newQuery.Where("partner_id = ?", *partnerID)
	}
	newQuery.Count(&new)
	stats["new"] = new

	// Купленные онлайн
	purchasedQuery := r.db.Model(&Coupon{}).Where("is_purchased = ?", true)
	if partnerID != nil {
		purchasedQuery = purchasedQuery.Where("partner_id = ?", *partnerID)
	}
	purchasedQuery.Count(&purchased)
	stats["purchased"] = purchased

	return stats, nil
}

// ResetCoupon сбрасывает купон в исходное состояние
func (r *CouponRepository) ResetCoupon(id uuid.UUID) error {
	return r.db.Model(&Coupon{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":             "new",
		"used_at":            nil,
		"is_blocked":         false,
		"is_purchased":       false,
		"purchase_email":     nil,
		"purchased_at":       nil,
		"original_image_url": nil,
		"preview_url":        nil,
		"schema_url":         nil,
		"schema_sent_email":  nil,
		"schema_sent_at":     nil,
	}).Error
}

// Reset - алиас для ResetCoupon (для совместимости с handler)
func (r *CouponRepository) Reset(id uuid.UUID) error {
	return r.ResetCoupon(id)
}

// CountByPartnerID возвращает количество купонов партнера
func (r *CouponRepository) CountByPartnerID(partnerID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&Coupon{}).Where("partner_id = ?", partnerID).Count(&count).Error
	return count, err
}

// CountActivatedByPartnerID возвращает количество активированных купонов партнера
func (r *CouponRepository) CountActivatedByPartnerID(partnerID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&Coupon{}).Where("partner_id = ? AND status = ?", partnerID, "used").Count(&count).Error
	return count, err
}

// CountPurchasedByPartnerID возвращает количество купленных онлайн купонов партнера
func (r *CouponRepository) CountPurchasedByPartnerID(partnerID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&Coupon{}).Where("partner_id = ? AND is_purchased = ?", partnerID, true).Count(&count).Error
	return count, err
}

// GetFiltered возвращает купоны с применением фильтров
func (r *CouponRepository) GetFiltered(filters map[string]interface{}) ([]*Coupon, error) {
	query := r.db.Model(&Coupon{})

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
	err := query.Order("created_at DESC").Find(&coupons).Error
	return coupons, err
}

// GetRecentActivated возвращает последние активированные купоны с сортировкой по дате активации и лимитом
func (r *CouponRepository) GetRecentActivated(limit int) ([]*Coupon, error) {
	var coupons []*Coupon
	err := r.db.Where("status = ? AND used_at IS NOT NULL", "used").
		Order("used_at DESC").
		Limit(limit).
		Find(&coupons).Error
	if err != nil {
		return nil, ErrFailedToFindRecentActivatedCoupons
	}
	return coupons, nil
}

// CodeExists проверяет, существует ли купон с данным кодом
func (r *CouponRepository) CodeExists(code string) (bool, error) {
	var count int64
	err := r.db.Model(&Coupon{}).Where("code = ?", code).Count(&count).Error
	if err != nil {
		return false, ErrFailedToCheckCodeExists
	}
	return count > 0, nil
}
