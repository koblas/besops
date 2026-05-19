package monitor

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"time"

	"github.com/koblas/besops/internal/broadcast"
	"github.com/koblas/besops/internal/domain/heartbeat"
	domainmonitor "github.com/koblas/besops/internal/domain/monitor"
	"github.com/koblas/besops/lib/status"
	"github.com/koblas/besops/lib/telemetry"
)

// Store loads monitor configuration from the database.
type Store interface {
	FindByID(ctx context.Context, id string) (*domainmonitor.Monitor, error)
	FindAllActiveIDs(ctx context.Context) ([]string, error)
}

// HeartbeatStore persists check results.
type HeartbeatStore interface {
	Insert(ctx context.Context, hb *heartbeat.Heartbeat) error
	GetLatest(ctx context.Context, monitorID string) (*heartbeat.Heartbeat, error)
}

// NotificationDispatcher sends alerts when status changes.
type NotificationDispatcher interface {
	Dispatch(ctx context.Context, monitorID string, current status.Status, previous status.Status, msg string)
}

// MetricsObserver receives every check result for metrics/telemetry export.
type MetricsObserver = telemetry.Observer

// TagProvider loads tag names for a monitor.
type TagProvider interface {
	GetTagsForMonitor(ctx context.Context, monitorID string) ([]string, error)
}

// MaintenanceChecker determines if a monitor is currently in a maintenance window.
type MaintenanceChecker interface {
	IsMonitorInMaintenance(ctx context.Context, monitorID string) (bool, error)
}

// Manager controls the lifecycle of all monitor runners via a priority-queue scheduler.
type Manager struct {
	scheduler *Scheduler
}

type ManagerOption func(*managerOpts)

type managerOpts struct {
	maxWorkers int
	metrics    MetricsObserver
	tags       TagProvider
}

func WithManagerMaxWorkers(n int) ManagerOption {
	return func(o *managerOpts) { o.maxWorkers = n }
}

func WithManagerMetrics(m MetricsObserver) ManagerOption {
	return func(o *managerOpts) { o.metrics = m }
}

func WithManagerTags(t TagProvider) ManagerOption {
	return func(o *managerOpts) { o.tags = t }
}

func NewManager(store Store, hbStore HeartbeatStore, registry *Registry, notify NotificationDispatcher, publisher broadcast.Publisher, maint MaintenanceChecker, opts ...ManagerOption) *Manager {
	o := &managerOpts{maxWorkers: 32}
	for _, opt := range opts {
		opt(o)
	}

	schedOpts := []SchedulerOption{WithMaxWorkers(o.maxWorkers)}
	if o.metrics != nil {
		schedOpts = append(schedOpts, WithMetrics(o.metrics))
	}
	if o.tags != nil {
		schedOpts = append(schedOpts, WithTags(o.tags))
	}

	sched := NewScheduler(store, hbStore, registry, notify, publisher, maint, schedOpts...)
	return &Manager{scheduler: sched}
}

func (m *Manager) Start(ctx context.Context) error {
	return m.scheduler.Start(ctx)
}

// Stop gracefully shuts down all running monitors.
func (m *Manager) Stop() {
	m.scheduler.Stop()
}

// StartMonitor adds or resumes a monitor in the schedule.
func (m *Manager) StartMonitor(ctx context.Context, id string) error {
	return m.scheduler.ScheduleMonitor(ctx, id)
}

// StopMonitor removes a monitor from the schedule.
func (m *Manager) StopMonitor(ctx context.Context, id string) {
	m.scheduler.RemoveMonitor(ctx, id)
}

// RestartMonitor reschedules a monitor with fresh configuration.
// Uses last fire time to determine when the next check should occur.
func (m *Manager) RestartMonitor(ctx context.Context, id string) error {
	return m.scheduler.ScheduleMonitor(ctx, id)
}

// CheckNow queues an immediate check for a monitor.
func (m *Manager) CheckNow(ctx context.Context, id string) {
	m.scheduler.CheckNow(ctx, id)
}

