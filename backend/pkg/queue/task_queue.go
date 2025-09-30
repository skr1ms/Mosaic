package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/skr1ms/mosaic/pkg/middleware"
)

type TaskQueue struct {
	name              string
	redisClient       *redis.Client
	ctx               context.Context
	cancel            context.CancelFunc
	logger            *middleware.Logger
}

// Task represents task in queue
type Task struct {
	ID          string         `json:"id"`
	Type        string         `json:"type"`
	Payload     map[string]any `json:"payload"`
	Priority    int            `json:"priority"`
	MaxRetries  int            `json:"max_retries"`
	Retries     int            `json:"retries"`
	CreatedAt   time.Time      `json:"created_at"`
	ScheduledAt *time.Time     `json:"scheduled_at,omitempty"`
	ProcessedAt *time.Time     `json:"processed_at,omitempty"`
	Error       string         `json:"error,omitempty"`
}

// TaskHandler function for task processing
type TaskHandler func(ctx context.Context, task *Task) error

func NewTaskQueue(name string, redisClient *redis.Client, logger *middleware.Logger) *TaskQueue {
	ctx, cancel := context.WithCancel(context.Background())

	queue := &TaskQueue{
		name:        name,
		redisClient: redisClient,
		ctx:         ctx,
		cancel:      cancel,
		logger:      logger,
	}

	return queue
}

