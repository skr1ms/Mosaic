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
	Password  string     `bun:"password,notnull" json:"password" validate:"required,secure_password"`
	LastLogin *time.Time `bun:"last_login" json:"last_login"`
	CreatedAt time.Time  `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
}

// - idx_admins_login: быстрый поиск по логину
// - idx_admins_last_login: аналитика по датам последнего входа
// - idx_admins_created_at: сортировка по дате создания

func (a *Admin) CreateIndex() string {
	return `
	CREATE INDEX IF NOT EXISTS idx_admins_login ON admins(login);
	CREATE INDEX IF NOT EXISTS idx_admins_last_login ON admins(last_login);
	CREATE INDEX IF NOT EXISTS idx_admins_created_at ON admins(created_at);
	`
}
