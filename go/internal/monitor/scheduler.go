package monitor

import (
	"container/heap"
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/koblas/besops/internal/broadcast"
	"github.com/koblas/besops/lib/status"
)

// Priority levels for the scheduler queue.
const (
	priorityImmediate = iota // Manual "check now" or first check after start
	priorityRetry            // Failed check, retrying
	priorityScheduled        // Normal interval-based check
)

// schedItem represents a single scheduled check in the priority queue.
type schedItem struct {
	monitorID string
	fireAt    time.Time
	priority  int
	retry     int // current retry attempt (0 = first try)
	index     int // heap index, managed by container/heap
}

// schedHeap implements heap.Interface for priority-based scheduling.
// Items are ordered by: (1) fire time, (2) priority level.
type schedHeap []*schedItem

func (h schedHeap) Len() int { return len(h) }

func (h schedHeap) Less(i, j int) bool {
	if h[i].fireAt.Equal(h[j].fireAt) {
		return h[i].priority < h[j].priority
	}
	return h[i].fireAt.Before(h[j].fireAt)
}

func (h schedHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

func (h *schedHeap) Push(x any) {
	item := x.(*schedItem)
	item.index = len(*h)
	*h = append(*h, item)
}

func (h *schedHeap) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*h = old[:n-1]
	return item
}

// Scheduler is a priority-queue-based monitor scheduler with a bounded worker pool.
type Scheduler struct {
	mu       sync.Mutex
	queue    schedHeap
	configs  map[string]*Config   // monitorID -> current config
	lastFire map[string]time.Time // monitorID -> last time check was dispatched
	wake     chan struct{}        // signal the loop when queue changes

	store     Store
	hbStore   HeartbeatStore
	registry  *Registry
	notify    NotificationDispatcher
	publisher broadcast.Publisher
	maint     MaintenanceChecker
	metrics   MetricsObserver
	tags      TagProvider

	workerSem chan struct{} // bounded worker pool semaphore
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
}

// SchedulerOption configures the Scheduler.
type SchedulerOption func(*Scheduler)

// WithMaxWorkers sets the maximum concurrent check workers.
func WithMaxWorkers(n int) SchedulerOption {
	return func(s *Scheduler) {
		s.workerSem = make(chan struct{}, n)
	}
}

// WithMetrics attaches a metrics observer to record telemetry for every check.
func WithMetrics(m MetricsObserver) SchedulerOption {
	return func(s *Scheduler) {
		s.metrics = m
	}
}

// WithTags attaches a tag provider to enrich telemetry with monitor tags.
func WithTags(t TagProvider) SchedulerOption {
	return func(s *Scheduler) {
		s.tags = t
	}
}

func NewScheduler(store Store, hbStore HeartbeatStore, registry *Registry, notify NotificationDispatcher, publisher broadcast.Publisher, maint MaintenanceChecker, opts ...SchedulerOption) *Scheduler {
	s := &Scheduler{
		queue:     make(schedHeap, 0),
		configs:   make(map[string]*Config),
		lastFire:  make(map[string]time.Time),
		wake:      make(chan struct{}, 1),
		store:     store,
		hbStore:   hbStore,
		registry:  registry,
		notify:    notify,
		publisher: publisher,
		maint:     maint,
		workerSem: make(chan struct{}, 32),
	}
	for _, opt := range opts {
		opt(s)
	}
	heap.Init(&s.queue)
	return s
}

// Start loads all active monitors and begins the scheduling loop.
func (s *Scheduler) Start(ctx context.Context) error {
	s.ctx, s.cancel = context.WithCancel(ctx)

	ids, err := s.store.FindAllActiveIDs(ctx)
	if err != nil {
		return fmt.Errorf("loading active monitor IDs: %w", err)
	}

	slog.InfoContext(ctx, "scheduler starting", slog.Int("active_monitors", len(ids)))

	for _, id := range ids {
		if err := s.addMonitor(id); err != nil {
			slog.ErrorContext(ctx, "failed to load monitor for scheduling", slog.String("id", id), slog.Any("error", err))
		}
	}

	s.wg.Add(1)
	go s.loop() //nolint:contextcheck // loop uses s.ctx set above

	return nil
}

// Stop gracefully shuts down the scheduler and waits for in-flight checks.
func (s *Scheduler) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
	s.wg.Wait()
}

// ScheduleMonitor adds or reschedules a monitor. Called when a monitor is created or updated.
// Fires an immediate check so the new configuration is validated right away.
func (s *Scheduler) ScheduleMonitor(ctx context.Context, id string) error {
	s.mu.Lock()
	s.removeLockedNoWake(id)
	s.mu.Unlock()

	if err := s.addMonitorImmediate(id); err != nil {
		return err
	}

	s.signal()
	return nil
}

// RemoveMonitor removes a monitor from the schedule (pause/delete).
func (s *Scheduler) RemoveMonitor(_ context.Context, id string) {
	s.mu.Lock()
	s.removeLockedNoWake(id)
	delete(s.configs, id)
	delete(s.lastFire, id)
	s.mu.Unlock()
	s.signal()
}

// CheckNow queues an immediate check for a monitor.
func (s *Scheduler) CheckNow(_ context.Context, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.configs[id]; !exists {
		return
	}

	s.removeFromQueue(id)
	heap.Push(&s.queue, &schedItem{
		monitorID: id,
		fireAt:    time.Now(),
		priority:  priorityImmediate,
	})
	s.signal()
}

func (s *Scheduler) addMonitor(id string) error {
	return s.loadAndSchedule(id, false)
}

