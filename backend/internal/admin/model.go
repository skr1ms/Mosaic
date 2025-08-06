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

// ProfileChangeLog представляет лог изменений профиля партнера
type ProfileChangeLog struct {
	bun.BaseModel `bun:"table:profile_changes,alias:pc"`

	ID        uuid.UUID `bun:"id,pk,type:uuid,default:gen_random_uuid()" json:"id"`
	PartnerID uuid.UUID `bun:"partner_id,notnull,type:uuid" json:"partner_id"`
	Field     string    `bun:"field,notnull" json:"field"`
	OldValue  string    `bun:"old_value" json:"old_value"`
	NewValue  string    `bun:"new_value" json:"new_value"`
	ChangedBy string    `bun:"changed_by,notnull" json:"changed_by"` // admin login
	ChangedAt time.Time `bun:"changed_at,nullzero,notnull,default:current_timestamp" json:"changed_at"`
	Reason    string    `bun:"reason" json:"reason"`
}

// UserFilterDB представляет сохраненный пользовательский фильтр в базе данных
type UserFilterDB struct {
	bun.BaseModel `bun:"table:user_filters,alias:uf"`

	ID          uuid.UUID `bun:"id,pk,type:uuid,default:gen_random_uuid()" json:"id"`
	AdminID     uuid.UUID `bun:"admin_id,notnull,type:uuid" json:"admin_id"`
	Name        string    `bun:"name,notnull" json:"name"`
	Description string    `bun:"description" json:"description"`
	FilterType  string    `bun:"filter_type,notnull" json:"filter_type"` // "coupons", "partners"
	FilterData  string    `bun:"filter_data,notnull" json:"filter_data"` // JSON с параметрами фильтра
	IsDefault   bool      `bun:"is_default,default:false" json:"is_default"`
	CreatedAt   time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt   time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`
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

// CreateProfileChangeLogIndex создает индексы для таблицы изменений профиля
func (pc *ProfileChangeLog) CreateIndex() string {
	return `
	CREATE INDEX IF NOT EXISTS idx_profile_changes_partner_id ON profile_changes(partner_id);
	CREATE INDEX IF NOT EXISTS idx_profile_changes_changed_at ON profile_changes(changed_at);
	CREATE INDEX IF NOT EXISTS idx_profile_changes_field ON profile_changes(field);
	CREATE INDEX IF NOT EXISTS idx_profile_changes_changed_by ON profile_changes(changed_by);
	`
}

// CreateUserFilterIndex создает индексы для таблицы пользовательских фильтров
func (uf *UserFilterDB) CreateIndex() string {
	return `
	CREATE INDEX IF NOT EXISTS idx_user_filters_admin_id ON user_filters(admin_id);
	CREATE INDEX IF NOT EXISTS idx_user_filters_filter_type ON user_filters(filter_type);
	CREATE INDEX IF NOT EXISTS idx_user_filters_is_default ON user_filters(is_default);
	CREATE INDEX IF NOT EXISTS idx_user_filters_created_at ON user_filters(created_at);
	`
}
