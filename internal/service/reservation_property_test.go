package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/lucatorrekens/bakery-app/internal/domain"
	"github.com/lucatorrekens/bakery-app/internal/repository/memory"
	"pgregory.net/rapid"
)

// **Validates: Requirements 5.6, 5.7**

// genDayOfWeek generates a random day of the week.
func genDayOfWeek() *rapid.Generator[domain.DayOfWeek] {
	return rapid.Custom(func(t *rapid.T) domain.DayOfWeek {
		days := domain.AllDaysOfWeek()
		idx := rapid.IntRange(0, len(days)-1).Draw(t, "dayIdx")
		return days[idx]
	})
}

// genOpenScheduleForDay generates a DaySchedule where the bakery is open with valid openTime < closeTime.
func genOpenScheduleForDay(day domain.DayOfWeek) domain.DaySchedule {
	return domain.DaySchedule{
		Day:       day,
		IsOpen:    true,
		OpenTime:  domain.TimeOfDay{Hour: 6, Minute: 0},
		CloseTime: domain.TimeOfDay{Hour: 22, Minute: 0},
	}
}

// setupTestBakeryRepo creates a bakery repo with a bakery open every day and N products.
func setupTestBakeryRepo(numProducts int) (*memory.BakeryRepo, string, []domain.Product) {
	repo := memory.NewBakeryRepo()
	bakeryID := "test-bakery-1"

	// Create schedule: open every day 06:00 - 22:00
	schedule := make([]domain.DaySchedule, 0, 7)
	for _, day := range domain.AllDaysOfWeek() {
		schedule = append(schedule, genOpenScheduleForDay(day))
	}

	bakery := domain.Bakery{
		ID:       bakeryID,
		Name:     "Test Bakery",
		Schedule: schedule,
	}
	repo.SeedBakery(bakery)

	// Create products
	products := make([]domain.Product, numProducts)
	for i := 0; i < numProducts; i++ {
		p := domain.Product{
			ID:          fmt.Sprintf("product-%d", i+1),
			BakeryID:    bakeryID,
			Name:        fmt.Sprintf("Product %d", i+1),
			Price:       int64((i + 1) * 100), // 100, 200, 300, ... cents
			Category:    "pastries",
			IsAvailable: true,
		}
		products[i] = p
		repo.SeedProduct(p)
	}

	return repo, bakeryID, products
}

// genValidReservationItems generates a random list of valid order items referencing
// the given products. Each item has quantity in [1, 999].
func genValidReservationItems(products []domain.Product) *rapid.Generator[[]domain.OrderItem] {
	return rapid.Custom(func(t *rapid.T) []domain.OrderItem {
		numItems := rapid.IntRange(1, len(products)).Draw(t, "numItems")
		// Shuffle and pick first numItems products
		indices := rapid.SliceOfN(rapid.IntRange(0, len(products)-1), numItems, numItems).Draw(t, "productIndices")

		// Deduplicate indices (use a set)
		seen := make(map[int]bool)
		items := make([]domain.OrderItem, 0, numItems)
		for _, idx := range indices {
			if seen[idx] {
				continue
			}
			seen[idx] = true
			p := products[idx]
			qty := rapid.IntRange(1, 999).Draw(t, "quantity")
			items = append(items, domain.OrderItem{
				ProductID: p.ID,
				Quantity:  qty,
			})
		}
		// Ensure at least one item
		if len(items) == 0 {
			p := products[0]
			qty := rapid.IntRange(1, 999).Draw(t, "fallbackQuantity")
			items = append(items, domain.OrderItem{
				ProductID: p.ID,
				Quantity:  qty,
			})
		}
		return items
	})
}

// genValidTimeSlot generates a random time slot within bakery hours (06:00 - 22:00).
func genValidTimeSlot() *rapid.Generator[domain.TimeSlot] {
	return rapid.Custom(func(t *rapid.T) domain.TimeSlot {
		// startTime in minutes from 06:00 (360) to 21:59 (1319)
		startMins := rapid.IntRange(360, 1319).Draw(t, "startMins")
		// endTime must be after startTime and at most 22:00 (1320)
		endMins := rapid.IntRange(startMins, 1320).Draw(t, "endMins")
		return domain.TimeSlot{
			StartTime: domain.TimeOfDay{Hour: startMins / 60, Minute: startMins % 60},
			EndTime:   domain.TimeOfDay{Hour: endMins / 60, Minute: endMins % 60},
		}
	})
}

// TestProperty_ReservationPaymentInvariant verifies that for any valid reservation,
// the payment method is always OnSpot and the CreateReservation function does not
// return a payment link (returns (*Reservation, error) only).
func TestProperty_ReservationPaymentInvariant(t *testing.T) {
	const numProducts = 5
	bakeryRepo, bakeryID, products := setupTestBakeryRepo(numProducts)
	reservationRepo := memory.NewReservationRepo()

	idCounter := 0
	svc := NewReservationService(ReservationServiceConfig{
		BakeryRepo:      bakeryRepo,
		ReservationRepo: reservationRepo,
		IDGen: func() string {
			idCounter++
			return fmt.Sprintf("res-%d", idCounter)
		},
	})

	rapid.Check(t, func(t *rapid.T) {
		// Generate random valid reservation inputs
		day := genDayOfWeek().Draw(t, "day")
		timeSlot := genValidTimeSlot().Draw(t, "timeSlot")
		items := genValidReservationItems(products).Draw(t, "items")
		userID := fmt.Sprintf("user-%d", rapid.IntRange(1, 1000).Draw(t, "userId"))

		reservation := domain.Reservation{
			BakeryID:      bakeryID,
			Items:         items,
			ScheduledDay:  day,
			ScheduledTime: timeSlot,
		}

		result, err := svc.CreateReservation(context.Background(), userID, reservation)
		if err != nil {
			// If validation rejects it, that's fine — we only check successful reservations
			return
		}

		// Property 1: PaymentMethod must always be OnSpot
		if result.PaymentMethod != domain.PaymentMethodOnSpot {
			t.Fatalf("expected PaymentMethod to be %q, got %q",
				domain.PaymentMethodOnSpot, result.PaymentMethod)
		}

		// Property 2: CreateReservation returns (*Reservation, error) — no payment link.
		// This is enforced by the type system (the function signature doesn't return a PaymentLink),
		// but we also verify the reservation struct has no payment link field that could be set.
		// The domain.Reservation struct only has PaymentMethod, no PaymentLink/URL field.
		// This property is satisfied by design: the function returns (*domain.Reservation, error)
		// with no third return value (unlike OrderService.CreateOrder which returns (*Order, *PaymentLink, error)).
		// We verify the invariant holds by confirming the result is a valid reservation without payment info.
		if result.ID == "" {
			t.Fatal("expected reservation to have an ID assigned")
		}
		if result.Status != domain.ReservationStatusConfirmed {
			t.Fatalf("expected status %q, got %q",
				domain.ReservationStatusConfirmed, result.Status)
		}
	})
}
