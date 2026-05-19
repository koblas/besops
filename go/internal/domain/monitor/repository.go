package monitor

import "context"

type Repository interface {
	FindByID(ctx context.Context, id string) (*Monitor, error)
	FindByUserID(ctx context.Context, userID string) ([]*Monitor, error)
	FindAllActiveIDs(ctx context.Context) ([]string, error)
	Create(ctx context.Context, m *Monitor) (string, error)
	Update(ctx context.Context, m *Monitor) error
	Delete(ctx context.Context, id string) error
	FindByTagIDs(ctx context.Context, tagIDs []string) ([]*Monitor, error)
}
