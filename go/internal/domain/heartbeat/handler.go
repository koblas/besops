package heartbeat

import (
	"context"
	"fmt"
	"time"

	oas "github.com/koblas/besops/internal/api/oas_generated"
	"github.com/koblas/besops/internal/api/oasutil"
)

var _ oas.HeartbeatHandler = (*Handler)(nil)

type ChartPoint struct {
	Timestamp  int64
	Up         int
	Down       int
	Latency    float64
	LatencyMin int64
	LatencyMax int64
}

type ChartRepository interface {
	GetMinutely(ctx context.Context, monitorID string, since int64) ([]ChartPoint, error)
	GetHourly(ctx context.Context, monitorID string, since int64) ([]ChartPoint, error)
	GetDaily(ctx context.Context, monitorID string, since int64) ([]ChartPoint, error)
}

type Handler struct {
	repo      Repository
	chartRepo ChartRepository
}

func NewHandler(repo Repository, chartRepo ChartRepository) *Handler {
	return &Handler{repo: repo, chartRepo: chartRepo}
}

func (h *Handler) GetHeartbeats(ctx context.Context, params oas.GetHeartbeatsParams) ([]oas.Heartbeat, error) {
	var limit int
	if params.Count.IsSet() {
		limit = params.Count.Value
		if limit > 500 {
			limit = 500
		}
	} else {
		hours := 24
		if params.Hours.IsSet() {
			hours = params.Hours.Value
		}
		limit = hours * 60
	}

	hbs, err := h.repo.GetByMonitorPaged(ctx, params.MonitorId.String(), 0, limit)
	if err != nil {
		return nil, fmt.Errorf("getting heartbeats: %w", err)
	}

	result := make([]oas.Heartbeat, 0, len(hbs))
	for _, hb := range hbs {
		result = append(result, heartbeatToOAS(hb))
	}
	return result, nil
}

func (h *Handler) GetImportantHeartbeats(ctx context.Context, params oas.GetImportantHeartbeatsParams) (*oas.GetImportantHeartbeatsOK, error) {
	limit := 100
	if params.Limit.IsSet() {
		limit = params.Limit.Value
	}
	offset := 0
	if params.Offset.IsSet() {
		offset = params.Offset.Value
	}

	hbs, total, err := h.repo.GetImportantByMonitor(ctx, params.MonitorId.String(), offset, limit)
	if err != nil {
		return nil, fmt.Errorf("getting important heartbeats: %w", err)
	}

	events := make([]oas.Heartbeat, 0, len(hbs))
	for _, hb := range hbs {
		events = append(events, heartbeatToOAS(hb))
	}
	return &oas.GetImportantHeartbeatsOK{
		Data:  events,
		Total: total,
	}, nil
}

func (h *Handler) ClearHeartbeats(ctx context.Context, params oas.ClearHeartbeatsParams) error {
	if err := h.repo.ClearByMonitor(ctx, params.MonitorId.String()); err != nil {
		return fmt.Errorf("clearing heartbeats: %w", err)
	}
	return nil
}

func (h *Handler) ClearEvents(ctx context.Context, params oas.ClearEventsParams) error {
	if err := h.repo.ClearByMonitor(ctx, params.MonitorId.String()); err != nil {
		return fmt.Errorf("clearing events: %w", err)
	}
	return nil
}

func (h *Handler) GetChartData(ctx context.Context, params oas.GetChartDataParams) ([]oas.ChartPoint, error) {
	hours := 24
	if params.Hours.IsSet() {
		hours = params.Hours.Value
	}

	since := time.Now().Add(-time.Duration(hours) * time.Hour).Unix()

	var points []ChartPoint
	var err error
	switch {
	case hours <= 24:
		points, err = h.chartRepo.GetMinutely(ctx, params.MonitorId.String(), since)
	case hours <= 720:
		points, err = h.chartRepo.GetHourly(ctx, params.MonitorId.String(), since)
	default:
		points, err = h.chartRepo.GetDaily(ctx, params.MonitorId.String(), since)
	}
	if err != nil {
		return nil, fmt.Errorf("getting chart data: %w", err)
	}

	result := make([]oas.ChartPoint, 0, len(points))
	for _, p := range points {
		result = append(result, oas.ChartPoint{
			Timestamp: oas.NewOptInt64(p.Timestamp),
			Up:        oas.NewOptInt(p.Up),
			Down:      oas.NewOptInt(p.Down),
			Latency:    oas.NewOptFloat64(p.Latency),
			LatencyMin: oas.NewOptInt(int(p.LatencyMin)),
			LatencyMax: oas.NewOptInt(int(p.LatencyMax)),
		})
	}
	return result, nil
}


func (h *Handler) ListRecentEvents(ctx context.Context, params oas.ListRecentEventsParams) (*oas.ListRecentEventsOK, error) {
	limit := 25
	if params.Limit.IsSet() {
		limit = params.Limit.Value
	}

	hbs, total, err := h.repo.GetAllImportant(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("getting recent events: %w", err)
	}

	events := make([]oas.Heartbeat, 0, len(hbs))
	for _, hb := range hbs {
		events = append(events, heartbeatToOAS(hb))
	}
	return &oas.ListRecentEventsOK{
		Data:  events,
		Total: total,
	}, nil
}

func heartbeatToOAS(hb *Heartbeat) oas.Heartbeat {
	result := oas.Heartbeat{
		ID:        oasutil.MustParseUUID(hb.ID),
		MonitorId: oasutil.MustParseUUID(hb.MonitorID),
		Status:    oas.HeartbeatStatus(hb.Status),
		Time:      time.Time(hb.Time),
	}
	if hb.Msg != "" {
		result.Msg = oas.NewOptString(hb.Msg)
	}
	if hb.Latency != nil {
		result.Latency = oas.NewOptInt64(*hb.Latency)
	}
	if hb.Important {
		result.Important = oas.NewOptBool(true)
	}
	if hb.Duration > 0 {
		result.Duration = oas.NewOptInt64(hb.Duration)
	}
	return result
}
