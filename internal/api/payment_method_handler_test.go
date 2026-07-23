package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/lucatorrekens/bakery-app/internal/payment"
	"github.com/stretchr/testify/assert"
)

// stubPaymentMethodService is a controllable stub for testing payment method handler.
type stubPaymentMethodService struct {
	methods      []payment.SavedPaymentMethod
	clientSecret string
	listErr      error
	setupErr     error
	detachErr    error
	defaultErr   error
}

func (s *stubPaymentMethodService) ListPaymentMethods(_ context.Context, _ string) ([]payment.SavedPaymentMethod, error) {
	if s.listErr != nil {
		return nil, s.listErr
	}
	return s.methods, nil
}

func (s *stubPaymentMethodService) CreateSetupIntent(_ context.Context, _ string) (string, error) {
	if s.setupErr != nil {
		return "", s.setupErr
	}
	return s.clientSecret, nil
}

func (s *stubPaymentMethodService) DetachPaymentMethod(_ context.Context, _ string, _ string) error {
	return s.detachErr
}

func (s *stubPaymentMethodService) SetDefaultPaymentMethod(_ context.Context, _ string, _ string) error {
	return s.defaultErr
}

func setupPaymentMethodTestRouter(svc PaymentMethodService) chi.Router {
	handler := NewPaymentMethodHandler(svc)
	r := chi.NewRouter()
	r.Use(testAuthMiddleware)
	handler.RegisterRoutes(r)
	return r
}

func TestPaymentMethodHandler_ListMethods_Unauthenticated(t *testing.T) {
	svc := &stubPaymentMethodService{}
	router := setupPaymentMethodTestRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/user/payment-methods", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestPaymentMethodHandler_ListMethods_Authenticated(t *testing.T) {
	svc := &stubPaymentMethodService{
		methods: []payment.SavedPaymentMethod{
			{ID: "pm_1", Brand: "visa", Last4: "4242", ExpMonth: 12, ExpYear: 2025, IsDefault: true},
		},
	}
	router := setupPaymentMethodTestRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/user/payment-methods", nil)
	req.Header.Set("X-User-ID", "user-1") // fallback auth for testing
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "pm_1")
	assert.Contains(t, w.Body.String(), "visa")
	assert.Contains(t, w.Body.String(), "4242")
}

func TestPaymentMethodHandler_CreateSetupIntent_Unauthenticated(t *testing.T) {
	svc := &stubPaymentMethodService{}
	router := setupPaymentMethodTestRouter(svc)

	req := httptest.NewRequest(http.MethodPost, "/api/user/payment-methods/setup", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestPaymentMethodHandler_CreateSetupIntent_Authenticated(t *testing.T) {
	svc := &stubPaymentMethodService{clientSecret: "seti_secret_123"}
	router := setupPaymentMethodTestRouter(svc)

	req := httptest.NewRequest(http.MethodPost, "/api/user/payment-methods/setup", nil)
	req.Header.Set("X-User-ID", "user-1")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "seti_secret_123")
}

func TestPaymentMethodHandler_DetachMethod_Unauthenticated(t *testing.T) {
	svc := &stubPaymentMethodService{}
	router := setupPaymentMethodTestRouter(svc)

	req := httptest.NewRequest(http.MethodDelete, "/api/user/payment-methods/pm_123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestPaymentMethodHandler_DetachMethod_Authenticated(t *testing.T) {
	svc := &stubPaymentMethodService{}
	router := setupPaymentMethodTestRouter(svc)

	req := httptest.NewRequest(http.MethodDelete, "/api/user/payment-methods/pm_123", nil)
	req.Header.Set("X-User-ID", "user-1")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestPaymentMethodHandler_DetachMethod_NotOwned(t *testing.T) {
	svc := &stubPaymentMethodService{detachErr: payment.ErrPaymentMethodNotOwned}
	router := setupPaymentMethodTestRouter(svc)

	req := httptest.NewRequest(http.MethodDelete, "/api/user/payment-methods/pm_123", nil)
	req.Header.Set("X-User-ID", "user-1")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestPaymentMethodHandler_SetDefault_Unauthenticated(t *testing.T) {
	svc := &stubPaymentMethodService{}
	router := setupPaymentMethodTestRouter(svc)

	req := httptest.NewRequest(http.MethodPut, "/api/user/payment-methods/pm_123/default", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestPaymentMethodHandler_SetDefault_Authenticated(t *testing.T) {
	svc := &stubPaymentMethodService{}
	router := setupPaymentMethodTestRouter(svc)

	req := httptest.NewRequest(http.MethodPut, "/api/user/payment-methods/pm_123/default", nil)
	req.Header.Set("X-User-ID", "user-1")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "default_set")
}

func TestPaymentMethodHandler_SetDefault_NotOwned(t *testing.T) {
	svc := &stubPaymentMethodService{defaultErr: payment.ErrPaymentMethodNotOwned}
	router := setupPaymentMethodTestRouter(svc)

	req := httptest.NewRequest(http.MethodPut, "/api/user/payment-methods/pm_123/default", nil)
	req.Header.Set("X-User-ID", "user-1")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}
