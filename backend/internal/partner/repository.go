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

// Create создаёт нового партнёра
func (r *PartnerRepository) Create(ctx context.Context, partner *Partner) error {
	_, err := r.db.NewInsert().Model(partner).Exec(ctx)
	if err != nil {
		return ErrFailedToCreatePartner
	}
	return nil
}

// GetByLogin находит партнёра по логину
func (r *PartnerRepository) GetByLogin(ctx context.Context, login string) (*Partner, error) {
	partner := new(Partner)
	err := r.db.NewSelect().Model(partner).Where("login = ?", login).Scan(ctx)
	if err != nil {
		return nil, ErrFailedToFindPartnerByLogin
	}
	return partner, nil
}

// GetByPartnerCode находит партнёра по коду
func (r *PartnerRepository) GetByPartnerCode(ctx context.Context, code string) (*Partner, error) {
	partner := new(Partner)
	err := r.db.NewSelect().Model(partner).Where("partner_code = ?", code).Scan(ctx)
	if err != nil {
		return nil, ErrFailedToFindPartnerByPartnerCode
	}
	return partner, nil
}

// GetByDomain находит партнёра по домену
func (r *PartnerRepository) GetByDomain(ctx context.Context, domain string) (*Partner, error) {
	partner := new(Partner)
	err := r.db.NewSelect().Model(partner).Where("domain = ?", domain).Scan(ctx)
	if err != nil {
		return nil, ErrFailedToFindPartnerByDomain
	}
	return partner, nil
}

// GetByID находит партнёра по ID
func (r *PartnerRepository) GetByID(ctx context.Context, id uuid.UUID) (*Partner, error) {
	partner := new(Partner)
	err := r.db.NewSelect().Model(partner).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return nil, ErrFailedToFindPartnerByID
	}
	return partner, nil
}

// GetByEmail находит партнёра по email
func (r *PartnerRepository) GetByEmail(ctx context.Context, email string) (*Partner, error) {
	partner := new(Partner)
	err := r.db.NewSelect().Model(partner).Where("email = ?", email).Scan(ctx)
	if err != nil {
		return nil, ErrFailedToFindPartnerByEmail
	}
	return partner, nil
}

// GetAll возвращает всех партнёров
func (r *PartnerRepository) GetAll(ctx context.Context) ([]*Partner, error) {
	var partners []*Partner
	err := r.db.NewSelect().Model(&partners).Scan(ctx)
	if err != nil {
		return nil, ErrFailedToFindAllPartners
	}
	return partners, nil
}

// GetActivePartners возвращает только активных партнёров
func (r *PartnerRepository) GetActivePartners(ctx context.Context) ([]*Partner, error) {
	var partners []*Partner
	err := r.db.NewSelect().Model(&partners).Where("status = ?", "active").Scan(ctx)
	if err != nil {
		return nil, ErrFailedToFindActivePartners
	}
	return partners, nil
}

// Update обновляет данные партнёра
func (r *PartnerRepository) Update(ctx context.Context, partner *Partner) error {
	_, err := r.db.NewUpdate().Model(partner).WherePK().Exec(ctx)
	if err != nil {
		return ErrFailedToUpdatePartner
	}
	return nil
}

// UpdateLastLogin обновляет время последнего входа
func (r *PartnerRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	_, err := r.db.NewUpdate().Model((*Partner)(nil)).Set("last_login = ?", &now).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return ErrFailedToUpdatePartnerLastLogin
	}
	return nil
}

// UpdateStatus обновляет статус партнёра
func (r *PartnerRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	_, err := r.db.NewUpdate().Model((*Partner)(nil)).Set("status = ?", status).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return ErrFailedToUpdatePartnerStatus
	}
	return nil
}

// Delete удаляет партнёра
func (r *PartnerRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.NewDelete().Model((*Partner)(nil)).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return ErrFailedToDeletePartner
	}
	return nil
}

// DeleteWithCoupons удаляет партнёра и все его купоны в транзакции
func (r *PartnerRepository) DeleteWithCoupons(ctx context.Context, id uuid.UUID) error {
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		// Проверяем, сколько купонов у партнера для логирования
		_, err := tx.NewSelect().Table("coupons").Where("partner_id = ?", id).Count(ctx)
		if err != nil {
			return ErrFailedToDeletePartner
		}

		// Удаляем партнера - купоны и изображения удалятся автоматически благодаря CASCADE
		_, err = tx.NewDelete().Model((*Partner)(nil)).Where("id = ?", id).Exec(ctx)
		if err != nil {
			return ErrFailedToDeletePartner
		}

		return nil
	})
}

