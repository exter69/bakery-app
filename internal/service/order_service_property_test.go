package service

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/lucatorrekens/bakery-app/internal/domain"
	"github.com/lucatorrekens/bakery-app/internal/payment"
	"github.com/lucatorrekens/bakery-app/internal/repository/memory"
	"pgregory.net/rapid"
)

// trackingPaymentService is a mock PaymentService that records refund calls.
type trackingPaymentService struct {
	mu          sync.Mutex
	refundCalls []refundCall
}

type refundCall struct {
	OrderID string
	Amount  int64
}

func newTrackingPaymentService() *trackingPaymentService {
	return &trackingPaymentService{}
}

func (s *trackingPaymentService) InitiatePayment(_ context.Context, orderID string, amount int64) (*domain.PaymentLink, error) {
	return &domain.PaymentLink{
		URL:       fmt.Sprintf("https://pay.example.com/checkout/%s?amount=%d", orderID, amount),
		ExpiresIn: 1800,
	}, nil
}

func (s *trackingPaymentService) ProcessPaymentCallback(_ context.Context, _ string, _ string) error {
	return nil
}

func (s *trackingPaymentService) InitiateRefund(_ context.Context, orderID string, amount int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.refundCalls = append(s.refundCalls, refundCall{OrderID: orderID, Amount: amount})
	return nil
}

func (s *trackingPaymentService) getRefundCalls() []refundCall {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]refundCall, len(s.refundCalls))
	copy(result, s.refundCalls)
	return result
}

// TestProperty_OwnershipEnforcement verifies that mutation operations (DeleteOrder)
// are rejected with ErrForbidden when the requesting user is not the owner.
// Also verifies that the owner can successfully delete the order.
// **Validates: Requirements 7.4, 7.5, 10.2**
func TestProperty_OwnershipEnforcement(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Set up infrastructure
		bakeryRepo := memory.NewBakeryRepo()
		orderRepo := memory.NewOrderRepo()
		paymentSvc := payment.NewMockPaymentService()

		bakeryID := "bakery-1"
		bakery := domain.Bakery{
			ID:   bakeryID,
			Name: "Test Bakery",
			Schedule: []domain.DaySchedule{
				{Day: domain.Monday, IsOpen: true, OpenTime: domain.TimeOfDay{Hour: 6, Minute: 0}, CloseTime: domain.TimeOfDay{Hour: 22, Minute: 0}},
				{Day: domain.Tuesday, IsOpen: true, OpenTime: domain.TimeOfDay{Hour: 6, Minute: 0}, CloseTime: domain.TimeOfDay{Hour: 22, Minute: 0}},
				{Day: domain.Wednesday, IsOpen: true, OpenTime: domain.TimeOfDay{Hour: 6, Minute: 0}, CloseTime: domain.TimeOfDay{Hour: 22, Minute: 0}},
				{Day: domain.Thursday, IsOpen: true, OpenTime: domain.TimeOfDay{Hour: 6, Minute: 0}, CloseTime: domain.TimeOfDay{Hour: 22, Minute: 0}},
				{Day: domain.Friday, IsOpen: true, OpenTime: domain.TimeOfDay{Hour: 6, Minute: 0}, CloseTime: domain.TimeOfDay{Hour: 22, Minute: 0}},
				{Day: domain.Saturday, IsOpen: true, OpenTime: domain.TimeOfDay{Hour: 8, Minute: 0}, CloseTime: domain.TimeOfDay{Hour: 20, Minute: 0}},
				{Day: domain.Sunday, IsOpen: true, OpenTime: domain.TimeOfDay{Hour: 8, Minute: 0}, CloseTime: domain.TimeOfDay{Hour: 18, Minute: 0}},
			},
		}
		bakeryRepo.SeedBakery(bakery)

		bakeryRepo.SeedProduct(domain.Product{
			ID:          "product-1",
			BakeryID:    bakeryID,
			Name:        "Croissant",
			Price:       350,
			Category:    "pastry",
			IsAvailable: true,
		})

		ownerUserID := "user-A"
		idCounter := 0
		svc := NewOrderService(OrderServiceConfig{
			OrderRepo:  orderRepo,
			BakeryRepo: bakeryRepo,
			PaymentSvc: paymentSvc,
			IDGen: func() string {
				idCounter++
				return fmt.Sprintf("order-%d", idCounter)
			},
			Now: func() time.Time { return time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC) },
		})

		// Create an order owned by user-A
		order := domain.Order{
			BakeryID:     bakeryID,
			Items:        []domain.OrderItem{{ProductID: "product-1", Quantity: 1}},
			ScheduledDay: domain.Monday,
			ScheduledTime: domain.TimeSlot{
				StartTime: domain.TimeOfDay{Hour: 10, Minute: 0},
				EndTime:   domain.TimeOfDay{Hour: 10, Minute: 30},
			},
		}
		created, _, err := svc.CreateOrder(context.Background(), ownerUserID, order)
		if err != nil {
			t.Fatalf("failed to create order: %v", err)
		}

		// Generate a random non-owner user ID (must differ from "user-A")
		nonOwnerID := rapid.StringMatching(`[a-z0-9\-]{3,20}`).
			Filter(func(s string) bool { return s != ownerUserID }).
			Draw(t, "nonOwnerUserID")

		// Property: DeleteOrder MUST return ErrForbidden when userID != order.UserID
		err = svc.DeleteOrder(context.Background(), created.ID, nonOwnerID)
		if err != ErrForbidden {
			t.Fatalf("expected ErrForbidden for non-owner %q, got: %v", nonOwnerID, err)
		}

		// Positive case: owner can delete their own order
		err = svc.DeleteOrder(context.Background(), created.ID, ownerUserID)
		if err != nil {
			t.Fatalf("expected owner %q to delete order successfully, got: %v", ownerUserID, err)
		}
	})
}

