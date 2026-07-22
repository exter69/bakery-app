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
	"github.com/lucatorrekens/bakery-app/internal/repository/memory"
	"github.com/lucatorrekens/bakery-app/internal/service"
)

const sellerTestSecret = "seller-test-secret"

func setupSellerTestRouter(t *testing.T) (*memory.BakeryRepo, chi.Router) {
	t.Helper()

	bakeryRepo := memory.NewBakeryRepo()
	orderRepo := memory.NewOrderRepo()
	reservationRepo := memory.NewReservationRepo()

	sellerSvc := service.NewSellerService(service.SellerServiceConfig{
		BakeryRepo:      bakeryRepo,
		OrderRepo:       orderRepo,
		ReservationRepo: reservationRepo,
	})

	handler := api.NewSellerHandler(sellerSvc)

	r := chi.NewRouter()
	r.Use(middleware.JWTAuth(sellerTestSecret))
	handler.RegisterRoutes(r)

	return bakeryRepo, r
}

func sellerToken(t *testing.T, userID string) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  userID,
		"role": float64(domain.RoleSeller),
		"exp":  time.Now().Add(1 * time.Hour).Unix(),
		"iat":  time.Now().Unix(),
	})
	signed, err := token.SignedString([]byte(sellerTestSecret))
	require.NoError(t, err)
	return signed
}

func seedSellerBakeryAndProduct(repo *memory.BakeryRepo, ownerID string) (string, string) {
	bakeryID := "bakery-seller-1"
	productID := "product-1"

	repo.SeedBakery(domain.Bakery{
		ID:      bakeryID,
		Name:    "Test Bakery",
		OwnerID: ownerID,
		Schedule: []domain.DaySchedule{
			{Day: domain.Monday, OpenTime: domain.TimeOfDay{Hour: 8, Minute: 0}, CloseTime: domain.TimeOfDay{Hour: 18, Minute: 0}, IsOpen: true},
		},
	})

	repo.SeedProduct(domain.Product{
		ID:          productID,
		BakeryID:    bakeryID,
		Name:        "Croissant",
		Price:       350,
		Category:    "Pastries",
		IsAvailable: true,
		Allergens:   []string{"gluten", "dairy"},
		HealthScore: intPtr(4),
	})

	return bakeryID, productID
}

func intPtr(i int) *int {
	return &i
}

