package monitor

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/koblas/besops/internal/domain/heartbeat"
	domainmonitor "github.com/koblas/besops/internal/domain/monitor"
	"github.com/koblas/besops/lib/status"
)

type memMonitorStore struct {
	mu       sync.Mutex
	monitors map[string]*domainmonitor.Monitor
}

func (s *memMonitorStore) FindByID(_ context.Context, id string) (*domainmonitor.Monitor, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	m, ok := s.monitors[id]
	if !ok {
		return nil, fmt.Errorf("monitor %s not found", id)
	}
	return m, nil
}

func (s *memMonitorStore) FindAllActiveIDs(_ context.Context) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ids := make([]string, 0, len(s.monitors))
	for id := range s.monitors {
		ids = append(ids, id)
	}
	return ids, nil
}

type memHeartbeatStore struct {
	mu         sync.Mutex
	heartbeats map[string][]*heartbeat.Heartbeat
}

func (s *memHeartbeatStore) Insert(_ context.Context, hb *heartbeat.Heartbeat) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.heartbeats[hb.MonitorID] = append(s.heartbeats[hb.MonitorID], hb)
	return nil
}

func (s *memHeartbeatStore) GetLatest(_ context.Context, monitorID string) (*heartbeat.Heartbeat, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	hbs := s.heartbeats[monitorID]
	if len(hbs) == 0 {
		return nil, fmt.Errorf("no previous heartbeat")
	}
	return hbs[len(hbs)-1], nil
}

func (s *memHeartbeatStore) getAll(monitorID string) []*heartbeat.Heartbeat {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := make([]*heartbeat.Heartbeat, len(s.heartbeats[monitorID]))
	copy(cp, s.heartbeats[monitorID])
	return cp
}

type memNotifyDispatcher struct {
	mu     sync.Mutex
	events []notifyEvent
}

type notifyEvent struct {
	monitorID string
	current   status.Status
	previous  status.Status
	msg       string
}

func (d *memNotifyDispatcher) Dispatch(_ context.Context, monitorID string, current status.Status, previous status.Status, msg string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.events = append(d.events, notifyEvent{monitorID: monitorID, current: current, previous: previous, msg: msg})
}

func (d *memNotifyDispatcher) getEvents() []notifyEvent {
	d.mu.Lock()
	defer d.mu.Unlock()
	cp := make([]notifyEvent, len(d.events))
	copy(cp, d.events)
	return cp
}

type alwaysUpChecker struct{}

func (c *alwaysUpChecker) Type() string { return "test" }
func (c *alwaysUpChecker) Check(_ context.Context, _ *Config) (CheckResult, error) {
	return CheckResult{Status: status.Up, Latency: 5, Message: "ok"}, nil
}

func TestManagerStartAndStop(t *testing.T) {
	store := &memMonitorStore{monitors: map[string]*domainmonitor.Monitor{
		"m1": {ID: "m1", Name: "Test 1", Type: "test", Interval: 60, Timeout: 5, Active: true},
		"m2": {ID: "m2", Name: "Test 2", Type: "test", Interval: 60, Timeout: 5, Active: true},
	}}
	hbStore := &memHeartbeatStore{heartbeats: make(map[string][]*heartbeat.Heartbeat)}
	notify := &memNotifyDispatcher{}

	registry := NewRegistry()
	registry.Register(&alwaysUpChecker{})

	mgr := NewManager(store, hbStore, registry, notify, nil, nil)

	require.NoError(t, mgr.Start(t.Context()))
	time.Sleep(50 * time.Millisecond)

	assert.Equal(t, 2, mgr.RunningCount())
	assert.True(t, mgr.IsRunning("m1"))
	assert.True(t, mgr.IsRunning("m2"))

	mgr.Stop()

	hbs := hbStore.getAll("m1")
	require.NotEmpty(t, hbs)
	assert.Equal(t, int(status.Up), hbs[0].Status)
}

func TestManagerStopMonitor(t *testing.T) {
	store := &memMonitorStore{monitors: map[string]*domainmonitor.Monitor{
		"m1": {ID: "m1", Name: "Test", Type: "test", Interval: 60, Timeout: 5, Active: true},
	}}
	hbStore := &memHeartbeatStore{heartbeats: make(map[string][]*heartbeat.Heartbeat)}
	registry := NewRegistry()
	registry.Register(&alwaysUpChecker{})

	mgr := NewManager(store, hbStore, registry, nil, nil, nil)
	require.NoError(t, mgr.Start(t.Context()))
	time.Sleep(30 * time.Millisecond)

	mgr.StopMonitor(t.Context(), "m1")

	assert.False(t, mgr.IsRunning("m1"))
	assert.Equal(t, 0, mgr.RunningCount())

	mgr.Stop()
}

