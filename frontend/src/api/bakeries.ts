import { apiFetch } from './client';
import type { Bakery, BakeryCard, ListResponse, Menu, Product } from '../types/bakery';

/** Fetch paginated bakery list, optionally sorted by distance from user */
export function fetchBakeries(
  page = 1,
  location?: { lat: number; lng: number },
): Promise<ListResponse<BakeryCard>> {
  const params = new URLSearchParams({ page: String(page) });
  if (location) {
    params.set('lat', String(location.lat));
    params.set('lng', String(location.lng));
  }
  return apiFetch<ListResponse<BakeryCard>>(`/bakeries?${params.toString()}`);
}

/** Fetch a single bakery by ID */
export function fetchBakery(id: string): Promise<Bakery> {
  return apiFetch<Bakery>(`/bakeries/${id}`);
}

/** Fetch menu for a bakery, grouped by category */
export function fetchMenu(bakeryId: string): Promise<Menu> {
  return apiFetch<Menu>(`/bakeries/${bakeryId}/menu`);
}

/** Parameters for product search */
export interface ProductSearchParams {
  q?: string;
  category?: string;
  excludeAllergens?: string[];
  minHealthScore?: number;
  page?: number;
}

/** A search result pairing a product with its bakery info */
export interface ProductSearchResult {
  product: Product;
  bakeryId: string;
  bakeryName: string;
}

/** Search products across all bakeries with optional filters */
export function searchProducts(params: ProductSearchParams): Promise<ListResponse<ProductSearchResult>> {
  const searchParams = new URLSearchParams();
  if (params.q) searchParams.set('q', params.q);
  if (params.category) searchParams.set('category', params.category);
  if (params.excludeAllergens?.length) searchParams.set('excludeAllergens', params.excludeAllergens.join(','));
  if (params.minHealthScore) searchParams.set('minHealthScore', String(params.minHealthScore));
  if (params.page) searchParams.set('page', String(params.page));
  return apiFetch<ListResponse<ProductSearchResult>>(`/products/search?${searchParams.toString()}`);
}