func TestUpdateProduct_ValidAllergens(t *testing.T) {
	repo, router := setupSellerTestRouter(t)
	ownerID := "seller-1"
	_, productID := seedSellerBakeryAndProduct(repo, ownerID)

	body := map[string]interface{}{
		"allergens": []string{"eggs", "fish", "peanuts"},
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/api/products/"+productID, bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+sellerToken(t, ownerID))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var product domain.Product
	err := json.NewDecoder(w.Body).Decode(&product)
	require.NoError(t, err)
	assert.Equal(t, []string{"eggs", "fish", "peanuts"}, product.Allergens)
	// HealthScore should remain unchanged
	require.NotNil(t, product.HealthScore)
	assert.Equal(t, 4, *product.HealthScore)
}

func TestUpdateProduct_EmptyAllergens(t *testing.T) {
	repo, router := setupSellerTestRouter(t)
	ownerID := "seller-1"
	_, productID := seedSellerBakeryAndProduct(repo, ownerID)

	body := map[string]interface{}{
		"allergens": []string{},
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/api/products/"+productID, bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+sellerToken(t, ownerID))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var product domain.Product
	err := json.NewDecoder(w.Body).Decode(&product)
	require.NoError(t, err)
	assert.Equal(t, []string{}, product.Allergens)
}

func TestUpdateProduct_InvalidAllergen(t *testing.T) {
	repo, router := setupSellerTestRouter(t)
	ownerID := "seller-1"
	_, productID := seedSellerBakeryAndProduct(repo, ownerID)

	body := map[string]interface{}{
		"allergens": []string{"gluten", "chocolate"},
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/api/products/"+productID, bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+sellerToken(t, ownerID))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errResp dto.ErrorResponse
	err := json.NewDecoder(w.Body).Decode(&errResp)
	require.NoError(t, err)
	assert.Equal(t, "INVALID_ALLERGEN", errResp.Code)
	assert.Contains(t, errResp.Message, "chocolate")
}

func TestUpdateProduct_ValidHealthScore(t *testing.T) {
	repo, router := setupSellerTestRouter(t)
	ownerID := "seller-1"
	_, productID := seedSellerBakeryAndProduct(repo, ownerID)

	body := map[string]interface{}{
		"healthScore": 3,
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/api/products/"+productID, bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+sellerToken(t, ownerID))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var product domain.Product
	err := json.NewDecoder(w.Body).Decode(&product)
	require.NoError(t, err)
	require.NotNil(t, product.HealthScore)
	assert.Equal(t, 3, *product.HealthScore)
	// Allergens should remain unchanged
	assert.Equal(t, []string{"gluten", "dairy"}, product.Allergens)
}

func TestUpdateProduct_NullHealthScore(t *testing.T) {
	repo, router := setupSellerTestRouter(t)
	ownerID := "seller-1"
	_, productID := seedSellerBakeryAndProduct(repo, ownerID)

	body := map[string]interface{}{
		"healthScore": nil,
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/api/products/"+productID, bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+sellerToken(t, ownerID))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var product domain.Product
	err := json.NewDecoder(w.Body).Decode(&product)
	require.NoError(t, err)
	assert.Nil(t, product.HealthScore)
}

func TestUpdateProduct_InvalidHealthScore_TooHigh(t *testing.T) {
	repo, router := setupSellerTestRouter(t)
	ownerID := "seller-1"
	_, productID := seedSellerBakeryAndProduct(repo, ownerID)

	body := map[string]interface{}{
		"healthScore": 10,
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/api/products/"+productID, bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+sellerToken(t, ownerID))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errResp dto.ErrorResponse
	err := json.NewDecoder(w.Body).Decode(&errResp)
	require.NoError(t, err)
	assert.Equal(t, "INVALID_HEALTH_SCORE", errResp.Code)
}

func TestUpdateProduct_InvalidHealthScore_Zero(t *testing.T) {
	repo, router := setupSellerTestRouter(t)
	ownerID := "seller-1"
	_, productID := seedSellerBakeryAndProduct(repo, ownerID)

	body := map[string]interface{}{
		"healthScore": 0,
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/api/products/"+productID, bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+sellerToken(t, ownerID))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errResp dto.ErrorResponse
	err := json.NewDecoder(w.Body).Decode(&errResp)
	require.NoError(t, err)
	assert.Equal(t, "INVALID_HEALTH_SCORE", errResp.Code)
}

func TestUpdateProduct_OmittedFieldsUnchanged(t *testing.T) {
	repo, router := setupSellerTestRouter(t)
	ownerID := "seller-1"
	_, productID := seedSellerBakeryAndProduct(repo, ownerID)

	// Only update name - allergens and healthScore should remain unchanged
	body := map[string]interface{}{
		"name": "Updated Croissant",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/api/products/"+productID, bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+sellerToken(t, ownerID))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var product domain.Product
	err := json.NewDecoder(w.Body).Decode(&product)
	require.NoError(t, err)
	assert.Equal(t, "Updated Croissant", product.Name)
	assert.Equal(t, []string{"gluten", "dairy"}, product.Allergens)
	require.NotNil(t, product.HealthScore)
	assert.Equal(t, 4, *product.HealthScore)
}

func TestUpdateProduct_BothAllergensAndHealthScore(t *testing.T) {
	repo, router := setupSellerTestRouter(t)
	ownerID := "seller-1"
	_, productID := seedSellerBakeryAndProduct(repo, ownerID)

	body := map[string]interface{}{
		"allergens":   []string{"soy", "sesame"},
		"healthScore": 2,
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/api/products/"+productID, bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+sellerToken(t, ownerID))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var product domain.Product
	err := json.NewDecoder(w.Body).Decode(&product)
	require.NoError(t, err)
	assert.Equal(t, []string{"soy", "sesame"}, product.Allergens)
	require.NotNil(t, product.HealthScore)
	assert.Equal(t, 2, *product.HealthScore)
}

func TestUpdateProduct_DuplicateAllergens(t *testing.T) {
	repo, router := setupSellerTestRouter(t)
	ownerID := "seller-1"
	_, productID := seedSellerBakeryAndProduct(repo, ownerID)

	body := map[string]interface{}{
		"allergens": []string{"gluten", "gluten"},
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/api/products/"+productID, bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+sellerToken(t, ownerID))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errResp dto.ErrorResponse
	err := json.NewDecoder(w.Body).Decode(&errResp)
	require.NoError(t, err)
	assert.Equal(t, "INVALID_ALLERGEN", errResp.Code)
	assert.Contains(t, errResp.Message, "duplicate")
}
