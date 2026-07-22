package domain

import "fmt"

// Allergen represents a valid EU-regulated allergen identifier.
type Allergen string

const (
	AllergenGluten      Allergen = "gluten"
	AllergenCrustaceans Allergen = "crustaceans"
	AllergenEggs        Allergen = "eggs"
	AllergenFish        Allergen = "fish"
	AllergenPeanuts     Allergen = "peanuts"
	AllergenSoy         Allergen = "soy"
	AllergenDairy       Allergen = "dairy"
	AllergenNuts        Allergen = "nuts"
	AllergenCelery      Allergen = "celery"
	AllergenMustard     Allergen = "mustard"
	AllergenSesame      Allergen = "sesame"
	AllergenSulphites   Allergen = "sulphites"
	AllergenLupin       Allergen = "lupin"
	AllergenMolluscs    Allergen = "molluscs"
)

// maxAllergens is the maximum number of allergen entries allowed per product.
const maxAllergens = 20

// ValidAllergens is the complete set of valid allergen identifiers for O(1) membership checks.
var ValidAllergens = map[Allergen]bool{
	AllergenGluten:      true,
	AllergenCrustaceans: true,
	AllergenEggs:        true,
	AllergenFish:        true,
	AllergenPeanuts:     true,
	AllergenSoy:         true,
	AllergenDairy:       true,
	AllergenNuts:        true,
	AllergenCelery:      true,
	AllergenMustard:     true,
	AllergenSesame:      true,
	AllergenSulphites:   true,
	AllergenLupin:       true,
	AllergenMolluscs:    true,
}

// ValidateAllergens checks that all provided values are in the valid set,
// are unique, and the array has at most maxAllergens elements.
func ValidateAllergens(allergens []string) error {
	if len(allergens) > maxAllergens {
		return fmt.Errorf("allergens array exceeds maximum of %d elements", maxAllergens)
	}

	seen := make(map[string]bool, len(allergens))
	for _, a := range allergens {
		if !ValidAllergens[Allergen(a)] {
			return fmt.Errorf("invalid allergen: %q", a)
		}
		if seen[a] {
			return fmt.Errorf("duplicate allergen: %q", a)
		}
		seen[a] = true
	}

	return nil
}

// ValidateHealthScore checks that score is nil or in the range [1, 5].
func ValidateHealthScore(score *int) error {
	if score == nil {
		return nil
	}
	if *score < 1 || *score > 5 {
		return fmt.Errorf("health score must be between 1 and 5, got %d", *score)
	}
	return nil
}