// Search выполняет поиск партнёров по различным критериям
func (r *PartnerRepository) Search(ctx context.Context, queryStr string, status string) ([]*Partner, error) {
	query := r.db.NewSelect().Model((*Partner)(nil))

	if queryStr != "" {
		query = query.Where("brand_name ILIKE ? OR domain ILIKE ? OR email ILIKE ?", "%"+queryStr+"%", "%"+queryStr+"%", "%"+queryStr+"%")
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var partners []*Partner
	err := query.Scan(ctx, &partners)
	if err != nil {
		return nil, ErrFailedToFindPartners
	}
	return partners, nil
}

// GetPartnerCouponsForExport возвращает купоны партнёра с данными для экспорта
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
			c.used_at
		FROM coupons c
		JOIN partners p ON c.partner_id = p.id
		WHERE c.partner_id = ?`
	args := []interface{}{partnerID}
	if status != "" {
		query += " AND c.status = ?"
		args = append(args, status)
	}
	query += " ORDER BY c.created_at DESC"
	err := r.db.NewRaw(query, args...).Scan(ctx, &coupons)
	if err != nil {
		return nil, ErrFailedToFindPartnerCouponsForExport
	}
	return coupons, nil
}

// GetAllCouponsForExport возвращает все купоны с данными партнёров для экспорта админом
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
			c.used_at
		FROM coupons c
		JOIN partners p ON c.partner_id = p.id
		ORDER BY p.id, c.created_at DESC`
	err := r.db.NewRaw(query).Scan(ctx, &coupons)
	if err != nil {
		return nil, ErrFailedToGetAllCouponsForExport
	}
	return coupons, nil
}

// GetCouponsStatistics возвращает статистику купонов партнёра
func (r *PartnerRepository) GetCouponsStatistics(ctx context.Context, partnerID uuid.UUID) (map[string]int64, error) {
	stats := make(map[string]int64)

	var count int
	var err error

	count, err = r.db.NewSelect().Model(map[string]interface{}{}).Table("coupons").Where("partner_id = ?", partnerID).Count(ctx)
	if err != nil {
		return nil, ErrFailedToGetCouponsStatistics
	}
	stats["total"] = int64(count)

	count, err = r.db.NewSelect().Model(map[string]interface{}{}).Table("coupons").Where("partner_id = ? AND status = ?", partnerID, "used").Count(ctx)
	if err != nil {
		return nil, ErrFailedToGetCouponsStatistics
	}
	stats["activated"] = int64(count)

	count, err = r.db.NewSelect().Model(map[string]interface{}{}).Table("coupons").Where("partner_id = ? AND status = ?", partnerID, "new").Count(ctx)
	if err != nil {
		return nil, ErrFailedToGetCouponsStatistics
	}
	stats["new"] = int64(count)

	count, err = r.db.NewSelect().Model(map[string]interface{}{}).Table("coupons").Where("partner_id = ? AND is_purchased = ?", partnerID, true).Count(ctx)
	if err != nil {
		return nil, ErrFailedToGetCouponsStatistics
	}
	stats["purchased"] = int64(count)

	return stats, nil
}

// UpdatePassword обновляет пароль партнёра
func (r *PartnerRepository) UpdatePassword(ctx context.Context, partnerID uuid.UUID, hashedPassword string) error {
	_, err := r.db.NewUpdate().Model((*Partner)(nil)).Set("password = ?", hashedPassword).Where("id = ?", partnerID).Exec(ctx)
	if err != nil {
		return ErrFailedToUpdatePartnerPassword
	}
	return nil
}

// GetNextPartnerCode возвращает следующий доступный код партнёра (начиная с 0001)
func (r *PartnerRepository) GetNextPartnerCode(ctx context.Context) (string, error) {
	var maxCode string
	err := r.db.NewSelect().Model((*Partner)(nil)).ColumnExpr("COALESCE(MAX(CAST(partner_code AS INTEGER)), 0)").Scan(ctx, &maxCode)
	if err != nil {
		return "", ErrFailedToGetPartnerCode
	}
	var nextCode int
	if maxCode == "" || maxCode == "0" {
		nextCode = 1
	} else {
		if _, err := fmt.Sscanf(maxCode, "%d", &nextCode); err != nil {
			return "", ErrFailedToGetPartnerCode
		}
		nextCode++
	}
	return fmt.Sprintf("%04d", nextCode), nil
}
