package memory

import (
	"context"
	"fmt"
	"sort"
	"testing"

	"github.com/lucatorrekens/bakery-app/internal/domain"
	"pgregory.net/rapid"
)

// validAllergenList is the complete list of valid EU-regulated allergens for use in generators.
var validAllergenList = []string{
	"gluten", "crustaceans", "eggs", "fish", "peanuts", "soy", "dairy",
	"nuts", "celery", "mustard", "sesame", "sulphites", "lupin", "molluscs",
}

// genAllergenSubset generates a random valid subset of allergens (0-14 items, no duplicates).
func genAllergenSubset(t *rapid.T) []string {
	// Generate a random subset by independently including/excluding each allergen
	subset := make([]string, 0)
	for _, a := range validAllergenList {
		include := rapid.Bool().Draw(t, "include_"+a)
		if include {
			subset = append(subset, a)
		}
	}
	return subset
}

// genHealthScore generates a valid health score: nil or 1-5.
func genHealthScore(t *rapid.T) *int {
	isNil := rapid.Bool().Draw(t, "isNil")
	if isNil {
		return nil
	}
	v := rapid.IntRange(1, 5).Draw(t, "score")
	return &v
}

// TestPropertyAllergenDataRoundTrip verifies that creating a product with a valid
// allergen subset and then fetching it returns the exact same set.
//
// **Validates: Requirements 1.1, 1.3, 9.1, 9.3, 9.7, 9.9**
func TestPropertyAllergenDataRoundTrip(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		repo := NewBakeryRepo()
		repo.SeedBakery(domain.Bakery{ID: "b1"})

		allergens := genAllergenSubset(t)
		productID := fmt.Sprintf("p-%d", rapid.IntRange(1, 100000).Draw(t, "productID"))

		product := &domain.Product{
			ID:        productID,
			BakeryID:  "b1",
			Name:      "Test Product",
			Allergens: allergens,
		}

		err := repo.CreateProduct(context.Background(), product)
		if err != nil {
			t.Fatalf("CreateProduct failed: %v", err)
		}

		got, err := repo.GetProductByID(context.Background(), productID)
		if err != nil {
			t.Fatalf("GetProductByID failed: %v", err)
		}
		if got == nil {
			t.Fatal("GetProductByID returned nil")
		}

		// Both should be non-nil (repo normalizes nil to empty slice)
		if got.Allergens == nil {
			t.Fatal("fetched Allergens is nil, expected non-nil slice")
		}

		// Sort both for comparison since order may vary
		expectedSorted := make([]string, len(allergens))
		copy(expectedSorted, allergens)
		sort.Strings(expectedSorted)

		gotSorted := make([]string, len(got.Allergens))
		copy(gotSorted, got.Allergens)
		sort.Strings(gotSorted)

		if len(gotSorted) != len(expectedSorted) {
			t.Fatalf("allergen count mismatch: expected %d, got %d", len(expectedSorted), len(gotSorted))
		}
		for i := range expectedSorted {
			if gotSorted[i] != expectedSorted[i] {
				t.Fatalf("allergen[%d] mismatch: expected %q, got %q", i, expectedSorted[i], gotSorted[i])
			}
		}
	})
}

// TestPropertyHealthScoreDataRoundTrip verifies that creating a product with a valid
// health score (nil or 1-5) and then fetching it returns the exact same value.
//
// **Validates: Requirements 1.2, 1.4, 9.2, 9.4, 9.8, 9.9**
func TestPropertyHealthScoreDataRoundTrip(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		repo := NewBakeryRepo()
		repo.SeedBakery(domain.Bakery{ID: "b1"})

		healthScore := genHealthScore(t)
		productID := fmt.Sprintf("p-%d", rapid.IntRange(1, 100000).Draw(t, "productID"))

		product := &domain.Product{
			ID:          productID,
			BakeryID:    "b1",
			Name:        "Test Product",
			HealthScore: healthScore,
		}

		err := repo.CreateProduct(context.Background(), product)
		if err != nil {
			t.Fatalf("CreateProduct failed: %v", err)
		}

		got, err := repo.GetProductByID(context.Background(), productID)
		if err != nil {
			t.Fatalf("GetProductByID failed: %v", err)
		}
		if got == nil {
			t.Fatal("GetProductByID returned nil")
		}

		// Verify health score round-trip
		if healthScore == nil {
			if got.HealthScore != nil {
				t.Fatalf("expected nil HealthScore, got %d", *got.HealthScore)
			}
		} else {
			if got.HealthScore == nil {
				t.Fatalf("expected HealthScore=%d, got nil", *healthScore)
			}
			if *got.HealthScore != *healthScore {
				t.Fatalf("expected HealthScore=%d, got %d", *healthScore, *got.HealthScore)
			}
		}
	})
}

