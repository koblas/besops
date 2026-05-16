package stats

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/koblas/besops/models"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/sqlite"
	"github.com/stephenafamo/bob/dialect/sqlite/dm"
	"github.com/stephenafamo/bob/dialect/sqlite/im"
	"github.com/stephenafamo/bob/dialect/sqlite/sm"
)

type sqliteRepo struct {
	db  *sql.DB
	exc bob.DB
}

func NewRepository(db *sql.DB) Repository {
	return &sqliteRepo{db: db, exc: bob.NewDB(db)}
}

func (r *sqliteRepo) UpsertMinutely(ctx context.Context, stat *StatMinutely) error {
	if stat.ID == "" {
		stat.ID = uuid.New().String()
	}
	return r.upsertStat(ctx, "stat_minutely", stat.ID, stat.MonitorID, stat.Timestamp, stat.Ping, stat.PingMin, stat.PingMax, stat.Up, stat.Down)
}

func (r *sqliteRepo) UpsertHourly(ctx context.Context, stat *StatHourly) error {
	if stat.ID == "" {
		stat.ID = uuid.New().String()
	}
	return r.upsertStat(ctx, "stat_hourly", stat.ID, stat.MonitorID, stat.Timestamp, stat.Ping, stat.PingMin, stat.PingMax, stat.Up, stat.Down)
}

func (r *sqliteRepo) UpsertDaily(ctx context.Context, stat *StatDaily) error {
	if stat.ID == "" {
		stat.ID = uuid.New().String()
	}
	return r.upsertStat(ctx, "stat_daily", stat.ID, stat.MonitorID, stat.Timestamp, stat.Ping, stat.PingMin, stat.PingMax, stat.Up, stat.Down)
}

func (r *sqliteRepo) upsertStat(ctx context.Context, table, id, monitorID string, timestamp int64, ping float64, pingMin, pingMax int64, up, down int) error {
	q := sqlite.Insert(
		im.Into(table, "id", "monitor_id", "timestamp", "ping", "ping_min", "ping_max", "up", "down"),
		im.Values(sqlite.Arg(id, monitorID, timestamp, ping, pingMin, pingMax, up, down)),
		im.OnConflict("monitor_id", "timestamp").DoUpdate(
			im.SetExcluded("ping", "ping_min", "ping_max", "up", "down"),
		),
	)

	_, err := bob.Exec(ctx, r.exc, q)
	if err != nil {
		return fmt.Errorf("upserting %s stat: %w", table, err)
	}
	return nil
}

func (r *sqliteRepo) GetMinutely(ctx context.Context, monitorID string, since int64) ([]*StatMinutely, error) {
	ms, err := models.StatMinutelies.Query(
		sm.Where(sqlite.Raw("monitor_id = ? AND timestamp >= ?", monitorID, since)),
		sm.OrderBy(models.StatMinutelies.Columns.Timestamp).Asc(),
	).All(ctx, r.exc)
	if err != nil {
		return nil, fmt.Errorf("querying minutely stats: %w", err)
	}

	result := make([]*StatMinutely, len(ms))
	for i, m := range ms {
		result[i] = &StatMinutely{
			ID:        m.ID,
			MonitorID: m.MonitorID,
			Timestamp: m.Timestamp,
			Ping:      m.Ping.GetOrZero(),
			PingMin:   m.PingMin.GetOrZero(),
			PingMax:   m.PingMax.GetOrZero(),
			Up:        int(m.Up),
			Down:      int(m.Down),
		}
	}
	return result, nil
}

func (r *sqliteRepo) GetHourly(ctx context.Context, monitorID string, since int64) ([]*StatHourly, error) {
	hs, err := models.StatHourlies.Query(
		sm.Where(sqlite.Raw("monitor_id = ? AND timestamp >= ?", monitorID, since)),
		sm.OrderBy(models.StatHourlies.Columns.Timestamp).Asc(),
	).All(ctx, r.exc)
	if err != nil {
		return nil, fmt.Errorf("querying hourly stats: %w", err)
	}

	result := make([]*StatHourly, len(hs))
	for i, m := range hs {
		result[i] = &StatHourly{
			ID:        m.ID,
			MonitorID: m.MonitorID,
			Timestamp: m.Timestamp,
			Ping:      m.Ping.GetOrZero(),
			PingMin:   m.PingMin.GetOrZero(),
			PingMax:   m.PingMax.GetOrZero(),
			Up:        int(m.Up),
			Down:      int(m.Down),
		}
	}
	return result, nil
}

func (r *sqliteRepo) GetDaily(ctx context.Context, monitorID string, since int64) ([]*StatDaily, error) {
	ds, err := models.StatDailies.Query(
		sm.Where(sqlite.Raw("monitor_id = ? AND timestamp >= ?", monitorID, since)),
		sm.OrderBy(models.StatDailies.Columns.Timestamp).Asc(),
	).All(ctx, r.exc)
	if err != nil {
		return nil, fmt.Errorf("querying daily stats: %w", err)
	}

	result := make([]*StatDaily, len(ds))
	for i, m := range ds {
		result[i] = &StatDaily{
			ID:        m.ID,
			MonitorID: m.MonitorID,
			Timestamp: m.Timestamp,
			Ping:      m.Ping.GetOrZero(),
			PingMin:   m.PingMin.GetOrZero(),
			PingMax:   m.PingMax.GetOrZero(),
			Up:        int(m.Up),
			Down:      int(m.Down),
		}
	}
	return result, nil
}

func (r *sqliteRepo) DeleteOlderThan(ctx context.Context, table string, timestamp int64) error {
	allowedTables := map[string]bool{
		"stat_minutely": true,
		"stat_hourly":   true,
		"stat_daily":    true,
	}
	if !allowedTables[table] {
		return fmt.Errorf("invalid table name: %s", table)
	}

	q := sqlite.Delete(
		dm.From(sqlite.Quote(table)),
		dm.Where(sqlite.Raw("timestamp < ?", timestamp)),
	)

	_, err := bob.Exec(ctx, r.exc, q)
	if err != nil {
		return fmt.Errorf("deleting old %s stats: %w", table, err)
	}
	return nil
}
