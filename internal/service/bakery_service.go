package service

import (
	"context"
	"fmt"
	"sort"

	"github.com/lucatorrekens/bakery-app/internal/domain"
)

const defaultRadius = 100.0 // km

// bakeryService is the concrete implementation of domain.BakeryService.
type bakeryService struct {
	repo domain.BakeryRepository
}

// NewBakeryService creates a new BakeryService backed by the given repository.
func NewBakeryService(repo domain.BakeryRepository) domain.BakeryService {
	return &bakeryService{repo: repo}
}

// ListBakeries returns a paginated list of bakeries with today's schedule.
// If lat/lng are provided, results are filtered by radius and sorted by distance.
func (s *bakeryService) ListBakeries(ctx context.Context, params domain.BakeryListParams) (*domain.ListResult[domain.BakeryWithDistance], error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 50
	}

	// If location provided, fetch all bakeries and filter/sort by distance in memory
	if params.Lat != nil && params.Lng != nil {
		return s.listBakeriesByDistance(ctx, params)
	}

	// No location — use standard paginated listing
	bakeries, total, err := s.repo.ListBakeries(ctx, params.PaginationParams)
	if err != nil {
		return nil, fmt.Errorf("listing bakeries: %w", err)
	}

	items := make([]domain.BakeryWithDistance, len(bakeries))
	for i, b := range bakeries {
		items[i] = domain.BakeryWithDistance{Bakery: b, Distance: nil}
	}

	return &domain.ListResult[domain.BakeryWithDistance]{
		Items:    items,
		Page:     params.Page,
		PageSize: params.PageSize,
		Total:    total,
	}, nil
}

// listBakeriesByDistance fetches all bakeries, calculates distance, filters by radius,
// sorts by distance, and then paginates.
func (s *bakeryService) listBakeriesByDistance(ctx context.Context, params domain.BakeryListParams) (*domain.ListResult[domain.BakeryWithDistance], error) {
	radius := params.Radius
	if radius <= 0 {
		radius = defaultRadius
	}

	// Fetch all bakeries (no pagination at repo level for distance sorting)
	allParams := domain.PaginationParams{Page: 1, PageSize: 10000}
	bakeries, _, err := s.repo.ListBakeries(ctx, allParams)
	if err != nil {
		return nil, fmt.Errorf("listing bakeries: %w", err)
	}

	lat := *params.Lat
	lng := *params.Lng

	// Calculate distance and filter by radius
	type bakeryDist struct {
		bakery   domain.Bakery
		distance float64
	}
	var filtered []bakeryDist
	for _, b := range bakeries {
		dist := domain.HaversineDistance(lat, lng, b.Latitude, b.Longitude)
		if dist <= radius {
			filtered = append(filtered, bakeryDist{bakery: b, distance: dist})
		}
	}

	// Sort by distance (closest first)
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].distance < filtered[j].distance
	})

	total := len(filtered)

	// Paginate
	start := (params.Page - 1) * params.PageSize
	if start >= total {
		return &domain.ListResult[domain.BakeryWithDistance]{
			Items:    []domain.BakeryWithDistance{},
			Page:     params.Page,
			PageSize: params.PageSize,
			Total:    total,
		}, nil
	}
	end := start + params.PageSize
	if end > total {
		end = total
	}

	items := make([]domain.BakeryWithDistance, 0, end-start)
	for _, bd := range filtered[start:end] {
		dist := bd.distance
		items = append(items, domain.BakeryWithDistance{
			Bakery:   bd.bakery,
			Distance: &dist,
		})
	}

	return &domain.ListResult[domain.BakeryWithDistance]{
		Items:    items,
		Page:     params.Page,
		PageSize: params.PageSize,
		Total:    total,
	}, nil
}

// GetMenu returns products grouped by category for a given bakery.
// Returns an error wrapping ErrBakeryNotFound if the bakery does not exist.
// Returns an empty map if the bakery exists but has no products.
func (s *bakeryService) GetMenu(ctx context.Context, bakeryID string) (map[string][]domain.Product, error) {
	// Verify bakery exists
	bakery, err := s.repo.GetBakery(ctx, bakeryID)
	if err != nil {
		return nil, fmt.Errorf("fetching bakery: %w", err)
	}
	if bakery == nil {
		return nil, ErrBakeryNotFound
	}

	// Fetch products
	products, err := s.repo.GetProductsByBakery(ctx, bakeryID)
	if err != nil {
		return nil, fmt.Errorf("fetching products: %w", err)
	}

	// Group by category
	menu := make(map[string][]domain.Product)
	for _, p := range products {
		menu[p.Category] = append(menu[p.Category], p)
	}

	return menu, nil
}

// SearchProducts searches products across bakeries with filters.
func (s *bakeryService) SearchProducts(ctx context.Context, params domain.ProductSearchParams) (*domain.ListResult[domain.ProductSearchResult], error) {
	if params.PageSize <= 0 {
		params.PageSize = 20
	}
	if params.Page < 1 {
		params.Page = 1
	}

	results, total, err := s.repo.SearchProducts(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("searching products: %w", err)
	}

	return &domain.ListResult[domain.ProductSearchResult]{
		Items:    results,
		Page:     params.Page,
		PageSize: params.PageSize,
		Total:    total,
	}, nil
}
