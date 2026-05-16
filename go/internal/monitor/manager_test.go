package monitor

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

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
	return CheckResult{Status: status.Up, Ping: 5, Message: "ok"}, nil
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

	mgr := NewManager(store, hbStore, registry, notify, nil)

	ctx := t.Context()
	if err := mgr.Start(ctx); err != nil {
		t.Fatalf("start: %v", err)
	}

	// Wait for immediate checks to fire
	time.Sleep(50 * time.Millisecond)

	if mgr.RunningCount() != 2 {
		t.Errorf("expected 2 running, got %d", mgr.RunningCount())
	}

	if !mgr.IsRunning("m1") || !mgr.IsRunning("m2") {
		t.Error("expected both monitors running")
	}

	mgr.Stop()

	// Verify heartbeats were stored
	hbs := hbStore.getAll("m1")
	if len(hbs) == 0 {
		t.Error("expected heartbeats for m1")
	}
	if hbs[0].Status != int(status.Up) {
		t.Errorf("expected Up heartbeat, got %d", hbs[0].Status)
	}
}

func TestManagerStopMonitor(t *testing.T) {
	store := &memMonitorStore{monitors: map[string]*domainmonitor.Monitor{
		"m1": {ID: "m1", Name: "Test", Type: "test", Interval: 60, Timeout: 5, Active: true},
	}}
	hbStore := &memHeartbeatStore{heartbeats: make(map[string][]*heartbeat.Heartbeat)}
	registry := NewRegistry()
	registry.Register(&alwaysUpChecker{})

	mgr := NewManager(store, hbStore, registry, nil, nil)
	if err := mgr.Start(t.Context()); err != nil {
		t.Fatal(err)
	}
	time.Sleep(30 * time.Millisecond)

	mgr.StopMonitor(t.Context(), "m1")

	if mgr.IsRunning("m1") {
		t.Error("m1 should be stopped")
	}
	if mgr.RunningCount() != 0 {
		t.Errorf("expected 0 running, got %d", mgr.RunningCount())
	}

	mgr.Stop()
}

func TestManagerRestartMonitor(t *testing.T) {
	store := &memMonitorStore{monitors: map[string]*domainmonitor.Monitor{
		"m1": {ID: "m1", Name: "Test", Type: "test", Interval: 60, Timeout: 5, Active: true},
	}}
	hbStore := &memHeartbeatStore{heartbeats: make(map[string][]*heartbeat.Heartbeat)}
	registry := NewRegistry()
	registry.Register(&alwaysUpChecker{})

	mgr := NewManager(store, hbStore, registry, nil, nil)
	if err := mgr.Start(t.Context()); err != nil {
		t.Fatal(err)
	}
	time.Sleep(30 * time.Millisecond)

	if err := mgr.RestartMonitor(t.Context(), "m1"); err != nil {
		t.Fatalf("restart: %v", err)
	}

	time.Sleep(30 * time.Millisecond)
	if !mgr.IsRunning("m1") {
		t.Error("m1 should be running after restart")
	}

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

	mgr := NewManager(store, hbStore, registry, notify, nil)
	if err := mgr.Start(t.Context()); err != nil {
		t.Fatal(err)
	}

	// Wait for the first check to complete
	time.Sleep(30 * time.Millisecond)

	// Replace heartbeats with a single "down" so the next check (Up) triggers a notification
	hbStore.mu.Lock()
	hbStore.heartbeats["m1"] = []*heartbeat.Heartbeat{{
		MonitorID: "m1",
		Status:    int(status.Down),
		Time:      heartbeat.RFC3339Time(time.Now().Add(-time.Minute)),
		Msg:       "was down",
	}}
	hbStore.mu.Unlock()

	// Restart to trigger a new immediate check
	if err := mgr.RestartMonitor(t.Context(), "m1"); err != nil {
		t.Fatal(err)
	}
	time.Sleep(50 * time.Millisecond)
	mgr.Stop()

	events := notify.getEvents()
	if len(events) == 0 {
		t.Error("expected notification on status change from Down to Up")
	} else {
		if events[0].current != status.Up || events[0].previous != status.Down {
			t.Errorf("unexpected transition: %v -> %v", events[0].previous, events[0].current)
		}
	}
}

func TestManagerUnknownType(t *testing.T) {
	store := &memMonitorStore{monitors: map[string]*domainmonitor.Monitor{
		"m1": {ID: "m1", Name: "Bad", Type: "nonexistent", Interval: 60, Timeout: 5, Active: true},
	}}
	hbStore := &memHeartbeatStore{heartbeats: make(map[string][]*heartbeat.Heartbeat)}
	registry := NewRegistry()

	mgr := NewManager(store, hbStore, registry, nil, nil)
	err := mgr.Start(t.Context())
	if err != nil {
		t.Fatal("Start should not fail for individual monitor errors")
	}

	// Monitor with unknown type should not be running
	time.Sleep(30 * time.Millisecond)
	if mgr.IsRunning("m1") {
		t.Error("monitor with unknown type should not be running")
	}

	mgr.Stop()
}
