package maintenance

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

type sqliteRepo struct {
	db bob.DB
}

func NewRepository(db *sql.DB) Repository {
	return &sqliteRepo{db: bob.NewDB(db)}
}

func (r *sqliteRepo) FindAll(ctx context.Context) ([]*Maintenance, error) {
	ms, err := models.Maintenances.Query().All(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("querying maintenance windows: %w", err)
	}

	result := make([]*Maintenance, len(ms))
	for i, m := range ms {
		result[i] = maintenanceFromModel(m)
	}
	return result, nil
}

func (r *sqliteRepo) FindByID(ctx context.Context, id string) (*Maintenance, error) {
	m, err := models.FindMaintenance(ctx, r.db, id)
	if err != nil {
		return nil, errs.WrapNotFound(err, "finding maintenance") //nolint:wrapcheck // WrapNotFound handles wrapping
	}
	return maintenanceFromModel(m), nil
}

func (r *sqliteRepo) GetMonitorMaintenanceIDs(ctx context.Context, monitorID string) ([]string, error) {
	links, err := models.MonitorMaintenances.Query(
		sm.Where(models.MonitorMaintenances.Columns.MonitorID.EQ(sqlite.Arg(monitorID))),
	).All(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("querying monitor maintenance IDs: %w", err)
	}

	ids := make([]string, len(links))
	for i, l := range links {
		ids[i] = l.MaintenanceID
	}
	return ids, nil
}

func (r *sqliteRepo) Create(ctx context.Context, m *Maintenance) (string, error) {
	m.ID = uuid.New().String()

	_, err := models.Maintenances.Insert(maintenanceToSetter(m)).One(ctx, r.db)
	if err != nil {
		return "", fmt.Errorf("creating maintenance: %w", err)
	}
	return m.ID, nil
}

func (r *sqliteRepo) Update(ctx context.Context, m *Maintenance) error {
	existing, err := models.FindMaintenance(ctx, r.db, m.ID)
	if err != nil {
		return fmt.Errorf("finding maintenance for update: %w", err)
	}

	setter := maintenanceToSetter(m)
	setter.ID = omit.Val[string]{}

	if err := existing.Update(ctx, r.db, setter); err != nil {
		return fmt.Errorf("updating maintenance: %w", err)
	}
	return nil
}

func (r *sqliteRepo) Delete(ctx context.Context, id string) error {
	m, err := models.FindMaintenance(ctx, r.db, id)
	if err != nil {
		return fmt.Errorf("finding maintenance for delete: %w", err)
	}

	if err := m.Delete(ctx, r.db); err != nil {
		return fmt.Errorf("deleting maintenance: %w", err)
	}
	return nil
}

func (r *sqliteRepo) SetMonitorIDs(ctx context.Context, maintenanceID string, monitorIDs []string) error {
	existing, err := models.MonitorMaintenances.Query(
		sm.Where(models.MonitorMaintenances.Columns.MaintenanceID.EQ(sqlite.Arg(maintenanceID))),
	).All(ctx, r.db)
	if err != nil {
		return fmt.Errorf("querying existing monitor maintenances: %w", err)
	}

	if err := existing.DeleteAll(ctx, r.db); err != nil {
		return fmt.Errorf("deleting existing monitor maintenances: %w", err)
	}

	for _, monitorID := range monitorIDs {
		_, insertErr := models.MonitorMaintenances.Insert(&models.MonitorMaintenanceSetter{
			ID:            omit.From(uuid.New().String()),
			MonitorID:     omit.From(monitorID),
			MaintenanceID: omit.From(maintenanceID),
		}).One(ctx, r.db)
		if insertErr != nil {
			return fmt.Errorf("inserting monitor maintenance link: %w", insertErr)
		}
	}
	return nil
}

func (r *sqliteRepo) GetStatusPageIDs(ctx context.Context, maintenanceID string) ([]string, error) {
	links, err := models.MaintenanceStatusPages.Query(
		sm.Where(models.MaintenanceStatusPages.Columns.MaintenanceID.EQ(sqlite.Arg(maintenanceID))),
	).All(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("querying maintenance status page IDs: %w", err)
	}

	ids := make([]string, len(links))
	for i, l := range links {
		ids[i] = l.StatusPageID
	}
	return ids, nil
}

func (r *sqliteRepo) SetStatusPageIDs(ctx context.Context, maintenanceID string, statusPageIDs []string) error {
	existing, err := models.MaintenanceStatusPages.Query(
		sm.Where(models.MaintenanceStatusPages.Columns.MaintenanceID.EQ(sqlite.Arg(maintenanceID))),
	).All(ctx, r.db)
	if err != nil {
		return fmt.Errorf("querying existing maintenance status pages: %w", err)
	}

	if err := existing.DeleteAll(ctx, r.db); err != nil {
		return fmt.Errorf("deleting existing maintenance status pages: %w", err)
	}

	for _, spID := range statusPageIDs {
		_, insertErr := models.MaintenanceStatusPages.Insert(&models.MaintenanceStatusPageSetter{
			ID:            omit.From(uuid.New().String()),
			MaintenanceID: omit.From(maintenanceID),
			StatusPageID:  omit.From(spID),
		}).One(ctx, r.db)
		if insertErr != nil {
			return fmt.Errorf("inserting maintenance status page link: %w", insertErr)
		}
	}
	return nil
}

func maintenanceToSetter(m *Maintenance) *models.MaintenanceSetter {
	s := &models.MaintenanceSetter{
		ID:          omit.From(m.ID),
		Title:       omit.From(m.Title),
		Description: omitnull.From(m.Description),
		UserID:      omit.From(m.UserID),
		Active:      omit.From(m.Active),
		Strategy:    omit.From(m.Strategy),
		StartTime:   omitnull.From(m.StartTime),
		EndTime:     omitnull.From(m.EndTime),
		Weekdays:    omitnull.From(m.Weekdays),
		DaysOfMonth: omitnull.From(m.DaysOfMonth),
		IntervalDay: omitnull.From(int64(m.IntervalDay)),
		Cron:        omitnull.From(m.CronExpression),
		Timezone:    omitnull.From(m.TimezoneOption),
		Duration:    omitnull.From(int64(m.DurationMinutes)),
	}

	if m.StartDate != nil {
		s.StartDate = omitnull.From(*m.StartDate)
	}
	if m.EndDate != nil {
		s.EndDate = omitnull.From(*m.EndDate)
	}

	return s
}

func maintenanceFromModel(m *models.Maintenance) *Maintenance {
	result := &Maintenance{
		ID:              m.ID,
		Title:           m.Title,
		Description:     m.Description.GetOrZero(),
		UserID:          m.UserID,
		Active:          m.Active,
		Strategy:        m.Strategy,
		StartTime:       m.StartTime.GetOrZero(),
		EndTime:         m.EndTime.GetOrZero(),
		Weekdays:        m.Weekdays.GetOrZero(),
		DaysOfMonth:     m.DaysOfMonth.GetOrZero(),
		IntervalDay:     int(m.IntervalDay.GetOrZero()),
		CronExpression:  m.Cron.GetOrZero(),
		DurationMinutes: int(m.Duration.GetOrZero()),
		TimezoneOption:  m.Timezone.GetOrZero(),
	}

	if v, ok := m.StartDate.Get(); ok {
		result.StartDate = &v
	}
	if v, ok := m.EndDate.Get(); ok {
		result.EndDate = &v
	}

	return result
}