func TestManagerRestartMonitor(t *testing.T) {
	store := &memMonitorStore{monitors: map[string]*domainmonitor.Monitor{
		"m1": {ID: "m1", Name: "Test", Type: "test", Interval: 60, Timeout: 5, Active: true},
	}}
	hbStore := &memHeartbeatStore{heartbeats: make(map[string][]*heartbeat.Heartbeat)}
	registry := NewRegistry()
	registry.Register(&alwaysUpChecker{})

	mgr := NewManager(store, hbStore, registry, nil, nil, nil)
	require.NoError(t, mgr.Start(t.Context()))
	time.Sleep(30 * time.Millisecond)

	require.NoError(t, mgr.RestartMonitor(t.Context(), "m1"))
	time.Sleep(30 * time.Millisecond)

	assert.True(t, mgr.IsRunning("m1"))

	mgr.Stop()
}

func TestManagerNotifiesOnStatusChange(t *testing.T) {
	store := &memMonitorStore{monitors: map[string]*domainmonitor.Monitor{
		"m1": {ID: "m1", Name: "Flip", Type: "test", Interval: 60, Timeout: 5, Active: true},
	}}
	hbStore := &memHeartbeatStore{heartbeats: make(map[string][]*heartbeat.Heartbeat)}
	notify := &memNotifyDispatcher{}
	registry := NewRegistry()
	registry.Register(&alwaysUpChecker{})

	mgr := NewManager(store, hbStore, registry, notify, nil, nil)
	require.NoError(t, mgr.Start(t.Context()))
	time.Sleep(30 * time.Millisecond)

	hbStore.mu.Lock()
	hbStore.heartbeats["m1"] = []*heartbeat.Heartbeat{{
		MonitorID: "m1",
		Status:    int(status.Down),
		Time:      heartbeat.RFC3339Time(time.Now().Add(-time.Minute)),
		Msg:       "was down",
	}}
	hbStore.mu.Unlock()

	require.NoError(t, mgr.RestartMonitor(t.Context(), "m1"))
	time.Sleep(50 * time.Millisecond)
	mgr.Stop()

	events := notify.getEvents()
	require.NotEmpty(t, events, "expected notification on status change from Down to Up")
	assert.Equal(t, status.Up, events[0].current)
	assert.Equal(t, status.Down, events[0].previous)
}

func TestModelToConfigParsesHeaders(t *testing.T) {
	mon := &domainmonitor.Monitor{
		ID:         "m1",
		Name:       "HTTP Test",
		Type:       "http",
		Interval:   60,
		Timeout:    10,
		ConfigJSON: `{"url":"https://example.com","method":"POST","headers":[{"name":"Content-Type","value":"application/json"},{"name":"Authorization","value":"Bearer tok"}],"body":"{\"ping\":true}"}`,
	}

	cfg := modelToConfig(mon)

	require.Len(t, cfg.HTTP.Headers, 2)
	assert.Equal(t, "Content-Type", cfg.HTTP.Headers[0].Name)
	assert.Equal(t, "application/json", cfg.HTTP.Headers[0].Value)
	assert.Equal(t, "Authorization", cfg.HTTP.Headers[1].Name)
	assert.Equal(t, "Bearer tok", cfg.HTTP.Headers[1].Value)
}

func TestManagerUnknownType(t *testing.T) {
	store := &memMonitorStore{monitors: map[string]*domainmonitor.Monitor{
		"m1": {ID: "m1", Name: "Bad", Type: "nonexistent", Interval: 60, Timeout: 5, Active: true},
	}}
	hbStore := &memHeartbeatStore{heartbeats: make(map[string][]*heartbeat.Heartbeat)}
	registry := NewRegistry()

	mgr := NewManager(store, hbStore, registry, nil, nil, nil)
	require.NoError(t, mgr.Start(t.Context()), "Start should not fail for individual monitor errors")

	time.Sleep(30 * time.Millisecond)
	assert.False(t, mgr.IsRunning("m1"), "monitor with unknown type should not be running")

	mgr.Stop()
}
