package dockerhost

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	oas "github.com/koblas/besops/internal/api/oas_generated"
	"github.com/koblas/besops/internal/api/oasutil"
)

var _ oas.DockerHostHandler = (*Handler)(nil)

type Handler struct {
	repo Repository
}

func NewHandler(repo Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) ListDockerHosts(ctx context.Context) ([]oas.DockerHost, error) {
	userID, err := oasutil.UserIDFromCtx(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting user from context: %w", err)
	}

	hosts, err := h.repo.FindAll(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("listing docker hosts: %w", err)
	}

	result := make([]oas.DockerHost, len(hosts))
	for i, dh := range hosts {
		result[i] = dockerHostToOAS(dh)
	}
	return result, nil
}

func (h *Handler) CreateDockerHost(ctx context.Context, req *oas.DockerHostInput) (*oas.CreateDockerHostCreated, error) {
	userID, err := oasutil.UserIDFromCtx(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting user from context: %w", err)
	}

	dh := &DockerHost{
		UserID:       userID,
		Name:         req.Name,
		DockerType:   string(req.DockerType),
		DockerDaemon: oasutil.OptStringValue(req.DockerDaemon),
	}

	id, createErr := h.repo.Create(ctx, dh)
	if createErr != nil {
		return nil, fmt.Errorf("creating docker host: %w", createErr)
	}

	return &oas.CreateDockerHostCreated{ID: uuid.MustParse(id)}, nil
}

func (h *Handler) UpdateDockerHost(ctx context.Context, req *oas.DockerHostInput, params oas.UpdateDockerHostParams) error {
	dh := &DockerHost{
		ID:           params.DockerHostId.String(),
		Name:         req.Name,
		DockerType:   string(req.DockerType),
		DockerDaemon: oasutil.OptStringValue(req.DockerDaemon),
	}

	if updateErr := h.repo.Update(ctx, dh); updateErr != nil {
		return fmt.Errorf("updating docker host: %w", updateErr)
	}
	return nil
}

func (h *Handler) DeleteDockerHost(ctx context.Context, params oas.DeleteDockerHostParams) error {
	if err := h.repo.Delete(ctx, params.DockerHostId.String()); err != nil {
		return fmt.Errorf("deleting docker host: %w", err)
	}
	return nil
}

func (h *Handler) TestDockerHost(_ context.Context, _ oas.TestDockerHostParams) (*oas.MessageResponse, error) {
	return &oas.MessageResponse{Message: "OK"}, nil
}

func dockerHostToOAS(dh *DockerHost) oas.DockerHost {
	return oas.DockerHost{
		ID:           uuid.MustParse(dh.ID),
		Name:         dh.Name,
		DockerType:   oas.DockerHostDockerType(dh.DockerType),
		DockerDaemon: oasutil.PtrToOptString(dh.DockerDaemon),
	}
}
