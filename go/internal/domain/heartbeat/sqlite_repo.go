package heartbeat

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/aarondl/opt/omit"
	"github.com/aarondl/opt/omitnull"
	"github.com/google/uuid"
	"github.com/koblas/besops/lib/errs"
	"github.com/koblas/besops/models"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/sqlite"
	"github.com/stephenafamo/bob/dialect/sqlite/dm"
	"github.com/stephenafamo/bob/dialect/sqlite/sm"
)

type sqliteRepo struct {
	db  *sql.DB
	exc bob.DB
}

func NewRepository(db *sql.DB) Repository {
	return &sqliteRepo{db: db, exc: bob.NewDB(db)}
}

func (r *sqliteRepo) Insert(ctx context.Context, hb *Heartbeat) error {
	if hb.ID == "" {
		hb.ID = uuid.New().String()
	}

	setter := &models.HeartbeatSetter{
		ID:        omit.From(hb.ID),
		MonitorID: omit.From(hb.MonitorID),
		Status:    omit.From(int64(hb.Status)),
		MSG:       omitnull.From(hb.Msg),
		Time:      omit.From(time.Time(hb.Time)),
		Important: omit.From(hb.Important),
		Duration:  omit.From(hb.Duration),
		Retries:   omit.From(int64(hb.Retries)),
	}
	if hb.Ping != nil {
		setter.Ping = omitnull.From(*hb.Ping)
	}
	if hb.Response != nil {
		setter.Response = omitnull.From(hb.Response)
	}

	_, err := models.Heartbeats.Insert(setter).One(ctx, r.exc)
	if err != nil {
		return fmt.Errorf("inserting heartbeat: %w", err)
	}
	return nil
}

func (r *sqliteRepo) GetLatest(ctx context.Context, monitorID string) (*Heartbeat, error) {
	hb, err := models.Heartbeats.Query(
		sm.Where(models.Heartbeats.Columns.MonitorID.EQ(sqlite.Arg(monitorID))),
		sm.OrderBy(models.Heartbeats.Columns.Time).Desc(),
		sm.Limit(1),
	).One(ctx, r.exc)
	if err != nil {
		return nil, errs.WrapNotFound(err, "querying latest heartbeat") //nolint:wrapcheck // WrapNotFound handles wrapping
	}
	return heartbeatFromModel(hb), nil
}

func (r *sqliteRepo) GetPrevious(ctx context.Context, monitorID string) (*Heartbeat, error) {
	hbs, err := models.Heartbeats.Query(
		sm.Where(models.Heartbeats.Columns.MonitorID.EQ(sqlite.Arg(monitorID))),
		sm.OrderBy(models.Heartbeats.Columns.Time).Desc(),
		sm.Limit(1),
		sm.Offset(1),
	).All(ctx, r.exc)
	if err != nil {
		return nil, fmt.Errorf("querying previous heartbeat: %w", err)
	}
	if len(hbs) == 0 {
		return nil, fmt.Errorf("previous heartbeat: %w", errs.ErrNotFound)
	}
	return heartbeatFromModel(hbs[0]), nil
}

func (r *sqliteRepo) GetByMonitorPaged(ctx context.Context, monitorID string, offset, limit int) ([]*Heartbeat, error) {
	hbs, err := models.Heartbeats.Query(
		sm.Where(models.Heartbeats.Columns.MonitorID.EQ(sqlite.Arg(monitorID))),
		sm.OrderBy(models.Heartbeats.Columns.Time).Desc(),
		sm.Limit(limit),
		sm.Offset(offset),
	).All(ctx, r.exc)
	if err != nil {
		return nil, fmt.Errorf("querying heartbeats: %w", err)
	}
	return heartbeatsFromModels(hbs), nil
}

func (r *sqliteRepo) GetImportantByMonitor(ctx context.Context, monitorID string, offset, limit int) ([]*Heartbeat, int64, error) {
	countQ := models.Heartbeats.Query(
		sm.Where(models.Heartbeats.Columns.MonitorID.EQ(sqlite.Arg(monitorID))),
		sm.Where(models.Heartbeats.Columns.Important.EQ(sqlite.Arg(true))),
	)

	total, err := countQ.Count(ctx, r.exc)
	if err != nil {
		return nil, 0, fmt.Errorf("counting important heartbeats: %w", err)
	}

	hbs, err := models.Heartbeats.Query(
		sm.Where(models.Heartbeats.Columns.MonitorID.EQ(sqlite.Arg(monitorID))),
		sm.Where(models.Heartbeats.Columns.Important.EQ(sqlite.Arg(true))),
		sm.OrderBy(models.Heartbeats.Columns.Time).Desc(),
		sm.Limit(limit),
		sm.Offset(offset),
	).All(ctx, r.exc)
	if err != nil {
		return nil, 0, fmt.Errorf("querying important heartbeats: %w", err)
	}
	return heartbeatsFromModels(hbs), total, nil
}

