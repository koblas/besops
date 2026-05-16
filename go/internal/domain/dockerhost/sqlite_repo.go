package dockerhost

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

func (r *sqliteRepo) FindAll(ctx context.Context, userID string) ([]*DockerHost, error) {
	hosts, err := models.DockerHosts.Query(
		sm.Where(models.DockerHosts.Columns.UserID.EQ(sqlite.Arg(userID))),
	).All(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("querying docker hosts: %w", err)
	}

	result := make([]*DockerHost, len(hosts))
	for i, h := range hosts {
		result[i] = dockerHostFromModel(h)
	}
	return result, nil
}

func (r *sqliteRepo) FindByID(ctx context.Context, id string) (*DockerHost, error) {
	h, err := models.FindDockerHost(ctx, r.db, id)
	if err != nil {
		return nil, errs.WrapNotFound(err, "finding docker host") //nolint:wrapcheck // WrapNotFound handles wrapping
	}
	return dockerHostFromModel(h), nil
}

func (r *sqliteRepo) Create(ctx context.Context, dh *DockerHost) (string, error) {
	dh.ID = uuid.New().String()

	_, err := models.DockerHosts.Insert(&models.DockerHostSetter{
		ID:           omit.From(dh.ID),
		UserID:       omit.From(dh.UserID),
		Name:         omit.From(dh.Name),
		DockerType:   omit.From(dh.DockerType),
		DockerDaemon: omitnull.From(dh.DockerDaemon),
	}).One(ctx, r.db)
	if err != nil {
		return "", fmt.Errorf("creating docker host: %w", err)
	}
	return dh.ID, nil
}

func (r *sqliteRepo) Update(ctx context.Context, dh *DockerHost) error {
	existing, err := models.FindDockerHost(ctx, r.db, dh.ID)
	if err != nil {
		return errs.WrapNotFound(err, "finding docker host for update") //nolint:wrapcheck // WrapNotFound handles wrapping
	}

	if err := existing.Update(ctx, r.db, &models.DockerHostSetter{
		Name:         omit.From(dh.Name),
		DockerType:   omit.From(dh.DockerType),
		DockerDaemon: omitnull.From(dh.DockerDaemon),
	}); err != nil {
		return fmt.Errorf("updating docker host: %w", err)
	}
	return nil
}

func (r *sqliteRepo) Delete(ctx context.Context, id string) error {
	h, err := models.FindDockerHost(ctx, r.db, id)
	if err != nil {
		return errs.WrapNotFound(err, "finding docker host for delete") //nolint:wrapcheck // WrapNotFound handles wrapping
	}

	if err := h.Delete(ctx, r.db); err != nil {
		return fmt.Errorf("deleting docker host: %w", err)
	}
	return nil
}

func dockerHostFromModel(m *models.DockerHost) *DockerHost {
	return &DockerHost{
		ID:           m.ID,
		UserID:       m.UserID,
		Name:         m.Name,
		DockerType:   m.DockerType,
		DockerDaemon: m.DockerDaemon.GetOrZero(),
	}
}
