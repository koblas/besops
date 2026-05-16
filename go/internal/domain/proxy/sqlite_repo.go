package proxy

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

func (r *sqliteRepo) FindAll(ctx context.Context, userID string) ([]*Proxy, error) {
	ps, err := models.Proxies.Query(
		sm.Where(models.Proxies.Columns.UserID.EQ(sqlite.Arg(userID))),
	).All(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("querying proxies: %w", err)
	}

	result := make([]*Proxy, len(ps))
	for i, p := range ps {
		result[i] = proxyFromModel(p)
	}
	return result, nil
}

func (r *sqliteRepo) FindByID(ctx context.Context, id string) (*Proxy, error) {
	p, err := models.FindProxy(ctx, r.db, id)
	if err != nil {
		return nil, errs.WrapNotFound(err, "finding proxy") //nolint:wrapcheck // WrapNotFound handles wrapping
	}
	return proxyFromModel(p), nil
}

func (r *sqliteRepo) Create(ctx context.Context, p *Proxy) (string, error) {
	p.ID = uuid.New().String()

	_, err := models.Proxies.Insert(&models.ProxySetter{
		ID:       omit.From(p.ID),
		UserID:   omit.From(p.UserID),
		Protocol: omit.From(p.Protocol),
		Host:     omit.From(p.Host),
		Port:     omit.From(p.Port),
		Auth:     omit.From(p.Auth),
		Username: omitnull.From(p.Username),
		Password: omitnull.From(p.Password),
		Active:   omit.From(p.Active),
		Default:  omit.From(p.Default),
	}).One(ctx, r.db)
	if err != nil {
		return "", fmt.Errorf("creating proxy: %w", err)
	}
	return p.ID, nil
}

func (r *sqliteRepo) Update(ctx context.Context, p *Proxy) error {
	existing, err := models.FindProxy(ctx, r.db, p.ID)
	if err != nil {
		return errs.WrapNotFound(err, "finding proxy for update") //nolint:wrapcheck // WrapNotFound handles wrapping
	}

	if err := existing.Update(ctx, r.db, &models.ProxySetter{
		Protocol: omit.From(p.Protocol),
		Host:     omit.From(p.Host),
		Port:     omit.From(p.Port),
		Auth:     omit.From(p.Auth),
		Username: omitnull.From(p.Username),
		Password: omitnull.From(p.Password),
		Active:   omit.From(p.Active),
		Default:  omit.From(p.Default),
	}); err != nil {
		return fmt.Errorf("updating proxy: %w", err)
	}
	return nil
}

func (r *sqliteRepo) Delete(ctx context.Context, id string) error {
	p, err := models.FindProxy(ctx, r.db, id)
	if err != nil {
		return errs.WrapNotFound(err, "finding proxy for delete") //nolint:wrapcheck // WrapNotFound handles wrapping
	}

	if err := p.Delete(ctx, r.db); err != nil {
		return fmt.Errorf("deleting proxy: %w", err)
	}
	return nil
}

func proxyFromModel(m *models.Proxy) *Proxy {
	return &Proxy{
		ID:       m.ID,
		UserID:   m.UserID,
		Protocol: m.Protocol,
		Host:     m.Host,
		Port:     m.Port,
		Auth:     m.Auth,
		Username: m.Username.GetOrZero(),
		Password: m.Password.GetOrZero(),
		Active:   m.Active,
		Default:  m.Default,
	}
}
