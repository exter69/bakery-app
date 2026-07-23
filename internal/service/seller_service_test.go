package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lucatorrekens/bakery-app/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Test helpers ---

type mockPaymentGateway struct {
	capturedIDs []string
	voidedIDs   []string
	refundedIDs []string
	captureErr  error
	voidErr     error
	refundErr   error
}

func (m *mockPaymentGateway) CreateCheckoutURL(_ context.Context, _ string, _ int64) (string, error) {
	return "https://example.com/checkout", nil
}

func (m *mockPaymentGateway) VerifyPayment(_ context.Context, _ string) (bool, error) {
	return true, nil
}

func (m *mockPaymentGateway) CapturePayment(_ context.Context, paymentIntentID string) error {
	if m.captureErr != nil {
		return m.captureErr
	}
	m.capturedIDs = append(m.capturedIDs, paymentIntentID)
	return nil
}

func (m *mockPaymentGateway) VoidAuthorization(_ context.Context, paymentIntentID string) error {
	if m.voidErr != nil {
		return m.voidErr
	}
	m.voidedIDs = append(m.voidedIDs, paymentIntentID)
	return nil
}

func (m *mockPaymentGateway) RefundPayment(_ context.Context, paymentIntentID string, _ int64) error {
	if m.refundErr != nil {
		return m.refundErr
	}
	m.refundedIDs = append(m.refundedIDs, paymentIntentID)
	return nil
}

type fakeBakeryRepo struct {
	bakeries map[string]*domain.Bakery
}

func newFakeBakeryRepo() *fakeBakeryRepo {
	return &fakeBakeryRepo{bakeries: make(map[string]*domain.Bakery)}
}

func (r *fakeBakeryRepo) GetBakery(_ context.Context, id string) (*domain.Bakery, error) {
	b, ok := r.bakeries[id]
	if !ok {
		return nil, nil
	}
	return b, nil
}

func (r *fakeBakeryRepo) GetBakeryByOwner(_ context.Context, ownerID string) (*domain.Bakery, error) {
	for _, b := range r.bakeries {
		if b.OwnerID == ownerID {
			return b, nil
		}
	}
	return nil, nil
}

func (r *fakeBakeryRepo) ListBakeries(_ context.Context, _ domain.PaginationParams) ([]domain.Bakery, int, error) {
	return nil, 0, nil
}

func (r *fakeBakeryRepo) UpdateBakery(_ context.Context, _ *domain.Bakery) error { return nil }
func (r *fakeBakeryRepo) GetProductsByBakery(_ context.Context, _ string) ([]domain.Product, error) {
	return nil, nil
}
func (r *fakeBakeryRepo) GetProductByID(_ context.Context, _ string) (*domain.Product, error) {
	return nil, nil
}
func (r *fakeBakeryRepo) CreateProduct(_ context.Context, _ *domain.Product) error { return nil }
func (r *fakeBakeryRepo) UpdateProduct(_ context.Context, _ *domain.Product) error { return nil }
func (r *fakeBakeryRepo) DeleteProduct(_ context.Context, _ string) error           { return nil }
func (r *fakeBakeryRepo) SearchProducts(_ context.Context, _ domain.ProductSearchParams) ([]domain.ProductSearchResult, int, error) {
	return nil, 0, nil
}

type fakeOrderRepo struct {
	orders map[string]*domain.Order
}

func newFakeOrderRepo() *fakeOrderRepo {
	return &fakeOrderRepo{orders: make(map[string]*domain.Order)}
}

func (r *fakeOrderRepo) Save(_ context.Context, order *domain.Order) error {
	r.orders[order.ID] = order
	return nil
}

func (r *fakeOrderRepo) GetByID(_ context.Context, id string) (*domain.Order, error) {
	o, ok := r.orders[id]
	if !ok {
		return nil, nil
	}
	return o, nil
}

func (r *fakeOrderRepo) ListByUser(_ context.Context, _ string, _ domain.OrderFilters, _ domain.PaginationParams) ([]domain.Order, int, error) {
	return nil, 0, nil
}

func (r *fakeOrderRepo) ListByBakery(_ context.Context, _ string, _ domain.OrderFilters, _ domain.PaginationParams) ([]domain.Order, int, error) {
	return nil, 0, nil
}

// --- Tests ---

