package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/lucatorrekens/bakery-app/internal/api/dto"
	"github.com/lucatorrekens/bakery-app/internal/domain"
	"github.com/lucatorrekens/bakery-app/internal/payment"
	"github.com/lucatorrekens/bakery-app/internal/repository/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubPaymentService is a controllable payment service stub for testing.
type stubPaymentService struct {
	processErr error
}

func (s *stubPaymentService) InitiatePayment(_ context.Context, _ string, _ int64) (*domain.PaymentLink, error) {
	return &domain.PaymentLink{URL: "https://pay.example.com/test", ExpiresIn: 1800}, nil
}

func (s *stubPaymentService) ProcessPaymentCallback(_ context.Context, _ string, _ string) error {
	return s.processErr
}

func (s *stubPaymentService) InitiateRefund(_ context.Context, _ string, _ int64) error {
	return nil
}

func setupPaymentTestRouter(paymentSvc domain.PaymentService, orderRepo *memory.OrderRepo) chi.Router {
	handler := NewPaymentHandler(paymentSvc, orderRepo)
	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	return r
}

func seedPendingOrder(repo *memory.OrderRepo, orderID string) {
	order := &domain.Order{
		ID:            orderID,
		BakeryID:      "bakery-1",
		UserID:        "user-1",
		Status:        domain.OrderStatusPendingPayment,
		TotalAmount:   1500,
		PaymentMethod: domain.PaymentMethodOnline,
	}
	_ = repo.Save(context.Background(), order)
}

func TestPaymentCallback_Success(t *testing.T) {
	orderRepo := memory.NewOrderRepo()
	seedPendingOrder(orderRepo, "order-1")

	paymentSvc := &stubPaymentService{processErr: nil}
	router := setupPaymentTestRouter(paymentSvc, orderRepo)

	reqBody := dto.PaymentCallbackRequest{
		OrderID:    "order-1",
		PaymentRef: "pay-ref-123",
		Status:     "success",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/payments/callback", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp dto.PaymentCallbackResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)

	assert.Equal(t, "order-1", resp.OrderID)
	assert.Equal(t, "confirmed", resp.Status)
}

func TestPaymentCallback_LinkExpired_CancelsOrder(t *testing.T) {
	orderRepo := memory.NewOrderRepo()
	seedPendingOrder(orderRepo, "order-2")

	paymentSvc := &stubPaymentService{processErr: payment.ErrLinkExpired}
	router := setupPaymentTestRouter(paymentSvc, orderRepo)

	reqBody := dto.PaymentCallbackRequest{
		OrderID:    "order-2",
		PaymentRef: "pay-ref-456",
		Status:     "success",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/payments/callback", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusGone, w.Code)

	var resp dto.ErrorResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "PAYMENT_LINK_EXPIRED", resp.Code)

	// Verify order was cancelled
	order, _ := orderRepo.GetByID(context.Background(), "order-2")
	assert.Equal(t, domain.OrderStatusCancelled, order.Status)
}

func TestPaymentCallback_Failure_TracksRetries(t *testing.T) {
	orderRepo := memory.NewOrderRepo()
	seedPendingOrder(orderRepo, "order-3")

	paymentSvc := &stubPaymentService{processErr: nil}
	router := setupPaymentTestRouter(paymentSvc, orderRepo)

	// First two failures should keep order in pending_payment
	for i := 0; i < 2; i++ {
		reqBody := dto.PaymentCallbackRequest{
			OrderID:    "order-3",
			PaymentRef: "",
			Status:     "failed",
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/payments/callback", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp dto.PaymentCallbackResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "pending_payment", resp.Status)
	}

	// Third failure should cancel the order
	reqBody := dto.PaymentCallbackRequest{
		OrderID:    "order-3",
		PaymentRef: "",
		Status:     "failed",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/payments/callback", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp dto.PaymentCallbackResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "cancelled", resp.Status)

	// Verify order was cancelled
	order, _ := orderRepo.GetByID(context.Background(), "order-3")
	assert.Equal(t, domain.OrderStatusCancelled, order.Status)
}

func TestPaymentCallback_InvalidJSON_Returns400(t *testing.T) {
	orderRepo := memory.NewOrderRepo()
	paymentSvc := &stubPaymentService{}
	router := setupPaymentTestRouter(paymentSvc, orderRepo)

	req := httptest.NewRequest(http.MethodPost, "/api/payments/callback", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPaymentCallback_MissingOrderID_Returns400(t *testing.T) {
	orderRepo := memory.NewOrderRepo()
	paymentSvc := &stubPaymentService{}
	router := setupPaymentTestRouter(paymentSvc, orderRepo)

	reqBody := dto.PaymentCallbackRequest{
		OrderID:    "",
		PaymentRef: "ref",
		Status:     "success",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/payments/callback", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPaymentCallback_InvalidStatus_Returns400(t *testing.T) {
	orderRepo := memory.NewOrderRepo()
	paymentSvc := &stubPaymentService{}
	router := setupPaymentTestRouter(paymentSvc, orderRepo)

	reqBody := dto.PaymentCallbackRequest{
		OrderID:    "order-1",
		PaymentRef: "ref",
		Status:     "invalid",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/payments/callback", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPaymentCallback_LinkNotFound_Returns404(t *testing.T) {
	orderRepo := memory.NewOrderRepo()
	paymentSvc := &stubPaymentService{processErr: payment.ErrLinkNotFound}
	router := setupPaymentTestRouter(paymentSvc, orderRepo)

	reqBody := dto.PaymentCallbackRequest{
		OrderID:    "nonexistent-order",
		PaymentRef: "ref",
		Status:     "success",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/payments/callback", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestPaymentCallback_LinkAlreadyUsed_Returns409(t *testing.T) {
	orderRepo := memory.NewOrderRepo()
	paymentSvc := &stubPaymentService{processErr: payment.ErrLinkUsed}
	router := setupPaymentTestRouter(paymentSvc, orderRepo)

	reqBody := dto.PaymentCallbackRequest{
		OrderID:    "order-1",
		PaymentRef: "ref",
		Status:     "success",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/payments/callback", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestPaymentCallback_InternalError_Returns500(t *testing.T) {
	orderRepo := memory.NewOrderRepo()
	paymentSvc := &stubPaymentService{processErr: errors.New("unexpected error")}
	router := setupPaymentTestRouter(paymentSvc, orderRepo)

	reqBody := dto.PaymentCallbackRequest{
		OrderID:    "order-1",
		PaymentRef: "ref",
		Status:     "success",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/payments/callback", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
