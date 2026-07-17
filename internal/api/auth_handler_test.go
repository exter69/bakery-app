package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/lucatorrekens/bakery-app/internal/repository/memory"
	"github.com/lucatorrekens/bakery-app/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testJWTSecret = "test-secret-key"

func setupAuthTestHandler() (*AuthHandler, chi.Router) {
	userRepo := memory.NewUserRepo()
	authSvc := service.NewAuthService(service.AuthServiceConfig{
		UserRepo:  userRepo,
		JWTSecret: testJWTSecret,
	})
	handler := NewAuthHandler(authSvc)

	r := chi.NewRouter()
	r.Post("/api/auth/register", handler.Register)
	r.Post("/api/auth/login", handler.Login)
	return handler, r
}

func TestRegister_Success(t *testing.T) {
	_, router := setupAuthTestHandler()

	body := `{"username":"alice","password":"secret123","role":2}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp registerResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)

	assert.NotEmpty(t, resp.ID)
	assert.Equal(t, "alice", resp.Username)
	assert.Equal(t, 2, resp.Role)
}

func TestRegister_DuplicateUsername(t *testing.T) {
	_, router := setupAuthTestHandler()

	body := `{"username":"bob","password":"secret123","role":2}`

	// First registration should succeed
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	// Second registration with same username should fail with 409
	req = httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)

	var errResp map[string]string
	err := json.NewDecoder(w.Body).Decode(&errResp)
	require.NoError(t, err)
	assert.Equal(t, "USERNAME_EXISTS", errResp["code"])
}

func TestRegister_ValidationError_EmptyUsername(t *testing.T) {
	_, router := setupAuthTestHandler()

	body := `{"username":"","password":"secret123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestRegister_ValidationError_ShortPassword(t *testing.T) {
	_, router := setupAuthTestHandler()

	body := `{"username":"carol","password":"abc"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestRegister_AdminRoleRejected(t *testing.T) {
	_, router := setupAuthTestHandler()

	// Trying to register as admin (role 0) should default to customer (role 2)
	body := `{"username":"mallory","password":"secret123","role":0}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp registerResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, 2, resp.Role) // Should be customer, not admin
}

func TestLogin_Success(t *testing.T) {
	_, router := setupAuthTestHandler()

	// Register first
	regBody := `{"username":"dave","password":"password123","role":2}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString(regBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	// Login
	loginBody := `{"username":"dave","password":"password123"}`
	req = httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(loginBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp loginResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)

	assert.NotEmpty(t, resp.Token)
	assert.Equal(t, "dave", resp.User.Username)
	assert.Equal(t, 2, resp.User.Role)
	assert.NotEmpty(t, resp.User.ID)
}

func TestLogin_WrongPassword(t *testing.T) {
	_, router := setupAuthTestHandler()

	// Register first
	regBody := `{"username":"eve","password":"correct-password","role":2}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString(regBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	// Login with wrong password
	loginBody := `{"username":"eve","password":"wrong-password"}`
	req = httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(loginBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var errResp map[string]string
	err := json.NewDecoder(w.Body).Decode(&errResp)
	require.NoError(t, err)
	assert.Equal(t, "INVALID_CREDENTIALS", errResp["code"])
}

func TestLogin_NonExistentUser(t *testing.T) {
	_, router := setupAuthTestHandler()

	loginBody := `{"username":"ghost","password":"doesnotmatter"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(loginBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var errResp map[string]string
	err := json.NewDecoder(w.Body).Decode(&errResp)
	require.NoError(t, err)
	assert.Equal(t, "INVALID_CREDENTIALS", errResp["code"])
}
