package api

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lucatorrekens/bakery-app/internal/api/dto"
	"github.com/lucatorrekens/bakery-app/internal/service"
)

// OAuthHandler handles OAuth/social login HTTP endpoints.
type OAuthHandler struct {
	oauthSvc  *service.OAuthService
	stateKey  []byte // HMAC key for signing OAuth state tokens
	stateTTL  time.Duration
	nowFunc   func() time.Time
}

// NewOAuthHandler creates a new OAuthHandler.
func NewOAuthHandler(oauthSvc *service.OAuthService, stateKey []byte) *OAuthHandler {
	return &OAuthHandler{
		oauthSvc: oauthSvc,
		stateKey: stateKey,
		stateTTL: 10 * time.Minute,
		nowFunc:  time.Now,
	}
}

// oauthURLResponse is the JSON response for GET /api/auth/oauth/{provider}.
type oauthURLResponse struct {
	URL   string `json:"url"`
	State string `json:"state"`
}

// oauthCallbackRequest is the JSON body for POST /api/auth/oauth/{provider}/callback.
type oauthCallbackRequest struct {
	Code  string `json:"code"`
	State string `json:"state"`
}

// oauthCallbackResponse is the JSON response for a successful OAuth callback.
type oauthCallbackResponse struct {
	Token string       `json:"token"`
	User  userResponse `json:"user"`
}

// generateOAuthState creates a signed, time-limited state token.
// Format: base64(nonce) + "." + hex(hmac(nonce + "|" + timestamp))
func (h *OAuthHandler) generateOAuthState() (string, error) {
	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	nonceB64 := base64.RawURLEncoding.EncodeToString(nonce)
	ts := h.nowFunc().Unix()
	payload := nonceB64 + "|" + time.Unix(ts, 0).Format(time.RFC3339)

	mac := hmac.New(sha256.New, h.stateKey)
	mac.Write([]byte(payload))
	sig := hex.EncodeToString(mac.Sum(nil))

	return nonceB64 + "." + sig + "." + time.Unix(ts, 0).Format(time.RFC3339), nil
}

// verifyOAuthState validates the state token signature and expiry.
func (h *OAuthHandler) verifyOAuthState(state string) bool {
	parts := strings.SplitN(state, ".", 3)
	if len(parts) != 3 {
		return false
	}
	nonceB64, sig, tsStr := parts[0], parts[1], parts[2]

	// Verify signature
	payload := nonceB64 + "|" + tsStr
	mac := hmac.New(sha256.New, h.stateKey)
	mac.Write([]byte(payload))
	expectedSig := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(sig), []byte(expectedSig)) {
		return false
	}

	// Verify expiry
	ts, err := time.Parse(time.RFC3339, tsStr)
	if err != nil {
		return false
	}
	if h.nowFunc().Sub(ts) > h.stateTTL {
		return false
	}

	return true
}

// GetAuthURL handles GET /api/auth/oauth/{provider}.
// Returns the OAuth authorization URL for the given provider with a server-generated signed state.
func (h *OAuthHandler) GetAuthURL(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	if provider == "" {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_PROVIDER",
			Message: "provider is required",
		})
		return
	}

	state, err := h.generateOAuthState()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "an unexpected error occurred",
		})
		return
	}

	authURL, err := h.oauthSvc.GetAuthURL(provider, state)
	if err != nil {
		if errors.Is(err, service.ErrUnsupportedProvider) {
			writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
				Code:    "UNSUPPORTED_PROVIDER",
				Message: "unsupported OAuth provider: " + provider,
			})
			return
		}
		writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "an unexpected error occurred",
		})
		return
	}

	writeJSON(w, http.StatusOK, oauthURLResponse{URL: authURL, State: state})
}

// HandleCallback handles POST /api/auth/oauth/{provider}/callback.
// Validates the signed state token, exchanges the authorization code for user info,
// creates or finds the user, and issues a JWT.
func (h *OAuthHandler) HandleCallback(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	if provider == "" {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_PROVIDER",
			Message: "provider is required",
		})
		return
	}

	var req oauthCallbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_BODY",
			Message: "invalid request body",
		})
		return
	}

	if req.Code == "" {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "MISSING_CODE",
			Message: "authorization code is required",
		})
		return
	}

	// Verify the server-generated state token (CSRF protection)
	if req.State == "" || !h.verifyOAuthState(req.State) {
		writeJSON(w, http.StatusForbidden, dto.ErrorResponse{
			Code:    "INVALID_STATE",
			Message: "invalid or expired OAuth state parameter",
		})
		return
	}

	result, err := h.oauthSvc.HandleCallback(r.Context(), provider, req.Code)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUnsupportedProvider):
			writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
				Code:    "UNSUPPORTED_PROVIDER",
				Message: "unsupported OAuth provider: " + provider,
			})
		case errors.Is(err, service.ErrOAuthNoEmail):
			writeJSON(w, http.StatusUnprocessableEntity, dto.ErrorResponse{
				Code:    "OAUTH_NO_EMAIL",
				Message: "the OAuth provider did not return an email address",
			})
		case errors.Is(err, service.ErrOAuthAccountLinkRequiresVerification):
			writeJSON(w, http.StatusConflict, dto.ErrorResponse{
				Code:    "ACCOUNT_LINK_REQUIRES_VERIFICATION",
				Message: "an account with this email already exists; please log in with your password to link this provider",
			})
		default:
			writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{
				Code:    "OAUTH_ERROR",
				Message: "OAuth authentication failed",
			})
		}
		return
	}

	writeJSON(w, http.StatusOK, oauthCallbackResponse{
		Token: result.Token,
		User: userResponse{
			ID:       result.User.ID,
			Username: result.User.Username,
			Role:     int(result.User.Role),
		},
	})
}
