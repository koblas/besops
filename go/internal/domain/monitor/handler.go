package monitor

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	oas "github.com/koblas/besops/internal/api/oas_generated"
	"github.com/koblas/besops/internal/api/oasutil"
)

var _ oas.MonitorHandler = (*Handler)(nil)

// MonitorScheduler is the interface the handler uses to notify the scheduler of changes.
type MonitorScheduler interface {
	StartMonitor(ctx context.Context, id string) error
	StopMonitor(ctx context.Context, id string)
	RestartMonitor(ctx context.Context, id string) error
}

// UptimeProvider computes uptime ratios for monitors.
type UptimeProvider interface {
	GetUptime(ctx context.Context, monitorID string, hours int) (float64, error)
}

// MonitorTagInfo holds resolved tag data for API responses.
type MonitorTagInfo struct {
	TagID string
	Name  string
	Color string
	Value string
}

// TagReader loads tags for monitors in API responses.
type TagReader interface {
	GetMonitorTags(ctx context.Context, monitorID string) ([]MonitorTagInfo, error)
}

type Handler struct {
	repo      Repository
	scheduler MonitorScheduler
	uptimes   UptimeProvider
	tags      TagReader
}

func NewHandler(repo Repository, scheduler MonitorScheduler, uptimes UptimeProvider, tags TagReader) *Handler {
	return &Handler{repo: repo, scheduler: scheduler, uptimes: uptimes, tags: tags}
}

func (h *Handler) ListMonitors(ctx context.Context) ([]oas.Monitor, error) {
	userID, err := oasutil.UserIDFromCtx(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting user from context: %w", err)
	}

	monitors, err := h.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("listing monitors: %w", err)
	}

	result := make([]oas.Monitor, 0, len(monitors))
	for _, m := range monitors {
		om := monitorToOAS(m)
		if h.tags != nil {
			om.Tags = h.loadOASTags(ctx, m.ID)
		}
		result = append(result, om)
	}
	return result, nil
}

func (h *Handler) GetMonitor(ctx context.Context, params oas.GetMonitorParams) (oas.GetMonitorRes, error) {
	m, err := h.repo.FindByID(ctx, params.MonitorId.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return &oas.ErrorResponse{Error: "monitor not found"}, nil
		}
		return nil, fmt.Errorf("finding monitor: %w", err)
	}
	result := monitorToOAS(m)
	if h.tags != nil {
		result.Tags = h.loadOASTags(ctx, m.ID)
	}
	return &result, nil
}

func (h *Handler) CreateMonitor(ctx context.Context, req *oas.MonitorInput) (*oas.CreateMonitorCreated, error) {
	userID, err := oasutil.UserIDFromCtx(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting user from context: %w", err)
	}

	m := monitorFromInput(req, userID)
	id, err := h.repo.Create(ctx, m)
	if err != nil {
		return nil, fmt.Errorf("creating monitor: %w", err)
	}

	if m.Active && h.scheduler != nil {
		_ = h.scheduler.StartMonitor(ctx, id)
	}

	return &oas.CreateMonitorCreated{ID: oasutil.MustParseUUID(id)}, nil
}

func (h *Handler) UpdateMonitor(ctx context.Context, req *oas.MonitorInput, params oas.UpdateMonitorParams) (oas.UpdateMonitorRes, error) {
	existing, err := h.repo.FindByID(ctx, params.MonitorId.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return &oas.ErrorResponse{Error: "monitor not found"}, nil
		}
		return nil, fmt.Errorf("finding monitor: %w", err)
	}

	applyMonitorInput(existing, req)
	if err := h.repo.Update(ctx, existing); err != nil {
		return nil, fmt.Errorf("updating monitor: %w", err)
	}

	if h.scheduler != nil {
		if existing.Active {
			_ = h.scheduler.RestartMonitor(ctx, existing.ID)
		} else {
			h.scheduler.StopMonitor(ctx, existing.ID)
		}
	}

	result := monitorToOAS(existing)
	return &result, nil
}

func (h *Handler) DeleteMonitor(ctx context.Context, params oas.DeleteMonitorParams) (oas.DeleteMonitorRes, error) {
	id := params.MonitorId.String()
	if h.scheduler != nil {
		h.scheduler.StopMonitor(ctx, id)
	}
	if err := h.repo.Delete(ctx, id); err != nil {
		return nil, fmt.Errorf("deleting monitor: %w", err)
	}
	return &oas.DeleteMonitorNoContent{}, nil
}

func (h *Handler) PauseMonitor(ctx context.Context, params oas.PauseMonitorParams) (*oas.MessageResponse, error) {
	m, err := h.repo.FindByID(ctx, params.MonitorId.String())
	if err != nil {
		return nil, fmt.Errorf("finding monitor: %w", err)
	}
	m.Active = false
	if err := h.repo.Update(ctx, m); err != nil {
		return nil, fmt.Errorf("pausing monitor: %w", err)
	}
	if h.scheduler != nil {
		h.scheduler.StopMonitor(ctx, m.ID)
	}
	return &oas.MessageResponse{Message: "paused"}, nil
}

