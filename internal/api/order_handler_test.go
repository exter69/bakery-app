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
	"github.com/lucatorrekens/bakery-app/internal/payment"
	"github.com/lucatorrekens/bakery-app/internal/repository/memory"
	"github.com/lucatorrekens/bakery-app/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupOrderTestRouter(bakeries []domain.Bakery, products []domain.Product) chi.Router {
	bakeryRepo := memory.NewBakeryRepo()
	for _, b := range bakeries {
		bakeryRepo.SeedBakery(b)
	}
	for _, p := range products {
		bakeryRepo.SeedProduct(p)
	}

	orderRepo := memory.NewOrderRepo()
	paymentSvc := payment.NewMockPaymentService()

	idCounter := 0
	orderSvc := service.NewOrderService(service.OrderServiceConfig{
		OrderRepo:  orderRepo,
		BakeryRepo: bakeryRepo,
		PaymentSvc: paymentSvc,
		IDGen: func() string {
			idCounter++
			return "test-order-" + string(rune('0'+idCounter))
		},
		Now: func() time.Time {
			return time.Date(2025, 1, 8, 10, 0, 0, 0, time.UTC)
		},
	})

	handler := NewOrderHandler(orderSvc)
	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	return r
}

func testBakery() domain.Bakery {
	return domain.Bakery{
		ID:       "bakery-1",
		Name:     "Test Bakery",
		PhotoURL: "https://example.com/bakery.jpg",
		Schedule: []domain.DaySchedule{
			{Day: domain.Monday, OpenTime: domain.TimeOfDay{Hour: 8, Minute: 0}, CloseTime: domain.TimeOfDay{Hour: 18, Minute: 0}, IsOpen: true},
			{Day: domain.Tuesday, OpenTime: domain.TimeOfDay{Hour: 8, Minute: 0}, CloseTime: domain.TimeOfDay{Hour: 18, Minute: 0}, IsOpen: true},
			{Day: domain.Wednesday, OpenTime: domain.TimeOfDay{Hour: 9, Minute: 0}, CloseTime: domain.TimeOfDay{Hour: 17, Minute: 0}, IsOpen: true},
			{Day: domain.Thursday, OpenTime: domain.TimeOfDay{Hour: 8, Minute: 0}, CloseTime: domain.TimeOfDay{Hour: 18, Minute: 0}, IsOpen: true},
			{Day: domain.Friday, OpenTime: domain.TimeOfDay{Hour: 8, Minute: 0}, CloseTime: domain.TimeOfDay{Hour: 20, Minute: 0}, IsOpen: true},
			{Day: domain.Saturday, OpenTime: domain.TimeOfDay{Hour: 10, Minute: 0}, CloseTime: domain.TimeOfDay{Hour: 16, Minute: 0}, IsOpen: true},
			{Day: domain.Sunday, IsOpen: false},
		},
	}
}

func testProducts() []domain.Product {
	return []domain.Product{
		{ID: "prod-1", BakeryID: "bakery-1", Name: "Croissant", Price: 350, Category: "pastries", IsAvailable: true},
		{ID: "prod-2", BakeryID: "bakery-1", Name: "Baguette", Price: 500, Category: "bread", IsAvailable: true},
		{ID: "prod-3", BakeryID: "bakery-1", Name: "Eclair", Price: 400, Category: "pastries", IsAvailable: false},
	}
}

