package notification

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	oas "github.com/koblas/besops/internal/api/oas_generated"
	"github.com/koblas/besops/internal/api/oasutil"
)

var _ oas.NotificationHandler = (*Handler)(nil)

type Handler struct {
	repo Repository
}

func NewHandler(repo Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) ListNotifications(ctx context.Context) ([]oas.Notification, error) {
	userID, err := oasutil.UserIDFromCtx(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting user from context: %w", err)
	}

	notifs, err := h.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("listing notifications: %w", err)
	}

	result := make([]oas.Notification, 0, len(notifs))
	for _, n := range notifs {
		result = append(result, notificationToOAS(n))
	}
	return result, nil
}

func (h *Handler) CreateNotification(ctx context.Context, req *oas.NotificationInput) (*oas.CreateNotificationCreated, error) {
	userID, err := oasutil.UserIDFromCtx(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting user from context: %w", err)
	}

	configJSON := marshalConfigWithType(req.Type, req.Config)

	n := &Notification{
		Name:   req.Name,
		UserID: userID,
		Config: string(configJSON),
		Active: oasutil.OptBoolValue(req.Active, true),
	}

	id, err := h.repo.Create(ctx, n)
	if err != nil {
		return nil, fmt.Errorf("creating notification: %w", err)
	}

	return &oas.CreateNotificationCreated{ID: oasutil.MustParseUUID(id)}, nil
}

func (h *Handler) UpdateNotification(ctx context.Context, req *oas.NotificationInput, params oas.UpdateNotificationParams) (*oas.Notification, error) {
	existing, err := h.repo.FindByID(ctx, params.NotificationId.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("notification not found: %w", err)
		}
		return nil, fmt.Errorf("finding notification: %w", err)
	}

	existing.Name = req.Name
	existing.Config = string(marshalConfigWithType(req.Type, req.Config))
	if req.Active.IsSet() {
		existing.Active = req.Active.Value
	}

	if err := h.repo.Update(ctx, existing); err != nil {
		return nil, fmt.Errorf("updating notification: %w", err)
	}

	result := notificationToOAS(existing)
	return &result, nil
}

func (h *Handler) DeleteNotification(ctx context.Context, params oas.DeleteNotificationParams) error {
	if err := h.repo.Delete(ctx, params.NotificationId.String()); err != nil {
		return fmt.Errorf("deleting notification: %w", err)
	}
	return nil
}

func (h *Handler) TestNotification(ctx context.Context, params oas.TestNotificationParams) (*oas.MessageResponse, error) {
	_, err := h.repo.FindByID(ctx, params.NotificationId.String())
	if err != nil {
		return nil, fmt.Errorf("finding notification: %w", err)
	}
	// TODO: dispatch test notification via notifier
	return &oas.MessageResponse{Message: "test notification sent"}, nil
}

func notificationToOAS(n *Notification) oas.Notification {
	result := oas.Notification{
		ID:     oasutil.MustParseUUID(n.ID),
		Name:   n.Name,
		Active: n.Active,
	}
	if n.Config != "" {
		var raw map[string]json.RawMessage
		if err := json.Unmarshal([]byte(n.Config), &raw); err == nil {
			if t, ok := raw["type"]; ok {
				var typ string
				if json.Unmarshal(t, &typ) == nil {
					result.Type = typ
				}
				delete(raw, "type")
			}
			cfg, _ := json.Marshal(raw)
			result.Config = cfg
		}
	}
	return result
}

func marshalConfigWithType(typ string, config []byte) []byte {
	var raw map[string]json.RawMessage
	if len(config) > 0 {
		_ = json.Unmarshal(config, &raw)
	}
	if raw == nil {
		raw = make(map[string]json.RawMessage)
	}
	typeJSON, _ := json.Marshal(typ)
	raw["type"] = typeJSON
	result, _ := json.Marshal(raw)
	return result
}
