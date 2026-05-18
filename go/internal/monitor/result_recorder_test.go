package monitor

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/koblas/besops/internal/domain/heartbeat"
	"github.com/koblas/besops/lib/status"
	"github.com/koblas/besops/lib/telemetry"
)

type capturedMetric struct {
	monitor   telemetry.MonitorInfo
	up        bool
	latencyMs int64
}

type capturingObserver struct {
	mu      sync.Mutex
	records []capturedMetric
}

func (o *capturingObserver) Record(_ context.Context, monitor telemetry.MonitorInfo, up bool, latencyMs int64) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.records = append(o.records, capturedMetric{monitor: monitor, up: up, latencyMs: latencyMs})
}

func (o *capturingObserver) getRecords() []capturedMetric {
	o.mu.Lock()
	defer o.mu.Unlock()
	cp := make([]capturedMetric, len(o.records))
	copy(cp, o.records)
	return cp
}

type staticMaintenanceChecker struct {
	inMaintenance bool
}

func (c *staticMaintenanceChecker) IsMonitorInMaintenance(_ context.Context, _ string) (bool, error) {
	return c.inMaintenance, nil
}

// --- Metrics Recording ---

func TestHandleResult_RecordsMetrics_WhenUp(t *testing.T) {
	hbStore := &memHeartbeatStore{heartbeats: make(map[string][]*heartbeat.Heartbeat)}
	observer := &capturingObserver{}

	recorder := &resultRecorder{
		hbStore:     hbStore,
		metrics:     observer,
		monitorName: "Web App",
		monitorType: "http",
		groupName:   "Production",
		tags:        []string{"critical", "env:prod"},
	}

	recorder.HandleResult(t.Context(), "mon-1", CheckResult{
		Status:  status.Up,
		Latency: 42,
		Message: "200 OK",
	}, 0)

	records := observer.getRecords()
	require.Len(t, records, 1)

	assert.True(t, records[0].up)
	assert.Equal(t, int64(42), records[0].latencyMs)
	assert.Equal(t, "mon-1", records[0].monitor.MonitorID())
	assert.Equal(t, "Web App", records[0].monitor.MonitorName())
	assert.Equal(t, "http", records[0].monitor.MonitorType())
	assert.Equal(t, "Production", records[0].monitor.GroupName())
	assert.Equal(t, []string{"critical", "env:prod"}, records[0].monitor.Tags())
}

func TestHandleResult_RecordsMetrics_WhenDown(t *testing.T) {
	hbStore := &memHeartbeatStore{heartbeats: make(map[string][]*heartbeat.Heartbeat)}
	observer := &capturingObserver{}

	recorder := &resultRecorder{
		hbStore:     hbStore,
		metrics:     observer,
		monitorName: "DNS Check",
		monitorType: "dns",
	}

	recorder.HandleResult(t.Context(), "mon-2", CheckResult{
		Status:  status.Down,
		Latency: 0,
		Message: "timeout",
	}, 1)

	records := observer.getRecords()
	require.Len(t, records, 1)

	assert.False(t, records[0].up)
	assert.Equal(t, int64(0), records[0].latencyMs)
	assert.Equal(t, "mon-2", records[0].monitor.MonitorID())
	assert.Equal(t, "dns", records[0].monitor.MonitorType())
}

func TestHandleResult_NoMetrics_WhenObserverNil(t *testing.T) {
	hbStore := &memHeartbeatStore{heartbeats: make(map[string][]*heartbeat.Heartbeat)}

	recorder := &resultRecorder{
		hbStore: hbStore,
		metrics: nil,
	}

	// Should not panic
	recorder.HandleResult(t.Context(), "mon-1", CheckResult{
		Status: status.Up, Latency: 10, Message: "ok",
	}, 0)

	hbs := hbStore.getAll("mon-1")
	require.Len(t, hbs, 1, "heartbeat should still be stored even without metrics")
}

func TestHandleResult_MetricsLatency_ZeroWhenNoPing(t *testing.T) {
	hbStore := &memHeartbeatStore{heartbeats: make(map[string][]*heartbeat.Heartbeat)}
	observer := &capturingObserver{}

	recorder := &resultRecorder{
		hbStore: hbStore,
		metrics: observer,
	}

	recorder.HandleResult(t.Context(), "mon-1", CheckResult{
		Status:  status.Up,
		Latency: 0,
		Message: "ok",
	}, 0)

	records := observer.getRecords()
	require.Len(t, records, 1)
	assert.Equal(t, int64(0), records[0].latencyMs)
}