// TestProperty_InitialOrderStatus verifies that all newly created orders
// have status PendingPayment regardless of the input data.
// **Validates: Requirements 3.10**
func TestProperty_InitialOrderStatus(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Set up test bakery with a full-week open schedule
		bakeryRepo := memory.NewBakeryRepo()
		orderRepo := memory.NewOrderRepo()
		paymentSvc := payment.NewMockPaymentService()

		bakeryID := "bakery-1"
		bakery := domain.Bakery{
			ID:   bakeryID,
			Name: "Test Bakery",
			Schedule: []domain.DaySchedule{
				{Day: domain.Monday, IsOpen: true, OpenTime: domain.TimeOfDay{Hour: 6, Minute: 0}, CloseTime: domain.TimeOfDay{Hour: 22, Minute: 0}},
				{Day: domain.Tuesday, IsOpen: true, OpenTime: domain.TimeOfDay{Hour: 6, Minute: 0}, CloseTime: domain.TimeOfDay{Hour: 22, Minute: 0}},
				{Day: domain.Wednesday, IsOpen: true, OpenTime: domain.TimeOfDay{Hour: 6, Minute: 0}, CloseTime: domain.TimeOfDay{Hour: 22, Minute: 0}},
				{Day: domain.Thursday, IsOpen: true, OpenTime: domain.TimeOfDay{Hour: 6, Minute: 0}, CloseTime: domain.TimeOfDay{Hour: 22, Minute: 0}},
				{Day: domain.Friday, IsOpen: true, OpenTime: domain.TimeOfDay{Hour: 6, Minute: 0}, CloseTime: domain.TimeOfDay{Hour: 22, Minute: 0}},
				{Day: domain.Saturday, IsOpen: true, OpenTime: domain.TimeOfDay{Hour: 8, Minute: 0}, CloseTime: domain.TimeOfDay{Hour: 20, Minute: 0}},
				{Day: domain.Sunday, IsOpen: true, OpenTime: domain.TimeOfDay{Hour: 8, Minute: 0}, CloseTime: domain.TimeOfDay{Hour: 18, Minute: 0}},
			},
		}
		bakeryRepo.SeedBakery(bakery)

		// Create a pool of available products
		numProducts := rapid.IntRange(1, 10).Draw(t, "numProducts")
		productIDs := make([]string, numProducts)
		for i := 0; i < numProducts; i++ {
			pid := fmt.Sprintf("product-%d", i+1)
			productIDs[i] = pid
			bakeryRepo.SeedProduct(domain.Product{
				ID:          pid,
				BakeryID:    bakeryID,
				Name:        fmt.Sprintf("Product %d", i+1),
				Price:       int64(rapid.IntRange(100, 50000).Draw(t, fmt.Sprintf("price-%d", i))),
				Category:    "pastry",
				IsAvailable: true,
			})
		}

		// Create the order service with a deterministic ID generator
		idCounter := 0
		svc := NewOrderService(OrderServiceConfig{
			OrderRepo:  orderRepo,
			BakeryRepo: bakeryRepo,
			PaymentSvc: paymentSvc,
			IDGen: func() string {
				idCounter++
				return fmt.Sprintf("test-order-%d", idCounter)
			},
			Now: func() time.Time { return time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC) },
		})

		// Generate random valid order items (1 to 5 items, using valid product IDs)
		numItems := rapid.IntRange(1, 5).Draw(t, "numItems")
		items := make([]domain.OrderItem, numItems)
		for i := range items {
			pidIdx := rapid.IntRange(0, len(productIDs)-1).Draw(t, fmt.Sprintf("itemProdIdx-%d", i))
			items[i] = domain.OrderItem{
				ProductID: productIDs[pidIdx],
				Quantity:  rapid.IntRange(1, 999).Draw(t, fmt.Sprintf("itemQty-%d", i)),
			}
		}

		// Generate a random valid schedule (random day, time within bakery hours)
		days := domain.AllDaysOfWeek()
		dayIdx := rapid.IntRange(0, len(days)-1).Draw(t, "dayIdx")
		day := days[dayIdx]

		// Get the schedule for the selected day to generate valid times
		var daySchedule domain.DaySchedule
		for _, ds := range bakery.Schedule {
			if ds.Day == day {
				daySchedule = ds
				break
			}
		}

		// Generate start and end times within operating hours
		openMins := daySchedule.OpenTime.Hour*60 + daySchedule.OpenTime.Minute
		closeMins := daySchedule.CloseTime.Hour*60 + daySchedule.CloseTime.Minute
		startMins := rapid.IntRange(openMins, closeMins).Draw(t, "startMins")
		endMins := rapid.IntRange(startMins, closeMins).Draw(t, "endMins")

		order := domain.Order{
			BakeryID: bakeryID,
			Items:    items,
			ScheduledDay: day,
			ScheduledTime: domain.TimeSlot{
				StartTime: domain.TimeOfDay{Hour: startMins / 60, Minute: startMins % 60},
				EndTime:   domain.TimeOfDay{Hour: endMins / 60, Minute: endMins % 60},
			},
		}

		// Call CreateOrder
		created, _, err := svc.CreateOrder(context.Background(), "user-123", order)
		if err != nil {
			t.Fatalf("unexpected error creating order: %v", err)
		}

		// Verify: the initial status must always be PendingPayment
		if created.Status != domain.OrderStatusPendingPayment {
			t.Fatalf("expected status %q, got %q", domain.OrderStatusPendingPayment, created.Status)
		}
	})
}

