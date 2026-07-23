package service

import (
	"context"
	"testing"
	"time"

	"github.com/lucatorrekens/bakery-app/internal/domain"
	"github.com/lucatorrekens/bakery-app/internal/repository/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupPayoutTest() (*PayoutService, *memory.OrderRepo, *memory.BakeryRepo, *memory.PayoutRepo) {
	orderRepo := memory.NewOrderRepo()
	bakeryRepo := memory.NewBakeryRepo()
	payoutRepo := memory.NewPayoutRepo()

	svc := NewPayoutService(PayoutServiceConfig{
		ConnectSvc: nil, // No Stripe in tests
		PayoutRepo: payoutRepo,
		BakeryRepo: bakeryRepo,
		OrderRepo:  orderRepo,
	})

	return svc, orderRepo, bakeryRepo, payoutRepo
}

func TestOnOrderDelivered_createsPayoutWithCorrectSplit(t *testing.T) {
	svc, orderRepo, bakeryRepo, payoutRepo := setupPayoutTest()
	ctx := context.Background()

	// Arrange: bakery with 15% commission
	bakeryRepo.SeedBakery(domain.Bakery{
		ID:             "bakery-1",
		OwnerID:        "seller-1",
		Name:           "Test Bakery",
		CommissionRate: 15,
	})

	order := &domain.Order{
		ID:          "order-1",
		BakeryID:    "bakery-1",
		UserID:      "user-1",
		TotalAmount: 2000, // 20 EUR
		Status:      domain.OrderStatusDelivered,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	_ = orderRepo.Save(ctx, order)

	// Act
	err := svc.OnOrderDelivered(ctx, "order-1")

	// Assert
	require.NoError(t, err)

	payout, err := payoutRepo.GetByOrderID(ctx, "order-1")
	require.NoError(t, err)
	require.NotNil(t, payout)

	assert.Equal(t, "order-1", payout.OrderID)
	assert.Equal(t, "bakery-1", payout.BakeryID)
	assert.Equal(t, int64(1700), payout.Amount)     // 2000 - 300 (15%)
	assert.Equal(t, int64(300), payout.Commission)   // 2000 * 15 / 100
	assert.Equal(t, domain.PayoutStatusPending, payout.Status)
}

func TestOnOrderDelivered_isIdempotent(t *testing.T) {
	svc, orderRepo, bakeryRepo, _ := setupPayoutTest()
	ctx := context.Background()

	bakeryRepo.SeedBakery(domain.Bakery{
		ID:             "bakery-1",
		OwnerID:        "seller-1",
		Name:           "Test Bakery",
		CommissionRate: 10,
	})

	order := &domain.Order{
		ID:          "order-1",
		BakeryID:    "bakery-1",
		UserID:      "user-1",
		TotalAmount: 1000,
		Status:      domain.OrderStatusDelivered,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	_ = orderRepo.Save(ctx, order)

	// First call succeeds
	err := svc.OnOrderDelivered(ctx, "order-1")
	require.NoError(t, err)

	// Second call returns duplicate error
	err = svc.OnOrderDelivered(ctx, "order-1")
	assert.Equal(t, ErrPayoutAlreadyExists, err)
}

func TestOnOrderDelivered_returnsErrorForMissingOrder(t *testing.T) {
	svc, _, _, _ := setupPayoutTest()
	ctx := context.Background()

	err := svc.OnOrderDelivered(ctx, "nonexistent-order")
	assert.ErrorIs(t, err, ErrOrderNotFound)
}

func TestOnOrderDelivered_returnsErrorForMissingBakery(t *testing.T) {
	svc, orderRepo, _, _ := setupPayoutTest()
	ctx := context.Background()

	order := &domain.Order{
		ID:          "order-1",
		BakeryID:    "nonexistent-bakery",
		UserID:      "user-1",
		TotalAmount: 1000,
		Status:      domain.OrderStatusDelivered,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	_ = orderRepo.Save(ctx, order)

	err := svc.OnOrderDelivered(ctx, "order-1")
	assert.ErrorIs(t, err, ErrBakeryNotFound)
}

func TestOnOrderRefunded_reversesPayoutStatus(t *testing.T) {
	svc, orderRepo, bakeryRepo, payoutRepo := setupPayoutTest()
	ctx := context.Background()

	bakeryRepo.SeedBakery(domain.Bakery{
		ID:             "bakery-1",
		OwnerID:        "seller-1",
		Name:           "Test Bakery",
		CommissionRate: 10,
	})

	order := &domain.Order{
		ID:          "order-1",
		BakeryID:    "bakery-1",
		UserID:      "user-1",
		TotalAmount: 1000,
		Status:      domain.OrderStatusDelivered,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	_ = orderRepo.Save(ctx, order)

	// Create a payout first
	err := svc.OnOrderDelivered(ctx, "order-1")
	require.NoError(t, err)

	// Refund
	err = svc.OnOrderRefunded(ctx, "order-1")
	require.NoError(t, err)

	payout, _ := payoutRepo.GetByOrderID(ctx, "order-1")
	assert.Equal(t, domain.PayoutStatusRefunded, payout.Status)
}

func TestOnOrderRefunded_noPayoutIsNoop(t *testing.T) {
	svc, _, _, _ := setupPayoutTest()
	ctx := context.Background()

	// No payout exists — should not error
	err := svc.OnOrderRefunded(ctx, "nonexistent-order")
	assert.NoError(t, err)
}

func TestListPayouts_returnsPaginatedResults(t *testing.T) {
	svc, _, _, payoutRepo := setupPayoutTest()
	ctx := context.Background()

	// Create 3 payouts
	for i := 0; i < 3; i++ {
		payout := &domain.Payout{
			OrderID:    "order-" + string(rune('a'+i)),
			BakeryID:   "bakery-1",
			Amount:     900,
			Commission: 100,
			Status:     domain.PayoutStatusTransferred,
			CreatedAt:  time.Now().Add(time.Duration(i) * time.Minute),
		}
		_ = payoutRepo.Create(ctx, payout)
	}

	result, err := svc.ListPayouts(ctx, "bakery-1", domain.PaginationParams{Page: 1, PageSize: 2})
	require.NoError(t, err)
	assert.Equal(t, 3, result.Total)
	assert.Len(t, result.Items, 2)
	assert.Equal(t, 1, result.Page)
	assert.Equal(t, 2, result.PageSize)
}

func TestListPayouts_emptyForUnknownBakery(t *testing.T) {
	svc, _, _, _ := setupPayoutTest()
	ctx := context.Background()

	result, err := svc.ListPayouts(ctx, "unknown-bakery", domain.PaginationParams{Page: 1, PageSize: 20})
	require.NoError(t, err)
	assert.Equal(t, 0, result.Total)
	assert.Empty(t, result.Items)
}

func TestGetConnectStatus_returnsDisconnectedWhenNoBakery(t *testing.T) {
	svc, _, _, _ := setupPayoutTest()
	ctx := context.Background()

	_, _, _, err := svc.GetConnectStatus(ctx, "nonexistent")
	assert.ErrorIs(t, err, ErrBakeryNotFound)
}

func TestGetConnectStatus_returnsDisconnectedWhenNoStripeID(t *testing.T) {
	svc, _, bakeryRepo, _ := setupPayoutTest()
	ctx := context.Background()

	bakeryRepo.SeedBakery(domain.Bakery{
		ID:      "bakery-1",
		OwnerID: "seller-1",
		Name:    "Test Bakery",
	})

	connected, charges, payouts, err := svc.GetConnectStatus(ctx, "bakery-1")
	require.NoError(t, err)
	assert.False(t, connected)
	assert.False(t, charges)
	assert.False(t, payouts)
}

// Task 6.1: Verify a pending/failed payout is marked "refunded" without calling Stripe.
func TestPayoutService_OnOrderRefunded_NonTransferredPayout(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name   string
		status domain.PayoutStatus
	}{
		{"pending payout", domain.PayoutStatusPending},
		{"failed payout", domain.PayoutStatusFailed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payoutRepo := memory.NewPayoutRepo()
			orderRepo := memory.NewOrderRepo()
			bakeryRepo := memory.NewBakeryRepo()

			// Create the payout service with nil ConnectSvc — any Stripe call would panic
			svc := NewPayoutService(PayoutServiceConfig{
				ConnectSvc: nil,
				PayoutRepo: payoutRepo,
				BakeryRepo: bakeryRepo,
				OrderRepo:  orderRepo,
			})

			// Directly seed a payout with non-transferred status
			payout := &domain.Payout{
				ID:         "payout-" + string(tt.status),
				OrderID:    "order-refund-test",
				BakeryID:   "bakery-1",
				Amount:     1700,
				Commission: 300,
				Status:     tt.status,
			}
			err := payoutRepo.Create(ctx, payout)
			require.NoError(t, err)

			// Act: refund the order
			err = svc.OnOrderRefunded(ctx, "order-refund-test")
			require.NoError(t, err)

			// Assert: payout is now marked as refunded
			updated, err := payoutRepo.GetByOrderID(ctx, "order-refund-test")
			require.NoError(t, err)
			require.NotNil(t, updated)
			assert.Equal(t, domain.PayoutStatusRefunded, updated.Status)
			// No Stripe call was made (ConnectSvc is nil — would panic if called)
			assert.Empty(t, updated.StripeTransferID, "no transfer ID should be set for non-transferred payouts")
		})
	}
}

// Task 6.2: Verify nil return when no payout exists for the given order.
func TestPayoutService_OnOrderRefunded_NoPayout(t *testing.T) {
	svc, _, _, _ := setupPayoutTest()
	ctx := context.Background()

	// No payout exists for this order ID — should return nil without error
	err := svc.OnOrderRefunded(ctx, "order-without-payout")
	assert.NoError(t, err)
}