func TestUpdateOrderStatus_CapturesPaymentOnDelivery(t *testing.T) {
	bakeryRepo := newFakeBakeryRepo()
	orderRepo := newFakeOrderRepo()
	gateway := &mockPaymentGateway{}

	bakery := &domain.Bakery{ID: "bakery-1", OwnerID: "owner-1"}
	bakeryRepo.bakeries[bakery.ID] = bakery

	order := &domain.Order{
		ID:              "order-1",
		BakeryID:        "bakery-1",
		UserID:          "user-1",
		Status:          domain.OrderStatusReady,
		PaymentIntentID: "pi_test_123",
		TotalAmount:     5000,
		UpdatedAt:       time.Now(),
	}
	_ = orderRepo.Save(context.Background(), order)

	svc := NewSellerService(SellerServiceConfig{
		BakeryRepo:      bakeryRepo,
		OrderRepo:       orderRepo,
		ReservationRepo: nil,
		PaymentGateway:  gateway,
	})

	result, err := svc.UpdateOrderStatus(context.Background(), "order-1", "owner-1", domain.OrderStatusDelivered)
	require.NoError(t, err)
	assert.Equal(t, domain.OrderStatusDelivered, result.Status)
	assert.Equal(t, []string{"pi_test_123"}, gateway.capturedIDs)
}

func TestUpdateOrderStatus_SkipsCaptureWhenNoPaymentIntentID(t *testing.T) {
	bakeryRepo := newFakeBakeryRepo()
	orderRepo := newFakeOrderRepo()
	gateway := &mockPaymentGateway{}

	bakery := &domain.Bakery{ID: "bakery-1", OwnerID: "owner-1"}
	bakeryRepo.bakeries[bakery.ID] = bakery

	order := &domain.Order{
		ID:          "order-1",
		BakeryID:    "bakery-1",
		UserID:      "user-1",
		Status:      domain.OrderStatusReady,
		TotalAmount: 5000,
		UpdatedAt:   time.Now(),
	}
	_ = orderRepo.Save(context.Background(), order)

	svc := NewSellerService(SellerServiceConfig{
		BakeryRepo:      bakeryRepo,
		OrderRepo:       orderRepo,
		ReservationRepo: nil,
		PaymentGateway:  gateway,
	})

	result, err := svc.UpdateOrderStatus(context.Background(), "order-1", "owner-1", domain.OrderStatusDelivered)
	require.NoError(t, err)
	assert.Equal(t, domain.OrderStatusDelivered, result.Status)
	assert.Empty(t, gateway.capturedIDs, "should not call capture when no PaymentIntentID")
}

func TestUpdateOrderStatus_CaptureFailureReturnsError(t *testing.T) {
	bakeryRepo := newFakeBakeryRepo()
	orderRepo := newFakeOrderRepo()
	gateway := &mockPaymentGateway{captureErr: errors.New("stripe error")}

	bakery := &domain.Bakery{ID: "bakery-1", OwnerID: "owner-1"}
	bakeryRepo.bakeries[bakery.ID] = bakery

	order := &domain.Order{
		ID:              "order-1",
		BakeryID:        "bakery-1",
		UserID:          "user-1",
		Status:          domain.OrderStatusReady,
		PaymentIntentID: "pi_test_123",
		TotalAmount:     5000,
		UpdatedAt:       time.Now(),
	}
	_ = orderRepo.Save(context.Background(), order)

	svc := NewSellerService(SellerServiceConfig{
		BakeryRepo:      bakeryRepo,
		OrderRepo:       orderRepo,
		ReservationRepo: nil,
		PaymentGateway:  gateway,
	})

	_, err := svc.UpdateOrderStatus(context.Background(), "order-1", "owner-1", domain.OrderStatusDelivered)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "capturing payment")

	// Order should NOT be saved as Delivered when capture fails
	saved, _ := orderRepo.GetByID(context.Background(), "order-1")
	assert.Equal(t, domain.OrderStatusReady, saved.Status)
}

func TestUpdateOrderStatus_NormalTransitionWithoutCapture(t *testing.T) {
	bakeryRepo := newFakeBakeryRepo()
	orderRepo := newFakeOrderRepo()
	gateway := &mockPaymentGateway{}

	bakery := &domain.Bakery{ID: "bakery-1", OwnerID: "owner-1"}
	bakeryRepo.bakeries[bakery.ID] = bakery

	order := &domain.Order{
		ID:              "order-1",
		BakeryID:        "bakery-1",
		UserID:          "user-1",
		Status:          domain.OrderStatusConfirmed,
		PaymentIntentID: "pi_test_123",
		TotalAmount:     5000,
		UpdatedAt:       time.Now(),
	}
	_ = orderRepo.Save(context.Background(), order)

	svc := NewSellerService(SellerServiceConfig{
		BakeryRepo:      bakeryRepo,
		OrderRepo:       orderRepo,
		ReservationRepo: nil,
		PaymentGateway:  gateway,
	})

	result, err := svc.UpdateOrderStatus(context.Background(), "order-1", "owner-1", domain.OrderStatusPreparing)
	require.NoError(t, err)
	assert.Equal(t, domain.OrderStatusPreparing, result.Status)
	assert.Empty(t, gateway.capturedIDs, "should not capture on non-Delivered transition")
}
