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
