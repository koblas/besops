package tag

import (
	"context"
	"fmt"

	oas "github.com/koblas/besops/internal/api/oas_generated"
	"github.com/koblas/besops/internal/api/oasutil"
)

var _ oas.TagHandler = (*Handler)(nil)

type Handler struct {
	repo Repository
}

func NewHandler(repo Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) ListTags(ctx context.Context) (oas.ListTagsRes, error) {
	tags, err := h.repo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing tags: %w", err)
	}
	result := make(oas.ListTagsOKApplicationJSON, 0, len(tags))
	for _, t := range tags {
		result = append(result, toOAS(t))
	}
	return &result, nil
}

func (h *Handler) CreateTag(ctx context.Context, req *oas.TagInput) (oas.CreateTagRes, error) {
	t := &Tag{
		Name:  req.Name,
		Color: req.Color,
	}
	id, err := h.repo.Create(ctx, t)
	if err != nil {
		return nil, fmt.Errorf("creating tag: %w", err)
	}
	return &oas.Tag{
		ID:    oasutil.MustParseUUID(id),
		Name:  req.Name,
		Color: req.Color,
	}, nil
}

func (h *Handler) UpdateTag(ctx context.Context, req *oas.TagInput, params oas.UpdateTagParams) (oas.UpdateTagRes, error) {
	t := &Tag{
		ID:    params.TagId.String(),
		Name:  req.Name,
		Color: req.Color,
	}
	if err := h.repo.Update(ctx, t); err != nil {
		return nil, fmt.Errorf("updating tag: %w", err)
	}
	return &oas.Tag{
		ID:    params.TagId,
		Name:  req.Name,
		Color: req.Color,
	}, nil
}

func (h *Handler) DeleteTag(ctx context.Context, params oas.DeleteTagParams) (oas.DeleteTagRes, error) {
	if err := h.repo.Delete(ctx, params.TagId.String()); err != nil {
		return nil, fmt.Errorf("deleting tag: %w", err)
	}
	return &oas.DeleteTagNoContent{}, nil
}

func (h *Handler) AddMonitorTag(ctx context.Context, req *oas.AddMonitorTagReq, params oas.AddMonitorTagParams) (oas.AddMonitorTagRes, error) {
	value := ""
	if req.Value.IsSet() {
		value = req.Value.Value
	}
	if err := h.repo.AddToMonitor(ctx, params.MonitorId.String(), req.TagId.String(), value); err != nil {
		return nil, fmt.Errorf("adding tag to monitor: %w", err)
	}
	return &oas.AddMonitorTagCreated{}, nil
}

func (h *Handler) DeleteMonitorTag(ctx context.Context, params oas.DeleteMonitorTagParams) (oas.DeleteMonitorTagRes, error) {
	if err := h.repo.RemoveFromMonitor(ctx, params.MonitorId.String(), params.TagId.String()); err != nil {
		return nil, fmt.Errorf("removing tag from monitor: %w", err)
	}
	return &oas.DeleteMonitorTagNoContent{}, nil
}

func (h *Handler) UpdateMonitorTag(ctx context.Context, req *oas.UpdateMonitorTagReq, params oas.UpdateMonitorTagParams) (oas.UpdateMonitorTagRes, error) {
	value := ""
	if req.Value.IsSet() {
		value = req.Value.Value
	}
	if err := h.repo.RemoveFromMonitor(ctx, params.MonitorId.String(), params.TagId.String()); err != nil {
		return nil, fmt.Errorf("removing existing tag: %w", err)
	}
	if err := h.repo.AddToMonitor(ctx, params.MonitorId.String(), params.TagId.String(), value); err != nil {
		return nil, fmt.Errorf("adding updated tag: %w", err)
	}
	return &oas.UpdateMonitorTagOK{}, nil
}

func toOAS(t *Tag) oas.Tag {
	return oas.Tag{
		ID:    oasutil.MustParseUUID(t.ID),
		Name:  t.Name,
		Color: t.Color,
	}
}