func TestCreateOrder_Success(t *testing.T) {
	router := setupOrderTestRouter([]domain.Bakery{testBakery()}, testProducts())

	reqBody := dto.CreateOrderRequest{
		BakeryID: "bakery-1",
		Items: []dto.OrderItemRequest{
			{ProductID: "prod-1", Quantity: 3},
			{ProductID: "prod-2", Quantity: 1},
		},
		ScheduledDay: domain.Wednesday,
		ScheduledTime: dto.TimeSlotRequest{
			StartTime: "10:00",
			EndTime:   "10:30",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "user-123")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp dto.OrderResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)

	// Verify order fields
	assert.NotEmpty(t, resp.ID)
	assert.Equal(t, "bakery-1", resp.BakeryID)
	assert.Equal(t, "pending_payment", resp.Status)
	assert.Equal(t, "online", resp.PaymentMethod)
	assert.Equal(t, "wednesday", resp.ScheduledDay)
	assert.Equal(t, "10:00", resp.ScheduledTime.StartTime)
	assert.Equal(t, "10:30", resp.ScheduledTime.EndTime)

	// Verify items: prod-1 (350 * 3 = 1050), prod-2 (500 * 1 = 500)
	require.Len(t, resp.Items, 2)
	assert.Equal(t, "prod-1", resp.Items[0].ProductID)
	assert.Equal(t, "Croissant", resp.Items[0].ProductName)
	assert.Equal(t, 3, resp.Items[0].Quantity)
	assert.Equal(t, int64(350), resp.Items[0].UnitPrice)
	assert.Equal(t, int64(1050), resp.Items[0].Subtotal)

	assert.Equal(t, "prod-2", resp.Items[1].ProductID)
	assert.Equal(t, "Baguette", resp.Items[1].ProductName)
	assert.Equal(t, 1, resp.Items[1].Quantity)
	assert.Equal(t, int64(500), resp.Items[1].UnitPrice)
	assert.Equal(t, int64(500), resp.Items[1].Subtotal)

	// Verify total: 1050 + 500 = 1550
	assert.Equal(t, int64(1550), resp.TotalAmount)

	// Verify payment link
	assert.NotEmpty(t, resp.PaymentLink)
	assert.Contains(t, resp.PaymentLink, "pay.example.com")
}

func TestCreateOrder_EmptyItems_ReturnsValidationError(t *testing.T) {
	router := setupOrderTestRouter([]domain.Bakery{testBakery()}, testProducts())

	reqBody := dto.CreateOrderRequest{
		BakeryID: "bakery-1",
		Items:    []dto.OrderItemRequest{},
		ScheduledDay: domain.Wednesday,
		ScheduledTime: dto.TimeSlotRequest{
			StartTime: "10:00",
			EndTime:   "10:30",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "user-123")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var resp dto.ValidationErrorResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)

	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
	require.Len(t, resp.Errors, 1)
	assert.Equal(t, "items", resp.Errors[0].Field)
	assert.Contains(t, resp.Errors[0].Message, "at least one item is required")
}

func TestCreateOrder_BakeryClosedOnDay_ReturnsValidationError(t *testing.T) {
	router := setupOrderTestRouter([]domain.Bakery{testBakery()}, testProducts())

	reqBody := dto.CreateOrderRequest{
		BakeryID: "bakery-1",
		Items: []dto.OrderItemRequest{
			{ProductID: "prod-1", Quantity: 1},
		},
		ScheduledDay: domain.Sunday, // bakery is closed on Sunday
		ScheduledTime: dto.TimeSlotRequest{
			StartTime: "10:00",
			EndTime:   "10:30",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "user-123")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var resp dto.ValidationErrorResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)

	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
	assert.True(t, len(resp.Errors) > 0)
	assert.Equal(t, "scheduledDay", resp.Errors[0].Field)
}

func TestCreateOrder_UnavailableProduct_ReturnsValidationError(t *testing.T) {
	router := setupOrderTestRouter([]domain.Bakery{testBakery()}, testProducts())

	reqBody := dto.CreateOrderRequest{
		BakeryID: "bakery-1",
		Items: []dto.OrderItemRequest{
			{ProductID: "prod-3", Quantity: 1}, // prod-3 is unavailable
		},
		ScheduledDay: domain.Wednesday,
		ScheduledTime: dto.TimeSlotRequest{
			StartTime: "10:00",
			EndTime:   "10:30",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "user-123")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var resp dto.ValidationErrorResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)

	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
	assert.True(t, len(resp.Errors) > 0)
}

func TestCreateOrder_BakeryNotFound_Returns404(t *testing.T) {
	router := setupOrderTestRouter([]domain.Bakery{testBakery()}, testProducts())

	reqBody := dto.CreateOrderRequest{
		BakeryID: "nonexistent-bakery",
		Items: []dto.OrderItemRequest{
			{ProductID: "prod-1", Quantity: 1},
		},
		ScheduledDay: domain.Wednesday,
		ScheduledTime: dto.TimeSlotRequest{
			StartTime: "10:00",
			EndTime:   "10:30",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "user-123")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp dto.ErrorResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "BAKERY_NOT_FOUND", resp.Code)
}

func TestCreateOrder_InvalidJSON_Returns400(t *testing.T) {
	router := setupOrderTestRouter([]domain.Bakery{testBakery()}, testProducts())

	req := httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp dto.ErrorResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "INVALID_REQUEST", resp.Code)
}

