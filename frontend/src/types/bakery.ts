/** Bakery card as returned by the list endpoint */
export interface BakeryCard {
  id: string;
  name: string;
  photoUrl: string;
  latitude: number;
  longitude: number;
  todaySchedule: TodaySchedule;
  distance?: number; // distance in km, only present when user location is available
}

/** Today's schedule for a bakery */
export interface TodaySchedule {
  openTime?: string;
  closeTime?: string;
  isOpen: boolean;
}

/** Full bakery info */
export interface Bakery {
  id: string;
  name: string;
  photoUrl: string;
  description: string;
  address: string;
  latitude?: number;
  longitude?: number;
  googlePlaceId?: string;
  minDeliveryAmount?: number; // in cents
  schedule: DaySchedule[];
  createdAt: string;
}

/** Schedule for a single day */
export interface DaySchedule {
  day: string;
  openTime: { hour: number; minute: number };
  closeTime: { hour: number; minute: number };
  isOpen: boolean;
}

/** Product in the bakery menu */
export interface Product {
  id: string;
  bakeryId: string;
  name: string;
  description: string;
  price: number; // in cents
  photoUrl: string;
  category: string;
  isAvailable: boolean;
  allergens: string[];
  healthScore: number | null;
}

/** Menu grouped by category */
export type Menu = Record<string, Product[]>;

/** Paginated list response */
export interface ListResponse<T> {
  items: T[];
  page: number;
  pageSize: number;
  total: number;
}
