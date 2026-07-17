package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lucatorrekens/bakery-app/internal/api"
	"github.com/lucatorrekens/bakery-app/internal/api/dto"
	"github.com/lucatorrekens/bakery-app/internal/domain"
	"github.com/lucatorrekens/bakery-app/internal/middleware"
	"github.com/lucatorrekens/bakery-app/internal/payment"
	"github.com/lucatorrekens/bakery-app/internal/repository/memory"
	"github.com/lucatorrekens/bakery-app/internal/service"
)

const testJWTSecret = "integration-test-secret"

// testServer holds the full server stack for integration testing.
type testServer struct {
	server     *httptest.Server
	bakeryRepo *memory.BakeryRepo
	orderRepo  *memory.OrderRepo
}

// setupTestServer creates a fully wired test server with all handlers, services, and middleware.
func setupTestServer(t *testing.T) *testServer {
	t.Helper()

	bakeryRepo := memory.NewBakeryRepo()
	orderRepo := memory.NewOrderRepo()
	reservationRepo := memory.NewReservationRepo()

	// Services
	bakerySvc := service.NewBakeryService(bakeryRepo)

	paymentGateway := payment.NewStubGateway()
	paymentSvc := payment.NewPaymentService(payment.ServiceConfig{
		Gateway:   paymentGateway,
		OrderRepo: orderRepo,
	})

	orderSvc := service.NewOrderService(service.OrderServiceConfig{
		OrderRepo:  orderRepo,
		BakeryRepo: bakeryRepo,
		PaymentSvc: paymentSvc,
	})

	reservationSvc := service.NewReservationService(service.ReservationServiceConfig{
		BakeryRepo:      bakeryRepo,
		ReservationRepo: reservationRepo,
	})

	// Handlers
	bakeryHandler := api.NewBakeryHandler(bakerySvc)
	orderHandler := api.NewOrderHandler(orderSvc)
	reservationHandler := api.NewReservationHandler(reservationSvc)
	paymentHandler := api.NewPaymentHandler(paymentSvc, orderRepo)

	// Router — bakery endpoints are public, everything else requires auth
	r := chi.NewRouter()

	// Public bakery routes (no auth)
	bakeryHandler.RegisterRoutes(r)

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.JWTAuth(testJWTSecret))
		orderHandler.RegisterRoutes(r)
		reservationHandler.RegisterRoutes(r)
		paymentHandler.RegisterRoutes(r)
	})

	server := httptest.NewServer(r)
	t.Cleanup(server.Close)

	return &testServer{
		server:     server,
		bakeryRepo: bakeryRepo,
		orderRepo:  orderRepo,
	}
}

// seedTestBakery seeds a bakery with products into the test repo.
func seedTestBakery(ts *testServer) (bakeryID string, productIDs []string) {
	bakeryID = "bakery-integration-1"

	bakery := domain.Bakery{
		ID:       bakeryID,
		Name:     "Integration Test Bakery",
		PhotoURL: "https://example.com/photo.jpg",
		Schedule: []domain.DaySchedule{
			{Day: domain.Monday, OpenTime: domain.TimeOfDay{Hour: 8, Minute: 0}, CloseTime: domain.TimeOfDay{Hour: 18, Minute: 0}, IsOpen: true},
			{Day: domain.Tuesday, OpenTime: domain.TimeOfDay{Hour: 8, Minute: 0}, CloseTime: domain.TimeOfDay{Hour: 18, Minute: 0}, IsOpen: true},
			{Day: domain.Wednesday, OpenTime: domain.TimeOfDay{Hour: 8, Minute: 0}, CloseTime: domain.TimeOfDay{Hour: 18, Minute: 0}, IsOpen: true},
			{Day: domain.Thursday, OpenTime: domain.TimeOfDay{Hour: 8, Minute: 0}, CloseTime: domain.TimeOfDay{Hour: 18, Minute: 0}, IsOpen: true},
			{Day: domain.Friday, OpenTime: domain.TimeOfDay{Hour: 8, Minute: 0}, CloseTime: domain.TimeOfDay{Hour: 18, Minute: 0}, IsOpen: true},
			{Day: domain.Saturday, OpenTime: domain.TimeOfDay{Hour: 9, Minute: 0}, CloseTime: domain.TimeOfDay{Hour: 14, Minute: 0}, IsOpen: true},
			{Day: domain.Sunday, IsOpen: false},
		},
	}
	ts.bakeryRepo.SeedBakery(bakery)

	products := []domain.Product{
		{ID: "prod-1", BakeryID: bakeryID, Name: "Croissant", Price: 350, Category: "Pastries", IsAvailable: true},
		{ID: "prod-2", BakeryID: bakeryID, Name: "Baguette", Price: 500, Category: "Breads", IsAvailable: true},
	}
	for _, p := range products {
		ts.bakeryRepo.SeedProduct(p)
		productIDs = append(productIDs, p.ID)
	}

	return bakeryID, productIDs
}

