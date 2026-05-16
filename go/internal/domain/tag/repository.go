package tag

import "context"

type Repository interface {
	FindAll(ctx context.Context) ([]*Tag, error)
	FindByID(ctx context.Context, id string) (*Tag, error)
	Create(ctx context.Context, t *Tag) (string, error)
	Update(ctx context.Context, t *Tag) error
	Delete(ctx context.Context, id string) error
	GetForMonitor(ctx context.Context, monitorID string) ([]*MonitorTag, error)
	AddToMonitor(ctx context.Context, monitorID, tagID string, value string) error
	RemoveFromMonitor(ctx context.Context, monitorID, tagID string) error
}
