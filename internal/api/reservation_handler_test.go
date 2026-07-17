package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lucatorrekens/bakery-app/internal/api/dto"
	"github.com/lucatorrekens/bakery-app/internal/domain"
	"github.com/lucatorrekens/bakery-app/internal/repository/memory"
	"github.com/lucatorrekens/bakery-app/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupReservationHandler(bakeries []domain.Bakery, products []domain.Product) (*ReservationHandler, chi.Router) {
	bakeryRepo := memory.NewBakeryRepo()
	for _, b := range bakeries {
		bakeryRepo.SeedBakery(b)
	}
	for _, p := range products {
		bakeryRepo.SeedProduct(p)
	}

	reservationRepo := memory.NewReservationRepo()

	svc := service.NewReservationService(service.ReservationServiceConfig{
		BakeryRepo:      bakeryRepo,
		ReservationRepo: reservationRepo,
		IDGen:           func() string { return "test-reservation-id" },
		Now:             func() time.Time { return time.Date(2025, 1, 8, 10, 0, 0, 0, time.UTC) },
	})

	handler := NewReservationHandler(svc)
	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	return handler, r
}

func TestCreateReservation_Success(t *testing.T) {
	bakeries := []domain.Bakery{sampleBakery("b1")}
	products := []domain.Product{
		{ID: "p1", BakeryID: "b1", Name: "Croissant", Category: "Pastries", Price: 350, IsAvailable: true},
		{ID: "p2", BakeryID: "b1", Name: "Baguette", Category: "Breads", Price: 500, IsAvailable: true},
	}
	_, router := setupReservationHandler(bakeries, products)

	body := dto.CreateReservationRequest{
		BakeryID: "b1",
		Items: []dto.OrderItemRequest{
			{ProductID: "p1", Quantity: 2},
			{ProductID: "p2", Quantity: 1},
		},
		ScheduledDay: domain.Wednesday,
		ScheduledTime: dto.TimeSlotRequest{
			StartTime: "10:00",
			EndTime:   "10:30",
		},
	}

	jsonBody, err := json.Marshal(body)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/reservations", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "user-123")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp dto.ReservationResponse
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)

	// Verify ID was assigned
	assert.Equal(t, "test-reservation-id", resp.ID)
	assert.Equal(t, "b1", resp.BakeryID)

	// Verify payment method is always OnSpot
	assert.Equal(t, "on_spot", resp.PaymentMethod)

	// Verify status is always Confirmed
	assert.Equal(t, "confirmed", resp.Status)

	// Verify items were enriched with product info
	require.Len(t, resp.Items, 2)
	assert.Equal(t, "Croissant", resp.Items[0].ProductName)
	assert.Equal(t, int64(350), resp.Items[0].UnitPrice)
	assert.Equal(t, 2, resp.Items[0].Quantity)
	assert.Equal(t, int64(700), resp.Items[0].Subtotal)

	assert.Equal(t, "Baguette", resp.Items[1].ProductName)
	assert.Equal(t, int64(500), resp.Items[1].UnitPrice)
	assert.Equal(t, 1, resp.Items[1].Quantity)
	assert.Equal(t, int64(500), resp.Items[1].Subtotal)

	// Verify total = 2*350 + 1*500 = 1200
	assert.Equal(t, int64(1200), resp.TotalAmount)

	// Verify schedule
	assert.Equal(t, "wednesday", resp.ScheduledDay)
	assert.Equal(t, "10:00", resp.ScheduledTime.StartTime)
	assert.Equal(t, "10:30", resp.ScheduledTime.EndTime)
}

func TestCreateReservation_EmptyItems(t *testing.T) {
	bakeries := []domain.Bakery{sampleBakery("b1")}
	_, router := setupReservationHandler(bakeries, nil)

	body := dto.CreateReservationRequest{
		BakeryID:     "b1",
		Items:        []dto.OrderItemRequest{},
		ScheduledDay: domain.Wednesday,
		ScheduledTime: dto.TimeSlotRequest{
			StartTime: "10:00",
			EndTime:   "10:30",
		},
	}

	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/reservations", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var resp dto.ValidationErrorResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
	require.Len(t, resp.Errors, 1)
	assert.Equal(t, "items", resp.Errors[0].Field)
}

func TestCreateReservation_InvalidQuantity(t *testing.T) {
	bakeries := []domain.Bakery{sampleBakery("b1")}
	products := []domain.Product{
		{ID: "p1", BakeryID: "b1", Name: "Croissant", Category: "Pastries", Price: 350, IsAvailable: true},
	}
	_, router := setupReservationHandler(bakeries, products)

	body := dto.CreateReservationRequest{
		BakeryID: "b1",
		Items: []dto.OrderItemRequest{
			{ProductID: "p1", Quantity: 0}, // invalid
		},
		ScheduledDay: domain.Wednesday,
		ScheduledTime: dto.TimeSlotRequest{
			StartTime: "10:00",
			EndTime:   "10:30",
		},
	}

	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/reservations", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var resp dto.ValidationErrorResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
}

