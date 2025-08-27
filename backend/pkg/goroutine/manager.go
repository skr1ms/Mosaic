package goroutine

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
)

type Manager struct {
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	mu           sync.RWMutex
	goroutines   map[string]*GoroutineInfo
	workerPools  map[string]*WorkerPool
	shutdownOnce sync.Once
	shutdownChan chan struct{}
	metrics      *Metrics
}

type GoroutineInfo struct {
	ID           string
	Name         string
	StartedAt    time.Time
	Status       string
	LastActivity time.Time
	Error        error
}

type WorkerPool struct {
	name      string
	workers   int
	tasks     chan func()
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	active    int32
	completed int64
	failed    int64
}

type Metrics struct {
	TotalGoroutines  int64
	ActiveGoroutines int64
	CompletedTasks   int64
	FailedTasks      int64
}

func NewManager(ctx context.Context) *Manager {
	ctx, cancel := context.WithCancel(ctx)
	return &Manager{
		ctx:          ctx,
		cancel:       cancel,
		goroutines:   make(map[string]*GoroutineInfo),
		workerPools:  make(map[string]*WorkerPool),
		shutdownChan: make(chan struct{}),
		metrics:      &Metrics{},
	}
}

func (m *Manager) StartGoroutine(name string, fn func() error) string {
	id := fmt.Sprintf("%s_%d", name, time.Now().UnixNano())

	m.mu.Lock()
	defer m.mu.Unlock()

	info := &GoroutineInfo{
		ID:        id,
		Name:      name,
		StartedAt: time.Now(),
		Status:    "running",
	}

	m.goroutines[id] = info
	atomic.AddInt64(&m.metrics.TotalGoroutines, 1)
	atomic.AddInt64(&m.metrics.ActiveGoroutines, 1)

	m.wg.Add(1)
	go func() {
		defer func() {
			m.wg.Done()
			atomic.AddInt64(&m.metrics.ActiveGoroutines, -1)

			if r := recover(); r != nil {
				info.Status = "panic"
				info.Error = fmt.Errorf("panic: %v", r)
				log.Error().
					Str("goroutine_id", id).
					Str("name", name).
					Interface("panic", r).
					Str("stack", string(debug.Stack())).
					Msg("Goroutine panic recovered")
			}
		}()

		select {
		case <-m.ctx.Done():
			info.Status = "cancelled"
			return
		default:
		}

		if err := fn(); err != nil {
			info.Status = "failed"
			info.Error = err
			log.Error().
				Err(err).
				Str("goroutine_id", id).
				Str("name", name).
				Msg("Goroutine failed")
		} else {
			info.Status = "completed"
		}

		info.LastActivity = time.Now()
	}()

	log.Info().
		Str("goroutine_id", id).
		Str("name", name).
		Msg("Goroutine started")

	return id
}

func (m *Manager) StartGoroutineWithContext(name string, ctx context.Context, fn func(context.Context) error) string {
	return m.StartGoroutine(name, func() error {
		return fn(ctx)
	})
}

func (m *Manager) StartGoroutineWithTimeout(name string, timeout time.Duration, fn func() error) string {
	return m.StartGoroutine(name, func() error {
		done := make(chan error, 1)
		go func() {
			done <- fn()
		}()

		select {
		case err := <-done:
			return err
		case <-time.After(timeout):
			log.Error().Str("goroutine_name", name).Dur("timeout", timeout).Msg("Goroutine timed out")
			return fmt.Errorf("goroutine timed out after %v", timeout)
		case <-m.ctx.Done():
			log.Error().Str("goroutine_name", name).Msg("Goroutine cancelled")
			return fmt.Errorf("goroutine cancelled")
		}
	})
}

