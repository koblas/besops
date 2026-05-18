package heartbeat

import "context"

type Repository interface {
	Insert(ctx context.Context, hb *Heartbeat) error
	GetLatest(ctx context.Context, monitorID string) (*Heartbeat, error)
	GetPrevious(ctx context.Context, monitorID string) (*Heartbeat, error)
	GetByMonitorPaged(ctx context.Context, monitorID string, offset, limit int) ([]*Heartbeat, error)
	GetImportantByMonitor(ctx context.Context, monitorID string, offset, limit int) ([]*Heartbeat, int64, error)
	GetByMonitorInRange(ctx context.Context, monitorID string, from, to int64) ([]*Heartbeat, error)
	DeleteOlderThan(ctx context.Context, monitorID string, hours int) error
	ClearByMonitor(ctx context.Context, monitorID string) error
	GetAverageLatency(ctx context.Context, monitorID string, hours int) (float64, error)
	GetAverageResponse(ctx context.Context, monitorID string, hours int) (float64, error)
	GetUptime(ctx context.Context, monitorID string, hours int) (float64, error)
}
