package queue

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/skr1ms/mosaic/pkg/middleware"
)

type QueueManager struct {
	redis      *redis.Client
	queues     map[string]*TaskQueue
	imageQueue *ImageTaskQueue
	aiQueue    *ImageTaskQueue // Special queue for AI tasks
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
	logger     *middleware.Logger
}

func NewQueueManager(redis *redis.Client, logger *middleware.Logger) *QueueManager {
	ctx, cancel := context.WithCancel(context.Background())

	return &QueueManager{
		redis:      redis,
		queues:     make(map[string]*TaskQueue),
		imageQueue: NewImageTaskQueue(redis, logger),
		aiQueue:    NewImageTaskQueue(redis, logger),
		ctx:        ctx,
		cancel:     cancel,
		logger:     logger,
	}
}

// GetImageQueue returns queue for image processing
func (qm *QueueManager) GetImageQueue() *ImageTaskQueue {
	return qm.imageQueue
}

// GetAIQueue returns queue for AI image processing
func (qm *QueueManager) GetAIQueue() *ImageTaskQueue {
	return qm.aiQueue
}

// CreateQueue creates new queue with specified name
func (qm *QueueManager) CreateQueue(name string) *TaskQueue {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	if queue, exists := qm.queues[name]; exists {
		return queue
	}

	queue := NewTaskQueue(name, qm.redis, qm.logger)
	qm.queues[name] = queue

	qm.logger.GetZerologLogger().Info().Str("queue_name", name).Msg("Queue created")
	return queue
}

// GetQueue returns queue by name
func (qm *QueueManager) GetQueue(name string) *TaskQueue {
	qm.mu.RLock()
	defer qm.mu.RUnlock()

	return qm.queues[name]
}

// StartAllWorkers starts workers for all queues
func (qm *QueueManager) StartAllWorkers(imageAdapter *ImageServiceAdapter, emailAdapter *EmailServiceAdapter) {
	imageHandlers := GetImageTaskHandlers(imageAdapter, emailAdapter, qm.logger)
	qm.imageQueue.StartWorker(imageHandlers)

	qm.aiQueue.StartWorker(imageHandlers)

	qm.mu.RLock()
	for name, queue := range qm.queues {
		qm.logger.GetZerologLogger().Info().Str("queue_name", name).Msg("Starting queue worker")
		queue.StartWorker(map[string]TaskHandler{})
	}
	qm.mu.RUnlock()

	qm.logger.GetZerologLogger().Info().Msg("All queue workers started")
}

// StopAll stops all queues
func (qm *QueueManager) StopAll() {
	qm.cancel()

	qm.imageQueue.Close()

	qm.aiQueue.Close()

	// Stop other queues
	qm.mu.RLock()
	for name, queue := range qm.queues {
		qm.logger.GetZerologLogger().Info().Str("queue_name", name).Msg("Stopping queue")
		queue.Close()
	}
	qm.mu.RUnlock()

	qm.logger.GetZerologLogger().Info().Msg("All queues stopped")
}

// GetStats returns statistics for all queues
func (qm *QueueManager) GetStats() map[string]QueueStats {
	stats := make(map[string]QueueStats)

	// Statistics for image queue
	imageStats := qm.getQueueStats("images")
	stats["images"] = imageStats

	// Statistics for AI queue
	aiStats := qm.getQueueStats("ai_images")
	stats["ai_images"] = aiStats

	// Statistics for other queues
	qm.mu.RLock()
	for name := range qm.queues {
		queueStats := qm.getQueueStats(name)
		stats[name] = queueStats
	}
	qm.mu.RUnlock()

	return stats
}

// QueueStats queue statistics
type QueueStats struct {
	Name           string `json:"name"`
	PendingTasks   int64  `json:"pending_tasks"`
	DelayedTasks   int64  `json:"delayed_tasks"`
	CompletedTasks int64  `json:"completed_tasks"`
	FailedTasks    int64  `json:"failed_tasks"`
}

// getQueueStats gets statistics for specific queue
func (qm *QueueManager) getQueueStats(queueName string) QueueStats {
	ctx := context.Background()

	stats := QueueStats{
		Name: queueName,
	}

	// Count tasks in queue (by all priorities)
	for priority := 0; priority <= 10; priority++ {
		queueKey := fmt.Sprintf("queue:%s:priority:%d", queueName, priority)
		count, err := qm.redis.LLen(ctx, queueKey).Result()
		if err == nil {
			stats.PendingTasks += count
		}
	}

	// Count delayed tasks
	delayedKey := fmt.Sprintf("queue:%s:delayed", queueName)
	delayedCount, err := qm.redis.ZCard(ctx, delayedKey).Result()
	if err == nil {
		stats.DelayedTasks = delayedCount
	}

	// Count completed tasks
	completedKey := fmt.Sprintf("queue:%s:completed", queueName)
	completedCount, err := qm.redis.ZCard(ctx, completedKey).Result()
	if err == nil {
		stats.CompletedTasks = completedCount
	}

	// Count failed tasks
	failedKey := fmt.Sprintf("queue:%s:failed", queueName)
	failedCount, err := qm.redis.ZCard(ctx, failedKey).Result()
	if err == nil {
		stats.FailedTasks = failedCount
	}

	return stats
}

// CleanupOldTasks cleans old completed and failed tasks
func (qm *QueueManager) CleanupOldTasks() error {
	ctx := context.Background()

	// Get all queue names
	queueNames := []string{"images", "ai_images"}

	qm.mu.RLock()
	for name := range qm.queues {
		queueNames = append(queueNames, name)
	}
	qm.mu.RUnlock()

	for _, queueName := range queueNames {
		// Remove completed tasks older than 24 hours
		completedKey := fmt.Sprintf("queue:%s:completed", queueName)
		yesterday := time.Now().Add(-24 * time.Hour).Unix()

		removedCompleted, err := qm.redis.ZRemRangeByScore(ctx, completedKey, "-inf", fmt.Sprintf("%d", yesterday)).Result()
		if err != nil {
			qm.logger.GetZerologLogger().Error().Err(err).Str("queue", queueName).Msg("Failed to cleanup completed tasks")
		} else if removedCompleted > 0 {
			qm.logger.GetZerologLogger().Info().Str("queue", queueName).Int64("removed", removedCompleted).Msg("Cleaned up completed tasks")
		}

		// Remove failed tasks older than 7 days
		failedKey := fmt.Sprintf("queue:%s:failed", queueName)
		weekAgo := time.Now().Add(-7 * 24 * time.Hour).Unix()

		removedFailed, err := qm.redis.ZRemRangeByScore(ctx, failedKey, "-inf", fmt.Sprintf("%d", weekAgo)).Result()
		if err != nil {
			qm.logger.GetZerologLogger().Error().Err(err).Str("queue", queueName).Msg("Failed to cleanup failed tasks")
		} else if removedFailed > 0 {
			qm.logger.GetZerologLogger().Info().Str("queue", queueName).Int64("removed", removedFailed).Msg("Cleaned up failed tasks")
		}
	}

	return nil
}
