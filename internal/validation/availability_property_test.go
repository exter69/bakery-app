package validation

import (
	"fmt"
	"testing"

	"github.com/lucatorrekens/bakery-app/internal/domain"
	"pgregory.net/rapid"
)

// **Validates: Requirements 6.5**

// genProductID generates a random product ID string.
func genProductID() *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		return fmt.Sprintf("prod-%d", rapid.IntRange(1, 10000).Draw(t, "id"))
	})
}

// genProduct generates a Product with the given ID and availability.
func genProduct(id string, available bool) domain.Product {
	return domain.Product{
		ID:          id,
		BakeryID:    "bakery-1",
		Name:        "Product " + id,
		Description: "A test product",
		Price:       100,
		Category:    "pastry",
		IsAvailable: available,
	}
}

// genAvailableProductList generates a non-empty list of products that are all available.
func genAvailableProductList() *rapid.Generator[[]domain.Product] {
	return rapid.Custom(func(t *rapid.T) []domain.Product {
		count := rapid.IntRange(1, 10).Draw(t, "productCount")
		products := make([]domain.Product, count)
		for i := 0; i < count; i++ {
			id := fmt.Sprintf("prod-%d", i+1)
			products[i] = genProduct(id, true)
		}
		return products
	})
}

// genOrderItemsFromProducts generates order items that only reference products in the given list.
func genOrderItemsFromProducts(products []domain.Product) *rapid.Generator[[]domain.OrderItem] {
	return rapid.Custom(func(t *rapid.T) []domain.OrderItem {
		itemCount := rapid.IntRange(1, len(products)).Draw(t, "itemCount")
		items := make([]domain.OrderItem, itemCount)
		for i := 0; i < itemCount; i++ {
			prodIdx := rapid.IntRange(0, len(products)-1).Draw(t, fmt.Sprintf("prodIdx_%d", i))
			items[i] = domain.OrderItem{
				ProductID: products[prodIdx].ID,
				Quantity:  rapid.IntRange(1, 10).Draw(t, fmt.Sprintf("qty_%d", i)),
				UnitPrice: int64(rapid.IntRange(100, 5000).Draw(t, fmt.Sprintf("price_%d", i))),
			}
		}
		return items
	})
}

func TestProperty_ProductAvailability_AcceptsAllAvailable(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a list of available products and order items referencing them
		products := genAvailableProductList().Draw(t, "products")
		items := genOrderItemsFromProducts(products).Draw(t, "items")

		result := ValidateProductAvailability(items, products)

		if result.HasErrors() {
			t.Fatalf("expected no errors when all products are available, got: %v", result.Errors)
		}
	})
}

func TestProperty_ProductAvailability_RejectsUnavailableProducts(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a mix of available and unavailable products, ensuring at least one is unavailable
		availableCount := rapid.IntRange(0, 5).Draw(t, "availableCount")
		unavailableCount := rapid.IntRange(1, 5).Draw(t, "unavailableCount")

		products := make([]domain.Product, 0, availableCount+unavailableCount)
		for i := 0; i < availableCount; i++ {
			id := fmt.Sprintf("avail-%d", i+1)
			products = append(products, genProduct(id, true))
		}
		for i := 0; i < unavailableCount; i++ {
			id := fmt.Sprintf("unavail-%d", i+1)
			products = append(products, genProduct(id, false))
		}

		// Generate items that reference at least one unavailable product
		unavailableIdx := rapid.IntRange(availableCount, len(products)-1).Draw(t, "unavailableItemIdx")
		items := []domain.OrderItem{
			{
				ProductID: products[unavailableIdx].ID,
				Quantity:  rapid.IntRange(1, 10).Draw(t, "qty"),
				UnitPrice: int64(rapid.IntRange(100, 5000).Draw(t, "price")),
			},
		}

		// Optionally add more items referencing any product
		extraItems := rapid.IntRange(0, 3).Draw(t, "extraItems")
		for i := 0; i < extraItems; i++ {
			prodIdx := rapid.IntRange(0, len(products)-1).Draw(t, fmt.Sprintf("extraProdIdx_%d", i))
			items = append(items, domain.OrderItem{
				ProductID: products[prodIdx].ID,
				Quantity:  rapid.IntRange(1, 10).Draw(t, fmt.Sprintf("extraQty_%d", i)),
				UnitPrice: int64(rapid.IntRange(100, 5000).Draw(t, fmt.Sprintf("extraPrice_%d", i))),
			})
		}

		result := ValidateProductAvailability(items, products)

		if !result.HasErrors() {
			t.Fatal("expected errors when items reference unavailable products, got none")
		}

		// Verify at least one error mentions productId field
		foundProductErr := false
		for _, err := range result.Errors {
			if hasSuffix(err.Field, ".productId") {
				foundProductErr = true
				break
			}
		}
		if !foundProductErr {
			t.Fatalf("expected a .productId error, got: %v", result.Errors)
		}
	})
}

func TestProperty_ProductAvailability_RejectsNonExistentProducts(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a product list
		productCount := rapid.IntRange(1, 5).Draw(t, "productCount")
		products := make([]domain.Product, productCount)
		for i := 0; i < productCount; i++ {
			id := fmt.Sprintf("existing-%d", i+1)
			products[i] = genProduct(id, true)
		}

		// Generate items that reference at least one non-existent product
		nonExistentID := fmt.Sprintf("nonexistent-%d", rapid.IntRange(1, 10000).Draw(t, "nonExistentID"))
		items := []domain.OrderItem{
			{
				ProductID: nonExistentID,
				Quantity:  rapid.IntRange(1, 10).Draw(t, "qty"),
				UnitPrice: int64(rapid.IntRange(100, 5000).Draw(t, "price")),
			},
		}

		// Optionally add items that reference existing products
		extraItems := rapid.IntRange(0, 3).Draw(t, "extraItems")
		for i := 0; i < extraItems; i++ {
			prodIdx := rapid.IntRange(0, len(products)-1).Draw(t, fmt.Sprintf("existProdIdx_%d", i))
			items = append(items, domain.OrderItem{
				ProductID: products[prodIdx].ID,
				Quantity:  rapid.IntRange(1, 10).Draw(t, fmt.Sprintf("existQty_%d", i)),
				UnitPrice: int64(rapid.IntRange(100, 5000).Draw(t, fmt.Sprintf("existPrice_%d", i))),
			})
		}

		result := ValidateProductAvailability(items, products)

		if !result.HasErrors() {
			t.Fatal("expected errors when items reference non-existent products, got none")
		}

		// Verify at least one error mentions "not found"
		foundNotFound := false
		for _, err := range result.Errors {
			if len(err.Message) >= 9 && contains(err.Message, "not found") {
				foundNotFound = true
				break
			}
		}
		if !foundNotFound {
			t.Fatalf("expected a 'not found' error message, got: %v", result.Errors)
		}
	})
}

// contains checks if substr is present in s.
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// hasSuffix checks if s ends with suffix.
func hasSuffix(s, suffix string) bool {
	if len(s) < len(suffix) {
		return false
	}
	return s[len(s)-len(suffix):] == suffix
}
