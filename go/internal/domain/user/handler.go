package user

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sync"

	oas "github.com/koblas/besops/internal/api/oas_generated"
	"github.com/koblas/besops/internal/api/oasutil"
	"github.com/koblas/besops/internal/auth"
	"github.com/koblas/besops/lib/errs"
)

var _ oas.AuthHandler = (*Handler)(nil)

type Handler struct {
	repo        Repository
	auth        *auth.Provider
	bootstrap   bool
	pendingTOTP sync.Map
}

func NewHandler(repo Repository, authProvider *auth.Provider, bootstrap bool) *Handler {
	return &Handler{
		repo:      repo,
		auth:      authProvider,
		bootstrap: bootstrap,
	}
}

func (h *Handler) NeedSetup(ctx context.Context) (oas.NeedSetupRes, error) {
	if h.bootstrap {
		return &oas.NeedSetupOK{NeedSetup: false}, nil
	}
	count, err := h.repo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("counting users: %w", err)
	}
	return &oas.NeedSetupOK{NeedSetup: count == 0}, nil
}

func (h *Handler) Setup(ctx context.Context, req *oas.SetupReq) (oas.SetupRes, error) {
	count, err := h.repo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("counting users: %w", err)
	}
	if count > 0 {
		return &oas.SetupConflict{}, nil
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}

	userID, err := h.auth.CreateUser(ctx, req.Username, hash)
	if err != nil {
		return nil, fmt.Errorf("creating user: %w", err)
	}

	tokens, err := h.auth.IssueTokens(ctx, userID, req.Username, "openid profile", "")
	if err != nil {
		return nil, fmt.Errorf("issuing tokens: %w", err)
	}

	return &oas.MessageResponse{Message: tokens.AccessToken}, nil
}

func (h *Handler) Login(ctx context.Context, req *oas.LoginRequest) (oas.LoginRes, error) {
	result, err := h.auth.Authenticate(ctx, req.Username, req.Password)
	if errors.Is(err, errs.ErrNotFound) && h.bootstrap {
		result, err = h.bootstrapUser(ctx, req.Username, req.Password)
	}
	if err != nil {
		return nil, errs.NewUnauthenticated(err, "invalid username/password")
	}

	if result.Requires2FA {
		if !req.Token.IsSet() || req.Token.Value == "" {
			return &oas.LoginResponse{
				TokenRequired: oas.NewOptBool(true),
			}, nil
		}
		if verifyErr := h.auth.Verify2FA(ctx, result.UserID, req.Token.Value); verifyErr != nil {
			return &oas.LoginUnauthorized{Error: "invalid 2FA token"}, nil //nolint:nilerr
		}
	}

	tokens, err := h.auth.IssueTokens(ctx, result.UserID, req.Username, "openid profile", "")
	if err != nil {
		return nil, fmt.Errorf("issuing tokens: %w", err)
	}

	return &oas.LoginResponse{
		Token:        tokens.AccessToken,
		RefreshToken: oas.NewOptString(tokens.RefreshToken),
	}, nil
}

func (h *Handler) bootstrapUser(ctx context.Context, username, password string) (*auth.LoginResult, error) {
	hash, err := auth.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}

	if _, createErr := h.auth.CreateUser(ctx, username, hash); createErr != nil {
		return nil, fmt.Errorf("creating bootstrap user: %w", createErr)
	}

	result, err := h.auth.Authenticate(ctx, username, password)
	if err != nil {
		return nil, fmt.Errorf("authenticating bootstrap user: %w", err)
	}
	return result, nil
}

func (h *Handler) Logout(_ context.Context) (oas.LogoutRes, error) {
	return &oas.LogoutNoContent{}, nil
}

func (h *Handler) RefreshToken(ctx context.Context, req *oas.RefreshTokenRequest) (oas.RefreshTokenRes, error) {
	tokens, err := h.auth.RefreshAccessToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, errs.NewUnauthenticated(err, "invalid or expired refresh token")
	}

	return &oas.TokenResponse{Token: tokens.AccessToken}, nil
}

