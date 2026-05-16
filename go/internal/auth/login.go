package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/koblas/besops/lib/errs"
)

// LoginResult describes the outcome of a login attempt.
type LoginResult struct {
	UserID      string
	Requires2FA bool
}

// Authenticate verifies username/password and returns the user identity.
// If the user has 2FA enabled, the caller must subsequently call Verify2FA.
func (p *Provider) Authenticate(ctx context.Context, username, password string) (*LoginResult, error) {
	user, err := p.users.FindByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return nil, errs.NewNotFound(err, "user not found")
		}
		return nil, errs.NewUnauthenticated(fmt.Errorf("user lookup failed: %w", err), "")
	}

	if !user.Active {
		return nil, fmt.Errorf("account disabled")
	}

	if !VerifyPassword(password, user.PasswordHash) {
		slog.InfoContext(ctx, "password mis-match")
		return nil, fmt.Errorf("invalid credentials")
	}

	if NeedsRehash(user.PasswordHash) {
		if newHash, err := HashPassword(password); err == nil {
			_ = p.users.UpdatePassword(ctx, user.ID, newHash)
		}
	}

	return &LoginResult{
		UserID:      user.ID,
		Requires2FA: user.TOTPEnabled,
	}, nil
}

// Verify2FA validates a TOTP token for a user during login.
func (p *Provider) Verify2FA(ctx context.Context, userID, token string) error {
	user, err := p.users.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	if !user.TOTPEnabled || user.TOTPSecret == "" {
		return fmt.Errorf("2FA not enabled")
	}

	if !VerifyTOTP(user.TOTPSecret, token) {
		return fmt.Errorf("invalid 2FA token")
	}

	return nil
}
