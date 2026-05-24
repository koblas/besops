package maintenance

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
	oas "github.com/koblas/besops/internal/api/oas_generated"
	"github.com/koblas/besops/internal/api/oasutil"
)

var _ oas.MaintenanceHandler = (*Handler)(nil)

type Handler struct {
	repo Repository
}

func NewHandler(repo Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) ListMaintenance(ctx context.Context) ([]oas.Maintenance, error) {
	list, err := h.repo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing maintenance windows: %w", err)
	}

	result := make([]oas.Maintenance, 0, len(list))
	for _, m := range list {
		result = append(result, maintenanceToOAS(m))
	}
	return result, nil
}

func (h *Handler) GetMaintenance(ctx context.Context, params oas.GetMaintenanceParams) (*oas.Maintenance, error) {
	m, err := h.repo.FindByID(ctx, params.MaintenanceId.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("maintenance not found: %w", err)
		}
		return nil, fmt.Errorf("finding maintenance: %w", err)
	}
	result := maintenanceToOAS(m)
	return &result, nil
}

func (h *Handler) CreateMaintenance(ctx context.Context, req *oas.MaintenanceInput) (*oas.CreateMaintenanceCreated, error) {
	userID, err := oasutil.UserIDFromCtx(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting user from context: %w", err)
	}

	m := maintenanceFromInput(req, userID)
	id, err := h.repo.Create(ctx, m)
	if err != nil {
		return nil, fmt.Errorf("creating maintenance: %w", err)
	}

	return &oas.CreateMaintenanceCreated{ID: oasutil.MustParseUUID(id)}, nil
}

func (h *Handler) UpdateMaintenance(ctx context.Context, req *oas.MaintenanceInput, params oas.UpdateMaintenanceParams) (*oas.Maintenance, error) {
	existing, err := h.repo.FindByID(ctx, params.MaintenanceId.String())
	if err != nil {
		return nil, fmt.Errorf("finding maintenance: %w", err)
	}

	applyMaintenanceInput(existing, req)
	if err := h.repo.Update(ctx, existing); err != nil {
		return nil, fmt.Errorf("updating maintenance: %w", err)
	}

	result := maintenanceToOAS(existing)
	return &result, nil
}

func (h *Handler) DeleteMaintenance(ctx context.Context, params oas.DeleteMaintenanceParams) error {
	if err := h.repo.Delete(ctx, params.MaintenanceId.String()); err != nil {
		return fmt.Errorf("deleting maintenance: %w", err)
	}
	return nil
}

func (h *Handler) PauseMaintenance(ctx context.Context, params oas.PauseMaintenanceParams) (*oas.MessageResponse, error) {
	m, err := h.repo.FindByID(ctx, params.MaintenanceId.String())
	if err != nil {
		return nil, fmt.Errorf("finding maintenance: %w", err)
	}
	m.Active = false
	if err := h.repo.Update(ctx, m); err != nil {
		return nil, fmt.Errorf("pausing maintenance: %w", err)
	}
	return &oas.MessageResponse{Message: "paused"}, nil
}

func (h *Handler) ResumeMaintenance(ctx context.Context, params oas.ResumeMaintenanceParams) (*oas.MessageResponse, error) {
	m, err := h.repo.FindByID(ctx, params.MaintenanceId.String())
	if err != nil {
		return nil, fmt.Errorf("finding maintenance: %w", err)
	}
	m.Active = true
	if err := h.repo.Update(ctx, m); err != nil {
		return nil, fmt.Errorf("resuming maintenance: %w", err)
	}
	return &oas.MessageResponse{Message: "resumed"}, nil
}

func (h *Handler) GetMaintenanceMonitors(ctx context.Context, params oas.GetMaintenanceMonitorsParams) ([]uuid.UUID, error) {
	ids, err := h.repo.GetMonitorMaintenanceIDs(ctx, params.MaintenanceId.String())
	if err != nil {
		return nil, fmt.Errorf("getting maintenance monitors: %w", err)
	}
	result := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		result = append(result, oasutil.MustParseUUID(id))
	}
	return result, nil
}

func (h *Handler) SetMaintenanceMonitors(ctx context.Context, req *oas.SetMaintenanceMonitorsReq, params oas.SetMaintenanceMonitorsParams) error {
	ids := make([]string, len(req.MonitorIds))
	for i, id := range req.MonitorIds {
		ids[i] = id.String()
	}

	if err := h.repo.SetMonitorIDs(ctx, params.MaintenanceId.String(), ids); err != nil {
		return fmt.Errorf("setting maintenance monitors: %w", err)
	}
	return nil
}

func (h *Handler) GetMaintenanceStatusPages(ctx context.Context, params oas.GetMaintenanceStatusPagesParams) ([]uuid.UUID, error) {
	ids, err := h.repo.GetStatusPageIDs(ctx, params.MaintenanceId.String())
	if err != nil {
		return nil, fmt.Errorf("getting maintenance status pages: %w", err)
	}

	result := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		result = append(result, uuid.MustParse(id))
	}
	return result, nil
}

func (h *Handler) SetMaintenanceStatusPages(ctx context.Context, req *oas.SetMaintenanceStatusPagesReq, params oas.SetMaintenanceStatusPagesParams) error {
	ids := make([]string, len(req.StatusPageIds))
	for i, id := range req.StatusPageIds {
		ids[i] = id.String()
	}

	if err := h.repo.SetStatusPageIDs(ctx, params.MaintenanceId.String(), ids); err != nil {
		return fmt.Errorf("setting maintenance status pages: %w", err)
	}
	return nil
}

