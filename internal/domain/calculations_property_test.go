package domain

import (
	"testing"

	"pgregory.net/rapid"
)

// TestSubtotalCorrectness verifies that after calling CalculateOrderTotal,
// each item's Subtotal equals quantity × unitPrice.
// **Validates: Requirements 6.2**
func TestSubtotalCorrectness(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a random slice of OrderItems with valid quantities and prices
		numItems := rapid.IntRange(1, 20).Draw(t, "numItems")
		items := make([]OrderItem, numItems)
		for i := range items {
			items[i] = OrderItem{
				ProductID:   "product-" + rapid.StringMatching(`[a-z]{5}`).Draw(t, "productId"),
				ProductName: "Product " + rapid.StringMatching(`[a-z]{3,10}`).Draw(t, "productName"),
				Quantity:    rapid.IntRange(1, 999).Draw(t, "quantity"),
				UnitPrice:   int64(rapid.IntRange(1, 100000).Draw(t, "unitPrice")),
			}
		}

		// Call the function under test
		CalculateOrderTotal(items)

		// Verify: each item's subtotal must equal quantity × unitPrice
		for i, item := range items {
			expected := int64(item.Quantity) * item.UnitPrice
			if item.Subtotal != expected {
				t.Fatalf("item[%d]: subtotal = %d, want %d (quantity=%d, unitPrice=%d)",
					i, item.Subtotal, expected, item.Quantity, item.UnitPrice)
			}
		}
	})
}
