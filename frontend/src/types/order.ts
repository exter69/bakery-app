/** Order status values */
export type OrderStatus =
  | 'pending_payment'
  | 'confirmed'
  | 'preparing'
  | 'ready'
  | 'delivered'
  | 'cancelled';

/** Reservation status values */
export type ReservationStatus = 'confirmed' | 'ready' | 'picked_up' | 'cancelled';

/** Combined status type for filtering */
export type ScheduleStatus = OrderStatus | ReservationStatus;

/** An item within an order or reservation */
export interface ScheduleItem {
  productId: string;
  productName: string;
  quantity: number;
  unitPrice: number;
  subtotal: number;
}

/** Time slot as returned by the API */
export interface TimeSlotResponse {
  startTime: string; // HH:MM format
  endTime: string;   // HH:MM format
}

/** An order or reservation entry in the schedule list */
export interface ScheduleEntry {
  id: string;
  type: 'order' | 'reservation';
  bakeryId: string;
  items: ScheduleItem[];
  scheduledDay: string;
  scheduledTime: TimeSlotResponse;
  status: OrderStatus | ReservationStatus;
  totalAmount: number;
  createdAt: string;
}

/** Query params for fetching schedule entries */
export interface ScheduleQueryParams {
  page?: number;
  status?: string;
  sortBy?: 'scheduledTime' | 'createdAt';
  sortDirection?: 'asc' | 'desc';
}

/** Paginated response for schedule entries */
export interface ScheduleListResponse {
  items: ScheduleEntry[];
  page: number;
  pageSize: number;
  total: number;
}
