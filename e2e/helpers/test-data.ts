export const USERS = {
  admin: { username: 'admin', password: 'admin123', role: 'admin' },
  baker: { username: 'baker_jean', password: 'baker123', role: 'seller' },
  baker2: { username: 'baker_marie', password: 'baker123', role: 'seller' },
  customer: { username: 'alice', password: 'customer123', role: 'customer' },
  customer2: { username: 'bob', password: 'customer123', role: 'customer' },
} as const;

export const BAKERIES = {
  bakery1: { id: 'bakery-1', name: 'La Boulangerie du Coin' },
  bakery2: { id: 'bakery-2', name: 'Mie & Beurre' },
  bakery3: { id: 'bakery-3', name: 'Le Fournil de Max' },
} as const;

export const REGISTRATION_CODE = 'DEMO1234';

export const SEEDED_COUNTS = {
  bakeries: 3,
  products: 15,
  orders: 3,
  reservations: 2,
  recurringOrders: 2,
} as const;
