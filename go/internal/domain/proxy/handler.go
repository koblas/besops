package proxy

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	oas "github.com/koblas/besops/internal/api/oas_generated"
	"github.com/koblas/besops/internal/api/oasutil"
)

var _ oas.ProxyHandler = (*Handler)(nil)

type Handler struct {
	repo Repository
}

func NewHandler(repo Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) ListProxies(ctx context.Context) (oas.ListProxiesRes, error) {
	userID, err := oasutil.UserIDFromCtx(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting user from context: %w", err)
	}

	proxies, err := h.repo.FindAll(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("listing proxies: %w", err)
	}

	result := make(oas.ListProxiesOKApplicationJSON, len(proxies))
	for i, p := range proxies {
		result[i] = proxyToOAS(p)
	}
	return &result, nil
}

func (h *Handler) CreateProxy(ctx context.Context, req *oas.ProxyInput) (oas.CreateProxyRes, error) {
	userID, err := oasutil.UserIDFromCtx(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting user from context: %w", err)
	}

	p := &Proxy{
		UserID:   userID,
		Protocol: string(req.Protocol),
		Host:     req.Host,
		Port:     int64(req.Port),
		Auth:     oasutil.OptBoolValue(req.Auth, false),
		Username: oasutil.OptStringValue(req.Username),
		Password: oasutil.OptStringValue(req.Password),
		Active:   oasutil.OptBoolValue(req.Active, true),
		Default:  oasutil.OptBoolValue(req.Default, false),
	}

	id, createErr := h.repo.Create(ctx, p)
	if createErr != nil {
		return nil, fmt.Errorf("creating proxy: %w", createErr)
	}

	return &oas.CreateProxyCreated{ID: uuid.MustParse(id)}, nil
}

func (h *Handler) UpdateProxy(ctx context.Context, req *oas.ProxyInput, params oas.UpdateProxyParams) (oas.UpdateProxyRes, error) {
	userID, err := oasutil.UserIDFromCtx(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting user from context: %w", err)
	}

	p := &Proxy{
		ID:       params.ProxyId.String(),
		UserID:   userID,
		Protocol: string(req.Protocol),
		Host:     req.Host,
		Port:     int64(req.Port),
		Auth:     oasutil.OptBoolValue(req.Auth, false),
		Username: oasutil.OptStringValue(req.Username),
		Password: oasutil.OptStringValue(req.Password),
		Active:   oasutil.OptBoolValue(req.Active, true),
		Default:  oasutil.OptBoolValue(req.Default, false),
	}

	if updateErr := h.repo.Update(ctx, p); updateErr != nil {
		return nil, fmt.Errorf("updating proxy: %w", updateErr)
	}
	return &oas.UpdateProxyOK{}, nil
}

func (h *Handler) DeleteProxy(ctx context.Context, params oas.DeleteProxyParams) (oas.DeleteProxyRes, error) {
	if err := h.repo.Delete(ctx, params.ProxyId.String()); err != nil {
		return nil, fmt.Errorf("deleting proxy: %w", err)
	}
	return &oas.DeleteProxyNoContent{}, nil
}

func proxyToOAS(p *Proxy) oas.Proxy {
	return oas.Proxy{
		ID:       uuid.MustParse(p.ID),
		Protocol: oas.ProxyProtocol(p.Protocol),
		Host:     p.Host,
		Port:     int32(p.Port), //nolint:gosec // port values are 0-65535, no overflow
		Active:   p.Active,
		Default:  oas.NewOptBool(p.Default),
		Auth:     oas.NewOptBool(p.Auth),
		Username: oasutil.PtrToOptString(p.Username),
		Password: oasutil.PtrToOptString(p.Password),
	}
}
