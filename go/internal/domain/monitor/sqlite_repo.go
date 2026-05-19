package monitor

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/aarondl/opt/omit"
	"github.com/aarondl/opt/omitnull"
	"github.com/google/uuid"
	"github.com/koblas/besops/lib/errs"
	"github.com/koblas/besops/models"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/sqlite"
	"github.com/stephenafamo/bob/dialect/sqlite/sm"
)

type repo struct {
	db bob.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repo{db: bob.NewDB(db)}
}

func (r *repo) FindByID(ctx context.Context, id string) (*Monitor, error) {
	m, err := models.FindMonitor(ctx, r.db, id)
	if err != nil {
		return nil, errs.WrapNotFound(err, "finding monitor") //nolint:wrapcheck // WrapNotFound handles wrapping
	}
	return monitorFromModel(m), nil
}

func (r *repo) FindByUserID(ctx context.Context, userID string) ([]*Monitor, error) {
	ms, err := models.Monitors.Query(
		sm.Where(models.Monitors.Columns.UserID.EQ(sqlite.Arg(userID))),
	).All(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("querying monitors by user: %w", err)
	}
	return monitorsFromModels(ms), nil
}

func (r *repo) FindAllActiveIDs(ctx context.Context) ([]string, error) {
	ms, err := models.Monitors.Query(
		sm.Where(models.Monitors.Columns.Active.EQ(sqlite.Arg(true))),
		sm.Columns(models.Monitors.Columns.Only("id")),
	).All(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("querying active monitor IDs: %w", err)
	}

	ids := make([]string, len(ms))
	for i, m := range ms {
		ids[i] = m.ID
	}
	return ids, nil
}

func (r *repo) Create(ctx context.Context, m *Monitor) (string, error) {
	m.ID = uuid.New().String()

	_, err := models.Monitors.Insert(monitorToSetter(m)).One(ctx, r.db)
	if err != nil {
		return "", fmt.Errorf("inserting monitor: %w", err)
	}
	return m.ID, nil
}

func (r *repo) Update(ctx context.Context, m *Monitor) error {
	existing, err := models.FindMonitor(ctx, r.db, m.ID)
	if err != nil {
		return fmt.Errorf("finding monitor for update: %w", err)
	}

	setter := monitorToSetter(m)
	setter.ID = omit.Val[string]{}

	if err := existing.Update(ctx, r.db, setter); err != nil {
		return fmt.Errorf("updating monitor: %w", err)
	}
	return nil
}

func (r *repo) Delete(ctx context.Context, id string) error {
	m, err := models.FindMonitor(ctx, r.db, id)
	if err != nil {
		return fmt.Errorf("finding monitor for delete: %w", err)
	}

	if err := m.Delete(ctx, r.db); err != nil {
		return fmt.Errorf("deleting monitor: %w", err)
	}
	return nil
}

func (r *repo) FindByTagIDs(ctx context.Context, tagIDs []string) ([]*Monitor, error) {
	if len(tagIDs) == 0 {
		return nil, nil
	}

	args := make([]bob.Expression, len(tagIDs))
	for i, id := range tagIDs {
		args[i] = sqlite.Arg(id)
	}

	ms, err := models.Monitors.Query(
		sm.InnerJoin(sqlite.Quote("monitor_tag")).OnEQ(
			sqlite.Quote("monitor_tag", "monitor_id"),
			models.Monitors.Columns.ID,
		),
		sm.Where(sqlite.Quote("monitor_tag", "tag_id").In(args...)),
		sm.GroupBy(models.Monitors.Columns.ID),
	).All(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("querying monitors by tag IDs: %w", err)
	}
	return monitorsFromModels(ms), nil
}

func monitorToSetter(m *Monitor) *models.MonitorSetter {
	return &models.MonitorSetter{
		ID:                 omit.From(m.ID),
		Name:               omit.From(m.Name),
		Active:             omit.From(m.Active),
		UserID:             omit.From(m.UserID),
		Interval:           omit.From(int64(m.Interval)),
		Type:               omit.From(m.Type),
		Weight:             omitnull.From(int64(m.Weight)),
		Maxretries:         omit.From(int64(m.MaxRetries)),
		UpsideDown:         omit.From(m.UpsideDown),
		RetryInterval:      omit.From(int64(m.RetryInterval)),
		Timeout:            omit.From(int64(m.Timeout)),
		Description:        omitnull.From(m.Description),
		ResendInterval:     omit.From(int64(m.ResendInterval)),
		ExpiryNotification: omitnull.From(m.ExpiryNotification),
		ConfigJSON:         omit.From(m.ConfigJSON),
	}
}

func monitorFromModel(m *models.Monitor) *Monitor {
	mon := &Monitor{
		ID:             m.ID,
		Name:           m.Name,
		Active:         m.Active,
		UserID:         m.UserID,
		Interval:       int(m.Interval),
		Type:           m.Type,
		Weight:         int(m.Weight.GetOrZero()),
		CreatedDate:    m.CreatedDate,
		MaxRetries:     int(m.Maxretries),
		UpsideDown:     m.UpsideDown,
		RetryInterval:  int(m.RetryInterval),
		Timeout:        int(m.Timeout),
		Description:    m.Description.GetOrZero(),
		ResendInterval: int(m.ResendInterval),
		ConfigJSON:     m.ConfigJSON,
	}

	if v, ok := m.ExpiryNotification.Get(); ok {
		mon.ExpiryNotification = v
	}

	return mon
}

func monitorsFromModels(ms models.MonitorSlice) []*Monitor {
	result := make([]*Monitor, len(ms))
	for i, m := range ms {
		result[i] = monitorFromModel(m)
	}
	return result
}
