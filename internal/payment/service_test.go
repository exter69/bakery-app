package payment

import (
	"context"
	"testing"
	"time"

	"github.com/lucatorrekens/bakery-app/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Test helpers ---

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

func TestInitiatePayment_ReturnsLinkWith30MinExpiry(t *testing.T) {
	repo := newFakeOrderRepo()
	now := time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)

	svc := NewPaymentService(ServiceConfig{
		Gateway:   NewStubGateway(),
		OrderRepo: repo,
		Clock:     func() time.Time { return now },
	})

	link, err := svc.InitiatePayment(context.Background(), "order-1", 5000)
	require.NoError(t, err)
	assert.NotEmpty(t, link.URL)
	assert.Contains(t, link.URL, "order-1")
	assert.Equal(t, 1800, link.ExpiresIn) // 30 minutes in seconds
}

func TestInitiatePayment_GeneratesUniqueTokens(t *testing.T) {
	repo := newFakeOrderRepo()
	svc := NewPaymentService(ServiceConfig{
		Gateway:   NewStubGateway(),
		OrderRepo: repo,
	})

	link1, err := svc.InitiatePayment(context.Background(), "order-1", 5000)
	require.NoError(t, err)

	link2, err := svc.InitiatePayment(context.Background(), "order-2", 3000)
	require.NoError(t, err)

	assert.NotEqual(t, link1.URL, link2.URL)
}

func TestProcessPaymentCallback_ConfirmsOrder(t *testing.T) {
	repo := newFakeOrderRepo()
	now := time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)

	order := &domain.Order{
		ID:     "order-1",
		Status: domain.OrderStatusPendingPayment,
	}
	_ = repo.Save(context.Background(), order)

	svc := NewPaymentService(ServiceConfig{
		Gateway:   NewStubGateway(),
		OrderRepo: repo,
		Clock:     func() time.Time { return now },
	})

	// First initiate payment to register the link
	_, err := svc.InitiatePayment(context.Background(), "order-1", 5000)
	require.NoError(t, err)

	// Process callback within the expiry window
	err = svc.ProcessPaymentCallback(context.Background(), "order-1", "ref-123")
	require.NoError(t, err)

	// Verify order status was updated
	updated, _ := repo.GetByID(context.Background(), "order-1")
	assert.Equal(t, domain.OrderStatusConfirmed, updated.Status)
}

func TestProcessPaymentCallback_RejectsExpiredLink(t *testing.T) {
	repo := newFakeOrderRepo()
	currentTime := time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)

	order := &domain.Order{
		ID:     "order-1",
		Status: domain.OrderStatusPendingPayment,
	}
	_ = repo.Save(context.Background(), order)

	svc := NewPaymentService(ServiceConfig{
		Gateway:   NewStubGateway(),
		OrderRepo: repo,
		Clock:     func() time.Time { return currentTime },
	})

	_, err := svc.InitiatePayment(context.Background(), "order-1", 5000)
	require.NoError(t, err)

	// Advance time past the 30 minute expiry
	currentTime = currentTime.Add(31 * time.Minute)

	err = svc.ProcessPaymentCallback(context.Background(), "order-1", "ref-123")
	assert.ErrorIs(t, err, ErrLinkExpired)

	// Verify order status was NOT changed
	unchanged, _ := repo.GetByID(context.Background(), "order-1")
	assert.Equal(t, domain.OrderStatusPendingPayment, unchanged.Status)
}

func TestProcessPaymentCallback_RejectsSingleUseViolation(t *testing.T) {
	repo := newFakeOrderRepo()
	now := time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)

	order := &domain.Order{
		ID:     "order-1",
		Status: domain.OrderStatusPendingPayment,
	}
	_ = repo.Save(context.Background(), order)

	svc := NewPaymentService(ServiceConfig{
		Gateway:   NewStubGateway(),
		OrderRepo: repo,
		Clock:     func() time.Time { return now },
	})

	_, err := svc.InitiatePayment(context.Background(), "order-1", 5000)
	require.NoError(t, err)

	// First callback succeeds
	err = svc.ProcessPaymentCallback(context.Background(), "order-1", "ref-123")
	require.NoError(t, err)

	// Second callback on same link should fail
	err = svc.ProcessPaymentCallback(context.Background(), "order-1", "ref-456")
	assert.ErrorIs(t, err, ErrLinkUsed)
}

func TestProcessPaymentCallback_RejectsUnknownOrder(t *testing.T) {
	repo := newFakeOrderRepo()
	svc := NewPaymentService(ServiceConfig{
		Gateway:   NewStubGateway(),
		OrderRepo: repo,
	})

	err := svc.ProcessPaymentCallback(context.Background(), "nonexistent-order", "ref-123")
	assert.ErrorIs(t, err, ErrLinkNotFound)
}

func TestProcessPaymentCallback_RejectsNonPendingOrder(t *testing.T) {
	repo := newFakeOrderRepo()
	now := time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)

	// Order already confirmed (not pending_payment)
	order := &domain.Order{
		ID:     "order-1",
		Status: domain.OrderStatusConfirmed,
	}
	_ = repo.Save(context.Background(), order)

	svc := NewPaymentService(ServiceConfig{
		Gateway:   NewStubGateway(),
		OrderRepo: repo,
		Clock:     func() time.Time { return now },
	})

	_, err := svc.InitiatePayment(context.Background(), "order-1", 5000)
	require.NoError(t, err)

	err = svc.ProcessPaymentCallback(context.Background(), "order-1", "ref-123")
	assert.ErrorIs(t, err, ErrInvalidOrderStatus)
}
