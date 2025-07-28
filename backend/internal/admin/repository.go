package admin

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type AdminRepository struct {
	db *bun.DB
}

func NewAdminRepository(db *bun.DB) *AdminRepository {
	return &AdminRepository{db: db}
}

func (r *AdminRepository) Create(admin *Admin) error {
	ctx := context.Background()
	_, err := r.db.NewInsert().Model(admin).Exec(ctx)
	if err != nil {
		return ErrFailedToCreateAdmin
	}
	return nil
}

// GetByLogin находит администратора по логину
func (r *AdminRepository) GetByLogin(login string) (*Admin, error) {
	ctx := context.Background()
	admin := new(Admin)
	err := r.db.NewSelect().Model(admin).Where("login = ?", login).Scan(ctx)
	if err != nil {
		return nil, ErrFailedToFindAdminByLogin
	}
	return admin, nil
}

// GetByID находит администратора по ID
func (r *AdminRepository) GetByID(id uuid.UUID) (*Admin, error) {
	ctx := context.Background()
	admin := new(Admin)
	err := r.db.NewSelect().Model(admin).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return nil, ErrFailedToFindAdminByID
	}
	return admin, nil
}

// GetAll возвращает всех администраторов
func (r *AdminRepository) GetAll() ([]*Admin, error) {
	ctx := context.Background()
	var admins []*Admin
	err := r.db.NewSelect().Model(&admins).Scan(ctx)
	if err != nil {
		return nil, ErrFailedToFindAllAdmins
	}
	return admins, nil
}

// UpdateLastLogin обновляет время последнего входа
func (r *AdminRepository) UpdateLastLogin(id uuid.UUID) error {
	ctx := context.Background()
	now := time.Now()
	_, err := r.db.NewUpdate().Model((*Admin)(nil)).Set("last_login = ?", &now).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return ErrFailedToUpdateLastLogin
	}
	return nil
}

// Delete удаляет администратора
func (r *AdminRepository) Delete(id uuid.UUID) error {
	ctx := context.Background()
	_, err := r.db.NewDelete().Model((*Admin)(nil)).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return ErrFailedToDeleteAdmin
	}
	return nil
}

// UpdatePassword обновляет пароль администратора
func (r *AdminRepository) UpdatePassword(id uuid.UUID, hashedPassword string) error {
	ctx := context.Background()
	_, err := r.db.NewUpdate().Model((*Admin)(nil)).Set("password = ?", hashedPassword).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return ErrFailedToUpdatePassword
	}
	return nil
}
