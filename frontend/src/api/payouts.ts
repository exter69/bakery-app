import { apiFetch } from './client';

export interface Payout {
  id: string;
  orderId: string;
  bakeryId: string;
  amount: number;
  commission: number;
  stripeTransferId?: string;
  status: 'pending' | 'transferred' | 'failed' | 'refunded';
  createdAt: string;
  transferredAt?: string;
}

export interface PayoutListResponse {
  items: Payout[];
  page: number;
  pageSize: number;
  total: number;
}

export interface ConnectStatus {
  connected: boolean;
  chargesEnabled: boolean;
  payoutsEnabled: boolean;
}

/** Fetch paginated payout history for the seller's bakery */
export function fetchPayouts(page = 1): Promise<PayoutListResponse> {
  const params = new URLSearchParams();
  if (page > 1) params.set('page', String(page));
  const query = params.toString();
  return apiFetch<PayoutListResponse>(`/seller/payouts${query ? `?${query}` : ''}`);
}

/** Get the current Stripe Connect status for the seller's bakery */
export function fetchConnectStatus(): Promise<ConnectStatus> {
  return apiFetch<ConnectStatus>('/seller/connect/status');
}

/** Start the Stripe Connect onboarding flow — returns a URL to redirect to */
export function startOnboarding(refreshUrl: string, returnUrl: string): Promise<{ url: string }> {
  return apiFetch<{ url: string }>('/seller/connect/onboard', {
    method: 'POST',
    body: JSON.stringify({ refreshUrl, returnUrl }),
  });
}
