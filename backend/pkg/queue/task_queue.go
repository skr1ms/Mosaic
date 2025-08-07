package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

// TaskQueue представляет очередь задач
type TaskQueue struct {
	redis  *redis.Client
	name   string
	ctx    context.Context
	cancel context.CancelFunc
}

// Task представляет задачу в очереди
type Task struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Payload     map[string]interface{} `json:"payload"`
	Priority    int                    `json:"priority"`
	MaxRetries  int                    `json:"max_retries"`
	Retries     int                    `json:"retries"`
	CreatedAt   time.Time              `json:"created_at"`
	ScheduledAt *time.Time             `json:"scheduled_at,omitempty"`
	ProcessedAt *time.Time             `json:"processed_at,omitempty"`
	Error       string                 `json:"error,omitempty"`
}

// TaskHandler функция для обработки задач
type TaskHandler func(ctx context.Context, task *Task) error

// NewTaskQueue создает новую очередь задач
func NewTaskQueue(redis *redis.Client, name string) *TaskQueue {
	ctx, cancel := context.WithCancel(context.Background())

	return &TaskQueue{
		redis:  redis,
		name:   name,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Enqueue добавляет задачу в очередь
func (q *TaskQueue) Enqueue(taskType string, payload map[string]interface{}, opts ...TaskOption) error {
	task := &Task{
		ID:         uuid.New().String(),
		Type:       taskType,
		Payload:    payload,
		Priority:   0,
		MaxRetries: 3,
		Retries:    0,
		CreatedAt:  time.Now(),
	}

	// Применяем опции
	for _, opt := range opts {
		opt(task)
	}

	// Определяем ключ для очереди
	queueKey := q.getQueueKey(task.Priority)

	// Если задача отложенная, добавляем в delayed set
	if task.ScheduledAt != nil && task.ScheduledAt.After(time.Now()) {
		return q.enqueueDelayed(task)
	}

	// Сериализуем задачу
	taskData, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	// Добавляем в очередь
	err = q.redis.LPush(q.ctx, queueKey, taskData).Err()
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	log.Info().
		Str("task_id", task.ID).
		Str("task_type", task.Type).
		Str("queue", q.name).
		Int("priority", task.Priority).
		Msg("Task enqueued")

	return nil
}

// enqueueDelayed добавляет отложенную задачу
func (q *TaskQueue) enqueueDelayed(task *Task) error {
	taskData, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal delayed task: %w", err)
	}

	delayedKey := q.getDelayedKey()
	score := float64(task.ScheduledAt.Unix())

	err = q.redis.ZAdd(q.ctx, delayedKey, redis.Z{
		Score:  score,
		Member: taskData,
	}).Err()

	if err != nil {
		return fmt.Errorf("failed to enqueue delayed task: %w", err)
	}

	log.Info().
		Str("task_id", task.ID).
		Str("task_type", task.Type).
		Time("scheduled_at", *task.ScheduledAt).
		Msg("Delayed task enqueued")

	return nil
}

// Dequeue извлекает задачу из очереди для обработки
func (q *TaskQueue) Dequeue(timeout time.Duration) (*Task, error) {
	// Пытаемся получить задачу с разными приоритетами (сначала высокий)
	for priority := 10; priority >= 0; priority-- {
		queueKey := q.getQueueKey(priority)

		result, err := q.redis.BRPop(q.ctx, timeout, queueKey).Result()
		if err != nil {
			if err == redis.Nil {
				continue // Очередь пуста, пробуем следующий приоритет
			}
			return nil, fmt.Errorf("failed to dequeue task: %w", err)
		}

		if len(result) != 2 {
			continue
		}

		var task Task
		if err := json.Unmarshal([]byte(result[1]), &task); err != nil {
			log.Error().Err(err).Msg("Failed to unmarshal task")
			continue
		}

		return &task, nil
	}

	return nil, nil // Нет задач
}

// ProcessDelayedTasks перемещает отложенные задачи в основную очередь
func (q *TaskQueue) ProcessDelayedTasks() error {
	delayedKey := q.getDelayedKey()
	now := float64(time.Now().Unix())

	// Получаем задачи, время которых подошло
	result, err := q.redis.ZRangeByScoreWithScores(q.ctx, delayedKey, &redis.ZRangeBy{
		Min: "-inf",
		Max: fmt.Sprintf("%f", now),
	}).Result()

	if err != nil {
		return fmt.Errorf("failed to get delayed tasks: %w", err)
	}

	for _, z := range result {
		taskData := z.Member.(string)

		var task Task
		if err := json.Unmarshal([]byte(taskData), &task); err != nil {
			log.Error().Err(err).Msg("Failed to unmarshal delayed task")
			continue
		}

		// Перемещаем в основную очередь
		queueKey := q.getQueueKey(task.Priority)

		pipe := q.redis.Pipeline()
		pipe.ZRem(q.ctx, delayedKey, taskData)
		pipe.LPush(q.ctx, queueKey, taskData)

		if _, err := pipe.Exec(q.ctx); err != nil {
			log.Error().Err(err).Str("task_id", task.ID).Msg("Failed to move delayed task")
			continue
		}

		log.Info().
			Str("task_id", task.ID).
			Str("task_type", task.Type).
			Msg("Delayed task moved to queue")
	}

	return nil
}

// MarkCompleted помечает задачу как завершенную
func (q *TaskQueue) MarkCompleted(task *Task) error {
	completedKey := q.getCompletedKey()
	now := time.Now()
	task.ProcessedAt = &now

	taskData, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal completed task: %w", err)
	}

	// Добавляем в completed set с TTL 24 часа
	err = q.redis.ZAdd(q.ctx, completedKey, redis.Z{
		Score:  float64(now.Unix()),
		Member: taskData,
	}).Err()

	if err != nil {
		return fmt.Errorf("failed to mark task as completed: %w", err)
	}

	// Устанавливаем TTL на ключ
	q.redis.Expire(q.ctx, completedKey, 24*time.Hour)

	log.Info().
		Str("task_id", task.ID).
		Str("task_type", task.Type).
		Msg("Task marked as completed")

	return nil
}

