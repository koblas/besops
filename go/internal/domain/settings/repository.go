package settings

import "context"

// Repository defines persistence operations for application settings (key-value pairs).
type Repository interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string) error
	GetAll(ctx context.Context) (map[string]string, error)
}
