package email

import (
	"fmt"
	"sync"
	"time"

	"github.com/skr1ms/mosaic/pkg/middleware"
)

// EmailJob represents email sending task
type EmailJob struct {
	ID          string
	To          string
	Subject     string
	TemplateID  string
	Data        map[string]any
	Attachments []Attachment
	CreatedAt   time.Time
	Attempts    int
	MaxRetries  int
	NextRetry   time.Time
}

// EmailQueue email queue with managed goroutines
type EmailQueue struct {
	mailer           *Mailer
	workers          int
	jobs             chan *EmailJob
	retryJobs        chan *EmailJob
	quit             chan struct{}
	running          bool
	mu               sync.RWMutex
	wg               sync.WaitGroup
	logger           *middleware.Logger
}

// NewEmailQueue creates new email queue
func NewEmailQueue(mailer *Mailer, workers int, logger *middleware.Logger) *EmailQueue {
	queue := &EmailQueue{
		mailer:    mailer,
		workers:   workers,
		jobs:      make(chan *EmailJob, 1000),
		retryJobs: make(chan *EmailJob, 1000),
		quit:      make(chan struct{}),
		logger:    logger,
	}

	return queue
}

// Start starts queue processing
func (q *EmailQueue) Start() {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.running {
		return
	}

	q.running = true
	q.logger.GetZerologLogger().Info().Int("workers", q.workers).Msg("Starting email queue")

	go func() {
		q.retryScheduler()
	}()
}

// Stop stops queue processing
func (q *EmailQueue) Stop() {
	q.mu.Lock()
	defer q.mu.Unlock()

	if !q.running {
		return
	}

	q.logger.GetZerologLogger().Info().Msg("Stopping email queue")
	close(q.quit)

	q.wg.Wait()
	q.running = false
	q.logger.GetZerologLogger().Info().Msg("Email queue stopped")
}

// SendEmail adds email to queue
func (q *EmailQueue) SendEmail(job *EmailJob) error {
	if job.MaxRetries == 0 {
		job.MaxRetries = 3
	}
	job.CreatedAt = time.Now()
	job.NextRetry = time.Now()

	select {
	case q.jobs <- job:
		q.logger.GetZerologLogger().Info().
			Str("job_id", job.ID).
			Str("to", job.To).
			Str("template", job.TemplateID).
			Msg("Email job queued")
		return nil
	default:
		q.logger.GetZerologLogger().Error().Msg("Email queue is full")
		return fmt.Errorf("email queue is full")
	}
}

func (q *EmailQueue) SendSchemaEmail(to, couponCode, schemaURL string) error {
	job := &EmailJob{
		ID:         fmt.Sprintf("schema_%s_%d", couponCode, time.Now().Unix()),
		To:         to,
		Subject:    "Your diamond mosaic schema is ready!",
		TemplateID: "schema_ready",
		Data: map[string]any{
			"CouponCode": couponCode,
			"SchemaURL":  schemaURL,
		},
		MaxRetries: 5,
	}

	return q.SendEmail(job)
}

func (q *EmailQueue) SendProcessingErrorEmail(to, couponCode, errorMessage string) error {
	job := &EmailJob{
		ID:         fmt.Sprintf("error_%s_%d", couponCode, time.Now().Unix()),
		To:         to,
		Subject:    "Error processing your image",
		TemplateID: "processing_error",
		Data: map[string]any{
			"CouponCode":   couponCode,
			"ErrorMessage": errorMessage,
		},
		MaxRetries: 3,
	}

	return q.SendEmail(job)
}

func (q *EmailQueue) SendStatusUpdateEmail(to, couponCode, status string, progress int) error {
	job := &EmailJob{
		ID:         fmt.Sprintf("status_%s_%d", couponCode, time.Now().Unix()),
		To:         to,
		Subject:    "Update on your order status",
		TemplateID: "status_update",
		Data: map[string]any{
			"CouponCode": couponCode,
			"Status":     status,
			"Progress":   progress,
		},
		MaxRetries: 3,
	}

	return q.SendEmail(job)
}

// retryScheduler processes tasks for retry
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

			for _, job := range pendingRetries {
				if now.After(job.NextRetry) {
					readyJobs = append(readyJobs, job)
				} else {
					stillPending = append(stillPending, job)
				}
			}

			for _, job := range readyJobs {
				select {
				case q.retryJobs <- job:
					q.logger.GetZerologLogger().Info().
						Str("job_id", job.ID).
						Int("attempt", job.Attempts).
						Msg("Retry job sent to queue")
				default:
					stillPending = append(stillPending, job)
				}
			}

			pendingRetries = stillPending

		case job := <-q.retryJobs:
			pendingRetries = append(pendingRetries, job)
		}
	}
}

// GetStats returns queue statistics
func (q *EmailQueue) GetStats() map[string]any {
	q.mu.RLock()
	defer q.mu.RUnlock()

	return map[string]any{
		"running":      q.running,
		"workers":      q.workers,
		"jobs_queued":  len(q.jobs),
		"retry_queued": len(q.retryJobs),
	}
}
