package jobs

import (
	"context"
	"fmt"
	"log/slog"
)

// ExpiredSessionCleaner is implemented by session stores that support pruning.
type ExpiredSessionCleaner interface {
	DeleteExpired(ctx context.Context) (int64, error)
}

// ClearExpiredSessionsJob removes expired refresh token sessions.
type ClearExpiredSessionsJob struct {
	store ExpiredSessionCleaner
}

func NewClearExpiredSessionsJob(store ExpiredSessionCleaner) *ClearExpiredSessionsJob {
	return &ClearExpiredSessionsJob{store: store}
}

func (j *ClearExpiredSessionsJob) Name() string     { return "clear-expired-sessions" }
func (j *ClearExpiredSessionsJob) Schedule() string { return "0 0 * * * *" } // every hour

func (j *ClearExpiredSessionsJob) Run(ctx context.Context) error {
	deleted, err := j.store.DeleteExpired(ctx)
	if err != nil {
		return fmt.Errorf("clearing expired sessions: %w", err)
	}
	if deleted > 0 {
		slog.InfoContext(ctx, "cleared expired sessions", slog.Int64("count", deleted))
	}
	return nil
}