func TestCreateOrder_InvalidTimeFormat_Returns400(t *testing.T) {
	router := setupOrderTestRouter([]domain.Bakery{testBakery()}, testProducts())

	reqBody := dto.CreateOrderRequest{
		BakeryID: "bakery-1",
		Items: []dto.OrderItemRequest{
			{ProductID: "prod-1", Quantity: 1},
		},
		ScheduledDay: domain.Wednesday,
		ScheduledTime: dto.TimeSlotRequest{
			StartTime: "invalid",
			EndTime:   "10:30",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "user-123")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// --- Tests for GET /api/orders ---

func setupOrderManagementRouter(orders []domain.Order) chi.Router {
	bakeryRepo := memory.NewBakeryRepo()
	bakeryRepo.SeedBakery(testBakery())
	for _, p := range testProducts() {
		bakeryRepo.SeedProduct(p)
	}

	orderRepo := memory.NewOrderRepo()
	for i := range orders {
		orderRepo.Save(nil, &orders[i])
	}

	paymentSvc := payment.NewMockPaymentService()

	orderSvc := service.NewOrderService(service.OrderServiceConfig{
		OrderRepo:  orderRepo,
		BakeryRepo: bakeryRepo,
		PaymentSvc: paymentSvc,
		Now:        func() time.Time { return time.Date(2025, 1, 8, 10, 0, 0, 0, time.UTC) },
	})

	handler := NewOrderHandler(orderSvc)
	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	return r
}

func sampleOrders(userID string) []domain.Order {
	return []domain.Order{
		{
			ID:       "order-1",
			BakeryID: "bakery-1",
			UserID:   userID,
			Items:    []domain.OrderItem{{ProductID: "prod-1", ProductName: "Croissant", Quantity: 2, UnitPrice: 350, Subtotal: 700}},
			ScheduledDay:  domain.Monday,
			ScheduledTime: domain.TimeSlot{StartTime: domain.TimeOfDay{Hour: 10, Minute: 0}, EndTime: domain.TimeOfDay{Hour: 10, Minute: 30}},
			Status:        domain.OrderStatusConfirmed,
			TotalAmount:   700,
			PaymentMethod: domain.PaymentMethodOnline,
			CreatedAt:     time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC),
			UpdatedAt:     time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC),
		},
		{
			ID:       "order-2",
			BakeryID: "bakery-1",
			UserID:   userID,
			Items:    []domain.OrderItem{{ProductID: "prod-2", ProductName: "Baguette", Quantity: 1, UnitPrice: 500, Subtotal: 500}},
			ScheduledDay:  domain.Tuesday,
			ScheduledTime: domain.TimeSlot{StartTime: domain.TimeOfDay{Hour: 14, Minute: 0}, EndTime: domain.TimeOfDay{Hour: 14, Minute: 30}},
			Status:        domain.OrderStatusPendingPayment,
			TotalAmount:   500,
			PaymentMethod: domain.PaymentMethodOnline,
			CreatedAt:     time.Date(2025, 1, 2, 10, 0, 0, 0, time.UTC),
			UpdatedAt:     time.Date(2025, 1, 2, 10, 0, 0, 0, time.UTC),
		},
		{
			ID:       "order-3",
			BakeryID: "bakery-1",
			UserID:   userID,
			Items:    []domain.OrderItem{{ProductID: "prod-1", ProductName: "Croissant", Quantity: 5, UnitPrice: 350, Subtotal: 1750}},
			ScheduledDay:  domain.Wednesday,
			ScheduledTime: domain.TimeSlot{StartTime: domain.TimeOfDay{Hour: 9, Minute: 30}, EndTime: domain.TimeOfDay{Hour: 10, Minute: 0}},
			Status:        domain.OrderStatusDelivered,
			TotalAmount:   1750,
			PaymentMethod: domain.PaymentMethodOnline,
			CreatedAt:     time.Date(2025, 1, 3, 10, 0, 0, 0, time.UTC),
			UpdatedAt:     time.Date(2025, 1, 3, 10, 0, 0, 0, time.UTC),
		},
	}
}

func TestListOrders_Success(t *testing.T) {
	orders := sampleOrders("user-123")
	router := setupOrderManagementRouter(orders)

	req := httptest.NewRequest(http.MethodGet, "/api/orders", nil)
	req.Header.Set("X-User-ID", "user-123")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp dto.ListResponse[dto.OrderResponse]
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)

	assert.Equal(t, 3, resp.Total)
	assert.Equal(t, 1, resp.Page)
	assert.Equal(t, 20, resp.PageSize)
	assert.Len(t, resp.Items, 3)
}