func TestCreateReservation_BakeryClosed(t *testing.T) {
	bakeries := []domain.Bakery{sampleBakery("b1")}
	products := []domain.Product{
		{ID: "p1", BakeryID: "b1", Name: "Croissant", Category: "Pastries", Price: 350, IsAvailable: true},
	}
	_, router := setupReservationHandler(bakeries, products)

	body := dto.CreateReservationRequest{
		BakeryID: "b1",
		Items: []dto.OrderItemRequest{
			{ProductID: "p1", Quantity: 1},
		},
		ScheduledDay: domain.Sunday, // bakery is closed on Sunday
		ScheduledTime: dto.TimeSlotRequest{
			StartTime: "10:00",
			EndTime:   "10:30",
		},
	}

	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/reservations", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var resp dto.ValidationErrorResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Contains(t, resp.Errors[0].Message, "closed")
}

func TestCreateReservation_TimeOutsideHours(t *testing.T) {
	bakeries := []domain.Bakery{sampleBakery("b1")}
	products := []domain.Product{
		{ID: "p1", BakeryID: "b1", Name: "Croissant", Category: "Pastries", Price: 350, IsAvailable: true},
	}
	_, router := setupReservationHandler(bakeries, products)

	// Wednesday schedule is 09:30-17:00, requesting 07:00-07:30
	body := dto.CreateReservationRequest{
		BakeryID: "b1",
		Items: []dto.OrderItemRequest{
			{ProductID: "p1", Quantity: 1},
		},
		ScheduledDay: domain.Wednesday,
		ScheduledTime: dto.TimeSlotRequest{
			StartTime: "07:00",
			EndTime:   "07:30",
		},
	}

	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/reservations", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestCreateReservation_ProductUnavailable(t *testing.T) {
	bakeries := []domain.Bakery{sampleBakery("b1")}
	products := []domain.Product{
		{ID: "p1", BakeryID: "b1", Name: "Croissant", Category: "Pastries", Price: 350, IsAvailable: false}, // unavailable
	}
	_, router := setupReservationHandler(bakeries, products)

	body := dto.CreateReservationRequest{
		BakeryID: "b1",
		Items: []dto.OrderItemRequest{
			{ProductID: "p1", Quantity: 1},
		},
		ScheduledDay: domain.Wednesday,
		ScheduledTime: dto.TimeSlotRequest{
			StartTime: "10:00",
			EndTime:   "10:30",
		},
	}

	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/reservations", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var resp dto.ValidationErrorResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Contains(t, resp.Errors[0].Message, "not available")
}

func TestCreateReservation_BakeryNotFound(t *testing.T) {
	// No bakeries seeded
	_, router := setupReservationHandler(nil, nil)

	body := dto.CreateReservationRequest{
		BakeryID: "nonexistent",
		Items: []dto.OrderItemRequest{
			{ProductID: "p1", Quantity: 1},
		},
		ScheduledDay: domain.Wednesday,
		ScheduledTime: dto.TimeSlotRequest{
			StartTime: "10:00",
			EndTime:   "10:30",
		},
	}

	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/reservations", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp dto.ErrorResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "BAKERY_NOT_FOUND", resp.Code)
}