// TestProperty_DeletionRestrictedForTerminalStates verifies that orders in
// Delivered or Cancelled states cannot be deleted, and their state remains unchanged.
// **Validates: Requirements 7.7**
func TestProperty_DeletionRestrictedForTerminalStates(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Set up dependencies
		orderRepo := memory.NewOrderRepo()
		paymentSvc := payment.NewMockPaymentService()

		svc := NewOrderService(OrderServiceConfig{
			OrderRepo:  orderRepo,
			PaymentSvc: paymentSvc,
			IDGen: func() string {
				return "unused"
			},
			Now: func() time.Time { return time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC) },
		})

		// Generate a random terminal status: Delivered or Cancelled
		terminalStatuses := []domain.OrderStatus{
			domain.OrderStatusDelivered,
			domain.OrderStatusCancelled,
		}
		statusIdx := rapid.IntRange(0, len(terminalStatuses)-1).Draw(t, "statusIdx")
		terminalStatus := terminalStatuses[statusIdx]

		// Generate a random order ID and user ID
		orderID := fmt.Sprintf("order-%s", rapid.StringMatching(`[a-z0-9]{6,12}`).Draw(t, "orderID"))
		userID := fmt.Sprintf("user-%s", rapid.StringMatching(`[a-z0-9]{4,8}`).Draw(t, "userID"))

		// Create and seed an order in the terminal state
		order := domain.Order{
			ID:       orderID,
			BakeryID: "bakery-1",
			UserID:   userID,
			Items: []domain.OrderItem{
				{ProductID: "product-1", ProductName: "Croissant", Quantity: 1, UnitPrice: 350, Subtotal: 350},
			},
			ScheduledDay: domain.Monday,
			ScheduledTime: domain.TimeSlot{
				StartTime: domain.TimeOfDay{Hour: 10, Minute: 0},
				EndTime:   domain.TimeOfDay{Hour: 11, Minute: 0},
			},
			Status:        terminalStatus,
			TotalAmount:   350,
			PaymentMethod: domain.PaymentMethodOnline,
			CreatedAt:     time.Date(2025, 1, 14, 9, 0, 0, 0, time.UTC),
			UpdatedAt:     time.Date(2025, 1, 14, 12, 0, 0, 0, time.UTC),
		}
		if err := orderRepo.Save(context.Background(), &order); err != nil {
			t.Fatalf("failed to seed order: %v", err)
		}

		// Attempt to delete the order with the correct owner
		err := svc.DeleteOrder(context.Background(), orderID, userID)

		// Property: DeleteOrder MUST return ErrOrderNotCancellable
		if err != ErrOrderNotCancellable {
			t.Fatalf("expected ErrOrderNotCancellable for status %q, got: %v", terminalStatus, err)
		}

		// Property: The order's status must remain unchanged
		persisted, getErr := orderRepo.GetByID(context.Background(), orderID)
		if getErr != nil {
			t.Fatalf("failed to retrieve order after delete attempt: %v", getErr)
		}
		if persisted.Status != terminalStatus {
			t.Fatalf("expected order status to remain %q, got %q", terminalStatus, persisted.Status)
		}
	})
}


