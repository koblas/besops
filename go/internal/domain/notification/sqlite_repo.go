package notification

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/aarondl/opt/omit"
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

func (r *sqliteRepo) FindByID(ctx context.Context, id string) (*Notification, error) {
	n, err := models.FindNotification(ctx, r.db, id)
	if err != nil {
		return nil, errs.WrapNotFound(err, "finding notification") //nolint:wrapcheck // WrapNotFound handles wrapping
	}
	return notificationFromModel(n), nil
}

func (r *sqliteRepo) FindByUserID(ctx context.Context, userID string) ([]*Notification, error) {
	ns, err := models.Notifications.Query(
		sm.Where(models.Notifications.Columns.UserID.EQ(sqlite.Arg(userID))),
	).All(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("finding notifications by user id: %w", err)
	}

	result := make([]*Notification, len(ns))
	for i, n := range ns {
		result[i] = notificationFromModel(n)
	}
	return result, nil
}

func (r *sqliteRepo) GetForMonitor(ctx context.Context, monitorID string) ([]*Notification, error) {
	links, err := models.MonitorNotifications.Query(
		sm.Where(models.MonitorNotifications.Columns.MonitorID.EQ(sqlite.Arg(monitorID))),
	).All(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("querying monitor notification links: %w", err)
	}

	if len(links) == 0 {
		return nil, nil
	}

	ids := make([]any, len(links))
	for i, l := range links {
		ids[i] = l.NotificationID
	}

	ns, err := models.Notifications.Query(
		sm.Where(models.Notifications.Columns.ID.In(sqlite.Arg(ids...))),
	).All(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("getting notifications for monitor: %w", err)
	}

	result := make([]*Notification, len(ns))
	for i, n := range ns {
		result[i] = notificationFromModel(n)
	}
	return result, nil
}

func (r *sqliteRepo) Create(ctx context.Context, n *Notification) (string, error) {
	n.ID = uuid.New().String()

	_, err := models.Notifications.Insert(&models.NotificationSetter{
		ID:     omit.From(n.ID),
		Name:   omit.From(n.Name),
		UserID: omit.From(n.UserID),
		Config: omit.From(n.Config),
		Active: omit.From(n.Active),
	}).One(ctx, r.db)
	if err != nil {
		return "", fmt.Errorf("creating notification: %w", err)
	}
	return n.ID, nil
}

func (r *sqliteRepo) Update(ctx context.Context, n *Notification) error {
	existing, err := models.FindNotification(ctx, r.db, n.ID)
	if err != nil {
		return fmt.Errorf("finding notification for update: %w", err)
	}

	if err := existing.Update(ctx, r.db, &models.NotificationSetter{
		Name:   omit.From(n.Name),
		Config: omit.From(n.Config),
		Active: omit.From(n.Active),
	}); err != nil {
		return fmt.Errorf("updating notification: %w", err)
	}
	return nil
}

func (r *sqliteRepo) Delete(ctx context.Context, id string) error {
	n, err := models.FindNotification(ctx, r.db, id)
	if err != nil {
		return fmt.Errorf("finding notification for delete: %w", err)
	}

	if err := n.Delete(ctx, r.db); err != nil {
		return fmt.Errorf("deleting notification: %w", err)
	}
	return nil
}

func notificationFromModel(m *models.Notification) *Notification {
	return &Notification{
		ID:     m.ID,
		Name:   m.Name,
		UserID: m.UserID,
		Config: m.Config,
		Active: m.Active,
	}
}
