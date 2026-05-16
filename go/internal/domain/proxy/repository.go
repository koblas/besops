package proxy

import "context"

type Repository interface {
	FindAll(ctx context.Context, userID string) ([]*Proxy, error)
	FindByID(ctx context.Context, id string) (*Proxy, error)
	Create(ctx context.Context, p *Proxy) (string, error)
	Update(ctx context.Context, p *Proxy) error
	Delete(ctx context.Context, id string) error
}
