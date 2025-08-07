package queue

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

type QueueManager struct {
	redis      *redis.Client
	queues     map[string]*TaskQueue
	imageQueue *ImageTaskQueue
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
}

func NewQueueManager(redis *redis.Client) *QueueManager {
	ctx, cancel := context.WithCancel(context.Background())

	return &QueueManager{
		redis:      redis,
		queues:     make(map[string]*TaskQueue),
		imageQueue: NewImageTaskQueue(redis),
		ctx:        ctx,
		cancel:     cancel,
	}
}

// GetImageQueue возвращает очередь для обработки изображений
func (qm *QueueManager) GetImageQueue() *ImageTaskQueue {
	return qm.imageQueue
}

// CreateQueue создает новую очередь с указанным именем
func (qm *QueueManager) CreateQueue(name string) *TaskQueue {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	if queue, exists := qm.queues[name]; exists {
		return queue
	}

	queue := NewTaskQueue(qm.redis, name)
	qm.queues[name] = queue

	log.Info().Str("queue_name", name).Msg("Queue created")
	return queue
}

// GetQueue возвращает очередь по имени
func (qm *QueueManager) GetQueue(name string) *TaskQueue {
	qm.mu.RLock()
	defer qm.mu.RUnlock()

	return qm.queues[name]
}

// StartAllWorkers запускает воркеры для всех очередей
func (qm *QueueManager) StartAllWorkers(imageAdapter *ImageServiceAdapter, emailAdapter *EmailServiceAdapter) {
	// Запускаем воркер для очереди изображений
	imageHandlers := GetImageTaskHandlers(imageAdapter, emailAdapter)
	qm.imageQueue.StartWorker(imageHandlers)

	// Запускаем воркеры для других очередей
	qm.mu.RLock()
	for name, queue := range qm.queues {
		log.Info().Str("queue_name", name).Msg("Starting queue worker")
		// Здесь можно добавить обработчики для других типов задач
		queue.StartWorker(map[string]TaskHandler{})
	}
	qm.mu.RUnlock()

	log.Info().Msg("All queue workers started")
}

// StopAll останавливает все очереди
func (qm *QueueManager) StopAll() {
	qm.cancel()

	// Останавливаем очередь изображений
	qm.imageQueue.Stop()

	// Останавливаем другие очереди
	qm.mu.RLock()
	for name, queue := range qm.queues {
		log.Info().Str("queue_name", name).Msg("Stopping queue")
		queue.Stop()
	}
	qm.mu.RUnlock()

	log.Info().Msg("All queues stopped")
}

// GetStats возвращает статистику по всем очередям
func (qm *QueueManager) GetStats() map[string]QueueStats {
	stats := make(map[string]QueueStats)

	// Статистика для очереди изображений
	imageStats := qm.getQueueStats("images")
	stats["images"] = imageStats

	// Статистика для других очередей
	qm.mu.RLock()
	for name := range qm.queues {
		queueStats := qm.getQueueStats(name)
		stats[name] = queueStats
	}
	qm.mu.RUnlock()

	return stats
}

// QueueStats статистика очереди
type QueueStats struct {
	Name           string `json:"name"`
	PendingTasks   int64  `json:"pending_tasks"`
	DelayedTasks   int64  `json:"delayed_tasks"`
	CompletedTasks int64  `json:"completed_tasks"`
	FailedTasks    int64  `json:"failed_tasks"`
}

// getQueueStats получает статистику для конкретной очереди
func (qm *QueueManager) getQueueStats(queueName string) QueueStats {
	ctx := context.Background()

	stats := QueueStats{
		Name: queueName,
	}

	// Подсчитываем задачи в очереди (по всем приоритетам)
	for priority := 0; priority <= 10; priority++ {
		queueKey := fmt.Sprintf("queue:%s:priority:%d", queueName, priority)
		count, err := qm.redis.LLen(ctx, queueKey).Result()
		if err == nil {
			stats.PendingTasks += count
		}
	}

	// Подсчитываем отложенные задачи
	delayedKey := fmt.Sprintf("queue:%s:delayed", queueName)
	delayedCount, err := qm.redis.ZCard(ctx, delayedKey).Result()
	if err == nil {
		stats.DelayedTasks = delayedCount
	}

	// Подсчитываем завершенные задачи
	completedKey := fmt.Sprintf("queue:%s:completed", queueName)
	completedCount, err := qm.redis.ZCard(ctx, completedKey).Result()
	if err == nil {
		stats.CompletedTasks = completedCount
	}

	// Подсчитываем неудачные задачи
	failedKey := fmt.Sprintf("queue:%s:failed", queueName)
	failedCount, err := qm.redis.ZCard(ctx, failedKey).Result()
	if err == nil {
		stats.FailedTasks = failedCount
	}

	return stats
}

// CleanupOldTasks очищает старые завершенные и неудачные задачи
func (qm *QueueManager) CleanupOldTasks() error {
	ctx := context.Background()

	// Получаем все имена очередей
	queueNames := []string{"images"}

	qm.mu.RLock()
	for name := range qm.queues {
		queueNames = append(queueNames, name)
	}
	qm.mu.RUnlock()

	for _, queueName := range queueNames {
		// Удаляем завершенные задачи старше 24 часов
		completedKey := fmt.Sprintf("queue:%s:completed", queueName)
		yesterday := time.Now().Add(-24 * time.Hour).Unix()

		removedCompleted, err := qm.redis.ZRemRangeByScore(ctx, completedKey, "-inf", fmt.Sprintf("%d", yesterday)).Result()
		if err != nil {
			log.Error().Err(err).Str("queue", queueName).Msg("Failed to cleanup completed tasks")
		} else if removedCompleted > 0 {
			log.Info().Str("queue", queueName).Int64("removed", removedCompleted).Msg("Cleaned up completed tasks")
		}

		// Удаляем неудачные задачи старше 7 дней
		failedKey := fmt.Sprintf("queue:%s:failed", queueName)
		weekAgo := time.Now().Add(-7 * 24 * time.Hour).Unix()

		removedFailed, err := qm.redis.ZRemRangeByScore(ctx, failedKey, "-inf", fmt.Sprintf("%d", weekAgo)).Result()
		if err != nil {
			log.Error().Err(err).Str("queue", queueName).Msg("Failed to cleanup failed tasks")
		} else if removedFailed > 0 {
			log.Info().Str("queue", queueName).Int64("removed", removedFailed).Msg("Cleaned up failed tasks")
		}
	}

	return nil
}
