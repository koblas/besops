package statuspage

import "context"

type Repository interface {
	// Status pages
	FindAll(ctx context.Context) ([]*StatusPage, error)
	FindBySlug(ctx context.Context, slug string) (*StatusPage, error)
	Create(ctx context.Context, sp *StatusPage) (string, error)
	Update(ctx context.Context, sp *StatusPage) error
	Delete(ctx context.Context, id string) error

	// Groups
	GetGroups(ctx context.Context, statusPageID string) ([]*Group, error)
	SaveGroups(ctx context.Context, statusPageID string, groups []*Group) error

	// Monitor groups
	GetMonitorGroups(ctx context.Context, groupID string) ([]*MonitorGroup, error)
	SaveMonitorGroups(ctx context.Context, groupID string, monitorGroups []*MonitorGroup) error

	// Incidents
	FindIncidentsByStatusPage(ctx context.Context, statusPageID string) ([]*Incident, error)
	FindIncidentByID(ctx context.Context, id string) (*Incident, error)
	CreateIncident(ctx context.Context, inc *Incident) (string, error)
	UpdateIncident(ctx context.Context, inc *Incident) error
	DeleteIncident(ctx context.Context, id string) error
}