func (h *Handler) ChangePassword(ctx context.Context, req *oas.ChangePasswordReq) (oas.ChangePasswordRes, error) {
	userID, err := oasutil.UserIDFromCtx(ctx)
	if err != nil {
		return nil, errs.NewUnauthenticated(err, "unauthorized")
	}

	user, findErr := h.repo.FindByID(ctx, userID)
	if findErr != nil {
		return nil, errs.NewNotFound(findErr, "user not found")
	}

	if !auth.VerifyPassword(req.CurrentPassword, user.Password) {
		return nil, errs.NewUnauthenticated(nil, "current password incorrect")
	}

	newHash, hashErr := auth.HashPassword(req.NewPassword)
	if hashErr != nil {
		return nil, fmt.Errorf("hashing password: %w", hashErr)
	}

	if updateErr := h.repo.UpdatePassword(ctx, userID, newHash); updateErr != nil {
		return nil, fmt.Errorf("updating password: %w", updateErr)
	}

	tokens, tokenErr := h.auth.IssueTokens(ctx, userID, user.Username, "openid profile", "")
	if tokenErr != nil {
		return nil, fmt.Errorf("issuing tokens: %w", tokenErr)
	}

	return &oas.TokenResponse{Token: tokens.AccessToken}, nil
}

func (h *Handler) Get2FAStatus(ctx context.Context) (oas.Get2FAStatusRes, error) {
	userID, err := oasutil.UserIDFromCtx(ctx)
	if err != nil {
		return nil, errs.NewUnauthenticated(err, "unauthorized")
	}

	user, err := h.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, errs.NewNotFound(err, "user not found")
	}

	return &oas.Get2FAStatusOK{Enabled: user.TwoFAStatus}, nil
}

func (h *Handler) Prepare2FA(ctx context.Context, req *oas.Prepare2FAReq) (oas.Prepare2FARes, error) {
	userID, err := oasutil.UserIDFromCtx(ctx)
	if err != nil {
		return nil, errs.NewUnauthenticated(err, "unauthorized")
	}

	user, err := h.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, errs.NewNotFound(err, "user not found")
	}

	if !auth.VerifyPassword(req.CurrentPassword, user.Password) {
		return nil, errs.NewUnauthenticated(nil, "current password incorrect")
	}

	secret, uri, genErr := h.auth.PrepareTOTP(user.Username)
	if genErr != nil {
		return nil, fmt.Errorf("preparing TOTP: %w", genErr)
	}

	h.pendingTOTP.Store(userID, secret)

	parsedURI, parseErr := url.Parse(uri)
	if parseErr != nil {
		return nil, fmt.Errorf("parsing TOTP URI: %w", parseErr)
	}

	return &oas.Prepare2FAOK{URI: *parsedURI}, nil
}

func (h *Handler) Enable2FA(ctx context.Context, req *oas.Enable2FAReq) (oas.Enable2FARes, error) {
	userID, err := oasutil.UserIDFromCtx(ctx)
	if err != nil {
		return nil, errs.NewUnauthenticated(err, "unauthorized")
	}

	user, err := h.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, errs.NewNotFound(err, "user not found")
	}

	if !auth.VerifyPassword(req.CurrentPassword, user.Password) {
		return nil, errs.NewUnauthenticated(nil, "current password incorrect")
	}

	secretVal, ok := h.pendingTOTP.Load(userID)
	if !ok {
		return nil, fmt.Errorf("no pending 2FA setup — call prepare2FA first")
	}
	secret := secretVal.(string) //nolint:forcetypeassert // stored by Prepare2FA

	if enableErr := h.auth.EnableTOTP(ctx, userID, secret, req.Token); enableErr != nil {
		return nil, fmt.Errorf("enabling 2FA: %w", enableErr)
	}

	h.pendingTOTP.Delete(userID)

	return &oas.MessageResponse{Message: "2FA enabled"}, nil
}

func (h *Handler) Disable2FA(ctx context.Context, req *oas.Disable2FAReq) (oas.Disable2FARes, error) {
	userID, err := oasutil.UserIDFromCtx(ctx)
	if err != nil {
		return nil, errs.NewUnauthenticated(err, "unauthorized")
	}

	user, err := h.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, errs.NewNotFound(err, "user not found")
	}

	if !auth.VerifyPassword(req.CurrentPassword, user.Password) {
		return nil, errs.NewUnauthenticated(nil, "current password incorrect")
	}

	if disableErr := h.auth.DisableTOTP(ctx, userID); disableErr != nil {
		return nil, fmt.Errorf("disabling 2FA: %w", disableErr)
	}

	return &oas.MessageResponse{Message: "2FA disabled"}, nil
}
