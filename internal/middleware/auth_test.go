package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSecret = "test-secret-key-for-jwt"

// createValidToken creates a valid JWT token with the given user ID and expiry.
func createValidToken(userID string, expiry time.Time) string {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": expiry.Unix(),
		"iat": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte(testSecret))
	return signed
}

// createTokenWithUserIDClaim creates a JWT with a "userId" claim instead of "sub".
func createTokenWithUserIDClaim(userID string, expiry time.Time) string {
	claims := jwt.MapClaims{
		"userId": userID,
		"exp":    expiry.Unix(),
		"iat":    time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte(testSecret))
	return signed
}

// createExpiredToken creates a JWT that has already expired.
func createExpiredToken(userID string) string {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(-1 * time.Hour).Unix(),
		"iat": time.Now().Add(-2 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte(testSecret))
	return signed
}

// createTokenWithWrongSecret creates a JWT signed with a different secret.
func createTokenWithWrongSecret(userID string) string {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(1 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte("wrong-secret"))
	return signed
}

func TestJWTAuth_ValidToken(t *testing.T) {
	var capturedUserID string

	handler := JWTAuth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUserID = GetUserIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	token := createValidToken("user-42", time.Now().Add(1*time.Hour))
	req := httptest.NewRequest(http.MethodGet, "/api/orders", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "user-42", capturedUserID)
}

func TestJWTAuth_ValidTokenWithUserIDClaim(t *testing.T) {
	var capturedUserID string

	handler := JWTAuth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUserID = GetUserIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	token := createTokenWithUserIDClaim("user-99", time.Now().Add(1*time.Hour))
	req := httptest.NewRequest(http.MethodGet, "/api/orders", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "user-99", capturedUserID)
}

func TestJWTAuth_MissingAuthorizationHeader(t *testing.T) {
	handler := JWTAuth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/orders", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var errResp authErrorResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&errResp))
	assert.Equal(t, "UNAUTHORIZED", errResp.Code)
	assert.Equal(t, "authentication required", errResp.Message)
}

func TestJWTAuth_MissingBearerPrefix(t *testing.T) {
	handler := JWTAuth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	token := createValidToken("user-1", time.Now().Add(1*time.Hour))
	req := httptest.NewRequest(http.MethodGet, "/api/orders", nil)
	req.Header.Set("Authorization", token) // no "Bearer " prefix

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var errResp authErrorResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&errResp))
	assert.Equal(t, "UNAUTHORIZED", errResp.Code)
	assert.Equal(t, "authentication required", errResp.Message)
}

func TestJWTAuth_ExpiredToken(t *testing.T) {
	handler := JWTAuth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	token := createExpiredToken("user-expired")
	req := httptest.NewRequest(http.MethodGet, "/api/orders", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var errResp authErrorResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&errResp))
	assert.Equal(t, "TOKEN_EXPIRED", errResp.Code)
	assert.Equal(t, "token has expired", errResp.Message)
}

func TestJWTAuth_TamperedToken(t *testing.T) {
	handler := JWTAuth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	token := createTokenWithWrongSecret("user-tampered")
	req := httptest.NewRequest(http.MethodGet, "/api/orders", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var errResp authErrorResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&errResp))
	assert.Equal(t, "INVALID_TOKEN", errResp.Code)
	assert.Equal(t, "invalid token", errResp.Message)
}

func TestJWTAuth_MalformedToken(t *testing.T) {
	handler := JWTAuth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/orders", nil)
	req.Header.Set("Authorization", "Bearer not-a-valid-jwt")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var errResp authErrorResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&errResp))
	assert.Equal(t, "INVALID_TOKEN", errResp.Code)
	assert.Equal(t, "invalid token", errResp.Message)
}

func TestJWTAuth_EmptyBearerToken(t *testing.T) {
	handler := JWTAuth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/orders", nil)
	req.Header.Set("Authorization", "Bearer ")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var errResp authErrorResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&errResp))
	assert.Equal(t, "UNAUTHORIZED", errResp.Code)
}

func TestJWTAuth_TokenWithNoSubClaim(t *testing.T) {
	// Token with no "sub" or "userId" claim
	claims := jwt.MapClaims{
		"exp": time.Now().Add(1 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte(testSecret))

	handler := JWTAuth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/orders", nil)
	req.Header.Set("Authorization", "Bearer "+signed)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var errResp authErrorResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&errResp))
	assert.Equal(t, "INVALID_TOKEN", errResp.Code)
	assert.Equal(t, "invalid token", errResp.Message)
}

func TestGetUserIDFromContext_NoValue(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	userID := GetUserIDFromContext(req.Context())
	assert.Equal(t, "", userID)
}