func (s *Scheduler) addMonitorImmediate(id string) error {
	return s.loadAndSchedule(id, true)
}

func (s *Scheduler) loadAndSchedule(id string, immediate bool) error {
	mon, err := s.store.FindByID(s.ctx, id)
	if err != nil {
		return fmt.Errorf("loading monitor %s: %w", id, err)
	}

	_, err = s.registry.Get(mon.Type)
	if err != nil {
		return fmt.Errorf("getting checker for type %s: %w", mon.Type, err)
	}

	cfg := modelToConfig(mon)

	if s.tags != nil {
		if tags, tagErr := s.tags.GetTagsForMonitor(s.ctx, id); tagErr == nil {
			cfg.Tags = tags
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.configs[id] = cfg

	var fireAt time.Time
	if immediate {
		fireAt = time.Now()
	} else if last, ok := s.lastFire[id]; ok {
		next := last.Add(cfg.Interval)
		if next.After(time.Now()) {
			fireAt = next
		} else {
			fireAt = time.Now()
		}
	} else {
		fireAt = time.Now()
	}

	heap.Push(&s.queue, &schedItem{
		monitorID: id,
		fireAt:    fireAt,
		priority:  priorityImmediate,
	})

	return nil
}

func (s *Scheduler) removeLockedNoWake(id string) {
	s.removeFromQueue(id)
}

func (s *Scheduler) removeFromQueue(id string) {
	for i, item := range s.queue {
		if item.monitorID == id {
			heap.Remove(&s.queue, i)
			return
		}
	}
}

func (s *Scheduler) signal() {
	select {
	case s.wake <- struct{}{}:
	default:
	}
}

func (s *Scheduler) loop() {
	defer s.wg.Done()

	var timer *time.Timer
	defer func() {
		if timer != nil {
			timer.Stop()
		}
	}()

	for {
		s.mu.Lock()
		if s.queue.Len() == 0 {
			s.mu.Unlock()
			select {
			case <-s.ctx.Done():
				return
			case <-s.wake:
				continue
			}
		}

		next := s.queue[0]
		now := time.Now()
		delay := next.fireAt.Sub(now)
		s.mu.Unlock()

		if delay <= 0 {
			s.dispatch()
			continue
		}

		if timer == nil {
			timer = time.NewTimer(delay)
		} else {
			timer.Reset(delay)
		}

		select {
		case <-s.ctx.Done():
			return
		case <-s.wake:
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			continue
		case <-timer.C:
			s.dispatch()
		}
	}
}

func (s *Scheduler) dispatch() {
	s.mu.Lock()
	if s.queue.Len() == 0 {
		s.mu.Unlock()
		return
	}

	item := heap.Pop(&s.queue).(*schedItem)
	cfg, exists := s.configs[item.monitorID]
	if !exists {
		s.mu.Unlock()
		return
	}

	cfgCopy := *cfg
	s.lastFire[item.monitorID] = time.Now()
	retry := item.retry
	s.mu.Unlock()

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		select {
		case s.workerSem <- struct{}{}:
		case <-s.ctx.Done():
			return
		}
		defer func() { <-s.workerSem }()

		s.executeCheck(&cfgCopy, retry)
	}()
}

func (s *Scheduler) executeCheck(cfg *Config, retry int) {
	checker, err := s.registry.Get(cfg.Type)
	if err != nil {
		slog.ErrorContext(s.ctx, "unknown checker type", slog.String("monitor", cfg.ID), slog.String("type", cfg.Type))
		s.scheduleNext(cfg.ID, priorityScheduled, 0)
		return
	}

	_, isAggregate := checker.(AggregateChecker)

	timeout := cfg.Timeout
	if isAggregate {
		timeout = 30 * time.Second
	}

	checkCtx, cancel := context.WithTimeout(s.ctx, timeout)
	result, err := checker.Check(checkCtx, cfg)
	cancel()

	if err != nil {
		result = CheckResult{Status: status.Down, Message: err.Error()}
	}

	// Aggregate checkers (groups) never retry — their result is deterministic.
	if isAggregate || result.Status == status.Up || retry >= cfg.MaxRetries {
		handler := &resultRecorder{
			hbStore:     s.hbStore,
			notify:      s.notify,
			publisher:   s.publisher,
			maint:       s.maint,
			metrics:     s.metrics,
			monitorName: cfg.Name,
			monitorType: cfg.Type,
			tags:        cfg.Tags,
		}
		handler.HandleResult(s.ctx, cfg.ID, result, retry)
		s.scheduleNext(cfg.ID, priorityScheduled, 0)
	} else {
		s.scheduleRetry(cfg.ID, retry+1, cfg.RetryInterval)
	}
}

func (s *Scheduler) scheduleNext(monitorID string, priority int, retry int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cfg, exists := s.configs[monitorID]
	if !exists {
		return
	}

	heap.Push(&s.queue, &schedItem{
		monitorID: monitorID,
		fireAt:    time.Now().Add(cfg.Interval),
		priority:  priority,
		retry:     retry,
	})
	s.signal()
}

func (s *Scheduler) scheduleRetry(monitorID string, retry int, retryInterval time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.configs[monitorID]; !exists {
		return
	}

	if retryInterval == 0 {
		retryInterval = 30 * time.Second
	}

	heap.Push(&s.queue, &schedItem{
		monitorID: monitorID,
		fireAt:    time.Now().Add(retryInterval),
		priority:  priorityRetry,
		retry:     retry,
	})
	s.signal()
}
