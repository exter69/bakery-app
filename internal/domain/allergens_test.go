package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidAllergens_Contains14Entries(t *testing.T) {
	assert.Equal(t, 14, len(ValidAllergens))
}

func TestValidateAllergens_EmptySlice(t *testing.T) {
	err := ValidateAllergens([]string{})
	require.NoError(t, err)
}

func TestValidateAllergens_NilSlice(t *testing.T) {
	err := ValidateAllergens(nil)
	require.NoError(t, err)
}

func TestValidateAllergens_ValidSingle(t *testing.T) {
	err := ValidateAllergens([]string{"gluten"})
	require.NoError(t, err)
}

func TestValidateAllergens_AllValid(t *testing.T) {
	all := []string{
		"gluten", "crustaceans", "eggs", "fish", "peanuts", "soy", "dairy",
		"nuts", "celery", "mustard", "sesame", "sulphites", "lupin", "molluscs",
	}
	err := ValidateAllergens(all)
	require.NoError(t, err)
}

func TestValidateAllergens_InvalidValue(t *testing.T) {
	err := ValidateAllergens([]string{"gluten", "wheat"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid allergen")
	assert.Contains(t, err.Error(), "wheat")
}

func TestValidateAllergens_Duplicate(t *testing.T) {
	err := ValidateAllergens([]string{"eggs", "eggs"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate allergen")
	assert.Contains(t, err.Error(), "eggs")
}

func TestValidateAllergens_ExceedsMax(t *testing.T) {
	// 21 elements (all 14 valid + 7 repeated to exceed 20)
	allergens := []string{
		"gluten", "crustaceans", "eggs", "fish", "peanuts", "soy", "dairy",
		"nuts", "celery", "mustard", "sesame", "sulphites", "lupin", "molluscs",
		"gluten", "crustaceans", "eggs", "fish", "peanuts", "soy", "dairy",
	}
	err := ValidateAllergens(allergens)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds maximum")
}

func TestValidateHealthScore_Nil(t *testing.T) {
	err := ValidateHealthScore(nil)
	require.NoError(t, err)
}

func TestValidateHealthScore_Valid(t *testing.T) {
	for _, v := range []int{1, 2, 3, 4, 5} {
		score := v
		err := ValidateHealthScore(&score)
		require.NoError(t, err, "score %d should be valid", v)
	}
}

func TestValidateHealthScore_TooLow(t *testing.T) {
	score := 0
	err := ValidateHealthScore(&score)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must be between 1 and 5")
}

func TestValidateHealthScore_TooHigh(t *testing.T) {
	score := 6
	err := ValidateHealthScore(&score)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must be between 1 and 5")
}

func TestValidateHealthScore_Negative(t *testing.T) {
	score := -1
	err := ValidateHealthScore(&score)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must be between 1 and 5")
}
