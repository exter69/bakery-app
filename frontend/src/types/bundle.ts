/** Bundle type: specific contents listed or a surprise bag */
export type BundleType = 'compose' | 'surprise';

/** Bundle lifecycle status */
export type BundleStatus = 'draft' | 'published' | 'expired' | 'sold_out';

/** Reservation lifecycle status */
export type BundleReservationStatus = 'pending' | 'confirmed' | 'picked_up' | 'released' | 'cancelled';

/** A single item in a composé bundle */
export interface BundleItem {
  productId?: string;
  description: string;
  quantity: number;
}

/** A surplus bundle as returned by the API */
export interface Bundle {
  id: string;
  bakeryId: string;
  bakeryName: string;
  bakeryLatitude: number;
  bakeryLongitude: number;
  name: string;
  type: BundleType;
  photoUrl: string;
  description: string;
  estimatedValue: number; // cents
  originalPrice: number; // cents
  discountedPrice: number; // cents
  quantityTotal: number;
  quantityRemaining: number;
  pickupStartTime: string; // "HH:MM"
  pickupEndTime: string; // "HH:MM"
  publishedDate: string;
  expiresAt: string;
  status: BundleStatus;
  items: BundleItem[];
  createdAt: string;
}

/** A customer's reservation on a bundle */
export interface BundleReservation {
  id: string;
  bundleId: string;
  bundleName: string;
  status: BundleReservationStatus;
  createdAt: string;
}

/** Community impact metrics */
export interface BundleImpact {
  totalSaved: number;
  weightAvoided: number;
}

/** Filters for the bundle list */
export interface BundleFilters {
  type?: BundleType;
  pickupBefore?: string; // "HH:MM"
  maxDistance?: number; // meters, applied client-side
}

/** Request body for creating a new bundle (seller) */
export interface CreateBundleRequest {
  name: string;
  type: BundleType;
  photoUrl?: string;
  description?: string;
  estimatedValue?: number; // cents
  originalPrice: number; // cents
  discountedPrice: number; // cents
  quantityTotal: number;
  pickupStartTime: string; // "HH:MM"
  pickupEndTime: string; // "HH:MM"
  items?: BundleItem[];
}