func TestCreateReservation_InvalidJSON(t *testing.T) {
	bakeries := []domain.Bakery{sampleBakery("b1")}
	_, router := setupReservationHandler(bakeries, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/reservations", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp dto.ErrorResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "INVALID_REQUEST", resp.Code)
}

func TestCreateReservation_InvalidTimeFormat(t *testing.T) {
	bakeries := []domain.Bakery{sampleBakery("b1")}
	products := []domain.Product{
		{ID: "p1", BakeryID: "b1", Name: "Croissant", Category: "Pastries", Price: 350, IsAvailable: true},
	}
	_, router := setupReservationHandler(bakeries, products)

	body := dto.CreateReservationRequest{
		BakeryID: "b1",
		Items: []dto.OrderItemRequest{
			{ProductID: "p1", Quantity: 1},
		},
		ScheduledDay: domain.Wednesday,
		ScheduledTime: dto.TimeSlotRequest{
			StartTime: "invalid",
			EndTime:   "10:30",
		},
	}

	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/reservations", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// --- Tests for DELETE /api/reservations/:id ---

func setupReservationManagementHandler(reservations []domain.Reservation) chi.Router {
	bakeryRepo := memory.NewBakeryRepo()
	bakeryRepo.SeedBakery(sampleBakery("b1"))

	reservationRepo := memory.NewReservationRepo()
	for _, r := range reservations {
		reservationRepo.Save(nil, r)
	}

	svc := service.NewReservationService(service.ReservationServiceConfig{
		BakeryRepo:      bakeryRepo,
		ReservationRepo: reservationRepo,
		Now:             func() time.Time { return time.Date(2025, 1, 8, 10, 0, 0, 0, time.UTC) },
	})

	handler := NewReservationHandler(svc)
	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	return r
}

func sampleReservations() []domain.Reservation {
	return []domain.Reservation{
		{
			ID:            "res-1",
			BakeryID:      "b1",
			UserID:        "user-123",
			Items:         []domain.OrderItem{{ProductID: "p1", ProductName: "Croissant", Quantity: 2, UnitPrice: 350, Subtotal: 700}},
			ScheduledDay:  domain.Monday,
			ScheduledTime: domain.TimeSlot{StartTime: domain.TimeOfDay{Hour: 10, Minute: 0}, EndTime: domain.TimeOfDay{Hour: 10, Minute: 30}},
			Status:        domain.ReservationStatusConfirmed,
			TotalAmount:   700,
			PaymentMethod: domain.PaymentMethodOnSpot,
			CreatedAt:     time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC),
		},
		{
			ID:            "res-2",
			BakeryID:      "b1",
			UserID:        "user-123",
			Items:         []domain.OrderItem{{ProductID: "p1", ProductName: "Croissant", Quantity: 1, UnitPrice: 350, Subtotal: 350}},
			ScheduledDay:  domain.Tuesday,
			ScheduledTime: domain.TimeSlot{StartTime: domain.TimeOfDay{Hour: 11, Minute: 0}, EndTime: domain.TimeOfDay{Hour: 11, Minute: 30}},
			Status:        domain.ReservationStatusPickedUp,
			TotalAmount:   350,
			PaymentMethod: domain.PaymentMethodOnSpot,
			CreatedAt:     time.Date(2025, 1, 2, 10, 0, 0, 0, time.UTC),
		},
		{
			ID:            "res-3",
			BakeryID:      "b1",
			UserID:        "user-123",
			Items:         []domain.OrderItem{{ProductID: "p1", ProductName: "Croissant", Quantity: 3, UnitPrice: 350, Subtotal: 1050}},
			ScheduledDay:  domain.Wednesday,
			ScheduledTime: domain.TimeSlot{StartTime: domain.TimeOfDay{Hour: 12, Minute: 0}, EndTime: domain.TimeOfDay{Hour: 12, Minute: 30}},
			Status:        domain.ReservationStatusReady,
			TotalAmount:   1050,
			PaymentMethod: domain.PaymentMethodOnSpot,
			CreatedAt:     time.Date(2025, 1, 3, 10, 0, 0, 0, time.UTC),
		},
	}
}

func TestDeleteReservation_Success_Confirmed(t *testing.T) {
	router := setupReservationManagementHandler(sampleReservations())

	req := httptest.NewRequest(http.MethodDelete, "/api/reservations/res-1", nil)
	req.Header.Set("X-User-ID", "user-123")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]string
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "reservation cancelled successfully", resp["message"])
}

func TestDeleteReservation_Success_Ready(t *testing.T) {
	router := setupReservationManagementHandler(sampleReservations())

	req := httptest.NewRequest(http.MethodDelete, "/api/reservations/res-3", nil)
	req.Header.Set("X-User-ID", "user-123")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDeleteReservation_NotFound(t *testing.T) {
	router := setupReservationManagementHandler(nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/reservations/nonexistent", nil)
	req.Header.Set("X-User-ID", "user-123")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp dto.ErrorResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "RESERVATION_NOT_FOUND", resp.Code)
}

func TestDeleteReservation_Forbidden(t *testing.T) {
	router := setupReservationManagementHandler(sampleReservations())

	req := httptest.NewRequest(http.MethodDelete, "/api/reservations/res-1", nil)
	req.Header.Set("X-User-ID", "user-other") // different user
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var resp dto.ErrorResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "FORBIDDEN", resp.Code)
}

func TestDeleteReservation_PickedUpNotCancellable(t *testing.T) {
	router := setupReservationManagementHandler(sampleReservations())

	req := httptest.NewRequest(http.MethodDelete, "/api/reservations/res-2", nil) // picked_up
	req.Header.Set("X-User-ID", "user-123")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)

	var resp dto.ErrorResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "RESERVATION_NOT_CANCELLABLE", resp.Code)
}