// Enqueue adds task to queue
func (q *TaskQueue) Enqueue(taskType string, payload map[string]any, opts ...TaskOption) error {
	task := &Task{
		ID:         uuid.New().String(),
		Type:       taskType,
		Payload:    payload,
		Priority:   0,
		MaxRetries: 3,
		Retries:    0,
		CreatedAt:  time.Now(),
	}

	for _, opt := range opts {
		opt(task)
	}

	queueKey := q.getQueueKey(task.Priority)

	if task.ScheduledAt != nil && task.ScheduledAt.After(time.Now()) {
		return q.enqueueDelayed(task)
	}

	taskData, err := json.Marshal(task)
	if err != nil {
		q.logger.GetZerologLogger().Error().Err(err).Str("task_id", task.ID).Str("task_type", task.Type).Msg("Failed to marshal task")
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	err = q.redisClient.LPush(q.ctx, queueKey, taskData).Err()
	if err != nil {
		q.logger.GetZerologLogger().Error().Err(err).Str("task_id", task.ID).Str("task_type", task.Type).Str("queue", q.name).Msg("Failed to enqueue task")
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	q.logger.GetZerologLogger().Info().
		Str("task_id", task.ID).
		Str("task_type", task.Type).
		Str("queue", q.name).
		Int("priority", task.Priority).
		Msg("Task enqueued")

	return nil
}

// enqueueDelayed adds delayed task
func (q *TaskQueue) enqueueDelayed(task *Task) error {
	taskData, err := json.Marshal(task)
	if err != nil {
		q.logger.GetZerologLogger().Error().Err(err).Str("task_id", task.ID).Str("task_type", task.Type).Msg("Failed to marshal delayed task")
		return fmt.Errorf("failed to marshal delayed task: %w", err)
	}

	delayedKey := q.getDelayedKey()
	score := float64(task.ScheduledAt.Unix())

	err = q.redisClient.ZAdd(q.ctx, delayedKey, redis.Z{
		Score:  score,
		Member: taskData,
	}).Err()

	if err != nil {
		q.logger.GetZerologLogger().Error().Err(err).Str("task_id", task.ID).Str("task_type", task.Type).Str("delayed_key", delayedKey).Msg("Failed to enqueue delayed task")
		return fmt.Errorf("failed to enqueue delayed task: %w", err)
	}

	q.logger.GetZerologLogger().Info().
		Str("task_id", task.ID).
		Str("task_type", task.Type).
		Time("scheduled_at", *task.ScheduledAt).
		Msg("Delayed task enqueued")

	return nil
}

func (q *TaskQueue) Dequeue(timeout time.Duration) (*Task, error) {
	for priority := 10; priority >= 0; priority-- {
		queueKey := q.getQueueKey(priority)

		result, err := q.redisClient.BRPop(q.ctx, timeout, queueKey).Result()
		if err != nil {
			if err == redis.Nil {
				continue
			}
			q.logger.GetZerologLogger().Error().Err(err).Str("queue", q.name).Int("priority", priority).Msg("Failed to dequeue task")
			return nil, fmt.Errorf("failed to dequeue task: %w", err)
		}

		if len(result) != 2 {
			continue
		}

		var task Task
		if err := json.Unmarshal([]byte(result[1]), &task); err != nil {
			q.logger.GetZerologLogger().Error().Err(err).Msg("Failed to unmarshal task")
			continue
		}

		return &task, nil
	}

	return nil, nil
}

func (q *TaskQueue) ProcessDelayedTasks() error {
	delayedKey := q.getDelayedKey()
	now := float64(time.Now().Unix())

	result, err := q.redisClient.ZRangeByScoreWithScores(q.ctx, delayedKey, &redis.ZRangeBy{
		Min: "-inf",
		Max: fmt.Sprintf("%f", now),
	}).Result()

	if err != nil {
		q.logger.GetZerologLogger().Error().Err(err).Str("queue", q.name).Msg("Failed to get delayed tasks")
		return fmt.Errorf("failed to get delayed tasks: %w", err)
	}

	for _, z := range result {
		taskData := z.Member.(string)

		var task Task
		if err := json.Unmarshal([]byte(taskData), &task); err != nil {
			q.logger.GetZerologLogger().Error().Err(err).Msg("Failed to unmarshal delayed task")
			continue
		}

		queueKey := q.getQueueKey(task.Priority)

		pipe := q.redisClient.Pipeline()
		pipe.ZRem(q.ctx, delayedKey, taskData)
		pipe.LPush(q.ctx, queueKey, taskData)

		if _, err := pipe.Exec(q.ctx); err != nil {
			q.logger.GetZerologLogger().Error().Err(err).Str("task_id", task.ID).Msg("Failed to move delayed task")
			continue
		}

		q.logger.GetZerologLogger().Info().
			Str("task_id", task.ID).
			Str("task_type", task.Type).
			Msg("Delayed task moved to queue")
	}

	return nil
}

// MarkCompleted marks task as completed
func (q *TaskQueue) MarkCompleted(task *Task) error {
	completedKey := q.getCompletedKey()
	now := time.Now()
	task.ProcessedAt = &now

	taskData, err := json.Marshal(task)
	if err != nil {
		q.logger.GetZerologLogger().Error().Err(err).Str("task_id", task.ID).Str("task_type", task.Type).Msg("Failed to marshal completed task")
		return fmt.Errorf("failed to marshal completed task: %w", err)
	}

	// Add to completed set with 24 hour TTL
	err = q.redisClient.ZAdd(q.ctx, completedKey, redis.Z{
		Score:  float64(now.Unix()),
		Member: taskData,
	}).Err()

	if err != nil {
		q.logger.GetZerologLogger().Error().Err(err).Str("task_id", task.ID).Str("task_type", task.Type).Msg("Failed to mark task as completed")
		return fmt.Errorf("failed to mark task as completed: %w", err)
	}

	// Set TTL on key
	q.redisClient.Expire(q.ctx, completedKey, 24*time.Hour)

	q.logger.GetZerologLogger().Info().
		Str("task_id", task.ID).
		Str("task_type", task.Type).
		Msg("Task marked as completed")

	return nil
}

// MarkFailed marks the task as unsuccessful
func (q *TaskQueue) MarkFailed(task *Task, err error) error {
	task.Retries++
	task.Error = err.Error()
	now := time.Now()
	task.ProcessedAt = &now

	// If there are more attempts, return to queue with delay
	if task.Retries < task.MaxRetries {
		delay := time.Duration(task.Retries*task.Retries) * time.Minute // Exponential delay
		scheduledAt := time.Now().Add(delay)
		task.ScheduledAt = &scheduledAt
		task.ProcessedAt = nil

		return q.enqueueDelayed(task)
	}

	// Otherwise add to failed set
	failedKey := q.getFailedKey()
	taskData, marshalErr := json.Marshal(task)
	if marshalErr != nil {
		q.logger.GetZerologLogger().Error().Err(marshalErr).Str("task_id", task.ID).Str("task_type", task.Type).Msg("Failed to marshal failed task")
		return fmt.Errorf("failed to marshal failed task: %w", marshalErr)
	}

	redisErr := q.redisClient.ZAdd(q.ctx, failedKey, redis.Z{
		Score:  float64(now.Unix()),
		Member: taskData,
	}).Err()

	if redisErr != nil {
		q.logger.GetZerologLogger().Error().Err(redisErr).Str("task_id", task.ID).Str("task_type", task.Type).Msg("Failed to mark task as failed")
		return fmt.Errorf("failed to mark task as failed: %w", redisErr)
	}

	// Set TTL on key
	q.redisClient.Expire(q.ctx, failedKey, 7*24*time.Hour) // 7 days

	q.logger.GetZerologLogger().Error().
		Err(err).
		Str("task_id", task.ID).
		Str("task_type", task.Type).
		Int("retries", task.Retries).
		Msg("Task failed permanently")

	return nil
}

func (q *TaskQueue) StartWorker(handlers map[string]TaskHandler) {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := q.ProcessDelayedTasks(); err != nil {
					q.logger.GetZerologLogger().Error().Err(err).Msg("Failed to process delayed tasks")
				}
			case <-q.ctx.Done():
				return
			}
		}
	}()

	go func() {
		for {
			select {
			case <-q.ctx.Done():
				q.logger.GetZerologLogger().Info().Str("queue", q.name).Msg("Task queue worker stopped")
				return
			default:
				task, err := q.Dequeue(5 * time.Second)
				if err != nil {
					q.logger.GetZerologLogger().Error().Err(err).Msg("Failed to dequeue task")
					continue
				}

				if task == nil {
					continue
				}

				handler, exists := handlers[task.Type]
				if !exists {
					q.logger.GetZerologLogger().Error().
						Str("task_id", task.ID).
						Str("task_type", task.Type).
						Msg("No handler found for task type")

					q.MarkFailed(task, fmt.Errorf("no handler found for task type: %s", task.Type))
					continue
				}

				q.logger.GetZerologLogger().Info().
					Str("task_id", task.ID).
					Str("task_type", task.Type).
					Msg("Processing task")

				if err := handler(q.ctx, task); err != nil {
					q.logger.GetZerologLogger().Error().
						Err(err).
						Str("task_id", task.ID).
						Str("task_type", task.Type).
						Msg("Task processing failed")

					q.MarkFailed(task, err)
				} else {
					q.MarkCompleted(task)
				}
			}
		}
	}()
}