// generateTestToken creates a valid JWT token for the given user.
func generateTestToken(t *testing.T, userID string) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(1 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	})
	signed, err := token.SignedString([]byte(testJWTSecret))
	require.NoError(t, err)
	return signed
}

// generateExpiredToken creates an expired JWT token.
func generateExpiredToken(t *testing.T, userID string) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(-1 * time.Hour).Unix(),
		"iat": time.Now().Add(-2 * time.Hour).Unix(),
	})
	signed, err := token.SignedString([]byte(testJWTSecret))
	require.NoError(t, err)
	return signed
}

// doRequest is a helper to send HTTP requests to the test server.
func doRequest(t *testing.T, method, url, token string, body interface{}) *http.Response {
	t.Helper()
	var reqBody *bytes.Buffer
	if body != nil {
		b, err := json.Marshal(body)
		require.NoError(t, err)
		reqBody = bytes.NewBuffer(b)
	} else {
		reqBody = &bytes.Buffer{}
	}

	req, err := http.NewRequest(method, url, reqBody)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

// TestIntegration_FullOrderFlow tests the complete order lifecycle:
// browse bakeries → view menu → create order → payment callback → confirmation → list orders
func TestIntegration_FullOrderFlow(t *testing.T) {
	ts := setupTestServer(t)
	bakeryID, productIDs := seedTestBakery(ts)
	token := generateTestToken(t, "user-order-flow")

	// Step 1: Browse bakeries
	resp := doRequest(t, "GET", ts.server.URL+"/api/bakeries", token, nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var bakeryList dto.ListResponse[dto.BakeryCardResponse]
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&bakeryList))
	resp.Body.Close()
	assert.GreaterOrEqual(t, bakeryList.Total, 1)

	// Find our seeded bakery
	var foundBakery bool
	for _, b := range bakeryList.Items {
		if b.ID == bakeryID {
			foundBakery = true
			assert.Equal(t, "Integration Test Bakery", b.Name)
			break
		}
	}
	assert.True(t, foundBakery, "seeded bakery should appear in list")

	// Step 2: View menu
	resp = doRequest(t, "GET", ts.server.URL+"/api/bakeries/"+bakeryID+"/menu", token, nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var menu map[string][]domain.Product
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&menu))
	resp.Body.Close()
	assert.Contains(t, menu, "Pastries")
	assert.Contains(t, menu, "Breads")

	// Step 3: Create order
	orderReq := dto.CreateOrderRequest{
		BakeryID: bakeryID,
		Items: []dto.OrderItemRequest{
			{ProductID: productIDs[0], Quantity: 2},
			{ProductID: productIDs[1], Quantity: 1},
		},
		ScheduledDay: domain.Monday,
		ScheduledTime: dto.TimeSlotRequest{
			StartTime: "10:00",
			EndTime:   "10:30",
		},
	}
	resp = doRequest(t, "POST", ts.server.URL+"/api/orders", token, orderReq)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var orderResp dto.OrderResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&orderResp))
	resp.Body.Close()
	assert.NotEmpty(t, orderResp.ID)
	assert.Equal(t, "pending_payment", orderResp.Status)
	assert.NotEmpty(t, orderResp.PaymentLink)
	// Total: 2*350 + 1*500 = 1200
	assert.Equal(t, int64(1200), orderResp.TotalAmount)

	// Step 4: Payment callback - success
	callbackReq := dto.PaymentCallbackRequest{
		OrderID:    orderResp.ID,
		PaymentRef: "pay-ref-123",
		Status:     "success",
	}
	resp = doRequest(t, "POST", ts.server.URL+"/api/payments/callback", token, callbackReq)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var callbackResp dto.PaymentCallbackResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&callbackResp))
	resp.Body.Close()
	assert.Equal(t, "confirmed", callbackResp.Status)

	// Step 5: List orders - verify order shows as confirmed
	resp = doRequest(t, "GET", ts.server.URL+"/api/orders", token, nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var orderList dto.ListResponse[dto.OrderResponse]
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&orderList))
	resp.Body.Close()
	assert.GreaterOrEqual(t, orderList.Total, 1)

	var confirmedOrder *dto.OrderResponse
	for i, o := range orderList.Items {
		if o.ID == orderResp.ID {
			confirmedOrder = &orderList.Items[i]
			break
		}
	}
	require.NotNil(t, confirmedOrder, "created order should appear in list")
	assert.Equal(t, "confirmed", confirmedOrder.Status)
}

