import { apiFetch } from './client';
import type { ListResponse } from '../types/bakery';

/** Item within a recurring order */
export interface RecurringOrderItem {
  productId: string;
  productName: string;
  quantity: number;
  unitPrice: number;
  subtotal: number;
}

/** Recurring order as returned by the API */
export interface RecurringOrder {
  id: string;
  bakeryId: string;
  bakeryName: string;
  items: RecurringOrderItem[];
  scheduledDay: string;
  scheduledTime: { startTime: string; endTime: string };
  frequency: string;
  selectionMode: string;
  active: boolean;
  createdAt: string;
  updatedAt: string;
}

/** User profile with holiday settings */
export interface UserProfile {
  id: string;
  username: string;
  role: number;
  holidayMode: boolean;
  holidayFrom?: string;
  holidayTo?: string;
  favoriteProducts: string[];
}

/** Holiday update payload */
export interface HolidayUpdate {
  holidayMode: boolean;
  holidayFrom?: string;
  holidayTo?: string;
}

/** Payload for creating a new recurring order */
export interface CreateRecurringOrderPayload {
  bakeryId: string;
  items: { productId: string; quantity: number }[];
  scheduledDay: string;
  scheduledTime: { startTime: string; endTime: string };
  frequency: string;
  selectionMode: string;
}

/** Create a new recurring order */
export function createRecurringOrder(data: CreateRecurringOrderPayload): Promise<RecurringOrder> {
  return apiFetch<RecurringOrder>('/recurring-orders', {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

/** Fetch paginated list of the current user's recurring orders */
export function fetchRecurringOrders(page?: number): Promise<ListResponse<RecurringOrder>> {
  const params = new URLSearchParams();
  if (page) params.set('page', String(page));
  const query = params.toString();
  return apiFetch<ListResponse<RecurringOrder>>(`/recurring-orders${query ? `?${query}` : ''}`);
}

/** Pause a recurring order */
export function pauseRecurringOrder(id: string): Promise<void> {
  return apiFetch<void>(`/recurring-orders/${id}/pause`, { method: 'PUT' });
}

/** Resume a recurring order */
export function resumeRecurringOrder(id: string): Promise<void> {
  return apiFetch<void>(`/recurring-orders/${id}/resume`, { method: 'PUT' });
}

/** Delete a recurring order */
export function deleteRecurringOrder(id: string): Promise<void> {
  return apiFetch<void>(`/recurring-orders/${id}`, { method: 'DELETE' });
}

/** Fetch the current user's profile */
export function fetchProfile(): Promise<UserProfile> {
  return apiFetch<UserProfile>('/user/profile');
}

/** Update holiday mode settings */
export function updateHoliday(data: HolidayUpdate): Promise<UserProfile> {
  return apiFetch<UserProfile>('/user/holiday', {
    method: 'PUT',
    body: JSON.stringify(data),
  });
}

/** Favorites response from the API */
export interface FavoritesResponse {
  productIds: string[];
}

/** Fetch the current user's favorite product IDs */
export function fetchFavorites(): Promise<string[]> {
  return apiFetch<FavoritesResponse>('/user/favorites').then((res) => res.productIds);
}

/** Update the user's favorite products list */
export function updateFavorites(productIds: string[]): Promise<void> {
  return apiFetch<void>('/user/favorites', {
    method: 'PUT',
    body: JSON.stringify({ productIds }),
  });
}
