package admin

import (
	"context"
	"fmt"
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
		return fmt.Errorf("failed to create admin: %w", err)
	}
	return nil
}

// GetByLogin находит администратора по логину
func (r *AdminRepository) GetByLogin(login string) (*Admin, error) {
	ctx := context.Background()
	admin := new(Admin)
	err := r.db.NewSelect().Model(admin).Where("login = ?", login).Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find admin by login: %w", err)
	}
	return admin, nil
}

// GetByID находит администратора по ID
func (r *AdminRepository) GetByID(id uuid.UUID) (*Admin, error) {
	ctx := context.Background()
	admin := new(Admin)
	err := r.db.NewSelect().Model(admin).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find admin by id: %w", err)
	}
	return admin, nil
}

// GetAll возвращает всех администраторов
func (r *AdminRepository) GetAll() ([]*Admin, error) {
	ctx := context.Background()
	var admins []*Admin
	err := r.db.NewSelect().Model(&admins).Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find all admins: %w", err)
	}
	return admins, nil
}

// UpdateLastLogin обновляет время последнего входа
func (r *AdminRepository) UpdateLastLogin(id uuid.UUID) error {
	ctx := context.Background()
	now := time.Now()
	_, err := r.db.NewUpdate().Model((*Admin)(nil)).Set("last_login = ?", &now).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}
	return nil
}

// Delete удаляет администратора
func (r *AdminRepository) Delete(id uuid.UUID) error {
	ctx := context.Background()
	_, err := r.db.NewDelete().Model((*Admin)(nil)).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete admin: %w", err)
	}
	return nil
}

// UpdatePassword обновляет пароль администратора
func (r *AdminRepository) UpdatePassword(id uuid.UUID, hashedPassword string) error {
	ctx := context.Background()
	_, err := r.db.NewUpdate().Model((*Admin)(nil)).Set("password = ?", hashedPassword).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}
	return nil
}

// CreateProfileChangeLog создает запись об изменении профиля партнера
func (r *AdminRepository) CreateProfileChangeLog(log *ProfileChangeLog) error {
	ctx := context.Background()
	_, err := r.db.NewInsert().Model(log).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to create profile change log: %w", err)
	}
	return nil
}

// GetProfileChangesByPartnerID возвращает историю изменений профиля партнера
func (r *AdminRepository) GetProfileChangesByPartnerID(partnerID uuid.UUID) ([]*ProfileChangeLog, error) {
	ctx := context.Background()
	var changes []*ProfileChangeLog
	err := r.db.NewSelect().Model(&changes).
		Where("partner_id = ?", partnerID).
		Order("changed_at DESC").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile changes: %w", err)
	}
	return changes, nil
}

// CreateUserFilter создает новый пользовательский фильтр
func (r *AdminRepository) CreateUserFilter(filter *UserFilterDB) error {
	ctx := context.Background()
	_, err := r.db.NewInsert().Model(filter).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to create user filter: %w", err)
	}
	return nil
}

// GetUserFiltersByAdminID возвращает все фильтры администратора
func (r *AdminRepository) GetUserFiltersByAdminID(adminID uuid.UUID, filterType string) ([]*UserFilterDB, error) {
	ctx := context.Background()
	var filters []*UserFilterDB
	query := r.db.NewSelect().Model(&filters).Where("admin_id = ?", adminID)

	if filterType != "" {
		query = query.Where("filter_type = ?", filterType)
	}

	err := query.Order("is_default DESC, created_at DESC").Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user filters: %w", err)
	}
	return filters, nil
}

// GetUserFilterByID возвращает фильтр по ID
func (r *AdminRepository) GetUserFilterByID(filterID uuid.UUID) (*UserFilterDB, error) {
	ctx := context.Background()
	filter := new(UserFilterDB)
	err := r.db.NewSelect().Model(filter).Where("id = ?", filterID).Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user filter: %w", err)
	}
	return filter, nil
}

// UpdateUserFilter обновляет пользовательский фильтр
func (r *AdminRepository) UpdateUserFilter(filter *UserFilterDB) error {
	ctx := context.Background()
	filter.UpdatedAt = time.Now()
	_, err := r.db.NewUpdate().Model(filter).WherePK().Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to update user filter: %w", err)
	}
	return nil
}

// DeleteUserFilter удаляет пользовательский фильтр
func (r *AdminRepository) DeleteUserFilter(filterID uuid.UUID) error {
	ctx := context.Background()
	_, err := r.db.NewDelete().Model((*UserFilterDB)(nil)).Where("id = ?", filterID).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete user filter: %w", err)
	}
	return nil
}
