package auth

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"
)

// OIDCDiscovery is the OpenID Connect discovery document.
type OIDCDiscovery struct {
	Issuer                            string   `json:"issuer"`
	AuthorizationEndpoint             string   `json:"authorization_endpoint"`
	TokenEndpoint                     string   `json:"token_endpoint"`
	UserInfoEndpoint                  string   `json:"userinfo_endpoint"`
	JwksURI                           string   `json:"jwks_uri"`
	ScopesSupported                   []string `json:"scopes_supported"`
	ResponseTypesSupported            []string `json:"response_types_supported"`
	GrantTypesSupported               []string `json:"grant_types_supported"`
	SubjectTypesSupported             []string `json:"subject_types_supported"`
	IDTokenSigningAlgValuesSupported  []string `json:"id_token_signing_alg_values_supported"`
	CodeChallengeMethodsSupported     []string `json:"code_challenge_methods_supported"`
	TokenEndpointAuthMethodsSupported []string `json:"token_endpoint_auth_methods_supported"`
}

// Discovery returns the OIDC discovery document for this provider.
func (p *Provider) Discovery() OIDCDiscovery {
	return OIDCDiscovery{
		Issuer:                            p.issuer,
		AuthorizationEndpoint:             p.issuer + "/authorize",
		TokenEndpoint:                     p.issuer + "/token",
		UserInfoEndpoint:                  p.issuer + "/userinfo",
		JwksURI:                           p.issuer + "/.well-known/jwks.json",
		ScopesSupported:                   []string{"openid", "profile"},
		ResponseTypesSupported:            []string{"code"},
		GrantTypesSupported:               []string{"authorization_code", "refresh_token"},
		SubjectTypesSupported:             []string{"public"},
		IDTokenSigningAlgValuesSupported:  []string{"ES256"},
		CodeChallengeMethodsSupported:     []string{"S256"},
		TokenEndpointAuthMethodsSupported: []string{"none"},
	}
}

// JWKS represents a JSON Web Key Set.
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// JWK represents a JSON Web Key (EC P-256 public key).
type JWK struct {
	Kty string `json:"kty"`
	Use string `json:"use"`
	Crv string `json:"crv"`
	Kid string `json:"kid"`
	Alg string `json:"alg"`
	X   string `json:"x"`
	Y   string `json:"y"`
}

// JWKS returns the JSON Web Key Set containing the provider's public key.
func (p *Provider) JWKS() JWKS {
	pub := p.PublicKey()
	ecdhKey, err := pub.ECDH()
	if err != nil {
		return JWKS{}
	}
	// ECDH Bytes() returns the uncompressed point without the 0x04 prefix
	raw := ecdhKey.Bytes()
	coordLen := len(raw) / 2
	x := raw[:coordLen]
	y := raw[coordLen:]

	return JWKS{
		Keys: []JWK{
			{
				Kty: "EC",
				Use: "sig",
				Crv: "P-256",
				Kid: p.keyID,
				Alg: "ES256",
				X:   base64.RawURLEncoding.EncodeToString(x),
				Y:   base64.RawURLEncoding.EncodeToString(y),
			},
		},
	}
}

// AuthorizeRequest represents a parsed authorization request.
type AuthorizeRequest struct {
	ClientID            string
	RedirectURI         string
	ResponseType        string
	Scope               string
	State               string
	Nonce               string
	CodeChallenge       string
	CodeChallengeMethod string
}

// ValidateAuthorizeRequest checks that an authorization request is well-formed.
func (p *Provider) ValidateAuthorizeRequest(req *AuthorizeRequest) error {
	if req.ResponseType != "code" {
		return fmt.Errorf("unsupported response_type: %s", req.ResponseType)
	}
	if req.CodeChallenge == "" {
		return fmt.Errorf("code_challenge is required (PKCE)")
	}
	if req.CodeChallengeMethod != "S256" {
		return fmt.Errorf("only S256 code_challenge_method is supported")
	}
	if req.RedirectURI == "" {
		return fmt.Errorf("redirect_uri is required")
	}
	return nil
}

// IssueCode creates an authorization code after the user has authenticated.
func (p *Provider) IssueCode(ctx context.Context, userID string, req *AuthorizeRequest) (string, error) {
	code, err := generateOpaqueToken()
	if err != nil {
		return "", fmt.Errorf("generating authorization code: %w", err)
	}

	authCode := &Code{
		Code:                code,
		UserID:              userID,
		ClientID:            req.ClientID,
		RedirectURI:         req.RedirectURI,
		Scope:               req.Scope,
		CodeChallenge:       req.CodeChallenge,
		CodeChallengeMethod: req.CodeChallengeMethod,
		ExpiresAt:           time.Now().Add(5 * time.Minute),
		Nonce:               req.Nonce,
	}

	if err := p.codes.Store(ctx, authCode); err != nil {
		return "", fmt.Errorf("storing authorization code: %w", err)
	}

	return code, nil
}

// TokenRequest represents a parsed token endpoint request.
type TokenRequest struct {
	GrantType    string
	Code         string
	RedirectURI  string
	CodeVerifier string
	RefreshToken string
	ClientID     string
}

// ExchangeCode exchanges an authorization code for tokens.
func (p *Provider) ExchangeCode(ctx context.Context, req *TokenRequest) (*TokenSet, error) {
	authCode, err := p.codes.Consume(ctx, req.Code)
	if err != nil {
		return nil, fmt.Errorf("invalid authorization code: %w", err)
	}

	if time.Now().After(authCode.ExpiresAt) {
		return nil, fmt.Errorf("authorization code expired")
	}

	if authCode.RedirectURI != req.RedirectURI {
		return nil, fmt.Errorf("redirect_uri mismatch")
	}

	if !verifyCodeChallenge(authCode.CodeChallenge, req.CodeVerifier) {
		return nil, fmt.Errorf("PKCE verification failed")
	}

	user, err := p.users.FindByID(ctx, authCode.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return p.IssueTokens(ctx, user.ID, user.Username, authCode.Scope, authCode.Nonce)
}

// HandleTokenRequest processes a token endpoint request (code exchange or refresh).
func (p *Provider) HandleTokenRequest(ctx context.Context, req *TokenRequest) (*TokenSet, error) {
	switch req.GrantType {
	case "authorization_code":
		return p.ExchangeCode(ctx, req)
	case "refresh_token":
		return p.RefreshAccessToken(ctx, req.RefreshToken)
	default:
		return nil, fmt.Errorf("unsupported grant_type: %s", req.GrantType)
	}
}

// UserInfo represents the claims returned by the userinfo endpoint.
type UserInfo struct {
	Sub               string `json:"sub"`
	PreferredUsername string `json:"preferred_username"`
}

// GetUserInfo returns user info for a given user ID.
func (p *Provider) GetUserInfo(ctx context.Context, userID string) (*UserInfo, error) {
	user, err := p.users.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("finding user for userinfo: %w", err)
	}
	return &UserInfo{
		Sub:               user.ID,
		PreferredUsername: user.Username,
	}, nil
}

func verifyCodeChallenge(challenge, verifier string) bool {
	h := sha256.Sum256([]byte(verifier))
	computed := base64.RawURLEncoding.EncodeToString(h[:])
	return computed == challenge
}
