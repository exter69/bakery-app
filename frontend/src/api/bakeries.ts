import { apiFetch } from './client';
import type { Bakery, BakeryCard, ListResponse, Menu } from '../types/bakery';

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