// IsRunning returns true if a monitor is currently scheduled.
func (m *Manager) IsRunning(id string) bool {
	m.scheduler.mu.Lock()
	defer m.scheduler.mu.Unlock()
	_, exists := m.scheduler.configs[id]
	return exists
}

// RunningCount returns the number of scheduled monitors.
func (m *Manager) RunningCount() int {
	m.scheduler.mu.Lock()
	defer m.scheduler.mu.Unlock()
	return len(m.scheduler.configs)
}

func modelToConfig(mon *domainmonitor.Monitor) *Config {
	cfg := &Config{
		ID:            mon.ID,
		Type:          mon.Type,
		Name:          mon.Name,
		Interval:      time.Duration(mon.Interval) * time.Second,
		Timeout:       time.Duration(mon.Timeout) * time.Second,
		MaxRetries:    mon.MaxRetries,
		RetryInterval: time.Duration(mon.RetryInterval) * time.Second,
	}

	if mon.ConfigJSON == "" || mon.ConfigJSON == "{}" {
		return cfg
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(mon.ConfigJSON), &raw); err != nil {
		return cfg
	}

	cfg.Hostname = jsonString(raw, "hostname")
	cfg.Port = jsonInt(raw, "port")
	cfg.IgnoreTLS = jsonBool(raw, "ignoreTls")
	cfg.Keyword = jsonString(raw, "keyword")
	cfg.JsonPath = jsonString(raw, "jsonPath")
	cfg.ExpectedValue = jsonString(raw, "expectedValue")

	if v := jsonString(raw, "proxyId"); v != "" {
		cfg.ProxyID = &v
	}

	switch mon.Type {
	case "http":
		cfg.URL = jsonString(raw, "url")
		cfg.HTTP.Method = jsonString(raw, "method")
		cfg.HTTP.Body = jsonString(raw, "body")
		cfg.HTTP.BasicAuthUser = jsonString(raw, "basicAuthUser")
		cfg.HTTP.BasicAuthPass = jsonString(raw, "basicAuthPass")
		if h, ok := raw["headers"]; ok {
			cfg.HTTP.Headers = jsonHeaderPairs(h)
		}
		if sc, ok := raw["acceptedStatusCodes"]; ok {
			_ = json.Unmarshal(sc, &cfg.HTTP.AcceptedStatusCodes)
		}
		if jsonBool(raw, "invertKeyword") {
			cfg.KeywordType = "not contain"
		} else if cfg.Keyword != "" {
			cfg.KeywordType = "contain"
		}
	case "dns":
		cfg.DNS.ResolveType = jsonString(raw, "dnsResolveType")
	case "mqtt":
		cfg.MQTT.Topic = jsonString(raw, "mqttTopic")
		cfg.MQTT.Username = jsonString(raw, "mqttUsername")
		cfg.MQTT.Password = jsonString(raw, "mqttPassword")
		cfg.MQTT.SuccessMessage = jsonString(raw, "mqttSuccessMessage")
	case "grpc-keyword":
		cfg.GRPC.URL = jsonString(raw, "grpcUrl")
		cfg.GRPC.ServiceName = jsonString(raw, "grpcServiceName")
		cfg.GRPC.Method = jsonString(raw, "grpcMethod")
		cfg.GRPC.EnableTLS = jsonBool(raw, "grpcEnableTls")
	case "redis":
		cfg.Redis.ConnectionString = jsonString(raw, "databaseQuery")
	case "group":
		if tids, ok := raw["tagIds"]; ok {
			_ = json.Unmarshal(tids, &cfg.GroupTagIDs)
		}
	}

	return cfg
}

func jsonString(raw map[string]json.RawMessage, key string) string {
	v, ok := raw[key]
	if !ok {
		return ""
	}
	var s string
	if err := json.Unmarshal(v, &s); err != nil {
		return ""
	}
	return s
}

func jsonInt(raw map[string]json.RawMessage, key string) int {
	v, ok := raw[key]
	if !ok {
		return 0
	}
	var n int
	if err := json.Unmarshal(v, &n); err != nil {
		return 0
	}
	return n
}

