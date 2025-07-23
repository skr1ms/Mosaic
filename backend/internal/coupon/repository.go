package coupon

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CouponRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *CouponRepository {
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
	return coupons, err
}

// GetAll возвращает все купоны
func (r *CouponRepository) GetAll() ([]*Coupon, error) {
	var coupons []*Coupon
	err := r.db.Find(&coupons).Error
	return coupons, err
}

// Update обновляет купон
func (r *CouponRepository) Update(coupon *Coupon) error {
	return r.db.Save(coupon).Error
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

// Search выполняет поиск купонов по различным критериям
func (r *CouponRepository) Search(code, status, size, style string, partnerID *uuid.UUID) ([]*Coupon, error) {
	if code != "" {
		r.db = r.db.Where("code ILIKE ?", "%"+code+"%")
	}

	if status != "" {
		r.db = r.db.Where("status = ?", status)
	}

	if size != "" {
		r.db = r.db.Where("size = ?", size)
	}

	if style != "" {
		r.db = r.db.Where("style = ?", style)
	}

	if partnerID != nil {
		r.db = r.db.Where("partner_id = ?", *partnerID)
	}

	var coupons []*Coupon
	err := r.db.Find(&coupons).Error
	return coupons, err
}

// GetStatistics возвращает статистику по купонам
func (r *CouponRepository) GetStatistics(partnerID *uuid.UUID) (map[string]int64, error) {
	stats := make(map[string]int64)

	r.db.Model(&Coupon{})
	if partnerID != nil {
		r.db = r.db.Where("partner_id = ?", *partnerID)
	}

	var total, used, new, purchased int64

	// Общее количество купонов
	r.db.Count(&total)
	stats["total"] = total

	// Активированные купоны
	r.db.Model(&Coupon{}).Where("status = ?", "used").Count(&used)
	if partnerID != nil {
		r.db.Model(&Coupon{}).Where("partner_id = ? AND status = ?", *partnerID, "used").Count(&used)
	}
	stats["used"] = used

	// Новые купоны
	r.db.Model(&Coupon{}).Where("status = ?", "new").Count(&new)
	if partnerID != nil {
		r.db.Model(&Coupon{}).Where("partner_id = ? AND status = ?", *partnerID, "new").Count(&new)
	}
	stats["new"] = new

	// Купленные онлайн
	if partnerID != nil {
		r.db.Model(&Coupon{}).Where("partner_id = ? AND is_purchased = ?", *partnerID, true).Count(&purchased)
	} else {
		r.db.Model(&Coupon{}).Where("is_purchased = ?", true).Count(&purchased)
	}
	stats["purchased"] = purchased

	return stats, nil
}

// ResetCoupon сбрасывает купон в исходное состояние
func (r *CouponRepository) ResetCoupon(id uuid.UUID) error {
	return r.db.Model(&Coupon{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":             "new",
		"used_at":            nil,
		"original_image_url": nil,
		"preview_url":        nil,
		"schema_url":         nil,
		"schema_sent_email":  nil,
		"schema_sent_at":     nil,
	}).Error
}
