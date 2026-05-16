package app

import (
	"context"
	"fmt"

	"github.com/koblas/besops/internal/auth"
	"github.com/koblas/besops/internal/domain/user"
)

// userStoreAdapter adapts user.Repository to auth.UserStore.
type userStoreAdapter struct {
	repo user.Repository
}

func newUserStoreAdapter(repo user.Repository) auth.UserStore {
	return &userStoreAdapter{repo: repo}
}

func (a *userStoreAdapter) FindByUsername(ctx context.Context, username string) (*auth.User, error) {
	u, err := a.repo.FindByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("finding user by username: %w", err)
	}

	return toAuthUser(u), nil
}

func (a *userStoreAdapter) FindByID(ctx context.Context, id string) (*auth.User, error) {
	u, err := a.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("finding user by ID: %w", err)
	}

	return toAuthUser(u), nil
}

func (a *userStoreAdapter) Create(ctx context.Context, username, passwordHash string) (string, error) {
	u := &user.User{
		Username: username,
		Password: passwordHash,
		Active:   true,
	}
	id, err := a.repo.Create(ctx, u)
	if err != nil {
		return "", fmt.Errorf("creating user: %w", err)
	}
	return id, nil
}

func (a *userStoreAdapter) UpdatePassword(ctx context.Context, id string, passwordHash string) error {
	if err := a.repo.UpdatePassword(ctx, id, passwordHash); err != nil {
		return fmt.Errorf("updating password: %w", err)
	}
	return nil
}

func (a *userStoreAdapter) Update2FA(ctx context.Context, id string, enabled bool, secret string) error {
	if err := a.repo.Update2FA(ctx, id, enabled, secret); err != nil {
		return fmt.Errorf("updating 2FA: %w", err)
	}
	return nil
}

func (a *userStoreAdapter) Count(ctx context.Context) (int64, error) {
	count, err := a.repo.Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("counting users: %w", err)
	}
	return count, nil
}

func toAuthUser(u *user.User) *auth.User {
	return &auth.User{
		ID:           u.ID,
		Username:     u.Username,
		PasswordHash: u.Password,
		Active:       u.Active,
		TOTPSecret:   u.TwoFASecret,
		TOTPEnabled:  u.TwoFAStatus,
	}
}
