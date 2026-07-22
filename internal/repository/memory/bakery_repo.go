package memory

import (
	"context"
	"sync"

	"github.com/lucatorrekens/bakery-app/internal/domain"
)

// BakeryRepo is an in-memory implementation of domain.BakeryRepository.
type BakeryRepo struct {
	mu       sync.RWMutex
	bakeries map[string]domain.Bakery
	products map[string][]domain.Product // keyed by bakeryID
}

// NewBakeryRepo creates a new in-memory BakeryRepo.
func NewBakeryRepo() *BakeryRepo {
	return &BakeryRepo{
		bakeries: make(map[string]domain.Bakery),
		products: make(map[string][]domain.Product),
	}
}

// SeedBakery adds a bakery to the store (for testing/seeding).
func (r *BakeryRepo) SeedBakery(b domain.Bakery) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.bakeries[b.ID] = b
}

// SeedProduct adds a product to the store (for testing/seeding).
func (r *BakeryRepo) SeedProduct(p domain.Product) {
	r.mu.Lock()
	defer r.mu.Unlock()
	// Ensure Allergens is never stored as nil
	if p.Allergens == nil {
		p.Allergens = []string{}
	}
	r.products[p.BakeryID] = append(r.products[p.BakeryID], p)
}

// ListBakeries returns a paginated slice of bakeries.
func (r *BakeryRepo) ListBakeries(_ context.Context, params domain.PaginationParams) ([]domain.Bakery, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	all := make([]domain.Bakery, 0, len(r.bakeries))
	for _, b := range r.bakeries {
		all = append(all, b)
	}

	total := len(all)
	start := (params.Page - 1) * params.PageSize
	if start >= total {
		return []domain.Bakery{}, total, nil
	}
	end := start + params.PageSize
	if end > total {
		end = total
	}

	return all[start:end], total, nil
}

// GetBakery returns a bakery by ID, or nil if not found.
func (r *BakeryRepo) GetBakery(_ context.Context, id string) (*domain.Bakery, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	b, ok := r.bakeries[id]
	if !ok {
		return nil, nil
	}
	return &b, nil
}

// GetBakeryByOwner returns the bakery owned by the given user, or nil if not found.
func (r *BakeryRepo) GetBakeryByOwner(_ context.Context, ownerID string) (*domain.Bakery, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, b := range r.bakeries {
		if b.OwnerID == ownerID {
			return &b, nil
		}
	}
	return nil, nil
}

// UpdateBakery persists changes to a bakery.
func (r *BakeryRepo) UpdateBakery(_ context.Context, bakery *domain.Bakery) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.bakeries[bakery.ID]; !ok {
		return nil
	}
	r.bakeries[bakery.ID] = *bakery
	return nil
}

// GetProductsByBakery returns all products for a bakery.
func (r *BakeryRepo) GetProductsByBakery(_ context.Context, bakeryID string) ([]domain.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	products := r.products[bakeryID]
	if products == nil {
		return []domain.Product{}, nil
	}
	// Ensure Allergens is never nil on returned products
	result := make([]domain.Product, len(products))
	copy(result, products)
	for i := range result {
		if result[i].Allergens == nil {
			result[i].Allergens = []string{}
		}
	}
	return result, nil
}

// GetProductByID returns a single product by ID, or nil if not found.
func (r *BakeryRepo) GetProductByID(_ context.Context, id string) (*domain.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, products := range r.products {
		for i := range products {
			if products[i].ID == id {
				p := products[i]
				// Ensure Allergens is never nil on returned product
				if p.Allergens == nil {
					p.Allergens = []string{}
				}
				return &p, nil
			}
		}
	}
	return nil, nil
}

// CreateProduct persists a new product.
func (r *BakeryRepo) CreateProduct(_ context.Context, product *domain.Product) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Ensure Allergens is never stored as nil
	if product.Allergens == nil {
		product.Allergens = []string{}
	}

	r.products[product.BakeryID] = append(r.products[product.BakeryID], *product)
	return nil
}

// UpdateProduct persists changes to a product.
func (r *BakeryRepo) UpdateProduct(_ context.Context, product *domain.Product) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Ensure Allergens is never stored as nil
	if product.Allergens == nil {
		product.Allergens = []string{}
	}

	products := r.products[product.BakeryID]
	for i := range products {
		if products[i].ID == product.ID {
			products[i] = *product
			r.products[product.BakeryID] = products
			return nil
		}
	}
	return nil
}

// DeleteProduct removes a product by ID.
func (r *BakeryRepo) DeleteProduct(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for bakeryID, products := range r.products {
		for i := range products {
			if products[i].ID == id {
				r.products[bakeryID] = append(products[:i], products[i+1:]...)
				return nil
			}
		}
	}
	return nil
}
