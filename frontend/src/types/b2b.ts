export interface BusinessProfile {
  id: string;
  userId: string;
  companyName: string;
  vatSiret: string;
  iban: string;
  billingEmail: string;
  billingContactName: string;
  createdAt: string;
  updatedAt: string;
}

export interface DeliverySite {
  id: string;
  userId: string;
  name: string;
  streetAddress: string;
  city: string;
  postalCode: string;
  country: string;
  deliveryInstructions?: string;
  createdAt: string;
  updatedAt: string;
}

export type B2BAccessStatus = 'pending' | 'approved' | 'rejected' | 'revoked';

export interface B2BAccess {
  id: string;
  bakeryId: string;
  businessUserId: string;
  status: B2BAccessStatus;
  createdAt: string;
  updatedAt: string;
}

export interface B2BConfig {
  id: string;
  bakeryId: string;
  cutoffTime: string;
  deliveryWindowStart: string;
  deliveryWindowEnd: string;
  orderMinimum: number;
  proDiscount: number;
}

export interface SavedList {
  id: string;
  userId: string;
  bakeryId: string;
  name: string;
  items: SavedListItem[];
  createdAt: string;
  updatedAt: string;
}

export interface SavedListItem {
  id: string;
  productId: string;
  quantity: number;
}

export interface B2BInvoice {
  id: string;
  orderId: string;
  bakeryId: string;
  businessProfileId: string;
  invoiceNumber: number;
  subtotalHt: number;
  discountAmount: number;
  tvaAmount: number;
  totalTtc: number;
  paymentStatus: 'pending' | 'paid' | 'overdue';
  issuedAt: string;
  paidAt?: string;
}

export interface B2BCartItem {
  productId: string;
  productName: string;
  unitPrice: number;
  quantity: number;
}

export interface B2BCartGroup {
  bakeryId: string;
  bakeryName: string;
  items: B2BCartItem[];
  subtotalHt: number;
}

export interface B2BCart {
  groups: B2BCartGroup[];
  totalHt: number;
}

export interface B2BOrderPricing {
  subtotalHt: number;
  discountRate: number;
  discountAmount: number;
  tvaRate: number;
  tvaAmount: number;
  totalTtc: number;
}