// TestIntegration_FullReservationFlow tests the complete reservation lifecycle:
// browse bakeries → view menu → create reservation → confirmed with on_spot payment
func TestIntegration_FullReservationFlow(t *testing.T) {
	ts := setupTestServer(t)
	bakeryID, productIDs := seedTestBakery(ts)
	token := generateTestToken(t, "user-reservation-flow")

	// Step 1: Browse bakeries
	resp := doRequest(t, "GET", ts.server.URL+"/api/bakeries", token, nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Step 2: View menu
	resp = doRequest(t, "GET", ts.server.URL+"/api/bakeries/"+bakeryID+"/menu", token, nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Step 3: Create reservation
	reservationReq := dto.CreateReservationRequest{
		BakeryID: bakeryID,
		Items: []dto.OrderItemRequest{
			{ProductID: productIDs[0], Quantity: 3},
			{ProductID: productIDs[1], Quantity: 2},
		},
		ScheduledDay: domain.Friday,
		ScheduledTime: dto.TimeSlotRequest{
			StartTime: "14:00",
			EndTime:   "14:30",
		},
	}
	resp = doRequest(t, "POST", ts.server.URL+"/api/reservations", token, reservationReq)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var reservationResp dto.ReservationResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&reservationResp))
	resp.Body.Close()

	assert.NotEmpty(t, reservationResp.ID)
	assert.Equal(t, "confirmed", reservationResp.Status)
	assert.Equal(t, "on_spot", reservationResp.PaymentMethod)
	// Total: 3*350 + 2*500 = 2050
	assert.Equal(t, int64(2050), reservationResp.TotalAmount)
}

