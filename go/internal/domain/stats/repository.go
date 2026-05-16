package stats

import "context"

type Repository interface {
	UpsertMinutely(ctx context.Context, stat *StatMinutely) error
	UpsertHourly(ctx context.Context, stat *StatHourly) error
	UpsertDaily(ctx context.Context, stat *StatDaily) error
	GetMinutely(ctx context.Context, monitorID string, since int64) ([]*StatMinutely, error)
	GetHourly(ctx context.Context, monitorID string, since int64) ([]*StatHourly, error)
	GetDaily(ctx context.Context, monitorID string, since int64) ([]*StatDaily, error)
	DeleteOlderThan(ctx context.Context, table string, timestamp int64) error
}