func (h *Handler) ResumeMonitor(ctx context.Context, params oas.ResumeMonitorParams) (*oas.MessageResponse, error) {
	m, err := h.repo.FindByID(ctx, params.MonitorId.String())
	if err != nil {
		return nil, fmt.Errorf("finding monitor: %w", err)
	}
	m.Active = true
	if err := h.repo.Update(ctx, m); err != nil {
		return nil, fmt.Errorf("resuming monitor: %w", err)
	}
	if h.scheduler != nil {
		_ = h.scheduler.StartMonitor(ctx, m.ID)
	}
	return &oas.MessageResponse{Message: "resumed"}, nil
}

func (h *Handler) CheckDomain(ctx context.Context, params oas.CheckDomainParams) (*oas.CheckDomainOK, error) {
	// TODO: implement domain/TLS check
	_ = params.MonitorId
	return &oas.CheckDomainOK{}, nil
}

func (h *Handler) loadOASTags(ctx context.Context, monitorID string) []oas.MonitorTag {
	tags, err := h.tags.GetMonitorTags(ctx, monitorID)
	if err != nil {
		return nil
	}
	result := make([]oas.MonitorTag, 0, len(tags))
	for _, t := range tags {
		result = append(result, oas.MonitorTag{
			TagId: oas.NewOptUUID(oasutil.MustParseUUID(t.TagID)),
			Name:  oas.NewOptString(t.Name),
			Color: oas.NewOptString(t.Color),
			Value: oas.NewOptString(t.Value),
		})
	}
	return result
}

func monitorToOAS(m *Monitor) oas.Monitor {
	result := oas.Monitor{
		ID:                 oasutil.MustParseUUID(m.ID),
		Name:               m.Name,
		Type:               oas.MonitorType(m.Type),
		Active:             m.Active,
		Interval:           oas.NewOptInt(m.Interval),
		Timeout:            oas.NewOptInt(m.Timeout),
		MaxRetries:         oas.NewOptInt(m.MaxRetries),
		RetryInterval:      oas.NewOptInt(m.RetryInterval),
		Description:        oasutil.PtrToOptString(m.Description),
		UpsideDown:         oas.NewOptBool(m.UpsideDown),
		ResendInterval:     oas.NewOptInt(m.ResendInterval),
		ExpiryNotification: oas.NewOptBool(m.ExpiryNotification),
	}

	var cfg oas.MonitorConfig
	if m.ConfigJSON != "" && m.ConfigJSON != "{}" {
		if err := json.Unmarshal([]byte(m.ConfigJSON), &cfg); err == nil {
			result.Config = oas.OptMonitorConfig{Value: cfg, Set: true}
		}
	}

	return result
}

func monitorFromInput(req *oas.MonitorInput, userID string) *Monitor {
	m := &Monitor{
		Name:               req.Name,
		Type:               req.Type,
		Active:             oasutil.OptBoolValue(req.Active, true),
		UserID:             userID,
		Interval:           oasutil.OptIntValue(req.Interval, 60),
		MaxRetries:         oasutil.OptIntValue(req.MaxRetries, 1),
		Timeout:            oasutil.OptIntValue(req.Timeout, 48),
		RetryInterval:      oasutil.OptIntValue(req.RetryInterval, 60),
		Description:        oasutil.OptStringValue(req.Description),
		UpsideDown:         oasutil.OptBoolValue(req.UpsideDown, false),
		ResendInterval:     oasutil.OptIntValue(req.ResendInterval, 0),
		ExpiryNotification: oasutil.OptBoolValue(req.ExpiryNotification, false),
	}

	m.ConfigJSON = marshalConfig(&req.Config)
	enforceGroupDefaults(m)

	return m
}

func applyMonitorInput(m *Monitor, req *oas.MonitorInput) {
	m.Name = req.Name
	m.Type = req.Type
	if req.Active.IsSet() {
		m.Active = req.Active.Value
	}
	if req.Interval.IsSet() {
		m.Interval = req.Interval.Value
	}
	if req.MaxRetries.IsSet() {
		m.MaxRetries = req.MaxRetries.Value
	}
	if req.Timeout.IsSet() {
		m.Timeout = req.Timeout.Value
	}
	if req.RetryInterval.IsSet() {
		m.RetryInterval = req.RetryInterval.Value
	}
	if req.Description.IsSet() {
		m.Description = req.Description.Value
	}
	if req.UpsideDown.IsSet() {
		m.UpsideDown = req.UpsideDown.Value
	}
	if req.ResendInterval.IsSet() {
		m.ResendInterval = req.ResendInterval.Value
	}
	if req.ExpiryNotification.IsSet() {
		m.ExpiryNotification = req.ExpiryNotification.Value
	}

	m.ConfigJSON = marshalConfig(&req.Config)
	enforceGroupDefaults(m)
}

func enforceGroupDefaults(m *Monitor) {
	if m.Type == "group" {
		m.Interval = 60
		m.MaxRetries = 0
		m.RetryInterval = 60
	}
}

func marshalConfig(cfg *oas.MonitorConfig) string {
	b, err := json.Marshal(cfg)
	if err != nil {
		return "{}"
	}
	return string(b)
}

func (h *Handler) GetMonitorUptimes(ctx context.Context) (oas.GetMonitorUptimesOK, error) {
	ids, err := h.repo.FindAllActiveIDs(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading active monitor IDs: %w", err)
	}

	result := make(oas.GetMonitorUptimesOK, len(ids))
	for _, id := range ids {
		up, upErr := h.uptimes.GetUptime(ctx, id, 24)
		if upErr == nil {
			result[id] = up
		}
	}
	return result, nil
}
