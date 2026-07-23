package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// OAuthProvider defines the interface for social login providers.
type OAuthProvider interface {
	// GetAuthURL returns the OAuth authorization URL to redirect the user to.
	GetAuthURL(state string, redirectURI string) string
	// ExchangeCode exchanges the authorization code for user info.
	ExchangeCode(ctx context.Context, code string, redirectURI string) (*OAuthUserInfo, error)
}

// OAuthUserInfo holds the user details returned from the OAuth provider.
type OAuthUserInfo struct {
	ProviderUserID string
	Email          string
	Name           string
}

// GoogleProvider implements OAuth for Google.
type GoogleProvider struct {
	ClientID     string
	ClientSecret string
	HTTPClient   *http.Client // optional; defaults to http.DefaultClient
}

func (g *GoogleProvider) httpClient() *http.Client {
	if g.HTTPClient != nil {
		return g.HTTPClient
	}
	return http.DefaultClient
}

// GetAuthURL returns the Google OAuth2 authorization URL.
func (g *GoogleProvider) GetAuthURL(state, redirectURI string) string {
	params := url.Values{
		"client_id":     {g.ClientID},
		"redirect_uri":  {redirectURI},
		"response_type": {"code"},
		"scope":         {"openid email profile"},
		"state":         {state},
		"access_type":   {"offline"},
	}
	return "https://accounts.google.com/o/oauth2/v2/auth?" + params.Encode()
}

// ExchangeCode exchanges a Google authorization code for user info.
func (g *GoogleProvider) ExchangeCode(ctx context.Context, code, redirectURI string) (*OAuthUserInfo, error) {
	// Exchange code for access token
	resp, err := g.httpClient().PostForm("https://oauth2.googleapis.com/token", url.Values{
		"code":          {code},
		"client_id":     {g.ClientID},
		"client_secret": {g.ClientSecret},
		"redirect_uri":  {redirectURI},
		"grant_type":    {"authorization_code"},
	})
	if err != nil {
		return nil, fmt.Errorf("exchanging code for token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token exchange failed (status %d): %s", resp.StatusCode, body)
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		IDToken     string `json:"id_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("decoding token response: %w", err)
	}
	if tokenResp.AccessToken == "" {
		return nil, fmt.Errorf("empty access token in response")
	}

	// Fetch user info from Google
	req, err := http.NewRequestWithContext(ctx, "GET", "https://www.googleapis.com/oauth2/v3/userinfo", nil)
	if err != nil {
		return nil, fmt.Errorf("creating userinfo request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)

	userResp, err := g.httpClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching user info: %w", err)
	}
	defer userResp.Body.Close()

	if userResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(userResp.Body)
		return nil, fmt.Errorf("userinfo request failed (status %d): %s", userResp.StatusCode, body)
	}

	body, err := io.ReadAll(userResp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading userinfo response: %w", err)
	}

	var userInfo struct {
		Sub   string `json:"sub"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, fmt.Errorf("decoding userinfo: %w", err)
	}
	if userInfo.Sub == "" {
		return nil, fmt.Errorf("empty subject in userinfo response")
	}

	return &OAuthUserInfo{
		ProviderUserID: userInfo.Sub,
		Email:          userInfo.Email,
		Name:           userInfo.Name,
	}, nil
}

// AppleProvider implements OAuth for Apple Sign In.
type AppleProvider struct {
	ClientID   string // Services ID (e.g., com.example.app)
	TeamID     string
	KeyID      string
	PrivateKey string // PEM-encoded private key
	HTTPClient *http.Client
}

func (a *AppleProvider) httpClient() *http.Client {
	if a.HTTPClient != nil {
		return a.HTTPClient
	}
	return http.DefaultClient
}

// GetAuthURL returns the Apple Sign In authorization URL.
func (a *AppleProvider) GetAuthURL(state, redirectURI string) string {
	params := url.Values{
		"client_id":     {a.ClientID},
		"redirect_uri":  {redirectURI},
		"response_type": {"code id_token"},
		"scope":         {"name email"},
		"state":         {state},
		"response_mode": {"form_post"},
	}
	return "https://appleid.apple.com/auth/authorize?" + params.Encode()
}

// ExchangeCode exchanges an Apple authorization code for user info.
// Note: Apple token exchange requires a client_secret JWT signed with the private key.
// This is a placeholder — real implementation requires JWT generation with the team/key credentials.
func (a *AppleProvider) ExchangeCode(ctx context.Context, code, redirectURI string) (*OAuthUserInfo, error) {
	return nil, fmt.Errorf("apple OAuth exchange not yet implemented (requires JWT client secret generation)")
}
