package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/koblas/besops/lib/errs"
)

const (
	accessTokenTTL  = 15 * time.Minute
	idTokenTTL      = 1 * time.Hour
	refreshTokenTTL = 7 * 24 * time.Hour
)

// TokenSet is the complete set of tokens returned by the token endpoint.
type TokenSet struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
}

// IssueTokens creates a full token set for an authenticated user.
func (p *Provider) IssueTokens(ctx context.Context, userID, username, scope, nonce string) (*TokenSet, error) {
	now := time.Now()

	accessToken, err := p.issueAccessToken(userID, scope, now)
	if err != nil {
		return nil, fmt.Errorf("issue access token: %w", err)
	}

	idToken, err := p.issueIDToken(userID, username, nonce, now)
	if err != nil {
		return nil, fmt.Errorf("issue id token: %w", err)
	}

	refreshToken, err := generateOpaqueToken()
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	session := &Session{
		ID:        mustRandomID(),
		UserID:    userID,
		Token:     refreshToken,
		ExpiresAt: now.Add(refreshTokenTTL),
		CreatedAt: now,
	}
	if err := p.sessions.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("store refresh token: %w", err)
	}

	return &TokenSet{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(accessTokenTTL.Seconds()),
		RefreshToken: refreshToken,
		IDToken:      idToken,
	}, nil
}

// RefreshAccessToken issues a new access token from a valid refresh token.
func (p *Provider) RefreshAccessToken(ctx context.Context, refreshToken string) (*TokenSet, error) {
	session, err := p.sessions.FindByToken(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	if time.Now().After(session.ExpiresAt) {
		_ = p.sessions.Revoke(ctx, refreshToken)
		return nil, fmt.Errorf("refresh token expired")
	}

	user, err := p.users.FindByID(ctx, session.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	if !user.Active {
		return nil, fmt.Errorf("user account disabled")
	}

	now := time.Now()
	accessToken, err := p.issueAccessToken(session.UserID, "openid profile", now)
	if err != nil {
		return nil, fmt.Errorf("issuing refreshed access token: %w", err)
	}

	return &TokenSet{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int(accessTokenTTL.Seconds()),
	}, nil
}

// RevokeSession invalidates a refresh token.
func (p *Provider) RevokeSession(ctx context.Context, refreshToken string) error {
	if err := p.sessions.Revoke(ctx, refreshToken); err != nil {
		return fmt.Errorf("revoking session: %w", err)
	}
	return nil
}

// ValidateAccessToken parses and validates an access token, returning the subject (user ID).
func (p *Provider) ValidateAccessToken(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AccessTokenClaims{}, func(t *jwt.Token) (any, error) {
		if t.Method.Alg() != jwt.SigningMethodES256.Alg() {
			return nil, fmt.Errorf("unexpected signing method: %s", t.Method.Alg())
		}
		return p.PublicKey(), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return "", errs.NewUnauthenticated(err, "token expired")
		}

		return "", errs.NewUnauthenticated(err, "invalid token")
	}
	if !token.Valid {
		return "", errs.NewUnauthenticated(nil, "invalid token claims")
	}

	claims, ok := token.Claims.(*AccessTokenClaims)
	if !ok {
		return "", errs.NewInternal(errors.New("unable to cast AccessTokenClaims"), "")
	}

	sub, err := claims.GetSubject()
	if err != nil || sub == "" {
		// err is always nil RegisteredClaims, but testing for it anyway
		return "", errs.NewUnauthenticated(err, "invalid token subject")
	}
	return sub, nil
}

func (p *Provider) issueAccessToken(userID, scope string, now time.Time) (string, error) {
	claims := AccessTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    p.issuer,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(accessTokenTTL)),
		},
		Scope: scope,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = p.keyID

	signed, err := token.SignedString(p.signingKey)
	if err != nil {
		return "", fmt.Errorf("signing access token: %w", err)
	}
	return signed, nil
}

func (p *Provider) issueIDToken(userID, username, nonce string, now time.Time) (string, error) {
	claims := IDTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    p.issuer,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(idTokenTTL)),
		},
		Nonce:             nonce,
		PreferredUsername: username,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = p.keyID

	signed, err := token.SignedString(p.signingKey)
	if err != nil {
		return "", fmt.Errorf("signing ID token: %w", err)
	}
	return signed, nil
}

func generateOpaqueToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating random bytes: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func mustRandomID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}
