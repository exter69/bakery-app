import { apiFetch, API_BASE, getToken } from './client';
import type {
  BusinessProfile,
  DeliverySite,
  B2BAccess,
  B2BConfig,
  SavedList,
  B2BInvoice,
  B2BPricingResult,
} from '../types/b2b';
import type { Bakery, Product } from '../types/bakery';

// --- Registration ---

export function registerBusiness(data: {
  username: string;
  password: string;
  companyName: string;
  vatSiret: string;
  iban: string;
  billingEmail: string;
  billingContactName: string;
}): Promise<{ token: string; profile: BusinessProfile }> {
  return apiFetch('/comptoir/register', {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

// --- Profile ---

export function getProfile(): Promise<BusinessProfile> {
  return apiFetch('/comptoir/profile');
}

export function updateProfile(data: {
  companyName: string;
  iban: string;
  billingEmail: string;
  billingContactName: string;
}): Promise<BusinessProfile> {
  return apiFetch('/comptoir/profile', {
    method: 'PUT',
    body: JSON.stringify(data),
  });
}

// --- Delivery Sites ---

export function listSites(): Promise<DeliverySite[]> {
  return apiFetch('/comptoir/sites');
}

export function createSite(data: {
  name: string;
  streetAddress: string;
  city: string;
  postalCode: string;
  country: string;
  deliveryInstructions?: string;
}): Promise<DeliverySite> {
  return apiFetch('/comptoir/sites', {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

export function updateSite(
  id: string,
  data: Partial<Omit<DeliverySite, 'id' | 'userId' | 'createdAt' | 'updatedAt'>>
): Promise<DeliverySite> {
  return apiFetch(`/comptoir/sites/${id}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  });
}

export function deleteSite(id: string): Promise<void> {
  return apiFetch(`/comptoir/sites/${id}`, { method: 'DELETE' });
}

// --- Access ---

export function requestAccess(bakeryId: string): Promise<B2BAccess> {
  return apiFetch(`/comptoir/access/request/${bakeryId}`, { method: 'POST' });
}

export function listApprovedBakeries(): Promise<Bakery[]> {
  return apiFetch('/comptoir/bakeries');
}

// --- Products & Config ---

export function getProducts(bakeryId: string): Promise<Record<string, Product[]>> {
  return apiFetch(`/comptoir/bakeries/${bakeryId}/products`);
}

export function getConfig(bakeryId: string): Promise<B2BConfig> {
  return apiFetch(`/comptoir/bakeries/${bakeryId}/config`);
}

// --- Checkout ---

export function checkout(data: {
  bakeryId: string;
  deliverySiteId: string;
  items: { productId: string; quantity: number }[];
}): Promise<{ id: string; status: string }> {
  return apiFetch('/comptoir/checkout', {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

export function editOrder(
  orderId: string,
  items: { productId: string; quantity: number }[]
): Promise<{ id: string; status: string }> {
  return apiFetch(`/comptoir/orders/${orderId}`, {
    method: 'PUT',
    body: JSON.stringify({ items }),
  });
}

export function computePricing(data: {
  bakeryId: string;
  items: { productId: string; quantity: number }[];
}): Promise<B2BPricingResult> {
  return apiFetch('/comptoir/pricing', {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

// --- Saved Lists ---

export function listSavedLists(bakeryId: string): Promise<SavedList[]> {
  return apiFetch(`/comptoir/lists?bakeryId=${bakeryId}`);
}

export function createSavedList(data: {
  bakeryId: string;
  name: string;
  items: { productId: string; quantity: number }[];
}): Promise<SavedList> {
  return apiFetch('/comptoir/lists', {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

export function deleteSavedList(id: string): Promise<void> {
  return apiFetch(`/comptoir/lists/${id}`, { method: 'DELETE' });
}

// --- Deliveries ---

export function listDeliveries(params?: {
  bakeryId?: string;
  status?: string;
  dateFrom?: string;
  dateTo?: string;
  page?: number;
}): Promise<{ items: Array<{ id: string; bakeryId: string; status: string; items: { productId: string; quantity: number }[]; createdAt: string; deliverySiteId?: string; subtotalHt?: number; discountAmount?: number; tvaAmount?: number }>; page: number; pageSize: number; total: number }> {
  const searchParams = new URLSearchParams();
  if (params?.bakeryId) searchParams.set('bakeryId', params.bakeryId);
  if (params?.status) searchParams.set('status', params.status);
  if (params?.dateFrom) searchParams.set('dateFrom', params.dateFrom);
  if (params?.dateTo) searchParams.set('dateTo', params.dateTo);
  if (params?.page) searchParams.set('page', String(params.page));
  const qs = searchParams.toString();
  return apiFetch(`/comptoir/deliveries${qs ? `?${qs}` : ''}`);
}

export function getLastOrder(bakeryId: string): Promise<{ id: string; items: { productId: string; quantity: number }[]; createdAt: string } | null> {
  return apiFetch(`/comptoir/orders/${bakeryId}/last`);
}

// --- Invoices ---

export function listInvoices(page?: number): Promise<{ items: B2BInvoice[]; page: number; pageSize: number; total: number }> {
  return apiFetch(`/comptoir/invoices?page=${page ?? 1}`);
}

export function downloadInvoicePDF(invoiceId: string): Promise<Response> {
  const token = getToken();
  return fetch(`${API_BASE}/comptoir/invoices/${invoiceId}/pdf`, {
    headers: token ? { Authorization: `Bearer ${token}` } : {},
  });
}

// --- Baker-facing B2B Management ---

export function listAccessRequests(): Promise<B2BAccess[]> {
  return apiFetch('/dashboard/b2b/access');
}

export function approveAccess(accessId: string): Promise<void> {
  return apiFetch(`/dashboard/b2b/access/${accessId}/approve`, { method: 'POST' });
}

export function rejectAccess(accessId: string): Promise<void> {
  return apiFetch(`/dashboard/b2b/access/${accessId}/reject`, { method: 'POST' });
}

export function revokeAccess(accessId: string): Promise<void> {
  return apiFetch(`/dashboard/b2b/access/${accessId}/revoke`, { method: 'POST' });
}

export function getBakerConfig(): Promise<B2BConfig | null> {
  return apiFetch('/dashboard/b2b/config');
}

export function saveBakerConfig(data: {
  cutoffTime: string;
  deliveryWindowStart: string;
  deliveryWindowEnd: string;
  orderMinimum: number;
  proDiscount: number;
}): Promise<B2BConfig> {
  return apiFetch('/dashboard/b2b/config', {
    method: 'PUT',
    body: JSON.stringify(data),
  });
}
