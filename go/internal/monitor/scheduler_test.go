package monitor

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/koblas/besops/internal/broadcast"
	domainmonitor "github.com/koblas/besops/internal/domain/monitor"
	"github.com/koblas/besops/internal/domain/heartbeat"
	"github.com/koblas/besops/lib/status"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockStore struct {
	mu       sync.Mutex
	monitors map[string]*domainmonitor.Monitor
}

func (s *mockStore) FindByID(_ context.Context, id string) (*domainmonitor.Monitor, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	m, ok := s.monitors[id]
	if !ok {
		return nil, assert.AnError
	}
	return m, nil
}

func (s *mockStore) FindAllActiveIDs(_ context.Context) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ids := make([]string, 0, len(s.monitors))
	for id := range s.monitors {
		ids = append(ids, id)
	}
	return ids, nil
}

type mockHBStore struct {
	mu   sync.Mutex
	data []*heartbeat.Heartbeat
}

func (s *mockHBStore) Insert(_ context.Context, hb *heartbeat.Heartbeat) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = append(s.data, hb)
	return nil
}

func (s *mockHBStore) GetLatest(_ context.Context, _ string) (*heartbeat.Heartbeat, error) {
	return nil, assert.AnError
}

func (s *mockHBStore) getAll() []*heartbeat.Heartbeat {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := make([]*heartbeat.Heartbeat, len(s.data))
	copy(cp, s.data)
	return cp
}

type mockCheckerSched struct {
	mu      sync.Mutex
	calls   int
	results []CheckResult
}

func (c *mockCheckerSched) Type() string { return "http" }

func (c *mockCheckerSched) Check(_ context.Context, _ *Config) (CheckResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	idx := c.calls
	if idx >= len(c.results) {
		idx = len(c.results) - 1
	}
	c.calls++
	return c.results[idx], nil
}

func (c *mockCheckerSched) callCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.calls
}

func newTestScheduler(store *mockStore, hbStore *mockHBStore, checker *mockCheckerSched) *Scheduler {
	registry := NewRegistry()
	registry.Register(checker)
	return NewScheduler(store, hbStore, registry, nil, (*noopPublisher)(nil), nil, WithMaxWorkers(4))
}

type noopPublisher struct{}

func (p *noopPublisher) Publish(_ broadcast.Event) {}

func TestSchedulerImmediateCheck(t *testing.T) {
	store := &mockStore{monitors: map[string]*domainmonitor.Monitor{
		"m1": {ID: "m1", Name: "test", Type: "http", Interval: 3600, Timeout: 5, Active: true, ConfigJSON: `{"kind":"http","url":"http://example.com","method":"GET"}`},
	}}
	hbStore := &mockHBStore{}
	checker := &mockCheckerSched{results: []CheckResult{{Status: status.Up, Latency: 42, Message: "ok"}}}
	sched := newTestScheduler(store, hbStore, checker)

	require.NoError(t, sched.Start(t.Context()))
	time.Sleep(100 * time.Millisecond)
	sched.Stop()

	assert.GreaterOrEqual(t, checker.callCount(), 1)
	hbs := hbStore.getAll()
	require.NotEmpty(t, hbs)
	assert.Equal(t, int(status.Up), hbs[0].Status)
}

func TestSchedulerInterval(t *testing.T) {
	store := &mockStore{monitors: map[string]*domainmonitor.Monitor{
		"m1": {ID: "m1", Name: "test", Type: "http", Interval: 1, Timeout: 5, Active: true, ConfigJSON: `{"kind":"http","url":"http://example.com","method":"GET"}`},
	}}
	hbStore := &mockHBStore{}
	checker := &mockCheckerSched{results: []CheckResult{{Status: status.Up, Latency: 10, Message: "ok"}}}
	sched := newTestScheduler(store, hbStore, checker)

	require.NoError(t, sched.Start(t.Context()))
	// Wait for ~2 intervals (1s each) plus the immediate check
	time.Sleep(2500 * time.Millisecond)
	sched.Stop()

	assert.GreaterOrEqual(t, checker.callCount(), 3)
}

func TestSchedulerRetry(t *testing.T) {
	store := &mockStore{monitors: map[string]*domainmonitor.Monitor{
		"m1": {ID: "m1", Name: "retry", Type: "http", Interval: 3600, Timeout: 5, MaxRetries: 3, RetryInterval: 1, Active: true, ConfigJSON: `{"kind":"http","url":"http://example.com","method":"GET"}`},
	}}
	hbStore := &mockHBStore{}
	checker := &mockCheckerSched{results: []CheckResult{
		{Status: status.Down, Message: "fail1"},
		{Status: status.Down, Message: "fail2"},
		{Status: status.Up, Message: "recovered"},
	}}
	sched := newTestScheduler(store, hbStore, checker)

	require.NoError(t, sched.Start(t.Context()))
	time.Sleep(3500 * time.Millisecond)
	sched.Stop()

	assert.GreaterOrEqual(t, checker.callCount(), 3)
	hbs := hbStore.getAll()
	require.NotEmpty(t, hbs)
	assert.Equal(t, int(status.Up), hbs[0].Status)
	assert.Equal(t, 2, hbs[0].Retries)
}

