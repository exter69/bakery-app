package memory

import (
	"context"
	"testing"

	"github.com/lucatorrekens/bakery-app/internal/domain"
)

func TestCreateProduct_NilAllergens_DefaultsToEmptySlice(t *testing.T) {
	repo := NewBakeryRepo()
	repo.SeedBakery(domain.Bakery{ID: "b1"})

	product := &domain.Product{
		ID:        "p1",
		BakeryID:  "b1",
		Name:      "Croissant",
		Allergens: nil, // nil input
	}

	if err := repo.CreateProduct(context.Background(), product); err != nil {
		t.Fatalf("CreateProduct failed: %v", err)
	}

	got, err := repo.GetProductByID(context.Background(), "p1")
	if err != nil {
		t.Fatalf("GetProductByID failed: %v", err)
	}
	if got.Allergens == nil {
		t.Fatal("expected non-nil Allergens slice, got nil")
	}
	if len(got.Allergens) != 0 {
		t.Fatalf("expected empty Allergens slice, got %v", got.Allergens)
	}
}

func TestCreateProduct_WithAllergens_Persisted(t *testing.T) {
	repo := NewBakeryRepo()
	repo.SeedBakery(domain.Bakery{ID: "b1"})

	allergens := []string{"gluten", "dairy", "eggs"}
	product := &domain.Product{
		ID:        "p1",
		BakeryID:  "b1",
		Name:      "Cake",
		Allergens: allergens,
	}

	if err := repo.CreateProduct(context.Background(), product); err != nil {
		t.Fatalf("CreateProduct failed: %v", err)
	}

	got, err := repo.GetProductByID(context.Background(), "p1")
	if err != nil {
		t.Fatalf("GetProductByID failed: %v", err)
	}
	if len(got.Allergens) != 3 {
		t.Fatalf("expected 3 allergens, got %d", len(got.Allergens))
	}
	for i, a := range allergens {
		if got.Allergens[i] != a {
			t.Errorf("allergen[%d]: expected %q, got %q", i, a, got.Allergens[i])
		}
	}
}

func TestCreateProduct_WithHealthScore_Persisted(t *testing.T) {
	repo := NewBakeryRepo()
	repo.SeedBakery(domain.Bakery{ID: "b1"})

	score := 4
	product := &domain.Product{
		ID:          "p1",
		BakeryID:    "b1",
		Name:        "Bread",
		HealthScore: &score,
	}

	if err := repo.CreateProduct(context.Background(), product); err != nil {
		t.Fatalf("CreateProduct failed: %v", err)
	}

	got, err := repo.GetProductByID(context.Background(), "p1")
	if err != nil {
		t.Fatalf("GetProductByID failed: %v", err)
	}
	if got.HealthScore == nil {
		t.Fatal("expected non-nil HealthScore")
	}
	if *got.HealthScore != 4 {
		t.Fatalf("expected HealthScore=4, got %d", *got.HealthScore)
	}
}

func TestCreateProduct_NilHealthScore_Persisted(t *testing.T) {
	repo := NewBakeryRepo()
	repo.SeedBakery(domain.Bakery{ID: "b1"})

	product := &domain.Product{
		ID:          "p1",
		BakeryID:    "b1",
		Name:        "Bread",
		HealthScore: nil,
	}

	if err := repo.CreateProduct(context.Background(), product); err != nil {
		t.Fatalf("CreateProduct failed: %v", err)
	}

	got, err := repo.GetProductByID(context.Background(), "p1")
	if err != nil {
		t.Fatalf("GetProductByID failed: %v", err)
	}
	if got.HealthScore != nil {
		t.Fatalf("expected nil HealthScore, got %d", *got.HealthScore)
	}
}

func TestUpdateProduct_AllergensAndHealthScore_Persisted(t *testing.T) {
	repo := NewBakeryRepo()
	repo.SeedBakery(domain.Bakery{ID: "b1"})

	product := &domain.Product{
		ID:        "p1",
		BakeryID:  "b1",
		Name:      "Muffin",
		Allergens: []string{"gluten"},
	}
	if err := repo.CreateProduct(context.Background(), product); err != nil {
		t.Fatalf("CreateProduct failed: %v", err)
	}

	// Update with new allergens and health score
	score := 3
	updated := &domain.Product{
		ID:          "p1",
		BakeryID:    "b1",
		Name:        "Muffin",
		Allergens:   []string{"gluten", "eggs", "dairy"},
		HealthScore: &score,
	}
	if err := repo.UpdateProduct(context.Background(), updated); err != nil {
		t.Fatalf("UpdateProduct failed: %v", err)
	}

	got, err := repo.GetProductByID(context.Background(), "p1")
	if err != nil {
		t.Fatalf("GetProductByID failed: %v", err)
	}
	if len(got.Allergens) != 3 {
		t.Fatalf("expected 3 allergens after update, got %d", len(got.Allergens))
	}
	if got.HealthScore == nil || *got.HealthScore != 3 {
		t.Fatal("expected HealthScore=3 after update")
	}
}

func TestUpdateProduct_NilAllergens_DefaultsToEmptySlice(t *testing.T) {
	repo := NewBakeryRepo()
	repo.SeedBakery(domain.Bakery{ID: "b1"})

	product := &domain.Product{
		ID:        "p1",
		BakeryID:  "b1",
		Name:      "Scone",
		Allergens: []string{"gluten"},
	}
	if err := repo.CreateProduct(context.Background(), product); err != nil {
		t.Fatalf("CreateProduct failed: %v", err)
	}

	// Update with nil allergens
	updated := &domain.Product{
		ID:        "p1",
		BakeryID:  "b1",
		Name:      "Scone",
		Allergens: nil,
	}
	if err := repo.UpdateProduct(context.Background(), updated); err != nil {
		t.Fatalf("UpdateProduct failed: %v", err)
	}

	got, err := repo.GetProductByID(context.Background(), "p1")
	if err != nil {
		t.Fatalf("GetProductByID failed: %v", err)
	}
	if got.Allergens == nil {
		t.Fatal("expected non-nil Allergens after update with nil, got nil")
	}
	if len(got.Allergens) != 0 {
		t.Fatalf("expected empty Allergens after update with nil, got %v", got.Allergens)
	}
}

func TestGetProductsByBakery_AllergensNeverNil(t *testing.T) {
	repo := NewBakeryRepo()
	repo.SeedBakery(domain.Bakery{ID: "b1"})

	// Seed a product with nil allergens (simulates legacy data)
	repo.SeedProduct(domain.Product{
		ID:        "p1",
		BakeryID:  "b1",
		Name:      "Old Product",
		Allergens: nil, // SeedProduct normalizes this
	})

	products, err := repo.GetProductsByBakery(context.Background(), "b1")
	if err != nil {
		t.Fatalf("GetProductsByBakery failed: %v", err)
	}
	if len(products) != 1 {
		t.Fatalf("expected 1 product, got %d", len(products))
	}
	if products[0].Allergens == nil {
		t.Fatal("expected non-nil Allergens in listed product, got nil")
	}
}