func jsonHeaderPairs(data json.RawMessage) []HeaderPair {
	var raw []struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil
	}
	pairs := make([]HeaderPair, len(raw))
	for i, r := range raw {
		pairs[i] = HeaderPair{Name: r.Name, Value: r.Value}
	}
	return pairs
}

func jsonBool(raw map[string]json.RawMessage, key string) bool {
	v, ok := raw[key]
	if !ok {
		return false
	}
	var b bool
	if err := json.Unmarshal(v, &b); err != nil {
		return false
	}
	return b
}

// resultRecorder implements ResultHandler by persisting heartbeats and dispatching notifications.
type resultRecorder struct {
	hbStore     HeartbeatStore
	notify      NotificationDispatcher
	publisher   broadcast.Publisher
	maint       MaintenanceChecker
	metrics     MetricsObserver
	monitorName string
	monitorType string
	tags        []string
}

func (r *resultRecorder) HandleResult(ctx context.Context, monitorID string, result CheckResult, retries int) {
	slog.DebugContext(ctx, "recording result", slog.String("monitor", monitorID), slog.String("status", result.Status.String()), slog.Int64("latency_ms", result.Latency), slog.Int("retries", retries))

	// Check if monitor is in a maintenance window — if so, override status and suppress notifications.
	inMaintenance := false
	if r.maint != nil {
		var err error
		inMaintenance, err = r.maint.IsMonitorInMaintenance(ctx, monitorID)
		if err != nil {
			slog.WarnContext(ctx, "maintenance check failed, proceeding normally", slog.String("monitor", monitorID), slog.Any("error", err))
		}
	}

	recordedStatus := result.Status
	if inMaintenance {
		recordedStatus = status.Maintenance
	}

	prev, prevErr := r.hbStore.GetLatest(ctx, monitorID)

	prevStatus := status.Status(-1)
	if prevErr == nil {
		prevStatus = status.Status(prev.Status)
	}

	hb := &heartbeat.Heartbeat{
		MonitorID: monitorID,
		Status:    int(recordedStatus),
		Time:      heartbeat.RFC3339Time(time.Now()),
		Msg:       truncateMessage(result.Message, 255),
		Retries:   retries,
		Important: prevErr != nil || prevStatus != recordedStatus,
	}
	if result.Latency > 0 {
		l := result.Latency
		hb.Latency = &l
	}

	if err := r.hbStore.Insert(ctx, hb); err != nil {
		slog.ErrorContext(ctx, "failed to insert heartbeat", slog.String("monitor", monitorID), slog.Any("error", err))
		return
	}

	if r.metrics != nil {
		var latency int64
		if hb.Latency != nil {
			latency = *hb.Latency
		}
		r.metrics.Record(ctx, &monitorMetricInfo{
			id:   monitorID,
			name: r.monitorName,
			typ:  r.monitorType,
			tags: r.tags,
		}, result.Status == status.Up, latency)
	}

	if r.publisher != nil {
		r.publisher.Publish(broadcast.Event{
			Type: "heartbeat",
			Data: hb,
		})
	}

	if inMaintenance {
		return
	}

	if prevStatus != recordedStatus && prevErr == nil && r.notify != nil {
		slog.DebugContext(ctx, "status changed, dispatching notification", slog.String("monitor", monitorID), slog.String("from", prevStatus.String()), slog.String("to", recordedStatus.String()))
		r.notify.Dispatch(ctx, monitorID, recordedStatus, prevStatus, result.Message)
	}
}

func truncateMessage(s string, maxLen int) string {
	s = strings.TrimSpace(s)
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

type monitorMetricInfo struct {
	id   string
	name string
	typ  string
	tags []string
}

func (m *monitorMetricInfo) MonitorID() string   { return m.id }
func (m *monitorMetricInfo) MonitorName() string { return m.name }
func (m *monitorMetricInfo) MonitorType() string { return m.typ }
func (m *monitorMetricInfo) GroupName() string   { return "" }
func (m *monitorMetricInfo) Tags() []string      { return m.tags }