// --- Maintenance Window ---

func TestHandleResult_Maintenance_OverridesStatusToMaintenance(t *testing.T) {
	hbStore := &memHeartbeatStore{heartbeats: make(map[string][]*heartbeat.Heartbeat)}

	recorder := &resultRecorder{
		hbStore: hbStore,
		maint:   &staticMaintenanceChecker{inMaintenance: true},
	}

	recorder.HandleResult(t.Context(), "mon-1", CheckResult{
		Status:  status.Down,
		Latency: 0,
		Message: "connection refused",
	}, 0)

	hbs := hbStore.getAll("mon-1")
	require.Len(t, hbs, 1)
	assert.Equal(t, int(status.Maintenance), hbs[0].Status, "status should be overridden to Maintenance")
}

func TestHandleResult_Maintenance_SuppressesNotification(t *testing.T) {
	hbStore := &memHeartbeatStore{heartbeats: make(map[string][]*heartbeat.Heartbeat)}
	notify := &memNotifyDispatcher{}

	// Seed a previous "Up" heartbeat so a Down would normally trigger notification
	hbStore.heartbeats["mon-1"] = []*heartbeat.Heartbeat{{
		MonitorID: "mon-1",
		Status:    int(status.Up),
		Time:      heartbeat.RFC3339Time(time.Now().Add(-time.Minute)),
	}}

	recorder := &resultRecorder{
		hbStore: hbStore,
		notify:  notify,
		maint:   &staticMaintenanceChecker{inMaintenance: true},
	}

	recorder.HandleResult(t.Context(), "mon-1", CheckResult{
		Status:  status.Down,
		Message: "connection refused",
	}, 0)

	events := notify.getEvents()
	assert.Empty(t, events, "notification should be suppressed during maintenance")
}

func TestHandleResult_NotInMaintenance_DispatchesNotification(t *testing.T) {
	hbStore := &memHeartbeatStore{heartbeats: make(map[string][]*heartbeat.Heartbeat)}
	notify := &memNotifyDispatcher{}

	// Seed previous Up heartbeat
	hbStore.heartbeats["mon-1"] = []*heartbeat.Heartbeat{{
		MonitorID: "mon-1",
		Status:    int(status.Up),
		Time:      heartbeat.RFC3339Time(time.Now().Add(-time.Minute)),
	}}

	recorder := &resultRecorder{
		hbStore: hbStore,
		notify:  notify,
		maint:   &staticMaintenanceChecker{inMaintenance: false},
	}

	recorder.HandleResult(t.Context(), "mon-1", CheckResult{
		Status:  status.Down,
		Message: "timeout",
	}, 0)

	events := notify.getEvents()
	require.Len(t, events, 1)
	assert.Equal(t, status.Down, events[0].current)
	assert.Equal(t, status.Up, events[0].previous)
}

// --- Tags Flow Through to Metrics ---

func TestHandleResult_TagsPassedToMetrics(t *testing.T) {
	hbStore := &memHeartbeatStore{heartbeats: make(map[string][]*heartbeat.Heartbeat)}
	observer := &capturingObserver{}

	tags := []string{"region:us-east", "tier:premium", "env:staging"}

	recorder := &resultRecorder{
		hbStore:     hbStore,
		metrics:     observer,
		monitorName: "API",
		monitorType: "http",
		tags:        tags,
	}

	recorder.HandleResult(t.Context(), "mon-1", CheckResult{
		Status: status.Up, Latency: 15, Message: "ok",
	}, 0)

	records := observer.getRecords()
	require.Len(t, records, 1)
	assert.Equal(t, tags, records[0].monitor.Tags())
}

func TestHandleResult_NilTags_PassesNilToMetrics(t *testing.T) {
	hbStore := &memHeartbeatStore{heartbeats: make(map[string][]*heartbeat.Heartbeat)}
	observer := &capturingObserver{}

	recorder := &resultRecorder{
		hbStore: hbStore,
		metrics: observer,
		tags:    nil,
	}

	recorder.HandleResult(t.Context(), "mon-1", CheckResult{
		Status: status.Up, Latency: 5, Message: "ok",
	}, 0)

	records := observer.getRecords()
	require.Len(t, records, 1)
	assert.Nil(t, records[0].monitor.Tags())
}
