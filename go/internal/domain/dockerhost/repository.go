package dockerhost

import "context"

type Repository interface {
	FindAll(ctx context.Context, userID string) ([]*DockerHost, error)
	FindByID(ctx context.Context, id string) (*DockerHost, error)
	Create(ctx context.Context, dh *DockerHost) (string, error)
	Update(ctx context.Context, dh *DockerHost) error
	Delete(ctx context.Context, id string) error
}
