package email

import (
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// EmailJob представляет задачу отправки email
type EmailJob struct {
	ID          string
	To          string
	Subject     string
	TemplateID  string
	Data        map[string]interface{}
	Attachments []Attachment
	CreatedAt   time.Time
	Attempts    int
	MaxRetries  int
	NextRetry   time.Time
}

// EmailQueue обеспечивает отправку email с retry механизмом
type EmailQueue struct {
	mailer    *Mailer
	jobs      chan *EmailJob
	retryJobs chan *EmailJob
	workers   int
	quit      chan bool
	wg        sync.WaitGroup
	running   bool
	mu        sync.RWMutex
}

// NewEmailQueue создает новую очередь email
func NewEmailQueue(mailer *Mailer, workers int) *EmailQueue {
	return &EmailQueue{
		mailer:    mailer,
		jobs:      make(chan *EmailJob, 1000),
		retryJobs: make(chan *EmailJob, 1000),
		workers:   workers,
		quit:      make(chan bool),
		running:   false,
	}
}

// Start запускает обработку очереди
func (q *EmailQueue) Start() {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.running {
		return
	}

	q.running = true
	log.Info().Int("workers", q.workers).Msg("Starting email queue")

	// Запускаем воркеров
	for i := 0; i < q.workers; i++ {
		q.wg.Add(1)
		go q.worker(i)
	}

	// Запускаем retry scheduler
	q.wg.Add(1)
	go q.retryScheduler()
}

// Stop останавливает обработку очереди
func (q *EmailQueue) Stop() {
	q.mu.Lock()
	defer q.mu.Unlock()

	if !q.running {
		return
	}

	log.Info().Msg("Stopping email queue")
	close(q.quit)
	q.wg.Wait()
	q.running = false
	log.Info().Msg("Email queue stopped")
}

// SendEmail добавляет email в очередь
func (q *EmailQueue) SendEmail(job *EmailJob) error {
	if job.MaxRetries == 0 {
		job.MaxRetries = 3
	}
	job.CreatedAt = time.Now()
	job.NextRetry = time.Now()

	select {
	case q.jobs <- job:
		log.Info().
			Str("job_id", job.ID).
			Str("to", job.To).
			Str("template", job.TemplateID).
			Msg("Email job queued")
		return nil
	default:
		return fmt.Errorf("email queue is full")
	}
}

// SendSchemaEmail отправляет готовую схему пользователю
func (q *EmailQueue) SendSchemaEmail(to, couponCode, schemaURL string) error {
	job := &EmailJob{
		ID:         fmt.Sprintf("schema_%s_%d", couponCode, time.Now().Unix()),
		To:         to,
		Subject:    "Ваша схема алмазной мозаики готова!",
		TemplateID: "schema_ready",
		Data: map[string]interface{}{
			"CouponCode": couponCode,
			"SchemaURL":  schemaURL,
		},
		MaxRetries: 5, // Для схем больше попыток
	}

	return q.SendEmail(job)
}

// SendProcessingErrorEmail отправляет уведомление об ошибке обработки
func (q *EmailQueue) SendProcessingErrorEmail(to, couponCode, errorMessage string) error {
	job := &EmailJob{
		ID:         fmt.Sprintf("error_%s_%d", couponCode, time.Now().Unix()),
		To:         to,
		Subject:    "Ошибка обработки вашего изображения",
		TemplateID: "processing_error",
		Data: map[string]interface{}{
			"CouponCode":   couponCode,
			"ErrorMessage": errorMessage,
		},
		MaxRetries: 3,
	}

	return q.SendEmail(job)
}

// SendStatusUpdateEmail отправляет уведомление об изменении статуса
func (q *EmailQueue) SendStatusUpdateEmail(to, couponCode, status string, progress int) error {
	job := &EmailJob{
		ID:         fmt.Sprintf("status_%s_%d", couponCode, time.Now().Unix()),
		To:         to,
		Subject:    "Обновление статуса вашего заказа",
		TemplateID: "status_update",
		Data: map[string]interface{}{
			"CouponCode": couponCode,
			"Status":     status,
			"Progress":   progress,
		},
		MaxRetries: 3,
	}

	return q.SendEmail(job)
}

// worker обрабатывает задачи из очереди
func (q *EmailQueue) worker(id int) {
	defer q.wg.Done()

	log.Info().Int("worker_id", id).Msg("Email worker started")

	for {
		select {
		case <-q.quit:
			log.Info().Int("worker_id", id).Msg("Email worker stopping")
			return

		case job := <-q.jobs:
			q.processJob(job, id)

		case job := <-q.retryJobs:
			q.processJob(job, id)
		}
	}
}

// processJob обрабатывает отдельную задачу
func (q *EmailQueue) processJob(job *EmailJob, workerID int) {
	logger := log.With().
		Str("job_id", job.ID).
		Str("to", job.To).
		Int("worker_id", workerID).
		Int("attempt", job.Attempts+1).
		Logger()

	logger.Info().Msg("Processing email job")

	job.Attempts++

	var err error
	switch job.TemplateID {
	case "schema_ready":
		schemaURL, ok := job.Data["SchemaURL"].(string)
		if !ok {
			logger.Error().Msg("Missing SchemaURL in job data")
			return
		}
		couponCode, ok := job.Data["CouponCode"].(string)
		if !ok {
			logger.Error().Msg("Missing CouponCode in job data")
			return
		}
		err = q.mailer.SendSchemaEmail(job.To, schemaURL, couponCode)

	case "processing_error":
		couponCode, ok := job.Data["CouponCode"].(string)
		if !ok {
			logger.Error().Msg("Missing CouponCode in job data")
			return
		}
		errorMessage, ok := job.Data["ErrorMessage"].(string)
		if !ok {
			logger.Error().Msg("Missing ErrorMessage in job data")
			return
		}
		err = q.mailer.SendProcessingErrorEmail(job.To, couponCode, errorMessage)

	case "status_update":
		couponCode, ok := job.Data["CouponCode"].(string)
		if !ok {
			logger.Error().Msg("Missing CouponCode in job data")
			return
		}
		status, ok := job.Data["Status"].(string)
		if !ok {
			logger.Error().Msg("Missing Status in job data")
			return
		}
		progress, ok := job.Data["Progress"].(int)
		if !ok {
			logger.Error().Msg("Missing Progress in job data")
			return
		}
		progressMsg := fmt.Sprintf("Прогресс выполнения: %d%%", progress)
		err = q.mailer.SendStatusUpdateEmail(job.To, couponCode, status, progressMsg)

	default:
		logger.Error().Str("template_id", job.TemplateID).Msg("Unknown template ID")
		return
	}

	if err != nil {
		logger.Error().Err(err).Msg("Failed to send email")

		// Если не превышено максимальное количество попыток, планируем retry
		if job.Attempts < job.MaxRetries {
			q.scheduleRetry(job)
		} else {
			logger.Error().
				Int("max_retries", job.MaxRetries).
				Msg("Email job failed after max retries")
		}
	} else {
		logger.Info().Msg("Email sent successfully")
	}
}

// scheduleRetry планирует повторную попытку отправки
func (q *EmailQueue) scheduleRetry(job *EmailJob) {
	// Экспоненциальный backoff: 1 мин, 5 мин, 15 мин, 30 мин
	delays := []time.Duration{
		1 * time.Minute,
		5 * time.Minute,
		15 * time.Minute,
		30 * time.Minute,
	}

	var delay time.Duration
	if job.Attempts-1 < len(delays) {
		delay = delays[job.Attempts-1]
	} else {
		delay = delays[len(delays)-1]
	}

	job.NextRetry = time.Now().Add(delay)

	log.Info().
		Str("job_id", job.ID).
		Time("next_retry", job.NextRetry).
		Dur("delay", delay).
		Int("attempt", job.Attempts).
		Msg("Email job scheduled for retry")
}

// retryScheduler обрабатывает задачи для повторной отправки
func (q *EmailQueue) retryScheduler() {
	defer q.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	var pendingRetries []*EmailJob

	for {
		select {
		case <-q.quit:
			return

		case <-ticker.C:
			now := time.Now()
			var readyJobs []*EmailJob
			var stillPending []*EmailJob

			// Проверяем какие задачи готовы к повторной отправке
			for _, job := range pendingRetries {
				if now.After(job.NextRetry) {
					readyJobs = append(readyJobs, job)
				} else {
					stillPending = append(stillPending, job)
				}
			}

			// Отправляем готовые задачи обратно в очередь
			for _, job := range readyJobs {
				select {
				case q.retryJobs <- job:
					log.Info().
						Str("job_id", job.ID).
						Int("attempt", job.Attempts).
						Msg("Retry job sent to queue")
				default:
					// Если очередь полная, оставляем на следующий раз
					stillPending = append(stillPending, job)
				}
			}

			pendingRetries = stillPending

		case job := <-q.retryJobs:
			// Новые задачи для retry добавляем в pending список
			pendingRetries = append(pendingRetries, job)
		}
	}
}

// GetStats возвращает статистику очереди
func (q *EmailQueue) GetStats() map[string]interface{} {
	q.mu.RLock()
	defer q.mu.RUnlock()

	return map[string]interface{}{
		"running":      q.running,
		"workers":      q.workers,
		"jobs_queued":  len(q.jobs),
		"retry_queued": len(q.retryJobs),
	}
}
