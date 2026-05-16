package jobs

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// MonitorLister provides active monitor IDs.
type MonitorLister interface {
	FindAllActiveIDs(ctx context.Context) ([]string, error)
}

// HeartbeatCleaner deletes old heartbeats.
type HeartbeatCleaner interface {
	DeleteOlderThan(ctx context.Context, monitorID string, hours int) error
}

// ClearOldDataJob removes heartbeats older than the configured retention period.
type ClearOldDataJob struct {
	monitors   MonitorLister
	heartbeats HeartbeatCleaner
	retentionH int
}

func NewClearOldDataJob(monitors MonitorLister, heartbeats HeartbeatCleaner, retentionHours int) *ClearOldDataJob {
	return &ClearOldDataJob{
		monitors:   monitors,
		heartbeats: heartbeats,
		retentionH: retentionHours,
	}
}

func (j *ClearOldDataJob) Name() string     { return "clear_old_data" }
func (j *ClearOldDataJob) Schedule() string { return "0 3 * * * *" } // every hour at :03

func (j *ClearOldDataJob) Run(ctx context.Context) error {
	ids, err := j.monitors.FindAllActiveIDs(ctx)
	if err != nil {
		return fmt.Errorf("listing monitor IDs: %w", err)
	}

	start := time.Now()
	var deleted int
	for _, id := range ids {
		if err := j.heartbeats.DeleteOlderThan(ctx, id, j.retentionH); err != nil {
			slog.WarnContext(ctx, "failed to clear heartbeats", slog.String("monitor_id", id), slog.Any("error", err))
			continue
		}
		deleted++
	}

	slog.InfoContext(ctx, "clear old data complete", slog.Int("monitors", deleted), slog.Duration("duration", time.Since(start)))
	return nil
}
