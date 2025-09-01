package image

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type ImageRepository struct {
	db *bun.DB
}

func NewRepository(db *bun.DB) *ImageRepository {
	return &ImageRepository{db: db}
}

func (r *ImageRepository) Create(ctx context.Context, task *Image) error {
	_, err := r.db.NewInsert().Model(task).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}
	return nil
}

func (r *ImageRepository) GetByID(ctx context.Context, id uuid.UUID) (*Image, error) {
	task := new(Image)
	err := r.db.NewSelect().Model(task).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find task by ID: %w", err)
	}
	return task, nil
}

func (r *ImageRepository) GetByCouponID(ctx context.Context, couponID uuid.UUID) (*Image, error) {
	task := new(Image)
	err := r.db.NewSelect().Model(task).Where("coupon_id = ?", couponID).Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find coupon by ID: %w", err)
	}
	return task, nil
}

func (r *ImageRepository) GetNextInQueue(ctx context.Context) (*Image, error) {
	task := new(Image)
	err := r.db.NewSelect().Model(task).
		Where("status = ?", "queued").
		OrderExpr("priority DESC, created_at ASC").
		Limit(1).
		Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no tasks in queue")
		}
		return nil, fmt.Errorf("failed to find next in queue: %w", err)
	}
	return task, nil
}

func (r *ImageRepository) GetQueuedTasks(ctx context.Context) ([]*Image, error) {
	var tasks []*Image
	err := r.db.NewSelect().Model(&tasks).
		Where("status = ?", "queued").
		OrderExpr("priority DESC, created_at ASC").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find queued tasks: %w", err)
	}
	return tasks, nil
}

func (r *ImageRepository) GetProcessingTasks(ctx context.Context) ([]*Image, error) {
	var tasks []*Image
	err := r.db.NewSelect().Model(&tasks).
		Where("status = ?", "processing").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find processing tasks: %w", err)
	}
	return tasks, nil
}

func (r *ImageRepository) StartProcessing(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	_, err := r.db.NewUpdate().Model((*Image)(nil)).
		Set("status = ?", "processing").
		Set("started_at = ?", &now).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to start processing: %w", err)
	}
	return nil
}

func (r *ImageRepository) CompleteProcessing(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	_, err := r.db.NewUpdate().Model((*Image)(nil)).
		Set("status = ?", "completed").
		Set("completed_at = ?", &now).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to mark task as completed: %w", err)
	}
	return nil
}

func (r *ImageRepository) FailProcessing(ctx context.Context, id uuid.UUID, errorMessage string) error {
	now := time.Now()
	_, err := r.db.NewUpdate().Model((*Image)(nil)).
		Set("status = ?", "failed").
		Set("error_message = ?", errorMessage).
		Set("completed_at = ?", &now).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to mark task as failed: %w", err)
	}
	return nil
}

func (r *ImageRepository) RetryTask(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.NewUpdate().Model((*Image)(nil)).
		Set("status = ?", "queued").
		Set("retry_count = retry_count + 1").
		Set("started_at = NULL").
		Set("error_message = NULL").
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to retry task: %w", err)
	}
	return nil
}

func (r *ImageRepository) Update(ctx context.Context, task *Image) error {
	_, err := r.db.NewUpdate().Model(task).
		WherePK().Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}
	return nil
}

func (r *ImageRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.NewDelete().Model((*Image)(nil)).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}
	return nil
}

func (r *ImageRepository) GetAll(ctx context.Context) ([]*Image, error) {
	var tasks []*Image
	err := r.db.NewSelect().Model(&tasks).Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find all tasks: %w", err)
	}
	return tasks, nil
}

func (r *ImageRepository) GetByStatus(ctx context.Context, status string) ([]*Image, error) {
	var tasks []*Image
	err := r.db.NewSelect().Model(&tasks).Where("status = ?", status).Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find tasks by status: %w", err)
	}
	return tasks, nil
}

func (r *ImageRepository) GetAllWithPartner(ctx context.Context) ([]*ImageWithPartner, error) {
	var tasks []*ImageWithPartner
	err := r.db.NewSelect().
		Model((*Image)(nil)).
		ColumnExpr("i.*, partners.id as partner_id, partners.partner_code as partner_code").
		Join("JOIN coupons ON coupons.id = i.coupon_id").
		Join("JOIN partners ON partners.id = coupons.partner_id").
		Order("i.created_at DESC").
		Scan(ctx, &tasks)
	if err != nil {
		return nil, fmt.Errorf("failed to find all tasks with partner info: %w", err)
	}
	return tasks, nil
}

func (r *ImageRepository) GetByStatusWithPartner(ctx context.Context, status string) ([]*ImageWithPartner, error) {
	var tasks []*ImageWithPartner
	err := r.db.NewSelect().
		Model((*Image)(nil)).
		ColumnExpr("i.*, partners.id as partner_id, partners.partner_code as partner_code").
		Join("JOIN coupons ON coupons.id = i.coupon_id").
		Join("JOIN partners ON partners.id = coupons.partner_id").
		Where("i.status = ?", status).
		Order("i.created_at DESC").
		Scan(ctx, &tasks)
	if err != nil {
		return nil, fmt.Errorf("failed to find tasks by status with partner info: %w", err)
	}
	return tasks, nil
}

func (r *ImageRepository) GetWithFilters(ctx context.Context, status, dateFrom, dateTo string) ([]*ImageWithPartner, error) {
	var tasks []*ImageWithPartner

	query := r.db.NewSelect().
		Model((*Image)(nil)).
		ColumnExpr("i.*, partners.id as partner_id, partners.partner_code as partner_code").
		Join("JOIN coupons ON coupons.id = i.coupon_id").
		Join("JOIN partners ON partners.id = coupons.partner_id")

	// Status filter
	if status != "" && status != "all" {
		query = query.Where("i.status = ?", status)
	}

	// Date from filter
	if dateFrom != "" {
		query = query.Where("DATE(i.created_at) >= ?", dateFrom)
	}

	// Date to filter
	if dateTo != "" {
		query = query.Where("DATE(i.created_at) <= ?", dateTo)
	}

	query = query.Order("i.created_at DESC")

	err := query.Scan(ctx, &tasks)
	if err != nil {
		return nil, fmt.Errorf("failed to find tasks with filters: %w", err)
	}
	return tasks, nil
}

func (r *ImageRepository) GetFailedTasksForRetry(ctx context.Context) ([]*Image, error) {
	var tasks []*Image
	err := r.db.NewSelect().Model(&tasks).
		Where("status = ? AND retry_count < max_retries", "failed").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find failed tasks for retry: %w", err)
	}
	return tasks, nil
}

func (r *ImageRepository) GetStatistics(ctx context.Context) (map[string]int64, error) {
	stats := make(map[string]int64)

	count, err := r.db.NewSelect().Model((*Image)(nil)).Where("status = ?", "queued").Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get statistics: %w", err)
	}
	stats["queued"] = int64(count)

	count, err = r.db.NewSelect().Model((*Image)(nil)).Where("status = ?", "processing").Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get statistics: %w", err)
	}
	stats["processing"] = int64(count)

	count, err = r.db.NewSelect().Model((*Image)(nil)).Where("status = ?", "completed").Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get statistics: %w", err)
	}
	stats["completed"] = int64(count)

	count, err = r.db.NewSelect().Model((*Image)(nil)).Where("status = ?", "failed").Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get statistics: %w", err)
	}
	stats["failed"] = int64(count)

	return stats, nil
}