func (q *TaskQueue) Close() error {
	q.cancel()

	q.logger.GetZerologLogger().Info().Str("queue", q.name).Msg("Task queue stopped")
	return nil
}

func (q *TaskQueue) getQueueKey(priority int) string {
	return fmt.Sprintf("queue:%s:priority:%d", q.name, priority)
}

func (q *TaskQueue) getDelayedKey() string {
	return fmt.Sprintf("queue:%s:delayed", q.name)
}

func (q *TaskQueue) getCompletedKey() string {
	return fmt.Sprintf("queue:%s:completed", q.name)
}

func (q *TaskQueue) getFailedKey() string {
	return fmt.Sprintf("queue:%s:failed", q.name)
}

type TaskOption func(*Task)

func WithPriority(priority int) TaskOption {
	return func(t *Task) {
		t.Priority = priority
	}
}

func WithMaxRetries(maxRetries int) TaskOption {
	return func(t *Task) {
		t.MaxRetries = maxRetries
	}
}

func WithDelay(delay time.Duration) TaskOption {
	return func(t *Task) {
		scheduledAt := time.Now().Add(delay)
		t.ScheduledAt = &scheduledAt
	}
}

// WithScheduledTime sets exact execution time
func WithScheduledTime(scheduledAt time.Time) TaskOption {
	return func(t *Task) {
		t.ScheduledAt = &scheduledAt
	}
}
