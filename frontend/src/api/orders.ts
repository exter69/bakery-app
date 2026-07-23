import { apiFetch } from './client';
import type { ScheduleListResponse, ScheduleQueryParams } from '../types/order';

/** Item to include in order/reservation creation */
export interface CreateOrderItem {
  productId: string;
  quantity: number;
}

/** Request body for creating an order */
export interface CreateOrderRequest {
  bakeryId: string;
  items: CreateOrderItem[];
  scheduledDay: string;
  scheduledTime: { startTime: string; endTime: string };
  recurring?: boolean;
  recurringDays?: string[];
}

/** Response from creating an order (includes payment link) */
export interface CreateOrderResponse {
  id: string;
  status: string;
  totalAmount: number;
  paymentUrl: string;
}

/** Request body for creating a reservation */
export interface CreateReservationRequest {
  bakeryId: string;
  items: CreateOrderItem[];
  scheduledDay: string;
  scheduledTime: { startTime: string; endTime: string };
}

/** Response from creating a reservation */
export interface CreateReservationResponse {
  id: string;
  status: string;
  totalAmount: number;
}

/** Create a delivery order - returns a payment URL for redirection */
export function createOrder(request: CreateOrderRequest): Promise<CreateOrderResponse> {
  return apiFetch<CreateOrderResponse>('/orders', {
    method: 'POST',
    body: JSON.stringify(request),
  });
}

/** Create a reservation (payment on spot, no redirect) */
export function createReservation(request: CreateReservationRequest): Promise<CreateReservationResponse> {
  return apiFetch<CreateReservationResponse>('/reservations', {
    method: 'POST',
    body: JSON.stringify(request),
  });
}

/** Fetch paginated list of orders and reservations */
export function fetchScheduleEntries(params: ScheduleQueryParams = {}): Promise<ScheduleListResponse> {
  const searchParams = new URLSearchParams();

  if (params.page) searchParams.set('page', String(params.page));
  if (params.status) searchParams.set('status', params.status);
  if (params.sortBy) searchParams.set('sortBy', params.sortBy);
  if (params.sortDirection) searchParams.set('sortDirection', params.sortDirection);

  const query = searchParams.toString();
  return apiFetch<ScheduleListResponse>(`/orders${query ? `?${query}` : ''}`);
}

/** Delete an order by ID (sets status to Cancelled) */
export function deleteOrder(orderId: string): Promise<void> {
  return apiFetch<void>(`/orders/${orderId}`, { method: 'DELETE' });
}

/** Delete a reservation by ID (sets status to Cancelled) */
export function deleteReservation(reservationId: string): Promise<void> {
  return apiFetch<void>(`/reservations/${reservationId}`, { method: 'DELETE' });
}

/** Fetch order history (delivered orders + picked-up reservations) */
export function fetchOrderHistory(page?: number): Promise<ScheduleListResponse> {
  const params = new URLSearchParams();
  if (page) params.set('page', String(page));
  params.set('sortBy', 'createdAt');
  params.set('sortDirection', 'desc');
  return apiFetch<ScheduleListResponse>(`/orders?${params.toString()}`).then((response) => {
    // Filter to terminal states client-side (delivered orders + picked-up reservations)
    const filtered = response.items.filter(
      (item) => item.status === 'delivered' || item.status === 'picked_up'
    );
    return {
      ...response,
      items: filtered,
      total: filtered.length,
    };
  });
}

/** Data structure for re-order items stored in sessionStorage */
export interface ReorderData {
  bakeryId: string;
  items: { productId: string; productName: string; quantity: number }[];
}

const REORDER_STORAGE_KEY = 'reorder_items';

/** Store re-order items in sessionStorage and return the bakeryId for navigation */
export function storeReorderData(data: ReorderData): void {
  sessionStorage.setItem(REORDER_STORAGE_KEY, JSON.stringify(data));
}

/** Consume re-order items from sessionStorage (returns null if none) */
export function consumeReorderData(): ReorderData | null {
  const raw = sessionStorage.getItem(REORDER_STORAGE_KEY);
  if (!raw) return null;
  sessionStorage.removeItem(REORDER_STORAGE_KEY);
  try {
    return JSON.parse(raw) as ReorderData;
  } catch {
    return null;
  }
}
