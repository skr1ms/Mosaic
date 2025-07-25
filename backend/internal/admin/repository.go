package admin

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AdminRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *AdminRepository {
	return &AdminRepository{db: db}
}

func (r *AdminRepository) Create(admin *Admin) error {
	return r.db.Create(admin).Error
}

// GetByLogin находит администратора по логину
func (r *AdminRepository) GetByLogin(login string) (*Admin, error) {
	var admin Admin
	err := r.db.Where("login = ?", login).First(&admin).Error
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

// GetByID находит администратора по ID
func (r *AdminRepository) GetByID(id uuid.UUID) (*Admin, error) {
	var admin Admin
	err := r.db.Where("id = ?", id).First(&admin).Error
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

// GetAll возвращает всех администраторов
func (r *AdminRepository) GetAll() ([]*Admin, error) {
	var admins []*Admin
	err := r.db.Find(&admins).Error
	return admins, err
}

// UpdateLastLogin обновляет время последнего входа
func (r *AdminRepository) UpdateLastLogin(id uuid.UUID) error {
	now := time.Now()
	return r.db.Model(&Admin{}).Where("id = ?", id).Update("last_login", &now).Error
}

// Delete удаляет администратора
func (r *AdminRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&Admin{}, id).Error
}