func (r *sqliteRepo) GetByMonitorInRange(ctx context.Context, monitorID string, from, to int64) ([]*Heartbeat, error) {
	hbs, err := models.Heartbeats.Query(
		sm.Where(sqlite.Raw("monitor_id = ? AND time >= datetime(?, 'unixepoch') AND time < datetime(?, 'unixepoch')", monitorID, from, to)),
		sm.OrderBy(models.Heartbeats.Columns.Time).Asc(),
	).All(ctx, r.exc)
	if err != nil {
		return nil, fmt.Errorf("querying heartbeats in range: %w", err)
	}
	return heartbeatsFromModels(hbs), nil
}

func (r *sqliteRepo) DeleteOlderThan(ctx context.Context, monitorID string, hours int) error {
	_, err := models.Heartbeats.Delete(
		dm.Where(sqlite.Raw("monitor_id = ? AND time < datetime('now', ?)", monitorID, fmt.Sprintf("-%d hours", hours))),
	).Exec(ctx, r.exc)
	if err != nil {
		return fmt.Errorf("deleting old heartbeats: %w", err)
	}
	return nil
}

func (r *sqliteRepo) ClearByMonitor(ctx context.Context, monitorID string) error {
	_, err := models.Heartbeats.Delete(
		dm.Where(models.Heartbeats.Columns.MonitorID.EQ(sqlite.Arg(monitorID))),
	).Exec(ctx, r.exc)
	if err != nil {
		return fmt.Errorf("clearing heartbeats for monitor: %w", err)
	}
	return nil
}

func (r *sqliteRepo) GetAveragePing(ctx context.Context, monitorID string, hours int) (float64, error) {
	var avg sql.NullFloat64
	err := r.db.QueryRowContext(ctx,
		`SELECT AVG(ping) FROM heartbeat WHERE monitor_id = ? AND time >= datetime('now', ?) AND ping IS NOT NULL`,
		monitorID, fmt.Sprintf("-%d hours", hours),
	).Scan(&avg)
	if err != nil {
		return 0, fmt.Errorf("querying average ping: %w", err)
	}
	return avg.Float64, nil
}

func (r *sqliteRepo) GetAverageResponse(ctx context.Context, monitorID string, hours int) (float64, error) {
	var avg sql.NullFloat64
	err := r.db.QueryRowContext(ctx,
		`SELECT AVG(duration) FROM heartbeat WHERE monitor_id = ? AND time >= datetime('now', ?)`,
		monitorID, fmt.Sprintf("-%d hours", hours),
	).Scan(&avg)
	if err != nil {
		return 0, fmt.Errorf("querying average response: %w", err)
	}
	return avg.Float64, nil
}

func (r *sqliteRepo) GetUptime(ctx context.Context, monitorID string, hours int) (float64, error) {
	var total, up int64
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*), COALESCE(SUM(CASE WHEN status = 1 THEN 1 ELSE 0 END), 0) FROM heartbeat WHERE monitor_id = ? AND time >= datetime('now', ?)`,
		monitorID, fmt.Sprintf("-%d hours", hours),
	).Scan(&total, &up)
	if err != nil {
		return 0, fmt.Errorf("querying uptime: %w", err)
	}
	if total == 0 {
		return 1.0, nil
	}
	return float64(up) / float64(total), nil
}

func heartbeatFromModel(m *models.Heartbeat) *Heartbeat {
	hb := &Heartbeat{
		ID:        m.ID,
		MonitorID: m.MonitorID,
		Status:    int(m.Status),
		Time:      RFC3339Time(m.Time),
		Msg:       m.MSG.GetOrZero(),
		Important: m.Important,
		Duration:  m.Duration,
		Retries:   int(m.Retries),
	}
	if v, ok := m.Ping.Get(); ok {
		hb.Ping = &v
	}
	if v, ok := m.Response.Get(); ok {
		hb.Response = v
	}
	return hb
}

func heartbeatsFromModels(ms models.HeartbeatSlice) []*Heartbeat {
	result := make([]*Heartbeat, len(ms))
	for i, m := range ms {
		result[i] = heartbeatFromModel(m)
	}
	return result
}