func TestListOrders_FilterByStatus(t *testing.T) {
	orders := sampleOrders("user-123")
	router := setupOrderManagementRouter(orders)

	req := httptest.NewRequest(http.MethodGet, "/api/orders?status=confirmed", nil)
	req.Header.Set("X-User-ID", "user-123")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp dto.ListResponse[dto.OrderResponse]
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)

	assert.Equal(t, 1, resp.Total)
	assert.Len(t, resp.Items, 1)
	assert.Equal(t, "confirmed", resp.Items[0].Status)
}

func TestListOrders_SortByScheduledTime(t *testing.T) {
	orders := sampleOrders("user-123")
	router := setupOrderManagementRouter(orders)

	req := httptest.NewRequest(http.MethodGet, "/api/orders?sortBy=scheduledTime&sortDir=asc", nil)
	req.Header.Set("X-User-ID", "user-123")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp dto.ListResponse[dto.OrderResponse]
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)

	require.Len(t, resp.Items, 3)
	// Sorted by scheduled start time ascending: 09:30, 10:00, 14:00
	assert.Equal(t, "order-3", resp.Items[0].ID) // 09:30
	assert.Equal(t, "order-1", resp.Items[1].ID) // 10:00
	assert.Equal(t, "order-2", resp.Items[2].ID) // 14:00
}

func TestListOrders_Pagination(t *testing.T) {
	orders := sampleOrders("user-123")
	router := setupOrderManagementRouter(orders)

	// Request page 2 - should be empty since all 3 orders fit on page 1 (pageSize=20)
	req := httptest.NewRequest(http.MethodGet, "/api/orders?page=2", nil)
	req.Header.Set("X-User-ID", "user-123")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp dto.ListResponse[dto.OrderResponse]
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)

	assert.Equal(t, 3, resp.Total)
	assert.Equal(t, 2, resp.Page)
	assert.Len(t, resp.Items, 0)
}

func TestListOrders_OnlyReturnsUserOrders(t *testing.T) {
	orders := sampleOrders("user-123")
	router := setupOrderManagementRouter(orders)

	// Different user should see no orders
	req := httptest.NewRequest(http.MethodGet, "/api/orders", nil)
	req.Header.Set("X-User-ID", "user-999")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp dto.ListResponse[dto.OrderResponse]
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)

	assert.Equal(t, 0, resp.Total)
	assert.Len(t, resp.Items, 0)
}

func TestListOrders_InvalidPage(t *testing.T) {
	router := setupOrderManagementRouter(nil)

	req := httptest.NewRequest(http.MethodGet, "/api/orders?page=abc", nil)
	req.Header.Set("X-User-ID", "user-123")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// --- Tests for DELETE /api/orders/:id ---

func TestDeleteOrder_Success_PendingPayment(t *testing.T) {
	orders := sampleOrders("user-123")
	router := setupOrderManagementRouter(orders)

	req := httptest.NewRequest(http.MethodDelete, "/api/orders/order-2", nil)
	req.Header.Set("X-User-ID", "user-123")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]string
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "order cancelled successfully", resp["message"])
}

func TestDeleteOrder_Success_Confirmed(t *testing.T) {
	orders := sampleOrders("user-123")
	router := setupOrderManagementRouter(orders)

	req := httptest.NewRequest(http.MethodDelete, "/api/orders/order-1", nil)
	req.Header.Set("X-User-ID", "user-123")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDeleteOrder_NotFound(t *testing.T) {
	router := setupOrderManagementRouter(nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/orders/nonexistent", nil)
	req.Header.Set("X-User-ID", "user-123")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp dto.ErrorResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "ORDER_NOT_FOUND", resp.Code)
}

func TestDeleteOrder_Forbidden(t *testing.T) {
	orders := sampleOrders("user-123")
	router := setupOrderManagementRouter(orders)

	req := httptest.NewRequest(http.MethodDelete, "/api/orders/order-1", nil)
	req.Header.Set("X-User-ID", "user-other") // different user
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var resp dto.ErrorResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "FORBIDDEN", resp.Code)
}

func TestDeleteOrder_DeliveredNotCancellable(t *testing.T) {
	orders := sampleOrders("user-123")
	router := setupOrderManagementRouter(orders)

	req := httptest.NewRequest(http.MethodDelete, "/api/orders/order-3", nil) // delivered
	req.Header.Set("X-User-ID", "user-123")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)

	var resp dto.ErrorResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "ORDER_NOT_CANCELLABLE", resp.Code)
}