// MarkFailed помечает задачу как неудачную
func (q *TaskQueue) MarkFailed(task *Task, err error) error {
	task.Retries++
	task.Error = err.Error()
	now := time.Now()
	task.ProcessedAt = &now

	// Если есть еще попытки, возвращаем в очередь с задержкой
	if task.Retries < task.MaxRetries {
		delay := time.Duration(task.Retries*task.Retries) * time.Minute // Экспоненциальная задержка
		scheduledAt := time.Now().Add(delay)
		task.ScheduledAt = &scheduledAt
		task.ProcessedAt = nil

		return q.enqueueDelayed(task)
	}

	// Иначе добавляем в failed set
	failedKey := q.getFailedKey()
	taskData, marshalErr := json.Marshal(task)
	if marshalErr != nil {
		return fmt.Errorf("failed to marshal failed task: %w", marshalErr)
	}

	redisErr := q.redis.ZAdd(q.ctx, failedKey, redis.Z{
		Score:  float64(now.Unix()),
		Member: taskData,
	}).Err()

	if redisErr != nil {
		return fmt.Errorf("failed to mark task as failed: %w", redisErr)
	}

	// Устанавливаем TTL на ключ
	q.redis.Expire(q.ctx, failedKey, 7*24*time.Hour) // 7 дней

	log.Error().
		Err(err).
		Str("task_id", task.ID).
		Str("task_type", task.Type).
		Int("retries", task.Retries).
		Msg("Task failed permanently")

	return nil
}

// StartWorker запускает воркер для обработки задач
func (q *TaskQueue) StartWorker(handlers map[string]TaskHandler) {
	go func() {
		log.Info().Str("queue", q.name).Msg("Task queue worker started")

		// Воркер для обработки отложенных задач
		go func() {
			ticker := time.NewTicker(30 * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					if err := q.ProcessDelayedTasks(); err != nil {
						log.Error().Err(err).Msg("Failed to process delayed tasks")
					}
				case <-q.ctx.Done():
					return
				}
			}
		}()

		// Основной воркер
		for {
			select {
			case <-q.ctx.Done():
				log.Info().Str("queue", q.name).Msg("Task queue worker stopped")
				return
			default:
				task, err := q.Dequeue(5 * time.Second)
				if err != nil {
					log.Error().Err(err).Msg("Failed to dequeue task")
					continue
				}

				if task == nil {
					continue // Нет задач
				}

				// Находим обработчик для типа задачи
				handler, exists := handlers[task.Type]
				if !exists {
					log.Error().
						Str("task_id", task.ID).
						Str("task_type", task.Type).
						Msg("No handler found for task type")

					q.MarkFailed(task, fmt.Errorf("no handler found for task type: %s", task.Type))
					continue
				}

				// Обрабатываем задачу
				log.Info().
					Str("task_id", task.ID).
					Str("task_type", task.Type).
					Msg("Processing task")

				if err := handler(q.ctx, task); err != nil {
					log.Error().
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

// Stop останавливает очередь задач
func (q *TaskQueue) Stop() {
	q.cancel()
	log.Info().Str("queue", q.name).Msg("Task queue stopped")
}

// Вспомогательные методы для генерации ключей Redis

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

// TaskOption функция для настройки задачи
type TaskOption func(*Task)

// WithPriority устанавливает приоритет задачи
func WithPriority(priority int) TaskOption {
	return func(t *Task) {
		t.Priority = priority
	}
}

// WithMaxRetries устанавливает максимальное количество попыток
func WithMaxRetries(maxRetries int) TaskOption {
	return func(t *Task) {
		t.MaxRetries = maxRetries
	}
}

// WithDelay добавляет задержку перед выполнением
func WithDelay(delay time.Duration) TaskOption {
	return func(t *Task) {
		scheduledAt := time.Now().Add(delay)
		t.ScheduledAt = &scheduledAt
	}
}

// WithScheduledTime устанавливает точное время выполнения
func WithScheduledTime(scheduledAt time.Time) TaskOption {
	return func(t *Task) {
		t.ScheduledAt = &scheduledAt
	}
}