// TestProperty_CancellationRefundInvariant verifies that when a confirmed or preparing
// order is cancelled via DeleteOrder, a refund is always initiated with the correct
// orderID and amount.
// **Validates: Requirements 7.8**
func TestProperty_CancellationRefundInvariant(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Set up dependencies
		orderRepo := memory.NewOrderRepo()
		trackingPay := newTrackingPaymentService()

		svc := NewOrderService(OrderServiceConfig{
			OrderRepo:  orderRepo,
			BakeryRepo: memory.NewBakeryRepo(),
			PaymentSvc: trackingPay,
			IDGen:      func() string { return "unused" },
			Now:        func() time.Time { return time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC) },
		})

		// Generate a random order with status Confirmed or Preparing
		status := rapid.SampledFrom([]domain.OrderStatus{
			domain.OrderStatusConfirmed,
			domain.OrderStatusPreparing,
		}).Draw(t, "status")

		orderID := fmt.Sprintf("order-%s", rapid.StringMatching(`[a-z0-9]{8}`).Draw(t, "orderID"))
		userID := fmt.Sprintf("user-%s", rapid.StringMatching(`[a-z0-9]{6}`).Draw(t, "userID"))
		totalAmount := int64(rapid.IntRange(100, 1000000).Draw(t, "totalAmount"))

		// Create an order directly in the repository with the chosen status
		order := domain.Order{
			ID:            orderID,
			BakeryID:      "bakery-1",
			UserID:        userID,
			Status:        status,
			TotalAmount:   totalAmount,
			PaymentMethod: domain.PaymentMethodOnline,
			Items: []domain.OrderItem{
				{ProductID: "prod-1", ProductName: "Croissant", Quantity: 1, UnitPrice: totalAmount, Subtotal: totalAmount},
			},
			ScheduledDay:  domain.Monday,
			ScheduledTime: domain.TimeSlot{StartTime: domain.TimeOfDay{Hour: 10, Minute: 0}, EndTime: domain.TimeOfDay{Hour: 11, Minute: 0}},
			CreatedAt:     time.Date(2025, 1, 14, 10, 0, 0, 0, time.UTC),
			UpdatedAt:     time.Date(2025, 1, 14, 10, 0, 0, 0, time.UTC),
		}
		if err := orderRepo.Save(context.Background(), &order); err != nil {
			t.Fatalf("failed to seed order: %v", err)
		}

		// Call DeleteOrder with the correct owner
		err := svc.DeleteOrder(context.Background(), orderID, userID)
		if err != nil {
			t.Fatalf("DeleteOrder failed unexpectedly: %v", err)
		}

		// Property: refund MUST have been called with correct orderID and amount
		refunds := trackingPay.getRefundCalls()
		if len(refunds) != 1 {
			t.Fatalf("expected exactly 1 refund call, got %d", len(refunds))
		}
		if refunds[0].OrderID != orderID {
			t.Fatalf("refund called with orderID %q, expected %q", refunds[0].OrderID, orderID)
		}
		if refunds[0].Amount != totalAmount {
			t.Fatalf("refund called with amount %d, expected %d", refunds[0].Amount, totalAmount)
		}
	})
}