// TestPropertyPartialUpdatePreservesOmittedFields verifies that updating only the name
// of a product preserves the allergens and health score fields.
//
// **Validates: Requirements 9.5, 9.6**
func TestPropertyPartialUpdatePreservesOmittedFields(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		repo := NewBakeryRepo()
		repo.SeedBakery(domain.Bakery{ID: "b1"})

		allergens := genAllergenSubset(t)
		healthScore := genHealthScore(t)
		productID := fmt.Sprintf("p-%d", rapid.IntRange(1, 100000).Draw(t, "productID"))

		// Create product with allergens and health score
		product := &domain.Product{
			ID:          productID,
			BakeryID:    "b1",
			Name:        "Original Name",
			Allergens:   allergens,
			HealthScore: healthScore,
		}

		err := repo.CreateProduct(context.Background(), product)
		if err != nil {
			t.Fatalf("CreateProduct failed: %v", err)
		}

		// Simulate partial update: only change the name, preserve allergens and healthScore
		// Fetch the product first, then modify only the name (mimics how the handler works)
		existing, err := repo.GetProductByID(context.Background(), productID)
		if err != nil {
			t.Fatalf("GetProductByID failed: %v", err)
		}

		newName := rapid.StringMatching(`[A-Za-z ]{3,20}`).Draw(t, "newName")
		existing.Name = newName

		err = repo.UpdateProduct(context.Background(), existing)
		if err != nil {
			t.Fatalf("UpdateProduct failed: %v", err)
		}

		// Fetch again and verify allergens + healthScore unchanged
		got, err := repo.GetProductByID(context.Background(), productID)
		if err != nil {
			t.Fatalf("GetProductByID after update failed: %v", err)
		}
		if got == nil {
			t.Fatal("GetProductByID returned nil after update")
		}

		// Verify name was updated
		if got.Name != newName {
			t.Fatalf("expected Name=%q after update, got %q", newName, got.Name)
		}

		// Verify allergens preserved (sort for comparison)
		expectedAllergens := allergens
		if expectedAllergens == nil {
			expectedAllergens = []string{}
		}

		expectedSorted := make([]string, len(expectedAllergens))
		copy(expectedSorted, expectedAllergens)
		sort.Strings(expectedSorted)

		gotSorted := make([]string, len(got.Allergens))
		copy(gotSorted, got.Allergens)
		sort.Strings(gotSorted)

		if len(gotSorted) != len(expectedSorted) {
			t.Fatalf("allergen count changed after name update: expected %d, got %d", len(expectedSorted), len(gotSorted))
		}
		for i := range expectedSorted {
			if gotSorted[i] != expectedSorted[i] {
				t.Fatalf("allergen[%d] changed after name update: expected %q, got %q", i, expectedSorted[i], gotSorted[i])
			}
		}

		// Verify health score preserved
		if healthScore == nil {
			if got.HealthScore != nil {
				t.Fatalf("expected nil HealthScore preserved, got %d", *got.HealthScore)
			}
		} else {
			if got.HealthScore == nil {
				t.Fatalf("expected HealthScore=%d preserved, got nil", *healthScore)
			}
			if *got.HealthScore != *healthScore {
				t.Fatalf("HealthScore changed after name update: expected %d, got %d", *healthScore, *got.HealthScore)
			}
		}
	})
}