// TestIntegration_CancelOrderWithRefund tests:
// create order → confirm payment → cancel order → verify cancelled
func TestIntegration_CancelOrderWithRefund(t *testing.T) {
	ts := setupTestServer(t)
	bakeryID, productIDs := seedTestBakery(ts)
	token := generateTestToken(t, "user-cancel-flow")

	// Step 1: Create order
	orderReq := dto.CreateOrderRequest{
		BakeryID: bakeryID,
		Items: []dto.OrderItemRequest{
			{ProductID: productIDs[0], Quantity: 1},
		},
		ScheduledDay: domain.Wednesday,
		ScheduledTime: dto.TimeSlotRequest{
			StartTime: "09:00",
			EndTime:   "09:30",
		},
	}
	resp := doRequest(t, "POST", ts.server.URL+"/api/orders", token, orderReq)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var orderResp dto.OrderResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&orderResp))
	resp.Body.Close()
	assert.Equal(t, "pending_payment", orderResp.Status)

	// Step 2: Confirm payment
	callbackReq := dto.PaymentCallbackRequest{
		OrderID:    orderResp.ID,
		PaymentRef: "pay-ref-cancel-test",
		Status:     "success",
	}
	resp = doRequest(t, "POST", ts.server.URL+"/api/payments/callback", token, callbackReq)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Step 3: Cancel (delete) the confirmed order
	resp = doRequest(t, "DELETE", ts.server.URL+"/api/orders/"+orderResp.ID, token, nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Step 4: Verify order is cancelled by listing orders
	resp = doRequest(t, "GET", ts.server.URL+"/api/orders", token, nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var orderList dto.ListResponse[dto.OrderResponse]
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&orderList))
	resp.Body.Close()

	var cancelledOrder *dto.OrderResponse
	for i, o := range orderList.Items {
		if o.ID == orderResp.ID {
			cancelledOrder = &orderList.Items[i]
			break
		}
	}
	require.NotNil(t, cancelledOrder, "order should still appear in list after cancellation")
	assert.Equal(t, "cancelled", cancelledOrder.Status)
}

// TestIntegration_UnauthorizedAccess tests that requests without valid tokens are rejected.
func TestIntegration_UnauthorizedAccess(t *testing.T) {
	ts := setupTestServer(t)

	t.Run("no authorization header", func(t *testing.T) {
		// Only order/reservation/payment endpoints require auth
		endpoints := []struct {
			method string
			path   string
		}{
			{"POST", "/api/orders"},
			{"GET", "/api/orders"},
			{"DELETE", "/api/orders/some-id"},
			{"POST", "/api/reservations"},
			{"DELETE", "/api/reservations/some-id"},
			{"POST", "/api/payments/callback"},
		}

		for _, ep := range endpoints {
			resp := doRequest(t, ep.method, ts.server.URL+ep.path, "", nil)
			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "endpoint %s %s should require auth", ep.method, ep.path)
			resp.Body.Close()
		}

		// Bakery endpoints are public and should NOT require auth
		publicEndpoints := []struct {
			method string
			path   string
		}{
			{"GET", "/api/bakeries"},
			{"GET", "/api/bakeries/some-id/menu"},
		}

		for _, ep := range publicEndpoints {
			resp := doRequest(t, ep.method, ts.server.URL+ep.path, "", nil)
			assert.NotEqual(t, http.StatusUnauthorized, resp.StatusCode, "endpoint %s %s should be public", ep.method, ep.path)
			resp.Body.Close()
		}
	})

	t.Run("expired token", func(t *testing.T) {
		expiredToken := generateExpiredToken(t, "user-expired")

		resp := doRequest(t, "GET", ts.server.URL+"/api/orders", expiredToken, nil)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

		var errResp dto.ErrorResponse
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&errResp))
		resp.Body.Close()
		assert.Equal(t, "TOKEN_EXPIRED", errResp.Code)
	})

	t.Run("invalid token signature", func(t *testing.T) {
		// Sign with a different secret
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub": "user-tampered",
			"exp": time.Now().Add(1 * time.Hour).Unix(),
		})
		signed, err := token.SignedString([]byte("wrong-secret"))
		require.NoError(t, err)

		resp := doRequest(t, "GET", ts.server.URL+"/api/orders", signed, nil)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

		var errResp dto.ErrorResponse
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&errResp))
		resp.Body.Close()
		assert.Equal(t, "INVALID_TOKEN", errResp.Code)
	})
}
