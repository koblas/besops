package stats

import (
	"context"
	"fmt"
	"time"
)

type Handler struct {
	repo Repository
}

func NewHandler(repo Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) ClearAll(ctx context.Context) error {
	now := time.Now().Unix()
	if err := h.repo.DeleteOlderThan(ctx, "stat_minutely", now); err != nil {
		return fmt.Errorf("clearing minutely stats: %w", err)
	}
	if err := h.repo.DeleteOlderThan(ctx, "stat_hourly", now); err != nil {
		return fmt.Errorf("clearing hourly stats: %w", err)
	}
	if err := h.repo.DeleteOlderThan(ctx, "stat_daily", now); err != nil {
		return fmt.Errorf("clearing daily stats: %w", err)
	}
	return nil
}

func (h *Handler) GetMinutely(ctx context.Context, monitorID string, since int64) ([]*StatMinutely, error) {
	return h.repo.GetMinutely(ctx, monitorID, since)
}

func (h *Handler) GetHourly(ctx context.Context, monitorID string, since int64) ([]*StatHourly, error) {
	return h.repo.GetHourly(ctx, monitorID, since)
}

func (h *Handler) GetDaily(ctx context.Context, monitorID string, since int64) ([]*StatDaily, error) {
	return h.repo.GetDaily(ctx, monitorID, since)
}
