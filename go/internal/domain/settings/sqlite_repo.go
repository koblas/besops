package settings

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

func (r *sqliteRepo) Get(ctx context.Context, key string) (string, error) {
	s, err := models.Settings.Query(
		sm.Where(models.Settings.Columns.Key.EQ(sqlite.Arg(key))),
	).One(ctx, r.db)
	if err != nil {
		return "", errs.WrapNotFound(err, fmt.Sprintf("getting setting %q", key)) //nolint:wrapcheck // WrapNotFound handles wrapping
	}
	return s.Value.GetOrZero(), nil
}

func (r *sqliteRepo) Set(ctx context.Context, key, value string) error {
	existing, err := models.Settings.Query(
		sm.Where(models.Settings.Columns.Key.EQ(sqlite.Arg(key))),
	).One(ctx, r.db)
	if err == nil {
		if updateErr := existing.Update(ctx, r.db, &models.SettingSetter{
			Value: omitnull.From(value),
		}); updateErr != nil {
			return fmt.Errorf("updating setting %q: %w", key, updateErr)
		}
		return nil
	}

	_, err = models.Settings.Insert(&models.SettingSetter{
		ID:    omit.From(uuid.New().String()),
		Key:   omit.From(key),
		Value: omitnull.From(value),
	}).One(ctx, r.db)
	if err != nil {
		return fmt.Errorf("inserting setting %q: %w", key, err)
	}
	return nil
}

func (r *sqliteRepo) GetAll(ctx context.Context) (map[string]string, error) {
	all, err := models.Settings.Query().All(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("querying settings: %w", err)
	}

	result := make(map[string]string, len(all))
	for _, s := range all {
		result[s.Key] = s.Value.GetOrZero()
	}
	return result, nil
}
