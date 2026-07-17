package api

import (
	"encoding/json"
	"fmt"
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

func fixedWednesday() time.Time {
	return time.Date(2025, 1, 8, 10, 0, 0, 0, time.UTC) // Wednesday
}

func fixedSunday() time.Time {
	return time.Date(2025, 1, 12, 10, 0, 0, 0, time.UTC) // Sunday
}

func sampleBakery(id string) domain.Bakery {
	return domain.Bakery{
		ID:       id,
		Name:     "Bakery " + id,
		PhotoURL: fmt.Sprintf("https://example.com/%s.jpg", id),
		Schedule: []domain.DaySchedule{
			{Day: domain.Monday, OpenTime: domain.TimeOfDay{Hour: 8, Minute: 0}, CloseTime: domain.TimeOfDay{Hour: 18, Minute: 0}, IsOpen: true},
			{Day: domain.Tuesday, OpenTime: domain.TimeOfDay{Hour: 8, Minute: 0}, CloseTime: domain.TimeOfDay{Hour: 18, Minute: 0}, IsOpen: true},
			{Day: domain.Wednesday, OpenTime: domain.TimeOfDay{Hour: 9, Minute: 30}, CloseTime: domain.TimeOfDay{Hour: 17, Minute: 0}, IsOpen: true},
			{Day: domain.Thursday, OpenTime: domain.TimeOfDay{Hour: 8, Minute: 0}, CloseTime: domain.TimeOfDay{Hour: 18, Minute: 0}, IsOpen: true},
			{Day: domain.Friday, OpenTime: domain.TimeOfDay{Hour: 8, Minute: 0}, CloseTime: domain.TimeOfDay{Hour: 20, Minute: 0}, IsOpen: true},
			{Day: domain.Saturday, OpenTime: domain.TimeOfDay{Hour: 10, Minute: 0}, CloseTime: domain.TimeOfDay{Hour: 16, Minute: 0}, IsOpen: true},
			{Day: domain.Sunday, IsOpen: false},
		},
	}
}

func setupTestHandler(bakeries []domain.Bakery, now func() time.Time) (*BakeryHandler, chi.Router) {
	return setupTestHandlerWithProducts(bakeries, nil, now)
}

func setupTestHandlerWithProducts(bakeries []domain.Bakery, products []domain.Product, now func() time.Time) (*BakeryHandler, chi.Router) {
	repo := memory.NewBakeryRepo()
	for _, b := range bakeries {
		repo.SeedBakery(b)
	}
	for _, p := range products {
		repo.SeedProduct(p)
	}
	svc := service.NewBakeryService(repo)
	handler := NewBakeryHandlerWithClock(svc, now)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	return handler, r
}

func TestListBakeries_ReturnsCardsWithTodaySchedule(t *testing.T) {
	bakeries := []domain.Bakery{
		sampleBakery("b1"),
		sampleBakery("b2"),
		sampleBakery("b3"),
	}
	_, router := setupTestHandler(bakeries, fixedWednesday)

	req := httptest.NewRequest(http.MethodGet, "/api/bakeries", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp dto.ListResponse[dto.BakeryCardResponse]
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)

	assert.Equal(t, 3, resp.Total)
	assert.Equal(t, 1, resp.Page)
	assert.Equal(t, 50, resp.PageSize)
	assert.Len(t, resp.Items, 3)

	// All bakeries are open on Wednesday 09:30-17:00
	for _, card := range resp.Items {
		assert.True(t, card.TodaySchedule.IsOpen)
		assert.Equal(t, "09:30", card.TodaySchedule.OpenTime)
		assert.Equal(t, "17:00", card.TodaySchedule.CloseTime)
		assert.NotEmpty(t, card.PhotoURL)
		assert.NotEmpty(t, card.Name)
	}
}

func TestListBakeries_EmptyList(t *testing.T) {
	_, router := setupTestHandler([]domain.Bakery{}, fixedWednesday)

	req := httptest.NewRequest(http.MethodGet, "/api/bakeries", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp dto.ListResponse[dto.BakeryCardResponse]
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)

	assert.Equal(t, 0, resp.Total)
	assert.Empty(t, resp.Items)
}

func TestListBakeries_PaginationAt50(t *testing.T) {
	bakeries := make([]domain.Bakery, 55)
	for i := range bakeries {
		bakeries[i] = sampleBakery(fmt.Sprintf("b%d", i+1))
	}
	_, router := setupTestHandler(bakeries, fixedWednesday)

	// Page 1 should return 50 items
	req := httptest.NewRequest(http.MethodGet, "/api/bakeries?page=1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var resp dto.ListResponse[dto.BakeryCardResponse]
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)

	assert.Equal(t, 55, resp.Total)
	assert.Len(t, resp.Items, 50)
	assert.Equal(t, 1, resp.Page)
	assert.Equal(t, 50, resp.PageSize)

	// Page 2 should return the remaining 5 items
	req = httptest.NewRequest(http.MethodGet, "/api/bakeries?page=2", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)

	assert.Equal(t, 55, resp.Total)
	assert.Len(t, resp.Items, 5)
	assert.Equal(t, 2, resp.Page)
}

func TestListBakeries_InvalidPageParam(t *testing.T) {
	_, router := setupTestHandler([]domain.Bakery{sampleBakery("b1")}, fixedWednesday)

	req := httptest.NewRequest(http.MethodGet, "/api/bakeries?page=abc", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errResp dto.ErrorResponse
	err := json.NewDecoder(w.Body).Decode(&errResp)
	require.NoError(t, err)
	assert.Equal(t, "INVALID_PAGE", errResp.Code)
}

func TestListBakeries_NegativePageParam(t *testing.T) {
	_, router := setupTestHandler([]domain.Bakery{sampleBakery("b1")}, fixedWednesday)

	req := httptest.NewRequest(http.MethodGet, "/api/bakeries?page=-1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestListBakeries_ClosedBakeryOnSunday(t *testing.T) {
	_, router := setupTestHandler([]domain.Bakery{sampleBakery("b1")}, fixedSunday)

	req := httptest.NewRequest(http.MethodGet, "/api/bakeries", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp dto.ListResponse[dto.BakeryCardResponse]
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)

	require.Len(t, resp.Items, 1)
	assert.False(t, resp.Items[0].TodaySchedule.IsOpen)
	assert.Empty(t, resp.Items[0].TodaySchedule.OpenTime)
	assert.Empty(t, resp.Items[0].TodaySchedule.CloseTime)
}

func TestListBakeries_PageBeyondTotal(t *testing.T) {
	_, router := setupTestHandler([]domain.Bakery{sampleBakery("b1")}, fixedWednesday)

	req := httptest.NewRequest(http.MethodGet, "/api/bakeries?page=100", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp dto.ListResponse[dto.BakeryCardResponse]
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)

	assert.Equal(t, 1, resp.Total)
	assert.Empty(t, resp.Items)
}

func TestListBakeries_DefaultsToPage1WhenNoParam(t *testing.T) {
	_, router := setupTestHandler([]domain.Bakery{sampleBakery("b1")}, fixedWednesday)

	req := httptest.NewRequest(http.MethodGet, "/api/bakeries", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var resp dto.ListResponse[dto.BakeryCardResponse]
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)

	assert.Equal(t, 1, resp.Page)
}

func TestGetMenu_MultipleCategories(t *testing.T) {
	bakeries := []domain.Bakery{sampleBakery("b1")}
	products := []domain.Product{
		{ID: "p1", BakeryID: "b1", Name: "Croissant", Category: "Pastries", Price: 350, IsAvailable: true},
		{ID: "p2", BakeryID: "b1", Name: "Baguette", Category: "Breads", Price: 500, IsAvailable: true},
		{ID: "p3", BakeryID: "b1", Name: "Pain au chocolat", Category: "Pastries", Price: 400, IsAvailable: true},
		{ID: "p4", BakeryID: "b1", Name: "Espresso", Category: "Drinks", Price: 250, IsAvailable: true},
	}
	_, router := setupTestHandlerWithProducts(bakeries, products, fixedWednesday)

	req := httptest.NewRequest(http.MethodGet, "/api/bakeries/b1/menu", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var menu map[string][]domain.Product
	err := json.NewDecoder(w.Body).Decode(&menu)
	require.NoError(t, err)

	// Should have 3 categories
	assert.Len(t, menu, 3)
	assert.Contains(t, menu, "Pastries")
	assert.Contains(t, menu, "Breads")
	assert.Contains(t, menu, "Drinks")

	// Pastries should contain 2 products
	assert.Len(t, menu["Pastries"], 2)
	// Breads should contain 1 product
	assert.Len(t, menu["Breads"], 1)
	// Drinks should contain 1 product
	assert.Len(t, menu["Drinks"], 1)
}

func TestGetMenu_ProductsInCorrectCategories(t *testing.T) {
	bakeries := []domain.Bakery{sampleBakery("b1")}
	products := []domain.Product{
		{ID: "p1", BakeryID: "b1", Name: "Croissant", Category: "Pastries", Price: 350, IsAvailable: true},
		{ID: "p2", BakeryID: "b1", Name: "Baguette", Category: "Breads", Price: 500, IsAvailable: true},
		{ID: "p3", BakeryID: "b1", Name: "Sourdough", Category: "Breads", Price: 600, IsAvailable: true},
	}
	_, router := setupTestHandlerWithProducts(bakeries, products, fixedWednesday)

	req := httptest.NewRequest(http.MethodGet, "/api/bakeries/b1/menu", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var menu map[string][]domain.Product
	err := json.NewDecoder(w.Body).Decode(&menu)
	require.NoError(t, err)

	// Verify Breads contains the correct products
	breadNames := make([]string, 0, len(menu["Breads"]))
	for _, p := range menu["Breads"] {
		breadNames = append(breadNames, p.Name)
	}
	assert.ElementsMatch(t, []string{"Baguette", "Sourdough"}, breadNames)

	// Verify Pastries contains the correct product
	require.Len(t, menu["Pastries"], 1)
	assert.Equal(t, "Croissant", menu["Pastries"][0].Name)
}

func TestGetMenu_EmptyMenu(t *testing.T) {
	bakeries := []domain.Bakery{sampleBakery("b1")}
	// No products seeded for bakery b1
	_, router := setupTestHandlerWithProducts(bakeries, nil, fixedWednesday)

	req := httptest.NewRequest(http.MethodGet, "/api/bakeries/b1/menu", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var menu map[string][]domain.Product
	err := json.NewDecoder(w.Body).Decode(&menu)
	require.NoError(t, err)

	// Empty bakery menu returns empty map
	assert.Empty(t, menu)
}

func TestGetMenu_BakeryNotFound(t *testing.T) {
	// No bakeries seeded at all
	_, router := setupTestHandlerWithProducts([]domain.Bakery{}, nil, fixedWednesday)

	req := httptest.NewRequest(http.MethodGet, "/api/bakeries/nonexistent/menu", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var errResp dto.ErrorResponse
	err := json.NewDecoder(w.Body).Decode(&errResp)
	require.NoError(t, err)
	assert.Equal(t, "BAKERY_NOT_FOUND", errResp.Code)
}