func maintenanceToOAS(m *Maintenance) oas.Maintenance {
	result := oas.Maintenance{
		ID:       oasutil.MustParseUUID(m.ID),
		Title:    m.Title,
		Active:   m.Active,
		Strategy: oas.MaintenanceStrategy(m.Strategy),
	}
	if m.Description != "" {
		result.Description = oas.NewOptString(m.Description)
	}
	if m.StartDate != nil {
		result.StartDate = oas.NewOptDateTime(*m.StartDate)
	}
	if m.EndDate != nil {
		result.EndDate = oas.NewOptDateTime(*m.EndDate)
	}
	if m.StartTime != "" {
		result.StartTime = oas.NewOptString(m.StartTime)
	}
	if m.EndTime != "" {
		result.EndTime = oas.NewOptString(m.EndTime)
	}
	if m.Weekdays != "" {
		result.Weekdays = parseInt32List(m.Weekdays)
	}
	if m.DaysOfMonth != "" {
		result.DaysOfMonth = parseInt32List(m.DaysOfMonth)
	}
	if m.IntervalDay > 0 {
		result.IntervalDay = oas.NewOptInt32(int32(m.IntervalDay)) //nolint:gosec // small config value
	}
	if m.CronExpression != "" {
		result.Cron = oas.NewOptString(m.CronExpression)
	}
	if m.DurationMinutes > 0 {
		result.DurationMinutes = oas.NewOptInt32(int32(m.DurationMinutes)) //nolint:gosec // small config value
	}
	if m.TimezoneOption != "" {
		result.TimezoneOption = oas.NewOptString(m.TimezoneOption)
	}
	return result
}

func maintenanceFromInput(req *oas.MaintenanceInput, userID string) *Maintenance {
	m := &Maintenance{
		Title:           req.Title,
		Description:     oasutil.OptStringValue(req.Description),
		UserID:          userID,
		Active:          oasutil.OptBoolValue(req.Active, true),
		Strategy:        string(req.Strategy),
		StartTime:       oasutil.OptStringValue(req.StartTime),
		EndTime:         oasutil.OptStringValue(req.EndTime),
		IntervalDay:     oasutil.OptIntValue(req.IntervalDay, 0),
		CronExpression:  oasutil.OptStringValue(req.Cron),
		DurationMinutes: oasutil.OptIntValue(req.DurationMinutes, 0),
		TimezoneOption:  oasutil.OptStringValue(req.TimezoneOption),
	}
	if req.StartDate.IsSet() {
		t := req.StartDate.Value
		m.StartDate = &t
	}
	if req.EndDate.IsSet() {
		t := req.EndDate.Value
		m.EndDate = &t
	}
	if len(req.Weekdays) > 0 {
		m.Weekdays = int32ListToJSON(req.Weekdays)
	}
	if len(req.DaysOfMonth) > 0 {
		m.DaysOfMonth = int32ListToJSON(req.DaysOfMonth)
	}
	return m
}

func applyMaintenanceInput(m *Maintenance, req *oas.MaintenanceInput) {
	m.Title = req.Title
	m.Strategy = string(req.Strategy)
	if req.Description.IsSet() {
		m.Description = req.Description.Value
	}
	if req.Active.IsSet() {
		m.Active = req.Active.Value
	}
	if req.StartDate.IsSet() {
		t := req.StartDate.Value
		m.StartDate = &t
	}
	if req.EndDate.IsSet() {
		t := req.EndDate.Value
		m.EndDate = &t
	}
	if req.StartTime.IsSet() {
		m.StartTime = req.StartTime.Value
	}
	if req.EndTime.IsSet() {
		m.EndTime = req.EndTime.Value
	}
	if len(req.Weekdays) > 0 {
		m.Weekdays = int32ListToJSON(req.Weekdays)
	}
	if len(req.DaysOfMonth) > 0 {
		m.DaysOfMonth = int32ListToJSON(req.DaysOfMonth)
	}
	if req.IntervalDay.IsSet() {
		m.IntervalDay = int(req.IntervalDay.Value)
	}
	if req.Cron.IsSet() {
		m.CronExpression = req.Cron.Value
	}
	if req.DurationMinutes.IsSet() {
		m.DurationMinutes = int(req.DurationMinutes.Value)
	}
	if req.TimezoneOption.IsSet() {
		m.TimezoneOption = req.TimezoneOption.Value
	}
}

func parseInt32List(s string) []int32 {
	var raw []int
	if err := json.Unmarshal([]byte(s), &raw); err == nil {
		result := make([]int32, len(raw))
		for i, v := range raw {
			result[i] = int32(v) //nolint:gosec // weekday/day-of-month values, no overflow
		}
		return result
	}
	var result []int32
	parts := strings.Split(s, ",")
	for _, p := range parts {
		if v, err := strconv.Atoi(strings.TrimSpace(p)); err == nil {
			result = append(result, int32(v)) //nolint:gosec // weekday/day-of-month values, no overflow
		}
	}
	return result
}

func int32ListToJSON(vals []int32) string {
	b, _ := json.Marshal(vals)
	return string(b)
}
