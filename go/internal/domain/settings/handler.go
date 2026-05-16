package settings

import (
	"context"
	"fmt"
	"strconv"

	oas "github.com/koblas/besops/internal/api/oas_generated"
)

var _ oas.SettingsHandler = (*Handler)(nil)

type Handler struct {
	repo Repository
}

func NewHandler(repo Repository) *Handler {
	return &Handler{repo: repo}
}

// GetSettings returns all application settings.
func (h *Handler) GetSettings(ctx context.Context) (*oas.Settings, error) {
	all, err := h.repo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting settings: %w", err)
	}

	result := &oas.Settings{}
	if v, ok := all["primaryBaseURL"]; ok {
		result.PrimaryBaseURL = oas.NewOptString(v)
	}
	if v, ok := all["serverTimezone"]; ok {
		result.ServerTimezone = oas.NewOptString(v)
	}
	if v, ok := all["entryPage"]; ok {
		result.EntryPage = oas.OptSettingsEntryPage{Value: oas.SettingsEntryPage(v), Set: true}
	}
	if v, ok := all["trustProxy"]; ok {
		result.TrustProxy = oas.NewOptBool(v == "true" || v == "1")
	}
	if v, ok := all["disableAuth"]; ok {
		result.DisableAuth = oas.NewOptBool(v == "true" || v == "1")
	}
	if v, ok := all["apiKeysEnabled"]; ok {
		result.ApiKeysEnabled = oas.NewOptBool(v == "true" || v == "1")
	}
	if v, ok := all["keepDataPeriodDays"]; ok {
		if days, parseErr := strconv.Atoi(v); parseErr == nil {
			result.KeepDataPeriodDays = oas.NewOptInt(days)
		}
	}
	if v, ok := all["statusPageSlug"]; ok {
		result.StatusPageSlug = oas.NewOptString(v)
	}
	return result, nil
}

// UpdateSettings persists the provided settings values.
func (h *Handler) UpdateSettings(ctx context.Context, req *oas.Settings) (*oas.MessageResponse, error) {
	if req.PrimaryBaseURL.IsSet() {
		if err := h.repo.Set(ctx, "primaryBaseURL", req.PrimaryBaseURL.Value); err != nil {
			return nil, fmt.Errorf("setting primaryBaseURL: %w", err)
		}
	}
	if req.ServerTimezone.IsSet() {
		if err := h.repo.Set(ctx, "serverTimezone", req.ServerTimezone.Value); err != nil {
			return nil, fmt.Errorf("setting serverTimezone: %w", err)
		}
	}
	if req.EntryPage.IsSet() {
		if err := h.repo.Set(ctx, "entryPage", string(req.EntryPage.Value)); err != nil {
			return nil, fmt.Errorf("setting entryPage: %w", err)
		}
	}
	if req.TrustProxy.IsSet() {
		if err := h.repo.Set(ctx, "trustProxy", strconv.FormatBool(req.TrustProxy.Value)); err != nil {
			return nil, fmt.Errorf("setting trustProxy: %w", err)
		}
	}
	if req.DisableAuth.IsSet() {
		if err := h.repo.Set(ctx, "disableAuth", strconv.FormatBool(req.DisableAuth.Value)); err != nil {
			return nil, fmt.Errorf("setting disableAuth: %w", err)
		}
	}
	if req.ApiKeysEnabled.IsSet() {
		if err := h.repo.Set(ctx, "apiKeysEnabled", strconv.FormatBool(req.ApiKeysEnabled.Value)); err != nil {
			return nil, fmt.Errorf("setting apiKeysEnabled: %w", err)
		}
	}
	if req.KeepDataPeriodDays.IsSet() {
		if err := h.repo.Set(ctx, "keepDataPeriodDays", strconv.Itoa(req.KeepDataPeriodDays.Value)); err != nil {
			return nil, fmt.Errorf("setting keepDataPeriodDays: %w", err)
		}
	}
	if req.StatusPageSlug.IsSet() {
		if err := h.repo.Set(ctx, "statusPageSlug", req.StatusPageSlug.Value); err != nil {
			return nil, fmt.Errorf("setting statusPageSlug: %w", err)
		}
	}
	return &oas.MessageResponse{Message: "settings updated"}, nil
}
