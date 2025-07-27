package admin

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AdminRepository struct {
	db *gorm.DB
}

func NewAdminRepository(db *gorm.DB) *AdminRepository {
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
		return nil, ErrFailedToFindAdminByLogin
	}
	return &admin, nil
}

// GetByID находит администратора по ID
func (r *AdminRepository) GetByID(id uuid.UUID) (*Admin, error) {
	var admin Admin
	err := r.db.Where("id = ?", id).First(&admin).Error
	if err != nil {
		return nil, ErrFailedToFindAdminByID
	}
	return &admin, nil
}

// GetAll возвращает всех администраторов
func (r *AdminRepository) GetAll() ([]*Admin, error) {
	var admins []*Admin
	err := r.db.Find(&admins).Error
	if err != nil {
		return nil, ErrFailedToFindAllAdmins
	}
	return admins, nil
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

// UpdatePassword обновляет пароль администратора
func (r *AdminRepository) UpdatePassword(id uuid.UUID, hashedPassword string) error {
	return r.db.Model(&Admin{}).Where("id = ?", id).Update("password", hashedPassword).Error
}