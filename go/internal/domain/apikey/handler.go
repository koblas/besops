package apikey

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	oas "github.com/koblas/besops/internal/api/oas_generated"
	"github.com/koblas/besops/internal/api/oasutil"
)

var _ oas.APIKeyHandler = (*Handler)(nil)

type Handler struct {
	repo Repository
}

func NewHandler(repo Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) ListAPIKeys(ctx context.Context) ([]oas.APIKey, error) {
	userID, err := oasutil.UserIDFromCtx(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting user from context: %w", err)
	}

	keys, err := h.repo.FindAll(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("listing API keys: %w", err)
	}

	result := make([]oas.APIKey, 0, len(keys))
	for _, k := range keys {
		result = append(result, apiKeyToOAS(k))
	}
	return result, nil
}

func (h *Handler) CreateAPIKey(ctx context.Context, req *oas.APIKeyInput) (*oas.CreateAPIKeyCreated, error) {
	userID, err := oasutil.UserIDFromCtx(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting user from context: %w", err)
	}

	keyBytes := make([]byte, 32)
	if _, randErr := rand.Read(keyBytes); randErr != nil {
		return nil, fmt.Errorf("generating key: %w", randErr)
	}
	keyStr := "uk_" + hex.EncodeToString(keyBytes)

	k := &APIKey{
		Key:       keyStr,
		Name:      req.Name,
		UserID:    userID,
		Active:    true,
		CreatedAt: time.Now().UTC(),
	}
	if req.Expires.IsSet() {
		t := req.Expires.Value
		k.Expires = &t
	}

	id, err := h.repo.Create(ctx, k)
	if err != nil {
		return nil, fmt.Errorf("creating API key: %w", err)
	}

	return &oas.CreateAPIKeyCreated{
		ID:  oasutil.MustParseUUID(id),
		Key: keyStr,
	}, nil
}

func (h *Handler) DeleteAPIKey(ctx context.Context, params oas.DeleteAPIKeyParams) error {
	if err := h.repo.Delete(ctx, params.KeyId.String()); err != nil {
		return fmt.Errorf("deleting API key: %w", err)
	}
	return nil
}

func (h *Handler) EnableAPIKey(ctx context.Context, params oas.EnableAPIKeyParams) (*oas.MessageResponse, error) {
	if err := h.repo.SetActive(ctx, params.KeyId.String(), true); err != nil {
		return nil, fmt.Errorf("enabling API key: %w", err)
	}
	return &oas.MessageResponse{Message: "enabled"}, nil
}

func (h *Handler) DisableAPIKey(ctx context.Context, params oas.DisableAPIKeyParams) (*oas.MessageResponse, error) {
	if err := h.repo.SetActive(ctx, params.KeyId.String(), false); err != nil {
		return nil, fmt.Errorf("disabling API key: %w", err)
	}
	return &oas.MessageResponse{Message: "disabled"}, nil
}

func apiKeyToOAS(k *APIKey) oas.APIKey {
	result := oas.APIKey{
		ID:          oasutil.MustParseUUID(k.ID),
		Name:        k.Name,
		Active:      k.Active,
		CreatedDate: oas.NewOptDateTime(k.CreatedAt),
	}
	if k.Expires != nil {
		result.Expires = oas.NewOptDateTime(*k.Expires)
	}
	return result
}
