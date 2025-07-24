package partner

import (
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
func (r *PartnerRepository) GetByPartnerCode(code int16) (*Partner, error) {
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
