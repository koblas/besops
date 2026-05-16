package uptime

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/koblas/besops/internal/domain/heartbeat"
	"github.com/koblas/besops/internal/domain/stats"
)

// MonitorLister provides active monitor IDs for aggregation.
type MonitorLister interface {
	FindAllActiveIDs(ctx context.Context) ([]string, error)
}

// HeartbeatReader reads heartbeats for aggregation.
type HeartbeatReader interface {
	GetByMonitorInRange(ctx context.Context, monitorID string, from, to int64) ([]*heartbeat.Heartbeat, error)
}

// StatWriter persists aggregated stats.
type StatWriter interface {
	UpsertMinutely(ctx context.Context, stat *stats.StatMinutely) error
	UpsertHourly(ctx context.Context, stat *stats.StatHourly) error
	UpsertDaily(ctx context.Context, stat *stats.StatDaily) error
}

// Calculator aggregates heartbeats into minutely, hourly, and daily stat buckets.
type Calculator struct {
	monitors   MonitorLister
	heartbeats HeartbeatReader
	stats      StatWriter
}

func NewCalculator(
	monitors MonitorLister,
	heartbeats HeartbeatReader,
	stats StatWriter,
) *Calculator {
	return &Calculator{
		monitors:   monitors,
		heartbeats: heartbeats,
		stats:      stats,
	}
}

// AggregateMinutely computes stats for the previous minute for all active monitors.
func (c *Calculator) AggregateMinutely(ctx context.Context) error {
	now := time.Now().UTC()
	bucketEnd := now.Truncate(time.Minute)
	bucketStart := bucketEnd.Add(-time.Minute)

	return c.aggregate(ctx, bucketStart, bucketEnd, func(ctx context.Context, monitorID string, s *statBucket) error {
		return c.stats.UpsertMinutely(ctx, &stats.StatMinutely{
			MonitorID: monitorID,
			Timestamp: bucketStart.Unix(),
			Ping:      s.avgPing(),
			PingMin:   s.pingMin,
			PingMax:   s.pingMax,
			Up:        s.up,
			Down:      s.down,
		})
	})
}

// AggregateHourly computes stats for the previous hour for all active monitors.
func (c *Calculator) AggregateHourly(ctx context.Context) error {
	now := time.Now().UTC()
	bucketEnd := now.Truncate(time.Hour)
	bucketStart := bucketEnd.Add(-time.Hour)

	return c.aggregate(ctx, bucketStart, bucketEnd, func(ctx context.Context, monitorID string, s *statBucket) error {
		return c.stats.UpsertHourly(ctx, &stats.StatHourly{
			MonitorID: monitorID,
			Timestamp: bucketStart.Unix(),
			Ping:      s.avgPing(),
			PingMin:   s.pingMin,
			PingMax:   s.pingMax,
			Up:        s.up,
			Down:      s.down,
		})
	})
}

// AggregateDaily computes stats for the previous day for all active monitors.
func (c *Calculator) AggregateDaily(ctx context.Context) error {
	now := time.Now().UTC()
	bucketEnd := now.Truncate(24 * time.Hour)
	bucketStart := bucketEnd.Add(-24 * time.Hour)

	return c.aggregate(ctx, bucketStart, bucketEnd, func(ctx context.Context, monitorID string, s *statBucket) error {
		return c.stats.UpsertDaily(ctx, &stats.StatDaily{
			MonitorID: monitorID,
			Timestamp: bucketStart.Unix(),
			Ping:      s.avgPing(),
			PingMin:   s.pingMin,
			PingMax:   s.pingMax,
			Up:        s.up,
			Down:      s.down,
		})
	})
}

type upsertFn func(ctx context.Context, monitorID string, s *statBucket) error

func (c *Calculator) aggregate(ctx context.Context, from, to time.Time, upsert upsertFn) error {
	ids, err := c.monitors.FindAllActiveIDs(ctx)
	if err != nil {
		return fmt.Errorf("listing active monitors: %w", err)
	}

	for _, id := range ids {
		beats, beatErr := c.heartbeats.GetByMonitorInRange(ctx, id, from.Unix(), to.Unix())
		if beatErr != nil {
			slog.WarnContext(ctx, "failed to get heartbeats for aggregation", slog.String("monitor_id", id), slog.Any("error", beatErr))
			continue
		}
		if len(beats) == 0 {
			continue
		}

		bucket := computeBucket(beats)
		if upsertErr := upsert(ctx, id, bucket); upsertErr != nil {
			slog.WarnContext(ctx, "failed to upsert stat", slog.String("monitor_id", id), slog.Any("error", upsertErr))
		}
	}
	return nil
}

type statBucket struct {
	up      int
	down    int
	pingSum int64
	pingMin int64
	pingMax int64
	pingN   int
}

func (s *statBucket) avgPing() float64 {
	if s.pingN == 0 {
		return 0
	}
	return float64(s.pingSum) / float64(s.pingN)
}

func computeBucket(beats []*heartbeat.Heartbeat) *statBucket {
	b := &statBucket{
		pingMin: int64(^uint64(0) >> 1), // max int64
	}

	for _, hb := range beats {
		switch hb.Status {
		case 1: // up
			b.up++
		case 0: // down
			b.down++
		}

		if hb.Ping != nil {
			p := *hb.Ping
			b.pingSum += p
			b.pingN++
			if p < b.pingMin {
				b.pingMin = p
			}
			if p > b.pingMax {
				b.pingMax = p
			}
		}
	}

	if b.pingN == 0 {
		b.pingMin = 0
	}
	return b
}