func TestSchedulerRetryExhausted(t *testing.T) {
	store := &mockStore{monitors: map[string]*domainmonitor.Monitor{
		"m1": {ID: "m1", Name: "exhaust", Type: "http", Interval: 3600, Timeout: 5, MaxRetries: 2, RetryInterval: 1, Active: true, ConfigJSON: `{"kind":"http","url":"http://example.com","method":"GET"}`},
	}}
	hbStore := &mockHBStore{}
	checker := &mockCheckerSched{results: []CheckResult{{Status: status.Down, Message: "fail"}}}
	sched := newTestScheduler(store, hbStore, checker)

	require.NoError(t, sched.Start(t.Context()))
	time.Sleep(3500 * time.Millisecond)
	sched.Stop()

	hbs := hbStore.getAll()
	require.NotEmpty(t, hbs)
	assert.Equal(t, int(status.Down), hbs[0].Status)
	assert.Equal(t, 2, hbs[0].Retries)
}

func TestSchedulerRescheduleOnConfigChange(t *testing.T) {
	store := &mockStore{monitors: map[string]*domainmonitor.Monitor{
		"m1": {ID: "m1", Name: "slow", Type: "http", Interval: 3600, Timeout: 5, Active: true, ConfigJSON: `{"kind":"http","url":"http://example.com","method":"GET"}`},
	}}
	hbStore := &mockHBStore{}
	checker := &mockCheckerSched{results: []CheckResult{{Status: status.Up, Latency: 10, Message: "ok"}}}
	sched := newTestScheduler(store, hbStore, checker)

	ctx := t.Context()
	require.NoError(t, sched.Start(ctx))

	// Wait for immediate check
	time.Sleep(100 * time.Millisecond)
	countAfterFirst := checker.callCount()
	require.Equal(t, 1, countAfterFirst)

	// Change interval to 1 second and reschedule
	store.mu.Lock()
	store.monitors["m1"].Interval = 1
	store.mu.Unlock()

	require.NoError(t, sched.ScheduleMonitor(ctx, "m1"))

	// The reschedule should compute next fire from lastFire + new interval (1s)
	time.Sleep(1500 * time.Millisecond)
	sched.Stop()

	// Should have fired again with the new 1s interval
	assert.GreaterOrEqual(t, checker.callCount(), 2)
}

func TestSchedulerRemoveMonitor(t *testing.T) {
	store := &mockStore{monitors: map[string]*domainmonitor.Monitor{
		"m1": {ID: "m1", Name: "toremove", Type: "http", Interval: 1, Timeout: 5, Active: true, ConfigJSON: `{"kind":"http","url":"http://example.com","method":"GET"}`},
	}}
	hbStore := &mockHBStore{}
	checker := &mockCheckerSched{results: []CheckResult{{Status: status.Up, Latency: 10, Message: "ok"}}}
	sched := newTestScheduler(store, hbStore, checker)

	ctx := t.Context()
	require.NoError(t, sched.Start(ctx))
	time.Sleep(100 * time.Millisecond)

	countBefore := checker.callCount()
	sched.RemoveMonitor(ctx, "m1")
	time.Sleep(2 * time.Second)
	sched.Stop()

	// No more checks after removal
	assert.Equal(t, countBefore, checker.callCount())
}

func TestSchedulerCheckNow(t *testing.T) {
	store := &mockStore{monitors: map[string]*domainmonitor.Monitor{
		"m1": {ID: "m1", Name: "checknow", Type: "http", Interval: 3600, Timeout: 5, Active: true, ConfigJSON: `{"kind":"http","url":"http://example.com","method":"GET"}`},
	}}
	hbStore := &mockHBStore{}
	checker := &mockCheckerSched{results: []CheckResult{{Status: status.Up, Latency: 10, Message: "ok"}}}
	sched := newTestScheduler(store, hbStore, checker)

	ctx := t.Context()
	require.NoError(t, sched.Start(ctx))

	// Wait for immediate check
	time.Sleep(100 * time.Millisecond)
	require.Equal(t, 1, checker.callCount())

	// Trigger immediate check
	sched.CheckNow(ctx, "m1")
	time.Sleep(100 * time.Millisecond)
	sched.Stop()

	assert.Equal(t, 2, checker.callCount())
}
