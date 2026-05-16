// Package auth implements an OpenID Connect Provider with a built-in
// username/password identity provider. The application acts as its own
// OIDC issuer — the frontend is a standard OIDC Relying Party using
// Authorization Code + PKCE.
package auth

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"math/big"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/hkdf"
)

// Provider is the OIDC provider that issues tokens and validates credentials.
type Provider struct {
	issuer     string
	signingKey *ecdsa.PrivateKey
	keyID      string
	users      UserStore
	sessions   SessionStore
	codes      CodeStore
}

// UserStore is the interface for looking up and verifying user credentials.
type UserStore interface {
	FindByUsername(ctx context.Context, username string) (*User, error)
	FindByID(ctx context.Context, id string) (*User, error)
	Create(ctx context.Context, username, passwordHash string) (string, error)
	UpdatePassword(ctx context.Context, id string, passwordHash string) error
	Update2FA(ctx context.Context, id string, enabled bool, secret string) error
	Count(ctx context.Context) (int64, error)
}

// SessionStore manages refresh token / session persistence.
type SessionStore interface {
	Create(ctx context.Context, session *Session) error
	FindByToken(ctx context.Context, token string) (*Session, error)
	Revoke(ctx context.Context, token string) error
	RevokeAllForUser(ctx context.Context, userID string) error
}

// CodeStore manages short-lived authorization codes for the OIDC flow.
type CodeStore interface {
	Store(ctx context.Context, code *Code) error
	Consume(ctx context.Context, code string) (*Code, error)
}

// User represents a user identity for authentication.
type User struct {
	ID           string
	Username     string
	PasswordHash string
	Active       bool
	TOTPSecret   string
	TOTPEnabled  bool
}

// Session represents an active refresh token.
type Session struct {
	ID        string
	UserID    string
	Token     string
	ExpiresAt time.Time
	CreatedAt time.Time
}

// Code represents an authorization code issued during the OIDC flow.
type Code struct {
	Code                string
	UserID              string
	ClientID            string
	RedirectURI         string
	Scope               string
	CodeChallenge       string
	CodeChallengeMethod string
	ExpiresAt           time.Time
	Nonce               string
}

// DeriveSigningKey deterministically derives an ECDSA P-256 key from a secret string
// using HKDF-SHA256. The same secret always produces the same key, so tokens survive restarts.
func DeriveSigningKey(secret string) (*ecdsa.PrivateKey, error) {
	hkdfReader := hkdf.New(sha256.New, []byte(secret), []byte("rupert-jwt-salt"), []byte("ecdsa-p256-signing-key"))

	// Read 32 bytes of key material and use as the private scalar d.
	// Reduce modulo (N-1) and add 1 to guarantee 1 <= d < N.
	buf := make([]byte, 32)
	if _, err := io.ReadFull(hkdfReader, buf); err != nil {
		return nil, fmt.Errorf("reading HKDF output: %w", err)
	}

	curve := elliptic.P256()
	n := curve.Params().N
	nMinusOne := new(big.Int).Sub(n, big.NewInt(1))

	d := new(big.Int).SetBytes(buf)
	d.Mod(d, nMinusOne)
	d.Add(d, big.NewInt(1))

	priv := &ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{Curve: curve},
		D:        d,
	}
	priv.PublicKey.X, priv.PublicKey.Y = curve.ScalarBaseMult(d.Bytes())

	return priv, nil
}

// NewProvider creates an OIDC provider. If signingKey is nil, a new ECDSA P-256
// key is generated (suitable for single-instance deployments).
func NewProvider(issuer string, signingKey *ecdsa.PrivateKey, users UserStore, sessions SessionStore, codes CodeStore) (*Provider, error) {
	if signingKey == nil {
		var err error
		signingKey, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return nil, fmt.Errorf("generating ECDSA signing key: %w", err)
		}
	}

	return &Provider{
		issuer:     issuer,
		signingKey: signingKey,
		keyID:      "rupert-1",
		users:      users,
		sessions:   sessions,
		codes:      codes,
	}, nil
}

// CreateUser registers a new user via the underlying UserStore.
func (p *Provider) CreateUser(ctx context.Context, username, passwordHash string) (string, error) {
	id, err := p.users.Create(ctx, username, passwordHash)
	if err != nil {
		return "", fmt.Errorf("creating user: %w", err)
	}
	return id, nil
}

// PrepareTOTP generates a new TOTP secret and returns the otpauth URI for QR display.
func (p *Provider) PrepareTOTP(username string) (secret, uri string, err error) {
	secret, err = GenerateTOTPSecret()
	if err != nil {
		return "", "", fmt.Errorf("generating TOTP secret: %w", err)
	}
	uri = TOTPKeyURI(secret, username)
	return secret, uri, nil
}

// EnableTOTP verifies the token against the secret, then persists 2FA as enabled.
func (p *Provider) EnableTOTP(ctx context.Context, userID, secret, token string) error {
	if !VerifyTOTP(secret, token) {
		return fmt.Errorf("invalid TOTP token")
	}
	if err := p.users.Update2FA(ctx, userID, true, secret); err != nil {
		return fmt.Errorf("enabling 2FA: %w", err)
	}
	return nil
}

// DisableTOTP removes TOTP from the user account.
func (p *Provider) DisableTOTP(ctx context.Context, userID string) error {
	if err := p.users.Update2FA(ctx, userID, false, ""); err != nil {
		return fmt.Errorf("disabling 2FA: %w", err)
	}
	return nil
}

// SigningKey returns the provider's ECDSA private key.
func (p *Provider) SigningKey() *ecdsa.PrivateKey {
	return p.signingKey
}

// PublicKey returns the ECDSA public key for token verification.
func (p *Provider) PublicKey() *ecdsa.PublicKey {
	return &p.signingKey.PublicKey
}

// KeyID returns the key identifier used in JWT headers and JWKS.
func (p *Provider) KeyID() string {
	return p.keyID
}

// Issuer returns the OIDC issuer URL.
func (p *Provider) Issuer() string {
	return p.issuer
}

// IDTokenClaims are the claims included in an OIDC ID token.
type IDTokenClaims struct {
	jwt.RegisteredClaims
	Nonce             string `json:"nonce,omitempty"`
	PreferredUsername string `json:"preferred_username,omitempty"`
}

// AccessTokenClaims are the claims included in an access token.
type AccessTokenClaims struct {
	jwt.RegisteredClaims
	Scope string `json:"scope,omitempty"`
}
