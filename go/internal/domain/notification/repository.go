package notification

import "context"

type Repository interface {
	FindByID(ctx context.Context, id string) (*Notification, error)
	FindByUserID(ctx context.Context, userID string) ([]*Notification, error)
	GetForMonitor(ctx context.Context, monitorID string) ([]*Notification, error)
	Create(ctx context.Context, n *Notification) (string, error)
	Update(ctx context.Context, n *Notification) error
	Delete(ctx context.Context, id string) error
}
