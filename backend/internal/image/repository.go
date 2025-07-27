package image

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ImageRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *ImageRepository {
	return &ImageRepository{db: db}
}

// Create добавляет новую задачу в очередь обработки
func (r *ImageRepository) Create(task *Image) error {
	return r.db.Create(task).Error
}

// GetByID возвращает задачу по ID
func (r *ImageRepository) GetByID(id uuid.UUID) (*Image, error) {
	var task Image
	err := r.db.Where("id = ?", id).First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// GetByCouponID возвращает задачу по ID купона
func (r *ImageRepository) GetByCouponID(couponID uuid.UUID) (*Image, error) {
	var task Image
	err := r.db.Where("coupon_id = ?", couponID).First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// GetNextInQueue возвращает следующую задачу в очереди для обработки
func (r *ImageRepository) GetNextInQueue() (*Image, error) {
	var task Image
	err := r.db.Where("status = ?", "queued").
		Order("priority DESC, created_at ASC").
		First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// GetQueuedTasks возвращает все задачи в очереди
func (r *ImageRepository) GetQueuedTasks() ([]*Image, error) {
	var tasks []*Image
	err := r.db.Where("status = ?", "queued").
		Order("priority DESC, created_at ASC").
		Find(&tasks).Error
	return tasks, err
}

// GetProcessingTasks возвращает все задачи в процессе обработки
func (r *ImageRepository) GetProcessingTasks() ([]*Image, error) {
	var tasks []*Image
	err := r.db.Where("status = ?", "processing").Find(&tasks).Error
	return tasks, err
}

// StartProcessing помечает задачу как обрабатываемую
func (r *ImageRepository) StartProcessing(id uuid.UUID) error {
	now := time.Now()
	return r.db.Model(&Image{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":     "processing",
		"started_at": &now,
	}).Error
}

// CompleteProcessing помечает задачу как завершенную
func (r *ImageRepository) CompleteProcessing(id uuid.UUID) error {
	now := time.Now()
	return r.db.Model(&Image{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":       "completed",
		"completed_at": &now,
	}).Error
}

// FailProcessing помечает задачу как неудачную
func (r *ImageRepository) FailProcessing(id uuid.UUID, errorMessage string) error {
	now := time.Now()
	return r.db.Model(&Image{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":        "failed",
		"error_message": errorMessage,
		"completed_at":  &now,
	}).Error
}

// RetryTask увеличивает счетчик попыток и возвращает задачу в очередь
func (r *ImageRepository) RetryTask(id uuid.UUID) error {
	return r.db.Model(&Image{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":        "queued",
		"retry_count":   gorm.Expr("retry_count + 1"),
		"started_at":    nil,
		"error_message": nil,
	}).Error
}

// Update обновляет задачу
func (r *ImageRepository) Update(task *Image) error {
	return r.db.Save(task).Error
}

// Delete удаляет задачу
func (r *ImageRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&Image{}, id).Error
}

// GetAll возвращает все задачи
func (r *ImageRepository) GetAll() ([]*Image, error) {
	var tasks []*Image
	err := r.db.Find(&tasks).Error
	return tasks, err
}

// GetByStatus возвращает задачи по статусу
func (r *ImageRepository) GetByStatus(status string) ([]*Image, error) {
	var tasks []*Image
	err := r.db.Where("status = ?", status).Find(&tasks).Error
	return tasks, err
}

// GetFailedTasksForRetry возвращает неудачные задачи, которые можно повторить
func (r *ImageRepository) GetFailedTasksForRetry() ([]*Image, error) {
	var tasks []*Image
	err := r.db.Where("status = ? AND retry_count < max_retries", "failed").Find(&tasks).Error
	return tasks, err
}

// GetStatistics возвращает статистику по обработке изображений
func (r *ImageRepository) GetStatistics() (map[string]int64, error) {
	stats := make(map[string]int64)

	var queued, processing, completed, failed int64

	r.db.Model(&Image{}).Where("status = ?", "queued").Count(&queued)
	stats["queued"] = queued

	r.db.Model(&Image{}).Where("status = ?", "processing").Count(&processing)
	stats["processing"] = processing

	r.db.Model(&Image{}).Where("status = ?", "completed").Count(&completed)
	stats["completed"] = completed

	r.db.Model(&Image{}).Where("status = ?", "failed").Count(&failed)
	stats["failed"] = failed

	return stats, nil
}
