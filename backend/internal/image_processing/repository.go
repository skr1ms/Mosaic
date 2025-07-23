package image_processing

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ImageProcessingRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *ImageProcessingRepository {
	return &ImageProcessingRepository{db: db}
}

// Create добавляет новую задачу в очередь обработки
func (r *ImageProcessingRepository) Create(task *ImageProcessingQueue) error {
	return r.db.Create(task).Error
}

// GetByID возвращает задачу по ID
func (r *ImageProcessingRepository) GetByID(id uuid.UUID) (*ImageProcessingQueue, error) {
	var task ImageProcessingQueue
	err := r.db.Where("id = ?", id).First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// GetByCouponID возвращает задачу по ID купона
func (r *ImageProcessingRepository) GetByCouponID(couponID uuid.UUID) (*ImageProcessingQueue, error) {
	var task ImageProcessingQueue
	err := r.db.Where("coupon_id = ?", couponID).First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// GetNextInQueue возвращает следующую задачу в очереди для обработки
func (r *ImageProcessingRepository) GetNextInQueue() (*ImageProcessingQueue, error) {
	var task ImageProcessingQueue
	err := r.db.Where("status = ?", "queued").
		Order("priority DESC, created_at ASC").
		First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// GetQueuedTasks возвращает все задачи в очереди
func (r *ImageProcessingRepository) GetQueuedTasks() ([]*ImageProcessingQueue, error) {
	var tasks []*ImageProcessingQueue
	err := r.db.Where("status = ?", "queued").
		Order("priority DESC, created_at ASC").
		Find(&tasks).Error
	return tasks, err
}

// GetProcessingTasks возвращает все задачи в процессе обработки
func (r *ImageProcessingRepository) GetProcessingTasks() ([]*ImageProcessingQueue, error) {
	var tasks []*ImageProcessingQueue
	err := r.db.Where("status = ?", "processing").Find(&tasks).Error
	return tasks, err
}

// StartProcessing помечает задачу как обрабатываемую
func (r *ImageProcessingRepository) StartProcessing(id uuid.UUID) error {
	now := time.Now()
	return r.db.Model(&ImageProcessingQueue{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":     "processing",
		"started_at": &now,
	}).Error
}

// CompleteProcessing помечает задачу как завершенную
func (r *ImageProcessingRepository) CompleteProcessing(id uuid.UUID) error {
	now := time.Now()
	return r.db.Model(&ImageProcessingQueue{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":       "completed",
		"completed_at": &now,
	}).Error
}

// FailProcessing помечает задачу как неудачную
func (r *ImageProcessingRepository) FailProcessing(id uuid.UUID, errorMessage string) error {
	now := time.Now()
	return r.db.Model(&ImageProcessingQueue{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":        "failed",
		"error_message": errorMessage,
		"completed_at":  &now,
	}).Error
}

// RetryTask увеличивает счетчик попыток и возвращает задачу в очередь
func (r *ImageProcessingRepository) RetryTask(id uuid.UUID) error {
	return r.db.Model(&ImageProcessingQueue{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":        "queued",
		"retry_count":   gorm.Expr("retry_count + 1"),
		"started_at":    nil,
		"error_message": nil,
	}).Error
}

// Update обновляет задачу
func (r *ImageProcessingRepository) Update(task *ImageProcessingQueue) error {
	return r.db.Save(task).Error
}

// Delete удаляет задачу
func (r *ImageProcessingRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&ImageProcessingQueue{}, id).Error
}

// GetAll возвращает все задачи
func (r *ImageProcessingRepository) GetAll() ([]*ImageProcessingQueue, error) {
	var tasks []*ImageProcessingQueue
	err := r.db.Find(&tasks).Error
	return tasks, err
}

// GetByStatus возвращает задачи по статусу
func (r *ImageProcessingRepository) GetByStatus(status string) ([]*ImageProcessingQueue, error) {
	var tasks []*ImageProcessingQueue
	err := r.db.Where("status = ?", status).Find(&tasks).Error
	return tasks, err
}

// GetFailedTasksForRetry возвращает неудачные задачи, которые можно повторить
func (r *ImageProcessingRepository) GetFailedTasksForRetry() ([]*ImageProcessingQueue, error) {
	var tasks []*ImageProcessingQueue
	err := r.db.Where("status = ? AND retry_count < max_retries", "failed").Find(&tasks).Error
	return tasks, err
}

// GetStatistics возвращает статистику по обработке изображений
func (r *ImageProcessingRepository) GetStatistics() (map[string]int64, error) {
	stats := make(map[string]int64)

	var queued, processing, completed, failed int64

	r.db.Model(&ImageProcessingQueue{}).Where("status = ?", "queued").Count(&queued)
	stats["queued"] = queued

	r.db.Model(&ImageProcessingQueue{}).Where("status = ?", "processing").Count(&processing)
	stats["processing"] = processing

	r.db.Model(&ImageProcessingQueue{}).Where("status = ?", "completed").Count(&completed)
	stats["completed"] = completed

	r.db.Model(&ImageProcessingQueue{}).Where("status = ?", "failed").Count(&failed)
	stats["failed"] = failed

	return stats, nil
}
