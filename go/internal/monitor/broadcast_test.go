package monitor

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/koblas/besops/internal/broadcast"
	"github.com/koblas/besops/internal/domain/heartbeat"
	domainmonitor "github.com/koblas/besops/internal/domain/monitor"
	"github.com/koblas/besops/lib/status"
)

func TestResultRecorderPublishesToHub(t *testing.T) {
	hbStore := &memHeartbeatStore{heartbeats: make(map[string][]*heartbeat.Heartbeat)}
	hub := broadcast.NewHub(8)

	recorder := &resultRecorder{
		hbStore:   hbStore,
		notify:    nil,
		publisher: hub,
	}

	sub := hub.Subscribe()
	defer hub.Unsubscribe(sub)

	recorder.HandleResult(t.Context(), "m1", CheckResult{
		Status:  status.Up,
		Ping:    12,
		Message: "ok",
	}, 0)

	select {
	case ev := <-sub:
		assert.Equal(t, "heartbeat", ev.Type)
		hb, ok := ev.Data.(*heartbeat.Heartbeat)
		require.True(t, ok, "expected *heartbeat.Heartbeat, got %T", ev.Data)
		assert.Equal(t, "m1", hb.MonitorID)
		assert.Equal(t, int(status.Up), hb.Status)
		assert.Equal(t, int64(12), *hb.Ping)
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for broadcast event")
	}
}

func TestManagerPublishesOnCheck(t *testing.T) {
	store := &memMonitorStore{monitors: map[string]*domainmonitor.Monitor{
		"m1": {ID: "m1", Name: "Pub Test", Type: "test", Interval: 60, Timeout: 5, Active: true},
	}}
	hbStore := &memHeartbeatStore{heartbeats: make(map[string][]*heartbeat.Heartbeat)}

	hub := broadcast.NewHub(8)
	sub := hub.Subscribe()
	defer hub.Unsubscribe(sub)

	registry := NewRegistry()
	registry.Register(&alwaysUpChecker{})

	mgr := NewManager(store, hbStore, registry, nil, hub, nil)
	require.NoError(t, mgr.Start(t.Context()))

	select {
	case ev := <-sub:
		assert.Equal(t, "heartbeat", ev.Type)
		hb, ok := ev.Data.(*heartbeat.Heartbeat)
		require.True(t, ok)
		assert.Equal(t, "m1", hb.MonitorID)
		assert.Equal(t, int(status.Up), hb.Status)
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for broadcast from monitor manager")
	}

	mgr.Stop()
}
