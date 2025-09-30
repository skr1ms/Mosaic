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

func (r *AdminRepository) GetByLogin(login string) (*Admin, error) {
	ctx := context.Background()
	admin := new(Admin)
	err := r.db.NewSelect().Model(admin).Where("login = ?", login).Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find admin by login: %w", err)
	}
	return admin, nil
}

func (r *AdminRepository) GetByEmail(email string) (*Admin, error) {
	ctx := context.Background()
	admin := new(Admin)
	err := r.db.NewSelect().Model(admin).Where("email = ?", email).Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find admin by email: %w", err)
	}
	return admin, nil
}

func (r *AdminRepository) GetByID(id uuid.UUID) (*Admin, error) {
	ctx := context.Background()
	admin := new(Admin)
	err := r.db.NewSelect().Model(admin).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find admin by id: %w", err)
	}
	return admin, nil
}

func (r *AdminRepository) GetAll() ([]*Admin, error) {
	ctx := context.Background()
	var admins []*Admin
	err := r.db.NewSelect().Model(&admins).Order("created_at DESC").Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all admins: %w", err)
	}
	return admins, nil
}

func (r *AdminRepository) UpdateLastLogin(id uuid.UUID) error {
	ctx := context.Background()
	now := time.Now()
	_, err := r.db.NewUpdate().Model((*Admin)(nil)).Set("last_login = ?", &now).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}
	return nil
}

func (r *AdminRepository) Delete(id uuid.UUID) error {
	ctx := context.Background()
	_, err := r.db.NewDelete().Model((*Admin)(nil)).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete admin: %w", err)
	}
	return nil
}

func (r *AdminRepository) UpdatePassword(id uuid.UUID, hashedPassword string) error {
	ctx := context.Background()
	_, err := r.db.NewUpdate().Model((*Admin)(nil)).Set("password = ?", hashedPassword).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}
	return nil
}

func (r *AdminRepository) UpdateEmail(id uuid.UUID, newEmail string) error {
	ctx := context.Background()
	_, err := r.db.NewUpdate().Model((*Admin)(nil)).Set("email = ?", newEmail).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to update email: %w", err)
	}
	return nil
}

func (r *AdminRepository) CreateProfileChangeLog(log *ProfileChangeLog) error {
	ctx := context.Background()
	_, err := r.db.NewInsert().Model(log).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to create profile change log: %w", err)
	}
	return nil
}

func (r *AdminRepository) GetProfileChangesByPartnerID(partnerID uuid.UUID) ([]*ProfileChangeLog, error) {
	ctx := context.Background()
	var changes []*ProfileChangeLog
	err := r.db.NewSelect().Model(&changes).Where("partner_id = ?", partnerID).OrderExpr("changed_at DESC").Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile changes: %w", err)
	}
	return changes, nil
}
