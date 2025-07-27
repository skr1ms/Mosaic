package partner

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PartnerRepository struct {
	db *gorm.DB
}

func NewPartnerRepository(db *gorm.DB) *PartnerRepository {
	return &PartnerRepository{db: db}
}

// Create создает нового партнера
func (r *PartnerRepository) Create(partner *Partner) error {
	return r.db.Create(partner).Error
}

// GetByLogin находит партнера по логину
func (r *PartnerRepository) GetByLogin(login string) (*Partner, error) {
	var partner Partner
	err := r.db.Where("login = ?", login).First(&partner).Error
	if err != nil {
		return nil, err
	}
	return &partner, nil
}

// GetByPartnerCode находит партнера по коду
func (r *PartnerRepository) GetByPartnerCode(code string) (*Partner, error) {
	var partner Partner
	err := r.db.Where("partner_code = ?", code).First(&partner).Error
	if err != nil {
		return nil, err
	}
	return &partner, nil
}

// GetByDomain находит партнера по домену
func (r *PartnerRepository) GetByDomain(domain string) (*Partner, error) {
	var partner Partner
	err := r.db.Where("domain = ?", domain).First(&partner).Error
	if err != nil {
		return nil, err
	}
	return &partner, nil
}

// GetByID находит партнера по ID
func (r *PartnerRepository) GetByID(id uuid.UUID) (*Partner, error) {
	var partner Partner
	err := r.db.Where("id = ?", id).First(&partner).Error
	if err != nil {
		return nil, err
	}
	return &partner, nil
}

// GetByEmail находит партнера по email
func (r *PartnerRepository) GetByEmail(email string) (*Partner, error) {
	var partner Partner
	err := r.db.Where("email = ?", email).First(&partner).Error
	if err != nil {
		return nil, err
	}
	return &partner, nil
}

// GetAll возвращает всех партнеров
func (r *PartnerRepository) GetAll() ([]*Partner, error) {
	var partners []*Partner
	err := r.db.Find(&partners).Error
	return partners, err
}

// GetActivePartners возвращает только активных партнеров
func (r *PartnerRepository) GetActivePartners() ([]*Partner, error) {
	var partners []*Partner
	err := r.db.Where("status = ?", "active").Find(&partners).Error
	return partners, err
}

// Update обновляет данные партнера
func (r *PartnerRepository) Update(partner *Partner) error {
	return r.db.Save(partner).Error
}

// UpdateLastLogin обновляет время последнего входа
func (r *PartnerRepository) UpdateLastLogin(id uuid.UUID) error {
	now := time.Now()
	return r.db.Model(&Partner{}).Where("id = ?", id).Update("last_login", &now).Error
}

// UpdateStatus обновляет статус партнера
func (r *PartnerRepository) UpdateStatus(id uuid.UUID, status string) error {
	return r.db.Model(&Partner{}).Where("id = ?", id).Update("status", status).Error
}

// Delete удаляет партнера
func (r *PartnerRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&Partner{}, id).Error
}

// DeleteWithCoupons удаляет партнера и все его купоны в транзакции
func (r *PartnerRepository) DeleteWithCoupons(id uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Получаем партнера для активации хука BeforeDelete
		var partner Partner
		if err := tx.First(&partner, id).Error; err != nil {
			return err
		}

		// Удаляем партнера (хук BeforeDelete автоматически удалит купоны)
		if err := tx.Delete(&partner).Error; err != nil {
			return err
		}

		return nil
	})
}

// Search выполняет поиск партнеров по различным критериям
func (r *PartnerRepository) Search(query string, status string) ([]*Partner, error) {
	db := r.db

	if query != "" {
		db = db.Where("brand_name ILIKE ? OR domain ILIKE ? OR email ILIKE ?",
			"%"+query+"%", "%"+query+"%", "%"+query+"%")
	}

	if status != "" {
		db = db.Where("status = ?", status)
	}

	var partners []*Partner
	err := db.Find(&partners).Error
	return partners, err
}

// GetPartnerCouponsForExport возвращает купоны партнера с данными для экспорта
func (r *PartnerRepository) GetPartnerCouponsForExport(partnerID uuid.UUID, status string) ([]*ExportCouponRequest, error) {
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
			c.used_at
		FROM coupons c
		JOIN partners p ON c.partner_id = p.id
		WHERE c.partner_id = ?
	`

	args := []interface{}{partnerID}

	if status != "" {
		query += " AND c.status = ?"
		args = append(args, status)
	}

	query += " ORDER BY c.created_at DESC"

	err := r.db.Raw(query, args...).Scan(&coupons).Error
	return coupons, err
}

// GetAllCouponsForExport возвращает все купоны с данными партнеров для экспорта админом
func (r *PartnerRepository) GetAllCouponsForExport() ([]*ExportCouponRequest, error) {
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
			c.used_at
		FROM coupons c
		JOIN partners p ON c.partner_id = p.id
		ORDER BY p.id, c.created_at DESC
	`

	err := r.db.Raw(query).Scan(&coupons).Error
	return coupons, err
}

// GetCouponsStatistics возвращает статистику купонов партнера
func (r *PartnerRepository) GetCouponsStatistics(partnerID uuid.UUID) (map[string]int64, error) {
	stats := make(map[string]int64)

	// Общее количество купонов
	var total int64
	err := r.db.Table("coupons").Where("partner_id = ?", partnerID).Count(&total).Error
	if err != nil {
		return nil, err
	}
	stats["total"] = total

	// Активированные купоны
	var activated int64
	err = r.db.Table("coupons").Where("partner_id = ? AND status = ?", partnerID, "used").Count(&activated).Error
	if err != nil {
		return nil, err
	}
	stats["activated"] = activated

	// Новые купоны
	var new int64
	err = r.db.Table("coupons").Where("partner_id = ? AND status = ?", partnerID, "new").Count(&new).Error
	if err != nil {
		return nil, err
	}
	stats["new"] = new

	// Купленные онлайн
	var purchased int64
	err = r.db.Table("coupons").Where("partner_id = ? AND is_purchased = ?", partnerID, true).Count(&purchased).Error
	if err != nil {
		return nil, err
	}
	stats["purchased"] = purchased

	return stats, nil
}

// UpdatePassword обновляет пароль партнера
func (r *PartnerRepository) UpdatePassword(partnerID uuid.UUID, hashedPassword string) error {
	return r.db.Model(&Partner{}).Where("id = ?", partnerID).Update("password", hashedPassword).Error
}

// GetNextPartnerCode возвращает следующий доступный код партнера (начиная с 0001)
func (r *PartnerRepository) GetNextPartnerCode() (string, error) {
	var maxCode string
	err := r.db.Model(&Partner{}).Select("COALESCE(MAX(CAST(partner_code AS INTEGER)), 0)").Scan(&maxCode).Error
	if err != nil {
		return "", err
	}

	// Если максимальный код 0 или меньше, начинаем с 1
	// Иначе инкрементируем на 1
	// Код 0000 зарезервирован для собственных купонов
	var nextCode int
	if maxCode == "" || maxCode == "0" {
		nextCode = 1
	} else {
		// Парсим строку в число
		if _, err := fmt.Sscanf(maxCode, "%d", &nextCode); err != nil {
			return "", err
		}
		nextCode++
	}

	// Форматируем как 4-значную строку с ведущими нулями
	return fmt.Sprintf("%04d", nextCode), nil
}