func (m *Manager) NewWorkerPool(name string, workers int, bufferSize int) *WorkerPool {
	m.mu.Lock()
	defer m.mu.Unlock()

	ctx, cancel := context.WithCancel(m.ctx)
	pool := &WorkerPool{
		name:    name,
		workers: workers,
		tasks:   make(chan func(), bufferSize),
		ctx:     ctx,
		cancel:  cancel,
	}

	m.workerPools[name] = pool

	for i := 0; i < workers; i++ {
		pool.wg.Add(1)
		go pool.worker(i)
	}

	log.Info().
		Str("pool_name", name).
		Int("workers", workers).
		Int("buffer_size", bufferSize).
		Msg("Worker pool created")

	return pool
}

// worker main worker loop
func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()

	for {
		select {
		case task, ok := <-wp.tasks:
			if !ok {
				return
			}

			atomic.AddInt32(&wp.active, 1)

			func() {
				defer func() {
					atomic.AddInt32(&wp.active, -1)
					if r := recover(); r != nil {
						atomic.AddInt64(&wp.failed, 1)
						log.Error().
							Str("pool_name", wp.name).
							Int("worker_id", id).
							Interface("panic", r).
							Str("stack", string(debug.Stack())).
							Msg("Worker panic recovered")
					}
				}()

				task()
				atomic.AddInt64(&wp.completed, 1)
			}()

		case <-wp.ctx.Done():
			return
		}
	}
}

func (wp *WorkerPool) SubmitTask(task func()) error {
	select {
	case wp.tasks <- task:
		return nil
	case <-wp.ctx.Done():
		log.Error().Str("pool_name", wp.name).Msg("Worker pool is shutting down")
		return fmt.Errorf("worker pool is shutting down")
	default:
		log.Error().Str("pool_name", wp.name).Msg("Worker pool is full")
		return fmt.Errorf("worker pool is full")
	}
}

func (wp *WorkerPool) SubmitTaskWithTimeout(task func(), timeout time.Duration) error {
	select {
	case wp.tasks <- task:
		return nil
	case <-time.After(timeout):
		log.Error().Str("pool_name", wp.name).Dur("timeout", timeout).Msg("Submit timeout in worker pool")
		return fmt.Errorf("submit timeout after %v", timeout)
	case <-wp.ctx.Done():
		log.Error().Str("pool_name", wp.name).Msg("Worker pool is shutting down during submit")
		return fmt.Errorf("worker pool is shutting down")
	}
}

func (m *Manager) GetMetrics() Metrics {
	return *m.metrics
}

func (m *Manager) GetGoroutineInfo(id string) (*GoroutineInfo, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	info, exists := m.goroutines[id]
	return info, exists
}

func (m *Manager) GetAllGoroutines() map[string]*GoroutineInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]*GoroutineInfo)
	for id, info := range m.goroutines {
		result[id] = info
	}
	return result
}

func (m *Manager) StopGoroutine(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	info, exists := m.goroutines[id]
	if !exists {
		return false
	}

	if info.Status == "running" {
		info.Status = "stopping"
	}

	return true
}

func (m *Manager) Shutdown(timeout time.Duration) error {
	var result error
	m.shutdownOnce.Do(func() {
		log.Info().Msg("Shutting down goroutine manager")

		close(m.shutdownChan)

		m.mu.Lock()
		for _, pool := range m.workerPools {
			pool.cancel()
		}
		m.mu.Unlock()

		m.cancel()

		done := make(chan struct{})
		go func() {
			m.wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			log.Info().Msg("All goroutines stopped gracefully")
			result = nil
		case <-time.After(timeout):
			log.Warn().Dur("timeout", timeout).Msg("Shutdown timeout, forcing stop")
			result = fmt.Errorf("shutdown timeout after %v", timeout)
		}
	})
	return result
}

// IsShuttingDown checks if shutdown process is in progress
func (m *Manager) IsShuttingDown() bool {
	select {
	case <-m.shutdownChan:
		return true
	default:
		return false
	}
}

// Close releases manager resources
func (m *Manager) Close() error {
	return m.Shutdown(30 * time.Second)
}
