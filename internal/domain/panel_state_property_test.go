package domain

import (
	"testing"

	"pgregory.net/rapid"
)

// TestPanelStateResetOnClose verifies that closing a panel (calling ResetPanelState)
// always resets its internal state to the initial empty state, regardless of what
// the panel state was before closing.
// **Validates: Requirements 9.3**
func TestPanelStateResetOnClose(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a random panel state simulating an open panel with user selections
		numItems := rapid.IntRange(0, 10).Draw(t, "numItems")
		items := make([]OrderItem, numItems)
		for i := range items {
			items[i] = OrderItem{
				ProductID:   "product-" + rapid.StringMatching(`[a-z]{5}`).Draw(t, "productId"),
				ProductName: "Product " + rapid.StringMatching(`[a-z]{3,10}`).Draw(t, "productName"),
				Quantity:    rapid.IntRange(1, 999).Draw(t, "quantity"),
				UnitPrice:   int64(rapid.IntRange(1, 100000).Draw(t, "unitPrice")),
				Subtotal:    int64(rapid.IntRange(1, 99900000).Draw(t, "subtotal")),
			}
		}

		days := []string{"", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}
		selectedDay := rapid.SampledFrom(days).Draw(t, "selectedDay")

		startTime := rapid.SampledFrom([]string{"", "08:00", "09:30", "10:00", "12:00", "14:30", "16:00"}).Draw(t, "startTime")
		endTime := rapid.SampledFrom([]string{"", "09:00", "10:30", "11:00", "13:00", "15:30", "17:00"}).Draw(t, "endTime")

		isOpen := rapid.Bool().Draw(t, "isOpen")

		state := PanelState{
			SelectedDay: selectedDay,
			StartTime:   startTime,
			EndTime:     endTime,
			Items:       items,
			IsOpen:      isOpen,
		}

		// Reset the panel state (simulating panel close)
		result := ResetPanelState(state)

		// Verify: result must equal the initial empty state
		expected := InitialPanelState()

		if result.SelectedDay != expected.SelectedDay {
			t.Fatalf("SelectedDay: got %q, want %q", result.SelectedDay, expected.SelectedDay)
		}
		if result.StartTime != expected.StartTime {
			t.Fatalf("StartTime: got %q, want %q", result.StartTime, expected.StartTime)
		}
		if result.EndTime != expected.EndTime {
			t.Fatalf("EndTime: got %q, want %q", result.EndTime, expected.EndTime)
		}
		if result.IsOpen != expected.IsOpen {
			t.Fatalf("IsOpen: got %v, want %v", result.IsOpen, expected.IsOpen)
		}
		if len(result.Items) != 0 {
			t.Fatalf("Items: got %d items, want 0", len(result.Items))
		}
	})
}
