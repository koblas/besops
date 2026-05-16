package auth

import (
	"encoding/json"
	"net/http"
)

// Handler exposes OIDC endpoints as HTTP handlers.
type Handler struct {
	provider *Provider
}

// NewHandler creates an HTTP handler for the OIDC provider endpoints.
func NewHandler(provider *Provider) *Handler {
	return &Handler{provider: provider}
}

// RegisterRoutes registers the OIDC endpoints on the given mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /.well-known/openid-configuration", h.discovery)
	mux.HandleFunc("GET /.well-known/jwks.json", h.jwks)
	mux.HandleFunc("POST /auth/token", h.token)
	mux.HandleFunc("POST /auth/login", h.login)
	mux.HandleFunc("POST /auth/login/2fa", h.login2FA)
	mux.HandleFunc("GET /auth/userinfo", h.userinfo)
	mux.HandleFunc("POST /auth/logout", h.logout)
}

func (h *Handler) discovery(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, h.provider.Discovery())
}

func (h *Handler) jwks(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, h.provider.JWKS())
}

// loginRequest is the body for the login endpoint (built-in IDP).
type loginRequest struct {
	Username            string `json:"username"`
	Password            string `json:"password"`
	RedirectURI         string `json:"redirect_uri"`
	ClientID            string `json:"client_id"`
	CodeChallenge       string `json:"code_challenge"`
	CodeChallengeMethod string `json:"code_challenge_method"`
	State               string `json:"state"`
	Nonce               string `json:"nonce"`
	Scope               string `json:"scope"`
}

// loginResponse is returned on successful authentication.
type loginResponse struct {
	Code        string `json:"code,omitempty"`
	State       string `json:"state,omitempty"`
	Requires2FA bool   `json:"requires_2fa,omitempty"`
	LoginToken  string `json:"login_token,omitempty"`
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorBody("invalid request body"))
		return
	}

	result, err := h.provider.Authenticate(r.Context(), req.Username, req.Password)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, errorBody("invalid credentials"))
		return
	}

	if result.Requires2FA {
		h.handleLogin2FAPending(w, r, result, &req)
		return
	}

	h.handleLoginIssueCode(w, r, result.UserID, &req)
}

func (h *Handler) handleLogin2FAPending(w http.ResponseWriter, r *http.Request, result *LoginResult, req *loginRequest) {
	pendingToken, err := generateOpaqueToken()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorBody("internal error"))
		return
	}

	_ = h.provider.codes.Store(r.Context(), &Code{
		Code:                pendingToken,
		UserID:              result.UserID,
		ClientID:            req.ClientID,
		RedirectURI:         req.RedirectURI,
		Scope:               req.Scope,
		CodeChallenge:       req.CodeChallenge,
		CodeChallengeMethod: req.CodeChallengeMethod,
		Nonce:               req.Nonce,
	})

	writeJSON(w, http.StatusOK, loginResponse{
		Requires2FA: true,
		LoginToken:  pendingToken,
		State:       req.State,
	})
}

func (h *Handler) handleLoginIssueCode(w http.ResponseWriter, r *http.Request, userID string, req *loginRequest) {
	authReq := &AuthorizeRequest{
		ClientID:            req.ClientID,
		RedirectURI:         req.RedirectURI,
		ResponseType:        "code",
		Scope:               req.Scope,
		State:               req.State,
		Nonce:               req.Nonce,
		CodeChallenge:       req.CodeChallenge,
		CodeChallengeMethod: req.CodeChallengeMethod,
	}

	if err := h.provider.ValidateAuthorizeRequest(authReq); err != nil {
		writeJSON(w, http.StatusBadRequest, errorBody(err.Error()))
		return
	}

	code, err := h.provider.IssueCode(r.Context(), userID, authReq)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorBody("failed to issue code"))
		return
	}

	writeJSON(w, http.StatusOK, loginResponse{
		Code:  code,
		State: req.State,
	})
}

type login2FARequest struct {
	LoginToken string `json:"login_token"`
	Token      string `json:"token"`
}

func (h *Handler) login2FA(w http.ResponseWriter, r *http.Request) {
	var req login2FARequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorBody("invalid request body"))
		return
	}

	pending, err := h.provider.codes.Consume(r.Context(), req.LoginToken)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, errorBody("invalid or expired login token"))
		return
	}

	if verifyErr := h.provider.Verify2FA(r.Context(), pending.UserID, req.Token); verifyErr != nil {
		writeJSON(w, http.StatusUnauthorized, errorBody("invalid 2FA token"))
		return
	}

	authReq := &AuthorizeRequest{
		ClientID:            pending.ClientID,
		RedirectURI:         pending.RedirectURI,
		ResponseType:        "code",
		Scope:               pending.Scope,
		Nonce:               pending.Nonce,
		CodeChallenge:       pending.CodeChallenge,
		CodeChallengeMethod: pending.CodeChallengeMethod,
	}

	code, err := h.provider.IssueCode(r.Context(), pending.UserID, authReq)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorBody("failed to issue code"))
		return
	}

	writeJSON(w, http.StatusOK, loginResponse{Code: code})
}

type tokenRequest struct {
	GrantType    string `json:"grant_type"`
	Code         string `json:"code"`
	RedirectURI  string `json:"redirect_uri"`
	CodeVerifier string `json:"code_verifier"`
	RefreshToken string `json:"refresh_token"`
	ClientID     string `json:"client_id"`
}

func (h *Handler) token(w http.ResponseWriter, r *http.Request) {
	var req tokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorBody("invalid request body"))
		return
	}

	tokens, err := h.provider.HandleTokenRequest(r.Context(), &TokenRequest{
		GrantType:    req.GrantType,
		Code:         req.Code,
		RedirectURI:  req.RedirectURI,
		CodeVerifier: req.CodeVerifier,
		RefreshToken: req.RefreshToken,
		ClientID:     req.ClientID,
	})
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorBody(err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, tokens)
}

func (h *Handler) userinfo(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		token := extractBearerToken(r)
		if token == "" {
			writeJSON(w, http.StatusUnauthorized, errorBody("unauthorized"))
			return
		}
		var err error
		userID, err = h.provider.ValidateAccessToken(token)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, errorBody("unauthorized"))
			return
		}
	}

	info, err := h.provider.GetUserInfo(r.Context(), userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorBody("user not found"))
		return
	}

	writeJSON(w, http.StatusOK, info)
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.RefreshToken == "" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	_ = h.provider.RevokeSession(r.Context(), body.RefreshToken)
	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func errorBody(msg string) map[string]string {
	return map[string]string{"error": msg}
}
