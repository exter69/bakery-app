import { apiFetch } from './client';
import type { ListResponse } from '../types/bakery';
import type { Bundle, BundleImpact, BundleReservation, CreateBundleRequest } from '../types/bundle';

/** Fetch paginated published bundles with optional filters */
export function listBundles(params?: {
  page?: number;
  type?: string;
  pickupBefore?: string;
}): Promise<ListResponse<Bundle>> {
  const searchParams = new URLSearchParams();
  if (params?.page) searchParams.set('page', String(params.page));
  if (params?.type) searchParams.set('type', params.type);
  if (params?.pickupBefore) searchParams.set('pickupBefore', params.pickupBefore);
  const query = searchParams.toString();
  return apiFetch<ListResponse<Bundle>>(`/bundles${query ? `?${query}` : ''}`);
}

/** Fetch a single bundle by ID */
export function getBundle(id: string): Promise<Bundle> {
  return apiFetch<Bundle>(`/bundles/${id}`);
}

/** Reserve a bundle for the current user */
export function reserveBundle(id: string): Promise<BundleReservation> {
  return apiFetch<BundleReservation>(`/bundles/${id}/reserve`, { method: 'POST' });
}

/** Confirm a pending reservation */
export function confirmReservation(bundleId: string): Promise<BundleReservation> {
  return apiFetch<BundleReservation>(`/bundles/${bundleId}/reserve/confirm`, { method: 'POST' });
}

/** Cancel an active reservation */
export function cancelBundleReservation(id: string): Promise<void> {
  return apiFetch<void>(`/bundle-reservations/${id}`, { method: 'DELETE' });
}

/** Fetch community impact metrics for the current month */
export function getBundleImpact(): Promise<BundleImpact> {
  return apiFetch<BundleImpact>('/bundles/impact');
}

/** Create a new bundle (seller) */
export function createBundle(data: CreateBundleRequest): Promise<Bundle> {
  return apiFetch<Bundle>('/bundles', {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

/** Publish a draft bundle (seller) */
export function publishBundle(id: string): Promise<Bundle> {
  return apiFetch<Bundle>(`/bundles/${id}/publish`, { method: 'POST' });
}
