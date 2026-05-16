package maintenance

import "context"

type Repository interface {
	FindAll(ctx context.Context) ([]*Maintenance, error)
	FindByID(ctx context.Context, id string) (*Maintenance, error)
	GetMonitorMaintenanceIDs(ctx context.Context, monitorID string) ([]string, error)
	Create(ctx context.Context, m *Maintenance) (string, error)
	Update(ctx context.Context, m *Maintenance) error
	Delete(ctx context.Context, id string) error
	SetMonitorIDs(ctx context.Context, maintenanceID string, monitorIDs []string) error
	GetStatusPageIDs(ctx context.Context, maintenanceID string) ([]string, error)
	SetStatusPageIDs(ctx context.Context, maintenanceID string, statusPageIDs []string) error
}
