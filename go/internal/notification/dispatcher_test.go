package notification

import (
	"context"
	"fmt"
	"sync"
	"testing"
)

type mockNotifier struct {
	name    string
	mu      sync.Mutex
	calls   []mockSendCall
	failErr error
}

type mockSendCall struct {
	config  map[string]any
	msg     string
	monitor *MonitorInfo
}

func (n *mockNotifier) Name() string { return n.name }

func (n *mockNotifier) Send(_ context.Context, config map[string]any, msg string, monitor *MonitorInfo, _ *HeartbeatInfo) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.calls = append(n.calls, mockSendCall{config: config, msg: msg, monitor: monitor})
	return n.failErr
}

func (n *mockNotifier) getCalls() []mockSendCall {
	n.mu.Lock()
	defer n.mu.Unlock()
	cp := make([]mockSendCall, len(n.calls))
	copy(cp, n.calls)
	return cp
}

type mockRuleStore struct {
	rules map[string][]Rule
}

func (s *mockRuleStore) GetRulesForMonitor(_ context.Context, monitorID string) ([]Rule, error) {
	return s.rules[monitorID], nil
}

func TestDispatcherSendsToAllActive(t *testing.T) {
	slack := &mockNotifier{name: "slack"}
	discord := &mockNotifier{name: "discord"}

	registry := NewRegistry()
	registry.Register(slack)
	registry.Register(discord)

	store := &mockRuleStore{rules: map[string][]Rule{
		"m1": {
			{ID: "n1", Name: "Slack Alert", Type: "slack", Config: map[string]any{"url": "x"}, Active: true},
			{ID: "n2", Name: "Discord Alert", Type: "discord", Config: map[string]any{"url": "y"}, Active: true},
		},
	}}

	dispatcher := NewDispatcher(registry, store)
	monitor := &MonitorInfo{Name: "Test Monitor"}
	heartbeat := &HeartbeatInfo{Status: 1, Message: "ok"}

	dispatcher.Dispatch(t.Context(), "m1", monitor, heartbeat, "Service is up")

	slackCalls := slack.getCalls()
	if len(slackCalls) != 1 {
		t.Fatalf("expected 1 slack call, got %d", len(slackCalls))
	}
	if slackCalls[0].msg != "Service is up" {
		t.Errorf("expected 'Service is up', got %s", slackCalls[0].msg)
	}

	discordCalls := discord.getCalls()
	if len(discordCalls) != 1 {
		t.Fatalf("expected 1 discord call, got %d", len(discordCalls))
	}
}

func TestDispatcherSkipsInactive(t *testing.T) {
	slack := &mockNotifier{name: "slack"}

	registry := NewRegistry()
	registry.Register(slack)

	store := &mockRuleStore{rules: map[string][]Rule{
		"m1": {
			{ID: "n1", Name: "Inactive", Type: "slack", Config: map[string]any{}, Active: false},
		},
	}}

	dispatcher := NewDispatcher(registry, store)
	dispatcher.Dispatch(t.Context(), "m1", nil, nil, "test")

	if len(slack.getCalls()) != 0 {
		t.Error("expected no calls to inactive provider")
	}
}

func TestDispatcherHandlesUnknownProvider(t *testing.T) {
	registry := NewRegistry()
	store := &mockRuleStore{rules: map[string][]Rule{
		"m1": {
			{ID: "n1", Name: "Unknown", Type: "nonexistent", Config: map[string]any{}, Active: true},
		},
	}}

	dispatcher := NewDispatcher(registry, store)
	// Should not panic
	dispatcher.Dispatch(t.Context(), "m1", nil, nil, "test")
}

func TestDispatcherContinuesOnError(t *testing.T) {
	failing := &mockNotifier{name: "failing", failErr: fmt.Errorf("send failed")}
	working := &mockNotifier{name: "working"}

	registry := NewRegistry()
	registry.Register(failing)
	registry.Register(working)

	store := &mockRuleStore{rules: map[string][]Rule{
		"m1": {
			{ID: "n1", Name: "Fail", Type: "failing", Config: map[string]any{}, Active: true},
			{ID: "n2", Name: "Work", Type: "working", Config: map[string]any{}, Active: true},
		},
	}}

	dispatcher := NewDispatcher(registry, store)
	dispatcher.Dispatch(t.Context(), "m1", nil, nil, "test")

	if len(working.getCalls()) != 1 {
		t.Error("working provider should still be called even when another fails")
	}
}

func TestSendTest(t *testing.T) {
	slack := &mockNotifier{name: "slack"}
	registry := NewRegistry()
	registry.Register(slack)

	dispatcher := NewDispatcher(registry, nil)
	err := dispatcher.SendTest(t.Context(), "slack", map[string]any{"url": "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	calls := slack.getCalls()
	if len(calls) != 1 {
		t.Fatal("expected 1 call")
	}
	if calls[0].monitor.Name != "Test Monitor" {
		t.Errorf("expected Test Monitor, got %s", calls[0].monitor.Name)
	}
}

func TestSendTestUnknownProvider(t *testing.T) {
	registry := NewRegistry()
	dispatcher := NewDispatcher(registry, nil)

	err := dispatcher.SendTest(t.Context(), "unknown", nil)
	if err == nil {
		t.Error("expected error for unknown provider")
	}
}
