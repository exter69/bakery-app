package domain

import (
	"testing"

	"pgregory.net/rapid"
)

// validAllergenSlice is the list of all 14 valid allergen strings for use in generators.
var validAllergenSlice = []string{
	"gluten", "crustaceans", "eggs", "fish", "peanuts", "soy", "dairy",
	"nuts", "celery", "mustard", "sesame", "sulphites", "lupin", "molluscs",
}

// validAllergenSet for quick membership checking in generators.
var validAllergenSet = map[string]bool{
	"gluten": true, "crustaceans": true, "eggs": true, "fish": true,
	"peanuts": true, "soy": true, "dairy": true, "nuts": true,
	"celery": true, "mustard": true, "sesame": true, "sulphites": true,
	"lupin": true, "molluscs": true,
}

// TestPropertyInvalidAllergensRejected verifies that any slice containing at least one
// string NOT in the valid 14 allergens set is always rejected by ValidateAllergens.
//
// **Validates: Requirements 1.6, 9.10**
func TestPropertyInvalidAllergensRejected(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a random string that is NOT a valid allergen.
		invalidAllergen := rapid.StringMatching(`[a-z]{3,20}`).
			Filter(func(s string) bool { return !validAllergenSet[s] }).
			Draw(t, "invalidAllergen")

		// Optionally include some valid allergens before the invalid one.
		numValid := rapid.IntRange(0, 5).Draw(t, "numValidPrefix")
		allergens := make([]string, 0, numValid+1)
		for i := 0; i < numValid && i < len(validAllergenSlice); i++ {
			allergens = append(allergens, validAllergenSlice[i])
		}
		allergens = append(allergens, invalidAllergen)

		err := ValidateAllergens(allergens)
		if err == nil {
			t.Fatalf("expected error for invalid allergen %q in %v, got nil", invalidAllergen, allergens)
		}
	})
}

// TestPropertyValidAllergenSubsetsAccepted verifies that any subset of the 14 valid
// allergens (0 to 14 items, no duplicates) is always accepted by ValidateAllergens.
//
// **Validates: Requirements 1.6, 9.10**
func TestPropertyValidAllergenSubsetsAccepted(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a random subset by picking a random bitmask over 14 allergens.
		subset := make([]string, 0)
		for _, a := range validAllergenSlice {
			include := rapid.Bool().Draw(t, "include_"+a)
			if include {
				subset = append(subset, a)
			}
		}

		err := ValidateAllergens(subset)
		if err != nil {
			t.Fatalf("expected no error for valid subset %v, got: %v", subset, err)
		}
	})
}

// TestPropertyInvalidHealthScoresRejected verifies that any integer outside [1,5]
// is always rejected by ValidateHealthScore.
//
// **Validates: Requirements 1.7, 9.11**
func TestPropertyInvalidHealthScoresRejected(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate an integer that is NOT in [1, 5].
		score := rapid.IntRange(-1000, 1000).
			Filter(func(v int) bool { return v < 1 || v > 5 }).
			Draw(t, "invalidScore")

		err := ValidateHealthScore(&score)
		if err == nil {
			t.Fatalf("expected error for invalid health score %d, got nil", score)
		}
	})
}

// TestPropertyValidHealthScoresAccepted verifies that any integer in [1,5] or nil
// is always accepted by ValidateHealthScore.
//
// **Validates: Requirements 1.7, 9.11**
func TestPropertyValidHealthScoresAccepted(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		isNil := rapid.Bool().Draw(t, "isNil")
		if isNil {
			err := ValidateHealthScore(nil)
			if err != nil {
				t.Fatalf("expected no error for nil health score, got: %v", err)
			}
		} else {
			score := rapid.IntRange(1, 5).Draw(t, "validScore")
			err := ValidateHealthScore(&score)
			if err != nil {
				t.Fatalf("expected no error for valid health score %d, got: %v", score, err)
			}
		}
	})
}

// TestPropertyAllergenArrayExceedingMaxRejected verifies that any allergen array with
// more than 20 elements is always rejected, even if all elements are valid.
//
// **Validates: Requirements 1.6, 9.10**
func TestPropertyAllergenArrayExceedingMaxRejected(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a count > 20 (up to 50 for good coverage).
		count := rapid.IntRange(21, 50).Draw(t, "count")
		allergens := make([]string, count)
		for i := 0; i < count; i++ {
			// Cycle through valid allergens to fill array (duplicates are fine here,
			// the length check should fire first).
			allergens[i] = validAllergenSlice[i%len(validAllergenSlice)]
		}

		err := ValidateAllergens(allergens)
		if err == nil {
			t.Fatalf("expected error for allergen array of length %d, got nil", count)
		}
	})
}
