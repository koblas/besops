package apikey

import "context"

type Repository interface {
	FindByID(ctx context.Context, id string) (*APIKey, error)
	FindAll(ctx context.Context, userID string) ([]*APIKey, error)
	Create(ctx context.Context, key *APIKey) (string, error)
	Delete(ctx context.Context, id string) error
	SetActive(ctx context.Context, id string, active bool) error
}
