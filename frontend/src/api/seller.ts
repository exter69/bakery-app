import { apiFetch, getToken, API_BASE } from './client';
import type { Bakery, DaySchedule, Product, ListResponse } from '../types/bakery';
import type { ScheduleEntry } from '../types/order';

/** Order type as returned by bakery-specific endpoints */
export type Order = ScheduleEntry;

/** Reservation type as returned by bakery-specific endpoints */
export type Reservation = ScheduleEntry;

/** Upload an image and return the stored URL */
export async function uploadImage(file: File, type: 'products' | 'bakeries'): Promise<string> {
  const formData = new FormData();
  formData.append('file', file);
  formData.append('type', type);

  const token = getToken();
  const response = await fetch(`${API_BASE}/uploads`, {
    method: 'POST',
    headers: token ? { Authorization: `Bearer ${token}` } : {},
    body: formData, // Don't set Content-Type — browser sets it with boundary
  });

  if (!response.ok) {
    const body = await response.json().catch(() => ({ message: 'Upload failed' }));
    throw new Error(body.message || 'Upload failed');
  }

  const result = await response.json();
  return result.url;
}

/** Fetch the current seller's bakery */
export async function fetchMyBakery(): Promise<Bakery | null> {
  try {
    return await apiFetch<Bakery>('/seller/bakery');
  } catch {
    return null;
  }
}

/** Update bakery info */
export function updateBakery(
  id: string,
  data: { name?: string; description?: string; address?: string; photoUrl?: string; googlePlaceId?: string }
): Promise<Bakery> {
  return apiFetch<Bakery>(`/bakeries/${id}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  });
}

/** Update bakery schedule */
export function updateSchedule(bakeryId: string, schedule: DaySchedule[]): Promise<Bakery> {
  return apiFetch<Bakery>(`/bakeries/${bakeryId}/schedule`, {
    method: 'PUT',
    body: JSON.stringify({ schedule }),
  });
}

/** Fetch products for a bakery */
export function fetchProducts(bakeryId: string): Promise<Product[]> {
  return apiFetch<Product[]>(`/bakeries/${bakeryId}/products`);
}

/** Create a new product */
export function createProduct(
  bakeryId: string,
  product: { name: string; description: string; price: number; photoUrl: string; category: string; allergens: string[]; healthScore: number | null }
): Promise<Product> {
  return apiFetch<Product>(`/bakeries/${bakeryId}/products`, {
    method: 'POST',
    body: JSON.stringify(product),
  });
}

/** Update an existing product */
export function updateProduct(id: string, updates: Partial<Product>): Promise<Product> {
  return apiFetch<Product>(`/products/${id}`, {
    method: 'PUT',
    body: JSON.stringify(updates),
  });
}

/** Delete a product */
export function deleteProduct(id: string): Promise<void> {
  return apiFetch<void>(`/products/${id}`, { method: 'DELETE' });
}

/** Fetch orders for a bakery */
export function fetchBakeryOrders(
  bakeryId: string,
  params?: { page?: number; status?: string; date?: string }
): Promise<ListResponse<Order>> {
  const searchParams = new URLSearchParams();
  if (params?.page) searchParams.set('page', String(params.page));
  if (params?.status) searchParams.set('status', params.status);
  if (params?.date) searchParams.set('date', params.date);
  const query = searchParams.toString();
  return apiFetch<ListResponse<Order>>(`/bakeries/${bakeryId}/orders${query ? `?${query}` : ''}`);
}

/** Fetch reservations for a bakery */
export function fetchBakeryReservations(
  bakeryId: string,
  params?: { page?: number; status?: string; date?: string }
): Promise<ListResponse<Reservation>> {
  const searchParams = new URLSearchParams();
  if (params?.page) searchParams.set('page', String(params.page));
  if (params?.status) searchParams.set('status', params.status);
  if (params?.date) searchParams.set('date', params.date);
  const query = searchParams.toString();
  return apiFetch<ListResponse<Reservation>>(`/bakeries/${bakeryId}/reservations${query ? `?${query}` : ''}`);
}

/** Update order status */
export function updateOrderStatus(orderId: string, status: string): Promise<Order> {
  return apiFetch<Order>(`/orders/${orderId}/status`, {
    method: 'PUT',
    body: JSON.stringify({ status }),
  });
}

/** Update reservation status */
export function updateReservationStatus(reservationId: string, status: string): Promise<Reservation> {
  return apiFetch<Reservation>(`/reservations/${reservationId}/status`, {
    method: 'PUT',
    body: JSON.stringify({ status }),
  });
}
