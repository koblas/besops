package apikey

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

func (r *sqliteRepo) FindByID(ctx context.Context, id string) (*APIKey, error) {
	k, err := models.FindAPIKey(ctx, r.db, id)
	if err != nil {
		return nil, errs.WrapNotFound(err, "finding api key") //nolint:wrapcheck // WrapNotFound handles wrapping
	}
	return apiKeyFromModel(k), nil
}

func (r *sqliteRepo) FindAll(ctx context.Context, userID string) ([]*APIKey, error) {
	keys, err := models.APIKeys.Query(
		sm.Where(models.APIKeys.Columns.UserID.EQ(sqlite.Arg(userID))),
	).All(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("querying api keys: %w", err)
	}

	result := make([]*APIKey, len(keys))
	for i, k := range keys {
		result[i] = apiKeyFromModel(k)
	}
	return result, nil
}

func (r *sqliteRepo) Create(ctx context.Context, key *APIKey) (string, error) {
	key.ID = uuid.New().String()

	setter := &models.APIKeySetter{
		ID:     omit.From(key.ID),
		Key:    omit.From(key.Key),
		Name:   omit.From(key.Name),
		UserID: omit.From(key.UserID),
		Active: omit.From(key.Active),
	}
	if key.Expires != nil {
		setter.Expires = omitnull.From(*key.Expires)
	}

	_, err := models.APIKeys.Insert(setter).One(ctx, r.db)
	if err != nil {
		return "", fmt.Errorf("inserting api key: %w", err)
	}
	return key.ID, nil
}

func (r *sqliteRepo) Delete(ctx context.Context, id string) error {
	k, err := models.FindAPIKey(ctx, r.db, id)
	if err != nil {
		return fmt.Errorf("finding api key for delete: %w", err)
	}

	if err := k.Delete(ctx, r.db); err != nil {
		return fmt.Errorf("deleting api key: %w", err)
	}
	return nil
}

func (r *sqliteRepo) SetActive(ctx context.Context, id string, active bool) error {
	k, err := models.FindAPIKey(ctx, r.db, id)
	if err != nil {
		return fmt.Errorf("finding api key for update: %w", err)
	}

	if err := k.Update(ctx, r.db, &models.APIKeySetter{
		Active: omit.From(active),
	}); err != nil {
		return fmt.Errorf("updating api key active status: %w", err)
	}
	return nil
}

func apiKeyFromModel(m *models.APIKey) *APIKey {
	k := &APIKey{
		ID:        m.ID,
		Key:       m.Key,
		Name:      m.Name,
		UserID:    m.UserID,
		CreatedAt: m.CreatedDate,
		Active:    m.Active,
	}
	if v, ok := m.Expires.Get(); ok {
		k.Expires = &v
	}
	return k
}
