package domain

import (
	"testing"

	"pgregory.net/rapid"
)

// TestProductSelectionAddsToItems verifies that clicking a product during
// selection mode either increases the item list length by 1 (new product)
// or increments the existing product's quantity (already selected product).
// **Validates: Requirements 3.6, 5.5**
func TestProductSelectionAddsToItems(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a pool of distinct products to click from
		numProducts := rapid.IntRange(1, 10).Draw(t, "numProducts")
		type productDef struct {
			id    string
			name  string
			price int64
		}
		products := make([]productDef, numProducts)
		for i := range products {
			products[i] = productDef{
				id:    rapid.StringMatching(`prod-[a-z]{4}`).Draw(t, "productId"),
				name:  "Product " + rapid.StringMatching(`[A-Z][a-z]{2,8}`).Draw(t, "productName"),
				price: int64(rapid.IntRange(50, 50000).Draw(t, "price")),
			}
		}

		// Generate a random sequence of product clicks
		numClicks := rapid.IntRange(1, 30).Draw(t, "numClicks")
		var items []OrderItem

		for click := 0; click < numClicks; click++ {
			// Pick a random product to click
			idx := rapid.IntRange(0, numProducts-1).Draw(t, "clickIdx")
			p := products[idx]

			itemsBefore := make([]OrderItem, len(items))
			copy(itemsBefore, items)
			lenBefore := len(items)

			// Check if the product was already in the list
			var existedBefore bool
			var qtyBefore int
			for _, item := range itemsBefore {
				if item.ProductID == p.id {
					existedBefore = true
					qtyBefore = item.Quantity
					break
				}
			}

			// Perform the click
			items = AddProductToSelection(items, p.id, p.name, p.price)

			// Property: the clicked product must now be in the list
			var found bool
			var qtyAfter int
			for _, item := range items {
				if item.ProductID == p.id {
					found = true
					qtyAfter = item.Quantity
					break
				}
			}
			if !found {
				t.Fatalf("click %d: product %s not found in items after click", click, p.id)
			}

			if existedBefore {
				// Property: list length stays the same, quantity incremented by 1
				if len(items) != lenBefore {
					t.Fatalf("click %d: existing product %s — expected list length %d, got %d",
						click, p.id, lenBefore, len(items))
				}
				if qtyAfter != qtyBefore+1 {
					t.Fatalf("click %d: existing product %s — expected quantity %d, got %d",
						click, p.id, qtyBefore+1, qtyAfter)
				}
			} else {
				// Property: list length increases by 1, new item has quantity 1
				if len(items) != lenBefore+1 {
					t.Fatalf("click %d: new product %s — expected list length %d, got %d",
						click, p.id, lenBefore+1, len(items))
				}
				if qtyAfter != 1 {
					t.Fatalf("click %d: new product %s — expected quantity 1, got %d",
						click, p.id, qtyAfter)
				}
			}
		}
	})
}

// TestProductSelectionSameProductQuantity verifies that after N clicks of
// the same product, the quantity equals N.
// **Validates: Requirements 3.6, 5.5**
func TestProductSelectionSameProductQuantity(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		productID := rapid.StringMatching(`prod-[a-z]{4}`).Draw(t, "productId")
		productName := "Croissant"
		unitPrice := int64(rapid.IntRange(100, 10000).Draw(t, "price"))
		numClicks := rapid.IntRange(1, 50).Draw(t, "numClicks")

		var items []OrderItem
		for i := 0; i < numClicks; i++ {
			items = AddProductToSelection(items, productID, productName, unitPrice)
		}

		// Property: after N clicks of the same product, quantity == N
		if len(items) != 1 {
			t.Fatalf("expected 1 item in list after clicking same product %d times, got %d", numClicks, len(items))
		}
		if items[0].Quantity != numClicks {
			t.Fatalf("expected quantity %d after %d clicks, got %d", numClicks, numClicks, items[0].Quantity)
		}
	})
}
