package api

import (
	"context"
	"fmt"

	oas "github.com/koblas/besops/internal/api/oas_generated"
	"github.com/koblas/besops/internal/auth"
)

type SecurityHandler struct {
	authProvider *auth.Provider
}

func NewSecurityHandler(authProvider *auth.Provider) *SecurityHandler {
	return &SecurityHandler{authProvider: authProvider}
}

func (s *SecurityHandler) HandleBearerAuth(ctx context.Context, _ oas.OperationName, t oas.BearerAuth) (context.Context, error) {
	userID, err := s.authProvider.ValidateAccessToken(t.Token)
	if err != nil {
		return ctx, fmt.Errorf("validating access token: %w", err)
	}
	return auth.ContextWithUserID(ctx, userID), nil
}

