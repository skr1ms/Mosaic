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

// Create добавляет новую задачу в очередь обработки
func (r *ImageRepository) Create(ctx context.Context, task *Image) error {
	_, err := r.db.NewInsert().Model(task).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}
	return nil
}

// GetByID возвращает задачу по ID
func (r *ImageRepository) GetByID(ctx context.Context, id uuid.UUID) (*Image, error) {
	task := new(Image)
	err := r.db.NewSelect().Model(task).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find task by ID: %w", err)
	}
	return task, nil
}

// GetByCouponID возвращает задачу по ID купона
func (r *ImageRepository) GetByCouponID(ctx context.Context, couponID uuid.UUID) (*Image, error) {
	task := new(Image)
	err := r.db.NewSelect().Model(task).Where("coupon_id = ?", couponID).Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find coupon by ID: %w", err)
	}
	return task, nil
}

// GetNextInQueue возвращает следующую задачу в очереди для обработки
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

// GetQueuedTasks возвращает все задачи в очереди
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

// GetProcessingTasks возвращает все задачи в процессе обработки
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

// StartProcessing помечает задачу как обрабатываемую
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

// CompleteProcessing помечает задачу как завершённую
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

// FailProcessing помечает задачу как неудачную
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

// RetryTask увеличивает счётчик попыток и возвращает задачу в очередь
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

// Update обновляет задачу
func (r *ImageRepository) Update(ctx context.Context, task *Image) error {
	_, err := r.db.NewUpdate().Model(task).
		WherePK().Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}
	return nil
}

// Delete удаляет задачу
func (r *ImageRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.NewDelete().Model((*Image)(nil)).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}
	return nil
}

// GetAll возвращает все задачи
func (r *ImageRepository) GetAll(ctx context.Context) ([]*Image, error) {
	var tasks []*Image
	err := r.db.NewSelect().Model(&tasks).Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find all tasks: %w", err)
	}
	return tasks, nil
}

// GetByStatus возвращает задачи по статусу
func (r *ImageRepository) GetByStatus(ctx context.Context, status string) ([]*Image, error) {
	var tasks []*Image
	err := r.db.NewSelect().Model(&tasks).Where("status = ?", status).Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find tasks by status: %w", err)
	}
	return tasks, nil
}

// GetFailedTasksForRetry возвращает неудачные задачи, которые можно повторить
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

// GetStatistics возвращает статистику по обработке изображений
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
