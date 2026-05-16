package user

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

func (r *repo) FindByUsername(ctx context.Context, username string) (*User, error) {
	u, err := models.Users.Query(
		sm.Where(models.Users.Columns.Username.EQ(sqlite.Arg(username))),
	).One(ctx, r.db)
	if err != nil {
		return nil, errs.WrapNotFound(err, "finding user by username") //nolint:wrapcheck // WrapNotFound handles wrapping
	}
	return userFromModel(u), nil
}

func (r *repo) FindByID(ctx context.Context, id string) (*User, error) {
	u, err := models.FindUser(ctx, r.db, id)
	if err != nil {
		return nil, errs.WrapNotFound(err, "finding user") //nolint:wrapcheck // WrapNotFound handles wrapping
	}
	return userFromModel(u), nil
}

func (r *repo) Create(ctx context.Context, u *User) (string, error) {
	u.ID = uuid.New().String()

	_, err := models.Users.Insert(&models.UserSetter{
		ID:             omit.From(u.ID),
		Username:       omit.From(u.Username),
		Password:       omit.From(u.Password),
		Active:         omit.From(u.Active),
		TwofaStatus:    omit.From(u.TwoFAStatus),
		TwofaSecret:    omitnull.From(u.TwoFASecret),
		TwofaLastToken: omitnull.From(u.TwoFALastToken),
	}).One(ctx, r.db)
	if err != nil {
		return "", fmt.Errorf("inserting user: %w", err)
	}
	return u.ID, nil
}

func (r *repo) UpdatePassword(ctx context.Context, id string, hash string) error {
	u, err := models.FindUser(ctx, r.db, id)
	if err != nil {
		return fmt.Errorf("finding user for password update: %w", err)
	}

	if err := u.Update(ctx, r.db, &models.UserSetter{
		Password: omit.From(hash),
	}); err != nil {
		return fmt.Errorf("updating password: %w", err)
	}
	return nil
}

func (r *repo) Update2FA(ctx context.Context, id string, enabled bool, secret string) error {
	u, err := models.FindUser(ctx, r.db, id)
	if err != nil {
		return fmt.Errorf("finding user for 2FA update: %w", err)
	}

	if err := u.Update(ctx, r.db, &models.UserSetter{
		TwofaStatus: omit.From(enabled),
		TwofaSecret: omitnull.From(secret),
	}); err != nil {
		return fmt.Errorf("updating 2FA: %w", err)
	}
	return nil
}

func (r *repo) Count(ctx context.Context) (int64, error) {
	count, err := models.Users.Query().Count(ctx, r.db)
	if err != nil {
		return 0, fmt.Errorf("counting users: %w", err)
	}
	return count, nil
}

func userFromModel(m *models.User) *User {
	return &User{
		ID:             m.ID,
		Username:       m.Username,
		Password:       m.Password,
		Active:         m.Active,
		TwoFAStatus:    m.TwofaStatus,
		TwoFASecret:    m.TwofaSecret.GetOrZero(),
		TwoFALastToken: m.TwofaLastToken.GetOrZero(),
	}
}
