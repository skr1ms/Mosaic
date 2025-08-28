package admin

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Admin struct {
	bun.BaseModel `bun:"table:admins,alias:a"`

	ID        uuid.UUID  `bun:"id,pk,type:uuid,default:gen_random_uuid()" json:"id" validate:"required"`
	Login     string     `bun:"login,unique,notnull" json:"login" validate:"required,secure_login"`
	Email     string     `bun:"email,unique,notnull" json:"email" validate:"required,email"`
	Password  string     `bun:"password,notnull" json:"password" validate:"required,secure_password"`
	Role      string     `bun:"role,notnull,default:'admin'" json:"role" validate:"required,oneof=main_admin admin"`
	LastLogin *time.Time `bun:"last_login" json:"last_login"`
	CreatedAt time.Time  `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
}

type ProfileChangeLog struct {
	bun.BaseModel `bun:"table:profile_changes,alias:pc"`

	ID        uuid.UUID `bun:"id,pk,type:uuid,default:gen_random_uuid()" json:"id"`
	PartnerID uuid.UUID `bun:"partner_id,notnull,type:uuid" json:"partner_id"`
	Field     string    `bun:"field,notnull" json:"field"`
	OldValue  string    `bun:"old_value" json:"old_value"`
	NewValue  string    `bun:"new_value" json:"new_value"`
	ChangedBy string    `bun:"changed_by,notnull" json:"changed_by"`
	ChangedAt time.Time `bun:"changed_at,nullzero,notnull,default:current_timestamp" json:"changed_at"`
	Reason    string    `bun:"reason" json:"reason"`
}

func (a *Admin) CreateIndex() string {
	return `
	CREATE INDEX IF NOT EXISTS idx_admins_login ON admins(login);
    CREATE INDEX IF NOT EXISTS idx_admins_email ON admins(email);
	CREATE INDEX IF NOT EXISTS idx_admins_last_login ON admins(last_login);
	CREATE INDEX IF NOT EXISTS idx_admins_created_at ON admins(created_at);
	`
}

func (pc *ProfileChangeLog) CreateIndex() string {
	return `
	CREATE INDEX IF NOT EXISTS idx_profile_changes_partner_id ON profile_changes(partner_id);
	CREATE INDEX IF NOT EXISTS idx_profile_changes_changed_at ON profile_changes(changed_at);
	CREATE INDEX IF NOT EXISTS idx_profile_changes_field ON profile_changes(field);
	CREATE INDEX IF NOT EXISTS idx_profile_changes_changed_by ON profile_changes(changed_by);
	`
}
