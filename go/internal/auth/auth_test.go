package auth

import (
	"context"
	"fmt"
	"testing"
)

type mockUserStore struct {
	users map[string]*User
}

func (m *mockUserStore) FindByUsername(_ context.Context, username string) (*User, error) {
	for _, u := range m.users {
		if u.Username == username {
			return u, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockUserStore) FindByID(_ context.Context, id string) (*User, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return u, nil
}

func (m *mockUserStore) Create(_ context.Context, username, passwordHash string) (string, error) {
	id := mustRandomID()
	m.users[id] = &User{ID: id, Username: username, PasswordHash: passwordHash, Active: true}
	return id, nil
}

func (m *mockUserStore) UpdatePassword(_ context.Context, id string, hash string) error {
	if u, ok := m.users[id]; ok {
		u.PasswordHash = hash
	}
	return nil
}

func (m *mockUserStore) Update2FA(_ context.Context, id string, enabled bool, secret string) error {
	if u, ok := m.users[id]; ok {
		u.TOTPEnabled = enabled
		u.TOTPSecret = secret
	}
	return nil
}

func (m *mockUserStore) Count(_ context.Context) (int64, error) {
	return int64(len(m.users)), nil
}

func newTestProvider(t *testing.T) (*Provider, *mockUserStore) {
	t.Helper()
	users := &mockUserStore{users: make(map[string]*User)}
	sessions := NewMemorySessionStore()
	codes := NewMemoryCodeStore()

	p, err := NewProvider("http://localhost:3001", nil, users, sessions, codes)
	if err != nil {
		t.Fatal(err)
	}
	return p, users
}

func TestPasswordHashAndVerify(t *testing.T) {
	hash, err := HashPassword("correcthorsebatterystaple")
	if err != nil {
		t.Fatal(err)
	}

	if !VerifyPassword("correcthorsebatterystaple", hash) {
		t.Error("expected password to verify")
	}

	if VerifyPassword("wrongpassword", hash) {
		t.Error("expected wrong password to fail")
	}
}

func TestPasswordNeedsRehash(t *testing.T) {
	argonHash, _ := HashPassword("test")
	if NeedsRehash(argonHash) {
		t.Error("argon2id hash should not need rehash")
	}

	bcryptHash := "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"
	if !NeedsRehash(bcryptHash) {
		t.Error("bcrypt hash should need rehash")
	}
}

func TestAuthenticateAndIssueTokens(t *testing.T) {
	p, users := newTestProvider(t)
	ctx := t.Context()

	hash, _ := HashPassword("secret123")
	users.users["user-1"] = &User{
		ID:           "user-1",
		Username:     "admin",
		PasswordHash: hash,
		Active:       true,
	}

	result, err := p.Authenticate(ctx, "admin", "secret123")
	if err != nil {
		t.Fatalf("authenticate: %v", err)
	}
	if result.UserID != "user-1" {
		t.Errorf("expected user-1, got %s", result.UserID)
	}
	if result.Requires2FA {
		t.Error("should not require 2FA")
	}

	tokens, err := p.IssueTokens(ctx, "user-1", "admin", "openid profile", "")
	if err != nil {
		t.Fatalf("issue tokens: %v", err)
	}

	if tokens.AccessToken == "" {
		t.Error("empty access token")
	}
	if tokens.IDToken == "" {
		t.Error("empty ID token")
	}
	if tokens.RefreshToken == "" {
		t.Error("empty refresh token")
	}
	if tokens.TokenType != "Bearer" {
		t.Errorf("expected Bearer, got %s", tokens.TokenType)
	}
}

func TestValidateAccessToken(t *testing.T) {
	p, users := newTestProvider(t)
	ctx := t.Context()

	hash, _ := HashPassword("pass")
	users.users["u1"] = &User{ID: "u1", Username: "test", PasswordHash: hash, Active: true}

	tokens, err := p.IssueTokens(ctx, "u1", "test", "openid", "")
	if err != nil {
		t.Fatal(err)
	}

	userID, err := p.ValidateAccessToken(tokens.AccessToken)
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if userID != "u1" {
		t.Errorf("expected u1, got %s", userID)
	}

	_, err = p.ValidateAccessToken("garbage")
	if err == nil {
		t.Error("expected error for invalid token")
	}
}

func TestOIDCCodeFlow(t *testing.T) {
	p, users := newTestProvider(t)
	ctx := t.Context()

	hash, _ := HashPassword("mypass")
	users.users["u1"] = &User{ID: "u1", Username: "admin", PasswordHash: hash, Active: true}

	authReq := &AuthorizeRequest{
		ClientID:            "frontend",
		RedirectURI:         "http://localhost:3000/callback",
		ResponseType:        "code",
		Scope:               "openid profile",
		State:               "random-state",
		Nonce:               "random-nonce",
		CodeChallenge:       "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM",
		CodeChallengeMethod: "S256",
	}

	if err := p.ValidateAuthorizeRequest(authReq); err != nil {
		t.Fatalf("validate: %v", err)
	}

	code, err := p.IssueCode(ctx, "u1", authReq)
	if err != nil {
		t.Fatalf("issue code: %v", err)
	}

	// Exchange code with correct verifier
	// The challenge above is SHA256("dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk") base64url encoded
	tokens, err := p.ExchangeCode(ctx, &TokenRequest{
		GrantType:    "authorization_code",
		Code:         code,
		RedirectURI:  "http://localhost:3000/callback",
		CodeVerifier: "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
	})
	if err != nil {
		t.Fatalf("exchange code: %v", err)
	}

	if tokens.AccessToken == "" || tokens.IDToken == "" {
		t.Error("expected tokens to be non-empty")
	}
}

func TestRefreshToken(t *testing.T) {
	p, users := newTestProvider(t)
	ctx := t.Context()

	hash, _ := HashPassword("pass")
	users.users["u1"] = &User{ID: "u1", Username: "test", PasswordHash: hash, Active: true}

	tokens, err := p.IssueTokens(ctx, "u1", "test", "openid", "")
	if err != nil {
		t.Fatal(err)
	}

	refreshed, err := p.RefreshAccessToken(ctx, tokens.RefreshToken)
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}

	if refreshed.AccessToken == "" {
		t.Error("expected new access token")
	}

	// Revoke and verify it's invalid
	_ = p.RevokeSession(ctx, tokens.RefreshToken)

	_, err = p.RefreshAccessToken(ctx, tokens.RefreshToken)
	if err == nil {
		t.Error("expected error after revocation")
	}
}

func TestDiscovery(t *testing.T) {
	p, _ := newTestProvider(t)
	disc := p.Discovery()

	if disc.Issuer != "http://localhost:3001" {
		t.Errorf("unexpected issuer: %s", disc.Issuer)
	}
	if disc.TokenEndpoint != "http://localhost:3001/token" {
		t.Errorf("unexpected token endpoint: %s", disc.TokenEndpoint)
	}
}

func TestJWKS(t *testing.T) {
	p, _ := newTestProvider(t)
	jwks := p.JWKS()

	if len(jwks.Keys) != 1 {
		t.Fatalf("expected 1 key, got %d", len(jwks.Keys))
	}
	key := jwks.Keys[0]
	if key.Kty != "EC" || key.Crv != "P-256" || key.Alg != "ES256" {
		t.Errorf("unexpected key params: %+v", key)
	}
	if key.X == "" || key.Y == "" {
		t.Error("expected non-empty X and Y coordinates")
	}
}

func TestDeriveSigningKeyDeterministic(t *testing.T) {
	key1, err := DeriveSigningKey("my-stable-secret")
	if err != nil {
		t.Fatal(err)
	}
	key2, err := DeriveSigningKey("my-stable-secret")
	if err != nil {
		t.Fatal(err)
	}

	if !key1.Equal(key2) {
		t.Error("same secret should produce identical keys")
	}
}

func TestDeriveSigningKeyDifferentSecrets(t *testing.T) {
	key1, err := DeriveSigningKey("secret-a")
	if err != nil {
		t.Fatal(err)
	}
	key2, err := DeriveSigningKey("secret-b")
	if err != nil {
		t.Fatal(err)
	}

	if key1.Equal(key2) {
		t.Error("different secrets should produce different keys")
	}
}
