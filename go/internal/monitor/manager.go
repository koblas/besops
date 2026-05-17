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
}

func WithManagerMaxWorkers(n int) ManagerOption {
	return func(o *managerOpts) { o.maxWorkers = n }
}

func NewManager(store Store, hbStore HeartbeatStore, registry *Registry, notify NotificationDispatcher, publisher broadcast.Publisher, maint MaintenanceChecker, opts ...ManagerOption) *Manager {
	o := &managerOpts{maxWorkers: 32}
	for _, opt := range opts {
		opt(o)
	}

	sched := NewScheduler(store, hbStore, registry, notify, publisher, maint, WithMaxWorkers(o.maxWorkers))
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
		URL:           mon.URL,
		Hostname:      mon.Hostname,
		Interval:      time.Duration(mon.Interval) * time.Second,
		Timeout:       time.Duration(mon.Timeout) * time.Second,
		MaxRetries:    mon.MaxRetries,
		RetryInterval: time.Duration(mon.RetryInterval) * time.Second,
		IgnoreTLS:     mon.IgnoreTLS,
		Keyword:       mon.Keyword,
		JsonPath:      mon.JsonPath,
		ExpectedValue: mon.ExpectedValue,
		ProxyID:       mon.ProxyID,
		PushToken:     mon.PushToken,
		HTTP: HTTPConfig{
			Method:        mon.Method,
			Body:          mon.Body,
			BasicAuthUser: mon.BasicAuthUser,
			BasicAuthPass: mon.BasicAuthPass,
		},
		DNS: DNSConfig{
			ResolveType: mon.DNSResolveType,
		},
		MQTT: MQTTConfig{
			Topic:          mon.MQTTTopic,
			Username:       mon.MQTTUsername,
			Password:       mon.MQTTPassword,
			SuccessMessage: mon.MQTTSuccessMessage,
		},
		GRPC: GRPCConfig{
			URL:         mon.GRPCUrl,
			ServiceName: mon.GRPCServiceName,
			Method:      mon.GRPCMethod,
			Body:        mon.GRPCBody,
			Protobuf:    mon.GRPCProtobuf,
			EnableTLS:   mon.GRPCEnableTLS,
		},
		SMTP: SMTPConfig{},
		Redis: RedisConfig{
			ConnectionString: mon.DatabaseQuery,
		},
	}

	if mon.Port != nil {
		cfg.Port = *mon.Port
	}

	if mon.Headers != "" {
		_ = json.Unmarshal([]byte(mon.Headers), &cfg.HTTP.Headers)
	}

	if mon.AcceptedStatusCodes != "" {
		_ = json.Unmarshal([]byte(mon.AcceptedStatusCodes), &cfg.HTTP.AcceptedStatusCodes)
	}

	if mon.InvertKeyword {
		cfg.KeywordType = "not contain"
	} else if mon.Keyword != "" {
		cfg.KeywordType = "contain"
	}

	return cfg
}

// resultRecorder implements ResultHandler by persisting heartbeats and dispatching notifications.
type resultRecorder struct {
	hbStore   HeartbeatStore
	notify    NotificationDispatcher
	publisher broadcast.Publisher
	maint     MaintenanceChecker
}

func (r *resultRecorder) HandleResult(ctx context.Context, monitorID string, result CheckResult, retries int) {
	slog.DebugContext(ctx, "recording result", slog.String("monitor", monitorID), slog.String("status", result.Status.String()), slog.Int64("ping", result.Ping), slog.Int("retries", retries))

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
	if result.Ping > 0 {
		ping := result.Ping
		hb.Ping = &ping
	}

	if err := r.hbStore.Insert(ctx, hb); err != nil {
		slog.ErrorContext(ctx, "failed to insert heartbeat", slog.String("monitor", monitorID), slog.Any("error", err))
		return
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
