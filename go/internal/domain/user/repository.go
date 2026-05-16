package user

import "context"

type Repository interface {
	FindByUsername(ctx context.Context, username string) (*User, error)
	FindByID(ctx context.Context, id string) (*User, error)
	Create(ctx context.Context, u *User) (string, error)
	UpdatePassword(ctx context.Context, id string, hash string) error
	Update2FA(ctx context.Context, id string, enabled bool, secret string) error
	Count(ctx context.Context) (int64, error)
}
